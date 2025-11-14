package font

import (
	"fmt"
)

// EncodingType represents the type of encoding used by a font.
type EncodingType int

const (
	// EncodingUnknown indicates no encoding information available
	EncodingUnknown EncodingType = iota
	// EncodingWinAnsi is Windows ANSI encoding (CP1252)
	EncodingWinAnsi
	// EncodingMacRoman is Mac OS Roman encoding
	EncodingMacRoman
	// EncodingPDFDoc is PDFDocEncoding
	EncodingPDFDoc
	// EncodingIdentity is Identity-H or Identity-V (for CID fonts)
	EncodingIdentity
	// EncodingCustom indicates a custom encoding with ToUnicode CMap
	EncodingCustom
)

// Font represents a PDF font with its encoding information.
type Font struct {
	// Name is the font resource name (e.g., "/F1", "/TT2")
	Name string

	// BaseFont is the actual font name (e.g., "Helvetica", "TimesNewRoman")
	BaseFont string

	// Encoding type (WinAnsi, MacRoman, etc.)
	Encoding EncodingType

	// ToUnicode CMap (if available)
	ToUnicode *CMap

	// Whether this font uses multi-byte character codes
	IsMultiByte bool
}

// NewFont creates a new Font with the given name.
func NewFont(name string) *Font {
	return &Font{
		Name:        name,
		Encoding:    EncodingUnknown,
		ToUnicode:   nil,
		IsMultiByte: false,
	}
}

// DecodeText decodes text bytes using this font's encoding.
// It prioritizes ToUnicode CMap if available, then falls back to standard encodings.
func (f *Font) DecodeText(data []byte) string {
	// If we have a ToUnicode CMap, use it
	if f.ToUnicode != nil {
		return f.ToUnicode.DecodeString(data)
	}

	// Fall back to standard encodings
	switch f.Encoding {
	case EncodingWinAnsi:
		return DecodeWinAnsi(data)
	case EncodingPDFDoc:
		return DecodePDFDoc(data)
	case EncodingIdentity:
		// Identity encoding - typically means we need ToUnicode
		// Without it, we can't decode properly
		return string(data) // Raw bytes as fallback
	default:
		// Unknown encoding - try as ASCII/Latin1
		return string(data)
	}
}

// String returns a debug representation of the font.
func (f *Font) String() string {
	hasToUnicode := "no"
	if f.ToUnicode != nil {
		hasToUnicode = "yes"
	}
	return fmt.Sprintf("Font{Name: %s, BaseFont: %s, Encoding: %v, ToUnicode: %s, MultiБyte: %v}",
		f.Name, f.BaseFont, f.Encoding, hasToUnicode, f.IsMultiByte)
}
