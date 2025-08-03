package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// AccessibilityManager handles accessibility features for the application
type AccessibilityManager struct {
	focusChain   []fyne.Focusable
	currentFocus int
	enabled      bool
}

// NewAccessibilityManager creates a new accessibility manager
func NewAccessibilityManager() *AccessibilityManager {
	return &AccessibilityManager{
		focusChain:   make([]fyne.Focusable, 0),
		currentFocus: -1,
		enabled:      true,
	}
}

// 新しいファイルとして staff_list.go を作成

// RegisterFocusable adds a focusable widget to the focus chain
func (am *AccessibilityManager) RegisterFocusable(widget fyne.Focusable) {
	am.focusChain = append(am.focusChain, widget)
}

// NextFocus moves focus to the next focusable element
func (am *AccessibilityManager) NextFocus() {
	if !am.enabled || len(am.focusChain) == 0 {
		return
	}

	am.currentFocus = (am.currentFocus + 1) % len(am.focusChain)
	am.focusChain[am.currentFocus].FocusGained()
}

// PreviousFocus moves focus to the previous focusable element
func (am *AccessibilityManager) PreviousFocus() {
	if !am.enabled || len(am.focusChain) == 0 {
		return
	}

	am.currentFocus--
	if am.currentFocus < 0 {
		am.currentFocus = len(am.focusChain) - 1
	}
	am.focusChain[am.currentFocus].FocusGained()
}

// ClearFocusChain removes all widgets from the focus chain
func (am *AccessibilityManager) ClearFocusChain() {
	am.focusChain = make([]fyne.Focusable, 0)
	am.currentFocus = -1
}

// SetEnabled enables or disables accessibility features
func (am *AccessibilityManager) SetEnabled(enabled bool) {
	am.enabled = enabled
}

// AccessibleButton extends widget.Button with accessibility features
type AccessibleButton struct {
	widget.Button
	description string
	shortcut    string
}

// NewAccessibleButton creates a new accessible button
func NewAccessibleButton(text string, description string, tapped func()) *AccessibleButton {
	btn := &AccessibleButton{
		description: description,
	}
	btn.Button.Text = text
	btn.Button.OnTapped = tapped
	btn.ExtendBaseWidget(btn)
	return btn
}

// SetShortcut sets a keyboard shortcut for the button
func (ab *AccessibleButton) SetShortcut(shortcut string) {
	ab.shortcut = shortcut
}

// GetDescription returns the accessibility description
func (ab *AccessibleButton) GetDescription() string {
	if ab.description != "" {
		return ab.description
	}
	return ab.Text
}

// AccessibleEntry extends widget.Entry with accessibility features
type AccessibleEntry struct {
	widget.Entry
	label       string
	placeholder string
	required    bool
}

// NewAccessibleEntry creates a new accessible entry
func NewAccessibleEntry(label string, placeholder string) *AccessibleEntry {
	entry := &AccessibleEntry{
		label:       label,
		placeholder: placeholder,
	}
	entry.Entry.PlaceHolder = placeholder
	entry.ExtendBaseWidget(entry)
	return entry
}

// SetRequired marks the entry as required
func (ae *AccessibleEntry) SetRequired(required bool) {
	ae.required = required
	if required && ae.placeholder != "" {
		ae.Entry.PlaceHolder = ae.placeholder + " (必須)"
	}
}

// GetLabel returns the accessibility label
func (ae *AccessibleEntry) GetLabel() string {
	return ae.label
}

// IsRequired returns whether the entry is required
func (ae *AccessibleEntry) IsRequired() bool {
	return ae.required
}

// AccessibleContainer wraps containers with accessibility features
type AccessibleContainer struct {
	*fyne.Container
	title       string
	description string
	manager     *AccessibilityManager
}

// NewAccessibleContainer creates a new accessible container
func NewAccessibleContainer(title string, description string) *AccessibleContainer {
	ac := &AccessibleContainer{
		Container:   container.NewWithoutLayout(),
		title:       title,
		description: description,
		manager:     NewAccessibilityManager(),
	}
	return ac
}

// AddFocusable adds a focusable widget and registers it with the accessibility manager
func (ac *AccessibleContainer) AddFocusable(widget fyne.Focusable) {
	ac.manager.RegisterFocusable(widget)
}

// GetTitle returns the container title for screen readers
func (ac *AccessibleContainer) GetTitle() string {
	return ac.title
}

// GetDescription returns the container description for screen readers
func (ac *AccessibleContainer) GetDescription() string {
	return ac.description
}

// HighContrastTheme provides a high contrast theme for accessibility
type HighContrastTheme struct {
	fyne.Theme
}

// NewHighContrastTheme creates a new high contrast theme
func NewHighContrastTheme() *HighContrastTheme {
	return &HighContrastTheme{}
}

// Color returns high contrast colors
func (hct *HighContrastTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.Black // Black background
	case theme.ColorNameForeground:
		return color.White // White text
	case theme.ColorNameButton:
		return color.RGBA{R: 255, G: 255, B: 0, A: 255} // Yellow buttons
	case theme.ColorNamePrimary:
		return color.RGBA{R: 255, G: 255, B: 0, A: 255} // Yellow primary
	case theme.ColorNameFocus:
		return color.RGBA{R: 255, G: 0, B: 0, A: 255} // Red focus
	default:
		// Fallback to default theme for other colors
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Font returns the default font for high contrast theme
func (hct *HighContrastTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon returns the default icon for high contrast theme
func (hct *HighContrastTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns larger sizes for better accessibility
func (hct *HighContrastTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case "text":
		return 16 // Larger text
	case "inline_icon":
		return 24 // Larger icons
	case "padding":
		return 8 // More padding
	case "inner_padding":
		return 12 // More inner padding
	default:
		return theme.DefaultTheme().Size(name)
	}
}

// AccessibilitySettings stores user accessibility preferences
type AccessibilitySettings struct {
	HighContrast   bool    `json:"high_contrast"`
	LargeFonts     bool    `json:"large_fonts"`
	ReducedMotion  bool    `json:"reduced_motion"`
	ScreenReader   bool    `json:"screen_reader"`
	KeyboardNav    bool    `json:"keyboard_nav"`
	FontScale      float32 `json:"font_scale"`
	AnimationSpeed float32 `json:"animation_speed"`
}

// DefaultAccessibilitySettings returns default accessibility settings
func DefaultAccessibilitySettings() *AccessibilitySettings {
	return &AccessibilitySettings{
		HighContrast:   false,
		LargeFonts:     false,
		ReducedMotion:  false,
		ScreenReader:   false,
		KeyboardNav:    true,
		FontScale:      1.0,
		AnimationSpeed: 1.0,
	}
}

// Apply applies the accessibility settings to the application
func (as *AccessibilitySettings) Apply(app fyne.App) {
	if as.HighContrast {
		app.Settings().SetTheme(NewHighContrastTheme())
	}

	// Additional settings would be applied here in a real implementation
	// For example, adjusting animation speeds, enabling screen reader mode, etc.
}
