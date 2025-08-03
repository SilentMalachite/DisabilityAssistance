package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// JapaneseTheme provides a theme optimized for Japanese accessibility
type JapaneseTheme struct {
	fyne.Theme
}

// NewJapaneseTheme creates a new Japanese theme instance
func NewJapaneseTheme() *JapaneseTheme {
	return &JapaneseTheme{
		Theme: theme.DefaultTheme(),
	}
}

// Font returns the appropriate font for Japanese text
func (jt *JapaneseTheme) Font(style fyne.TextStyle) fyne.Resource {
	// For production use, embed Noto Sans CJK font files here
	// For development, rely on system fonts that support Japanese
	// The OS should provide appropriate fallback fonts for Japanese text
	return theme.DefaultTheme().Font(style)
}

// Color returns colors optimized for accessibility
func (jt *JapaneseTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 48, G: 48, B: 48, A: 255}
		}
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	case theme.ColorNameForeground:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}
		return color.NRGBA{R: 33, G: 33, B: 33, A: 255}
	case theme.ColorNamePrimary:
		// Use professional blue for welfare systems
		return color.NRGBA{R: 30, G: 100, B: 180, A: 255}
	case theme.ColorNameError:
		// Use accessible red with sufficient contrast
		return color.NRGBA{R: 200, G: 40, B: 30, A: 255}
	case theme.ColorNameWarning:
		// Use amber for warnings (better accessibility than orange)
		return color.NRGBA{R: 230, G: 150, B: 20, A: 255}
	case theme.ColorNameSuccess:
		// Use accessible green
		return color.NRGBA{R: 20, G: 130, B: 20, A: 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Size returns sizes optimized for Japanese text readability
func (jt *JapaneseTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case "text":
		return 14 // Slightly larger text for better readability with Japanese
	case "caption":
		return 12 // Larger caption text
	case "inline_icon":
		return 20 // Larger inline icons
	case "padding":
		return 6 // More padding for better spacing
	case "inner_padding":
		return 10 // More inner padding
	case "scrollbar":
		return 16 // Wider scroll bars for better accessibility
	case "separator":
		return 2 // Thicker separators for better visibility
	default:
		return theme.DefaultTheme().Size(name)
	}
}
