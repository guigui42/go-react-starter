package handlers

import (
"encoding/json"
"net/http"
"time"

"github.com/guigui42/go-react-starter/internal/models"
"github.com/guigui42/go-react-starter/internal/repository/scopes"
"github.com/guigui42/go-react-starter/pkg/response"
"gorm.io/gorm"
)

// UserExportHandler handles GDPR user data export.
type UserExportHandler struct {
db *gorm.DB
}

// NewUserExportHandler creates a new user export handler.
func NewUserExportHandler(db *gorm.DB) *UserExportHandler {
return &UserExportHandler{db: db}
}

// UserExportData contains all user data for GDPR export.
type UserExportData struct {
ExportedAt string           `json:"exported_at"`
User       UserExportUser   `json:"user"`
Notes      []UserExportNote `json:"notes"`
}

// UserExportUser contains user profile data.
type UserExportUser struct {
ID            string `json:"id"`
Email         string `json:"email"`
EmailVerified bool   `json:"email_verified"`
CreatedAt     string `json:"created_at"`
}

// UserExportNote contains note data.
type UserExportNote struct {
ID        string `json:"id"`
Title     string `json:"title"`
Content   string `json:"content"`
CreatedAt string `json:"created_at"`
UpdatedAt string `json:"updated_at"`
}

// ExportUserData exports all user data as JSON (GDPR compliance).
func (h *UserExportHandler) ExportUserData(w http.ResponseWriter, r *http.Request) {
userID, ok := getUserIDFromContext(r)
if !ok {
response.Error(w, http.StatusUnauthorized, "unauthorized", "Unauthorized", nil)
return
}

var user models.User
if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
response.Error(w, http.StatusNotFound, "not_found", "User not found", nil)
return
}

exportData := UserExportData{
ExportedAt: time.Now().UTC().Format(time.RFC3339),
User: UserExportUser{
ID:            user.ID.String(),
Email:         user.Email,
EmailVerified: user.EmailVerified,
CreatedAt:     user.CreatedAt.Format(time.RFC3339),
},
Notes: []UserExportNote{},
}

var notes []models.Note
if err := h.db.Scopes(scopes.ForUser(userID)).Order("created_at DESC").Find(&notes).Error; err == nil {
exportData.Notes = make([]UserExportNote, len(notes))
for i, n := range notes {
exportData.Notes[i] = UserExportNote{
ID:        n.ID.String(),
Title:     n.Title,
Content:   n.Content,
CreatedAt: n.CreatedAt.Format(time.RFC3339),
UpdatedAt: n.UpdatedAt.Format(time.RFC3339),
}
}
}

w.Header().Set("Content-Type", "application/json")
w.Header().Set("Content-Disposition", "attachment; filename=user-data-export.json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(exportData)
}
