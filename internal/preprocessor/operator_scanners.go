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
	" splitBy ", " joinBy ", " replace ",
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
	leftStart := findLeftOperandStartBytesWithStops(result, defaultStopBytes([]rune{':', '&', '|'}))
	if leftStart >= len(result) {
		return leftStart
	}

	leftOp := string(result[leftStart:])
	lastBoundary := -1
	for _, op := range typedOperatorLeftBoundaryOps {
		pos := strings.LastIndex(leftOp, op)
		if pos > lastBoundary {
			lastBoundary = pos + len(op)
		}
	}
	if lastBoundary > 0 {
		return leftStart + lastBoundary
	}
	return leftStart
}

// findModuloLeftOperandStartBytes includes additive and multiplicative
// arithmetic in the left operand. DataWeave's infix mod binds less tightly
// than those operators, but more tightly than comparisons and logical
// operators.
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
	case '=', '<', '>', '!', '&', '|', ',', ';', ':', '(', '[', '{':
		return true
	default:
		return false
	}
}

func shouldStopModuloRightOperand(input string, pos, _ int) bool {
	switch input[pos] {
	case '=', '<', '>', '!', '&', '|', ';', ':':
		return true
	default:
		return false
	}
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
	end := start
	for end < len(input) && IsTypeExprChar(input[end]) {
		if allowConfig && end+4 <= len(input) && input[end:end+4] == "map[" {
			break
		}
		end++
	}

	typeStart, typeEnd := trimSpaceBounds(input, start, end)
	if typeStart >= typeEnd {
		return typedOperatorRightSpan{}, false
	}

	typeExpr := input[typeStart:typeEnd]
	next := end
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
