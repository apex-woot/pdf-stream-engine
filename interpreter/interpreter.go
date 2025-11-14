package interpreter

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/apex-woot/pdf-stream-engine/parser"
)

// Interpreter processes a stream of PDF operations.
type Interpreter struct {
	parser       *parser.Parser
	textBuilder  strings.Builder
	inTextObject bool
	textState    TextState
	stateStack   []TextState // For q/Q operators
}

// NewInterpreter creates a new interpreter.
func NewInterpreter() *Interpreter {
	return &Interpreter{
		textBuilder:  strings.Builder{},
		inTextObject: false,
		textState:    NewTextState(),
		stateStack:   make([]TextState, 0),
	}
}

// ProcessStream reads from an io.Reader, parses the content stream,
// and interprets the operations.
func (interp *Interpreter) ProcessStream(r io.Reader) error {
	interp.parser = parser.NewParser(r)
	operations, err := interp.parser.Parse()
	if err != nil {
		return fmt.Errorf("parser failed: %w", err)
	}

	for _, op := range operations {
		if err := interp.processOperation(op); err != nil {
			// Log warnings but continue processing
			log.Printf("Warning: error processing op '%s': %v", op.Name, err)
		}
	}
	return nil
}

// GetText returns the accumulated text extracted from the stream.
func (interp *Interpreter) GetText() string {
	// Trim leading/trailing whitespace and normalize newlines
	s := strings.TrimSpace(interp.textBuilder.String())
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return s
}

// processOperation handles a single PDF operation.
func (interp *Interpreter) processOperation(op parser.Operation) error {
	// Text can only be drawn inside a BT/ET block.
	if !interp.inTextObject && isTextShowingOp(op.Name) {
		return fmt.Errorf("text showing op '%s' outside BT/ET block", op.Name)
	}

	switch op.Name {
	// --- Graphics State ---
	case "q":
		// Save graphics state
		interp.stateStack = append(interp.stateStack, interp.textState.Copy())
	case "Q":
		// Restore graphics state
		if len(interp.stateStack) == 0 {
			return errors.New("unbalanced 'Q' operator")
		}
		interp.textState = interp.stateStack[len(interp.stateStack)-1]
		interp.stateStack = interp.stateStack[:len(interp.stateStack)-1]

	// --- Text Object ---
	case "BT":
		interp.inTextObject = true
		// Reset text matrices (not fully implemented)
		interp.textState = NewTextState()
		interp.textState.LastY = 0 // Assume start at Y=0
	case "ET":
		interp.inTextObject = false

	// --- Text State ---
	case "Tf":
		// Set font and size. e.g., /F1 12 Tf
		if len(op.Operands) < 2 {
			return fmt.Errorf("Tf expects 2 operands, got %d", len(op.Operands))
		}
		fontName, ok := op.Operands[0].(string)
		if !ok {
			return fmt.Errorf("Tf font name not a string")
		}
		fontSize, err := operandToFloat(op.Operands[1])
		if err != nil {
			return fmt.Errorf("Tf font size not a number")
		}
		interp.textState.FontName = fontName
		interp.textState.FontSize = fontSize

	// --- Text Showing ---
	case "Tj":
		// Show text
		if len(op.Operands) < 1 {
			return fmt.Errorf("Tj expects 1 operand, got %d", len(op.Operands))
		}
		if err := interp.showText(op.Operands[0]); err != nil {
			return fmt.Errorf("Tj: %w", err)
		}

	case "TJ":
		// Show text with spacing
		if len(op.Operands) < 1 {
			return fmt.Errorf("TJ expects 1 operand, got %d", len(op.Operands))
		}
		arr, ok := op.Operands[0].([]any)
		if !ok {
			return fmt.Errorf("TJ operand not an array")
		}
		for _, val := range arr {
			if str, ok := val.(string); ok {
				if err := interp.showText(str); err != nil {
					return fmt.Errorf("TJ: %w", err)
				}
			} else if b, ok := val.([]byte); ok {
				if err := interp.showText(b); err != nil {
					return fmt.Errorf("TJ: %w", err)
				}
			} else if num, ok := val.(float64); ok {
				// A number indicates a spacing adjustment.
				// We are just extracting text, so we ignore it.
				// A more advanced layout engine would use this.
				_ = num // (silence linter)
			}
		}

	case "T*":
		// Move to start of next line
		interp.textBuilder.WriteString("\n")
		// Simulate a line break (font size is a decent guess)
		interp.textState.LastY -= interp.textState.FontSize

	// --- Other common ops to ignore gracefully ---
	case "Tm": // Set text matrix [a b c d e f]
		if len(op.Operands) < 6 {
			break // Ignore malformed op
		}
		if f, err := operandToFloat(op.Operands[5]); err == nil {
			// Check if Y position (f) has changed significantly
			if math.Abs(f-interp.textState.LastY) > interp.textState.FontSize*0.5 {
				interp.textBuilder.WriteString("\n")
			}
			interp.textState.LastY = f
		}
	case "Td": // Move text position [tx ty]
		if len(op.Operands) < 2 {
			break // Ignore malformed op
		}
		if tx, err := operandToFloat(op.Operands[0]); err == nil {
			if ty, err := operandToFloat(op.Operands[1]); err == nil {
				if ty != 0 {
					// Vertical move
					interp.textBuilder.WriteString("\n")
					interp.textState.LastY += ty
				} else if tx > 1.0 { // Arbitrary "space" threshold
					// Horizontal move that's not kerning
					interp.textBuilder.WriteString(" ")
				}
			}
		}
	case "TD": // Move text position and set leading
		if len(op.Operands) < 2 {
			break // Ignore malformed op
		}
		if tx, err := operandToFloat(op.Operands[0]); err == nil {
			if ty, err := operandToFloat(op.Operands[1]); err == nil {
				if ty != 0 {
					// Vertical move
					interp.textBuilder.WriteString("\n")
					interp.textState.LastY += ty
				} else if tx > 1.0 { // Arbitrary "space" threshold
					// Horizontal move that's not kerning
					interp.textBuilder.WriteString(" ")
				}
			}
		}
	case "rg": // Set fill color (non-stroking)
	case "RG": // Set stroke color
	case "g": // Set fill gray
	case "G": // Set stroke gray
	case "Tc": // Set character spacing
	case "Tw": // Set word spacing
	case "re": // Append rectangle
	case "W": // Set clipping path
	case "n": // End path
	case "gs": // Set graphics state
	case "cm": // Concatenate matrix
	case "Do": // Draw XObject (e.g., image)
		// We ignore these as we only care about text content

	default:
		// log.Printf("Ignoring unhandled operator: %s", op.Name)
	}
	return nil
}

// showText is a helper to append text.
// It handles simple string/byte conversion and encoding.
func (interp *Interpreter) showText(val any) error {
	switch s := val.(type) {
	case string:
		// This comes from a Literal String ( ... )
		// We assume it's mostly OK, but a real parser would
		// decode octal escapes here.
		interp.textBuilder.WriteString(s)
	case []byte:
		// This comes from a Hex String < ... >
		// We must decode it from WinAnsiEncoding.
		interp.textBuilder.WriteString(DecodeWinAnsi(s))
	default:
		// This will catch operands that are not text, e.g., numbers.
		return fmt.Errorf("operand not a string or []byte, got %T", val)
	}
	return nil
}

func isTextShowingOp(opName string) bool {
	switch opName {
	case "Tj", "TJ", "'", "\"":
		return true
	default:
		return false
	}
}

// Helper to convert an operand to float64, with a fallback.
func operandToFloat(val any) (float64, error) {
	if f, ok := val.(float64); ok {
		return f, nil
	}
	if i, ok := val.(int); ok {
		return float64(i), nil
	}
	if s, ok := val.(string); ok {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f, nil
		}
	}
	return 0, fmt.Errorf("cannot convert %v to float", val)
}
