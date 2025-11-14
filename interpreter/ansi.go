package interpreter

import (
	"strings"
)

// winAnsiToUnicode maps common WinAnsiEncoding bytes (above 127) to Unicode runes.
// This is not a complete map, but covers common punctuation.
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
	0x9F: '\u0178', // Latin Capital Letter Y with Diaeresis
}

// DecodeWinAnsi converts a byte slice (from a PDF hex string)
// into a proper UTF-8 Go string, handling common WinAnsi characters.
func DecodeWinAnsi(data []byte) string {
	var b strings.Builder
	b.Grow(len(data))
	for _, byteVal := range data {
		if byteVal < 128 {
			// It's standard ASCII
			b.WriteByte(byteVal)
		} else {
			// Look up in our WinAnsi map
			if r, ok := winAnsiToUnicode[byteVal]; ok {
				b.WriteRune(r)
			} else {
				// Unknown byte, replace with Unicode Replacement Character
				b.WriteRune('\uFFFD') // ''
			}
		}
	}
	return b.String()
}
