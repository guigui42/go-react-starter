package handlers

import (
"encoding/json"
"net/http"

"github.com/guigui42/go-react-starter/internal/middleware"
"github.com/guigui42/go-react-starter/internal/models"
"github.com/guigui42/go-react-starter/pkg/response"
"gorm.io/gorm"
)

// UserPreferencesHandler handles user preferences-related HTTP requests.
type UserPreferencesHandler struct {
db *gorm.DB
}

// NewUserPreferencesHandler creates a new user preferences handler.
func NewUserPreferencesHandler(db *gorm.DB) *UserPreferencesHandler {
return &UserPreferencesHandler{db: db}
}

// UpdatePreferencesRequest represents the request payload for updating user preferences.
type UpdatePreferencesRequest struct {
Language               *string `json:"language,omitempty"`
ColorScheme            *string `json:"color_scheme,omitempty"`
}

// GetPreferences returns the current user's preferences.
//
// @Summary      Get user preferences
// @Description  Returns the authenticated user's preferences including language and color scheme
// @Tags         User Preferences
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.UserPreferences "User preferences"
// @Failure      401 {object} response.ErrorResponse "User not authenticated"
// @Failure      404 {object} response.ErrorResponse "Preferences not found"
// @Router       /api/v1/preferences [get]
func (h *UserPreferencesHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
// Get user ID from context (set by auth middleware)
userID, ok := middleware.GetUserID(r)
if !ok {
response.Unauthorized(w, "User ID not found in context")
return
}

// Fetch preferences from database
var prefs models.UserPreferences
if err := h.db.Where("user_id = ?", userID).First(&prefs).Error; err != nil {
if err == gorm.ErrRecordNotFound {
response.NotFound(w, "Preferences not found")
return
}
response.InternalServerError(w, "Failed to fetch preferences")
return
}

response.Success(w, http.StatusOK, prefs)
}

// UpdatePreferences updates the current user's preferences.
//
// @Summary      Update user preferences
// @Description  Updates the authenticated user's preferences. Only provided fields will be updated.
// @Tags         User Preferences
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        preferences body UpdatePreferencesRequest true "Preferences to update"
// @Success      200 {object} models.UserPreferences "Updated preferences"
// @Failure      400 {object} response.ErrorResponse "Validation failed"
// @Failure      401 {object} response.ErrorResponse "User not authenticated"
// @Failure      404 {object} response.ErrorResponse "Preferences not found"
// @Router       /api/v1/preferences [put]
func (h *UserPreferencesHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
// Get user ID from context (set by auth middleware)
userID, ok := middleware.GetUserID(r)
if !ok {
response.Unauthorized(w, "User ID not found in context")
return
}

// Decode request body
var req UpdatePreferencesRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
response.BadRequest(w, "Invalid request body", map[string]interface{}{
"error": err.Error(),
})
return
}

// Fetch existing preferences
var prefs models.UserPreferences
if err := h.db.Where("user_id = ?", userID).First(&prefs).Error; err != nil {
if err == gorm.ErrRecordNotFound {
response.NotFound(w, "Preferences not found")
return
}
response.InternalServerError(w, "Failed to fetch preferences")
return
}

// Update fields if provided
if req.Language != nil {
prefs.Language = *req.Language
}
if req.ColorScheme != nil {
prefs.ColorScheme = *req.ColorScheme
}

// Validate updated preferences
if err := prefs.Validate(); err != nil {
response.ValidationError(w, map[string]interface{}{
"error": err.Error(),
})
return
}

// Save updated preferences
if err := h.db.Save(&prefs).Error; err != nil {
response.InternalServerError(w, "Failed to update preferences")
return
}

response.Success(w, http.StatusOK, prefs)
}
