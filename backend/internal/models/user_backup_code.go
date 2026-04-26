package models

import (
	"time"

	"github.com/google/uuid"
)

// UserBackupCode represents a backup code for account recovery.
// Backup codes are one-time use codes that allow users to authenticate
// when they lose access to their passkeys.
type UserBackupCode struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	CodeHash  string     `gorm:"type:text;not null" json:"-"` // bcrypt hash of the code
	Used      bool       `gorm:"not null;default:false" json:"used"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `gorm:"not null" json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"` // Optional expiration

	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName specifies the table name for the UserBackupCode model.
func (UserBackupCode) TableName() string {
	return "user_backup_codes"
}
