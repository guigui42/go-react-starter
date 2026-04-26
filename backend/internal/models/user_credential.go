// Package models provides GORM models for the application.
// It includes domain entities with validation and database constraints.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserCredential represents a WebAuthn credential (passkey) for a user.
// Stores all necessary information for WebAuthn authentication including
// the public key, sign count for cloning detection, and device metadata.
type UserCredential struct {
	ID                        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID                    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	CredentialID              []byte     `gorm:"not null;uniqueIndex" json:"-"` // GORM auto-selects bytea/blob
	PublicKey                 []byte     `gorm:"not null" json:"-"`             // GORM auto-selects bytea/blob
	SignCount                 uint32     `gorm:"not null;default:0" json:"sign_count"`
	AAGUID                    []byte     `gorm:"not null" json:"-"`                             // GORM auto-selects bytea/blob
	Transports                string     `gorm:"type:text" json:"transports"`                   // JSON array
	BackupEligible            bool       `gorm:"not null;default:false" json:"backup_eligible"` // Can be synced
	BackupState               bool       `gorm:"not null;default:false" json:"backup_state"`    // Currently synced
	AuthenticatorAttachment   string     `gorm:"type:text" json:"authenticator_attachment"`     // "platform" or "cross-platform"
	AttestationType           string     `gorm:"type:text" json:"attestation_type"`             // Attestation format used
	AttestationObject         []byte     `json:"-"`                                             // Original attestation object
	AttestationClientDataJSON []byte     `json:"-"`                                             // Original client data JSON
	UserVerified              bool       `gorm:"not null;default:false" json:"user_verified"`   // Biometric/PIN used
	Flags                     uint8      `gorm:"not null" json:"flags"`                         // Raw protocol flags
	FriendlyName              string     `gorm:"type:text" json:"friendly_name"`                // User-assigned name
	CreatedAt                 time.Time  `gorm:"not null" json:"created_at"`
	LastUsedAt                *time.Time `json:"last_used_at,omitempty"`

	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// BeforeCreate is a GORM hook that generates a UUID for the credential before creation.
func (uc *UserCredential) BeforeCreate(tx *gorm.DB) error {
	if uc.ID == uuid.Nil {
		uc.ID = NewID()
	}
	return nil
}
