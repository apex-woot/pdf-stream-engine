package parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// Operation represents a PDF operator and its operands.
type Operation struct {
	Name     string
	Operands []any
}

// Parser tokenizes a PDF content stream.
// This is a simplified parser; a production-parser would need to be
// more robust, especially around string parsing and error handling.
type Parser struct {
	scanner *bufio.Scanner
}

// NewParser creates a new parser for a given reader.
func NewParser(r io.Reader) *Parser {
	scanner := bufio.NewScanner(r)
	scanner.Split(pdfTokenSplit) // Use our custom tokenizer
	return &Parser{scanner: scanner}
}

// Parse processes the entire stream and returns a list of operations.
func (p *Parser) Parse() ([]Operation, error) {
	var operations []Operation
	var operands []any
	var arrayStack [][]any // stack of arrays being built
	arrayLevel := 0

	for p.scanner.Scan() {
		token := p.scanner.Bytes()
		if len(token) == 0 {
			continue
		}

		// Check if it's an operator (alphabetic)
		if arrayLevel == 0 && isOperator(token) {
			op := Operation{
				Name:     string(token),
				Operands: make([]any, len(operands)),
			}
			copy(op.Operands, operands)
			operations = append(operations, op)
			operands = operands[:0] // Clear the operand stack
		} else {
			// It's an operand, or we are inside an array
			operand, err := parseOperand(token)
			if err != nil {
				// For now, we'll just skip bad operands
				fmt.Printf("Warning: skipping unparsable operand '%s': %v\n", string(token), err)
				continue
			}

			if s, ok := operand.(string); ok {
				if s == "[" {
					// Start new array
					arrayStack = append(arrayStack, make([]any, 0))
					arrayLevel++
					continue // Don't add "[" to operand stack
				} else if s == "]" {
					// Close current array
					if arrayLevel == 0 {
						return nil, errors.New("unexpected ']' outside of array")
					}
					arrayLevel--
					closedArray := arrayStack[len(arrayStack)-1]
					arrayStack = arrayStack[:len(arrayStack)-1] // pop

					if arrayLevel == 0 {
						// Top-level array finished, add to main operands
						operands = append(operands, closedArray)
					} else {
						// Nested array finished, add to parent array
						parentArray := arrayStack[len(arrayStack)-1]
						arrayStack[len(arrayStack)-1] = append(parentArray, closedArray)
					}
					continue // Don't add "]" to operand stack
				}
			}

			// Add operand
			if arrayLevel > 0 {
				// Add to current array
				currentArray := arrayStack[len(arrayStack)-1]
				arrayStack[len(arrayStack)-1] = append(currentArray, operand)
			} else {
				// Add to main operand stack
				operands = append(operands, operand)
			}
		}
	}

	if err := p.scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	if arrayLevel > 0 {
		return nil, errors.New("unclosed array at end of stream")
	}

	return operations, nil
}

// isOperator checks if a token is a PDF operator.
// This is a simplification: valid operators can contain '*' or "'"
func isOperator(token []byte) bool {
	if len(token) == 0 {
		return false
	}
	// Check if all characters are letters (or special operator chars)
	for _, b := range token {
		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '*' || b == '\'' {
			continue
		}
		return false // Not a simple operator
	}
	return true
}

// parseOperand converts a token into a Go type.
func parseOperand(token []byte) (any, error) {
	if len(token) == 0 {
		return nil, errors.New("empty token")
	}

	switch token[0] {
	case '(':
		return parseLiteralString(token)
	case '<':
		if len(token) > 1 && token[1] == '<' {
			// Dictionary token, e.g., <</MCID 0>>
			// For text extraction, we don't need to parse its contents,
			// so we just return it as a string to be consumed by an operator.
			return string(token), nil
		}
		return parseHexString(token) // This now returns ([]byte, error)
	case '/':
		return string(token[1:]), nil // Name
	case '[':
		// This case is now handled by the logic below
	case ']':
		// This case is now handled by the logic below
	default:
		// Try to parse as a number (float or int)
		if f, err := strconv.ParseFloat(string(token), 64); err == nil {
			return f, nil
		}

		sToken := string(token)
		if sToken == "[" {
			return "[", nil // Array start marker
		}
		if sToken == "]" {
			return "]", nil // Array end marker
		}

		// If not a number, it might be an inline operator we missed,
		// but for operands, we'll error out.
		return nil, fmt.Errorf("unrecognized operand type: %s", string(token))
	}
	// Fallback for cases like '[' and ']' which are now handled in default
	sToken := string(token)
	if sToken == "[" {
		return "[", nil
	}
	if sToken == "]" {
		return "]", nil
	}
	// This should be unreachable, but as a safeguard:
	return nil, fmt.Errorf("unrecognized operand type: %s", string(token))
}

// parseLiteralString handles (string) with escapes.
func parseLiteralString(token []byte) (string, error) {
	if len(token) < 2 || token[0] != '(' || token[len(token)-1] != ')' {
		return "", fmt.Errorf("invalid literal string: %s", string(token))
	}
	// Trim parens
	s := token[1 : len(token)-1]
	var b strings.Builder
	escaping := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if escaping {
			switch c {
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case 'b':
				b.WriteByte('\b')
			case 'f':
				b.WriteByte('\f')
			case '(', ')', '\\':
				b.WriteByte(c)
			default:
				// Octal escape (e.g., \123)
				if c >= '0' && c <= '7' {
					octal := []byte{c}
					j := 1
					for j < 3 && i+j < len(s) && s[i+j] >= '0' && s[i+j] <= '7' {
						octal = append(octal, s[i+j])
						j++
					}
					i += (j - 1)
					val, _ := strconv.ParseInt(string(octal), 8, 32)
					b.WriteByte(byte(val))
				} else {
					// Ignored escape (e.g. \g)
				}
			}
			escaping = false
		} else if c == '\\' {
			escaping = true
		} else {
			b.WriteByte(c)
		}
	}
	return b.String(), nil
}

// parseHexString handles <hexstring>.
func parseHexString(token []byte) ([]byte, error) {
	if len(token) < 2 || token[0] != '<' || token[len(token)-1] != '>' {
		return nil, fmt.Errorf("invalid hex string: %s", string(token))
	}
	// Trim angle brackets
	s := token[1 : len(token)-1]

	// PDF hex strings can contain whitespace, so let's strip it.
	s = bytes.ReplaceAll(s, []byte(" "), []byte(""))
	s = bytes.ReplaceAll(s, []byte("\n"), []byte(""))
	s = bytes.ReplaceAll(s, []byte("\r"), []byte(""))
	s = bytes.ReplaceAll(s, []byte("\t"), []byte(""))

	// Ensure even length, pad with 0 if not
	if len(s)%2 != 0 {
		s = append(s, '0')
	}

	var b bytes.Buffer
	for i := 0; i < len(s); i += 2 {
		hexByte := string(s[i : i+2])
		val, err := strconv.ParseUint(hexByte, 16, 8)
		if err != nil {
			// PDF spec says to ignore bad hex chars, but we'll be strict
			return nil, fmt.Errorf("invalid hex byte '%s': %w", hexByte, err)
		}
		b.WriteByte(byte(val))
	}
	// Note: This returns the raw bytes.
	// The interpreter will need to handle encoding.
	return b.Bytes(), nil
}

// pdfTokenSplit is a custom bufio.SplitFunc for PDF content streams.
// This is a *major* simplification. A real implementation is much more complex.
func pdfTokenSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	start := 0
	// Skip leading whitespace and comments
	for start < len(data) {
		r := rune(data[start])
		if unicode.IsSpace(r) {
			start++
			continue
		}

		if r == '%' {
			// Skip comment
			start++ // consume the '%'
			for start < len(data) {
				if data[start] == '\n' || data[start] == '\r' {
					break // Stop at newline
				}
				start++
			}
			continue // Re-run the loop to skip whitespace/more comments
		}

		break // Found start of a token
	}

	if start == len(data) {
		return start, nil, nil // Need more data
	}

	pos := start
	switch data[pos] {
	case '(': // Literal String
		parenLevel := 1
		pos++
		escaping := false
		for pos < len(data) {
			b := data[pos]
			if escaping {
				escaping = false
			} else if b == '\\' {
				escaping = true
			} else if b == '(' {
				parenLevel++
			} else if b == ')' {
				parenLevel--
				if parenLevel == 0 {
					pos++
					return pos, data[start:pos], nil
				}
			}
			pos++
		}
	case '<': // Hex String or Dictionary
		if pos+1 < len(data) && data[pos+1] == '<' {
			// Dictionary start: <<
			pos += 2
			dictLevel := 1
			for pos < len(data) {
				// This is a simplified dictionary tokenizer.
				// It doesn't handle nested strings, arrays, etc.
				// but will work for simple dictionaries.
				if pos+1 < len(data) {
					if data[pos] == '>' && data[pos+1] == '>' {
						dictLevel--
						pos += 2
						if dictLevel == 0 {
							return pos, data[start:pos], nil
						}
						continue
					} else if data[pos] == '<' && data[pos+1] == '<' {
						dictLevel++
						pos += 2
						continue
					}
				}
				// Handle nested literal strings to avoid premature stops
				if data[pos] == '(' {
					parenLevel := 1
					pos++
					escaping := false
					for pos < len(data) {
						b := data[pos]
						if escaping {
							escaping = false
						} else if b == '\\' {
							escaping = true
						} else if b == '(' {
							parenLevel++
						} else if b == ')' {
							parenLevel--
							if parenLevel == 0 {
								break
							}
						}
						pos++
					}
				}
				pos++
			}
		} else {
			// Hex String start: <
			pos++
			for pos < len(data) {
				b := data[pos]
				if b == '>' {
					pos++
					return pos, data[start:pos], nil
				}
				// Allow whitespace and hex chars
				if (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F') || unicode.IsSpace(rune(b)) {
					pos++
				} else {
					// Invalid char, treat as end of token
					return pos, data[start:pos], nil
				}
			}
		}
	case '[': // Array Start
		return pos + 1, data[start : pos+1], nil
	case ']': // Array End
		return pos + 1, data[start : pos+1], nil
	case '/': // Name
		pos++
		for pos < len(data) {
			if unicode.IsSpace(rune(data[pos])) || isDelimiter(data[pos]) {
				return pos, data[start:pos], nil
			}
			pos++
		}
	default: // Number or Operator
		for pos < len(data) {
			if unicode.IsSpace(rune(data[pos])) || isDelimiter(data[pos]) {
				return pos, data[start:pos], nil
			}
			pos++
		}
	}

	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}

	return start, nil, nil // Need more data
}

func isDelimiter(b byte) bool {
	switch b {
	case '(', ')', '<', '>', '[', ']', '{', '}', '/', '%':
		return true
	default:
		return false
	}
}
