package font

import (
	"fmt"
	"sync"
)

// FontRegistry manages a collection of fonts and provides lookup by name.
// It's safe for concurrent use.
type FontRegistry struct {
	mu    sync.RWMutex
	fonts map[string]*Font

	// Default font used when a font is not found
	defaultFont *Font
}

// NewFontRegistry creates a new font registry.
func NewFontRegistry() *FontRegistry {
	// Create a default font with WinAnsi encoding
	defaultFont := NewFont("DefaultFont")
	defaultFont.Encoding = EncodingWinAnsi

	return &FontRegistry{
		fonts:       make(map[string]*Font),
		defaultFont: defaultFont,
	}
}

// Register adds a font to the registry.
// If a font with the same name already exists, it will be replaced.
func (fr *FontRegistry) Register(font *Font) {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	fr.fonts[font.Name] = font
}

// RegisterSimple is a convenience method to register a font with basic info.
func (fr *FontRegistry) RegisterSimple(name string, encoding EncodingType) *Font {
	font := NewFont(name)
	font.Encoding = encoding
	fr.Register(font)
	return font
}

// RegisterWithToUnicode registers a font with a ToUnicode CMap.
func (fr *FontRegistry) RegisterWithToUnicode(name string, cmap *CMap) *Font {
	font := NewFont(name)
	font.ToUnicode = cmap
	font.Encoding = EncodingCustom // Mark as having custom encoding
	fr.Register(font)
	return font
}

// Lookup retrieves a font by name.
// If the font is not found, returns the default font and false.
func (fr *FontRegistry) Lookup(name string) (*Font, bool) {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	if font, ok := fr.fonts[name]; ok {
		return font, true
	}
	return fr.defaultFont, false
}

// MustLookup retrieves a font by name, returning the default font if not found.
func (fr *FontRegistry) MustLookup(name string) *Font {
	font, _ := fr.Lookup(name)
	return font
}

// SetDefaultFont sets the default font used when a font is not found.
func (fr *FontRegistry) SetDefaultFont(font *Font) {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	fr.defaultFont = font
}

// Count returns the number of registered fonts.
func (fr *FontRegistry) Count() int {
	fr.mu.RLock()
	defer fr.mu.RUnlock()
	return len(fr.fonts)
}

// List returns a slice of all registered font names.
func (fr *FontRegistry) List() []string {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	names := make([]string, 0, len(fr.fonts))
	for name := range fr.fonts {
		names = append(names, name)
	}
	return names
}

// Clear removes all registered fonts.
func (fr *FontRegistry) Clear() {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	fr.fonts = make(map[string]*Font)
}

// String returns a debug representation of the registry.
func (fr *FontRegistry) String() string {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	return fmt.Sprintf("FontRegistry with %d fonts: %v", len(fr.fonts), fr.List())
}
