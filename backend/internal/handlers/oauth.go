// Package handlers provides HTTP request handlers for the API.
package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/guigui42/go-react-starter/internal/config"
	"github.com/guigui42/go-react-starter/internal/middleware"
	"github.com/guigui42/go-react-starter/internal/models"
	"github.com/guigui42/go-react-starter/internal/services"
	"github.com/guigui42/go-react-starter/pkg/response"
	"github.com/rs/zerolog"
)

const (
	// OAuthStateCookieName is the name of the cookie used to store the OAuth state.
	OAuthStateCookieName = "oauth_state"
)

// OAuthHandler handles OAuth-related HTTP requests.
type OAuthHandler struct {
	oauthService *services.OAuthService
	auditService *services.AuditService
	authCfg      *config.AuthConfig
	serverCfg    *config.ServerConfig
	emailCfg     *config.EmailConfig
}

// NewOAuthHandler creates a new OAuth handler.
func NewOAuthHandler(oauthService *services.OAuthService, auditService *services.AuditService, authCfg *config.AuthConfig, serverCfg *config.ServerConfig, emailCfg *config.EmailConfig) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
		auditService: auditService,
		authCfg:      authCfg,
		serverCfg:    serverCfg,
		emailCfg:     emailCfg,
	}
}

// ListProviders returns the list of enabled OAuth providers.
//
// @Summary      List OAuth providers
// @Description  Returns the list of enabled OAuth providers for frontend feature flags
// @Tags         OAuth
// @Produce      json
// @Success      200 {object} models.OAuthProvidersResponse "List of OAuth providers"
// @Router       /auth/oauth/providers [get]
func (h *OAuthHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.oauthService.GetEnabledProviders()
	response.Success(w, http.StatusOK, models.OAuthProvidersResponse{
		Providers: providers,
	})
}

// BeginOAuth initiates the OAuth flow for a provider.
// It generates a state JWT, stores it in an httpOnly cookie, and redirects to the provider.
//
// @Summary      Begin OAuth flow
// @Description  Initiates OAuth authentication by redirecting to the provider's authorization URL
// @Tags         OAuth
// @Param        provider path string true "OAuth provider (github, google, facebook)"
// @Success      302 "Redirect to OAuth provider"
// @Failure      400 {object} response.ErrorResponse "Invalid provider"
// @Failure      500 {object} response.ErrorResponse "Failed to generate state token"
// @Router       /auth/oauth/{provider} [get]
func (h *OAuthHandler) BeginOAuth(w http.ResponseWriter, r *http.Request) {
	providerStr := chi.URLParam(r, "provider")
	provider := models.OAuthProvider(providerStr)

	if !provider.IsValid() {
		response.BadRequest(w, "Invalid OAuth provider", map[string]interface{}{
			"provider": providerStr,
		})
		return
	}

	if !h.oauthService.IsProviderEnabled(provider) {
		response.BadRequest(w, "OAuth provider is not enabled", map[string]interface{}{
			"provider": providerStr,
		})
		return
	}

	// Validate redirect_uri against allowlist to prevent open redirect attacks
	redirectURI := r.URL.Query().Get("redirect_uri")
	if redirectURI != "" && !h.isAllowedRedirectURI(redirectURI) {
		response.BadRequest(w, "Invalid redirect_uri", nil)
		return
	}

	// Generate state token
	state, err := h.oauthService.GenerateStateToken(provider, redirectURI)
	if err != nil {
		response.InternalServerError(w, "Failed to generate state token")
		return
	}

	// Store state in httpOnly cookie
	h.setOAuthStateCookie(w, state)

	// Get authorization URL
	var authURL string
	switch provider {
	case models.OAuthProviderGitHub:
		authURL, err = h.oauthService.GetGitHubAuthURL(state)
	case models.OAuthProviderGoogle:
		authURL, err = h.oauthService.GetGoogleAuthURL(state)
	case models.OAuthProviderFacebook:
		authURL, err = h.oauthService.GetFacebookAuthURL(state)
	default:
		response.BadRequest(w, "OAuth provider not yet implemented", map[string]interface{}{
			"provider": providerStr,
		})
		return
	}

	if err != nil {
		if errors.Is(err, services.ErrOAuthProviderNotEnabled) {
			response.BadRequest(w, "OAuth provider is not enabled", nil)
			return
		}
		response.InternalServerError(w, "Failed to generate authorization URL")
		return
	}

	// Redirect to provider
	http.Redirect(w, r, authURL, http.StatusFound)
}

// OAuthCallback handles the OAuth callback from the provider.
// It validates the state, exchanges the code for tokens, and logs in or creates the user.
//
// @Summary      OAuth callback
// @Description  Handles the OAuth callback, exchanges code for tokens, and authenticates the user
// @Tags         OAuth
// @Param        provider path string true "OAuth provider (github, google, facebook)"
// @Param        code query string true "Authorization code from OAuth provider"
// @Param        state query string true "State parameter for CSRF protection"
// @Success      302 "Redirect to frontend with success or error"
// @Failure      400 {object} response.ErrorResponse "Invalid state or code"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /auth/oauth/{provider}/callback [get]
func (h *OAuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	providerStr := chi.URLParam(r, "provider")
	provider := models.OAuthProvider(providerStr)

	if !provider.IsValid() {
		h.redirectWithError(w, r, "invalid_provider", "Invalid OAuth provider")
		return
	}

	// Get OAuth parameters
	code := r.URL.Query().Get("code")
	stateParam := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")
	errorDesc := r.URL.Query().Get("error_description")

	// Check for error from provider
	if errorParam != "" {
		h.redirectWithError(w, r, errorParam, errorDesc)
		return
	}

	if code == "" {
		h.redirectWithError(w, r, "missing_code", "No authorization code received")
		return
	}

	// Get state from cookie
	stateCookie, err := r.Cookie(OAuthStateCookieName)
	if err != nil || stateCookie.Value == "" {
		h.redirectWithError(w, r, "invalid_state", "OAuth state cookie not found")
		return
	}

	// Validate state
	if stateParam != stateCookie.Value {
		h.redirectWithError(w, r, "invalid_state", "OAuth state mismatch")
		return
	}

	// Validate state token
	_, err = h.oauthService.ValidateStateToken(stateParam, provider)
	if err != nil {
		if errors.Is(err, services.ErrOAuthStateExpired) {
			h.redirectWithError(w, r, "state_expired", "OAuth state has expired, please try again")
			return
		}
		h.redirectWithError(w, r, "invalid_state", "Invalid OAuth state")
		return
	}

	// Clear state cookie
	h.clearOAuthStateCookie(w)

	// Exchange code for user info
	var providerUserID string
	var email string

	switch provider {
	case models.OAuthProviderGitHub:
		user, primaryEmail, err := h.oauthService.ExchangeGitHubCode(r.Context(), code)
		if err != nil {
			if errors.Is(err, services.ErrOAuthEmailRequired) {
				h.redirectWithError(w, r, "email_required", "Email is required from GitHub. Please ensure your email is public or verify your GitHub email.")
				return
			}
			h.redirectWithError(w, r, "exchange_failed", "Failed to authenticate with GitHub")
			return
		}
		providerUserID = fmt.Sprintf("%d", user.ID)
		email = primaryEmail
	case models.OAuthProviderGoogle:
		user, primaryEmail, err := h.oauthService.ExchangeGoogleCode(r.Context(), code)
		if err != nil {
			if errors.Is(err, services.ErrOAuthEmailRequired) {
				h.redirectWithError(w, r, "email_required", "Email is required from Google. Please ensure your Google account has a verified email.")
				return
			}
			h.redirectWithError(w, r, "exchange_failed", "Failed to authenticate with Google")
			return
		}
		providerUserID = user.ID
		email = primaryEmail
	case models.OAuthProviderFacebook:
		user, primaryEmail, err := h.oauthService.ExchangeFacebookCode(r.Context(), code)
		if err != nil {
			if errors.Is(err, services.ErrOAuthEmailRequired) {
				h.redirectWithError(w, r, "email_required", "Email is required from Facebook. Please ensure your Facebook account has a verified email.")
				return
			}
			h.redirectWithError(w, r, "exchange_failed", "Failed to authenticate with Facebook")
			return
		}
		providerUserID = user.ID
		email = primaryEmail
	default:
		h.redirectWithError(w, r, "not_implemented", "OAuth provider not yet implemented")
		return
	}

	// Handle OAuth callback (create/link user)
	result, err := h.oauthService.HandleOAuthCallback(r.Context(), provider, providerUserID, email)
	if err != nil {
		// Log failed OAuth login
		if h.auditService != nil {
			_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
				EventType: "auth.oauth.login",
				Action:    "Failed OAuth authentication",
				Status:    "failure",
				IPAddress: services.GetClientIP(r),
				UserAgent: services.GetUserAgent(r),
				Metadata: map[string]interface{}{
					"provider": string(provider),
					"email":    email,
				},
			})
		}
		if errors.Is(err, services.ErrOAuthAlreadyLinked) {
			h.redirectWithError(w, r, "already_linked", "This OAuth account is already linked to another user")
			return
		}
		h.redirectWithError(w, r, "auth_failed", "Failed to authenticate")
		return
	}

	// Generate JWT token for the user
	token, err := generateJWT(result.User.ID.String(), h.authCfg.JWTSecret, h.authCfg.SessionDuration)
	if err != nil {
		h.redirectWithError(w, r, "token_failed", "Failed to generate authentication token")
		return
	}

	// Set auth cookie
	setAuthCookie(w, token, h.authCfg.SessionDuration, h.serverCfg.IsSecure())

	// Log successful OAuth login
	if h.auditService != nil {
		_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
			EventType: "auth.oauth.login",
			ActorID:   &result.User.ID,
			Action:    "User logged in via OAuth",
			Status:    "success",
			IPAddress: services.GetClientIP(r),
			UserAgent: services.GetUserAgent(r),
			Metadata: map[string]interface{}{
				"provider":    string(provider),
				"is_new_user": result.IsNewUser,
			},
		})
	}

	// Redirect to frontend with success
	h.redirectWithSuccess(w, r, result.IsNewUser)
}

// ListLinkedAccounts returns the OAuth accounts linked to the current user.
//
// @Summary      List linked OAuth accounts
// @Description  Returns all OAuth accounts linked to the current user
// @Tags         OAuth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.UserOAuthAccountsResponse "List of linked OAuth accounts"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /api/v1/auth/oauth/accounts [get]
func (h *OAuthHandler) ListLinkedAccounts(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User not found in context")
		return
	}

	accounts, err := h.oauthService.GetLinkedOAuthAccounts(r.Context(), user.ID)
	if err != nil {
		response.InternalServerError(w, "Failed to fetch linked accounts")
		return
	}

	// Convert to response format
	accountInfos := make([]models.UserOAuthAccountInfo, len(accounts))
	for i, account := range accounts {
		accountInfos[i] = account.ToInfo()
	}

	response.Success(w, http.StatusOK, models.UserOAuthAccountsResponse{
		Accounts: accountInfos,
	})
}

// UnlinkOAuth removes an OAuth provider link from the current user's account.
//
// @Summary      Unlink OAuth provider
// @Description  Removes the specified OAuth provider from the current user's account. Requires at least one other authentication method.
// @Tags         OAuth
// @Param        provider path string true "OAuth provider to unlink (github, google, facebook)"
// @Security     BearerAuth
// @Success      204 "OAuth provider unlinked successfully"
// @Failure      400 {object} response.ErrorResponse "Invalid provider or last auth method"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      404 {object} response.ErrorResponse "OAuth provider not linked"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /api/v1/auth/oauth/{provider} [delete]
func (h *OAuthHandler) UnlinkOAuth(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User not found in context")
		return
	}

	providerStr := chi.URLParam(r, "provider")
	provider := models.OAuthProvider(providerStr)

	if !provider.IsValid() {
		response.BadRequest(w, "Invalid OAuth provider", map[string]interface{}{
			"provider": providerStr,
		})
		return
	}

	err := h.oauthService.UnlinkOAuthAccount(r.Context(), user.ID, provider)
	if err != nil {
		if errors.Is(err, services.ErrOAuthNotLinked) {
			response.NotFound(w, "OAuth provider is not linked to your account")
			return
		}
		if errors.Is(err, services.ErrOAuthLastAuthMethod) {
			response.BadRequest(w, "Cannot unlink last authentication method. Please set up a password or passkey first.", nil)
			return
		}
		response.InternalServerError(w, "Failed to unlink OAuth provider")
		return
	}

	// Log OAuth unlink
	if h.auditService != nil {
		_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
			EventType: "auth.oauth.unlink",
			ActorID:   &user.ID,
			Action:    "User unlinked OAuth provider",
			Status:    "success",
			IPAddress: services.GetClientIP(r),
			UserAgent: services.GetUserAgent(r),
			Metadata: map[string]interface{}{
				"provider": string(provider),
			},
		})
	}

	w.WriteHeader(http.StatusNoContent)
}

// BeginOAuthLink initiates OAuth flow for linking a provider to an existing authenticated user.
// This is similar to BeginOAuth but includes the user ID in the state for linking.
//
// @Summary      Begin OAuth linking flow
// @Description  Initiates OAuth flow to link a new provider to the current user's account
// @Tags         OAuth
// @Param        provider path string true "OAuth provider (github, google, facebook)"
// @Security     BearerAuth
// @Success      302 "Redirect to OAuth provider"
// @Failure      400 {object} response.ErrorResponse "Invalid provider"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /api/v1/auth/oauth/{provider}/link [get]
func (h *OAuthHandler) BeginOAuthLink(w http.ResponseWriter, r *http.Request) {
	// This uses the same flow as BeginOAuth, but with authentication required
	// The linking is handled in the callback based on whether the user is authenticated
	h.BeginOAuth(w, r)
}

// setOAuthStateCookie sets the OAuth state in an httpOnly cookie.
func (h *OAuthHandler) setOAuthStateCookie(w http.ResponseWriter, state string) {
	secure := h.serverCfg.IsSecure()
	http.SetCookie(w, &http.Cookie{
		Name:     OAuthStateCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode, // Lax required for OAuth redirect (Strict blocks cross-site redirects)
	})
}

// clearOAuthStateCookie clears the OAuth state cookie.
func (h *OAuthHandler) clearOAuthStateCookie(w http.ResponseWriter) {
	secure := h.serverCfg.IsSecure()
	http.SetCookie(w, &http.Cookie{
		Name:     OAuthStateCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// redirectWithError redirects to the frontend with an error code.
// The error message is logged for debugging but not sent to the frontend.
// The frontend maps known error codes to localized messages via i18n.
func (h *OAuthHandler) redirectWithError(w http.ResponseWriter, r *http.Request, errorCode, errorMsg string) {
	// Clear state cookie on error
	h.clearOAuthStateCookie(w)

	if logger := middleware.GetLogger(r.Context()); logger != nil {
		logger.WithMetadata(zerolog.WarnLevel, "OAuth error redirect", map[string]interface{}{
			"error_code": errorCode,
			"error_msg":  errorMsg,
		})
	}

	frontendURL := h.getFrontendURL()
	redirectURL := fmt.Sprintf("%s/auth/callback?error=%s&error_description=%s",
		frontendURL,
		url.QueryEscape(errorCode),
		url.QueryEscape(errorMsg),
	)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// redirectWithSuccess redirects to the frontend with success.
func (h *OAuthHandler) redirectWithSuccess(w http.ResponseWriter, r *http.Request, isNewUser bool) {
	frontendURL := h.getFrontendURL()
	redirectURL := fmt.Sprintf("%s/auth/callback?success=true&new_user=%t", frontendURL, isNewUser)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// getFrontendURL returns the frontend URL from config.
// This is where the backend redirects after OAuth processing.
func (h *OAuthHandler) getFrontendURL() string {
	if h.emailCfg.BaseURL != "" {
		return h.emailCfg.BaseURL
	}
	return "http://localhost:5173"
}

// CheckAuthMethodsWithOAuthResponse adds OAuth info to the existing auth methods response.
type CheckAuthMethodsWithOAuthResponse struct {
	HasPassword    bool     `json:"has_password"`
	HasPasskey     bool     `json:"has_passkey"`
	HasOAuth       bool     `json:"has_oauth"`
	OAuthProviders []string `json:"oauth_providers,omitempty"`
}

// isAllowedRedirectURI validates a redirect URI against the allowlist of permitted origins.
// Only URIs that start with the configured frontend URL or relative paths are accepted.
func (h *OAuthHandler) isAllowedRedirectURI(uri string) bool {
	// Reject protocol-relative URLs (e.g., "//evil.com") which could redirect to external domains
	if strings.HasPrefix(uri, "//") {
		return false
	}

	// Allow relative paths (e.g., "/dashboard", "/settings")
	if strings.HasPrefix(uri, "/") {
		return true
	}

	// Check against allowed origins
	allowedOrigins := []string{
		h.getFrontendURL(),
	}

	for _, origin := range allowedOrigins {
		normalizedOrigin := strings.TrimRight(origin, "/")
		if normalizedOrigin == "" {
			continue
		}
		// Exact match
		if uri == normalizedOrigin {
			return true
		}
		// Prefix match with path boundary (origin + "/")
		if strings.HasPrefix(uri, normalizedOrigin+"/") {
			return true
		}
	}

	return false
}
