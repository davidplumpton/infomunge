package preprocessor

import (
	"strings"
	"unicode"

	"infomunge/internal/stringutils"
)

// replaceKeyAttributes converts "Key @(Attributes): Value" to "Key: __with_attrs(Value, Attributes)".
func replaceKeyAttributes(s string) string {
	var result []rune
	sc := stringutils.NewExpressionScanner(s)

	for sc.Pos() < len(s) {
		if !sc.IsInString() && sc.Peek() == '@' && sc.Pos()+1 < len(s) && s[sc.Pos()+1] == '(' {
			if info, ok := parseKeyAttributeInfo(s, sc, result); ok {
				if !info.hasColon {
					sc.SetPos(info.attrEnd + 1)
					result = appendKeyAttributesLiteral(result, info.attrs)
					continue
				}
				result = rewriteKeyAttributesExpression(s, sc, result, info)
				continue
			}
		}
		result = append(result, sc.NextRune())
	}
	return string(result)
}

type keyAttributeInfo struct {
	keyStart int
	key      string
	attrs    string
	attrEnd  int
	colonPos int
	hasColon bool
}

func parseKeyAttributeInfo(s string, sc *stringutils.ExpressionScanner, result []rune) (keyAttributeInfo, bool) {
	start := stringutils.FindLeftOperandStart(result, nil)
	if start >= len(result) {
		return keyAttributeInfo{}, false
	}

	keyStr := strings.TrimSpace(string(result[start:]))
	if !isPotentialKey(keyStr) {
		return keyAttributeInfo{}, false
	}

	openParen := sc.Pos() + 1
	attrEnd := sc.FindMatchingCloseBracket(openParen)
	if attrEnd == -1 {
		return keyAttributeInfo{}, false
	}

	attrs := strings.TrimSpace(s[openParen+1 : attrEnd])
	colonPos, hasColon := findKeyAttributeColon(s, attrEnd+1)
	return keyAttributeInfo{
		keyStart: start,
		key:      keyStr,
		attrs:    attrs,
		attrEnd:  attrEnd,
		colonPos: colonPos,
		hasColon: hasColon,
	}, true
}

func findKeyAttributeColon(s string, pos int) (int, bool) {
	for pos < len(s) && (s[pos] == ' ' || s[pos] == '\t' || s[pos] == '\n' || s[pos] == '\r') {
		pos++
	}
	if pos < len(s) && s[pos] == ':' {
		return pos, true
	}
	return -1, false
}

func appendKeyAttributesLiteral(result []rune, attrs string) []rune {
	result = append(result, '@')
	result = append(result, '(')
	result = append(result, []rune(attrs)...)
	result = append(result, ')')
	return result
}

func rewriteKeyAttributesExpression(s string, sc *stringutils.ExpressionScanner, result []rune, info keyAttributeInfo) []rune {
	result = result[:info.keyStart]
	result = append(result, []rune(quoteKeyAttributeKey(info.key))...)
	result = append(result, ':')
	result = append(result, []rune(" __with_attrs(")...)

	sc.SetPos(info.colonPos + 1)
	sc.SkipWhitespace()

	val := readKeyAttributeValue(s, sc)
	result = append(result, []rune(val)...)
	result = append(result, []rune(", ")...)
	result = append(result, []rune(rewriteKeyAttributesObject(info.attrs))...)
	result = append(result, ')')
	return result
}

func quoteKeyAttributeKey(key string) string {
	if strings.HasPrefix(key, "\"") || strings.HasPrefix(key, "'") || strings.HasPrefix(key, "(") {
		return key
	}
	return "\"" + key + "\""
}

func readKeyAttributeValue(s string, sc *stringutils.ExpressionScanner) string {
	valStart := sc.Pos()
	depth := 0
	inStr := false
	for sc.Pos() < len(s) {
		ch := s[sc.Pos()]
		if ch == '"' && (sc.Pos() == 0 || s[sc.Pos()-1] != '\\') {
			inStr = !inStr
		}
		if !inStr {
			if ch == '(' || ch == '[' || ch == '{' {
				depth++
			} else if ch == ')' || ch == ']' || ch == '}' {
				if depth == 0 {
					break
				}
				depth--
			} else if ch == ',' && depth == 0 {
				break
			}
		}
		sc.Advance(1)
	}
	return strings.TrimSpace(s[valStart:sc.Pos()])
}

func rewriteKeyAttributesObject(attrs string) string {
	attrRewriter := newRewriter("{"+attrs+"}", Options{})
	rewrittenAttrs, _, rewriteErr := attrRewriter.Rewrite()
	if rewriteErr != nil {
		return "{" + attrs + "}"
	}
	return rewrittenAttrs
}

func isPotentialKey(s string) bool {
	if len(s) == 0 {
		return false
	}
	if s[0] == '"' || s[0] == '\'' || s[0] == '(' {
		return true
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '#' {
			return false
		}
	}
	return true
}
