package repository

import (
"context"

"github.com/example/go-react-starter/internal/models"
)

// NoteRepositoryInterface defines the interface for note data access.
type NoteRepositoryInterface interface {
Create(ctx context.Context, note *models.Note) error
FindByIDForUser(ctx context.Context, id string, userID string) (*models.Note, error)
FindByUserID(ctx context.Context, userID string, limit, offset int) ([]models.Note, int, error)
Update(ctx context.Context, note *models.Note) error
Delete(ctx context.Context, id string, userID string) error
}
