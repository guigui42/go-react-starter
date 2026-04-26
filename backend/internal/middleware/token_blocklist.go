package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/example/go-react-starter/internal/models"
	"github.com/example/go-react-starter/pkg/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TokenBlocklist provides JWT token revocation by tracking blocked JTI (JWT ID) claims.
// Uses an in-memory cache backed by database persistence. Expired entries are
// automatically cleaned up every 10 minutes from both memory and database.
type TokenBlocklist struct {
	mu      sync.RWMutex
	entries map[string]time.Time // jti -> expiration time (in-memory cache)
	db      *gorm.DB
	log     *logger.Logger
	done    chan struct{}
}

// NewTokenBlocklist creates a new token blocklist backed by the given database.
// It loads existing non-expired blocked tokens from the database into memory
// and starts background cleanup of expired entries.
// If db is nil, the blocklist operates in memory-only mode (useful for tests).
func NewTokenBlocklist(db *gorm.DB, opts ...func(*TokenBlocklist)) *TokenBlocklist {
	bl := &TokenBlocklist{
		entries: make(map[string]time.Time),
		db:      db,
		done:    make(chan struct{}),
	}
	for _, opt := range opts {
		opt(bl)
	}
	bl.loadFromDB()
	go bl.cleanup()
	return bl
}

// WithLogger sets the logger for the token blocklist.
func WithLogger(log *logger.Logger) func(*TokenBlocklist) {
	return func(bl *TokenBlocklist) {
		bl.log = log
	}
}

// Block adds a JTI to the blocklist. It persists to the database and caches in memory.
func (bl *TokenBlocklist) Block(jti string, expiresAt time.Time) {
	bl.mu.Lock()
	bl.entries[jti] = expiresAt
	bl.mu.Unlock()

	// Persist to database (best-effort; memory cache ensures immediate blocking)
	if bl.db != nil {
		if err := bl.db.WithContext(context.Background()).Clauses(clause.OnConflict{DoNothing: true}).Create(&models.BlockedToken{
			JTI:       jti,
			ExpiresAt: expiresAt,
			CreatedAt: time.Now(),
		}).Error; err != nil {
			if bl.log != nil {
				bl.log.Error("Failed to persist blocked token to database", err)
			}
		}
	}
}

// IsBlocked returns true if the given JTI has been revoked.
// Checks in-memory cache first for performance, falls back to database.
func (bl *TokenBlocklist) IsBlocked(jti string) bool {
	// Fast path: check in-memory cache
	bl.mu.RLock()
	_, blocked := bl.entries[jti]
	bl.mu.RUnlock()
	if blocked {
		return true
	}

	// Slow path: check database (handles server restart scenario)
	if bl.db != nil {
		var token models.BlockedToken
		if err := bl.db.WithContext(context.Background()).Where("jti = ? AND expires_at > ?", jti, time.Now()).First(&token).Error; err == nil {
			// Cache the hit for future fast-path lookups
			bl.mu.Lock()
			bl.entries[jti] = token.ExpiresAt
			bl.mu.Unlock()
			return true
		}
	}

	return false
}

// Stop terminates the background cleanup goroutine.
func (bl *TokenBlocklist) Stop() {
	close(bl.done)
}

// loadFromDB loads all non-expired blocked tokens from the database into memory.
func (bl *TokenBlocklist) loadFromDB() {
	if bl.db == nil {
		return
	}

	var tokens []models.BlockedToken
	if err := bl.db.WithContext(context.Background()).Where("expires_at > ?", time.Now()).Find(&tokens).Error; err != nil {
		if bl.log != nil {
			bl.log.Error("Failed to load blocked tokens from database", err)
		}
		return
	}

	bl.mu.Lock()
	for _, t := range tokens {
		bl.entries[t.JTI] = t.ExpiresAt
	}
	bl.mu.Unlock()
}

// cleanup periodically removes expired entries from both memory and database.
func (bl *TokenBlocklist) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-bl.done:
			return
		case now := <-ticker.C:
			bl.mu.Lock()
			for jti, exp := range bl.entries {
				if now.After(exp) {
					delete(bl.entries, jti)
				}
			}
			bl.mu.Unlock()

			if bl.db != nil {
				if err := bl.db.WithContext(context.Background()).Where("expires_at <= ?", now).Delete(&models.BlockedToken{}).Error; err != nil {
					if bl.log != nil {
						bl.log.Error("Failed to cleanup expired blocked tokens from database", err)
					}
				}
			}
		}
	}
}
