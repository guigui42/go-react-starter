package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/example/go-react-starter/internal/middleware"
	"github.com/example/go-react-starter/internal/models"
	"github.com/example/go-react-starter/pkg/logger"
	_ "github.com/example/go-react-starter/pkg/response" // imported for swagger type resolution
	"github.com/rs/zerolog"
)

// CSPViolationRepository defines the interface for CSP violation storage.
type CSPViolationRepository interface {
	Create(ctx context.Context, violation *models.CSPViolation) error
	GetSummary(ctx context.Context) (*models.CSPViolationSummary, error)
}

// CSPHandler handles Content Security Policy violation reports.
type CSPHandler struct {
	log  *logger.Logger
	repo CSPViolationRepository
}

// NewCSPHandler creates a new CSP handler with the given logger.
// If repo is nil, violations are only logged (not persisted to database).
func NewCSPHandler(log *logger.Logger, repo CSPViolationRepository) *CSPHandler {
	return &CSPHandler{
		log:  log,
		repo: repo,
	}
}

// CSPReport represents the structure of a CSP violation report.
// We only extract the key fields to avoid logging excessive data.
type CSPReport struct {
	CSPReport struct {
		DocumentURI        string `json:"document-uri"`
		ViolatedDirective  string `json:"violated-directive"`
		BlockedURI         string `json:"blocked-uri"`
		EffectiveDirective string `json:"effective-directive"`
	} `json:"csp-report"`
}

// ReportViolation handles CSP violation reports from the browser.
// The browser sends a JSON report when a CSP policy is violated.
// This endpoint logs the violation and persists it to the database for monitoring.
//
// @Summary      Report CSP violation
// @Description  Receives Content Security Policy violation reports from browsers
// @Tags         Security
// @Accept       json
// @Success      204 "No content"
// @Router       /api/csp-report [post]
func (h *CSPHandler) ReportViolation(w http.ResponseWriter, r *http.Request) {
	// Limit body size to prevent memory exhaustion (10KB is generous for CSP reports)
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024)

	var report CSPReport

	// Decode the CSP violation report
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		// Get request-scoped logger if available, fallback to base logger
		if reqLogger := middleware.GetLogger(r.Context()); reqLogger != nil {
			reqLogger.WithMetadata(zerolog.WarnLevel, "Failed to decode CSP report", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			h.log.WithMetadata(zerolog.WarnLevel, "Failed to decode CSP report", map[string]interface{}{
				"error": err.Error(),
			})
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Log only key CSP violation fields to avoid excessive log sizes
	logData := map[string]interface{}{
		"document_uri":        report.CSPReport.DocumentURI,
		"violated_directive":  report.CSPReport.ViolatedDirective,
		"blocked_uri":         report.CSPReport.BlockedURI,
		"effective_directive": report.CSPReport.EffectiveDirective,
		"user_agent":          r.UserAgent(),
		"remote_addr":         r.RemoteAddr,
	}

	// Use request-scoped logger if available, fallback to base logger
	if reqLogger := middleware.GetLogger(r.Context()); reqLogger != nil {
		reqLogger.WithMetadata(zerolog.WarnLevel, "CSP violation reported", logData)
	} else {
		h.log.WithMetadata(zerolog.WarnLevel, "CSP violation reported", logData)
	}

	// Persist to database if repository is available
	if h.repo != nil {
		violation := &models.CSPViolation{
			DocumentURI:        report.CSPReport.DocumentURI,
			ViolatedDirective:  report.CSPReport.ViolatedDirective,
			BlockedURI:         report.CSPReport.BlockedURI,
			EffectiveDirective: report.CSPReport.EffectiveDirective,
			UserAgent:          r.UserAgent(),
			RemoteAddr:         r.RemoteAddr,
		}

		if err := h.repo.Create(r.Context(), violation); err != nil {
			// Log error but don't fail the request
			if reqLogger := middleware.GetLogger(r.Context()); reqLogger != nil {
				reqLogger.WithMetadata(zerolog.ErrorLevel, "Failed to persist CSP violation", map[string]interface{}{
					"error": err.Error(),
				})
			} else {
				h.log.WithMetadata(zerolog.ErrorLevel, "Failed to persist CSP violation", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
	}

	// Return 204 No Content as per CSP reporting spec
	w.WriteHeader(http.StatusNoContent)
}

// GetSummary returns an aggregated summary of CSP violations.
// This is an admin-only endpoint.
//
// @Summary      Get CSP violation summary
// @Description  Returns aggregated statistics about CSP violations
// @Tags         Admin
// @Produce      json
// @Success      200 {object} models.CSPViolationSummary
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/admin/csp-violations [get]
// @Security     BearerAuth
func (h *CSPHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	if h.repo == nil {
		http.Error(w, "CSP violation tracking not enabled", http.StatusServiceUnavailable)
		return
	}

	summary, err := h.repo.GetSummary(r.Context())
	if err != nil {
		h.log.Error("Failed to get CSP violation summary", err)
		http.Error(w, "Failed to get CSP violation summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(summary); err != nil {
		h.log.Error("Failed to encode CSP violation summary", err)
	}
}
