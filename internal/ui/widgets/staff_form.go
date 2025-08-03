package widgets

import (
	"context"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"shien-system/internal/domain"
	"shien-system/internal/usecase"
)

// StaffForm represents a form for creating and editing staff
type StaffForm struct {
	useCase usecase.StaffUseCase

	// Form widgets
	nameEntry   *widget.Entry
	roleSelect  *widget.Select

	// Control buttons
	saveButton   *widget.Button
	cancelButton *widget.Button

	// State
	isEditing   bool
	staffID     string
	currentUser *domain.Staff

	// Callbacks
	onSaved     func(*domain.Staff)
	onCancelled func()
}

// NewStaffForm creates a new staff form
func NewStaffForm(useCase usecase.StaffUseCase) *StaffForm {
	form := &StaffForm{
		useCase: useCase,
	}
	form.createWidgets()
	form.setupEventHandlers()
	return form
}

// createWidgets creates all form widgets
func (sf *StaffForm) createWidgets() {
	// Basic information
	sf.nameEntry = widget.NewEntry()
	sf.nameEntry.SetPlaceHolder("職員名を入力")

	// Role selection
	sf.roleSelect = widget.NewSelect([]string{
		"管理者",
		"職員", 
		"閲覧専用",
	}, nil)
	sf.roleSelect.SetSelected("職員") // Default to staff

	// Control buttons
	sf.saveButton = widget.NewButton("保存", sf.handleSave)
	sf.cancelButton = widget.NewButton("キャンセル", sf.handleCancel)
}

// setupEventHandlers sets up event handlers for form widgets
func (sf *StaffForm) setupEventHandlers() {
	// Enter key handling for save
	sf.nameEntry.OnSubmitted = func(string) {
		sf.handleSave()
	}
}

// SetForEdit configures the form for editing an existing staff member
func (sf *StaffForm) SetForEdit(staff *domain.Staff, currentUser *domain.Staff) {
	sf.isEditing = true
	sf.staffID = staff.ID
	sf.currentUser = currentUser

	// Populate form fields
	sf.nameEntry.SetText(staff.Name)
	sf.roleSelect.SetSelected(sf.formatRoleForSelect(staff.Role))

	sf.saveButton.SetText("更新")
}

// SetForCreate configures the form for creating a new staff member
func (sf *StaffForm) SetForCreate(currentUser *domain.Staff) {
	sf.isEditing = false
	sf.staffID = ""
	sf.currentUser = currentUser

	sf.clearForm()
	sf.saveButton.SetText("作成")
}

// clearForm clears all form fields
func (sf *StaffForm) clearForm() {
	sf.nameEntry.SetText("")
	sf.roleSelect.SetSelected("職員")
}

// handleSave processes the save action
func (sf *StaffForm) handleSave() {
	if sf.isEditing {
		sf.handleUpdate()
	} else {
		sf.handleCreate()
	}
}

// handleCreate creates a new staff member
func (sf *StaffForm) handleCreate() {
	if !sf.validateForm() {
		return
	}

	req := sf.buildCreateRequest()
	sf.setFormEnabled(false)

	ctx := context.Background()
	staff, err := sf.useCase.CreateStaff(ctx, req)
	sf.setFormEnabled(true)

	if err != nil {
		sf.showError("作成エラー", err)
		return
	}

	// Success - notify callback
	if sf.onSaved != nil {
		sf.onSaved(staff)
	}
}

// handleUpdate updates an existing staff member
func (sf *StaffForm) handleUpdate() {
	if !sf.validateForm() {
		return
	}

	req := sf.buildUpdateRequest()
	sf.setFormEnabled(false)

	ctx := context.Background()
	staff, err := sf.useCase.UpdateStaff(ctx, req)
	sf.setFormEnabled(true)

	if err != nil {
		sf.showError("更新エラー", err)
		return
	}

	// Success - notify callback
	if sf.onSaved != nil {
		sf.onSaved(staff)
	}
}

// buildCreateRequest builds a create request from form data
func (sf *StaffForm) buildCreateRequest() usecase.CreateStaffRequest {
	return usecase.CreateStaffRequest{
		Name:    strings.TrimSpace(sf.nameEntry.Text),
		Role:    sf.parseRoleFromSelect(sf.roleSelect.Selected),
		ActorID: sf.currentUser.ID,
	}
}

// buildUpdateRequest builds an update request from form data
func (sf *StaffForm) buildUpdateRequest() usecase.UpdateStaffRequest {
	return usecase.UpdateStaffRequest{
		ID:      sf.staffID,
		Name:    strings.TrimSpace(sf.nameEntry.Text),
		Role:    sf.parseRoleFromSelect(sf.roleSelect.Selected),
		ActorID: sf.currentUser.ID,
	}
}

// validateForm validates form input
func (sf *StaffForm) validateForm() bool {
	var errors []string

	// Validate name
	name := strings.TrimSpace(sf.nameEntry.Text)
	if name == "" {
		errors = append(errors, "職員名は必須です")
	}

	// Validate role selection
	if sf.roleSelect.Selected == "" {
		errors = append(errors, "ロールを選択してください")
	}

	if len(errors) > 0 {
		sf.showError("入力エラー", fmt.Errorf("以下の問題を解決してください:\n• %s", strings.Join(errors, "\n• ")))
		return false
	}

	return true
}

// formatRoleForSelect converts role enum to display string
func (sf *StaffForm) formatRoleForSelect(role domain.StaffRole) string {
	switch role {
	case domain.RoleAdmin:
		return "管理者"
	case domain.RoleStaff:
		return "職員"
	case domain.RoleReadOnly:
		return "閲覧専用"
	default:
		return "職員"
	}
}

// parseRoleFromSelect converts display string to role enum
func (sf *StaffForm) parseRoleFromSelect(selected string) domain.StaffRole {
	switch selected {
	case "管理者":
		return domain.RoleAdmin
	case "職員":
		return domain.RoleStaff
	case "閲覧専用":
		return domain.RoleReadOnly
	default:
		return domain.RoleStaff
	}
}

// setFormEnabled enables or disables form controls
func (sf *StaffForm) setFormEnabled(enabled bool) {
	sf.nameEntry.Disable()
	sf.roleSelect.Disable()
	sf.saveButton.Disable()
	sf.cancelButton.Disable()

	if enabled {
		sf.nameEntry.Enable()
		sf.roleSelect.Enable()
		sf.saveButton.Enable()
		sf.cancelButton.Enable()
	}
}

// handleCancel handles form cancellation
func (sf *StaffForm) handleCancel() {
	if sf.onCancelled != nil {
		sf.onCancelled()
	}
}

// showError displays an error message
func (sf *StaffForm) showError(title string, err error) {
	// This should be connected to the global error dialog
	// For now, we'll use a simple dialog
	fmt.Printf("Error: %s - %v\n", title, err)
}

// CreateDialog creates a dialog containing the form
func (sf *StaffForm) CreateDialog(parent fyne.Window) dialog.Dialog {
	content := sf.CreateObject()
	
	var title string
	if sf.isEditing {
		title = "職員情報編集"
	} else {
		title = "新規職員登録"
	}

	dlg := dialog.NewCustom(title, "閉じる", content, parent)
	dlg.Resize(fyne.NewSize(400, 300))
	return dlg
}

// CreateObject creates the form's UI object
func (sf *StaffForm) CreateObject() fyne.CanvasObject {
	// Form content
	form := container.NewVBox(
		// Basic information section
		widget.NewCard("基本情報", "", container.NewGridWithColumns(2,
			widget.NewLabel("職員名 *"),
			sf.nameEntry,
			widget.NewLabel("ロール *"),
			sf.roleSelect,
		)),

		// Action buttons
		container.NewHBox(
			sf.saveButton,
			sf.cancelButton,
		),
	)

	return container.NewScroll(form)
}

// SetOnSaved sets the callback for when a staff is saved
func (sf *StaffForm) SetOnSaved(callback func(*domain.Staff)) {
	sf.onSaved = callback
}

// SetOnCancelled sets the callback for when the form is cancelled
func (sf *StaffForm) SetOnCancelled(callback func()) {
	sf.onCancelled = callback
}