package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"shien-system/internal/domain"
)

// Mock repositories for testing
type mockRecipientRepository struct {
	recipients map[domain.ID]*domain.Recipient
	nextError  error
}

func (m *mockRecipientRepository) Create(ctx context.Context, recipient *domain.Recipient) error {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return err
	}
	if m.recipients == nil {
		m.recipients = make(map[domain.ID]*domain.Recipient)
	}
	m.recipients[recipient.ID] = recipient
	return nil
}

func (m *mockRecipientRepository) GetByID(ctx context.Context, id domain.ID) (*domain.Recipient, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return nil, err
	}
	recipient, exists := m.recipients[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	return recipient, nil
}

func (m *mockRecipientRepository) Update(ctx context.Context, recipient *domain.Recipient) error {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return err
	}
	if _, exists := m.recipients[recipient.ID]; !exists {
		return domain.ErrNotFound
	}
	m.recipients[recipient.ID] = recipient
	return nil
}

func (m *mockRecipientRepository) Delete(ctx context.Context, id domain.ID) error {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return err
	}
	if _, exists := m.recipients[id]; !exists {
		return domain.ErrNotFound
	}
	delete(m.recipients, id)
	return nil
}

func (m *mockRecipientRepository) List(ctx context.Context, limit, offset int) ([]*domain.Recipient, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return nil, err
	}
	var recipients []*domain.Recipient
	for _, recipient := range m.recipients {
		recipients = append(recipients, recipient)
	}
	// Simple pagination simulation
	start := offset
	if start > len(recipients) {
		return []*domain.Recipient{}, nil
	}
	end := start + limit
	if end > len(recipients) {
		end = len(recipients)
	}
	return recipients[start:end], nil
}

func (m *mockRecipientRepository) Search(ctx context.Context, query string, limit, offset int) ([]*domain.Recipient, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return nil, err
	}
	// Simple search implementation for testing
	var results []*domain.Recipient
	for _, recipient := range m.recipients {
		// Simple name search
		if strings.Contains(strings.ToLower(recipient.Name), strings.ToLower(query)) {
			results = append(results, recipient)
		}
	}
	// Apply pagination
	start := offset
	if start > len(results) {
		return []*domain.Recipient{}, nil
	}
	end := start + limit
	if end > len(results) {
		end = len(results)
	}
	return results[start:end], nil
}

func (m *mockRecipientRepository) GetByStaffID(ctx context.Context, staffID domain.ID) ([]*domain.Recipient, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return nil, err
	}
	// For testing, return empty slice
	return []*domain.Recipient{}, nil
}

func (m *mockRecipientRepository) GetActive(ctx context.Context, limit, offset int) ([]*domain.Recipient, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return nil, err
	}
	var activeRecipients []*domain.Recipient
	for _, recipient := range m.recipients {
		if recipient.DischargeDate == nil {
			activeRecipients = append(activeRecipients, recipient)
		}
	}
	// Apply pagination
	start := offset
	if start > len(activeRecipients) {
		return []*domain.Recipient{}, nil
	}
	end := start + limit
	if end > len(activeRecipients) {
		end = len(activeRecipients)
	}
	return activeRecipients[start:end], nil
}

func (m *mockRecipientRepository) CountActive(ctx context.Context) (int, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return 0, err
	}
	count := 0
	for _, recipient := range m.recipients {
		if recipient.DischargeDate == nil {
			count++
		}
	}
	return count, nil
}

func (m *mockRecipientRepository) Count(ctx context.Context) (int, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return 0, err
	}
	return len(m.recipients), nil
}

type mockStaffAssignmentRepository struct {
	assignments map[domain.ID]*domain.StaffAssignment
	nextError   error
}

func (m *mockStaffAssignmentRepository) Create(ctx context.Context, assignment *domain.StaffAssignment) error {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return err
	}
	if m.assignments == nil {
		m.assignments = make(map[domain.ID]*domain.StaffAssignment)
	}
	m.assignments[assignment.ID] = assignment
	return nil
}

func (m *mockStaffAssignmentRepository) GetByID(ctx context.Context, id domain.ID) (*domain.StaffAssignment, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return nil, err
	}
	assignment, exists := m.assignments[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	return assignment, nil
}

func (m *mockStaffAssignmentRepository) Update(ctx context.Context, assignment *domain.StaffAssignment) error {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return err
	}
	if _, exists := m.assignments[assignment.ID]; !exists {
		return domain.ErrNotFound
	}
	m.assignments[assignment.ID] = assignment
	return nil
}

func (m *mockStaffAssignmentRepository) Delete(ctx context.Context, id domain.ID) error {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return err
	}
	if _, exists := m.assignments[id]; !exists {
		return domain.ErrNotFound
	}
	delete(m.assignments, id)
	return nil
}

func (m *mockStaffAssignmentRepository) GetByRecipientID(ctx context.Context, recipientID domain.ID) ([]*domain.StaffAssignment, error) {
	var assignments []*domain.StaffAssignment
	for _, assignment := range m.assignments {
		if assignment.RecipientID == recipientID {
			assignments = append(assignments, assignment)
		}
	}
	return assignments, nil
}

func (m *mockStaffAssignmentRepository) GetByStaffID(ctx context.Context, staffID domain.ID) ([]*domain.StaffAssignment, error) {
	var assignments []*domain.StaffAssignment
	for _, assignment := range m.assignments {
		if assignment.StaffID == staffID {
			assignments = append(assignments, assignment)
		}
	}
	return assignments, nil
}

func (m *mockStaffAssignmentRepository) GetActiveByRecipientID(ctx context.Context, recipientID domain.ID) ([]*domain.StaffAssignment, error) {
	var assignments []*domain.StaffAssignment
	for _, assignment := range m.assignments {
		if assignment.RecipientID == recipientID && assignment.UnassignedAt == nil {
			assignments = append(assignments, assignment)
		}
	}
	return assignments, nil
}

func (m *mockStaffAssignmentRepository) GetActiveByStaffID(ctx context.Context, staffID domain.ID) ([]*domain.StaffAssignment, error) {
	var assignments []*domain.StaffAssignment
	for _, assignment := range m.assignments {
		if assignment.StaffID == staffID && assignment.UnassignedAt == nil {
			assignments = append(assignments, assignment)
		}
	}
	return assignments, nil
}

func (m *mockStaffAssignmentRepository) UnassignAll(ctx context.Context, recipientID domain.ID, unassignedAt time.Time) error {
	for _, assignment := range m.assignments {
		if assignment.RecipientID == recipientID && assignment.UnassignedAt == nil {
			assignment.UnassignedAt = &unassignedAt
		}
	}
	return nil
}

func (m *mockStaffAssignmentRepository) List(ctx context.Context, limit, offset int) ([]*domain.StaffAssignment, error) {
	var assignments []*domain.StaffAssignment
	for _, assignment := range m.assignments {
		assignments = append(assignments, assignment)
	}
	// Simple pagination simulation
	start := offset
	if start > len(assignments) {
		return []*domain.StaffAssignment{}, nil
	}
	end := start + limit
	if end > len(assignments) {
		end = len(assignments)
	}
	return assignments[start:end], nil
}

func (m *mockStaffAssignmentRepository) Count(ctx context.Context) (int, error) {
	return len(m.assignments), nil
}

type mockStaffRepository struct {
	staff     map[domain.ID]*domain.Staff
	nextError error
}

func (m *mockStaffRepository) GetByID(ctx context.Context, id domain.ID) (*domain.Staff, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return nil, err
	}
	staff, exists := m.staff[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	return staff, nil
}

func (m *mockStaffRepository) GetByExactName(ctx context.Context, name string) (*domain.Staff, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return nil, err
	}
	for _, staff := range m.staff {
		if staff.Name == name {
			return staff, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockStaffRepository) Create(ctx context.Context, staff *domain.Staff) error {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return err
	}
	if m.staff == nil {
		m.staff = make(map[domain.ID]*domain.Staff)
	}
	m.staff[staff.ID] = staff
	return nil
}

func (m *mockStaffRepository) Update(ctx context.Context, staff *domain.Staff) error {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return err
	}
	if _, exists := m.staff[staff.ID]; !exists {
		return domain.ErrNotFound
	}
	m.staff[staff.ID] = staff
	return nil
}

func (m *mockStaffRepository) Delete(ctx context.Context, id domain.ID) error {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return err
	}
	if _, exists := m.staff[id]; !exists {
		return domain.ErrNotFound
	}
	delete(m.staff, id)
	return nil
}

func (m *mockStaffRepository) List(ctx context.Context, limit, offset int) ([]*domain.Staff, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return nil, err
	}
	var staff []*domain.Staff
	for _, s := range m.staff {
		staff = append(staff, s)
	}
	// Simple pagination simulation
	start := offset
	if start > len(staff) {
		return []*domain.Staff{}, nil
	}
	end := start + limit
	if end > len(staff) {
		end = len(staff)
	}
	return staff[start:end], nil
}

func (m *mockStaffRepository) GetByRole(ctx context.Context, role domain.StaffRole) ([]*domain.Staff, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return nil, err
	}
	var result []*domain.Staff
	for _, staff := range m.staff {
		if staff.Role == role {
			result = append(result, staff)
		}
	}
	return result, nil
}

func (m *mockStaffRepository) GetByName(ctx context.Context, name string) ([]*domain.Staff, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return nil, err
	}
	var result []*domain.Staff
	for _, staff := range m.staff {
		if strings.Contains(strings.ToLower(staff.Name), strings.ToLower(name)) {
			result = append(result, staff)
		}
	}
	return result, nil
}

func (m *mockStaffRepository) Count(ctx context.Context) (int, error) {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return 0, err
	}
	return len(m.staff), nil
}

type mockAuditLogRepository struct {
	logs      []*domain.AuditLog
	nextError error
}

func (m *mockAuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	if m.nextError != nil {
		err := m.nextError
		m.nextError = nil
		return err
	}
	m.logs = append(m.logs, log)
	return nil
}

func (m *mockAuditLogRepository) GetByID(ctx context.Context, id domain.ID) (*domain.AuditLog, error) {
	return nil, nil
}

func (m *mockAuditLogRepository) GetByActorID(ctx context.Context, actorID domain.ID, limit, offset int) ([]*domain.AuditLog, error) {
	return nil, nil
}

func (m *mockAuditLogRepository) GetByAction(ctx context.Context, action string, limit, offset int) ([]*domain.AuditLog, error) {
	return nil, nil
}

func (m *mockAuditLogRepository) GetByTarget(ctx context.Context, target string, limit, offset int) ([]*domain.AuditLog, error) {
	return nil, nil
}

func (m *mockAuditLogRepository) GetByTimeRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*domain.AuditLog, error) {
	return nil, nil
}

func (m *mockAuditLogRepository) Search(ctx context.Context, query domain.AuditLogQuery, limit, offset int) ([]*domain.AuditLog, error) {
	return nil, nil
}

func (m *mockAuditLogRepository) List(ctx context.Context, limit, offset int) ([]*domain.AuditLog, error) {
	return nil, nil
}

func (m *mockAuditLogRepository) Count(ctx context.Context) (int, error) {
	return 0, nil
}

func TestRecipientUseCase_CreateRecipient(t *testing.T) {
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
	mockAssignmentRepo := &mockStaffAssignmentRepository{}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewRecipientUseCase(mockRecipientRepo, mockStaffRepo, mockAssignmentRepo, mockAuditRepo)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	req := CreateRecipientRequest{
		Name:             "テスト利用者",
		Kana:             "テストリヨウシャ",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		DisabilityName:   "知的障害",
		HasDisabilityID:  true,
		Grade:            "B1",
		Address:          "東京都渋谷区1-1-1",
		Phone:            "03-1234-5678",
		Email:            "test@example.com",
		PublicAssistance: false,
		AdmissionDate:    &now,
		ActorID:          "staff-001",
	}

	recipient, err := usecase.CreateRecipient(ctx, req)
	if err != nil {
		t.Errorf("CreateRecipient() error = %v", err)
	}

	if recipient == nil {
		t.Fatal("CreateRecipient() returned nil recipient")
	}

	// Verify recipient fields
	if recipient.Name != req.Name {
		t.Errorf("Name = %v, want %v", recipient.Name, req.Name)
	}
	if recipient.Sex != req.Sex {
		t.Errorf("Sex = %v, want %v", recipient.Sex, req.Sex)
	}
	if !recipient.BirthDate.Equal(req.BirthDate) {
		t.Errorf("BirthDate = %v, want %v", recipient.BirthDate, req.BirthDate)
	}

	// Verify audit log was created
	if len(mockAuditRepo.logs) != 1 {
		t.Errorf("Expected 1 audit log, got %d", len(mockAuditRepo.logs))
	}

	if len(mockAuditRepo.logs) > 0 {
		auditLog := mockAuditRepo.logs[0]
		if auditLog.ActorID != req.ActorID {
			t.Errorf("Audit log ActorID = %v, want %v", auditLog.ActorID, req.ActorID)
		}
		if auditLog.Action != "CREATE" {
			t.Errorf("Audit log Action = %v, want CREATE", auditLog.Action)
		}
	}
}

func TestRecipientUseCase_CreateRecipient_ValidationError(t *testing.T) {
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
	mockAssignmentRepo := &mockStaffAssignmentRepository{}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewRecipientUseCase(mockRecipientRepo, mockStaffRepo, mockAssignmentRepo, mockAuditRepo)

	ctx := context.Background()

	// Test with empty name
	req := CreateRecipientRequest{
		Name:    "", // Invalid: empty name
		Sex:     domain.SexMale,
		ActorID: "staff-001",
	}

	_, err := usecase.CreateRecipient(ctx, req)
	if err == nil {
		t.Error("CreateRecipient() should return validation error for empty name")
	}

	var useCaseErr *UseCaseError
	if !errors.As(err, &useCaseErr) || useCaseErr.Code != "VALIDATION_FAILED" {
		t.Errorf("Expected VALIDATION_FAILED error, got %v", err)
	}
}

func TestRecipientUseCase_GetRecipient(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	existingRecipient := &domain.Recipient{
		ID:               "recipient-001",
		Name:             "既存利用者",
		Sex:              domain.SexFemale,
		BirthDate:        time.Date(1985, 5, 15, 0, 0, 0, 0, time.UTC),
		HasDisabilityID:  true,
		PublicAssistance: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	mockRecipientRepo := &mockRecipientRepository{
		recipients: map[domain.ID]*domain.Recipient{
			"recipient-001": existingRecipient,
		},
	}
	mockStaffRepo := &mockStaffRepository{}
	mockAssignmentRepo := &mockStaffAssignmentRepository{}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewRecipientUseCase(mockRecipientRepo, mockStaffRepo, mockAssignmentRepo, mockAuditRepo)

	ctx := context.Background()

	recipient, err := usecase.GetRecipient(ctx, "recipient-001")
	if err != nil {
		t.Errorf("GetRecipient() error = %v", err)
	}

	if recipient == nil {
		t.Fatal("GetRecipient() returned nil")
	}

	if recipient.ID != "recipient-001" {
		t.Errorf("ID = %v, want recipient-001", recipient.ID)
	}
	if recipient.Name != "既存利用者" {
		t.Errorf("Name = %v, want 既存利用者", recipient.Name)
	}
}

func TestRecipientUseCase_GetRecipient_NotFound(t *testing.T) {
	mockRecipientRepo := &mockRecipientRepository{
		recipients: map[domain.ID]*domain.Recipient{},
	}
	mockStaffRepo := &mockStaffRepository{}
	mockAssignmentRepo := &mockStaffAssignmentRepository{}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewRecipientUseCase(mockRecipientRepo, mockStaffRepo, mockAssignmentRepo, mockAuditRepo)

	ctx := context.Background()

	_, err := usecase.GetRecipient(ctx, "nonexistent")
	if err == nil {
		t.Error("GetRecipient() should return error for nonexistent recipient")
	}

	var useCaseErr *UseCaseError
	if !errors.As(err, &useCaseErr) || useCaseErr.Code != "RECIPIENT_NOT_FOUND" {
		t.Errorf("Expected RECIPIENT_NOT_FOUND error, got %v", err)
	}
}

func TestRecipientUseCase_AssignStaff(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	existingRecipient := &domain.Recipient{
		ID:        "recipient-001",
		Name:      "テスト利用者",
		CreatedAt: now,
		UpdatedAt: now,
	}
	existingStaff := &domain.Staff{
		ID:   "staff-001",
		Name: "テスト職員",
		Role: domain.RoleStaff,
	}

	mockRecipientRepo := &mockRecipientRepository{
		recipients: map[domain.ID]*domain.Recipient{
			"recipient-001": existingRecipient,
		},
	}
	mockStaffRepo := &mockStaffRepository{
		staff: map[domain.ID]*domain.Staff{
			"staff-001": existingStaff,
		},
	}
	mockAssignmentRepo := &mockStaffAssignmentRepository{}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewRecipientUseCase(mockRecipientRepo, mockStaffRepo, mockAssignmentRepo, mockAuditRepo)

	ctx := context.Background()

	req := AssignStaffRequest{
		RecipientID: "recipient-001",
		StaffID:     "staff-001",
		Role:        "主担当",
		ActorID:     "staff-001",
	}

	err := usecase.AssignStaff(ctx, req)
	if err != nil {
		t.Errorf("AssignStaff() error = %v", err)
	}

	// Verify assignment was created
	assignments := mockAssignmentRepo.assignments
	if len(assignments) != 1 {
		t.Errorf("Expected 1 assignment, got %d", len(assignments))
	}

	// Verify audit log
	if len(mockAuditRepo.logs) != 1 {
		t.Errorf("Expected 1 audit log, got %d", len(mockAuditRepo.logs))
	}
}
