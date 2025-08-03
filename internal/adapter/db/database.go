package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"shien-system/internal/domain"

	_ "github.com/mattn/go-sqlite3"
)

// Database represents a SQLite database connection
type Database struct {
	db           *sql.DB
	migrationDir string
}

// Config holds database configuration
type Config struct {
	Path         string
	MigrationDir string
}

// NewDatabase creates a new database connection
func NewDatabase(config Config) (*Database, error) {
	if config.Path == "" {
		return nil, fmt.Errorf("database path is required")
	}

	if config.MigrationDir == "" {
		config.MigrationDir = "migrations"
	}

	// Ensure the directory exists
	dir := filepath.Dir(config.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection with proper settings for concurrent access
	dsn := fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL&_foreign_keys=1&_busy_timeout=30000", config.Path)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{
		db:           db,
		migrationDir: config.MigrationDir,
	}

	// Initialize migration tracking table
	if err := database.initMigrationTable(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize migration table: %w", err)
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// DB returns the underlying sql.DB instance
func (d *Database) DB() *sql.DB {
	return d.db
}

// WithTransaction executes a function within a database transaction
func (d *Database) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Store transaction in context for repository use
	ctx = context.WithValue(ctx, "tx", tx)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			// Log the panic but don't re-panic to prevent application crash
			if panicErr, ok := r.(error); ok {
				err = fmt.Errorf("transaction panic recovered: %w", panicErr)
			} else {
				err = fmt.Errorf("transaction panic recovered: %v", r)
			}
		}
	}()

	err = fn(ctx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// initMigrationTable creates the migration tracking table
func (d *Database) initMigrationTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			version TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TEXT NOT NULL
		)`

	_, err := d.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

// RunMigrations runs all pending migrations
func (d *Database) RunMigrations(ctx context.Context) error {
	// Get all migration files
	files, err := d.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := d.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, file := range files {
		version, name := parseMigrationFilename(file.Name())
		if version == "" {
			continue // Skip invalid filenames
		}

		// Check if already applied
		if _, applied := appliedMigrations[version]; applied {
			continue
		}

		if err := d.runMigration(ctx, file, version, name); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", version, err)
		}
	}

	return nil
}

// GetMigrationStatus returns the status of all migrations
func (d *Database) GetMigrationStatus(ctx context.Context) ([]domain.MigrationStatus, error) {
	query := `SELECT version, name, applied_at FROM migrations ORDER BY version`

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	var migrations []domain.MigrationStatus
	for rows.Next() {
		var migration domain.MigrationStatus
		var appliedAtStr string

		if err := rows.Scan(&migration.Version, &migration.Name, &appliedAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan migration row: %w", err)
		}

		appliedAt, err := time.Parse(time.RFC3339, appliedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse applied_at time: %w", err)
		}
		migration.AppliedAt = appliedAt

		migrations = append(migrations, migration)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return migrations, nil
}

// getMigrationFiles returns all migration files sorted by version
func (d *Database) getMigrationFiles() ([]os.DirEntry, error) {
	entries, err := os.ReadDir(d.migrationDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No migration directory is fine
		}
		return nil, fmt.Errorf("failed to read migration directory: %w", err)
	}

	var migrationFiles []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".sql") {
			migrationFiles = append(migrationFiles, entry)
		}
	}

	// Sort by filename (which should include version number)
	sort.Slice(migrationFiles, func(i, j int) bool {
		return migrationFiles[i].Name() < migrationFiles[j].Name()
	})

	return migrationFiles, nil
}

// getAppliedMigrations returns a map of applied migration versions
func (d *Database) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	query := `SELECT version FROM migrations`

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return applied, nil
}

// runMigration executes a single migration file
func (d *Database) runMigration(ctx context.Context, file os.DirEntry, version, name string) error {
	filePath := filepath.Join(d.migrationDir, file.Name())

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration within a transaction
	return d.WithTransaction(ctx, func(ctx context.Context) error {
		tx := ctx.Value("tx").(*sql.Tx)

		// Execute the migration SQL
		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration SQL: %w", err)
		}

		// Record the migration as applied
		recordQuery := `INSERT INTO migrations (version, name, applied_at) VALUES (?, ?, ?)`
		if _, err := tx.ExecContext(ctx, recordQuery, version, name, time.Now().UTC().Format(time.RFC3339)); err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}

		return nil
	})
}

// parseMigrationFilename extracts version and name from a migration filename
// Expected format: NNNN_name.sql (e.g., "0001_init.sql")
func parseMigrationFilename(filename string) (version, name string) {
	if !strings.HasSuffix(filename, ".sql") {
		return "", ""
	}

	nameWithoutExt := strings.TrimSuffix(filename, ".sql")
	parts := strings.SplitN(nameWithoutExt, "_", 2)

	if len(parts) < 2 {
		return "", ""
	}

	return parts[0], parts[1]
}

// Health checks database connectivity
func (d *Database) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := d.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Test a simple query
	var result int
	if err := d.db.QueryRowContext(ctx, "SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("database query test failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("database query returned unexpected result: %d", result)
	}

	return nil
}
