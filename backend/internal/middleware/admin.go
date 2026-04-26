package middleware

import (
	"net/http"

	"github.com/example/go-react-starter/internal/models"
	"github.com/example/go-react-starter/pkg/response"
	"gorm.io/gorm"
)

// AdminMiddleware verifies that the authenticated user has admin privileges.
// It must be used after AuthMiddleware which sets the user ID in context.
// Returns 403 Forbidden if the user is not an admin.
func AdminMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from context (set by AuthMiddleware)
			userID, ok := GetUserID(r)
			if !ok || userID == "" {
				response.Unauthorized(w, "Authentication required")
				return
			}

			// Fetch user from database to check admin status
			var user models.User
			if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
				response.Unauthorized(w, "User not found")
				return
			}

			// Check admin status
			if !user.IsAdmin {
				response.Forbidden(w, "Admin access required")
				return
			}

			// User is admin, proceed with request
			next.ServeHTTP(w, r)
		})
	}
}
