# Database & Migrations

## PostgreSQL Setup

### Development (Docker)

```bash
make db-up      # Start PostgreSQL 17 on port 5434
make db-down    # Stop container
make db-shell   # Open psql shell
make db-reset   # Destroy and recreate
```

Default connection:

| Setting | Value |
|---------|-------|
| Host | `localhost` |
| Port | `5434` (configurable via `DB_PORT`) |
| User | `postgres` |
| Password | `postgres` |
| Database | `starter` |

### Connection Configuration

```bash
# backend/.env
PGHOST=localhost
PGPORT=5434
PGUSER=postgres
PGPASSWORD=postgres
PGDATABASE=starter
PGSSLMODE=disable
PGSCHEMA=public
AUTO_MIGRATE=true
```

### Two Database Connections

The backend uses two separate GORM connections:

1. **Migration connection** (`InitMigrationDatabase`) — Uses `PreferSimpleProtocol` for multi-statement DDL. Closed after migrations complete.
2. **Main connection** (`InitDatabase`) — Uses `PrepareStmt: true` for performance. Used for all runtime queries.

## GORM Configuration

- **PrepareStmt**: Enabled for the main connection (prepared statement cache)
- **Logger**: Zerolog integration, query logging in development
- **NamingStrategy**: Default GORM snake_case
- **SoftDelete**: Supported via `gorm.DeletedAt` field

## Migration System

### Overview

Versioned Go migrations with automatic tracking:

```
backend/internal/migrations/
├── runner.go              # Migration runner (Up, Down, Status)
└── 001_initial_schema.go  # Initial schema migration
```

### Creating a Migration

```go
// backend/internal/migrations/002_add_tasks.go
package migrations

import "gorm.io/gorm"

func init() {
    Register(Migration{
        Version: "002",
        Name:    "add_tasks",
        Up: func(db *gorm.DB) error {
            type Task struct {
                ID     string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
                UserID string `gorm:"type:uuid;not null;index"`
                Title  string `gorm:"not null;size:255"`
            }
            return db.AutoMigrate(&Task{})
        },
        Down: func(db *gorm.DB) error {
            return db.Migrator().DropTable("tasks")
        },
    })
}
```

### Conventions

| Rule | Detail |
|------|--------|
| Version format | Zero-padded strings: `"001"`, `"002"`, `"010"` |
| File naming | `{version}_{description}.go` |
| Registration | Via `init()` function calling `Register()` |
| `Down` function | Required — must not be nil |
| Transaction | Each migration runs in a transaction |

### Tracking Table

`schema_migrations` stores applied versions:

| Column | Type | Purpose |
|--------|------|---------|
| `version` | string (unique) | Migration version |
| `name` | string | Human-readable description |
| `applied_at` | timestamp | When it was applied |
| `status` | string | `"applied"` or `"failed"` |
| `error_message` | text | Error details on failure |

### Runner API

```go
runner := migrations.NewRunner(db, logger)
runner.Up()       // Apply all pending migrations
runner.Down()     // Roll back the last applied migration
runner.Status()   // Get status of all known migrations
```

### Admin Dashboard

Migration status is visible in the admin panel under the **Migrations** tab, showing applied/pending/failed status for each migration.

## AutoMigrate

In addition to versioned migrations, GORM `AutoMigrate` runs on startup when `AUTO_MIGRATE=true`:

```go
db.AutoMigrate(models.AllModels()...)
```

This handles schema drift (adding new columns from model changes) without requiring explicit migrations. **Versioned migrations** are for:
- Data transformations
- Renaming columns/tables
- Complex schema changes
- Anything that `AutoMigrate` can't handle

## Model Registry

`models.AllModels()` in `backend/internal/models/registry.go` is the single source of truth for all database models. It drives:

1. Server startup `AutoMigrate`
2. Test database setup
3. Migration 001 (initial schema)

When adding a new model, always add it to `AllModels()`.

## Row-Level Security

See [Security](security.md#row-level-security) for details on GORM scopes and the user scope guard.
