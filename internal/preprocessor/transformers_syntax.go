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

// replaceArrayRangeIndexing converts each arr[start to end] selector to an
// inclusive range-index call. It rewrites innermost selectors first so nested
// brackets and multiple independent selectors are handled without confusing an
// array literal for the indexed expression.
func replaceArrayRangeIndexing(s string) string {
	result := s
	for {
		next, changed := replaceFirstArrayRangeIndex(result)
		if !changed {
			return result
		}
		result = next
	}
}

func replaceFirstArrayRangeIndex(s string) (string, bool) {
	var bracketStack []int
	var quote byte

	for pos := 0; pos < len(s); pos++ {
		ch := s[pos]
		if quote != 0 {
			if ch == quote && !stringutils.IsEscapedAt(s, pos) {
				quote = 0
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			quote = ch
			continue
		}

		switch ch {
		case '[':
			bracketStack = append(bracketStack, pos)
		case ']':
			if len(bracketStack) == 0 {
				continue
			}
			open := bracketStack[len(bracketStack)-1]
			bracketStack = bracketStack[:len(bracketStack)-1]
			start, end, ok := splitRangeIndexBounds(s[open+1 : pos])
			if !ok {
				continue
			}

			prefix := []rune(s[:open])
			exprStart := stringutils.FindLeftOperandStart(prefix, nil)
			operandStart := exprStart
			for operandStart < len(prefix) && unicode.IsSpace(prefix[operandStart]) {
				operandStart++
			}
			indexedExpr := strings.TrimSpace(string(prefix[operandStart:]))
			if indexedExpr == "" {
				continue
			}

			beforeExpr := string(prefix[:operandStart])
			replacement := fmt.Sprintf("__rangeIndex(%s, %s, %s)", indexedExpr, start, end)
			return beforeExpr + replacement + s[pos+1:], true
		}
	}

	return s, false
}

func splitRangeIndexBounds(content string) (string, string, bool) {
	var quote byte
	parenDepth := 0
	bracketDepth := 0
	braceDepth := 0
	rangePos := -1

	for pos := 0; pos < len(content); pos++ {
		ch := content[pos]
		if quote != 0 {
			if ch == quote && !stringutils.IsEscapedAt(content, pos) {
				quote = 0
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			quote = ch
			continue
		}

		switch ch {
		case '(':
			parenDepth++
		case ')':
			parenDepth--
		case '[':
			bracketDepth++
		case ']':
			bracketDepth--
		case '{':
			braceDepth++
		case '}':
			braceDepth--
		default:
			if parenDepth == 0 && bracketDepth == 0 && braceDepth == 0 &&
				strings.HasPrefix(content[pos:], " to ") {
				if rangePos != -1 {
					return "", "", false
				}
				rangePos = pos
				pos += len(" to ") - 1
			}
		}
	}

	if rangePos == -1 {
		return "", "", false
	}
	start := strings.TrimSpace(content[:rangePos])
	end := strings.TrimSpace(content[rangePos+len(" to "):])
	if start == "" || end == "" {
		return "", "", false
	}
	return start, end, true
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

func replaceRecursiveDescentWithMapping(s string) (string, []int) {
	buf := newMappedBuffer(len(s) + 32)
	sc := stringutils.NewExpressionScanner(s)

	for sc.Pos() < len(s) {
		if !sc.IsInString() && sc.Peek2() == ".*" {
			leftStart := findLeftOperandStartBytesWithStops(buf.bytes, defaultStopBytes(nil))
			if leftStart >= buf.Len() {
				buf.AppendOriginal(s, sc.Pos(), sc.Pos()+1)
				sc.Advance(1)
				continue
			}
			leftTrimStart, _ := trimSpaceBounds(buf.String(), leftStart, buf.Len())
			leftBytes, leftMapping := buf.Slice(leftTrimStart)
			buf.Truncate(leftStart)
			nextPos := sc.Pos() + 2
			// Check if there's a field name after .*
			if nextPos < len(s) && (IsIdentifierStart(s[nextPos]) || s[nextPos] == '@') {
				// Extract the field name
				fieldStart := nextPos
				fieldEnd := fieldStart
				for fieldEnd < len(s) && IsIdentifierPart(s[fieldEnd]) {
					fieldEnd++
				}
				// Transform to __multival(obj, "field")
				buf.AppendLiteral("__multival(", sc.Pos())
				buf.AppendBytes(leftBytes, leftMapping)
				buf.AppendLiteral(", \"", fieldStart)
				buf.AppendOriginal(s, fieldStart, fieldEnd)
				buf.AppendLiteral("\")", fieldEnd-1)
				sc.SetPos(fieldEnd)
			} else {
				// No field name, just get all values
				buf.AppendLiteral("__objvalues(", sc.Pos())
				buf.AppendBytes(leftBytes, leftMapping)
				buf.AppendLiteral(")", sc.Pos()+1)
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
				buf.AppendOriginal(s, sc.Pos(), sc.Pos()+1)
				sc.Advance(1)
				continue
			}
			fieldEnd := fieldStart + 1
			for fieldEnd < len(s) && IsIdentifierPart(s[fieldEnd]) {
				fieldEnd++
			}
			leftStart := findLeftOperandStartBytesWithStops(buf.bytes, defaultStopBytes(nil))
			if leftStart >= buf.Len() {
				buf.AppendOriginal(s, sc.Pos(), sc.Pos()+1)
				sc.Advance(1)
				continue
			}
			leftTrimStart, _ := trimSpaceBounds(buf.String(), leftStart, buf.Len())
			leftBytes, leftMapping := buf.Slice(leftTrimStart)
			buf.Truncate(leftStart)
			buf.AppendLiteral("__deep(", sc.Pos())
			buf.AppendBytes(leftBytes, leftMapping)
			buf.AppendLiteral(", \"", fieldStart)
			buf.AppendOriginal(s, fieldStart, fieldEnd)
			buf.AppendLiteral("\")", fieldEnd-1)
			sc.SetPos(fieldEnd)
			continue
		}
		buf.AppendOriginal(s, sc.Pos(), sc.Pos()+1)
		sc.Advance(1)
	}
	return buf.String(), buf.mapping
}

// replaceFilterSelectors converts "obj[?(expr)]" to "__filter_selector(obj, __lambda(\"__arg, __idx\", expr))".
func replaceFilterSelectors(s string) string {
	var result []rune
	sc := stringutils.NewExpressionScanner(s)

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
			predicate, _, _ = rewriteImplicitParams(predicate)

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
	sc := stringutils.NewExpressionScanner(s)

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

func replaceDotNotationWithMapping(s string) (string, []int) {
	buf := newMappedBuffer(len(s) + 16)
	sc := stringutils.NewExpressionScanner(s)

	for sc.Pos() < len(s) {
		if !sc.IsInString() && sc.Peek() == '.' && sc.Pos()+1 < len(s) && (IsIdentifierStart(s[sc.Pos()+1]) || s[sc.Pos()+1] == '@' || s[sc.Pos()+1] == '#') {
			if buf.Len() > 0 {
				lastCharByte := buf.bytes[buf.Len()-1]
				lastChar := rune(lastCharByte)
				if IsIdentifierPart(lastCharByte) || lastChar == ')' || lastChar == ']' || lastChar == '"' || lastChar == '}' {
					dotPos := sc.Pos()
					sc.Next()
					isNamespace := false
					markerPos := -1
					if sc.Peek() == '@' {
						markerPos = sc.Pos()
						sc.Next()
					} else if sc.Peek() == '#' {
						isNamespace = true
						markerPos = sc.Pos()
						sc.Next()
					}
					fieldName := "#"
					fieldStart := sc.Pos()
					if !isNamespace {
						for sc.Pos() < len(s) && IsIdentifierPart(s[sc.Pos()]) {
							sc.Advance(1)
						}
						fieldName = s[fieldStart:sc.Pos()]
					}
					fieldEnd := sc.Pos()
					isOptional := false
					isAssert := false
					suffixPos := -1
					if !isNamespace && sc.Pos() < len(s) && s[sc.Pos()] == '?' {
						isOptional = true
						suffixPos = sc.Pos()
						sc.Advance(1)
					} else if !isNamespace && sc.Pos() < len(s) && s[sc.Pos()] == '!' {
						isAssert = true
						suffixPos = sc.Pos()
						sc.Advance(1)
					}
					selectorSuffix := rune(0)
					if isOptional {
						selectorSuffix = '?'
					} else if isAssert {
						selectorSuffix = '!'
					}
					appendDotNotationSelectorWithMapping(buf, s, dotPos, markerPos, fieldStart, fieldEnd, fieldName, selectorSuffix, suffixPos, isNamespace)
					continue
				}
			}
		}
		buf.AppendOriginal(s, sc.Pos(), sc.Pos()+1)
		sc.Advance(1)
	}
	return buf.String(), buf.mapping
}

func appendDotNotationSelectorWithMapping(buf *mappedBuffer, src string, dotPos, markerPos, fieldStart, fieldEnd int, fieldName string, suffix rune, suffixPos int, isNamespace bool) {
	buf.AppendLiteral("[", dotPos)
	quotePos := fieldStart
	if markerPos >= 0 {
		quotePos = markerPos
	}
	buf.AppendLiteral("\"", quotePos)
	if isNamespace {
		buf.AppendOriginal(src, markerPos, markerPos+1)
	} else {
		if markerPos >= 0 {
			buf.AppendOriginal(src, markerPos, markerPos+1)
		}
		buf.AppendOriginal(src, fieldStart, fieldEnd)
	}
	if suffix != 0 && suffixPos >= 0 {
		buf.AppendOriginal(src, suffixPos, suffixPos+1)
	}
	closePos := fieldEnd - 1
	if suffixPos >= 0 {
		closePos = suffixPos
	}
	if closePos < 0 {
		closePos = dotPos
	}
	buf.AppendLiteral("\"", closePos)
	buf.AppendLiteral("]", closePos)
}

// replaceCaseStatements converts case statements.
func replaceCaseStatements(s string) string {
	var result []rune
	sc := stringutils.NewExpressionScanner(s)

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
func processCaseStatement(s string, sc *stringutils.ExpressionScanner, result []rune, casePos int, kwLen int) ([]rune, bool) {
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
	sc := stringutils.NewExpressionScanner(s)

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
	sc := stringutils.NewExpressionScanner(s)

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
	sc := stringutils.NewExpressionScanner(s)

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
