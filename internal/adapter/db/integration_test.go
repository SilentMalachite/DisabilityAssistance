package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestDatabase_RunActualMigrations(t *testing.T) {
	tmpDir := t.TempDir()

	// Use the actual migrations directory
	migrationDir := "../../../migrations"

	config := Config{
		Path:         filepath.Join(tmpDir, "integration_test.db"),
		MigrationDir: migrationDir,
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test running actual migrations
	err = db.RunMigrations(ctx)
	if err != nil {
		t.Errorf("RunMigrations() error = %v", err)
	}

	// Verify all expected tables were created
	expectedTables := []string{
		"staff",
		"recipients",
		"benefit_certificates",
		"staff_assignments",
		"consents",
		"audit_logs",
		"migrations", // Migration tracking table
	}

	for _, table := range expectedTables {
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

	// Verify all expected indexes were created
	expectedIndexes := []string{
		"idx_assignments_staff",
		"idx_assignments_recipient",
		"idx_certificates_recipient",
		"idx_consents_recipient",
		"idx_audit_actor",
		"idx_audit_at",
	}

	for _, index := range expectedIndexes {
		var count int
		query := `SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?`
		err = db.DB().QueryRowContext(ctx, query, index).Scan(&count)
		if err != nil {
			t.Errorf("failed to check index %s: %v", index, err)
		}
		if count != 1 {
			t.Errorf("index %s not found", index)
		}
	}

	// Verify initial admin user was created
	var adminCount int
	query := `SELECT COUNT(*) FROM staff WHERE role = 'admin'`
	err = db.DB().QueryRowContext(ctx, query).Scan(&adminCount)
	if err != nil {
		t.Errorf("failed to check admin user: %v", err)
	}
	if adminCount < 1 {
		t.Error("no admin user found after migration")
	}

	// Test migration status
	status, err := db.GetMigrationStatus(ctx)
	if err != nil {
		t.Errorf("GetMigrationStatus() error = %v", err)
	}

	if len(status) == 0 {
		t.Error("no migrations recorded")
	}

	// First migration should be 0001_init
	if len(status) > 0 {
		if status[0].Version != "0001" {
			t.Errorf("first migration version = %v, want %v", status[0].Version, "0001")
		}
		if status[0].Name != "init" {
			t.Errorf("first migration name = %v, want %v", status[0].Name, "init")
		}
	}

	// Test schema constraints by trying to insert invalid data
	t.Run("test schema constraints", func(t *testing.T) {
		// Test staff role constraint
		_, err := db.DB().ExecContext(ctx, `INSERT INTO staff (id, name, role, created_at, updated_at) VALUES ('test', 'Test', 'invalid_role', datetime('now'), datetime('now'))`)
		if err == nil {
			t.Error("invalid staff role should be rejected")
		}

		// Test foreign key constraint
		_, err = db.DB().ExecContext(ctx, `INSERT INTO recipients (id, name_cipher, sex_cipher, birth_date_cipher, has_disability_id_cipher, public_assistance_cipher, created_at, updated_at) VALUES ('test-recipient', 'test', 'test', 'test', 'test', 'test', datetime('now'), datetime('now'))`)
		if err != nil {
			t.Errorf("failed to insert test recipient: %v", err)
		}

		// This should fail due to foreign key constraint
		_, err = db.DB().ExecContext(ctx, `INSERT INTO staff_assignments (id, recipient_id, staff_id, assigned_at) VALUES ('test-assignment', 'test-recipient', 'nonexistent-staff', datetime('now'))`)
		if err == nil {
			t.Error("invalid staff_id should be rejected by foreign key constraint")
		}
	})
}

func TestDatabase_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		Path: filepath.Join(tmpDir, "concurrent_test.db"),
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test concurrent transactions
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			err := db.WithTransaction(ctx, func(ctx context.Context) error {
				// Create a temporary table for this transaction
				tx := ctx.Value("tx").(*sql.Tx)
				_, err := tx.ExecContext(ctx, `CREATE TEMP TABLE test_concurrent_`+string(rune('0'+id))+` (id INTEGER PRIMARY KEY)`)
				return err
			})

			if err != nil {
				t.Errorf("concurrent transaction %d failed: %v", id, err)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
