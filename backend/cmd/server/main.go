package main

import (
"context"
"crypto/subtle"
"encoding/json"
"fmt"
"net/http"
"os"
"os/signal"
"syscall"
"time"

"github.com/go-chi/chi/v5"
"github.com/go-chi/chi/v5/middleware"
"github.com/example/go-react-starter/internal/config"
"github.com/example/go-react-starter/internal/handlers"
appMiddleware "github.com/example/go-react-starter/internal/middleware"
"github.com/example/go-react-starter/internal/migrations"
"github.com/example/go-react-starter/internal/models"
"github.com/example/go-react-starter/internal/providers/email"
"github.com/example/go-react-starter/internal/repository"
"github.com/example/go-react-starter/internal/repository/scopes"
"github.com/example/go-react-starter/internal/services"
"github.com/example/go-react-starter/pkg/logger"
"github.com/example/go-react-starter/pkg/observability"
"github.com/rs/zerolog"
"github.com/uptrace/opentelemetry-go-extra/otelgorm"
"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
// Initialize logger with ring buffer for admin log viewing
logLevel := getLogLevel()
env := getEnv("ENVIRONMENT", "development")
consoleMode := env != "production" && env != "prod"
log, logBuffer := logger.NewWithBuffer(logLevel, consoleMode)

// Route GORM SQL logs through zerolog
config.SetGormLogger(logger.NewGormLogger(log.ZLogger()))

version := getEnv("APP_VERSION", "dev")
log.Info(fmt.Sprintf("Starting API server (version %s)", version))

// Load configuration from environment
cfg, err := config.Load()
if err != nil {
log.Error("Failed to load configuration", err)
os.Exit(1)
}
for _, section := range cfg.LogSections() {
log.WithMetadata(zerolog.InfoLevel, fmt.Sprintf("Config [%s]", section.Name), section.Values)
}

// Load observability configuration and initialize OpenTelemetry
obsCfg := config.LoadObservability()
log.WithMetadata(zerolog.InfoLevel, fmt.Sprintf("Config [%s]", obsCfg.LogSection().Name), obsCfg.LogSection().Values)

otelShutdown, err := observability.Init(context.Background(), observability.Config{
ServiceName:       obsCfg.ServiceName,
ServiceVersion:    obsCfg.ServiceVersion,
Environment:       obsCfg.Environment,
OTelEnabled:       obsCfg.OTelEnabled,
PrometheusEnabled: obsCfg.PrometheusEnabled,
TracingEnabled:    obsCfg.TracingEnabled,
LogsEnabled:       obsCfg.LogsEnabled,
TraceEndpoint:     obsCfg.TraceEndpoint,
TraceProtocol:     obsCfg.TraceProtocol,
TraceInsecure:     obsCfg.TraceInsecure,
TraceSampleRate:   obsCfg.TraceSampleRate,
OTLPHeaders:       obsCfg.OTLPHeaders,
})
if err != nil {
log.Error("Failed to initialize OpenTelemetry", err)
os.Exit(1)
}
defer func() {
shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
if err := otelShutdown(shutdownCtx); err != nil {
log.Error("Error shutting down OpenTelemetry", err)
}
}()
if obsCfg.OTelEnabled {
log.Info("OpenTelemetry initialized successfully")
}

// Attach OTel log bridge hook to zerolog
if obsCfg.LogsEnabled {
log.AddHook(observability.NewZerologOTelHook(obsCfg.ServiceName))
log.Info("OTel log pipeline enabled — logs forwarded to collector")
}

// Initialize database connection
db, err := config.InitDatabase(cfg.Database)
if err != nil {
log.Error("Failed to connect to database", err)
os.Exit(1)
}
log.Info("Connected to database (PostgreSQL)")

// Register GORM OpenTelemetry plugin
if obsCfg.OTelEnabled {
if err := db.Use(otelgorm.NewPlugin(
otelgorm.WithDBName(cfg.Database.Database),
otelgorm.WithoutQueryVariables(),
)); err != nil {
log.Error("Failed to register GORM OTel plugin", err)
} else {
log.Info("GORM OpenTelemetry plugin registered")
}

sqlDB, err := db.DB()
if err == nil {
if err := observability.RegisterDBPoolMetrics(sqlDB); err != nil {
log.Error("Failed to register DB pool metrics", err)
} else {
log.Info("Database pool metrics registered")
}
} else {
log.Error("Failed to get sql.DB for pool metrics", err)
}
}

// Register GORM user scope guard
strictGuard := env != "production" && env != "prod"
scopes.RegisterUserScopeGuard(db, strictGuard)

// Run versioned database migrations
log.Info("Running database migrations...")
migDB, err := config.InitMigrationDatabase(cfg.Database)
if err != nil {
log.Error("Failed to open migration database connection", err)
os.Exit(1)
}
migrationRunner := migrations.NewRunner(migDB, log)
if err := migrationRunner.Up(); err != nil {
log.Error("Database migrations failed", err)
os.Exit(1)
}
if sqlDB, err := migDB.DB(); err == nil {
sqlDB.Close()
}

// Run GORM AutoMigrate
if cfg.Database.AutoMigrate {
log.Info("Running AutoMigrate...")
if err := db.AutoMigrate(models.AllModels()...); err != nil {
log.Error("AutoMigrate failed", err)
os.Exit(1)
}
log.Info("Database migrations completed successfully")
}

// Initialize background context
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Initialize logs handler
logsHandler := handlers.NewLogsHandler(logBuffer)

// Initialize CSP violation repository and handler
cspViolationRepo := repository.NewCSPViolationRepository(db)
cspHandler := handlers.NewCSPHandler(log, cspViolationRepo)

// Initialize email service
var emailService *services.EmailService
var emailProvider email.EmailProvider

if cfg.Email.SMTPHost != "" {
smtpConfig := email.SMTPConfig{
Host:        cfg.Email.SMTPHost,
Port:        cfg.Email.SMTPPort,
Username:    cfg.Email.SMTPUsername,
Password:    cfg.Email.SMTPPassword,
UseTLS:      cfg.Email.SMTPUseTLS,
FromAddress: cfg.Email.FromAddress,
FromName:    cfg.Email.FromName,
Timeout:     30 * time.Second,
}
emailProvider = email.NewSMTPProvider(smtpConfig)
log.Info("SMTP email provider initialized")
} else {
emailProvider = email.NewNoOpProvider()
if cfg.Email.VerifyEmails {
log.Warn("Email verification is enabled but SMTP is not configured. Emails will not be sent.")
}
}

emailTemplateService := services.NewEmailTemplateService(cfg.Email.TemplatesPath)
emailService = services.NewEmailService(db, emailProvider, emailTemplateService, &cfg.Email)

if cfg.Email.VerifyEmails {
log.Info("Email verification is enabled")
} else {
log.Info("Email verification is disabled")
}

// Initialize audit service
auditService := services.NewAuditService(db, log.ZLogger())

// Initialize admin service and handler
adminService := services.NewAdminService(db)
adminHandler := handlers.NewAdminHandler(adminService, db, auditService)

// Initialize admin users
if err := adminService.InitializeAdminUsers(ctx, cfg.Admin.AdminEmails); err != nil {
log.Error("Failed to initialize admin users", err)
}

// Initialize note repository and handler
noteRepo := repository.NewNoteRepository(db)
noteHandler := handlers.NewNoteHandler(noteRepo)

// Initialize token blocklist for JWT revocation on logout
tokenBlocklist := appMiddleware.NewTokenBlocklist(db, appMiddleware.WithLogger(log))
defer tokenBlocklist.Stop()

// Initialize account lockout service
lockoutService := services.NewAccountLockoutService(db, log)

// Initialize auth handler
authHandler := handlers.NewAuthHandlerWithEmail(db, emailService, auditService, &cfg.Auth, &cfg.Server, cfg.Admin.AdminEmails, lockoutService, tokenBlocklist)

// Initialize monitoring handler
monitoringHandler := handlers.NewMonitoringHandler()

// Initialize user preferences handler
userPreferencesHandler := handlers.NewUserPreferencesHandler(db)

// Initialize WebAuthn/Passkey authentication
webAuthnService, err := services.NewWebAuthnService(&cfg.WebAuthn)
if err != nil {
log.Error("Failed to initialize WebAuthn service", err)
os.Exit(1)
}
sessionStore := services.NewSessionStore()
defer sessionStore.Stop()

backupCodeService := services.NewBackupCodeService(db, 10, 16)
credentialRepo := repository.NewUserCredentialRepository(db)
passkeyHandler := handlers.NewPasskeyHandlerWithEmail(db, webAuthnService, sessionStore, credentialRepo, backupCodeService, auditService, emailService, &cfg.Auth, &cfg.Server)
log.Info("Passkey authentication initialized")

// Initialize OAuth service and handler
oauthService := services.NewOAuthService(db, &cfg.OAuth, &cfg.Auth, cfg.Admin.AdminEmails)
oauthHandler := handlers.NewOAuthHandler(oauthService, auditService, &cfg.Auth, &cfg.Server, &cfg.Email)
if cfg.OAuth.IsGitHubEnabled() {
log.Info("GitHub OAuth authentication enabled")
}
if cfg.OAuth.IsGoogleEnabled() {
log.Info("Google OAuth authentication enabled")
	if cfg.OAuth.IsGitHubEnabled() {
		log.Info("GitHub OAuth authentication enabled")
	}
	if cfg.OAuth.IsFacebookEnabled() {
		log.Info("Facebook OAuth authentication enabled")
	}
}
if cfg.OAuth.IsFacebookEnabled() {
log.Info("Facebook OAuth authentication enabled")
}

// Initialize Chi router
r := chi.NewRouter()

// Create auth and CSRF middleware instances
authMiddleware := appMiddleware.NewAuthMiddleware(cfg.Auth.JWTSecret, tokenBlocklist)
authMiddlewareWithUser := appMiddleware.NewAuthMiddlewareWithUser(db, cfg.Auth.JWTSecret, tokenBlocklist)
csrfMiddleware := appMiddleware.NewCSRFMiddleware(cfg.Server.IsSecure())

// Add Chi built-in middleware
r.Use(middleware.RequestID)
r.Use(middleware.RealIP)

// Add OpenTelemetry route labeler
if obsCfg.OTelEnabled {
r.Use(appMiddleware.OTelRouteLabeler)
}

// Add custom logging middleware
r.Use(appMiddleware.LoggingMiddleware(log))

// Add custom recovery middleware
r.Use(appMiddleware.RecoveryMiddleware)

// Enable CORS for local development
if !cfg.Server.IsSecure() {
log.Info("CORS middleware enabled for local development")
r.Use(appMiddleware.CORSMiddleware(appMiddleware.DefaultCORSConfig()))
}

// Health check endpoint
r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
})

// Prometheus metrics endpoint
if obsCfg.PrometheusEnabled {
metricsHandler := observability.PrometheusHandler()

if !cfg.Server.IsSecure() {
r.Handle("/metrics", metricsHandler)
log.Info("Prometheus metrics endpoint enabled at /metrics (unauthenticated, non-secure mode)")
} else {
metricsToken := obsCfg.MetricsAuthToken
if metricsToken == "" {
log.Warn("Prometheus metrics endpoint disabled in secure mode: METRICS_AUTH_TOKEN not set")
} else {
r.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
provided := r.Header.Get("X-Metrics-Token")
if provided == "" {
http.NotFound(w, r)
return
}
if subtle.ConstantTimeCompare([]byte(provided), []byte(metricsToken)) != 1 {
http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
return
}
metricsHandler.ServeHTTP(w, r)
}))
log.Info("Prometheus metrics endpoint enabled at /metrics (token-protected in secure mode)")
}
}
}

// CSP violation reporting endpoint
r.With(appMiddleware.MaxBodySize(10 * 1024)).Post("/api/csp-report", cspHandler.ReportViolation)

// Auth routes
r.Route("/auth", func(r chi.Router) {
r.Use(appMiddleware.MaxBodySize(appMiddleware.DefaultMaxBodySize))

// Public auth routes
r.Post("/register", authHandler.Register)
r.Post("/login", authHandler.Login)
r.Post("/check-methods", authHandler.CheckAuthMethods)
r.Post("/logout", authHandler.Logout)

// Email verification routes
r.Get("/verify-email", authHandler.VerifyEmail)
r.Post("/resend-verification", authHandler.ResendVerification)

// WebAuthn/Passkey Authentication
r.Post("/passkey/authenticate/begin", passkeyHandler.BeginAuthentication)
r.Post("/passkey/authenticate/finish", passkeyHandler.FinishAuthentication)

// Backup code authentication
r.Post("/backup-code/authenticate", passkeyHandler.AuthenticateWithBackupCode)

// OAuth routes
r.Route("/oauth", func(r chi.Router) {
r.Get("/providers", oauthHandler.ListProviders)
r.Get("/{provider}", oauthHandler.BeginOAuth)
r.Get("/{provider}/callback", oauthHandler.OAuthCallback)
})

// Protected auth routes
r.With(authMiddleware).Get("/me", authHandler.GetMe)

// Protected passkey routes - Registration
r.With(authMiddlewareWithUser).Post("/passkey/register/begin", passkeyHandler.BeginRegistration)
r.With(authMiddlewareWithUser).Post("/passkey/register/finish", passkeyHandler.FinishRegistration)

// Protected passkey routes - Migration
r.With(authMiddlewareWithUser).Route("/migration", func(r chi.Router) {
r.Get("/status", passkeyHandler.GetMigrationStatus)
r.With(csrfMiddleware).Post("/disable-password", passkeyHandler.DisablePasswordLogin)
r.With(csrfMiddleware).Post("/generate-backup-codes", passkeyHandler.GenerateBackupCodes)
})
})

// Public API routes
r.Get("/api/v1/auth/verify-email", authHandler.VerifyEmail)

// Protected API routes
r.Route("/api/v1", func(r chi.Router) {
r.Use(authMiddleware)
r.Use(csrfMiddleware)

r.Group(func(r chi.Router) {
r.Use(appMiddleware.MaxBodySize(appMiddleware.DefaultMaxBodySize))

// OAuth account management routes
r.Route("/auth/oauth", func(r chi.Router) {
r.Use(authMiddlewareWithUser)
r.Get("/accounts", oauthHandler.ListLinkedAccounts)
r.Delete("/{provider}", oauthHandler.UnlinkOAuth)
r.Get("/{provider}/link", oauthHandler.BeginOAuthLink)
})

// Notes CRUD routes
r.Route("/notes", func(r chi.Router) {
r.Post("/", noteHandler.Create)
r.Get("/", noteHandler.List)
r.Get("/{id}", noteHandler.Get)
r.Put("/{id}", noteHandler.Update)
r.Delete("/{id}", noteHandler.Delete)
})

// Monitoring routes
r.Route("/monitoring", func(r chi.Router) {
r.Get("/health", monitoringHandler.GetHealthStatus)
})

// Passkey Management routes
r.With(authMiddlewareWithUser).Route("/passkeys", func(r chi.Router) {
r.Get("/", passkeyHandler.ListCredentials)
r.Delete("/{id}", passkeyHandler.DeleteCredential)
r.Put("/{id}/name", passkeyHandler.UpdateCredentialName)
})

// User preferences routes
r.Route("/preferences", func(r chi.Router) {
r.Get("/", userPreferencesHandler.GetPreferences)
r.Put("/", userPreferencesHandler.UpdatePreferences)
})

// User account deletion route
r.Delete("/user", authHandler.DeleteUser)

// Admin routes
r.Route("/admin", func(r chi.Router) {
r.Use(appMiddleware.AdminMiddleware(db))

r.Get("/stats", adminHandler.GetStats)
r.Get("/users", adminHandler.ListUsers)

// Log viewing routes
r.Get("/logs", logsHandler.GetLogs)
r.Delete("/logs", logsHandler.ClearLogs)

// Email configuration and testing routes
r.Get("/email/config", adminHandler.GetEmailConfig)
r.Post("/email/test", adminHandler.SendTestEmail)

// CSP violation monitoring
r.Get("/csp-violations", cspHandler.GetSummary)

// Audit logs
r.Get("/audit-logs", adminHandler.GetAuditLogs)

// Database migration status
r.Get("/migrations", adminHandler.GetMigrationStatus)
})
})
})

// Start server
port := cfg.Server.Port
addr := ":" + port

var handler http.Handler = r
if obsCfg.OTelEnabled {
handler = otelhttp.NewHandler(handler, obsCfg.ServiceName,
otelhttp.WithFilter(func(r *http.Request) bool {
return r.URL.Path != "/health" && r.URL.Path != "/metrics"
}),
)
}

server := &http.Server{
Addr:              addr,
Handler:           handler,
ReadTimeout:       15 * time.Second,
ReadHeaderTimeout: 5 * time.Second,
WriteTimeout:      60 * time.Second,
IdleTimeout:       120 * time.Second,
}

// Channel to listen for interrupt signals
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

// Start server in a goroutine
go func() {
log.Info(fmt.Sprintf("Server starting on port %s", port))
if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
log.Error("Failed to start server", err)
os.Exit(1)
}
}()

// Wait for interrupt signal
<-quit
log.Info("Shutting down server...")

// Cancel background jobs context
cancel()

// Graceful shutdown with timeout
shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
defer shutdownCancel()

if err := server.Shutdown(shutdownCtx); err != nil {
log.Error("Server forced to shutdown", err)
os.Exit(1)
}

log.Info("Server exited properly")
}

// getLogLevel parses the LOG_LEVEL environment variable
func getLogLevel() zerolog.Level {
levelStr := getEnv("LOG_LEVEL", "info")
level, err := zerolog.ParseLevel(levelStr)
if err != nil {
return zerolog.InfoLevel
}
return level
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
value := os.Getenv(key)
if value == "" {
return defaultValue
}
return value
}
