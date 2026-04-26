package config

import (
"fmt"
"os"
"strings"
"time"

"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
Server   ServerConfig
Database DatabaseConfig
Auth     AuthConfig
WebAuthn WebAuthnConfig
Admin    AdminConfig
Email    EmailConfig
OAuth    OAuthConfig
}

// ServerConfig holds server settings
type ServerConfig struct {
Port        string
TrustProxy  bool   // Only trust X-Forwarded-For/X-Real-IP headers when behind a known reverse proxy
Environment string // Deployment environment (e.g., "production", "development")
}

// IsSecure returns true if the application is running in a production environment.
func (c *ServerConfig) IsSecure() bool {
return c.Environment == "production" || c.Environment == "prod"
}

// AuthConfig holds authentication settings
type AuthConfig struct {
JWTSecret       string
SessionDuration time.Duration
EncryptionKey   string
}

// WebAuthnConfig holds WebAuthn/Passkey configuration
type WebAuthnConfig struct {
RPID     string // Relying Party ID (e.g., "localhost")
RPOrigin string // Relying Party Origin (e.g., "http://localhost:5173")
RPName   string // Display name (e.g., "My App")
Timeout  int    // Timeout in milliseconds (default: 60000)
}

// AdminConfig holds admin configuration settings
type AdminConfig struct {
AdminEmails []string // Comma-separated list of admin email addresses
}

// EmailConfig holds email service configuration
type EmailConfig struct {
SMTPHost      string // SMTP server hostname
SMTPPort      int    // SMTP server port (default: 587)
SMTPUsername  string // SMTP authentication username
SMTPPassword  string // SMTP authentication password
SMTPUseTLS    bool   // Whether to use TLS (default: true)
FromAddress   string // Sender email address
FromName      string // Sender display name
VerifyEmails  bool   // Whether email verification is required (default: false)
TemplatesPath string // Path to email templates directory
BaseURL       string // Base URL for email links (e.g., "http://localhost:5173")
}

// OAuthConfig holds OAuth provider configuration
type OAuthConfig struct {
// GitHub OAuth configuration
GitHubClientID     string
GitHubClientSecret string

// Google OAuth configuration
GoogleClientID     string
GoogleClientSecret string

// Facebook OAuth configuration
FacebookClientID     string
FacebookClientSecret string

// CallbackBaseURL is the backend URL where OAuth providers redirect after authorization.
CallbackBaseURL string
StateDuration   time.Duration // Duration for OAuth state JWT validity (default: 10 minutes)
}

// IsGitHubEnabled returns true if GitHub OAuth is configured.
func (c *OAuthConfig) IsGitHubEnabled() bool {
return c.GitHubClientID != "" && c.GitHubClientSecret != ""
}

// IsGoogleEnabled returns true if Google OAuth is configured.
func (c *OAuthConfig) IsGoogleEnabled() bool {
return c.GoogleClientID != "" && c.GoogleClientSecret != ""
}

// IsFacebookEnabled returns true if Facebook OAuth is configured.
func (c *OAuthConfig) IsFacebookEnabled() bool {
return c.FacebookClientID != "" && c.FacebookClientSecret != ""
}

// ConfigSection holds a named group of configuration values for structured logging.
type ConfigSection struct {
Name   string
Values map[string]interface{}
}

// LogSections returns all configuration sections for structured logging.
func (c *Config) LogSections() []ConfigSection {
return []ConfigSection{
{
Name: "Server",
Values: map[string]interface{}{
"port":        c.Server.Port,
"environment": c.Server.Environment,
"trust_proxy": c.Server.TrustProxy,
},
},
{
Name: "Database",
Values: map[string]interface{}{
"host":         c.Database.Host,
"port":         c.Database.Port,
"database":     c.Database.Database,
"schema":       c.Database.Schema,
"ssl_mode":     c.Database.SSLMode,
"auto_migrate": c.Database.AutoMigrate,
},
},
{
Name: "Auth",
Values: map[string]interface{}{
"session_duration": c.Auth.SessionDuration.String(),
"jwt_secret_set":  c.Auth.JWTSecret != "",
},
},
{
Name: "WebAuthn",
Values: map[string]interface{}{
"rp_id":     c.WebAuthn.RPID,
"rp_origin": c.WebAuthn.RPOrigin,
"rp_name":   c.WebAuthn.RPName,
"timeout":   c.WebAuthn.Timeout,
},
},
{
Name: "Email",
Values: map[string]interface{}{
"smtp_host":      c.Email.SMTPHost,
"from_address":   c.Email.FromAddress,
"verify_emails":  c.Email.VerifyEmails,
"base_url":       c.Email.BaseURL,
},
},
{
Name: "OAuth",
Values: map[string]interface{}{
"github_enabled":    c.OAuth.IsGitHubEnabled(),
"google_enabled":    c.OAuth.IsGoogleEnabled(),
"facebook_enabled":  c.OAuth.IsFacebookEnabled(),
"callback_base_url": c.OAuth.CallbackBaseURL,
},
},
}
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
// Load .env file (ignore error if not found - use env vars)
_ = godotenv.Load()

cfg := &Config{
Server: ServerConfig{
Port:        getEnv("PORT", "8080"),
TrustProxy:  parseBool(getEnv("TRUST_PROXY", "false")),
Environment: getEnv("ENVIRONMENT", "development"),
},
Database: DatabaseConfig{
Host:        getEnv("PGHOST", "localhost"),
Port:        getEnv("PGPORT", "5432"),
User:        getEnv("PGUSER", "postgres"),
Password:    getEnv("PGPASSWORD", ""),
Database:    getEnv("PGDATABASE", "starter"),
SSLMode:     getEnv("PGSSLMODE", "require"),
Schema:      getEnv("PGSCHEMA", "public"),
AutoMigrate: parseBool(getEnv("AUTO_MIGRATE", "true")),
},
Auth: AuthConfig{
JWTSecret:       getEnv("JWT_SECRET", ""),
SessionDuration: parseDuration(getEnv("SESSION_DURATION", "24h")),
EncryptionKey:   getEnv("ENCRYPTION_KEY", ""),
},
WebAuthn: WebAuthnConfig{
RPID:     getEnv("WEBAUTHN_RP_ID", "localhost"),
RPOrigin: getEnv("WEBAUTHN_RP_ORIGIN", "http://localhost:5173"),
RPName:   getEnv("WEBAUTHN_RP_NAME", "Go React Starter"),
Timeout:  parseInt(getEnv("WEBAUTHN_TIMEOUT", "60000")),
},
Admin: AdminConfig{
AdminEmails: parseStringSlice(getEnv("ADMIN_EMAILS", "")),
},
Email: EmailConfig{
SMTPHost:      getEnv("SMTP_HOST", ""),
SMTPPort:      parseInt(getEnv("SMTP_PORT", "587")),
SMTPUsername:  getEnv("SMTP_USERNAME", ""),
SMTPPassword:  getEnv("SMTP_PASSWORD", ""),
SMTPUseTLS:    parseBool(getEnv("SMTP_USE_TLS", "true")),
FromAddress:   getEnv("EMAIL_FROM", "noreply@example.com"),
FromName:      getEnv("EMAIL_FROM_NAME", "Go React Starter"),
VerifyEmails:  parseBool(getEnv("EMAIL_VERIFY_REQUIRED", "false")),
TemplatesPath: getEnv("EMAIL_TEMPLATES_PATH", "./templates/emails"),
BaseURL:       getEnv("EMAIL_BASE_URL", "http://localhost:5173"),
},
OAuth: OAuthConfig{
GitHubClientID:       getEnv("OAUTH_GITHUB_CLIENT_ID", ""),
GitHubClientSecret:   getEnv("OAUTH_GITHUB_CLIENT_SECRET", ""),
GoogleClientID:       getEnv("OAUTH_GOOGLE_CLIENT_ID", ""),
GoogleClientSecret:   getEnv("OAUTH_GOOGLE_CLIENT_SECRET", ""),
FacebookClientID:     getEnv("OAUTH_FACEBOOK_CLIENT_ID", ""),
FacebookClientSecret: getEnv("OAUTH_FACEBOOK_CLIENT_SECRET", ""),
CallbackBaseURL:      getEnv("OAUTH_CALLBACK_URL", "http://localhost:8080"),
StateDuration:        parseDuration(getEnv("OAUTH_STATE_DURATION", "10m")),
},
}

if err := cfg.Validate(); err != nil {
return nil, err
}

return cfg, nil
}

// Validate checks required configuration values
func (c *Config) Validate() error {
if c.Auth.JWTSecret == "" {
return fmt.Errorf("JWT_SECRET is required")
}
if len(c.Auth.JWTSecret) < 32 {
return fmt.Errorf("JWT_SECRET must be at least 32 characters for security")
}
if c.Auth.EncryptionKey == "" {
return fmt.Errorf("ENCRYPTION_KEY is required")
}
if c.WebAuthn.RPID == "" {
return fmt.Errorf("WEBAUTHN_RP_ID is required")
}
if c.WebAuthn.RPOrigin == "" {
return fmt.Errorf("WEBAUTHN_RP_ORIGIN is required")
}
if c.WebAuthn.RPName == "" {
return fmt.Errorf("WEBAUTHN_RP_NAME is required")
}
if c.Database.Host == "" {
return fmt.Errorf("PGHOST is required")
}
if c.Database.User == "" {
return fmt.Errorf("PGUSER is required")
}
if c.Database.Database == "" {
return fmt.Errorf("PGDATABASE is required")
}
if c.Database.Password == "" {
return fmt.Errorf("PGPASSWORD is required")
}

return nil
}

// Helper functions

func getEnv(key, defaultValue string) string {
if value, exists := os.LookupEnv(key); exists {
return value
}
return defaultValue
}

func parseBool(s string) bool {
return s == "true" || s == "1" || s == "yes"
}

func parseInt(s string) int {
var n int
fmt.Sscanf(s, "%d", &n)
return n
}

func parseDuration(s string) time.Duration {
d, err := time.ParseDuration(s)
if err != nil {
return 24 * time.Hour
}
return d
}

func parseStringSlice(s string) []string {
if s == "" {
return nil
}
parts := strings.Split(s, ",")
result := make([]string, 0, len(parts))
for _, p := range parts {
trimmed := strings.TrimSpace(p)
if trimmed != "" {
result = append(result, trimmed)
}
}
return result
}
