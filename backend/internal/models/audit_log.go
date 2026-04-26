package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLog represents a security audit trail entry for sensitive operations.
// Stores structured log data for investigation and compliance purposes.
type AuditLog struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	EventType string         `gorm:"not null;index;size:50" json:"event_type"` // e.g., "user.delete", "auth.login"
	ActorID   *uuid.UUID     `gorm:"type:uuid;index" json:"actor_id"`          // User who performed the action
	TargetID  *uuid.UUID     `gorm:"type:uuid;index" json:"target_id"`         // User or resource affected
	Action    string         `gorm:"not null;size:100" json:"action"`          // Human-readable action description
	Status    string         `gorm:"not null;size:20" json:"status"`           // "success" or "failure"
	IPAddress *string        `gorm:"size:45" json:"ip_address"`                // IPv4 or IPv6
	UserAgent *string        `gorm:"size:500" json:"user_agent"`               // Browser/client info
	Metadata  json.RawMessage `gorm:"type:jsonb" json:"metadata"`                // Structured additional data
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete for compliance
}

// BeforeCreate generates a UUIDv7 for the audit log before creation.
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = NewID()
	}
	return nil
}

// TableName specifies the table name for AuditLog
func (AuditLog) TableName() string {
	return "audit_logs"
}
