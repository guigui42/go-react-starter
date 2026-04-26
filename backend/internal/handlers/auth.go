// Package handlers provides HTTP request handlers for the API.
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/example/go-react-starter/internal/config"
	"github.com/example/go-react-starter/internal/middleware"
	"github.com/example/go-react-starter/internal/models"
	"github.com/example/go-react-starter/internal/services"
	"github.com/example/go-react-starter/pkg/response"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// setAuthCookie sets the JWT token as an httpOnly cookie and a CSRF token cookie.
// Uses SameSite=Lax to allow cookies on top-level navigations (required for OAuth redirect flows)
// while still protecting against most CSRF attacks.
// The CSRF cookie is set here so that the frontend has a valid token for subsequent
// state-changing requests without requiring a separate GET request.
func setAuthCookie(w http.ResponseWriter, token string, sessionDuration time.Duration, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.AuthTokenCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})

	// Set CSRF token cookie for the frontend to use in state-changing requests.
	// Error is ignored as SetCSRFCookie can only fail on crypto/rand failure,
	// which indicates a severe system issue; authentication should not be blocked.
	_ = middleware.SetCSRFCookie(w, secure)
}

// clearAuthCookie clears the authentication cookie
func clearAuthCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.AuthTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})
}

// dummyPasswordHash is a pre-computed bcrypt hash used for constant-time login operations
// to prevent timing attacks when a user doesn't exist. For production (bcrypt cost 12),
// we use a pre-computed hash. For tests, we generate one dynamically to match the test cost.
// Generated using: bcrypt.GenerateFromPassword([]byte("dummy"), 12)
const productionDummyHash = "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.Vu0rFV4xqQ0yGy"

// getDummyHash returns an appropriate dummy hash for timing attack prevention.
// In production, uses the pre-computed cost-12 hash. In tests with reduced cost,
// generates a hash matching the current bcrypt cost for consistent timing.
func getDummyHash() []byte {
	currentCost := models.GetCurrentBcryptCost()
	if currentCost == models.BcryptCost {
		// Production: use pre-computed hash for efficiency
		return []byte(productionDummyHash)
	}
	// Tests: generate hash with current cost (only happens once per test run)
	hash, _ := models.GenerateDummyHash()
	return hash
}

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	db             *gorm.DB
	emailService   EmailServiceInterface
	auditService   *services.AuditService
	authCfg        *config.AuthConfig
	serverCfg      *config.ServerConfig
	adminEmails    []string
	tokenBlocklist *middleware.TokenBlocklist
	lockoutService *services.AccountLockoutService
}

// EmailServiceInterface defines the methods required from the email service.
// This allows for easy testing with mock implementations.
type EmailServiceInterface interface {
	SendVerificationEmail(ctx context.Context, user *models.User, language string) error
	VerifyEmail(ctx context.Context, token string) (*models.User, error)
	ResendVerification(ctx context.Context, userEmail string, language string) error
	IsEmailVerificationEnabled() bool
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(db *gorm.DB, auditService *services.AuditService, authCfg *config.AuthConfig, serverCfg *config.ServerConfig, lockoutService *services.AccountLockoutService, blocklist ...*middleware.TokenBlocklist) *AuthHandler {
	var bl *middleware.TokenBlocklist
	if len(blocklist) > 0 {
		bl = blocklist[0]
	}
	return &AuthHandler{
		db:             db,
		auditService:   auditService,
		authCfg:        authCfg,
		serverCfg:      serverCfg,
		tokenBlocklist: bl,
		lockoutService: lockoutService,
	}
}

// NewAuthHandlerWithEmail creates a new authentication handler with email service.
func NewAuthHandlerWithEmail(db *gorm.DB, emailService EmailServiceInterface, auditService *services.AuditService, authCfg *config.AuthConfig, serverCfg *config.ServerConfig, adminEmails []string, lockoutService *services.AccountLockoutService, blocklist ...*middleware.TokenBlocklist) *AuthHandler {
	normalized := make([]string, 0, len(adminEmails))
	for _, e := range adminEmails {
		if trimmed := strings.TrimSpace(strings.ToLower(e)); trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	var bl *middleware.TokenBlocklist
	if len(blocklist) > 0 {
		bl = blocklist[0]
	}
	return &AuthHandler{
		db:             db,
		emailService:   emailService,
		auditService:   auditService,
		authCfg:        authCfg,
		serverCfg:      serverCfg,
		adminEmails:    normalized,
		tokenBlocklist: bl,
		lockoutService: lockoutService,
	}
}

// RegisterRequest represents the registration payload.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents the login payload.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CheckAuthMethodsRequest represents the email check payload.
type CheckAuthMethodsRequest struct {
	Email string `json:"email"`
}

// CheckAuthMethodsResponse indicates available authentication methods.
// To prevent user enumeration, this response structure is always returned
// regardless of whether the user exists.
type CheckAuthMethodsResponse struct {
	HasPassword bool `json:"has_password"`
	HasPasskey  bool `json:"has_passkey"`
}

// Register handles user registration requests.
// It creates a new user with hashed password and default preferences.
//
// @Summary      Register a new user
// @Description  Creates a new user account with email and optional password. If no password is provided, user must set up passkey authentication. If email verification is enabled, returns verification pending response.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        user body RegisterRequest true "User registration data"
// @Success      201 {object} models.AuthResponse "User created successfully with JWT token"
// @Success      202 {object} models.VerificationPendingResponse "Verification email sent"
// @Failure      400 {object} response.ErrorResponse "Validation failed"
// @Failure      409 {object} response.ErrorResponse "Email already exists"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Don't expose internal Go error details (struct names, field types)
		// to prevent information disclosure that aids reconnaissance.
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	// Validate required fields
	if req.Email == "" {
		response.BadRequest(w, "Validation failed", map[string]interface{}{
			"email": "email is required",
		})
		return
	}

	// Validate email format
	if err := models.ValidateEmail(req.Email); err != nil {
		response.ValidationError(w, map[string]interface{}{
			"email": err.Error(),
		})
		return
	}

	// Normalize email to lowercase for case-insensitive comparison
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))

	// Create user instance
	isAdmin := false
	for _, adminEmail := range h.adminEmails {
		if adminEmail == normalizedEmail {
			isAdmin = true
			break
		}
	}
	user := models.User{
		Email:   normalizedEmail,
		IsAdmin: isAdmin,
	}

	// Password is optional for passkey-only accounts
	// If password is provided, validate and hash it
	if req.Password != "" {
		// SetPassword validates minimum 8 characters and hashes the password
		if err := user.SetPassword(req.Password); err != nil {
			response.ValidationError(w, map[string]interface{}{
				"password": err.Error(),
			})
			return
		}
	}
	// If password is empty, user will need to set up passkey authentication

	// Check if email already exists (efficient existence check)
	var exists bool
	if err := h.db.Model(&models.User{}).Select("count(*) > 0").Where("email = ?", normalizedEmail).Find(&exists).Error; err == nil && exists {
		response.Error(w, http.StatusConflict, "CONFLICT", "Email already exists", nil)
		return
	}

	// Create user in database

	if err := h.db.Create(&user).Error; err != nil {
		response.InternalServerError(w, "Failed to create user")
		return
	}

	// Create default preferences
	prefs := models.UserPreferences{
		UserID:       user.ID,
		Language:     "en",
	}

	if err := h.db.Create(&prefs).Error; err != nil {
		response.InternalServerError(w, "Failed to create preferences")
		return
	}

	// Check if email verification is required
	if h.emailService != nil && h.emailService.IsEmailVerificationEnabled() {
		// Send verification email
		if err := h.emailService.SendVerificationEmail(r.Context(), &user, prefs.Language); err != nil {
			// Log error but don't fail registration - user can resend
			// Note: Proper logging would require injecting a logger into AuthHandler.
			// For now, the error is silently ignored since registration should still succeed.
			// The user can request a new verification email using the resend endpoint.
		}

		// Log registration
		if h.auditService != nil {
			_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
				EventType: "auth.register",
				ActorID:   &user.ID,
				Action:    "New user registered",
				Status:    "success",
				IPAddress: services.GetClientIP(r),
				UserAgent: services.GetUserAgent(r),
				Metadata: map[string]interface{}{
					"email":              normalizedEmail,
					"email_verification": true,
				},
			})
		}

		// Return verification pending response (no auth cookie)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(models.VerificationPendingResponse{
			Message: "verification_email_sent",
			Email:   user.Email,
		})
		return
	}

	// Email verification not required - auto-login
	// Generate JWT token to automatically authenticate the user
	token, err := generateJWT(user.ID.String(), h.authCfg.JWTSecret, h.authCfg.SessionDuration)
	if err != nil {
		response.InternalServerError(w, "Failed to generate token")
		return
	}

	// Set JWT token in httpOnly cookie
	setAuthCookie(w, token, h.authCfg.SessionDuration, h.serverCfg.IsSecure())

	// Log registration
	if h.auditService != nil {
		_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
			EventType: "auth.register",
			ActorID:   &user.ID,
			Action:    "New user registered",
			Status:    "success",
			IPAddress: services.GetClientIP(r),
			UserAgent: services.GetUserAgent(r),
			Metadata: map[string]interface{}{
				"email":              normalizedEmail,
				"email_verification": false,
			},
		})
	}

	// Return user data (token is in httpOnly cookie only, not in response body for security)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.AuthResponse{
		User: user,
	})
}

// Login handles user login requests.
// It authenticates the user and returns a JWT token.
//
// Security: This function implements constant-time login to prevent account enumeration
// via timing attacks. A bcrypt comparison is always performed regardless of whether
// the user exists, ensuring both code paths take similar time (~400ms).
//
// @Summary      Login user
// @Description  Authenticates user with email and password, returns JWT token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        credentials body LoginRequest true "User login credentials"
// @Success      200 {object} models.AuthResponse "Login successful with JWT token"
// @Failure      400 {object} response.ErrorResponse "Invalid request body"
// @Failure      401 {object} response.ErrorResponse "Invalid credentials"
// @Failure      403 {object} models.EmailNotVerifiedResponse "Email not verified"
// @Failure      429 {object} response.ErrorResponse "Account locked due to too many failed attempts"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Don't expose internal Go error details to prevent information disclosure.
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		response.BadRequest(w, "Email and password are required", nil)
		return
	}

	// Normalize email to lowercase for case-insensitive comparison
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))

	// Check if account is locked due to too many failed attempts
	if h.lockoutService != nil {
		locked, remaining, err := h.lockoutService.IsLocked(r.Context(), normalizedEmail)
		if err != nil {
			if logger := middleware.GetLogger(r.Context()); logger != nil {
				logger.WithMetadata(zerolog.ErrorLevel, "failed to check account lockout", map[string]interface{}{
					"error": err.Error(),
					"email": normalizedEmail,
				})
			}
			response.InternalServerError(w, "Unable to process login request")
			return
		}
		if locked {
			if h.auditService != nil {
				_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
					EventType: "auth.login.locked",
					Action:    "Login attempt on locked account",
					Status:    "failure",
					IPAddress: services.GetClientIP(r),
					UserAgent: services.GetUserAgent(r),
					Metadata: map[string]interface{}{
						"email":            normalizedEmail,
						"locked_remaining": remaining.String(),
					},
				})
			}
			response.Error(w, http.StatusTooManyRequests, "ACCOUNT_LOCKED",
				"Too many failed login attempts. Please try again later.", nil)
			return
		}
	}

	// Find user by email
	var user models.User
	userFound := h.db.Where("email = ?", normalizedEmail).First(&user).Error == nil

	// Always perform bcrypt comparison to maintain constant timing and prevent
	// account enumeration attacks. When user is not found, we compare against
	// a dummy hash to ensure the operation takes similar time (~400ms in production).
	if !userFound {
		// Perform comparison to maintain consistent timing
		models.CompareHashAndPassword(getDummyHash(), []byte(req.Password))

		// Log failed login attempt
		if h.auditService != nil {
			_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
				EventType: "auth.login.failed",
				Action:    "Failed login attempt - user not found",
				Status:    "failure",
				IPAddress: services.GetClientIP(r),
				UserAgent: services.GetUserAgent(r),
				Metadata: map[string]interface{}{
					"email": normalizedEmail,
				},
			})
		}

		if h.lockoutService != nil {
			// Use context.WithoutCancel to preserve request-scoped values (trace/span/logger)
			// while ensuring recording completes even if the client disconnects
			lockoutCtx := context.WithoutCancel(r.Context())
			if err := h.lockoutService.RecordFailure(lockoutCtx, normalizedEmail); err != nil {
				if logger := middleware.GetLogger(r.Context()); logger != nil {
					logger.WithMetadata(zerolog.ErrorLevel, "failed to record login failure", map[string]interface{}{
						"error": err.Error(),
						"email": normalizedEmail,
					})
				}
			}
		}
		response.Unauthorized(w, "Invalid credentials")
		return
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		// Log failed login attempt with wrong password
		if h.auditService != nil {
			_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
				EventType: "auth.login.failed",
				ActorID:   &user.ID,
				Action:    "Failed login attempt - invalid password",
				Status:    "failure",
				IPAddress: services.GetClientIP(r),
				UserAgent: services.GetUserAgent(r),
				Metadata: map[string]interface{}{
					"email": normalizedEmail,
				},
			})
		}

		if h.lockoutService != nil {
			// Use context.WithoutCancel to preserve request-scoped values (trace/span/logger)
			// while ensuring recording completes even if the client disconnects
			lockoutCtx := context.WithoutCancel(r.Context())
			if err := h.lockoutService.RecordFailure(lockoutCtx, normalizedEmail); err != nil {
				if logger := middleware.GetLogger(r.Context()); logger != nil {
					logger.WithMetadata(zerolog.ErrorLevel, "failed to record login failure", map[string]interface{}{
						"error": err.Error(),
						"email": normalizedEmail,
					})
				}
			}
		}
		response.Unauthorized(w, "Invalid credentials")
		return
	}

	// Reset failed login attempts on successful authentication
	if h.lockoutService != nil {
		// Use context.WithoutCancel to preserve request-scoped values (trace/span/logger)
		// while ensuring cleanup completes even if the client disconnects
		lockoutCtx := context.WithoutCancel(r.Context())
		if err := h.lockoutService.RecordSuccess(lockoutCtx, normalizedEmail); err != nil {
			if logger := middleware.GetLogger(r.Context()); logger != nil {
				logger.WithMetadata(zerolog.ErrorLevel, "failed to clear login attempts", map[string]interface{}{
					"error": err.Error(),
					"email": normalizedEmail,
				})
			}
		}
	}

	// Check if email verification is required and email is not verified
	if h.emailService != nil && h.emailService.IsEmailVerificationEnabled() && !user.EmailVerified {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(models.EmailNotVerifiedResponse{
			Code:    "EMAIL_NOT_VERIFIED",
			Message: "Please verify your email address",
			Email:   user.Email,
		})
		return
	}

	// Generate JWT token
	token, err := generateJWT(user.ID.String(), h.authCfg.JWTSecret, h.authCfg.SessionDuration)
	if err != nil {
		response.InternalServerError(w, "Failed to generate token")
		return
	}

	// Set JWT token in httpOnly cookie
	setAuthCookie(w, token, h.authCfg.SessionDuration, h.serverCfg.IsSecure())

	// Log successful login
	if h.auditService != nil {
		_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
			EventType: "auth.login.success",
			ActorID:   &user.ID,
			Action:    "User logged in successfully",
			Status:    "success",
			IPAddress: services.GetClientIP(r),
			UserAgent: services.GetUserAgent(r),
		})
	}

	// Return user data (token is in httpOnly cookie only, not in response body for security)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.AuthResponse{
		User: user,
	})
}

// CheckAuthMethods checks available authentication methods for an email address.
// This endpoint is used by the email-first login flow to determine which
// authentication options to present to the user.
//
// Security: This function implements constant-time responses to prevent user
// enumeration attacks. The same response structure and timing is maintained
// regardless of whether the email exists in the database.
//
// @Summary      Check authentication methods
// @Description  Returns available auth methods (password, passkey) for a given email. Returns same structure for non-existent users to prevent enumeration.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        email body CheckAuthMethodsRequest true "Email to check"
// @Success      200 {object} CheckAuthMethodsResponse "Authentication methods available"
// @Failure      400 {object} response.ErrorResponse "Invalid request body"
// @Router       /api/v1/auth/check-methods [post]
func (h *AuthHandler) CheckAuthMethods(w http.ResponseWriter, r *http.Request) {
	var req CheckAuthMethodsRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Don't expose internal Go error details to prevent information disclosure.
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	// Validate email is provided
	if req.Email == "" {
		response.BadRequest(w, "Email is required", nil)
		return
	}

	// Normalize email to lowercase for case-insensitive comparison
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))

	// Default response for non-existent users - prevents enumeration
	// We default to has_password=true to always show password option as fallback
	resp := CheckAuthMethodsResponse{
		HasPassword: true,
		HasPasskey:  false,
	}

	// Find user by email
	var user models.User
	userFound := h.db.Where("email = ?", normalizedEmail).First(&user).Error == nil

	if userFound {
		// Check if user has a password set (password_hash is not empty)
		resp.HasPassword = user.PasswordHash != nil && *user.PasswordHash != ""

		// Check if user has any passkeys registered
		var credentialCount int64
		h.db.Model(&models.UserCredential{}).Where("user_id = ?", user.ID).Count(&credentialCount)
		resp.HasPasskey = credentialCount > 0
	}

	// Always perform a dummy database query for constant timing when user not found
	// This ensures timing is consistent whether or not the user exists
	if !userFound {
		var dummyCount int64
		h.db.Model(&models.UserCredential{}).Where("user_id = ?", "00000000-0000-0000-0000-000000000000").Count(&dummyCount)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Logout handles user logout requests.
// Clears the authentication cookie and validates the request has valid auth.
//
// @Summary      Logout user
// @Description  Logs out the current user by clearing the authentication cookie.
// @Tags         Authentication
// @Security     BearerAuth
// @Success      204 "Successfully logged out"
// @Failure      401 {object} response.ErrorResponse "Authorization required"
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Check for authentication via cookie or Authorization header
	hasAuth := false

	// Check cookie first
	if cookie, err := r.Cookie(middleware.AuthTokenCookieName); err == nil && cookie.Value != "" {
		hasAuth = true
	}

	// Also check Authorization header for backward compatibility
	if !hasAuth {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				hasAuth = true
			}
		}
	}

	if !hasAuth {
		response.Unauthorized(w, "Authorization required")
		return
	}

	// Clear the authentication cookie
	clearAuthCookie(w, h.serverCfg.IsSecure())

	// Revoke the JWT token by adding its JTI to the blocklist
	if h.tokenBlocklist != nil {
		tokenString := middleware.ExtractTokenFromRequest(r)
		if tokenString != "" {
			if token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(h.authCfg.JWTSecret), nil
			}); err == nil {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if jti, ok := claims["jti"].(string); ok {
						// Use token expiration or default to session duration
						exp := time.Now().Add(h.authCfg.SessionDuration)
						if expFloat, ok := claims["exp"].(float64); ok {
							exp = time.Unix(int64(expFloat), 0)
						}
						h.tokenBlocklist.Block(jti, exp)
					}
				}
			}
		}
	}

	// Log logout event
	// Try to get user ID from context (may not be available if using header auth)
	if userID, ok := middleware.GetUserID(r); ok {
		if parsedUserID, err := uuid.Parse(userID); err == nil && h.auditService != nil {
			_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
				EventType: "auth.logout",
				ActorID:   &parsedUserID,
				Action:    "User logged out",
				Status:    "success",
				IPAddress: services.GetClientIP(r),
				UserAgent: services.GetUserAgent(r),
			})
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetMe returns the currently authenticated user's information.
// Requires a valid JWT token in the Authorization header.
//
// @Summary      Get current user
// @Description  Returns the currently authenticated user's information
// @Tags         Authentication
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.UserInfoResponse "User information"
// @Failure      401 {object} response.ErrorResponse "User not authenticated"
// @Failure      404 {object} response.ErrorResponse "User not found"
// @Router       /api/v1/auth/me [get]
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}
	if userID == "" {
		response.Unauthorized(w, "User ID is empty")
		return
	}

	// Fetch user from database
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		response.NotFound(w, "User not found")
		return
	}

	// Return user data (excluding password hash)
	response.Success(w, http.StatusOK, models.UserInfoResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		IsAdmin:       user.IsAdmin,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt.Format(time.RFC3339),
	})
}

// DeleteUserRequest represents the request payload for deleting a user account.
type DeleteUserRequest struct {
	ConfirmEmail string `json:"confirmEmail"`
}

// DeleteUser handles user account deletion requests.
// Requires authentication and email confirmation for security.
// Permanently deletes all user data including trades, broker accounts, preferences, etc.
// This implements GDPR Article 17 (Right to Erasure).
//
// @Summary      Delete user account
// @Description  Permanently deletes the authenticated user's account and all associated data. This action cannot be undone.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        confirmation body DeleteUserRequest true "Email confirmation"
// @Security     BearerAuth
// @Success      204 "Account deleted successfully"
// @Failure      400 {object} response.ErrorResponse "Invalid request or email mismatch"
// @Failure      401 {object} response.ErrorResponse "Not authenticated"
// @Failure      500 {object} response.ErrorResponse "Failed to delete account"
// @Router       /api/v1/user [delete]
func (h *AuthHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Unauthorized(w, "User ID not found in context")
		return
	}
	if userID == "" {
		response.Unauthorized(w, "User ID is empty")
		return
	}

	// Decode request body
	var req DeleteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Don't expose internal Go error details to prevent information disclosure.
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	// Fetch user from database to verify email
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		response.NotFound(w, "User not found")
		return
	}

	// Verify email confirmation
	if req.ConfirmEmail != user.Email {
		response.BadRequest(w, "Email confirmation does not match", map[string]interface{}{
			"confirmEmail": "Email does not match your account email",
		})
		return
	}

	// Delete user and all associated data.
	// All child tables have ON DELETE CASCADE foreign keys to users,
	// so deleting the user record automatically cascades to:
	// trades, dividends, broker_accounts, user_preferences, tax_reports,
	// user_credentials, user_backup_codes, user_auth_migrations,
	// email_verifications, and user_oauth_accounts.
	err := h.db.Delete(&user).Error

	if err != nil {
		// Log failed deletion attempt
		if parsedUserID, parseErr := uuid.Parse(userID); parseErr == nil && h.auditService != nil {
			_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
				EventType: "user.delete.failed",
				ActorID:   &parsedUserID,
				TargetID:  &parsedUserID,
				Action:    "Failed to delete user account",
				Status:    "failure",
				IPAddress: services.GetClientIP(r),
				UserAgent: services.GetUserAgent(r),
				Metadata: map[string]interface{}{
					"error": err.Error(),
					"email": user.Email,
				},
			})
		}

		response.InternalServerError(w, "Failed to delete account")
		return
	}

	// Log successful account deletion
	if parsedUserID, parseErr := uuid.Parse(userID); parseErr == nil && h.auditService != nil {
		_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
			EventType: "user.delete.success",
			ActorID:   &parsedUserID,
			TargetID:  &parsedUserID,
			Action:    "User account deleted",
			Status:    "success",
			IPAddress: services.GetClientIP(r),
			UserAgent: services.GetUserAgent(r),
			Metadata: map[string]interface{}{
				"email": user.Email,
			},
		})
	}

	// Clear authentication cookie
	clearAuthCookie(w, h.serverCfg.IsSecure())

	// Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// generateJWT creates a JWT token for a user.
// It uses the provided JWT secret and session duration from the centralized config.
func generateJWT(userID string, jwtSecret string, sessionDuration time.Duration) (string, error) {
	if jwtSecret == "" {
		return "", fmt.Errorf("JWT_SECRET not configured")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     now.Add(sessionDuration).Unix(),
		"iat":     now.Unix(),
		"nbf":     now.Add(-30 * time.Second).Unix(), // RFC 7519: 30s leeway for clock skew across nodes
		"jti":     models.NewID().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// ResendVerificationRequest represents the resend verification request.
type ResendVerificationRequest struct {
	Email string `json:"email"`
}

// VerifyEmail handles email verification requests.
// It validates the token and marks the user's email as verified.
//
// @Summary      Verify email address
// @Description  Verifies a user's email address using the verification token
// @Tags         Authentication
// @Produce      json
// @Param        token query string true "Verification token"
// @Success      200 {object} models.EmailVerifiedResponse "Email verified successfully"
// @Failure      400 {object} response.ErrorResponse "Invalid or expired token"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /auth/verify-email [get]
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		response.BadRequest(w, "Verification token is required", nil)
		return
	}

	if h.emailService == nil {
		response.InternalServerError(w, "Email service not configured")
		return
	}

	user, err := h.emailService.VerifyEmail(r.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrTokenNotFound):
			response.BadRequest(w, "Invalid verification token", nil)
		case errors.Is(err, services.ErrTokenExpired):
			response.BadRequest(w, "Verification token has expired. Please request a new one.", nil)
		case errors.Is(err, services.ErrTokenAlreadyUsed):
			response.BadRequest(w, "This email has already been verified", nil)
		default:
			response.InternalServerError(w, "Failed to verify email")
		}
		return
	}

	// Log email verification
	if h.auditService != nil {
		_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
			EventType: "auth.email.verified",
			ActorID:   &user.ID,
			Action:    "Email address verified",
			Status:    "success",
			IPAddress: services.GetClientIP(r),
			UserAgent: services.GetUserAgent(r),
			Metadata: map[string]interface{}{
				"email": user.Email,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.EmailVerifiedResponse{
		Message:  "Email verified successfully",
		Verified: true,
	})
}

// ResendVerification handles resend verification email requests.
// It is rate-limited to 1 request per minute per user.
//
// @Summary      Resend verification email
// @Description  Resends the verification email to the user. Rate limited to 1 request per minute.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        email body ResendVerificationRequest true "Email address"
// @Success      200 {object} models.MessageResponse "Verification email sent (if email exists)"
// @Failure      400 {object} response.ErrorResponse "Invalid request"
// @Failure      429 {object} response.ErrorResponse "Rate limit exceeded"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /auth/resend-verification [post]
func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req ResendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Don't expose internal Go error details to prevent information disclosure.
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	if req.Email == "" {
		response.BadRequest(w, "Email is required", nil)
		return
	}

	if h.emailService == nil {
		response.InternalServerError(w, "Email service not configured")
		return
	}

	// Use default language (en) - in production, you might want to look up user preferences
	err := h.emailService.ResendVerification(r.Context(), req.Email, "en")
	if err != nil {
		if errors.Is(err, services.ErrResendRateLimit) {
			response.Error(w, http.StatusTooManyRequests, "RATE_LIMIT", "Please wait before requesting another verification email", nil)
			return
		}
		// Log the error for debugging, but still return success to prevent email enumeration
		if logger := middleware.GetLogger(r.Context()); logger != nil {
			logger.WithMetadata(zerolog.WarnLevel, "resend verification failed", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// Always return success to prevent email enumeration attacks
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.MessageResponse{
		Message: "If an account exists with this email, a verification email has been sent",
	})
}
