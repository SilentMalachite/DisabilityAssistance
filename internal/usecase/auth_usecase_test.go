package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"shien-system/internal/domain"
)

// Mock repositories for testing
type MockStaffRepository struct {
	mock.Mock
}

func (m *MockStaffRepository) Create(ctx context.Context, staff *domain.Staff) error {
	args := m.Called(ctx, staff)
	return args.Error(0)
}

func (m *MockStaffRepository) GetByID(ctx context.Context, id domain.ID) (*domain.Staff, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Staff), args.Error(1)
}

func (m *MockStaffRepository) GetByExactName(ctx context.Context, name string) (*domain.Staff, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Staff), args.Error(1)
}

func (m *MockStaffRepository) GetByName(ctx context.Context, name string) ([]*domain.Staff, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Staff), args.Error(1)
}

func (m *MockStaffRepository) Update(ctx context.Context, staff *domain.Staff) error {
	args := m.Called(ctx, staff)
	return args.Error(0)
}

func (m *MockStaffRepository) Delete(ctx context.Context, id domain.ID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStaffRepository) List(ctx context.Context, limit, offset int) ([]*domain.Staff, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.Staff), args.Error(1)
}

func (m *MockStaffRepository) GetByRole(ctx context.Context, role domain.StaffRole) ([]*domain.Staff, error) {
	args := m.Called(ctx, role)
	return args.Get(0).([]*domain.Staff), args.Error(1)
}

func (m *MockStaffRepository) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAuditLogRepository) GetByID(ctx context.Context, id domain.ID) (*domain.AuditLog, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) List(ctx context.Context, limit, offset int) ([]*domain.AuditLog, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) GetByActorID(ctx context.Context, actorID domain.ID, limit, offset int) ([]*domain.AuditLog, error) {
	args := m.Called(ctx, actorID, limit, offset)
	return args.Get(0).([]*domain.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) GetByAction(ctx context.Context, action string, limit, offset int) ([]*domain.AuditLog, error) {
	args := m.Called(ctx, action, limit, offset)
	return args.Get(0).([]*domain.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) GetByTarget(ctx context.Context, target string, limit, offset int) ([]*domain.AuditLog, error) {
	args := m.Called(ctx, target, limit, offset)
	return args.Get(0).([]*domain.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) GetByTimeRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*domain.AuditLog, error) {
	args := m.Called(ctx, start, end, limit, offset)
	return args.Get(0).([]*domain.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) Search(ctx context.Context, query domain.AuditLogQuery, limit, offset int) ([]*domain.AuditLog, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*domain.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

type MockPasswordHasher struct {
	mock.Mock
}

func (m *MockPasswordHasher) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHasher) CheckPassword(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

type MockSessionManager struct {
	mock.Mock
}

func (m *MockSessionManager) CreateSession(ctx context.Context, userID domain.ID, userRole domain.StaffRole) (*Session, error) {
	args := m.Called(ctx, userID, userRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Session), args.Error(1)
}

func (m *MockSessionManager) ValidateSession(ctx context.Context, sessionID string) (*Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Session), args.Error(1)
}

func (m *MockSessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionManager) RefreshSession(ctx context.Context, sessionID string) (*Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Session), args.Error(1)
}

func (m *MockSessionManager) CleanupExpiredSessions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Test functions
func TestAuthUseCase_Login_Success(t *testing.T) {
	// Arrange
	staffRepo := &MockStaffRepository{}
	auditRepo := &MockAuditLogRepository{}
	hasher := &MockPasswordHasher{}
	sessionMgr := &MockSessionManager{}

	authUseCase := NewAuthUseCase(staffRepo, auditRepo, hasher, sessionMgr, nil)

	ctx := context.Background()
	staffID := uuid.New().String()
	hashedPassword := "$2a$12$hashedpassword"

	staff := &domain.Staff{
		ID:           staffID,
		Name:         "テスト職員",
		Role:         domain.RoleStaff,
		PasswordHash: hashedPassword,
	}

	session := &Session{
		ID:        uuid.New().String(),
		UserID:    staffID,
		UserRole:  domain.RoleStaff,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(8 * time.Hour),
	}

	staffRepo.On("GetByExactName", ctx, "test-user").Return(staff, nil)
	hasher.On("CheckPassword", hashedPassword, "password123").Return(nil)
	sessionMgr.On("CreateSession", ctx, staffID, domain.RoleStaff).Return(session, nil)
	auditRepo.On("Create", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil)

	req := LoginRequest{
		Username: "test-user",
		Password: "password123",
		ClientIP: "192.168.1.1",
	}

	// Act
	result, err := authUseCase.Login(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, session.ID, result.SessionID)
	assert.Equal(t, staff.ID, result.User.ID)
	assert.Equal(t, staff.Name, result.User.Name)
	assert.Equal(t, staff.Role, result.User.Role)

	staffRepo.AssertExpectations(t)
	hasher.AssertExpectations(t)
	sessionMgr.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestAuthUseCase_Login_InvalidCredentials(t *testing.T) {
	// Arrange
	staffRepo := &MockStaffRepository{}
	auditRepo := &MockAuditLogRepository{}
	hasher := &MockPasswordHasher{}
	sessionMgr := &MockSessionManager{}

	authUseCase := NewAuthUseCase(staffRepo, auditRepo, hasher, sessionMgr, nil)

	ctx := context.Background()

	staffRepo.On("GetByExactName", ctx, "invalid-user").Return(nil, domain.ErrNotFound)
	auditRepo.On("Create", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil)

	req := LoginRequest{
		Username: "invalid-user",
		Password: "wrong-password",
		ClientIP: "192.168.1.1",
	}

	// Act
	result, err := authUseCase.Login(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInvalidCredentials, err)

	staffRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestAuthUseCase_Login_WrongPassword(t *testing.T) {
	// Arrange
	staffRepo := &MockStaffRepository{}
	auditRepo := &MockAuditLogRepository{}
	hasher := &MockPasswordHasher{}
	sessionMgr := &MockSessionManager{}

	authUseCase := NewAuthUseCase(staffRepo, auditRepo, hasher, sessionMgr, nil)

	ctx := context.Background()
	staffID := uuid.New().String()
	hashedPassword := "$2a$12$hashedpassword"

	staff := &domain.Staff{
		ID:           staffID,
		Name:         "テスト職員",
		Role:         domain.RoleStaff,
		PasswordHash: hashedPassword,
	}

	staffRepo.On("GetByExactName", ctx, "test-user").Return(staff, nil)
	hasher.On("CheckPassword", hashedPassword, "wrong-password").Return(ErrInvalidPassword)
	auditRepo.On("Create", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil)

	req := LoginRequest{
		Username: "test-user",
		Password: "wrong-password",
		ClientIP: "192.168.1.1",
	}

	// Act
	result, err := authUseCase.Login(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInvalidCredentials, err)

	staffRepo.AssertExpectations(t)
	hasher.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestAuthUseCase_Logout_Success(t *testing.T) {
	// Arrange
	staffRepo := &MockStaffRepository{}
	auditRepo := &MockAuditLogRepository{}
	hasher := &MockPasswordHasher{}
	sessionMgr := &MockSessionManager{}

	authUseCase := NewAuthUseCase(staffRepo, auditRepo, hasher, sessionMgr, nil)

	ctx := context.Background()
	sessionID := uuid.New().String()
	userID := uuid.New().String()

	session := &Session{
		ID:       sessionID,
		UserID:   userID,
		UserRole: domain.RoleStaff,
	}

	sessionMgr.On("ValidateSession", ctx, sessionID).Return(session, nil)
	sessionMgr.On("DeleteSession", ctx, sessionID).Return(nil)
	auditRepo.On("Create", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil)

	req := LogoutRequest{
		SessionID: sessionID,
		ClientIP:  "192.168.1.1",
	}

	// Act
	err := authUseCase.Logout(ctx, req)

	// Assert
	assert.NoError(t, err)

	sessionMgr.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestAuthUseCase_ValidateSession_Success(t *testing.T) {
	// Arrange
	staffRepo := &MockStaffRepository{}
	auditRepo := &MockAuditLogRepository{}
	hasher := &MockPasswordHasher{}
	sessionMgr := &MockSessionManager{}

	authUseCase := NewAuthUseCase(staffRepo, auditRepo, hasher, sessionMgr, nil)

	ctx := context.Background()
	sessionID := uuid.New().String()
	userID := uuid.New().String()

	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		UserRole:  domain.RoleAdmin,
		CreatedAt: time.Now().Add(-1 * time.Hour),
		ExpiresAt: time.Now().Add(7 * time.Hour),
	}

	staff := &domain.Staff{
		ID:   userID,
		Name: "管理者",
		Role: domain.RoleAdmin,
	}

	sessionMgr.On("ValidateSession", ctx, sessionID).Return(session, nil)
	staffRepo.On("GetByID", ctx, userID).Return(staff, nil)

	// Act
	result, err := authUseCase.ValidateSession(ctx, sessionID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, session.ID, result.SessionID)
	assert.Equal(t, staff.ID, result.User.ID)
	assert.Equal(t, staff.Name, result.User.Name)
	assert.Equal(t, staff.Role, result.User.Role)

	sessionMgr.AssertExpectations(t)
	staffRepo.AssertExpectations(t)
}

func TestAuthUseCase_ValidateSession_InvalidSession(t *testing.T) {
	// Arrange
	staffRepo := &MockStaffRepository{}
	auditRepo := &MockAuditLogRepository{}
	hasher := &MockPasswordHasher{}
	sessionMgr := &MockSessionManager{}

	authUseCase := NewAuthUseCase(staffRepo, auditRepo, hasher, sessionMgr, nil)

	ctx := context.Background()
	sessionID := "invalid-session-id"

	sessionMgr.On("ValidateSession", ctx, sessionID).Return(nil, ErrInvalidSession)

	// Act
	result, err := authUseCase.ValidateSession(ctx, sessionID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInvalidSession, err)

	sessionMgr.AssertExpectations(t)
}

func TestAuthUseCase_ChangePassword_Success(t *testing.T) {
	// Arrange
	staffRepo := &MockStaffRepository{}
	auditRepo := &MockAuditLogRepository{}
	hasher := &MockPasswordHasher{}
	sessionMgr := &MockSessionManager{}

	authUseCase := NewAuthUseCase(staffRepo, auditRepo, hasher, sessionMgr, nil)

	ctx := context.Background()
	userID := uuid.New().String()
	oldHashedPassword := "$2a$12$oldhashedpassword"
	newHashedPassword := "$2a$12$newhashedpassword"

	staff := &domain.Staff{
		ID:           userID,
		Name:         "テスト職員",
		Role:         domain.RoleStaff,
		PasswordHash: oldHashedPassword,
	}

	staffRepo.On("GetByID", ctx, userID).Return(staff, nil)
	hasher.On("CheckPassword", oldHashedPassword, "oldpassword").Return(nil)
	hasher.On("HashPassword", "newpassword").Return(newHashedPassword, nil)
	staffRepo.On("Update", ctx, mock.AnythingOfType("*domain.Staff")).Return(nil)
	auditRepo.On("Create", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil)

	req := ChangePasswordRequest{
		UserID:      userID,
		OldPassword: "oldpassword",
		NewPassword: "newpassword",
		ClientIP:    "192.168.1.1",
	}

	// Act
	err := authUseCase.ChangePassword(ctx, req)

	// Assert
	assert.NoError(t, err)

	staffRepo.AssertExpectations(t)
	hasher.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}
