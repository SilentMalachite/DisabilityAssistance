package widgets

import (
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"shien-system/internal/usecase"
)

// FormBase provides common functionality for forms with CSRF protection
type FormBase struct {
	widget.BaseWidget
	
	// CSRF protection
	sessionManager usecase.SessionManager
	sessionID      string
	csrfToken      string
	csrfTokenEntry *widget.Entry
	
	// Form state
	isSubmitting bool
	errorLabel   *widget.Label
}

// NewFormBase creates a new FormBase with CSRF protection
func NewFormBase(sessionManager usecase.SessionManager) *FormBase {
	fb := &FormBase{
		sessionManager: sessionManager,
		errorLabel:     widget.NewLabel(""),
	}
	
	// Create hidden CSRF token field
	fb.csrfTokenEntry = widget.NewEntry()
	fb.csrfTokenEntry.Hide()
	
	return fb
}

// SetSession sets the session information for CSRF protection
func (fb *FormBase) SetSession(sessionID, csrfToken string) {
	fb.sessionID = sessionID
	fb.csrfToken = csrfToken
	fb.csrfTokenEntry.SetText(csrfToken)
}

// ValidateCSRF validates the CSRF token before form submission
func (fb *FormBase) ValidateCSRF(ctx context.Context) error {
	if fb.sessionManager == nil || fb.sessionID == "" || fb.csrfToken == "" {
		return nil // No CSRF validation if not configured
	}
	
	// Type assert to CSRFProtectedSessionManager if available
	if csrfManager, ok := fb.sessionManager.(usecase.CSRFProtectedSessionManager); ok {
		return csrfManager.ValidateCSRFToken(ctx, fb.sessionID, fb.csrfToken)
	}
	return nil // No CSRF validation if not supported
}

// GetCSRFProtectedContext returns a context with CSRF token
func (fb *FormBase) GetCSRFProtectedContext(ctx context.Context) context.Context {
	if fb.csrfToken != "" {
		return context.WithValue(ctx, usecase.ContextKeyCSRFToken, fb.csrfToken)
	}
	return ctx
}

// SetSubmitting sets the form submission state
func (fb *FormBase) SetSubmitting(submitting bool) {
	fb.isSubmitting = submitting
}

// IsSubmitting returns whether the form is currently being submitted
func (fb *FormBase) IsSubmitting() bool {
	return fb.isSubmitting
}

// ShowError displays an error message
func (fb *FormBase) ShowError(message string) {
	fb.errorLabel.SetText(message)
	fb.errorLabel.Show()
}

// ClearError clears any error message
func (fb *FormBase) ClearError() {
	fb.errorLabel.SetText("")
	fb.errorLabel.Hide()
}

// GetErrorLabel returns the error label widget
func (fb *FormBase) GetErrorLabel() *widget.Label {
	return fb.errorLabel
}

// GetCSRFTokenEntry returns the hidden CSRF token entry
func (fb *FormBase) GetCSRFTokenEntry() *widget.Entry {
	return fb.csrfTokenEntry
}

// CreateFormContainer creates a standard form container with CSRF protection
func (fb *FormBase) CreateFormContainer(fields ...fyne.CanvasObject) *fyne.Container {
	// Add CSRF token field and error label to form
	allFields := make([]fyne.CanvasObject, 0, len(fields)+2)
	allFields = append(allFields, fb.csrfTokenEntry) // Hidden CSRF field
	allFields = append(allFields, fields...)
	allFields = append(allFields, fb.errorLabel) // Error display
	
	return container.NewVBox(allFields...)
}