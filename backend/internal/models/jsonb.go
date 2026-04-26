package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONB is a PostgreSQL jsonb column type.
// It replaces gorm.io/datatypes.JSON to avoid pulling in MySQL/SQLite drivers.
type JSONB []byte

// Scan implements sql.Scanner for reading jsonb from the database.
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = append((*j)[:0], v...)
	case string:
		*j = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}
	return nil
}

// Value implements driver.Valuer for writing jsonb to the database.
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return string(j), nil
}

// MarshalJSON implements json.Marshaler — returns raw JSON bytes.
func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *JSONB) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*j = nil
		return nil
	}
	// Validate it's valid JSON
	if !json.Valid(b) {
		return fmt.Errorf("invalid JSON")
	}
	*j = append((*j)[:0], b...)
	return nil
}
