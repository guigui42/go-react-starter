package services

import (
"context"
"errors"
"fmt"
"strings"
"sync"
"time"

"github.com/google/uuid"
"github.com/example/go-react-starter/internal/config"
"github.com/example/go-react-starter/internal/models"
"github.com/example/go-react-starter/internal/providers/email"
"github.com/example/go-react-starter/internal/repository/scopes"
"gorm.io/gorm"
)

// Common errors for email service.
var (
ErrTokenNotFound    = errors.New("verification token not found")
ErrTokenExpired     = errors.New("verification token has expired")
ErrTokenAlreadyUsed = errors.New("verification token has already been used")
ErrResendRateLimit  = errors.New("please wait before requesting another verification email")
)

// EmailService handles email sending and verification operations.
type EmailService struct {
db              *gorm.DB
provider        email.EmailProvider
templateService *EmailTemplateService
config          *config.EmailConfig
resendLimiter   map[uuid.UUID]time.Time
limiterMu       sync.RWMutex
}

// NewEmailService creates a new email service.
func NewEmailService(
db *gorm.DB,
provider email.EmailProvider,
templateService *EmailTemplateService,
emailConfig *config.EmailConfig,
) *EmailService {
return &EmailService{
db:              db,
provider:        provider,
templateService: templateService,
config:          emailConfig,
resendLimiter:   make(map[uuid.UUID]time.Time),
}
}

// SendVerificationEmail sends a verification email to the user.
func (s *EmailService) SendVerificationEmail(ctx context.Context, user *models.User, language string) error {
verification, err := models.NewEmailVerification(user.ID)
if err != nil {
return fmt.Errorf("failed to create verification token: %w", err)
}

if err := s.db.WithContext(ctx).Where("user_id = ?", user.ID).Delete(&models.EmailVerification{}).Error; err != nil {
return fmt.Errorf("failed to clean up old tokens: %w", err)
}

if err := s.db.WithContext(ctx).Create(verification).Error; err != nil {
return fmt.Errorf("failed to save verification token: %w", err)
}

verificationURL := fmt.Sprintf("%s/auth/verify-email?token=%s", s.config.BaseURL, verification.Token)

templateData := map[string]interface{}{
"VerificationURL": verificationURL,
"Year":            time.Now().Year(),
}

rendered, err := s.templateService.RenderTemplate("verification", language, templateData)
if err != nil {
return fmt.Errorf("failed to render email template: %w", err)
}

emailMsg := &email.Email{
To:       user.Email,
Subject:  rendered.Subject,
HTMLBody: rendered.HTMLBody,
TextBody: rendered.TextBody,
Language: language,
}

if err := s.provider.SendEmail(ctx, emailMsg); err != nil {
return fmt.Errorf("failed to send verification email: %w", err)
}

return nil
}

// VerifyEmail verifies an email address using the provided token.
func (s *EmailService) VerifyEmail(ctx context.Context, token string) (*models.User, error) {
guardCtx := scopes.SkipUserScopeGuard(ctx)

var verification models.EmailVerification
if err := s.db.WithContext(guardCtx).Where("token = ?", token).First(&verification).Error; err != nil {
if errors.Is(err, gorm.ErrRecordNotFound) {
return nil, ErrTokenNotFound
}
return nil, fmt.Errorf("failed to find verification token: %w", err)
}

if verification.IsVerified() {
return nil, ErrTokenAlreadyUsed
}

if verification.IsExpired() {
return nil, ErrTokenExpired
}

var user models.User
if err := s.db.WithContext(ctx).Where("id = ?", verification.UserID).First(&user).Error; err != nil {
return nil, fmt.Errorf("failed to find user: %w", err)
}

now := time.Now()
err := s.db.WithContext(guardCtx).Transaction(func(tx *gorm.DB) error {
verification.VerifiedAt = &now
if err := tx.Save(&verification).Error; err != nil {
return err
}

user.EmailVerified = true
user.EmailVerifiedAt = &now
return tx.Save(&user).Error
})

if err != nil {
return nil, fmt.Errorf("failed to mark email as verified: %w", err)
}

return &user, nil
}

// ResendVerification resends the verification email.
func (s *EmailService) ResendVerification(ctx context.Context, userEmail string, language string) error {
normalizedEmail := strings.ToLower(strings.TrimSpace(userEmail))

var user models.User
if err := s.db.WithContext(ctx).Where("email = ?", normalizedEmail).First(&user).Error; err != nil {
if errors.Is(err, gorm.ErrRecordNotFound) {
return nil
}
return fmt.Errorf("failed to find user: %w", err)
}

if user.EmailVerified {
return nil
}

s.limiterMu.RLock()
lastSent, exists := s.resendLimiter[user.ID]
s.limiterMu.RUnlock()

if exists && time.Since(lastSent) < time.Minute {
return ErrResendRateLimit
}

s.limiterMu.Lock()
s.resendLimiter[user.ID] = time.Now()
s.limiterMu.Unlock()

return s.SendVerificationEmail(ctx, &user, language)
}

// SendSecurityAlert sends a security alert email to the user.
func (s *EmailService) SendSecurityAlert(ctx context.Context, user *models.User, alertType string, alertTitle string, alertMessage string, details map[string]string, language string) error {
templateData := map[string]interface{}{
"AlertTitle":   alertTitle,
"AlertMessage": alertMessage,
"Details":      details,
"Year":         time.Now().Year(),
}

rendered, err := s.templateService.RenderTemplate("security_alert", language, templateData)
if err != nil {
return fmt.Errorf("failed to render security alert template: %w", err)
}

emailMsg := &email.Email{
To:       user.Email,
Subject:  rendered.Subject,
HTMLBody: rendered.HTMLBody,
TextBody: rendered.TextBody,
Language: language,
}

if err := s.provider.SendEmail(ctx, emailMsg); err != nil {
return fmt.Errorf("failed to send security alert email: %w", err)
}

return nil
}

// IsEmailVerificationEnabled returns whether email verification is enabled.
func (s *EmailService) IsEmailVerificationEnabled() bool {
return s.config.VerifyEmails
}

// CleanupExpiredTokens removes expired verification tokens from the database.
func (s *EmailService) CleanupExpiredTokens(ctx context.Context) error {
result := s.db.WithContext(ctx).
Where("expires_at < ? AND verified_at IS NULL", time.Now()).
Delete(&models.EmailVerification{})

if result.Error != nil {
return fmt.Errorf("failed to cleanup expired tokens: %w", result.Error)
}

return nil
}
