// Package models provides GORM models for the application.
// It includes domain entities with validation and database constraints.
package models

import (
	"time"

	"github.com/google/uuid"
)

// UserAuthMigration tracks user authentication migration status from password to passkey.
// One record per user (user_id is primary key).
type UserAuthMigration struct {
	UserID               uuid.UUID  `gorm:"type:uuid;primaryKey" json:"user_id"`
	HasPassword          bool       `gorm:"not null;default:true" json:"has_password"`
	HasPasskey           bool       `gorm:"not null;default:false;index:idx_migration_status" json:"has_passkey"`
	PasswordLoginEnabled bool       `gorm:"not null;default:true;index:idx_migration_status" json:"password_login_enabled"`
	PasskeyLoginEnabled  bool       `gorm:"not null;default:false" json:"passkey_login_enabled"`
	MigrationStartedAt   *time.Time `json:"migration_started_at,omitempty"`
	MigrationCompletedAt *time.Time `json:"migration_completed_at,omitempty"`
	LastPasswordLogin    *time.Time `json:"last_password_login,omitempty"`

	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}
