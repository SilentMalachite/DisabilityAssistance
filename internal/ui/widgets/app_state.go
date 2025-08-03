package widgets

import (
	"context"
	"fmt"

	"shien-system/internal/adapter/pdf"
	"shien-system/internal/config"
	"shien-system/internal/domain"
	"shien-system/internal/usecase"

	"fyne.io/fyne/v2"
)

// StateObserver defines an interface for observing app state changes
type StateObserver interface {
	OnStateChanged()
}

// AppState represents the application state with authentication
type AppState struct {
	// Authentication
	isAuthenticated bool
	currentUser     *domain.Staff
	sessionID       string
	csrfToken       string

	// Setup state
	needsSetup bool

	// Window reference for dialogs
	window fyne.Window

	// Configuration
	config *config.Config

	// Use cases
	authUseCase        usecase.AuthUseCase
	recipientUseCase   usecase.RecipientUseCase
	certificateUseCase usecase.CertificateUseCase
	staffUseCase       usecase.StaffUseCase
	setupUseCase       usecase.SetupUseCase
	backupUseCase      *usecase.BackupUseCase

	// Services
	pdfService *pdf.PDFService

	// Repositories for direct access
	auditRepo domain.AuditLogRepository
	staffRepo domain.StaffRepository

	// UI state
	currentView string

	// UI components (lazy loading)
	setupForm           *SetupForm
	loginForm           *LoginForm
	recipientList       *RecipientList
	recipientForm       *RecipientForm
	certificateForm     *CertificateForm
	certificateList     *CertificateList
	auditLogList        *AuditLogList
	staffList           *StaffList
	staffForm           *StaffForm
	settingsView        *SettingsView
	accessibilityManager *AccessibilityManager

	// Error handling (set from main window)
	feedbackManager *FeedbackManager
	errorDialog     *ErrorDialog

	// Observers for reactive updates
	observers []StateObserver
}

// NewAppState creates a new application state
func NewAppState(authUseCase usecase.AuthUseCase, recipientUseCase usecase.RecipientUseCase, certificateUseCase usecase.CertificateUseCase, staffUseCase usecase.StaffUseCase, setupUseCase usecase.SetupUseCase, backupUseCase *usecase.BackupUseCase, auditRepo domain.AuditLogRepository, staffRepo domain.StaffRepository, pdfService *pdf.PDFService, cfg *config.Config) *AppState {
	appState := &AppState{
		isAuthenticated:    false,
		currentUser:        nil,
		sessionID:          "",
		config:             cfg,
		authUseCase:        authUseCase,
		recipientUseCase:   recipientUseCase,
		certificateUseCase: certificateUseCase,
		staffUseCase:       staffUseCase,
		setupUseCase:       setupUseCase,
		backupUseCase:      backupUseCase,
		pdfService:         pdfService,
		auditRepo:          auditRepo,
		staffRepo:          staffRepo,
		currentView:        "login",
		observers:          make([]StateObserver, 0),
	}

	// Initialize accessibility manager
	appState.accessibilityManager = NewAccessibilityManager()

	// Check if initial setup is needed
	ctx := context.Background()
	needsSetup, err := setupUseCase.NeedsInitialSetup(ctx)
	if err != nil {
		// Log error but continue
		fmt.Printf("Warning: failed to check setup status: %v\n", err)
	} else {
		appState.needsSetup = needsSetup
		if needsSetup {
			appState.currentView = "setup"
		}
	}

	return appState
}

// Login sets the authenticated state
func (as *AppState) Login(sessionID string, user *domain.Staff) {
	as.isAuthenticated = true
	as.currentUser = user
	as.sessionID = sessionID
	as.currentView = "recipients" // Default view after login
	
	// Generate CSRF token for session protection
	// Note: This should be retrieved from the session in a real implementation
	as.csrfToken = "" // Will be set by LoginWithCSRF
	
	as.notifyObservers()
}

// LoginWithCSRF sets the authenticated state with CSRF token
func (as *AppState) LoginWithCSRF(sessionID string, user *domain.Staff, csrfToken string) {
	as.isAuthenticated = true
	as.currentUser = user
	as.sessionID = sessionID
	as.csrfToken = csrfToken
	as.currentView = "recipients" // Default view after login
	as.notifyObservers()
}

// Logout clears the authenticated state
func (as *AppState) Logout() {
	as.isAuthenticated = false
	as.currentUser = nil
	as.sessionID = ""
	as.csrfToken = ""
	as.currentView = "login"

	// Clear UI components to reset state
	as.loginForm = nil
	as.recipientList = nil
	as.recipientForm = nil
	as.certificateForm = nil
	as.certificateList = nil
	as.auditLogList = nil
	as.staffList = nil
	as.staffForm = nil
	as.settingsView = nil

	as.notifyObservers()
}

// IsAuthenticated returns whether user is authenticated
func (as *AppState) IsAuthenticated() bool {
	return as.isAuthenticated
}

// GetCurrentUser returns the current authenticated user
func (as *AppState) GetCurrentUser() *domain.Staff {
	return as.currentUser
}

// GetSessionID returns the current session ID
func (as *AppState) GetSessionID() string {
	return as.sessionID
}

// GetCSRFToken returns the current CSRF token
func (as *AppState) GetCSRFToken() string {
	return as.csrfToken
}

// GetCurrentView returns the current view name
func (as *AppState) GetCurrentView() string {
	return as.currentView
}

// SetCurrentView changes the current view (only if authenticated)
func (as *AppState) SetCurrentView(view string) {
	if !as.isAuthenticated && view != "login" {
		// Non-authenticated users can only access login view
		return
	}
	oldView := as.currentView
	as.currentView = view

	// Notify observers if view actually changed
	if oldView != as.currentView {
		as.notifyObservers()
	}
}

// GetSetupForm returns the setup form (lazy loading)
func (as *AppState) GetSetupForm() *SetupForm {
	if as.setupForm == nil {
		as.setupForm = NewSetupForm(as.setupUseCase, func() {
			// Setup complete - switch to login view
			as.needsSetup = false
			as.currentView = "login"
			as.notifyObservers()
		}, as.feedbackManager)
	}
	return as.setupForm
}

// GetLoginForm returns the login form (lazy loading)
func (as *AppState) GetLoginForm() *LoginForm {
	if as.loginForm == nil {
		as.loginForm = NewLoginForm(as.authUseCase)

		// Set up login success callback
		as.loginForm.SetOnLoginSuccess(func(sessionID string, staff *domain.Staff, csrfToken string) {
			as.LoginWithCSRF(sessionID, staff, csrfToken)
		})
	}
	return as.loginForm
}

// GetRecipientList returns the recipient list (lazy loading, auth required)
func (as *AppState) GetRecipientList() *RecipientList {
	if !as.isAuthenticated {
		return nil
	}

	if as.recipientList == nil && as.recipientUseCase != nil {
		as.recipientList = NewRecipientList(
			as.recipientUseCase,
			as.certificateUseCase,
			as.staffUseCase,
			as.pdfService,
		)

		// Set up event handlers
		as.recipientList.SetOnNewRecipient(func() {
			as.showRecipientForm(nil) // nil means create new
		})

		as.recipientList.SetOnEditRecipient(func(recipientID string) {
			as.showRecipientForm(&recipientID) // edit existing
		})

		// Load initial data
		go as.recipientList.LoadData()
	}

	return as.recipientList
}

// GetStaffList returns the staff list (lazy loading, auth required)
func (as *AppState) GetStaffList() *StaffList {
	if !as.isAuthenticated {
		return nil
	}

	if as.staffList == nil && as.staffUseCase != nil {
		as.staffList = NewStaffList(as.staffUseCase, as.pdfService)

		// Set up event handlers
		as.staffList.SetOnNewStaff(func() {
			as.showStaffForm(nil) // nil means create new
		})

		as.staffList.SetOnEditStaff(func(staffID string) {
			as.showStaffForm(&staffID) // edit existing
		})

		// Load initial data
		go as.staffList.LoadData()
	}

	return as.staffList
}

// GetStaffForm returns the staff form (lazy loading, auth required)
func (as *AppState) GetStaffForm() *StaffForm {
	if !as.isAuthenticated {
		return nil
	}

	if as.staffForm == nil && as.staffUseCase != nil {
		as.staffForm = NewStaffForm(as.staffUseCase)

		// Set up event handlers
		as.staffForm.SetOnSaved(func(staff *domain.Staff) {
			// Show success message and refresh list
			if as.feedbackManager != nil {
				as.feedbackManager.ShowSuccess(fmt.Sprintf("職員「%s」を保存しました", staff.Name))
			}

			// Refresh staff list if visible
			if as.staffList != nil {
				as.staffList.LoadData()
			}
		})

		as.staffForm.SetOnCancelled(func() {
			// Handle form cancellation - nothing needed for now
		})
	}

	return as.staffForm
}

// showStaffForm displays the staff form dialog
func (as *AppState) showStaffForm(staffID *string) {
	if !as.isAuthenticated {
		return
	}

	form := as.GetStaffForm()
	if form == nil {
		if as.errorDialog != nil {
			as.errorDialog.ShowError("エラー", fmt.Errorf("フォームを初期化できませんでした"))
		}
		return
	}

	// Configure form for create or edit
	if staffID == nil {
		// Create new staff
		form.SetForCreate(as.currentUser)
	} else {
		// Edit existing staff - load data first
		staff, err := as.loadStaffForEdit(*staffID)
		if err != nil {
			if as.errorDialog != nil {
				as.errorDialog.ShowError("読み込みエラー", fmt.Errorf("職員情報を読み込めませんでした: %v", err))
			}
			return
		}
		form.SetForEdit(staff, as.currentUser)
	}

	// Show form in dialog
	if as.window == nil {
		if as.errorDialog != nil {
			as.errorDialog.ShowError("エラー", fmt.Errorf("ウィンドウが初期化されていません"))
		}
		return
	}

	// Create and show dialog
	dlg := form.CreateDialog(as.window)
	
	// Set up close handler
	form.SetOnSaved(func(staff *domain.Staff) {
		dlg.Hide()
		// Show success message and refresh list
		if as.feedbackManager != nil {
			as.feedbackManager.ShowSuccess(fmt.Sprintf("職員「%s」を保存しました", staff.Name))
		}
		// Refresh staff list if visible
		if as.staffList != nil {
			as.staffList.LoadData()
		}
	})
	
	form.SetOnCancelled(func() {
		dlg.Hide()
	})

	dlg.Show()
}

// loadStaffForEdit loads a staff for editing
func (as *AppState) loadStaffForEdit(staffID string) (*domain.Staff, error) {
	ctx := context.Background()
	return as.staffUseCase.GetStaff(ctx, staffID)
}

// CreateContent creates the main UI content based on current state
func (as *AppState) CreateContent() fyne.CanvasObject {
	// Check if initial setup is needed
	if as.needsSetup && as.currentView == "setup" {
		return as.GetSetupForm().CreateContent()
	}

	if !as.isAuthenticated {
		return as.GetLoginForm().CreateObject()
	}

	switch as.currentView {
	case "recipients":
		recipientList := as.GetRecipientList()
		if recipientList != nil {
			return recipientList.CreateObject()
		}
		fallthrough
	case "certificates":
		certificateList := as.GetCertificateList()
		if certificateList != nil {
			return certificateList.CreateObject()
		}
		fallthrough
	case "audit":
		auditLogList := as.GetAuditLogList()
		if auditLogList != nil {
			return auditLogList.CreateObject()
		}
		fallthrough
	case "staff":
		staffList := as.GetStaffList()
		if staffList != nil {
			return staffList.CreateObject()
		}
		fallthrough
	case "settings":
		settingsView := as.GetSettingsView()
		if settingsView != nil {
			return settingsView.CreateObject()
		}
		fallthrough
	default:
		// Default to recipients view for authenticated users
		as.currentView = "recipients"
		recipientList := as.GetRecipientList()
		if recipientList != nil {
			return recipientList.CreateObject()
		}
		// Fallback to login if no recipient use case
		return as.GetLoginForm().CreateObject()
	}
}

// AddObserver adds a state observer
func (as *AppState) AddObserver(observer StateObserver) {
	as.observers = append(as.observers, observer)
}

// RemoveObserver removes a state observer
func (as *AppState) RemoveObserver(observer StateObserver) {
	for i, obs := range as.observers {
		if obs == observer {
			as.observers = append(as.observers[:i], as.observers[i+1:]...)
			break
		}
	}
}

// GetObservers returns the list of observers (for testing)
func (as *AppState) GetObservers() []StateObserver {
	return as.observers
}

// notifyObservers notifies all observers of state changes
func (as *AppState) notifyObservers() {
	for _, observer := range as.observers {
		observer.OnStateChanged()
	}
}

// SetErrorHandling sets the error handling components
func (as *AppState) SetErrorHandling(feedbackManager *FeedbackManager, errorDialog *ErrorDialog) {
	as.feedbackManager = feedbackManager
	as.errorDialog = errorDialog
}

// SetWindow sets the main window reference for dialogs
func (as *AppState) SetWindow(window fyne.Window) {
	as.window = window
}

// GetFeedbackManager returns the feedback manager
func (as *AppState) GetFeedbackManager() *FeedbackManager {
	return as.feedbackManager
}

// GetErrorDialog returns the error dialog
func (as *AppState) GetErrorDialog() *ErrorDialog {
	return as.errorDialog
}

// GetRecipientForm returns the recipient form (lazy loading, auth required)
func (as *AppState) GetRecipientForm() *RecipientForm {
	if !as.isAuthenticated {
		return nil
	}

	if as.recipientForm == nil && as.recipientUseCase != nil {
		as.recipientForm = NewRecipientForm(as.recipientUseCase)

		// Set up event handlers
		as.recipientForm.SetOnSaved(func(recipient *domain.Recipient) {
			// Show success message and refresh list
			if as.feedbackManager != nil {
				as.feedbackManager.ShowSuccess(fmt.Sprintf("利用者「%s」を保存しました", recipient.Name))
			}

			// Refresh recipient list if visible
			if as.recipientList != nil {
				as.recipientList.LoadData()
			}
		})

		as.recipientForm.SetOnCancelled(func() {
			// Handle form cancellation - nothing needed for now
		})
	}

	return as.recipientForm
}

// showRecipientForm displays the recipient form dialog
func (as *AppState) showRecipientForm(recipientID *string) {
	if !as.isAuthenticated {
		return
	}

	form := as.GetRecipientForm()
	if form == nil {
		if as.errorDialog != nil {
			as.errorDialog.ShowError("エラー", fmt.Errorf("フォームを初期化できませんでした"))
		}
		return
	}

	// Configure form for create or edit
	if recipientID == nil {
		// Create new recipient
		form.SetForCreate(as.currentUser)
	} else {
		// Edit existing recipient - load data first
		recipient, err := as.loadRecipientForEdit(*recipientID)
		if err != nil {
			if as.errorDialog != nil {
				as.errorDialog.ShowError("読み込みエラー", fmt.Errorf("利用者情報を読み込めませんでした: %v", err))
			}
			return
		}
		form.SetForEdit(recipient, as.currentUser)
	}

	// Show form in dialog
	if as.window == nil {
		if as.errorDialog != nil {
			as.errorDialog.ShowError("エラー", fmt.Errorf("ウィンドウが初期化されていません"))
		}
		return
	}

	// Create and show dialog
	dlg := form.CreateDialog(as.window)
	
	// Set up close handler
	form.SetOnSaved(func(recipient *domain.Recipient) {
		dlg.Hide()
		// Show success message and refresh list
		if as.feedbackManager != nil {
			as.feedbackManager.ShowSuccess(fmt.Sprintf("利用者「%s」を保存しました", recipient.Name))
		}
		// Refresh recipient list if visible
		if as.recipientList != nil {
			as.recipientList.LoadData()
		}
	})
	
	form.SetOnCancelled(func() {
		dlg.Hide()
	})

	dlg.Show()
}

// loadRecipientForEdit loads a recipient for editing
func (as *AppState) loadRecipientForEdit(recipientID string) (*domain.Recipient, error) {
	ctx := context.Background()
	return as.recipientUseCase.GetRecipient(ctx, recipientID)
}

// GetCertificateList returns the certificate list (lazy loading, auth required)
func (as *AppState) GetCertificateList() *CertificateList {
	if !as.isAuthenticated {
		return nil
	}

	if as.certificateList == nil && as.certificateUseCase != nil && as.recipientUseCase != nil {
		as.certificateList = NewCertificateList(as.certificateUseCase, as.recipientUseCase, as.pdfService)

		// Set up event handlers
		as.certificateList.SetOnNewCertificate(func() {
			as.showCertificateForm(nil) // nil means create new
		})

		as.certificateList.SetOnEditCertificate(func(certificateID string) {
			as.showCertificateForm(&certificateID) // edit existing
		})

		// Load initial data
		go as.certificateList.LoadData()
	}

	return as.certificateList
}

// GetCertificateForm returns the certificate form (lazy loading, auth required)
func (as *AppState) GetCertificateForm() *CertificateForm {
	if !as.isAuthenticated {
		return nil
	}

	if as.certificateForm == nil && as.certificateUseCase != nil && as.recipientUseCase != nil {
		as.certificateForm = NewCertificateForm(as.certificateUseCase, as.recipientUseCase, as.currentUser)

		// Set up event handlers
		as.certificateForm.SetOnSaved(func(certificate *domain.BenefitCertificate) {
			// Show success message and refresh list
			if as.feedbackManager != nil {
				as.feedbackManager.ShowSuccess("受給者証を保存しました")
			}

			// Refresh certificate list if visible
			if as.certificateList != nil {
				as.certificateList.LoadData()
			}
		})

		as.certificateForm.SetOnCancelled(func() {
			// Handle form cancellation - nothing needed for now
		})
	}

	return as.certificateForm
}

// showCertificateForm displays the certificate form dialog
func (as *AppState) showCertificateForm(certificateID *string) {
	if !as.isAuthenticated {
		return
	}

	form := as.GetCertificateForm()
	if form == nil {
		if as.errorDialog != nil {
			as.errorDialog.ShowError("エラー", fmt.Errorf("フォームを初期化できませんでした"))
		}
		return
	}

	// Configure form for create or edit
	if certificateID == nil {
		// Create new certificate
		form.SetForCreate(nil)
	} else {
		// Edit existing certificate - load data first
		certificate, err := as.loadCertificateForEdit(*certificateID)
		if err != nil {
			if as.errorDialog != nil {
				as.errorDialog.ShowError("読み込みエラー", fmt.Errorf("受給者証情報を読み込めませんでした: %v", err))
			}
			return
		}
		form.SetForEdit(certificate)
	}

	// Show form in dialog
	if as.window == nil {
		if as.errorDialog != nil {
			as.errorDialog.ShowError("エラー", fmt.Errorf("ウィンドウが初期化されていません"))
		}
		return
	}

	// Create and show dialog
	dlg := form.CreateDialog()
	
	// Set up close handler
	form.SetOnSaved(func(certificate *domain.BenefitCertificate) {
		dlg.Hide()
		// Show success message and refresh list
		if as.feedbackManager != nil {
			as.feedbackManager.ShowSuccess("受給者証を保存しました")
		}
		// Refresh certificate list if visible
		if as.certificateList != nil {
			as.certificateList.LoadData()
		}
	})
	
	form.SetOnCancelled(func() {
		dlg.Hide()
	})

	dlg.Show()
}

// loadCertificateForEdit loads a certificate for editing
func (as *AppState) loadCertificateForEdit(certificateID string) (*domain.BenefitCertificate, error) {
	ctx := context.Background()
	return as.certificateUseCase.GetCertificate(ctx, certificateID)
}

// GetSettingsView returns the settings view (lazy loading, auth required)
func (as *AppState) GetSettingsView() *SettingsView {
	if !as.isAuthenticated {
		return nil
	}

	if as.settingsView == nil && as.config != nil {
		as.settingsView = NewSettingsView(as.config)

		// Set up event handlers
		as.settingsView.SetOnSaved(func() {
			// Show success message
			if as.feedbackManager != nil {
				as.feedbackManager.ShowSuccess("設定を保存しました")
			}
		})

		as.settingsView.SetOnCancelled(func() {
			// Handle settings cancellation - nothing needed for now
		})
	}

	return as.settingsView
}

// GetAccessibilityManager returns the accessibility manager
func (as *AppState) GetAccessibilityManager() *AccessibilityManager {
	return as.accessibilityManager
}

// GetBackupUseCase returns the backup use case
func (as *AppState) GetBackupUseCase() *usecase.BackupUseCase {
	return as.backupUseCase
}

// GetAuditLogList returns the audit log list (lazy loading, auth required)
func (as *AppState) GetAuditLogList() *AuditLogList {
	if !as.isAuthenticated {
		return nil
	}

	if as.auditLogList == nil && as.auditRepo != nil && as.staffRepo != nil {
		as.auditLogList = NewAuditLogList(as.auditRepo, as.staffRepo, as.pdfService)

		// Load initial data
		go as.auditLogList.LoadData()
	}

	return as.auditLogList
}
