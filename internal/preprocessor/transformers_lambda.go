package preprocessor

import (
	"infomunge/internal/stringutils"
	"strings"
	"unicode"
)

var implicitLambdaOperators = []string{
	" map ",
	" filter ",
	" reduce ",
	" groupBy ",
	" sort ",
	" maxBy ",
	" minBy ",
	" orderBy ",
	" distinctBy ",
	" filterObject ",
	" flatMap ",
	" pluck ",
}

// replaceImplicitLambdas converts implicit lambda syntax using $ and $$ to explicit arrow functions.
func replaceImplicitLambdas(s string) string {
	for _, op := range implicitLambdaOperators {
		s = replaceImplicitLambdaForOp(s, op)
	}
	return s
}

// wrapImplicitObjectLiteralBodies adds braces around map bodies that start with a key/value pair.
func wrapImplicitObjectLiteralBodies(s string) string {
	var result []rune
	runes := []rune(s)
	op := " map "

	for i := 0; i < len(runes); i++ {
		if runes[i] == '"' {
			i, result = copyStringLiteral(runes, i, result)
			continue
		}

		if !matchesAt(runes, i, op) {
			result = append(result, runes[i])
			continue
		}

		result = append(result, []rune(op)...)
		i += len(op)

		for i < len(runes) && unicode.IsSpace(runes[i]) {
			result = append(result, runes[i])
			i++
		}

		bodyStart := i
		bodyEnd, hasArrow := scanLambdaBody(runes, bodyStart)
		if hasArrow {
			i--
			continue
		}

		bodyStr := string(runes[bodyStart:bodyEnd])
		trimmed := strings.TrimSpace(bodyStr)
		if trimmed != "" && !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") && shouldWrapTopLevelObjectLiteral(trimmed) {
			result = append(result, '{')
			result = append(result, []rune(bodyStr)...)
			result = append(result, '}')
			i = bodyEnd - 1
			continue
		}

		result = append(result, []rune(bodyStr)...)
		i = bodyEnd - 1
	}

	return string(result)
}

// replaceImplicitLambdaForOp handles implicit lambda replacement for a specific operator.
// Operator op should include trailing space (e.g., " reduce ") for matching with space,
// but we also handle the case where the operator is followed directly by "(" (no space).
func replaceImplicitLambdaForOp(s string, op string) string {
	var result []rune
	runes := []rune(s)
	// opNoTrailingSpace is the operator without the trailing space (e.g., " reduce")
	opNoTrailingSpace := strings.TrimSuffix(op, " ")

	for i := 0; i < len(runes); i++ {
		if runes[i] == '"' {
			i, result = copyStringLiteral(runes, i, result)
			continue
		}

		// Try to match the operator - first try with trailing space, then without
		matched := false
		matchLen := 0
		needSpace := false
		if matchesAt(runes, i, op) {
			// Matched with trailing space (e.g., " reduce ")
			matched = true
			matchLen = len(op)
		} else if matchesAt(runes, i, opNoTrailingSpace) {
			// Check if operator is followed by "(" or whitespace (e.g., " reduce(" or " map\n")
			afterOp := i + len(opNoTrailingSpace)
			if afterOp < len(runes) && (runes[afterOp] == '(' || unicode.IsSpace(runes[afterOp])) {
				matched = true
				matchLen = len(opNoTrailingSpace)
				needSpace = true // Need to add space before lambda
			}
		}

		if !matched {
			result = append(result, runes[i])
			continue
		}

		result = append(result, runes[i:i+matchLen]...)
		i += matchLen

		hasWhitespace := false
		for i < len(runes) && unicode.IsSpace(runes[i]) {
			result = append(result, runes[i])
			hasWhitespace = true
			i++
		}
		// Add space if we matched without trailing space and there was no whitespace
		if needSpace && !hasWhitespace {
			result = append(result, ' ')
		}

		bodyEnd, hasArrow := scanLambdaBody(runes, i)
		if hasArrow {
			i--
			continue
		}

		bodyStr := string(runes[i:bodyEnd])
		newBody, hasDollar, hasDoubleDollar := rewriteImplicitParams(bodyStr)

		if !hasDollar && !hasDoubleDollar {
			i--
			continue
		}

		params := "__arg"
		if hasDoubleDollar {
			params = "__arg, __idx"
		}

		result = append(result, []rune("("+params+") -> "+newBody)...)
		i = bodyEnd - 1
	}

	return string(result)
}

func copyStringLiteral(runes []rune, i int, result []rune) (int, []rune) {
	result = append(result, runes[i])
	i++
	for i < len(runes) {
		result = append(result, runes[i])
		if runes[i] == '"' && !stringutils.IsEscapedRuneAt(runes, i) {
			break
		}
		i++
	}
	return i, result
}

func matchesAt(runes []rune, i int, pattern string) bool {
	return i+len(pattern) <= len(runes) && string(runes[i:i+len(pattern)]) == pattern
}

// rewriteImplicitParams replaces executable $ and $$ references while respecting
// quoted payloads. Dollars in ordinary strings are implicit interpolation and
// therefore become concatenation expressions rooted in an empty string so the
// evaluator applies language string coercion. Strings passed to regex remain
// literal pattern/flag payloads.
func rewriteImplicitParams(s string) (rewritten string, hasDollar, hasDoubleDollar bool) {
	var result strings.Builder
	runes := []rune(s)
	var regexCallStack []bool
	i := 0

	for i < len(runes) {
		switch runes[i] {
		case '"':
			end := quotedLiteralEnd(runes, i, '"')
			if insideRegexCall(regexCallStack) {
				result.WriteString(string(runes[i:end]))
			} else {
				interpolated, stringHasDollar, stringHasDoubleDollar := rewriteImplicitString(runes[i:end])
				result.WriteString(interpolated)
				hasDollar = hasDollar || stringHasDollar
				hasDoubleDollar = hasDoubleDollar || stringHasDoubleDollar
			}
			i = end
			continue
		case '\'':
			end := quotedLiteralEnd(runes, i, '\'')
			result.WriteString(string(runes[i:end]))
			i = end
			continue
		case '(':
			parentRegexCall := insideRegexCall(regexCallStack)
			regexCallStack = append(regexCallStack, parentRegexCall || precedingIdentifierIs(runes, i, "regex"))
		case ')':
			if len(regexCallStack) > 0 {
				regexCallStack = regexCallStack[:len(regexCallStack)-1]
			}
		case '$':
			paramLen, replacement := implicitParamAt(runes, i)
			if paramLen == 0 {
				break
			}
			result.WriteString(replacement)
			if paramLen == 2 {
				hasDoubleDollar = true
			} else {
				hasDollar = true
			}
			i += paramLen
			continue
		}
		result.WriteRune(runes[i])
		i++
	}

	return result.String(), hasDollar, hasDoubleDollar
}

func rewriteImplicitString(literal []rune) (string, bool, bool) {
	if len(literal) < 2 || literal[0] != '"' || literal[len(literal)-1] != '"' {
		return string(literal), false, false
	}

	content := literal[1 : len(literal)-1]
	parts := []string{`""`}
	literalStart := 0
	hasDollar := false
	hasDoubleDollar := false

	for i := 0; i < len(content); {
		if content[i] != '$' || stringutils.IsEscapedRuneAt(content, i) {
			i++
			continue
		}
		paramLen, replacement := implicitParamAt(content, i)
		if paramLen == 0 {
			i++
			continue
		}
		if i > literalStart {
			parts = append(parts, `"`+string(content[literalStart:i])+`"`)
		}
		parts = append(parts, "("+replacement+")")
		if paramLen == 2 {
			hasDoubleDollar = true
		} else {
			hasDollar = true
		}
		i += paramLen
		literalStart = i
	}

	if !hasDollar && !hasDoubleDollar {
		return string(literal), false, false
	}
	if literalStart < len(content) {
		parts = append(parts, `"`+string(content[literalStart:])+`"`)
	}
	return "(" + strings.Join(parts, " + ") + ")", hasDollar, hasDoubleDollar
}

func implicitParamAt(runes []rune, i int) (int, string) {
	if i >= len(runes) || runes[i] != '$' {
		return 0, ""
	}

	paramLen := 1
	replacement := "__arg"
	if i+1 < len(runes) && runes[i+1] == '$' {
		paramLen = 2
		replacement = "__idx"
	}
	after := i + paramLen
	if after < len(runes) && (unicode.IsLetter(runes[after]) || unicode.IsDigit(runes[after]) || runes[after] == '_') {
		return 0, ""
	}
	return paramLen, replacement
}

func quotedLiteralEnd(runes []rune, start int, quote rune) int {
	for i := start + 1; i < len(runes); i++ {
		if runes[i] == quote && !stringutils.IsEscapedRuneAt(runes, i) {
			return i + 1
		}
	}
	return len(runes)
}

func insideRegexCall(stack []bool) bool {
	return len(stack) > 0 && stack[len(stack)-1]
}

func precedingIdentifierIs(runes []rune, pos int, want string) bool {
	end := pos
	for end > 0 && unicode.IsSpace(runes[end-1]) {
		end--
	}
	start := end
	for start > 0 && (unicode.IsLetter(runes[start-1]) || unicode.IsDigit(runes[start-1]) || runes[start-1] == '_') {
		start--
	}
	if start > 0 && (runes[start-1] == '.' || runes[start-1] == '$') {
		return false
	}
	return string(runes[start:end]) == want
}

// replaceArrowFunctions converts arrow function syntax.
func replaceArrowFunctions(s string) string {
	var result []rune
	sc := stringutils.NewExpressionScanner(s)

	for sc.Pos() < len(s) {
		if !sc.IsInString() && sc.Peek() == '(' {
			closeIdx := sc.FindMatchingCloseBracket(sc.Pos())
			if closeIdx != -1 {
				arrowIdx := closeIdx + 1
				for arrowIdx < len(s) && s[arrowIdx] == ' ' {
					arrowIdx++
				}

				if arrowIdx+1 < len(s) && s[arrowIdx:arrowIdx+2] == "->" {
					paramsStr := strings.TrimSpace(s[sc.Pos()+1 : closeIdx])
					pos := arrowIdx + 2
					for pos < len(s) && (s[pos] == ' ' || s[pos] == '\t') {
						pos++
					}
					bodyStart := pos
					// Track depth fresh from the body start position
					var bodyState ScanState
					for pos < len(s) {
						ch := s[pos]
						if !bodyState.InString() {
							if (ch == ')' || ch == ']' || ch == '}') && bodyState.Depth() == 0 {
								break
							}
							if bodyState.Depth() == 0 && ch == ',' {
								break
							}
							if bodyState.Depth() == 0 && ch == ' ' {
								// Check for operators that end the body
								if pos+4 <= len(s) && s[pos:pos+4] == " or " {
									break
								}
								if isCollectionOperatorWithSpacesAt(s, pos) {
									break
								}
							}
						}
						bodyState.Advance(ch)
						pos++
					}

					bodyStr := strings.TrimSpace(s[bodyStart:pos])
					// Escape backslashes and quotes in paramsStr since it's
					// embedded inside a Go string literal. This handles cases
					// where the rewriter has converted object literals in default
					// values (e.g., {sum: 0} -> map[string]interface{}{"sum": 0,}).
					escapedParams := strings.ReplaceAll(paramsStr, `\`, `\\`)
					escapedParams = strings.ReplaceAll(escapedParams, `"`, `\"`)
					result = append(result, []rune("__lambda(\"")...)
					result = append(result, []rune(escapedParams)...)
					result = append(result, []rune("\", ")...)
					result = append(result, []rune(bodyStr)...)
					result = append(result, ')')
					sc.SetPos(pos)
					continue
				}
			}
		}
		result = append(result, sc.NextRune())
	}
	return string(result)
}
