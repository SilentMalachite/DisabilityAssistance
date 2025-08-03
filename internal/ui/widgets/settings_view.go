package widgets

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/layout"

	"shien-system/internal/config"
)

// SettingsView represents the application settings interface
type SettingsView struct {
	// Configuration
	config *config.Config

	// Theme settings section
	themeGroup       *widget.Card
	themeModeSelect  *widget.Select
	fontSizeSlider   *widget.Slider
	fontSizeLabel    *widget.Label

	// Application settings section
	applicationGroup       *widget.Card
	dbPathEntry           *widget.Entry
	dbPathBrowseButton    *widget.Button
	backupEnabledCheck    *widget.Check
	backupIntervalSelect  *widget.Select
	backupPathEntry       *widget.Entry
	backupPathBrowseButton *widget.Button

	// Control buttons
	saveButton   *widget.Button
	resetButton  *widget.Button
	cancelButton *widget.Button

	// Event handlers
	onSaved     func()
	onCancelled func()

	// State
	hasChanges bool
}

// NewSettingsView creates a new settings view
func NewSettingsView(cfg *config.Config) *SettingsView {
	sv := &SettingsView{
		config: cfg,
	}

	sv.createWidgets()
	sv.setupEventHandlers()
	sv.loadCurrentSettings()

	return sv
}

// createWidgets initializes all setting widgets
func (sv *SettingsView) createWidgets() {
	// Theme settings
	sv.themeModeSelect = widget.NewSelect(
		[]string{"ライトモード", "ダークモード", "システム設定に従う"},
		nil,
	)
	sv.themeModeSelect.PlaceHolder = "テーマモードを選択"

	sv.fontSizeSlider = widget.NewSlider(10, 24)
	sv.fontSizeSlider.Step = 1
	sv.fontSizeLabel = widget.NewLabel("フォントサイズ: 14px")

	sv.themeGroup = widget.NewCard("テーマ設定", "",
		container.NewVBox(
			widget.NewLabel("表示モード"),
			sv.themeModeSelect,
			container.NewVBox(
				sv.fontSizeLabel,
				sv.fontSizeSlider,
			),
		),
	)

	// Application settings
	sv.dbPathEntry = widget.NewEntry()
	sv.dbPathEntry.SetPlaceHolder("データベースファイルのパス")
	sv.dbPathBrowseButton = widget.NewButton("参照...", nil)

	sv.backupEnabledCheck = widget.NewCheck("自動バックアップを有効にする", nil)
	sv.backupIntervalSelect = widget.NewSelect(
		[]string{"毎日", "毎週", "毎月", "手動のみ"},
		nil,
	)
	sv.backupIntervalSelect.PlaceHolder = "バックアップ間隔"

	sv.backupPathEntry = widget.NewEntry()
	sv.backupPathEntry.SetPlaceHolder("バックアップ保存先")
	sv.backupPathBrowseButton = widget.NewButton("参照...", nil)

	sv.applicationGroup = widget.NewCard("アプリケーション設定", "",
		container.NewVBox(
			container.NewBorder(nil, nil, nil, sv.dbPathBrowseButton, sv.dbPathEntry),
			sv.backupEnabledCheck,
			widget.NewLabel("バックアップ間隔"),
			sv.backupIntervalSelect,
			container.NewBorder(nil, nil, nil, sv.backupPathBrowseButton, sv.backupPathEntry),
		),
	)

	// Control buttons
	sv.saveButton = widget.NewButton("設定を保存", nil)
	sv.resetButton = widget.NewButton("デフォルトに戻す", nil)
	sv.cancelButton = widget.NewButton("キャンセル", nil)
}

// setupEventHandlers configures event handlers for settings widgets
func (sv *SettingsView) setupEventHandlers() {
	// Theme settings handlers
	sv.themeModeSelect.OnChanged = func(string) {
		sv.hasChanges = true
	}

	sv.fontSizeSlider.OnChanged = func(value float64) {
		sv.fontSizeLabel.SetText(fmt.Sprintf("フォントサイズ: %.0fpx", value))
		sv.hasChanges = true
	}

	// Application settings handlers
	sv.dbPathEntry.OnChanged = func(string) {
		sv.hasChanges = true
	}

	sv.dbPathBrowseButton.OnTapped = func() {
		sv.showDBPathDialog()
	}

	sv.backupEnabledCheck.OnChanged = func(enabled bool) {
		sv.backupIntervalSelect.Enable()
		sv.backupPathEntry.Enable()
		sv.backupPathBrowseButton.Enable()
		if !enabled {
			sv.backupIntervalSelect.Disable()
			sv.backupPathEntry.Disable()
			sv.backupPathBrowseButton.Disable()
		}
		sv.hasChanges = true
	}

	sv.backupIntervalSelect.OnChanged = func(string) {
		sv.hasChanges = true
	}

	sv.backupPathEntry.OnChanged = func(string) {
		sv.hasChanges = true
	}

	sv.backupPathBrowseButton.OnTapped = func() {
		sv.showBackupPathDialog()
	}

	// Control button handlers
	sv.saveButton.OnTapped = sv.handleSave
	sv.resetButton.OnTapped = sv.handleReset
	sv.cancelButton.OnTapped = sv.handleCancel
}

// loadCurrentSettings loads current configuration values into widgets
func (sv *SettingsView) loadCurrentSettings() {
	if sv.config == nil {
		return
	}

	// Theme settings
	switch sv.config.UI.Theme {
	case "light":
		sv.themeModeSelect.SetSelected("ライトモード")
	case "dark":
		sv.themeModeSelect.SetSelected("ダークモード")
	default:
		sv.themeModeSelect.SetSelected("システム設定に従う")
	}

	sv.fontSizeSlider.SetValue(float64(sv.config.UI.FontSize))
	sv.fontSizeLabel.SetText(fmt.Sprintf("フォントサイズ: %dpx", sv.config.UI.FontSize))

	// Application settings
	sv.dbPathEntry.SetText(sv.config.Database.Path)
	sv.backupEnabledCheck.SetChecked(sv.config.Backup.Enabled)

	switch sv.config.Backup.ScheduleInterval {
	case "daily":
		sv.backupIntervalSelect.SetSelected("毎日")
	case "weekly":
		sv.backupIntervalSelect.SetSelected("毎週")
	case "monthly":
		sv.backupIntervalSelect.SetSelected("毎月")
	default:
		sv.backupIntervalSelect.SetSelected("手動のみ")
	}

	sv.backupPathEntry.SetText(sv.config.Backup.BackupDir)

	// Enable/disable backup settings based on backup enabled state
	if !sv.config.Backup.Enabled {
		sv.backupIntervalSelect.Disable()
		sv.backupPathEntry.Disable()
		sv.backupPathBrowseButton.Disable()
	}

	sv.hasChanges = false
}

// showDBPathDialog shows a file dialog for selecting database path
func (sv *SettingsView) showDBPathDialog() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()
		
		sv.dbPathEntry.SetText(reader.URI().Path())
		sv.hasChanges = true
	}, nil)
}

// showBackupPathDialog shows a folder dialog for selecting backup path
func (sv *SettingsView) showBackupPathDialog() {
	dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
		if err != nil || dir == nil {
			return
		}
		
		sv.backupPathEntry.SetText(dir.Path())
		sv.hasChanges = true
	}, nil)
}

// handleSave processes the save action
func (sv *SettingsView) handleSave() {
	if !sv.hasChanges {
		if sv.onSaved != nil {
			sv.onSaved()
		}
		return
	}

	// Validate settings
	if err := sv.validateSettings(); err != nil {
		sv.showError("設定エラー", err)
		return
	}

	// Apply settings to config
	sv.applySettings()

	// Save configuration
	if err := config.SaveConfig(sv.config); err != nil {
		sv.showError("保存エラー", fmt.Errorf("設定の保存に失敗しました: %v", err))
		return
	}

	sv.hasChanges = false
	sv.showInfo("設定を保存しました")

	if sv.onSaved != nil {
		sv.onSaved()
	}
}

// handleReset resets all settings to defaults
func (sv *SettingsView) handleReset() {
	dialog.ShowConfirm("設定をリセット", 
		"すべての設定をデフォルト値に戻しますか？\nこの操作は元に戻せません。",
		func(confirmed bool) {
			if confirmed {
				sv.resetToDefaults()
			}
		}, nil)
}

// handleCancel processes the cancel action
func (sv *SettingsView) handleCancel() {
	if sv.hasChanges {
		dialog.ShowConfirm("変更を破棄", 
			"未保存の変更があります。変更を破棄してよろしいですか？",
			func(confirmed bool) {
				if confirmed {
					sv.loadCurrentSettings() // Reload original settings
					if sv.onCancelled != nil {
						sv.onCancelled()
					}
				}
			}, nil)
	} else {
		if sv.onCancelled != nil {
			sv.onCancelled()
		}
	}
}

// validateSettings validates all setting values
func (sv *SettingsView) validateSettings() error {
	// Validate database path
	if sv.dbPathEntry.Text == "" {
		return fmt.Errorf("データベースファイルのパスを入力してください")
	}

	// Validate backup settings if enabled
	if sv.backupEnabledCheck.Checked {
		if sv.backupPathEntry.Text == "" {
			return fmt.Errorf("バックアップ保存先を設定してください")
		}
		if sv.backupIntervalSelect.Selected == "" {
			return fmt.Errorf("バックアップ間隔を選択してください")
		}
	}

	return nil
}

// applySettings applies widget values to configuration
func (sv *SettingsView) applySettings() {
	// Theme settings
	switch sv.themeModeSelect.Selected {
	case "ライトモード":
		sv.config.UI.Theme = "light"
	case "ダークモード":
		sv.config.UI.Theme = "dark"
	default:
		sv.config.UI.Theme = "auto"
	}

	sv.config.UI.FontSize = int(sv.fontSizeSlider.Value)

	// Application settings
	sv.config.Database.Path = sv.dbPathEntry.Text
	sv.config.Backup.Enabled = sv.backupEnabledCheck.Checked
	sv.config.Backup.BackupDir = sv.backupPathEntry.Text

	switch sv.backupIntervalSelect.Selected {
	case "毎日":
		sv.config.Backup.ScheduleInterval = "daily"
	case "毎週":
		sv.config.Backup.ScheduleInterval = "weekly"
	case "毎月":
		sv.config.Backup.ScheduleInterval = "monthly"
	default:
		sv.config.Backup.ScheduleInterval = "manual"
	}
}

// resetToDefaults resets all settings to default values
func (sv *SettingsView) resetToDefaults() {
	// Reset theme settings
	sv.themeModeSelect.SetSelected("システム設定に従う")
	sv.fontSizeSlider.SetValue(14)
	sv.fontSizeLabel.SetText("フォントサイズ: 14px")

	// Reset application settings
	sv.dbPathEntry.SetText("shien-system.db")
	sv.backupEnabledCheck.SetChecked(true)
	sv.backupIntervalSelect.SetSelected("毎日")
	sv.backupPathEntry.SetText("./backups")

	sv.hasChanges = true
}

// showError displays an error message
func (sv *SettingsView) showError(title string, err error) {
	message := title
	if err != nil {
		message = fmt.Sprintf("%s\n\nエラー詳細: %v", title, err)
	}
	dialog.ShowError(fmt.Errorf(message), nil)
}

// showInfo displays an information message
func (sv *SettingsView) showInfo(message string) {
	dialog.ShowInformation("設定", message, nil)
}

// CreateObject creates the complete settings layout
func (sv *SettingsView) CreateObject() fyne.CanvasObject {
	// Create main content with scrolling
	content := container.NewVBox(
		sv.themeGroup,
		sv.applicationGroup,
	)

	scrollContent := container.NewScroll(content)
	scrollContent.SetMinSize(content.MinSize())

	// Create button bar
	buttonBar := container.NewHBox(
		layout.NewSpacer(),
		sv.resetButton,
		sv.cancelButton,
		sv.saveButton,
	)

	// Combine content and buttons
	return container.NewBorder(
		nil,          // top
		buttonBar,    // bottom
		nil,          // left
		nil,          // right
		scrollContent, // center
	)
}

// SetOnSaved sets the callback for when settings are saved
func (sv *SettingsView) SetOnSaved(callback func()) {
	sv.onSaved = callback
}

// SetOnCancelled sets the callback for when settings dialog is cancelled
func (sv *SettingsView) SetOnCancelled(callback func()) {
	sv.onCancelled = callback
}

// HasChanges returns whether there are unsaved changes
func (sv *SettingsView) HasChanges() bool {
	return sv.hasChanges
}