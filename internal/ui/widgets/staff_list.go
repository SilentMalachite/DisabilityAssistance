package widgets

import (
	"context"
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"shien-system/internal/adapter/pdf"
	"shien-system/internal/domain"
	"shien-system/internal/usecase"
)

// StaffList provides a widget for managing staff members
type StaffList struct {
	useCase    usecase.StaffUseCase
	pdfService *pdf.PDFService

	// UI components
	searchEntry   *widget.Entry
	table         *widget.Table
	newButton     *widget.Button
	refreshButton *widget.Button
	exportButton  *widget.Button
	roleFilter    *widget.Select

	// Data
	staff         []*domain.Staff
	filteredData  []*domain.Staff
	currentSearch string
	currentRole   string

	// Event handlers
	onNewStaff  func()
	onEditStaff func(staffID domain.ID)
}

// NewStaffList creates a new staff list widget
func NewStaffList(useCase usecase.StaffUseCase, pdfService *pdf.PDFService) *StaffList {
	s := &StaffList{
		useCase:    useCase,
		pdfService: pdfService,
	}

	s.createWidgets()
	s.setupTable()
	s.setupEventHandlers()

	return s
}

// createWidgets initializes the UI components
func (s *StaffList) createWidgets() {
	// Search entry
	s.searchEntry = widget.NewEntry()
	s.searchEntry.SetPlaceHolder("職員名で検索...")

	// Role filter
	s.roleFilter = widget.NewSelect([]string{"全て", "管理者", "職員", "閲覧のみ"}, nil)
	s.roleFilter.SetSelected("全て")

	// Buttons
	s.newButton = widget.NewButton("新規職員登録", nil)
	s.refreshButton = widget.NewButton("更新", nil)
	s.exportButton = widget.NewButton("PDFエクスポート", func() {
		s.exportToPDF()
	})

	// Table
	s.table = widget.NewTable(
		func() (int, int) { return s.Length(), 6 }, // 6 columns
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			s.updateTableCell(id, cell)
		},
	)
}

// setupTable configures the table headers and properties
func (s *StaffList) setupTable() {
	s.table.SetColumnWidth(0, 100) // ID
	s.table.SetColumnWidth(1, 150) // 名前
	s.table.SetColumnWidth(2, 100) // ロール
	s.table.SetColumnWidth(3, 120) // 作成日
	s.table.SetColumnWidth(4, 120) // 更新日
	s.table.SetColumnWidth(5, 80)  // 状態
}

// setupEventHandlers configures event handlers
func (s *StaffList) setupEventHandlers() {
	s.searchEntry.OnChanged = s.onSearchChanged
	s.roleFilter.OnChanged = s.onRoleFilterChanged
	s.refreshButton.OnTapped = func() {
		s.LoadData()
	}

	s.table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 && id.Row-1 < len(s.filteredData) {
			staff := s.filteredData[id.Row-1]
			if s.onEditStaff != nil {
				s.onEditStaff(staff.ID)
			}
		}
	}
}

// updateTableCell updates table cell content
func (s *StaffList) updateTableCell(id widget.TableCellID, cell fyne.CanvasObject) {
	label := cell.(*widget.Label)

	// Header row
	if id.Row == 0 {
		s.createTableHeader(id.Col, label)
		return
	}

	// Data rows
	rowIndex := id.Row - 1
	if rowIndex >= len(s.filteredData) {
		label.SetText("")
		return
	}

	staff := s.filteredData[rowIndex]

	switch id.Col {
	case 0: // ID
		label.SetText(staff.ID[:8]) // 短縮表示
	case 1: // 名前
		label.SetText(staff.Name)
	case 2: // ロール
		label.SetText(s.formatRole(staff.Role))
	case 3: // 作成日
		label.SetText(staff.CreatedAt.Format("2006-01-02"))
	case 4: // 更新日
		label.SetText(staff.UpdatedAt.Format("2006-01-02"))
	case 5: // 状態
		label.SetText("アクティブ") // TODO: 状態管理を実装
	}
}

// formatRole formats staff role for display
func (s *StaffList) formatRole(role domain.StaffRole) string {
	switch role {
	case domain.RoleAdmin:
		return "管理者"
	case domain.RoleStaff:
		return "職員"
	case domain.RoleReadOnly:
		return "閲覧のみ"
	default:
		return string(role)
	}
}

// LoadData loads staff data from the use case
func (s *StaffList) LoadData() {
	req := usecase.ListStaffRequest{
		Limit:  1000, // TODO: ページング実装
		Offset: 0,
	}

	result, err := s.useCase.ListStaff(context.Background(), req)
	if err != nil {
		// TODO: エラーハンドリング
		s.staff = []*domain.Staff{}
	} else {
		s.staff = result.Staff
	}

	s.applyFilters()
	s.table.Refresh()
}

// onSearchChanged handles search input changes
func (s *StaffList) onSearchChanged(text string) {
	s.currentSearch = text
	s.applyFilters()
}

// onRoleFilterChanged handles role filter changes
func (s *StaffList) onRoleFilterChanged(selected string) {
	s.currentRole = selected
	s.applyFilters()
}

// applyFilters applies search and role filters to the data
func (s *StaffList) applyFilters() {
	s.filteredData = []*domain.Staff{}

	for _, staff := range s.staff {
		// Apply search filter
		if s.currentSearch != "" && !s.matchesSearch(staff) {
			continue
		}

		// Apply role filter
		if s.currentRole != "" && s.currentRole != "全て" && !s.matchesRole(staff) {
			continue
		}

		s.filteredData = append(s.filteredData, staff)
	}

	s.table.Refresh()
}

// matchesSearch checks if staff matches search criteria
func (s *StaffList) matchesSearch(staff *domain.Staff) bool {
	searchLower := strings.ToLower(s.currentSearch)
	return strings.Contains(strings.ToLower(staff.Name), searchLower)
}

// matchesRole checks if staff matches role filter
func (s *StaffList) matchesRole(staff *domain.Staff) bool {
	switch s.currentRole {
	case "管理者":
		return staff.Role == domain.RoleAdmin
	case "職員":
		return staff.Role == domain.RoleStaff
	case "閲覧のみ":
		return staff.Role == domain.RoleReadOnly
	default:
		return true
	}
}

// CreateObject creates the main container for the staff list
func (s *StaffList) CreateObject() fyne.CanvasObject {
	// Top controls
	searchContainer := container.NewHBox(
		widget.NewLabel("検索:"),
		s.searchEntry,
		widget.NewLabel("ロール:"),
		s.roleFilter,
		s.refreshButton,
		s.exportButton,
	)

	// Button bar
	buttonContainer := container.NewHBox(
		s.newButton,
	)

	// Main layout
	return container.NewBorder(
		container.NewVBox(searchContainer, buttonContainer),
		nil,
		nil,
		nil,
		s.table,
	)
}

// createTableHeader creates table header labels
func (s *StaffList) createTableHeader(col int, label *widget.Label) {
	headers := []string{"ID", "名前", "ロール", "作成日", "更新日", "状態"}
	if col < len(headers) {
		label.SetText(headers[col])
	}
}

// SetOnNewStaff sets the callback for new staff button
func (s *StaffList) SetOnNewStaff(callback func()) {
	s.onNewStaff = callback
	if s.newButton != nil {
		s.newButton.OnTapped = callback
	}
}

// SetOnEditStaff sets the callback for editing staff
func (s *StaffList) SetOnEditStaff(callback func(staffID domain.ID)) {
	s.onEditStaff = callback
}

// Length returns the number of filtered staff entries
func (s *StaffList) Length() int {
	return len(s.filteredData) + 1 // +1 for header
}

// exportToPDF exports the filtered staff list to a PDF report
func (s *StaffList) exportToPDF() {
	if s.pdfService == nil {
		dialog.ShowError(fmt.Errorf("PDFサービスが利用できません"), fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	if len(s.filteredData) == 0 {
		dialog.ShowInformation("情報", "エクスポートする職員データがありません。", fyne.CurrentApp().Driver().AllWindows()[0])
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
		staffValues := make([]domain.Staff, len(s.filteredData))
		for i, staff := range s.filteredData {
			if staff != nil {
				staffValues[i] = *staff
			}
		}

		// Generate PDF for staff list
		ctx := context.Background()
		pdfBytes, err := s.pdfService.GenerateStaffReport(ctx, staffValues)
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

		dialog.ShowInformation("成功", fmt.Sprintf("%d人の職員データをPDFにエクスポートしました。", len(s.filteredData)), fyne.CurrentApp().Driver().AllWindows()[0])
	}, fyne.CurrentApp().Driver().AllWindows()[0])

	// Set default filename with timestamp
	defaultName := fmt.Sprintf("職員一覧_%s.pdf", time.Now().Format("20060102_150405"))
	saveDialog.SetFileName(defaultName)
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".pdf"}))
	saveDialog.Show()
}
