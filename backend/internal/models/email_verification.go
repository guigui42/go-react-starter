// Package models provides GORM models for the application.
package models

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailVerification represents a pending email verification token.
// Tokens are 32-byte crypto-random values with 24-hour expiry.
type EmailVerification struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Token      string     `gorm:"type:text;not null;uniqueIndex" json:"-"`
	ExpiresAt  time.Time  `gorm:"not null" json:"expires_at"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	CreatedAt  time.Time  `gorm:"not null" json:"created_at"`

	// Relationship
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// BeforeCreate is a GORM hook that generates a UUID before creation.
func (e *EmailVerification) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = NewID()
	}
	return nil
}

// GenerateVerificationToken generates a cryptographically secure verification token.
// Returns a URL-safe base64 encoded 32-byte random token.
func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// IsExpired returns true if the verification token has expired.
func (e *EmailVerification) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// IsVerified returns true if the email has been verified.
func (e *EmailVerification) IsVerified() bool {
	return e.VerifiedAt != nil
}

// NewEmailVerification creates a new email verification with a 24-hour expiry.
func NewEmailVerification(userID uuid.UUID) (*EmailVerification, error) {
	token, err := GenerateVerificationToken()
	if err != nil {
		return nil, err
	}

	return &EmailVerification{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}, nil
}
