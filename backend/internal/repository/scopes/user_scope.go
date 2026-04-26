package scopes

import (
	"gorm.io/gorm"
)

// ForUser returns a GORM scope that filters records by user_id.
// Apply this to any query on a user-scoped model to ensure data isolation.
// Pass an optional table name to qualify the column for JOINed queries
// (e.g., ForUser(userID, "trades") produces "trades.user_id = ?").
func ForUser(userID string, table ...string) func(db *gorm.DB) *gorm.DB {
	col := "user_id"
	if len(table) > 0 && table[0] != "" {
		col = table[0] + ".user_id"
	}
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(col+" = ?", userID)
	}
}
