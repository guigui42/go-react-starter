// Package middleware provides HTTP middleware for the the API.
package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/guigui42/go-react-starter/pkg/response"
)

const (
	// CSRFTokenCookieName is the name of the cookie used to store the CSRF token
	CSRFTokenCookieName = "csrf_token"
	// CSRFTokenHeaderName is the name of the HTTP header used to send the CSRF token
	CSRFTokenHeaderName = "X-CSRF-Token"
	// CSRFTokenLength is the length of the generated CSRF token in bytes (before base64 encoding)
	CSRFTokenLength = 32
)

// generateCSRFToken generates a cryptographically secure random token
func generateCSRFToken() (string, error) {
	bytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// SetCSRFCookie sets a new CSRF token cookie on the response.
// The cookie is readable by JavaScript (HttpOnly: false) so the frontend can include it in headers.
// It uses SameSite=Strict for additional protection.
// The secure parameter controls the Secure flag on the cookie (should be true in production).
func SetCSRFCookie(w http.ResponseWriter, secure bool) error {
	token, err := generateCSRFToken()
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     CSRFTokenCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false, // Must be readable by JavaScript to send in header
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})

	return nil
}

// NewCSRFMiddleware creates a CSRF protection middleware using the double-submit cookie pattern.
// The secure parameter controls the Secure flag on rotated CSRF cookies.
// Safe methods (GET, HEAD, OPTIONS) pass through without modification.
// Unsafe methods (POST, PUT, DELETE, PATCH) validate that the token
// in the X-CSRF-Token header matches the token in the csrf_token cookie.
// The CSRF cookie is set during authentication (login/register/OAuth) via SetCSRFCookie.
func NewCSRFMiddleware(secure bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Safe methods: pass through without CSRF validation.
			// Re-issue the CSRF cookie if it's missing (e.g., after browser restart:
			// the auth cookie persists via MaxAge but the CSRF session cookie is lost).
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				if _, err := r.Cookie(CSRFTokenCookieName); err != nil {
					_ = SetCSRFCookie(w, secure)
				}
				next.ServeHTTP(w, r)
				return
			}

			// Unsafe methods (POST, PUT, DELETE, PATCH): validate CSRF token
			cookie, err := r.Cookie(CSRFTokenCookieName)
			if err != nil || cookie.Value == "" {
				response.Forbidden(w, "CSRF token cookie missing")
				return
			}

			headerToken := r.Header.Get(CSRFTokenHeaderName)
			if headerToken == "" {
				response.Forbidden(w, "CSRF token header missing")
				return
			}

			// Validate that cookie and header tokens match (double-submit pattern)
			if cookie.Value != headerToken {
				response.Forbidden(w, "CSRF token mismatch")
				return
			}

			// Token is valid — no per-request rotation to avoid race conditions
			// with concurrent SPA requests. Token is rotated on auth events
			// (login/register/OAuth) and SameSite=Strict provides protection.
			next.ServeHTTP(w, r)
		})
	}
}
