package preprocessor

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// replaceRegexLiterals converts /pattern/ regex literals to string literals "pattern".
// This must run early in the pipeline before other operators try to parse the slashes.
//
// Context rules for distinguishing regex from division:
// - After operators, open brackets, colon, comma → regex
// - After identifier or closing bracket → division
// - At start of expression → regex
func replaceRegexLiterals(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	i := 0
	inString := false
	inSingleQuoteString := false

	for i < len(s) {
		ch := s[i]

		// Track string state
		if ch == '"' && !inSingleQuoteString && (i == 0 || s[i-1] != '\\') {
			inString = !inString
			result.WriteByte(ch)
			i++
			continue
		}

		if ch == '\'' && !inString && (i == 0 || s[i-1] != '\\') {
			inSingleQuoteString = !inSingleQuoteString
			result.WriteByte(ch)
			i++
			continue
		}

		// Skip if in string
		if inString || inSingleQuoteString {
			result.WriteByte(ch)
			i++
			continue
		}

		// Check for regex literal starting with /
		if ch == '/' {
			// Check if this is a regex literal based on context
			if isRegexContext(s, i, result.String()) {
				// Parse the regex literal
				regexEnd, pattern, flags := parseRegexLiteral(s, i)
				if regexEnd > i {
					// Convert to regex function call: regex("pattern", "flags")
					escaped := escapeRegexForString(pattern)
					result.WriteString("regex(\"")
					result.WriteString(escaped)
					result.WriteString("\"")
					
					// Add flags argument if present
					if flags != "" {
						result.WriteString(", \"")
						result.WriteString(flags)
						result.WriteString("\"")
					}
					
					result.WriteByte(')')
					
					i = regexEnd
					continue
				}
			}
		}

		result.WriteByte(ch)
		i++
	}

	return result.String()
}

// IsRegexContext exposes regex context detection for comment stripping.
func IsRegexContext(s string, i int, resultSoFar string) bool {
	return isRegexContext(s, i, resultSoFar)
}

// isRegexContext determines if a slash at position i should be interpreted as
// the start of a regex literal based on the preceding context.
func isRegexContext(s string, i int, resultSoFar string) bool {
	// Find the last non-whitespace character before position i
	prevIdx := i - 1
	for prevIdx >= 0 && unicode.IsSpace(rune(s[prevIdx])) {
		prevIdx--
	}

	if prevIdx < 0 {
		// Start of input → regex
		return true
	}

	prevChar := s[prevIdx]

	// After these characters, we expect a regex
	regexStarters := "([{,;:=<>!&|+-*%^~?"
	if strings.ContainsRune(regexStarters, rune(prevChar)) {
		return true
	}

	// Check for word operators that precede regex
	// Look at the result so far to check for keywords
	trimmed := strings.TrimRightFunc(resultSoFar, unicode.IsSpace)
	wordOperators := []string{
		"contains", "find", "match", "matches", "scan", "splitBy", "replace",
		"return", "if", "else", "then", "and", "or", "not",
	}

	for _, op := range wordOperators {
		if strings.HasSuffix(trimmed, op) {
			// Make sure it's a word boundary
			opStart := len(trimmed) - len(op)
			if opStart == 0 || !isIdentChar(trimmed[opStart-1]) {
				return true
			}
		}
	}

	// After identifier or closing bracket → division
	if isIdentChar(prevChar) || prevChar == ')' || prevChar == ']' || prevChar == '}' || prevChar == '"' || prevChar == '\'' {
		return false
	}

	// Default to regex for safety
	return true
}

// ParseRegexLiteral exposes regex literal parsing for comment stripping.
func ParseRegexLiteral(s string, start int) (end int, pattern string, flags string) {
	return parseRegexLiteral(s, start)
}

// parseRegexLiteral parses a regex literal starting at position i.
// Returns the end position (after flags), the pattern, and any flags.
func parseRegexLiteral(s string, start int) (end int, pattern string, flags string) {
	if start >= len(s) || s[start] != '/' {
		return start, "", ""
	}

	i := start + 1
	var patternBuilder strings.Builder
	escaped := false

	for i < len(s) {
		ch := s[i]

		if escaped {
			// Handle escaped characters in regex
			if ch == '/' {
				// Escaped slash - include just the slash
				patternBuilder.WriteByte('/')
			} else {
				// Other escaped char - keep the backslash
				patternBuilder.WriteByte('\\')
				patternBuilder.WriteByte(ch)
			}
			escaped = false
			i++
			continue
		}

		if ch == '\\' {
			escaped = true
			i++
			continue
		}

		if ch == '/' {
			// End of regex pattern
			i++ // Move past closing /

			// Parse flags (i, m, s, g, etc.)
			var flagsBuilder strings.Builder
			for i < len(s) && isRegexFlag(s[i]) {
				flagsBuilder.WriteByte(s[i])
				i++
			}

			return i, patternBuilder.String(), flagsBuilder.String()
		}

		if ch == '\n' || ch == '\r' {
			// Regex literals can't span lines - this isn't a regex
			return start, "", ""
		}

		patternBuilder.WriteByte(ch)
		i++
	}

	// Never found closing / - not a valid regex literal
	return start, "", ""
}

// isRegexFlag returns true if ch is a valid regex flag character.
func isRegexFlag(ch byte) bool {
	return ch == 'i' || ch == 'm' || ch == 's' || ch == 'g' || ch == 'u' || ch == 'x'
}

// escapeRegexForString escapes a regex pattern for use in a double-quoted string.
func escapeRegexForString(pattern string) string {
	var result strings.Builder
	result.Grow(len(pattern) + 10)

	for i := 0; i < len(pattern); i++ {
		ch := pattern[i]
		switch ch {
		case '"':
			result.WriteString(`\"`)
		case '\\':
			// Keep backslashes - they're regex escapes
			result.WriteByte('\\')
			result.WriteByte('\\')
		case '\n':
			result.WriteString(`\n`)
		case '\r':
			result.WriteString(`\r`)
		case '\t':
			result.WriteString(`\t`)
		default:
			result.WriteByte(ch)
		}
	}

	return result.String()
}

// replaceReplaceOperator converts "str replace pattern with replacement" to "replace(str, pattern, replacement)"
// This handles both string patterns and regex literals (after regex literal conversion).
func replaceReplaceOperator(s string) string {
	const replaceKeyword = " replace "
	const withKeyword = " with "

	var result []rune
	inString := false
	i := 0

	for i < len(s) {
		if s[i] == '"' && (i == 0 || s[i-1] != '\\') {
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
		if s[i] == '"' && (i == 0 || s[i-1] != '\\') {
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

		if ch == '"' && (i == 0 || s[i-1] != '\\') {
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
