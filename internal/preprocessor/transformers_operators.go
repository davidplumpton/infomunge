package preprocessor

import (
	unifiederrors "infomunge/internal/errors"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"infomunge/internal/stringutils"
)

const (
	binaryOpDefault     = "default"
	binaryOpOnNull      = "onNull"
	binaryOpThen        = "then"
	binaryOpUpdate      = "update"
	binaryOpFind        = "find"
	binaryOpConcatenate = "concatenate"
	binaryOpRemove      = "remove"
	binaryOpSplitBy     = "splitBy"
	binaryOpJoinBy      = "joinBy"
	binaryOpTo          = "to"
	binaryOpMatch       = "match"
	binaryOpContains    = "contains"
	binaryOpMatches     = "matches"
	binaryOpRepeat      = "repeat"
	binaryOpMod         = "mod"
)

var binaryOperatorConfigs = map[string]stringutils.BinaryOperatorConfig{
	binaryOpDefault: {
		Operator:     " default ",
		FuncName:     "__default",
		RightStopOps: []string{" and "},
	},
	binaryOpOnNull: {
		Operator: " onNull ",
		FuncName: "onNull",
	},
	binaryOpThen: {
		Operator: " then ",
		FuncName: "then",
	},
	binaryOpUpdate: {
		Operator:     " ~ ",
		FuncName:     "__update",
		ExtraStops:   []rune{'~'},
		RightStopOps: []string{" ~ "},
	},
	binaryOpFind: {
		Operator: " find ",
		FuncName: "find",
	},
	binaryOpConcatenate: {
		Operator:     " ++ ",
		FuncName:     "__concat",
		ExtraStops:   []rune{'+'},
		RightStopOps: []string{" ++ "},
	},
	binaryOpRemove: {
		Operator:     " -- ",
		FuncName:     "__remove",
		ExtraStops:   []rune{'-'},
		RightStopOps: []string{" -- "},
	},
	binaryOpSplitBy: {
		Operator:     " splitBy ",
		FuncName:     "splitBy",
		RightStopOps: []string{" splitBy "},
	},
	binaryOpJoinBy: {
		Operator:     " joinBy ",
		FuncName:     "joinBy",
		RightStopOps: []string{" joinBy "},
	},
	binaryOpTo: {
		Operator:        " to ",
		FuncName:        "to",
		RightStopOps:    []string{" to "},
		UseMinimalStops: true, // Allow negative numbers like -2 to 5
	},
	binaryOpMatch: {
		Operator:     " match ",
		FuncName:     "match",
		RightStopOps: []string{" match "},
	},
	binaryOpContains: {
		Operator:     " contains ",
		FuncName:     "contains",
		RightStopOps: []string{" && ", " || ", " and ", " or ", " contains "},
	},
	binaryOpMatches: {
		Operator:     " matches ",
		FuncName:     "matches",
		RightStopOps: []string{" matches "},
	},
	binaryOpRepeat: {
		Operator:     " repeat ",
		FuncName:     "repeat",
		RightStopOps: []string{" repeat "},
	},
	binaryOpMod: {
		Operator:     " mod ",
		FuncName:     "mod",
		RightStopOps: []string{"==", "!=", "<", ">", "<=", ">=", " matches "},
	},
}

func replaceConfiguredBinaryOperator(s string, key string) (string, error) {
	config, ok := binaryOperatorConfigs[key]
	if !ok {
		return s, unifiederrors.ParseErrorf("missing binary operator config: %s", key)
	}
	return stringutils.ReplaceBinaryOperator(s, config), nil
}

// replaceDefaultOperator converts "expr default value" to "__default(expr, value)"
func replaceDefaultOperator(s string) string {
	result, _ := replaceDefaultOperatorErr(s)
	return result
}

func replaceDefaultOperatorErr(s string) (string, error) {
	return replaceConfiguredBinaryOperator(s, binaryOpDefault)
}

// replaceOnNullOperator converts "expr onNull value" to "onNull(expr, value)"
func replaceOnNullOperator(s string) string {
	result, _ := replaceOnNullOperatorErr(s)
	return result
}

func replaceOnNullOperatorErr(s string) (string, error) {
	return replaceConfiguredBinaryOperator(s, binaryOpOnNull)
}

// replaceThenOperator converts "expr then value" to "then(expr, value)"
func replaceThenOperator(s string) string {
	result, _ := replaceThenOperatorErr(s)
	return result
}

func replaceThenOperatorErr(s string) (string, error) {
	return replaceConfiguredBinaryOperator(s, binaryOpThen)
}

// replaceUpdateOperator converts "obj ~ {field: value}" to "__update(obj, {field: value})"
func replaceUpdateOperator(s string) string {
	result, _ := replaceUpdateOperatorErr(s)
	return result
}

func replaceUpdateOperatorErr(s string) (string, error) {
	return replaceConfiguredBinaryOperator(s, binaryOpUpdate)
}

// replaceFindOperator converts "source find value" to "find(source, value)"
func replaceFindOperator(s string) string {
	result, _ := replaceFindOperatorErr(s)
	return result
}

func replaceFindOperatorErr(s string) (string, error) {
	return replaceConfiguredBinaryOperator(s, binaryOpFind)
}

// replaceAsOperator converts "value as TypeExpr" to "__coerce(value, \"TypeExpr\")"
func replaceAsOperator(s string) string {
	var result []rune
	inString := false
	i := 0

	for i < len(s) {
		if s[i] == '"' && (i == 0 || s[i-1] != '\\') {
			inString = !inString
			result = append(result, '"')
			i++
			continue
		}

		if !inString && i+4 <= len(s) && s[i:i+4] == " as " {
			leftStart := findLeftOperandForAs(result)

			if leftStart >= len(result) {
				result = append(result, []rune(" as ")...)
				i += 4
				continue
			}

			leftOp := strings.TrimSpace(string(result[leftStart:]))
			result = result[:leftStart]

			i += 4 // skip " as "

			typeStart := i
			for i < len(s) && IsTypeExprChar(s[i]) {
				if i+4 <= len(s) && s[i:i+4] == "map[" {
					break
				}
				i++
			}

			if i == typeStart {
				result = append(result, []rune(leftOp)...)
				result = append(result, []rune(" as ")...)
				continue
			}

			typeExpr := strings.TrimSpace(s[typeStart:i])
			if typeExpr == "" {
				result = append(result, []rune(leftOp)...)
				result = append(result, []rune(" as ")...)
				continue
			}

			configStart := i
			for configStart < len(s) && unicode.IsSpace(rune(s[configStart])) {
				configStart++
			}

			var configArg string
			prefixes := []string{GoObjectPrefix, GoObjectPrefixSpace}
			for _, prefix := range prefixes {
				if configStart+len(prefix) <= len(s) && s[configStart:configStart+len(prefix)] == prefix {
					bracePos := configStart + len(prefix) - 1
					sc := stringutils.NewExpressionScanner(s)
					closePos := sc.FindMatchingCloseBracket(bracePos)

					if closePos != -1 {
						configArg = s[configStart : closePos+1]
						i = closePos + 1
						break
					}
				}
			}

			quotedType := strconv.Quote(typeExpr)
			result = append(result, []rune("__coerce(")...)
			result = append(result, []rune(leftOp)...)
			result = append(result, []rune(", ")...)
			result = append(result, []rune(quotedType)...)

			if configArg != "" {
				result = append(result, []rune(", ")...)
				result = append(result, []rune(configArg)...)
			}

			result = append(result, ')')
			continue
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		result = append(result, r)
		i += size
	}

	return string(result)
}

// replaceIsOperator converts "value is Type" to "__isType(value, \"Type\")"
func replaceIsOperator(s string) string {
	var result []rune
	inString := false
	i := 0

	for i < len(s) {
		if s[i] == '"' && (i == 0 || s[i-1] != '\\') {
			inString = !inString
			result = append(result, '"')
			i++
			continue
		}

		if !inString && i+4 <= len(s) && s[i:i+4] == " is " {
			leftStart := findLeftOperandForAs(result)

			if leftStart >= len(result) {
				result = append(result, []rune(" is ")...)
				i += 4
				continue
			}

			leftOp := strings.TrimSpace(string(result[leftStart:]))
			result = result[:leftStart]

			i += 4 // skip " is "

			typeStart := i
			for i < len(s) && IsTypeExprChar(s[i]) {
				i++
			}

			if i == typeStart {
				result = append(result, []rune(leftOp)...)
				result = append(result, []rune(" is ")...)
				continue
			}

			typeExpr := strings.TrimSpace(s[typeStart:i])
			if typeExpr == "" {
				result = append(result, []rune(leftOp)...)
				result = append(result, []rune(" is ")...)
				continue
			}

			quotedType := strconv.Quote(typeExpr)
			result = append(result, []rune("__isType(")...)
			result = append(result, []rune(leftOp)...)
			result = append(result, []rune(", ")...)
			result = append(result, []rune(quotedType)...)
			result = append(result, ')')
			continue
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		result = append(result, r)
		i += size
	}

	return string(result)
}

// replaceConcatenateOperator converts "++" to "__concat".
func replaceConcatenateOperator(s string) string {
	result, _ := replaceConcatenateOperatorErr(s)
	return result
}

func replaceConcatenateOperatorErr(s string) (string, error) {
	return replaceConfiguredBinaryOperator(s, binaryOpConcatenate)
}

// replaceRemoveOperator converts "--" to "__remove".
func replaceRemoveOperator(s string) string {
	result, _ := replaceRemoveOperatorErr(s)
	return result
}

func replaceRemoveOperatorErr(s string) (string, error) {
	return replaceConfiguredBinaryOperator(s, binaryOpRemove)
}

// replaceExponentOperator converts "**" to "pow".
func replaceExponentOperator(s string) string {
	var result []rune
	var topState ScanState
	i := 0

	for i < len(s) {
		if !topState.InString() && i+2 <= len(s) && s[i:i+2] == "**" {
			leftStart := stringutils.FindLeftOperandStart(result, nil)
			if leftStart >= len(result) {
				result = append(result, []rune("**")...)
				i += 2
				continue
			}

			leftOp := strings.TrimSpace(string(result[leftStart:]))
			if leftOp == "" {
				result = append(result, []rune("**")...)
				i += 2
				continue
			}

			rightStart := i + 2
			for rightStart < len(s) && unicode.IsSpace(rune(s[rightStart])) {
				rightStart++
			}

			rightEnd := rightStart
			var rightState ScanState

			for rightEnd < len(s) {
				ch := s[rightEnd]
				if !rightState.InString() {
					if ch == '(' || ch == '[' || ch == '{' {
						rightState.Advance(ch)
						rightEnd++
						continue
					}
					if (ch == ')' || ch == ']' || ch == '}') && rightState.Depth() == 0 {
						break
					}
					if rightState.Depth() == 0 && (ch == ',' || shouldStopExponentAtOperator(s, rightEnd, rightStart)) {
						break
					}
				}
				rightState.Advance(ch)
				rightEnd++
			}

			rightOp := strings.TrimSpace(s[rightStart:rightEnd])
			if rightOp == "" {
				result = append(result, []rune("**")...)
				i += 2
				continue
			}

			result = result[:leftStart]
			result = append(result, []rune("pow(")...)
			result = append(result, []rune(leftOp)...)
			result = append(result, []rune(", ")...)
			result = append(result, []rune(rightOp)...)
			result = append(result, ')')
			i = rightEnd
			continue
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		result = append(result, r)
		topState.Advance(s[i])
		i += size
	}

	return string(result)
}

func shouldStopExponentAtOperator(s string, pos int, start int) bool {
	ch := s[pos]
	if (ch == '+' || ch == '-') && pos == start {
		return false
	}

	if ch == '*' {
		if (pos+1 < len(s) && s[pos+1] == '*') || (pos > 0 && s[pos-1] == '*') {
			return false
		}
	}

	for _, stop := range stringutils.DefaultOperatorStops {
		if rune(ch) == stop {
			return true
		}
	}

	return false
}

// replaceSplitByOperator converts "splitBy".
func replaceSplitByOperator(s string) string {
	result, _ := replaceSplitByOperatorErr(s)
	return result
}

func replaceSplitByOperatorErr(s string) (string, error) {
	return replaceConfiguredBinaryOperator(s, binaryOpSplitBy)
}

// replaceJoinByOperator converts "joinBy".
func replaceJoinByOperator(s string) string {
	result, _ := replaceJoinByOperatorErr(s)
	return result
}

func replaceJoinByOperatorErr(s string) (string, error) {
	return replaceConfiguredBinaryOperator(s, binaryOpJoinBy)
}

// replaceToOperator converts "start to end" to "to(start, end)" for range expressions.
func replaceToOperator(s string) string {
	result, _ := replaceToOperatorErr(s)
	return result
}

func replaceToOperatorErr(s string) (string, error) {
	return replaceConfiguredBinaryOperator(s, binaryOpTo)
}

// replacePipeToFunctionOperator converts "expr | funcName" to "funcName(expr)"
// This handles the DataWeave-style pipe-to-function syntax where a single identifier
// follows the pipe operator.
func replacePipeToFunctionOperator(s string) string {
	var result []rune
	inString := false
	i := 0

	for i < len(s) {
		if s[i] == '"' && (i == 0 || s[i-1] != '\\') {
			inString = !inString
			result = append(result, '"')
			i++
			continue
		}

		// Look for " | " followed by an identifier (not another operator or value)
		if !inString && i+3 <= len(s) && s[i:i+3] == " | " {
			// Skip " | "
			i += 3

			// Skip whitespace after pipe
			for i < len(s) && unicode.IsSpace(rune(s[i])) {
				i++
			}

			// Check if what follows is a valid identifier (not a number, string, or bracket)
			if i >= len(s) {
				result = append(result, []rune(" | ")...)
				continue
			}

			ch := s[i]
			// Check if this is the start of an identifier
			if !isIdentStart(ch) {
				// Not an identifier, restore the pipe and continue
				result = append(result, []rune(" | ")...)
				continue
			}

			// Scan the identifier
			identStart := i
			for i < len(s) && isIdentChar(s[i]) {
				i++
			}
			funcName := s[identStart:i]

			// Skip whitespace after identifier
			afterIdent := i
			for afterIdent < len(s) && unicode.IsSpace(rune(s[afterIdent])) {
				afterIdent++
			}

			// Check if the next non-whitespace character is '(' - if so, this is a function call
			// not a pipe-to-function, so don't transform it
			if afterIdent < len(s) && s[afterIdent] == '(' {
				// This is already a function call like "x | func(args)", don't transform
				result = append(result, []rune(" | ")...)
				result = append(result, []rune(funcName)...)
				continue
			}

			// This is a pipe-to-function: transform "expr | funcName" to "funcName(expr)"
			// Find the left operand
			leftStart := stringutils.FindLeftOperandStart(result, nil)

			if leftStart >= len(result) {
				// No left operand found, restore original
				result = append(result, []rune(" | ")...)
				result = append(result, []rune(funcName)...)
				continue
			}

			leftOp := strings.TrimSpace(string(result[leftStart:]))
			result = result[:leftStart]

			result = append(result, []rune(funcName+"(")...)
			result = append(result, []rune(leftOp)...)
			result = append(result, ')')
			continue
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		result = append(result, r)
		i += size
	}

	return string(result)
}

// isIdentStart returns true if ch can start an identifier (letter or underscore)
func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// findLeftOperandForAs finds the left operand for the 'as' operator.
// It stops at word operators like 'match', 'default', etc. to ensure proper precedence.
func findLeftOperandForAs(result []rune) int {
	// First, find the basic left operand using default rules
	leftStart := stringutils.FindLeftOperandStart(result, []rune{':', '&', '|'})

	if leftStart >= len(result) {
		return leftStart
	}

	// Now check if the operand contains word operators that 'as' should not cross
	leftOp := string(result[leftStart:])

	// List of word operators that 'as' should not cross (as has higher precedence)
	wordOps := []string{
		" match ", " matches ", " scan ",
		" default ", " onNull ", " then ",
		" find ", " contains ",
		" splitBy ", " joinBy ", " replace ",
		" and ", " or ", " && ", " || ",
	}

	// Find the last occurrence of any word operator
	lastWordOpPos := -1
	for _, op := range wordOps {
		pos := strings.LastIndex(leftOp, op)
		if pos > lastWordOpPos {
			lastWordOpPos = pos + len(op) // Position after the operator
		}
	}

	if lastWordOpPos > 0 {
		// Adjust leftStart to after the word operator
		return leftStart + lastWordOpPos
	}

	return leftStart
}

// replaceMatchOperator converts "str match pattern" to "match(str, pattern)"
// This transforms the infix match operator into a function call.
// Note: In DataWeave, `match` returns capture groups, while `matches` returns a boolean.
func replaceMatchOperator(s string) string {
	result, _ := replaceMatchOperatorErr(s)
	return result
}

func replaceMatchOperatorErr(s string) (string, error) {
	return replaceConfiguredBinaryOperator(s, binaryOpMatch)
}

// replaceContainsOperator converts "arr contains value" to "contains(arr, value)"
func replaceContainsOperator(s string) string {
	result, _ := replaceContainsOperatorErr(s)
	return result
}

func replaceContainsOperatorErr(s string) (string, error) {
	return replaceConfiguredBinaryOperator(s, binaryOpContains)
}

// replaceMatchesOperator converts "str matches pattern" to "matches(str, pattern)"
// This transforms the infix matches operator into a function call.
func replaceMatchesOperator(s string) string {
	result, _ := replaceMatchesOperatorErr(s)
	return result
}

func replaceMatchesOperatorErr(s string) (string, error) {
	return replaceConfiguredBinaryOperator(s, binaryOpMatches)
}

// replaceRepeatOperator converts "str repeat n" to "repeat(str, n)"
func replaceRepeatOperator(s string) string {
	result, _ := replaceRepeatOperatorErr(s)
	return result
}

func replaceRepeatOperatorErr(s string) (string, error) {
	return replaceConfiguredBinaryOperator(s, binaryOpRepeat)
}

// replaceModOperator converts "a mod b" to "mod(a, b)"
func replaceModOperator(s string) string {
	result, _ := replaceModOperatorErr(s)
	return result
}

func replaceModOperatorErr(s string) (string, error) {
	return replaceConfiguredBinaryOperator(s, binaryOpMod)
}

// replaceMethodCallOperator is a generic transformer for "left operator(args)" → "operator(left, args)".
// It handles operators that can be called in method-like syntax with parentheses.
func replaceMethodCallOperator(s string, operatorName string) string {
	token := " " + operatorName + "("
	tokenLen := len(token)

	var result []rune
	var topState ScanState
	i := 0

	for i < len(s) {
		// Look for " operator(" pattern
		if !topState.InString() && i+tokenLen <= len(s) && s[i:i+tokenLen] == token {
			// Find the left operand
			leftStart := stringutils.FindLeftOperandStart(result, nil)

			if leftStart >= len(result) {
				// No left operand found, continue
				result = append(result, []rune(token)...)
				i += tokenLen
				continue
			}

			leftOp := strings.TrimSpace(string(result[leftStart:]))
			result = result[:leftStart]

			// Build the function call
			result = append(result, []rune(operatorName+"(")...)
			result = append(result, []rune(leftOp)...)
			result = append(result, []rune(", ")...)
			i += tokenLen // skip " operator("

			// Copy the rest of the arguments until we find the closing paren
			argState := ScanState{DepthParen: 1}
			for i < len(s) && argState.DepthParen > 0 {
				ch := s[i]
				argState.Advance(ch)
				result = append(result, rune(ch))
				i++
			}
			continue
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		result = append(result, r)
		topState.Advance(s[i])
		i += size
	}

	return string(result)
}

// replaceSubstringOperator converts "str substring(start, end)" to "substring(str, start, end)"
func replaceSubstringOperator(s string) string {
	return replaceMethodCallOperator(s, "substring")
}

// replaceContainsMethodCall converts "str contains(pattern)" to "contains(str, pattern)"
func replaceContainsMethodCall(s string) string {
	return replaceMethodCallOperator(s, "contains")
}

// replaceFindMethodCall converts "str find(pattern)" to "find(str, pattern)"
func replaceFindMethodCall(s string) string {
	return replaceMethodCallOperator(s, "find")
}

// replaceMatchMethodCall converts "str match(pattern)" to "match(str, pattern)"
func replaceMatchMethodCall(s string) string {
	return replaceMethodCallOperator(s, "match")
}

// replaceMatchesMethodCall converts "str matches(pattern)" to "matches(str, pattern)"
func replaceMatchesMethodCall(s string) string {
	return replaceMethodCallOperator(s, "matches")
}

// replaceScanMethodCall converts "str scan(pattern)" to "scan(str, pattern)"
func replaceScanMethodCall(s string) string {
	return replaceMethodCallOperator(s, "scan")
}

// replaceSplitByMethodCall converts "str splitBy(pattern)" to "splitBy(str, pattern)"
func replaceSplitByMethodCall(s string) string {
	return replaceMethodCallOperator(s, "splitBy")
}
