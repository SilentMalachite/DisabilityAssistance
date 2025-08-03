package widgets

import (
	"testing"

	"fyne.io/fyne/v2/widget"
	"github.com/stretchr/testify/assert"
)

func TestNewAccessibilityManager(t *testing.T) {
	am := NewAccessibilityManager()

	assert.NotNil(t, am)
	assert.True(t, am.enabled)
	assert.Equal(t, -1, am.currentFocus)
	assert.Empty(t, am.focusChain)
}

func TestAccessibilityManager_RegisterFocusable(t *testing.T) {
	am := NewAccessibilityManager()
	entry := widget.NewEntry()

	am.RegisterFocusable(entry)

	assert.Len(t, am.focusChain, 1)
	assert.Equal(t, entry, am.focusChain[0])
}

func TestAccessibilityManager_NextFocus(t *testing.T) {
	am := NewAccessibilityManager()
	entry1 := widget.NewEntry()
	entry2 := widget.NewEntry()

	am.RegisterFocusable(entry1)
	am.RegisterFocusable(entry2)

	// First call should focus the first widget
	am.NextFocus()
	assert.Equal(t, 0, am.currentFocus)

	// Second call should focus the second widget
	am.NextFocus()
	assert.Equal(t, 1, am.currentFocus)

	// Third call should wrap around to the first widget
	am.NextFocus()
	assert.Equal(t, 0, am.currentFocus)
}

func TestAccessibilityManager_PreviousFocus(t *testing.T) {
	am := NewAccessibilityManager()
	entry1 := widget.NewEntry()
	entry2 := widget.NewEntry()

	am.RegisterFocusable(entry1)
	am.RegisterFocusable(entry2)

	// First call should focus the last widget (wrap around)
	am.PreviousFocus()
	assert.Equal(t, 1, am.currentFocus)

	// Second call should focus the first widget
	am.PreviousFocus()
	assert.Equal(t, 0, am.currentFocus)
}

func TestAccessibilityManager_SetEnabled(t *testing.T) {
	am := NewAccessibilityManager()
	entry := widget.NewEntry()
	am.RegisterFocusable(entry)

	// Disable accessibility
	am.SetEnabled(false)
	assert.False(t, am.enabled)

	// NextFocus should not work when disabled
	originalFocus := am.currentFocus
	am.NextFocus()
	assert.Equal(t, originalFocus, am.currentFocus)

	// Re-enable accessibility
	am.SetEnabled(true)
	assert.True(t, am.enabled)

	// NextFocus should work again
	am.NextFocus()
	assert.Equal(t, 0, am.currentFocus)
}

func TestAccessibilityManager_ClearFocusChain(t *testing.T) {
	am := NewAccessibilityManager()
	entry := widget.NewEntry()
	am.RegisterFocusable(entry)

	assert.Len(t, am.focusChain, 1)

	am.ClearFocusChain()

	assert.Empty(t, am.focusChain)
	assert.Equal(t, -1, am.currentFocus)
}

func TestNewAccessibleButton(t *testing.T) {
	called := false
	btn := NewAccessibleButton("テストボタン", "テスト用のボタンです", func() {
		called = true
	})

	assert.NotNil(t, btn)
	assert.Equal(t, "テストボタン", btn.Text)
	assert.Equal(t, "テスト用のボタンです", btn.GetDescription())

	// Test button tap
	btn.OnTapped()
	assert.True(t, called)
}

func TestAccessibleButton_SetShortcut(t *testing.T) {
	btn := NewAccessibleButton("テスト", "説明", func() {})

	btn.SetShortcut("Ctrl+T")
	assert.Equal(t, "Ctrl+T", btn.shortcut)
}

func TestAccessibleButton_GetDescription(t *testing.T) {
	// Test with custom description
	btn1 := NewAccessibleButton("ボタン", "カスタム説明", func() {})
	assert.Equal(t, "カスタム説明", btn1.GetDescription())

	// Test fallback to button text
	btn2 := NewAccessibleButton("ボタンテキスト", "", func() {})
	assert.Equal(t, "ボタンテキスト", btn2.GetDescription())
}

func TestNewAccessibleEntry(t *testing.T) {
	entry := NewAccessibleEntry("名前", "お名前を入力してください")

	assert.NotNil(t, entry)
	assert.Equal(t, "名前", entry.GetLabel())
	assert.Equal(t, "お名前を入力してください", entry.PlaceHolder)
	assert.False(t, entry.IsRequired())
}

func TestAccessibleEntry_SetRequired(t *testing.T) {
	entry := NewAccessibleEntry("名前", "お名前を入力")

	entry.SetRequired(true)

	assert.True(t, entry.IsRequired())
	assert.Equal(t, "お名前を入力 (必須)", entry.PlaceHolder)

	entry.SetRequired(false)
	assert.False(t, entry.IsRequired())
}

func TestNewAccessibleContainer(t *testing.T) {
	container := NewAccessibleContainer("テストコンテナ", "テスト用のコンテナです")

	assert.NotNil(t, container)
	assert.Equal(t, "テストコンテナ", container.GetTitle())
	assert.Equal(t, "テスト用のコンテナです", container.GetDescription())
	assert.NotNil(t, container.manager)
}

func TestAccessibleContainer_AddFocusable(t *testing.T) {
	container := NewAccessibleContainer("テスト", "説明")
	entry := widget.NewEntry()

	container.AddFocusable(entry)

	assert.Len(t, container.manager.focusChain, 1)
	assert.Equal(t, entry, container.manager.focusChain[0])
}

func TestNewHighContrastTheme(t *testing.T) {
	theme := NewHighContrastTheme()

	assert.NotNil(t, theme)

	// Test high contrast colors
	bg := theme.Color("background", 0) // Use 0 for light variant
	assert.NotNil(t, bg)

	fg := theme.Color("foreground", 0) // Use 0 for light variant
	assert.NotNil(t, fg)

	// Test larger text size
	textSize := theme.Size("text")
	assert.Equal(t, float32(16), textSize)
}

func TestDefaultAccessibilitySettings(t *testing.T) {
	settings := DefaultAccessibilitySettings()

	assert.NotNil(t, settings)
	assert.False(t, settings.HighContrast)
	assert.False(t, settings.LargeFonts)
	assert.False(t, settings.ReducedMotion)
	assert.False(t, settings.ScreenReader)
	assert.True(t, settings.KeyboardNav)
	assert.Equal(t, float32(1.0), settings.FontScale)
	assert.Equal(t, float32(1.0), settings.AnimationSpeed)
}

func BenchmarkAccessibilityManager_NextFocus(b *testing.B) {
	am := NewAccessibilityManager()

	// Add many focusable widgets
	for i := 0; i < 100; i++ {
		entry := widget.NewEntry()
		am.RegisterFocusable(entry)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		am.NextFocus()
	}
}

func BenchmarkAccessibilityManager_RegisterFocusable(b *testing.B) {
	am := NewAccessibilityManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry := widget.NewEntry()
		am.RegisterFocusable(entry)
	}
}
