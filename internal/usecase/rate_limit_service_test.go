package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"shien-system/internal/domain"
)

// MockLoginAttemptRepository for testing
type MockLoginAttemptRepository struct {
	attempts map[string][]*domain.LoginAttempt
	failures map[string]int
}

func NewMockLoginAttemptRepository() *MockLoginAttemptRepository {
	return &MockLoginAttemptRepository{
		attempts: make(map[string][]*domain.LoginAttempt),
		failures: make(map[string]int),
	}
}

func (m *MockLoginAttemptRepository) Create(ctx context.Context, attempt *domain.LoginAttempt) error {
	if attempt.ID == "" {
		attempt.ID = uuid.New().String()
	}
	key := attempt.IPAddress + ":" + attempt.Username
	m.attempts[key] = append(m.attempts[key], attempt)
	if !attempt.Success {
		m.failures[key]++
	}
	return nil
}

func (m *MockLoginAttemptRepository) GetByIPAddress(ctx context.Context, ipAddress string, since time.Time) ([]*domain.LoginAttempt, error) {
	var result []*domain.LoginAttempt
	for _, attempts := range m.attempts {
		for _, attempt := range attempts {
			if attempt.IPAddress == ipAddress && attempt.AttemptedAt.After(since) {
				result = append(result, attempt)
			}
		}
	}
	return result, nil
}

func (m *MockLoginAttemptRepository) GetByUsername(ctx context.Context, username string, since time.Time) ([]*domain.LoginAttempt, error) {
	var result []*domain.LoginAttempt
	for _, attempts := range m.attempts {
		for _, attempt := range attempts {
			if attempt.Username == username && attempt.AttemptedAt.After(since) {
				result = append(result, attempt)
			}
		}
	}
	return result, nil
}

func (m *MockLoginAttemptRepository) GetFailedAttempts(ctx context.Context, ipAddress, username string, since time.Time) ([]*domain.LoginAttempt, error) {
	var result []*domain.LoginAttempt
	for _, attempts := range m.attempts {
		for _, attempt := range attempts {
			if !attempt.Success && attempt.AttemptedAt.After(since) {
				if (ipAddress != "" && attempt.IPAddress == ipAddress) ||
					(username != "" && attempt.Username == username) {
					result = append(result, attempt)
				}
			}
		}
	}
	return result, nil
}

func (m *MockLoginAttemptRepository) CountRecentFailures(ctx context.Context, ipAddress, username string, since time.Time) (int, error) {
	count := 0
	for _, attempts := range m.attempts {
		for _, attempt := range attempts {
			if !attempt.Success && attempt.AttemptedAt.After(since) {
				if (ipAddress != "" && attempt.IPAddress == ipAddress) ||
					(username != "" && attempt.Username == username) {
					count++
				}
			}
		}
	}
	return count, nil
}

func (m *MockLoginAttemptRepository) DeleteOldAttempts(ctx context.Context, before time.Time) error {
	return nil
}

// MockAccountLockoutRepository for testing
type MockAccountLockoutRepository struct {
	lockouts map[string]*domain.AccountLockout
}

func NewMockAccountLockoutRepository() *MockAccountLockoutRepository {
	return &MockAccountLockoutRepository{
		lockouts: make(map[string]*domain.AccountLockout),
	}
}

func (m *MockAccountLockoutRepository) Create(ctx context.Context, lockout *domain.AccountLockout) error {
	if lockout.ID == "" {
		lockout.ID = uuid.New().String()
	}
	m.lockouts[lockout.ID] = lockout
	return nil
}

func (m *MockAccountLockoutRepository) GetByID(ctx context.Context, id domain.ID) (*domain.AccountLockout, error) {
	if lockout, exists := m.lockouts[id]; exists {
		return lockout, nil
	}
	return nil, domain.ErrNotFound
}

func (m *MockAccountLockoutRepository) GetActiveByUsername(ctx context.Context, username string) (*domain.AccountLockout, error) {
	for _, lockout := range m.lockouts {
		if lockout.Username == username && lockout.UnlockedAt == nil {
			// Check if expired
			if lockout.Duration > 0 {
				expiry := lockout.LockedAt.Add(time.Duration(lockout.Duration) * time.Second)
				if time.Now().After(expiry) {
					return nil, nil // Expired
				}
			}
			return lockout, nil
		}
	}
	return nil, nil
}

func (m *MockAccountLockoutRepository) GetActiveByIPAddress(ctx context.Context, ipAddress string) (*domain.AccountLockout, error) {
	for _, lockout := range m.lockouts {
		if lockout.IPAddress == ipAddress && lockout.UnlockedAt == nil {
			// Check if expired
			if lockout.Duration > 0 {
				expiry := lockout.LockedAt.Add(time.Duration(lockout.Duration) * time.Second)
				if time.Now().After(expiry) {
					return nil, nil // Expired
				}
			}
			return lockout, nil
		}
	}
	return nil, nil
}

func (m *MockAccountLockoutRepository) GetActiveLockouts(ctx context.Context) ([]*domain.AccountLockout, error) {
	var result []*domain.AccountLockout
	for _, lockout := range m.lockouts {
		if lockout.UnlockedAt == nil {
			result = append(result, lockout)
		}
	}
	return result, nil
}

func (m *MockAccountLockoutRepository) Unlock(ctx context.Context, id domain.ID, unlockedAt time.Time) error {
	if lockout, exists := m.lockouts[id]; exists {
		lockout.UnlockedAt = &unlockedAt
		return nil
	}
	return domain.ErrNotFound
}

func (m *MockAccountLockoutRepository) UnlockByUsername(ctx context.Context, username string, unlockedAt time.Time) error {
	for _, lockout := range m.lockouts {
		if lockout.Username == username && lockout.UnlockedAt == nil {
			lockout.UnlockedAt = &unlockedAt
		}
	}
	return nil
}

func (m *MockAccountLockoutRepository) UnlockByIPAddress(ctx context.Context, ipAddress string, unlockedAt time.Time) error {
	for _, lockout := range m.lockouts {
		if lockout.IPAddress == ipAddress && lockout.UnlockedAt == nil {
			lockout.UnlockedAt = &unlockedAt
		}
	}
	return nil
}

func (m *MockAccountLockoutRepository) CleanupExpiredLockouts(ctx context.Context) error {
	return nil
}

func (m *MockAccountLockoutRepository) List(ctx context.Context, limit, offset int) ([]*domain.AccountLockout, error) {
	var result []*domain.AccountLockout
	for _, lockout := range m.lockouts {
		result = append(result, lockout)
	}
	return result, nil
}

func (m *MockAccountLockoutRepository) Count(ctx context.Context) (int, error) {
	return len(m.lockouts), nil
}

// MockRateLimitConfigRepository for testing
type MockRateLimitConfigRepository struct {
	config *domain.RateLimitConfig
}

func NewMockRateLimitConfigRepository() *MockRateLimitConfigRepository {
	return &MockRateLimitConfigRepository{
		config: &domain.RateLimitConfig{
			ID:                       "test-config",
			MaxAttemptsPerIP:         5,
			MaxAttemptsPerUser:       3,
			WindowSizeMinutes:        15,
			LockoutDurationMinutes:   30,
			BackoffMultiplier:        2.0,
			MaxLockoutHours:          24,
			WhitelistIPs:             []string{"127.0.0.1"},
			EnableProgressiveLockout: true,
			CreatedAt:                time.Now(),
			UpdatedAt:                time.Now(),
		},
	}
}

func (m *MockRateLimitConfigRepository) Get(ctx context.Context) (*domain.RateLimitConfig, error) {
	return m.config, nil
}

func (m *MockRateLimitConfigRepository) Update(ctx context.Context, config *domain.RateLimitConfig) error {
	m.config = config
	return nil
}

func (m *MockRateLimitConfigRepository) GetDefault() *domain.RateLimitConfig {
	return m.config
}

// Tests

func TestRateLimitService_CheckLoginAttempt_Allowed(t *testing.T) {
	attemptRepo := NewMockLoginAttemptRepository()
	lockoutRepo := NewMockAccountLockoutRepository()
	configRepo := NewMockRateLimitConfigRepository()
	auditRepo := &MockAuditLogRepository{}

	service := NewRateLimitService(attemptRepo, lockoutRepo, configRepo, auditRepo)

	ctx := context.Background()
	result, err := service.CheckLoginAttempt(ctx, "192.168.1.1", "testuser", "TestAgent")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.Allowed {
		t.Errorf("Expected login to be allowed")
	}

	if result.Reason != "within_limits" {
		t.Errorf("Expected reason 'within_limits', got %s", result.Reason)
	}
}

func TestRateLimitService_CheckLoginAttempt_WhitelistedIP(t *testing.T) {
	attemptRepo := NewMockLoginAttemptRepository()
	lockoutRepo := NewMockAccountLockoutRepository()
	configRepo := NewMockRateLimitConfigRepository()
	auditRepo := &MockAuditLogRepository{}

	service := NewRateLimitService(attemptRepo, lockoutRepo, configRepo, auditRepo)

	ctx := context.Background()
	result, err := service.CheckLoginAttempt(ctx, "127.0.0.1", "testuser", "TestAgent")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.Allowed {
		t.Errorf("Expected login to be allowed for whitelisted IP")
	}

	if result.Reason != "whitelisted_ip" {
		t.Errorf("Expected reason 'whitelisted_ip', got %s", result.Reason)
	}
}

func TestRateLimitService_CheckLoginAttempt_IPLockout(t *testing.T) {
	attemptRepo := NewMockLoginAttemptRepository()
	lockoutRepo := NewMockAccountLockoutRepository()
	configRepo := NewMockRateLimitConfigRepository()
	auditRepo := &MockAuditLogRepository{}

	service := NewRateLimitService(attemptRepo, lockoutRepo, configRepo, auditRepo)

	ctx := context.Background()

	// Simulate exceeding IP limit
	ipAddress := "192.168.1.100"
	username := "testuser"

	// Add failed attempts to exceed limit
	for i := 0; i < 6; i++ {
		attemptRepo.Create(ctx, &domain.LoginAttempt{
			IPAddress:   ipAddress,
			Username:    username,
			Success:     false,
			AttemptedAt: time.Now(),
			UserAgent:   "TestAgent",
		})
	}

	result, err := service.CheckLoginAttempt(ctx, ipAddress, username, "TestAgent")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Allowed {
		t.Errorf("Expected login to be blocked due to IP rate limit")
	}

	if result.Reason != "ip_rate_limit_exceeded" {
		t.Errorf("Expected reason 'ip_rate_limit_exceeded', got %s", result.Reason)
	}
}

func TestRateLimitService_RecordLoginAttempt(t *testing.T) {
	attemptRepo := NewMockLoginAttemptRepository()
	lockoutRepo := NewMockAccountLockoutRepository()
	configRepo := NewMockRateLimitConfigRepository()
	auditRepo := &MockAuditLogRepository{}

	service := NewRateLimitService(attemptRepo, lockoutRepo, configRepo, auditRepo)

	ctx := context.Background()

	err := service.RecordLoginAttempt(ctx, "192.168.1.1", "testuser", "TestAgent", true)
	if err != nil {
		t.Fatalf("Expected no error recording login attempt, got %v", err)
	}

	// Verify attempt was recorded
	attempts, err := attemptRepo.GetByIPAddress(ctx, "192.168.1.1", time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatalf("Expected no error getting attempts, got %v", err)
	}

	if len(attempts) != 1 {
		t.Errorf("Expected 1 attempt recorded, got %d", len(attempts))
	}

	if !attempts[0].Success {
		t.Errorf("Expected successful attempt")
	}
}

func TestRateLimitService_ManualLockout(t *testing.T) {
	attemptRepo := NewMockLoginAttemptRepository()
	lockoutRepo := NewMockAccountLockoutRepository()
	configRepo := NewMockRateLimitConfigRepository()
	auditRepo := &MockAuditLogRepository{}

	service := NewRateLimitService(attemptRepo, lockoutRepo, configRepo, auditRepo)

	ctx := context.Background()

	err := service.ManualLockout(ctx, "admin-001", "testuser", "192.168.1.1", "Security violation", 60)
	if err != nil {
		t.Fatalf("Expected no error creating manual lockout, got %v", err)
	}

	// Verify lockout was created
	lockout, err := lockoutRepo.GetActiveByUsername(ctx, "testuser")
	if err != nil {
		t.Fatalf("Expected no error getting lockout, got %v", err)
	}

	if lockout == nil {
		t.Fatalf("Expected lockout to be created")
	}

	if lockout.LockoutType != domain.LockoutTypeManual {
		t.Errorf("Expected manual lockout type, got %s", lockout.LockoutType)
	}

	if lockout.Reason != "Security violation" {
		t.Errorf("Expected reason 'Security violation', got %s", lockout.Reason)
	}
}

func TestRateLimitService_UnlockAccount(t *testing.T) {
	attemptRepo := NewMockLoginAttemptRepository()
	lockoutRepo := NewMockAccountLockoutRepository()
	configRepo := NewMockRateLimitConfigRepository()
	auditRepo := &MockAuditLogRepository{}

	service := NewRateLimitService(attemptRepo, lockoutRepo, configRepo, auditRepo)

	ctx := context.Background()

	// Create a lockout first
	lockout := &domain.AccountLockout{
		ID:          "test-lockout",
		Username:    "testuser",
		IPAddress:   "192.168.1.1",
		LockoutType: domain.LockoutTypeAccount,
		LockedAt:    time.Now(),
		Reason:      "Test lockout",
		Duration:    3600,
	}
	lockoutRepo.Create(ctx, lockout)

	// Unlock the account
	err := service.UnlockAccount(ctx, "admin-001", "testuser")
	if err != nil {
		t.Fatalf("Expected no error unlocking account, got %v", err)
	}

	// Verify account is unlocked
	activeLockout, err := lockoutRepo.GetActiveByUsername(ctx, "testuser")
	if err != nil {
		t.Fatalf("Expected no error checking lockout, got %v", err)
	}

	if activeLockout != nil {
		t.Errorf("Expected no active lockout after unlock")
	}
}
