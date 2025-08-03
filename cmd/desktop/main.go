package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"shien-system/internal/adapter/backup"
	"shien-system/internal/adapter/crypto"
	"shien-system/internal/adapter/db"
	"shien-system/internal/adapter/pdf"
	"shien-system/internal/adapter/session"
	"shien-system/internal/config"
	"shien-system/internal/ui/theme"
	"shien-system/internal/ui/widgets"
	"shien-system/internal/usecase"
)

// consoleLogger implements backup.Logger interface for console output
type consoleLogger struct{}

func (c *consoleLogger) Info(msg string, fields ...interface{}) {
	fmt.Printf("[INFO] "+msg+"\n", fields...)
}

func (c *consoleLogger) Error(msg string, fields ...interface{}) {
	fmt.Printf("[ERROR] "+msg+"\n", fields...)
}

func (c *consoleLogger) Warn(msg string, fields ...interface{}) {
	fmt.Printf("[WARN] "+msg+"\n", fields...)
}

// Dependencies holds all initialized dependencies
type Dependencies struct {
	config             *config.Config
	database           *db.Database
	authUseCase        usecase.AuthUseCase
	recipientUseCase   usecase.RecipientUseCase
	certificateUseCase usecase.CertificateUseCase
	staffUseCase       usecase.StaffUseCase
	setupUseCase       usecase.SetupUseCase
	backupUseCase      *usecase.BackupUseCase
	pdfService         *pdf.PDFService

	// Repositories for direct access
	auditRepo *db.AuditLogRepository
	staffRepo *db.StaffRepository
}

// MainWindow holds the main window and reactive content
type MainWindow struct {
	window               fyne.Window
	appState             *widgets.AppState
	reactiveContent      *widgets.ReactiveContainer
	logoutHandler        *widgets.LogoutHandler
	feedbackManager      *widgets.FeedbackManager
	errorDialog          *widgets.ErrorDialog
	accessibilityManager *widgets.AccessibilityManager
}

func main() {
	// Load configuration first
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	myApp := app.New()
	myApp.Settings().SetTheme(theme.NewJapaneseTheme())

	myWindow := myApp.NewWindow("障害者サービス管理システム")
	myWindow.Resize(fyne.NewSize(1200, 800))

	// Initialize database and repositories with configuration
	dependencies, err := initializeDependencies(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}
	defer dependencies.database.Close()

	// Create main app state with authentication
	appState := widgets.NewAppState(dependencies.authUseCase, dependencies.recipientUseCase, dependencies.certificateUseCase, dependencies.staffUseCase, dependencies.setupUseCase, dependencies.backupUseCase, dependencies.auditRepo, dependencies.staffRepo, dependencies.pdfService, cfg)

	// Create main window with reactive content
	mainWindow := createMainWindow(myWindow, appState)
	defer mainWindow.reactiveContent.Destroy()

	myWindow.ShowAndRun()
}

// initializeDependencies initializes database and use cases
func initializeDependencies(cfg *config.Config) (*Dependencies, error) {
	// Initialize database with secure configuration
	dbConfig := db.Config{
		Path:         cfg.Database.Path,
		MigrationDir: config.GetMigrationDir(),
	}

	database, err := db.NewDatabase(dbConfig)
	if err != nil {
		return nil, err
	}

	// Run migrations
	ctx := context.Background()
	if err := database.RunMigrations(ctx); err != nil {
		database.Close()
		return nil, err
	}

	// Initialize repositories
	recipientRepo, err := db.NewRecipientRepository(database)
	if err != nil {
		database.Close()
		return nil, fmt.Errorf("failed to create recipient repository: %w", err)
	}
	
	staffRepo := db.NewStaffRepository(database)
	assignmentRepo := db.NewStaffAssignmentRepository(database)
	
	certificateRepo, err := db.NewBenefitCertificateRepository(database)
	if err != nil {
		database.Close()
		return nil, fmt.Errorf("failed to create certificate repository: %w", err)
	}
	
	auditRepo := db.NewAuditLogRepository(database)

	// Initialize crypto components
	passwordHasher := crypto.NewBcryptPasswordHasher()

	// Initialize session manager
	sessionManager := session.NewMemorySessionManager(24 * time.Hour)

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

	// Initialize PDF service with font path and cipher for encrypted fields
	fontPath := "assets/fonts"  // Default font path
	fieldCipher, err := crypto.NewFieldCipher()
	if err != nil {
		database.Close()
		return nil, fmt.Errorf("failed to create field cipher: %w", err)
	}
	pdfService := pdf.NewPDFService(fontPath, fieldCipher)

	// Initialize use cases
	authUseCase := usecase.NewAuthUseCase(
		staffRepo,
		auditRepo,
		passwordHasher,
		sessionManager,
		rateLimitSvc,
	)

	recipientUseCase := usecase.NewRecipientUseCase(
		recipientRepo,
		staffRepo,
		assignmentRepo,
		auditRepo,
	)

	certificateUseCase := usecase.NewCertificateUseCase(
		certificateRepo,
		recipientRepo,
		staffRepo,
		auditRepo,
	)

	staffUseCase := usecase.NewStaffUseCase(
		staffRepo,
		assignmentRepo,
		auditRepo,
	)

	setupUseCase := usecase.NewSetupUseCase(
		staffRepo,
		auditRepo,
		passwordHasher,
	)

	// Initialize backup service with proper logger
	backupLogger := &consoleLogger{}
	
	backupService := backup.NewService(database.DB(), fieldCipher, &cfg.Backup, backupLogger)
	
	// Initialize backup scheduler
	backupScheduler := backup.NewScheduler(backupService, &cfg.Backup, backupLogger)
	
	// Initialize backup use case
	backupUseCase := usecase.NewBackupUseCase(backupService, backupScheduler, auditRepo, backupLogger)

	return &Dependencies{
		config:             cfg,
		database:           database,
		authUseCase:        authUseCase,
		recipientUseCase:   recipientUseCase,
		certificateUseCase: certificateUseCase,
		staffUseCase:       staffUseCase,
		setupUseCase:       setupUseCase,
		backupUseCase:      backupUseCase,
		pdfService:         pdfService,
		auditRepo:          auditRepo,
		staffRepo:          staffRepo,
	}, nil
}

// createMainWindow creates the main window with reactive content
func createMainWindow(window fyne.Window, appState *widgets.AppState) *MainWindow {
	// Create error handling components
	logoutHandler := widgets.NewLogoutHandler(appState)
	feedbackManager := widgets.NewFeedbackManager()
	errorDialog := widgets.NewErrorDialog(window)

	// Create accessibility manager
	accessibilityManager := widgets.NewAccessibilityManager()
	accessibilityManager.SetEnabled(true)

	// Set error handling in app state
	appState.SetErrorHandling(feedbackManager, errorDialog)
	
	// Set window reference for dialogs
	appState.SetWindow(window)

	// Create reactive content that updates automatically
	reactiveContent := widgets.NewReactiveContainer(appState, func(state *widgets.AppState) fyne.CanvasObject {
		if !state.IsAuthenticated() {
			// Show login form when not authenticated
			return container.NewBorder(
				nil,
				feedbackManager.GetContainer(), // Show feedback at bottom
				nil, nil,
				container.NewCenter(state.CreateContent()),
			)
		}

		// Show main application when authenticated
		return container.NewBorder(
			createHeader(state, logoutHandler, errorDialog),
			container.NewVBox(
				feedbackManager.GetContainer(), // Show feedback above footer
				createFooter(),
			),
			createSidebar(state, feedbackManager, errorDialog, accessibilityManager),
			nil,
			state.CreateContent(),
		)
	})

	// Set window content
	window.SetContent(reactiveContent.GetContainer())

	// Setup keyboard shortcuts for accessibility
	setupKeyboardShortcuts(window, accessibilityManager)

	return &MainWindow{
		window:               window,
		appState:             appState,
		reactiveContent:      reactiveContent,
		logoutHandler:        logoutHandler,
		feedbackManager:      feedbackManager,
		errorDialog:          errorDialog,
		accessibilityManager: accessibilityManager,
	}
}

// setupKeyboardShortcuts sets up keyboard shortcuts for accessibility
func setupKeyboardShortcuts(window fyne.Window, accessibilityManager *widgets.AccessibilityManager) {
	canvas := window.Canvas()
	
	// Tab key for next focus
	canvas.AddShortcut(&desktop.CustomShortcut{
		KeyName: fyne.KeyTab,
	}, func(shortcut fyne.Shortcut) {
		accessibilityManager.NextFocus()
	})
	
	// Shift+Tab for previous focus
	canvas.AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyTab,
		Modifier: fyne.KeyModifierShift,
	}, func(shortcut fyne.Shortcut) {
		accessibilityManager.PreviousFocus()
	})
	
	// Alt+F for file menu
	canvas.AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyF,
		Modifier: fyne.KeyModifierAlt,
	}, func(shortcut fyne.Shortcut) {
		// TODO: Open file menu
	})
	
	// Ctrl+L for logout
	canvas.AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyL,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// TODO: Trigger logout
	})
	
	// F1 for help
	canvas.AddShortcut(&desktop.CustomShortcut{
		KeyName: fyne.KeyF1,
	}, func(shortcut fyne.Shortcut) {
		// TODO: Show help dialog
	})
}

func createHeader(appState *widgets.AppState, logoutHandler *widgets.LogoutHandler, errorDialog *widgets.ErrorDialog) *fyne.Container {
	var userInfo string
	if user := appState.GetCurrentUser(); user != nil {
		userInfo = user.Name + " (" + string(user.Role) + ")"
	}

	return container.NewBorder(
		nil, nil,
		widget.NewLabel("障害者サービス管理システム"),
		container.NewHBox(
			widget.NewLabel(userInfo),
			widget.NewButton("ログアウト", func() {
				// Show confirmation dialog before logout
				errorDialog.ShowConfirmation(
					"ログアウト確認",
					logoutHandler.GetConfirmationMessage(),
					func() {
						// Perform proper logout with session invalidation
						if err := logoutHandler.PerformLogout(); err != nil {
							log.Printf("Logout warning: %v", err)
							errorDialog.ShowError("ログアウトエラー", err)
						}
						// UI will automatically refresh via reactive container
					},
					func() {
						// Cancel logout - do nothing
					},
				)
			}),
		),
		nil,
	)
}

func createSidebar(appState *widgets.AppState, feedbackManager *widgets.FeedbackManager, errorDialog *widgets.ErrorDialog, accessibilityManager *widgets.AccessibilityManager) *fyne.Container {
	// Create accessible buttons for navigation
	recipientsBtn := widgets.NewAccessibleButton("利用者一覧", "利用者情報の一覧を表示します", func() {
		feedbackManager.ShowInfo("利用者一覧を表示中...")
		appState.SetCurrentView("recipients")
		// UI will automatically refresh via reactive container
	})
	recipientsBtn.SetShortcut("Alt+1")
	accessibilityManager.RegisterFocusable(recipientsBtn)

	staffBtn := widgets.NewAccessibleButton("担当者管理", "担当者の管理画面を表示します", func() {
		feedbackManager.ShowInfo("担当者管理を表示中...")
		appState.SetCurrentView("staff")
	})
	staffBtn.SetShortcut("Alt+2")
	accessibilityManager.RegisterFocusable(staffBtn)

	certificatesBtn := widgets.NewAccessibleButton("受給者証管理", "受給者証の管理画面を表示します", func() {
		feedbackManager.ShowInfo("受給者証管理を表示中...")
		appState.SetCurrentView("certificates")
	})
	certificatesBtn.SetShortcut("Alt+3")
	accessibilityManager.RegisterFocusable(certificatesBtn)

	auditBtn := widgets.NewAccessibleButton("監査ログ", "システムの監査ログを表示します", func() {
		feedbackManager.ShowInfo("監査ログを表示中...")
		appState.SetCurrentView("audit")
	})
	auditBtn.SetShortcut("Alt+4")
	accessibilityManager.RegisterFocusable(auditBtn)

	settingsBtn := widgets.NewAccessibleButton("設定", "システム設定画面を表示します", func() {
		feedbackManager.ShowInfo("設定を表示中...")
		appState.SetCurrentView("settings")
	})
	settingsBtn.SetShortcut("Alt+5")
	accessibilityManager.RegisterFocusable(settingsBtn)

	return container.NewVBox(
		recipientsBtn,
		staffBtn,
		certificatesBtn,
		auditBtn,
		widget.NewSeparator(),
		settingsBtn,
	)
}

func createFooter() *fyne.Container {
	return container.NewHBox(
		widget.NewLabel("Ready"),
	)
}
