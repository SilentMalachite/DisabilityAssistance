package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"shien-system/internal/domain"
)

// staffUseCase implements StaffUseCase interface
type staffUseCase struct {
	staffRepo      domain.StaffRepository
	assignmentRepo domain.StaffAssignmentRepository
	auditRepo      domain.AuditLogRepository
}

// NewStaffUseCase creates a new staff usecase
func NewStaffUseCase(
	staffRepo domain.StaffRepository,
	assignmentRepo domain.StaffAssignmentRepository,
	auditRepo domain.AuditLogRepository,
) StaffUseCase {
	return &staffUseCase{
		staffRepo:      staffRepo,
		assignmentRepo: assignmentRepo,
		auditRepo:      auditRepo,
	}
}

// CreateStaff creates a new staff member with validation
func (uc *staffUseCase) CreateStaff(ctx context.Context, req CreateStaffRequest) (*domain.Staff, error) {
	// Validate input
	if err := uc.validateCreateStaffRequest(req); err != nil {
		return nil, &UseCaseError{
			Code:    "VALIDATION_FAILED",
			Message: "入力値が不正です",
			Cause:   err,
		}
	}

	// Verify actor exists and has permission
	actor, err := uc.staffRepo.GetByID(ctx, req.ActorID)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, ErrUnauthorized
		}
		return nil, &UseCaseError{
			Code:    "INTERNAL_ERROR",
			Message: "内部エラーが発生しました",
			Cause:   err,
		}
	}

	// Check if actor has permission to create staff
	if !uc.canManageStaff(actor.Role, req.Role) {
		return nil, ErrUnauthorized
	}

	// Create staff
	now := time.Now().UTC()
	staff := &domain.Staff{
		ID:        domain.ID(uuid.New().String()),
		Name:      req.Name,
		Role:      req.Role,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = uc.staffRepo.Create(ctx, staff)
	if err != nil {
		return nil, &UseCaseError{
			Code:    "CREATION_FAILED",
			Message: "職員の作成に失敗しました",
			Cause:   err,
		}
	}

	// Log the action
	auditLog := &domain.AuditLog{
		ID:      domain.ID(uuid.New().String()),
		ActorID: req.ActorID,
		Action:  "CREATE",
		Target:  fmt.Sprintf("staff:%s", staff.ID),
		At:      now,
		IP:      uc.getClientIP(ctx),
		Details: fmt.Sprintf("職員「%s」を作成しました (役割: %s)", staff.Name, staff.Role),
	}

	err = uc.auditRepo.Create(ctx, auditLog)
	if err != nil {
		// Log audit failure but don't fail the operation
	}

	return staff, nil
}

// GetStaff retrieves a staff member by ID
func (uc *staffUseCase) GetStaff(ctx context.Context, id domain.ID) (*domain.Staff, error) {
	staff, err := uc.staffRepo.GetByID(ctx, id)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, ErrStaffNotFound
		}
		return nil, &UseCaseError{
			Code:    "RETRIEVAL_FAILED",
			Message: "職員の取得に失敗しました",
			Cause:   err,
		}
	}

	return staff, nil
}

// UpdateStaff updates staff information
func (uc *staffUseCase) UpdateStaff(ctx context.Context, req UpdateStaffRequest) (*domain.Staff, error) {
	// Validate input
	if err := uc.validateUpdateStaffRequest(req); err != nil {
		return nil, &UseCaseError{
			Code:    "VALIDATION_FAILED",
			Message: "入力値が不正です",
			Cause:   err,
		}
	}

	// Verify actor exists and has permission
	actor, err := uc.staffRepo.GetByID(ctx, req.ActorID)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, ErrUnauthorized
		}
		return nil, &UseCaseError{
			Code:    "INTERNAL_ERROR",
			Message: "内部エラーが発生しました",
			Cause:   err,
		}
	}

	// Get existing staff
	existing, err := uc.staffRepo.GetByID(ctx, req.ID)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, ErrStaffNotFound
		}
		return nil, &UseCaseError{
			Code:    "RETRIEVAL_FAILED",
			Message: "職員の取得に失敗しました",
			Cause:   err,
		}
	}

	// Check if actor has permission to update this staff
	if !uc.canManageStaff(actor.Role, existing.Role) || !uc.canManageStaff(actor.Role, req.Role) {
		return nil, ErrUnauthorized
	}

	// Update staff
	now := time.Now().UTC()
	staff := &domain.Staff{
		ID:        req.ID,
		Name:      req.Name,
		Role:      req.Role,
		CreatedAt: existing.CreatedAt, // Preserve original creation time
		UpdatedAt: now,
	}

	err = uc.staffRepo.Update(ctx, staff)
	if err != nil {
		return nil, &UseCaseError{
			Code:    "UPDATE_FAILED",
			Message: "職員の更新に失敗しました",
			Cause:   err,
		}
	}

	// Log the action
	auditLog := &domain.AuditLog{
		ID:      domain.ID(uuid.New().String()),
		ActorID: req.ActorID,
		Action:  "UPDATE",
		Target:  fmt.Sprintf("staff:%s", staff.ID),
		At:      now,
		IP:      uc.getClientIP(ctx),
		Details: fmt.Sprintf("職員「%s」を更新しました (役割: %s)", staff.Name, staff.Role),
	}

	err = uc.auditRepo.Create(ctx, auditLog)
	if err != nil {
		// Log audit failure but don't fail the operation
	}

	return staff, nil
}

// DeleteStaff deletes a staff member with assignment validation
func (uc *staffUseCase) DeleteStaff(ctx context.Context, id domain.ID) error {
	// Get existing staff
	staff, err := uc.staffRepo.GetByID(ctx, id)
	if err != nil {
		if err == domain.ErrNotFound {
			return ErrStaffNotFound
		}
		return &UseCaseError{
			Code:    "RETRIEVAL_FAILED",
			Message: "職員の取得に失敗しました",
			Cause:   err,
		}
	}

	// Check for active assignments
	activeAssignments, err := uc.assignmentRepo.GetActiveByStaffID(ctx, id)
	if err != nil {
		return &UseCaseError{
			Code:    "INTERNAL_ERROR",
			Message: "内部エラーが発生しました",
			Cause:   err,
		}
	}

	// Cannot delete staff with active assignments
	if len(activeAssignments) > 0 {
		return ErrCannotDeleteStaff
	}

	// Delete staff
	err = uc.staffRepo.Delete(ctx, id)
	if err != nil {
		return &UseCaseError{
			Code:    "DELETION_FAILED",
			Message: "職員の削除に失敗しました",
			Cause:   err,
		}
	}

	// Log the action (get actor from context)
	actorID := uc.getActorID(ctx)
	if actorID != "" {
		auditLog := &domain.AuditLog{
			ID:      domain.ID(uuid.New().String()),
			ActorID: actorID,
			Action:  "DELETE",
			Target:  fmt.Sprintf("staff:%s", id),
			At:      time.Now().UTC(),
			IP:      uc.getClientIP(ctx),
			Details: fmt.Sprintf("職員「%s」を削除しました", staff.Name),
		}

		uc.auditRepo.Create(ctx, auditLog)
	}

	return nil
}

// ListStaff retrieves paginated list of staff
func (uc *staffUseCase) ListStaff(ctx context.Context, req ListStaffRequest) (*PaginatedStaff, error) {
	var staff []*domain.Staff
	var err error

	// Apply filtering if specified
	if req.FilterBy.Role != nil {
		staff, err = uc.staffRepo.GetByRole(ctx, *req.FilterBy.Role)
		if err != nil {
			return nil, &UseCaseError{
				Code:    "LIST_FAILED",
				Message: "職員一覧の取得に失敗しました",
				Cause:   err,
			}
		}
		// Apply pagination manually for filtered results
		start := req.Offset
		if start > len(staff) {
			staff = []*domain.Staff{}
		} else {
			end := start + req.Limit
			if end > len(staff) {
				end = len(staff)
			}
			staff = staff[start:end]
		}
	} else {
		// Use repository pagination for unfiltered results
		staff, err = uc.staffRepo.List(ctx, req.Limit, req.Offset)
		if err != nil {
			return nil, &UseCaseError{
				Code:    "LIST_FAILED",
				Message: "職員一覧の取得に失敗しました",
				Cause:   err,
			}
		}
	}

	total, err := uc.staffRepo.Count(ctx)
	if err != nil {
		return nil, &UseCaseError{
			Code:    "COUNT_FAILED",
			Message: "職員数の取得に失敗しました",
			Cause:   err,
		}
	}

	return &PaginatedStaff{
		Staff:  staff,
		Total:  total,
		Limit:  req.Limit,
		Offset: req.Offset,
	}, nil
}

// GetStaffByRole retrieves staff members by role
func (uc *staffUseCase) GetStaffByRole(ctx context.Context, role domain.StaffRole) ([]*domain.Staff, error) {
	staff, err := uc.staffRepo.GetByRole(ctx, role)
	if err != nil {
		return nil, &UseCaseError{
			Code:    "RETRIEVAL_FAILED",
			Message: "役割別職員の取得に失敗しました",
			Cause:   err,
		}
	}

	return staff, nil
}

// GetAssignments retrieves assignments for a staff member
func (uc *staffUseCase) GetAssignments(ctx context.Context, staffID domain.ID) ([]*domain.StaffAssignment, error) {
	// Verify staff exists
	_, err := uc.staffRepo.GetByID(ctx, staffID)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, ErrStaffNotFound
		}
		return nil, &UseCaseError{
			Code:    "INTERNAL_ERROR",
			Message: "内部エラーが発生しました",
			Cause:   err,
		}
	}

	assignments, err := uc.assignmentRepo.GetByStaffID(ctx, staffID)
	if err != nil {
		return nil, &UseCaseError{
			Code:    "RETRIEVAL_FAILED",
			Message: "担当者割り当ての取得に失敗しました",
			Cause:   err,
		}
	}

	return assignments, nil
}

// Validation functions

func (uc *staffUseCase) validateCreateStaffRequest(req CreateStaffRequest) error {
	var errors []string

	if strings.TrimSpace(req.Name) == "" {
		errors = append(errors, "名前は必須です")
	}

	if req.Role == "" {
		errors = append(errors, "役割は必須です")
	}

	// Validate role value
	switch req.Role {
	case domain.RoleAdmin, domain.RoleStaff, domain.RoleReadOnly:
		// Valid roles
	default:
		errors = append(errors, "無効な役割です")
	}

	if req.ActorID == "" {
		errors = append(errors, "実行者IDは必須です")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, ", "))
	}

	return nil
}

func (uc *staffUseCase) validateUpdateStaffRequest(req UpdateStaffRequest) error {
	var errors []string

	if req.ID == "" {
		errors = append(errors, "IDは必須です")
	}

	if strings.TrimSpace(req.Name) == "" {
		errors = append(errors, "名前は必須です")
	}

	if req.Role == "" {
		errors = append(errors, "役割は必須です")
	}

	// Validate role value
	switch req.Role {
	case domain.RoleAdmin, domain.RoleStaff, domain.RoleReadOnly:
		// Valid roles
	default:
		errors = append(errors, "無効な役割です")
	}

	if req.ActorID == "" {
		errors = append(errors, "実行者IDは必須です")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, ", "))
	}

	return nil
}

// Permission helper functions

func (uc *staffUseCase) canManageStaff(actorRole, targetRole domain.StaffRole) bool {
	// Only admins can manage other staff
	if actorRole != domain.RoleAdmin {
		return false
	}

	// Admins can manage all roles
	return true
}

// Helper functions

func (uc *staffUseCase) getActorID(ctx context.Context) domain.ID {
	if actorID := ctx.Value(ContextKeyUserID); actorID != nil {
		if id, ok := actorID.(string); ok {
			return domain.ID(id)
		}
	}
	return ""
}

func (uc *staffUseCase) getClientIP(ctx context.Context) string {
	if ip := ctx.Value(ContextKeyClientIP); ip != nil {
		if clientIP, ok := ip.(string); ok {
			return clientIP
		}
	}
	return "unknown"
}
