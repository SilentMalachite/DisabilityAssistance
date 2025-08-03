package db

import (
	"context"
	"testing"
	"time"

	"shien-system/internal/domain"
)

func TestBenefitCertificateRepository_Create(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	// First create a recipient (foreign key dependency)
	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	recipient := &domain.Recipient{
		ID:               "cert-test-recipient-001",
		Name:             "受給者証テスト太郎",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		HasDisabilityID:  true,
		PublicAssistance: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err := recipientRepo.Create(ctx, recipient)
	if err != nil {
		t.Fatalf("Create recipient error = %v", err)
	}

	// Now test certificate creation
	certRepo, err := NewBenefitCertificateRepository(db)
	require.NoError(t, err)

	startDate := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC)

	certificate := &domain.BenefitCertificate{
		ID:                     "cert-create-001",
		RecipientID:            recipient.ID,
		StartDate:              startDate,
		EndDate:                endDate,
		Issuer:                 "東京都渋谷区",
		ServiceType:            "生活介護",
		MaxBenefitDaysPerMonth: 22,
		BenefitDetails:         "月額上限37,200円、食事提供加算あり",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	err = certRepo.Create(ctx, certificate)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	// Verify the certificate was created by retrieving it
	retrieved, err := certRepo.GetByID(ctx, certificate.ID)
	if err != nil {
		t.Errorf("GetByID() error = %v", err)
	}

	if retrieved == nil {
		t.Fatal("GetByID() returned nil")
	}

	// Compare all fields
	if retrieved.ID != certificate.ID {
		t.Errorf("ID = %v, want %v", retrieved.ID, certificate.ID)
	}
	if retrieved.RecipientID != certificate.RecipientID {
		t.Errorf("RecipientID = %v, want %v", retrieved.RecipientID, certificate.RecipientID)
	}
	if !retrieved.StartDate.Equal(certificate.StartDate) {
		t.Errorf("StartDate = %v, want %v", retrieved.StartDate, certificate.StartDate)
	}
	if !retrieved.EndDate.Equal(certificate.EndDate) {
		t.Errorf("EndDate = %v, want %v", retrieved.EndDate, certificate.EndDate)
	}
	if retrieved.Issuer != certificate.Issuer {
		t.Errorf("Issuer = %v, want %v", retrieved.Issuer, certificate.Issuer)
	}
	if retrieved.ServiceType != certificate.ServiceType {
		t.Errorf("ServiceType = %v, want %v", retrieved.ServiceType, certificate.ServiceType)
	}
	if retrieved.MaxBenefitDaysPerMonth != certificate.MaxBenefitDaysPerMonth {
		t.Errorf("MaxBenefitDaysPerMonth = %v, want %v", retrieved.MaxBenefitDaysPerMonth, certificate.MaxBenefitDaysPerMonth)
	}
	if retrieved.BenefitDetails != certificate.BenefitDetails {
		t.Errorf("BenefitDetails = %v, want %v", retrieved.BenefitDetails, certificate.BenefitDetails)
	}
}

func TestBenefitCertificateRepository_GetByRecipientID(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	// Create recipient
	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	recipient := &domain.Recipient{
		ID:               "cert-test-recipient-002",
		Name:             "受給者証複数テスト太郎",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		HasDisabilityID:  true,
		PublicAssistance: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err := recipientRepo.Create(ctx, recipient)
	if err != nil {
		t.Fatalf("Create recipient error = %v", err)
	}

	certRepo, err := NewBenefitCertificateRepository(db)
	require.NoError(t, err)

	// Create multiple certificates for the same recipient
	certificates := []*domain.BenefitCertificate{
		{
			ID:                     "cert-multi-001",
			RecipientID:            recipient.ID,
			StartDate:              time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC),
			EndDate:                time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC),
			Issuer:                 "東京都渋谷区",
			ServiceType:            "生活介護",
			MaxBenefitDaysPerMonth: 20,
			BenefitDetails:         "旧受給者証",
			CreatedAt:              now,
			UpdatedAt:              now,
		},
		{
			ID:                     "cert-multi-002",
			RecipientID:            recipient.ID,
			StartDate:              time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			EndDate:                time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC),
			Issuer:                 "東京都渋谷区",
			ServiceType:            "生活介護・就労継続支援B型",
			MaxBenefitDaysPerMonth: 22,
			BenefitDetails:         "新受給者証",
			CreatedAt:              now,
			UpdatedAt:              now,
		},
	}

	for _, cert := range certificates {
		err := certRepo.Create(ctx, cert)
		if err != nil {
			t.Fatalf("Create certificate error = %v", err)
		}
	}

	// Test GetByRecipientID
	retrievedCerts, err := certRepo.GetByRecipientID(ctx, recipient.ID)
	if err != nil {
		t.Errorf("GetByRecipientID() error = %v", err)
	}

	if len(retrievedCerts) != 2 {
		t.Errorf("GetByRecipientID() returned %d certificates, want 2", len(retrievedCerts))
	}

	// Verify certificates are for the correct recipient
	for _, cert := range retrievedCerts {
		if cert.RecipientID != recipient.ID {
			t.Errorf("Certificate %s has RecipientID %s, want %s", cert.ID, cert.RecipientID, recipient.ID)
		}
	}
}

func TestBenefitCertificateRepository_GetExpiringSoon(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	// Create recipient
	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	recipient := &domain.Recipient{
		ID:               "cert-test-recipient-003",
		Name:             "期限テスト太郎",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		HasDisabilityID:  true,
		PublicAssistance: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err := recipientRepo.Create(ctx, recipient)
	if err != nil {
		t.Fatalf("Create recipient error = %v", err)
	}

	certRepo, err := NewBenefitCertificateRepository(db)
	require.NoError(t, err)

	// Create certificates with different expiration dates
	certificates := []*domain.BenefitCertificate{
		{
			ID:                     "cert-expiring-001",
			RecipientID:            recipient.ID,
			StartDate:              now.Add(-365 * 24 * time.Hour),
			EndDate:                now.Add(15 * 24 * time.Hour), // Expires in 15 days
			Issuer:                 "東京都",
			ServiceType:            "生活介護",
			MaxBenefitDaysPerMonth: 22,
			BenefitDetails:         "期限間近の受給者証",
			CreatedAt:              now,
			UpdatedAt:              now,
		},
		{
			ID:                     "cert-not-expiring-001",
			RecipientID:            recipient.ID,
			StartDate:              now,
			EndDate:                now.Add(365 * 24 * time.Hour), // Expires in 1 year
			Issuer:                 "東京都",
			ServiceType:            "生活介護",
			MaxBenefitDaysPerMonth: 22,
			BenefitDetails:         "まだ期限の長い受給者証",
			CreatedAt:              now,
			UpdatedAt:              now,
		},
		{
			ID:                     "cert-expired-001",
			RecipientID:            recipient.ID,
			StartDate:              now.Add(-365 * 24 * time.Hour),
			EndDate:                now.Add(-1 * 24 * time.Hour), // Already expired
			Issuer:                 "東京都",
			ServiceType:            "生活介護",
			MaxBenefitDaysPerMonth: 22,
			BenefitDetails:         "既に期限切れの受給者証",
			CreatedAt:              now,
			UpdatedAt:              now,
		},
	}

	for _, cert := range certificates {
		err := certRepo.Create(ctx, cert)
		if err != nil {
			t.Fatalf("Create certificate error = %v", err)
		}
	}

	// Test GetExpiringSoon (within 30 days)
	expiring, err := certRepo.GetExpiringSoon(ctx, 30*24*time.Hour)
	if err != nil {
		t.Errorf("GetExpiringSoon() error = %v", err)
	}

	// Should return both the expiring (15 days) and expired (-1 day) certificates
	if len(expiring) != 2 {
		t.Errorf("GetExpiringSoon() returned %d certificates, want 2", len(expiring))
	}

	// Verify the returned certificates
	foundExpiring := false
	foundExpired := false
	for _, cert := range expiring {
		if cert.ID == "cert-expiring-001" {
			foundExpiring = true
		}
		if cert.ID == "cert-expired-001" {
			foundExpired = true
		}
		if cert.ID == "cert-not-expiring-001" {
			t.Error("GetExpiringSoon() should not return certificates expiring in > 30 days")
		}
	}

	if !foundExpiring {
		t.Error("GetExpiringSoon() did not return the expiring certificate")
	}
	if !foundExpired {
		t.Error("GetExpiringSoon() did not return the expired certificate")
	}
}

func TestBenefitCertificateRepository_GetActiveByRecipientID(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	// Create recipient
	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	recipient := &domain.Recipient{
		ID:               "cert-test-recipient-004",
		Name:             "有効性テスト太郎",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		HasDisabilityID:  true,
		PublicAssistance: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err := recipientRepo.Create(ctx, recipient)
	if err != nil {
		t.Fatalf("Create recipient error = %v", err)
	}

	certRepo, err := NewBenefitCertificateRepository(db)
	require.NoError(t, err)

	// Create certificates with different validity periods
	certificates := []*domain.BenefitCertificate{
		{
			ID:                     "cert-past-001",
			RecipientID:            recipient.ID,
			StartDate:              time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
			EndDate:                time.Date(2023, 3, 31, 0, 0, 0, 0, time.UTC),
			Issuer:                 "東京都",
			ServiceType:            "生活介護",
			MaxBenefitDaysPerMonth: 20,
			BenefitDetails:         "過去の受給者証",
			CreatedAt:              now,
			UpdatedAt:              now,
		},
		{
			ID:                     "cert-current-001",
			RecipientID:            recipient.ID,
			StartDate:              time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			EndDate:                time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC),
			Issuer:                 "東京都",
			ServiceType:            "生活介護",
			MaxBenefitDaysPerMonth: 22,
			BenefitDetails:         "現在有効な受給者証",
			CreatedAt:              now,
			UpdatedAt:              now,
		},
		{
			ID:                     "cert-future-001",
			RecipientID:            recipient.ID,
			StartDate:              time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
			EndDate:                time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
			Issuer:                 "東京都",
			ServiceType:            "生活介護",
			MaxBenefitDaysPerMonth: 25,
			BenefitDetails:         "将来の受給者証",
			CreatedAt:              now,
			UpdatedAt:              now,
		},
	}

	for _, cert := range certificates {
		err := certRepo.Create(ctx, cert)
		if err != nil {
			t.Fatalf("Create certificate error = %v", err)
		}
	}

	// Test GetActiveByRecipientID for current date (2024-2025 period)
	checkDate := time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)
	activeCert, err := certRepo.GetActiveByRecipientID(ctx, recipient.ID, checkDate)
	if err != nil {
		t.Errorf("GetActiveByRecipientID() error = %v", err)
	}

	if activeCert == nil {
		t.Fatal("GetActiveByRecipientID() returned nil")
	}

	if activeCert.ID != "cert-current-001" {
		t.Errorf("GetActiveByRecipientID() returned certificate %s, want cert-current-001", activeCert.ID)
	}

	// Test for a date where no certificate is active
	noActiveDate := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)
	noCert, err := certRepo.GetActiveByRecipientID(ctx, recipient.ID, noActiveDate)
	if err != domain.ErrNotFound {
		t.Errorf("GetActiveByRecipientID() for inactive date error = %v, want %v", err, domain.ErrNotFound)
	}
	if noCert != nil {
		t.Error("GetActiveByRecipientID() for inactive date should return nil")
	}
}

func TestBenefitCertificateRepository_Update(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	// Create recipient
	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	recipient := &domain.Recipient{
		ID:               "cert-test-recipient-005",
		Name:             "更新テスト太郎",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		HasDisabilityID:  true,
		PublicAssistance: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err := recipientRepo.Create(ctx, recipient)
	if err != nil {
		t.Fatalf("Create recipient error = %v", err)
	}

	certRepo, err := NewBenefitCertificateRepository(db)
	require.NoError(t, err)

	// Create initial certificate
	certificate := &domain.BenefitCertificate{
		ID:                     "cert-update-001",
		RecipientID:            recipient.ID,
		StartDate:              time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
		EndDate:                time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC),
		Issuer:                 "東京都渋谷区",
		ServiceType:            "生活介護",
		MaxBenefitDaysPerMonth: 20,
		BenefitDetails:         "更新前の内容",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	err = certRepo.Create(ctx, certificate)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Update the certificate
	updatedTime := now.Add(time.Hour)
	certificate.EndDate = time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC)
	certificate.ServiceType = "生活介護・就労継続支援B型"
	certificate.MaxBenefitDaysPerMonth = 25
	certificate.BenefitDetails = "更新後の内容・サービス追加"
	certificate.UpdatedAt = updatedTime

	err = certRepo.Update(ctx, certificate)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	// Verify the updates
	retrieved, err := certRepo.GetByID(ctx, certificate.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if !retrieved.EndDate.Equal(certificate.EndDate) {
		t.Errorf("EndDate = %v, want %v", retrieved.EndDate, certificate.EndDate)
	}
	if retrieved.ServiceType != certificate.ServiceType {
		t.Errorf("ServiceType = %v, want %v", retrieved.ServiceType, certificate.ServiceType)
	}
	if retrieved.MaxBenefitDaysPerMonth != certificate.MaxBenefitDaysPerMonth {
		t.Errorf("MaxBenefitDaysPerMonth = %v, want %v", retrieved.MaxBenefitDaysPerMonth, certificate.MaxBenefitDaysPerMonth)
	}
	if retrieved.BenefitDetails != certificate.BenefitDetails {
		t.Errorf("BenefitDetails = %v, want %v", retrieved.BenefitDetails, certificate.BenefitDetails)
	}
	if !retrieved.UpdatedAt.Equal(updatedTime) {
		t.Errorf("UpdatedAt = %v, want %v", retrieved.UpdatedAt, updatedTime)
	}
	// CreatedAt should remain unchanged
	if !retrieved.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", retrieved.CreatedAt, now)
	}
}

func TestBenefitCertificateRepository_Delete(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	// Create recipient
	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	recipient := &domain.Recipient{
		ID:               "cert-test-recipient-006",
		Name:             "削除テスト太郎",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		HasDisabilityID:  true,
		PublicAssistance: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err := recipientRepo.Create(ctx, recipient)
	if err != nil {
		t.Fatalf("Create recipient error = %v", err)
	}

	certRepo, err := NewBenefitCertificateRepository(db)
	require.NoError(t, err)

	certificate := &domain.BenefitCertificate{
		ID:                     "cert-delete-001",
		RecipientID:            recipient.ID,
		StartDate:              time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
		EndDate:                time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC),
		Issuer:                 "東京都",
		ServiceType:            "生活介護",
		MaxBenefitDaysPerMonth: 22,
		BenefitDetails:         "削除テスト用受給者証",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	err = certRepo.Create(ctx, certificate)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Delete the certificate
	err = certRepo.Delete(ctx, certificate.ID)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify the certificate was deleted
	retrieved, err := certRepo.GetByID(ctx, certificate.ID)
	if err != domain.ErrNotFound {
		t.Errorf("GetByID() after delete error = %v, want %v", err, domain.ErrNotFound)
	}
	if retrieved != nil {
		t.Error("GetByID() after delete should return nil")
	}
}

func TestBenefitCertificateRepository_List(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	// Create recipient
	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	recipients := []*domain.Recipient{
		{
			ID:               "cert-list-recipient-001",
			Name:             "一覧テスト太郎",
			Sex:              domain.SexMale,
			BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			HasDisabilityID:  true,
			PublicAssistance: false,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "cert-list-recipient-002",
			Name:             "一覧テスト花子",
			Sex:              domain.SexFemale,
			BirthDate:        time.Date(1985, 5, 15, 0, 0, 0, 0, time.UTC),
			HasDisabilityID:  true,
			PublicAssistance: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	for _, recipient := range recipients {
		err := recipientRepo.Create(ctx, recipient)
		if err != nil {
			t.Fatalf("Create recipient error = %v", err)
		}
	}

	certRepo, err := NewBenefitCertificateRepository(db)
	require.NoError(t, err)

	// Create multiple certificates
	certificates := []*domain.BenefitCertificate{
		{
			ID:                     "cert-list-001",
			RecipientID:            recipients[0].ID,
			StartDate:              time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			EndDate:                time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC),
			Issuer:                 "東京都",
			ServiceType:            "生活介護",
			MaxBenefitDaysPerMonth: 22,
			BenefitDetails:         "一覧テスト用証書1",
			CreatedAt:              now,
			UpdatedAt:              now,
		},
		{
			ID:                     "cert-list-002",
			RecipientID:            recipients[1].ID,
			StartDate:              time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			EndDate:                time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC),
			Issuer:                 "神奈川県横浜市",
			ServiceType:            "就労継続支援B型",
			MaxBenefitDaysPerMonth: 20,
			BenefitDetails:         "一覧テスト用証書2",
			CreatedAt:              now.Add(time.Minute),
			UpdatedAt:              now.Add(time.Minute),
		},
	}

	for _, cert := range certificates {
		err := certRepo.Create(ctx, cert)
		if err != nil {
			t.Fatalf("Create certificate error = %v", err)
		}
	}

	// Test listing all
	allCerts, err := certRepo.List(ctx, 10, 0)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}

	if len(allCerts) < 2 {
		t.Errorf("List() returned %d certificates, want at least 2", len(allCerts))
	}

	// Test pagination
	firstPage, err := certRepo.List(ctx, 1, 0)
	if err != nil {
		t.Errorf("List() first page error = %v", err)
	}

	secondPage, err := certRepo.List(ctx, 1, 1)
	if err != nil {
		t.Errorf("List() second page error = %v", err)
	}

	if len(firstPage) != 1 {
		t.Errorf("First page returned %d certificates, want 1", len(firstPage))
	}

	if len(secondPage) != 1 {
		t.Errorf("Second page returned %d certificates, want 1", len(secondPage))
	}

	// Verify no duplicates between pages
	if len(firstPage) > 0 && len(secondPage) > 0 && firstPage[0].ID == secondPage[0].ID {
		t.Error("Duplicate certificate found between pages")
	}
}

func TestBenefitCertificateRepository_Count(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	certRepo, err := NewBenefitCertificateRepository(db)
	require.NoError(t, err)
	ctx := context.Background()

	// Initial count should be 0
	count, err := certRepo.Count(ctx)
	if err != nil {
		t.Errorf("Count() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Initial count = %d, want 0", count)
	}

	// Note: We would need to create recipient first to test certificate creation
	// This is a basic count test - full test would require setting up recipients
}
