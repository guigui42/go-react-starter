package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/guigui42/go-react-starter/internal/models"
	"github.com/guigui42/go-react-starter/internal/repository/scopes"
	"gorm.io/gorm"
)

// UserCredentialRepository handles WebAuthn credential data access
type UserCredentialRepository struct {
	db *gorm.DB
}

// NewUserCredentialRepository creates a new user credential repository
func NewUserCredentialRepository(db *gorm.DB) *UserCredentialRepository {
	return &UserCredentialRepository{db: db}
}

// Create creates a new credential
func (r *UserCredentialRepository) Create(ctx context.Context, credential *models.UserCredential) error {
	return r.db.WithContext(ctx).Create(credential).Error
}

// FindByIDForUser finds a credential by its UUID if it belongs to the specified user.
// This method combines lookup and authorization in a single query to prevent timing side-channels.
func (r *UserCredentialRepository) FindByIDForUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.UserCredential, error) {
	var credential models.UserCredential
	err := r.db.WithContext(ctx).Scopes(scopes.ForUser(userID.String())).Where("id = ?", id).First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

// FindByCredentialID finds a credential by its WebAuthn credential ID (binary)
func (r *UserCredentialRepository) FindByCredentialID(ctx context.Context, credentialID []byte) (*models.UserCredential, error) {
	// Auth flow lookup — user is unknown at this point, skip the user scope guard.
	ctx = scopes.SkipUserScopeGuard(ctx)
	var credential models.UserCredential
	err := r.db.WithContext(ctx).Where("credential_id = ?", credentialID).First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

// FindByUserID finds all credentials for a user, ordered by creation date (newest first)
func (r *UserCredentialRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserCredential, error) {
	var credentials []models.UserCredential
	err := r.db.WithContext(ctx).
		Scopes(scopes.ForUser(userID.String())).
		Order("created_at DESC").
		Find(&credentials).Error
	return credentials, err
}

// Update updates a credential
func (r *UserCredentialRepository) Update(ctx context.Context, credential *models.UserCredential) error {
	return r.db.WithContext(ctx).Save(credential).Error
}

// Delete deletes a credential by ID, scoped to the given user
func (r *UserCredentialRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Scopes(scopes.ForUser(userID.String())).Delete(&models.UserCredential{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CountByUserID returns the number of credentials for a user
func (r *UserCredentialRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.UserCredential{}).
		Scopes(scopes.ForUser(userID.String())).
		Count(&count).Error
	return count, err
}

// UpdateSignCount updates the sign count and last used timestamp for a credential
func (r *UserCredentialRepository) UpdateSignCount(ctx context.Context, id uuid.UUID, userID uuid.UUID, signCount uint32) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.UserCredential{}).
		Scopes(scopes.ForUser(userID.String())).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"sign_count":   signCount,
			"last_used_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
