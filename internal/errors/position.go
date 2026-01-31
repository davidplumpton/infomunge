package errors

import (
	"go/token"
	"strings"
)

// PositionFromOffset builds a token.Position from a byte offset in the source.
func PositionFromOffset(source string, offset int, filename string) token.Position {
	if offset < 0 {
		offset = 0
	}
	if offset > len(source) {
		offset = len(source)
	}

	line := 1
	col := 1
	for i := 0; i < offset && i < len(source); i++ {
		if source[i] == '\n' {
			line++
			col = 1
			continue
		}
		col++
	}

	return token.Position{
		Filename: filename,
		Offset:   offset,
		Line:     line,
		Column:   col,
	}
}

// LineTextAtOffset returns the full line of source containing the offset.
func LineTextAtOffset(source string, offset int) string {
	if source == "" {
		return ""
	}
	if offset < 0 {
		offset = 0
	}
	if offset > len(source) {
		offset = len(source)
	}

	lineStart := strings.LastIndex(source[:offset], "\n")
	if lineStart == -1 {
		lineStart = 0
	} else {
		lineStart++
	}

	lineEnd := strings.Index(source[offset:], "\n")
	if lineEnd == -1 {
		lineEnd = len(source)
	} else {
		lineEnd = offset + lineEnd
	}

	return source[lineStart:lineEnd]
}
