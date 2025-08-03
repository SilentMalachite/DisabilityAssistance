package widgets

import (
	"context"
	"fmt"

	"shien-system/internal/domain"
	"shien-system/internal/usecase"
	"shien-system/internal/validation"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// LoginForm represents the login form widget
type LoginForm struct {
	authUseCase usecase.AuthUseCase

	// UI components
	usernameEntry *widget.Entry
	passwordEntry *widget.Entry
	loginButton   *widget.Button
	statusLabel   *widget.Label

	// CSRF token field
	csrfTokenEntry *widget.Entry

	// State
	isLoggingIn bool
	sessionID   string
	csrfToken   string

	// Event handlers
	onLoginSuccess func(sessionID string, staff *domain.Staff, csrfToken string)
}

// NewLoginForm creates a new LoginForm widget
func NewLoginForm(authUseCase usecase.AuthUseCase) *LoginForm {
	lf := &LoginForm{
		authUseCase: authUseCase,
		isLoggingIn: false,
	}

	lf.createWidgets()
	lf.setupEventHandlers()

	return lf
}

// createWidgets initializes all UI components
func (lf *LoginForm) createWidgets() {
	// Username entry
	lf.usernameEntry = widget.NewEntry()
	lf.usernameEntry.SetPlaceHolder("ユーザー名")

	// Password entry
	lf.passwordEntry = widget.NewPasswordEntry()
	lf.passwordEntry.SetPlaceHolder("パスワード")

	// CSRF token entry (hidden)
	lf.csrfTokenEntry = widget.NewEntry()
	lf.csrfTokenEntry.Hide()

	// Login button
	lf.loginButton = widget.NewButton("ログイン", func() {
		lf.performLogin()
	})
	lf.loginButton.Importance = widget.HighImportance

	// Status label
	lf.statusLabel = widget.NewLabel("")
	lf.statusLabel.Wrapping = fyne.TextWrapWord
}

// setupEventHandlers configures event handlers
func (lf *LoginForm) setupEventHandlers() {
	// Enter key handling on password field
	lf.passwordEntry.OnSubmitted = func(text string) {
		lf.performLogin()
	}

	// Enter key handling on username field
	lf.usernameEntry.OnSubmitted = func(text string) {
		lf.passwordEntry.FocusGained()
	}
}

// performLogin handles the login process
func (lf *LoginForm) performLogin() {
	// Clear previous status
	lf.statusLabel.SetText("")

	// Get and sanitize input
	formValidator := validation.NewFormValidator()
	username := formValidator.SanitizeInput(lf.usernameEntry.Text)
	password := lf.passwordEntry.Text
	csrfToken := lf.csrfTokenEntry.Text

	// Validate input
	loginData := map[string]string{
		"username":   username,
		"password":   password,
		"csrf_token": csrfToken,
	}

	if validationErrors := formValidator.ValidateLoginForm(loginData); len(validationErrors) > 0 {
		lf.statusLabel.SetText(validationErrors.Error())
		return
	}

	// Set loading state
	lf.setLoggingInState(true)

	// Perform login
	ctx := context.Background()
	
	// Add CSRF token to context if available
	if lf.csrfToken != "" {
		ctx = context.WithValue(ctx, usecase.ContextKeyCSRFToken, lf.csrfToken)
	}
	
	req := usecase.LoginRequest{
		Username: username,
		Password: password,
	}

	resp, err := lf.authUseCase.Login(ctx, req)
	lf.setLoggingInState(false)

	if err != nil {
		lf.statusLabel.SetText(fmt.Sprintf("ログインに失敗しました: %v", err))
		return
	}

	// Success callback
	if lf.onLoginSuccess != nil {
		lf.onLoginSuccess(resp.SessionID, resp.User, resp.CSRFToken)
	}
}

// setLoggingInState updates the UI state during login
func (lf *LoginForm) setLoggingInState(loggingIn bool) {
	lf.isLoggingIn = loggingIn

	// Update UI on main thread
	lf.usernameEntry.Disable()
	lf.passwordEntry.Disable()
	lf.loginButton.Disable()

	if !loggingIn {
		lf.usernameEntry.Enable()
		lf.passwordEntry.Enable()
		lf.loginButton.Enable()
	}

	if loggingIn {
		lf.statusLabel.SetText("ログイン中...")
	}
}

// CreateObject creates the main UI object for this widget
func (lf *LoginForm) CreateObject() fyne.CanvasObject {
	// Title with larger text
	title := widget.NewLabel("障害者サービス管理システム")
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}

	// Subtitle
	subtitle := widget.NewLabel("ログイン")
	subtitle.Alignment = fyne.TextAlignCenter

	// Form fields with improved spacing
	form := container.NewVBox(
		widget.NewLabel("ユーザー名:"),
		lf.usernameEntry,
		widget.NewLabel(""), // Spacer
		widget.NewLabel("パスワード:"),
		lf.passwordEntry,
		widget.NewLabel(""), // Spacer
		lf.loginButton,
	)

	// Status area
	statusContainer := container.NewVBox(
		lf.statusLabel,
	)

	// Main content
	content := container.NewVBox(
		title,
		widget.NewLabel(""), // Spacer
		subtitle,
		widget.NewSeparator(),
		widget.NewLabel(""), // Spacer
		form,
		widget.NewLabel(""), // Spacer
		widget.NewSeparator(),
		statusContainer,
	)

	// Create a card-like container with padding
	loginCard := container.NewPadded(content)

	// Center the content with proper spacing
	return container.NewCenter(
		container.NewVBox(
			widget.NewLabel(""), // Top spacer
			widget.NewLabel(""), // Additional top spacer
			loginCard,
			widget.NewLabel(""), // Bottom spacer
			widget.NewLabel(""), // Additional bottom spacer
		),
	)
}

// SetOnLoginSuccess sets the callback for successful login
func (lf *LoginForm) SetOnLoginSuccess(callback func(sessionID string, staff *domain.Staff, csrfToken string)) {
	lf.onLoginSuccess = callback
}

// ClearForm clears all form fields and resets state
func (lf *LoginForm) ClearForm() {
	lf.usernameEntry.SetText("")
	lf.passwordEntry.SetText("")
	lf.statusLabel.SetText("")
	lf.isLoggingIn = false

	// Re-enable form fields
	lf.usernameEntry.Enable()
	lf.passwordEntry.Enable()
	lf.loginButton.Enable()
}
