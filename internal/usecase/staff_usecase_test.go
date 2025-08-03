package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"shien-system/internal/domain"
)

func TestStaffUseCase_CreateStaff(t *testing.T) {
	mockStaffRepo := &mockStaffRepository{
		staff: map[domain.ID]*domain.Staff{
			"admin-001": {
				ID:   "admin-001",
				Name: "管理者",
				Role: domain.RoleAdmin,
			},
		},
	}
	mockAssignmentRepo := &mockStaffAssignmentRepository{}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewStaffUseCase(mockStaffRepo, mockAssignmentRepo, mockAuditRepo)

	ctx := context.Background()

	req := CreateStaffRequest{
		Name:    "新規職員",
		Role:    domain.RoleStaff,
		ActorID: "admin-001",
	}

	staff, err := usecase.CreateStaff(ctx, req)
	if err != nil {
		t.Errorf("CreateStaff() error = %v", err)
	}

	if staff == nil {
		t.Fatal("CreateStaff() returned nil staff")
	}

	// Verify staff fields
	if staff.Name != req.Name {
		t.Errorf("Name = %v, want %v", staff.Name, req.Name)
	}
	if staff.Role != req.Role {
		t.Errorf("Role = %v, want %v", staff.Role, req.Role)
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

func TestStaffUseCase_CreateStaff_ValidationError(t *testing.T) {
	mockStaffRepo := &mockStaffRepository{
		staff: map[domain.ID]*domain.Staff{
			"admin-001": {
				ID:   "admin-001",
				Name: "管理者",
				Role: domain.RoleAdmin,
			},
		},
	}
	mockAssignmentRepo := &mockStaffAssignmentRepository{}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewStaffUseCase(mockStaffRepo, mockAssignmentRepo, mockAuditRepo)

	ctx := context.Background()

	// Test with empty name
	req := CreateStaffRequest{
		Name:    "", // Invalid: empty name
		Role:    domain.RoleStaff,
		ActorID: "admin-001",
	}

	_, err := usecase.CreateStaff(ctx, req)
	if err == nil {
		t.Error("CreateStaff() should return validation error for empty name")
	}

	var useCaseErr *UseCaseError
	if !errors.As(err, &useCaseErr) || useCaseErr.Code != "VALIDATION_FAILED" {
		t.Errorf("Expected VALIDATION_FAILED error, got %v", err)
	}
}

func TestStaffUseCase_GetStaff(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	existingStaff := &domain.Staff{
		ID:        "staff-001",
		Name:      "既存職員",
		Role:      domain.RoleStaff,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockStaffRepo := &mockStaffRepository{
		staff: map[domain.ID]*domain.Staff{
			"staff-001": existingStaff,
		},
	}
	mockAssignmentRepo := &mockStaffAssignmentRepository{}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewStaffUseCase(mockStaffRepo, mockAssignmentRepo, mockAuditRepo)

	ctx := context.Background()

	staff, err := usecase.GetStaff(ctx, "staff-001")
	if err != nil {
		t.Errorf("GetStaff() error = %v", err)
	}

	if staff == nil {
		t.Fatal("GetStaff() returned nil")
	}

	if staff.ID != "staff-001" {
		t.Errorf("ID = %v, want staff-001", staff.ID)
	}
	if staff.Name != "既存職員" {
		t.Errorf("Name = %v, want 既存職員", staff.Name)
	}
}

func TestStaffUseCase_GetStaff_NotFound(t *testing.T) {
	mockStaffRepo := &mockStaffRepository{
		staff: map[domain.ID]*domain.Staff{},
	}
	mockAssignmentRepo := &mockStaffAssignmentRepository{}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewStaffUseCase(mockStaffRepo, mockAssignmentRepo, mockAuditRepo)

	ctx := context.Background()

	_, err := usecase.GetStaff(ctx, "nonexistent")
	if err == nil {
		t.Error("GetStaff() should return error for nonexistent staff")
	}

	var useCaseErr *UseCaseError
	if !errors.As(err, &useCaseErr) || useCaseErr.Code != "STAFF_NOT_FOUND" {
		t.Errorf("Expected STAFF_NOT_FOUND error, got %v", err)
	}
}

func TestStaffUseCase_UpdateStaff(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	existingStaff := &domain.Staff{
		ID:        "staff-001",
		Name:      "既存職員",
		Role:      domain.RoleStaff,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockStaffRepo := &mockStaffRepository{
		staff: map[domain.ID]*domain.Staff{
			"staff-001": existingStaff,
			"admin-001": {
				ID:   "admin-001",
				Name: "管理者",
				Role: domain.RoleAdmin,
			},
		},
	}
	mockAssignmentRepo := &mockStaffAssignmentRepository{}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewStaffUseCase(mockStaffRepo, mockAssignmentRepo, mockAuditRepo)

	ctx := context.Background()

	req := UpdateStaffRequest{
		ID:      "staff-001",
		Name:    "更新後職員",
		Role:    domain.RoleAdmin,
		ActorID: "admin-001",
	}

	staff, err := usecase.UpdateStaff(ctx, req)
	if err != nil {
		t.Errorf("UpdateStaff() error = %v", err)
	}

	if staff == nil {
		t.Fatal("UpdateStaff() returned nil")
	}

	if staff.Name != req.Name {
		t.Errorf("Name = %v, want %v", staff.Name, req.Name)
	}
	if staff.Role != req.Role {
		t.Errorf("Role = %v, want %v", staff.Role, req.Role)
	}

	// Verify audit log was created
	if len(mockAuditRepo.logs) != 1 {
		t.Errorf("Expected 1 audit log, got %d", len(mockAuditRepo.logs))
	}
}

func TestStaffUseCase_DeleteStaff_WithActiveAssignments(t *testing.T) {
	existingStaff := &domain.Staff{
		ID:   "staff-001",
		Name: "削除対象職員",
		Role: domain.RoleStaff,
	}

	mockStaffRepo := &mockStaffRepository{
		staff: map[domain.ID]*domain.Staff{
			"staff-001": existingStaff,
		},
	}

	// Mock with active assignments
	mockAssignmentRepo := &mockStaffAssignmentRepository{
		assignments: map[domain.ID]*domain.StaffAssignment{
			"assignment-001": {
				ID:           "assignment-001",
				StaffID:      "staff-001",
				RecipientID:  "recipient-001",
				Role:         "主担当",
				AssignedAt:   time.Now().UTC(),
				UnassignedAt: nil, // Active assignment
			},
		},
	}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewStaffUseCase(mockStaffRepo, mockAssignmentRepo, mockAuditRepo)

	ctx := context.Background()

	err := usecase.DeleteStaff(ctx, "staff-001")
	if err == nil {
		t.Error("DeleteStaff() should return error when staff has active assignments")
	}

	var useCaseErr *UseCaseError
	if !errors.As(err, &useCaseErr) || useCaseErr.Code != "CANNOT_DELETE_STAFF" {
		t.Errorf("Expected CANNOT_DELETE_STAFF error, got %v", err)
	}
}

func TestStaffUseCase_DeleteStaff_Success(t *testing.T) {
	existingStaff := &domain.Staff{
		ID:   "staff-001",
		Name: "削除対象職員",
		Role: domain.RoleStaff,
	}

	mockStaffRepo := &mockStaffRepository{
		staff: map[domain.ID]*domain.Staff{
			"staff-001": existingStaff,
		},
	}

	// Mock with no active assignments
	mockAssignmentRepo := &mockStaffAssignmentRepository{
		assignments: map[domain.ID]*domain.StaffAssignment{},
	}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewStaffUseCase(mockStaffRepo, mockAssignmentRepo, mockAuditRepo)

	// Set context with actor information
	ctx := context.WithValue(context.Background(), ContextKeyUserID, "admin-001")

	err := usecase.DeleteStaff(ctx, "staff-001")
	if err != nil {
		t.Errorf("DeleteStaff() error = %v", err)
	}

	// Verify staff was deleted from mock
	if _, exists := mockStaffRepo.staff["staff-001"]; exists {
		t.Error("Staff should be deleted from repository")
	}
}

func TestStaffUseCase_GetStaffByRole(t *testing.T) {
	mockStaffRepo := &mockStaffRepository{
		staff: map[domain.ID]*domain.Staff{
			"staff-001": {
				ID:   "staff-001",
				Name: "職員1",
				Role: domain.RoleStaff,
			},
			"staff-002": {
				ID:   "staff-002",
				Name: "職員2",
				Role: domain.RoleStaff,
			},
			"admin-001": {
				ID:   "admin-001",
				Name: "管理者",
				Role: domain.RoleAdmin,
			},
		},
	}
	mockAssignmentRepo := &mockStaffAssignmentRepository{}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewStaffUseCase(mockStaffRepo, mockAssignmentRepo, mockAuditRepo)

	ctx := context.Background()

	staffMembers, err := usecase.GetStaffByRole(ctx, domain.RoleStaff)
	if err != nil {
		t.Errorf("GetStaffByRole() error = %v", err)
	}

	if len(staffMembers) != 2 {
		t.Errorf("GetStaffByRole() returned %d staff, want 2", len(staffMembers))
	}

	// Verify all returned staff have the correct role
	for _, staff := range staffMembers {
		if staff.Role != domain.RoleStaff {
			t.Errorf("Staff %s has role %v, want %v", staff.ID, staff.Role, domain.RoleStaff)
		}
	}
}

func TestStaffUseCase_GetAssignments(t *testing.T) {
	mockStaffRepo := &mockStaffRepository{
		staff: map[domain.ID]*domain.Staff{
			"staff-001": {
				ID:   "staff-001",
				Name: "テスト職員",
				Role: domain.RoleStaff,
			},
		},
	}

	now := time.Now().UTC()
	mockAssignmentRepo := &mockStaffAssignmentRepository{
		assignments: map[domain.ID]*domain.StaffAssignment{
			"assignment-001": {
				ID:          "assignment-001",
				StaffID:     "staff-001",
				RecipientID: "recipient-001",
				Role:        "主担当",
				AssignedAt:  now,
			},
			"assignment-002": {
				ID:          "assignment-002",
				StaffID:     "staff-001",
				RecipientID: "recipient-002",
				Role:        "副担当",
				AssignedAt:  now,
			},
			"assignment-003": {
				ID:          "assignment-003",
				StaffID:     "staff-002", // Different staff
				RecipientID: "recipient-003",
				Role:        "主担当",
				AssignedAt:  now,
			},
		},
	}
	mockAuditRepo := &mockAuditLogRepository{}

	usecase := NewStaffUseCase(mockStaffRepo, mockAssignmentRepo, mockAuditRepo)

	ctx := context.Background()

	assignments, err := usecase.GetAssignments(ctx, "staff-001")
	if err != nil {
		t.Errorf("GetAssignments() error = %v", err)
	}

	if len(assignments) != 2 {
		t.Errorf("GetAssignments() returned %d assignments, want 2", len(assignments))
	}

	// Verify all assignments belong to the correct staff
	for _, assignment := range assignments {
		if assignment.StaffID != "staff-001" {
			t.Errorf("Assignment %s has StaffID %s, want staff-001", assignment.ID, assignment.StaffID)
		}
	}
}
