package widgets

import (
	"testing"

	"shien-system/internal/config"
	"shien-system/internal/domain"

	"fyne.io/fyne/v2/app"
)

// Note: AppState is defined in app_state.go

func TestNewAppState(t *testing.T) {
	mockAuth := &MockAuthUseCase{}
	mockRecipient := &MockRecipientUseCase{}
	mockCertificate := &MockCertificateUseCase{}
	mockSetup := &MockSetupUseCase{needsSetup: false}
	mockAuditRepo := &MockAuditLogRepository{}
	mockStaffRepo := &MockStaffRepository{}

	mockStaff := &MockStaffUseCase{}
	mockConfig := &config.Config{}
	appState := NewAppState(mockAuth, mockRecipient, mockCertificate, mockStaff, mockSetup, nil, mockAuditRepo, mockStaffRepo, nil, mockConfig)

	if appState == nil {
		t.Fatal("NewAppState returned nil")
	}

	if appState.isAuthenticated {
		t.Error("Should not be authenticated initially")
	}

	if appState.currentUser != nil {
		t.Error("Current user should be nil initially")
	}

	if appState.sessionID != "" {
		t.Error("Session ID should be empty initially")
	}

	if appState.currentView != "login" {
		t.Error("Current view should be 'login' initially")
	}
}

func TestAppState_Login(t *testing.T) {
	mockAuth := &MockAuthUseCase{}
	mockRecipient := &MockRecipientUseCase{}
	mockCertificate := &MockCertificateUseCase{}
	mockSetup := &MockSetupUseCase{needsSetup: false}
	mockAuditRepo := &MockAuditLogRepository{}
	mockStaffRepo := &MockStaffRepository{}
	mockStaff := &MockStaffUseCase{}
	mockConfig := &config.Config{}
	appState := NewAppState(mockAuth, mockRecipient, mockCertificate, mockStaff, mockSetup, nil, mockAuditRepo, mockStaffRepo, nil, mockConfig)

	// Test data
	sessionID := "test-session-123"
	staff := &domain.Staff{
		ID:   "admin-001",
		Name: "管理者",
		Role: domain.RoleAdmin,
	}

	// Perform login
	appState.Login(sessionID, staff)

	// Verify state changes
	if !appState.isAuthenticated {
		t.Error("Should be authenticated after login")
	}

	if appState.currentUser != staff {
		t.Error("Current user not set correctly")
	}

	if appState.sessionID != sessionID {
		t.Error("Session ID not set correctly")
	}

	if appState.currentView != "recipients" {
		t.Error("Current view should change to 'recipients' after login")
	}
}

func TestAppState_Logout(t *testing.T) {
	mockAuth := &MockAuthUseCase{}
	mockRecipient := &MockRecipientUseCase{}
	mockCertificate := &MockCertificateUseCase{}
	mockSetup := &MockSetupUseCase{needsSetup: false}
	mockAuditRepo := &MockAuditLogRepository{}
	mockStaffRepo := &MockStaffRepository{}
	mockStaff := &MockStaffUseCase{}
	mockConfig := &config.Config{}
	appState := NewAppState(mockAuth, mockRecipient, mockCertificate, mockStaff, mockSetup, nil, mockAuditRepo, mockStaffRepo, nil, mockConfig)

	// Set up authenticated state
	staff := &domain.Staff{
		ID:   "admin-001",
		Name: "管理者",
		Role: domain.RoleAdmin,
	}
	appState.Login("test-session-123", staff)

	// Perform logout
	appState.Logout()

	// Verify state reset
	if appState.isAuthenticated {
		t.Error("Should not be authenticated after logout")
	}

	if appState.currentUser != nil {
		t.Error("Current user should be nil after logout")
	}

	if appState.sessionID != "" {
		t.Error("Session ID should be empty after logout")
	}

	if appState.currentView != "login" {
		t.Error("Current view should be 'login' after logout")
	}
}

func TestAppState_IsAuthenticated(t *testing.T) {
	mockAuth := &MockAuthUseCase{}
	mockRecipient := &MockRecipientUseCase{}
	mockCertificate := &MockCertificateUseCase{}
	mockSetup := &MockSetupUseCase{needsSetup: false}
	mockAuditRepo := &MockAuditLogRepository{}
	mockStaffRepo := &MockStaffRepository{}
	mockStaff := &MockStaffUseCase{}
	mockConfig := &config.Config{}
	appState := NewAppState(mockAuth, mockRecipient, mockCertificate, mockStaff, mockSetup, nil, mockAuditRepo, mockStaffRepo, nil, mockConfig)

	// Initially not authenticated
	if appState.IsAuthenticated() {
		t.Error("Should not be authenticated initially")
	}

	// After login
	staff := &domain.Staff{ID: "admin-001", Name: "管理者", Role: domain.RoleAdmin}
	appState.Login("test-session", staff)

	if !appState.IsAuthenticated() {
		t.Error("Should be authenticated after login")
	}

	// After logout
	appState.Logout()

	if appState.IsAuthenticated() {
		t.Error("Should not be authenticated after logout")
	}
}

func TestAppState_GetCurrentUser(t *testing.T) {
	mockAuth := &MockAuthUseCase{}
	mockRecipient := &MockRecipientUseCase{}
	mockCertificate := &MockCertificateUseCase{}
	mockSetup := &MockSetupUseCase{needsSetup: false}
	mockAuditRepo := &MockAuditLogRepository{}
	mockStaffRepo := &MockStaffRepository{}
	mockStaff := &MockStaffUseCase{}
	mockConfig := &config.Config{}
	appState := NewAppState(mockAuth, mockRecipient, mockCertificate, mockStaff, mockSetup, nil, mockAuditRepo, mockStaffRepo, nil, mockConfig)

	// Initially no user
	if appState.GetCurrentUser() != nil {
		t.Error("Should have no current user initially")
	}

	// After login
	staff := &domain.Staff{
		ID:   "admin-001",
		Name: "管理者",
		Role: domain.RoleAdmin,
	}
	appState.Login("test-session", staff)

	currentUser := appState.GetCurrentUser()
	if currentUser == nil {
		t.Fatal("Should have current user after login")
	}

	if currentUser.ID != staff.ID {
		t.Error("Current user ID does not match")
	}

	if currentUser.Name != staff.Name {
		t.Error("Current user name does not match")
	}
}

func TestAppState_GetCurrentView(t *testing.T) {
	mockAuth := &MockAuthUseCase{}
	mockRecipient := &MockRecipientUseCase{}
	mockCertificate := &MockCertificateUseCase{}
	mockSetup := &MockSetupUseCase{needsSetup: false}
	mockAuditRepo := &MockAuditLogRepository{}
	mockStaffRepo := &MockStaffRepository{}
	mockStaff := &MockStaffUseCase{}
	mockConfig := &config.Config{}
	appState := NewAppState(mockAuth, mockRecipient, mockCertificate, mockStaff, mockSetup, nil, mockAuditRepo, mockStaffRepo, nil, mockConfig)

	// Initially login view
	if appState.GetCurrentView() != "login" {
		t.Error("Should be on login view initially")
	}

	// After login
	staff := &domain.Staff{ID: "admin-001", Name: "管理者", Role: domain.RoleAdmin}
	appState.Login("test-session", staff)

	if appState.GetCurrentView() != "recipients" {
		t.Error("Should be on recipients view after login")
	}

	// Change view
	appState.SetCurrentView("staff")

	if appState.GetCurrentView() != "staff" {
		t.Error("View should change to staff")
	}
}

func TestAppState_SetCurrentView_AuthenticationRequired(t *testing.T) {
	mockAuth := &MockAuthUseCase{}
	mockRecipient := &MockRecipientUseCase{}
	mockCertificate := &MockCertificateUseCase{}
	mockSetup := &MockSetupUseCase{needsSetup: false}
	mockAuditRepo := &MockAuditLogRepository{}
	mockStaffRepo := &MockStaffRepository{}
	mockStaff := &MockStaffUseCase{}
	mockConfig := &config.Config{}
	appState := NewAppState(mockAuth, mockRecipient, mockCertificate, mockStaff, mockSetup, nil, mockAuditRepo, mockStaffRepo, nil, mockConfig)

	// Try to change view when not authenticated
	appState.SetCurrentView("staff")

	// Should remain on login view
	if appState.GetCurrentView() != "login" {
		t.Error("Should remain on login view when not authenticated")
	}
}

func TestAppState_GetLoginForm(t *testing.T) {
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

	loginForm := appState.GetLoginForm()

	if loginForm == nil {
		t.Fatal("Login form should not be nil")
	}

	// Should return same instance on subsequent calls
	loginForm2 := appState.GetLoginForm()
	if loginForm != loginForm2 {
		t.Error("Should return same login form instance")
	}
}

func TestAppState_GetRecipientList(t *testing.T) {
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

	// Should return nil when not authenticated
	recipientList := appState.GetRecipientList()
	if recipientList != nil {
		t.Error("Should return nil when not authenticated")
	}

	// Login first
	staff := &domain.Staff{ID: "admin-001", Name: "管理者", Role: domain.RoleAdmin}
	appState.Login("test-session", staff)

	// Should return recipient list when authenticated
	recipientList = appState.GetRecipientList()
	if recipientList == nil {
		t.Error("Should return recipient list when authenticated")
	}
}
