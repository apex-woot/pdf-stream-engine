# CLAUDE.md - AI Assistant Guide for pdf-stream-engine

## Project Overview

**pdf-stream-engine** is a Go-based PDF content stream parser and interpreter designed to extract text content from PDF files. The project focuses on parsing PDF content streams (the low-level drawing commands within PDF files) and interpreting text-showing operations to extract readable text.

**Repository**: `github.com/apex-woot/pdf-stream-engine`
**Language**: Go 1.25.4
**Current State**: Early development (bootstrap phase)

## Project Purpose

This engine processes PDF content streams which contain drawing instructions (operators and operands). It specifically focuses on:
- Tokenizing and parsing PDF content stream syntax
- Interpreting text-showing operators (Tj, TJ, etc.)
- Maintaining text state (fonts, positioning, matrices)
- Extracting human-readable text from complex PDF instructions
- Handling PDF encoding schemes (WinAnsi, etc.)

## Architecture Overview

### High-Level Structure

```
pdf-stream-engine/
├── main.go              # Entry point with sample demonstration
├── go.mod               # Go module definition
├── parser/              # PDF content stream parsing
│   └── parser.go        # Tokenizer and parser implementation
└── interpreter/         # PDF operation interpretation
    ├── interpreter.go   # Main interpreter logic
    ├── textstate.go     # Text state management
    └── ansi.go          # WinAnsi encoding handling
```

### Component Responsibilities

#### 1. Parser Package (`parser/`)

**Location**: `/home/user/pdf-stream-engine/parser/parser.go`

**Purpose**: Low-level tokenization and parsing of PDF content streams

**Key Types**:
- `Operation`: Represents a PDF operator with its operands
  ```go
  type Operation struct {
      Name     string   // e.g., "Tj", "Tm", "BT"
      Operands []any    // Operands (strings, numbers, arrays)
  }
  ```
- `Parser`: Main parsing engine with custom tokenizer

**Capabilities**:
- Tokenizes PDF content streams using `pdfTokenSplit` (custom `bufio.SplitFunc`)
- Handles complex PDF syntax:
  - Literal strings: `(text)` with escape sequences
  - Hex strings: `<68656c6c6f>`
  - Names: `/FontName`
  - Numbers: integers and floats
  - Arrays: `[item1 item2 item3]`
  - Dictionaries: `<</Key Value>>`
  - Comments: `% comment text`
- Manages nested arrays and proper operator/operand pairing
- Strips whitespace and comments automatically

**Important Functions**:
- `NewParser(r io.Reader)`: Creates parser from any reader
- `Parse()`: Returns slice of all operations in stream
- `parseOperand(token []byte)`: Converts token to Go type
- `parseLiteralString(token []byte)`: Handles string escapes
- `parseHexString(token []byte)`: Decodes hex strings to bytes
- `pdfTokenSplit()`: Custom tokenizer for PDF syntax

#### 2. Interpreter Package (`interpreter/`)

**Location**: `/home/user/pdf-stream-engine/interpreter/`

**Purpose**: Interprets PDF operations and extracts text content

**Key Types**:
- `Interpreter`: Main execution engine
  ```go
  type Interpreter struct {
      parser       *parser.Parser
      textBuilder  strings.Builder  // Accumulated text output
      inTextObject bool             // Whether inside BT/ET block
      textState    TextState        // Current text rendering state
      stateStack   []TextState      // For q/Q state save/restore
  }
  ```
- `TextState`: Maintains current text rendering state
  ```go
  type TextState struct {
      FontName string
      FontSize float64
      LastY    float64  // Track Y position for line breaks
  }
  ```

**Capabilities**:
- Processes PDF operations sequentially
- Enforces PDF rules (e.g., text operators only inside BT/ET)
- Maintains graphics state stack (q/Q operators)
- Tracks text positioning for intelligent line breaking
- Handles multiple text encodings (WinAnsi)
- Gracefully handles errors (logs warnings, continues processing)

**Supported PDF Operators**:

| Operator | Description | Implementation |
|----------|-------------|----------------|
| `BT` | Begin Text Object | Sets `inTextObject = true`, resets text state |
| `ET` | End Text Object | Sets `inTextObject = false` |
| `Tf` | Set Font & Size | Updates `textState.FontName` and `FontSize` |
| `Tj` | Show Text | Appends text to builder |
| `TJ` | Show Text Array | Iterates array, shows text, ignores spacing numbers |
| `T*` | Next Line | Adds newline, adjusts Y position |
| `Tm` | Set Text Matrix | Detects Y position changes for line breaks |
| `Td`, `TD` | Move Text Position | Adds newlines or spaces based on movement |
| `q` | Save Graphics State | Pushes current state to stack |
| `Q` | Restore Graphics State | Pops state from stack |
| Others | Graphics ops | Gracefully ignored (rg, RG, g, G, re, W, n, gs, cm, Do, etc.) |

**Important Functions**:
- `NewInterpreter()`: Creates new interpreter instance
- `ProcessStream(r io.Reader)`: Main entry point, processes entire stream
- `GetText()`: Returns extracted text with normalized whitespace
- `processOperation(op Operation)`: Handles individual operations
- `showText(val any)`: Appends text (handles string/[]byte types)
- `DecodeWinAnsi(data []byte)`: Converts WinAnsi bytes to UTF-8

#### 3. Main Package (`main.go`)

**Location**: `/home/user/pdf-stream-engine/main.go`

**Purpose**: Demonstrates usage with real PDF content stream sample

**Features**:
- Contains realistic PDF content stream from actual PDF
- Shows complete usage pattern
- Demonstrates text extraction workflow

## Code Conventions and Patterns

### 1. Error Handling Philosophy

**Graceful Degradation**: The interpreter logs warnings but continues processing when encountering errors. This is intentional for PDF parsing where some malformed content shouldn't stop the entire extraction.

```go
// In interpreter.go:60
for _, op := range operations {
    if err := interp.processOperation(op); err != nil {
        log.Printf("Warning: error processing op '%s': %v", op.Name, err)
        // Continue processing, don't return error
    }
}
```

**When to use this pattern**: For non-critical errors during content processing.

**When NOT to use this pattern**: For critical errors like parser initialization failures.

### 2. Type Assertions Pattern

The codebase uses type assertions extensively because PDF operands can be various types (`string`, `[]byte`, `float64`, `[]any`).

```go
// Common pattern in interpreter.go
if str, ok := val.(string); ok {
    // handle string
} else if b, ok := val.([]byte); ok {
    // handle bytes
} else if num, ok := val.(float64); ok {
    // handle number
}
```

**Best Practice**: Always use the comma-ok idiom for type assertions to avoid panics.

### 3. String Building

Use `strings.Builder` for efficient string accumulation (see `interpreter.go:18`).

```go
// Good
textBuilder strings.Builder
textBuilder.WriteString("text")

// Avoid (inefficient)
text := ""
text += "more text"  // Creates new strings repeatedly
```

### 4. State Management

The interpreter uses a stack-based approach for state management, mirroring PDF's graphics state model:

```go
// Save state (q operator)
stateStack = append(stateStack, textState.Copy())

// Restore state (Q operator)
textState = stateStack[len(stateStack)-1]
stateStack = stateStack[:len(stateStack)-1]
```

**Important**: Always create deep copies when saving state (see `textstate.go:22`).

### 5. Custom Tokenization

The parser uses a custom `bufio.SplitFunc` for tokenization instead of regex or manual parsing. This is efficient and handles PDF's complex syntax rules.

**Location**: `parser.go:269` (`pdfTokenSplit` function)

**Key Aspects**:
- Handles nested parentheses in literal strings
- Tracks escape sequences
- Manages dictionary nesting
- Strips comments inline

## Development Workflow

### Building and Running

```bash
# Build the project
go build -o pdf-stream-engine .

# Run the demo
go run main.go

# Run with Go
./pdf-stream-engine
```

### Testing Strategy

Currently, there are no formal tests. When adding tests:

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./parser
go test ./interpreter
```

### Adding New PDF Operators

To add support for a new PDF operator:

1. **Identify the operator** in PDF specification
2. **Add case to switch statement** in `interpreter.go:67` (`processOperation` method)
3. **Extract operands** with appropriate type assertions
4. **Update state or text** as needed
5. **Handle errors** gracefully

Example:
```go
case "Ts":  // Set text rise
    if len(op.Operands) < 1 {
        return fmt.Errorf("Ts expects 1 operand, got %d", len(op.Operands))
    }
    rise, err := operandToFloat(op.Operands[0])
    if err != nil {
        return fmt.Errorf("Ts: invalid rise value")
    }
    interp.textState.Rise = rise
```

### Extending the Parser

To support new token types:

1. **Modify `pdfTokenSplit`** in `parser.go:269` to recognize the token
2. **Add parsing logic** in `parseOperand` function at `parser.go:132`
3. **Update `Operation.Operands`** type handling in interpreter

## Common AI Assistant Tasks

### Task 1: Debugging Text Extraction Issues

**Problem**: Text not being extracted correctly or appearing in wrong order.

**Investigation Steps**:
1. Check if text operators are inside BT/ET blocks (`interpreter.go:63`)
2. Verify text positioning logic in Tm/Td/TD handlers (`interpreter.go:149-191`)
3. Examine Y-position tracking for line breaks (`textstate.go:8`)
4. Test with sample stream in `main.go:12`

**Key Files**: `interpreter/interpreter.go:149-191`, `interpreter/textstate.go`

### Task 2: Adding Support for New Encodings

**Current State**: Only WinAnsi encoding is supported (`ansi.go`).

**To Add New Encoding**:
1. Create new file in `interpreter/` (e.g., `utf16.go`)
2. Implement decoder function similar to `DecodeWinAnsi`
3. Add encoding detection logic in `showText` method
4. Update hex string handling in `interpreter.go:221-224`

**Reference**: `interpreter/ansi.go:39-57`

### Task 3: Improving Parser Robustness

**Common Issues**:
- Nested structures (dictionaries in arrays, etc.)
- Malformed strings
- Binary data in streams

**Improvement Areas**:
1. **Error recovery**: `parser.go:59-62` (currently skips bad operands)
2. **Dictionary parsing**: `parser.go:324-369` (simplified implementation)
3. **String escape handling**: `parser.go:185-233`

### Task 4: Performance Optimization

**Hot Paths**:
1. `pdfTokenSplit` function (`parser.go:271`) - called for every token
2. `processOperation` switch statement (`interpreter.go:67`) - called for every operation
3. String building (`interpreter.go:18`) - already optimized with `strings.Builder`

**Optimization Opportunities**:
- Pool allocations for common types
- Reduce string conversions
- Batch text appends

## Key Design Decisions

### 1. Why Separate Parser and Interpreter?

**Rationale**: Clean separation of concerns
- **Parser**: Syntax-level concerns (tokenization, structure)
- **Interpreter**: Semantic-level concerns (meaning, text extraction)

This allows:
- Reusing parser for different purposes (validation, transformation)
- Testing each component independently
- Easier maintenance

### 2. Why `any` Type for Operands?

PDF operands can be many types (strings, numbers, arrays, dictionaries). Using `[]any` provides flexibility while maintaining type safety through type assertions.

**Alternative Considered**: Type-specific operand slices would require complex union types.

### 3. Why Continue on Errors?

PDF files from the real world often contain minor issues. Stopping on first error would make the tool impractical. The approach:
- Log warnings for debugging
- Continue processing to extract as much text as possible
- Return fatal errors only for parser-level failures

## File Reference Quick Guide

| File | Lines | Primary Purpose |
|------|-------|-----------------|
| `main.go` | 174 | Entry point, sample demonstration |
| `parser/parser.go` | 424 | PDF content stream tokenization and parsing |
| `interpreter/interpreter.go` | 256 | Operation interpretation and text extraction |
| `interpreter/textstate.go` | 29 | Text rendering state management |
| `interpreter/ansi.go` | 58 | WinAnsi encoding conversion |
| `go.mod` | 4 | Go module definition |
| `.gitignore` | 39 | Git ignore rules |

## Important Implementation Notes

### Text Positioning Logic

The interpreter attempts to add line breaks intelligently:

1. **Vertical movement** (Tm, Td, TD with Y change > 0.5 * font size) → Add `\n`
2. **Horizontal movement** (Td, TD with X > 1.0) → Add space
3. **T*** operator → Explicit newline

**Threshold Values**: See `interpreter.go:155` and `interpreter.go:170`

These are heuristics and may need tuning based on specific PDF files.

### Array Handling

Arrays can contain mixed types and are used heavily in TJ (show text array) operations:

```go
// Example from TJ operation
arr := [("Hello") -250 ("World")]
// "Hello" and "World" are text
// -250 is spacing adjustment (ignored for simple extraction)
```

**Implementation**: `interpreter.go:121-140`

### State Stack

Graphics state saving/restoring is crucial for maintaining correct text properties across complex PDF structures.

**Key Files**: `interpreter.go:69-78`

## Git Workflow

**Current Branch**: `claude/claude-md-mhyq3s7uyp948xpx-01E6UB9R5msnPHUBxXYbz5fw`

**Recent Commits**:
- `a2fba9e` - chore: add gitignore
- `00ad4a4` - feat: init bootstrap

**Development Practices**:
1. Feature work should be done on feature branches
2. Commit messages should follow conventional commits format
3. Use descriptive commit messages that explain the "why"

## Future Enhancement Areas

Based on current implementation, these areas could benefit from expansion:

1. **Text Matrix Support**: Full 6-parameter matrix operations (currently simplified)
2. **Additional Encodings**: UTF-16, PDFDocEncoding, custom encodings
3. **Font Metrics**: For accurate text positioning
4. **Advanced Layout**: Column detection, reading order determination
5. **Image Handling**: Currently ignored (Do operator)
6. **Unit Tests**: No tests exist yet
7. **Benchmarks**: Performance measurement suite
8. **Documentation**: API docs with examples
9. **Error Types**: Structured error types instead of strings
10. **Stream Filters**: Support for compressed streams (Flate, LZW, etc.)

## Dependencies

**Current**: None (uses only Go standard library)

**Potential Future Needs**:
- Compression libraries (for stream filters)
- PDF reference parser (for indirect object resolution)
- OCR libraries (for image-based text)

## Troubleshooting Guide

### Issue: Text appears garbled

**Check**:
1. Encoding handling in `ansi.go` - may need different encoding
2. Font encoding in PDF (font may use custom encoding)
3. Hex string parsing in `parser.go:236`

### Issue: Missing text

**Check**:
1. Text showing operators are recognized (`interpreter.go:232`)
2. Text is inside BT/ET block (`interpreter.go:63`)
3. Operators aren't being skipped due to parse errors (check logs)

### Issue: Wrong text order

**Check**:
1. Text positioning logic in Tm/Td handlers (`interpreter.go:149-191`)
2. Y-position threshold values (`interpreter.go:155`)
3. Whether PDF uses unusual text ordering (may need layout analysis)

### Issue: Parser fails

**Check**:
1. Token split logic for specific construct (`parser.go:271`)
2. Nested structure handling (dictionaries, arrays, strings)
3. Error messages for specific token type

## Best Practices for AI Assistants

When working on this codebase:

1. **Always check both parser AND interpreter** when debugging text issues
2. **Test with the sample in main.go** after changes
3. **Preserve error handling philosophy** (log, don't fail)
4. **Add test cases** when fixing bugs (even though none exist yet)
5. **Document PDF spec references** when implementing operators
6. **Consider real-world PDFs** which may violate spec
7. **Use type assertions carefully** with comma-ok idiom
8. **Keep parser and interpreter concerns separate**
9. **Update this CLAUDE.md** when making architectural changes

## Learning Resources

- **PDF Reference 1.7**: Adobe's official PDF specification
- **PDF 32000-1:2008**: ISO standard for PDF
- Specifically useful sections:
  - Section 8.2: Graphics Objects
  - Section 9: Text
  - Section 9.4: Text Showing Operators
  - Appendix D: Character Sets and Encodings

## Quick Command Reference

```bash
# Build
go build

# Run
go run main.go

# Format code
go fmt ./...

# Vet code
go vet ./...

# Run tests (when they exist)
go test ./...

# Build for different platforms
GOOS=linux GOARCH=amd64 go build
GOOS=windows GOARCH=amd64 go build
GOOS=darwin GOARCH=amd64 go build
```

---

**Last Updated**: 2025-11-14
**Document Version**: 1.0.0
**Codebase State**: Early development (bootstrap)
