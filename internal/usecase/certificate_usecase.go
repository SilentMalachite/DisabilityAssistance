package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"shien-system/internal/domain"
)

// certificateUseCase implements CertificateUseCase interface
type certificateUseCase struct {
	certRepo      domain.BenefitCertificateRepository
	recipientRepo domain.RecipientRepository
	staffRepo     domain.StaffRepository
	auditRepo     domain.AuditLogRepository
}

// NewCertificateUseCase creates a new certificate usecase
func NewCertificateUseCase(
	certRepo domain.BenefitCertificateRepository,
	recipientRepo domain.RecipientRepository,
	staffRepo domain.StaffRepository,
	auditRepo domain.AuditLogRepository,
) CertificateUseCase {
	return &certificateUseCase{
		certRepo:      certRepo,
		recipientRepo: recipientRepo,
		staffRepo:     staffRepo,
		auditRepo:     auditRepo,
	}
}

// CreateCertificate creates a new benefit certificate
func (uc *certificateUseCase) CreateCertificate(ctx context.Context, req CreateCertificateRequest) (*domain.BenefitCertificate, error) {
	// Validate input
	if err := uc.validateCreateCertificateRequest(req); err != nil {
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

	// Verify recipient exists
	_, err = uc.recipientRepo.GetByID(ctx, req.RecipientID)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, ErrRecipientNotFound
		}
		return nil, &UseCaseError{
			Code:    "INTERNAL_ERROR",
			Message: "内部エラーが発生しました",
			Cause:   err,
		}
	}

	// Create certificate
	now := time.Now().UTC()
	certificate := &domain.BenefitCertificate{
		ID:                     domain.ID(uuid.New().String()),
		RecipientID:            req.RecipientID,
		StartDate:              req.StartDate,
		EndDate:                req.EndDate,
		Issuer:                 req.Issuer,
		ServiceType:            req.ServiceType,
		MaxBenefitDaysPerMonth: req.MaxBenefitDaysPerMonth,
		BenefitDetails:         req.BenefitDetails,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	err = uc.certRepo.Create(ctx, certificate)
	if err != nil {
		return nil, &UseCaseError{
			Code:    "CREATION_FAILED",
			Message: "受給者証の作成に失敗しました",
			Cause:   err,
		}
	}

	// Log the action
	auditLog := &domain.AuditLog{
		ID:      domain.ID(uuid.New().String()),
		ActorID: req.ActorID,
		Action:  "CREATE",
		Target:  fmt.Sprintf("certificate:%s", certificate.ID),
		At:      now,
		IP:      uc.getClientIP(ctx),
		Details: fmt.Sprintf("受給者証を作成しました (サービス種別: %s, 有効期限: %s)",
			certificate.ServiceType, certificate.EndDate.Format("2006-01-02")),
	}

	err = uc.auditRepo.Create(ctx, auditLog)
	if err != nil {
		// Log audit failure but don't fail the operation
	}

	return certificate, nil
}

// GetCertificate retrieves a certificate by ID
func (uc *certificateUseCase) GetCertificate(ctx context.Context, id domain.ID) (*domain.BenefitCertificate, error) {
	certificate, err := uc.certRepo.GetByID(ctx, id)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, ErrCertificateNotFound
		}
		return nil, &UseCaseError{
			Code:    "RETRIEVAL_FAILED",
			Message: "受給者証の取得に失敗しました",
			Cause:   err,
		}
	}

	return certificate, nil
}

// UpdateCertificate updates certificate information
func (uc *certificateUseCase) UpdateCertificate(ctx context.Context, req UpdateCertificateRequest) (*domain.BenefitCertificate, error) {
	// Validate input
	if err := uc.validateUpdateCertificateRequest(req); err != nil {
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

	// Get existing certificate
	existing, err := uc.certRepo.GetByID(ctx, req.ID)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, ErrCertificateNotFound
		}
		return nil, &UseCaseError{
			Code:    "RETRIEVAL_FAILED",
			Message: "受給者証の取得に失敗しました",
			Cause:   err,
		}
	}

	// Update certificate
	now := time.Now().UTC()
	certificate := &domain.BenefitCertificate{
		ID:                     req.ID,
		RecipientID:            existing.RecipientID, // Cannot change recipient
		StartDate:              req.StartDate,
		EndDate:                req.EndDate,
		Issuer:                 req.Issuer,
		ServiceType:            req.ServiceType,
		MaxBenefitDaysPerMonth: req.MaxBenefitDaysPerMonth,
		BenefitDetails:         req.BenefitDetails,
		CreatedAt:              existing.CreatedAt, // Preserve original creation time
		UpdatedAt:              now,
	}

	err = uc.certRepo.Update(ctx, certificate)
	if err != nil {
		return nil, &UseCaseError{
			Code:    "UPDATE_FAILED",
			Message: "受給者証の更新に失敗しました",
			Cause:   err,
		}
	}

	// Log the action
	auditLog := &domain.AuditLog{
		ID:      domain.ID(uuid.New().String()),
		ActorID: req.ActorID,
		Action:  "UPDATE",
		Target:  fmt.Sprintf("certificate:%s", certificate.ID),
		At:      now,
		IP:      uc.getClientIP(ctx),
		Details: fmt.Sprintf("受給者証を更新しました (サービス種別: %s, 有効期限: %s)",
			certificate.ServiceType, certificate.EndDate.Format("2006-01-02")),
	}

	err = uc.auditRepo.Create(ctx, auditLog)
	if err != nil {
		// Log audit failure but don't fail the operation
	}

	return certificate, nil
}

// DeleteCertificate deletes a certificate
func (uc *certificateUseCase) DeleteCertificate(ctx context.Context, id domain.ID) error {
	// Get existing certificate
	certificate, err := uc.certRepo.GetByID(ctx, id)
	if err != nil {
		if err == domain.ErrNotFound {
			return ErrCertificateNotFound
		}
		return &UseCaseError{
			Code:    "RETRIEVAL_FAILED",
			Message: "受給者証の取得に失敗しました",
			Cause:   err,
		}
	}

	// Delete certificate
	err = uc.certRepo.Delete(ctx, id)
	if err != nil {
		return &UseCaseError{
			Code:    "DELETION_FAILED",
			Message: "受給者証の削除に失敗しました",
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
			Target:  fmt.Sprintf("certificate:%s", id),
			At:      time.Now().UTC(),
			IP:      uc.getClientIP(ctx),
			Details: fmt.Sprintf("受給者証を削除しました (サービス種別: %s)", certificate.ServiceType),
		}

		uc.auditRepo.Create(ctx, auditLog)
	}

	return nil
}

// GetCertificatesByRecipient retrieves all certificates for a recipient
func (uc *certificateUseCase) GetCertificatesByRecipient(ctx context.Context, recipientID domain.ID) ([]*domain.BenefitCertificate, error) {
	// Verify recipient exists
	_, err := uc.recipientRepo.GetByID(ctx, recipientID)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, ErrRecipientNotFound
		}
		return nil, &UseCaseError{
			Code:    "INTERNAL_ERROR",
			Message: "内部エラーが発生しました",
			Cause:   err,
		}
	}

	certificates, err := uc.certRepo.GetByRecipientID(ctx, recipientID)
	if err != nil {
		return nil, &UseCaseError{
			Code:    "RETRIEVAL_FAILED",
			Message: "受給者証一覧の取得に失敗しました",
			Cause:   err,
		}
	}

	return certificates, nil
}

// GetExpiringSoon retrieves certificates expiring soon
func (uc *certificateUseCase) GetExpiringSoon(ctx context.Context, days int) ([]*domain.BenefitCertificate, error) {
	if days <= 0 {
		return nil, &UseCaseError{
			Code:    "VALIDATION_FAILED",
			Message: "日数は正の値である必要があります",
		}
	}

	within := time.Duration(days) * 24 * time.Hour
	certificates, err := uc.certRepo.GetExpiringSoon(ctx, within)
	if err != nil {
		return nil, &UseCaseError{
			Code:    "RETRIEVAL_FAILED",
			Message: "期限切れ間近の受給者証の取得に失敗しました",
			Cause:   err,
		}
	}

	return certificates, nil
}

// ValidateCertificate checks if a certificate is valid for a given date
func (uc *certificateUseCase) ValidateCertificate(ctx context.Context, certificateID domain.ID, date time.Time) (*ValidationResult, error) {
	certificate, err := uc.certRepo.GetByID(ctx, certificateID)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, ErrCertificateNotFound
		}
		return nil, &UseCaseError{
			Code:    "RETRIEVAL_FAILED",
			Message: "受給者証の取得に失敗しました",
			Cause:   err,
		}
	}

	result := &ValidationResult{}

	// Check if date is within certificate validity period
	if date.Before(certificate.StartDate) {
		result.IsValid = false
		result.Reason = fmt.Sprintf("受給者証の開始日(%s)より前の日付です", certificate.StartDate.Format("2006-01-02"))
		return result, nil
	}

	if date.After(certificate.EndDate) {
		result.IsValid = false
		result.Reason = fmt.Sprintf("受給者証の終了日(%s)を過ぎています", certificate.EndDate.Format("2006-01-02"))
		expiry := certificate.EndDate
		result.ExpiresAt = &expiry
		return result, nil
	}

	// Certificate is valid
	result.IsValid = true
	result.Reason = "受給者証は有効です"
	expiry := certificate.EndDate
	result.ExpiresAt = &expiry

	return result, nil
}

// Validation functions

func (uc *certificateUseCase) validateCreateCertificateRequest(req CreateCertificateRequest) error {
	var errors []string

	if req.RecipientID == "" {
		errors = append(errors, "利用者IDは必須です")
	}

	if req.StartDate.IsZero() {
		errors = append(errors, "開始日は必須です")
	}

	if req.EndDate.IsZero() {
		errors = append(errors, "終了日は必須です")
	}

	if !req.StartDate.IsZero() && !req.EndDate.IsZero() && !req.StartDate.Before(req.EndDate) {
		errors = append(errors, "開始日は終了日より前である必要があります")
	}

	if strings.TrimSpace(req.Issuer) == "" {
		errors = append(errors, "発行者は必須です")
	}

	if strings.TrimSpace(req.ServiceType) == "" {
		errors = append(errors, "サービス種別は必須です")
	}

	if req.MaxBenefitDaysPerMonth <= 0 {
		errors = append(errors, "月間最大給付日数は正の値である必要があります")
	}

	if req.ActorID == "" {
		errors = append(errors, "実行者IDは必須です")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, ", "))
	}

	return nil
}

func (uc *certificateUseCase) validateUpdateCertificateRequest(req UpdateCertificateRequest) error {
	var errors []string

	if req.ID == "" {
		errors = append(errors, "IDは必須です")
	}

	if req.StartDate.IsZero() {
		errors = append(errors, "開始日は必須です")
	}

	if req.EndDate.IsZero() {
		errors = append(errors, "終了日は必須です")
	}

	if !req.StartDate.IsZero() && !req.EndDate.IsZero() && !req.StartDate.Before(req.EndDate) {
		errors = append(errors, "開始日は終了日より前である必要があります")
	}

	if strings.TrimSpace(req.Issuer) == "" {
		errors = append(errors, "発行者は必須です")
	}

	if strings.TrimSpace(req.ServiceType) == "" {
		errors = append(errors, "サービス種別は必須です")
	}

	if req.MaxBenefitDaysPerMonth <= 0 {
		errors = append(errors, "月間最大給付日数は正の値である必要があります")
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

func (uc *certificateUseCase) getActorID(ctx context.Context) domain.ID {
	if actorID := ctx.Value(ContextKeyUserID); actorID != nil {
		if id, ok := actorID.(string); ok {
			return domain.ID(id)
		}
	}
	return ""
}

func (uc *certificateUseCase) getClientIP(ctx context.Context) string {
	if ip := ctx.Value(ContextKeyClientIP); ip != nil {
		if clientIP, ok := ip.(string); ok {
			return clientIP
		}
	}
	return "unknown"
}
