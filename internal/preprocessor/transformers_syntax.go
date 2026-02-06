package preprocessor

import (
	"fmt"
	"infomunge/internal/stringutils"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// replaceStringInterpolation converts strings with $(expr) to concatenation expressions.
func replaceStringInterpolation(s string) string {
	var result []rune
	i := 0

	for i < len(s) {
		if s[i] != '"' {
			r, size := utf8.DecodeRuneInString(s[i:])
			result = append(result, r)
			i += size
			continue
		}

		stringStart := i
		endPos, hasInterpolation := scanString(s, i)
		i = endPos

		if !hasInterpolation {
			for _, ch := range s[stringStart:i] {
				result = append(result, ch)
			}
			continue
		}

		stringContent := s[stringStart+1 : i-1]
		interpolated := interpolateString(stringContent)
		result = append(result, []rune(interpolated)...)
	}

	return string(result)
}

// scanString scans a double-quoted string.
func scanString(s string, pos int) (endPos int, hasInterpolation bool) {
	pos++
	for pos < len(s) {
		if s[pos] == '\\' && pos+1 < len(s) {
			pos += 2
			continue
		}
		if s[pos] == '"' {
			return pos + 1, hasInterpolation
		}
		if pos+1 < len(s) && s[pos] == '$' && s[pos+1] == '(' {
			hasInterpolation = true
		}
		pos++
	}
	return pos, hasInterpolation
}

// interpolateString converts string content with $(expr) to a concatenation expression.
func interpolateString(content string) string {
	var parts []string
	var literal strings.Builder
	i := 0

	for i < len(content) {
		if content[i] == '\\' && i+1 < len(content) {
			literal.WriteString(content[i : i+2])
			i += 2
			continue
		}

		if i+1 < len(content) && content[i] == '$' && content[i+1] == '(' {
			if literal.Len() > 0 {
				parts = append(parts, "\""+literal.String()+"\"")
				literal.Reset()
			}

			expr, newPos := extractInterpolationExpr(content, i+2)
			parts = append(parts, "("+expr+")")
			i = newPos
			continue
		}

		literal.WriteByte(content[i])
		i++
	}

	if literal.Len() > 0 {
		parts = append(parts, "\""+literal.String()+"\"")
	}

	return "(" + strings.Join(parts, " + ") + ")"
}

// replaceArrayRangeIndexing converts arr[start to end] to slice(arr, start, end+1)
func replaceArrayRangeIndexing(s string) string {
	// Find patterns like expr[ start to end ]
	// Replace with slice(expr, start, end+1)

	i := strings.LastIndex(s, "[")
	if i == -1 {
		return s
	}
	if i+1 >= len(s) || !unicode.IsDigit(rune(s[i+1])) && !unicode.IsLetter(rune(s[i+1])) && s[i+1] != '(' {
		return s
	}
	j := strings.LastIndex(s, "]")
	if j <= i {
		return s
	}
	bracketContent := s[i+1 : j]
	if strings.Contains(bracketContent, " to ") {
		parts := strings.Split(bracketContent, " to ")
		if len(parts) == 2 {
			start := strings.TrimSpace(parts[0])
			endStr := strings.TrimSpace(parts[1])
			arrayExpr := s[:i]
			endInt, err := strconv.Atoi(endStr)
			if err != nil {
				return s
			}
			endPlusOne := endInt + 1
			return fmt.Sprintf("slice(%s, %s, %d)", arrayExpr, start, endPlusOne)
		}
	}
	return s
}

// extractInterpolationExpr extracts the expression inside $(...).
func extractInterpolationExpr(content string, pos int) (expr string, endPos int) {
	depth := 1
	start := pos
	inStr := false

	for pos < len(content) && depth > 0 {
		ch := content[pos]
		if ch == '"' && (pos == 0 || content[pos-1] != '\\') {
			inStr = !inStr
		}
		if !inStr {
			if ch == '(' {
				depth++
			} else if ch == ')' {
				depth--
			}
		}
		if depth > 0 {
			pos++
		}
	}

	return content[start:pos], pos + 1
}

// replaceRecursiveDescent converts "obj..field" to "__deep(obj, \"field\")"
func replaceRecursiveDescent(s string) string {
	var result []rune
	sc := NewScanner(s)

	for sc.Pos() < len(s) {
		if !sc.IsInString() && sc.Peek2() == ".*" {
			leftExpr, newResult, ok := extractLeftOperand(result)
			if !ok {
				result = append(result, sc.NextRune())
				continue
			}
			result = newResult
			nextPos := sc.Pos() + 2
			// Check if there's a field name after .*
			if nextPos < len(s) && (IsIdentifierStart(s[nextPos]) || s[nextPos] == '@') {
				// Extract the field name
				fieldStart := nextPos
				fieldEnd := fieldStart
				for fieldEnd < len(s) && IsIdentifierPart(s[fieldEnd]) {
					fieldEnd++
				}
				fieldName := s[fieldStart:fieldEnd]
				// Transform to __multival(obj, "field")
				result = append(result, []rune("__multival(")...)
				result = append(result, []rune(leftExpr)...)
				result = append(result, []rune(", \"")...)
				result = append(result, []rune(fieldName)...)
				result = append(result, []rune("\")")...)
				sc.SetPos(fieldEnd)
			} else {
				// No field name, just get all values
				result = append(result, []rune("__objvalues(")...)
				result = append(result, []rune(leftExpr)...)
				result = append(result, []rune(")")...)
				sc.Advance(2)
			}
			continue
		}

		if !sc.IsInString() && sc.Peek2() == ".." {
			fieldStart := sc.Pos() + 2
			for fieldStart < len(s) && unicode.IsSpace(rune(s[fieldStart])) {
				fieldStart++
			}
			if fieldStart >= len(s) || s[fieldStart] == '.' || !IsIdentifierStart(s[fieldStart]) {
				result = append(result, sc.NextRune())
				continue
			}
			fieldEnd := fieldStart + 1
			for fieldEnd < len(s) && IsIdentifierPart(s[fieldEnd]) {
				fieldEnd++
			}
			fieldName := s[fieldStart:fieldEnd]
			leftExpr, newResult, ok := extractLeftOperand(result)
			if !ok {
				result = append(result, sc.NextRune())
				continue
			}
			result = newResult
			result = append(result, []rune("__deep(")...)
			result = append(result, []rune(leftExpr)...)
			result = append(result, []rune(", \"")...)
			result = append(result, []rune(fieldName)...)
			result = append(result, []rune("\")")...)
			sc.SetPos(fieldEnd)
			continue
		}
		result = append(result, sc.NextRune())
	}
	return string(result)
}

// replaceFilterSelectors converts "obj[?(expr)]" to "__filter_selector(obj, __lambda(\"__arg, __idx\", expr))".
func replaceFilterSelectors(s string) string {
	var result []rune
	sc := NewScanner(s)

	for sc.Pos() < len(s) {
		if !sc.IsInString() && sc.Peek() == '[' && sc.Pos()+2 < len(s) && s[sc.Pos()+1] == '?' && s[sc.Pos()+2] == '(' {
			leftExpr, newResult, ok := extractLeftOperand(result)
			if !ok {
				result = append(result, sc.NextRune())
				continue
			}

			openParen := sc.Pos() + 2
			sc.SetPos(openParen)
			closeParen := sc.FindMatchingCloseBracket(sc.Pos())
			if closeParen == -1 {
				sc.SetPos(openParen - 2)
				result = append(result, sc.NextRune())
				continue
			}

			closeBracket := closeParen + 1
			for closeBracket < len(s) && unicode.IsSpace(rune(s[closeBracket])) {
				closeBracket++
			}
			if closeBracket >= len(s) || s[closeBracket] != ']' {
				sc.SetPos(openParen - 2)
				result = append(result, sc.NextRune())
				continue
			}

			predicate := strings.TrimSpace(s[openParen+1 : closeParen])
			if predicate == "" {
				predicate = "false"
			}
			predicate = replaceImplicitParam(predicate, "$$", "__idx")
			predicate = replaceImplicitParam(predicate, "$", "__arg")

			result = newResult
			result = append(result, []rune("__filter_selector(")...)
			result = append(result, []rune(leftExpr)...)
			result = append(result, []rune(", __lambda(\"__arg, __idx\", ")...)
			result = append(result, []rune(predicate)...)
			result = append(result, []rune("))")...)

			sc.SetPos(closeBracket + 1)
			continue
		}
		result = append(result, sc.NextRune())
	}

	return string(result)
}

// replaceMetadataSelectors converts "obj.^meta" to "__metadata(obj, \"meta\")".
func replaceMetadataSelectors(s string) string {
	var result []rune
	sc := NewScanner(s)

	for sc.Pos() < len(s) {
		if !sc.IsInString() && sc.Peek2() == ".^" {
			metaStart := sc.Pos() + 2
			if metaStart < len(s) && IsIdentifierStart(s[metaStart]) {
				metaEnd := metaStart + 1
				for metaEnd < len(s) && IsIdentifierPart(s[metaEnd]) {
					metaEnd++
				}

				leftExpr, newResult, ok := extractLeftOperand(result)
				if !ok {
					result = append(result, sc.NextRune())
					continue
				}
				metaName := s[metaStart:metaEnd]

				result = newResult
				result = append(result, []rune("__metadata(")...)
				result = append(result, []rune(leftExpr)...)
				result = append(result, []rune(", \"")...)
				result = append(result, []rune(metaName)...)
				result = append(result, []rune("\")")...)
				sc.SetPos(metaEnd)
				continue
			}
		}
		result = append(result, sc.NextRune())
	}

	return string(result)
}

// replaceDotNotation converts "obj.field" to "obj[\"field\"]"
func replaceDotNotation(s string) string {
	var result []rune
	sc := NewScanner(s)

	for sc.Pos() < len(s) {
		if !sc.IsInString() && sc.Peek() == '.' && sc.Pos()+1 < len(s) && (IsIdentifierStart(s[sc.Pos()+1]) || s[sc.Pos()+1] == '@' || s[sc.Pos()+1] == '#') {
			if len(result) > 0 {
				lastChar := result[len(result)-1]
				lastCharByte := byte(lastChar)
				if IsIdentifierPart(lastCharByte) || lastChar == ')' || lastChar == ']' || lastChar == '"' || lastChar == '}' {
					sc.Next()
					isAt := false
					isNamespace := false
					if sc.Peek() == '@' {
						isAt = true
						sc.Next()
					} else if sc.Peek() == '#' {
						isNamespace = true
						sc.Next()
					}
					fieldName := "#"
					if !isNamespace {
						fieldStart := sc.Pos()
						for sc.Pos() < len(s) && IsIdentifierPart(s[sc.Pos()]) {
							sc.Advance(1)
						}
						fieldName = s[fieldStart:sc.Pos()]
					}
					isOptional := false
					isAssert := false
					if !isNamespace && sc.Pos() < len(s) && s[sc.Pos()] == '?' {
						isOptional = true
						sc.Advance(1)
					} else if !isNamespace && sc.Pos() < len(s) && s[sc.Pos()] == '!' {
						isAssert = true
						sc.Advance(1)
					}
					selectorSuffix := rune(0)
					if isOptional {
						selectorSuffix = '?'
					} else if isAssert {
						selectorSuffix = '!'
					}
					result = appendDotNotationSelector(result, isAt, fieldName, selectorSuffix)
					continue
				}
			}
		}
		result = append(result, sc.NextRune())
	}
	return string(result)
}

func appendDotNotationSelector(result []rune, isAt bool, fieldName string, suffix rune) []rune {
	result = append(result, '[')
	result = append(result, '"')
	if isAt {
		result = append(result, '@')
	}
	for _, ch := range fieldName {
		result = append(result, ch)
	}
	if suffix != 0 {
		result = append(result, suffix)
	}
	result = append(result, '"')
	result = append(result, ']')
	return result
}

// replaceCaseStatements converts case statements.
func replaceCaseStatements(s string) string {
	var result []rune
	sc := NewScanner(s)

	for sc.Pos() < len(s) {
		casePos, kwLen, isPatternMatchKeywordFound := isPatternMatchKeyword(s, sc)
		if isPatternMatchKeywordFound {
			if processed, ok := processCaseStatement(s, sc, result, casePos, kwLen); ok {
				result = processed
				continue
			}
		}
		result = append(result, sc.NextRune())
	}
	return string(result)
}

// processCaseStatement processes a single case statement.
func processCaseStatement(s string, sc *StringScanner, result []rune, casePos int, kwLen int) ([]rune, bool) {
	afterCase := casePos + kwLen
	bracePos, ok := findCaseBrace(s, afterCase)
	if !ok {
		return result, false
	}

	exprStart, ok := findExprStart(result)
	if !ok {
		return result, false
	}

	expr := strings.TrimSpace(string(result[exprStart:]))
	result = result[:exprStart]

	sc.SetPos(bracePos)
	closeBrace := sc.FindMatchingCloseBracket(sc.Pos())
	if closeBrace == -1 {
		return result, false
	}

	itemsStr := s[sc.Pos()+1 : closeBrace]
	sc.SetPos(closeBrace + 1)

	caseItems := parseCaseItems(itemsStr)

	result = append(result, []rune("__case(")...)
	result = append(result, []rune(expr)...)
	result = append(result, []rune(", []interface{} {")...)
	result = append(result, []rune(strings.Join(caseItems, ", "))...)
	result = append(result, []rune("})")...)
	return result, true
}

func parseCaseItems(s string) []string {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, ",") {
		s = s[:len(s)-1]
	}

	var items []string
	sc := NewScanner(s)

	for sc.Pos() < len(s) {
		sc.SkipWhitespace()
		if sc.Pos() >= len(s) {
			break
		}

		start := sc.Pos()
		for sc.Pos() < len(s) {
			if !sc.IsInString() && sc.Peek2() == "->" {
				break
			}
			sc.Next()
		}

		if sc.Pos() >= len(s) {
			break
		}

		pattern := strings.TrimSpace(s[start:sc.Pos()])
		if strings.HasPrefix(pattern, "case ") {
			pattern = strings.TrimSpace(strings.TrimPrefix(pattern, "case "))
		}

		sc.Advance(2)
		sc.SkipWhitespace()

		bodyStart := sc.Pos()
		for sc.Pos() < len(s) {
			if !sc.IsInString() && sc.Depth() == 0 {
				peek := sc.Peek()
				if peek == ',' || peek == '\n' {
					break
				}
				if strings.HasPrefix(s[sc.Pos():], "case ") || strings.HasPrefix(s[sc.Pos():], "else ") {
					break
				}
			}
			sc.Next()
		}

		body := strings.TrimSpace(s[bodyStart:sc.Pos()])
		if sc.Pos() < len(s) && sc.Peek() == ',' {
			sc.Next()
		}

		items = append(items, fmt.Sprintf(`map[string]interface{}{ "pattern": %q, "result": %s }`, pattern, body))
	}

	return items
}

// replaceModuleCall converts module calls.
func replaceModuleCall(s string) string {
	var result []rune
	sc := NewScanner(s)

	for sc.Pos() < len(s) {
		if sc.IsInString() {
			result = append(result, sc.NextRune())
			continue
		}

		if sc.Pos()+2 <= len(s) && s[sc.Pos():sc.Pos()+2] == "::" {
			modEnd := len(result)
			modStart := modEnd - 1
			for modStart >= 0 && IsIdentifierPart(byte(result[modStart])) {
				modStart--
			}
			modStart++

			if modStart >= modEnd {
				result = append(result, sc.NextRune())
				continue
			}

			modName := string(result[modStart:modEnd])
			result = result[:modStart]
			sc.Advance(2)
			funcStart := sc.Pos()
			for sc.Pos() < len(s) && IsIdentifierPart(s[sc.Pos()]) {
				sc.Advance(1)
			}

			if sc.Pos() == funcStart {
				result = append(result, []rune(modName+"::")...)
				continue
			}

			funcName := s[funcStart:sc.Pos()]
			for sc.Pos() < len(s) && (s[sc.Pos()] == ' ' || s[sc.Pos()] == '\t') {
				sc.Advance(1)
			}

			if sc.Pos() >= len(s) || s[sc.Pos()] != '(' {
				result = append(result, []rune(modName+"::"+funcName)...)
				continue
			}

			closeIdx := sc.FindMatchingCloseBracket(sc.Pos())
			if closeIdx < 0 {
				result = append(result, []rune(modName+"::"+funcName)...)
				continue
			}

			args := s[sc.Pos()+1 : closeIdx]
			sc.SetPos(closeIdx + 1)
			result = append(result, []rune("__modcall(\"")...)
			result = append(result, []rune(modName)...)
			result = append(result, []rune("\", \"")...)
			result = append(result, []rune(funcName)...)
			result = append(result, []rune("\"")...)
			if strings.TrimSpace(args) != "" {
				result = append(result, []rune(", ")...)
				result = append(result, []rune(args)...)
			}
			result = append(result, ')')
			continue
		}
		result = append(result, sc.NextRune())
	}
	return string(result)
}

// replaceMultiStatementSequences converts sequences of statements to __seq calls.
func replaceMultiStatementSequences(s string) string {
	lines := strings.Split(s, "\n")
	var nonEmptyLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			nonEmptyLines = append(nonEmptyLines, trimmed)
		}
	}
	if len(nonEmptyLines) <= 1 {
		return s
	}
	result := "__seq("
	for i, stmt := range nonEmptyLines {
		result += stmt
		if i < len(nonEmptyLines)-1 {
			result += ", "
		}
	}
	result += ")"
	return result
}

// replaceAssignmentExpressions converts assignment expressions to __assign calls.
func replaceAssignmentExpressions(s string) string {
	var result []rune
	sc := NewScanner(s)

	for sc.Pos() < len(s) {
		if sc.IsInString() {
			result = append(result, sc.NextRune())
			continue
		}

		if s[sc.Pos()] == '=' {
			prevChar := byte(0)
			if len(result) > 0 {
				prevChar = byte(result[len(result)-1])
			}
			nextChar := byte(0)
			if sc.Pos()+1 < len(s) {
				nextChar = s[sc.Pos()+1]
			}

			if isComparisonOperatorEq(prevChar, nextChar) {
				result = append(result, sc.NextRune())
				continue
			}

			varName, varStart, _, ok := extractVariableNameAtPosition(result)
			if !ok {
				result = append(result, sc.NextRune())
				continue
			}

			result = result[:varStart]
			sc.Advance(1)
			for sc.Pos() < len(s) && (s[sc.Pos()] == ' ' || s[sc.Pos()] == '\t') {
				sc.Advance(1)
			}

			valueStart := sc.Pos()
			valueEnd := findValueExpressionEnd(s, sc)

			if valueStart >= valueEnd {
				result = append(result, []rune(varName)...)
				result = append(result, '=')
				continue
			}

			value := strings.TrimSpace(s[valueStart:valueEnd])
			result = append(result, []rune("__assign(")...)
			result = append(result, []rune(strconv.Quote(varName))...)
			result = append(result, []rune(", ")...)
			result = append(result, []rune(value)...)
			result = append(result, ')')
			continue
		}
		result = append(result, sc.NextRune())
	}
	return string(result)
}

// replaceKeyAttributes converts "Key @(Attributes): Value" to "Key: __with_attrs(Value, Attributes)"
func replaceKeyAttributes(s string) string {
	var result []rune
	sc := NewScanner(s)

	for sc.Pos() < len(s) {
		if !sc.IsInString() && sc.Peek() == '@' && sc.Pos()+1 < len(s) && s[sc.Pos()+1] == '(' {
			// Found @(
			// Try to find the key to the left. A key can be an identifier or a string.
			start := stringutils.FindLeftOperandStart(result, nil)
			if start < len(result) {
				keyStr := strings.TrimSpace(string(result[start:]))
				// Verify it's a potential key: identifier or string
				if isPotentialKey(keyStr) {
					sc.Advance(1) // Skip @
					attrEnd := sc.FindMatchingCloseBracket(sc.Pos())
					if attrEnd != -1 {
						attrs := strings.TrimSpace(s[sc.Pos()+1 : attrEnd])
						// Look for : after attrEnd
						pos := attrEnd + 1
						for pos < len(s) && (s[pos] == ' ' || s[pos] == '\t' || s[pos] == '\n' || s[pos] == '\r') {
							pos++
						}
						if pos < len(s) && s[pos] == ':' {
							// Found it! Transform to Key: __with_attrs(Value, Attrs)
							result = result[:start]
							// Ensure Key is quoted if it was an unquoted identifier
							if !strings.HasPrefix(keyStr, "\"") && !strings.HasPrefix(keyStr, "'") && !strings.HasPrefix(keyStr, "(") {
								keyStr = "\"" + keyStr + "\""
							}
							result = append(result, []rune(keyStr)...)
							result = append(result, ':')
							result = append(result, []rune(" __with_attrs(")...)

							// Skip the : and whitespace
							sc.SetPos(pos + 1)
							sc.SkipWhitespace()

							// Now we need to find the end of the Value.
							// Value ends at the next comma or closing brace at the same depth.
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
							val := strings.TrimSpace(s[valStart:sc.Pos()])
							result = append(result, []rune(val)...)
							result = append(result, []rune(", ")...)

							// Rewrite attributes using the rewriter to handle unquoted keys and braces
							attrRewriter := newRewriter("{"+attrs+"}", Options{})
							rewrittenAttrs, _, _ := attrRewriter.Rewrite()
							result = append(result, []rune(rewrittenAttrs)...)

							result = append(result, ')')
							// sc.Pos() is at the delimiter (comma or closing brace)
							continue
						} else {
							// Not followed by :, reset scanner to AFTER the @( block
							sc.SetPos(attrEnd + 1)
							result = append(result, '@')
							result = append(result, '(')
							result = append(result, []rune(attrs)...)
							result = append(result, ')')
							continue
						}
					}
				}
			}
		}
		result = append(result, sc.NextRune())
	}
	return string(result)
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
