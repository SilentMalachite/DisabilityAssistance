package widgets

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/layout"

	"shien-system/internal/domain"
	"shien-system/internal/usecase"
)

// CertificateForm represents a form for creating/editing benefit certificates
type CertificateForm struct {
	useCase     usecase.CertificateUseCase
	recipientUC usecase.RecipientUseCase

	// UI components - Certificate Information
	recipientSelect         *widget.Select
	startDateEntry          *widget.Entry
	endDateEntry            *widget.Entry
	issuerEntry             *widget.Entry
	serviceTypeEntry        *widget.Entry
	maxBenefitDaysEntry     *widget.Entry
	benefitDetailsEntry     *widget.Entry

	// Form controls
	saveButton   *widget.Button
	cancelButton *widget.Button

	// State
	isEditing     bool
	certificateID *domain.ID
	currentUser   *domain.Staff
	recipients    []*domain.Recipient

	// Event handlers
	onSaved     func(*domain.BenefitCertificate)
	onCancelled func()
}

// NewCertificateForm creates a new certificate form
func NewCertificateForm(certificateUC usecase.CertificateUseCase, recipientUC usecase.RecipientUseCase, currentUser *domain.Staff) *CertificateForm {
	cf := &CertificateForm{
		useCase:     certificateUC,
		recipientUC: recipientUC,
		currentUser: currentUser,
	}

	cf.createWidgets()
	cf.setupEventHandlers()

	return cf
}

// createWidgets initializes all form widgets
func (cf *CertificateForm) createWidgets() {
	// Certificate Information
	cf.recipientSelect = widget.NewSelect([]string{}, nil)
	cf.recipientSelect.PlaceHolder = "利用者を選択してください"

	cf.startDateEntry = widget.NewEntry()
	cf.startDateEntry.SetPlaceHolder("YYYY-MM-DD")
	cf.startDateEntry.Validator = cf.validateDateFormat

	cf.endDateEntry = widget.NewEntry()
	cf.endDateEntry.SetPlaceHolder("YYYY-MM-DD")
	cf.endDateEntry.Validator = cf.validateDateFormat

	cf.issuerEntry = widget.NewEntry()
	cf.issuerEntry.SetPlaceHolder("発行機関名を入力")

	cf.serviceTypeEntry = widget.NewEntry()
	cf.serviceTypeEntry.SetPlaceHolder("サービス種別を入力")

	cf.maxBenefitDaysEntry = widget.NewEntry()
	cf.maxBenefitDaysEntry.SetPlaceHolder("月あたりの給付日数上限")

	cf.benefitDetailsEntry = widget.NewMultiLineEntry()
	cf.benefitDetailsEntry.SetPlaceHolder("給付内容の詳細を入力")
	cf.benefitDetailsEntry.Wrapping = fyne.TextWrapWord

	// Form controls
	cf.saveButton = widget.NewButton("保存", nil)
	cf.cancelButton = widget.NewButton("キャンセル", nil)
}

// setupEventHandlers configures event handlers for form widgets
func (cf *CertificateForm) setupEventHandlers() {
	cf.saveButton.OnTapped = cf.handleSave
	cf.cancelButton.OnTapped = cf.handleCancel
}

// validateDateFormat validates date format
func (cf *CertificateForm) validateDateFormat(text string) error {
	if text == "" {
		return nil
	}
	_, err := time.Parse("2006-01-02", text)
	if err != nil {
		return fmt.Errorf("日付は YYYY-MM-DD 形式で入力してください")
	}
	return nil
}

// SetForEdit configures the form for editing an existing certificate
func (cf *CertificateForm) SetForEdit(certificate *domain.BenefitCertificate) {
	cf.isEditing = true
	cf.certificateID = &certificate.ID
	cf.saveButton.SetText("更新")

	// Load recipient data first
	cf.loadRecipientData()

	// Set form values
	cf.setRecipientInSelect(certificate.RecipientID)
	cf.startDateEntry.SetText(certificate.StartDate.Format("2006-01-02"))
	cf.endDateEntry.SetText(certificate.EndDate.Format("2006-01-02"))
	cf.issuerEntry.SetText(certificate.Issuer)
	cf.serviceTypeEntry.SetText(certificate.ServiceType)
	cf.maxBenefitDaysEntry.SetText(strconv.Itoa(certificate.MaxBenefitDaysPerMonth))
	cf.benefitDetailsEntry.SetText(certificate.BenefitDetails)
}

// SetForCreate configures the form for creating a new certificate
func (cf *CertificateForm) SetForCreate(recipientID *domain.ID) {
	cf.isEditing = false
	cf.certificateID = nil
	cf.saveButton.SetText("登録")
	cf.clearForm()

	// Load recipient data
	cf.loadRecipientData()

	// Pre-select recipient if provided
	if recipientID != nil {
		cf.setRecipientInSelect(*recipientID)
	}
}

// clearForm resets all form fields
func (cf *CertificateForm) clearForm() {
	cf.recipientSelect.SetSelected("")
	cf.startDateEntry.SetText("")
	cf.endDateEntry.SetText("")
	cf.issuerEntry.SetText("")
	cf.serviceTypeEntry.SetText("")
	cf.maxBenefitDaysEntry.SetText("")
	cf.benefitDetailsEntry.SetText("")
}

// loadRecipientData loads recipient options for the select widget
func (cf *CertificateForm) loadRecipientData() {
	ctx := context.Background()
	recipients, err := cf.recipientUC.GetActiveRecipients(ctx)
	if err != nil {
		cf.showError("利用者データの読み込みに失敗しました", err)
		return
	}

	cf.recipients = recipients
	options := make([]string, len(recipients))
	for i, recipient := range recipients {
		options[i] = fmt.Sprintf("%s (%s)", recipient.Name, recipient.Kana)
	}
	cf.recipientSelect.Options = options
}

// setRecipientInSelect sets the selected recipient by ID
func (cf *CertificateForm) setRecipientInSelect(recipientID domain.ID) {
	for i, recipient := range cf.recipients {
		if recipient.ID == recipientID {
			cf.recipientSelect.SetSelectedIndex(i)
			break
		}
	}
}

// getSelectedRecipientID returns the ID of the selected recipient
func (cf *CertificateForm) getSelectedRecipientID() *domain.ID {
	selectedIndex := cf.recipientSelect.SelectedIndex()
	if selectedIndex < 0 || selectedIndex >= len(cf.recipients) {
		return nil
	}
	return &cf.recipients[selectedIndex].ID
}

// handleSave processes the save action
func (cf *CertificateForm) handleSave() {
	if cf.isEditing {
		cf.handleUpdate()
	} else {
		cf.handleCreate()
	}
}

// handleCreate creates a new certificate
func (cf *CertificateForm) handleCreate() {
	cf.setFormEnabled(false)
	defer cf.setFormEnabled(true)

	req, err := cf.buildCreateRequest()
	if err != nil {
		cf.showError("入力内容を確認してください", err)
		return
	}

	ctx := context.Background()
	certificate, err := cf.useCase.CreateCertificate(ctx, *req)
	if err != nil {
		cf.showError("受給者証の登録に失敗しました", err)
		return
	}

	if cf.onSaved != nil {
		cf.onSaved(certificate)
	}
}

// handleUpdate updates an existing certificate
func (cf *CertificateForm) handleUpdate() {
	cf.setFormEnabled(false)
	defer cf.setFormEnabled(true)

	req, err := cf.buildUpdateRequest()
	if err != nil {
		cf.showError("入力内容を確認してください", err)
		return
	}

	ctx := context.Background()
	certificate, err := cf.useCase.UpdateCertificate(ctx, *req)
	if err != nil {
		cf.showError("受給者証の更新に失敗しました", err)
		return
	}

	if cf.onSaved != nil {
		cf.onSaved(certificate)
	}
}

// buildCreateRequest builds a certificate creation request from form data
func (cf *CertificateForm) buildCreateRequest() (*usecase.CreateCertificateRequest, error) {
	recipientID := cf.getSelectedRecipientID()
	if recipientID == nil {
		return nil, fmt.Errorf("利用者を選択してください")
	}

	startDate, err := cf.parseDateField(cf.startDateEntry.Text, "開始日")
	if err != nil {
		return nil, err
	}

	endDate, err := cf.parseDateField(cf.endDateEntry.Text, "終了日")
	if err != nil {
		return nil, err
	}

	if endDate.Before(startDate) {
		return nil, fmt.Errorf("終了日は開始日より後の日付を設定してください")
	}

	maxBenefitDays, err := cf.parseIntField(cf.maxBenefitDaysEntry.Text, "月あたりの給付日数上限")
	if err != nil {
		return nil, err
	}

	if err := cf.validateForm(); err != nil {
		return nil, err
	}

	return &usecase.CreateCertificateRequest{
		RecipientID:            *recipientID,
		StartDate:              startDate,
		EndDate:                endDate,
		Issuer:                 strings.TrimSpace(cf.issuerEntry.Text),
		ServiceType:            strings.TrimSpace(cf.serviceTypeEntry.Text),
		MaxBenefitDaysPerMonth: maxBenefitDays,
		BenefitDetails:         strings.TrimSpace(cf.benefitDetailsEntry.Text),
		ActorID:                cf.currentUser.ID,
	}, nil
}

// buildUpdateRequest builds a certificate update request from form data
func (cf *CertificateForm) buildUpdateRequest() (*usecase.UpdateCertificateRequest, error) {
	if cf.certificateID == nil {
		return nil, fmt.Errorf("更新対象の受給者証が選択されていません")
	}

	startDate, err := cf.parseDateField(cf.startDateEntry.Text, "開始日")
	if err != nil {
		return nil, err
	}

	endDate, err := cf.parseDateField(cf.endDateEntry.Text, "終了日")
	if err != nil {
		return nil, err
	}

	if endDate.Before(startDate) {
		return nil, fmt.Errorf("終了日は開始日より後の日付を設定してください")
	}

	maxBenefitDays, err := cf.parseIntField(cf.maxBenefitDaysEntry.Text, "月あたりの給付日数上限")
	if err != nil {
		return nil, err
	}

	if err := cf.validateForm(); err != nil {
		return nil, err
	}

	return &usecase.UpdateCertificateRequest{
		ID:                     *cf.certificateID,
		StartDate:              startDate,
		EndDate:                endDate,
		Issuer:                 strings.TrimSpace(cf.issuerEntry.Text),
		ServiceType:            strings.TrimSpace(cf.serviceTypeEntry.Text),
		MaxBenefitDaysPerMonth: maxBenefitDays,
		BenefitDetails:         strings.TrimSpace(cf.benefitDetailsEntry.Text),
		ActorID:                cf.currentUser.ID,
	}, nil
}

// validateForm validates the entire form
func (cf *CertificateForm) validateForm() error {
	if strings.TrimSpace(cf.issuerEntry.Text) == "" {
		return fmt.Errorf("発行機関名を入力してください")
	}

	if strings.TrimSpace(cf.serviceTypeEntry.Text) == "" {
		return fmt.Errorf("サービス種別を入力してください")
	}

	return nil
}

// parseDateField parses a date field with error context
func (cf *CertificateForm) parseDateField(text, fieldName string) (time.Time, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return time.Time{}, fmt.Errorf("%sを入力してください", fieldName)
	}

	date, err := time.Parse("2006-01-02", text)
	if err != nil {
		return time.Time{}, fmt.Errorf("%sは YYYY-MM-DD 形式で入力してください", fieldName)
	}

	return date, nil
}

// parseIntField parses an integer field with error context
func (cf *CertificateForm) parseIntField(text, fieldName string) (int, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, fmt.Errorf("%sを入力してください", fieldName)
	}

	value, err := strconv.Atoi(text)
	if err != nil {
		return 0, fmt.Errorf("%sは数値で入力してください", fieldName)
	}

	if value < 0 {
		return 0, fmt.Errorf("%sは0以上の値を入力してください", fieldName)
	}

	return value, nil
}

// setFormEnabled enables/disables form inputs
func (cf *CertificateForm) setFormEnabled(enabled bool) {
	cf.recipientSelect.Disable()
	if enabled {
		cf.recipientSelect.Enable()
	}

	cf.startDateEntry.Disable()
	if enabled {
		cf.startDateEntry.Enable()
	}

	cf.endDateEntry.Disable()
	if enabled {
		cf.endDateEntry.Enable()
	}

	cf.issuerEntry.Disable()
	if enabled {
		cf.issuerEntry.Enable()
	}

	cf.serviceTypeEntry.Disable()
	if enabled {
		cf.serviceTypeEntry.Enable()
	}

	cf.maxBenefitDaysEntry.Disable()
	if enabled {
		cf.maxBenefitDaysEntry.Enable()
	}

	cf.benefitDetailsEntry.Disable()
	if enabled {
		cf.benefitDetailsEntry.Enable()
	}

	cf.saveButton.Disable()
	if enabled {
		cf.saveButton.Enable()
	}

	cf.cancelButton.Disable()
	if enabled {
		cf.cancelButton.Enable()
	}
}

// handleCancel processes the cancel action
func (cf *CertificateForm) handleCancel() {
	if cf.onCancelled != nil {
		cf.onCancelled()
	}
}

// showError displays an error message
func (cf *CertificateForm) showError(title string, err error) {
	message := title
	if err != nil {
		message = fmt.Sprintf("%s\n\nエラー詳細: %v", title, err)
	}
	dialog.ShowError(fmt.Errorf(message), nil)
}

// CreateDialog creates a dialog containing the form
func (cf *CertificateForm) CreateDialog() dialog.Dialog {
	content := cf.CreateObject()
	
	title := "受給者証登録"
	if cf.isEditing {
		title = "受給者証編集"
	}

	return dialog.NewCustom(title, "閉じる", content, nil)
}

// CreateObject creates the complete form layout
func (cf *CertificateForm) CreateObject() fyne.CanvasObject {
	// Create form sections
	recipientSection := container.NewVBox(
		widget.NewLabel("利用者"),
		cf.recipientSelect,
	)

	datesSection := container.NewVBox(
		widget.NewLabel("有効期間"),
		container.NewGridWithColumns(2,
			container.NewVBox(
				widget.NewLabel("開始日"),
				cf.startDateEntry,
			),
			container.NewVBox(
				widget.NewLabel("終了日"),
				cf.endDateEntry,
			),
		),
	)

	serviceSection := container.NewVBox(
		widget.NewLabel("サービス情報"),
		container.NewVBox(
			widget.NewLabel("発行機関"),
			cf.issuerEntry,
		),
		container.NewVBox(
			widget.NewLabel("サービス種別"),
			cf.serviceTypeEntry,
		),
		container.NewVBox(
			widget.NewLabel("月あたりの給付日数上限"),
			cf.maxBenefitDaysEntry,
		),
	)

	detailsSection := container.NewVBox(
		widget.NewLabel("給付内容詳細"),
		container.NewBorder(nil, nil, nil, nil, cf.benefitDetailsEntry),
	)

	buttonsSection := container.NewHBox(
		layout.NewSpacer(),
		cf.cancelButton,
		cf.saveButton,
	)

	// Combine all sections
	return container.NewVBox(
		recipientSection,
		widget.NewSeparator(),
		datesSection,
		widget.NewSeparator(),
		serviceSection,
		widget.NewSeparator(),
		detailsSection,
		widget.NewSeparator(),
		buttonsSection,
	)
}

// SetOnSaved sets the callback for when a certificate is saved
func (cf *CertificateForm) SetOnSaved(callback func(*domain.BenefitCertificate)) {
	cf.onSaved = callback
}

// SetOnCancelled sets the callback for when the form is cancelled
func (cf *CertificateForm) SetOnCancelled(callback func()) {
	cf.onCancelled = callback
}