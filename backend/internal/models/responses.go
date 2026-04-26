package models

// AuthResponse represents the response for authentication endpoints (login/register).
// JWT token is set in httpOnly cookie, not returned in response body for security.
type AuthResponse struct {
User User `json:"user"`
}

// UserInfoResponse represents the response for the GetMe endpoint.
type UserInfoResponse struct {
ID            string `json:"id"`
Email         string `json:"email"`
IsAdmin       bool   `json:"is_admin"`
EmailVerified bool   `json:"email_verified"`
CreatedAt     string `json:"created_at"`
}

// PasskeyRegistrationBeginResponse represents the response when beginning passkey registration.
type PasskeyRegistrationBeginResponse struct {
Options   interface{} `json:"options"`
SessionID string      `json:"sessionId"`
}

// PasskeyRegistrationFinishResponse represents the response when completing passkey registration.
type PasskeyRegistrationFinishResponse struct {
Verified     bool   `json:"verified"`
CredentialID string `json:"credentialId"`
FriendlyName string `json:"friendlyName"`
CreatedAt    string `json:"createdAt"`
}

// PasskeyAuthenticationBeginResponse represents the response when beginning passkey authentication.
type PasskeyAuthenticationBeginResponse struct {
Options   interface{} `json:"options"`
SessionID string      `json:"sessionId"`
}

// PasskeyAuthenticationFinishResponse represents the response when completing passkey authentication.
type PasskeyAuthenticationFinishResponse struct {
Verified bool             `json:"verified"`
User     UserInfoResponse `json:"user"`
Token    string           `json:"token"`
}

// PasskeyCredentialInfo represents information about a passkey credential.
type PasskeyCredentialInfo struct {
ID                      string  `json:"id"`
FriendlyName            string  `json:"friendlyName"`
AuthenticatorAttachment string  `json:"authenticatorAttachment,omitempty"`
BackupEligible          bool    `json:"backupEligible"`
BackupState             bool    `json:"backupState"`
CreatedAt               string  `json:"createdAt"`
LastUsedAt              *string `json:"lastUsedAt,omitempty"`
}

// PasskeyCredentialsListResponse represents the response for listing passkey credentials.
type PasskeyCredentialsListResponse struct {
Credentials []PasskeyCredentialInfo `json:"credentials"`
}

// MigrationStatusResponse represents the response for migration status.
type MigrationStatusResponse struct {
HasPassword          bool `json:"has_password"`
HasPasskey           bool `json:"has_passkey"`
PasswordLoginEnabled bool `json:"password_login_enabled"`
PasskeyLoginEnabled  bool `json:"passkey_login_enabled"`
CanDisablePassword   bool `json:"can_disable_password"`
}

// DisablePasswordResponse represents the response when disabling password login.
type DisablePasswordResponse struct {
PasswordLoginEnabled bool     `json:"password_login_enabled"`
BackupCodes          []string `json:"backup_codes"`
Message              string   `json:"message"`
}

// BackupCodesResponse represents the response containing backup codes.
type BackupCodesResponse struct {
BackupCodes []string `json:"backup_codes"`
Count       int      `json:"count"`
Message     string   `json:"message"`
}

// BackupCodeAuthResponse represents the response for backup code authentication.
type BackupCodeAuthResponse struct {
Token                string           `json:"token"`
User                 UserInfoResponse `json:"user"`
RemainingBackupCodes int64            `json:"remaining_backup_codes"`
Message              string           `json:"message"`
}

// MessageResponse represents a simple message response.
type MessageResponse struct {
Message string `json:"message"`
}

// VerificationPendingResponse represents the response when email verification is required.
type VerificationPendingResponse struct {
Message string `json:"message"`
Email   string `json:"email"`
}

// EmailVerifiedResponse represents the response after successful email verification.
type EmailVerifiedResponse struct {
Message  string `json:"message"`
Verified bool   `json:"verified"`
}

// EmailNotVerifiedResponse represents the error response when email is not verified.
type EmailNotVerifiedResponse struct {
Code    string `json:"code"`
Message string `json:"message"`
Email   string `json:"email"`
}

// PasskeyCredentialUpdateResponse represents the response after updating a passkey credential.
type PasskeyCredentialUpdateResponse struct {
ID           string `json:"id"`
FriendlyName string `json:"friendlyName"`
}
