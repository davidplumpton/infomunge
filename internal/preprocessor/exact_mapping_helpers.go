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
			for _, stop := range stops {
				if ch == stop {
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
