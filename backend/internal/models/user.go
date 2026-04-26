package models

import (
	_ "embed"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/unicode/norm"
	"gorm.io/gorm"
)

//go:embed common-passwords.txt
var commonPasswordsData string

// BcryptCost is the default cost factor for bcrypt password hashing.
// Recommended minimum is 12 for production systems per OWASP guidelines.
// Each increment doubles the computation time (~400ms at cost 12).
// Note: bcrypt automatically generates and embeds a unique salt per hash.
const BcryptCost = 12

// bcryptCost is the runtime cost factor, which can be overridden for testing.
// Defaults to BcryptCost (12) for production security.
var bcryptCost = BcryptCost

// SetBcryptCostForTesting allows tests to use a lower bcrypt cost for speed.
// This should ONLY be called from test code. Returns a cleanup function
// that restores the original cost.
//
// Example usage in TestMain:
//
//	func TestMain(m *testing.M) {
//	    cleanup := models.SetBcryptCostForTesting(4)
//	    defer cleanup()
//	    os.Exit(m.Run())
//	}
func SetBcryptCostForTesting(cost int) func() {
	original := bcryptCost
	bcryptCost = cost
	return func() {
		bcryptCost = original
	}
}

// GetCurrentBcryptCost returns the current bcrypt cost factor.
// This is useful for tests to verify the cost factor being used.
func GetCurrentBcryptCost() int {
	return bcryptCost
}

// dummyHash is a cached dummy password hash for constant-time login operations.
// Generated lazily on first use to match the current bcrypt cost.
var dummyHash []byte
var dummyHashOnce sync.Once

// GenerateDummyHash returns a dummy bcrypt hash for constant-time login operations.
// The hash is generated once and cached for the current bcrypt cost setting.
// This is used to prevent timing attacks when a user doesn't exist.
func GenerateDummyHash() ([]byte, error) {
	var err error
	dummyHashOnce.Do(func() {
		dummyHash, err = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcryptCost)
	})
	return dummyHash, err
}

// User represents an investor using the application.
// Passwords are hashed using bcrypt and email addresses must be unique.
// PasswordHash is now nullable to support passkey-only users.
//
// Note: EmailVerified defaults to false for new users. Existing users are
// "grandfathered" and have EmailVerified set to true during migration
// (see main.go AutoMigrate section) to prevent lockout when email
// verification is enabled.
type User struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Email           string     `gorm:"type:text;uniqueIndex;not null" json:"email"`
	PasswordHash    *string    `gorm:"type:text" json:"-"` // Nullable for passkey-only users
	IsAdmin         bool       `gorm:"type:boolean;default:false" json:"is_admin"`
	IsTestUser      bool       `gorm:"type:boolean;default:false;index" json:"is_test_user"`
	EmailVerified   bool       `gorm:"type:boolean;default:false" json:"email_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"not null" json:"updated_at"`

	// Relationships for passkey authentication
	Credentials   []UserCredential   `gorm:"foreignKey:UserID" json:"-"`
	AuthMigration *UserAuthMigration `gorm:"foreignKey:UserID" json:"-"`
}

// BeforeCreate is a GORM hook that generates a UUID for the user before creation.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = NewID()
	}
	return nil
}

// commonPasswordsSet is a lazy-loaded set of common passwords for validation.
var commonPasswordsSet map[string]bool
var commonPasswordsOnce sync.Once

// loadCommonPasswords loads the embedded common passwords file into a set.
// Thread-safe using sync.Once for lazy initialization.
func loadCommonPasswords() map[string]bool {
	commonPasswordsOnce.Do(func() {
		commonPasswordsSet = make(map[string]bool)
		lines := strings.Split(commonPasswordsData, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				commonPasswordsSet[strings.ToLower(line)] = true
			}
		}
	})
	return commonPasswordsSet
}

// extractUsername extracts the local part (before @) from an email address.
// Returns empty string if email format is invalid or username would be too short.
func extractUsername(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) < 2 || parts[0] == "" {
		return ""
	}
	username := parts[0]
	// Only check username in password if it's at least 3 characters
	// to avoid overly restrictive validation for very short usernames
	if len(username) < 3 {
		return ""
	}
	return username
}

// containsSubstringConstantTime performs constant-time substring containment check
// to prevent timing attacks that could reveal password composition information.
func containsSubstringConstantTime(haystack, needle string) bool {
	if len(needle) == 0 {
		return false
	}

	haystackLower := strings.ToLower(haystack)
	needleLower := strings.ToLower(needle)

	// If needle is longer than haystack, it can't be contained
	if len(needleLower) > len(haystackLower) {
		return false
	}

	found := 0

	// Check all possible positions to maintain constant time
	// This prevents timing attacks that could determine substring location
	for i := 0; i <= len(haystackLower)-len(needleLower); i++ {
		match := 1
		for j := 0; j < len(needleLower); j++ {
			if haystackLower[i+j] != needleLower[j] {
				match = 0
			}
		}
		// Use bitwise OR to avoid early exit (constant time)
		found |= match
	}

	return found == 1
}

// ValidatePassword validates a password against security requirements.
// Requirements based on NIST SP 800-63B:
// - Minimum 12 characters (NIST recommended)
// - Maximum 72 characters (bcrypt limit, also prevents DoS)
// - Not in common passwords list (top 10,000)
// - Does not contain email username
// - Unicode normalized (NFC)
// - No null bytes or control characters
//
// Uses constant-time comparisons where possible to prevent timing attacks.
func ValidatePassword(password, email string) error {
	// Normalize password to Unicode NFC form (NIST SP 800-63B requirement)
	// This prevents bypasses using equivalent Unicode representations
	password = norm.NFC.String(password)

	// Check for empty or whitespace-only passwords
	if strings.TrimSpace(password) == "" {
		return errors.New("password cannot be empty")
	}

	// Check for null bytes (security issue)
	if strings.Contains(password, "\x00") {
		return errors.New("password contains invalid characters")
	}

	// Validate printable characters only
	for _, r := range password {
		if !unicode.IsPrint(r) && !unicode.IsSpace(r) {
			return errors.New("password contains invalid characters")
		}
	}

	// Pre-lowercase password once to avoid timing leaks in later checks
	passwordLower := strings.ToLower(password)

	// Check minimum length
	if len(password) < 12 {
		return errors.New("password must be at least 12 characters")
	}

	// Check maximum length (bcrypt limit is 72 bytes, prevents DoS)
	if len(password) > 72 {
		return errors.New("password must not exceed 72 characters")
	}

	// Check if password contains email username (constant-time check)
	username := extractUsername(email)
	if username != "" && containsSubstringConstantTime(password, username) {
		return errors.New("password cannot contain your email address")
	}

	// Check against common passwords list
	// Map lookup is generally constant-time; we've already pre-lowercased
	commonPasswords := loadCommonPasswords()
	if commonPasswords[passwordLower] {
		return errors.New("password is too common, please choose a stronger one")
	}

	return nil
}

// SetPassword hashes and sets the user's password using bcrypt.
// It validates password strength according to NIST SP 800-63B guidelines.
// Returns an error if validation fails or if hashing fails.
func (u *User) SetPassword(password string) error {
	// Validate password strength
	if err := ValidatePassword(password, u.Email); err != nil {
		return err
	}

	// Hash password with bcrypt using the configured cost factor
	// (12 for production per OWASP guidelines, lower for tests)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return err
	}

	hashStr := string(hash)
	u.PasswordHash = &hashStr
	return nil
}

// CheckPassword verifies if the provided password matches the stored hash.
// Returns true if the password is correct, false otherwise.
// Returns false if the user has no password (passkey-only user).
func (u *User) CheckPassword(password string) bool {
	if u.PasswordHash == nil {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(password))
	return err == nil
}

// CompareHashAndPassword is a wrapper around bcrypt.CompareHashAndPassword
// exposed for constant-time login operations to prevent timing attacks.
// This allows the auth handler to perform bcrypt comparisons even when
// a user doesn't exist, maintaining consistent response times.
func CompareHashAndPassword(hash, password []byte) error {
	return bcrypt.CompareHashAndPassword(hash, password)
}

// emailRegex is a regular expression for basic email validation.
// It checks for the pattern: local-part@domain with basic character restrictions.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail checks if the provided email address has a valid format.
// Returns an error if the email format is invalid.
func ValidateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

// WebAuthn User Interface Implementation
// These methods implement the webauthn.User interface required by go-webauthn/webauthn

// WebAuthnID returns the user's ID in byte format for WebAuthn
func (u *User) WebAuthnID() []byte {
	return u.ID[:]
}

// WebAuthnName returns the user's email as the WebAuthn username
func (u *User) WebAuthnName() string {
	return u.Email
}

// WebAuthnDisplayName returns the user's email as the display name
func (u *User) WebAuthnDisplayName() string {
	return u.Email
}

// WebAuthnIcon returns an empty string (no icon URL)
func (u *User) WebAuthnIcon() string {
	return ""
}

// WebAuthnCredentials returns the user's WebAuthn credentials in the format expected by the library
func (u *User) WebAuthnCredentials() []webauthn.Credential {
	credentials := make([]webauthn.Credential, len(u.Credentials))

	for i, cred := range u.Credentials {
		// Parse transports from JSON string
		var transports []string
		if cred.Transports != "" {
			_ = json.Unmarshal([]byte(cred.Transports), &transports)
		}

		// Convert to protocol.AuthenticatorTransport
		authTransports := make([]protocol.AuthenticatorTransport, len(transports))
		for j, t := range transports {
			authTransports[j] = protocol.AuthenticatorTransport(t)
		}

		credentials[i] = webauthn.Credential{
			ID:              cred.CredentialID,
			PublicKey:       cred.PublicKey,
			AttestationType: cred.AttestationType,
			Transport:       authTransports,
			Flags: webauthn.CredentialFlags{
				UserPresent:    true,
				UserVerified:   cred.UserVerified,
				BackupEligible: cred.BackupEligible,
				BackupState:    cred.BackupState,
			},
			Authenticator: webauthn.Authenticator{
				AAGUID:    cred.AAGUID,
				SignCount: cred.SignCount,
			},
		}
	}

	return credentials
}
