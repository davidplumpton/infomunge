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

// updateLeftOperandStartBytes keeps an update expression inside an unresolved
// collection callback. Core rewriting runs before collection and arrow
// transforms, so the generic minimal-stop scan would otherwise consume the
// collection source and operator as part of the update's left operand.
func updateLeftOperandStartBytes(result []byte) int {
	start := findLeftOperandStartBytes(result)
	bodyStart := collectionLambdaBodyStart(string(result))
	if bodyStart < 0 {
		return start
	}
	if isCompleteGroupedLambdaBody(string(result[bodyStart:])) {
		return start
	}

	expressionStart := collectionLambdaExpressionStart(result, bodyStart)
	if expressionStart > start {
		return expressionStart
	}
	return start
}

// collectionLambdaExpressionStart returns the expression after an explicit
// callback arrow, or the implicit callback body start when no arrow is present.
func collectionLambdaExpressionStart(input []byte, bodyStart int) int {
	pos := bodyStart
	for pos < len(input) && isWhitespace(input[pos]) {
		pos++
	}
	implicitStart := pos

	var state ScanState
	for pos+1 < len(input) {
		if !state.InString() && state.Depth() == 0 &&
			input[pos] == '-' && input[pos+1] == '>' {
			pos += 2
			for pos < len(input) && isWhitespace(input[pos]) {
				pos++
			}
			return pos
		}
		state.Advance(input[pos])
		pos++
	}
	return implicitStart
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

// selectorLeftOperandStart keeps selector operands inside collection lambda
// bodies that have not yet been rewritten. Selector processing runs before
// functional processing, so a raw expression such as "items map $.values"
// would otherwise be treated as one left operand.
func selectorLeftOperandStart(result []rune, stops []rune) int {
	start := stringutils.FindLeftOperandStartWithStops(result, stops)
	input := string(result)
	bodyStart := collectionLambdaBodyStart(input)
	if bodyStart < 0 {
		return start
	}
	bodyStartRune := runeIndexAtByteOffset(input, bodyStart)
	if bodyStartRune > start {
		return bodyStartRune
	}
	return start
}

func extractSelectorLeftOperand(result []rune) (leftExpr string, newResult []rune, ok bool) {
	start := selectorLeftOperandStart(result, stringutils.DefaultOperatorStops)
	if start >= len(result) {
		return "", result, false
	}

	leftExpr = strings.TrimSpace(string(result[start:]))
	newResult = result[:start]
	return leftExpr, newResult, true
}

func selectorLeftOperandStartBytes(result []byte, stops []byte) int {
	start := findLeftOperandStartBytesWithStops(result, stops)
	bodyStart := collectionLambdaBodyStart(string(result))
	if bodyStart > start {
		return bodyStart
	}
	return start
}

func collectionLambdaBodyStart(input string) int {
	type candidate struct {
		paren int
		brack int
		brace int
		start int
	}

	var state ScanState
	var candidates []candidate
	for pos := 0; pos < len(input); pos++ {
		if !state.InString() {
			if bodyStart, ok := collectionOperatorBodyStartAt(input, pos); ok {
				candidates = append(candidates, candidate{
					paren: state.DepthParen,
					brack: state.DepthBrack,
					brace: state.DepthBrace,
					start: bodyStart,
				})
			}
		}
		state.Advance(input[pos])
	}

	for i := len(candidates) - 1; i >= 0; i-- {
		if candidates[i].paren == state.DepthParen &&
			candidates[i].brack == state.DepthBrack &&
			candidates[i].brace == state.DepthBrace {
			return candidates[i].start
		}
	}
	return -1
}

func collectionOperatorBodyStartAt(input string, pos int) (int, bool) {
	if pos == 0 || !isWhitespace(input[pos-1]) {
		return 0, false
	}
	for _, operator := range CollectionOperators {
		end := pos + len(operator)
		if end > len(input) || input[pos:end] != operator {
			continue
		}
		if end >= len(input) || (!isWhitespace(input[end]) && input[end] != '(') {
			continue
		}
		for end < len(input) && isWhitespace(input[end]) {
			end++
		}
		return end, true
	}
	return 0, false
}

// isPatternMatchKeyword checks if the scanner is at a pattern match keyword position.
func isPatternMatchKeyword(s string, sc *stringutils.ExpressionScanner) (int, int, bool) {
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
	sc := stringutils.NewExpressionScanner(s)
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
		if strings.HasPrefix(s[bracePos:], GoObjectPrefix) {
			bracePos += len(GoObjectPrefix) - 1 // Point to the '{'
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
func findValueExpressionEnd(s string, sc *stringutils.ExpressionScanner) int {
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
	var state ScanState
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
				lineStartDepth = state.DepthBrace
				lineStartInString = state.InString()
			}
			continue
		}

		if !state.InString() && ch == '/' && i+1 < len(raw) && raw[i+1] == '/' {
			inLineComment = true
			continue
		}

		state.Advance(ch)

		if ch == '\n' {
			if lineStartDepth == 0 && !lineStartInString {
				line := raw[lineStart:i]
				if strings.TrimSpace(line) == "---" {
					return lineStart, i + 1, true
				}
			}
			lineStart = i + 1
			lineStartDepth = state.DepthBrace
			lineStartInString = state.InString()
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
	var state ScanState
	inLineComment := false

	for i := 0; i < len(raw); i++ {
		ch := raw[i]

		if inLineComment {
			if ch == '\n' {
				inLineComment = false
			}
			continue
		}

		if !state.InString() && ch == '/' && i+1 < len(raw) && raw[i+1] == '/' {
			inLineComment = true
			continue
		}

		if state.DepthBrace == 0 && !state.InString() && !inLineComment && strings.HasPrefix(raw[i:], separator) {
			return i, i + len(separator), true
		}

		state.Advance(ch)
	}

	return 0, 0, false
}
