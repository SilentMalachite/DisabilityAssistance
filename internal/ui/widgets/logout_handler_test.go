package widgets

import (
	"context"
	"errors"
	"testing"

	"shien-system/internal/config"
	"shien-system/internal/domain"
	"shien-system/internal/usecase"

	"fyne.io/fyne/v2/app"
)

// Note: LogoutHandler is defined in logout_handler.go

func TestNewLogoutHandler(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	mockAuth := &MockAuthUseCase{}
	mockRecipient := &MockRecipientUseCase{}
	mockCertificate := &MockCertificateUseCase{}
	mockSetup := &MockSetupUseCase{needsSetup: false}
	mockAuditRepo := &MockAuditLogRepository{}
	mockStaffRepo := &MockStaffRepository{}
	mockStaff := &MockStaffUseCase{}
	mockConfig := &config.Config{}
	appState := NewAppState(mockAuth, mockRecipient, mockCertificate, mockStaff, mockSetup, nil, mockAuditRepo, mockStaffRepo, nil, mockConfig)

	handler := NewLogoutHandler(appState)

	if handler == nil {
		t.Fatal("NewLogoutHandler returned nil")
	}

	if handler.appState != appState {
		t.Error("App state not set correctly")
	}
}

func TestLogoutHandler_PerformLogout_Success(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	logoutCalled := false
	mockAuth := &MockAuthUseCase{
		logoutFunc: func(ctx context.Context, req usecase.LogoutRequest) error {
			logoutCalled = true
			if req.SessionID != "test-session-123" {
				t.Errorf("Expected session ID 'test-session-123', got '%s'", req.SessionID)
			}
			return nil
		},
	}

	mockRecipient := &MockRecipientUseCase{}
	mockCertificate := &MockCertificateUseCase{}
	mockSetup := &MockSetupUseCase{needsSetup: false}
	mockAuditRepo := &MockAuditLogRepository{}
	mockStaffRepo := &MockStaffRepository{}
	mockStaff := &MockStaffUseCase{}
	mockConfig := &config.Config{}
	appState := NewAppState(mockAuth, mockRecipient, mockCertificate, mockStaff, mockSetup, nil, mockAuditRepo, mockStaffRepo, nil, mockConfig)

	// Login first
	staff := &domain.Staff{
		ID:   "admin-001",
		Name: "管理者",
		Role: domain.RoleAdmin,
	}
	appState.Login("test-session-123", staff)

	handler := NewLogoutHandler(appState)

	// Add observer to verify state change
	observer := &MockStateObserver{}
	appState.AddObserver(observer)

	// Perform logout
	err := handler.PerformLogout()

	if err != nil {
		t.Errorf("Logout failed: %v", err)
	}

	if !logoutCalled {
		t.Error("AuthUseCase.Logout was not called")
	}

	if appState.IsAuthenticated() {
		t.Error("Should not be authenticated after logout")
	}

	if appState.GetCurrentUser() != nil {
		t.Error("Current user should be nil after logout")
	}

	if appState.GetSessionID() != "" {
		t.Error("Session ID should be empty after logout")
	}

	if !observer.WasNotified() {
		t.Error("Observer was not notified of state change")
	}
}

func TestLogoutHandler_PerformLogout_NotAuthenticated(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	mockAuth := &MockAuthUseCase{}
	mockRecipient := &MockRecipientUseCase{}
	mockCertificate := &MockCertificateUseCase{}
	mockSetup := &MockSetupUseCase{needsSetup: false}
	mockAuditRepo := &MockAuditLogRepository{}
	mockStaffRepo := &MockStaffRepository{}
	mockStaff := &MockStaffUseCase{}
	mockConfig := &config.Config{}
	appState := NewAppState(mockAuth, mockRecipient, mockCertificate, mockStaff, mockSetup, nil, mockAuditRepo, mockStaffRepo, nil, mockConfig)

	handler := NewLogoutHandler(appState)

	// Try to logout when not authenticated
	err := handler.PerformLogout()

	if err == nil {
		t.Error("Expected error when logging out while not authenticated")
	}
}

func TestLogoutHandler_PerformLogout_SessionError(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	mockAuth := &MockAuthUseCase{
		logoutFunc: func(ctx context.Context, req usecase.LogoutRequest) error {
			return errors.New("session invalidation failed")
		},
	}

	mockRecipient := &MockRecipientUseCase{}
	mockCertificate := &MockCertificateUseCase{}
	mockSetup := &MockSetupUseCase{needsSetup: false}
	mockAuditRepo := &MockAuditLogRepository{}
	mockStaffRepo := &MockStaffRepository{}
	mockStaff := &MockStaffUseCase{}
	mockConfig := &config.Config{}
	appState := NewAppState(mockAuth, mockRecipient, mockCertificate, mockStaff, mockSetup, nil, mockAuditRepo, mockStaffRepo, nil, mockConfig)

	// Login first
	staff := &domain.Staff{
		ID:   "admin-001",
		Name: "管理者",
		Role: domain.RoleAdmin,
	}
	appState.Login("test-session-123", staff)

	handler := NewLogoutHandler(appState)

	// Perform logout
	err := handler.PerformLogout()

	if err == nil {
		t.Error("Expected error from session invalidation")
	}

	// App state should still be cleared even if session invalidation fails
	if appState.IsAuthenticated() {
		t.Error("Should not be authenticated after logout attempt")
	}
}

func TestLogoutHandler_PerformLogout_WithClientIP(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	logoutCalled := false
	mockAuth := &MockAuthUseCase{
		logoutFunc: func(ctx context.Context, req usecase.LogoutRequest) error {
			logoutCalled = true
			if req.ClientIP != "127.0.0.1" {
				t.Errorf("Expected client IP '127.0.0.1', got '%s'", req.ClientIP)
			}
			return nil
		},
	}

	mockRecipient := &MockRecipientUseCase{}
	mockCertificate := &MockCertificateUseCase{}
	mockSetup := &MockSetupUseCase{needsSetup: false}
	mockAuditRepo := &MockAuditLogRepository{}
	mockStaffRepo := &MockStaffRepository{}
	mockStaff := &MockStaffUseCase{}
	mockConfig := &config.Config{}
	appState := NewAppState(mockAuth, mockRecipient, mockCertificate, mockStaff, mockSetup, nil, mockAuditRepo, mockStaffRepo, nil, mockConfig)

	// Login first
	staff := &domain.Staff{
		ID:   "admin-001",
		Name: "管理者",
		Role: domain.RoleAdmin,
	}
	appState.Login("test-session-123", staff)

	handler := NewLogoutHandler(appState)

	// Perform logout with client IP
	err := handler.PerformLogoutWithIP("127.0.0.1")

	if err != nil {
		t.Errorf("Logout failed: %v", err)
	}

	if !logoutCalled {
		t.Error("AuthUseCase.Logout was not called")
	}
}

func TestLogoutHandler_GetConfirmationMessage(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	mockAuth := &MockAuthUseCase{}
	mockRecipient := &MockRecipientUseCase{}
	mockCertificate := &MockCertificateUseCase{}
	mockSetup := &MockSetupUseCase{needsSetup: false}
	mockAuditRepo := &MockAuditLogRepository{}
	mockStaffRepo := &MockStaffRepository{}
	mockStaff := &MockStaffUseCase{}
	mockConfig := &config.Config{}
	appState := NewAppState(mockAuth, mockRecipient, mockCertificate, mockStaff, mockSetup, nil, mockAuditRepo, mockStaffRepo, nil, mockConfig)

	// Login first
	staff := &domain.Staff{
		ID:   "admin-001",
		Name: "管理者",
		Role: domain.RoleAdmin,
	}
	appState.Login("test-session-123", staff)

	handler := NewLogoutHandler(appState)

	message := handler.GetConfirmationMessage()

	expectedMessage := "管理者さん、ログアウトしますか？"
	if message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, message)
	}
}

func TestLogoutHandler_GetConfirmationMessage_NotAuthenticated(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	mockAuth := &MockAuthUseCase{}
	mockRecipient := &MockRecipientUseCase{}
	mockCertificate := &MockCertificateUseCase{}
	mockSetup := &MockSetupUseCase{needsSetup: false}
	mockAuditRepo := &MockAuditLogRepository{}
	mockStaffRepo := &MockStaffRepository{}
	mockStaff := &MockStaffUseCase{}
	mockConfig := &config.Config{}
	appState := NewAppState(mockAuth, mockRecipient, mockCertificate, mockStaff, mockSetup, nil, mockAuditRepo, mockStaffRepo, nil, mockConfig)

	handler := NewLogoutHandler(appState)

	message := handler.GetConfirmationMessage()

	expectedMessage := "ログアウトしますか？"
	if message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, message)
	}
}
