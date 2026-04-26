package services

import (
"fmt"
"time"

"github.com/go-webauthn/webauthn/protocol"
"github.com/go-webauthn/webauthn/webauthn"
"github.com/example/go-react-starter/internal/config"
"github.com/example/go-react-starter/internal/models"
)

// WebAuthnService handles WebAuthn operations for passkey authentication
type WebAuthnService struct {
webAuthn *webauthn.WebAuthn
}

// NewWebAuthnService creates a new WebAuthn service with the given configuration
func NewWebAuthnService(cfg *config.WebAuthnConfig) (*WebAuthnService, error) {
if cfg.RPID == "" {
return nil, fmt.Errorf("RPID is required")
}
if cfg.RPOrigin == "" {
return nil, fmt.Errorf("RPOrigin is required")
}
if cfg.RPName == "" {
return nil, fmt.Errorf("RPName is required")
}

timeout := time.Duration(cfg.Timeout) * time.Millisecond

webAuthnConfig := &webauthn.Config{
RPDisplayName: cfg.RPName,
RPID:          cfg.RPID,
RPOrigins:     []string{cfg.RPOrigin},
Timeouts: webauthn.TimeoutsConfig{
Login: webauthn.TimeoutConfig{
Enforce:    true,
Timeout:    timeout,
TimeoutUVD: timeout,
},
Registration: webauthn.TimeoutConfig{
Enforce:    true,
Timeout:    timeout,
TimeoutUVD: timeout,
},
},
}

wa, err := webauthn.New(webAuthnConfig)
if err != nil {
return nil, fmt.Errorf("failed to create WebAuthn instance: %w", err)
}

return &WebAuthnService{
webAuthn: wa,
}, nil
}

// BeginRegistration initiates the WebAuthn registration process
func (s *WebAuthnService) BeginRegistration(user *models.User) (*protocol.CredentialCreation, *webauthn.SessionData, error) {
options := []webauthn.RegistrationOption{
webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
AuthenticatorAttachment: protocol.Platform,
ResidentKey:             protocol.ResidentKeyRequirementRequired,
RequireResidentKey:      protocol.ResidentKeyRequired(),
UserVerification:        protocol.VerificationRequired,
}),
}

creation, sessionData, err := s.webAuthn.BeginRegistration(user, options...)
if err != nil {
return nil, nil, fmt.Errorf("failed to begin registration: %w", err)
}

return creation, sessionData, nil
}

// FinishRegistration completes the WebAuthn registration process
func (s *WebAuthnService) FinishRegistration(user *models.User, sessionData webauthn.SessionData, response *protocol.ParsedCredentialCreationData) (*webauthn.Credential, error) {
credential, err := s.webAuthn.CreateCredential(user, sessionData, response)
if err != nil {
return nil, fmt.Errorf("failed to finish registration: %w", err)
}

return credential, nil
}

// BeginAuthentication initiates the WebAuthn authentication process
func (s *WebAuthnService) BeginAuthentication(user *models.User) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
options := []webauthn.LoginOption{
webauthn.WithUserVerification(protocol.VerificationRequired),
}

assertion, sessionData, err := s.webAuthn.BeginLogin(user, options...)
if err != nil {
return nil, nil, fmt.Errorf("failed to begin authentication: %w", err)
}

return assertion, sessionData, nil
}

// FinishAuthentication completes the WebAuthn authentication process
func (s *WebAuthnService) FinishAuthentication(user *models.User, sessionData webauthn.SessionData, response *protocol.ParsedCredentialAssertionData) (*webauthn.Credential, error) {
credential, err := s.webAuthn.ValidateLogin(user, sessionData, response)
if err != nil {
return nil, fmt.Errorf("failed to finish authentication: %w", err)
}

return credential, nil
}

// GetWebAuthn returns the underlying WebAuthn instance
func (s *WebAuthnService) GetWebAuthn() *webauthn.WebAuthn {
return s.webAuthn
}
