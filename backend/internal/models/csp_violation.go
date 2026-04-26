package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CSPViolation represents a Content Security Policy violation report from the browser.
// These are logged for security monitoring and analysis.
type CSPViolation struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DocumentURI        string    `gorm:"type:text;not null;index" json:"document_uri"`
	ViolatedDirective  string    `gorm:"type:text;not null;index" json:"violated_directive"`
	BlockedURI         string    `gorm:"type:text;not null" json:"blocked_uri"`
	EffectiveDirective string    `gorm:"type:text;not null" json:"effective_directive"`
	UserAgent          string    `gorm:"type:text" json:"user_agent"`
	RemoteAddr         string    `gorm:"type:text" json:"remote_addr"`
	CreatedAt          time.Time `gorm:"not null;index" json:"created_at"`
}

// BeforeCreate generates UUID for the CSP violation before creation.
func (c *CSPViolation) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = NewID()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	return nil
}

// CSPViolationSummary provides aggregated statistics about CSP violations.
type CSPViolationSummary struct {
	TotalCount         int64            `json:"total_count"`
	Last24HoursCount   int64            `json:"last_24_hours_count"`
	Last7DaysCount     int64            `json:"last_7_days_count"`
	TopViolatedDomains []ViolationCount `json:"top_violated_domains"`
	TopDirectives      []ViolationCount `json:"top_directives"`
	RecentViolations   []CSPViolation   `json:"recent_violations"`
}

// ViolationCount represents a count of violations by a specific category.
type ViolationCount struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}
