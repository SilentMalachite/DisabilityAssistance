package usecase

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"

	"shien-system/internal/domain"
)

// RateLimitService provides brute force attack protection
type RateLimitService struct {
	attemptRepo domain.LoginAttemptRepository
	lockoutRepo domain.AccountLockoutRepository
	configRepo  domain.RateLimitConfigRepository
	auditRepo   domain.AuditLogRepository
}

// NewRateLimitService creates a new RateLimitService instance
func NewRateLimitService(
	attemptRepo domain.LoginAttemptRepository,
	lockoutRepo domain.AccountLockoutRepository,
	configRepo domain.RateLimitConfigRepository,
	auditRepo domain.AuditLogRepository,
) *RateLimitService {
	return &RateLimitService{
		attemptRepo: attemptRepo,
		lockoutRepo: lockoutRepo,
		configRepo:  configRepo,
		auditRepo:   auditRepo,
	}
}

// LoginAttemptResult represents the result of a login attempt check
type LoginAttemptResult struct {
	Allowed      bool
	LockoutType  string
	LockoutUntil *time.Time
	Reason       string
	AttemptID    string
}

// CheckLoginAttempt validates if a login attempt should be allowed
func (s *RateLimitService) CheckLoginAttempt(ctx context.Context, ipAddress, username, userAgent string) (*LoginAttemptResult, error) {
	// Get current configuration
	config, err := s.configRepo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit config: %w", err)
	}

	// Check if IP is whitelisted
	if s.isWhitelistedIP(ipAddress, config.WhitelistIPs) {
		return &LoginAttemptResult{
			Allowed: true,
			Reason:  "whitelisted_ip",
		}, nil
	}

	// Check for existing account lockouts
	userLockout, err := s.lockoutRepo.GetActiveByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check user lockout: %w", err)
	}

	if userLockout != nil {
		lockoutUntil := s.calculateLockoutExpiration(userLockout)
		return &LoginAttemptResult{
			Allowed:      false,
			LockoutType:  string(userLockout.LockoutType),
			LockoutUntil: lockoutUntil,
			Reason:       "account_locked",
		}, nil
	}

	ipLockout, err := s.lockoutRepo.GetActiveByIPAddress(ctx, ipAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to check IP lockout: %w", err)
	}

	if ipLockout != nil {
		lockoutUntil := s.calculateLockoutExpiration(ipLockout)
		return &LoginAttemptResult{
			Allowed:      false,
			LockoutType:  string(ipLockout.LockoutType),
			LockoutUntil: lockoutUntil,
			Reason:       "ip_locked",
		}, nil
	}

	// Check recent failure rates
	windowStart := time.Now().Add(-time.Duration(config.WindowSizeMinutes) * time.Minute)

	// Count recent failures for this IP
	ipFailures, err := s.attemptRepo.CountRecentFailures(ctx, ipAddress, "", windowStart)
	if err != nil {
		return nil, fmt.Errorf("failed to count IP failures: %w", err)
	}

	// Count recent failures for this username
	userFailures, err := s.attemptRepo.CountRecentFailures(ctx, "", username, windowStart)
	if err != nil {
		return nil, fmt.Errorf("failed to count user failures: %w", err)
	}

	// Check if limits are exceeded
	if ipFailures >= config.MaxAttemptsPerIP {
		// Create IP lockout
		lockout := &domain.AccountLockout{
			ID:           uuid.New().String(),
			Username:     username,
			IPAddress:    ipAddress,
			LockoutType:  domain.LockoutTypeIP,
			LockedAt:     time.Now(),
			Reason:       fmt.Sprintf("IP exceeded %d failed attempts", config.MaxAttemptsPerIP),
			FailureCount: ipFailures,
			Duration:     config.LockoutDurationMinutes * 60,
		}

		if err := s.lockoutRepo.Create(ctx, lockout); err != nil {
			return nil, fmt.Errorf("failed to create IP lockout: %w", err)
		}

		// Log security event
		s.logSecurityEvent(ctx, "", "IP_LOCKOUT_TRIGGERED", ipAddress,
			fmt.Sprintf("IP locked after %d failed attempts", ipFailures))

		lockoutUntil := s.calculateLockoutExpiration(lockout)
		return &LoginAttemptResult{
			Allowed:      false,
			LockoutType:  string(lockout.LockoutType),
			LockoutUntil: lockoutUntil,
			Reason:       "ip_rate_limit_exceeded",
		}, nil
	}

	if userFailures >= config.MaxAttemptsPerUser {
		// Calculate progressive lockout duration if enabled
		duration := config.LockoutDurationMinutes * 60
		if config.EnableProgressiveLockout {
			duration = s.calculateProgressiveLockoutDuration(ctx, username, config)
		}

		// Create account lockout
		lockout := &domain.AccountLockout{
			ID:           uuid.New().String(),
			Username:     username,
			IPAddress:    ipAddress,
			LockoutType:  domain.LockoutTypeAccount,
			LockedAt:     time.Now(),
			Reason:       fmt.Sprintf("Account exceeded %d failed attempts", config.MaxAttemptsPerUser),
			FailureCount: userFailures,
			Duration:     duration,
		}

		if err := s.lockoutRepo.Create(ctx, lockout); err != nil {
			return nil, fmt.Errorf("failed to create account lockout: %w", err)
		}

		// Log security event
		s.logSecurityEvent(ctx, username, "ACCOUNT_LOCKOUT_TRIGGERED", ipAddress,
			fmt.Sprintf("Account locked after %d failed attempts", userFailures))

		lockoutUntil := s.calculateLockoutExpiration(lockout)
		return &LoginAttemptResult{
			Allowed:      false,
			LockoutType:  string(lockout.LockoutType),
			LockoutUntil: lockoutUntil,
			Reason:       "account_rate_limit_exceeded",
		}, nil
	}

	// Attempt is allowed
	return &LoginAttemptResult{
		Allowed: true,
		Reason:  "within_limits",
	}, nil
}

// RecordLoginAttempt records a login attempt
func (s *RateLimitService) RecordLoginAttempt(ctx context.Context, ipAddress, username, userAgent string, success bool) error {
	attempt := &domain.LoginAttempt{
		ID:          uuid.New().String(),
		IPAddress:   ipAddress,
		Username:    username,
		Success:     success,
		AttemptedAt: time.Now(),
		UserAgent:   userAgent,
	}

	if err := s.attemptRepo.Create(ctx, attempt); err != nil {
		return fmt.Errorf("failed to record login attempt: %w", err)
	}

	// If this was a successful login, we should unlock any existing lockouts for this user
	if success {
		now := time.Now()
		if err := s.lockoutRepo.UnlockByUsername(ctx, username, now); err != nil {
			// Log but don't fail - the login was successful
			s.logSecurityEvent(ctx, username, "LOCKOUT_UNLOCK_FAILED", ipAddress,
				fmt.Sprintf("Failed to unlock account after successful login: %v", err))
		}
	}

	return nil
}

// DetectAttackPatterns analyzes login attempts to detect attack patterns
func (s *RateLimitService) DetectAttackPatterns(ctx context.Context) error {
	config, err := s.configRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get rate limit config: %w", err)
	}

	// Analyze recent login attempts for patterns
	windowStart := time.Now().Add(-time.Duration(config.WindowSizeMinutes*2) * time.Minute)

	// Detect distributed brute force attacks (multiple IPs, same username)
	if err := s.detectDistributedAttacks(ctx, windowStart); err != nil {
		return fmt.Errorf("failed to detect distributed attacks: %w", err)
	}

	// Detect credential stuffing (same IP, multiple usernames)
	if err := s.detectCredentialStuffing(ctx, windowStart); err != nil {
		return fmt.Errorf("failed to detect credential stuffing: %w", err)
	}

	return nil
}

// ManualLockout manually locks an account (admin function)
func (s *RateLimitService) ManualLockout(ctx context.Context, actorID, username, ipAddress, reason string, durationMinutes int) error {
	lockout := &domain.AccountLockout{
		ID:           uuid.New().String(),
		Username:     username,
		IPAddress:    ipAddress,
		LockoutType:  domain.LockoutTypeManual,
		LockedAt:     time.Now(),
		Reason:       reason,
		FailureCount: 0,
		Duration:     durationMinutes * 60,
	}

	if err := s.lockoutRepo.Create(ctx, lockout); err != nil {
		return fmt.Errorf("failed to create manual lockout: %w", err)
	}

	// Log admin action
	s.logSecurityEvent(ctx, actorID, "MANUAL_LOCKOUT_CREATED", ipAddress,
		fmt.Sprintf("Manual lockout created for %s: %s", username, reason))

	return nil
}

// UnlockAccount manually unlocks an account (admin function)
func (s *RateLimitService) UnlockAccount(ctx context.Context, actorID, username string) error {
	now := time.Now()
	if err := s.lockoutRepo.UnlockByUsername(ctx, username, now); err != nil {
		return fmt.Errorf("failed to unlock account: %w", err)
	}

	// Log admin action
	s.logSecurityEvent(ctx, actorID, "MANUAL_UNLOCK_PERFORMED", "",
		fmt.Sprintf("Manual unlock performed for %s", username))

	return nil
}

// CleanupOldRecords removes old login attempts and expired lockouts
func (s *RateLimitService) CleanupOldRecords(ctx context.Context) error {
	// Clean up login attempts older than 30 days
	cutoff := time.Now().Add(-30 * 24 * time.Hour)
	if err := s.attemptRepo.DeleteOldAttempts(ctx, cutoff); err != nil {
		return fmt.Errorf("failed to cleanup old attempts: %w", err)
	}

	// Clean up expired lockouts
	if err := s.lockoutRepo.CleanupExpiredLockouts(ctx); err != nil {
		return fmt.Errorf("failed to cleanup expired lockouts: %w", err)
	}

	return nil
}

// Helper methods

func (s *RateLimitService) isWhitelistedIP(ipAddress string, whitelist []string) bool {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return false
	}

	for _, whitelistEntry := range whitelist {
		// Check if it's a CIDR range
		if strings.Contains(whitelistEntry, "/") {
			_, cidr, err := net.ParseCIDR(whitelistEntry)
			if err == nil && cidr.Contains(ip) {
				return true
			}
		} else {
			// Check exact IP match
			if whitelistEntry == ipAddress {
				return true
			}
		}
	}

	return false
}

func (s *RateLimitService) calculateLockoutExpiration(lockout *domain.AccountLockout) *time.Time {
	if lockout.Duration <= 0 {
		return nil // Permanent lockout
	}

	expiration := lockout.LockedAt.Add(time.Duration(lockout.Duration) * time.Second)
	return &expiration
}

func (s *RateLimitService) calculateProgressiveLockoutDuration(ctx context.Context, username string, config *domain.RateLimitConfig) int {
	// Count previous lockouts for this user in the last 24 hours
	_ = time.Now().Add(-24 * time.Hour) // For future use

	// This is a simplified implementation - in a real system you might want to
	// query the lockout history more precisely
	baseDuration := config.LockoutDurationMinutes * 60
	maxDuration := config.MaxLockoutHours * 3600

	// Apply exponential backoff
	progressiveDuration := int(float64(baseDuration) * config.BackoffMultiplier)

	if progressiveDuration > maxDuration {
		progressiveDuration = maxDuration
	}

	return progressiveDuration
}

func (s *RateLimitService) detectDistributedAttacks(ctx context.Context, since time.Time) error {
	// This is a placeholder for distributed attack detection logic
	// In a real implementation, you would analyze login attempts to find patterns
	// like multiple IPs targeting the same username
	return nil
}

func (s *RateLimitService) detectCredentialStuffing(ctx context.Context, since time.Time) error {
	// This is a placeholder for credential stuffing detection logic
	// In a real implementation, you would analyze login attempts to find patterns
	// like the same IP trying multiple usernames
	return nil
}

func (s *RateLimitService) logSecurityEvent(ctx context.Context, actorID, action, ipAddress, details string) {
	auditLog := &domain.AuditLog{
		ID:      uuid.New().String(),
		ActorID: actorID,
		Action:  action,
		Target:  "SECURITY",
		At:      time.Now(),
		IP:      ipAddress,
		Details: details,
	}

	// Asynchronously log the security event
	go func() {
		if err := s.auditRepo.Create(context.Background(), auditLog); err != nil {
			// In production, you might want to log this error to a different system
		}
	}()
}
