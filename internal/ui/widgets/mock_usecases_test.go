package widgets

import (
	"context"
	"time"

	"shien-system/internal/domain"
	"shien-system/internal/usecase"
)

// MockRecipientUseCase implements usecase.RecipientUseCase for testing
type MockRecipientUseCase struct {
	recipients []*domain.Recipient
	err        error
}

func (m *MockRecipientUseCase) CreateRecipient(ctx context.Context, req usecase.CreateRecipientRequest) (*domain.Recipient, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &domain.Recipient{}, nil
}

func (m *MockRecipientUseCase) GetRecipient(ctx context.Context, id domain.ID) (*domain.Recipient, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.recipients) > 0 {
		return m.recipients[0], nil
	}
	return nil, nil
}

func (m *MockRecipientUseCase) UpdateRecipient(ctx context.Context, req usecase.UpdateRecipientRequest) (*domain.Recipient, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &domain.Recipient{}, nil
}

func (m *MockRecipientUseCase) DeleteRecipient(ctx context.Context, id domain.ID) error {
	return m.err
}

func (m *MockRecipientUseCase) ListRecipients(ctx context.Context, req usecase.ListRecipientsRequest) (*usecase.PaginatedRecipients, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &usecase.PaginatedRecipients{
		Recipients: m.recipients,
		Total:      len(m.recipients),
		Offset:     req.Offset,
		Limit:      req.Limit,
	}, nil
}

func (m *MockRecipientUseCase) GetActiveRecipients(ctx context.Context) ([]*domain.Recipient, error) {
	return m.recipients, m.err
}

func (m *MockRecipientUseCase) AssignStaff(ctx context.Context, req usecase.AssignStaffRequest) error {
	return m.err
}

func (m *MockRecipientUseCase) UnassignStaff(ctx context.Context, req usecase.UnassignStaffRequest) error {
	return m.err
}

// MockCertificateUseCase implements usecase.CertificateUseCase for testing
type MockCertificateUseCase struct {
	certificates []*domain.BenefitCertificate
	err          error
}

func (m *MockCertificateUseCase) CreateCertificate(ctx context.Context, req usecase.CreateCertificateRequest) (*domain.BenefitCertificate, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &domain.BenefitCertificate{}, nil
}

func (m *MockCertificateUseCase) GetCertificate(ctx context.Context, id domain.ID) (*domain.BenefitCertificate, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.certificates) > 0 {
		return m.certificates[0], nil
	}
	return nil, nil
}

func (m *MockCertificateUseCase) UpdateCertificate(ctx context.Context, req usecase.UpdateCertificateRequest) (*domain.BenefitCertificate, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &domain.BenefitCertificate{}, nil
}

func (m *MockCertificateUseCase) DeleteCertificate(ctx context.Context, id domain.ID) error {
	return m.err
}

func (m *MockCertificateUseCase) GetCertificatesByRecipient(ctx context.Context, recipientID domain.ID) ([]*domain.BenefitCertificate, error) {
	return m.certificates, m.err
}

func (m *MockCertificateUseCase) GetExpiringSoon(ctx context.Context, days int) ([]*domain.BenefitCertificate, error) {
	return m.certificates, m.err
}

func (m *MockCertificateUseCase) ValidateCertificate(ctx context.Context, certificateID domain.ID, date time.Time) (*usecase.ValidationResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &usecase.ValidationResult{
		IsValid: true,
		Reason:  "",
	}, nil
}

// MockSetupUseCase implements usecase.SetupUseCase for testing
type MockSetupUseCase struct {
	needsSetup bool
	err        error
}

func (m *MockSetupUseCase) NeedsInitialSetup(ctx context.Context) (bool, error) {
	return m.needsSetup, m.err
}

func (m *MockSetupUseCase) CreateInitialAdmin(ctx context.Context, name, password string) error {
	return m.err
}

// MockAuditLogRepository implements domain.AuditLogRepository for testing
type MockAuditLogRepository struct {
	logs []*domain.AuditLog
	err  error
}

func (m *MockAuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	return m.err
}

func (m *MockAuditLogRepository) GetByID(ctx context.Context, id domain.ID) (*domain.AuditLog, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.logs) > 0 {
		return m.logs[0], nil
	}
	return nil, nil
}

func (m *MockAuditLogRepository) GetByActorID(ctx context.Context, actorID domain.ID, limit, offset int) ([]*domain.AuditLog, error) {
	return m.logs, m.err
}

func (m *MockAuditLogRepository) GetByAction(ctx context.Context, action string, limit, offset int) ([]*domain.AuditLog, error) {
	return m.logs, m.err
}

func (m *MockAuditLogRepository) GetByTarget(ctx context.Context, target string, limit, offset int) ([]*domain.AuditLog, error) {
	return m.logs, m.err
}

func (m *MockAuditLogRepository) GetByTimeRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*domain.AuditLog, error) {
	return m.logs, m.err
}

func (m *MockAuditLogRepository) Search(ctx context.Context, query domain.AuditLogQuery, limit, offset int) ([]*domain.AuditLog, error) {
	return m.logs, m.err
}

func (m *MockAuditLogRepository) List(ctx context.Context, limit, offset int) ([]*domain.AuditLog, error) {
	return m.logs, m.err
}

func (m *MockAuditLogRepository) Count(ctx context.Context) (int, error) {
	return len(m.logs), m.err
}

// MockStaffRepository implements domain.StaffRepository for testing
type MockStaffRepository struct {
	staff []*domain.Staff
	err   error
}

// MockStaffUseCase implements usecase.StaffUseCase for testing
type MockStaffUseCase struct {
	staff []*domain.Staff
	err   error
}

func (m *MockStaffUseCase) CreateStaff(ctx context.Context, req usecase.CreateStaffRequest) (*domain.Staff, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &domain.Staff{}, nil
}

func (m *MockStaffUseCase) GetStaff(ctx context.Context, id domain.ID) (*domain.Staff, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.staff) > 0 {
		return m.staff[0], nil
	}
	return nil, nil
}

func (m *MockStaffUseCase) UpdateStaff(ctx context.Context, req usecase.UpdateStaffRequest) (*domain.Staff, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &domain.Staff{}, nil
}

func (m *MockStaffUseCase) DeleteStaff(ctx context.Context, id domain.ID) error {
	return m.err
}

func (m *MockStaffUseCase) ListStaff(ctx context.Context, req usecase.ListStaffRequest) (*usecase.PaginatedStaff, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &usecase.PaginatedStaff{
		Staff:  m.staff,
		Total:  len(m.staff),
		Offset: req.Offset,
		Limit:  req.Limit,
	}, nil
}

func (m *MockStaffUseCase) GetStaffByRole(ctx context.Context, role domain.StaffRole) ([]*domain.Staff, error) {
	return m.staff, m.err
}

func (m *MockStaffUseCase) GetAssignments(ctx context.Context, staffID domain.ID) ([]*domain.StaffAssignment, error) {
	return nil, m.err
}

func (m *MockStaffRepository) Create(ctx context.Context, staff *domain.Staff) error {
	return m.err
}

func (m *MockStaffRepository) GetByID(ctx context.Context, id domain.ID) (*domain.Staff, error) {
	return nil, m.err
}

func (m *MockStaffRepository) GetByExactName(ctx context.Context, name string) (*domain.Staff, error) {
	return nil, m.err
}

func (m *MockStaffRepository) Update(ctx context.Context, staff *domain.Staff) error {
	return m.err
}

func (m *MockStaffRepository) Delete(ctx context.Context, id domain.ID) error {
	return m.err
}

func (m *MockStaffRepository) List(ctx context.Context, limit, offset int) ([]*domain.Staff, error) {
	return m.staff, m.err
}

func (m *MockStaffRepository) GetByRole(ctx context.Context, role domain.StaffRole) ([]*domain.Staff, error) {
	return m.staff, m.err
}

func (m *MockStaffRepository) GetByName(ctx context.Context, name string) ([]*domain.Staff, error) {
	return m.staff, m.err
}

func (m *MockStaffRepository) Count(ctx context.Context) (int, error) {
	return len(m.staff), m.err
}


