package main

import (
	"fmt"
	"strings"

	"github.com/apex-woot/pdf-stream-engine/font"
	"github.com/apex-woot/pdf-stream-engine/streamengine"
)

func main() {
	fmt.Println("=== PDF Stream Engine - ToUnicode CMap Support ===\n")

	// Example 1: Simple extraction (default WinAnsi encoding)
	runSimpleExample()

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")

	// Example 2: Advanced extraction with ToUnicode CMap
	runAdvancedExample()
}

// runSimpleExample demonstrates basic text extraction using default WinAnsi encoding.
func runSimpleExample() {
	fmt.Println("EXAMPLE 1: Simple Text Extraction (WinAnsi encoding)")
	fmt.Println(strings.Repeat("-", 60))

	// Sample PDF content stream
	sampleStream := `
BT
/F1 12 Tf
72 720 Td
(Hello World) Tj
0 -14 Td
(This uses WinAnsi encoding) Tj
ET
`

	// Simple API - uses default WinAnsi encoding for all fonts
	text := streamengine.ExtractText([]byte(sampleStream))

	fmt.Println("Content Stream:")
	fmt.Println(sampleStream)
	fmt.Println("\nExtracted Text:")
	fmt.Println(text)
}

// runAdvancedExample demonstrates text extraction with custom font encodings
// and ToUnicode CMaps for handling CID fonts and complex encodings.
func runAdvancedExample() {
	fmt.Println("EXAMPLE 2: Advanced Text Extraction (ToUnicode CMap)")
	fmt.Println(strings.Repeat("-", 60))

	// Sample ToUnicode CMap (simplified CMap syntax)
	// This maps custom character codes to Unicode values
	sampleCMap := `
/CIDInit /ProcSet findresource begin
12 dict begin
begincmap
/CMapName /Custom-UTF16 def
/CMapType 2 def
1 begincodespacerange
<00> <FF>
endcodespacerange
4 beginbfchar
<01> <0048>
<02> <0065>
<03> <006C>
<04> <006F>
endbfchar
2 beginbfrange
<20> <7E> <0020>
<A0> <A9> <00A0>
endbfrange
endcmap
`

	// Parse the ToUnicode CMap
	cmapReader := strings.NewReader(sampleCMap)
	cmap, err := font.ParseToUnicodeCMap(cmapReader)
	if err != nil {
		fmt.Printf("Error parsing CMap: %v\n", err)
		return
	}

	// Verify the CMap parsed correctly
	fmt.Println("Sample CMap Mappings:")
	fmt.Printf("  0x01 -> U+%04X (%s)\n", 0x0048, "H")
	fmt.Printf("  0x02 -> U+%04X (%s)\n", 0x0065, "e")
	fmt.Printf("  0x03 -> U+%04X (%s)\n", 0x006C, "l")
	fmt.Printf("  0x04 -> U+%04X (%s)\n", 0x006F, "o")
	fmt.Printf("  0x20-0x7E -> ASCII range\n\n")

	// Create a font registry and register a font with the ToUnicode CMap
	fontRegistry := font.NewFontRegistry()
	// NOTE: Font names should NOT include the leading slash (/)
	// The parser strips it during parsing
	fontRegistry.RegisterWithToUnicode("CustomFont", cmap)

	// Also register a standard font with WinAnsi encoding
	fontRegistry.RegisterSimple("F1", font.EncodingWinAnsi)

	// Sample content stream using the custom font
	contentStream := `
BT
/CustomFont 12 Tf
72 720 Td
<01020304> Tj
0 -14 Td
(Regular ASCII text) Tj
ET
`

	// Extract text using the font registry
	text := streamengine.ExtractTextWithFonts([]byte(contentStream), fontRegistry)

	fmt.Println("Content Stream:")
	fmt.Println(contentStream)
	fmt.Println("\nExtracted Text:")
	fmt.Println(text)
	fmt.Println("\nHow it works:")
	fmt.Println("  1. Hex string <01020304> contains bytes [0x01, 0x02, 0x03, 0x04]")
	fmt.Println("  2. ToUnicode CMap maps each byte to Unicode:")
	fmt.Println("     0x01 -> U+0048 (H)")
	fmt.Println("     0x02 -> U+0065 (e)")
	fmt.Println("     0x03 -> U+006C (l)")
	fmt.Println("     0x04 -> U+006F (o)")
	fmt.Println("  3. Result: 'Helo' (followed by 'Regular ASCII text' on next line)")
}

// ===================================================================
// Integration with pdfcpu
// ===================================================================
//
// To use this with pdfcpu for extracting text from real PDFs:
//
// 1. Read the PDF file and get page content:
//    ```go
//    ctx, err := pdfcpu.ReadContext("document.pdf", pdfConf)
//    page := ctx.Pages[0]
//    streamDict := page.StreamDict
//    streamData, err := streamDict.Decode()
//    ```
//
// 2. Extract font resources and ToUnicode CMaps:
//    ```go
//    fontRegistry := font.NewFontRegistry()
//
//    resources := page.Resources
//    if fonts, ok := resources["Font"].(map[string]any); ok {
//        for fontName, fontObj := range fonts {
//            fontDict := fontObj.(pdfcpu.Dict)
//
//            // Check for ToUnicode CMap
//            if toUnicode, ok := fontDict["ToUnicode"]; ok {
//                cmapStream := toUnicode.(pdfcpu.StreamDict)
//                cmapData, _ := cmapStream.Decode()
//
//                // Parse the CMap
//                cmap, err := font.ParseToUnicodeCMap(bytes.NewReader(cmapData))
//                if err == nil {
//                    fontRegistry.RegisterWithToUnicode(fontName, cmap)
//                }
//            } else {
//                // No ToUnicode - use standard encoding
//                encoding := font.EncodingWinAnsi // or detect from Encoding dict
//                fontRegistry.RegisterSimple(fontName, encoding)
//            }
//        }
//    }
//    ```
//
// 3. Extract text with proper encoding:
//    ```go
//    text := streamengine.ExtractTextWithFonts(streamData, fontRegistry)
//    ```
//
// This approach ensures that:
//   - CID fonts are decoded using their ToUnicode CMaps
//   - Standard fonts use appropriate encodings (WinAnsi, MacRoman, etc.)
//   - Custom font encodings are handled correctly
//   - Multi-byte character codes are properly supported

