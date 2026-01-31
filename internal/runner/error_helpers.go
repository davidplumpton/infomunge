package runner

import (
	"fmt"
	"strings"

	unifiederrors "infomunge/internal/errors"
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
	if err == nil {
		return nil
	}

	unifiedErr, ok := err.(*unifiederrors.Error)
	if !ok {
		return err
	}
	if unifiedErr.Position != nil {
		return err
	}

	lineText := strings.TrimSpace(line)
	if lineText == "" {
		lineText = strings.TrimSpace(unifiederrors.LineTextAtOffset(source, offset))
	}
	if lineText != "" && !strings.Contains(unifiedErr.Message, "(line:") {
		unifiedErr.Message = fmt.Sprintf("%s (line: %s)", unifiedErr.Message, lineText)
	}

	pos := unifiederrors.PositionFromOffset(source, offset, "")
	unifiedErr.WithPosition(pos)
	return unifiedErr
}
