package streamengine

import (
	"bytes"

	"github.com/apex-woot/pdf-stream-engine/interpreter"
)

// ExtractText processes decoded PDF content streams and extracts text.
// Input: raw decoded stream bytes from pdfcpu (after StreamDict.Decode())
// Output: extracted text string
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
	// Create a reader from the byte slice
	reader := bytes.NewReader(streamData)

	// Create interpreter and process the stream
	interp := interpreter.NewInterpreter()
	if err := interp.ProcessStream(reader); err != nil {
		// Log error but still return any text that was extracted
		// This follows the graceful degradation philosophy
		return interp.GetText()
	}

	return interp.GetText()
}
