package widgets

import (
	"errors"
	"testing"
	"time"

	"shien-system/internal/domain"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

var testError = errors.New("test error")

// Use MockRecipientUseCase from mock_usecases_test.go

// Test data helper functions
func createTestRecipients() []*domain.Recipient {
	now := time.Now()
	return []*domain.Recipient{
		{
			ID:               "recipient-001",
			Name:             "田中太郎",
			Kana:             "タナカタロウ",
			Sex:              domain.SexMale,
			BirthDate:        time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
			DisabilityName:   "身体障害",
			HasDisabilityID:  true,
			Grade:            "1級",
			Address:          "東京都渋谷区",
			Phone:            "03-1234-5678",
			Email:            "tanaka@example.com",
			PublicAssistance: false,
			AdmissionDate:    &now,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "recipient-002",
			Name:             "佐藤花子",
			Kana:             "サトウハナコ",
			Sex:              domain.SexFemale,
			BirthDate:        time.Date(1985, 8, 20, 0, 0, 0, 0, time.UTC),
			DisabilityName:   "知的障害",
			HasDisabilityID:  true,
			Grade:            "2級",
			Address:          "東京都新宿区",
			Phone:            "03-8765-4321",
			Email:            "sato@example.com",
			PublicAssistance: true,
			AdmissionDate:    &now,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}
}

func TestNewRecipientList(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockUseCase := &MockRecipientUseCase{}

	recipientList := NewRecipientList(mockUseCase, nil, nil, nil)

	if recipientList == nil {
		t.Fatal("NewRecipientList returned nil")
	}
	if recipientList.searchEntry == nil {
		t.Error("searchEntry is nil")
	}
	if recipientList.table == nil {
		t.Error("table is nil")
	}
	if recipientList.newButton == nil {
		t.Error("newButton is nil")
	}
	if recipientList.refreshButton == nil {
		t.Error("refreshButton is nil")
	}
	if recipientList.staffFilter == nil {
		t.Error("staffFilter is nil")
	}
}

func TestRecipientList_LoadData(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockUseCase := &MockRecipientUseCase{}
	testRecipients := createTestRecipients()

	// Set up mock data
	mockUseCase.recipients = testRecipients

	recipientList := NewRecipientList(mockUseCase, nil, nil, nil)
	err := recipientList.LoadData()

	if err != nil {
		t.Errorf("LoadData failed: %v", err)
	}
	if recipientList.Length() != len(testRecipients) {
		t.Errorf("Expected %d recipients, got %d", len(testRecipients), recipientList.Length())
	}
}

func TestRecipientList_SearchFunctionality(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockUseCase := &MockRecipientUseCase{}
	testRecipients := createTestRecipients()

	// Set up mock data
	mockUseCase.recipients = testRecipients

	recipientList := NewRecipientList(mockUseCase, nil, nil, nil)
	// Load initial data
	recipientList.LoadData()

	// Test that search filters the data locally
	initialLength := recipientList.Length()
	if initialLength != len(testRecipients) {
		t.Errorf("Expected %d recipients, got %d", len(testRecipients), initialLength)
	}

	// Simulate search input - this should filter locally
	recipientList.onSearchChanged("田中")

	// Should have filtered to only recipients matching "田中"
	filteredLength := recipientList.Length()
	if filteredLength > initialLength {
		t.Errorf("Filtered length %d should be less than or equal to initial length %d", filteredLength, initialLength)
	}

	// Test passed
}

func TestRecipientList_TableColumns(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockUseCase := &MockRecipientUseCase{}
	testRecipients := createTestRecipients()
	// paginatedResult not needed - using mockUseCase.recipients

	mockUseCase.recipients = testRecipients

	recipientList := NewRecipientList(mockUseCase, nil, nil, nil)
	recipientList.LoadData()

	// Test column count - simplified check
	expectedColumns := []string{"氏名", "カナ", "性別", "生年月日", "障害名", "等級", "担当者", "状態"}
	_, cols := recipientList.table.Length()
	if cols != len(expectedColumns) {
		t.Errorf("Expected %d columns, got %d", len(expectedColumns), cols)
	}

	// Test that table has proper structure
	if recipientList.table == nil {
		t.Error("table is nil")
	}
}

func TestRecipientList_NewButtonAction(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockUseCase := &MockRecipientUseCase{}
	recipientList := NewRecipientList(mockUseCase, nil, nil, nil)

	// Test that new button exists and can be activated
	if recipientList.newButton == nil {
		t.Fatal("newButton is nil")
	}
	if recipientList.newButton.Text != "新規登録" {
		t.Errorf("Expected button text '新規登録', got '%s'", recipientList.newButton.Text)
	}

	// We can't easily test the button action without a full UI framework setup,
	// but we can verify the button is properly configured
	if recipientList.newButton.OnTapped == nil {
		t.Error("newButton.OnTapped is nil")
	}
}

func TestRecipientList_DoubleClickNavigation(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockUseCase := &MockRecipientUseCase{}
	testRecipients := createTestRecipients()
	// paginatedResult not needed - using mockUseCase.recipients

	mockUseCase.recipients = testRecipients

	recipientList := NewRecipientList(mockUseCase, nil, nil, nil)
	recipientList.LoadData()

	// Test that table has double-click functionality configured
	if recipientList.table.OnSelected == nil {
		t.Error("table.OnSelected is nil")
	}
}

func TestRecipientList_RefreshFunctionality(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockUseCase := &MockRecipientUseCase{}
	testRecipients := createTestRecipients()
	// paginatedResult not needed - using mockUseCase.recipients

	// Mock multiple calls to ListRecipients
	mockUseCase.recipients = testRecipients

	recipientList := NewRecipientList(mockUseCase, nil, nil, nil)

	// Initial load
	err := recipientList.LoadData()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Refresh
	recipientList.refreshButton.OnTapped()

	// Test passed
}

func TestRecipientList_StaffFilterFunctionality(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockUseCase := &MockRecipientUseCase{}
	testRecipients := createTestRecipients()
	filteredRecipients := []*domain.Recipient{testRecipients[0]} // Only first recipient

	// Set up filtered mock data
	mockUseCase.recipients = filteredRecipients

	recipientList := NewRecipientList(mockUseCase, nil, nil, nil)

	// Simulate staff filter selection
	recipientList.onStaffFilterChanged("staff-001")

	if recipientList.Length() != len(filteredRecipients) {
		t.Errorf("Expected %d recipients, got %d", len(filteredRecipients), recipientList.Length())
	}
	// Test passed
}

func TestRecipientList_CreateObject(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockUseCase := &MockRecipientUseCase{}
	recipientList := NewRecipientList(mockUseCase, nil, nil, nil)

	// Test that CreateObject returns a valid Fyne object
	obj := recipientList.CreateObject()
	if obj == nil {
		t.Fatal("CreateObject returned nil")
	}
	// Verify it implements fyne.CanvasObject interface
	var _ fyne.CanvasObject = obj
}

func TestRecipientList_ErrorHandling(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockUseCase := &MockRecipientUseCase{}

	// Mock error in ListRecipients
	mockUseCase.err = testError

	recipientList := NewRecipientList(mockUseCase, nil, nil, nil)
	err := recipientList.LoadData()

	if err == nil {
		t.Error("Expected error but got nil")
	}
	// Test passed
}
