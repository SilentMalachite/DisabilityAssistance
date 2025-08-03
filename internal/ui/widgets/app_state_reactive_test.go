package widgets

import (
	"testing"

	"shien-system/internal/config"
	"shien-system/internal/domain"

	"fyne.io/fyne/v2/app"
)

// Note: StateObserver is defined in app_state.go

func TestAppState_AddObserver(t *testing.T) {
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

	observer := &MockStateObserver{}
	appState.AddObserver(observer)

	// Verify observer was added
	if len(appState.GetObservers()) != 1 {
		t.Error("Observer was not added")
	}
}

func TestAppState_RemoveObserver(t *testing.T) {
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

	observer := &MockStateObserver{}
	appState.AddObserver(observer)
	appState.RemoveObserver(observer)

	// Verify observer was removed
	if len(appState.GetObservers()) != 0 {
		t.Error("Observer was not removed")
	}
}

func TestAppState_NotifyObserversOnLogin(t *testing.T) {
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

	observer := &MockStateObserver{}
	appState.AddObserver(observer)

	// Perform login
	staff := &domain.Staff{
		ID:   "admin-001",
		Name: "管理者",
		Role: domain.RoleAdmin,
	}
	appState.Login("test-session", staff)

	// Verify observer was notified
	if !observer.WasNotified() {
		t.Error("Observer was not notified on login")
	}
}

func TestAppState_NotifyObserversOnLogout(t *testing.T) {
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

	observer := &MockStateObserver{}

	// Login first
	staff := &domain.Staff{
		ID:   "admin-001",
		Name: "管理者",
		Role: domain.RoleAdmin,
	}
	appState.Login("test-session", staff)

	// Add observer after login to reset notification state
	appState.AddObserver(observer)
	observer.Reset()

	// Perform logout
	appState.Logout()

	// Verify observer was notified
	if !observer.WasNotified() {
		t.Error("Observer was not notified on logout")
	}
}

func TestAppState_NotifyObserversOnViewChange(t *testing.T) {
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

	// Login first to enable view changes
	staff := &domain.Staff{
		ID:   "admin-001",
		Name: "管理者",
		Role: domain.RoleAdmin,
	}
	appState.Login("test-session", staff)

	observer := &MockStateObserver{}
	appState.AddObserver(observer)
	observer.Reset()

	// Change view
	appState.SetCurrentView("staff")

	// Verify observer was notified
	if !observer.WasNotified() {
		t.Error("Observer was not notified on view change")
	}
}

func TestAppState_MultipleObservers(t *testing.T) {
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

	observer1 := &MockStateObserver{}
	observer2 := &MockStateObserver{}

	appState.AddObserver(observer1)
	appState.AddObserver(observer2)

	// Perform login
	staff := &domain.Staff{
		ID:   "admin-001",
		Name: "管理者",
		Role: domain.RoleAdmin,
	}
	appState.Login("test-session", staff)

	// Both observers should be notified
	if !observer1.WasNotified() {
		t.Error("Observer1 was not notified")
	}

	if !observer2.WasNotified() {
		t.Error("Observer2 was not notified")
	}
}

func TestAppState_NotifyOnlyValidObservers(t *testing.T) {
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

	observer1 := &MockStateObserver{}
	observer2 := &MockStateObserver{}

	appState.AddObserver(observer1)
	appState.AddObserver(observer2)

	// Remove one observer
	appState.RemoveObserver(observer1)

	// Perform login
	staff := &domain.Staff{
		ID:   "admin-001",
		Name: "管理者",
		Role: domain.RoleAdmin,
	}
	appState.Login("test-session", staff)

	// Only remaining observer should be notified
	if observer1.WasNotified() {
		t.Error("Removed observer should not be notified")
	}

	if !observer2.WasNotified() {
		t.Error("Active observer was not notified")
	}
}

// MockStateObserver implements StateObserver for testing
type MockStateObserver struct {
	notified bool
}

func (m *MockStateObserver) OnStateChanged() {
	m.notified = true
}

func (m *MockStateObserver) WasNotified() bool {
	return m.notified
}

func (m *MockStateObserver) Reset() {
	m.notified = false
}
