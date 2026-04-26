package models

import (
"errors"
"time"

"github.com/google/uuid"
"gorm.io/gorm"
)

// UserPreferences represents user-specific settings and preferences.
type UserPreferences struct {
ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
UserID          uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
Language        string    `gorm:"type:text;not null;default:'en'" json:"language"`
ColorScheme     string    `gorm:"type:text;not null;default:'auto'" json:"color_scheme"`
DigestFrequency string    `gorm:"type:text;not null;default:'never'" json:"digest_frequency"`
CreatedAt       time.Time `gorm:"not null" json:"created_at"`
UpdatedAt       time.Time `gorm:"not null" json:"updated_at"`

// Relationship
User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// BeforeCreate is a GORM hook that generates a UUID for the preferences before creation.
func (p *UserPreferences) BeforeCreate(tx *gorm.DB) error {
if p.ID == uuid.Nil {
p.ID = NewID()
}
return nil
}

// Validate checks all validation rules for user preferences.
func (p *UserPreferences) Validate() error {
if p.Language != "en" && p.Language != "fr" {
return errors.New("language must be 'en' or 'fr'")
}

if p.ColorScheme != "light" && p.ColorScheme != "dark" && p.ColorScheme != "auto" {
return errors.New("color_scheme must be 'light', 'dark', or 'auto'")
}

if p.DigestFrequency == "" {
p.DigestFrequency = "never"
}
switch p.DigestFrequency {
case "never", "daily", "weekly", "monthly":
// valid
default:
return errors.New("digest_frequency must be 'never', 'daily', 'weekly', or 'monthly'")
}

return nil
}
