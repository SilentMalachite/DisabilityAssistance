package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"shien-system/internal/adapter/crypto"
	"shien-system/internal/adapter/db"
	"shien-system/internal/adapter/session"
	"shien-system/internal/domain"
	"shien-system/internal/usecase"
)

func TestAuthenticationIntegration_CompleteFlow(t *testing.T) {
	// Initialize all components
	database := setupTestDatabase(t)
	defer database.Close()

	staffRepo := db.NewStaffRepository(database)
	auditRepo := db.NewAuditLogRepository(database)
	passwordHasher := crypto.NewBcryptPasswordHasher()
	sessionManager := session.NewMemorySessionManager(8 * time.Hour)

	// Initialize rate limiting components
	attemptRepo := db.NewLoginAttemptRepository(database)
	lockoutRepo := db.NewAccountLockoutRepository(database)
	configRepo := db.NewRateLimitConfigRepository(database)
	
	// Initialize rate limit service
	rateLimitSvc := usecase.NewRateLimitService(
		attemptRepo,
		lockoutRepo,
		configRepo,
		auditRepo,
	)

	authUseCase := usecase.NewAuthUseCase(staffRepo, auditRepo, passwordHasher, sessionManager, rateLimitSvc)

	ctx := context.Background()

	// Step 1: Create a test user with hashed password
	testPassword := "TestPassword123!"
	hashedPassword, err := passwordHasher.HashPassword(testPassword)
	require.NoError(t, err)

	testStaff := &domain.Staff{
		ID:           "test-staff-001",
		Name:         "テスト職員",
		Role:         domain.RoleStaff,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = staffRepo.Create(ctx, testStaff)
	require.NoError(t, err)

	// Step 2: Test successful login
	loginReq := usecase.LoginRequest{
		Username: "テスト職員",
		Password: testPassword,
		ClientIP: "192.168.1.100",
	}

	loginResp, err := authUseCase.Login(ctx, loginReq)
	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.SessionID)
	assert.Equal(t, testStaff.ID, loginResp.User.ID)
	assert.Equal(t, testStaff.Name, loginResp.User.Name)
	assert.Equal(t, testStaff.Role, loginResp.User.Role)
	assert.True(t, loginResp.ExpiresAt.After(time.Now().Add(7*time.Hour)))

	sessionID := loginResp.SessionID

	// Step 3: Test session validation
	sessionInfo, err := authUseCase.ValidateSession(ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, sessionID, sessionInfo.SessionID)
	assert.Equal(t, testStaff.ID, sessionInfo.User.ID)
	assert.Equal(t, testStaff.Name, sessionInfo.User.Name)
	assert.Equal(t, testStaff.Role, sessionInfo.User.Role)

	// Step 4: Test password change
	newPassword := "NewTestPassword456!"
	changePasswordReq := usecase.ChangePasswordRequest{
		UserID:      testStaff.ID,
		OldPassword: testPassword,
		NewPassword: newPassword,
		ClientIP:    "192.168.1.100",
	}

	err = authUseCase.ChangePassword(ctx, changePasswordReq)
	require.NoError(t, err)

	// Step 5: Test login with new password
	loginReq2 := usecase.LoginRequest{
		Username: "テスト職員",
		Password: newPassword,
		ClientIP: "192.168.1.100",
	}

	loginResp2, err := authUseCase.Login(ctx, loginReq2)
	require.NoError(t, err)
	assert.NotEmpty(t, loginResp2.SessionID)
	assert.NotEqual(t, sessionID, loginResp2.SessionID) // Should be a new session

	// Step 6: Test login with old password fails
	loginReq3 := usecase.LoginRequest{
		Username: "テスト職員",
		Password: testPassword, // Old password
		ClientIP: "192.168.1.100",
	}

	loginResp3, err := authUseCase.Login(ctx, loginReq3)
	assert.Error(t, err)
	assert.Nil(t, loginResp3)
	assert.Equal(t, usecase.ErrInvalidCredentials, err)

	// Step 7: Test session refresh
	refreshedSession, err := authUseCase.RefreshSession(ctx, loginResp2.SessionID)
	require.NoError(t, err)
	assert.Equal(t, loginResp2.SessionID, refreshedSession.SessionID)
	assert.True(t, refreshedSession.ExpiresAt.After(loginResp2.ExpiresAt))

	// Step 8: Test logout
	logoutReq := usecase.LogoutRequest{
		SessionID: loginResp2.SessionID,
		ClientIP:  "192.168.1.100",
	}

	err = authUseCase.Logout(ctx, logoutReq)
	require.NoError(t, err)

	// Step 9: Test session validation after logout fails
	sessionInfo2, err := authUseCase.ValidateSession(ctx, loginResp2.SessionID)
	assert.Error(t, err)
	assert.Nil(t, sessionInfo2)
	assert.Equal(t, usecase.ErrInvalidSession, err)
}

func TestAuthenticationIntegration_InvalidCredentials(t *testing.T) {
	// Initialize all components
	database := setupTestDatabase(t)
	defer database.Close()

	staffRepo := db.NewStaffRepository(database)
	auditRepo := db.NewAuditLogRepository(database)
	passwordHasher := crypto.NewBcryptPasswordHasher()
	sessionManager := session.NewMemorySessionManager(8 * time.Hour)

	// Initialize rate limiting components
	attemptRepo := db.NewLoginAttemptRepository(database)
	lockoutRepo := db.NewAccountLockoutRepository(database)
	configRepo := db.NewRateLimitConfigRepository(database)
	
	// Initialize rate limit service
	rateLimitSvc := usecase.NewRateLimitService(
		attemptRepo,
		lockoutRepo,
		configRepo,
		auditRepo,
	)

	authUseCase := usecase.NewAuthUseCase(staffRepo, auditRepo, passwordHasher, sessionManager, rateLimitSvc)

	ctx := context.Background()

	// Test login with non-existent user
	loginReq := usecase.LoginRequest{
		Username: "non-existent-user",
		Password: "password123",
		ClientIP: "192.168.1.100",
	}

	loginResp, err := authUseCase.Login(ctx, loginReq)
	assert.Error(t, err)
	assert.Nil(t, loginResp)
	assert.Equal(t, usecase.ErrInvalidCredentials, err)
}

func TestAuthenticationIntegration_SessionExpiry(t *testing.T) {
	// Initialize components with very short session expiry
	database := setupTestDatabase(t)
	defer database.Close()

	staffRepo := db.NewStaffRepository(database)
	auditRepo := db.NewAuditLogRepository(database)
	passwordHasher := crypto.NewBcryptPasswordHasher()
	sessionManager := session.NewMemorySessionManager(100 * time.Millisecond) // Very short

	// Initialize rate limiting components
	attemptRepo := db.NewLoginAttemptRepository(database)
	lockoutRepo := db.NewAccountLockoutRepository(database)
	configRepo := db.NewRateLimitConfigRepository(database)
	
	// Initialize rate limit service
	rateLimitSvc := usecase.NewRateLimitService(
		attemptRepo,
		lockoutRepo,
		configRepo,
		auditRepo,
	)

	authUseCase := usecase.NewAuthUseCase(staffRepo, auditRepo, passwordHasher, sessionManager, rateLimitSvc)

	ctx := context.Background()

	// Create test user
	testPassword := "TestPassword123!"
	hashedPassword, err := passwordHasher.HashPassword(testPassword)
	require.NoError(t, err)

	testStaff := &domain.Staff{
		ID:           "test-staff-002",
		Name:         "テスト職員2",
		Role:         domain.RoleAdmin,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = staffRepo.Create(ctx, testStaff)
	require.NoError(t, err)

	// Login
	loginReq := usecase.LoginRequest{
		Username: "テスト職員2",
		Password: testPassword,
		ClientIP: "192.168.1.100",
	}

	loginResp, err := authUseCase.Login(ctx, loginReq)
	require.NoError(t, err)

	// Wait for session to expire
	time.Sleep(200 * time.Millisecond)

	// Try to validate expired session
	sessionInfo, err := authUseCase.ValidateSession(ctx, loginResp.SessionID)
	assert.Error(t, err)
	assert.Nil(t, sessionInfo)
	assert.Equal(t, usecase.ErrInvalidSession, err) // Session should be cleaned up automatically
}

func TestAuthenticationIntegration_MultipleUsers(t *testing.T) {
	// Initialize all components
	database := setupTestDatabase(t)
	defer database.Close()

	staffRepo := db.NewStaffRepository(database)
	auditRepo := db.NewAuditLogRepository(database)
	passwordHasher := crypto.NewBcryptPasswordHasher()
	sessionManager := session.NewMemorySessionManager(8 * time.Hour)

	// Initialize rate limiting components
	attemptRepo := db.NewLoginAttemptRepository(database)
	lockoutRepo := db.NewAccountLockoutRepository(database)
	configRepo := db.NewRateLimitConfigRepository(database)
	
	// Initialize rate limit service
	rateLimitSvc := usecase.NewRateLimitService(
		attemptRepo,
		lockoutRepo,
		configRepo,
		auditRepo,
	)

	authUseCase := usecase.NewAuthUseCase(staffRepo, auditRepo, passwordHasher, sessionManager, rateLimitSvc)

	ctx := context.Background()

	// Create multiple test users with unique IDs
	users := []struct {
		id       string
		name     string
		role     domain.StaffRole
		password string
	}{
		{"test-admin-001", "テスト管理者", domain.RoleAdmin, "AdminPassword123!"},
		{"test-staff-001", "テスト職員1", domain.RoleStaff, "StaffPassword123!"},
		{"test-readonly-001", "テスト閲覧者", domain.RoleReadOnly, "ReadOnlyPassword123!"},
	}

	// Create users in database
	for _, user := range users {
		hashedPassword, err := passwordHasher.HashPassword(user.password)
		require.NoError(t, err)

		staff := &domain.Staff{
			ID:           user.id,
			Name:         user.name,
			Role:         user.role,
			PasswordHash: hashedPassword,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err = staffRepo.Create(ctx, staff)
		require.NoError(t, err)
	}

	// Login all users simultaneously
	sessions := make(map[string]string) // username -> sessionID

	for _, user := range users {
		loginReq := usecase.LoginRequest{
			Username: user.name,
			Password: user.password,
			ClientIP: "192.168.1.100",
		}

		loginResp, err := authUseCase.Login(ctx, loginReq)
		require.NoError(t, err)
		assert.Equal(t, user.id, loginResp.User.ID)
		assert.Equal(t, user.role, loginResp.User.Role)

		sessions[user.name] = loginResp.SessionID
	}

	// Validate all sessions are active
	for username, sessionID := range sessions {
		sessionInfo, err := authUseCase.ValidateSession(ctx, sessionID)
		require.NoError(t, err)
		assert.Equal(t, sessionID, sessionInfo.SessionID)

		// Find the expected user
		var expectedUser *struct {
			id       string
			name     string
			role     domain.StaffRole
			password string
		}
		for _, user := range users {
			if user.name == username {
				expectedUser = &user
				break
			}
		}
		require.NotNil(t, expectedUser)

		assert.Equal(t, expectedUser.id, sessionInfo.User.ID)
		assert.Equal(t, expectedUser.role, sessionInfo.User.Role)
	}

	// Logout one user
	logoutReq := usecase.LogoutRequest{
		SessionID: sessions["テスト管理者"],
		ClientIP:  "192.168.1.100",
	}

	err := authUseCase.Logout(ctx, logoutReq)
	require.NoError(t, err)

	// Verify the logged out user's session is invalid
	sessionInfo, err := authUseCase.ValidateSession(ctx, sessions["テスト管理者"])
	assert.Error(t, err)
	assert.Nil(t, sessionInfo)

	// Verify other users' sessions are still valid
	for username, sessionID := range sessions {
		if username == "テスト管理者" {
			continue // Skip the logged out user
		}

		sessionInfo, err := authUseCase.ValidateSession(ctx, sessionID)
		assert.NoError(t, err)
		assert.NotNil(t, sessionInfo)
	}
}

// setupTestDatabase creates an in-memory database for testing
func setupTestDatabase(t *testing.T) *db.Database {
	config := db.Config{
		Path:         ":memory:",
		MigrationDir: "../../migrations",
	}

	database, err := db.NewDatabase(config)
	require.NoError(t, err)

	// Run migrations to create tables
	ctx := context.Background()
	err = database.RunMigrations(ctx)
	require.NoError(t, err)

	return database
}
