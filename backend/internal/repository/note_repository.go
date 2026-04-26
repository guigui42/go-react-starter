package repository

import (
"context"

"github.com/example/go-react-starter/internal/models"
"github.com/example/go-react-starter/internal/repository/scopes"
"gorm.io/gorm"
)

// NoteRepository implements NoteRepositoryInterface.
type NoteRepository struct {
db *gorm.DB
}

// NewNoteRepository creates a new note repository.
func NewNoteRepository(db *gorm.DB) *NoteRepository {
return &NoteRepository{db: db}
}

// Create inserts a new note.
func (r *NoteRepository) Create(ctx context.Context, note *models.Note) error {
return r.db.WithContext(ctx).Create(note).Error
}

// FindByIDForUser returns a single note by ID for a specific user.
func (r *NoteRepository) FindByIDForUser(ctx context.Context, id string, userID string) (*models.Note, error) {
var note models.Note
err := r.db.WithContext(ctx).
Scopes(scopes.ForUser(userID)).
Where("id = ?", id).
First(&note).Error
if err != nil {
return nil, err
}
return &note, nil
}

// FindByUserID returns all notes for a user with pagination.
func (r *NoteRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]models.Note, int, error) {
var notes []models.Note
var total int64

baseQuery := r.db.WithContext(ctx).
Scopes(scopes.ForUser(userID)).
Model(&models.Note{})

if err := baseQuery.Count(&total).Error; err != nil {
return nil, 0, err
}

if err := baseQuery.
Order("created_at DESC").
Limit(limit).
Offset(offset).
Find(&notes).Error; err != nil {
return nil, 0, err
}

return notes, int(total), nil
}

// Update saves changes to an existing note.
func (r *NoteRepository) Update(ctx context.Context, note *models.Note) error {
return r.db.WithContext(ctx).Save(note).Error
}

// Delete removes a note by ID for a specific user.
func (r *NoteRepository) Delete(ctx context.Context, id string, userID string) error {
result := r.db.WithContext(ctx).
Scopes(scopes.ForUser(userID)).
Where("id = ?", id).
Delete(&models.Note{})
if result.Error != nil {
return result.Error
}
if result.RowsAffected == 0 {
return gorm.ErrRecordNotFound
}
return nil
}
