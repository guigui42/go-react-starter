package models

import (
	"time"
)

// LoginAttempt tracks failed login attempts per email address for account lockout.
// After MaxFailedAttempts consecutive failures, the account is locked for LockoutDuration.
type LoginAttempt struct {
	Email        string     `gorm:"primaryKey;size:255" json:"email"`
	FailedCount  int        `gorm:"not null;default:0" json:"failed_count"`
	LastFailedAt *time.Time `json:"last_failed_at"`
	LockedUntil  *time.Time `json:"locked_until"`
}

const (
	// MaxFailedAttempts is the number of consecutive failures before lockout.
	MaxFailedAttempts = 5
	// LockoutDuration is how long an account is locked after max failures.
	LockoutDuration = 15 * time.Minute
)

// IsLocked returns true if the account is currently locked.
func (la *LoginAttempt) IsLocked() bool {
	return la.LockedUntil != nil && time.Now().Before(*la.LockedUntil)
}
