package runner

import (
	"infomunge/internal/sourcemap"
)

func leadingWhitespaceOffset(line string) int {
	offset := 0
	for i := 0; i < len(line); i++ {
		if line[i] != ' ' && line[i] != '\t' {
			break
		}
		offset++
	}
	return offset
}

func attachLineContext(err error, source string, offset int, line string) error {
	return sourcemap.Identity(source).AttachLineContext(err, offset, line)
}
