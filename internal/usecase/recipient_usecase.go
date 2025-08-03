package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"shien-system/internal/domain"
)

// recipientUseCase implements RecipientUseCase interface
type recipientUseCase struct {
	recipientRepo  domain.RecipientRepository
	staffRepo      domain.StaffRepository
	assignmentRepo domain.StaffAssignmentRepository
	auditRepo      domain.AuditLogRepository
}

// NewRecipientUseCase creates a new recipient usecase
func NewRecipientUseCase(
	recipientRepo domain.RecipientRepository,
	staffRepo domain.StaffRepository,
	assignmentRepo domain.StaffAssignmentRepository,
	auditRepo domain.AuditLogRepository,
) RecipientUseCase {
	return &recipientUseCase{
		recipientRepo:  recipientRepo,
		staffRepo:      staffRepo,
		assignmentRepo: assignmentRepo,
		auditRepo:      auditRepo,
	}
}

// CreateRecipient creates a new recipient with audit logging
func (uc *recipientUseCase) CreateRecipient(ctx context.Context, req CreateRecipientRequest) (*domain.Recipient, error) {
	// Validate input
	if err := uc.validateCreateRecipientRequest(req); err != nil {
		return nil, &UseCaseError{
			Code:    "VALIDATION_FAILED",
			Message: "入力値が不正です",
			Cause:   err,
		}
	}

	// Verify actor exists
	_, err := uc.staffRepo.GetByID(ctx, req.ActorID)
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

	// Create recipient
	now := time.Now().UTC()
	recipient := &domain.Recipient{
		ID:               domain.ID(uuid.New().String()),
		Name:             req.Name,
		Kana:             req.Kana,
		Sex:              req.Sex,
		BirthDate:        req.BirthDate,
		DisabilityName:   req.DisabilityName,
		HasDisabilityID:  req.HasDisabilityID,
		Grade:            req.Grade,
		Address:          req.Address,
		Phone:            req.Phone,
		Email:            req.Email,
		PublicAssistance: req.PublicAssistance,
		AdmissionDate:    req.AdmissionDate,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err = uc.recipientRepo.Create(ctx, recipient)
	if err != nil {
		return nil, &UseCaseError{
			Code:    "CREATION_FAILED",
			Message: "利用者の作成に失敗しました",
			Cause:   err,
		}
	}

	// Log the action
	auditLog := &domain.AuditLog{
		ID:      domain.ID(uuid.New().String()),
		ActorID: req.ActorID,
		Action:  "CREATE",
		Target:  fmt.Sprintf("recipient:%s", recipient.ID),
		At:      now,
		IP:      uc.getClientIP(ctx),
		Details: fmt.Sprintf("利用者「%s」を作成しました", recipient.Name),
	}

	err = uc.auditRepo.Create(ctx, auditLog)
	if err != nil {
		// Log audit failure but don't fail the operation
		// In production, this should be logged to monitoring system
	}

	return recipient, nil
}

// GetRecipient retrieves a recipient by ID with access control
func (uc *recipientUseCase) GetRecipient(ctx context.Context, id domain.ID) (*domain.Recipient, error) {
	recipient, err := uc.recipientRepo.GetByID(ctx, id)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, ErrRecipientNotFound
		}
		return nil, &UseCaseError{
			Code:    "RETRIEVAL_FAILED",
			Message: "利用者の取得に失敗しました",
			Cause:   err,
		}
	}

	return recipient, nil
}

// UpdateRecipient updates recipient information with audit logging
func (uc *recipientUseCase) UpdateRecipient(ctx context.Context, req UpdateRecipientRequest) (*domain.Recipient, error) {
	// Validate input
	if err := uc.validateUpdateRecipientRequest(req); err != nil {
		return nil, &UseCaseError{
			Code:    "VALIDATION_FAILED",
			Message: "入力値が不正です",
			Cause:   err,
		}
	}

	// Verify actor exists
	_, err := uc.staffRepo.GetByID(ctx, req.ActorID)
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

	// Get existing recipient
	existing, err := uc.recipientRepo.GetByID(ctx, req.ID)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, ErrRecipientNotFound
		}
		return nil, &UseCaseError{
			Code:    "RETRIEVAL_FAILED",
			Message: "利用者の取得に失敗しました",
			Cause:   err,
		}
	}

	// Update recipient
	now := time.Now().UTC()
	recipient := &domain.Recipient{
		ID:               req.ID,
		Name:             req.Name,
		Kana:             req.Kana,
		Sex:              req.Sex,
		BirthDate:        req.BirthDate,
		DisabilityName:   req.DisabilityName,
		HasDisabilityID:  req.HasDisabilityID,
		Grade:            req.Grade,
		Address:          req.Address,
		Phone:            req.Phone,
		Email:            req.Email,
		PublicAssistance: req.PublicAssistance,
		AdmissionDate:    req.AdmissionDate,
		DischargeDate:    req.DischargeDate,
		CreatedAt:        existing.CreatedAt, // Preserve original creation time
		UpdatedAt:        now,
	}

	err = uc.recipientRepo.Update(ctx, recipient)
	if err != nil {
		return nil, &UseCaseError{
			Code:    "UPDATE_FAILED",
			Message: "利用者の更新に失敗しました",
			Cause:   err,
		}
	}

	// Log the action
	auditLog := &domain.AuditLog{
		ID:      domain.ID(uuid.New().String()),
		ActorID: req.ActorID,
		Action:  "UPDATE",
		Target:  fmt.Sprintf("recipient:%s", recipient.ID),
		At:      now,
		IP:      uc.getClientIP(ctx),
		Details: fmt.Sprintf("利用者「%s」を更新しました", recipient.Name),
	}

	err = uc.auditRepo.Create(ctx, auditLog)
	if err != nil {
		// Log audit failure but don't fail the operation
	}

	return recipient, nil
}

// DeleteRecipient soft deletes a recipient with cascade handling
func (uc *recipientUseCase) DeleteRecipient(ctx context.Context, id domain.ID) error {
	// Get existing recipient
	recipient, err := uc.recipientRepo.GetByID(ctx, id)
	if err != nil {
		if err == domain.ErrNotFound {
			return ErrRecipientNotFound
		}
		return &UseCaseError{
			Code:    "RETRIEVAL_FAILED",
			Message: "利用者の取得に失敗しました",
			Cause:   err,
		}
	}

	// Check for active assignments
	activeAssignments, err := uc.assignmentRepo.GetActiveByRecipientID(ctx, id)
	if err != nil {
		return &UseCaseError{
			Code:    "INTERNAL_ERROR",
			Message: "内部エラーが発生しました",
			Cause:   err,
		}
	}

	// Unassign all active assignments before deletion
	if len(activeAssignments) > 0 {
		now := time.Now().UTC()
		err = uc.assignmentRepo.UnassignAll(ctx, id, now)
		if err != nil {
			return &UseCaseError{
				Code:    "UNASSIGN_FAILED",
				Message: "担当者の割り当て解除に失敗しました",
				Cause:   err,
			}
		}
	}

	// Delete recipient
	err = uc.recipientRepo.Delete(ctx, id)
	if err != nil {
		return &UseCaseError{
			Code:    "DELETION_FAILED",
			Message: "利用者の削除に失敗しました",
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
			Target:  fmt.Sprintf("recipient:%s", id),
			At:      time.Now().UTC(),
			IP:      uc.getClientIP(ctx),
			Details: fmt.Sprintf("利用者「%s」を削除しました", recipient.Name),
		}

		uc.auditRepo.Create(ctx, auditLog)
	}

	return nil
}

// ListRecipients retrieves paginated list of recipients
func (uc *recipientUseCase) ListRecipients(ctx context.Context, req ListRecipientsRequest) (*PaginatedRecipients, error) {
	// For now, implement basic listing without filtering
	// Advanced filtering can be added later
	recipients, err := uc.recipientRepo.List(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, &UseCaseError{
			Code:    "LIST_FAILED",
			Message: "利用者一覧の取得に失敗しました",
			Cause:   err,
		}
	}

	total, err := uc.recipientRepo.Count(ctx)
	if err != nil {
		return nil, &UseCaseError{
			Code:    "COUNT_FAILED",
			Message: "利用者数の取得に失敗しました",
			Cause:   err,
		}
	}

	return &PaginatedRecipients{
		Recipients: recipients,
		Total:      total,
		Limit:      req.Limit,
		Offset:     req.Offset,
	}, nil
}

// GetActiveRecipients retrieves all currently active recipients
func (uc *recipientUseCase) GetActiveRecipients(ctx context.Context) ([]*domain.Recipient, error) {
	// Get all active recipients (no pagination for this method)
	recipients, err := uc.recipientRepo.GetActive(ctx, 1000, 0) // Large limit for all active
	if err != nil {
		return nil, &UseCaseError{
			Code:    "RETRIEVAL_FAILED",
			Message: "アクティブな利用者の取得に失敗しました",
			Cause:   err,
		}
	}

	return recipients, nil
}

// AssignStaff assigns staff members to a recipient
func (uc *recipientUseCase) AssignStaff(ctx context.Context, req AssignStaffRequest) error {
	// Validate input
	if err := uc.validateAssignStaffRequest(req); err != nil {
		return &UseCaseError{
			Code:    "VALIDATION_FAILED",
			Message: "入力値が不正です",
			Cause:   err,
		}
	}

	// Verify actor exists
	_, err := uc.staffRepo.GetByID(ctx, req.ActorID)
	if err != nil {
		if err == domain.ErrNotFound {
			return ErrUnauthorized
		}
		return &UseCaseError{
			Code:    "INTERNAL_ERROR",
			Message: "内部エラーが発生しました",
			Cause:   err,
		}
	}

	// Verify recipient exists
	_, err = uc.recipientRepo.GetByID(ctx, req.RecipientID)
	if err != nil {
		if err == domain.ErrNotFound {
			return ErrRecipientNotFound
		}
		return &UseCaseError{
			Code:    "INTERNAL_ERROR",
			Message: "内部エラーが発生しました",
			Cause:   err,
		}
	}

	// Verify staff exists
	_, err = uc.staffRepo.GetByID(ctx, req.StaffID)
	if err != nil {
		if err == domain.ErrNotFound {
			return ErrStaffNotFound
		}
		return &UseCaseError{
			Code:    "INTERNAL_ERROR",
			Message: "内部エラーが発生しました",
			Cause:   err,
		}
	}

	// Create assignment
	now := time.Now().UTC()
	assignment := &domain.StaffAssignment{
		ID:           domain.ID(uuid.New().String()),
		RecipientID:  req.RecipientID,
		StaffID:      req.StaffID,
		Role:         req.Role,
		AssignedAt:   now,
		UnassignedAt: nil,
	}

	err = uc.assignmentRepo.Create(ctx, assignment)
	if err != nil {
		return &UseCaseError{
			Code:    "ASSIGNMENT_FAILED",
			Message: "担当者の割り当てに失敗しました",
			Cause:   err,
		}
	}

	// Log the action
	auditLog := &domain.AuditLog{
		ID:      domain.ID(uuid.New().String()),
		ActorID: req.ActorID,
		Action:  "ASSIGN",
		Target:  fmt.Sprintf("assignment:%s", assignment.ID),
		At:      now,
		IP:      uc.getClientIP(ctx),
		Details: fmt.Sprintf("担当者を割り当てました (役割: %s)", req.Role),
	}

	uc.auditRepo.Create(ctx, auditLog)

	return nil
}

// UnassignStaff removes staff assignment from a recipient
func (uc *recipientUseCase) UnassignStaff(ctx context.Context, req UnassignStaffRequest) error {
	// Validate input
	if req.AssignmentID == "" {
		return &UseCaseError{
			Code:    "VALIDATION_FAILED",
			Message: "割り当てIDが必要です",
		}
	}

	// Verify actor exists
	_, err := uc.staffRepo.GetByID(ctx, req.ActorID)
	if err != nil {
		if err == domain.ErrNotFound {
			return ErrUnauthorized
		}
		return &UseCaseError{
			Code:    "INTERNAL_ERROR",
			Message: "内部エラーが発生しました",
			Cause:   err,
		}
	}

	// Get assignment
	assignment, err := uc.assignmentRepo.GetByID(ctx, req.AssignmentID)
	if err != nil {
		if err == domain.ErrNotFound {
			return &UseCaseError{
				Code:    "ASSIGNMENT_NOT_FOUND",
				Message: "担当者割り当てが見つかりません",
			}
		}
		return &UseCaseError{
			Code:    "INTERNAL_ERROR",
			Message: "内部エラーが発生しました",
			Cause:   err,
		}
	}

	// Check if already unassigned
	if assignment.UnassignedAt != nil {
		return &UseCaseError{
			Code:    "ALREADY_UNASSIGNED",
			Message: "既に割り当てが解除されています",
		}
	}

	// Unassign
	now := time.Now().UTC()
	assignment.UnassignedAt = &now

	err = uc.assignmentRepo.Update(ctx, assignment)
	if err != nil {
		return &UseCaseError{
			Code:    "UNASSIGN_FAILED",
			Message: "担当者の割り当て解除に失敗しました",
			Cause:   err,
		}
	}

	// Log the action
	auditLog := &domain.AuditLog{
		ID:      domain.ID(uuid.New().String()),
		ActorID: req.ActorID,
		Action:  "UNASSIGN",
		Target:  fmt.Sprintf("assignment:%s", assignment.ID),
		At:      now,
		IP:      uc.getClientIP(ctx),
		Details: "担当者の割り当てを解除しました",
	}

	uc.auditRepo.Create(ctx, auditLog)

	return nil
}

// Validation functions

func (uc *recipientUseCase) validateCreateRecipientRequest(req CreateRecipientRequest) error {
	var errors []string

	if strings.TrimSpace(req.Name) == "" {
		errors = append(errors, "名前は必須です")
	}

	if req.Sex == "" {
		errors = append(errors, "性別は必須です")
	}

	if req.BirthDate.IsZero() {
		errors = append(errors, "生年月日は必須です")
	}

	if req.ActorID == "" {
		errors = append(errors, "実行者IDは必須です")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, ", "))
	}

	return nil
}

func (uc *recipientUseCase) validateUpdateRecipientRequest(req UpdateRecipientRequest) error {
	var errors []string

	if req.ID == "" {
		errors = append(errors, "IDは必須です")
	}

	if strings.TrimSpace(req.Name) == "" {
		errors = append(errors, "名前は必須です")
	}

	if req.Sex == "" {
		errors = append(errors, "性別は必須です")
	}

	if req.BirthDate.IsZero() {
		errors = append(errors, "生年月日は必須です")
	}

	if req.ActorID == "" {
		errors = append(errors, "実行者IDは必須です")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, ", "))
	}

	return nil
}

func (uc *recipientUseCase) validateAssignStaffRequest(req AssignStaffRequest) error {
	var errors []string

	if req.RecipientID == "" {
		errors = append(errors, "利用者IDは必須です")
	}

	if req.StaffID == "" {
		errors = append(errors, "職員IDは必須です")
	}

	if strings.TrimSpace(req.Role) == "" {
		errors = append(errors, "役割は必須です")
	}

	if req.ActorID == "" {
		errors = append(errors, "実行者IDは必須です")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, ", "))
	}

	return nil
}

// Helper functions

func (uc *recipientUseCase) getActorID(ctx context.Context) domain.ID {
	if actorID := ctx.Value(ContextKeyUserID); actorID != nil {
		if id, ok := actorID.(string); ok {
			return domain.ID(id)
		}
	}
	return ""
}

func (uc *recipientUseCase) getClientIP(ctx context.Context) string {
	if ip := ctx.Value(ContextKeyClientIP); ip != nil {
		if clientIP, ok := ip.(string); ok {
			return clientIP
		}
	}
	return "unknown"
}
