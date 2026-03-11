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

func findTypedOperatorLeftOperandStart(result []rune) int {
	leftStart := stringutils.FindLeftOperandStart(result, []rune{':', '&', '|'})
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

func scanTypedOperatorRight(input string, start int, allowConfig bool) (string, string, int, bool) {
	end := start
	for end < len(input) && IsTypeExprChar(input[end]) {
		if allowConfig && end+4 <= len(input) && input[end:end+4] == "map[" {
			break
		}
		end++
	}

	typeStart, typeEnd := trimSpaceBounds(input, start, end)
	if typeStart >= typeEnd {
		return "", "", start, false
	}

	typeExpr := input[typeStart:typeEnd]
	next := end
	configArg := ""
	if !allowConfig {
		return typeExpr, configArg, next, true
	}

	configStart := next
	for configStart < len(input) && unicode.IsSpace(rune(input[configStart])) {
		configStart++
	}

	prefixes := []string{GoObjectPrefix, GoObjectPrefixSpace}
	for _, prefix := range prefixes {
		if configStart+len(prefix) > len(input) || input[configStart:configStart+len(prefix)] != prefix {
			continue
		}
		bracePos := configStart + len(prefix) - 1
		scanner := stringutils.NewExpressionScanner(input)
		closePos := scanner.FindMatchingCloseBracket(bracePos)
		if closePos == -1 {
			break
		}
		configArg = input[configStart : closePos+1]
		next = closePos + 1
		break
	}

	return typeExpr, configArg, next, true
}
