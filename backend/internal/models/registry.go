package models

// AllModels returns all GORM model structs in dependency order for AutoMigrate.
// This is the single source of truth for the list of models used by:
//   - cmd/server/main.go (server startup AutoMigrate)
//   - internal/testutil/database.go (test setup)
//
// Order matters: parent tables must come before child tables
// to satisfy foreign key constraints.
func AllModels() []interface{} {
return []interface{}{
&User{},
&UserPreferences{},
&UserCredential{},
&UserAuthMigration{},
&UserBackupCode{},
&EmailVerification{},
&UserOAuthAccount{},
&LoginAttempt{},
&BlockedToken{},
&AuditLog{},
&CSPViolation{},
&Note{},
}
}
