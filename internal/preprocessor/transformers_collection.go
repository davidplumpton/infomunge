package preprocessor

import (
	"strings"

	"infomunge/internal/stringutils"
)

// replaceFilterOperator converts "array filter __lambda(...)" to "__filter(array, __lambda(...))"
func replaceFilterOperator(s string) string {
	return replaceCollectionOperator(s, " filter", "__filter")
}

// replaceMapOperator converts "array map __lambda(...)" to "__map(array, __lambda(...))"
func replaceMapOperator(s string) string {
	return replaceCollectionOperator(s, " map", "__map")
}

// replaceReduceOperator converts "array reduce __lambda(...)" to "__reduce(array, __lambda(...))"
func replaceReduceOperator(s string) string {
	return replaceCollectionOperator(s, " reduce", "__reduce")
}

// replaceGroupByOperator converts "array groupBy __lambda(...)" to "__groupBy(array, __lambda(...))"
func replaceGroupByOperator(s string) string {
	return replaceCollectionOperator(s, " groupBy", "__groupBy")
}

// replaceFlatMapOperator converts "array flatMap __lambda(...)" to "__flatMap(array, __lambda(...))"
func replaceFlatMapOperator(s string) string {
	return replaceCollectionOperator(s, " flatMap", "__flatMap")
}

// replaceMaxByOperator converts "array maxBy __lambda(...)" to "maxBy(array, __lambda(...))"
func replaceMaxByOperator(s string) string {
	return replaceCollectionOperator(s, " maxBy", "maxBy")
}

// replaceMinByOperator converts "array minBy __lambda(...)" to "minBy(array, __lambda(...))"
func replaceMinByOperator(s string) string {
	return replaceCollectionOperator(s, " minBy", "minBy")
}

// replaceOrderByOperator converts "array orderBy __lambda(...)" to "orderBy(array, __lambda(...))"
func replaceOrderByOperator(s string) string {
	return replaceCollectionOperator(s, " orderBy", "orderBy")
}

// replaceSortOperator converts "array sort __lambda(...)" to "orderBy(array, __lambda(...))"
func replaceSortOperator(s string) string {
	return replaceCollectionOperator(s, " sort", "orderBy")
}

// replaceDistinctByOperator converts "array distinctBy __lambda(...)" to "distinctBy(array, __lambda(...))"
func replaceDistinctByOperator(s string) string {
	return replaceCollectionOperator(s, " distinctBy", "distinctBy")
}

// replaceFilterObjectOperator converts "object filterObject __lambda(...)" to "filterObject(object, __lambda(...))"
func replaceFilterObjectOperator(s string) string {
	return replaceCollectionOperator(s, " filterObject", "filterObject")
}

// replaceMapObjectOperator converts "object mapObject __lambda(...)" to "mapObject(object, __lambda(...))"
func replaceMapObjectOperator(s string) string {
	return replaceCollectionOperator(s, " mapObject", "mapObject")
}

// replacePluckOperator converts "array pluck lambda" to "__pluck(array, lambda)"
func replacePluckOperator(s string) string {
	return replaceCollectionOperator(s, " pluck", "__pluck")
}

// replaceCollectionOperator handles transformation of "array op lambda" into "func(array, lambda)".
func replaceCollectionOperator(s string, opKey string, funcName string) string {
	var result []rune
	sc := stringutils.NewExpressionScanner(s)
	opLen := len(opKey)

	for sc.Pos() < len(s) {
		if !sc.IsInString() && sc.Pos()+opLen <= len(s) && s[sc.Pos():sc.Pos()+opLen] == opKey {
			afterOp := sc.Pos() + opLen
			// Match operator followed by space, newline, or parenthesis
			if afterOp < len(s) && (s[afterOp] == ' ' || s[afterOp] == '\n' || s[afterOp] == '\r' || s[afterOp] == '\t' || s[afterOp] == '(') {
				arrayStart := stringutils.FindLeftOperandStart(result, nil)
				if arrayStart >= len(result) {
					result = append(result, []rune(opKey)...)
					sc.Advance(opLen)
					continue
				}
				arrayOp := strings.TrimSpace(string(result[arrayStart:]))
				result = result[:arrayStart]
				result = append(result, []rune(funcName+"(")...)
				result = append(result, []rune(arrayOp)...)
				result = append(result, []rune(", ")...)
				pos := sc.Pos() + opLen
				// Skip all whitespace including newlines
				for pos < len(s) && (s[pos] == ' ' || s[pos] == '\t' || s[pos] == '\n' || s[pos] == '\r') {
					pos++
				}
				lambdaStart := pos
				// Track depth fresh from the lambda start position
				var state ScanState
				wasInGroup := false
				for pos < len(s) {
					ch := s[pos]
					if !state.InString() {
						if ch == '(' || ch == '[' || ch == '{' {
							state.Advance(ch)
							wasInGroup = true
						} else if ch == ')' || ch == ']' || ch == '}' {
							if state.Depth() == 0 {
								break
							}
							state.Advance(ch)
							// If we just exited a group back to depth 0, check for collection operators
							if state.Depth() == 0 && wasInGroup {
								// Look ahead for collection operators that would start a new expression
								j := pos + 1
								for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
									j++
								}
								// Check for collection operators
								shouldBreak := false
								for _, op := range CollectionOperators {
									opWithSpace := op + " "
									if j+len(opWithSpace) <= len(s) && s[j:j+len(opWithSpace)] == opWithSpace {
										shouldBreak = true
										break
									}
								}
								if shouldBreak {
									pos++ // include the closing bracket
									break
								}
							}
						} else if state.Depth() == 0 && (ch == ',' || (ch == ' ' && pos+opLen <= len(s) && s[pos:pos+opLen] == opKey)) {
							break
						}
					} else {
						state.Advance(ch)
					}
					pos++
				}
				lambdaStr := strings.TrimSpace(s[lambdaStart:pos])
				result = append(result, []rune(lambdaStr)...)
				result = append(result, ')')
				sc.SetPos(pos)
				continue
			}
		}
		result = append(result, sc.NextRune())
	}
	return string(result)
}
