// Package handlers provides HTTP request handlers for the API.
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/guigui42/go-react-starter/internal/config"
	"github.com/guigui42/go-react-starter/internal/middleware"
	"github.com/guigui42/go-react-starter/internal/models"
	"github.com/guigui42/go-react-starter/internal/repository"
	"github.com/guigui42/go-react-starter/internal/services"
	"github.com/guigui42/go-react-starter/pkg/response"
	"gorm.io/gorm"
)

// PasskeyHandler handles WebAuthn passkey authentication requests
type PasskeyHandler struct {
	db                *gorm.DB
	webAuthn          *services.WebAuthnService
	sessionStore      *services.SessionStore
	credentialRepo    *repository.UserCredentialRepository
	backupCodeService *services.BackupCodeService
	auditService      *services.AuditService
	emailService      EmailServiceInterface
	authCfg           *config.AuthConfig
	serverCfg         *config.ServerConfig
}

// NewPasskeyHandler creates a new passkey handler
func NewPasskeyHandler(
	db *gorm.DB,
	webAuthn *services.WebAuthnService,
	sessionStore *services.SessionStore,
	credentialRepo *repository.UserCredentialRepository,
	backupCodeService *services.BackupCodeService,
	auditService *services.AuditService,
	authCfg *config.AuthConfig,
	serverCfg *config.ServerConfig,
) *PasskeyHandler {
	return &PasskeyHandler{
		db:                db,
		webAuthn:          webAuthn,
		sessionStore:      sessionStore,
		credentialRepo:    credentialRepo,
		backupCodeService: backupCodeService,
		auditService:      auditService,
		authCfg:           authCfg,
		serverCfg:         serverCfg,
	}
}

// NewPasskeyHandlerWithEmail creates a new passkey handler with email service
func NewPasskeyHandlerWithEmail(
	db *gorm.DB,
	webAuthn *services.WebAuthnService,
	sessionStore *services.SessionStore,
	credentialRepo *repository.UserCredentialRepository,
	backupCodeService *services.BackupCodeService,
	auditService *services.AuditService,
	emailService EmailServiceInterface,
	authCfg *config.AuthConfig,
	serverCfg *config.ServerConfig,
) *PasskeyHandler {
	return &PasskeyHandler{
		db:                db,
		webAuthn:          webAuthn,
		sessionStore:      sessionStore,
		credentialRepo:    credentialRepo,
		backupCodeService: backupCodeService,
		auditService:      auditService,
		emailService:      emailService,
		authCfg:           authCfg,
		serverCfg:         serverCfg,
	}
}

// BeginRegistration starts the WebAuthn passkey registration flow
//
// @Summary      Begin passkey registration
// @Description  Starts the WebAuthn passkey registration flow for the authenticated user
// @Tags         Passkey
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.PasskeyRegistrationBeginResponse "WebAuthn registration options and session ID"
// @Failure      401 {object} response.ErrorResponse "Authentication required"
// @Failure      500 {object} response.ErrorResponse "Failed to begin registration"
// @Router       /api/v1/auth/passkey/registration/begin [post]
func (h *PasskeyHandler) BeginRegistration(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Begin registration with WebAuthn service
	options, sessionData, err := h.webAuthn.BeginRegistration(user)
	if err != nil {
		response.InternalServerError(w, "Failed to begin registration")
		return
	}

	// Store session data for verification
	sessionID := models.NewID().String()
	h.sessionStore.Set(sessionID, *sessionData)

	// Return options to client with session ID
	// Note: options.Response contains the publicKey credential creation options
	response.Success(w, http.StatusOK, models.PasskeyRegistrationBeginResponse{
		Options:   options.Response, // Return the publicKey options directly
		SessionID: sessionID,
	})
}

// FinishRegistrationRequest represents the credential creation response from client
type FinishRegistrationRequest struct {
	SessionID    string                                 `json:"sessionId"`
	FriendlyName string                                 `json:"friendlyName"`
	Credential   *protocol.ParsedCredentialCreationData `json:"-"` // Parsed from body
}

// FinishRegistration completes the WebAuthn passkey registration flow
//
// @Summary      Finish passkey registration
// @Description  Completes the WebAuthn passkey registration by verifying the credential
// @Tags         Passkey
// @Accept       json
// @Produce      json
// @Param        sessionId query string true "Session ID from begin registration"
// @Param        friendlyName query string false "Friendly name for the passkey" default(Unnamed Device)
// @Security     BearerAuth
// @Success      201 {object} models.PasskeyRegistrationFinishResponse "Registration verified with credential details"
// @Failure      400 {object} response.ErrorResponse "Invalid credential or session"
// @Failure      401 {object} response.ErrorResponse "Authentication required"
// @Failure      500 {object} response.ErrorResponse "Failed to store credential"
// @Router       /api/v1/auth/passkey/registration/finish [post]
func (h *PasskeyHandler) FinishRegistration(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Parse the credential creation response
	parsedResponse, err := protocol.ParseCredentialCreationResponse(r)
	if err != nil {
		response.BadRequest(w, "Invalid credential response", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get session ID and friendly name from query parameters
	sessionID := r.URL.Query().Get("sessionId")
	friendlyName := r.URL.Query().Get("friendlyName")

	if sessionID == "" {
		response.BadRequest(w, "Session ID is required", nil)
		return
	}

	// If no friendly name provided, use a default
	if friendlyName == "" {
		friendlyName = "Unnamed Device"
	}

	// Retrieve session data
	sessionData, exists := h.sessionStore.Get(sessionID)
	if !exists {
		response.BadRequest(w, "Invalid or expired session", nil)
		return
	}

	// Verify and create credential
	credential, err := h.webAuthn.FinishRegistration(user, sessionData, parsedResponse)
	if err != nil {
		response.BadRequest(w, "Failed to verify credential", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Store credential in database
	userCredential := &models.UserCredential{
		UserID:         user.ID,
		CredentialID:   credential.ID,
		PublicKey:      credential.PublicKey,
		SignCount:      credential.Authenticator.SignCount,
		AAGUID:         credential.Authenticator.AAGUID,
		BackupEligible: credential.Flags.BackupEligible,
		BackupState:    credential.Flags.BackupState,
		UserVerified:   credential.Flags.UserVerified,
		Flags:          uint8(credential.Flags.ProtocolValue()),
		FriendlyName:   friendlyName,
		CreatedAt:      time.Now().UTC(),
	}

	// Store transports if available
	if len(credential.Transport) > 0 {
		transportsJSON, _ := json.Marshal(credential.Transport)
		userCredential.Transports = string(transportsJSON)
	}

	if err := h.credentialRepo.Create(r.Context(), userCredential); err != nil {
		response.InternalServerError(w, "Failed to store credential")
		return
	}

	// Update user's auth migration status
	var migration models.UserAuthMigration
	err = h.db.WithContext(r.Context()).Where("user_id = ?", user.ID).First(&migration).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create migration record
		migration = models.UserAuthMigration{
			UserID:               user.ID,
			HasPassword:          user.PasswordHash != nil,
			HasPasskey:           true,
			PasswordLoginEnabled: true,
			PasskeyLoginEnabled:  true,
		}
		if err := h.db.WithContext(r.Context()).Create(&migration).Error; err != nil {
			response.InternalServerError(w, "Failed to create auth migration record")
			return
		}
	} else if err != nil {
		response.InternalServerError(w, "Failed to query auth migration status")
		return
	} else {
		// Update existing
		migration.HasPasskey = true
		migration.PasskeyLoginEnabled = true
		if err := h.db.WithContext(r.Context()).Save(&migration).Error; err != nil {
			response.InternalServerError(w, "Failed to update auth migration status")
			return
		}
	}

	// Delete session data
	h.sessionStore.Delete(sessionID)

	// Log passkey registration
	if h.auditService != nil {
		_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
			EventType: "auth.passkey.register",
			ActorID:   &user.ID,
			Action:    "User registered a new passkey",
			Status:    "success",
			IPAddress: services.GetClientIP(r),
			UserAgent: services.GetUserAgent(r),
			Metadata: map[string]interface{}{
				"credential_id": userCredential.ID.String(),
				"friendly_name": friendlyName,
			},
		})
	}

	response.Success(w, http.StatusCreated, models.PasskeyRegistrationFinishResponse{
		Verified:     true,
		CredentialID: userCredential.ID.String(),
		FriendlyName: userCredential.FriendlyName,
		CreatedAt:    userCredential.CreatedAt.Format(time.RFC3339),
	})
}

// BeginAuthenticationRequest represents the authentication begin request
type BeginAuthenticationRequest struct {
	Email string `json:"email"`
}

// BeginAuthentication starts the WebAuthn passkey authentication flow
//
// @Summary      Begin passkey authentication
// @Description  Starts the WebAuthn passkey authentication flow for a user
// @Tags         Passkey
// @Accept       json
// @Produce      json
// @Param        auth body BeginAuthenticationRequest true "User email for authentication"
// @Success      200 {object} models.PasskeyAuthenticationBeginResponse "WebAuthn authentication options and session ID"
// @Failure      400 {object} response.ErrorResponse "Email required or no passkeys registered"
// @Failure      404 {object} response.ErrorResponse "User not found"
// @Failure      500 {object} response.ErrorResponse "Failed to begin authentication"
// @Router       /api/v1/auth/passkey/authentication/begin [post]
func (h *PasskeyHandler) BeginAuthentication(w http.ResponseWriter, r *http.Request) {
	var req BeginAuthenticationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	if req.Email == "" {
		response.BadRequest(w, "Email is required", nil)
		return
	}

	// Normalize email to lowercase for case-insensitive comparison
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))

	// Find user by email
	var user models.User
	// ALWAYS preload credentials for constant timing regardless of whether user exists
	err := h.db.WithContext(r.Context()).Preload("Credentials").Where("email = ?", normalizedEmail).First(&user).Error

	// Check for database errors (but not "not found" which we handle below)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		response.InternalServerError(w, "Database error")
		return
	}

	userFound := err == nil
	hasCredentials := userFound && len(user.Credentials) > 0

	// ALWAYS call BeginAuthentication to maintain constant timing and prevent timing-based user enumeration
	var options *protocol.CredentialAssertion
	var sessionData *webauthn.SessionData
	var authErr error

	if hasCredentials {
		options, sessionData, authErr = h.webAuthn.BeginAuthentication(&user)
	} else {
		// Create dummy user with realistic credential structure to match timing of successful path
		// This prevents timing attacks that could distinguish between non-existent users and users without passkeys
		// Use pseudo-random data based on email to be consistent per email but unpredictable
		dummyCredID := make([]byte, 32)
		dummyPubKey := make([]byte, 65)
		dummyAAGUID := make([]byte, 16)
		// Fill with pseudo-random data based on email to be consistent per email but unpredictable
		for i := range dummyCredID {
			dummyCredID[i] = byte((int(req.Email[i%len(req.Email)]) * (i + 1)) % 256)
		}
		for i := range dummyPubKey {
			dummyPubKey[i] = byte((int(req.Email[i%len(req.Email)]) * (i + 7)) % 256)
		}

		dummyUser := &models.User{
			ID:    models.NewID(),
			Email: req.Email,
			Credentials: []models.UserCredential{
				{
					CredentialID:   dummyCredID,
					PublicKey:      dummyPubKey,
					AAGUID:         dummyAAGUID, // Can stay zero as it's rarely checked
					SignCount:      0,
					BackupEligible: false,
					BackupState:    false,
					UserVerified:   true,
					Flags:          0x45, // User Present (UP) + User Verified (UV) flags
				},
			},
		}
		// Call and assign to explicit dummy variables for clarity (result discarded for timing only)
		dummyOptions, dummySessionData, err := h.webAuthn.BeginAuthentication(dummyUser)
		_, _ = dummyOptions, dummySessionData // Explicitly discard
		authErr = err
	}

	// Handle authentication errors from both paths to maintain timing consistency
	if authErr != nil {
		response.InternalServerError(w, "Authentication service error")
		return
	}

	// Generic error for both "user not found" and "no credentials" to prevent user enumeration
	// Both paths took similar time before reaching this point
	if !hasCredentials {
		response.BadRequest(w, "Unable to start passkey authentication. Please use password login or register a passkey.", nil)
		return
	}

	// Store session data
	sessionID := models.NewID().String()
	h.sessionStore.Set(sessionID, *sessionData)

	response.Success(w, http.StatusOK, models.PasskeyAuthenticationBeginResponse{
		Options:   options.Response, // Return the publicKey options directly
		SessionID: sessionID,
	})
}

// FinishAuthentication completes the WebAuthn passkey authentication flow
//
// @Summary      Finish passkey authentication
// @Description  Completes the WebAuthn passkey authentication and returns JWT token
// @Tags         Passkey
// @Accept       json
// @Produce      json
// @Param        sessionId query string true "Session ID from begin authentication"
// @Success      200 {object} models.PasskeyAuthenticationFinishResponse "Authentication verified with JWT token"
// @Failure      400 {object} response.ErrorResponse "Invalid credential or session"
// @Failure      500 {object} response.ErrorResponse "Failed to verify credential or generate token"
// @Router       /api/v1/auth/passkey/authentication/finish [post]
func (h *PasskeyHandler) FinishAuthentication(w http.ResponseWriter, r *http.Request) {
	// Parse the credential assertion response
	parsedResponse, err := protocol.ParseCredentialRequestResponse(r)
	if err != nil {
		response.BadRequest(w, "Invalid credential response", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get session ID from query parameter
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		response.BadRequest(w, "Session ID is required", nil)
		return
	}

	// Retrieve session data
	sessionData, exists := h.sessionStore.Get(sessionID)
	if !exists {
		response.BadRequest(w, "Invalid or expired session", nil)
		return
	}

	// Get user from session data (need to implement this)
	// For now, we'll look up the credential and get the user from there
	credential, err := h.credentialRepo.FindByCredentialID(r.Context(), parsedResponse.RawID)
	if err != nil {
		response.BadRequest(w, "Credential not found", nil)
		return
	}

	// Load user
	var user models.User
	if err := h.db.WithContext(r.Context()).Preload("Credentials").Where("id = ?", credential.UserID).First(&user).Error; err != nil {
		response.InternalServerError(w, "Failed to load user")
		return
	}

	// Verify the assertion
	_, err = h.webAuthn.FinishAuthentication(&user, sessionData, parsedResponse)
	if err != nil {
		// Log failed passkey login
		if h.auditService != nil {
			_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
				EventType: "auth.passkey.login",
				ActorID:   &user.ID,
				Action:    "Failed passkey authentication",
				Status:    "failure",
				IPAddress: services.GetClientIP(r),
				UserAgent: services.GetUserAgent(r),
			})
		}
		response.BadRequest(w, "Failed to verify credential", map[string]interface{}{
			"error": err.Error(),
		})
		return
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

	// Update sign count and last used timestamp
	err = h.credentialRepo.UpdateSignCount(r.Context(), credential.ID, user.ID, parsedResponse.Response.AuthenticatorData.Counter)
	if err != nil {
		// Log but don't fail - authentication was successful
		// TODO: Add proper logging
	}

	// Update last password login timestamp
	var migration models.UserAuthMigration
	if err := h.db.WithContext(r.Context()).Where("user_id = ?", user.ID).First(&migration).Error; err == nil {
		if err := h.db.WithContext(r.Context()).Model(&migration).Update("last_password_login", nil).Error; err != nil {
			response.InternalServerError(w, "Failed to update auth migration status")
			return
		}
		if err := h.db.WithContext(r.Context()).Save(&migration).Error; err != nil {
			response.InternalServerError(w, "Failed to save auth migration status")
			return
		}
	}

	// Generate JWT token
	token, err := generateJWT(user.ID.String(), h.authCfg.JWTSecret, h.authCfg.SessionDuration)
	if err != nil {
		response.InternalServerError(w, "Failed to generate token")
		return
	}

	// Set JWT token in httpOnly cookie
	setAuthCookie(w, token, h.authCfg.SessionDuration, h.serverCfg.IsSecure())

	// Delete session data
	h.sessionStore.Delete(sessionID)

	// Log successful passkey login
	if h.auditService != nil {
		_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
			EventType: "auth.passkey.login",
			ActorID:   &user.ID,
			Action:    "User logged in via passkey",
			Status:    "success",
			IPAddress: services.GetClientIP(r),
			UserAgent: services.GetUserAgent(r),
		})
	}

	response.Success(w, http.StatusOK, models.PasskeyAuthenticationFinishResponse{
		Verified: true,
		User: models.UserInfoResponse{
			ID:            user.ID.String(),
			Email:         user.Email,
			IsAdmin:       user.IsAdmin,
			EmailVerified: user.EmailVerified,
		},
		Token: token,
	})
}

// ListCredentials returns all passkeys for the authenticated user
//
// @Summary      List passkeys
// @Description  Returns all registered passkeys for the authenticated user
// @Tags         Passkey
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.PasskeyCredentialsListResponse "List of passkey credentials"
// @Failure      401 {object} response.ErrorResponse "Authentication required"
// @Failure      500 {object} response.ErrorResponse "Failed to load credentials"
// @Router       /api/v1/auth/passkey/credentials [get]
func (h *PasskeyHandler) ListCredentials(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	credentials, err := h.credentialRepo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		response.InternalServerError(w, "Failed to load credentials")
		return
	}

	// Format credentials for response
	credentialsList := make([]models.PasskeyCredentialInfo, 0, len(credentials))
	for _, cred := range credentials {
		item := models.PasskeyCredentialInfo{
			ID:                      cred.ID.String(),
			FriendlyName:            cred.FriendlyName,
			AuthenticatorAttachment: cred.AuthenticatorAttachment,
			BackupEligible:          cred.BackupEligible,
			BackupState:             cred.BackupState,
			CreatedAt:               cred.CreatedAt.Format(time.RFC3339),
		}
		if cred.LastUsedAt != nil {
			lastUsed := cred.LastUsedAt.Format(time.RFC3339)
			item.LastUsedAt = &lastUsed
		}
		credentialsList = append(credentialsList, item)
	}

	response.Success(w, http.StatusOK, models.PasskeyCredentialsListResponse{
		Credentials: credentialsList,
	})
}

// DeleteCredential removes a passkey
//
// @Summary      Delete passkey
// @Description  Removes a passkey for the authenticated user (cannot delete last credential)
// @Tags         Passkey
// @Produce      json
// @Param        id path string true "Credential ID" format(uuid)
// @Security     BearerAuth
// @Success      204 "Credential deleted"
// @Failure      400 {object} response.ErrorResponse "Credential ID required or cannot delete last credential"
// @Failure      401 {object} response.ErrorResponse "Authentication required"
// @Failure      500 {object} response.ErrorResponse "Failed to delete credential"
// @Router       /api/v1/auth/passkey/credentials/{id} [delete]
func (h *PasskeyHandler) DeleteCredential(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	credentialID := chi.URLParam(r, "id")
	if credentialID == "" {
		response.BadRequest(w, "Credential ID is required", nil)
		return
	}

	id, err := uuid.Parse(credentialID)
	if err != nil {
		response.BadRequest(w, "Invalid credential ID", nil)
		return
	}

	// Check if this is the last credential
	count, err := h.credentialRepo.CountByUserID(r.Context(), user.ID)
	if err != nil {
		response.InternalServerError(w, "Failed to check credentials")
		return
	}

	if count <= 1 {
		response.BadRequest(w, "Cannot delete last credential", nil)
		return
	}

	// Verify ownership before deleting
	_, err = h.credentialRepo.FindByIDForUser(r.Context(), id, user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(w, "Credential not found")
			return
		}
		response.InternalServerError(w, "Failed to load credential")
		return
	}

	// Delete the credential
	if err := h.credentialRepo.Delete(r.Context(), id, user.ID); err != nil {
		response.InternalServerError(w, "Failed to delete credential")
		return
	}

	// Log passkey deletion
	if h.auditService != nil {
		_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
			EventType: "auth.passkey.delete",
			ActorID:   &user.ID,
			Action:    "User deleted a passkey",
			Status:    "success",
			IPAddress: services.GetClientIP(r),
			UserAgent: services.GetUserAgent(r),
			Metadata: map[string]interface{}{
				"credential_id": credentialID,
			},
		})
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateCredentialNameRequest represents the update request
type UpdateCredentialNameRequest struct {
	FriendlyName string `json:"friendlyName"`
}

// UpdateCredentialName updates the friendly name of a passkey
//
// @Summary      Update passkey name
// @Description  Updates the friendly name of a passkey for the authenticated user
// @Tags         Passkey
// @Accept       json
// @Produce      json
// @Param        id path string true "Credential ID" format(uuid)
// @Param        name body UpdateCredentialNameRequest true "New friendly name"
// @Security     BearerAuth
// @Success      200 {object} models.PasskeyCredentialUpdateResponse "Updated credential"
// @Failure      400 {object} response.ErrorResponse "Invalid request"
// @Failure      401 {object} response.ErrorResponse "Authentication required"
// @Failure      403 {object} response.ErrorResponse "Not authorized to modify credential"
// @Failure      404 {object} response.ErrorResponse "Credential not found"
// @Failure      500 {object} response.ErrorResponse "Failed to update credential"
// @Router       /api/v1/auth/passkey/credentials/{id} [put]
func (h *PasskeyHandler) UpdateCredentialName(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	credentialID := chi.URLParam(r, "id")
	if credentialID == "" {
		response.BadRequest(w, "Credential ID is required", nil)
		return
	}

	id, err := uuid.Parse(credentialID)
	if err != nil {
		response.BadRequest(w, "Invalid credential ID", nil)
		return
	}

	var req UpdateCredentialNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	if req.FriendlyName == "" {
		response.BadRequest(w, "Friendly name is required", nil)
		return
	}

	// Get the credential with authorization check in single query
	credential, err := h.credentialRepo.FindByIDForUser(r.Context(), id, user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(w, "Credential not found")
			return
		}
		response.InternalServerError(w, "Failed to load credential")
		return
	}

	// Update friendly name
	credential.FriendlyName = req.FriendlyName
	if err := h.credentialRepo.Update(r.Context(), credential); err != nil {
		response.InternalServerError(w, "Failed to update credential")
		return
	}

	response.Success(w, http.StatusOK, models.PasskeyCredentialUpdateResponse{
		ID:           credential.ID.String(),
		FriendlyName: credential.FriendlyName,
	})
}

// GetMigrationStatus returns the user's migration status
//
// @Summary      Get migration status
// @Description  Returns the user's authentication migration status (password/passkey enabled)
// @Tags         Passkey
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.MigrationStatusResponse "Migration status"
// @Failure      401 {object} response.ErrorResponse "Authentication required"
// @Router       /api/v1/auth/migration/status [get]
func (h *PasskeyHandler) GetMigrationStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Get or create migration status
	var migration models.UserAuthMigration
	err := h.db.WithContext(r.Context()).Where("user_id = ?", user.ID).First(&migration).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create default migration record
		count, _ := h.credentialRepo.CountByUserID(r.Context(), user.ID)
		migration = models.UserAuthMigration{
			UserID:               user.ID,
			HasPassword:          user.PasswordHash != nil,
			HasPasskey:           count > 0,
			PasswordLoginEnabled: true,
			PasskeyLoginEnabled:  count > 0,
		}
	}

	// Check if user can disable password
	canDisablePassword := migration.HasPasskey && migration.PasswordLoginEnabled

	response.Success(w, http.StatusOK, models.MigrationStatusResponse{
		HasPassword:          migration.HasPassword,
		HasPasskey:           migration.HasPasskey,
		PasswordLoginEnabled: migration.PasswordLoginEnabled,
		PasskeyLoginEnabled:  migration.PasskeyLoginEnabled,
		CanDisablePassword:   canDisablePassword,
	})
}

// DisablePasswordRequest represents the disable password request
type DisablePasswordRequest struct {
	Confirmed bool `json:"confirmed"`
}

// DisablePasswordLogin disables password login for the user
//
// @Summary      Disable password login
// @Description  Disables password login for the user (requires at least one passkey). Returns backup codes.
// @Tags         Passkey
// @Accept       json
// @Produce      json
// @Param        confirm body DisablePasswordRequest true "Confirmation"
// @Security     BearerAuth
// @Success      200 {object} models.DisablePasswordResponse "Password disabled with backup codes"
// @Failure      400 {object} response.ErrorResponse "Confirmation required or no passkey registered"
// @Failure      401 {object} response.ErrorResponse "Authentication required"
// @Failure      500 {object} response.ErrorResponse "Failed to update or generate backup codes"
// @Router       /api/v1/auth/migration/disable-password [post]
func (h *PasskeyHandler) DisablePasswordLogin(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	var req DisablePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	if !req.Confirmed {
		response.BadRequest(w, "Confirmation required", nil)
		return
	}

	// Verify user has at least one passkey
	count, err := h.credentialRepo.CountByUserID(r.Context(), user.ID)
	if err != nil {
		response.InternalServerError(w, "Failed to check credentials")
		return
	}

	if count == 0 {
		response.BadRequest(w, "At least one passkey is required before disabling password login", nil)
		return
	}

	// Update migration status
	var migration models.UserAuthMigration
	err = h.db.WithContext(r.Context()).Where("user_id = ?", user.ID).First(&migration).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		migration = models.UserAuthMigration{
			UserID: user.ID,
		}
	} else if err != nil {
		response.InternalServerError(w, "Failed to query auth migration status")
		return
	}

	migration.PasswordLoginEnabled = false
	migration.PasskeyLoginEnabled = true

	if err := h.db.WithContext(r.Context()).Save(&migration).Error; err != nil {
		response.InternalServerError(w, "Failed to update migration status")
		return
	}

	// Generate backup codes
	backupCodes, err := h.backupCodeService.GenerateBackupCodes(user.ID)
	if err != nil {
		response.InternalServerError(w, "Failed to generate backup codes")
		return
	}

	// Log password disable
	if h.auditService != nil {
		_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
			EventType: "auth.password.disable",
			ActorID:   &user.ID,
			Action:    "User disabled password login",
			Status:    "success",
			IPAddress: services.GetClientIP(r),
			UserAgent: services.GetUserAgent(r),
		})
	}

	response.Success(w, http.StatusOK, models.DisablePasswordResponse{
		PasswordLoginEnabled: false,
		BackupCodes:          backupCodes,
		Message:              "Password login disabled. Save these backup codes securely.",
	})
}

// GenerateBackupCodes generates new backup codes for the authenticated user
//
// @Summary      Generate backup codes
// @Description  Generates new backup codes for the authenticated user (requires at least one passkey)
// @Tags         Passkey
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.BackupCodesResponse "Generated backup codes"
// @Failure      400 {object} response.ErrorResponse "No passkey registered"
// @Failure      401 {object} response.ErrorResponse "Authentication required"
// @Failure      500 {object} response.ErrorResponse "Failed to generate backup codes"
// @Router       /api/v1/auth/migration/generate-backup-codes [post]
func (h *PasskeyHandler) GenerateBackupCodes(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Verify user has at least one passkey
	count, err := h.credentialRepo.CountByUserID(r.Context(), user.ID)
	if err != nil {
		response.InternalServerError(w, "Failed to check credentials")
		return
	}

	if count == 0 {
		response.BadRequest(w, "At least one passkey is required before generating backup codes", nil)
		return
	}

	// Generate backup codes
	backupCodes, err := h.backupCodeService.GenerateBackupCodes(user.ID)
	if err != nil {
		response.InternalServerError(w, "Failed to generate backup codes")
		return
	}

	response.Success(w, http.StatusOK, models.BackupCodesResponse{
		BackupCodes: backupCodes,
		Count:       len(backupCodes),
		Message:     "Backup codes generated. Save them securely - they won't be shown again.",
	})
}

// BackupCodeAuthRequest represents a backup code authentication request
type BackupCodeAuthRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

// AuthenticateWithBackupCode authenticates a user with a backup code
//
// @Summary      Authenticate with backup code
// @Description  Authenticates a user with a backup code (code is consumed after use)
// @Tags         Passkey
// @Accept       json
// @Produce      json
// @Param        auth body BackupCodeAuthRequest true "Email and backup code"
// @Success      200 {object} models.BackupCodeAuthResponse "Authentication successful with JWT token"
// @Failure      400 {object} response.ErrorResponse "Email and code required"
// @Failure      401 {object} response.ErrorResponse "Invalid credentials or backup code"
// @Failure      500 {object} response.ErrorResponse "Failed to generate token"
// @Router       /api/v1/auth/backup-code/authenticate [post]
func (h *PasskeyHandler) AuthenticateWithBackupCode(w http.ResponseWriter, r *http.Request) {
	var req BackupCodeAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	if req.Email == "" {
		response.BadRequest(w, "Email is required", nil)
		return
	}

	if req.Code == "" {
		response.BadRequest(w, "Backup code is required", nil)
		return
	}

	// Normalize email to lowercase for case-insensitive comparison
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))

	// Find user by email
	var user models.User
	if err := h.db.WithContext(r.Context()).Where("email = ?", normalizedEmail).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Unauthorized(w, "Invalid credentials")
			return
		}
		response.InternalServerError(w, "Database error")
		return
	}

	// Verify and consume backup code
	if err := h.backupCodeService.VerifyAndConsumeBackupCode(user.ID, req.Code); err != nil {
		// Log failed backup code attempt
		if h.auditService != nil {
			_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
				EventType: "auth.passkey.backup_code",
				ActorID:   &user.ID,
				Action:    "Failed backup code authentication",
				Status:    "failure",
				IPAddress: services.GetClientIP(r),
				UserAgent: services.GetUserAgent(r),
			})
		}
		response.Unauthorized(w, "Invalid or already used backup code")
		return
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

	// Get remaining backup codes count
	remainingCodes, _ := h.backupCodeService.CountUnusedCodes(user.ID)

	// Log backup code authentication
	if h.auditService != nil {
		_ = h.auditService.LogAuditEvent(r.Context(), services.AuditEvent{
			EventType: "auth.passkey.backup_code",
			ActorID:   &user.ID,
			Action:    "User logged in via backup code",
			Status:    "success",
			IPAddress: services.GetClientIP(r),
			UserAgent: services.GetUserAgent(r),
			Metadata: map[string]interface{}{
				"remaining_codes": remainingCodes,
			},
		})
	}

	response.Success(w, http.StatusOK, models.BackupCodeAuthResponse{
		Token: token,
		User: models.UserInfoResponse{
			ID:            user.ID.String(),
			Email:         user.Email,
			IsAdmin:       user.IsAdmin,
			EmailVerified: user.EmailVerified,
		},
		RemainingBackupCodes: remainingCodes,
		Message:              "Authentication successful. Consider setting up a new passkey.",
	})
}
