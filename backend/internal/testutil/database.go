// Package testutil provides utilities for testing, including database setup and helpers.
package testutil

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/example/go-react-starter/internal/models"
)

// TestBcryptCost is the bcrypt cost used during tests.
// bcrypt.MinCost (4) is the fastest option and is safe for tests.
const TestBcryptCost = 4

var bcryptSetupOnce sync.Once

// Singleton PostgreSQL container shared across all test packages.
// Container cleanup is handled by Ryuk (testcontainers reaper).
var (
	pgContainer *tcpostgres.PostgresContainer
	pgConnStr   string
	pgStartOnce sync.Once
	pgStartErr  error
)

// Template schema: AutoMigrate runs once into a template schema.
// Each test clones the template instead of re-running AutoMigrate.
// Template name includes PID to prevent cross-package interference
// (each package runs in its own process via `go test`).
var (
	templateOnce   sync.Once
	templateSchema string
	templateTables []string
	templateFKs    []fkConstraint
	templateErr    error
)

// Shared admin connection for schema create/drop operations.
// Avoids opening a new connection per test just for DDL.
var (
	adminDBOnce sync.Once
	adminDB     *gorm.DB
	adminDBErr  error
)

type fkConstraint struct {
	TableName      string `gorm:"column:table_name"`
	ConstraintName string `gorm:"column:constraint_name"`
	Definition     string `gorm:"column:definition"`
}

// SetupFastBcrypt reduces bcrypt cost for all tests in this package.
// This makes password hashing ~50x faster (from ~200ms to ~4ms per hash).
// Should be called once at the start of test runs.
// Returns a cleanup function that restores the original cost.
func SetupFastBcrypt() func() {
	return models.SetBcryptCostForTesting(TestBcryptCost)
}

// getTestDBURL returns the PostgreSQL connection string for testing.
// If TEST_DB_URL is set, it uses that directly.
// Otherwise, it starts a PostgreSQL container via testcontainers-go (singleton).
// Container cleanup is handled automatically by Ryuk (testcontainers reaper).
func getTestDBURL(t testing.TB) string {
	t.Helper()

	// Check for explicit override
	if url := os.Getenv("TEST_DB_URL"); url != "" {
		return url
	}

	// Start singleton container
	pgStartOnce.Do(func() {
		ctx := context.Background()
		pgContainer, pgStartErr = tcpostgres.Run(ctx,
			"postgres:17-alpine",
			tcpostgres.WithDatabase("starter_test"),
			tcpostgres.WithUsername("test"),
			tcpostgres.WithPassword("test"),
			tcpostgres.BasicWaitStrategies(),
		)
		if pgStartErr != nil {
			return
		}
		pgConnStr, pgStartErr = pgContainer.ConnectionString(ctx, "sslmode=disable")
	})
	require.NoError(t, pgStartErr, "Failed to start PostgreSQL test container")

	return pgConnStr
}

// quoteIdentifier safely quotes a PostgreSQL identifier to prevent SQL injection.
func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// getAdminDB returns a shared connection for schema DDL operations.
// This avoids opening a new connection per test just for CREATE/DROP SCHEMA.
func getAdminDB(t testing.TB, connStr string) *gorm.DB {
	t.Helper()
	adminDBOnce.Do(func() {
		adminDB, adminDBErr = gorm.Open(postgres.Open(connStr), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	})
	require.NoError(t, adminDBErr, "Failed to connect admin DB")
	return adminDB
}

// ensureTemplate creates a template schema with all models migrated.
// AutoMigrate runs once; each test then clones the template via CREATE TABLE ... LIKE.
func ensureTemplate(t testing.TB, connStr string) {
	templateOnce.Do(func() {
		db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			templateErr = fmt.Errorf("connect for template: %w", err)
			return
		}
		defer func() {
			if sqlDB, err := db.DB(); err == nil {
				sqlDB.Close()
			}
		}()

		// Pin to a single connection so SET search_path applies to all AutoMigrate queries
		sqlDB, err := db.DB()
		if err != nil {
			templateErr = fmt.Errorf("get sql.DB for template: %w", err)
			return
		}
		sqlDB.SetMaxOpenConns(1)

		// Create template schema and run AutoMigrate (once for the entire process)
		templateSchema = fmt.Sprintf("_tmpl_%d", os.Getpid())
		db.Exec("DROP SCHEMA IF EXISTS " + quoteIdentifier(templateSchema) + " CASCADE")
		if err := db.Exec("CREATE SCHEMA " + quoteIdentifier(templateSchema)).Error; err != nil {
			templateErr = fmt.Errorf("create template schema: %w", err)
			return
		}
		if err := db.Exec("SET search_path TO " + quoteIdentifier(templateSchema)).Error; err != nil {
			templateErr = fmt.Errorf("set search_path for template: %w", err)
			return
		}
		if err := db.AutoMigrate(models.AllModels()...); err != nil {
			templateErr = fmt.Errorf("automigrate template: %w", err)
			return
		}

		// Cache the list of tables
		rows, err := db.Raw(
			"SELECT tablename FROM pg_tables WHERE schemaname = $1 ORDER BY tablename",
			templateSchema,
		).Rows()
		if err != nil {
			templateErr = fmt.Errorf("list template tables: %w", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var name string
			rows.Scan(&name)
			templateTables = append(templateTables, name)
		}

		// Cache foreign key constraint definitions for cloning
		if err := db.Raw(`
			SELECT cl.relname AS table_name,
			       con.conname AS constraint_name,
			       pg_get_constraintdef(con.oid) AS definition
			FROM pg_constraint con
			JOIN pg_class cl ON cl.oid = con.conrelid
			JOIN pg_namespace ns ON ns.oid = cl.relnamespace
			WHERE con.contype = 'f' AND ns.nspname = $1
		`, templateSchema).Scan(&templateFKs).Error; err != nil {
			templateErr = fmt.Errorf("cache FK constraints: %w", err)
			return
		}
	})
	require.NoError(t, templateErr, "Failed to create template schema")
}

// SetupTestDB creates an isolated PostgreSQL test database.
// Uses testcontainers-go by default, or TEST_DB_URL if set.
// Each call creates a unique schema cloned from a pre-migrated template,
// avoiding the expensive AutoMigrate introspection per test.
//
// Example usage:
//
//	func TestMyFeature(t *testing.T) {
//	    db := testutil.SetupTestDB(t)
//	    // Use db for testing... (cleanup is automatic via t.Cleanup)
//	}
func SetupTestDB(t testing.TB) *gorm.DB {
	t.Helper()

	connStr := getTestDBURL(t)

	// Ensure template schema exists (AutoMigrate runs once for the process)
	ensureTemplate(t, connStr)

	schemaName := fmt.Sprintf("test_%s", uuid.New().String()[:8])

	// Create schema via shared admin connection
	admin := getAdminDB(t, connStr)
	err := admin.Exec("CREATE SCHEMA " + quoteIdentifier(schemaName)).Error
	require.NoError(t, err, "Failed to create test schema")

	// Open the test connection with search_path baked into the DSN.
	// This ensures ALL connections in the pool use the correct schema,
	// which is required for t.Parallel() safety.
	separator := "&"
	if !strings.Contains(connStr, "?") {
		separator = "?"
	}
	connStrWithSchema := connStr + separator + "search_path=" + schemaName

	db, err := gorm.Open(postgres.Open(connStrWithSchema), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Silent),
		PrepareStmt: true,
	})
	require.NoError(t, err, "Failed to connect to PostgreSQL test database with schema")

	// Limit pool to prevent unnecessary connections during tests
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)

	// Clone tables from template (fast: no GORM introspection needed)
	for _, table := range templateTables {
		err := db.Exec(fmt.Sprintf(
			"CREATE TABLE %s (LIKE %s.%s INCLUDING ALL)",
			quoteIdentifier(table),
			quoteIdentifier(templateSchema),
			quoteIdentifier(table),
		)).Error
		require.NoError(t, err, "Failed to clone table from template: "+table)
	}

	// Add foreign key constraints (unqualified refs resolve via search_path)
	for _, fk := range templateFKs {
		err := db.Exec(fmt.Sprintf(
			"ALTER TABLE %s ADD CONSTRAINT %s %s",
			quoteIdentifier(fk.TableName),
			quoteIdentifier(fk.ConstraintName),
			fk.Definition,
		)).Error
		require.NoError(t, err, "Failed to add FK constraint: "+fk.ConstraintName)
	}

	// Register cleanup to drop the schema when the test finishes
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
		admin := getAdminDB(t, connStr)
		admin.Exec("DROP SCHEMA IF EXISTS " + quoteIdentifier(schemaName) + " CASCADE")
	})

	return db
}

// CleanupTestDB cleans all data from the test database while preserving the schema.
// Useful for resetting state between test cases.
func CleanupTestDB(t testing.TB, db *gorm.DB) {
	t.Helper()

	// Delete in reverse dependency order (children first, then parents)
	// to respect foreign key constraints.
	allModels := models.AllModels()
	for i := len(allModels) - 1; i >= 0; i-- {
		err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(allModels[i]).Error
		require.NoError(t, err, fmt.Sprintf("Failed to cleanup model: %T", allModels[i]))
	}
}

// TruncateTable removes all data from a specific table using GORM.
func TruncateTable(t testing.TB, db *gorm.DB, model interface{}) {
	t.Helper()
	err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(model).Error
	require.NoError(t, err, fmt.Sprintf("Failed to truncate table for model: %T", model))
}

// CreateTestUser creates a test user in the database and returns the user ID.
// The password is automatically hashed.
func CreateTestUser(t *testing.T, db *gorm.DB, email, password string) uuid.UUID {
	t.Helper()

	user := &models.User{
		Email: email,
	}
	err := user.SetPassword(password)
	require.NoError(t, err, "Failed to set user password")

	err = db.Create(user).Error
	require.NoError(t, err, "Failed to create test user")

	return user.ID
}






// CreateTestNote creates a test note in the database and returns the note.
func CreateTestNote(t *testing.T, db *gorm.DB, userID uuid.UUID, title, content string) *models.Note {
	t.Helper()
	note := &models.Note{
		UserID:  userID,
		Title:   title,
		Content: content,
	}
	err := db.Create(note).Error
	require.NoError(t, err, "Failed to create test note")
	return note
}
