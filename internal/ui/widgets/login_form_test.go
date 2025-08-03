package widgets

import (
	"context"
	"errors"
	"testing"

	"shien-system/internal/domain"
	"shien-system/internal/usecase"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// MockAuthUseCase implements AuthUseCase for testing
type MockAuthUseCase struct {
	loginFunc      func(ctx context.Context, req usecase.LoginRequest) (*usecase.LoginResponse, error)
	logoutFunc     func(ctx context.Context, req usecase.LogoutRequest) error
	validateFunc   func(ctx context.Context, sessionID string) (*usecase.SessionInfo, error)
	changePassFunc func(ctx context.Context, req usecase.ChangePasswordRequest) error
	refreshFunc    func(ctx context.Context, sessionID string) (*usecase.SessionInfo, error)
}

func (m *MockAuthUseCase) Login(ctx context.Context, req usecase.LoginRequest) (*usecase.LoginResponse, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, req)
	}
	return nil, errors.New("mock not configured")
}

func (m *MockAuthUseCase) Logout(ctx context.Context, req usecase.LogoutRequest) error {
	if m.logoutFunc != nil {
		return m.logoutFunc(ctx, req)
	}
	return errors.New("mock not configured")
}

func (m *MockAuthUseCase) ValidateSession(ctx context.Context, sessionID string) (*usecase.SessionInfo, error) {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, sessionID)
	}
	return nil, errors.New("mock not configured")
}

func (m *MockAuthUseCase) ChangePassword(ctx context.Context, req usecase.ChangePasswordRequest) error {
	if m.changePassFunc != nil {
		return m.changePassFunc(ctx, req)
	}
	return errors.New("mock not configured")
}

func (m *MockAuthUseCase) RefreshSession(ctx context.Context, sessionID string) (*usecase.SessionInfo, error) {
	if m.refreshFunc != nil {
		return m.refreshFunc(ctx, sessionID)
	}
	return nil, errors.New("mock not configured")
}

func TestNewLoginForm(t *testing.T) {
	mockAuth := &MockAuthUseCase{}

	loginForm := NewLoginForm(mockAuth)

	// Verify form is created successfully
	if loginForm == nil {
		t.Fatal("NewLoginForm returned nil")
	}

	// Verify use case is set
	if loginForm.authUseCase == nil {
		t.Error("authUseCase not set")
	}

	// Verify UI components are initialized
	if loginForm.usernameEntry == nil {
		t.Error("usernameEntry not initialized")
	}

	if loginForm.passwordEntry == nil {
		t.Error("passwordEntry not initialized")
	}

	if loginForm.loginButton == nil {
		t.Error("loginButton not initialized")
	}

	if loginForm.statusLabel == nil {
		t.Error("statusLabel not initialized")
	}
}

func TestLoginForm_InitialState(t *testing.T) {
	mockAuth := &MockAuthUseCase{}
	loginForm := NewLoginForm(mockAuth)

	// Verify initial state
	if loginForm.usernameEntry.Text != "" {
		t.Error("Username field should be empty initially")
	}

	if loginForm.passwordEntry.Text != "" {
		t.Error("Password field should be empty initially")
	}

	if loginForm.isLoggingIn {
		t.Error("Should not be in logging in state initially")
	}

	if loginForm.statusLabel.Text != "" {
		t.Error("Status label should be empty initially")
	}

	if loginForm.loginButton.Disabled() {
		t.Error("Login button should not be disabled initially")
	}
}

func TestLoginForm_SuccessfulLogin(t *testing.T) {
	var onSuccessCallback bool

	mockAuth := &MockAuthUseCase{
		loginFunc: func(ctx context.Context, req usecase.LoginRequest) (*usecase.LoginResponse, error) {
			// Verify request parameters
			if req.Username != "admin" {
				t.Errorf("Expected username 'admin', got '%s'", req.Username)
			}
			if req.Password != "password123" {
				t.Errorf("Expected password 'password123', got '%s'", req.Password)
			}

			return &usecase.LoginResponse{
				SessionID: "test-session-token",
				User: &domain.Staff{
					ID:   "admin-001",
					Name: "管理者",
					Role: domain.RoleAdmin,
				},
			}, nil
		},
	}

	loginForm := NewLoginForm(mockAuth)
	loginForm.SetOnLoginSuccess(func(sessionID string, staff *domain.Staff, csrfToken string) {
		onSuccessCallback = true
		if sessionID != "test-session-token" {
			t.Errorf("Expected session token 'test-session-token', got '%s'", sessionID)
		}
		if staff.Name != "管理者" {
			t.Errorf("Expected staff name '管理者', got '%s'", staff.Name)
		}
	})

	// Simulate user input
	test.Type(loginForm.usernameEntry, "admin")
	test.Type(loginForm.passwordEntry, "password123")

	// Simulate login button click
	test.Tap(loginForm.loginButton)

	// Verify callback was called
	if !onSuccessCallback {
		t.Error("OnLoginSuccess callback was not called")
	}
}

func TestLoginForm_FailedLogin(t *testing.T) {
	mockAuth := &MockAuthUseCase{
		loginFunc: func(ctx context.Context, req usecase.LoginRequest) (*usecase.LoginResponse, error) {
			return nil, errors.New("invalid credentials")
		},
	}

	loginForm := NewLoginForm(mockAuth)

	// Simulate user input
	test.Type(loginForm.usernameEntry, "admin")
	test.Type(loginForm.passwordEntry, "wrongpassword")

	// Simulate login button click
	test.Tap(loginForm.loginButton)

	// Verify error is displayed
	if loginForm.statusLabel.Text == "" {
		t.Error("Error message should be displayed in status label")
	}

	// Verify form is not in logging in state after error
	if loginForm.isLoggingIn {
		t.Error("Form should not be in logging in state after error")
	}
}

func TestLoginForm_ValidationEmptyFields(t *testing.T) {
	mockAuth := &MockAuthUseCase{}
	loginForm := NewLoginForm(mockAuth)

	// Try to login with empty fields
	test.Tap(loginForm.loginButton)

	// Verify validation error is shown
	if loginForm.statusLabel.Text == "" {
		t.Error("Validation error should be displayed for empty fields")
	}
}

func TestLoginForm_DisableUIOnLogin(t *testing.T) {
	loginCalled := false

	var lf *LoginForm

	mockAuth := &MockAuthUseCase{
		loginFunc: func(ctx context.Context, req usecase.LoginRequest) (*usecase.LoginResponse, error) {
			loginCalled = true
			// During login, UI should be disabled
			if !lf.loginButton.Disabled() {
				t.Error("Login button should be disabled during login")
			}
			if !lf.usernameEntry.Disabled() {
				t.Error("Username field should be disabled during login")
			}
			if !lf.passwordEntry.Disabled() {
				t.Error("Password field should be disabled during login")
			}

			return &usecase.LoginResponse{
				SessionID: "test-token",
				User: &domain.Staff{
					ID:   "admin-001",
					Name: "管理者",
					Role: domain.RoleAdmin,
				},
			}, nil
		},
	}

	lf = NewLoginForm(mockAuth)

	// Set up form
	test.Type(lf.usernameEntry, "admin")
	test.Type(lf.passwordEntry, "password123")

	// Simulate login
	test.Tap(lf.loginButton)

	if !loginCalled {
		t.Error("Login function was not called")
	}
}

func TestLoginForm_CreateObject(t *testing.T) {
	mockAuth := &MockAuthUseCase{}
	loginForm := NewLoginForm(mockAuth)

	// Create UI object
	obj := loginForm.CreateObject()

	if obj == nil {
		t.Fatal("CreateObject returned nil")
	}

	// Verify object is a container
	if _, ok := obj.(*fyne.Container); !ok {
		t.Error("CreateObject should return a Container")
	}
}

func TestLoginForm_ClearForm(t *testing.T) {
	mockAuth := &MockAuthUseCase{}
	loginForm := NewLoginForm(mockAuth)

	// Fill form with data
	test.Type(loginForm.usernameEntry, "testuser")
	test.Type(loginForm.passwordEntry, "testpass")
	loginForm.statusLabel.SetText("Some error message")

	// Clear form
	loginForm.ClearForm()

	// Verify form is cleared
	if loginForm.usernameEntry.Text != "" {
		t.Error("Username field should be cleared")
	}

	if loginForm.passwordEntry.Text != "" {
		t.Error("Password field should be cleared")
	}

	if loginForm.statusLabel.Text != "" {
		t.Error("Status label should be cleared")
	}

	if loginForm.isLoggingIn {
		t.Error("Should not be in logging in state after clear")
	}
}

func TestLoginForm_SetOnLoginSuccess(t *testing.T) {
	mockAuth := &MockAuthUseCase{}
	loginForm := NewLoginForm(mockAuth)

	// Set callback
	loginForm.SetOnLoginSuccess(func(sessionID string, staff *domain.Staff, csrfToken string) {
		// Callback function for testing
	})

	// Verify callback is set (indirect verification through successful login test)
	if loginForm.onLoginSuccess == nil {
		t.Error("onLoginSuccess callback not set")
	}
}

func TestLoginForm_PasswordFieldSecurity(t *testing.T) {
	mockAuth := &MockAuthUseCase{}
	loginForm := NewLoginForm(mockAuth)

	// Verify password field is configured for password input
	if !loginForm.passwordEntry.Password {
		t.Error("Password field should be configured as password input")
	}
}

func TestLoginForm_EnterKeyHandling(t *testing.T) {
	var loginCalled bool

	mockAuth := &MockAuthUseCase{
		loginFunc: func(ctx context.Context, req usecase.LoginRequest) (*usecase.LoginResponse, error) {
			loginCalled = true
			return &usecase.LoginResponse{
				SessionID: "test-token",
				User: &domain.Staff{
					ID:   "admin-001",
					Name: "管理者",
					Role: domain.RoleAdmin,
				},
			}, nil
		},
	}

	loginForm := NewLoginForm(mockAuth)

	// Fill form
	test.Type(loginForm.usernameEntry, "admin")
	test.Type(loginForm.passwordEntry, "password123")

	// Simulate Enter key press on password field
	loginForm.passwordEntry.OnSubmitted("password123")

	if !loginCalled {
		t.Error("Login should be triggered by Enter key on password field")
	}
}
