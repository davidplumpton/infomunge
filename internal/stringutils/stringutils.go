package stringutils

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// OperatorReplacement describes a pattern to find and what to replace it with.
type OperatorReplacement struct {
	Pattern     string
	Replacement []rune
}

// ReplaceOperatorsOutsideStrings is a generic operator replacement function.
// It replaces patterns only when they appear outside of quoted strings.
func ReplaceOperatorsOutsideStrings(s string, replacements []OperatorReplacement) string {
	var result []rune
	inString := false
	i := 0

	for i < len(s) {
		if s[i] == '"' && !IsEscapedAt(s, i) {
			inString = !inString
			result = append(result, '"')
			i++
			continue
		}

		if !inString {
			matched := false
			for _, repl := range replacements {
				if i+len(repl.Pattern) <= len(s) && s[i:i+len(repl.Pattern)] == repl.Pattern {
					result = append(result, repl.Replacement...)
					i += len(repl.Pattern)
					matched = true
					break
				}
			}
			if matched {
				continue
			}
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		result = append(result, r)
		i += size
	}
	return string(result)
}

// BinaryOperatorConfig defines configuration for a binary infix operator transformation.
type BinaryOperatorConfig struct {
	Operator        string
	FuncName        string
	ExtraStops      []rune
	RightStopOps    []string
	UseMinimalStops bool // If true, use MinimalStops instead of DefaultOperatorStops for left operand
}

// ReplaceBinaryOperator is a generic helper that converts "left OP right" to "funcName(left, right)".
func ReplaceBinaryOperator(s string, cfg BinaryOperatorConfig) string {
	var result []rune
	inString := false
	i := 0
	opLen := len(cfg.Operator)

	for i < len(s) {
		if s[i] == '"' && !IsEscapedAt(s, i) {
			inString = !inString
			result = append(result, '"')
			i++
			continue
		}

		if !inString && i+opLen <= len(s) && s[i:i+opLen] == cfg.Operator {
			var leftStart int
			if cfg.UseMinimalStops {
				leftStart = FindLeftOperandStartWithStops(result, MinimalStops)
			} else {
				leftStart = FindLeftOperandStart(result, cfg.ExtraStops)
			}

			if leftStart >= len(result) {
				result = append(result, []rune(cfg.Operator)...)
				i += opLen
				continue
			}

			leftOp := strings.TrimSpace(string(result[leftStart:]))
			if leftOp == "" {
				result = append(result, []rune(cfg.Operator)...)
				i += opLen
				continue
			}

			rightStart := i + opLen
			rightEnd := rightStart
			depth := 0
			inStringLocal := false

			for rightEnd < len(s) {
				ch := s[rightEnd]
				if ch == '"' && !IsEscapedAt(s, rightEnd) {
					inStringLocal = !inStringLocal
				}

				if !inStringLocal {
					if ch == '(' || ch == '[' || ch == '{' {
						depth++
					} else if ch == ')' || ch == ']' || ch == '}' {
						if depth == 0 {
							break
						}
						depth--
					} else if depth == 0 && ch == ',' {
						break
					}
					stopMatched := false
					for _, stopOp := range cfg.RightStopOps {
						if depth == 0 && rightEnd+len(stopOp) <= len(s) && s[rightEnd:rightEnd+len(stopOp)] == stopOp {
							stopMatched = true
							break
						}
					}
					if stopMatched {
						break
					}
				}
				rightEnd++
			}

			rightOp := strings.TrimSpace(s[rightStart:rightEnd])
			if rightOp == "" {
				result = append(result, []rune(cfg.Operator)...)
				i += opLen
				continue
			}

			result = result[:leftStart]
			result = append(result, []rune(cfg.FuncName+"(")...)
			result = append(result, []rune(leftOp)...)
			result = append(result, []rune(", ")...)
			result = append(result, []rune(rightOp)...)
			result = append(result, ')')
			i = rightEnd
			continue
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		result = append(result, r)
		i += size
	}

	return string(result)
}

// DefaultOperatorStops contains the default stop characters for expression parsing.
// These include arithmetic operators, comparison operators, brackets, and delimiters.
var DefaultOperatorStops = []rune{'+', '-', '*', '/', '=', '<', '>', '!', '(', '[', '{', ',', ':'}

// MinimalStops contains minimal stop characters (delimiters and brackets only).
// Used when operators should be included in the operand.
var MinimalStops = []rune{',', ';', ':', '(', '[', '{'}

// FindLeftOperandStart walks backwards through result to find where the left operand begins.
// Uses DefaultOperatorStops plus any extraStops provided.
func FindLeftOperandStart(result []rune, extraStops []rune) int {
	allStops := append(DefaultOperatorStops, extraStops...)
	return FindLeftOperandStartWithStops(result, allStops)
}

// FindLeftOperandStartWithStops walks backwards through result to find where the left operand begins.
// Only stops at the explicitly provided stop characters.
func FindLeftOperandStartWithStops(result []rune, stops []rune) int {
	pos := len(result) - 1

	// Skip trailing whitespace
	for pos >= 0 && unicode.IsSpace(result[pos]) {
		pos--
	}

	if pos < 0 {
		return 0
	}

	depth := 0
	for pos >= 0 {
		ch := result[pos]

		// Handle strings when walking back
		if ch == '"' {
			pos--
			for pos >= 0 {
				if result[pos] == '"' {
					backslashCount := 0
					for k := pos - 1; k >= 0 && result[k] == '\\'; k-- {
						backslashCount++
					}
					if backslashCount%2 == 0 {
						break // Found start of string
					}
				}
				pos--
			}
			pos--
			continue
		}

		// Check for stop characters at depth 0
		if depth == 0 {
			for _, stop := range stops {
				if ch == stop {
					return pos + 1
				}
			}
		}

		if ch == ')' || ch == ']' || ch == '}' {
			depth++
		} else if ch == '(' || ch == '[' || ch == '{' {
			depth--
			if depth < 0 {
				return 0
			}
		}
		pos--
	}

	return 0
}
