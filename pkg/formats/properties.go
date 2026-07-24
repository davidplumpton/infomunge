package formats

import (
	"infomunge/pkg/values"
	"strings"
)

func init() {
	RegisterReader("text/x-java-properties", readProperties)
	RegisterExtension(".properties", "text/x-java-properties")
}

// readProperties parses a practical subset of Java .properties syntax:
// - comments: lines starting with # or !
// - separators: first unescaped '=', ':', or whitespace
// - escapes: \\, \=, \:, \t, \n, \r, \f
// - continuations: trailing unescaped '\' joins with the next line
func readProperties(content string) (interface{}, error) {
	result := values.NewObject(0)
	lines := strings.Split(content, "\n")
	var logicalLine strings.Builder

	for i := 0; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")

		if logicalLine.Len() > 0 {
			// Java properties continuation lines skip leading horizontal spaces.
			line = strings.TrimLeft(line, " \t\f")
		}
		logicalLine.WriteString(line)

		combined := logicalLine.String()
		if hasUnescapedLineContinuation(combined) {
			logicalLine.Reset()
			logicalLine.WriteString(combined[:len(combined)-1])
			continue
		}

		logical := strings.TrimSpace(logicalLine.String())
		logicalLine.Reset()

		if logical == "" || strings.HasPrefix(logical, "#") || strings.HasPrefix(logical, "!") {
			continue
		}

		keyPart, valuePart := splitPropertyLine(logical)
		key := unescapePropertyToken(strings.TrimSpace(keyPart))
		value := unescapePropertyToken(strings.TrimSpace(valuePart))
		values.SetObjectValue(result, key, value)
	}

	if logicalLine.Len() > 0 {
		logical := strings.TrimSpace(logicalLine.String())
		if logical != "" && !strings.HasPrefix(logical, "#") && !strings.HasPrefix(logical, "!") {
			keyPart, valuePart := splitPropertyLine(logical)
			key := unescapePropertyToken(strings.TrimSpace(keyPart))
			value := unescapePropertyToken(strings.TrimSpace(valuePart))
			values.SetObjectValue(result, key, value)
		}
	}

	return result, nil
}

func hasUnescapedLineContinuation(s string) bool {
	if s == "" {
		return false
	}
	count := 0
	for i := len(s) - 1; i >= 0 && s[i] == '\\'; i-- {
		count++
	}
	return count%2 == 1
}

func splitPropertyLine(line string) (key, value string) {
	sep := -1
	sepByWhitespace := false
	escaped := false

	for i := 0; i < len(line); i++ {
		ch := line[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '=' || ch == ':' {
			sep = i
			break
		}
		if ch == ' ' || ch == '\t' || ch == '\f' {
			sep = i
			sepByWhitespace = true
			break
		}
	}

	if sep == -1 {
		return line, ""
	}

	key = line[:sep]
	pos := sep
	if sepByWhitespace {
		for pos < len(line) && (line[pos] == ' ' || line[pos] == '\t' || line[pos] == '\f') {
			pos++
		}
		if pos < len(line) && (line[pos] == '=' || line[pos] == ':') {
			pos++
			for pos < len(line) && (line[pos] == ' ' || line[pos] == '\t' || line[pos] == '\f') {
				pos++
			}
		}
	} else {
		pos++
		for pos < len(line) && (line[pos] == ' ' || line[pos] == '\t' || line[pos] == '\f') {
			pos++
		}
	}
	value = line[pos:]
	return key, value
}

func unescapePropertyToken(s string) string {
	var out strings.Builder
	out.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			out.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 't':
			out.WriteByte('\t')
		case 'n':
			out.WriteByte('\n')
		case 'r':
			out.WriteByte('\r')
		case 'f':
			out.WriteByte('\f')
		case '\\', ':', '=':
			out.WriteByte(s[i])
		default:
			out.WriteByte(s[i])
		}
	}
	return out.String()
}
