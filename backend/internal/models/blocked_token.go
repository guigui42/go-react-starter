package models

import (
	"time"
)

// BlockedToken represents a revoked JWT token stored in the database.
// Tokens are blocked by their JTI (JWT ID) claim and automatically
// cleaned up after expiration.
type BlockedToken struct {
	JTI       string    `gorm:"primaryKey;size:36" json:"jti"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
