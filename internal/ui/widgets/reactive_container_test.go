package widgets

import (
	"testing"

	"shien-system/internal/config"
	"shien-system/internal/domain"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

// Note: ReactiveContainer is defined in reactive_container.go

func TestNewReactiveContainer(t *testing.T) {
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

	contentFunc := func(state *AppState) fyne.CanvasObject {
		return widget.NewLabel("Test Content")
	}

	reactive := NewReactiveContainer(appState, contentFunc)

	if reactive == nil {
		t.Fatal("NewReactiveContainer returned nil")
	}

	if reactive.appState != appState {
		t.Error("App state not set correctly")
	}

	if reactive.contentFunc == nil {
		t.Error("Content function not set")
	}

	if reactive.container == nil {
		t.Error("Container not initialized")
	}
}

func TestReactiveContainer_InitialContent(t *testing.T) {
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

	contentFunc := func(state *AppState) fyne.CanvasObject {
		if state.IsAuthenticated() {
			return widget.NewLabel("Authenticated Content")
		}
		return widget.NewLabel("Login Content")
	}

	reactive := NewReactiveContainer(appState, contentFunc)

	// Initial state should show login content
	content := reactive.GetContent()
	if label, ok := content.(*widget.Label); ok {
		if label.Text != "Login Content" {
			t.Errorf("Expected 'Login Content', got '%s'", label.Text)
		}
	} else {
		t.Error("Content is not a label")
	}
}

func TestReactiveContainer_UpdateOnStateChange(t *testing.T) {
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

	contentFunc := func(state *AppState) fyne.CanvasObject {
		if state.IsAuthenticated() {
			user := state.GetCurrentUser()
			if user != nil {
				return widget.NewLabel("Welcome " + user.Name)
			}
			return widget.NewLabel("Authenticated")
		}
		return widget.NewLabel("Please Login")
	}

	reactive := NewReactiveContainer(appState, contentFunc)

	// Initially not authenticated
	initialContent := reactive.GetContent()
	if label, ok := initialContent.(*widget.Label); ok {
		if label.Text != "Please Login" {
			t.Errorf("Expected 'Please Login', got '%s'", label.Text)
		}
	} else {
		t.Error("Initial content is not a label")
	}

	// Simulate login
	staff := &domain.Staff{
		ID:   "admin-001",
		Name: "管理者",
		Role: domain.RoleAdmin,
	}
	appState.Login("test-session", staff)

	// Update content
	reactive.Update()

	// Content should now show welcome message
	updatedContent := reactive.GetContent()
	if label, ok := updatedContent.(*widget.Label); ok {
		if label.Text != "Welcome 管理者" {
			t.Errorf("Expected 'Welcome 管理者', got '%s'", label.Text)
		}
	} else {
		t.Error("Updated content is not a label")
	}
}

func TestReactiveContainer_UpdateOnLogout(t *testing.T) {
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

	contentFunc := func(state *AppState) fyne.CanvasObject {
		if state.IsAuthenticated() {
			return widget.NewLabel("Main Application")
		}
		return widget.NewLabel("Login Required")
	}

	reactive := NewReactiveContainer(appState, contentFunc)

	// Login first
	staff := &domain.Staff{
		ID:   "admin-001",
		Name: "管理者",
		Role: domain.RoleAdmin,
	}
	appState.Login("test-session", staff)
	reactive.Update()

	// Verify authenticated content
	authContent := reactive.GetContent()
	if label, ok := authContent.(*widget.Label); ok {
		if label.Text != "Main Application" {
			t.Errorf("Expected 'Main Application', got '%s'", label.Text)
		}
	} else {
		t.Error("Authenticated content is not a label")
	}

	// Logout
	appState.Logout()
	reactive.Update()

	// Content should revert to login
	logoutContent := reactive.GetContent()
	if label, ok := logoutContent.(*widget.Label); ok {
		if label.Text != "Login Required" {
			t.Errorf("Expected 'Login Required', got '%s'", label.Text)
		}
	} else {
		t.Error("Logout content is not a label")
	}
}

func TestReactiveContainer_ComplexContentUpdates(t *testing.T) {
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

	contentFunc := func(state *AppState) fyne.CanvasObject {
		if !state.IsAuthenticated() {
			return state.GetLoginForm().CreateObject()
		}

		switch state.GetCurrentView() {
		case "recipients":
			return widget.NewLabel("Recipients View")
		case "staff":
			return widget.NewLabel("Staff View")
		default:
			return widget.NewLabel("Default View")
		}
	}

	reactive := NewReactiveContainer(appState, contentFunc)

	// Initially shows login form
	initialContent := reactive.GetContent()
	if initialContent == nil {
		t.Error("Initial content should not be nil")
	}

	// Login
	staff := &domain.Staff{
		ID:   "admin-001",
		Name: "管理者",
		Role: domain.RoleAdmin,
	}
	appState.Login("test-session", staff)
	reactive.Update()

	// Should show recipients view (default after login)
	recipientContent := reactive.GetContent()
	if label, ok := recipientContent.(*widget.Label); ok {
		if label.Text != "Recipients View" {
			t.Errorf("Expected 'Recipients View', got '%s'", label.Text)
		}
	} else {
		t.Error("Recipients content is not a label")
	}

	// Change view to staff
	appState.SetCurrentView("staff")
	reactive.Update()

	// Should show staff view
	staffContent := reactive.GetContent()
	if label, ok := staffContent.(*widget.Label); ok {
		if label.Text != "Staff View" {
			t.Errorf("Expected 'Staff View', got '%s'", label.Text)
		}
	} else {
		t.Error("Staff content is not a label")
	}
}

func TestReactiveContainer_GetContainer(t *testing.T) {
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

	contentFunc := func(state *AppState) fyne.CanvasObject {
		return widget.NewLabel("Test")
	}

	reactive := NewReactiveContainer(appState, contentFunc)

	container := reactive.GetContainer()
	if container == nil {
		t.Error("GetContainer returned nil")
	}

	// Should return same container on subsequent calls
	container2 := reactive.GetContainer()
	if container != container2 {
		t.Error("Should return same container instance")
	}
}

func TestReactiveContainer_SetUpdateCallback(t *testing.T) {
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

	contentFunc := func(state *AppState) fyne.CanvasObject {
		return widget.NewLabel("Test")
	}

	reactive := NewReactiveContainer(appState, contentFunc)

	callbackCalled := false
	reactive.SetUpdateCallback(func() {
		callbackCalled = true
	})

	// Trigger update
	reactive.Update()

	if !callbackCalled {
		t.Error("Update callback was not called")
	}
}

func TestReactiveContainer_AutoRefresh(t *testing.T) {
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

	updateCount := 0
	contentFunc := func(state *AppState) fyne.CanvasObject {
		updateCount++
		return widget.NewLabel("Updated")
	}

	reactive := NewReactiveContainer(appState, contentFunc)

	initialCount := updateCount

	// First update
	reactive.Update()
	if updateCount != initialCount+1 {
		t.Error("Content function should be called on update")
	}

	// Second update
	reactive.Update()
	if updateCount != initialCount+2 {
		t.Error("Content function should be called on each update")
	}
}
