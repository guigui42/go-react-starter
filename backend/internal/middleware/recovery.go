package middleware

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/guigui42/go-react-starter/pkg/response"
)

// RecoveryMiddleware recovers from panics and returns a JSON error response.
// This replaces Chi's default Recoverer which returns HTML.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				logger := GetLogger(r.Context())
				if logger != nil {
					logger.Error(
						fmt.Sprintf("Panic recovered: %v", err),
						fmt.Errorf("%v", err),
					)
					logger.Error(
						"Stack trace",
						fmt.Errorf("%s", debug.Stack()),
					)
				} else {
					// Log to stderr as fallback, never expose to client
					fmt.Fprintf(os.Stderr, "Panic recovered (no logger): %v\n%s\n", err, debug.Stack())
				}

				// Return JSON error response
				response.Error(
					w,
					http.StatusInternalServerError,
					"INTERNAL_ERROR",
					"An unexpected error occurred",
					nil,
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
