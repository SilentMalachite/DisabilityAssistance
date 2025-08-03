package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"shien-system/internal/domain"
)

// authUseCase implements AuthUseCase interface
type authUseCase struct {
	staffRepo      domain.StaffRepository
	auditRepo      domain.AuditLogRepository
	passwordHasher PasswordHasher
	sessionMgr     SessionManager
	rateLimitSvc   *RateLimitService
}

// NewAuthUseCase creates a new AuthUseCase instance
func NewAuthUseCase(
	staffRepo domain.StaffRepository,
	auditRepo domain.AuditLogRepository,
	passwordHasher PasswordHasher,
	sessionMgr SessionManager,
	rateLimitSvc *RateLimitService,
) AuthUseCase {
	return &authUseCase{
		staffRepo:      staffRepo,
		auditRepo:      auditRepo,
		passwordHasher: passwordHasher,
		sessionMgr:     sessionMgr,
		rateLimitSvc:   rateLimitSvc,
	}
}

// Login authenticates a user and creates a session
func (a *authUseCase) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Validate input
	if req.Username == "" || req.Password == "" {
		a.logAuditEvent(ctx, "", "LOGIN_FAILED", "AUTH", req.ClientIP, "Empty username or password")
		return nil, ErrValidationFailed
	}

	// Check rate limiting and lockout status
	if a.rateLimitSvc != nil {
		rateLimitResult, err := a.rateLimitSvc.CheckLoginAttempt(ctx, req.ClientIP, req.Username, req.UserAgent)
		if err != nil {
			a.logAuditEvent(ctx, "", "RATE_LIMIT_CHECK_FAILED", "AUTH", req.ClientIP, fmt.Sprintf("Rate limit check failed: %v", err))
			return nil, fmt.Errorf("rate limit check failed: %w", err)
		}

		if !rateLimitResult.Allowed {
			// Log the blocked attempt
			a.logAuditEvent(ctx, "", "LOGIN_BLOCKED", "AUTH", req.ClientIP,
				fmt.Sprintf("Login blocked for %s: %s", req.Username, rateLimitResult.Reason))

			// Record the blocked attempt
			if recordErr := a.rateLimitSvc.RecordLoginAttempt(ctx, req.ClientIP, req.Username, req.UserAgent, false); recordErr != nil {
				// Log but don't fail the response
				a.logAuditEvent(ctx, "", "RECORD_ATTEMPT_FAILED", "AUTH", req.ClientIP, fmt.Sprintf("Failed to record blocked attempt: %v", recordErr))
			}

			// Return appropriate error based on lockout type
			switch rateLimitResult.Reason {
			case "account_locked":
				return nil, &UseCaseError{
					Code:    "ACCOUNT_LOCKED",
					Message: "アカウントがロックされています。しばらく時間をおいてから再試行してください。",
					Cause:   ErrAccountLocked,
				}
			case "ip_locked":
				return nil, &UseCaseError{
					Code:    "IP_BLOCKED",
					Message: "このIPアドレスからのアクセスが一時的に制限されています。",
					Cause:   ErrIPBlocked,
				}
			case "account_rate_limit_exceeded", "ip_rate_limit_exceeded":
				return nil, &UseCaseError{
					Code:    "TOO_MANY_ATTEMPTS",
					Message: "ログイン試行回数が上限に達しました。しばらく時間をおいてから再試行してください。",
					Cause:   ErrTooManyAttempts,
				}
			default:
				return nil, &UseCaseError{
					Code:    "LOGIN_RESTRICTED",
					Message: "ログインが制限されています。",
					Cause:   ErrLoginRestricted,
				}
			}
		}
	}

	// Find staff by username
	staff, err := a.staffRepo.GetByExactName(ctx, req.Username)
	if err != nil {
		if err == domain.ErrNotFound {
			a.logAuditEvent(ctx, "", "LOGIN_FAILED", "AUTH", req.ClientIP, fmt.Sprintf("User not found: %s", req.Username))
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get staff by name: %w", err)
	}

	// Check password (assuming staff has a password hash field)
	if err := a.passwordHasher.CheckPassword(staff.PasswordHash, req.Password); err != nil {
		a.logAuditEvent(ctx, staff.ID, "LOGIN_FAILED", "AUTH", req.ClientIP, "Invalid password")

		// Record failed login attempt
		if a.rateLimitSvc != nil {
			if recordErr := a.rateLimitSvc.RecordLoginAttempt(ctx, req.ClientIP, req.Username, req.UserAgent, false); recordErr != nil {
				a.logAuditEvent(ctx, staff.ID, "RECORD_ATTEMPT_FAILED", "AUTH", req.ClientIP, fmt.Sprintf("Failed to record failed attempt: %v", recordErr))
			}
		}

		return nil, ErrInvalidCredentials
	}

	// Create session
	session, err := a.sessionMgr.CreateSession(ctx, staff.ID, staff.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Log successful login
	a.logAuditEvent(ctx, staff.ID, "LOGIN_SUCCESS", "AUTH", req.ClientIP, "Successful login")

	// Record successful login attempt
	if a.rateLimitSvc != nil {
		if recordErr := a.rateLimitSvc.RecordLoginAttempt(ctx, req.ClientIP, req.Username, req.UserAgent, true); recordErr != nil {
			a.logAuditEvent(ctx, staff.ID, "RECORD_ATTEMPT_FAILED", "AUTH", req.ClientIP, fmt.Sprintf("Failed to record successful attempt: %v", recordErr))
		}
	}

	return &LoginResponse{
		SessionID: session.ID,
		User:      staff,
		ExpiresAt: session.ExpiresAt,
		CSRFToken: session.CSRFToken,
	}, nil
}

// Logout ends a user session
func (a *authUseCase) Logout(ctx context.Context, req LogoutRequest) error {
	// Validate session first to get user info for audit log
	session, err := a.sessionMgr.ValidateSession(ctx, req.SessionID)
	if err != nil {
		// Even if session is invalid, we should still try to delete it
		a.sessionMgr.DeleteSession(ctx, req.SessionID)
		return ErrInvalidSession
	}

	// Delete session
	if err := a.sessionMgr.DeleteSession(ctx, req.SessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Log logout
	a.logAuditEvent(ctx, session.UserID, "LOGOUT", "AUTH", req.ClientIP, "User logged out")

	return nil
}

// ValidateSession validates a session and returns user information
func (a *authUseCase) ValidateSession(ctx context.Context, sessionID string) (*SessionInfo, error) {
	// Validate session
	session, err := a.sessionMgr.ValidateSession(ctx, sessionID)
	if err != nil {
		return nil, ErrInvalidSession
	}

	// Get user information
	staff, err := a.staffRepo.GetByID(ctx, session.UserID)
	if err != nil {
		if err == domain.ErrNotFound {
			// Clean up invalid session
			a.sessionMgr.DeleteSession(ctx, sessionID)
			return nil, ErrInvalidSession
		}
		return nil, fmt.Errorf("failed to get staff by ID: %w", err)
	}

	return &SessionInfo{
		SessionID: session.ID,
		User:      staff,
		CreatedAt: session.CreatedAt,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

// ChangePassword changes user password with validation
func (a *authUseCase) ChangePassword(ctx context.Context, req ChangePasswordRequest) error {
	// Validate input
	if req.UserID == "" || req.OldPassword == "" || req.NewPassword == "" {
		return ErrValidationFailed
	}

	// Get user
	staff, err := a.staffRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if err == domain.ErrNotFound {
			return ErrStaffNotFound
		}
		return fmt.Errorf("failed to get staff by ID: %w", err)
	}

	// Verify old password
	if err := a.passwordHasher.CheckPassword(staff.PasswordHash, req.OldPassword); err != nil {
		a.logAuditEvent(ctx, req.UserID, "PASSWORD_CHANGE_FAILED", "AUTH", req.ClientIP, "Invalid old password")
		return ErrInvalidPassword
	}

	// Hash new password
	hashedPassword, err := a.passwordHasher.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	staff.PasswordHash = hashedPassword
	staff.UpdatedAt = time.Now()

	if err := a.staffRepo.Update(ctx, staff); err != nil {
		return fmt.Errorf("failed to update staff password: %w", err)
	}

	// Log password change
	a.logAuditEvent(ctx, req.UserID, "PASSWORD_CHANGED", "AUTH", req.ClientIP, "Password changed successfully")

	return nil
}

// RefreshSession extends session expiration time
func (a *authUseCase) RefreshSession(ctx context.Context, sessionID string) (*SessionInfo, error) {
	// Refresh session
	session, err := a.sessionMgr.RefreshSession(ctx, sessionID)
	if err != nil {
		return nil, ErrInvalidSession
	}

	// Get user information
	staff, err := a.staffRepo.GetByID(ctx, session.UserID)
	if err != nil {
		if err == domain.ErrNotFound {
			// Clean up invalid session
			a.sessionMgr.DeleteSession(ctx, sessionID)
			return nil, ErrInvalidSession
		}
		return nil, fmt.Errorf("failed to get staff by ID: %w", err)
	}

	return &SessionInfo{
		SessionID: session.ID,
		User:      staff,
		CreatedAt: session.CreatedAt,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

// logAuditEvent is a helper function to log audit events
func (a *authUseCase) logAuditEvent(ctx context.Context, actorID domain.ID, action, target, clientIP, details string) {
	auditLog := &domain.AuditLog{
		ID:      uuid.New().String(),
		ActorID: actorID,
		Action:  action,
		Target:  target,
		At:      time.Now(),
		IP:      clientIP,
		Details: details,
	}

	// For testing purposes, make this synchronous
	// In production, you might want to make this asynchronous
	if err := a.auditRepo.Create(ctx, auditLog); err != nil {
		// In a production system, you might want to log this error somewhere
		// For now, we'll just ignore audit logging failures
	}
}
