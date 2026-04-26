// Package response provides standardized HTTP response helpers for the the API.
// It includes consistent error and success response structures that match the OpenAPI specification.
package response

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents the standardized API error structure.
// It matches the Error schema defined in the OpenAPI specification (contracts/api.yaml).
type ErrorResponse struct {
	Code    string                 `json:"code"`              // Error code (e.g., "VALIDATION_ERROR", "UNAUTHORIZED")
	Message string                 `json:"message"`           // Human-readable error message
	Details map[string]interface{} `json:"details,omitempty"` // Additional error details (optional)
}

// SuccessResponse represents the standardized API success response structure.
type SuccessResponse struct {
	Data interface{} `json:"data"` // Response data
}

// Error sends a standardized error response with the specified status code, error code, message, and optional details.
// It sets the Content-Type header to application/json and writes the response to the http.ResponseWriter.
//
// Example:
//
//	response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid input", map[string]interface{}{
//	    "field": "email",
//	    "reason": "invalid format",
//	})
func Error(w http.ResponseWriter, statusCode int, code string, message string, details map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	})
}

// Success sends a standardized success response with the specified status code and data.
// It sets the Content-Type header to application/json and writes the response to the http.ResponseWriter.
//
// Example:
//
//	response.Success(w, http.StatusOK, map[string]interface{}{
//	    "id": 123,
//	    "name": "Trade",
//	})
func Success(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(SuccessResponse{
		Data: data,
	})
}

// BadRequest sends a 400 Bad Request error response.
// Use this when the client sends invalid request data.
//
// Example:
//
//	response.BadRequest(w, "Invalid request body", map[string]interface{}{
//	    "error": "JSON parse error",
//	})
func BadRequest(w http.ResponseWriter, message string, details map[string]interface{}) {
	Error(w, http.StatusBadRequest, "BAD_REQUEST", message, details)
}

// Unauthorized sends a 401 Unauthorized error response.
// Use this when authentication is required but not provided or invalid.
//
// Example:
//
//	response.Unauthorized(w, "Authentication required")
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

// Forbidden sends a 403 Forbidden error response.
// Use this when the authenticated user lacks sufficient permissions.
//
// Example:
//
//	response.Forbidden(w, "Insufficient permissions to access this resource")
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, "FORBIDDEN", message, nil)
}

// NotFound sends a 404 Not Found error response.
// Use this when the requested resource does not exist.
//
// Example:
//
//	response.NotFound(w, "Trade not found")
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, "NOT_FOUND", message, nil)
}

// InternalServerError sends a 500 Internal Server Error response.
// Use this for unexpected server errors.
//
// Example:
//
//	response.InternalServerError(w, "An unexpected error occurred")
func InternalServerError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message, nil)
}

// ValidationError sends a 400 Bad Request with validation error details.
// Use this when field validation fails.
//
// Example:
//
//	response.ValidationError(w, map[string]interface{}{
//	    "email": "Invalid email format",
//	    "password": "Must be at least 8 characters",
//	})
func ValidationError(w http.ResponseWriter, errors map[string]interface{}) {
	Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", errors)
}
