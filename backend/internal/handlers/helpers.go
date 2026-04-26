package handlers

import (
"net/http"
"strconv"

"github.com/guigui42/go-react-starter/internal/middleware"
)

// getUserIDFromContext extracts the user ID from the request context.
func getUserIDFromContext(r *http.Request) (string, bool) {
userID, ok := r.Context().Value(middleware.UserIDKey).(string)
return userID, ok
}

// parsePagination extracts limit and offset from query parameters.
func parsePagination(r *http.Request) (int, int) {
limit := 50
offset := 0

if l := r.URL.Query().Get("limit"); l != "" {
if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
limit = parsed
}
}

if o := r.URL.Query().Get("offset"); o != "" {
if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
offset = parsed
}
}

return limit, offset
}
