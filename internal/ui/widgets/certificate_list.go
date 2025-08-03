package widgets

import (
	"context"
	"fmt"
	"time"

	"shien-system/internal/adapter/pdf"
	"shien-system/internal/domain"
	"shien-system/internal/usecase"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// CertificateList represents the benefit certificate list widget
type CertificateList struct {
	certificateUseCase usecase.CertificateUseCase
	recipientUseCase   usecase.RecipientUseCase
	pdfService         *pdf.PDFService

	// UI components
	table           *widget.Table
	newButton       *widget.Button
	refreshButton   *widget.Button
	exportButton    *widget.Button
	recipientFilter *widget.Select
	statusFilter    *widget.Select

	// Data
	certificates       []*domain.BenefitCertificate
	filteredData       []*domain.BenefitCertificate
	recipientMap       map[domain.ID]*domain.Recipient // For recipient name lookup
	currentRecipientID string
	currentStatus      string

	// Event handlers
	onNewCertificate  func()
	onEditCertificate func(certificateID domain.ID)
}

// CertificateForm.go - Created to handle benefit certificate registration and editing

// NewCertificateList creates a new CertificateList widget
func NewCertificateList(certificateUseCase usecase.CertificateUseCase, recipientUseCase usecase.RecipientUseCase, pdfService *pdf.PDFService) *CertificateList {
	cl := &CertificateList{
		certificateUseCase: certificateUseCase,
		recipientUseCase:   recipientUseCase,
		pdfService:         pdfService,
		certificates:       make([]*domain.BenefitCertificate, 0),
		filteredData:       make([]*domain.BenefitCertificate, 0),
		recipientMap:       make(map[domain.ID]*domain.Recipient),
	}

	cl.createWidgets()
	cl.setupTable()
	cl.setupEventHandlers()

	return cl
}

// createWidgets initializes all UI components
func (cl *CertificateList) createWidgets() {
	// Table
	cl.table = widget.NewTable(
		func() (int, int) {
			return len(cl.filteredData), 7 // 7 columns
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			cl.updateTableCell(id, obj.(*widget.Label))
		},
	)

	// Buttons
	cl.newButton = widget.NewButton("新規登録", func() {
		if cl.onNewCertificate != nil {
			cl.onNewCertificate()
		}
	})

	cl.refreshButton = widget.NewButton("更新", func() {
		cl.LoadData()
	})

	cl.exportButton = widget.NewButton("PDFエクスポート", func() {
		cl.exportToPDF()
	})

	// Filters
	cl.recipientFilter = widget.NewSelect([]string{"全利用者"}, func(selected string) {
		cl.onRecipientFilterChanged(selected)
	})
	cl.recipientFilter.Selected = "全利用者"

	cl.statusFilter = widget.NewSelect(
		[]string{"全て", "有効", "期限切れ", "期限間近"},
		func(selected string) {
			cl.onStatusFilterChanged(selected)
		},
	)
	cl.statusFilter.Selected = "全て"
}

// setupTable configures the table widget
func (cl *CertificateList) setupTable() {
	// Set column widths
	cl.table.SetColumnWidth(0, 120) // 利用者名
	cl.table.SetColumnWidth(1, 100) // サービス種別
	cl.table.SetColumnWidth(2, 100) // 開始日
	cl.table.SetColumnWidth(3, 100) // 終了日
	cl.table.SetColumnWidth(4, 80)  // 支給日数
	cl.table.SetColumnWidth(5, 100) // 発行者
	cl.table.SetColumnWidth(6, 80)  // 状態

	// Set up selection handler for double-click navigation
	cl.table.OnSelected = func(id widget.TableCellID) {
		if id.Row < len(cl.filteredData) && cl.onEditCertificate != nil {
			certificate := cl.filteredData[id.Row]
			cl.onEditCertificate(certificate.ID)
		}
	}
}

// setupEventHandlers configures event handlers
func (cl *CertificateList) setupEventHandlers() {
	// No specific event handlers needed for this widget currently
}

// updateTableCell updates a specific table cell with certificate data
func (cl *CertificateList) updateTableCell(id widget.TableCellID, label *widget.Label) {
	if id.Row >= len(cl.filteredData) {
		label.SetText("")
		return
	}

	certificate := cl.filteredData[id.Row]

	switch id.Col {
	case 0: // 利用者名
		if recipient, ok := cl.recipientMap[certificate.RecipientID]; ok {
			label.SetText(recipient.Name)
		} else {
			label.SetText("未設定")
		}
	case 1: // サービス種別
		label.SetText(certificate.ServiceType)
	case 2: // 開始日
		label.SetText(certificate.StartDate.Format("2006/01/02"))
	case 3: // 終了日
		label.SetText(certificate.EndDate.Format("2006/01/02"))
	case 4: // 支給日数
		label.SetText(fmt.Sprintf("%d日", certificate.MaxBenefitDaysPerMonth))
	case 5: // 発行者
		label.SetText(certificate.Issuer)
	case 6: // 状態
		status := cl.calculateStatus(certificate)
		label.SetText(status)

		// Color coding for status
		switch status {
		case "期限切れ":
			label.TextStyle.Bold = true
			// Note: Fyne doesn't support text color easily, we'd need custom rendering
		case "期限間近":
			label.TextStyle.Italic = true
		}
	default:
		label.SetText("")
	}
}

// calculateStatus calculates the certificate status based on dates
func (cl *CertificateList) calculateStatus(cert *domain.BenefitCertificate) string {
	now := time.Now()

	if cert.EndDate.Before(now) {
		return "期限切れ"
	}

	// Check if expiring within 30 days
	daysUntilExpiry := cert.EndDate.Sub(now).Hours() / 24
	if daysUntilExpiry <= 30 {
		return "期限間近"
	}

	return "有効"
}

// LoadData loads certificate data from the use case
func (cl *CertificateList) LoadData() error {
	ctx := context.Background()

	// Since ListCertificates doesn't exist, we'll get expiring certificates as a starting point
	// In production, you would want to implement a proper list method in the use case
	certificates, err := cl.certificateUseCase.GetExpiringSoon(ctx, 365) // Get all certificates expiring within 1 year
	if err != nil {
		return fmt.Errorf("failed to load certificates: %w", err)
	}

	cl.certificates = certificates

	// Load recipient information for display
	if err := cl.loadRecipientData(ctx); err != nil {
		return fmt.Errorf("failed to load recipient data: %w", err)
	}

	cl.applyFilters()
	cl.table.Refresh()

	return nil
}

// loadRecipientData loads recipient information for display purposes
func (cl *CertificateList) loadRecipientData(ctx context.Context) error {
	// Extract unique recipient IDs
	recipientIDs := make(map[domain.ID]bool)
	for _, cert := range cl.certificates {
		recipientIDs[cert.RecipientID] = true
	}

	// Load recipient information (this would require a batch get method in the use case)
	// For now, we'll simulate this - in production, you'd want to optimize this
	for recipientID := range recipientIDs {
		// Note: This is inefficient, but works for demonstration
		// In production, implement a batch get method
		recipient, err := cl.recipientUseCase.GetRecipient(ctx, recipientID)
		if err != nil {
			// Log error but continue with other recipients
			fmt.Printf("Failed to load recipient %s: %v\n", recipientID, err)
			continue
		}
		cl.recipientMap[recipientID] = recipient
	}

	return nil
}

// onRecipientFilterChanged handles recipient filter changes
func (cl *CertificateList) onRecipientFilterChanged(recipientName string) {
	// Convert recipient name to ID (simplified for testing)
	if recipientName == "全利用者" {
		cl.currentRecipientID = ""
	} else {
		// In production, you'd maintain a name-to-ID mapping
		cl.currentRecipientID = recipientName // placeholder
	}

	cl.LoadData() // Reload with new filter
}

// onStatusFilterChanged handles status filter changes
func (cl *CertificateList) onStatusFilterChanged(status string) {
	cl.currentStatus = status
	cl.applyFilters()
	cl.table.Refresh()
}

// getRecipientIDFilter returns the current recipient ID filter
func (cl *CertificateList) getRecipientIDFilter() *domain.ID {
	if cl.currentRecipientID == "" {
		return nil
	}
	return &cl.currentRecipientID
}

// applyFilters applies local filters to the certificate data
func (cl *CertificateList) applyFilters() {
	cl.filteredData = make([]*domain.BenefitCertificate, 0)

	for _, certificate := range cl.certificates {
		// Apply status filter
		if cl.currentStatus != "" && cl.currentStatus != "全て" {
			status := cl.calculateStatus(certificate)
			if status != cl.currentStatus {
				continue
			}
		}

		cl.filteredData = append(cl.filteredData, certificate)
	}
}

// CreateObject creates the main UI object for this widget
func (cl *CertificateList) CreateObject() fyne.CanvasObject {
	// Header with filters and controls
	header := container.NewBorder(
		nil, nil,
		container.NewHBox(
			widget.NewLabel("利用者:"),
			cl.recipientFilter,
			widget.NewLabel("状態:"),
			cl.statusFilter,
		),
		container.NewHBox(
			cl.newButton,
			cl.refreshButton,
			cl.exportButton,
		),
		nil,
	)

	// Table with headers
	tableContainer := container.NewBorder(
		cl.createTableHeader(),
		nil, nil, nil,
		cl.table,
	)

	// Complete layout
	return container.NewBorder(
		header,
		nil, nil, nil,
		tableContainer,
	)
}

// createTableHeader creates the table header
func (cl *CertificateList) createTableHeader() fyne.CanvasObject {
	headers := []string{"利用者名", "サービス種別", "開始日", "終了日", "支給日数", "発行者", "状態"}
	headerWidgets := make([]fyne.CanvasObject, len(headers))

	for i, header := range headers {
		label := widget.NewLabel(header)
		label.TextStyle.Bold = true
		headerWidgets[i] = label
	}

	return container.NewHBox(headerWidgets...)
}

// SetOnNewCertificate sets the callback for new certificate action
func (cl *CertificateList) SetOnNewCertificate(callback func()) {
	cl.onNewCertificate = callback
}

// SetOnEditCertificate sets the callback for edit certificate action
func (cl *CertificateList) SetOnEditCertificate(callback func(certificateID domain.ID)) {
	cl.onEditCertificate = callback
}

// Length returns the number of visible items in the table (for testing)
func (cl *CertificateList) Length() int {
	return len(cl.filteredData)
}

// exportToPDF exports the filtered certificate list to a PDF report
func (cl *CertificateList) exportToPDF() {
	if cl.pdfService == nil {
		dialog.ShowError(fmt.Errorf("PDFサービスが利用できません"), fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	if len(cl.filteredData) == 0 {
		dialog.ShowInformation("情報", "エクスポートする受給者証データがありません。", fyne.CurrentApp().Driver().AllWindows()[0])
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

		// Convert pointer slice to value slice for PDF service
		certificateValues := make([]domain.BenefitCertificate, len(cl.filteredData))
		for i, cert := range cl.filteredData {
			if cert != nil {
				certificateValues[i] = *cert
			}
		}

		// Generate PDF for certificate list
		ctx := context.Background()
		pdfBytes, err := cl.pdfService.GenerateCertificateReport(ctx, certificateValues, cl.recipientMap)
		if err != nil {
			dialog.ShowError(fmt.Errorf("PDF生成に失敗しました: %w", err), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		// Write PDF to file
		_, err = writer.Write(pdfBytes)
		if err != nil {
			dialog.ShowError(fmt.Errorf("ファイルの書き込みに失敗しました: %w", err), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		dialog.ShowInformation("成功", fmt.Sprintf("%d件の受給者証データをPDFにエクスポートしました。", len(cl.filteredData)), fyne.CurrentApp().Driver().AllWindows()[0])
	}, fyne.CurrentApp().Driver().AllWindows()[0])

	// Set default filename with timestamp
	defaultName := fmt.Sprintf("受給者証一覧_%s.pdf", time.Now().Format("20060102_150405"))
	saveDialog.SetFileName(defaultName)
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".pdf"}))
	saveDialog.Show()
}
