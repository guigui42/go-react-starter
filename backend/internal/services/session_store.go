package services

import (
"fmt"
"sync"
"time"

"github.com/go-webauthn/webauthn/webauthn"
"github.com/google/uuid"
)

// sessionEntry holds a WebAuthn session with expiration time
type sessionEntry struct {
data      *webauthn.SessionData
expiresAt time.Time
}

// SessionStore provides in-memory storage for WebAuthn challenge sessions
// with automatic expiration and cleanup
type SessionStore struct {
sessions      map[uuid.UUID]*sessionEntry
mu            sync.RWMutex
ttl           time.Duration
cleanupTicker *time.Ticker
stopCleanup   chan bool
stopOnce      sync.Once
}

// NewSessionStore creates a new session store with default 5-minute TTL
func NewSessionStore() *SessionStore {
return NewSessionStoreWithTTL(5 * time.Minute)
}

// NewSessionStoreWithTTL creates a new session store with custom TTL
func NewSessionStoreWithTTL(ttl time.Duration) *SessionStore {
store := &SessionStore{
sessions:    make(map[uuid.UUID]*sessionEntry),
ttl:         ttl,
stopCleanup: make(chan bool),
}

store.cleanupTicker = time.NewTicker(1 * time.Minute)
go store.automaticCleanup()

return store
}

// SaveSession stores a WebAuthn session for the given user
func (s *SessionStore) SaveSession(userID uuid.UUID, sessionData *webauthn.SessionData) error {
if sessionData == nil {
return fmt.Errorf("sessionData cannot be nil")
}

s.mu.Lock()
defer s.mu.Unlock()

s.sessions[userID] = &sessionEntry{
data:      sessionData,
expiresAt: time.Now().Add(s.ttl),
}

return nil
}

// GetSession retrieves a WebAuthn session for the given user
func (s *SessionStore) GetSession(userID uuid.UUID) (*webauthn.SessionData, error) {
s.mu.RLock()
defer s.mu.RUnlock()

entry, exists := s.sessions[userID]
if !exists {
return nil, fmt.Errorf("session not found for user %s", userID)
}

if time.Now().After(entry.expiresAt) {
return nil, fmt.Errorf("session not found for user %s", userID)
}

return entry.data, nil
}

// DeleteSession removes a WebAuthn session for the given user
func (s *SessionStore) DeleteSession(userID uuid.UUID) error {
s.mu.Lock()
defer s.mu.Unlock()

delete(s.sessions, userID)
return nil
}

// Cleanup removes all expired sessions from the store
func (s *SessionStore) Cleanup() {
s.mu.Lock()
defer s.mu.Unlock()

now := time.Now()
for userID, entry := range s.sessions {
if now.After(entry.expiresAt) {
delete(s.sessions, userID)
}
}
}

func (s *SessionStore) automaticCleanup() {
for {
select {
case <-s.cleanupTicker.C:
s.Cleanup()
case <-s.stopCleanup:
s.cleanupTicker.Stop()
return
}
}
}

// Stop gracefully stops the session store's cleanup goroutine.
func (s *SessionStore) Stop() {
s.stopOnce.Do(func() {
close(s.stopCleanup)
})
}

// Size returns the current number of sessions (including expired ones)
func (s *SessionStore) Size() int {
s.mu.RLock()
defer s.mu.RUnlock()
return len(s.sessions)
}

// Set stores a session with a string key
func (s *SessionStore) Set(key string, sessionData webauthn.SessionData) {
id, err := uuid.Parse(key)
if err != nil {
id = uuid.NewSHA1(uuid.NameSpaceURL, []byte(key))
}

s.mu.Lock()
defer s.mu.Unlock()

s.sessions[id] = &sessionEntry{
data:      &sessionData,
expiresAt: time.Now().Add(s.ttl),
}
}

// Get retrieves a session by string key
func (s *SessionStore) Get(key string) (webauthn.SessionData, bool) {
id, err := uuid.Parse(key)
if err != nil {
id = uuid.NewSHA1(uuid.NameSpaceURL, []byte(key))
}

s.mu.RLock()
defer s.mu.RUnlock()

entry, exists := s.sessions[id]
if !exists {
return webauthn.SessionData{}, false
}

if time.Now().After(entry.expiresAt) {
return webauthn.SessionData{}, false
}

return *entry.data, true
}

// Delete removes a session by string key
func (s *SessionStore) Delete(key string) {
id, err := uuid.Parse(key)
if err != nil {
id = uuid.NewSHA1(uuid.NameSpaceURL, []byte(key))
}

s.mu.Lock()
defer s.mu.Unlock()

delete(s.sessions, id)
}
