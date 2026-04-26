package handlers

import (
"net/http"
"time"

"github.com/guigui42/go-react-starter/pkg/response"
)

// MonitoringHandler handles system monitoring endpoints.
type MonitoringHandler struct{}

// NewMonitoringHandler creates a new monitoring handler.
func NewMonitoringHandler() *MonitoringHandler {
return &MonitoringHandler{}
}

// GetHealthStatus returns the system health status.
func (h *MonitoringHandler) GetHealthStatus(w http.ResponseWriter, r *http.Request) {
response.Success(w, http.StatusOK, map[string]interface{}{
"status":    "ok",
"timestamp": time.Now().UTC().Format(time.RFC3339),
})
}
