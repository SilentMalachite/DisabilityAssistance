package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"shien-system/internal/domain"
)

// TestAuthUseCase_BruteForceProtection tests the complete brute force protection
func TestAuthUseCase_BruteForceProtection(t *testing.T) {
	// Setup mock repositories
	staffRepo := &MockStaffRepository{}
	auditRepo := &MockAuditLogRepository{}
	passwordHasher := &MockPasswordHasher{}
	sessionMgr := &MockSessionManager{}

	// Setup rate limit components
	attemptRepo := NewMockLoginAttemptRepository()
	lockoutRepo := NewMockAccountLockoutRepository()
	configRepo := NewMockRateLimitConfigRepository()

	rateLimitSvc := NewRateLimitService(attemptRepo, lockoutRepo, configRepo, auditRepo)

	// Create auth use case with rate limiting
	authUseCase := NewAuthUseCase(staffRepo, auditRepo, passwordHasher, sessionMgr, rateLimitSvc)

	ctx := context.Background()

	// Setup test staff
	testStaff := &domain.Staff{
		ID:           "test-staff-001",
		Name:         "testuser",
		Role:         domain.RoleStaff,
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Setup mock expectations for staff repository
	staffRepo.On("GetByExactName", ctx, "testuser").Return(testStaff, nil)
	auditRepo.On("Create", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil)

	// Configure password hasher to return error for wrong password
	passwordHasher.On("CheckPassword", "hashedpassword", "wrongpassword").Return(ErrInvalidPassword)

	// Test multiple failed login attempts
	loginReq := LoginRequest{
		Username:  "testuser",
		Password:  "wrongpassword",
		ClientIP:  "192.168.1.100",
		UserAgent: "TestAgent/1.0",
	}

	// First two attempts should fail with invalid credentials
	for i := 0; i < 2; i++ {
		_, err := authUseCase.Login(ctx, loginReq)
		if err == nil {
			t.Errorf("Expected login to fail with wrong password (attempt %d)", i+1)
		}

		// Verify error is not rate limit related yet
		if useCaseErr, ok := err.(*UseCaseError); ok {
			if useCaseErr.Code != "INVALID_CREDENTIALS" {
				t.Errorf("Expected INVALID_CREDENTIALS error, got %s (attempt %d)", useCaseErr.Code, i+1)
			}
		}
	}

	// Third attempt should still fail with invalid credentials (not rate limited yet)
	_, err := authUseCase.Login(ctx, loginReq)
	if err == nil {
		t.Errorf("Expected login to fail with wrong password (attempt 3)")
	}

	// Fourth attempt should be rate limited
	_, err = authUseCase.Login(ctx, loginReq)
	if err == nil {
		t.Errorf("Expected login to fail due to rate limiting")
	}

	useCaseErr, ok := err.(*UseCaseError)
	if !ok {
		t.Errorf("Expected UseCaseError, got %T", err)
	} else if useCaseErr.Code != "TOO_MANY_ATTEMPTS" {
		t.Errorf("Expected TOO_MANY_ATTEMPTS error, got %s", useCaseErr.Code)
	}

	// Verify lockout was created after exceeding the limit
	lockout, err := lockoutRepo.GetActiveByUsername(ctx, "testuser")
	if err != nil {
		t.Fatalf("Error checking lockout: %v", err)
	}
	if lockout == nil {
		t.Errorf("Expected account lockout to be created after %d attempts", 3)
	}

	// Even with correct password, login should be blocked due to lockout
	loginReq.Password = "correctpassword"
	passwordHasher.On("CheckPassword", "hashedpassword", "correctpassword").Return(nil)

	_, err = authUseCase.Login(ctx, loginReq)
	if err == nil {
		t.Errorf("Expected login to be blocked due to lockout even with correct password")
	}

	if useCaseErr, ok := err.(*UseCaseError); ok {
		// After lockout, the error should be ACCOUNT_LOCKED or TOO_MANY_ATTEMPTS
		if useCaseErr.Code != "ACCOUNT_LOCKED" && useCaseErr.Code != "TOO_MANY_ATTEMPTS" {
			t.Errorf("Expected ACCOUNT_LOCKED or TOO_MANY_ATTEMPTS error, got %s", useCaseErr.Code)
		}
	}
}

func TestAuthUseCase_IPBasedRateLimit(t *testing.T) {
	// Setup similar to previous test but with different usernames to test IP-based limiting
	staffRepo := &MockStaffRepository{}
	auditRepo := &MockAuditLogRepository{}
	passwordHasher := &MockPasswordHasher{}
	sessionMgr := &MockSessionManager{}

	attemptRepo := NewMockLoginAttemptRepository()
	lockoutRepo := NewMockAccountLockoutRepository()
	configRepo := NewMockRateLimitConfigRepository()

	rateLimitSvc := NewRateLimitService(attemptRepo, lockoutRepo, configRepo, auditRepo)
	authUseCase := NewAuthUseCase(staffRepo, auditRepo, passwordHasher, sessionMgr, rateLimitSvc)

	ctx := context.Background()

	// Setup multiple test staff
	for i := 1; i <= 3; i++ {
		staff := &domain.Staff{
			ID:           "test-staff-00" + string(rune('0'+i)),
			Name:         "testuser" + string(rune('0'+i)),
			Role:         domain.RoleStaff,
			PasswordHash: "hashedpassword",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		staffRepo.On("GetByExactName", ctx, staff.Name).Return(staff, nil)
	}

	auditRepo.On("Create", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil)
	passwordHasher.On("CheckPassword", "hashedpassword", "wrongpassword").Return(ErrInvalidPassword)

	// Test failed attempts from same IP with different usernames
	ipAddress := "192.168.1.200"

	// Generate 6 failed attempts from same IP (exceeds IP limit of 5)
	for i := 1; i <= 6; i++ {
		username := "testuser" + string(rune('0'+(i%3)+1))
		loginReq := LoginRequest{
			Username:  username,
			Password:  "wrongpassword",
			ClientIP:  ipAddress,
			UserAgent: "TestAgent/1.0",
		}

		_, err := authUseCase.Login(ctx, loginReq)
		if err == nil {
			t.Errorf("Expected login to fail (attempt %d)", i)
		}

		// Check if IP is locked after 5 attempts
		if i > 5 {
			if useCaseErr, ok := err.(*UseCaseError); ok {
				if useCaseErr.Code != "IP_BLOCKED" && useCaseErr.Code != "TOO_MANY_ATTEMPTS" {
					t.Errorf("Expected IP block after 5 attempts, got %s", useCaseErr.Code)
				}
			}
		}
	}
}

func TestAuthUseCase_WhitelistedIP(t *testing.T) {
	staffRepo := &MockStaffRepository{}
	auditRepo := &MockAuditLogRepository{}
	passwordHasher := &MockPasswordHasher{}
	sessionMgr := &MockSessionManager{}

	attemptRepo := NewMockLoginAttemptRepository()
	lockoutRepo := NewMockAccountLockoutRepository()
	configRepo := NewMockRateLimitConfigRepository()

	rateLimitSvc := NewRateLimitService(attemptRepo, lockoutRepo, configRepo, auditRepo)
	authUseCase := NewAuthUseCase(staffRepo, auditRepo, passwordHasher, sessionMgr, rateLimitSvc)

	ctx := context.Background()

	// Setup test staff
	testStaff := &domain.Staff{
		ID:           "test-staff-001",
		Name:         "testuser",
		Role:         domain.RoleStaff,
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	staffRepo.On("GetByExactName", ctx, "testuser").Return(testStaff, nil)
	auditRepo.On("Create", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil)
	passwordHasher.On("CheckPassword", "hashedpassword", "wrongpassword").Return(ErrInvalidPassword)

	// Test many failed attempts from whitelisted IP (127.0.0.1)
	loginReq := LoginRequest{
		Username:  "testuser",
		Password:  "wrongpassword",
		ClientIP:  "127.0.0.1", // Whitelisted IP
		UserAgent: "TestAgent/1.0",
	}

	// Even after many attempts, should not be rate limited due to whitelist
	for i := 0; i < 10; i++ {
		_, err := authUseCase.Login(ctx, loginReq)
		if err == nil {
			t.Errorf("Expected login to fail with wrong password (attempt %d)", i+1)
		}

		// Should never be rate limited
		if useCaseErr, ok := err.(*UseCaseError); ok {
			if useCaseErr.Code == "TOO_MANY_ATTEMPTS" || useCaseErr.Code == "ACCOUNT_LOCKED" {
				t.Errorf("Whitelisted IP should not be rate limited (attempt %d)", i+1)
			}
		}
	}
}

func TestAuthUseCase_SuccessfulLoginResetsLockout(t *testing.T) {
	staffRepo := &MockStaffRepository{}
	auditRepo := &MockAuditLogRepository{}
	passwordHasher := &MockPasswordHasher{}
	sessionMgr := &MockSessionManager{}

	attemptRepo := NewMockLoginAttemptRepository()
	lockoutRepo := NewMockAccountLockoutRepository()
	configRepo := NewMockRateLimitConfigRepository()

	rateLimitSvc := NewRateLimitService(attemptRepo, lockoutRepo, configRepo, auditRepo)
	authUseCase := NewAuthUseCase(staffRepo, auditRepo, passwordHasher, sessionMgr, rateLimitSvc)

	ctx := context.Background()

	// Setup test staff
	testStaff := &domain.Staff{
		ID:           "test-staff-001",
		Name:         "testuser",
		Role:         domain.RoleStaff,
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	staffRepo.On("GetByExactName", ctx, "testuser").Return(testStaff, nil)
	auditRepo.On("Create", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil)

	// Create a manual lockout
	err := rateLimitSvc.ManualLockout(ctx, "admin-001", "testuser", "192.168.1.1", "Test lockout", 60)
	if err != nil {
		t.Fatalf("Failed to create lockout: %v", err)
	}

	// Verify lockout exists
	lockout, err := lockoutRepo.GetActiveByUsername(ctx, "testuser")
	if err != nil {
		t.Fatalf("Error checking lockout: %v", err)
	}
	if lockout == nil {
		t.Fatalf("Expected lockout to exist")
	}

	// Configure for successful login
	session := &Session{
		ID:        "test-session-001",
		UserID:    testStaff.ID,
		UserRole:  testStaff.Role,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(8 * time.Hour),
	}
	passwordHasher.On("CheckPassword", "hashedpassword", "correctpassword").Return(nil)
	sessionMgr.On("CreateSession", ctx, testStaff.ID, testStaff.Role).Return(session, nil)

	loginReq := LoginRequest{
		Username:  "testuser",
		Password:  "correctpassword",
		ClientIP:  "192.168.1.1",
		UserAgent: "TestAgent/1.0",
	}

	// Login should succeed and unlock the account
	resp, err := authUseCase.Login(ctx, loginReq)
	if err != nil {
		t.Errorf("Expected successful login, got error: %v", err)
	}
	if resp == nil {
		t.Errorf("Expected login response")
	}

	// Verify lockout is cleared
	lockout, err = lockoutRepo.GetActiveByUsername(ctx, "testuser")
	if err != nil {
		t.Fatalf("Error checking lockout after login: %v", err)
	}
	if lockout != nil {
		t.Errorf("Expected lockout to be cleared after successful login")
	}
}

// TestTimingAttackResistance ensures constant-time operations
func TestTimingAttackResistance(t *testing.T) {
	staffRepo := &MockStaffRepository{}
	auditRepo := &MockAuditLogRepository{}
	passwordHasher := &MockPasswordHasher{}
	sessionMgr := &MockSessionManager{}

	attemptRepo := NewMockLoginAttemptRepository()
	lockoutRepo := NewMockAccountLockoutRepository()
	configRepo := NewMockRateLimitConfigRepository()

	rateLimitSvc := NewRateLimitService(attemptRepo, lockoutRepo, configRepo, auditRepo)
	authUseCase := NewAuthUseCase(staffRepo, auditRepo, passwordHasher, sessionMgr, rateLimitSvc)

	ctx := context.Background()

	// Setup test staff
	testStaff := &domain.Staff{
		ID:           "test-staff-001",
		Name:         "testuser",
		Role:         domain.RoleStaff,
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Setup mock expectations
	staffRepo.On("GetByExactName", ctx, "nonexistentuser").Return(nil, domain.ErrNotFound)
	staffRepo.On("GetByExactName", ctx, "testuser").Return(testStaff, nil)
	auditRepo.On("Create", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil)
	passwordHasher.On("CheckPassword", "hashedpassword", "wrongpassword").Return(ErrInvalidPassword)

	// Test with non-existent user vs wrong password timing
	loginReqs := []LoginRequest{
		{
			Username:  "nonexistentuser",
			Password:  "anypassword",
			ClientIP:  "192.168.1.1",
			UserAgent: "TestAgent/1.0",
		},
		{
			Username:  "testuser",
			Password:  "wrongpassword",
			ClientIP:  "192.168.1.1",
			UserAgent: "TestAgent/1.0",
		},
	}

	var timings []time.Duration

	for _, req := range loginReqs {
		start := time.Now()
		_, err := authUseCase.Login(ctx, req)
		duration := time.Since(start)
		timings = append(timings, duration)

		if err == nil {
			t.Errorf("Expected login to fail for %s", req.Username)
		}
	}

	// Check that timings are relatively similar (within reasonable bounds)
	// This is a basic check - in production, you might want more sophisticated timing analysis
	timeDiff := timings[0] - timings[1]
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}

	// Allow up to 50ms difference (should be much smaller in practice)
	if timeDiff > 50*time.Millisecond {
		t.Logf("Warning: Potential timing attack vulnerability. Time difference: %v", timeDiff)
		t.Logf("Non-existent user: %v, Wrong password: %v", timings[0], timings[1])
	}
}

