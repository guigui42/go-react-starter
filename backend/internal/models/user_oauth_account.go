// Package models provides data structures for the application API.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OAuthProvider represents supported OAuth providers.
type OAuthProvider string

const (
	OAuthProviderGitHub   OAuthProvider = "github"
	OAuthProviderGoogle   OAuthProvider = "google"
	OAuthProviderFacebook OAuthProvider = "facebook"
)

// IsValid checks if the OAuth provider is a supported provider.
func (p OAuthProvider) IsValid() bool {
	switch p {
	case OAuthProviderGitHub, OAuthProviderGoogle, OAuthProviderFacebook:
		return true
	default:
		return false
	}
}

// UserOAuthAccount represents a linked OAuth provider account for a user.
// Users can link multiple OAuth providers (GitHub, Google, Facebook) to their account
// for convenient authentication. The email is stored for reference but the primary
// user email is always the one in the User model.
//
// Note: OAuth access tokens are NOT stored - they are discarded after authentication
// since we only need them for login, not for API access to the provider.
type UserOAuthAccount struct {
	ID             uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	UserID         uuid.UUID     `gorm:"type:uuid;not null;index" json:"user_id"`
	Provider       OAuthProvider `gorm:"type:varchar(50);not null;index:idx_oauth_provider_user,unique" json:"provider"`
	ProviderUserID string        `gorm:"type:varchar(255);not null;index:idx_oauth_provider_user,unique" json:"provider_user_id"`
	ProviderEmail  string        `gorm:"type:varchar(255)" json:"provider_email,omitempty"`
	CreatedAt      time.Time     `gorm:"not null" json:"created_at"`
	UpdatedAt      time.Time     `gorm:"not null" json:"updated_at"`

	// Relationship
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName returns the table name for GORM.
func (UserOAuthAccount) TableName() string {
	return "user_oauth_accounts"
}

// BeforeCreate is a GORM hook that generates a UUID for the account before creation.
func (a *UserOAuthAccount) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = NewID()
	}
	return nil
}

// UserOAuthAccountInfo represents a simplified view of a linked OAuth account
// for API responses, excluding internal fields.
// swagger:model UserOAuthAccountInfo
type UserOAuthAccountInfo struct {
	ID            string `json:"id"`
	Provider      string `json:"provider"`
	ProviderEmail string `json:"provider_email,omitempty"`
	CreatedAt     string `json:"created_at"`
}

// ToInfo converts a UserOAuthAccount to UserOAuthAccountInfo for API responses.
func (a *UserOAuthAccount) ToInfo() UserOAuthAccountInfo {
	return UserOAuthAccountInfo{
		ID:            a.ID.String(),
		Provider:      string(a.Provider),
		ProviderEmail: a.ProviderEmail,
		CreatedAt:     a.CreatedAt.Format(time.RFC3339),
	}
}

// OAuthProvidersResponse represents the response for listing enabled OAuth providers.
// swagger:model OAuthProvidersResponse
type OAuthProvidersResponse struct {
	Providers []OAuthProviderInfo `json:"providers"`
}

// OAuthProviderInfo represents information about an OAuth provider.
// swagger:model OAuthProviderInfo
type OAuthProviderInfo struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// UserOAuthAccountsResponse represents the response for listing linked OAuth accounts.
// swagger:model UserOAuthAccountsResponse
type UserOAuthAccountsResponse struct {
	Accounts []UserOAuthAccountInfo `json:"accounts"`
}
