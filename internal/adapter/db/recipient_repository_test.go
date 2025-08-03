package db

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"shien-system/internal/domain"
)

func setupTestDatabase(t *testing.T) *Database {
	tmpDir := t.TempDir()

	config := Config{
		Path:         filepath.Join(tmpDir, "test.db"),
		MigrationDir: "../../../migrations",
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	ctx := context.Background()
	err = db.RunMigrations(ctx)
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

func TestRecipientRepository_Create(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	// This will be implemented
	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	recipient := &domain.Recipient{
		ID:               "recipient-001",
		Name:             "山田太郎",
		Kana:             "ヤマダタロウ",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
		DisabilityName:   "知的障害",
		HasDisabilityID:  true,
		Grade:            "B1",
		Address:          "東京都渋谷区1-1-1",
		Phone:            "03-1234-5678",
		Email:            "yamada@example.com",
		PublicAssistance: false,
		AdmissionDate:    &now,
		DischargeDate:    nil,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err := recipientRepo.Create(ctx, recipient)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	// Verify the recipient was created by retrieving it
	retrieved, err := recipientRepo.GetByID(ctx, recipient.ID)
	if err != nil {
		t.Errorf("GetByID() error = %v", err)
	}

	if retrieved == nil {
		t.Fatal("GetByID() returned nil")
	}

	// Compare all fields
	if retrieved.ID != recipient.ID {
		t.Errorf("ID = %v, want %v", retrieved.ID, recipient.ID)
	}
	if retrieved.Name != recipient.Name {
		t.Errorf("Name = %v, want %v", retrieved.Name, recipient.Name)
	}
	if retrieved.Kana != recipient.Kana {
		t.Errorf("Kana = %v, want %v", retrieved.Kana, recipient.Kana)
	}
	if retrieved.Sex != recipient.Sex {
		t.Errorf("Sex = %v, want %v", retrieved.Sex, recipient.Sex)
	}
	if !retrieved.BirthDate.Equal(recipient.BirthDate) {
		t.Errorf("BirthDate = %v, want %v", retrieved.BirthDate, recipient.BirthDate)
	}
	if retrieved.DisabilityName != recipient.DisabilityName {
		t.Errorf("DisabilityName = %v, want %v", retrieved.DisabilityName, recipient.DisabilityName)
	}
	if retrieved.HasDisabilityID != recipient.HasDisabilityID {
		t.Errorf("HasDisabilityID = %v, want %v", retrieved.HasDisabilityID, recipient.HasDisabilityID)
	}
	if retrieved.Grade != recipient.Grade {
		t.Errorf("Grade = %v, want %v", retrieved.Grade, recipient.Grade)
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
	if retrieved.PublicAssistance != recipient.PublicAssistance {
		t.Errorf("PublicAssistance = %v, want %v", retrieved.PublicAssistance, recipient.PublicAssistance)
	}
}

func TestRecipientRepository_CreateDuplicate(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	recipient := &domain.Recipient{
		ID:               "duplicate-001",
		Name:             "重複太郎",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		HasDisabilityID:  false,
		PublicAssistance: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// First creation should succeed
	err := recipientRepo.Create(ctx, recipient)
	if err != nil {
		t.Errorf("First Create() error = %v", err)
	}

	// Second creation with same ID should fail
	err = recipientRepo.Create(ctx, recipient)
	if err == nil {
		t.Error("Second Create() should return error for duplicate ID")
	}
}

func TestRecipientRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	recipient, err := recipientRepo.GetByID(ctx, "nonexistent-id")
	if err != domain.ErrNotFound {
		t.Errorf("GetByID() error = %v, want %v", err, domain.ErrNotFound)
	}
	if recipient != nil {
		t.Error("GetByID() should return nil for nonexistent ID")
	}
}

func TestRecipientRepository_Update(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	// Create initial recipient
	recipient := &domain.Recipient{
		ID:               "update-001",
		Name:             "更新前太郎",
		Sex:              domain.SexMale,
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

	// Update the recipient
	updatedTime := now.Add(time.Hour)
	recipient.Name = "更新後太郎"
	recipient.Kana = "コウシンゴタロウ"
	recipient.DisabilityName = "身体障害"
	recipient.HasDisabilityID = true
	recipient.Grade = "A1"
	recipient.UpdatedAt = updatedTime

	err = recipientRepo.Update(ctx, recipient)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	// Verify the updates
	retrieved, err := recipientRepo.GetByID(ctx, recipient.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if retrieved.Name != "更新後太郎" {
		t.Errorf("Name = %v, want %v", retrieved.Name, "更新後太郎")
	}
	if retrieved.Kana != "コウシンゴタロウ" {
		t.Errorf("Kana = %v, want %v", retrieved.Kana, "コウシンゴタロウ")
	}
	if retrieved.DisabilityName != "身体障害" {
		t.Errorf("DisabilityName = %v, want %v", retrieved.DisabilityName, "身体障害")
	}
	if retrieved.HasDisabilityID != true {
		t.Errorf("HasDisabilityID = %v, want %v", retrieved.HasDisabilityID, true)
	}
	if retrieved.Grade != "A1" {
		t.Errorf("Grade = %v, want %v", retrieved.Grade, "A1")
	}
	if !retrieved.UpdatedAt.Equal(updatedTime) {
		t.Errorf("UpdatedAt = %v, want %v", retrieved.UpdatedAt, updatedTime)
	}
	// CreatedAt should remain unchanged
	if !retrieved.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", retrieved.CreatedAt, now)
	}
}

func TestRecipientRepository_Delete(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	recipient := &domain.Recipient{
		ID:               "delete-001",
		Name:             "削除太郎",
		Sex:              domain.SexMale,
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

	// Delete the recipient
	err = recipientRepo.Delete(ctx, recipient.ID)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify the recipient was deleted
	retrieved, err := recipientRepo.GetByID(ctx, recipient.ID)
	if err != domain.ErrNotFound {
		t.Errorf("GetByID() after delete error = %v, want %v", err, domain.ErrNotFound)
	}
	if retrieved != nil {
		t.Error("GetByID() after delete should return nil")
	}
}

func TestRecipientRepository_List(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	// Create multiple recipients
	recipients := []*domain.Recipient{
		{
			ID:               "list-001",
			Name:             "一覧太郎",
			Sex:              domain.SexMale,
			BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			HasDisabilityID:  false,
			PublicAssistance: false,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "list-002",
			Name:             "一覧花子",
			Sex:              domain.SexFemale,
			BirthDate:        time.Date(1985, 5, 15, 0, 0, 0, 0, time.UTC),
			HasDisabilityID:  true,
			PublicAssistance: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "list-003",
			Name:             "一覧次郎",
			Sex:              domain.SexMale,
			BirthDate:        time.Date(1995, 12, 25, 0, 0, 0, 0, time.UTC),
			HasDisabilityID:  true,
			PublicAssistance: false,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	for _, recipient := range recipients {
		err := recipientRepo.Create(ctx, recipient)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Test listing all
	allRecipients, err := recipientRepo.List(ctx, 10, 0)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}

	// Should have at least our 3 recipients (plus admin user doesn't count)
	if len(allRecipients) < 3 {
		t.Errorf("List() returned %d recipients, want at least 3", len(allRecipients))
	}

	// Test pagination
	firstPage, err := recipientRepo.List(ctx, 2, 0)
	if err != nil {
		t.Errorf("List() first page error = %v", err)
	}

	secondPage, err := recipientRepo.List(ctx, 2, 2)
	if err != nil {
		t.Errorf("List() second page error = %v", err)
	}

	if len(firstPage) > 2 {
		t.Errorf("First page returned %d recipients, want at most 2", len(firstPage))
	}

	// Verify no duplicates between pages
	for _, first := range firstPage {
		for _, second := range secondPage {
			if first.ID == second.ID {
				t.Errorf("Duplicate recipient %s found between pages", first.ID)
			}
		}
	}
}

func TestRecipientRepository_Count(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	// Initial count should be 0
	count, err := recipientRepo.Count(ctx)
	if err != nil {
		t.Errorf("Count() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Initial count = %d, want 0", count)
	}

	// Create recipients and verify count increases
	now := time.Now().UTC().Truncate(time.Second)

	for i := 1; i <= 3; i++ {
		recipient := &domain.Recipient{
			ID:               domain.ID("count-" + string(rune('0'+i))),
			Name:             "カウント太郎" + string(rune('0'+i)),
			Sex:              domain.SexMale,
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

		count, err := recipientRepo.Count(ctx)
		if err != nil {
			t.Errorf("Count() error = %v", err)
		}
		if count != i {
			t.Errorf("Count after %d creates = %d, want %d", i, count, i)
		}
	}
}

func TestRecipientRepository_GetActive(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	discharged := now.Add(-24 * time.Hour)

	// Create active recipient (no discharge date)
	activeRecipient := &domain.Recipient{
		ID:               "active-001",
		Name:             "現役太郎",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		HasDisabilityID:  false,
		PublicAssistance: false,
		AdmissionDate:    &now,
		DischargeDate:    nil, // Still active
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Create discharged recipient
	dischargedRecipient := &domain.Recipient{
		ID:               "discharged-001",
		Name:             "退所太郎",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		HasDisabilityID:  false,
		PublicAssistance: false,
		AdmissionDate:    &now,
		DischargeDate:    &discharged, // Already discharged
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err := recipientRepo.Create(ctx, activeRecipient)
	if err != nil {
		t.Fatalf("Create active recipient error = %v", err)
	}

	err = recipientRepo.Create(ctx, dischargedRecipient)
	if err != nil {
		t.Fatalf("Create discharged recipient error = %v", err)
	}

	// Get active recipients only
	activeRecipients, err := recipientRepo.GetActive(ctx, 10, 0)
	if err != nil {
		t.Errorf("GetActive() error = %v", err)
	}

	// Should only return the active recipient
	found := false
	for _, recipient := range activeRecipients {
		if recipient.ID == activeRecipient.ID {
			found = true
		}
		if recipient.ID == dischargedRecipient.ID {
			t.Error("GetActive() returned discharged recipient")
		}
	}

	if !found {
		t.Error("GetActive() did not return active recipient")
	}

	// Test active count
	activeCount, err := recipientRepo.CountActive(ctx)
	if err != nil {
		t.Errorf("CountActive() error = %v", err)
	}

	if activeCount < 1 {
		t.Errorf("CountActive() = %d, want at least 1", activeCount)
	}
}

// Note: Search functionality will be implemented with encrypted fields
// This is a placeholder for future implementation
func TestRecipientRepository_Search_Placeholder(t *testing.T) {
	t.Skip("Search functionality with encrypted fields to be implemented")

	// This test will be implemented when we decide on the search approach
	// Options:
	// 1. Store searchable hash of names (less secure but searchable)
	// 2. Decrypt and search in memory (more secure but slower)
	// 3. Use specialized encrypted search techniques
}
