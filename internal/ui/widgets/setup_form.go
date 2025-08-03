package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"shien-system/internal/usecase"
)

type SetupForm struct {
	setupUseCase    usecase.SetupUseCase
	onSetupComplete func()
	feedbackManager *FeedbackManager
}

func NewSetupForm(
	setupUseCase usecase.SetupUseCase,
	onSetupComplete func(),
	feedbackManager *FeedbackManager,
) *SetupForm {
	return &SetupForm{
		setupUseCase:    setupUseCase,
		onSetupComplete: onSetupComplete,
		feedbackManager: feedbackManager,
	}
}

func (f *SetupForm) CreateContent() fyne.CanvasObject {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("管理者名")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("パスワード（8文字以上）")

	confirmPasswordEntry := widget.NewPasswordEntry()
	confirmPasswordEntry.SetPlaceHolder("パスワード（確認）")

	submitButton := widget.NewButton("初期設定を完了", func() {
		name := nameEntry.Text
		password := passwordEntry.Text
		confirmPassword := confirmPasswordEntry.Text

		// Validation
		if name == "" {
			f.feedbackManager.ShowError("管理者名を入力してください")
			return
		}

		if password != confirmPassword {
			f.feedbackManager.ShowError("パスワードが一致しません")
			return
		}

		if len(password) < 8 {
			f.feedbackManager.ShowError("パスワードは8文字以上で入力してください")
			return
		}

		// Create initial admin
		if err := f.setupUseCase.CreateInitialAdmin(nil, name, password); err != nil {
			f.feedbackManager.ShowError("初期設定に失敗しました: " + err.Error())
			return
		}

		f.feedbackManager.ShowSuccess("初期設定が完了しました")
		if f.onSetupComplete != nil {
			f.onSetupComplete()
		}
	})

	form := container.NewVBox(
		widget.NewLabelWithStyle("システム初期設定", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("最初の管理者アカウントを作成してください"),
		widget.NewSeparator(),
		widget.NewForm(
			widget.NewFormItem("管理者名", nameEntry),
			widget.NewFormItem("パスワード", passwordEntry),
			widget.NewFormItem("パスワード（確認）", confirmPasswordEntry),
		),
		submitButton,
	)

	return container.NewCenter(
		container.NewMax(form),
	)
}
