package interpreter

// TextState holds the current state relevant to text rendering.
// A full implementation would include matrices, spacing, and more.
type TextState struct {
	FontName string
	FontSize float64
	LastY    float64 // Track the last Y position
	// We would also track TextMatrix, LineMatrix, WordSpacing, CharSpacing, etc.
}

// NewTextState creates a new, default text state.
func NewTextState() TextState {
	return TextState{
		FontName: "default",
		FontSize: 1.0,
		LastY:    0,
	}
}

// Copy creates a deep copy of the TextState.
func (ts TextState) Copy() TextState {
	return TextState{
		FontName: ts.FontName,
		FontSize: ts.FontSize,
		LastY:    ts.LastY,
	}
}
