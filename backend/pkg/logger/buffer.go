// Package logger provides structured logging utilities for the the API.
// This file provides a thread-safe ring buffer for capturing log entries.
package logger

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"time"
)

// DefaultBufferSize is the default number of log entries to keep in memory
const DefaultBufferSize = 1000

// MaxBufferSize is the maximum allowed buffer size
const MaxBufferSize = 10000

// LogEntry represents a single log entry captured in the buffer
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// RingBuffer is a thread-safe circular buffer for storing log entries.
// It implements io.Writer to capture zerolog JSON output.
type RingBuffer struct {
	mu      sync.RWMutex
	entries []LogEntry
	head    int  // Next write position
	size    int  // Current number of entries
	cap     int  // Maximum capacity
	full    bool // Whether the buffer has wrapped
}

// NewRingBuffer creates a new ring buffer with the specified capacity.
// If capacity is 0 or negative, it uses DefaultBufferSize.
// If capacity exceeds MaxBufferSize, it caps at MaxBufferSize.
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = DefaultBufferSize
	}
	if capacity > MaxBufferSize {
		capacity = MaxBufferSize
	}
	return &RingBuffer{
		entries: make([]LogEntry, capacity),
		cap:     capacity,
	}
}

// NewRingBufferFromEnv creates a ring buffer with capacity from LOG_BUFFER_SIZE env var
func NewRingBufferFromEnv() *RingBuffer {
	capacity := DefaultBufferSize
	if envSize := os.Getenv("LOG_BUFFER_SIZE"); envSize != "" {
		if parsed, err := strconv.Atoi(envSize); err == nil && parsed > 0 {
			capacity = parsed
		}
	}
	return NewRingBuffer(capacity)
}

// Write implements io.Writer interface to capture zerolog JSON output
func (rb *RingBuffer) Write(p []byte) (n int, err error) {
	n = len(p)
	if n == 0 {
		return 0, nil
	}

	// Parse the JSON log entry
	var raw map[string]interface{}
	if err := json.Unmarshal(p, &raw); err != nil {
		// If it's not valid JSON, store it as a raw message
		entry := LogEntry{
			Timestamp: time.Now(),
			Level:     "unknown",
			Message:   string(p),
		}
		rb.append(entry)
		return n, nil
	}

	// Extract standard zerolog fields
	entry := LogEntry{
		Timestamp: time.Now(),
		Fields:    make(map[string]interface{}),
	}

	// Extract timestamp
	if ts, ok := raw["time"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
			entry.Timestamp = parsed
		}
		delete(raw, "time")
	}

	// Extract level
	if level, ok := raw["level"].(string); ok {
		entry.Level = level
		delete(raw, "level")
	}

	// Extract message
	if msg, ok := raw["message"].(string); ok {
		entry.Message = msg
		delete(raw, "message")
	}

	// Store remaining fields (after sanitization)
	for k, v := range raw {
		// Sanitize string values for sensitive data
		if strVal, ok := v.(string); ok {
			entry.Fields[k] = Sanitize(strVal)
		} else {
			entry.Fields[k] = v
		}
	}

	rb.append(entry)
	return n, nil
}

// append adds an entry to the buffer (must be called with lock held externally or handles its own locking)
func (rb *RingBuffer) append(entry LogEntry) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.entries[rb.head] = entry
	rb.head = (rb.head + 1) % rb.cap

	if rb.full {
		// Buffer was already full, size stays at cap
	} else if rb.size < rb.cap {
		rb.size++
		if rb.size == rb.cap {
			rb.full = true
		}
	}
}

// GetEntries returns log entries, optionally filtered by level and/or since timestamp.
// Results are returned in chronological order (oldest first).
func (rb *RingBuffer) GetEntries(level string, since time.Time, limit int) []LogEntry {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.size == 0 {
		return []LogEntry{}
	}

	// Calculate start position (oldest entry)
	start := 0
	if rb.full {
		start = rb.head
	}

	// Collect entries in chronological order
	result := make([]LogEntry, 0, rb.size)
	for i := 0; i < rb.size; i++ {
		idx := (start + i) % rb.cap
		entry := rb.entries[idx]

		// Apply level filter
		if level != "" && entry.Level != level {
			continue
		}

		// Apply since filter
		if !since.IsZero() && !entry.Timestamp.After(since) {
			continue
		}

		result = append(result, entry)
	}

	// Apply limit (from the end, most recent entries)
	if limit > 0 && len(result) > limit {
		result = result[len(result)-limit:]
	}

	return result
}

// Clear removes all entries from the buffer
func (rb *RingBuffer) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.head = 0
	rb.size = 0
	rb.full = false
	// Reset entries to zero values
	for i := range rb.entries {
		rb.entries[i] = LogEntry{}
	}
}

// Size returns the current number of entries in the buffer
func (rb *RingBuffer) Size() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.size
}

// Capacity returns the maximum capacity of the buffer
func (rb *RingBuffer) Capacity() int {
	return rb.cap
}
