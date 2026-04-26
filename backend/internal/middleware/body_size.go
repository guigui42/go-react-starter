package middleware

import (
	"errors"
	"net/http"

	"github.com/example/go-react-starter/pkg/response"
)

// DefaultMaxBodySize is the default maximum request body size (1 MB).
// Used for JSON API endpoints.
const DefaultMaxBodySize = 1 << 20 // 1 MB

// UploadMaxBodySize is the maximum request body size for file upload endpoints (10 MB).
// Used for CSV import and other multipart/form-data routes.
const UploadMaxBodySize = 10 << 20 // 10 MB

// MaxBodySize returns middleware that limits the size of incoming request bodies.
// When the limit is exceeded, http.MaxBytesReader causes the request body read
// to return an *http.MaxBytesError. Handlers can use IsMaxBytesError to detect
// this and return a consistent JSON error response.
//
// This prevents denial-of-service attacks via memory exhaustion from oversized payloads.
//
// Apply different limits per route group:
//
//	r.Use(middleware.MaxBodySize(middleware.DefaultMaxBodySize))  // 1 MB for JSON
//	r.With(middleware.MaxBodySize(middleware.UploadMaxBodySize))  // 10 MB for uploads
func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// IsMaxBytesError reports whether err is or wraps an *http.MaxBytesError.
// Handlers should use this to detect body-too-large errors and call
// HandleMaxBytesError to return a consistent JSON error response.
func IsMaxBytesError(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}

// HandleMaxBytesError writes a standardized JSON 413 error response for oversized request bodies.
// Returns true if err was a MaxBytesError and was handled, false otherwise.
func HandleMaxBytesError(w http.ResponseWriter, err error) bool {
	if IsMaxBytesError(err) {
		response.Error(w, http.StatusRequestEntityTooLarge, "REQUEST_ENTITY_TOO_LARGE", "Request body too large", nil)
		return true
	}
	return false
}
