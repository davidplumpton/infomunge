package preprocessor

import "unicode/utf8"

func scanLambdaBody(runes []rune, start int) (end int, hasArrow bool) {
	return scanLambdaBodyWithOptions(runes, start, true)
}

func scanCollectionLambdaBody(runes []rune, start int) int {
	end, _ := scanLambdaBodyWithOptions(runes, start, false)
	return end
}

// scanLambdaBodyWithOptions finds where a lambda body segment ends while tracking nested
// groups and quoted strings. When detectArrow is true, explicit arrows short-circuit with
// hasArrow=true for callers that should skip implicit-lambda rewrites.
func scanLambdaBodyWithOptions(runes []rune, start int, detectArrow bool) (end int, hasArrow bool) {
	var state ScanState
	wasInGroup := false

	for i := start; i < len(runes); i++ {
		ch := runes[i]
		if state.InString() {
			state.AdvanceRune(ch)
			continue
		}
		switch ch {
		case '(', '[', '{':
			state.AdvanceRune(ch)
			wasInGroup = true
		case ')', ']', '}':
			if state.Depth() == 0 {
				return i, false
			}
			state.AdvanceRune(ch)
			if state.Depth() == 0 && wasInGroup {
				j := i + 1
				for j < len(runes) && (runes[j] == ' ' || runes[j] == '\t') {
					j++
				}
				if detectArrow && j+1 < len(runes) && runes[j] == '-' && runes[j+1] == '>' {
					continue
				}
				if isCollectionOperatorAtRunes(runes, j) {
					return i + 1, false
				}
			}
		case ',':
			if state.Depth() == 0 {
				return i, false
			}
		case ' ':
			if state.Depth() == 0 && isCollectionOperatorAtRunes(runes, i+1) {
				return i, false
			}
		default:
			state.AdvanceRune(ch)
		}
		if detectArrow && matchesAt(runes, i, "->") {
			return i, true
		}
	}
	return len(runes), false
}

func runeIndexAtByteOffset(s string, byteOffset int) int {
	if byteOffset <= 0 {
		return 0
	}
	if byteOffset >= len(s) {
		return utf8.RuneCountInString(s)
	}
	return utf8.RuneCountInString(s[:byteOffset])
}

func byteOffsetAtRuneIndex(s string, runeIndex int) int {
	if runeIndex <= 0 {
		return 0
	}
	currentRune := 0
	for byteOffset := range s {
		if currentRune == runeIndex {
			return byteOffset
		}
		currentRune++
	}
	return len(s)
}
