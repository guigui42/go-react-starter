// Package middleware provides HTTP middleware for the the API.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/example/go-react-starter/pkg/logger"
)

// contextKey is the type for context keys
type contextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey contextKey = "request_id"
	// LoggerKey is the context key for logger
	LoggerKey contextKey = "logger"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// Write captures that a response was written
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// LoggingMiddleware logs HTTP requests with structured logging
func LoggingMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate request ID
			requestID := middleware.GetReqID(r.Context())
			if requestID == "" {
				requestID = fmt.Sprintf("%d", middleware.NextRequestID())
			}

			// Create request logger with OTel trace correlation
			reqLogger := log.RequestLoggerWithTrace(r.Context(), requestID)

			// Add request ID and logger to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, RequestIDKey, requestID)
			ctx = context.WithValue(ctx, LoggerKey, reqLogger)
			r = r.WithContext(ctx)

			// Wrap response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				written:        false,
			}

			// Set request ID header
			w.Header().Set("X-Request-ID", requestID)

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Log request completion
			duration := time.Since(start)
			log.WithRequest(
				requestID,
				r.Method,
				r.URL.Path,
				wrapped.statusCode,
				duration,
			)
		})
	}
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GetLogger retrieves the logger from the context
func GetLogger(ctx context.Context) *logger.RequestLogger {
	if log, ok := ctx.Value(LoggerKey).(*logger.RequestLogger); ok {
		return log
	}
	return nil
}
