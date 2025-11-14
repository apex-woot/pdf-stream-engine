package font

import (
	"strings"
)

// winAnsiToUnicode maps WinAnsiEncoding bytes (0x80-0xFF) to Unicode runes.
// Based on Windows Code Page 1252.
var winAnsiToUnicode = map[byte]rune{
	0x80: '\u20AC', // Euro
	0x82: '\u201A', // Single Low-9 Quotation Mark
	0x83: '\u0192', // Latin Small Letter F with Hook
	0x84: '\u201E', // Double Low-9 Quotation Mark
	0x85: '\u2026', // Ellipsis
	0x86: '\u2020', // Dagger
	0x87: '\u2021', // Double Dagger
	0x88: '\u02C6', // Modifier Letter Circumflex Accent
	0x89: '\u2030', // Per Mille Sign
	0x8A: '\u0160', // Latin Capital Letter S with Caron
	0x8B: '\u2039', // Single Left-Pointing Angle Quotation Mark
	0x8C: '\u0152', // Latin Capital Ligature OE
	0x8E: '\u017D', // Latin Capital Letter Z with Caron
	0x91: '\u2018', // Left Single Quotation Mark
	0x92: '\u2019', // Right Single Quotation Mark
	0x93: '\u201C', // Left Double Quotation Mark
	0x94: '\u201D', // Right Double Quotation Mark
	0x95: '\u2022', // Bullet
	0x96: '\u2013', // En Dash
	0x97: '\u2014', // Em Dash
	0x98: '\u02DC', // Small Tilde
	0x99: '\u2122', // Trade Mark Sign
	0x9A: '\u0161', // Latin Small Letter S with Caron
	0x9B: '\u203A', // Single Right-Pointing Angle Quotation Mark
	0x9C: '\u0153', // Latin Small Ligature OE
	0x9E: '\u017E', // Latin Small Letter Z with Caron
	0x9F: '\u0178', // Latin Capital Letter Y with Diaeresis
}

// DecodeWinAnsi converts bytes from WinAnsiEncoding to UTF-8 string.
// Characters 0x00-0x7F are standard ASCII.
// Characters 0x80-0xFF are mapped according to Windows CP1252.
func DecodeWinAnsi(data []byte) string {
	var b strings.Builder
	b.Grow(len(data))
	for _, byteVal := range data {
		if byteVal < 0x80 {
			// Standard ASCII
			b.WriteByte(byteVal)
		} else {
			// Look up in WinAnsi map
			if r, ok := winAnsiToUnicode[byteVal]; ok {
				b.WriteRune(r)
			} else {
				// For unmapped bytes in 0x80-0xFF range, use ISO Latin-1
				b.WriteRune(rune(byteVal))
			}
		}
	}
	return b.String()
}

// pdfDocToUnicode maps PDFDocEncoding bytes (0x80-0xFF) to Unicode runes.
// PDFDocEncoding is similar to ISO Latin-1 but with some differences in 0x80-0x9F range.
var pdfDocToUnicode = map[byte]rune{
	0x80: '\u2022', // Bullet
	0x81: '\u2020', // Dagger
	0x82: '\u2021', // Double Dagger
	0x83: '\u2026', // Ellipsis
	0x84: '\u2014', // Em Dash
	0x85: '\u2013', // En Dash
	0x86: '\u0192', // Latin Small Letter F with Hook
	0x87: '\u2044', // Fraction Slash
	0x88: '\u2039', // Single Left-Pointing Angle Quotation Mark
	0x89: '\u203A', // Single Right-Pointing Angle Quotation Mark
	0x8A: '\u2212', // Minus Sign
	0x8B: '\u2030', // Per Mille Sign
	0x8C: '\u201E', // Double Low-9 Quotation Mark
	0x8D: '\u201C', // Left Double Quotation Mark
	0x8E: '\u201D', // Right Double Quotation Mark
	0x8F: '\u2018', // Left Single Quotation Mark
	0x90: '\u2019', // Right Single Quotation Mark
	0x91: '\u201A', // Single Low-9 Quotation Mark
	0x92: '\u2122', // Trade Mark Sign
	0x93: '\uFB01', // Latin Small Ligature FI
	0x94: '\uFB02', // Latin Small Ligature FL
	0x95: '\u0141', // Latin Capital Letter L with Stroke
	0x96: '\u0152', // Latin Capital Ligature OE
	0x97: '\u0160', // Latin Capital Letter S with Caron
	0x98: '\u0178', // Latin Capital Letter Y with Diaeresis
	0x99: '\u017D', // Latin Capital Letter Z with Caron
	0x9A: '\u0131', // Latin Small Letter Dotless I
	0x9B: '\u0142', // Latin Small Letter L with Stroke
	0x9C: '\u0153', // Latin Small Ligature OE
	0x9D: '\u0161', // Latin Small Letter S with Caron
	0x9E: '\u017E', // Latin Small Letter Z with Caron
	0x9F: '\uFFFD', // Replacement Character
	// 0xA0-0xFF are same as ISO Latin-1
}

// DecodePDFDoc converts bytes from PDFDocEncoding to UTF-8 string.
// Characters 0x00-0x7F are standard ASCII.
// Characters 0x80-0x9F use special PDFDocEncoding mappings.
// Characters 0xA0-0xFF are ISO Latin-1.
func DecodePDFDoc(data []byte) string {
	var b strings.Builder
	b.Grow(len(data))
	for _, byteVal := range data {
		if byteVal < 0x80 {
			// Standard ASCII
			b.WriteByte(byteVal)
		} else if byteVal < 0xA0 {
			// PDFDocEncoding special range
			if r, ok := pdfDocToUnicode[byteVal]; ok {
				b.WriteRune(r)
			} else {
				// Unmapped - use replacement character
				b.WriteRune('\uFFFD')
			}
		} else {
			// 0xA0-0xFF: same as ISO Latin-1
			b.WriteRune(rune(byteVal))
		}
	}
	return b.String()
}

// DecodeMacRoman converts bytes from MacRomanEncoding to UTF-8 string.
// This is a simplified version - a complete implementation would need
// the full MacRoman character set mapping.
func DecodeMacRoman(data []byte) string {
	// For now, treat as ISO Latin-1 (not accurate but reasonable fallback)
	// A full implementation would map Mac-specific characters
	var b strings.Builder
	b.Grow(len(data))
	for _, byteVal := range data {
		b.WriteRune(rune(byteVal))
	}
	return b.String()
}
