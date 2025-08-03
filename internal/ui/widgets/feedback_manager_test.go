package widgets

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

// Note: FeedbackType, constants, and FeedbackManager are defined in feedback_manager.go

func TestNewFeedbackManager(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	manager := NewFeedbackManager()

	if manager == nil {
		t.Fatal("NewFeedbackManager returned nil")
	}

	if manager.container == nil {
		t.Error("Container not initialized")
	}

	if manager.hideDelay != 5*time.Second {
		t.Error("Default hide delay should be 5 seconds")
	}

	if !manager.autoHide {
		t.Error("Auto hide should be enabled by default")
	}
}

func TestFeedbackManager_ShowInfo(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	manager := NewFeedbackManager()

	message := "情報メッセージ"
	manager.ShowInfo(message)

	// Verify message is shown
	if manager.currentWidget == nil {
		t.Error("Current widget should not be nil after showing message")
	}

	// Verify it's a label with the correct text
	if label, ok := manager.currentWidget.(*widget.Label); ok {
		if label.Text != message {
			t.Errorf("Expected message '%s', got '%s'", message, label.Text)
		}
	} else {
		t.Error("Current widget should be a label")
	}
}

func TestFeedbackManager_ShowSuccess(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	manager := NewFeedbackManager()

	message := "成功メッセージ"
	manager.ShowSuccess(message)

	// Verify message is shown
	if manager.currentWidget == nil {
		t.Error("Current widget should not be nil after showing success")
	}

	// Verify it's a label with success styling
	if label, ok := manager.currentWidget.(*widget.Label); ok {
		if label.Text != message {
			t.Errorf("Expected message '%s', got '%s'", message, label.Text)
		}
		// Success messages should have different importance
		if label.Importance != widget.SuccessImportance {
			t.Error("Success message should have SuccessImportance")
		}
	} else {
		t.Error("Current widget should be a label")
	}
}

func TestFeedbackManager_ShowWarning(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	manager := NewFeedbackManager()

	message := "警告メッセージ"
	manager.ShowWarning(message)

	// Verify message is shown
	if manager.currentWidget == nil {
		t.Error("Current widget should not be nil after showing warning")
	}

	// Verify it's a label with warning styling
	if label, ok := manager.currentWidget.(*widget.Label); ok {
		if label.Text != message {
			t.Errorf("Expected message '%s', got '%s'", message, label.Text)
		}
		if label.Importance != widget.WarningImportance {
			t.Error("Warning message should have WarningImportance")
		}
	} else {
		t.Error("Current widget should be a label")
	}
}

func TestFeedbackManager_ShowError(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	manager := NewFeedbackManager()

	message := "エラーメッセージ"
	manager.ShowError(message)

	// Verify message is shown
	if manager.currentWidget == nil {
		t.Error("Current widget should not be nil after showing error")
	}

	// Verify it's a label with error styling
	if label, ok := manager.currentWidget.(*widget.Label); ok {
		if label.Text != message {
			t.Errorf("Expected message '%s', got '%s'", message, label.Text)
		}
		if label.Importance != widget.DangerImportance {
			t.Error("Error message should have DangerImportance")
		}
	} else {
		t.Error("Current widget should be a label")
	}
}

func TestFeedbackManager_ShowErrorFromError(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	manager := NewFeedbackManager()

	err := &mockError{message: "テストエラー"}
	manager.ShowErrorFromError(err)

	// Verify error message is shown
	if label, ok := manager.currentWidget.(*widget.Label); ok {
		expected := "エラー: テストエラー"
		if label.Text != expected {
			t.Errorf("Expected message '%s', got '%s'", expected, label.Text)
		}
	} else {
		t.Error("Current widget should be a label")
	}
}

func TestFeedbackManager_ShowErrorFromError_Nil(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	manager := NewFeedbackManager()

	manager.ShowErrorFromError(nil)

	// Should show generic error message
	if label, ok := manager.currentWidget.(*widget.Label); ok {
		expected := "不明なエラーが発生しました"
		if label.Text != expected {
			t.Errorf("Expected message '%s', got '%s'", expected, label.Text)
		}
	} else {
		t.Error("Current widget should be a label")
	}
}

func TestFeedbackManager_Clear(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	manager := NewFeedbackManager()

	// Show a message first
	manager.ShowInfo("テストメッセージ")

	// Clear it
	manager.Clear()

	// Verify it's cleared
	if manager.currentWidget != nil {
		t.Error("Current widget should be nil after clear")
	}

	// Container should be empty
	if len(manager.container.Objects) != 0 {
		t.Error("Container should be empty after clear")
	}
}

func TestFeedbackManager_SetAutoHide(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	manager := NewFeedbackManager()

	// Disable auto hide
	manager.SetAutoHide(false, 0)

	if manager.autoHide {
		t.Error("Auto hide should be disabled")
	}

	// Enable auto hide with custom delay
	manager.SetAutoHide(true, 2*time.Second)

	if !manager.autoHide {
		t.Error("Auto hide should be enabled")
	}

	if manager.hideDelay != 2*time.Second {
		t.Error("Hide delay should be 2 seconds")
	}
}

func TestFeedbackManager_SetOnHideCallback(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	manager := NewFeedbackManager()

	callbackCalled := false
	manager.SetOnHideCallback(func() {
		callbackCalled = true
	})

	// Manually trigger hide
	manager.Clear()

	if manager.onHide != nil {
		manager.onHide()
	}

	if !callbackCalled {
		t.Error("Hide callback should be called")
	}
}

func TestFeedbackManager_GetContainer(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	manager := NewFeedbackManager()

	container := manager.GetContainer()

	if container == nil {
		t.Error("GetContainer should not return nil")
	}

	if container != manager.container {
		t.Error("Should return the same container instance")
	}
}

func TestFeedbackManager_IsVisible(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	manager := NewFeedbackManager()

	// Initially not visible
	if manager.IsVisible() {
		t.Error("Should not be visible initially")
	}

	// Show message
	manager.ShowInfo("テストメッセージ")

	// Should be visible
	if !manager.IsVisible() {
		t.Error("Should be visible after showing message")
	}

	// Clear message
	manager.Clear()

	// Should not be visible
	if manager.IsVisible() {
		t.Error("Should not be visible after clear")
	}
}

func TestFeedbackManager_ShowLoading(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	manager := NewFeedbackManager()

	message := "読み込み中..."
	manager.ShowLoading(message)

	// Verify loading message is shown
	if !manager.IsVisible() {
		t.Error("Should be visible when showing loading")
	}

	// Should contain progress bar (may be nested in a container)
	found := false
	for _, obj := range manager.container.Objects {
		if _, ok := obj.(*widget.ProgressBarInfinite); ok {
			found = true
			break
		}
		// Check if it's a container with progress bar
		if container, ok := obj.(*fyne.Container); ok {
			for _, child := range container.Objects {
				if _, ok := child.(*widget.ProgressBarInfinite); ok {
					found = true
					break
				}
			}
		}
	}

	if !found {
		t.Error("Loading should display a progress bar")
	}
}

func TestFeedbackManager_ShowConfirmation(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	manager := NewFeedbackManager()

	message := "保存しました"
	manager.ShowConfirmation(message)

	// Verify confirmation message is shown with checkmark
	if !manager.IsVisible() {
		t.Error("Should be visible when showing confirmation")
	}

	// Should be styled as success
	if label, ok := manager.currentWidget.(*widget.Label); ok {
		if label.Importance != widget.SuccessImportance {
			t.Error("Confirmation should have SuccessImportance")
		}
	}
}

// Mock error for testing
type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}

func TestFeedbackManager_AutoHideAfterDelay(t *testing.T) {
	myApp := app.New()
	defer myApp.Quit()

	manager := NewFeedbackManager()
	manager.SetAutoHide(true, 50*time.Millisecond) // Very short delay for testing

	hideCallbackCalled := false
	manager.SetOnHideCallback(func() {
		hideCallbackCalled = true
	})

	// Show message
	manager.ShowInfo("テストメッセージ")

	// Initially visible
	if !manager.IsVisible() {
		t.Error("Should be visible immediately after showing")
	}

	// Wait for auto hide
	time.Sleep(100 * time.Millisecond)

	// Should be hidden
	if manager.IsVisible() {
		t.Error("Should be hidden after delay")
	}

	if !hideCallbackCalled {
		t.Error("Hide callback should be called after auto hide")
	}
}
