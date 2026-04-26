package models

import (
	"database/sql/driver"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Decimal is a custom type for handling decimal values in PostgreSQL
// Stored as NUMERIC(precision, scale)
type Decimal struct {
	Float64   float64
	Precision int
	Scale     int
}

// Scan implements sql.Scanner interface
func (d *Decimal) Scan(value interface{}) error {
	if value == nil {
		d.Float64 = 0
		return nil
	}

	switch v := value.(type) {
	case float64:
		d.Float64 = v
	case []byte:
		_, err := fmt.Sscanf(string(v), "%f", &d.Float64)
		return err
	case string:
		_, err := fmt.Sscanf(v, "%f", &d.Float64)
		return err
	default:
		return fmt.Errorf("cannot scan type %T into Decimal", value)
	}

	return nil
}

// Value implements driver.Valuer interface
func (d Decimal) Value() (driver.Value, error) {
	return d.Float64, nil
}

// GormDataType returns the data type for GORM
func (Decimal) GormDataType() string {
	return "decimal"
}

// GormDBDataType returns database-specific data type
func (d Decimal) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	precision := 12
	scale := 4

	// Check if precision and scale are specified in the tag
	if val, ok := field.TagSettings["PRECISION"]; ok {
		fmt.Sscanf(val, "%d", &precision)
	}
	if val, ok := field.TagSettings["SCALE"]; ok {
		fmt.Sscanf(val, "%d", &scale)
	}

	return fmt.Sprintf("NUMERIC(%d,%d)", precision, scale)
}
