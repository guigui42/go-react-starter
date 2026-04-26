// Package migrations provides a versioned database migration system.
//
// Migrations are Go functions registered in a global registry.
// Each migration has a version string, a description, and up/down functions.
// The migration runner tracks applied versions in a schema_migrations table
// and applies them in order.
//
// This system uses GORM for PostgreSQL database migrations,
// allowing migration code to work with the PostgreSQL database engine.
package migrations

import (
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
)

// SchemaMigration tracks applied migrations in the database.
type SchemaMigration struct {
	ID           uint      `gorm:"primaryKey"`
	Version      string    `gorm:"uniqueIndex;not null;size:255"`
	Name         string    `gorm:"size:255;default:''"` // Default empty string for backwards compatibility
	AppliedAt    time.Time `gorm:"not null"`
	Status       string    `gorm:"size:20;default:'applied'"` // "applied" or "failed"
	ErrorMessage string    `gorm:"type:text;default:''"` // Error details when status is "failed"
}

// Migration represents a single versioned migration with up and down functions.
type Migration struct {
	// Version is the unique identifier for the migration (e.g., "001", "002").
	// Versions must be zero-padded strings so that lexicographic ordering
	// matches numeric ordering (e.g., "001" < "002" < "010").
	Version string
	// Name is a human-readable description of the migration.
	Name string
	// Up applies the migration (schema changes, data seeding, etc.).
	Up func(db *gorm.DB) error
	// Down rolls back the migration. It is required and must not be nil.
	Down func(db *gorm.DB) error
}

// registry holds all registered migrations.
var registry []Migration

// Register adds a migration to the global registry.
// Migrations should be registered via init() in their respective files.
// Panics if a migration with the same version is already registered.
func Register(m Migration) {
	for _, existing := range registry {
		if existing.Version == m.Version {
			panic(fmt.Sprintf("duplicate migration version %s (%q and %q)", m.Version, existing.Name, m.Name))
		}
	}
	registry = append(registry, m)
}

// All returns all registered migrations sorted by version.
func All() []Migration {
	sorted := make([]Migration, len(registry))
	copy(sorted, registry)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Version < sorted[j].Version
	})
	return sorted
}

// Logger is the interface required by the migration runner for logging.
type Logger interface {
	Info(message string)
	Error(message string, err error)
}

// Runner executes migrations against a database.
type Runner struct {
	db  *gorm.DB
	log Logger
}

// NewRunner creates a new migration runner.
func NewRunner(db *gorm.DB, log Logger) *Runner {
	return &Runner{db: db, log: log}
}

// ensureTable creates the schema_migrations tracking table if it doesn't exist.
func (r *Runner) ensureTable() error {
	return r.db.AutoMigrate(&SchemaMigration{})
}

// applied returns the set of migration versions that have been successfully applied.
func (r *Runner) applied() (map[string]bool, error) {
	var migrations []SchemaMigration
	if err := r.db.Where("status = ? OR status IS NULL OR status = ''", "applied").Find(&migrations).Error; err != nil {
		return nil, fmt.Errorf("query applied migrations: %w", err)
	}
	result := make(map[string]bool, len(migrations))
	for _, m := range migrations {
		result[m.Version] = true
	}
	return result, nil
}

// Up applies all pending migrations in order.
func (r *Runner) Up() error {
	if err := r.ensureTable(); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	appliedSet, err := r.applied()
	if err != nil {
		return err
	}

	all := All()
	applied := 0
	for _, m := range all {
		if appliedSet[m.Version] {
			continue
		}

		// Clear any previous failure record for this version before retrying
		r.db.Where("version = ? AND status = ?", m.Version, "failed").Delete(&SchemaMigration{})

		r.log.Info(fmt.Sprintf("applying migration %s: %s", m.Version, m.Name))

		// Wrap the migration and its tracking record in a transaction so that
		// a partial failure does not leave the database in an inconsistent state.
		// Note: some DDL statements (e.g., CREATE TABLE in certain databases)
		// may not be fully transactional; PostgreSQL supports
		// transactional DDL for most operations.
		txErr := r.db.Transaction(func(tx *gorm.DB) error {
			if err := m.Up(tx); err != nil {
				return err
			}

			record := SchemaMigration{
				Version:   m.Version,
				Name:      m.Name,
				AppliedAt: time.Now(),
				Status:    "applied",
			}
			if err := tx.Create(&record).Error; err != nil {
				return fmt.Errorf("record migration %s: %w", m.Version, err)
			}

			return nil
		})
		if txErr != nil {
			// Record the failure outside the rolled-back transaction
			r.recordFailure(m.Version, m.Name, txErr)
			return fmt.Errorf("migration %s (%s) failed: %w", m.Version, m.Name, txErr)
		}

		applied++
	}

	if applied == 0 {
		r.log.Info("no pending migrations")
	} else {
		r.log.Info(fmt.Sprintf("applied %d migration(s)", applied))
	}
	return nil
}

// recordFailure persists a failed migration record so the admin UI can display it.
func (r *Runner) recordFailure(version, name string, migErr error) {
	errMsg := migErr.Error()
	// Truncate very long error messages
	if len(errMsg) > 2000 {
		errMsg = errMsg[:2000] + "..."
	}
	record := SchemaMigration{
		Version:      version,
		Name:         name,
		AppliedAt:    time.Now(),
		Status:       "failed",
		ErrorMessage: errMsg,
	}
	if err := r.db.Create(&record).Error; err != nil {
		r.log.Error(fmt.Sprintf("failed to record migration failure for %s", version), err)
	}
}

// Down rolls back the last applied migration.
func (r *Runner) Down() error {
	if err := r.ensureTable(); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	// Find the last successfully applied migration (skip failed ones)
	var last SchemaMigration
	result := r.db.Where("status = ?", "applied").Order("version DESC").First(&last)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			r.log.Info("no migrations to roll back")
			return nil
		}
		return fmt.Errorf("find last migration: %w", result.Error)
	}

	// Find the migration definition
	all := All()
	var migration *Migration
	for i := range all {
		if all[i].Version == last.Version {
			migration = &all[i]
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration %s not found in registry", last.Version)
	}

	if migration.Down == nil {
		return fmt.Errorf("migration %s (%s) does not support rollback", migration.Version, migration.Name)
	}

	r.log.Info(fmt.Sprintf("rolling back migration %s: %s", migration.Version, migration.Name))

	if err := migration.Down(r.db); err != nil {
		return fmt.Errorf("rollback migration %s (%s) failed: %w", migration.Version, migration.Name, err)
	}

	if err := r.db.Where("version = ?", migration.Version).Delete(&SchemaMigration{}).Error; err != nil {
		return fmt.Errorf("remove migration record %s: %w", migration.Version, err)
	}

	r.log.Info(fmt.Sprintf("rolled back migration %s: %s", migration.Version, migration.Name))
	return nil
}

// Status returns the current migration status.
func (r *Runner) Status() ([]MigrationStatus, error) {
	if err := r.ensureTable(); err != nil {
		return nil, fmt.Errorf("create schema_migrations table: %w", err)
	}

	var allRecords []SchemaMigration
	if err := r.db.Order("version ASC").Find(&allRecords).Error; err != nil {
		return nil, fmt.Errorf("query migrations: %w", err)
	}
	recordMap := make(map[string]SchemaMigration, len(allRecords))
	for _, rec := range allRecords {
		recordMap[rec.Version] = rec
	}

	all := All()
	statuses := make([]MigrationStatus, 0, len(all))
	for _, m := range all {
		status := MigrationStatus{
			Version: m.Version,
			Name:    m.Name,
		}
		if rec, ok := recordMap[m.Version]; ok {
			status.AppliedAt = &rec.AppliedAt
			if rec.Status == "failed" {
				status.Applied = false
				status.Failed = true
				status.ErrorMessage = rec.ErrorMessage
			} else {
				status.Applied = true
			}
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// MigrationStatus represents the status of a single migration.
type MigrationStatus struct {
	Version      string
	Name         string
	Applied      bool
	Failed       bool
	AppliedAt    *time.Time
	ErrorMessage string
}
