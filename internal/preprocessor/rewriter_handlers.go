package preprocessor

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/stringutils"
)

// handleDoubleQuote manages double-quoted strings, including escape detection.
func (r *rewriter) handleDoubleQuote(char byte) bool {
	if char != '"' {
		return false
	}

	if !r.inString {
		// Start of double-quoted string
		r.inString = true
		r.stringEscape = false
		r.append(char, r.pos)
		return true
	}

	if r.stringEscape {
		// Quote is escaped, continue in string
		r.stringEscape = false
		r.append(char, r.pos)
		return true
	}

	// End of double-quoted string
	r.inString = false
	r.stringEscape = false
	r.append(char, r.pos)
	return true
}

func (r *rewriter) setInvariantError(format string, args ...interface{}) {
	if r.err != nil {
		return
	}
	r.err = unifiederrors.ParseErrorf(format, args...)
}

// handleSingleQuote converts single-quoted strings to double-quoted.
func (r *rewriter) handleSingleQuote(char byte) bool {
	if char != '\'' || r.inString {
		return false
	}

	// Start of single-quoted string - convert to double-quoted
	r.append('"', r.pos)
	r.pos++
	for r.pos < len(r.input) && r.input[r.pos] != '\'' {
		r.processSingleQuoteContent()
	}
	if r.pos >= len(r.input) {
		r.setInvariantError("unterminated single-quoted string at position %d", r.pos)
		return true
	}
	r.append('"', r.pos)
	return true
}

// processSingleQuoteContent handles the content inside a single-quoted string.
func (r *rewriter) processSingleQuoteContent() {
	ch := r.input[r.pos]

	// Handle escape sequences
	if ch == '\\' && r.pos+1 < len(r.input) {
		next := r.input[r.pos+1]
		if next == '\'' {
			// \' becomes just ' in double-quoted string
			r.append('\'', r.pos)
			r.pos += 2
			return
		} else if next == '\\' {
			// \\ stays as \\
			r.append('\\', r.pos)
			r.append('\\', r.pos)
			r.pos += 2
			return
		} else { // Other backslash sequences: escape the backslash
			r.append('\\', r.pos)
			r.append('\\', r.pos)
			r.pos++
			return
		}
	}

	// Escape double quotes when converting from single to double quotes
	if ch == '"' {
		r.append('\\', r.pos)
	}

	r.append(ch, r.pos)
	r.pos++
}

func (r *rewriter) handleIfElse() bool {
	startPos, condStartPos, condEndPos, ok := r.findKeywordCondition("if")
	if !ok {
		return false
	}

	branches, ok := r.findBranches(condEndPos)
	if !ok {
		return false
	}

	condition, conditionMapping, ok := r.rewriteBranchWithMapping(r.input[condStartPos+1:condEndPos], condStartPos+1)
	if !ok {
		return true
	}
	trueBranch, trueMapping, ok := r.rewriteBranchWithMapping(r.input[branches.TrueStart:branches.TrueEnd], branches.TrueStart)
	if !ok {
		return true
	}

	falseBranch := "nil"
	falseMapping := identityMapping(len(falseBranch))
	if branches.FalseStart > 0 && branches.FalseEnd > 0 {
		falseBranch, falseMapping, ok = r.rewriteBranchWithMapping(r.input[branches.FalseStart:branches.FalseEnd], branches.FalseStart)
		if !ok {
			return true
		}
	}

	r.appendIfElse(startPos, condition, conditionMapping, trueBranch, trueMapping, falseBranch, falseMapping)
	r.updatePositionAfterIfElse(branches.TrueEnd, branches.FalseStart, branches.FalseEnd)
	return true
}

func (r *rewriter) handleWhileLoop() bool {
	startPos, condStartPos, condEndPos, ok := r.findKeywordCondition("while")
	if !ok {
		return false
	}

	bodyStartPos, bodyEndPos, ok := r.findMatchingBraces(condEndPos + 1)
	if !ok {
		return false
	}

	condition, ok := r.rewriteBranch(r.input[condStartPos+1 : condEndPos])
	if !ok {
		return true
	}
	body, ok := r.rewriteBranch(r.input[bodyStartPos+1 : bodyEndPos])
	if !ok {
		return true
	}

	body = replaceMultiStatementSequences(body)
	r.appendWhile(startPos, condition, body)
	r.pos = bodyEndPos
	return true
}

func (r *rewriter) handleBreak() bool {
	if !strings.HasPrefix(r.input[r.pos:], "break") {
		return false
	}

	afterBreak := r.input[r.pos+5:]
	if len(afterBreak) > 0 {
		ch := afterBreak[0]
		if !isWordBoundary(ch) && ch != '\n' && ch != ';' && ch != '}' && ch != ')' && ch != ',' {
			return false
		}
	}

	startPos := r.pos
	r.appendString("__break()", startPos)
	r.pos += 5 // Move past "break"
	return true
}

func (r *rewriter) handleContinue() bool {
	if !strings.HasPrefix(r.input[r.pos:], "continue") {
		return false
	}

	afterContinue := r.input[r.pos+8:]
	if len(afterContinue) > 0 {
		ch := afterContinue[0]
		if !isWordBoundary(ch) && ch != '\n' && ch != ';' && ch != '}' && ch != ')' && ch != ',' {
			return false
		}
	}

	startPos := r.pos
	r.appendString("__continue()", startPos)
	r.pos += 8 // Move past "continue"
	return true
}

// findMatchingBraces locates a brace-enclosed block starting from a given position.
//
// Skips whitespace and newlines, then looks for an opening brace '{' and finds
// its matching closing brace '}', handling nested braces and string literals.
//
// Parameters:
//   - afterCondPos: Position to start searching from (typically after a condition)
//
// Returns:
//   - braceStartPos: Position of the opening brace '{'
//   - braceEndPos: Position of the closing brace '}' (inclusive)
//   - ok: false if no opening brace found or braces are unbalanced
func (r *rewriter) findMatchingBraces(afterCondPos int) (int, int, bool) {
	return findMatchingDelimited(r.input, afterCondPos, '{', '}', true)
}

// findMatchingParens locates a parenthesized block starting from a given position.
//
// Skips horizontal space, then expects an opening '(' and finds the matching ')'.
// Newlines are not allowed between the keyword and the opening parenthesis.
func (r *rewriter) findMatchingParens(afterIfPos int) (int, int, bool) {
	return findMatchingDelimited(r.input, afterIfPos, '(', ')', false)
}

// findBranches locates the true and false branches of an if-else expression.
//
// Starting from the position after the condition's closing parenthesis, it scans
// for the true branch (required) and optional false branch (after " else ").
//
// Parameters:
//   - condEndPos: Position of the closing parenthesis of the condition
//
// Returns:
//   - trueStartPos: Start position of the true branch expression
//   - trueEndPos: End position of the true branch expression (exclusive)
//   - falseBranchStart: Start position of the false branch (0 if no else)
//   - falseBranchEnd: End position of the false branch (0 if no else)
//   - ok: false if parsing failed (e.g., mismatched brackets)
//
// The function handles nested brackets/braces and string literals to correctly
// identify branch boundaries, then stops at the first top-level " else " or
// newline boundary.
func (r *rewriter) findBranches(condEndPos int) (BranchPositions, bool) {
	var trueStartPos int
	if r.opts.AllowMultilineIfElse {
		trueStartPos = skipSpace(r.input, condEndPos+1, true)
	} else {
		trueStartPos = skipSpace(r.input, condEndPos+1, false)
	}
	scan := scanBranchEnd(r.input, trueStartPos, r.opts.AllowMultilineIfElse)
	if !scan.OK {
		r.setInvariantError("branch scan closed before open at position %d", scan.ErrPos)
		return BranchPositions{}, false
	}

	if scan.ElsePos < 0 {
		return BranchPositions{TrueStart: trueStartPos, TrueEnd: scan.BranchEnd}, true
	}

	falseBranchStart := skipSpace(r.input, scan.ElsePos+len("else"), true)
	var falseBranchEnd int
	if r.opts.AllowMultilineIfElse {
		falseBranchEnd = findExpressionEnd(r.input, falseBranchStart, true)
	} else {
		falseBranchEnd = findLineEnd(r.input, falseBranchStart)
	}
	return BranchPositions{
		TrueStart:  trueStartPos,
		TrueEnd:    scan.BranchEnd,
		FalseStart: falseBranchStart,
		FalseEnd:   falseBranchEnd,
	}, true
}

func (r *rewriter) handleAndOr() bool {
	if strings.HasPrefix(r.input[r.pos:], " and ") {
		r.appendString(" && ", r.pos)
		r.pos += 4
		return true
	}
	if strings.HasPrefix(r.input[r.pos:], " or ") {
		r.appendString(" || ", r.pos)
		r.pos += 3
		return true
	}
	return false
}

func (r *rewriter) handleNot() bool {
	if !strings.HasPrefix(r.input[r.pos:], "not ") {
		return false
	}
	if r.pos > 0 {
		prevChar := r.input[r.pos-1]
		if prevChar != ' ' && prevChar != '(' && prevChar != '[' && prevChar != ',' &&
			prevChar != ':' && prevChar != '!' && prevChar != '&' && prevChar != '|' {
			return false
		}
	}
	r.appendString("!", r.pos)
	r.pos += 3
	return true
}

func (r *rewriter) handleCoerceEquals() bool {
	if !strings.HasPrefix(r.input[r.pos:], "~=") {
		return false
	}
	r.appendString("^", r.pos)
	r.pos += 1
	return true
}

func (r *rewriter) handleUsing() bool {
	if !strings.HasPrefix(r.input[r.pos:], "using") {
		return false
	}
	if r.pos > 0 {
		prev := r.input[r.pos-1]
		if unicode.IsLetter(rune(prev)) || unicode.IsDigit(rune(prev)) || prev == '_' {
			return false
		}
	}

	parenStart := r.pos + len("using")
	for parenStart < len(r.input) && unicode.IsSpace(rune(r.input[parenStart])) {
		parenStart++
	}
	if parenStart >= len(r.input) || r.input[parenStart] != '(' {
		return false
	}

	parenEnd := parenStart + 1
	state := ScanState{DepthParen: 1}
	for parenEnd < len(r.input) && state.DepthParen > 0 {
		state.Advance(r.input[parenEnd])
		parenEnd++
	}
	if state.DepthParen != 0 {
		return false
	}

	assignments := r.input[parenStart+1 : parenEnd-1]
	exprStart := parenEnd
	for exprStart < len(r.input) && unicode.IsSpace(rune(r.input[exprStart])) {
		exprStart++
	}
	if exprStart >= len(r.input) {
		return false
	}

	exprEnd := findUsingExpressionEnd(r.input, exprStart)
	expression := strings.TrimSpace(r.input[exprStart:exprEnd])
	declarations := buildUsingDeclarations(assignments)

	declarations = strconv.Quote(declarations)
	expression = strconv.Quote(expression)

	output := fmt.Sprintf("__do(%s, %s)", declarations, expression)
	for i := 0; i < len(output); i++ {
		r.append(output[i], r.pos)
	}

	r.pos = exprEnd - 1
	return true
}

func findUsingExpressionEnd(input string, start int) int {
	state := ScanState{}

	for i := start; i < len(input); i++ {
		ch := input[i]
		if !state.InString() {
			switch ch {
			case ')':
				if state.DepthParen == 0 {
					return i
				}
			case ']':
				if state.DepthBrack == 0 {
					return i
				}
			case '}':
				if state.DepthBrace == 0 {
					return i
				}
			case ',', '\n', '\r':
				if state.AtTopLevel() {
					return i
				}
			}
		}
		state.Advance(ch)
	}

	return len(input)
}

func buildUsingDeclarations(assignments string) string {
	parts := splitTopLevelArgs(assignments)
	lines := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "var ") || strings.HasPrefix(trimmed, "fun ") {
			lines = append(lines, trimmed)
		} else {
			lines = append(lines, "var "+trimmed)
		}
	}
	return strings.Join(lines, "\n")
}

func splitTopLevelArgs(input string) []string {
	var parts []string
	start := 0
	state := ScanState{}

	for i := 0; i < len(input); i++ {
		ch := input[i]
		if ch == ',' && state.AtTopLevel() {
			parts = append(parts, input[start:i])
			start = i + 1
			continue
		}
		state.Advance(ch)
	}

	parts = append(parts, input[start:])
	return parts
}

func (r *rewriter) handleDoBlock() bool {
	if !strings.HasPrefix(r.input[r.pos:], "do ") && !strings.HasPrefix(r.input[r.pos:], "do{") {
		return false
	}

	braceStart, braceEnd, ok := r.findMatchingBraces(r.pos + len("do"))
	if !ok {
		return false
	}

	content := r.input[braceStart+1 : braceEnd]
	separatorIdx := strings.Index(content, "---")
	if separatorIdx == -1 {
		return false
	}

	declarations := strings.TrimSpace(content[:separatorIdx])
	expression := strings.TrimSpace(content[separatorIdx+3:])

	declarations = strconv.Quote(declarations)
	expression = strconv.Quote(expression)

	output := fmt.Sprintf("__do(%s, %s)", declarations, expression)
	for i := 0; i < len(output); i++ {
		r.append(output[i], r.pos)
	}

	r.pos = braceEnd
	return true
}

func (r *rewriter) handleUpdateExprBlock() bool {
	if !strings.HasPrefix(r.input[r.pos:], "update ") && !strings.HasPrefix(r.input[r.pos:], "update{") {
		return false
	}

	braceStart, braceEnd, ok := r.findMatchingBraces(r.pos + len("update"))
	if !ok {
		return false
	}

	casesContent := strings.TrimSpace(r.input[braceStart+1 : braceEnd])
	leftStart := findLeftOperandStartBytes(r.result)
	if leftStart >= len(r.result) {
		return false
	}

	leftOp := strings.TrimSpace(string(r.result[leftStart:]))
	r.truncate(leftStart)

	casesQuoted := strconv.Quote(casesContent)
	output := fmt.Sprintf("__updateExpr(%s, %s)", leftOp, casesQuoted)
	for i := 0; i < len(output); i++ {
		r.append(output[i], r.pos)
	}

	r.pos = braceEnd
	return true
}

func (r *rewriter) handleRangeBuiltin() bool {
	if !strings.HasPrefix(r.input[r.pos:], "range") {
		return false
	}
	if r.pos > 0 {
		prev := r.input[r.pos-1]
		if unicode.IsLetter(rune(prev)) || unicode.IsDigit(rune(prev)) || prev == '_' {
			return false
		}
	}

	afterRange := r.input[r.pos+5:]
	if len(afterRange) == 0 || afterRange[0] != '(' {
		return false
	}

	startPos := r.pos
	r.appendString("__range", startPos)
	r.pos += 4
	return true
}

func (r *rewriter) findKeywordCondition(keyword string) (int, int, int, bool) {
	if !strings.HasPrefix(r.input[r.pos:], keyword) {
		return 0, 0, 0, false
	}

	afterKeyword := r.input[r.pos+len(keyword):]
	if len(afterKeyword) == 0 || (afterKeyword[0] != ' ' && afterKeyword[0] != '(') {
		return 0, 0, 0, false
	}

	startPos := r.pos
	condStartPos, condEndPos, ok := r.findMatchingParens(r.pos + len(keyword))
	if !ok {
		return 0, 0, 0, false
	}
	return startPos, condStartPos, condEndPos, true
}

func normalizeCondition(condition string) string {
	return stringutils.ReplaceOperatorsOutsideStrings(strings.TrimSpace(condition), []stringutils.OperatorReplacement{
		{Pattern: " and ", Replacement: []rune(" && ")},
		{Pattern: " or ", Replacement: []rune(" || ")},
	})
}

func (r *rewriter) rewriteBranch(raw string) (string, bool) {
	rewritten, _, ok := r.rewriteBranchWithMapping(raw, 0)
	return rewritten, ok
}

func (r *rewriter) rewriteBranchWithMapping(raw string, basePos int) (string, []int, bool) {
	trimmed := strings.TrimSpace(raw)
	trimStart := strings.Index(raw, trimmed)
	if trimStart < 0 {
		trimStart = 0
	}
	branchBase := basePos + trimStart
	branch := stringutils.ReplaceOperatorsOutsideStrings(trimmed, []stringutils.OperatorReplacement{
		{Pattern: " and ", Replacement: []rune(" && ")},
		{Pattern: " or ", Replacement: []rune(" || ")},
	})
	branchMapping := inferStageMapping(trimmed, branch)
	branchRewriter := newRewriter(branch, r.opts)
	rewritten, rewrittenMapping, err := branchRewriter.RewriteCoreWithDepth(0)
	if err != nil {
		r.err = err
		return "", nil, false
	}
	composed := composeMappings(branchMapping, rewrittenMapping)
	for i := range composed {
		composed[i] += branchBase
	}
	return rewritten, composed, true
}

func (r *rewriter) appendIfElse(startPos int, condition string, conditionMapping []int, trueBranch string, trueMapping []int, falseBranch string, falseMapping []int) {
	conditionEnd := mappedFallback(conditionMapping, startPos)
	trueEnd := mappedFallback(trueMapping, conditionEnd)
	falseEnd := mappedFallback(falseMapping, trueEnd)

	r.appendString("__ifelse(", startPos)
	r.appendMappedString(condition, conditionMapping, startPos)
	r.appendString(", ", conditionEnd)
	r.appendMappedString(trueBranch, trueMapping, startPos)
	r.appendString(", ", trueEnd)
	r.appendMappedString(falseBranch, falseMapping, startPos)
	r.append(')', falseEnd)
}

func (r *rewriter) appendWhile(startPos int, condition, body string) {
	r.appendString("__while(", startPos)
	r.appendString(condition, startPos)
	r.appendString(", (", startPos)
	r.appendString(body, startPos)
	r.append(')', startPos)
	r.append(')', startPos)
}

func (r *rewriter) updatePositionAfterIfElse(trueEndPos int, falseBranchStart int, falseBranchEnd int) {
	if falseBranchStart > 0 && falseBranchEnd > 0 {
		r.pos = falseBranchEnd - 1
		return
	}
	r.pos = trueEndPos - 1
}

func skipSpace(input string, pos int, includeNewlines bool) int {
	for pos < len(input) && (input[pos] == ' ' || input[pos] == '\t' || (includeNewlines && (input[pos] == '\n' || input[pos] == '\r'))) {
		pos++
	}
	return pos
}

func mappedFallback(mapping []int, fallback int) int {
	if len(mapping) == 0 {
		return fallback
	}
	return mapping[len(mapping)-1]
}

func delimiterDepth(state ScanState, opener byte) int {
	switch opener {
	case '(':
		return state.DepthParen
	case '[':
		return state.DepthBrack
	case '{':
		return state.DepthBrace
	default:
		return 0
	}
}

// findMatchingDelimited locates a delimited block starting at start, handling nesting.
//
// Returns the opener and closer positions (inclusive) or ok=false when the opener
// is missing or the delimiters are unbalanced.
func findMatchingDelimited(input string, start int, opener byte, closer byte, allowNewlines bool) (int, int, bool) {
	pos := start
	if allowNewlines {
		pos = skipSpace(input, pos, true)
	} else {
		pos = skipSpace(input, pos, false)
	}

	if pos >= len(input) || input[pos] != opener {
		return 0, 0, false
	}

	startPos := pos
	pos++
	state := ScanState{}
	switch opener {
	case '(':
		state.DepthParen = 1
	case '[':
		state.DepthBrack = 1
	case '{':
		state.DepthBrace = 1
	default:
		return 0, 0, false
	}
	for pos < len(input) && delimiterDepth(state, opener) > 0 {
		state.Advance(input[pos])
		pos++
	}
	if delimiterDepth(state, opener) != 0 {
		return 0, 0, false
	}

	endPos := pos - 1
	return startPos, endPos, true
}

// scanBranchEnd finds the end of a branch expression starting at start.
//
// It tracks nesting across (), {}, [] and ignores content inside string literals.
// Returns the end position, the else position (or -1), and the position of
// an unmatched closer if the scan fails.
func scanBranchEnd(input string, start int, allowMultiline bool) BranchScanResult {
	pos := start
	state := ScanState{}
	for pos < len(input) {
		ch := input[pos]
		if !state.InString() {
			if isBranchCloser(ch) && state.Depth() == 0 {
				return BranchScanResult{ErrPos: pos}
			} else if state.Depth() == 0 && isElseKeywordAt(input, pos) {
				// branchEnd should be before any whitespace leading to else
				branchEnd := pos
				for branchEnd > start && isWhitespace(input[branchEnd-1]) {
					branchEnd--
				}
				return BranchScanResult{BranchEnd: branchEnd, ElsePos: pos, ErrPos: -1, OK: true}
			} else if allowMultiline && state.Depth() == 0 && (ch == '\n' || ch == '\r') {
				next := pos + 1
				for next < len(input) && isWhitespace(input[next]) {
					next++
				}
				if next < len(input) && isElseKeywordAt(input, next) {
					// allow multiline if/else by continuing until "else"
				} else {
					return BranchScanResult{BranchEnd: pos, ElsePos: -1, ErrPos: -1, OK: true}
				}
			}
			if !allowMultiline && state.Depth() == 0 && (ch == '\n' || pos == len(input)-1) {
				if ch != '\n' {
					pos++
				}
				return BranchScanResult{BranchEnd: pos, ElsePos: -1, ErrPos: -1, OK: true}
			}
		}
		state.Advance(ch)
		pos++
	}

	return BranchScanResult{BranchEnd: pos, ElsePos: -1, ErrPos: -1, OK: true}
}

func isElseKeywordAt(input string, pos int) bool {
	if pos+4 > len(input) || input[pos:pos+4] != "else" {
		return false
	}
	if pos > 0 && !isWordBoundary(input[pos-1]) {
		return false
	}
	afterPos := pos + 4
	if afterPos >= len(input) {
		return true
	}
	return isWordBoundary(input[afterPos])
}

func findExpressionEnd(input string, start int, allowNewlines bool) int {
	state := ScanState{}

	for i := start; i < len(input); i++ {
		ch := input[i]
		if !state.InString() {
			switch ch {
			case ')':
				if state.DepthParen == 0 {
					return i
				}
			case ']':
				if state.DepthBrack == 0 {
					return i
				}
			case '}':
				if state.DepthBrace == 0 {
					return i
				}
			case ',':
				if state.AtTopLevel() {
					return i
				}
			case '\n', '\r':
				if !allowNewlines && state.AtTopLevel() {
					return i
				}
			}
		}
		state.Advance(ch)
	}

	return len(input)
}

func isBranchOpener(ch byte) bool {
	return ch == '(' || ch == '{' || ch == '['
}

func isBranchCloser(ch byte) bool {
	return ch == ')' || ch == '}' || ch == ']'
}

// findLineEnd returns the position of the next newline or the end of input.
func findLineEnd(input string, start int) int {
	for pos := start; pos < len(input); pos++ {
		if input[pos] == '\n' {
			return pos
		}
	}
	return len(input)
}
