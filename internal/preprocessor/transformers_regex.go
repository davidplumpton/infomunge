package preprocessor

import (
	"infomunge/internal/stringutils"
	"strings"
)

type regexTokenKind int

const (
	regexTokenNone regexTokenKind = iota
	regexTokenValue
	regexTokenOperator
)

var regexPrefixKeywords = map[string]struct{}{
	// These keywords syntactically behave like operators in DW-like expressions.
	// If a slash follows one of them, we treat it like an expression start, so
	// `/.../` should be parsed as a regex literal (not division).
	"and":      {},
	"case":     {},
	"contains": {},
	"else":     {},
	"find":     {},
	"if":       {},
	"in":       {},
	"match":    {},
	"matches":  {},
	"not":      {},
	"or":       {},
	"replace":  {},
	"return":   {},
	"scan":     {},
	"splitBy":  {},
	"then":     {},
	"using":    {},
	"while":    {},
}

// replaceRegexLiterals converts /pattern/ regex literals to string literals "pattern".
// This must run early in the pipeline before other operators try to parse the slashes.
//
// Context rules for distinguishing regex from division:
// - After operators, open brackets, colon, comma → regex
// - After identifier or closing bracket → division
// - At start of expression → regex
func replaceRegexLiterals(s string) string {
	result, _ := replaceRegexLiteralsWithMapping(s)
	return result
}

// replaceRegexLiteralsWithMapping converts /pattern/ regex literals to regex("pattern", "flags")
// function calls, returning exact source-position mappings for the sourcemap pipeline.
func replaceRegexLiteralsWithMapping(s string) (string, []int) {
	buf := newMappedBuffer(len(s) + 32)

	i := 0
	inString := false
	inSingleQuoteString := false
	lastKind := regexTokenNone

	// Single-pass lexer-style rewrite:
	// 1) Track string contexts so we never rewrite slashes inside string literals.
	// 2) Maintain lastKind (value/operator/none) to decide whether '/' can start regex.
	// 3) When context allows, parse `/pattern/flags` and rewrite to regex("pattern", "flags").
	// 4) Otherwise keep '/' as-is so downstream stages can interpret division/comments.
	for i < len(s) {
		ch := s[i]

		// Track string state
		if ch == '"' && !inSingleQuoteString && !stringutils.IsEscapedAt(s, i) {
			inString = !inString
			buf.AppendOriginal(s, i, i+1)
			if !inString {
				lastKind = regexTokenValue
			}
			i++
			continue
		}

		if ch == '\'' && !inString && !stringutils.IsEscapedAt(s, i) {
			inSingleQuoteString = !inSingleQuoteString
			buf.AppendOriginal(s, i, i+1)
			if !inSingleQuoteString {
				lastKind = regexTokenValue
			}
			i++
			continue
		}

		// Skip if in string
		if inString || inSingleQuoteString {
			buf.AppendOriginal(s, i, i+1)
			i++
			continue
		}

		if isWhitespace(ch) {
			buf.AppendOriginal(s, i, i+1)
			i++
			continue
		}

		if IsIdentifierStart(ch) {
			start := i
			i++
			for i < len(s) && IsIdentifierPart(s[i]) {
				i++
			}
			ident := s[start:i]
			buf.AppendOriginal(s, start, i)
			if isRegexPrefixKeyword(ident) {
				lastKind = regexTokenOperator
			} else {
				lastKind = regexTokenValue
			}
			continue
		}

		if ch >= '0' && ch <= '9' {
			start := i
			i++
			for i < len(s) && ((s[i] >= '0' && s[i] <= '9') || s[i] == '.') {
				i++
			}
			buf.AppendOriginal(s, start, i)
			lastKind = regexTokenValue
			continue
		}

		if IsOpeningBracket(ch) || isRegexOperatorChar(ch) {
			buf.AppendOriginal(s, i, i+1)
			lastKind = regexTokenOperator
			i++
			continue
		}

		if IsClosingBracket(ch) {
			buf.AppendOriginal(s, i, i+1)
			lastKind = regexTokenValue
			i++
			continue
		}

		// Check for regex literal starting with /
		if ch == '/' {
			// Check if this is a regex literal based on context
			if canStartRegexAfter(lastKind) {
				// Parse the regex literal
				regexEnd, pattern, flags := parseRegexLiteral(s, i)
				if regexEnd > i {
					// Convert to regex function call: regex("pattern", "flags")
					// Map all generated characters to the opening slash position.
					escaped := escapeRegexForString(pattern)
					buf.AppendLiteral("regex(\"", i)
					buf.AppendLiteral(escaped, i)
					buf.AppendLiteral("\"", i)

					// Add flags argument if present
					if flags != "" {
						buf.AppendLiteral(", \"", i)
						buf.AppendLiteral(flags, i)
						buf.AppendLiteral("\"", i)
					}

					buf.AppendLiteral(")", i)
					lastKind = regexTokenValue

					i = regexEnd
					continue
				}
			}
		}

		buf.AppendOriginal(s, i, i+1)
		if ch == '/' || isRegexOperatorChar(ch) {
			lastKind = regexTokenOperator
		}
		i++
	}

	return buf.String(), buf.mapping
}

// isRegexContext determines if a slash at position i should be interpreted as
// the start of a regex literal based on the preceding context.
func isRegexContext(s string, i int) bool {
	if i <= 0 {
		return true
	}
	lastKind := scanRegexPrefixKind(s, i)
	return canStartRegexAfter(lastKind)
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

func isRegexPrefixKeyword(word string) bool {
	_, ok := regexPrefixKeywords[word]
	return ok
}

func isRegexOperatorChar(ch byte) bool {
	switch ch {
	// Grouped categories:
	// - Delimiters: ',', ';', ':'
	// - Assignment/comparison: '=', '<', '>', '!'
	// - Boolean/bitwise: '&', '|', '^', '~'
	// - Arithmetic/prefix/suffix: '+', '-', '*', '%', '?'
	// - Brackets are included for compatibility with historical context checks.
	case '(', '[', '{', ',', ';', ':', '=', '<', '>', '!', '&', '|', '+', '-', '*', '%', '^', '~', '?':
		return true
	default:
		return false
	}
}

func canStartRegexAfter(kind regexTokenKind) bool {
	return kind == regexTokenNone || kind == regexTokenOperator
}

// scanRegexPrefixKind classifies the token kind immediately before position end.
// This mirrors replaceRegexLiterals context tracking but only for the prefix.
// It intentionally re-parses nested/quoted constructs so comment stripping and
// other slash-aware passes can answer "would '/' start a regex here?" without
// running the full rewrite pipeline.
func scanRegexPrefixKind(s string, end int) regexTokenKind {
	if end > len(s) {
		end = len(s)
	}
	i := 0
	lastKind := regexTokenNone
	for i < end {
		ch := s[i]
		if isWhitespace(ch) {
			i++
			continue
		}
		if ch == '"' || ch == '\'' {
			i = scanStringLiteral(s, i, ch)
			lastKind = regexTokenValue
			continue
		}
		if IsIdentifierStart(ch) {
			start := i
			i++
			for i < end && IsIdentifierPart(s[i]) {
				i++
			}
			if isRegexPrefixKeyword(s[start:i]) {
				lastKind = regexTokenOperator
			} else {
				lastKind = regexTokenValue
			}
			continue
		}
		if ch >= '0' && ch <= '9' {
			i++
			for i < end && ((s[i] >= '0' && s[i] <= '9') || s[i] == '.') {
				i++
			}
			lastKind = regexTokenValue
			continue
		}
		if ch == '/' {
			// Reuse full literal parsing here so escaped slashes and flags are
			// interpreted exactly the same as the main regex rewrite pass.
			if canStartRegexAfter(lastKind) {
				regexEnd, _, _ := parseRegexLiteral(s, i)
				if regexEnd > i && regexEnd <= end {
					i = regexEnd
					lastKind = regexTokenValue
					continue
				}
			}
			i++
			lastKind = regexTokenOperator
			continue
		}
		if IsOpeningBracket(ch) || isRegexOperatorChar(ch) {
			i++
			lastKind = regexTokenOperator
			continue
		}
		if IsClosingBracket(ch) {
			i++
			lastKind = regexTokenValue
			continue
		}
		i++
	}
	return lastKind
}

func scanStringLiteral(s string, start int, quote byte) int {
	var state ScanState
	state.Advance(quote)
	for i := start + 1; i < len(s); i++ {
		state.Advance(s[i])
		if !state.InString() {
			return i + 1
		}
	}
	return len(s)
}
