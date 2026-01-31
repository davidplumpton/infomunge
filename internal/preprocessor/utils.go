package preprocessor

import (
	"strings"
	"unicode"

	"infomunge/internal/stringutils"
)

// isWordBoundary checks if a character is a word boundary.
func isWordBoundary(ch byte) bool {
	return !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '#')
}

// isWhitespace checks if a character is whitespace (space, tab, newline, carriage return).
func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// findLeftOperandStartBytes finds the start of the left operand in a byte slice.
// Uses MinimalStops (delimiters and brackets only, no operators).
func findLeftOperandStartBytes(result []byte) int {
	runes := []rune(string(result))
	return stringutils.FindLeftOperandStartWithStops(runes, stringutils.MinimalStops)
}

// extractLeftOperand extracts the left operand expression.
func extractLeftOperand(result []rune) (leftExpr string, newResult []rune, ok bool) {
	start := stringutils.FindLeftOperandStart(result, nil)
	if start >= len(result) {
		return "", result, false
	}

	leftExpr = strings.TrimSpace(string(result[start:]))
	newResult = result[:start]
	return leftExpr, newResult, true
}

// isPatternMatchKeyword checks if the scanner is at a pattern match keyword position.
func isPatternMatchKeyword(s string, sc *StringScanner) (int, int, bool) {
	if sc.IsInString() {
		return -1, 0, false
	}

	pos := sc.Pos()

	// Check for "match"
	if pos+5 <= len(s) && s[pos:pos+5] == "match" {
		if (pos == 0 || !IsIdentifierPart(s[pos-1])) &&
			(pos+5 == len(s) || !IsIdentifierPart(s[pos+5])) {
			return pos, 5, true
		}
	} else if pos > 0 && s[pos] == ' ' && pos+6 <= len(s) && s[pos+1:pos+6] == "match" {
		matchPos := pos + 1
		if (matchPos == 0 || !IsIdentifierPart(s[matchPos-1])) &&
			(matchPos+5 == len(s) || !IsIdentifierPart(s[matchPos+5])) {
			return matchPos, 5, true
		}
	}

	// Check for "case"
	if pos+4 <= len(s) && s[pos:pos+4] == "case" {
		// Verify it's not part of a longer identifier
		if (pos == 0 || !IsIdentifierPart(s[pos-1])) &&
			(pos+4 == len(s) || !IsIdentifierPart(s[pos+4])) {
			return pos, 4, true
		}
	} else if pos > 0 && s[pos] == ' ' && pos+5 <= len(s) && s[pos+1:pos+5] == "case" {
		casePos := pos + 1
		if (casePos == 0 || !IsIdentifierPart(s[casePos-1])) &&
			(casePos+4 == len(s) || !IsIdentifierPart(s[casePos+4])) {
			return casePos, 4, true
		}
	}
	return -1, 0, false
}

func containsIfKeywordOutsideStrings(s string) bool {
	sc := NewScanner(s)
	for sc.Pos() < len(s) {
		if sc.IsInString() {
			sc.Next()
			continue
		}
		pos := sc.Pos()
		if pos+2 <= len(s) && s[pos:pos+2] == "if" {
			beforeOk := pos == 0 || !IsIdentifierPart(s[pos-1])
			if beforeOk {
				afterPos := pos + 2
				for afterPos < len(s) && (s[afterPos] == ' ' || s[afterPos] == '\t') {
					afterPos++
				}
				if afterPos < len(s) && s[afterPos] == '(' {
					return true
				}
			}
		}
		sc.Next()
	}
	return false
}

// findCaseBrace locates the opening brace after the case keyword.
func findCaseBrace(s string, afterCasePos int) (int, bool) {
	bracePos := afterCasePos
	for bracePos < len(s) {
		if unicode.IsSpace(rune(s[bracePos])) {
			bracePos++
			continue
		}
		if strings.HasPrefix(s[bracePos:], "map[string]interface{}{") {
			bracePos += len("map[string]interface{}{") - 1 // Point to the '{'
			break
		}
		if s[bracePos] == '{' {
			break
		}
		bracePos++
	}
	return bracePos, bracePos < len(s) && s[bracePos] == '{'
}

// afterOp returns the position after whitespace from start.
func afterOp(s string, pos int) int {
	for pos < len(s) && unicode.IsSpace(rune(s[pos])) {
		pos++
	}
	return pos
}

// exprStartStops contains stop characters for findExprStart.
// Similar to DefaultOperatorStops but excludes ':' which is used in case expressions.
var exprStartStops = []rune{'+', '-', '*', '/', '=', '<', '>', '!', '(', '[', '{', ','}

// findExprStart finds where an expression starts by scanning backwards.
// Uses stringutils.FindLeftOperandStartWithStops for the core logic.
func findExprStart(runes []rune) (int, bool) {
	if len(runes) == 0 {
		return 0, false
	}

	// Check if all whitespace
	hasNonWhitespace := false
	for _, r := range runes {
		if !unicode.IsSpace(r) {
			hasNonWhitespace = true
			break
		}
	}
	if !hasNonWhitespace {
		return 0, false
	}

	return stringutils.FindLeftOperandStartWithStops(runes, exprStartStops), true
}

// isIdentRune returns true if r can be part of an identifier.
func isIdentRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '#'
}

// isIdentChar returns true if ch can be part of an identifier.
func isIdentChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '#'
}

// isComparisonOperatorEq checks if a '=' is part of a comparison operator.
func isComparisonOperatorEq(prevChar, nextChar byte) bool {
	return nextChar == '=' || prevChar == '=' || prevChar == '!' || prevChar == '<' || prevChar == '>'
}

// extractVariableNameAtPosition walks back from the end of result to find a variable name.
func extractVariableNameAtPosition(result []rune) (string, int, int, bool) {
	varEnd := len(result) - 1

	// Skip whitespace before '='
	for varEnd >= 0 && unicode.IsSpace(rune(result[varEnd])) {
		varEnd--
	}

	if varEnd < 0 {
		return "", 0, 0, false
	}

	// Walk back to find the start of the variable name
	varStart := varEnd
	for varStart >= 0 && isIdentRune(result[varStart]) {
		varStart--
	}
	varStart++

	if varStart > varEnd {
		return "", 0, 0, false
	}

	// Check if variable name is valid
	if varStart > 0 {
		prevCh := result[varStart-1]
		if isIdentRune(prevCh) || prevCh == ']' || prevCh == ')' {
			return "", 0, 0, false
		}
	}

	varName := string(result[varStart : varEnd+1])
	return varName, varStart, varEnd, true
}

// findValueExpressionEnd scans forward to find where the value expression ends.
func findValueExpressionEnd(s string, sc *StringScanner) int {
	depth := 0
	valueEnd := sc.Pos()

	for sc.Pos() < len(s) {
		if !sc.IsInString() {
			ch := s[sc.Pos()]
			if ch == '(' || ch == '[' || ch == '{' {
				depth++
			} else if ch == ')' || ch == ']' || ch == '}' {
				if depth == 0 {
					break
				}
				depth--
			} else if (ch == '\n' || ch == ';' || ch == ',') && depth == 0 {
				break
			}
		}
		valueEnd = sc.Pos()
		sc.Next()
	}

	return valueEnd + 1
}

// FindHeaderSeparator locates the top-level header/body separator.
// It returns headerEnd (exclusive), bodyStart (start of body), and whether it was found.
func FindHeaderSeparator(raw string) (int, int, bool) {
	if headerEnd, bodyStart, ok := findLineSeparatorTopLevel(raw); ok {
		return headerEnd, bodyStart, true
	}
	return findInlineSeparatorTopLevel(raw)
}

// ExtractHeaderAndBody separates the header from the body and returns the body's offset.
func ExtractHeaderAndBody(raw string) (header string, body string, offset int) {
	headerEnd, bodyStart, ok := FindHeaderSeparator(raw)
	if !ok {
		return "", raw, 0
	}

	headerRaw := raw[:headerEnd]
	bodyRaw := raw[bodyStart:]
	header = strings.TrimSpace(headerRaw)
	body = strings.TrimSpace(bodyRaw)
	offset = bodyStart
	trimmedBody := strings.TrimLeftFunc(bodyRaw, unicode.IsSpace)
	offset += len(bodyRaw) - len(trimmedBody)
	return header, body, offset
}

func findLineSeparatorTopLevel(raw string) (int, int, bool) {
	depth := 0
	inString := false
	inLineComment := false
	lineStart := 0
	lineStartDepth := 0
	lineStartInString := false

	for i := 0; i < len(raw); i++ {
		ch := raw[i]

		if inLineComment {
			if ch == '\n' {
				inLineComment = false
				lineStart = i + 1
				lineStartDepth = depth
				lineStartInString = inString
			}
			continue
		}

		if !inString && ch == '/' && i+1 < len(raw) && raw[i+1] == '/' {
			inLineComment = true
			continue
		}

		if ch == '"' && (i == 0 || raw[i-1] != '\\') {
			inString = !inString
			continue
		}

		if !inString {
			switch ch {
			case '{':
				depth++
			case '}':
				if depth > 0 {
					depth--
				}
			}
		}

		if ch == '\n' {
			if lineStartDepth == 0 && !lineStartInString {
				line := raw[lineStart:i]
				if strings.TrimSpace(line) == "---" {
					return lineStart, i + 1, true
				}
			}
			lineStart = i + 1
			lineStartDepth = depth
			lineStartInString = inString
		}
	}

	if lineStartDepth == 0 && !lineStartInString {
		line := raw[lineStart:]
		if strings.TrimSpace(line) == "---" {
			return lineStart, len(raw), true
		}
	}

	return 0, 0, false
}

func findInlineSeparatorTopLevel(raw string) (int, int, bool) {
	const separator = " --- "
	depth := 0
	inString := false
	inLineComment := false

	for i := 0; i < len(raw); i++ {
		ch := raw[i]

		if inLineComment {
			if ch == '\n' {
				inLineComment = false
			}
			continue
		}

		if !inString && ch == '/' && i+1 < len(raw) && raw[i+1] == '/' {
			inLineComment = true
			continue
		}

		if depth == 0 && !inString && !inLineComment && strings.HasPrefix(raw[i:], separator) {
			return i, i + len(separator), true
		}

		if ch == '"' && (i == 0 || raw[i-1] != '\\') {
			inString = !inString
			continue
		}

		if !inString {
			switch ch {
			case '{':
				depth++
			case '}':
				if depth > 0 {
					depth--
				}
			}
		}
	}

	return 0, 0, false
}
