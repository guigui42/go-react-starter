package scopes

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

type contextKey string

const (
	// SkipUserScopeGuardKey is set in context to bypass the user scope guard
	// for legitimate system/admin queries on user-scoped tables.
	SkipUserScopeGuardKey contextKey = "skip_user_scope_guard"
)

// SkipUserScopeGuard returns a context that disables the user scope guard
// for the duration of the query. Use for system-level operations
// (migrations, admin tools, auth lookups by email, etc.).
func SkipUserScopeGuard(ctx context.Context) context.Context {
	return context.WithValue(ctx, SkipUserScopeGuardKey, true)
}

// userScopedTables lists all tables that require user_id filtering.
var userScopedTables = map[string]bool{
	"user_preferences":     true,
	"user_credentials":     true,
	"user_backup_codes":    true,
	"user_auth_migrations": true,
	"user_oauth_accounts":  true,
	"email_verifications":  true,
	"notes":                true,
}

// IsUserScopedTable returns true if the given table requires user_id filtering.
func IsUserScopedTable(table string) bool {
	return userScopedTables[table]
}

// RegisterUserScopeGuard registers GORM callbacks that detect queries
// on user-scoped models without a user_id filter.
//
// Covers Query and Delete operations. Update/Save operations are excluded
// because GORM's Save() uses the primary key (not user_id) in WHERE,
// causing false positives on legitimately loaded objects.
//
// In strict mode (development/test), it panics to catch missing filters early.
// In non-strict mode (production), it logs a warning.
func RegisterUserScopeGuard(db *gorm.DB, strict bool) {
	guard := newGuardCallback(strict)
	db.Callback().Query().Before("gorm:query").Register("app:user_scope_guard_query", guard)
	db.Callback().Delete().Before("gorm:delete").Register("app:user_scope_guard_delete", guard)
}

func newGuardCallback(strict bool) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		if db.Statement == nil {
			return
		}

		// Check if guard is bypassed via context
		if db.Statement.Context != nil {
			if skip, ok := db.Statement.Context.Value(SkipUserScopeGuardKey).(bool); ok && skip {
				return
			}
		}

		// Determine the table name from the statement
		tableName := resolveTableName(db)
		if tableName == "" {
			return
		}

		// Only check user-scoped tables
		if !userScopedTables[tableName] {
			return
		}

		// Check if the query already contains a user_id clause
		// by inspecting the WHERE conditions built so far
		if hasUserIDClause(db) {
			return
		}

		msg := "SECURITY: query on user-scoped table '" + tableName + "' without user_id filter"
		if strict {
			panic(msg)
		}
		db.Logger.Warn(db.Statement.Context, msg)
	}
}

// resolveTableName extracts the target table name from a GORM statement.
func resolveTableName(db *gorm.DB) string {
	if db.Statement.Table != "" {
		return db.Statement.Table
	}
	if db.Statement.Schema != nil {
		return db.Statement.Schema.Table
	}
	return ""
}

// hasUserIDClause checks whether the current GORM statement includes
// a user_id filter in the WHERE clause.
func hasUserIDClause(db *gorm.DB) bool {
	stmt := db.Statement
	if stmt == nil {
		return false
	}

	// Inspect only the WHERE clause expressions to avoid false positives
	// from column projections in the SELECT list.
	if whereClause, ok := stmt.Clauses["WHERE"]; ok {
		if whereClause.Expression != nil {
			tmpStmt := &gorm.Statement{DB: db}
			whereClause.Expression.Build(tmpStmt)
			if containsUserID(tmpStmt.SQL.String()) {
				return true
			}
		}
	}

	return false
}

func containsUserID(s string) bool {
	lower := strings.ToLower(s)
	return strings.Contains(lower, "user_id")
}
