package streamengine

import (
	"bytes"

	"github.com/apex-woot/pdf-stream-engine/font"
	"github.com/apex-woot/pdf-stream-engine/interpreter"
)

// ExtractText processes decoded PDF content streams and extracts text.
// Input: raw decoded stream bytes from pdfcpu (after StreamDict.Decode())
// Output: extracted text string
//
// This is the simple API that uses default WinAnsi encoding for all fonts.
// For PDFs with custom font encodings or ToUnicode CMaps, use ExtractTextWithFonts.
//
// The function parses PDF content stream operators (BT, ET, Tj, TJ, Td, Tm, etc.)
// and extracts readable text while maintaining proper text positioning and reading order.
//
// Example content stream:
//
//	BT
//	/F1 12 Tf
//	1 0 0 1 72 720 Tm
//	(Hello World) Tj
//	ET
//
// Returns: "Hello World"
func ExtractText(streamData []byte) string {
	return ExtractTextWithFonts(streamData, nil)
}

// ExtractTextWithFonts processes decoded PDF content streams and extracts text
// using a custom font registry for proper encoding handling.
//
// Input:
//   - streamData: raw decoded stream bytes from pdfcpu (after StreamDict.Decode())
//   - fontRegistry: FontRegistry with registered fonts and their ToUnicode CMaps
//     If nil, a default registry with WinAnsi encoding will be used.
//
// Output: extracted text string
//
// Use this when you need to handle:
//   - Custom font encodings
//   - ToUnicode CMaps for CID fonts
//   - Multi-byte character encodings
func ExtractTextWithFonts(streamData []byte, fontRegistry *font.FontRegistry) string {
	// Create a reader from the byte slice
	reader := bytes.NewReader(streamData)

	// Create interpreter with font registry
	interp := interpreter.NewInterpreter(fontRegistry)
	if err := interp.ProcessStream(reader); err != nil {
		// Log error but still return any text that was extracted
		// This follows the graceful degradation philosophy
		return interp.GetText()
	}

	return interp.GetText()
}
