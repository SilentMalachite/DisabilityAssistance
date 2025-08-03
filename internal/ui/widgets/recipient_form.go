package widgets

import (
	"context"
	"fmt"
	"strings"
	"time"

	"shien-system/internal/domain"
	"shien-system/internal/usecase"
	"shien-system/internal/validation"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// RecipientForm represents the recipient creation/editing form
type RecipientForm struct {
	useCase usecase.RecipientUseCase

	// UI components - Basic Information
	nameEntry      *widget.Entry
	kanaEntry      *widget.Entry
	sexSelect      *widget.Select
	birthDateEntry *widget.Entry

	// UI components - Disability Information
	disabilityNameEntry  *widget.Entry
	hasDisabilityIDCheck *widget.Check
	gradeEntry           *widget.Entry

	// UI components - Contact Information
	addressEntry *widget.Entry
	phoneEntry   *widget.Entry
	emailEntry   *widget.Entry

	// UI components - Service Information
	publicAssistanceCheck *widget.Check
	admissionDateEntry    *widget.Entry
	dischargeDateEntry    *widget.Entry

	// Form controls
	saveButton   *widget.Button
	cancelButton *widget.Button

	// State
	isEditing   bool
	recipientID *domain.ID
	currentUser *domain.Staff

	// Event handlers
	onSaved     func(*domain.Recipient)
	onCancelled func()
}

// NewRecipientForm creates a new recipient form
func NewRecipientForm(useCase usecase.RecipientUseCase) *RecipientForm {
	form := &RecipientForm{
		useCase: useCase,
	}
	form.createWidgets()
	form.setupEventHandlers()
	return form
}

// createWidgets initializes all UI components
func (rf *RecipientForm) createWidgets() {
	// Basic Information
	rf.nameEntry = widget.NewEntry()
	rf.nameEntry.SetPlaceHolder("氏名（必須）")

	rf.kanaEntry = widget.NewEntry()
	rf.kanaEntry.SetPlaceHolder("フリガナ")

	rf.sexSelect = widget.NewSelect(
		[]string{"男性", "女性", "その他", "未設定"},
		func(selected string) {
			// Sex selection handler
		},
	)
	rf.sexSelect.Selected = "未設定"

	rf.birthDateEntry = widget.NewEntry()
	rf.birthDateEntry.SetPlaceHolder("生年月日 (YYYY/MM/DD)")

	// Disability Information
	rf.disabilityNameEntry = widget.NewEntry()
	rf.disabilityNameEntry.SetPlaceHolder("障害名")

	rf.hasDisabilityIDCheck = widget.NewCheck("身体障害者手帳等を保持", func(checked bool) {
		// Disability ID checkbox handler
	})

	rf.gradeEntry = widget.NewEntry()
	rf.gradeEntry.SetPlaceHolder("等級")

	// Contact Information
	rf.addressEntry = widget.NewEntry()
	rf.addressEntry.SetPlaceHolder("住所")
	rf.addressEntry.MultiLine = true

	rf.phoneEntry = widget.NewEntry()
	rf.phoneEntry.SetPlaceHolder("電話番号")

	rf.emailEntry = widget.NewEntry()
	rf.emailEntry.SetPlaceHolder("メールアドレス")

	// Service Information
	rf.publicAssistanceCheck = widget.NewCheck("生活保護", func(checked bool) {
		// Public assistance checkbox handler
	})

	rf.admissionDateEntry = widget.NewEntry()
	rf.admissionDateEntry.SetPlaceHolder("入所日 (YYYY/MM/DD)")

	rf.dischargeDateEntry = widget.NewEntry()
	rf.dischargeDateEntry.SetPlaceHolder("退所日 (YYYY/MM/DD) - 空白の場合は在籍中")

	// Form controls
	rf.saveButton = widget.NewButton("保存", func() {
		rf.handleSave()
	})
	rf.saveButton.Importance = widget.HighImportance

	rf.cancelButton = widget.NewButton("キャンセル", func() {
		rf.handleCancel()
	})
}

// setupEventHandlers configures event handlers
func (rf *RecipientForm) setupEventHandlers() {
	// Date validation on birth date
	rf.birthDateEntry.OnChanged = func(text string) {
		rf.validateDateFormat(text, "生年月日")
	}

	// Date validation on admission date
	rf.admissionDateEntry.OnChanged = func(text string) {
		if text != "" {
			rf.validateDateFormat(text, "入所日")
		}
	}

	// Date validation on discharge date
	rf.dischargeDateEntry.OnChanged = func(text string) {
		if text != "" {
			rf.validateDateFormat(text, "退所日")
		}
	}
}

// validateDateFormat validates date format and shows visual feedback
func (rf *RecipientForm) validateDateFormat(dateStr, fieldName string) {
	if dateStr == "" {
		return
	}

	_, err := time.Parse("2006/01/02", dateStr)
	if err != nil {
		// Show validation error in status or tooltip
		// For now, we'll just log it
		fmt.Printf("Invalid date format for %s: %s\n", fieldName, dateStr)
	}
}

// SetForEdit configures the form for editing an existing recipient
func (rf *RecipientForm) SetForEdit(recipient *domain.Recipient, currentUser *domain.Staff) {
	rf.isEditing = true
	rf.recipientID = &recipient.ID
	rf.currentUser = currentUser

	// Populate form fields
	rf.nameEntry.SetText(recipient.Name)
	rf.kanaEntry.SetText(recipient.Kana)
	rf.sexSelect.Selected = rf.formatSexForSelect(recipient.Sex)
	rf.birthDateEntry.SetText(recipient.BirthDate.Format("2006/01/02"))

	rf.disabilityNameEntry.SetText(recipient.DisabilityName)
	rf.hasDisabilityIDCheck.SetChecked(recipient.HasDisabilityID)
	rf.gradeEntry.SetText(recipient.Grade)

	rf.addressEntry.SetText(recipient.Address)
	rf.phoneEntry.SetText(recipient.Phone)
	rf.emailEntry.SetText(recipient.Email)

	rf.publicAssistanceCheck.SetChecked(recipient.PublicAssistance)

	if recipient.AdmissionDate != nil {
		rf.admissionDateEntry.SetText(recipient.AdmissionDate.Format("2006/01/02"))
	}
	if recipient.DischargeDate != nil {
		rf.dischargeDateEntry.SetText(recipient.DischargeDate.Format("2006/01/02"))
	}

	// Update button text
	rf.saveButton.SetText("更新")
}

// SetForCreate configures the form for creating a new recipient
func (rf *RecipientForm) SetForCreate(currentUser *domain.Staff) {
	rf.isEditing = false
	rf.recipientID = nil
	rf.currentUser = currentUser
	rf.clearForm()

	// Update button text
	rf.saveButton.SetText("保存")
}

// clearForm clears all form fields
func (rf *RecipientForm) clearForm() {
	rf.nameEntry.SetText("")
	rf.kanaEntry.SetText("")
	rf.sexSelect.Selected = "未設定"
	rf.birthDateEntry.SetText("")

	rf.disabilityNameEntry.SetText("")
	rf.hasDisabilityIDCheck.SetChecked(false)
	rf.gradeEntry.SetText("")

	rf.addressEntry.SetText("")
	rf.phoneEntry.SetText("")
	rf.emailEntry.SetText("")

	rf.publicAssistanceCheck.SetChecked(false)
	rf.admissionDateEntry.SetText("")
	rf.dischargeDateEntry.SetText("")
}

// handleSave processes form submission
func (rf *RecipientForm) handleSave() {
	// Validate required fields
	if err := rf.validateForm(); err != nil {
		rf.showError("入力エラー", err)
		return
	}

	// Disable form during save
	rf.setFormEnabled(false)

	ctx := context.Background()

	if rf.isEditing {
		rf.handleUpdate(ctx)
	} else {
		rf.handleCreate(ctx)
	}
}

// handleCreate creates a new recipient
func (rf *RecipientForm) handleCreate(ctx context.Context) {
	req, err := rf.buildCreateRequest()
	if err != nil {
		rf.setFormEnabled(true)
		rf.showError("データ作成エラー", err)
		return
	}

	recipient, err := rf.useCase.CreateRecipient(ctx, *req)
	if err != nil {
		rf.setFormEnabled(true)
		rf.showError("保存エラー", err)
		return
	}

	rf.setFormEnabled(true)
	if rf.onSaved != nil {
		rf.onSaved(recipient)
	}
}

// handleUpdate updates an existing recipient
func (rf *RecipientForm) handleUpdate(ctx context.Context) {
	req, err := rf.buildUpdateRequest()
	if err != nil {
		rf.setFormEnabled(true)
		rf.showError("データ更新エラー", err)
		return
	}

	recipient, err := rf.useCase.UpdateRecipient(ctx, *req)
	if err != nil {
		rf.setFormEnabled(true)
		rf.showError("更新エラー", err)
		return
	}

	rf.setFormEnabled(true)
	if rf.onSaved != nil {
		rf.onSaved(recipient)
	}
}

// buildCreateRequest builds a create request from form data
func (rf *RecipientForm) buildCreateRequest() (*usecase.CreateRecipientRequest, error) {
	birthDate, err := rf.parseDateField(rf.birthDateEntry.Text, "生年月日")
	if err != nil {
		return nil, err
	}

	req := &usecase.CreateRecipientRequest{
		Name:             strings.TrimSpace(rf.nameEntry.Text),
		Kana:             strings.TrimSpace(rf.kanaEntry.Text),
		Sex:              rf.parseSexFromSelect(rf.sexSelect.Selected),
		BirthDate:        birthDate,
		DisabilityName:   strings.TrimSpace(rf.disabilityNameEntry.Text),
		HasDisabilityID:  rf.hasDisabilityIDCheck.Checked,
		Grade:            strings.TrimSpace(rf.gradeEntry.Text),
		Address:          strings.TrimSpace(rf.addressEntry.Text),
		Phone:            strings.TrimSpace(rf.phoneEntry.Text),
		Email:            strings.TrimSpace(rf.emailEntry.Text),
		PublicAssistance: rf.publicAssistanceCheck.Checked,
		ActorID:          rf.currentUser.ID,
	}

	// Optional dates
	if admissionText := strings.TrimSpace(rf.admissionDateEntry.Text); admissionText != "" {
		admissionDate, err := rf.parseDateField(admissionText, "入所日")
		if err != nil {
			return nil, err
		}
		req.AdmissionDate = &admissionDate
	}

	// DischargeDate is not set during creation - only during updates

	return req, nil
}

// buildUpdateRequest builds an update request from form data
func (rf *RecipientForm) buildUpdateRequest() (*usecase.UpdateRecipientRequest, error) {
	if rf.recipientID == nil {
		return nil, fmt.Errorf("recipient ID is required for update")
	}

	birthDate, err := rf.parseDateField(rf.birthDateEntry.Text, "生年月日")
	if err != nil {
		return nil, err
	}

	req := &usecase.UpdateRecipientRequest{
		ID:               *rf.recipientID,
		Name:             strings.TrimSpace(rf.nameEntry.Text),
		Kana:             strings.TrimSpace(rf.kanaEntry.Text),
		Sex:              rf.parseSexFromSelect(rf.sexSelect.Selected),
		BirthDate:        birthDate,
		DisabilityName:   strings.TrimSpace(rf.disabilityNameEntry.Text),
		HasDisabilityID:  rf.hasDisabilityIDCheck.Checked,
		Grade:            strings.TrimSpace(rf.gradeEntry.Text),
		Address:          strings.TrimSpace(rf.addressEntry.Text),
		Phone:            strings.TrimSpace(rf.phoneEntry.Text),
		Email:            strings.TrimSpace(rf.emailEntry.Text),
		PublicAssistance: rf.publicAssistanceCheck.Checked,
		ActorID:          rf.currentUser.ID,
	}

	// Optional dates
	if admissionText := strings.TrimSpace(rf.admissionDateEntry.Text); admissionText != "" {
		admissionDate, err := rf.parseDateField(admissionText, "入所日")
		if err != nil {
			return nil, err
		}
		req.AdmissionDate = &admissionDate
	}

	if dischargeText := strings.TrimSpace(rf.dischargeDateEntry.Text); dischargeText != "" {
		dischargeDate, err := rf.parseDateField(dischargeText, "退所日")
		if err != nil {
			return nil, err
		}
		req.DischargeDate = &dischargeDate
	}

	return req, nil
}

// validateForm validates all form fields
func (rf *RecipientForm) validateForm() error {
	// Create form validator
	formValidator := validation.NewFormValidator()
	
	// Collect and sanitize form data
	formData := map[string]string{
		"name":            formValidator.SanitizeInput(rf.nameEntry.Text),
		"kana":            formValidator.SanitizeInput(rf.kanaEntry.Text),
		"birth_date":      strings.ReplaceAll(rf.birthDateEntry.Text, "/", "-"), // Convert format for validator
		"disability_name": formValidator.SanitizeInput(rf.disabilityNameEntry.Text),
		"grade":           formValidator.SanitizeInput(rf.gradeEntry.Text),
		"address":         formValidator.SanitizeInput(rf.addressEntry.Text),
		"phone":           formValidator.SanitizeInput(rf.phoneEntry.Text),
		"email":           formValidator.SanitizeInput(rf.emailEntry.Text),
		"admission_date":  strings.ReplaceAll(rf.admissionDateEntry.Text, "/", "-"),
		"discharge_date":  strings.ReplaceAll(rf.dischargeDateEntry.Text, "/", "-"),
	}
	
	// Validate using the comprehensive form validator
	if validationErrors := formValidator.ValidateRecipientForm(formData); len(validationErrors) > 0 {
		return fmt.Errorf(validationErrors.Error())
	}
	
	return nil
}

// parseDateField parses a date field with error handling
func (rf *RecipientForm) parseDateField(dateStr, fieldName string) (time.Time, error) {
	parsed, err := time.Parse("2006/01/02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("%sの形式が正しくありません (YYYY/MM/DD形式で入力してください)", fieldName)
	}
	return parsed, nil
}

// formatSexForSelect converts domain.Sex to select option
func (rf *RecipientForm) formatSexForSelect(sex domain.Sex) string {
	switch sex {
	case domain.SexMale:
		return "男性"
	case domain.SexFemale:
		return "女性"
	case domain.SexOther:
		return "その他"
	case domain.SexNA:
		return "未設定"
	default:
		return "未設定"
	}
}

// parseSexFromSelect converts select option to domain.Sex
func (rf *RecipientForm) parseSexFromSelect(selected string) domain.Sex {
	switch selected {
	case "男性":
		return domain.SexMale
	case "女性":
		return domain.SexFemale
	case "その他":
		return domain.SexOther
	default:
		return domain.SexNA
	}
}

// setFormEnabled enables or disables all form controls
func (rf *RecipientForm) setFormEnabled(enabled bool) {
	if enabled {
		rf.nameEntry.Enable()
		rf.kanaEntry.Enable()
		rf.sexSelect.Enable()
		rf.birthDateEntry.Enable()
		rf.disabilityNameEntry.Enable()
		rf.hasDisabilityIDCheck.Enable()
		rf.gradeEntry.Enable()
		rf.addressEntry.Enable()
		rf.phoneEntry.Enable()
		rf.emailEntry.Enable()
		rf.publicAssistanceCheck.Enable()
		rf.admissionDateEntry.Enable()
		rf.dischargeDateEntry.Enable()
		rf.saveButton.Enable()
		rf.cancelButton.Enable()
	} else {
		rf.nameEntry.Disable()
		rf.kanaEntry.Disable()
		rf.sexSelect.Disable()
		rf.birthDateEntry.Disable()
		rf.disabilityNameEntry.Disable()
		rf.hasDisabilityIDCheck.Disable()
		rf.gradeEntry.Disable()
		rf.addressEntry.Disable()
		rf.phoneEntry.Disable()
		rf.emailEntry.Disable()
		rf.publicAssistanceCheck.Disable()
		rf.admissionDateEntry.Disable()
		rf.dischargeDateEntry.Disable()
		rf.saveButton.Disable()
		rf.cancelButton.Disable()
	}
}

// handleCancel handles form cancellation
func (rf *RecipientForm) handleCancel() {
	if rf.onCancelled != nil {
		rf.onCancelled()
	}
}

// showError displays an error dialog
func (rf *RecipientForm) showError(title string, err error) {
	// This should be provided by the parent or app state
	fmt.Printf("Error %s: %v\n", title, err)
}

// CreateDialog creates a modal dialog containing the form
func (rf *RecipientForm) CreateDialog(parent fyne.Window) *dialog.CustomDialog {
	content := rf.CreateObject()

	var title string
	if rf.isEditing {
		title = "利用者情報編集"
	} else {
		title = "新規利用者登録"
	}

	return dialog.NewCustom(title, "", content, parent)
}

// CreateObject creates the main UI object for this form
func (rf *RecipientForm) CreateObject() fyne.CanvasObject {
	// Basic Information Section
	basicInfo := container.NewVBox(
		widget.NewLabel("基本情報"),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			widget.NewLabel("氏名*:"), rf.nameEntry,
			widget.NewLabel("フリガナ:"), rf.kanaEntry,
			widget.NewLabel("性別:"), rf.sexSelect,
			widget.NewLabel("生年月日*:"), rf.birthDateEntry,
		),
	)

	// Disability Information Section
	disabilityInfo := container.NewVBox(
		widget.NewLabel("障害情報"),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			widget.NewLabel("障害名:"), rf.disabilityNameEntry,
			widget.NewLabel("等級:"), rf.gradeEntry,
		),
		rf.hasDisabilityIDCheck,
	)

	// Contact Information Section
	contactInfo := container.NewVBox(
		widget.NewLabel("連絡先情報"),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			widget.NewLabel("住所:"), rf.addressEntry,
			widget.NewLabel("電話番号:"), rf.phoneEntry,
			widget.NewLabel("メール:"), rf.emailEntry,
		),
	)

	// Service Information Section
	serviceInfo := container.NewVBox(
		widget.NewLabel("サービス情報"),
		widget.NewSeparator(),
		rf.publicAssistanceCheck,
		container.NewGridWithColumns(2,
			widget.NewLabel("入所日:"), rf.admissionDateEntry,
			widget.NewLabel("退所日:"), rf.dischargeDateEntry,
		),
	)

	// Form controls
	controls := container.NewHBox(
		rf.saveButton,
		rf.cancelButton,
	)

	// Main layout with scroll container for better UX
	formContent := container.NewVBox(
		basicInfo,
		widget.NewSeparator(),
		disabilityInfo,
		widget.NewSeparator(),
		contactInfo,
		widget.NewSeparator(),
		serviceInfo,
		widget.NewSeparator(),
		controls,
	)

	return container.NewScroll(formContent)
}

// SetOnSaved sets the callback for successful save
func (rf *RecipientForm) SetOnSaved(callback func(*domain.Recipient)) {
	rf.onSaved = callback
}

// SetOnCancelled sets the callback for form cancellation
func (rf *RecipientForm) SetOnCancelled(callback func()) {
	rf.onCancelled = callback
}
