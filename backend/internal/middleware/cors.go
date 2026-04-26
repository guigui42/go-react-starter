package middleware

import (
	"net/http"
	"os"
	"strings"
)

// CORSConfig holds the configuration for the CORS middleware
type CORSConfig struct {
	// AllowedOrigins is a list of origins that are allowed to make cross-origin requests.
	// Use "*" to allow all origins (not recommended for production).
	AllowedOrigins []string

	// AllowedMethods is a list of methods that are allowed for cross-origin requests.
	AllowedMethods []string

	// AllowedHeaders is a list of headers that are allowed in cross-origin requests.
	AllowedHeaders []string

	// AllowCredentials indicates whether the response to the request can be exposed when credentials flag is true.
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a preflight request can be cached.
	MaxAge string
}

// DefaultCORSConfig returns the default CORS configuration for local development
// Uses explicit localhost origins (wildcard with credentials is rejected by browsers)
func DefaultCORSConfig() CORSConfig {
	// Read allowed origins from environment or use secure localhost defaults
	originsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	var allowedOrigins []string
	
	if originsEnv != "" {
		// Split by comma and trim whitespace
		parts := strings.Split(originsEnv, ",")
		allowedOrigins = make([]string, 0, len(parts))
		for _, origin := range parts {
			trimmed := strings.TrimSpace(origin)
			if trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
		// Fall back to defaults if env var is empty or contains only whitespace/commas
		if len(allowedOrigins) == 0 {
			allowedOrigins = []string{
				"http://localhost:5173", // Vite frontend default
				"http://localhost:5174", // Alternative Vite port
				"http://localhost:8080", // Backend API
			}
		}
	} else {
		// Secure defaults for local development
		allowedOrigins = []string{
			"http://localhost:5173", // Vite frontend default
			"http://localhost:5174", // Alternative Vite port
			"http://localhost:8080", // Backend API
		}
	}

	return CORSConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
			"PATCH",
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-Requested-With",
			"X-CSRF-Token",
		},
		AllowCredentials: true,
		MaxAge:           "86400", // 24 hours
	}
}

// CORSMiddleware returns a middleware handler that adds CORS headers to responses.
// This middleware is intended for local development when not running behind nginx.
func CORSMiddleware(config CORSConfig) func(http.Handler) http.Handler {
	allowAll := len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*"

	allowedOriginsMap := make(map[string]bool)
	if !allowAll {
		for _, origin := range config.AllowedOrigins {
			allowedOriginsMap[origin] = true
		}
	}

	allowedMethodsStr := strings.Join(config.AllowedMethods, ", ")
	allowedHeadersStr := strings.Join(config.AllowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if the origin is allowed (allow all or specific origin)
			if origin != "" && (allowAll || allowedOriginsMap[origin]) {
				// When allowing credentials, we must echo back the specific origin, not "*"
				w.Header().Set("Access-Control-Allow-Origin", origin)

				if config.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				// Handle preflight requests
				if r.Method == http.MethodOptions {
					w.Header().Set("Access-Control-Allow-Methods", allowedMethodsStr)
					w.Header().Set("Access-Control-Allow-Headers", allowedHeadersStr)
					w.Header().Set("Access-Control-Max-Age", config.MaxAge)
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

