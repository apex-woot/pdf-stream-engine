package font

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// CMap represents a character code to Unicode mapping (ToUnicode CMap).
type CMap struct {
	// Mappings from character code (as hex string) to Unicode string
	mappings map[string]string
}

// NewCMap creates an empty CMap.
func NewCMap() *CMap {
	return &CMap{
		mappings: make(map[string]string),
	}
}

// ParseToUnicodeCMap parses a ToUnicode CMap stream and returns a CMap.
// ToUnicode CMaps use PostScript-like syntax with operators:
//   - beginbfchar/endbfchar: single character mappings
//   - beginbfrange/endbfrange: range mappings
func ParseToUnicodeCMap(r io.Reader) (*CMap, error) {
	cmap := NewCMap()
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords) // Token by token

	for scanner.Scan() {
		token := scanner.Text()

		switch token {
		case "beginbfchar":
			if err := cmap.parseBfChar(scanner); err != nil {
				return nil, fmt.Errorf("parsing bfchar: %w", err)
			}
		case "beginbfrange":
			if err := cmap.parseBfRange(scanner); err != nil {
				return nil, fmt.Errorf("parsing bfrange: %w", err)
			}
		}
		// Ignore other tokens (CMap headers, versions, etc.)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	return cmap, nil
}

// parseBfChar parses a beginbfchar/endbfchar section.
// Format: <srcCode> <dstUnicode>
// Example: <01> <0041>  maps byte 0x01 to Unicode U+0041 (A)
func (cm *CMap) parseBfChar(scanner *bufio.Scanner) error {
	for scanner.Scan() {
		token := scanner.Text()
		if token == "endbfchar" {
			return nil
		}

		// Expect hex string for source code
		srcCode := strings.TrimSpace(token)
		if !isHexString(srcCode) {
			continue // Skip non-hex tokens
		}

		// Next token should be destination Unicode hex string
		if !scanner.Scan() {
			return fmt.Errorf("unexpected EOF in bfchar")
		}
		dstUnicode := strings.TrimSpace(scanner.Text())
		if !isHexString(dstUnicode) {
			continue // Skip malformed entries
		}

		// Store mapping
		srcHex := stripHexBrackets(srcCode)
		dstHex := stripHexBrackets(dstUnicode)

		// Convert destination to Unicode string
		unicodeStr, err := hexToUnicodeString(dstHex)
		if err != nil {
			continue // Skip invalid Unicode
		}

		cm.mappings[srcHex] = unicodeStr
	}
	return fmt.Errorf("endbfchar not found")
}

// parseBfRange parses a beginbfrange/endbfrange section.
// Format: <srcCodeStart> <srcCodeEnd> <dstUnicodeStart>
// Example: <0020> <007E> <0020>  maps 0x20-0x7E to U+0020-U+007E
func (cm *CMap) parseBfRange(scanner *bufio.Scanner) error {
	for scanner.Scan() {
		token := scanner.Text()
		if token == "endbfrange" {
			return nil
		}

		// Expect: <start> <end> <dstStart>
		srcStart := strings.TrimSpace(token)
		if !isHexString(srcStart) {
			continue
		}

		if !scanner.Scan() {
			return fmt.Errorf("unexpected EOF in bfrange (end)")
		}
		srcEnd := strings.TrimSpace(scanner.Text())
		if !isHexString(srcEnd) {
			continue
		}

		if !scanner.Scan() {
			return fmt.Errorf("unexpected EOF in bfrange (dst)")
		}
		dstStart := strings.TrimSpace(scanner.Text())

		// Handle array form: <start> <end> [<unicode1> <unicode2> ...]
		if strings.HasPrefix(dstStart, "[") {
			// This is an array of destination values (less common)
			// For simplicity, we'll skip this for now
			// In a full implementation, we'd parse the array
			continue
		}

		if !isHexString(dstStart) {
			continue
		}

		// Parse hex values
		srcStartHex := stripHexBrackets(srcStart)
		srcEndHex := stripHexBrackets(srcEnd)
		dstStartHex := stripHexBrackets(dstStart)

		startCode, err := hexStringToInt(srcStartHex)
		if err != nil {
			continue
		}
		endCode, err := hexStringToInt(srcEndHex)
		if err != nil {
			continue
		}
		dstCode, err := hexStringToInt(dstStartHex)
		if err != nil {
			continue
		}

		// Create mappings for the range
		for code := startCode; code <= endCode; code++ {
			srcHex := fmt.Sprintf("%02x", code)
			unicodeValue := dstCode + (code - startCode)
			unicodeStr := string(rune(unicodeValue))
			cm.mappings[srcHex] = unicodeStr
		}
	}
	return fmt.Errorf("endbfrange not found")
}

// Lookup returns the Unicode string for a given character code.
// The code should be provided as raw bytes.
func (cm *CMap) Lookup(code []byte) (string, bool) {
	// Convert code to hex string (lowercase, with leading zeros)
	// Format as zero-padded hex to match the format used when storing mappings
	hexKey := ""
	for _, b := range code {
		hexKey += fmt.Sprintf("%02x", b)
	}
	unicode, ok := cm.mappings[hexKey]
	return unicode, ok
}

// LookupByte is a convenience method for single-byte codes.
func (cm *CMap) LookupByte(code byte) (string, bool) {
	return cm.Lookup([]byte{code})
}

// Helper functions

func isHexString(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "<") && strings.HasSuffix(s, ">")
}

func stripHexBrackets(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "<")
	s = strings.TrimSuffix(s, ">")
	return s
}

// hexStringToInt converts a hex string to an integer.
func hexStringToInt(hexStr string) (int, error) {
	val, err := strconv.ParseInt(hexStr, 16, 64)
	return int(val), err
}

// hexToUnicodeString converts a hex-encoded Unicode string to a Go string.
// For multi-byte Unicode (e.g., <FEFF0041> for BOM + A), this handles UTF-16BE.
func hexToUnicodeString(hexStr string) (string, error) {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}

	// If length is 2 or more bytes, treat as UTF-16BE
	if len(data) >= 2 {
		// Check for BOM (0xFEFF)
		if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
			data = data[2:] // Strip BOM
		}

		// Decode UTF-16BE to runes
		var runes []rune
		for i := 0; i < len(data); i += 2 {
			if i+1 < len(data) {
				r := rune(data[i])<<8 | rune(data[i+1])
				runes = append(runes, r)
			}
		}
		return string(runes), nil
	}

	// Single byte - treat as direct Unicode codepoint
	if len(data) == 1 {
		return string(rune(data[0])), nil
	}

	return "", fmt.Errorf("invalid Unicode hex: %s", hexStr)
}

// DecodeString decodes a byte sequence using this CMap.
// For multi-byte encodings, this attempts to find the longest matching prefix.
func (cm *CMap) DecodeString(data []byte) string {
	var result strings.Builder
	result.Grow(len(data))

	i := 0
	for i < len(data) {
		matched := false

		// Try 2-byte code
		if i+1 < len(data) {
			code := data[i : i+2]
			if unicode, ok := cm.Lookup(code); ok {
				result.WriteString(unicode)
				i += 2
				matched = true
				continue
			}
		}

		// Try 1-byte code
		if unicode, ok := cm.LookupByte(data[i]); ok {
			result.WriteString(unicode)
			i++
			matched = true
		}

		// No mapping found - output replacement character
		if !matched {
			result.WriteRune('\uFFFD')
			i++
		}
	}

	return result.String()
}

// String returns a debug representation of the CMap.
func (cm *CMap) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("CMap with %d mappings:\n", len(cm.mappings)))
	for code, unicode := range cm.mappings {
		buf.WriteString(fmt.Sprintf("  %s -> %q\n", code, unicode))
	}
	return buf.String()
}
