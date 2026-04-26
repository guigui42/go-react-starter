package services

import (
"context"
"time"

"github.com/guigui42/go-react-starter/internal/models"
"github.com/guigui42/go-react-starter/internal/repository/scopes"
"github.com/rs/zerolog/log"
"gorm.io/gorm"
)

// AdminStats represents system overview statistics for the admin dashboard.
type AdminStats struct {
TotalUsers int `json:"total_users"`
TotalNotes int `json:"total_notes"`
}

// AdminUser represents a user as seen by an admin.
type AdminUser struct {
ID            string    `json:"id"`
Email         string    `json:"email"`
IsAdmin       bool      `json:"is_admin"`
IsTestUser    bool      `json:"is_test_user"`
EmailVerified bool      `json:"email_verified"`
NoteCount     int       `json:"note_count"`
CreatedAt     time.Time `json:"created_at"`
}

// AdminService provides admin-related operations.
type AdminService struct {
db *gorm.DB
}

// NewAdminService creates a new admin service.
func NewAdminService(db *gorm.DB) *AdminService {
return &AdminService{db: db}
}

// InitializeAdminUsers ensures that users with admin emails have the is_admin flag set.
func (s *AdminService) InitializeAdminUsers(ctx context.Context, adminEmails []string) error {
if len(adminEmails) == 0 {
return nil
}

result := s.db.WithContext(ctx).
Model(&models.User{}).
Where("email IN ?", adminEmails).
Update("is_admin", true)

if result.Error != nil {
return result.Error
}

if result.RowsAffected > 0 {
log.Info().Int64("count", result.RowsAffected).Msg("Admin users initialized")
}

return nil
}

// GetStats returns system-wide statistics.
func (s *AdminService) GetStats(ctx context.Context) (*AdminStats, error) {
stats := &AdminStats{}
// Admin queries bypass user scope guard for global counts
adminCtx := scopes.SkipUserScopeGuard(ctx)

var totalUsers int64
s.db.WithContext(adminCtx).Model(&models.User{}).Count(&totalUsers)
stats.TotalUsers = int(totalUsers)

var totalNotes int64
s.db.WithContext(adminCtx).Model(&models.Note{}).Count(&totalNotes)
stats.TotalNotes = int(totalNotes)

return stats, nil
}

// ListUsers returns all users with their note counts.
func (s *AdminService) ListUsers(ctx context.Context) ([]AdminUser, error) {
adminCtx := scopes.SkipUserScopeGuard(ctx)

var users []models.User
if err := s.db.WithContext(adminCtx).Order("created_at DESC").Find(&users).Error; err != nil {
return nil, err
}

adminUsers := make([]AdminUser, len(users))
for i, u := range users {
var noteCount int64
s.db.WithContext(adminCtx).Model(&models.Note{}).Where("user_id = ?", u.ID).Count(&noteCount)

adminUsers[i] = AdminUser{
ID:            u.ID.String(),
Email:         u.Email,
IsAdmin:       u.IsAdmin,
IsTestUser:    u.IsTestUser,
EmailVerified: u.EmailVerified,
NoteCount:     int(noteCount),
CreatedAt:     u.CreatedAt,
}
}

return adminUsers, nil
}
