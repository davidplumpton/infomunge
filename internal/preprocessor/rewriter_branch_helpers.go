package preprocessor

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
// Returns the end position and the else position (or -1). Top-level commas
// and closing delimiters terminate the branch; they belong to the containing
// object, array, call, or grouped expression, not to the branch body.
func scanBranchEnd(input string, start int, allowMultiline bool) BranchScanResult {
	pos := start
	state := ScanState{}
	for pos < len(input) {
		ch := input[pos]
		if !state.InString() {
			if state.Depth() == 0 && isBranchTerminator(ch) {
				branchEnd := pos
				for branchEnd > start && isWhitespace(input[branchEnd-1]) {
					branchEnd--
				}
				return BranchScanResult{BranchEnd: branchEnd, ElsePos: -1, ErrPos: -1, OK: true}
			} else if state.Depth() == 0 && isElseKeywordAt(input, pos) {
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
					// Continue until "else".
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

func isBranchCloser(ch byte) bool {
	return ch == ')' || ch == '}' || ch == ']'
}

func isBranchTerminator(ch byte) bool {
	return ch == ',' || isBranchCloser(ch)
}
