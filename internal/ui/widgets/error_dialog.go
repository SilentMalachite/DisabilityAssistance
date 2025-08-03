package widgets

import (
	"errors"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// ErrorDialog provides user-friendly error dialogs
type ErrorDialog struct {
	window fyne.Window
}

// NewErrorDialog creates a new error dialog manager
func NewErrorDialog(window fyne.Window) *ErrorDialog {
	return &ErrorDialog{
		window: window,
	}
}

// ShowError displays a general error dialog
func (ed *ErrorDialog) ShowError(title string, err error) {
	message := ed.FormatError(err)

	dialog.ShowError(
		errors.New(message),
		ed.window,
	)
}

// ShowValidationError displays validation errors in a user-friendly format
func (ed *ErrorDialog) ShowValidationError(errors []string) {
	message := ed.FormatValidationErrors(errors)

	d := dialog.NewInformation(
		"入力エラー",
		message,
		ed.window,
	)
	d.Show()
}

// ShowConnectionError displays a connection error dialog
func (ed *ErrorDialog) ShowConnectionError() {
	message := ed.GetConnectionErrorMessage()

	d := dialog.NewError(
		errors.New(message),
		ed.window,
	)
	d.Show()
}

// ShowPermissionError displays a permission error dialog
func (ed *ErrorDialog) ShowPermissionError(customMessage string) {
	message := ed.GetPermissionErrorMessage(customMessage)

	d := dialog.NewError(
		errors.New(message),
		ed.window,
	)
	d.Show()
}

// ShowConfirmation displays a confirmation dialog
func (ed *ErrorDialog) ShowConfirmation(title, message string, onConfirm, onCancel func()) {
	d := dialog.NewConfirm(
		title,
		message,
		func(confirmed bool) {
			if confirmed && onConfirm != nil {
				onConfirm()
			} else if !confirmed && onCancel != nil {
				onCancel()
			}
		},
		ed.window,
	)
	d.Show()
}

// ShowInfo displays an information dialog
func (ed *ErrorDialog) ShowInfo(title, message string) {
	d := dialog.NewInformation(title, message, ed.window)
	d.Show()
}

// FormatError formats an error for display
func (ed *ErrorDialog) FormatError(err error) string {
	if err == nil {
		return "不明なエラーが発生しました"
	}
	return "エラーが発生しました:\n" + err.Error()
}

// FormatValidationErrors formats validation errors for display
func (ed *ErrorDialog) FormatValidationErrors(errors []string) string {
	if len(errors) == 0 {
		return "入力内容に問題があります"
	}

	var formatted strings.Builder
	formatted.WriteString("入力内容に問題があります:")

	for _, err := range errors {
		formatted.WriteString("\n• ")
		formatted.WriteString(err)
	}

	return formatted.String()
}

// GetConnectionErrorMessage returns a standard connection error message
func (ed *ErrorDialog) GetConnectionErrorMessage() string {
	return "接続エラーが発生しました。\n\nネットワーク接続を確認して、\n再度お試しください。"
}

// GetPermissionErrorMessage returns a permission error message
func (ed *ErrorDialog) GetPermissionErrorMessage(customMessage string) string {
	if customMessage != "" {
		return "権限エラー:\n" + customMessage
	}
	return "この操作を実行する権限がありません。\n\n管理者にお問い合わせください。"
}
