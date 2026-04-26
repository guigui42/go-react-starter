package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/guigui42/go-react-starter/internal/models"
	"github.com/guigui42/go-react-starter/pkg/response"
	"gorm.io/gorm"
)

const (
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
	// UserContextKey is the context key for the full user object
	UserContextKey contextKey = "user"
	// AuthTokenCookieName is the name of the cookie used to store the JWT token
	AuthTokenCookieName = "auth_token"
)

// Claims represents JWT token claims
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// extractToken extracts the JWT token from the request.
// It first checks the auth_token cookie, then falls back to the Authorization header.
func extractToken(r *http.Request) string {
	// Check cookie first (preferred method)
	if cookie, err := r.Cookie(AuthTokenCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Fall back to Authorization header for backward compatibility with API clients
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	return ""
}

// ExtractTokenFromRequest extracts the JWT token from the request (exported for use by handlers).
func ExtractTokenFromRequest(r *http.Request) string {
	return extractToken(r)
}

// NewAuthMiddleware creates a middleware that verifies JWT tokens and sets user context.
// The jwtSecret is validated at startup via config.Load() and injected here to avoid runtime env lookups.
// The blocklist parameter is optional; pass nil to skip token revocation checks.
func NewAuthMiddleware(jwtSecret string, blocklist ...*TokenBlocklist) func(http.Handler) http.Handler {
	var bl *TokenBlocklist
	if len(blocklist) > 0 {
		bl = blocklist[0]
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from cookie or header
			tokenString := extractToken(r)
			if tokenString == "" {
				response.Unauthorized(w, "Authorization required")
				return
			}

			// Parse and validate token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Validate signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				if jwtSecret == "" {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				response.Unauthorized(w, "Invalid or expired token")
				return
			}

			// Extract claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				response.Unauthorized(w, "Invalid token claims")
				return
			}

			// Check if token has been revoked via blocklist
			if bl != nil {
				if jti, ok := claims["jti"].(string); ok && bl.IsBlocked(jti) {
					response.Unauthorized(w, "Token has been revoked")
					return
				}
			}

			// Set user ID in context
			userID, ok := claims["user_id"].(string)
			if !ok {
				response.Unauthorized(w, "Invalid user ID in token")
				return
			}

			// Add user ID to request context
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID retrieves user ID from request context.
// It returns the user ID and a boolean indicating whether it was found.
func GetUserID(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	return userID, ok
}

// GetUserFromContext retrieves the full user object from request context.
// It returns the user and a boolean indicating whether it was found.
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	return user, ok
}

// WithUser adds a user to the context
func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

// NewAuthMiddlewareWithUser creates a middleware like NewAuthMiddleware but also loads the full user object.
// The jwtSecret is validated at startup via config.Load() and injected here to avoid runtime env lookups.
// The blocklist parameter is optional; pass nil to skip token revocation checks.
func NewAuthMiddlewareWithUser(db *gorm.DB, jwtSecret string, blocklist ...*TokenBlocklist) func(http.Handler) http.Handler {
	var bl *TokenBlocklist
	if len(blocklist) > 0 {
		bl = blocklist[0]
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from cookie or header
			tokenString := extractToken(r)
			if tokenString == "" {
				response.Unauthorized(w, "Authorization required")
				return
			}

			// Parse and validate token using injected secret
			token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				if jwtSecret == "" {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				response.Unauthorized(w, "Invalid or expired token")
				return
			}

			// Extract claims
			claims, ok := token.Claims.(*Claims)
			if !ok {
				response.Unauthorized(w, "Invalid token claims")
				return
			}

			// Check if token has been revoked via blocklist
			if bl != nil && claims.ID != "" && bl.IsBlocked(claims.ID) {
				response.Unauthorized(w, "Token has been revoked")
				return
			}

			// Load user from database
			var user models.User
			if err := db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
				response.Unauthorized(w, "User not found")
				return
			}

			// Add user to context
			ctx := WithUser(r.Context(), &user)
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID.String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
