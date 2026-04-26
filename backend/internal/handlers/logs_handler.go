package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/example/go-react-starter/pkg/logger"
	"github.com/example/go-react-starter/pkg/response"
)

// Maximum number of log entries that can be returned in a single request
const maxLogEntriesLimit = 500

// Default number of log entries returned if no limit is specified
const defaultLogEntriesLimit = 100

// LogsHandler handles admin log viewing HTTP requests.
type LogsHandler struct {
	buffer *logger.RingBuffer
}

// NewLogsHandler creates a new logs handler with the given ring buffer.
// If buffer is nil, the handler will return empty responses.
func NewLogsHandler(buffer *logger.RingBuffer) *LogsHandler {
	return &LogsHandler{
		buffer: buffer,
	}
}

// LogsResponse represents the response for the logs endpoint
type LogsResponse struct {
	Entries  []logger.LogEntry `json:"entries"`
	Total    int               `json:"total"`
	Capacity int               `json:"capacity"`
}

// GetLogs returns log entries from the ring buffer with optional filtering.
//
// @Summary      Get backend logs
// @Description  Returns log entries from the in-memory buffer with optional filtering by level and time
// @Tags         Admin
// @Produce      json
// @Param        level query string false "Filter by log level (debug, info, warn, error)"
// @Param        since query int false "Unix timestamp to get logs after"
// @Param        limit query int false "Maximum number of entries to return (default: 100, max: 500)"
// @Security     BearerAuth
// @Success      200 {object} LogsResponse "Log entries"
// @Failure      400 {object} response.ErrorResponse "Invalid parameters"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      403 {object} response.ErrorResponse "Admin access required"
// @Router       /api/v1/admin/logs [get]
func (h *LogsHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	if h.buffer == nil {
		response.Success(w, http.StatusOK, LogsResponse{
			Entries:  []logger.LogEntry{},
			Total:    0,
			Capacity: 0,
		})
		return
	}

	// Parse query parameters
	level := r.URL.Query().Get("level")

	// Validate level if provided
	if level != "" {
		validLevels := map[string]bool{
			"debug": true,
			"info":  true,
			"warn":  true,
			"error": true,
		}
		if !validLevels[level] {
			response.BadRequest(w, "Invalid log level. Valid values: debug, info, warn, error", nil)
			return
		}
	}

	// Parse since timestamp
	var since time.Time
	if sinceStr := r.URL.Query().Get("since"); sinceStr != "" {
		sinceUnix, err := strconv.ParseInt(sinceStr, 10, 64)
		if err != nil {
			response.BadRequest(w, "Invalid since parameter. Must be a Unix timestamp", nil)
			return
		}
		since = time.Unix(sinceUnix, 0)
	}

	// Parse limit with defaults
	limit := defaultLogEntriesLimit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 1 {
			response.BadRequest(w, "Invalid limit parameter. Must be a positive integer", nil)
			return
		}
		limit = parsedLimit
	}

	// Cap limit at maximum
	if limit > maxLogEntriesLimit {
		limit = maxLogEntriesLimit
	}

	entries := h.buffer.GetEntries(level, since, limit)

	response.Success(w, http.StatusOK, LogsResponse{
		Entries:  entries,
		Total:    h.buffer.Size(),
		Capacity: h.buffer.Capacity(),
	})
}

// ClearLogs clears all entries from the log buffer.
//
// @Summary      Clear backend logs
// @Description  Clears all log entries from the in-memory buffer
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]string "Success message"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      403 {object} response.ErrorResponse "Admin access required"
// @Router       /api/v1/admin/logs [delete]
func (h *LogsHandler) ClearLogs(w http.ResponseWriter, r *http.Request) {
	if h.buffer == nil {
		response.Success(w, http.StatusOK, map[string]string{
			"message": "Log buffer not available",
		})
		return
	}

	h.buffer.Clear()

	response.Success(w, http.StatusOK, map[string]string{
		"message": "Log buffer cleared successfully",
	})
}
