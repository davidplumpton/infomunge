package stringutils

import (
	"unicode"
	"unicode/utf8"
)

// ExpressionScanner tracks position in a string while respecting string boundaries and bracket depth.
type ExpressionScanner struct {
	s        string
	pos      int
	inString bool
	depth    int // bracket/paren depth
}

// NewExpressionScanner creates a new ExpressionScanner for the given string starting at position 0.
func NewExpressionScanner(s string) *ExpressionScanner {
	return &ExpressionScanner{s: s, pos: 0, inString: false, depth: 0}
}

// Pos returns the current position.
func (sc *ExpressionScanner) Pos() int {
	return sc.pos
}

// SetPos sets the current position without updating state.
// Use with caution - prefer Advance or Next for normal scanning.
func (sc *ExpressionScanner) SetPos(pos int) {
	sc.pos = pos
}

// Len returns the length of the underlying string.
func (sc *ExpressionScanner) Len() int {
	return len(sc.s)
}

// String returns the underlying string.
func (sc *ExpressionScanner) String() string {
	return sc.s
}

// Peek returns the character at the current position without advancing, or 0 if at end.
func (sc *ExpressionScanner) Peek() byte {
	if sc.pos < len(sc.s) {
		return sc.s[sc.pos]
	}
	return 0
}

// PeekAhead returns the character at pos + offset without advancing, or 0 if out of bounds.
func (sc *ExpressionScanner) PeekAhead(offset int) byte {
	if sc.pos+offset < len(sc.s) {
		return sc.s[sc.pos+offset]
	}
	return 0
}

// Peek2 peeks two characters ahead and returns them as a string (or shorter if at end).
func (sc *ExpressionScanner) Peek2() string {
	if sc.pos+2 <= len(sc.s) {
		return sc.s[sc.pos : sc.pos+2]
	}
	if sc.pos < len(sc.s) {
		return sc.s[sc.pos:]
	}
	return ""
}

// PeekN peeks n characters ahead and returns them as a string.
func (sc *ExpressionScanner) PeekN(n int) string {
	if sc.pos+n <= len(sc.s) {
		return sc.s[sc.pos : sc.pos+n]
	}
	if sc.pos < len(sc.s) {
		return sc.s[sc.pos:]
	}
	return ""
}

// Next advances by one position and returns the character, updating string and bracket state.
// Returns 0 if at end.
func (sc *ExpressionScanner) Next() byte {
	if sc.pos >= len(sc.s) {
		return 0
	}
	ch := sc.s[sc.pos]
	sc.updateState(ch)
	sc.pos++
	return ch
}

// NextRune advances by one rune and returns it, properly handling multi-byte UTF-8 characters.
// State updates only happen for ASCII characters (quotes, brackets).
// Returns utf8.RuneError if at end or invalid UTF-8.
func (sc *ExpressionScanner) NextRune() rune {
	if sc.pos >= len(sc.s) {
		return utf8.RuneError
	}
	r, size := utf8.DecodeRuneInString(sc.s[sc.pos:])
	// Only update state for ASCII characters that affect parsing
	if size == 1 {
		sc.updateState(sc.s[sc.pos])
	}
	sc.pos += size
	return r
}

// Advance moves forward by n positions, updating state for each character.
func (sc *ExpressionScanner) Advance(n int) {
	for i := 0; i < n && sc.pos < len(sc.s); i++ {
		sc.Next()
	}
}

// IsInString returns whether the scanner is currently inside a string.
func (sc *ExpressionScanner) IsInString() bool {
	return sc.inString
}

// Depth returns the current bracket/paren nesting depth (0 outside brackets).
func (sc *ExpressionScanner) Depth() int {
	return sc.depth
}

// updateState updates the inString and depth trackers based on a character.
func (sc *ExpressionScanner) updateState(ch byte) {
	// Handle string state
	if ch == '"' && !IsEscapedAt(sc.s, sc.pos) {
		sc.inString = !sc.inString
		return
	}

	// Update bracket depth only when not in a string
	if !sc.inString {
		switch ch {
		case '(', '[', '{':
			sc.depth++
		case ')', ']', '}':
			sc.depth--
		}
	}
}

// SkipWhitespace advances past all whitespace characters while respecting string boundaries.
func (sc *ExpressionScanner) SkipWhitespace() {
	for sc.pos < len(sc.s) && unicode.IsSpace(rune(sc.s[sc.pos])) && !sc.inString {
		sc.Next()
	}
}

// FindMatchingCloseBracket finds the position of the closing bracket that matches
// the opening bracket at startPos. Returns -1 if not found.
// Handles nested brackets and string boundaries.
func (sc *ExpressionScanner) FindMatchingCloseBracket(startPos int) int {
	if startPos >= len(sc.s) {
		return -1
	}

	ch := sc.s[startPos]
	if ch != '(' && ch != '[' && ch != '{' {
		return -1
	}

	closeBracket := byte(')')
	if ch == '[' {
		closeBracket = ']'
	} else if ch == '{' {
		closeBracket = '}'
	}

	depth := 1
	inStr := false
	i := startPos + 1

	for i < len(sc.s) {
		c := sc.s[i]

		if c == '"' && !IsEscapedAt(sc.s, i) {
			inStr = !inStr
		}

		if !inStr {
			if c == ch {
				depth++
			} else if c == closeBracket {
				depth--
				if depth == 0 {
					return i
				}
			}
		}
		i++
	}

	return -1
}

// SubstringUntilDepth returns a substring from the current position until bracket depth
// reaches the specified target depth (or end of string). Does not advance the scanner.
func (sc *ExpressionScanner) SubstringUntilDepth(targetDepth int) string {
	startPos := sc.pos
	tempDepth := sc.depth
	inStr := sc.inString
	i := sc.pos

	for i < len(sc.s) {
		ch := sc.s[i]

		if ch == '"' && !IsEscapedAt(sc.s, i) {
			inStr = !inStr
		}

		if !inStr {
			switch ch {
			case '(', '[', '{':
				tempDepth++
			case ')', ']', '}':
				tempDepth--
				if tempDepth <= targetDepth {
					return sc.s[startPos:i]
				}
			}
		}
		i++
	}

	return sc.s[startPos:]
}

// ScanUntilAtDepthZero scans forward until a stop condition is met at depth 0.
// The stopFn is called with the current character and position.
// Returns the end position (exclusive).
func (sc *ExpressionScanner) ScanUntilAtDepthZero(stopFn func(ch byte, pos int) bool) int {
	for sc.pos < len(sc.s) {
		ch := sc.s[sc.pos]
		if !sc.inString && sc.depth == 0 && stopFn(ch, sc.pos) {
			return sc.pos
		}
		sc.Next()
	}
	return sc.pos
}
