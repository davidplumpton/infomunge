package preprocessor

import (
	"infomunge/internal/stringutils"
	"strings"
	"unicode"
)

// CollectionOperators is the canonical list of collection operators used for body boundary detection.
// These operators signal the end of a lambda body when chaining expressions.
var CollectionOperators = []string{
	"map", "filter", "reduce", "flatMap", "groupBy", "pluck",
	"sort", "orderBy", "maxBy", "minBy", "distinctBy", "filterObject", "mapObject",
}

// replaceImplicitLambdas converts implicit lambda syntax using $ and $$ to explicit arrow functions.
func replaceImplicitLambdas(s string) string {
	for _, op := range []string{" map ", " filter ", " reduce ", " groupBy ", " sort ",
		" maxBy ", " minBy ", " orderBy ", " distinctBy ", " filterObject ", " flatMap ", " pluck "} {
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
		hasDollar, hasDoubleDollar := detectDollarParams(bodyStr)

		if !hasDollar && !hasDoubleDollar {
			i--
			continue
		}

		newBody := bodyStr
		if hasDoubleDollar {
			newBody = replaceImplicitParam(newBody, "$$", "__idx")
		}
		if hasDollar {
			newBody = replaceImplicitParam(newBody, "$", "__arg")
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
		if runes[i] == '"' && runes[i-1] != '\\' {
			break
		}
		i++
	}
	return i, result
}

func matchesAt(runes []rune, i int, pattern string) bool {
	return i+len(pattern) <= len(runes) && string(runes[i:i+len(pattern)]) == pattern
}

func scanLambdaBody(runes []rune, start int) (end int, hasArrow bool) {
	var state ScanState
	wasInGroup := false
	// This scanner decides where an implicit lambda body stops.
	// It tracks nesting and strings so commas/operators only terminate at depth 0.
	// Special case: after closing a top-level group, we look ahead for either
	// `->` (explicit lambda, so caller should not rewrite) or another collection
	// operator that starts the next pipeline segment.
	for i := start; i < len(runes); i++ {
		ch := runes[i]
		if state.InString() {
			state.AdvanceRune(ch)
			continue
		}
		switch ch {
		case '(', '[', '{':
			state.AdvanceRune(ch)
			wasInGroup = true
		case ')', ']', '}':
			if state.Depth() == 0 {
				return i, false
			}
			state.AdvanceRune(ch)
			// If we just exited a group back to depth 0, check what follows
			if state.Depth() == 0 && wasInGroup {
				// Look ahead for "->" or collection operators after optional whitespace
				j := i + 1
				for j < len(runes) && (runes[j] == ' ' || runes[j] == '\t') {
					j++
				}
				if j+1 < len(runes) && runes[j] == '-' && runes[j+1] == '>' {
					// There's an arrow, continue scanning to find it
					continue
				}
				// Check for collection operators that would end the body
				rest := string(runes[j:])
				for _, op := range CollectionOperators {
					opWithSpace := op + " "
					if len(rest) >= len(opWithSpace) && rest[:len(opWithSpace)] == opWithSpace {
						// Collection operator found, body ends here
						return i + 1, false
					}
				}
				// No arrow or collection operator, continue scanning
			}
		case ',':
			if state.Depth() == 0 {
				return i, false
			}
		case ' ':
			// Check for collection operators at depth 0
			if state.Depth() == 0 {
				rest := string(runes[i+1:])
				for _, op := range CollectionOperators {
					opWithSpace := op + " "
					if len(rest) >= len(opWithSpace) && rest[:len(opWithSpace)] == opWithSpace {
						// Collection operator found, body ends here
						return i, false
					}
				}
			}
		default:
			state.AdvanceRune(ch)
		}
		if matchesAt(runes, i, "->") {
			return i, true
		}
	}
	return len(runes), false
}

func detectDollarParams(s string) (hasDollar, hasDoubleDollar bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == '$' {
			if i+1 < len(s) && s[i+1] == '$' {
				hasDoubleDollar = true
				i++
			} else {
				hasDollar = true
			}
		}
	}
	return
}

// replaceImplicitParam replaces $ or $$ with the actual parameter name.
func replaceImplicitParam(s string, placeholder string, replacement string) string {
	var result strings.Builder
	runes := []rune(s)
	placeholderRunes := []rune(placeholder)
	i := 0

	for i < len(runes) {
		if i+len(placeholderRunes) <= len(runes) && string(runes[i:i+len(placeholderRunes)]) == placeholder {
			if placeholder == "$" && i+1 < len(runes) && runes[i+1] == '$' {
				result.WriteRune(runes[i])
				i++
				continue
			}
			afterPlaceholder := i + len(placeholderRunes)
			if afterPlaceholder < len(runes) && (unicode.IsLetter(runes[afterPlaceholder]) || unicode.IsDigit(runes[afterPlaceholder]) || runes[afterPlaceholder] == '_') {
				result.WriteRune(runes[i])
				i++
				continue
			}
			result.WriteString(replacement)
			i += len(placeholderRunes)
			continue
		}
		result.WriteRune(runes[i])
		i++
	}

	return result.String()
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
								// Check for collection operators
								shouldBreak := false
								for _, op := range CollectionOperators {
									opWithSpaces := " " + op + " "
									if pos+len(opWithSpaces) <= len(s) && s[pos:pos+len(opWithSpaces)] == opWithSpaces {
										shouldBreak = true
										break
									}
								}
								if shouldBreak {
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
