package db

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"shien-system/internal/domain"
)

func TestRepositoryIntegration_EncryptionRoundTrip(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	// Test with comprehensive Japanese data
	recipient := &domain.Recipient{
		ID:               "integration-001",
		Name:             "統合テスト太郎",
		Kana:             "トウゴウテストタロウ",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1985, 3, 15, 0, 0, 0, 0, time.UTC),
		DisabilityName:   "知的障害・自閉症スペクトラム障害",
		HasDisabilityID:  true,
		Grade:            "B2",
		Address:          "〒150-0001 東京都渋谷区神宮前1-1-1 ○○マンション101号室",
		Phone:            "090-1234-5678",
		Email:            "integration@example.co.jp",
		PublicAssistance: true,
		AdmissionDate:    &now,
		DischargeDate:    nil,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Create the recipient
	err := recipientRepo.Create(ctx, recipient)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Retrieve and verify all fields are correctly decrypted
	retrieved, err := recipientRepo.GetByID(ctx, recipient.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	// Verify all sensitive fields match exactly
	if retrieved.Name != recipient.Name {
		t.Errorf("Name = %v, want %v", retrieved.Name, recipient.Name)
	}
	if retrieved.Kana != recipient.Kana {
		t.Errorf("Kana = %v, want %v", retrieved.Kana, recipient.Kana)
	}
	if retrieved.DisabilityName != recipient.DisabilityName {
		t.Errorf("DisabilityName = %v, want %v", retrieved.DisabilityName, recipient.DisabilityName)
	}
	if retrieved.Address != recipient.Address {
		t.Errorf("Address = %v, want %v", retrieved.Address, recipient.Address)
	}
	if retrieved.Phone != recipient.Phone {
		t.Errorf("Phone = %v, want %v", retrieved.Phone, recipient.Phone)
	}
	if retrieved.Email != recipient.Email {
		t.Errorf("Email = %v, want %v", retrieved.Email, recipient.Email)
	}
}

func TestRepositoryIntegration_DataIsActuallyEncrypted(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	sensitiveData := "機密情報・個人情報"

	recipient := &domain.Recipient{
		ID:               "encryption-test-001",
		Name:             sensitiveData,
		Sex:              domain.SexFemale,
		BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		HasDisabilityID:  false,
		PublicAssistance: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err := recipientRepo.Create(ctx, recipient)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Directly query the database to verify data is encrypted
	var nameCipher []byte
	query := `SELECT name_cipher FROM recipients WHERE id = ?`
	err = db.DB().QueryRowContext(ctx, query, recipient.ID).Scan(&nameCipher)
	if err != nil {
		t.Fatalf("Direct database query error = %v", err)
	}

	// Verify that the encrypted data is NOT the same as plaintext
	if string(nameCipher) == sensitiveData {
		t.Error("Sensitive data is stored in plaintext - encryption failed!")
	}

	// Verify encrypted data is not empty
	if len(nameCipher) == 0 {
		t.Error("Encrypted data is empty")
	}

	// The encrypted data should be longer than plaintext due to nonce
	if len(nameCipher) <= len(sensitiveData) {
		t.Error("Encrypted data should be longer than plaintext")
	}

	t.Logf("Original data length: %d, Encrypted data length: %d", len(sensitiveData), len(nameCipher))
}

func TestRepositoryIntegration_TransactionRollback(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	recipient := &domain.Recipient{
		ID:               "transaction-test-001",
		Name:             "トランザクションテスト太郎",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		HasDisabilityID:  false,
		PublicAssistance: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Test transaction rollback
	err := db.WithTransaction(ctx, func(ctx context.Context) error {
		// Create recipient within transaction
		err := recipientRepo.Create(ctx, recipient)
		if err != nil {
			return err
		}

		// Verify it exists within transaction
		_, err = recipientRepo.GetByID(ctx, recipient.ID)
		if err != nil {
			t.Errorf("GetByID() within transaction error = %v", err)
		}

		// Force rollback by returning an error
		return domain.ErrConstraint
	})

	if err != domain.ErrConstraint {
		t.Errorf("WithTransaction() error = %v, want %v", err, domain.ErrConstraint)
	}

	// Verify recipient does not exist after rollback
	_, err = recipientRepo.GetByID(ctx, recipient.ID)
	if err != domain.ErrNotFound {
		t.Errorf("GetByID() after rollback error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestRepositoryIntegration_TransactionCommit(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	recipient := &domain.Recipient{
		ID:               "transaction-commit-001",
		Name:             "コミットテスト太郎",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		HasDisabilityID:  false,
		PublicAssistance: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Test successful transaction
	err := db.WithTransaction(ctx, func(ctx context.Context) error {
		return recipientRepo.Create(ctx, recipient)
	})

	if err != nil {
		t.Errorf("WithTransaction() error = %v", err)
	}

	// Verify recipient exists after successful transaction
	retrieved, err := recipientRepo.GetByID(ctx, recipient.ID)
	if err != nil {
		t.Errorf("GetByID() after commit error = %v", err)
	}

	if retrieved == nil || retrieved.Name != recipient.Name {
		t.Error("Recipient not properly committed")
	}
}

func TestRepositoryIntegration_ConcurrentReads(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	// Pre-create recipients for concurrent reading
	recipients := make([]*domain.Recipient, 5)
	for i := 0; i < 5; i++ {
		recipients[i] = &domain.Recipient{
			ID:               domain.ID("concurrent-read-" + string(rune('0'+i))),
			Name:             "並行読み取りテスト太郎" + string(rune('0'+i)),
			Sex:              domain.SexMale,
			BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			HasDisabilityID:  false,
			PublicAssistance: false,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		err := recipientRepo.Create(ctx, recipients[i])
		if err != nil {
			t.Fatalf("Pre-create recipient %d error = %v", i, err)
		}
	}

	// Test concurrent reads (should not cause locking issues)
	done := make(chan error, 20)

	for i := 0; i < 20; i++ {
		go func(id int) {
			recipientID := recipients[id%5].ID
			expectedName := recipients[id%5].Name

			retrieved, err := recipientRepo.GetByID(ctx, recipientID)
			if err != nil {
				done <- err
				return
			}

			if retrieved.Name != expectedName {
				done <- fmt.Errorf("concurrent read data corruption: got %v, want %v", retrieved.Name, expectedName)
				return
			}

			done <- nil
		}(i)
	}

	// Wait for all goroutines and check for errors
	for i := 0; i < 20; i++ {
		if err := <-done; err != nil {
			t.Errorf("Concurrent read error: %v", err)
		}
	}
}

func TestRepositoryIntegration_SequentialWrites(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	// Test sequential writes (more realistic for SQLite)
	for i := 0; i < 5; i++ {
		recipient := &domain.Recipient{
			ID:               domain.ID("sequential-" + string(rune('0'+i))),
			Name:             "逐次書き込みテスト太郎" + string(rune('0'+i)),
			Sex:              domain.SexMale,
			BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			HasDisabilityID:  false,
			PublicAssistance: false,
			CreatedAt:        now.Add(time.Duration(i) * time.Millisecond),
			UpdatedAt:        now.Add(time.Duration(i) * time.Millisecond),
		}

		err := recipientRepo.Create(ctx, recipient)
		if err != nil {
			t.Errorf("Sequential create %d error = %v", i, err)
		}

		// Verify each creation immediately
		retrieved, err := recipientRepo.GetByID(ctx, recipient.ID)
		if err != nil {
			t.Errorf("Sequential retrieve %d error = %v", i, err)
		}

		if retrieved.Name != recipient.Name {
			t.Errorf("Sequential data integrity %d: got %v, want %v", i, retrieved.Name, recipient.Name)
		}
	}
}

func TestRepositoryIntegration_LargeDataHandling(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	// Create a recipient with large text fields
	largeText := ""
	for i := 0; i < 1000; i++ {
		largeText += "大容量テストデータ"
	}

	recipient := &domain.Recipient{
		ID:               "large-data-001",
		Name:             "大容量データテスト太郎",
		Kana:             "ダイヨウリョウデータテストタロウ",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		DisabilityName:   largeText, // Large encrypted field
		HasDisabilityID:  true,
		Grade:            "A1",
		Address:          largeText, // Another large encrypted field
		Phone:            "090-1234-5678",
		Email:            "large@example.com",
		PublicAssistance: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Test creation with large data
	start := time.Now()
	err := recipientRepo.Create(ctx, recipient)
	createDuration := time.Since(start)

	if err != nil {
		t.Errorf("Create() with large data error = %v", err)
	}

	// Test retrieval with large data
	start = time.Now()
	retrieved, err := recipientRepo.GetByID(ctx, recipient.ID)
	retrieveDuration := time.Since(start)

	if err != nil {
		t.Errorf("GetByID() with large data error = %v", err)
	}

	// Verify data integrity
	if retrieved.DisabilityName != largeText {
		t.Error("Large DisabilityName data corrupted")
	}
	if retrieved.Address != largeText {
		t.Error("Large Address data corrupted")
	}

	t.Logf("Large data performance - Create: %v, Retrieve: %v", createDuration, retrieveDuration)
	t.Logf("Large text length: %d characters", len(largeText))

	// Performance check (should complete within reasonable time)
	if createDuration > 5*time.Second {
		t.Errorf("Create with large data too slow: %v", createDuration)
	}
	if retrieveDuration > 5*time.Second {
		t.Errorf("Retrieve with large data too slow: %v", retrieveDuration)
	}
}

func TestRepositoryIntegration_EncryptionConsistency(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	recipient := &domain.Recipient{
		ID:               "consistency-001",
		Name:             "一貫性テスト太郎",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		HasDisabilityID:  false,
		PublicAssistance: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Create recipient
	err := recipientRepo.Create(ctx, recipient)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Retrieve multiple times to ensure consistency
	for i := 0; i < 10; i++ {
		retrieved, err := recipientRepo.GetByID(ctx, recipient.ID)
		if err != nil {
			t.Errorf("GetByID() iteration %d error = %v", i, err)
		}

		if retrieved.Name != recipient.Name {
			t.Errorf("Iteration %d: Name = %v, want %v", i, retrieved.Name, recipient.Name)
		}
	}

	// Update and verify consistency
	recipient.Name = "更新後一貫性テスト太郎"
	recipient.UpdatedAt = now.Add(time.Hour)

	err = recipientRepo.Update(ctx, recipient)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	// Verify update consistency
	for i := 0; i < 10; i++ {
		retrieved, err := recipientRepo.GetByID(ctx, recipient.ID)
		if err != nil {
			t.Errorf("GetByID() after update iteration %d error = %v", i, err)
		}

		if retrieved.Name != recipient.Name {
			t.Errorf("After update iteration %d: Name = %v, want %v", i, retrieved.Name, recipient.Name)
		}
	}
}

func BenchmarkRepositoryIntegration_EncryptionPerformance(b *testing.B) {
	tmpDir := b.TempDir()

	config := Config{
		Path:         filepath.Join(tmpDir, "benchmark.db"),
		MigrationDir: "../../../migrations",
	}

	db, err := NewDatabase(config)
	if err != nil {
		b.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	err = db.RunMigrations(ctx)
	if err != nil {
		b.Fatalf("failed to run migrations: %v", err)
	}

	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)
	now := time.Now().UTC().Truncate(time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recipient := &domain.Recipient{
			ID:               domain.ID(fmt.Sprintf("bench-%d-%d", now.UnixNano(), i)),
			Name:             "ベンチマーク太郎" + string(rune('0'+i%10)),
			Sex:              domain.SexMale,
			BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			HasDisabilityID:  i%2 == 0,
			PublicAssistance: i%3 == 0,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		err := recipientRepo.Create(ctx, recipient)
		if err != nil {
			b.Fatalf("Create() error = %v", err)
		}

		_, err = recipientRepo.GetByID(ctx, recipient.ID)
		if err != nil {
			b.Fatalf("GetByID() error = %v", err)
		}
	}
}
