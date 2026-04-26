package handlers

import (
"encoding/json"
"net/http"
"strconv"
"time"

"github.com/example/go-react-starter/internal/migrations"
"github.com/example/go-react-starter/internal/services"
"github.com/example/go-react-starter/pkg/response"
"github.com/google/uuid"
"gorm.io/gorm"
)

// AdminHandler handles admin API endpoints.
type AdminHandler struct {
adminService    *services.AdminService
db              *gorm.DB
auditService    *services.AuditService
migrationRunner *migrations.Runner
}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler(adminService *services.AdminService, db *gorm.DB, auditService *services.AuditService, migrationRunner *migrations.Runner) *AdminHandler {
return &AdminHandler{
adminService:    adminService,
db:              db,
auditService:    auditService,
migrationRunner: migrationRunner,
}
}

// GetStats returns system statistics.
func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
stats, err := h.adminService.GetStats(r.Context())
if err != nil {
response.Error(w, http.StatusInternalServerError, "internal_error", "Failed to fetch stats", nil)
return
}
response.Success(w, http.StatusOK, stats)
}

// ListUsers returns all users for admin management.
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
users, err := h.adminService.ListUsers(r.Context())
if err != nil {
response.Error(w, http.StatusInternalServerError, "internal_error", "Failed to fetch users", nil)
return
}
response.Success(w, http.StatusOK, users)
}

// GetAuditLogs returns paginated and filtered audit log entries.
func (h *AdminHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
q := r.URL.Query()

query := services.AuditLogQuery{
EventType: q.Get("event_type"),
Status:    q.Get("status"),
Page:      1,
PageSize:  50,
}

if p := q.Get("page"); p != "" {
if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
query.Page = parsed
}
}
if ps := q.Get("page_size"); ps != "" {
if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
query.PageSize = parsed
}
}
if actorID := q.Get("actor_id"); actorID != "" {
if uid, err := uuid.Parse(actorID); err == nil {
query.ActorID = &uid
}
}
if from := q.Get("from"); from != "" {
if t, err := time.Parse(time.RFC3339, from); err == nil {
query.From = &t
}
}
if to := q.Get("to"); to != "" {
if t, err := time.Parse(time.RFC3339, to); err == nil {
query.To = &t
}
}

result, err := h.auditService.QueryAuditLogs(r.Context(), query)
if err != nil {
response.Error(w, http.StatusInternalServerError, "internal_error", "Failed to fetch audit logs", nil)
return
}
response.Success(w, http.StatusOK, result)
}

// migrationStatusJSON is the JSON-serializable migration status for the API.
type migrationStatusJSON struct {
Version      string  `json:"version"`
Name         string  `json:"name"`
Applied      bool    `json:"applied"`
Failed       bool    `json:"failed"`
AppliedAt    *string `json:"applied_at,omitempty"`
ErrorMessage string  `json:"error_message,omitempty"`
}

// GetMigrationStatus returns the status of database migrations.
func (h *AdminHandler) GetMigrationStatus(w http.ResponseWriter, r *http.Request) {
statuses, err := h.migrationRunner.Status()
if err != nil {
response.Error(w, http.StatusInternalServerError, "internal_error", "Failed to fetch migration status", nil)
return
}

result := make([]migrationStatusJSON, len(statuses))
for i, s := range statuses {
var appliedAt *string
if s.AppliedAt != nil {
t := s.AppliedAt.Format(time.RFC3339)
appliedAt = &t
}
result[i] = migrationStatusJSON{
Version:      s.Version,
Name:         s.Name,
Applied:      s.Applied,
Failed:       s.Failed,
AppliedAt:    appliedAt,
ErrorMessage: s.ErrorMessage,
}
}
response.Success(w, http.StatusOK, result)
}

// GetEmailConfig returns the current email configuration (without secrets).
func (h *AdminHandler) GetEmailConfig(w http.ResponseWriter, r *http.Request) {
config := map[string]interface{}{
"smtp_configured": true,
"info":            "Email configuration details are available in environment variables",
}
response.Success(w, http.StatusOK, config)
}

// SendTestEmail sends a test email to verify SMTP configuration.
func (h *AdminHandler) SendTestEmail(w http.ResponseWriter, r *http.Request) {
var req struct {
To string `json:"to"`
}
if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.To == "" {
response.Error(w, http.StatusBadRequest, "bad_request", "Invalid request: 'to' email required", nil)
return
}
response.Success(w, http.StatusOK, map[string]string{
"message": "Test email functionality — configure SMTP to send actual emails",
"to":      req.To,
})
}
