package db

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestNewDatabase(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		config    Config
		wantErr   bool
		errString string
	}{
		{
			name: "valid config",
			config: Config{
				Path:         filepath.Join(tmpDir, "test.db"),
				MigrationDir: filepath.Join(tmpDir, "migrations"),
			},
			wantErr: false,
		},
		{
			name: "config with default migration dir",
			config: Config{
				Path: filepath.Join(tmpDir, "test2.db"),
			},
			wantErr: false,
		},
		{
			name: "empty path",
			config: Config{
				Path: "",
			},
			wantErr:   true,
			errString: "database path is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewDatabase(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewDatabase() expected error but got none")
					return
				}
				if tt.errString != "" && err.Error() != tt.errString {
					t.Errorf("NewDatabase() error = %v, want %v", err.Error(), tt.errString)
				}
				return
			}

			if err != nil {
				t.Errorf("NewDatabase() unexpected error = %v", err)
				return
			}

			if db == nil {
				t.Error("NewDatabase() returned nil database")
				return
			}

			// Test that database is accessible
			if db.DB() == nil {
				t.Error("Database.DB() returned nil")
			}

			// Clean up
			db.Close()
		})
	}
}

func TestDatabase_Close(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		Path: filepath.Join(tmpDir, "test.db"),
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Test closing again should not error
	err = db.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

func TestDatabase_Health(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		Path: filepath.Join(tmpDir, "test.db"),
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	err = db.Health(ctx)
	if err != nil {
		t.Errorf("Health() error = %v", err)
	}

	// Test with cancelled context
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()

	err = db.Health(cancelledCtx)
	if err == nil {
		t.Error("Health() with cancelled context should return error")
	}
}

func TestDatabase_WithTransaction(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		Path: filepath.Join(tmpDir, "test.db"),
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("successful transaction", func(t *testing.T) {
		err := db.WithTransaction(ctx, func(ctx context.Context) error {
			tx := ctx.Value("tx").(*sql.Tx)
			if tx == nil {
				t.Error("transaction not found in context")
			}

			// Execute a test query
			_, err := tx.ExecContext(ctx, "CREATE TEMP TABLE test_temp (id INTEGER)")
			return err
		})

		if err != nil {
			t.Errorf("WithTransaction() error = %v", err)
		}
	})

	t.Run("transaction rollback on error", func(t *testing.T) {
		testErr := sql.ErrNoRows

		err := db.WithTransaction(ctx, func(ctx context.Context) error {
			tx := ctx.Value("tx").(*sql.Tx)
			if tx == nil {
				t.Error("transaction not found in context")
			}

			// This should be rolled back
			_, err := tx.ExecContext(ctx, "CREATE TEMP TABLE test_rollback (id INTEGER)")
			if err != nil {
				return err
			}

			return testErr
		})

		if err != testErr {
			t.Errorf("WithTransaction() error = %v, want %v", err, testErr)
		}
	})
}

func TestParseMigrationFilename(t *testing.T) {
	tests := []struct {
		filename    string
		wantVersion string
		wantName    string
	}{
		{
			filename:    "0001_init.sql",
			wantVersion: "0001",
			wantName:    "init",
		},
		{
			filename:    "0002_add_users.sql",
			wantVersion: "0002",
			wantName:    "add_users",
		},
		{
			filename:    "0010_complex_migration_name.sql",
			wantVersion: "0010",
			wantName:    "complex_migration_name",
		},
		{
			filename:    "invalid.sql",
			wantVersion: "",
			wantName:    "",
		},
		{
			filename:    "0001.sql",
			wantVersion: "",
			wantName:    "",
		},
		{
			filename:    "0001_init.txt",
			wantVersion: "",
			wantName:    "",
		},
		{
			filename:    "init.sql",
			wantVersion: "",
			wantName:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			version, name := parseMigrationFilename(tt.filename)

			if version != tt.wantVersion {
				t.Errorf("parseMigrationFilename() version = %v, want %v", version, tt.wantVersion)
			}

			if name != tt.wantName {
				t.Errorf("parseMigrationFilename() name = %v, want %v", name, tt.wantName)
			}
		})
	}
}

func TestDatabase_Migrations(t *testing.T) {
	tmpDir := t.TempDir()
	migrationDir := filepath.Join(tmpDir, "migrations")

	// Create migration directory
	err := os.MkdirAll(migrationDir, 0755)
	if err != nil {
		t.Fatalf("failed to create migration directory: %v", err)
	}

	// Create test migration files
	migration1 := `CREATE TABLE test_table1 (id INTEGER PRIMARY KEY);`
	migration2 := `CREATE TABLE test_table2 (id INTEGER PRIMARY KEY);`

	err = os.WriteFile(filepath.Join(migrationDir, "0001_first.sql"), []byte(migration1), 0644)
	if err != nil {
		t.Fatalf("failed to write migration file: %v", err)
	}

	err = os.WriteFile(filepath.Join(migrationDir, "0002_second.sql"), []byte(migration2), 0644)
	if err != nil {
		t.Fatalf("failed to write migration file: %v", err)
	}

	config := Config{
		Path:         filepath.Join(tmpDir, "test.db"),
		MigrationDir: migrationDir,
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test running migrations
	err = db.RunMigrations(ctx)
	if err != nil {
		t.Errorf("RunMigrations() error = %v", err)
	}

	// Verify tables were created
	tables := []string{"test_table1", "test_table2"}
	for _, table := range tables {
		var count int
		query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`
		err = db.DB().QueryRowContext(ctx, query, table).Scan(&count)
		if err != nil {
			t.Errorf("failed to check table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("table %s not found", table)
		}
	}

	// Test getting migration status
	status, err := db.GetMigrationStatus(ctx)
	if err != nil {
		t.Errorf("GetMigrationStatus() error = %v", err)
	}

	if len(status) != 2 {
		t.Errorf("GetMigrationStatus() returned %d migrations, want 2", len(status))
	}

	// Verify migration details
	expectedMigrations := []struct {
		version string
		name    string
	}{
		{"0001", "first"},
		{"0002", "second"},
	}

	for i, expected := range expectedMigrations {
		if i >= len(status) {
			t.Errorf("migration %d not found", i)
			continue
		}

		if status[i].Version != expected.version {
			t.Errorf("migration %d version = %v, want %v", i, status[i].Version, expected.version)
		}

		if status[i].Name != expected.name {
			t.Errorf("migration %d name = %v, want %v", i, status[i].Name, expected.name)
		}

		if status[i].AppliedAt.IsZero() {
			t.Errorf("migration %d AppliedAt is zero", i)
		}
	}

	// Test running migrations again (should be idempotent)
	err = db.RunMigrations(ctx)
	if err != nil {
		t.Errorf("Second RunMigrations() error = %v", err)
	}

	// Status should still be the same
	status2, err := db.GetMigrationStatus(ctx)
	if err != nil {
		t.Errorf("Second GetMigrationStatus() error = %v", err)
	}

	if len(status2) != len(status) {
		t.Errorf("Second GetMigrationStatus() returned %d migrations, want %d", len(status2), len(status))
	}
}

func TestDatabase_MigrationErrors(t *testing.T) {
	tmpDir := t.TempDir()
	migrationDir := filepath.Join(tmpDir, "migrations")

	// Create migration directory
	err := os.MkdirAll(migrationDir, 0755)
	if err != nil {
		t.Fatalf("failed to create migration directory: %v", err)
	}

	// Create invalid migration file
	invalidMigration := `INVALID SQL SYNTAX;`
	err = os.WriteFile(filepath.Join(migrationDir, "0001_invalid.sql"), []byte(invalidMigration), 0644)
	if err != nil {
		t.Fatalf("failed to write migration file: %v", err)
	}

	config := Config{
		Path:         filepath.Join(tmpDir, "test.db"),
		MigrationDir: migrationDir,
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test that invalid migration fails
	err = db.RunMigrations(ctx)
	if err == nil {
		t.Error("RunMigrations() with invalid SQL should return error")
	}

	// Verify no migration was recorded
	status, err := db.GetMigrationStatus(ctx)
	if err != nil {
		t.Errorf("GetMigrationStatus() error = %v", err)
	}

	if len(status) != 0 {
		t.Errorf("GetMigrationStatus() returned %d migrations after failed migration, want 0", len(status))
	}
}

func BenchmarkDatabase_WithTransaction(b *testing.B) {
	tmpDir := b.TempDir()

	config := Config{
		Path: filepath.Join(tmpDir, "benchmark.db"),
	}

	db, err := NewDatabase(config)
	if err != nil {
		b.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := db.WithTransaction(ctx, func(ctx context.Context) error {
			tx := ctx.Value("tx").(*sql.Tx)
			_, err := tx.ExecContext(ctx, "SELECT 1")
			return err
		})
		if err != nil {
			b.Fatalf("WithTransaction() error = %v", err)
		}
	}
}
