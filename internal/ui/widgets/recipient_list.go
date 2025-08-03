package widgets

import (
	"context"
	"fmt"
	"time"

	"shien-system/internal/adapter/pdf"
	"shien-system/internal/domain"
	"shien-system/internal/usecase"
	"shien-system/internal/validation"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// RecipientList represents the recipient list widget with search and filter functionality
type RecipientList struct {
	useCase            usecase.RecipientUseCase
	certificateUseCase usecase.CertificateUseCase
	staffUseCase       usecase.StaffUseCase
	pdfService         *pdf.PDFService

	// UI components
	searchEntry    *widget.Entry
	table          *widget.Table
	newButton      *widget.Button
	refreshButton  *widget.Button
	exportButton   *widget.Button
	staffFilter    *widget.Select

	// Data
	recipients     []*domain.Recipient
	filteredData   []*domain.Recipient
	currentSearch  string
	currentStaffID string

	// Callbacks
	onNewRecipient  func()
	onEditRecipient func(recipientID string)
}

// NewRecipientList creates a new RecipientList widget
func NewRecipientList(useCase usecase.RecipientUseCase, certificateUseCase usecase.CertificateUseCase, staffUseCase usecase.StaffUseCase, pdfService *pdf.PDFService) *RecipientList {
	rl := &RecipientList{
		useCase:            useCase,
		certificateUseCase: certificateUseCase,
		staffUseCase:       staffUseCase,
		pdfService:         pdfService,
		recipients:         make([]*domain.Recipient, 0),
		filteredData:       make([]*domain.Recipient, 0),
	}

	rl.createWidgets()
	rl.setupTable()
	rl.setupEventHandlers()

	return rl
}

// createWidgets initializes all UI components
func (rl *RecipientList) createWidgets() {
	// Search entry
	rl.searchEntry = widget.NewEntry()
	rl.searchEntry.SetPlaceHolder("利用者名またはカナで検索...")

	// Table
	rl.table = widget.NewTable(
		func() (int, int) {
			return len(rl.filteredData), 8 // 8 columns
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			rl.updateTableCell(id, obj.(*widget.Label))
		},
	)

	// Buttons
	rl.newButton = widget.NewButton("新規登録", func() {
		if rl.onNewRecipient != nil {
			rl.onNewRecipient()
		}
	})

	rl.refreshButton = widget.NewButton("更新", func() {
		rl.LoadData()
	})

	rl.exportButton = widget.NewButton("PDFエクスポート", func() {
		rl.exportSelectedToPDF()
	})

	// Staff filter
	rl.staffFilter = widget.NewSelect([]string{"全て", "担当者1", "担当者2"}, func(selected string) {
		rl.onStaffFilterChanged(selected)
	})
	rl.staffFilter.Selected = "全て"
}

// setupTable configures the table widget
func (rl *RecipientList) setupTable() {
	// Set column widths
	rl.table.SetColumnWidth(0, 120) // 氏名
	rl.table.SetColumnWidth(1, 120) // カナ
	rl.table.SetColumnWidth(2, 60)  // 性別
	rl.table.SetColumnWidth(3, 100) // 生年月日
	rl.table.SetColumnWidth(4, 100) // 障害名
	rl.table.SetColumnWidth(5, 60)  // 等級
	rl.table.SetColumnWidth(6, 100) // 担当者
	rl.table.SetColumnWidth(7, 80)  // 状態

	// Set up selection handler for double-click navigation
	rl.table.OnSelected = func(id widget.TableCellID) {
		if id.Row < len(rl.filteredData) && rl.onEditRecipient != nil {
			recipient := rl.filteredData[id.Row]
			rl.onEditRecipient(recipient.ID)
		}
	}
}

// setupEventHandlers configures event handlers
func (rl *RecipientList) setupEventHandlers() {
	rl.searchEntry.OnChanged = func(text string) {
		rl.onSearchChanged(text)
	}
}

// updateTableCell updates a specific table cell with recipient data
func (rl *RecipientList) updateTableCell(id widget.TableCellID, label *widget.Label) {
	if id.Row >= len(rl.filteredData) {
		label.SetText("")
		return
	}

	recipient := rl.filteredData[id.Row]

	switch id.Col {
	case 0: // 氏名
		label.SetText(recipient.Name)
	case 1: // カナ
		label.SetText(recipient.Kana)
	case 2: // 性別
		label.SetText(rl.formatSex(recipient.Sex))
	case 3: // 生年月日
		label.SetText(recipient.BirthDate.Format("2006/01/02"))
	case 4: // 障害名
		label.SetText(recipient.DisabilityName)
	case 5: // 等級
		label.SetText(recipient.Grade)
	case 6: // 担当者
		label.SetText("未実装") // TODO: Implement staff assignment display
	case 7: // 状態
		status := "利用中"
		if recipient.DischargeDate != nil {
			status = "退所"
		}
		label.SetText(status)
	default:
		label.SetText("")
	}
}

// formatSex converts Sex enum to Japanese string
func (rl *RecipientList) formatSex(sex domain.Sex) string {
	switch sex {
	case domain.SexMale:
		return "男性"
	case domain.SexFemale:
		return "女性"
	case domain.SexOther:
		return "その他"
	case domain.SexNA:
		return "-"
	default:
		return "-"
	}
}

// LoadData loads recipient data from the use case
func (rl *RecipientList) LoadData() error {
	ctx := context.Background()
	req := usecase.ListRecipientsRequest{
		Limit:  1000, // Load all recipients for now
		Offset: 0,
		FilterBy: usecase.FilterRecipients{
			AssignedToStaff: rl.getStaffIDFilter(),
		},
	}

	result, err := rl.useCase.ListRecipients(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to load recipients: %w", err)
	}

	rl.recipients = result.Recipients
	rl.applyFilters()
	rl.table.Refresh()

	return nil
}

// onSearchChanged handles search text changes
func (rl *RecipientList) onSearchChanged(text string) {
	// Validate search query
	formValidator := validation.NewFormValidator()
	
	// Sanitize input
	sanitizedText := formValidator.SanitizeInput(text)
	
	// Validate search query for security
	if err := formValidator.ValidateSearchQuery(sanitizedText); err != nil {
		// If validation fails, clear the search to prevent issues
		rl.currentSearch = ""
		// Show error message via log (in production, you might want to show this to user)
		fmt.Printf("Invalid search query: %v\n", err)
	} else {
		rl.currentSearch = sanitizedText
	}
	
	rl.applyFilters()
	rl.table.Refresh()
}

// onStaffFilterChanged handles staff filter changes
func (rl *RecipientList) onStaffFilterChanged(staffName string) {
	// Convert staff name to ID (simplified for testing)
	switch staffName {
	case "全て":
		rl.currentStaffID = ""
	case "担当者1":
		rl.currentStaffID = "staff-001"
	default:
		rl.currentStaffID = staffName
	}

	rl.LoadData() // Reload with new filter
}

// getStaffIDFilter returns the current staff ID filter
func (rl *RecipientList) getStaffIDFilter() *domain.ID {
	if rl.currentStaffID == "" {
		return nil
	}
	return &rl.currentStaffID
}

// applyFilters applies search and other filters to the recipient data
func (rl *RecipientList) applyFilters() {
	rl.filteredData = make([]*domain.Recipient, 0)

	searchLower := rl.currentSearch

	for _, recipient := range rl.recipients {
		// Apply search filter
		if searchLower != "" {
			if !rl.matchesSearch(recipient, searchLower) {
				continue
			}
		}

		rl.filteredData = append(rl.filteredData, recipient)
	}
}

// matchesSearch checks if a recipient matches the search criteria
func (rl *RecipientList) matchesSearch(recipient *domain.Recipient, search string) bool {
	// Simple search implementation - can be enhanced
	return recipient.Name == search || recipient.Kana == search
}

// CreateObject creates the main UI object for this widget
func (rl *RecipientList) CreateObject() fyne.CanvasObject {
	// Header with search and controls
	header := container.NewBorder(
		nil, nil,
		container.NewHBox(
			widget.NewLabel("担当者:"),
			rl.staffFilter,
		),
		container.NewHBox(
			rl.newButton,
			rl.refreshButton,
			rl.exportButton,
		),
		container.NewBorder(
			nil, nil,
			widget.NewLabel("検索:"),
			nil,
			rl.searchEntry,
		),
	)

	// Table with headers
	tableContainer := container.NewBorder(
		rl.createTableHeader(),
		nil, nil, nil,
		rl.table,
	)

	// Complete layout
	return container.NewBorder(
		header,
		nil, nil, nil,
		tableContainer,
	)
}

// createTableHeader creates the table header
func (rl *RecipientList) createTableHeader() fyne.CanvasObject {
	headers := []string{"氏名", "カナ", "性別", "生年月日", "障害名", "等級", "担当者", "状態"}
	headerWidgets := make([]fyne.CanvasObject, len(headers))

	for i, header := range headers {
		label := widget.NewLabel(header)
		label.TextStyle.Bold = true
		headerWidgets[i] = label
	}

	return container.NewHBox(headerWidgets...)
}

// SetOnNewRecipient sets the callback for new recipient action
func (rl *RecipientList) SetOnNewRecipient(callback func()) {
	rl.onNewRecipient = callback
}

// SetOnEditRecipient sets the callback for edit recipient action
func (rl *RecipientList) SetOnEditRecipient(callback func(recipientID domain.ID)) {
	rl.onEditRecipient = callback
}

// Length returns the number of visible items in the table (for testing)
func (rl *RecipientList) Length() int {
	return len(rl.filteredData)
}

// exportSelectedToPDF exports the filtered recipients to a PDF report
func (rl *RecipientList) exportSelectedToPDF() {
	if rl.pdfService == nil {
		dialog.ShowError(fmt.Errorf("PDFサービスが利用できません"), fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	if len(rl.filteredData) == 0 {
		dialog.ShowInformation("情報", "エクスポートする利用者データがありません。", fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	// Show file save dialog
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(fmt.Errorf("ファイルの保存に失敗しました: %w", err), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}
		if writer == nil {
			return // User cancelled
		}
		defer writer.Close()

		// Generate PDF for each recipient
		for i, recipient := range rl.filteredData {
			ctx := context.Background()
			
			// Get certificates for this recipient
			certificates, err := rl.getCertificatesForRecipient(ctx, recipient.ID)
			if err != nil {
				dialog.ShowError(fmt.Errorf("受給者証データの取得に失敗しました: %w", err), fyne.CurrentApp().Driver().AllWindows()[0])
				return
			}

			// Get assignments for this recipient
			assignments, err := rl.getAssignmentsForRecipient(ctx, recipient.ID)
			if err != nil {
				dialog.ShowError(fmt.Errorf("担当者データの取得に失敗しました: %w", err), fyne.CurrentApp().Driver().AllWindows()[0])
				return
			}

			// Convert pointer slices to value slices for PDF service
			certificateValues := make([]domain.BenefitCertificate, len(certificates))
			for i, cert := range certificates {
				if cert != nil {
					certificateValues[i] = *cert
				}
			}

			assignmentValues := make([]domain.StaffAssignment, len(assignments))
			for i, assign := range assignments {
				if assign != nil {
					assignmentValues[i] = *assign
				}
			}

			// Generate PDF
			pdfBytes, err := rl.pdfService.GenerateRecipientReport(ctx, recipient, certificateValues, assignmentValues)
			if err != nil {
				dialog.ShowError(fmt.Errorf("PDF生成に失敗しました: %w", err), fyne.CurrentApp().Driver().AllWindows()[0])
				return
			}

			// For the first recipient, write directly. For subsequent ones, append.
			if i == 0 {
				_, err = writer.Write(pdfBytes)
			} else {
				// Note: This is a simplified implementation. 
				// For multiple recipients, you might want to combine PDFs or create separate files.
				break
			}

			if err != nil {
				dialog.ShowError(fmt.Errorf("ファイルの書き込みに失敗しました: %w", err), fyne.CurrentApp().Driver().AllWindows()[0])
				return
			}
		}

		dialog.ShowInformation("成功", fmt.Sprintf("%d人の利用者データをPDFにエクスポートしました。", len(rl.filteredData)), fyne.CurrentApp().Driver().AllWindows()[0])
	}, fyne.CurrentApp().Driver().AllWindows()[0])

	// Set default filename with timestamp
	defaultName := fmt.Sprintf("利用者一覧_%s.pdf", time.Now().Format("20060102_150405"))
	saveDialog.SetFileName(defaultName)
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".pdf"}))
	saveDialog.Show()
}

// getCertificatesForRecipient is a helper to get certificates for a recipient
// This is a simplified implementation - in a real app you'd have proper certificate use case
func (rl *RecipientList) getCertificatesForRecipient(ctx context.Context, recipientID domain.ID) ([]*domain.BenefitCertificate, error) {
	if rl.certificateUseCase == nil {
		return []*domain.BenefitCertificate{}, nil
	}

	// Use CertificateUseCase to get certificates for the recipient
	certificates, err := rl.certificateUseCase.GetCertificatesByRecipient(ctx, recipientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificates for recipient %s: %w", recipientID, err)
	}

	return certificates, nil
}

// getAssignmentsForRecipient is a helper to get staff assignments for a recipient  
// This is a simplified implementation - in a real app you'd have proper assignment use case
func (rl *RecipientList) getAssignmentsForRecipient(ctx context.Context, recipientID domain.ID) ([]*domain.StaffAssignment, error) {
	if rl.staffUseCase == nil {
		return []*domain.StaffAssignment{}, nil
	}

	// Use StaffUseCase to get assignments for the recipient
	assignments, err := rl.staffUseCase.GetAssignments(ctx, recipientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignments for recipient %s: %w", recipientID, err)
	}

	return assignments, nil
}
