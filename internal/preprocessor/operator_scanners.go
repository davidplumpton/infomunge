package preprocessor

import (
	"strings"
	"unicode"

	"infomunge/internal/stringutils"
)

type leftOperandScanConfig struct {
	ExtraStops      []rune
	UseMinimalStops bool
	CustomFinder    func([]rune) int
}

type rightOperandScanConfig struct {
	StopOps       []string
	StopPredicate func(input string, pos int, operandStart int) bool
}

type typedOperatorRightSpan struct {
	TypeExpr    string
	TypeStart   int
	TypeEnd     int
	ConfigArg   string
	ConfigStart int
	ConfigEnd   int
	Next        int
}

// These lower-precedence infix operators define the left boundary for typed
// operators so `as`/`is` only rewrite the immediate operand to their left.
var typedOperatorLeftBoundaryOps = []string{
	" match ", " matches ", " scan ",
	" default ", " onNull ", " then ",
	" find ", " contains ",
	" splitBy ", " joinBy ", " replace ", " with ",
	" to ", " repeat ", " mod ",
	" map ", " filter ", " reduce ", " flatMap ", " groupBy ", " pluck ",
	" sort ", " orderBy ", " maxBy ", " minBy ", " distinctBy ",
	" filterObject ", " mapObject ",
	" and ", " or ", " && ", " || ",
}

func findLeftOperandBoundsRunes(result []rune, cfg leftOperandScanConfig) (int, string, bool) {
	leftStart := 0
	switch {
	case cfg.CustomFinder != nil:
		leftStart = cfg.CustomFinder(result)
	case cfg.UseMinimalStops:
		leftStart = stringutils.FindLeftOperandStartWithStops(result, stringutils.MinimalStops)
	default:
		leftStart = stringutils.FindLeftOperandStart(result, cfg.ExtraStops)
	}

	if leftStart >= len(result) {
		return 0, "", false
	}

	leftOp := strings.TrimSpace(string(result[leftStart:]))
	if leftOp == "" {
		return 0, "", false
	}

	return leftStart, leftOp, true
}

func findTypedOperatorLeftOperandStartBytes(result []byte) int {
	leftStart := findLeftOperandStartBytesWithStops(result, defaultStopBytes([]rune{':', '&', '|', '~'}))
	if leftStart >= len(result) {
		return leftStart
	}

	leftOp := string(result[leftStart:])
	var state ScanState
	lastBoundary := 0
	for pos := 0; pos < len(leftOp); pos++ {
		if state.AtTopLevel() {
			for _, op := range typedOperatorLeftBoundaryOps {
				if strings.HasPrefix(leftOp[pos:], op) {
					lastBoundary = pos + len(op)
					break
				}
			}
		}
		state.Advance(leftOp[pos])
	}
	if lastBoundary > 0 {
		return leftStart + lastBoundary
	}
	return leftStart
}

// findDefaultLeftOperandStartBytes includes native unary, arithmetic, logical,
// comparison, selector, and call syntax in default's left operand. Default has
// lower precedence than those expressions. An ungrouped lambda arrow remains a
// boundary so default in an explicit collection callback applies to the
// callback result rather than the collection source.
func findDefaultLeftOperandStartBytes(result []byte) int {
	pos := len(result) - 1
	for pos >= 0 && isWhitespace(result[pos]) {
		pos--
	}
	if pos < 0 {
		return 0
	}

	input := string(result)
	depth := 0
	for pos >= 0 {
		ch := result[pos]

		if ch == '"' {
			pos--
			for pos >= 0 {
				if result[pos] == '"' && !stringutils.IsEscapedAt(input, pos) {
					break
				}
				pos--
			}
			pos--
			continue
		}

		if depth == 0 {
			if isDefaultLeftStructuralBoundary(ch) {
				return pos + 1
			}
			if ch == '>' && pos > 0 && result[pos-1] == '-' {
				return pos + 1
			}
			if ch == '=' && isStandaloneAssignmentAt(result, pos) {
				return pos + 1
			}
		}

		switch ch {
		case ')', ']', '}':
			depth++
		case '(', '[', '{':
			depth--
			if depth < 0 {
				return pos + 1
			}
		}
		pos--
	}

	return 0
}

func isDefaultLeftStructuralBoundary(ch byte) bool {
	switch ch {
	case ',', ';', ':':
		return true
	default:
		return false
	}
}

func isStandaloneAssignmentAt(input []byte, pos int) bool {
	if pos > 0 {
		switch input[pos-1] {
		case '=', '!', '<', '>':
			return false
		}
	}
	if pos+1 < len(input) {
		switch input[pos+1] {
		case '=', '>':
			return false
		}
	}
	return true
}

// findModuloLeftOperandStartBytes includes logical, native comparison, type,
// additive, and multiplicative expressions in the left operand. DataWeave's
// infix mod binds less tightly than those operators. Lower-precedence keyword
// operators have already wrapped their operands or provide structural call
// boundaries before this scanner runs.
func findModuloLeftOperandStartBytes(result []byte) int {
	pos := len(result) - 1
	for pos >= 0 && isWhitespace(result[pos]) {
		pos--
	}
	if pos < 0 {
		return 0
	}

	depth := 0
	for pos >= 0 {
		ch := result[pos]

		if ch == '"' {
			pos--
			for pos >= 0 {
				if result[pos] == '"' && !stringutils.IsEscapedAt(string(result), pos) {
					break
				}
				pos--
			}
			pos--
			continue
		}

		if depth == 0 && isModuloLeftBoundary(result[pos]) {
			return pos + 1
		}

		switch ch {
		case ')', ']', '}':
			depth++
		case '(', '[', '{':
			depth--
			if depth < 0 {
				return 0
			}
		}
		pos--
	}

	return 0
}

func isModuloLeftBoundary(ch byte) bool {
	switch ch {
	case ',', ';', ':', '(', '[', '{':
		return true
	default:
		return false
	}
}

func shouldStopModuloRightOperand(input string, pos, _ int) bool {
	switch input[pos] {
	case ';', ':':
		return true
	default:
		return false
	}
}

// shouldStopRangeRightOperand keeps lower-precedence collection operations
// outside an infix range. Arithmetic, comparisons, and type operators remain
// part of the upper bound, matching DataWeave's grouping, while a completed
// range can become the source of an array pipeline.
func shouldStopRangeRightOperand(input string, pos, operandStart int) bool {
	if !isWhitespace(input[pos]) {
		return false
	}
	trimStart, trimEnd := trimSpaceBounds(input, operandStart, pos)
	if trimStart >= trimEnd {
		return false
	}

	operatorStart := pos
	for operatorStart < len(input) && isWhitespace(input[operatorStart]) {
		operatorStart++
	}

	for _, operator := range CollectionOperators {
		if isRangeDownstreamKeywordAt(input, operatorStart, operator) {
			return true
		}
	}
	for _, operator := range []string{"find", "contains", "joinBy"} {
		if isRangeDownstreamKeywordAt(input, operatorStart, operator) {
			return true
		}
	}

	return strings.HasPrefix(input[operatorStart:], "++ ") ||
		strings.HasPrefix(input[operatorStart:], "-- ")
}

func isRangeDownstreamKeywordAt(input string, start int, operator string) bool {
	if start+len(operator) > len(input) || input[start:start+len(operator)] != operator {
		return false
	}
	end := start + len(operator)
	return end < len(input) && (isWhitespace(input[end]) || input[end] == '(')
}

func scanRightOperandBounds(input string, start int, cfg rightOperandScanConfig) (int, int, int, bool) {
	end := start
	var state ScanState

	for end < len(input) {
		ch := input[end]
		if !state.InString() {
			if ch == '(' || ch == '[' || ch == '{' {
				state.Advance(ch)
				end++
				continue
			}
			if (ch == ')' || ch == ']' || ch == '}') && state.Depth() == 0 {
				break
			}
			if state.Depth() == 0 && ch == ',' {
				break
			}
			for _, stop := range cfg.StopOps {
				if state.Depth() == 0 && end+len(stop) <= len(input) && input[end:end+len(stop)] == stop {
					trimStart, trimEnd := trimSpaceBounds(input, start, end)
					return trimStart, trimEnd, end, trimStart < trimEnd
				}
			}
			if cfg.StopPredicate != nil && state.Depth() == 0 && cfg.StopPredicate(input, end, start) {
				break
			}
		}

		state.Advance(ch)
		end++
	}

	trimStart, trimEnd := trimSpaceBounds(input, start, end)
	return trimStart, trimEnd, end, trimStart < trimEnd
}

func scanTypedOperatorRightSpan(input string, start int, allowConfig bool) (typedOperatorRightSpan, bool) {
	typeStart := start
	for typeStart < len(input) && unicode.IsSpace(rune(input[typeStart])) {
		typeStart++
	}

	typeEnd, ok := scanTypeExpressionEnd(input, typeStart)
	if !ok {
		return typedOperatorRightSpan{}, false
	}

	typeExpr := input[typeStart:typeEnd]
	// Leave separator whitespace after the type in the input stream. Later
	// transforms match spaced operator tokens such as " mod " and " ++ ".
	next := typeEnd
	configArg := ""
	configStart := -1
	configEnd := -1
	if !allowConfig {
		return typedOperatorRightSpan{
			TypeExpr:  typeExpr,
			TypeStart: typeStart,
			TypeEnd:   typeEnd,
			Next:      next,
		}, true
	}

	configCandidateStart := next
	for configCandidateStart < len(input) && unicode.IsSpace(rune(input[configCandidateStart])) {
		configCandidateStart++
	}

	prefixes := []string{GoObjectPrefix, GoObjectPrefixSpace}
	for _, prefix := range prefixes {
		if configCandidateStart+len(prefix) > len(input) || input[configCandidateStart:configCandidateStart+len(prefix)] != prefix {
			continue
		}
		bracePos := configCandidateStart + len(prefix) - 1
		scanner := stringutils.NewExpressionScanner(input)
		closePos := scanner.FindMatchingCloseBracket(bracePos)
		if closePos == -1 {
			break
		}
		configStart = configCandidateStart
		configEnd = closePos + 1
		configArg = input[configStart:configEnd]
		next = closePos + 1
		break
	}

	return typedOperatorRightSpan{
		TypeExpr:    typeExpr,
		TypeStart:   typeStart,
		TypeEnd:     typeEnd,
		ConfigArg:   configArg,
		ConfigStart: configStart,
		ConfigEnd:   configEnd,
		Next:        next,
	}, true
}

func scanTypeExpressionEnd(input string, start int) (int, bool) {
	end, ok := scanTypeNameEnd(input, start)
	if !ok {
		return start, false
	}
	if end < len(input) && input[end] == '?' {
		end++
	}

	for {
		separatorStart := end
		pos := end
		for pos < len(input) && unicode.IsSpace(rune(input[pos])) {
			pos++
		}
		if pos >= len(input) || input[pos] != '|' || pos+1 < len(input) && input[pos+1] == '|' {
			return end, true
		}

		pos++
		for pos < len(input) && unicode.IsSpace(rune(input[pos])) {
			pos++
		}
		nextEnd, nextOK := scanTypeNameEnd(input, pos)
		if !nextOK {
			return separatorStart, true
		}
		end = nextEnd
		if end < len(input) && input[end] == '?' {
			end++
		}
	}
}

func scanTypeNameEnd(input string, start int) (int, bool) {
	if start >= len(input) || !IsIdentifierStart(input[start]) {
		return start, false
	}

	end := start + 1
	for end < len(input) && IsIdentifierPart(input[end]) {
		end++
	}
	for end < len(input) && input[end] == '.' {
		segmentStart := end + 1
		if segmentStart >= len(input) || !IsIdentifierStart(input[segmentStart]) {
			break
		}
		end = segmentStart + 1
		for end < len(input) && IsIdentifierPart(input[end]) {
			end++
		}
	}
	return end, true
}
