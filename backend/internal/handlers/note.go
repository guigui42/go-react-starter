package handlers

import (
"encoding/json"
"net/http"
"strings"

"github.com/go-chi/chi/v5"
"github.com/guigui42/go-react-starter/internal/models"
"github.com/guigui42/go-react-starter/internal/repository"
"github.com/guigui42/go-react-starter/pkg/response"
"github.com/google/uuid"
"gorm.io/gorm"
)

// NoteHandler handles CRUD operations for notes.
type NoteHandler struct {
repo *repository.NoteRepository
}

// NewNoteHandler creates a new note handler.
func NewNoteHandler(repo *repository.NoteRepository) *NoteHandler {
return &NoteHandler{repo: repo}
}

// CreateNoteRequest is the request body for creating a note.
type CreateNoteRequest struct {
Title   string `json:"title"`
Content string `json:"content"`
}

// UpdateNoteRequest is the request body for updating a note.
type UpdateNoteRequest struct {
Title   string `json:"title"`
Content string `json:"content"`
}

// Create handles POST /api/v1/notes
func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
userID, ok := getUserIDFromContext(r)
if !ok {
response.Error(w, http.StatusUnauthorized, "unauthorized", "Unauthorized", nil)
return
}

var req CreateNoteRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
response.Error(w, http.StatusBadRequest, "bad_request", "Invalid request body", nil)
return
}

req.Title = strings.TrimSpace(req.Title)
if req.Title == "" {
response.Error(w, http.StatusBadRequest, "bad_request", "Title is required", nil)
return
}
if len(req.Title) > 255 {
response.Error(w, http.StatusBadRequest, "bad_request", "Title must be 255 characters or less", nil)
return
}

note := &models.Note{
UserID:  uuid.MustParse(userID),
Title:   req.Title,
Content: strings.TrimSpace(req.Content),
}

if err := h.repo.Create(r.Context(), note); err != nil {
response.Error(w, http.StatusInternalServerError, "internal_error", "Failed to create note", nil)
return
}

response.Success(w, http.StatusCreated, note.ToResponse())
}

// List handles GET /api/v1/notes
func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
userID, ok := getUserIDFromContext(r)
if !ok {
response.Error(w, http.StatusUnauthorized, "unauthorized", "Unauthorized", nil)
return
}

limit, offset := parsePagination(r)

notes, total, err := h.repo.FindByUserID(r.Context(), userID, limit, offset)
if err != nil {
response.Error(w, http.StatusInternalServerError, "internal_error", "Failed to fetch notes", nil)
return
}

noteResponses := make([]models.NoteResponse, len(notes))
for i, n := range notes {
noteResponses[i] = n.ToResponse()
}

response.Success(w, http.StatusOK, map[string]interface{}{
"notes":  noteResponses,
"total":  total,
"limit":  limit,
"offset": offset,
})
}

// Get handles GET /api/v1/notes/{id}
func (h *NoteHandler) Get(w http.ResponseWriter, r *http.Request) {
userID, ok := getUserIDFromContext(r)
if !ok {
response.Error(w, http.StatusUnauthorized, "unauthorized", "Unauthorized", nil)
return
}

id := chi.URLParam(r, "id")
note, err := h.repo.FindByIDForUser(r.Context(), id, userID)
if err != nil {
if err == gorm.ErrRecordNotFound {
response.Error(w, http.StatusNotFound, "not_found", "Note not found", nil)
return
}
response.Error(w, http.StatusInternalServerError, "internal_error", "Failed to fetch note", nil)
return
}

response.Success(w, http.StatusOK, note.ToResponse())
}

// Update handles PUT /api/v1/notes/{id}
func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
userID, ok := getUserIDFromContext(r)
if !ok {
response.Error(w, http.StatusUnauthorized, "unauthorized", "Unauthorized", nil)
return
}

id := chi.URLParam(r, "id")
note, err := h.repo.FindByIDForUser(r.Context(), id, userID)
if err != nil {
if err == gorm.ErrRecordNotFound {
response.Error(w, http.StatusNotFound, "not_found", "Note not found", nil)
return
}
response.Error(w, http.StatusInternalServerError, "internal_error", "Failed to fetch note", nil)
return
}

var req UpdateNoteRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
response.Error(w, http.StatusBadRequest, "bad_request", "Invalid request body", nil)
return
}

req.Title = strings.TrimSpace(req.Title)
if req.Title == "" {
response.Error(w, http.StatusBadRequest, "bad_request", "Title is required", nil)
return
}
if len(req.Title) > 255 {
response.Error(w, http.StatusBadRequest, "bad_request", "Title must be 255 characters or less", nil)
return
}

note.Title = req.Title
note.Content = strings.TrimSpace(req.Content)

if err := h.repo.Update(r.Context(), note); err != nil {
response.Error(w, http.StatusInternalServerError, "internal_error", "Failed to update note", nil)
return
}

response.Success(w, http.StatusOK, note.ToResponse())
}

// Delete handles DELETE /api/v1/notes/{id}
func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
userID, ok := getUserIDFromContext(r)
if !ok {
response.Error(w, http.StatusUnauthorized, "unauthorized", "Unauthorized", nil)
return
}

id := chi.URLParam(r, "id")
if err := h.repo.Delete(r.Context(), id, userID); err != nil {
if err == gorm.ErrRecordNotFound {
response.Error(w, http.StatusNotFound, "not_found", "Note not found", nil)
return
}
response.Error(w, http.StatusInternalServerError, "internal_error", "Failed to delete note", nil)
return
}

response.Success(w, http.StatusOK, map[string]string{"message": "Note deleted"})
}
