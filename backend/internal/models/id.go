package models

import "github.com/google/uuid"

// NewID generates a new UUIDv7 for use as a primary key.
// UUIDv7 is time-ordered, which provides better database index locality
// and natural chronological sorting compared to random UUIDv4.
func NewID() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}
