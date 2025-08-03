package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"shien-system/internal/domain"
)

func TestPhase2Integration_CompleteWorkflow(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	// Initialize all repositories
	staffRepo := NewStaffRepository(db)
	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)
	assignmentRepo := NewStaffAssignmentRepository(db)
	certRepo, err := NewBenefitCertificateRepository(db)
	require.NoError(t, err)
	auditRepo := NewAuditLogRepository(db)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	// Step 1: Create staff member
	staff := &domain.Staff{
		ID:        "integration-staff-001",
		Name:      "統合テスト担当者",
		Role:      domain.RoleStaff,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := staffRepo.Create(ctx, staff)
	if err != nil {
		t.Fatalf("Create staff error = %v", err)
	}

	// Log staff creation
	auditLog1 := &domain.AuditLog{
		ID:      "audit-integration-001",
		ActorID: "admin-001", // Initial admin from migration
		Action:  "CREATE",
		Target:  "staff:" + staff.ID,
		At:      now,
		IP:      "127.0.0.1",
		Details: "職員を作成しました",
	}

	err = auditRepo.Create(ctx, auditLog1)
	if err != nil {
		t.Fatalf("Create audit log error = %v", err)
	}

	// Step 2: Create recipient
	recipient := &domain.Recipient{
		ID:               "integration-recipient-001",
		Name:             "統合テスト利用者",
		Kana:             "トウゴウテストリヨウシャ",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
		DisabilityName:   "知的障害",
		HasDisabilityID:  true,
		Grade:            "B1",
		Address:          "東京都渋谷区1-1-1",
		Phone:            "03-1234-5678",
		Email:            "integration@example.com",
		PublicAssistance: false,
		AdmissionDate:    &now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err = recipientRepo.Create(ctx, recipient)
	if err != nil {
		t.Fatalf("Create recipient error = %v", err)
	}

	// Log recipient creation
	auditLog2 := &domain.AuditLog{
		ID:      "audit-integration-002",
		ActorID: staff.ID,
		Action:  "CREATE",
		Target:  "recipient:" + recipient.ID,
		At:      now.Add(time.Minute),
		IP:      "127.0.0.1",
		Details: "利用者を作成しました",
	}

	err = auditRepo.Create(ctx, auditLog2)
	if err != nil {
		t.Fatalf("Create audit log 2 error = %v", err)
	}

	// Step 3: Create staff assignment
	assignment := &domain.StaffAssignment{
		ID:           "integration-assignment-001",
		RecipientID:  recipient.ID,
		StaffID:      staff.ID,
		Role:         "主担当",
		AssignedAt:   now.Add(2 * time.Minute),
		UnassignedAt: nil,
	}

	err = assignmentRepo.Create(ctx, assignment)
	if err != nil {
		t.Fatalf("Create assignment error = %v", err)
	}

	// Log assignment creation
	auditLog3 := &domain.AuditLog{
		ID:      "audit-integration-003",
		ActorID: staff.ID,
		Action:  "ASSIGN",
		Target:  "assignment:" + assignment.ID,
		At:      now.Add(2 * time.Minute),
		IP:      "127.0.0.1",
		Details: "担当者を割り当てました",
	}

	err = auditRepo.Create(ctx, auditLog3)
	if err != nil {
		t.Fatalf("Create audit log 3 error = %v", err)
	}

	// Step 4: Create benefit certificate
	startDate := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC)

	certificate := &domain.BenefitCertificate{
		ID:                     "integration-cert-001",
		RecipientID:            recipient.ID,
		StartDate:              startDate,
		EndDate:                endDate,
		Issuer:                 "東京都渋谷区",
		ServiceType:            "生活介護",
		MaxBenefitDaysPerMonth: 22,
		BenefitDetails:         "月額上限37,200円、食事提供加算あり",
		CreatedAt:              now.Add(3 * time.Minute),
		UpdatedAt:              now.Add(3 * time.Minute),
	}

	err = certRepo.Create(ctx, certificate)
	if err != nil {
		t.Fatalf("Create certificate error = %v", err)
	}

	// Log certificate creation
	auditLog4 := &domain.AuditLog{
		ID:      "audit-integration-004",
		ActorID: staff.ID,
		Action:  "CREATE",
		Target:  "certificate:" + certificate.ID,
		At:      now.Add(3 * time.Minute),
		IP:      "127.0.0.1",
		Details: "受給者証を作成しました",
	}

	err = auditRepo.Create(ctx, auditLog4)
	if err != nil {
		t.Fatalf("Create audit log 4 error = %v", err)
	}

	// Step 5: Verify all data was created correctly

	// Verify staff
	retrievedStaff, err := staffRepo.GetByID(ctx, staff.ID)
	if err != nil {
		t.Errorf("GetByID staff error = %v", err)
	}
	if retrievedStaff.Name != staff.Name {
		t.Errorf("Staff name = %v, want %v", retrievedStaff.Name, staff.Name)
	}

	// Verify recipient (including encrypted fields)
	retrievedRecipient, err := recipientRepo.GetByID(ctx, recipient.ID)
	if err != nil {
		t.Errorf("GetByID recipient error = %v", err)
	}
	if retrievedRecipient.Name != recipient.Name {
		t.Errorf("Recipient name = %v, want %v", retrievedRecipient.Name, recipient.Name)
	}
	if retrievedRecipient.DisabilityName != recipient.DisabilityName {
		t.Errorf("Recipient disability = %v, want %v", retrievedRecipient.DisabilityName, recipient.DisabilityName)
	}

	// Verify assignment
	retrievedAssignment, err := assignmentRepo.GetByID(ctx, assignment.ID)
	if err != nil {
		t.Errorf("GetByID assignment error = %v", err)
	}
	if retrievedAssignment.Role != assignment.Role {
		t.Errorf("Assignment role = %v, want %v", retrievedAssignment.Role, assignment.Role)
	}

	// Verify certificate (including encrypted fields)
	retrievedCert, err := certRepo.GetByID(ctx, certificate.ID)
	if err != nil {
		t.Errorf("GetByID certificate error = %v", err)
	}
	if retrievedCert.Issuer != certificate.Issuer {
		t.Errorf("Certificate issuer = %v, want %v", retrievedCert.Issuer, certificate.Issuer)
	}
	if retrievedCert.ServiceType != certificate.ServiceType {
		t.Errorf("Certificate service type = %v, want %v", retrievedCert.ServiceType, certificate.ServiceType)
	}

	// Verify audit logs
	allLogs, err := auditRepo.List(ctx, 10, 0)
	if err != nil {
		t.Errorf("List audit logs error = %v", err)
	}
	if len(allLogs) < 4 {
		t.Errorf("Expected at least 4 audit logs, got %d", len(allLogs))
	}

	// Step 6: Test cross-repository queries

	// Get assignments by staff
	staffAssignments, err := assignmentRepo.GetByStaffID(ctx, staff.ID)
	if err != nil {
		t.Errorf("GetByStaffID error = %v", err)
	}
	if len(staffAssignments) != 1 {
		t.Errorf("Expected 1 assignment for staff, got %d", len(staffAssignments))
	}

	// Get certificates by recipient
	recipientCerts, err := certRepo.GetByRecipientID(ctx, recipient.ID)
	if err != nil {
		t.Errorf("GetByRecipientID certificates error = %v", err)
	}
	if len(recipientCerts) != 1 {
		t.Errorf("Expected 1 certificate for recipient, got %d", len(recipientCerts))
	}

	// Get audit logs by actor
	staffLogs, err := auditRepo.GetByActorID(ctx, staff.ID, 10, 0)
	if err != nil {
		t.Errorf("GetByActorID audit logs error = %v", err)
	}
	if len(staffLogs) < 3 {
		t.Errorf("Expected at least 3 audit logs for staff, got %d", len(staffLogs))
	}
}

func TestPhase2Integration_TransactionConsistency(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	staffRepo := NewStaffRepository(db)
	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)
	assignmentRepo := NewStaffAssignmentRepository(db)
	auditRepo := NewAuditLogRepository(db)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	// Test transaction rollback scenario
	err := db.WithTransaction(ctx, func(ctx context.Context) error {
		// Create staff
		staff := &domain.Staff{
			ID:        "transaction-staff-001",
			Name:      "トランザクションテスト職員",
			Role:      domain.RoleStaff,
			CreatedAt: now,
			UpdatedAt: now,
		}

		err := staffRepo.Create(ctx, staff)
		if err != nil {
			return err
		}

		// Create recipient
		recipient := &domain.Recipient{
			ID:               "transaction-recipient-001",
			Name:             "トランザクションテスト利用者",
			Sex:              domain.SexMale,
			BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			HasDisabilityID:  false,
			PublicAssistance: false,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		err = recipientRepo.Create(ctx, recipient)
		if err != nil {
			return err
		}

		// Create assignment
		assignment := &domain.StaffAssignment{
			ID:           "transaction-assignment-001",
			RecipientID:  recipient.ID,
			StaffID:      staff.ID,
			Role:         "テスト担当",
			AssignedAt:   now,
			UnassignedAt: nil,
		}

		err = assignmentRepo.Create(ctx, assignment)
		if err != nil {
			return err
		}

		// Create audit log
		auditLog := &domain.AuditLog{
			ID:      "transaction-audit-001",
			ActorID: staff.ID,
			Action:  "TEST",
			Target:  "transaction",
			At:      now,
			IP:      "127.0.0.1",
			Details: "トランザクションテスト",
		}

		err = auditRepo.Create(ctx, auditLog)
		if err != nil {
			return err
		}

		// Force rollback
		return domain.ErrConstraint
	})

	if err != domain.ErrConstraint {
		t.Errorf("WithTransaction error = %v, want %v", err, domain.ErrConstraint)
	}

	// Verify nothing was committed
	_, err = staffRepo.GetByID(ctx, "transaction-staff-001")
	if err != domain.ErrNotFound {
		t.Error("Staff should not exist after transaction rollback")
	}

	_, err = recipientRepo.GetByID(ctx, "transaction-recipient-001")
	if err != domain.ErrNotFound {
		t.Error("Recipient should not exist after transaction rollback")
	}

	_, err = assignmentRepo.GetByID(ctx, "transaction-assignment-001")
	if err != domain.ErrNotFound {
		t.Error("Assignment should not exist after transaction rollback")
	}

	_, err = auditRepo.GetByID(ctx, "transaction-audit-001")
	if err != domain.ErrNotFound {
		t.Error("Audit log should not exist after transaction rollback")
	}
}

func TestPhase2Integration_ForeignKeyConstraints(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	staffRepo := NewStaffRepository(db)
	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)
	assignmentRepo := NewStaffAssignmentRepository(db)
	certRepo, err := NewBenefitCertificateRepository(db)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	// Test foreign key constraint enforcement

	// Try to create assignment with non-existent staff
	assignment := &domain.StaffAssignment{
		ID:           "fk-test-assignment-001",
		RecipientID:  "nonexistent-recipient",
		StaffID:      "nonexistent-staff",
		Role:         "テスト",
		AssignedAt:   now,
		UnassignedAt: nil,
	}

	err := assignmentRepo.Create(ctx, assignment)
	if err == nil {
		t.Error("Assignment creation with non-existent staff/recipient should fail")
	}

	// Try to create certificate with non-existent recipient
	certificate := &domain.BenefitCertificate{
		ID:                     "fk-test-cert-001",
		RecipientID:            "nonexistent-recipient",
		StartDate:              time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:                time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		Issuer:                 "テスト市",
		ServiceType:            "テストサービス",
		MaxBenefitDaysPerMonth: 20,
		BenefitDetails:         "テスト用",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	err = certRepo.Create(ctx, certificate)
	if err == nil {
		t.Error("Certificate creation with non-existent recipient should fail")
	}

	// Create valid entities for cascade delete test
	staff := &domain.Staff{
		ID:        "fk-test-staff-001",
		Name:      "外部キーテスト職員",
		Role:      domain.RoleStaff,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = staffRepo.Create(ctx, staff)
	if err != nil {
		t.Fatalf("Create staff for FK test error = %v", err)
	}

	recipient := &domain.Recipient{
		ID:               "fk-test-recipient-001",
		Name:             "外部キーテスト利用者",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		HasDisabilityID:  false,
		PublicAssistance: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err = recipientRepo.Create(ctx, recipient)
	if err != nil {
		t.Fatalf("Create recipient for FK test error = %v", err)
	}

	// Create dependent records
	validAssignment := &domain.StaffAssignment{
		ID:           "fk-test-assignment-002",
		RecipientID:  recipient.ID,
		StaffID:      staff.ID,
		Role:         "テスト担当",
		AssignedAt:   now,
		UnassignedAt: nil,
	}

	err = assignmentRepo.Create(ctx, validAssignment)
	if err != nil {
		t.Fatalf("Create valid assignment error = %v", err)
	}

	validCertificate := &domain.BenefitCertificate{
		ID:                     "fk-test-cert-002",
		RecipientID:            recipient.ID,
		StartDate:              time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:                time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		Issuer:                 "テスト市",
		ServiceType:            "テストサービス",
		MaxBenefitDaysPerMonth: 20,
		BenefitDetails:         "テスト用",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	err = certRepo.Create(ctx, validCertificate)
	if err != nil {
		t.Fatalf("Create valid certificate error = %v", err)
	}

	// Test cascade delete by deleting recipient
	err = recipientRepo.Delete(ctx, recipient.ID)
	if err != nil {
		t.Errorf("Delete recipient error = %v", err)
	}

	// Verify dependent records were cascade deleted
	_, err = assignmentRepo.GetByID(ctx, validAssignment.ID)
	if err != domain.ErrNotFound {
		t.Error("Assignment should be cascade deleted when recipient is deleted")
	}

	_, err = certRepo.GetByID(ctx, validCertificate.ID)
	if err != domain.ErrNotFound {
		t.Error("Certificate should be cascade deleted when recipient is deleted")
	}
}

func TestPhase2Integration_EncryptionConsistency(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)
	certRepo, err := NewBenefitCertificateRepository(db)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	// Test encryption consistency across operations
	recipient := &domain.Recipient{
		ID:               "encryption-test-recipient-001",
		Name:             "暗号化テスト利用者",
		Kana:             "アンゴウカテストリヨウシャ",
		Sex:              domain.SexFemale,
		BirthDate:        time.Date(1985, 3, 15, 0, 0, 0, 0, time.UTC),
		DisabilityName:   "身体障害・視覚障害",
		HasDisabilityID:  true,
		Grade:            "A2",
		Address:          "〒100-0001 東京都千代田区千代田1-1 皇居内",
		Phone:            "090-9876-5432",
		Email:            "encryption-test@example.jp",
		PublicAssistance: true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err := recipientRepo.Create(ctx, recipient)
	if err != nil {
		t.Fatalf("Create encrypted recipient error = %v", err)
	}

	certificate := &domain.BenefitCertificate{
		ID:                     "encryption-test-cert-001",
		RecipientID:            recipient.ID,
		StartDate:              time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
		EndDate:                time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC),
		Issuer:                 "神奈川県横浜市青葉区",
		ServiceType:            "生活介護・就労継続支援A型・共同生活援助",
		MaxBenefitDaysPerMonth: 25,
		BenefitDetails:         "月額上限46,800円、食事提供加算・送迎加算・入浴サービス加算あり",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	err = certRepo.Create(ctx, certificate)
	if err != nil {
		t.Fatalf("Create encrypted certificate error = %v", err)
	}

	// Verify encryption round-trips multiple times
	for i := 0; i < 5; i++ {
		retrievedRecipient, err := recipientRepo.GetByID(ctx, recipient.ID)
		if err != nil {
			t.Errorf("GetByID recipient iteration %d error = %v", i, err)
		}

		// Check all encrypted fields
		if retrievedRecipient.Name != recipient.Name {
			t.Errorf("Iteration %d: Name = %v, want %v", i, retrievedRecipient.Name, recipient.Name)
		}
		if retrievedRecipient.Kana != recipient.Kana {
			t.Errorf("Iteration %d: Kana = %v, want %v", i, retrievedRecipient.Kana, recipient.Kana)
		}
		if retrievedRecipient.DisabilityName != recipient.DisabilityName {
			t.Errorf("Iteration %d: DisabilityName = %v, want %v", i, retrievedRecipient.DisabilityName, recipient.DisabilityName)
		}
		if retrievedRecipient.Address != recipient.Address {
			t.Errorf("Iteration %d: Address = %v, want %v", i, retrievedRecipient.Address, recipient.Address)
		}

		retrievedCert, err := certRepo.GetByID(ctx, certificate.ID)
		if err != nil {
			t.Errorf("GetByID certificate iteration %d error = %v", i, err)
		}

		// Check all encrypted fields
		if retrievedCert.Issuer != certificate.Issuer {
			t.Errorf("Iteration %d: Issuer = %v, want %v", i, retrievedCert.Issuer, certificate.Issuer)
		}
		if retrievedCert.ServiceType != certificate.ServiceType {
			t.Errorf("Iteration %d: ServiceType = %v, want %v", i, retrievedCert.ServiceType, certificate.ServiceType)
		}
		if retrievedCert.BenefitDetails != certificate.BenefitDetails {
			t.Errorf("Iteration %d: BenefitDetails = %v, want %v", i, retrievedCert.BenefitDetails, certificate.BenefitDetails)
		}
	}
}

func BenchmarkPhase2Integration_CompleteWorkflow(b *testing.B) {
	tmpDir := b.TempDir()

	config := Config{
		Path:         tmpDir + "/benchmark.db",
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

	staffRepo := NewStaffRepository(db)
	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)
	assignmentRepo := NewStaffAssignmentRepository(db)
	certRepo, err := NewBenefitCertificateRepository(db)
	require.NoError(t, err)
	auditRepo := NewAuditLogRepository(db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		now := time.Now().UTC().Truncate(time.Second)

		// Create complete workflow
		staff := &domain.Staff{
			ID:        domain.ID(fmt.Sprintf("bench-staff-%d", i)),
			Name:      fmt.Sprintf("ベンチマーク職員%d", i),
			Role:      domain.RoleStaff,
			CreatedAt: now,
			UpdatedAt: now,
		}

		err := staffRepo.Create(ctx, staff)
		if err != nil {
			b.Fatalf("Create staff error = %v", err)
		}

		recipient := &domain.Recipient{
			ID:               domain.ID(fmt.Sprintf("bench-recipient-%d", i)),
			Name:             fmt.Sprintf("ベンチマーク利用者%d", i),
			Sex:              domain.SexMale,
			BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			HasDisabilityID:  true,
			PublicAssistance: false,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		err = recipientRepo.Create(ctx, recipient)
		if err != nil {
			b.Fatalf("Create recipient error = %v", err)
		}

		assignment := &domain.StaffAssignment{
			ID:           domain.ID(fmt.Sprintf("bench-assignment-%d", i)),
			RecipientID:  recipient.ID,
			StaffID:      staff.ID,
			Role:         "担当",
			AssignedAt:   now,
			UnassignedAt: nil,
		}

		err = assignmentRepo.Create(ctx, assignment)
		if err != nil {
			b.Fatalf("Create assignment error = %v", err)
		}

		certificate := &domain.BenefitCertificate{
			ID:                     domain.ID(fmt.Sprintf("bench-cert-%d", i)),
			RecipientID:            recipient.ID,
			StartDate:              time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			EndDate:                time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC),
			Issuer:                 "ベンチマーク市",
			ServiceType:            "生活介護",
			MaxBenefitDaysPerMonth: 22,
			BenefitDetails:         "ベンチマークテスト",
			CreatedAt:              now,
			UpdatedAt:              now,
		}

		err = certRepo.Create(ctx, certificate)
		if err != nil {
			b.Fatalf("Create certificate error = %v", err)
		}

		auditLog := &domain.AuditLog{
			ID:      domain.ID(fmt.Sprintf("bench-audit-%d", i)),
			ActorID: staff.ID,
			Action:  "BENCHMARK",
			Target:  "workflow",
			At:      now,
			IP:      "127.0.0.1",
			Details: "ベンチマークテスト実行",
		}

		err = auditRepo.Create(ctx, auditLog)
		if err != nil {
			b.Fatalf("Create audit log error = %v", err)
		}
	}
}
