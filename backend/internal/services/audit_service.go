package services

import (
"context"
"encoding/json"
"fmt"
"net/http"
"strings"
"time"

"github.com/google/uuid"
"github.com/example/go-react-starter/internal/models"
"github.com/rs/zerolog"
"gorm.io/gorm"
)

// AuditService provides structured audit logging for sensitive operations.
type AuditService struct {
db     *gorm.DB
logger zerolog.Logger
}

// NewAuditService creates a new audit service instance.
func NewAuditService(db *gorm.DB, logger zerolog.Logger) *AuditService {
return &AuditService{
db:     db,
logger: logger,
}
}

// AuditEvent represents an audit event to be logged.
type AuditEvent struct {
EventType string
ActorID   *uuid.UUID
TargetID  *uuid.UUID
Action    string
Status    string
IPAddress string
UserAgent string
Metadata  map[string]interface{}
}

// LogAuditEvent logs an audit event to both Zerolog and the database.
func (s *AuditService) LogAuditEvent(ctx context.Context, event AuditEvent) error {
logEvent := s.logger.Info().
Str("audit", event.EventType).
Str("action", event.Action).
Str("status", event.Status)

if event.ActorID != nil {
logEvent = logEvent.Str("actor_id", event.ActorID.String())
}
if event.TargetID != nil {
logEvent = logEvent.Str("target_id", event.TargetID.String())
}
if event.IPAddress != "" {
logEvent = logEvent.Str("ip_address", event.IPAddress)
}
if len(event.Metadata) > 0 {
logEvent = logEvent.Interface("metadata", event.Metadata)
}

logEvent.Msg(event.Action)

var metadataJSON json.RawMessage
if len(event.Metadata) > 0 {
jsonBytes, err := json.Marshal(event.Metadata)
if err != nil {
s.logger.Warn().
Err(err).
Str("event_type", event.EventType).
Msg("Failed to serialize audit metadata")
} else {
metadataJSON = jsonBytes
}
}

auditLog := &models.AuditLog{
EventType: event.EventType,
ActorID:   event.ActorID,
TargetID:  event.TargetID,
Action:    event.Action,
Status:    event.Status,
Metadata:  metadataJSON,
}

if event.IPAddress != "" {
auditLog.IPAddress = &event.IPAddress
}
if event.UserAgent != "" {
auditLog.UserAgent = &event.UserAgent
}

if err := s.db.WithContext(ctx).Create(auditLog).Error; err != nil {
s.logger.Error().
Err(err).
Str("event_type", event.EventType).
Msg("Failed to store audit log in database")
}

return nil
}

// GetClientIP extracts the client IP address from an HTTP request.
func GetClientIP(r *http.Request) string {
if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
ips := strings.Split(xff, ",")
if len(ips) > 0 {
return strings.TrimSpace(ips[0])
}
}

if xri := r.Header.Get("X-Real-IP"); xri != "" {
return xri
}

return r.RemoteAddr
}

// GetUserAgent extracts the User-Agent from an HTTP request.
func GetUserAgent(r *http.Request) string {
return r.Header.Get("User-Agent")
}

// AuditLogQuery represents filter parameters for querying audit logs.
type AuditLogQuery struct {
EventType string
ActorID   *uuid.UUID
Status    string
From      *time.Time
To        *time.Time
Page      int
PageSize  int
}

// AuditLogResponse represents paginated audit log results.
type AuditLogResponse struct {
Logs       []models.AuditLog `json:"logs"`
Total      int64             `json:"total"`
Page       int               `json:"page"`
PageSize   int               `json:"page_size"`
TotalPages int               `json:"total_pages"`
Actors     map[string]string `json:"actors"`
}

// QueryAuditLogs retrieves audit logs with filtering and pagination.
func (s *AuditService) QueryAuditLogs(ctx context.Context, query AuditLogQuery) (*AuditLogResponse, error) {
if query.Page < 1 {
query.Page = 1
}
if query.PageSize < 1 {
query.PageSize = 50
} else if query.PageSize > 100 {
query.PageSize = 100
}

db := s.db.WithContext(ctx).Model(&models.AuditLog{})

if query.EventType != "" {
db = db.Where("event_type LIKE ?", query.EventType+"%")
}
if query.ActorID != nil {
db = db.Where("actor_id = ?", *query.ActorID)
}
if query.Status != "" {
db = db.Where("status = ?", query.Status)
}
if query.From != nil {
db = db.Where("created_at >= ?", *query.From)
}
if query.To != nil {
db = db.Where("created_at <= ?", *query.To)
}

var total int64
if err := db.Count(&total).Error; err != nil {
return nil, fmt.Errorf("count audit logs: %w", err)
}

var logs []models.AuditLog
offset := (query.Page - 1) * query.PageSize
if err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&logs).Error; err != nil {
return nil, fmt.Errorf("query audit logs: %w", err)
}

totalPages := int(total) / query.PageSize
if int(total)%query.PageSize > 0 {
totalPages++
}

actors := make(map[string]string)
var actorIDs []uuid.UUID
seen := make(map[uuid.UUID]bool)
for _, log := range logs {
if log.ActorID != nil && !seen[*log.ActorID] {
seen[*log.ActorID] = true
actorIDs = append(actorIDs, *log.ActorID)
}
}
if len(actorIDs) > 0 {
var users []struct {
ID    uuid.UUID
Email string
}
if err := s.db.WithContext(ctx).Table("users").Select("id, email").Where("id IN ?", actorIDs).Find(&users).Error; err != nil {
s.logger.Warn().Err(err).Msg("Failed to look up actor emails for audit log response")
} else {
for _, u := range users {
actors[u.ID.String()] = u.Email
}
}
}

return &AuditLogResponse{
Logs:       logs,
Total:      total,
Page:       query.Page,
PageSize:   query.PageSize,
TotalPages: totalPages,
Actors:     actors,
}, nil
}
