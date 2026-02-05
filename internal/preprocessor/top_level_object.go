package preprocessor

import (
	"strings"
	"unicode"
)

func wrapTopLevelObjectLiteral(input string) (string, bool) {
	if !shouldWrapTopLevelObjectLiteral(input) {
		return input, false
	}
	return "{" + input + "}", true
}

func shouldWrapTopLevelObjectLiteral(input string) bool {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return false
	}
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return false
	}

	var state ScanState
	for i := 0; i < len(input); i++ {
		ch := input[i]
		state.Advance(ch)
		if state.InString() {
			continue
		}

		switch ch {
		case ':':
			if state.Depth() != 0 {
				continue
			}
			if i+1 < len(input) && input[i+1] == ':' {
				continue
			}
			if !hasTopLevelKeyBeforeColon(input, i) {
				continue
			}
			return true
		}
	}

	return false
}

func hasTopLevelKeyBeforeColon(input string, colonPos int) bool {
	for j := colonPos - 1; j >= 0; j-- {
		ch := input[j]
		if unicode.IsSpace(rune(ch)) {
			continue
		}
		switch ch {
		case ',', '{', '[', '(', ':':
			return false
		default:
			return true
		}
	}
	return false
}
