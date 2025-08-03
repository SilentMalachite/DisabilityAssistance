package usecase

import (
	"context"
	"fmt"
	"shien-system/internal/domain"
)

type SetupUseCase interface {
	NeedsInitialSetup(ctx context.Context) (bool, error)
	CreateInitialAdmin(ctx context.Context, name, password string) error
}

type setupUseCase struct {
	staffRepo      domain.StaffRepository
	auditRepo      domain.AuditLogRepository
	passwordHasher domain.PasswordHasher
}

func NewSetupUseCase(
	staffRepo domain.StaffRepository,
	auditRepo domain.AuditLogRepository,
	passwordHasher domain.PasswordHasher,
) SetupUseCase {
	return &setupUseCase{
		staffRepo:      staffRepo,
		auditRepo:      auditRepo,
		passwordHasher: passwordHasher,
	}
}

func (u *setupUseCase) NeedsInitialSetup(ctx context.Context) (bool, error) {
	// Check if any admin user exists
	staffList, err := u.staffRepo.GetByRole(ctx, domain.RoleAdmin)
	if err != nil {
		return false, fmt.Errorf("checking admin existence: %w", err)
	}

	return len(staffList) == 0, nil
}

func (u *setupUseCase) CreateInitialAdmin(ctx context.Context, name, password string) error {
	// Validate password strength
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Hash password
	hashedPassword, err := u.passwordHasher.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	// Create admin user
	admin := &domain.Staff{
		ID:           "admin-001",
		Name:         name,
		Role:         domain.RoleAdmin,
		PasswordHash: hashedPassword,
	}

	if err := u.staffRepo.Create(ctx, admin); err != nil {
		return fmt.Errorf("creating admin: %w", err)
	}

	// Log the setup
	audit := &domain.AuditLog{
		ActorID: admin.ID,
		Action:  "initial_setup",
		Target:  "system",
		Details: "Initial admin account created",
	}

	if err := u.auditRepo.Create(ctx, audit); err != nil {
		// Non-critical error, just log it
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	return nil
}
