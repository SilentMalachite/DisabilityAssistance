package pdf

import (
	"embed"
	"io/fs"
)

// Embedded Japanese fonts for PDF generation
// In production, you would include the actual Noto Sans CJK font files here
// embeddedFonts would contain embedded font files
// Currently using empty filesystem as placeholder
var embeddedFonts embed.FS

// GetEmbeddedFontPath returns the path to embedded font files
func GetEmbeddedFontPath() fs.FS {
	return embeddedFonts
}

// FontManager manages font resources for PDF generation
type FontManager struct {
	fontFS fs.FS
}

// NewFontManager creates a new font manager
func NewFontManager() *FontManager {
	return &FontManager{
		fontFS: embeddedFonts,
	}
}

// GetFontBytes returns the bytes of the specified font file
func (fm *FontManager) GetFontBytes(fontName string) ([]byte, error) {
	return fs.ReadFile(fm.fontFS, "fonts/"+fontName)
}

// HasFont checks if the specified font exists
func (fm *FontManager) HasFont(fontName string) bool {
	_, err := fs.Stat(fm.fontFS, "fonts/"+fontName)
	return err == nil
}
