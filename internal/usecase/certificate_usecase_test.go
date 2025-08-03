package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"shien-system/internal/domain"
)

// Mock certificate repository
type mockCertificateRepository struct {
	certificates map[domain.ID]*domain.BenefitCertificate
	nextError    error
}

func (m *mockCertificateRepository) Create(ctx context.Context, cert *domain.BenefitCertificate) error {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return err
	}
	if m.certificates == nil {
		m.certificates = make(map[domain.ID]*domain.BenefitCertificate)
	}
	m.certificates[cert.ID] = cert
	return nil
}

func (m *mockCertificateRepository) GetByID(ctx context.Context, id domain.ID) (*domain.BenefitCertificate, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return nil, err
	}
	cert, exists := m.certificates[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	return cert, nil
}

func (m *mockCertificateRepository) Update(ctx context.Context, cert *domain.BenefitCertificate) error {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return err
	}
	if _, exists := m.certificates[cert.ID]; !exists {
		return domain.ErrNotFound
	}
	m.certificates[cert.ID] = cert
	return nil
}

func (m *mockCertificateRepository) Delete(ctx context.Context, id domain.ID) error {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return err
	}
	if _, exists := m.certificates[id]; !exists {
		return domain.ErrNotFound
	}
	delete(m.certificates, id)
	return nil
}

func (m *mockCertificateRepository) GetByRecipientID(ctx context.Context, recipientID domain.ID) ([]*domain.BenefitCertificate, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return nil, err
	}
	var certificates []*domain.BenefitCertificate
	for _, cert := range m.certificates {
		if cert.RecipientID == recipientID {
			certificates = append(certificates, cert)
		}
	}
	return certificates, nil
}

func (m *mockCertificateRepository) GetExpiringSoon(ctx context.Context, within time.Duration) ([]*domain.BenefitCertificate, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return nil, err
	}
	var expiring []*domain.BenefitCertificate
	cutoff := time.Now().UTC().Add(within)

	for _, cert := range m.certificates {
		if cert.EndDate.Before(cutoff) || cert.EndDate.Equal(cutoff) {
			expiring = append(expiring, cert)
		}
	}
	return expiring, nil
}

func (m *mockCertificateRepository) GetActiveByRecipientID(ctx context.Context, recipientID domain.ID, asOf time.Time) (*domain.BenefitCertificate, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return nil, err
	}
	for _, cert := range m.certificates {
		if cert.RecipientID == recipientID &&
			!cert.StartDate.After(asOf) &&
			cert.EndDate.After(asOf) {
			return cert, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockCertificateRepository) List(ctx context.Context, limit, offset int) ([]*domain.BenefitCertificate, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return nil, err
	}
	var certificates []*domain.BenefitCertificate
	for _, cert := range m.certificates {
		certificates = append(certificates, cert)
	}
	// Simple pagination simulation
	start := offset
	if start > len(certificates) {
		return []*domain.BenefitCertificate{}, nil
	}
	end := start + limit
	if end > len(certificates) {
		end = len(certificates)
	}
	return certificates[start:end], nil
}

func (m *mockCertificateRepository) Count(ctx context.Context) (int, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return 0, err
	}
	return len(m.certificates), nil
}

func TestCertificateUseCase_CreateCertificate(t *testing.T) {
	mockCertRepo := &mockCertificateRepository{}
	mockRecipientRepo := &mockRecipientRepository{
		recipients: map[domain.ID]*domain.Recipient{
			"recipient-001": {
				ID:   "recipient-001",
				Name: "テスト利用者",
			},
		},
	}
	mockStaffRepo := &mockStaffRepository{
		staff: map[domain.ID]*domain.Staff{
			"staff-001": {
				ID:   "staff-001",
				Name: "テスト職員",
				Role: domain.RoleStaff,
			},
		},
	}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewCertificateUseCase(mockCertRepo, mockRecipientRepo, mockStaffRepo, mockAuditRepo)

	ctx := context.Background()

	req := CreateCertificateRequest{
		RecipientID:            "recipient-001",
		StartDate:              time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
		EndDate:                time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC),
		Issuer:                 "東京都渋谷区",
		ServiceType:            "生活介護",
		MaxBenefitDaysPerMonth: 22,
		BenefitDetails:         "月額上限37,200円",
		ActorID:                "staff-001",
	}

	certificate, err := usecase.CreateCertificate(ctx, req)
	if err != nil {
		t.Errorf("CreateCertificate() error = %v", err)
	}

	if certificate == nil {
		t.Fatal("CreateCertificate() returned nil certificate")
	}

	// Verify certificate fields
	if certificate.RecipientID != req.RecipientID {
		t.Errorf("RecipientID = %v, want %v", certificate.RecipientID, req.RecipientID)
	}
	if certificate.Issuer != req.Issuer {
		t.Errorf("Issuer = %v, want %v", certificate.Issuer, req.Issuer)
	}
	if certificate.ServiceType != req.ServiceType {
		t.Errorf("ServiceType = %v, want %v", certificate.ServiceType, req.ServiceType)
	}

	// Verify audit log was created
	if len(mockAuditRepo.logs) != 1 {
		t.Errorf("Expected 1 audit log, got %d", len(mockAuditRepo.logs))
	}
}

func TestCertificateUseCase_CreateCertificate_ValidationError(t *testing.T) {
	mockCertRepo := &mockCertificateRepository{}
	mockRecipientRepo := &mockRecipientRepository{}
	mockStaffRepo := &mockStaffRepository{
		staff: map[domain.ID]*domain.Staff{
			"staff-001": {
				ID:   "staff-001",
				Name: "テスト職員",
				Role: domain.RoleStaff,
			},
		},
	}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewCertificateUseCase(mockCertRepo, mockRecipientRepo, mockStaffRepo, mockAuditRepo)

	ctx := context.Background()

	// Test with empty recipient ID
	req := CreateCertificateRequest{
		RecipientID: "", // Invalid: empty recipient ID
		StartDate:   time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC),
		ActorID:     "staff-001",
	}

	_, err := usecase.CreateCertificate(ctx, req)
	if err == nil {
		t.Error("CreateCertificate() should return validation error for empty recipient ID")
	}

	var useCaseErr *UseCaseError
	if !errors.As(err, &useCaseErr) || useCaseErr.Code != "VALIDATION_FAILED" {
		t.Errorf("Expected VALIDATION_FAILED error, got %v", err)
	}
}

func TestCertificateUseCase_GetCertificate(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	existingCert := &domain.BenefitCertificate{
		ID:                     "cert-001",
		RecipientID:            "recipient-001",
		StartDate:              time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
		EndDate:                time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC),
		Issuer:                 "テスト市",
		ServiceType:            "生活介護",
		MaxBenefitDaysPerMonth: 22,
		BenefitDetails:         "テスト用受給者証",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	mockCertRepo := &mockCertificateRepository{
		certificates: map[domain.ID]*domain.BenefitCertificate{
			"cert-001": existingCert,
		},
	}
	mockRecipientRepo := &mockRecipientRepository{}
	mockStaffRepo := &mockStaffRepository{}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewCertificateUseCase(mockCertRepo, mockRecipientRepo, mockStaffRepo, mockAuditRepo)

	ctx := context.Background()

	certificate, err := usecase.GetCertificate(ctx, "cert-001")
	if err != nil {
		t.Errorf("GetCertificate() error = %v", err)
	}

	if certificate == nil {
		t.Fatal("GetCertificate() returned nil")
	}

	if certificate.ID != "cert-001" {
		t.Errorf("ID = %v, want cert-001", certificate.ID)
	}
	if certificate.ServiceType != "生活介護" {
		t.Errorf("ServiceType = %v, want 生活介護", certificate.ServiceType)
	}
}

func TestCertificateUseCase_GetExpiringSoon(t *testing.T) {
	now := time.Now().UTC()

	// Create certificates with different expiration dates
	soonToExpire := &domain.BenefitCertificate{
		ID:          "cert-soon",
		RecipientID: "recipient-001",
		StartDate:   now.Add(-365 * 24 * time.Hour), // Started 1 year ago
		EndDate:     now.Add(15 * 24 * time.Hour),   // Expires in 15 days
		Issuer:      "テスト市",
		ServiceType: "生活介護",
	}

	notExpiring := &domain.BenefitCertificate{
		ID:          "cert-future",
		RecipientID: "recipient-002",
		StartDate:   now.Add(-30 * 24 * time.Hour), // Started 30 days ago
		EndDate:     now.Add(100 * 24 * time.Hour), // Expires in 100 days
		Issuer:      "テスト市",
		ServiceType: "就労継続支援A型",
	}

	mockCertRepo := &mockCertificateRepository{
		certificates: map[domain.ID]*domain.BenefitCertificate{
			"cert-soon":   soonToExpire,
			"cert-future": notExpiring,
		},
	}
	mockRecipientRepo := &mockRecipientRepository{}
	mockStaffRepo := &mockStaffRepository{}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewCertificateUseCase(mockCertRepo, mockRecipientRepo, mockStaffRepo, mockAuditRepo)

	ctx := context.Background()

	// Get certificates expiring within 30 days
	expiring, err := usecase.GetExpiringSoon(ctx, 30)
	if err != nil {
		t.Errorf("GetExpiringSoon() error = %v", err)
	}

	if len(expiring) != 1 {
		t.Errorf("GetExpiringSoon() returned %d certificates, want 1", len(expiring))
	}

	if len(expiring) > 0 && expiring[0].ID != "cert-soon" {
		t.Errorf("GetExpiringSoon() returned certificate %s, want cert-soon", expiring[0].ID)
	}
}

func TestCertificateUseCase_ValidateCertificate(t *testing.T) {
	now := time.Now().UTC()
	validCert := &domain.BenefitCertificate{
		ID:          "cert-valid",
		RecipientID: "recipient-001",
		StartDate:   now.Add(-30 * 24 * time.Hour), // Started 30 days ago
		EndDate:     now.Add(30 * 24 * time.Hour),  // Expires in 30 days
		Issuer:      "テスト市",
		ServiceType: "生活介護",
	}

	mockCertRepo := &mockCertificateRepository{
		certificates: map[domain.ID]*domain.BenefitCertificate{
			"cert-valid": validCert,
		},
	}
	mockRecipientRepo := &mockRecipientRepository{}
	mockStaffRepo := &mockStaffRepository{}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewCertificateUseCase(mockCertRepo, mockRecipientRepo, mockStaffRepo, mockAuditRepo)

	ctx := context.Background()

	// Test validation for current date (should be valid)
	result, err := usecase.ValidateCertificate(ctx, "cert-valid", now)
	if err != nil {
		t.Errorf("ValidateCertificate() error = %v", err)
	}

	if result == nil {
		t.Fatal("ValidateCertificate() returned nil result")
	}

	if !result.IsValid {
		t.Error("ValidateCertificate() IsValid = false, want true")
	}

	// Test validation for future date beyond expiration (should be invalid)
	futureDate := now.Add(60 * 24 * time.Hour) // 60 days from now
	result, err = usecase.ValidateCertificate(ctx, "cert-valid", futureDate)
	if err != nil {
		t.Errorf("ValidateCertificate() error = %v", err)
	}

	if result.IsValid {
		t.Error("ValidateCertificate() IsValid = true, want false for expired certificate")
	}

	if result.Reason == "" {
		t.Error("ValidateCertificate() should provide reason for invalid certificate")
	}
}
