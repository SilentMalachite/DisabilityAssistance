package widgets

import (
	"context"
	"fmt"
	"time"

	"shien-system/internal/adapter/pdf"
	"shien-system/internal/domain"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// AuditLogList represents the audit log list widget
type AuditLogList struct {
	auditRepo domain.AuditLogRepository
	staffRepo domain.StaffRepository
	pdfService *pdf.PDFService

	// UI components
	table         *widget.Table
	refreshButton *widget.Button
	exportButton  *widget.Button
	actionFilter  *widget.Select
	dateFromEntry *widget.Entry
	dateToEntry   *widget.Entry
	actorFilter   *widget.Select

	// Data
	auditLogs       []*domain.AuditLog
	filteredData    []*domain.AuditLog
	staffMap        map[domain.ID]*domain.Staff // For staff name lookup
	currentAction   string
	currentActorID  string
	currentDateFrom time.Time
	currentDateTo   time.Time
}

// NewAuditLogList creates a new AuditLogList widget
func NewAuditLogList(auditRepo domain.AuditLogRepository, staffRepo domain.StaffRepository, pdfService *pdf.PDFService) *AuditLogList {
	al := &AuditLogList{
		auditRepo:    auditRepo,
		staffRepo:    staffRepo,
		pdfService:   pdfService,
		auditLogs:    make([]*domain.AuditLog, 0),
		filteredData: make([]*domain.AuditLog, 0),
		staffMap:     make(map[domain.ID]*domain.Staff),
	}

	al.createWidgets()
	al.setupTable()
	al.setupEventHandlers()

	return al
}

// createWidgets initializes all UI components
func (al *AuditLogList) createWidgets() {
	// Table
	al.table = widget.NewTable(
		func() (int, int) {
			return len(al.filteredData), 6 // 6 columns
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			al.updateTableCell(id, obj.(*widget.Label))
		},
	)

	// Buttons
	al.refreshButton = widget.NewButton("更新", func() {
		al.LoadData()
	})

	al.exportButton = widget.NewButton("PDFエクスポート", func() {
		al.exportToPDF()
	})

	// Filters
	al.actionFilter = widget.NewSelect(
		[]string{"全て", "LOGIN_SUCCESS", "LOGIN_FAILED", "LOGOUT", "CREATE_RECIPIENT", "UPDATE_RECIPIENT", "DELETE_RECIPIENT"},
		func(selected string) {
			al.onActionFilterChanged(selected)
		},
	)
	al.actionFilter.Selected = "全て"

	al.actorFilter = widget.NewSelect([]string{"全職員"}, func(selected string) {
		al.onActorFilterChanged(selected)
	})
	al.actorFilter.Selected = "全職員"

	// Date filters
	al.dateFromEntry = widget.NewEntry()
	al.dateFromEntry.SetPlaceHolder("開始日 (YYYY/MM/DD)")

	al.dateToEntry = widget.NewEntry()
	al.dateToEntry.SetPlaceHolder("終了日 (YYYY/MM/DD)")
}

// setupTable configures the table widget
func (al *AuditLogList) setupTable() {
	// Set column widths
	al.table.SetColumnWidth(0, 150) // 日時
	al.table.SetColumnWidth(1, 100) // 操作者
	al.table.SetColumnWidth(2, 120) // アクション
	al.table.SetColumnWidth(3, 100) // 対象
	al.table.SetColumnWidth(4, 100) // IPアドレス
	al.table.SetColumnWidth(5, 200) // 詳細
}

// setupEventHandlers configures event handlers
func (al *AuditLogList) setupEventHandlers() {
	// Date validation
	al.dateFromEntry.OnChanged = func(text string) {
		if text != "" {
			al.validateAndSetDate(text, true)
		}
	}

	al.dateToEntry.OnChanged = func(text string) {
		if text != "" {
			al.validateAndSetDate(text, false)
		}
	}
}

// validateAndSetDate validates and sets date filter
func (al *AuditLogList) validateAndSetDate(dateStr string, isFrom bool) {
	parsed, err := time.Parse("2006/01/02", dateStr)
	if err != nil {
		// Show validation error
		fmt.Printf("Invalid date format: %s\n", dateStr)
		return
	}

	if isFrom {
		al.currentDateFrom = parsed
	} else {
		al.currentDateTo = parsed
	}

	// Auto-refresh when date changes
	al.applyFilters()
	al.table.Refresh()
}

// updateTableCell updates a specific table cell with audit log data
func (al *AuditLogList) updateTableCell(id widget.TableCellID, label *widget.Label) {
	if id.Row >= len(al.filteredData) {
		label.SetText("")
		return
	}

	log := al.filteredData[id.Row]

	switch id.Col {
	case 0: // 日時
		label.SetText(log.At.Format("2006/01/02 15:04:05"))
	case 1: // 操作者
		if staff, ok := al.staffMap[log.ActorID]; ok {
			label.SetText(staff.Name)
		} else {
			label.SetText(string(log.ActorID))
		}
	case 2: // アクション
		label.SetText(al.formatAction(log.Action))
	case 3: // 対象
		label.SetText(log.Target)
	case 4: // IPアドレス
		label.SetText(log.IP)
	case 5: // 詳細
		label.SetText(log.Details)
	default:
		label.SetText("")
	}
}

// formatAction formats action codes to Japanese
func (al *AuditLogList) formatAction(action string) string {
	switch action {
	case "LOGIN_SUCCESS":
		return "ログイン成功"
	case "LOGIN_FAILED":
		return "ログイン失敗"
	case "LOGOUT":
		return "ログアウト"
	case "CREATE_RECIPIENT":
		return "利用者作成"
	case "UPDATE_RECIPIENT":
		return "利用者更新"
	case "DELETE_RECIPIENT":
		return "利用者削除"
	case "CREATE_CERTIFICATE":
		return "受給者証作成"
	case "UPDATE_CERTIFICATE":
		return "受給者証更新"
	case "DELETE_CERTIFICATE":
		return "受給者証削除"
	default:
		return action
	}
}

// LoadData loads audit log data from the repository
func (al *AuditLogList) LoadData() error {
	ctx := context.Background()

	// Load recent audit logs (simplified)
	auditLogs, err := al.auditRepo.List(ctx, 1000, 0) // Get last 1000 entries
	if err != nil {
		return fmt.Errorf("failed to load audit logs: %w", err)
	}

	al.auditLogs = auditLogs

	// Load staff information for display
	if err := al.loadStaffData(ctx); err != nil {
		return fmt.Errorf("failed to load staff data: %w", err)
	}

	al.applyFilters()
	al.table.Refresh()

	return nil
}

// loadStaffData loads staff information for display purposes
func (al *AuditLogList) loadStaffData(ctx context.Context) error {
	// Extract unique actor IDs
	actorIDs := make(map[domain.ID]bool)
	for _, log := range al.auditLogs {
		actorIDs[log.ActorID] = true
	}

	// Load staff information
	// Note: This is inefficient but works for demonstration
	// In production, implement a batch get method
	for actorID := range actorIDs {
		staff, err := al.staffRepo.GetByID(ctx, actorID)
		if err != nil {
			// Log error but continue with other staff
			fmt.Printf("Failed to load staff %s: %v\n", actorID, err)
			continue
		}
		al.staffMap[actorID] = staff
	}

	return nil
}

// onActionFilterChanged handles action filter changes
func (al *AuditLogList) onActionFilterChanged(action string) {
	if action == "全て" {
		al.currentAction = ""
	} else {
		al.currentAction = action
	}

	al.applyFilters()
	al.table.Refresh()
}

// onActorFilterChanged handles actor filter changes
func (al *AuditLogList) onActorFilterChanged(actorName string) {
	if actorName == "全職員" {
		al.currentActorID = ""
	} else {
		// In production, you'd maintain a name-to-ID mapping
		al.currentActorID = actorName // placeholder
	}

	al.applyFilters()
	al.table.Refresh()
}

// applyFilters applies local filters to the audit log data
func (al *AuditLogList) applyFilters() {
	al.filteredData = make([]*domain.AuditLog, 0)

	for _, log := range al.auditLogs {
		// Apply action filter
		if al.currentAction != "" && log.Action != al.currentAction {
			continue
		}

		// Apply actor filter
		if al.currentActorID != "" && string(log.ActorID) != al.currentActorID {
			continue
		}

		// Apply date filters
		if !al.currentDateFrom.IsZero() && log.At.Before(al.currentDateFrom) {
			continue
		}

		if !al.currentDateTo.IsZero() && log.At.After(al.currentDateTo.Add(24*time.Hour)) {
			continue
		}

		al.filteredData = append(al.filteredData, log)
	}
}

// CreateObject creates the main UI object for this widget
func (al *AuditLogList) CreateObject() fyne.CanvasObject {
	// Filter controls
	filterControls := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("アクション:"),
			al.actionFilter,
			widget.NewLabel("操作者:"),
			al.actorFilter,
			al.refreshButton,
			al.exportButton,
		),
		container.NewHBox(
			widget.NewLabel("期間:"),
			al.dateFromEntry,
			widget.NewLabel("〜"),
			al.dateToEntry,
		),
	)

	// Table with headers
	tableContainer := container.NewBorder(
		al.createTableHeader(),
		nil, nil, nil,
		al.table,
	)

	// Complete layout
	return container.NewBorder(
		filterControls,
		nil, nil, nil,
		tableContainer,
	)
}

// createTableHeader creates the table header
func (al *AuditLogList) createTableHeader() fyne.CanvasObject {
	headers := []string{"日時", "操作者", "アクション", "対象", "IPアドレス", "詳細"}
	headerWidgets := make([]fyne.CanvasObject, len(headers))

	for i, header := range headers {
		label := widget.NewLabel(header)
		label.TextStyle.Bold = true
		headerWidgets[i] = label
	}

	return container.NewHBox(headerWidgets...)
}

// Length returns the number of visible items in the table (for testing)
func (al *AuditLogList) Length() int {
	return len(al.filteredData)
}

// exportToPDF exports the filtered audit logs to a PDF report
func (al *AuditLogList) exportToPDF() {
	if al.pdfService == nil {
		dialog.ShowError(fmt.Errorf("PDFサービスが利用できません"), fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	if len(al.filteredData) == 0 {
		dialog.ShowInformation("情報", "エクスポートする監査ログデータがありません。", fyne.CurrentApp().Driver().AllWindows()[0])
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
		auditLogValues := make([]domain.AuditLog, len(al.filteredData))
		for i, log := range al.filteredData {
			if log != nil {
				auditLogValues[i] = *log
			}
		}

		// Determine date range for the report
		var startDate, endDate time.Time
		if len(auditLogValues) > 0 {
			startDate = auditLogValues[len(auditLogValues)-1].At // Oldest (assuming reverse chronological order)
			endDate = auditLogValues[0].At                        // Newest
		}

		// If custom date filters are applied, use those instead
		if !al.currentDateFrom.IsZero() {
			startDate = al.currentDateFrom
		}
		if !al.currentDateTo.IsZero() {
			endDate = al.currentDateTo
		}

		// If still no dates, use a default range
		if startDate.IsZero() || endDate.IsZero() {
			endDate = time.Now()
			startDate = endDate.AddDate(0, -1, 0) // Last month
		}

		// Generate PDF
		ctx := context.Background()
		pdfBytes, err := al.pdfService.GenerateAuditReport(ctx, auditLogValues, startDate, endDate)
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

		dialog.ShowInformation("成功", fmt.Sprintf("%d件の監査ログデータをPDFにエクスポートしました。", len(al.filteredData)), fyne.CurrentApp().Driver().AllWindows()[0])
	}, fyne.CurrentApp().Driver().AllWindows()[0])

	// Set default filename with timestamp
	defaultName := fmt.Sprintf("監査ログ_%s.pdf", time.Now().Format("20060102_150405"))
	saveDialog.SetFileName(defaultName)
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".pdf"}))
	saveDialog.Show()
}
