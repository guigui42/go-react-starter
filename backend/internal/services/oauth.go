package services

import (
"context"
"encoding/json"
"errors"
"fmt"
"io"
"net/http"
"net/url"
"strings"
"time"

"github.com/golang-jwt/jwt/v5"
"github.com/google/uuid"
"github.com/example/go-react-starter/internal/config"
"github.com/example/go-react-starter/internal/models"
"gorm.io/gorm"
)

// Common errors for OAuth service.
var (
ErrOAuthProviderNotEnabled = errors.New("OAuth provider is not enabled")
ErrOAuthProviderInvalid    = errors.New("invalid OAuth provider")
ErrOAuthStateMismatch      = errors.New("OAuth state mismatch")
ErrOAuthStateExpired       = errors.New("OAuth state has expired")
ErrOAuthCodeExchange       = errors.New("failed to exchange OAuth code")
ErrOAuthUserInfo           = errors.New("failed to get user info from OAuth provider")
ErrOAuthEmailRequired      = errors.New("email is required from OAuth provider")
ErrOAuthAlreadyLinked      = errors.New("this OAuth account is already linked to another user")
ErrOAuthNotLinked          = errors.New("this OAuth provider is not linked to your account")
ErrOAuthLastAuthMethod     = errors.New("cannot unlink last authentication method")
)

type GitHubURLs struct {
AuthorizeURL string
TokenURL     string
UserURL      string
EmailsURL    string
}

func DefaultGitHubURLs() GitHubURLs {
return GitHubURLs{
AuthorizeURL: "https://github.com/login/oauth/authorize",
TokenURL:     "https://github.com/login/oauth/access_token",
UserURL:      "https://api.github.com/user",
EmailsURL:    "https://api.github.com/user/emails",
}
}

type GitHubUser struct {
ID    int64  `json:"id"`
Login string `json:"login"`
Email string `json:"email"`
Name  string `json:"name"`
}

type GitHubEmail struct {
Email    string `json:"email"`
Primary  bool   `json:"primary"`
Verified bool   `json:"verified"`
}

type GoogleURLs struct {
AuthorizeURL string
TokenURL     string
UserInfoURL  string
}

func DefaultGoogleURLs() GoogleURLs {
return GoogleURLs{
AuthorizeURL: "https://accounts.google.com/o/oauth2/v2/auth",
TokenURL:     "https://oauth2.googleapis.com/token",
UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
}
}

type GoogleUser struct {
ID            string `json:"id"`
Email         string `json:"email"`
VerifiedEmail bool   `json:"verified_email"`
Name          string `json:"name"`
Picture       string `json:"picture"`
}

type FacebookURLs struct {
AuthorizeURL string
TokenURL     string
UserInfoURL  string
}

func DefaultFacebookURLs() FacebookURLs {
return FacebookURLs{
AuthorizeURL: "https://www.facebook.com/v18.0/dialog/oauth",
TokenURL:     "https://graph.facebook.com/v18.0/oauth/access_token",
UserInfoURL:  "https://graph.facebook.com/me",
}
}

type FacebookUser struct {
ID    string `json:"id"`
Email string `json:"email"`
Name  string `json:"name"`
}

type OAuthService struct {
db           *gorm.DB
config       *config.OAuthConfig
authConfig   *config.AuthConfig
adminEmails  []string
githubURLs   GitHubURLs
googleURLs   GoogleURLs
facebookURLs FacebookURLs
httpClient   *http.Client
}

func NewOAuthService(db *gorm.DB, oauthConfig *config.OAuthConfig, authConfig *config.AuthConfig, adminEmails []string) *OAuthService {
normalized := make([]string, 0, len(adminEmails))
for _, e := range adminEmails {
if trimmed := strings.TrimSpace(strings.ToLower(e)); trimmed != "" {
normalized = append(normalized, trimmed)
}
}
return &OAuthService{
db:           db,
config:       oauthConfig,
authConfig:   authConfig,
adminEmails:  normalized,
githubURLs:   DefaultGitHubURLs(),
googleURLs:   DefaultGoogleURLs(),
facebookURLs: DefaultFacebookURLs(),
httpClient:   &http.Client{Timeout: 30 * time.Second},
}
}

func (s *OAuthService) isAdminEmail(email string) bool {
normalized := strings.TrimSpace(strings.ToLower(email))
for _, adminEmail := range s.adminEmails {
if adminEmail == normalized {
return true
}
}
return false
}

type OAuthStateClaims struct {
Provider    string `json:"provider"`
RedirectURI string `json:"redirect_uri,omitempty"`
jwt.RegisteredClaims
}

func (s *OAuthService) GenerateStateToken(provider models.OAuthProvider, redirectURI string) (string, error) {
now := time.Now()
claims := OAuthStateClaims{
Provider:    string(provider),
RedirectURI: redirectURI,
RegisteredClaims: jwt.RegisteredClaims{
ID:        models.NewID().String(),
IssuedAt:  jwt.NewNumericDate(now),
ExpiresAt: jwt.NewNumericDate(now.Add(s.config.StateDuration)),
},
}

token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
return token.SignedString([]byte(s.authConfig.JWTSecret))
}

func (s *OAuthService) ValidateStateToken(tokenString string, expectedProvider models.OAuthProvider) (*OAuthStateClaims, error) {
token, err := jwt.ParseWithClaims(tokenString, &OAuthStateClaims{}, func(token *jwt.Token) (interface{}, error) {
if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
return nil, jwt.ErrSignatureInvalid
}
return []byte(s.authConfig.JWTSecret), nil
})

if err != nil {
if errors.Is(err, jwt.ErrTokenExpired) {
return nil, ErrOAuthStateExpired
}
return nil, ErrOAuthStateMismatch
}

claims, ok := token.Claims.(*OAuthStateClaims)
if !ok || !token.Valid {
return nil, ErrOAuthStateMismatch
}

if claims.Provider != string(expectedProvider) {
return nil, ErrOAuthStateMismatch
}

return claims, nil
}

func (s *OAuthService) GetGitHubAuthURL(state string) (string, error) {
if !s.config.IsGitHubEnabled() {
return "", ErrOAuthProviderNotEnabled
}

params := url.Values{}
params.Set("client_id", s.config.GitHubClientID)
params.Set("redirect_uri", s.getCallbackURL(models.OAuthProviderGitHub))
params.Set("scope", "user:email")
params.Set("state", state)

return s.githubURLs.AuthorizeURL + "?" + params.Encode(), nil
}

func (s *OAuthService) ExchangeGitHubCode(ctx context.Context, code string) (*GitHubUser, string, error) {
if !s.config.IsGitHubEnabled() {
return nil, "", ErrOAuthProviderNotEnabled
}

params := url.Values{}
params.Set("client_id", s.config.GitHubClientID)
params.Set("client_secret", s.config.GitHubClientSecret)
params.Set("code", code)
params.Set("redirect_uri", s.getCallbackURL(models.OAuthProviderGitHub))

req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.githubURLs.TokenURL, strings.NewReader(params.Encode()))
if err != nil {
return nil, "", fmt.Errorf("%w: %v", ErrOAuthCodeExchange, err)
}
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
req.Header.Set("Accept", "application/json")

resp, err := s.httpClient.Do(req)
if err != nil {
return nil, "", fmt.Errorf("%w: %v", ErrOAuthCodeExchange, err)
}
defer resp.Body.Close()

body, err := io.ReadAll(resp.Body)
if err != nil {
return nil, "", fmt.Errorf("%w: failed to read response: %v", ErrOAuthCodeExchange, err)
}

var tokenResp struct {
AccessToken string `json:"access_token"`
TokenType   string `json:"token_type"`
Scope       string `json:"scope"`
Error       string `json:"error"`
ErrorDesc   string `json:"error_description"`
}
if err := json.Unmarshal(body, &tokenResp); err != nil {
return nil, "", fmt.Errorf("%w: failed to parse response: %v", ErrOAuthCodeExchange, err)
}

if tokenResp.Error != "" {
return nil, "", fmt.Errorf("%w: %s - %s", ErrOAuthCodeExchange, tokenResp.Error, tokenResp.ErrorDesc)
}

if tokenResp.AccessToken == "" {
return nil, "", fmt.Errorf("%w: no access token in response", ErrOAuthCodeExchange)
}

user, err := s.fetchGitHubUser(ctx, tokenResp.AccessToken)
if err != nil {
return nil, "", err
}

ghEmail, err := s.fetchGitHubPrimaryEmail(ctx, tokenResp.AccessToken)
if err != nil {
return nil, "", err
}

if user.Email == "" {
user.Email = ghEmail
}

return user, ghEmail, nil
}

func (s *OAuthService) fetchGitHubUser(ctx context.Context, accessToken string) (*GitHubUser, error) {
req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.githubURLs.UserURL, nil)
if err != nil {
return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
}
req.Header.Set("Authorization", "Bearer "+accessToken)
req.Header.Set("Accept", "application/json")

resp, err := s.httpClient.Do(req)
if err != nil {
return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
return nil, fmt.Errorf("%w: GitHub API returned status %d", ErrOAuthUserInfo, resp.StatusCode)
}

var user GitHubUser
if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
return nil, fmt.Errorf("%w: failed to parse user info: %v", ErrOAuthUserInfo, err)
}

return &user, nil
}

func (s *OAuthService) fetchGitHubPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.githubURLs.EmailsURL, nil)
if err != nil {
return "", fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
}
req.Header.Set("Authorization", "Bearer "+accessToken)
req.Header.Set("Accept", "application/json")

resp, err := s.httpClient.Do(req)
if err != nil {
return "", fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
return "", fmt.Errorf("%w: GitHub API returned status %d for emails", ErrOAuthUserInfo, resp.StatusCode)
}

var emails []GitHubEmail
if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
return "", fmt.Errorf("%w: failed to parse emails: %v", ErrOAuthUserInfo, err)
}

for _, email := range emails {
if email.Primary && email.Verified {
return email.Email, nil
}
}

for _, email := range emails {
if email.Verified {
return email.Email, nil
}
}

return "", ErrOAuthEmailRequired
}

func (s *OAuthService) GetGoogleAuthURL(state string) (string, error) {
if !s.config.IsGoogleEnabled() {
return "", ErrOAuthProviderNotEnabled
}

params := url.Values{}
params.Set("client_id", s.config.GoogleClientID)
params.Set("redirect_uri", s.getCallbackURL(models.OAuthProviderGoogle))
params.Set("response_type", "code")
params.Set("scope", "email profile")
params.Set("state", state)
params.Set("access_type", "online")

return s.googleURLs.AuthorizeURL + "?" + params.Encode(), nil
}

func (s *OAuthService) ExchangeGoogleCode(ctx context.Context, code string) (*GoogleUser, string, error) {
if !s.config.IsGoogleEnabled() {
return nil, "", ErrOAuthProviderNotEnabled
}

params := url.Values{}
params.Set("client_id", s.config.GoogleClientID)
params.Set("client_secret", s.config.GoogleClientSecret)
params.Set("code", code)
params.Set("grant_type", "authorization_code")
params.Set("redirect_uri", s.getCallbackURL(models.OAuthProviderGoogle))

req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.googleURLs.TokenURL, strings.NewReader(params.Encode()))
if err != nil {
return nil, "", fmt.Errorf("%w: %v", ErrOAuthCodeExchange, err)
}
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
req.Header.Set("Accept", "application/json")

resp, err := s.httpClient.Do(req)
if err != nil {
return nil, "", fmt.Errorf("%w: %v", ErrOAuthCodeExchange, err)
}
defer resp.Body.Close()

body, err := io.ReadAll(resp.Body)
if err != nil {
return nil, "", fmt.Errorf("%w: failed to read response: %v", ErrOAuthCodeExchange, err)
}

var tokenResp struct {
AccessToken string `json:"access_token"`
TokenType   string `json:"token_type"`
ExpiresIn   int    `json:"expires_in"`
Scope       string `json:"scope"`
Error       string `json:"error"`
ErrorDesc   string `json:"error_description"`
}
if err := json.Unmarshal(body, &tokenResp); err != nil {
return nil, "", fmt.Errorf("%w: failed to parse response: %v", ErrOAuthCodeExchange, err)
}

if tokenResp.Error != "" {
return nil, "", fmt.Errorf("%w: %s - %s", ErrOAuthCodeExchange, tokenResp.Error, tokenResp.ErrorDesc)
}

if tokenResp.AccessToken == "" {
return nil, "", fmt.Errorf("%w: no access token in response", ErrOAuthCodeExchange)
}

user, err := s.fetchGoogleUser(ctx, tokenResp.AccessToken)
if err != nil {
return nil, "", err
}

if user.Email == "" {
return nil, "", ErrOAuthEmailRequired
}

return user, user.Email, nil
}

func (s *OAuthService) fetchGoogleUser(ctx context.Context, accessToken string) (*GoogleUser, error) {
req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.googleURLs.UserInfoURL, nil)
if err != nil {
return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
}
req.Header.Set("Authorization", "Bearer "+accessToken)
req.Header.Set("Accept", "application/json")

resp, err := s.httpClient.Do(req)
if err != nil {
return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
return nil, fmt.Errorf("%w: Google API returned status %d", ErrOAuthUserInfo, resp.StatusCode)
}

var user GoogleUser
if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
return nil, fmt.Errorf("%w: failed to parse user info: %v", ErrOAuthUserInfo, err)
}

return &user, nil
}

func (s *OAuthService) GetFacebookAuthURL(state string) (string, error) {
if !s.config.IsFacebookEnabled() {
return "", ErrOAuthProviderNotEnabled
}

params := url.Values{}
params.Set("client_id", s.config.FacebookClientID)
params.Set("redirect_uri", s.getCallbackURL(models.OAuthProviderFacebook))
params.Set("scope", "email,public_profile")
params.Set("state", state)
params.Set("response_type", "code")

return s.facebookURLs.AuthorizeURL + "?" + params.Encode(), nil
}

func (s *OAuthService) ExchangeFacebookCode(ctx context.Context, code string) (*FacebookUser, string, error) {
if !s.config.IsFacebookEnabled() {
return nil, "", ErrOAuthProviderNotEnabled
}

params := url.Values{}
params.Set("client_id", s.config.FacebookClientID)
params.Set("client_secret", s.config.FacebookClientSecret)
params.Set("code", code)
params.Set("redirect_uri", s.getCallbackURL(models.OAuthProviderFacebook))

tokenURL := s.facebookURLs.TokenURL + "?" + params.Encode()
req, err := http.NewRequestWithContext(ctx, http.MethodGet, tokenURL, nil)
if err != nil {
return nil, "", fmt.Errorf("%w: %v", ErrOAuthCodeExchange, err)
}
req.Header.Set("Accept", "application/json")

resp, err := s.httpClient.Do(req)
if err != nil {
return nil, "", fmt.Errorf("%w: %v", ErrOAuthCodeExchange, err)
}
defer resp.Body.Close()

body, err := io.ReadAll(resp.Body)
if err != nil {
return nil, "", fmt.Errorf("%w: failed to read response: %v", ErrOAuthCodeExchange, err)
}

var tokenResp struct {
AccessToken string `json:"access_token"`
TokenType   string `json:"token_type"`
ExpiresIn   int    `json:"expires_in"`
Error       struct {
Message string `json:"message"`
Type    string `json:"type"`
Code    int    `json:"code"`
} `json:"error"`
}
if err := json.Unmarshal(body, &tokenResp); err != nil {
return nil, "", fmt.Errorf("%w: failed to parse response: %v", ErrOAuthCodeExchange, err)
}

if tokenResp.Error.Message != "" {
return nil, "", fmt.Errorf("%w: %s - %s", ErrOAuthCodeExchange, tokenResp.Error.Type, tokenResp.Error.Message)
}

if tokenResp.AccessToken == "" {
return nil, "", fmt.Errorf("%w: no access token in response", ErrOAuthCodeExchange)
}

user, err := s.fetchFacebookUser(ctx, tokenResp.AccessToken)
if err != nil {
return nil, "", err
}

if user.Email == "" {
return nil, "", ErrOAuthEmailRequired
}

return user, user.Email, nil
}

func (s *OAuthService) fetchFacebookUser(ctx context.Context, accessToken string) (*FacebookUser, error) {
userURL := s.facebookURLs.UserInfoURL + "?fields=id,email,name&access_token=" + url.QueryEscape(accessToken)
req, err := http.NewRequestWithContext(ctx, http.MethodGet, userURL, nil)
if err != nil {
return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
}
req.Header.Set("Accept", "application/json")

resp, err := s.httpClient.Do(req)
if err != nil {
return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
return nil, fmt.Errorf("%w: Facebook API returned status %d", ErrOAuthUserInfo, resp.StatusCode)
}

var user FacebookUser
if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
return nil, fmt.Errorf("%w: failed to parse user info: %v", ErrOAuthUserInfo, err)
}

return &user, nil
}

type OAuthLoginResult struct {
User      *models.User
IsNewUser bool
Linked    bool
}

func (s *OAuthService) HandleOAuthCallback(ctx context.Context, provider models.OAuthProvider, providerUserID string, email string) (*OAuthLoginResult, error) {
if !provider.IsValid() {
return nil, ErrOAuthProviderInvalid
}

if email == "" {
return nil, ErrOAuthEmailRequired
}

email = strings.ToLower(strings.TrimSpace(email))

var result OAuthLoginResult

err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
var existingOAuth models.UserOAuthAccount
if err := tx.Where("provider = ? AND provider_user_id = ?", provider, providerUserID).First(&existingOAuth).Error; err == nil {
var user models.User
if err := tx.First(&user, "id = ?", existingOAuth.UserID).Error; err != nil {
return fmt.Errorf("failed to load user: %w", err)
}
result.User = &user
result.IsNewUser = false
result.Linked = false
return nil
} else if !errors.Is(err, gorm.ErrRecordNotFound) {
return fmt.Errorf("failed to check OAuth account: %w", err)
}

var existingUser models.User
if err := tx.Where("email = ?", email).First(&existingUser).Error; err == nil {
oauthAccount := models.UserOAuthAccount{
UserID:         existingUser.ID,
Provider:       provider,
ProviderUserID: providerUserID,
ProviderEmail:  email,
}
if err := tx.Create(&oauthAccount).Error; err != nil {
if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
return ErrOAuthAlreadyLinked
}
return fmt.Errorf("failed to link OAuth account: %w", err)
}

if !existingUser.EmailVerified {
now := time.Now()
existingUser.EmailVerified = true
existingUser.EmailVerifiedAt = &now
if err := tx.Save(&existingUser).Error; err != nil {
return fmt.Errorf("failed to update user email verification: %w", err)
}
}

result.User = &existingUser
result.IsNewUser = false
result.Linked = true
return nil
} else if !errors.Is(err, gorm.ErrRecordNotFound) {
return fmt.Errorf("failed to check existing user: %w", err)
}

now := time.Now()
newUser := models.User{
Email:           email,
PasswordHash:    nil,
EmailVerified:   true,
EmailVerifiedAt: &now,
IsAdmin:         s.isAdminEmail(email),
}
if err := tx.Create(&newUser).Error; err != nil {
return fmt.Errorf("failed to create user: %w", err)
}

prefs := models.UserPreferences{
UserID:   newUser.ID,
Language: "en",
}
if err := tx.Create(&prefs).Error; err != nil {
return fmt.Errorf("failed to create preferences: %w", err)
}

oauthAccount := models.UserOAuthAccount{
UserID:         newUser.ID,
Provider:       provider,
ProviderUserID: providerUserID,
ProviderEmail:  email,
}
if err := tx.Create(&oauthAccount).Error; err != nil {
return fmt.Errorf("failed to link OAuth account: %w", err)
}

result.User = &newUser
result.IsNewUser = true
result.Linked = true
return nil
})

if err != nil {
return nil, err
}

return &result, nil
}

func (s *OAuthService) LinkOAuthAccount(ctx context.Context, userID uuid.UUID, provider models.OAuthProvider, providerUserID string, email string) error {
if !provider.IsValid() {
return ErrOAuthProviderInvalid
}

var existingOAuth models.UserOAuthAccount
if err := s.db.WithContext(ctx).Where("provider = ? AND provider_user_id = ?", provider, providerUserID).First(&existingOAuth).Error; err == nil {
if existingOAuth.UserID != userID {
return ErrOAuthAlreadyLinked
}
return nil
} else if !errors.Is(err, gorm.ErrRecordNotFound) {
return fmt.Errorf("failed to check OAuth account: %w", err)
}

oauthAccount := models.UserOAuthAccount{
UserID:         userID,
Provider:       provider,
ProviderUserID: providerUserID,
ProviderEmail:  strings.ToLower(strings.TrimSpace(email)),
}
if err := s.db.WithContext(ctx).Create(&oauthAccount).Error; err != nil {
if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
return ErrOAuthAlreadyLinked
}
return fmt.Errorf("failed to link OAuth account: %w", err)
}

return nil
}

func (s *OAuthService) UnlinkOAuthAccount(ctx context.Context, userID uuid.UUID, provider models.OAuthProvider) error {
if !provider.IsValid() {
return ErrOAuthProviderInvalid
}

return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
var oauthAccount models.UserOAuthAccount
if err := tx.Where("user_id = ? AND provider = ?", userID, provider).First(&oauthAccount).Error; err != nil {
if errors.Is(err, gorm.ErrRecordNotFound) {
return ErrOAuthNotLinked
}
return fmt.Errorf("failed to find OAuth account: %w", err)
}

hasOtherAuthMethods, err := s.hasOtherAuthMethods(tx, userID, provider)
if err != nil {
return err
}

if !hasOtherAuthMethods {
return ErrOAuthLastAuthMethod
}

if err := tx.Delete(&oauthAccount).Error; err != nil {
return fmt.Errorf("failed to unlink OAuth account: %w", err)
}

return nil
})
}

func (s *OAuthService) hasOtherAuthMethods(tx *gorm.DB, userID uuid.UUID, excludeProvider models.OAuthProvider) (bool, error) {
var user models.User
if err := tx.First(&user, "id = ?", userID).Error; err != nil {
return false, fmt.Errorf("failed to load user: %w", err)
}
if user.PasswordHash != nil && *user.PasswordHash != "" {
return true, nil
}

var passkeyCount int64
if err := tx.Model(&models.UserCredential{}).Where("user_id = ?", userID).Count(&passkeyCount).Error; err != nil {
return false, fmt.Errorf("failed to count passkeys: %w", err)
}
if passkeyCount > 0 {
return true, nil
}

var otherOAuthCount int64
if err := tx.Model(&models.UserOAuthAccount{}).Where("user_id = ? AND provider != ?", userID, excludeProvider).Count(&otherOAuthCount).Error; err != nil {
return false, fmt.Errorf("failed to count other OAuth accounts: %w", err)
}
if otherOAuthCount > 0 {
return true, nil
}

return false, nil
}

func (s *OAuthService) GetLinkedOAuthAccounts(ctx context.Context, userID uuid.UUID) ([]models.UserOAuthAccount, error) {
var accounts []models.UserOAuthAccount
if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
return nil, fmt.Errorf("failed to fetch OAuth accounts: %w", err)
}
return accounts, nil
}

func (s *OAuthService) GetEnabledProviders() []models.OAuthProviderInfo {
return []models.OAuthProviderInfo{
{Name: string(models.OAuthProviderGitHub), Enabled: s.config.IsGitHubEnabled()},
{Name: string(models.OAuthProviderGoogle), Enabled: s.config.IsGoogleEnabled()},
{Name: string(models.OAuthProviderFacebook), Enabled: s.config.IsFacebookEnabled()},
}
}

func (s *OAuthService) getCallbackURL(provider models.OAuthProvider) string {
return fmt.Sprintf("%s/auth/oauth/%s/callback", s.config.CallbackBaseURL, provider)
}

func (s *OAuthService) IsProviderEnabled(provider models.OAuthProvider) bool {
switch provider {
case models.OAuthProviderGitHub:
return s.config.IsGitHubEnabled()
case models.OAuthProviderGoogle:
return s.config.IsGoogleEnabled()
case models.OAuthProviderFacebook:
return s.config.IsFacebookEnabled()
default:
return false
}
}
