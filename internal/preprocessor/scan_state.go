package preprocessor

// isEscapedAt checks if the character at position i in string s is escaped
// by counting consecutive backslashes before it. An odd number means escaped.
// Example: "hello\"" -> quote at end is escaped (1 backslash)
// Example: "hello\\"  -> quote at end is NOT escaped (2 backslashes = escaped backslash)
func isEscapedAt(s string, i int) bool {
	backslashCount := 0
	for j := i - 1; j >= 0 && s[j] == '\\'; j-- {
		backslashCount++
	}
	return backslashCount%2 == 1
}

// stringState tracks whether we are inside a string literal, handling
// both single and double quotes with backslash escape detection.
type stringState struct {
	inString bool
	escaped  bool
	quote    byte
}

func (s *stringState) Advance(ch byte) {
	if s.inString {
		if s.escaped {
			s.escaped = false
			return
		}
		if ch == '\\' {
			s.escaped = true
			return
		}
		if ch == s.quote {
			s.inString = false
			s.quote = 0
		}
		return
	}
	if ch == '"' || ch == '\'' {
		s.inString = true
		s.escaped = false
		s.quote = ch
	}
}

// ScanState combines string-literal tracking with bracket depth counting.
// It tracks parentheses, brackets, and braces independently.
type ScanState struct {
	Str        stringState
	DepthParen int
	DepthBrack int
	DepthBrace int
}

// Advance processes one character, updating string state and bracket depths.
// Bracket changes are ignored when inside a string literal.
func (s *ScanState) Advance(ch byte) {
	s.Str.Advance(ch)
	if s.Str.inString {
		return
	}
	switch ch {
	case '(':
		s.DepthParen++
	case ')':
		if s.DepthParen > 0 {
			s.DepthParen--
		}
	case '[':
		s.DepthBrack++
	case ']':
		if s.DepthBrack > 0 {
			s.DepthBrack--
		}
	case '{':
		s.DepthBrace++
	case '}':
		if s.DepthBrace > 0 {
			s.DepthBrace--
		}
	}
}

// InString returns true if we are inside a string literal.
func (s *ScanState) InString() bool {
	return s.Str.inString
}

// Depth returns the total nesting depth across all bracket types.
func (s *ScanState) Depth() int {
	return s.DepthParen + s.DepthBrack + s.DepthBrace
}

// AtTopLevel returns true if all bracket depths are zero and we're not in a string.
func (s *ScanState) AtTopLevel() bool {
	return !s.Str.inString && s.DepthParen == 0 && s.DepthBrack == 0 && s.DepthBrace == 0
}
