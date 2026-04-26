package config

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// gormLogger is set by SetGormLogger and used for all new DB connections.
// When nil, falls back to GORM's default logger.
var gormLogger gormlogger.Interface

// DatabaseConfig holds the PostgreSQL database configuration
type DatabaseConfig struct {
	Host        string
	Port        string
	User        string
	Password    string
	Database    string
	SSLMode     string
	Schema      string
	AutoMigrate bool // Run GORM AutoMigrate on startup (default: true)
}

// dsn builds the PostgreSQL connection string from the config.
func (cfg DatabaseConfig) dsn() string {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)
	if cfg.Schema != "" {
		dsn += fmt.Sprintf(" search_path=%s", cfg.Schema)
	}
	return dsn
}

// SetGormLogger configures the GORM logger used by all subsequent InitDatabase
// and InitMigrationDatabase calls. Call this once during application startup,
// after the application logger is initialised.
func SetGormLogger(l gormlogger.Interface) {
	gormLogger = l
}

// activeGormLogger returns the configured logger or GORM's default as fallback.
func activeGormLogger() gormlogger.Interface {
	if gormLogger != nil {
		return gormLogger
	}
	return gormlogger.Default.LogMode(gormlogger.Error)
}

// InitDatabase initializes GORM connection to PostgreSQL
func InitDatabase(cfg DatabaseConfig) (*gorm.DB, error) {
	config := &gorm.Config{
		Logger:      activeGormLogger(),
		PrepareStmt: true,
		//DisableForeignKeyConstraintWhenMigrating: true, // FK constraints are managed by migrations, not AutoMigrate
	}

	db, err := gorm.Open(postgres.Open(cfg.dsn()), config)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Configure connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	return db, nil
}

// InitMigrationDatabase opens a PostgreSQL connection using the simple protocol
// (no prepared statements) for running multi-statement migrations.
func InitMigrationDatabase(cfg DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.dsn(),
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: activeGormLogger(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open migration database: %w", err)
	}
	return db, nil
}

// CloseDatabase closes the database connection
func CloseDatabase(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}
	return sqlDB.Close()
}
