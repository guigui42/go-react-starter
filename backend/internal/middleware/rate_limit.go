// Package middleware provides HTTP middleware for the the API.
package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/guigui42/go-react-starter/pkg/logger"
	"github.com/guigui42/go-react-starter/pkg/response"
	"golang.org/x/time/rate"
)

// RateLimiter implements IP-based rate limiting for HTTP handlers.
// It uses a token bucket algorithm (via golang.org/x/time/rate) to control
// the rate of requests per IP address.
//
// Note: IP-based rate limiting for auth endpoints is now handled by Nginx
// (see docker/nginx.conf). This middleware is kept for:
// - Local development (without Nginx)
// - Future use cases requiring custom rate limiting logic
type RateLimiter struct {
	visitors   map[string]*visitor
	mu         sync.Mutex
	rate       rate.Limit
	burst      int
	log        *logger.Logger
	trustProxy bool
}

// visitor represents a single IP address with its rate limiter and last access time.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new RateLimiter with the specified rate limit and burst size.
// The rate is expressed as requests per second, and burst is the maximum number
// of requests allowed in a burst. When trustProxy is true, X-Forwarded-For and
// X-Real-IP headers are used to determine the client IP; otherwise only RemoteAddr is used.
//
// Example:
//
//	// Allow 5 requests per minute with burst of 10
//	limiter := NewRateLimiter(rate.Limit(5.0/60.0), 10, log, false)
func NewRateLimiter(r rate.Limit, burst int, log *logger.Logger, trustProxy bool) *RateLimiter {
	rl := &RateLimiter{
		visitors:   make(map[string]*visitor),
		rate:       r,
		burst:      burst,
		log:        log,
		trustProxy: trustProxy,
	}
	go rl.cleanupVisitors()
	return rl
}

// getVisitor returns the rate limiter for the given IP address.
// If the IP is not in the map, a new limiter is created.
func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}
	v.lastSeen = time.Now()
	return v.limiter
}

// cleanupVisitors removes visitors that haven't been seen for more than 3 minutes.
// This runs in a background goroutine to prevent memory leaks.
func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// getClientIP extracts the client IP address from the request.
// When trustProxy is true, it checks X-Forwarded-For and X-Real-IP headers
// (for proxied requests). Otherwise, it only uses RemoteAddr to prevent
// attackers from spoofing their IP via headers to bypass rate limiting.
func getClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		// Check X-Forwarded-For header (may contain multiple IPs)
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			// X-Forwarded-For typically contains comma-separated IPs like "client, proxy1, proxy2"
			// Take the first IP (original client)
			parts := strings.Split(xff, ",")
			if len(parts) > 0 {
				ip := strings.TrimSpace(parts[0])
				if ip != "" {
					return ip
				}
			}
		}

		// Check X-Real-IP header
		if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
			return xrip
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// Limit is a middleware that rate limits requests based on client IP.
// Returns 429 Too Many Requests when the rate limit is exceeded.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r, rl.trustProxy)
		limiter := rl.getVisitor(ip)

		if !limiter.Allow() {
			// Log rate limit violation
			if rl.log != nil {
				rl.log.Warn("Rate limit exceeded for IP: " + ip)
			}
			w.Header().Set("Retry-After", "60")
			response.Error(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests. Please try again later.", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// VisitorCount returns the number of tracked visitors.
// Useful for testing and monitoring.
func (rl *RateLimiter) VisitorCount() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.visitors)
}

// UserRateLimiter implements per-user rate limiting for HTTP handlers.
// Unlike RateLimiter which uses IP addresses, this uses the authenticated user ID
// from the request context (set by AuthMiddleware).
//
// This is used for endpoints where the limit should apply per user account,
// not per IP address (e.g., GDPR data export which should be 1 request/hour per user).
type UserRateLimiter struct {
	users      map[string]*visitor
	mu         sync.Mutex
	rate       rate.Limit
	burst      int
	log        *logger.Logger
	trustProxy bool
}

// NewUserRateLimiter creates a new UserRateLimiter with the specified rate limit and burst size.
// The rate is expressed as requests per second, and burst is the maximum number
// of requests allowed in a burst. When trustProxy is true, X-Forwarded-For and
// X-Real-IP headers are used for IP fallback; otherwise only RemoteAddr is used.
//
// Example:
//
//	// Allow 1 request per hour with burst of 1 (for GDPR export)
//	limiter := NewUserRateLimiter(rate.Limit(1.0/3600.0), 1, log, false)
func NewUserRateLimiter(r rate.Limit, burst int, log *logger.Logger, trustProxy bool) *UserRateLimiter {
	rl := &UserRateLimiter{
		users:      make(map[string]*visitor),
		rate:       r,
		burst:      burst,
		log:        log,
		trustProxy: trustProxy,
	}
	go rl.cleanupUsers()
	return rl
}

// getUser returns the rate limiter for the given user ID.
// If the user is not in the map, a new limiter is created.
func (rl *UserRateLimiter) getUser(userID string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.users[userID]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.users[userID] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}
	v.lastSeen = time.Now()
	return v.limiter
}

// cleanupUsers removes users that haven't been seen for more than the rate period + buffer.
// For a 1 request/hour limit, this cleans up after 2 hours of inactivity.
// This runs in a background goroutine to prevent memory leaks.
func (rl *UserRateLimiter) cleanupUsers() {
	// Calculate cleanup threshold based on rate
	// For 1 request/hour (1/3600), cleanup after 2 hours
	cleanupThreshold := time.Duration(float64(2*time.Second) / float64(rl.rate))
	if cleanupThreshold < 5*time.Minute {
		cleanupThreshold = 5 * time.Minute
	}
	if cleanupThreshold > 4*time.Hour {
		cleanupThreshold = 4 * time.Hour
	}

	for {
		time.Sleep(cleanupThreshold / 2)
		rl.mu.Lock()
		for userID, v := range rl.users {
			if time.Since(v.lastSeen) > cleanupThreshold {
				delete(rl.users, userID)
			}
		}
		rl.mu.Unlock()
	}
}

// Limit is a middleware that rate limits requests based on authenticated user ID.
// It expects the user ID to be set in context by AuthMiddleware.
// Returns 429 Too Many Requests when the rate limit is exceeded.
func (rl *UserRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context (set by AuthMiddleware)
		userID, ok := GetUserID(r)
		if !ok {
			// If no user ID in context, fall back to IP-based limiting
			// This shouldn't happen for authenticated endpoints
			userID = getClientIP(r, rl.trustProxy)
			if rl.log != nil {
				rl.log.Warn("UserRateLimiter: no user ID in context, falling back to IP: " + userID)
			}
		}

		limiter := rl.getUser(userID)

		if !limiter.Allow() {
			// Log rate limit violation
			if rl.log != nil {
				rl.log.Warn("Rate limit exceeded for user: " + userID)
			}
			// Calculate retry-after based on rate (in seconds)
			retryAfter := int(1.0 / float64(rl.rate))
			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
			response.Error(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests. Please try again later.", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// UserCount returns the number of tracked users.
// Useful for testing and monitoring.
func (rl *UserRateLimiter) UserCount() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.users)
}
