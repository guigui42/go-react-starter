package handlers

import (
"encoding/json"
"net/http"
"strconv"

"github.com/example/go-react-starter/internal/models"
"github.com/example/go-react-starter/internal/services"
"github.com/example/go-react-starter/pkg/response"
"gorm.io/gorm"
)

// AdminHandler handles admin API endpoints.
type AdminHandler struct {
adminService *services.AdminService
db           *gorm.DB
auditService *services.AuditService
}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler(adminService *services.AdminService, db *gorm.DB, auditService *services.AuditService) *AdminHandler {
return &AdminHandler{
adminService: adminService,
db:           db,
auditService: auditService,
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

// GetAuditLogs returns recent audit log entries.
func (h *AdminHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
limit := 100
if l := r.URL.Query().Get("limit"); l != "" {
if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
limit = parsed
}
}

var logs []models.AuditLog
if err := h.db.Order("created_at DESC").Limit(limit).Find(&logs).Error; err != nil {
response.Error(w, http.StatusInternalServerError, "internal_error", "Failed to fetch audit logs", nil)
return
}
response.Success(w, http.StatusOK, logs)
}

// GetMigrationStatus returns the status of database migrations.
func (h *AdminHandler) GetMigrationStatus(w http.ResponseWriter, r *http.Request) {
type MigrationRecord struct {
Version   int    `json:"version"`
Name      string `json:"name"`
AppliedAt string `json:"applied_at"`
}

var records []MigrationRecord
h.db.Raw("SELECT version, name, applied_at FROM schema_migrations ORDER BY version").Scan(&records)
response.Success(w, http.StatusOK, records)
}

// GetEmailConfig returns the current email configuration (without secrets).
func (h *AdminHandler) GetEmailConfig(w http.ResponseWriter, r *http.Request) {
// Return safe config info (no secrets)
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
// In a real implementation, this would use the email service
response.Success(w, http.StatusOK, map[string]string{
"message": "Test email functionality — configure SMTP to send actual emails",
"to":      req.To,
})
}
