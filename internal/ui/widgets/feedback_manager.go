package widgets

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// FeedbackType represents the type of feedback message
type FeedbackType int

const (
	FeedbackTypeInfo FeedbackType = iota
	FeedbackTypeSuccess
	FeedbackTypeWarning
	FeedbackTypeError
)

// FeedbackManager manages user feedback messages and error handling
type FeedbackManager struct {
	container     *fyne.Container
	currentWidget fyne.CanvasObject
	autoHide      bool
	hideDelay     time.Duration
	onHide        func()
	hideTimer     *time.Timer
}

// NewFeedbackManager creates a new feedback manager
func NewFeedbackManager() *FeedbackManager {
	return &FeedbackManager{
		container: container.NewWithoutLayout(),
		autoHide:  true,
		hideDelay: 5 * time.Second,
	}
}

// ShowInfo displays an informational message
func (fm *FeedbackManager) ShowInfo(message string) {
	label := widget.NewLabel(message)
	label.Importance = widget.MediumImportance
	fm.showWidget(label)
}

// ShowSuccess displays a success message
func (fm *FeedbackManager) ShowSuccess(message string) {
	label := widget.NewLabel(message)
	label.Importance = widget.SuccessImportance
	fm.showWidget(label)
}

// ShowWarning displays a warning message
func (fm *FeedbackManager) ShowWarning(message string) {
	label := widget.NewLabel(message)
	label.Importance = widget.WarningImportance
	fm.showWidget(label)
}

// ShowError displays an error message
func (fm *FeedbackManager) ShowError(message string) {
	label := widget.NewLabel(message)
	label.Importance = widget.DangerImportance
	fm.showWidget(label)
}

// ShowErrorFromError displays an error message from an error object
func (fm *FeedbackManager) ShowErrorFromError(err error) {
	var message string
	if err != nil {
		message = "エラー: " + err.Error()
	} else {
		message = "不明なエラーが発生しました"
	}
	fm.ShowError(message)
}

// ShowLoading displays a loading message with progress indicator
func (fm *FeedbackManager) ShowLoading(message string) {
	progressBar := widget.NewProgressBarInfinite()
	progressBar.Start()

	label := widget.NewLabel(message)
	label.Importance = widget.MediumImportance

	content := container.NewVBox(
		label,
		progressBar,
	)

	fm.showWidget(content)
}

// ShowConfirmation displays a confirmation message (typically for successful actions)
func (fm *FeedbackManager) ShowConfirmation(message string) {
	fm.ShowSuccess(message)
}

// Clear removes any currently displayed message
func (fm *FeedbackManager) Clear() {
	// Stop any running timer
	if fm.hideTimer != nil {
		fm.hideTimer.Stop()
		fm.hideTimer = nil
	}

	// Stop any progress bars
	if fm.currentWidget != nil {
		fm.stopProgressBars(fm.currentWidget)
	}

	// Clear the container
	fm.container.RemoveAll()
	fm.currentWidget = nil

	// Call hide callback
	if fm.onHide != nil {
		fm.onHide()
	}
}

// SetAutoHide configures automatic hiding of messages
func (fm *FeedbackManager) SetAutoHide(autoHide bool, delay time.Duration) {
	fm.autoHide = autoHide
	if delay > 0 {
		fm.hideDelay = delay
	}
}

// SetOnHideCallback sets a callback to be called when messages are hidden
func (fm *FeedbackManager) SetOnHideCallback(callback func()) {
	fm.onHide = callback
}

// GetContainer returns the container for embedding in UI
func (fm *FeedbackManager) GetContainer() *fyne.Container {
	return fm.container
}

// IsVisible returns whether a message is currently displayed
func (fm *FeedbackManager) IsVisible() bool {
	return fm.currentWidget != nil && len(fm.container.Objects) > 0
}

// showWidget displays a widget and handles auto-hide
func (fm *FeedbackManager) showWidget(widget fyne.CanvasObject) {
	// Clear any existing message
	fm.Clear()

	// Add new widget
	fm.currentWidget = widget
	fm.container.Add(widget)

	// Set up auto-hide timer
	if fm.autoHide && fm.hideDelay > 0 {
		fm.hideTimer = time.AfterFunc(fm.hideDelay, func() {
			fm.Clear()
		})
	}
}

// stopProgressBars recursively stops any progress bars in the widget tree
func (fm *FeedbackManager) stopProgressBars(obj fyne.CanvasObject) {
	if progressBar, ok := obj.(*widget.ProgressBarInfinite); ok {
		progressBar.Stop()
		return
	}

	if container, ok := obj.(*fyne.Container); ok {
		for _, child := range container.Objects {
			fm.stopProgressBars(child)
		}
	}
}
