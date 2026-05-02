package preprocessor

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"infomunge/internal/stringutils"
)

// replaceReplaceOperator converts "str replace pattern with replacement" to "replace(str, pattern, replacement)"
// This handles both string patterns and regex literals (after regex literal conversion).
func replaceReplaceOperator(s string) string {
	const replaceKeyword = " replace "
	const withKeyword = " with "

	var result []rune
	inString := false
	i := 0

	for i < len(s) {
		if s[i] == '"' && !stringutils.IsEscapedAt(s, i) {
			inString = !inString
			result = append(result, '"')
			i++
			continue
		}

		// Look for " replace " keyword
		if !inString && i+len(replaceKeyword) <= len(s) && s[i:i+len(replaceKeyword)] == replaceKeyword {
			// Find the left operand
			leftStart := findLeftOperandForReplace(result)

			if leftStart >= len(result) {
				// No left operand, skip
				result = append(result, []rune(replaceKeyword)...)
				i += len(replaceKeyword)
				continue
			}

			leftOp := strings.TrimSpace(string(result[leftStart:]))
			result = result[:leftStart]

			i += len(replaceKeyword) // Skip " replace "

			// Find the pattern (until " with ")
			patternStart := i
			patternEnd := findWithKeyword(s, i)
			if patternEnd == -1 {
				// No " with " found, restore and continue
				result = append(result, []rune(leftOp)...)
				result = append(result, []rune(replaceKeyword)...)
				continue
			}

			pattern := strings.TrimSpace(s[patternStart:patternEnd])
			i = patternEnd + len(withKeyword) // Skip " with "

			// Find the replacement
			replacement, replEnd := findReplaceRightOperand(s, i)
			if replacement == "" {
				// No replacement found, restore and continue
				result = append(result, []rune(leftOp)...)
				result = append(result, []rune(replaceKeyword)...)
				result = append(result, []rune(pattern)...)
				result = append(result, []rune(withKeyword)...)
				continue
			}

			// Build the function call: replace(str, pattern, replacement)
			result = append(result, []rune("replace(")...)
			result = append(result, []rune(leftOp)...)
			result = append(result, []rune(", ")...)
			result = append(result, []rune(pattern)...)
			result = append(result, []rune(", ")...)
			result = append(result, []rune(replacement)...)
			result = append(result, ')')

			i = replEnd
			continue
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		result = append(result, r)
		i += size
	}

	return string(result)
}

// findLeftOperandForReplace finds the start of the left operand for replace.
func findLeftOperandForReplace(result []rune) int {
	if len(result) == 0 {
		return 0
	}

	// Work backwards to find the start of the operand
	i := len(result) - 1

	// Skip trailing whitespace
	for i >= 0 && unicode.IsSpace(result[i]) {
		i--
	}

	if i < 0 {
		return len(result)
	}

	// Handle closing brackets
	if result[i] == ')' || result[i] == ']' || result[i] == '}' || result[i] == '"' {
		closer := result[i]
		var opener rune
		switch closer {
		case ')':
			opener = '('
		case ']':
			opener = '['
		case '}':
			opener = '{'
		case '"':
			opener = '"'
		}

		depth := 1
		i--
		inStr := closer == '"'

		for i >= 0 && depth > 0 {
			ch := result[i]
			if ch == '"' && (i == 0 || result[i-1] != '\\') {
				if closer == '"' {
					depth--
				} else {
					inStr = !inStr
				}
			} else if !inStr {
				if ch == closer {
					depth++
				} else if ch == opener {
					depth--
				}
			}
			i--
		}

		// Now find the start of the expression
		for i >= 0 && (isIdentRune(result[i]) || result[i] == '.') {
			i--
		}

		return i + 1
	}

	// Handle identifier
	if isIdentRune(result[i]) {
		for i >= 0 && (isIdentRune(result[i]) || result[i] == '.') {
			i--
		}
		return i + 1
	}

	return len(result)
}

// findWithKeyword finds the position of " with " keyword after position start.
func findWithKeyword(s string, start int) int {
	const withKeyword = " with "
	inString := false
	depth := 0

	for i := start; i+len(withKeyword) <= len(s); i++ {
		if s[i] == '"' && !stringutils.IsEscapedAt(s, i) {
			inString = !inString
			continue
		}

		if !inString {
			ch := s[i]
			if ch == '(' || ch == '[' || ch == '{' {
				depth++
			} else if ch == ')' || ch == ']' || ch == '}' {
				depth--
			}

			if depth == 0 && s[i:i+len(withKeyword)] == withKeyword {
				return i
			}
		}
	}

	return -1
}

// findReplaceRightOperand finds the replacement value after " with ".
func findReplaceRightOperand(s string, start int) (string, int) {
	// Skip leading whitespace
	i := start
	for i < len(s) && unicode.IsSpace(rune(s[i])) {
		i++
	}

	if i >= len(s) {
		return "", start
	}

	// Parse the operand
	operandStart := i
	depth := 0
	inString := false

	for i < len(s) {
		ch := s[i]

		if ch == '"' && !stringutils.IsEscapedAt(s, i) {
			inString = !inString
			i++
			continue
		}

		if !inString {
			if ch == '(' || ch == '[' || ch == '{' {
				depth++
			} else if ch == ')' || ch == ']' || ch == '}' {
				if depth == 0 {
					break
				}
				depth--
			} else if depth == 0 {
				// Stop at operators or delimiters
				if ch == ',' || ch == '\n' {
					break
				}
				// Check for word operators
				stopOps := []string{" and ", " or ", " then ", " default ", " ++ ", " -- "}
				for _, op := range stopOps {
					if i+len(op) <= len(s) && s[i:i+len(op)] == op {
						return strings.TrimSpace(s[operandStart:i]), i
					}
				}
			}
		}

		i++
	}

	return strings.TrimSpace(s[operandStart:i]), i
}
