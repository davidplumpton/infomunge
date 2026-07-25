package preprocessor

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func scanLambdaBody(runes []rune, start int) (end int, hasArrow bool) {
	return scanLambdaBodyWithOptions(runes, start, true)
}

func scanCollectionLambdaBody(runes []rune, start int) int {
	end, _ := scanLambdaBodyWithOptions(runes, start, false)
	return end
}

// scanLambdaBodyWithOptions finds where a lambda body segment ends while tracking nested
// groups and quoted strings. Configured infix operators apply to the completed collection
// result, while ordinary arithmetic, comparisons, logical operators, and type operators
// remain part of the lambda body. When detectArrow is true, explicit arrows short-circuit
// with hasArrow=true for callers that should skip implicit-lambda rewrites.
func scanLambdaBodyWithOptions(runes []rune, start int, detectArrow bool) (end int, hasArrow bool) {
	var state ScanState
	wasInGroup := false

	for i := start; i < len(runes); i++ {
		ch := runes[i]
		if state.InString() {
			state.AdvanceRune(ch)
			continue
		}
		if state.Depth() == 0 && unicode.IsSpace(ch) {
			j := skipLambdaWhitespace(runes, i+1)
			if isCollectionLambdaResultOperatorAtRunes(runes, j) {
				return i, false
			}
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
				j := skipLambdaWhitespace(runes, i+1)
				if detectArrow && j+1 < len(runes) && runes[j] == '-' && runes[j+1] == '>' {
					continue
				}
				if isCollectionLambdaResultOperatorAtRunes(runes, j) {
					return i + 1, false
				}
			}
		case ',':
			if state.Depth() == 0 {
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

func skipLambdaWhitespace(runes []rune, start int) int {
	for start < len(runes) && unicode.IsSpace(runes[start]) {
		start++
	}
	return start
}

func isCollectionLambdaResultOperatorAtRunes(runes []rune, pos int) bool {
	for _, operator := range CollectionOperators {
		if isDelimitedLambdaOperatorAtRunes(runes, pos, operator) {
			return true
		}
	}
	return isNonCollectionLambdaResultOperatorAtRunes(runes, pos)
}

func isNonCollectionLambdaResultOperatorAtRunes(runes []rune, pos int) bool {
	for _, config := range binaryOperatorConfigs {
		if isDelimitedLambdaOperatorAtRunes(runes, pos, strings.TrimSpace(config.Operator)) {
			return true
		}
	}
	return isDelimitedLambdaOperatorAtRunes(runes, pos, "replace")
}

func isDelimitedLambdaOperatorAtRunes(runes []rune, pos int, operator string) bool {
	operatorRunes := []rune(operator)
	if len(operatorRunes) == 0 || pos < 0 || pos+len(operatorRunes) > len(runes) {
		return false
	}
	if string(runes[pos:pos+len(operatorRunes)]) != operator {
		return false
	}

	end := pos + len(operatorRunes)
	if end == len(runes) {
		return true
	}
	return unicode.IsSpace(runes[end]) || runes[end] == '('
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
