package preprocessor

import "infomunge/internal/stringutils"

type mappedBuffer struct {
	bytes   []byte
	mapping []int
}

func newMappedBuffer(capacity int) *mappedBuffer {
	return &mappedBuffer{
		bytes:   make([]byte, 0, capacity),
		mapping: make([]int, 0, capacity),
	}
}

func (mb *mappedBuffer) Len() int {
	return len(mb.bytes)
}

func (mb *mappedBuffer) Truncate(n int) {
	mb.bytes = mb.bytes[:n]
	mb.mapping = mb.mapping[:n]
}

func (mb *mappedBuffer) AppendOriginal(src string, start, end int) {
	for i := start; i < end; i++ {
		mb.bytes = append(mb.bytes, src[i])
		mb.mapping = append(mb.mapping, i)
	}
}

func (mb *mappedBuffer) AppendLiteral(lit string, originalPos int) {
	for i := 0; i < len(lit); i++ {
		mb.bytes = append(mb.bytes, lit[i])
		mb.mapping = append(mb.mapping, originalPos)
	}
}

func (mb *mappedBuffer) AppendBytes(data []byte, mapping []int) {
	mb.bytes = append(mb.bytes, data...)
	mb.mapping = append(mb.mapping, mapping...)
}

func (mb *mappedBuffer) Slice(start int) ([]byte, []int) {
	data := make([]byte, len(mb.bytes[start:]))
	copy(data, mb.bytes[start:])
	mapping := make([]int, len(mb.mapping[start:]))
	copy(mapping, mb.mapping[start:])
	return data, mapping
}

func (mb *mappedBuffer) String() string {
	return string(mb.bytes)
}

func findLeftOperandStartBytesWithStops(result []byte, stops []byte) int {
	return findLeftOperandStartBytesWithIgnoredOperators(result, stops, nil)
}

func findLeftOperandStartBytesWithIgnoredOperators(result []byte, stops []byte, ignoredOps []string) int {
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

		if depth == 0 {
			if operatorStart, ok := ignoredOperatorEndingAt(result, pos, ignoredOps); ok {
				pos = operatorStart - 1
				continue
			}
			for _, stop := range stops {
				if ch == stop {
					if isSignedDecimalExponentBytes(result, pos) {
						break
					}
					return pos + 1
				}
			}
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

func isSignedDecimalExponentBytes(result []byte, signPos int) bool {
	if signPos < 2 || signPos+1 >= len(result) {
		return false
	}
	if result[signPos] != '+' && result[signPos] != '-' {
		return false
	}
	if result[signPos-1] != 'e' && result[signPos-1] != 'E' {
		return false
	}
	return isASCIIDigit(result[signPos-2]) && isASCIIDigit(result[signPos+1])
}

func isASCIIDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func ignoredOperatorEndingAt(result []byte, pos int, operators []string) (int, bool) {
	for _, operator := range operators {
		tokenStart := 0
		for tokenStart < len(operator) && isWhitespace(operator[tokenStart]) {
			tokenStart++
		}
		tokenEnd := len(operator)
		for tokenEnd > tokenStart && isWhitespace(operator[tokenEnd-1]) {
			tokenEnd--
		}
		token := operator[tokenStart:tokenEnd]
		start := pos - len(token) + 1
		if token == "" || start < 0 || string(result[start:pos+1]) != token {
			continue
		}
		if tokenStart > 0 && (start == 0 || !isWhitespace(result[start-1])) {
			continue
		}
		if tokenEnd < len(operator) && (pos+1 >= len(result) || !isWhitespace(result[pos+1])) {
			continue
		}
		return start, true
	}
	return 0, false
}

func defaultStopBytes(extraStops []rune) []byte {
	stops := make([]byte, 0, len(stringutils.DefaultOperatorStops)+len(extraStops))
	for _, stop := range stringutils.DefaultOperatorStops {
		stops = append(stops, byte(stop))
	}
	for _, stop := range extraStops {
		stops = append(stops, byte(stop))
	}
	return stops
}

func minimalStopBytes() []byte {
	stops := make([]byte, 0, len(stringutils.MinimalStops))
	for _, stop := range stringutils.MinimalStops {
		stops = append(stops, byte(stop))
	}
	return stops
}

func trimSpaceBounds(s string, start, end int) (int, int) {
	for start < end && isWhitespace(s[start]) {
		start++
	}
	for end > start && isWhitespace(s[end-1]) {
		end--
	}
	return start, end
}
