package stringutils

import "unicode/utf8"

// StringState tracks whether we are inside a string literal, handling
// both single and double quotes with backslash escape detection.
type StringState struct {
	InString bool
	Escaped  bool
	Quote    byte
}

func (s *StringState) Advance(ch byte) {
	if s.InString {
		if s.Escaped {
			s.Escaped = false
			return
		}
		if ch == '\\' {
			s.Escaped = true
			return
		}
		if ch == s.Quote {
			s.InString = false
			s.Quote = 0
		}
		return
	}
	if ch == '"' || ch == '\'' {
		s.InString = true
		s.Escaped = false
		s.Quote = ch
	}
}

// ScanState combines string-literal tracking with bracket depth counting.
// It tracks parentheses, brackets, and braces independently.
type ScanState struct {
	Str        StringState
	DepthParen int
	DepthBrack int
	DepthBrace int
}

// Advance processes one character, updating string state and bracket depths.
// Bracket changes are ignored when inside a string literal.
func (s *ScanState) Advance(ch byte) {
	s.Str.Advance(ch)
	if s.Str.InString {
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

// AdvanceRune updates state for a rune-oriented scanner.
// Non-ASCII runes do not affect delimiter/depth tracking.
func (s *ScanState) AdvanceRune(ch rune) {
	if ch >= 0 && ch < utf8.RuneSelf {
		s.Advance(byte(ch))
	}
}

// InString returns true if we are inside a string literal.
func (s *ScanState) InString() bool {
	return s.Str.InString
}

// Depth returns the total nesting depth across all bracket types.
func (s *ScanState) Depth() int {
	return s.DepthParen + s.DepthBrack + s.DepthBrace
}

// AtTopLevel returns true if all bracket depths are zero and we're not in a string.
func (s *ScanState) AtTopLevel() bool {
	return !s.Str.InString && s.DepthParen == 0 && s.DepthBrack == 0 && s.DepthBrace == 0
}
