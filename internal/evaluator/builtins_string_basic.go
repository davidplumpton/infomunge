package evaluator

import (
	"fmt"
	"go/ast"
	"strings"
	"unicode/utf8"
)

// callBuiltinTrim implements the trim(string) function.
func callBuiltinTrim(args []Value, e *ast.CallExpr) (Value, error) {
	str, isNull, err := requireOneNullableScalarStringArg(args, "trim", e)
	if err != nil {
		return nil, err
	}
	if isNull {
		return nil, nil
	}

	return strings.TrimSpace(str), nil
}

// callBuiltinLength implements the length(value) function.
// For strings, it returns the number of Unicode code points (runes), not bytes.
// For arrays, it returns the number of elements.
// For objects, it returns the number of keys.
func callBuiltinLength(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "length requires exactly 1 argument", e); err != nil {
		return nil, err
	}

	switch v := args[0].(type) {
	case string:
		// Count Unicode code points (runes), not bytes
		return utf8.RuneCountInString(v), nil
	case Array:
		return len(v), nil
	case Object:
		return len(v), nil
	case nil:
		return 0, nil
	default:
		return nil, newUnsupportedTypeError("length", args[0], e.Pos())
	}
}

// callBuiltinRepeat implements the repeat(string, count) function.
// Returns the string repeated count times.
func callBuiltinRepeat(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 2, "repeat requires exactly 2 arguments: string, count", e); err != nil {
		return nil, err
	}

	str, err := assertStringArg(args[0], 1, "repeat", e)
	if err != nil {
		return nil, err
	}

	count, err := assertIntArg(args[1], 2, "repeat", e)
	if err != nil {
		return nil, err
	}

	if count < 0 {
		return nil, newPosError("repeat count must be non-negative", e.Pos())
	}

	return strings.Repeat(str, count), nil
}

// callBuiltinSplit implements the split(string, separator) function.
func callBuiltinSplit(args []Value, e *ast.CallExpr) (Value, error) {
	strs, err := assertStringArgs(args, 2, "split", e)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strs[0], strs[1])
	result := make(Array, len(parts))
	for i, part := range parts {
		result[i] = part
	}

	return result, nil
}

// callBuiltinSubstring implements the substring(string, start, end) function.
// Uses rune (Unicode code point) indices, not byte indices.
func callBuiltinSubstring(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, newPosError("substring requires 2 or 3 arguments: string, start[, end]", e.Pos())
	}

	str, err := assertStringArg(args[0], 1, "substring", e)
	if err != nil {
		return nil, err
	}

	start, err := assertIntArg(args[1], 2, "substring", e)
	if err != nil {
		return nil, err
	}

	// Convert string to runes for proper Unicode handling
	runes := []rune(str)
	runeLen := len(runes)

	if start < 0 || start > runeLen {
		return nil, newPosError(fmt.Sprintf("substring start index %d out of range (string has %d characters)", start, runeLen), e.Pos())
	}

	if len(args) == 3 {
		end, err := assertIntArg(args[2], 3, "substring", e)
		if err != nil {
			return nil, err
		}

		if end < start || end > runeLen {
			return nil, newPosError(fmt.Sprintf("substring end index %d out of range (string has %d characters)", end, runeLen), e.Pos())
		}

		return string(runes[start:end]), nil
	}

	return string(runes[start:]), nil
}

func runeIndexFromByteOffset(str string, byteOffset int) int {
	return utf8.RuneCountInString(str[:byteOffset])
}

// callBuiltinCharAt implements the charAt(string, index) function.
func callBuiltinCharAt(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 2, "charAt requires exactly 2 arguments: string, index", e); err != nil {
		return nil, err
	}

	str, err := assertStringArg(args[0], 1, "charAt", e)
	if err != nil {
		return nil, err
	}

	idx, err := assertIntArg(args[1], 2, "charAt", e)
	if err != nil {
		return nil, err
	}

	runes := []rune(str)
	if idx < 0 || idx >= len(runes) {
		return nil, newPosError(fmt.Sprintf("charAt index %d out of range", idx), e.Pos())
	}

	return string(runes[idx]), nil
}

// callBuiltinCharCodeAt implements the charCodeAt(string, index) function.
// Uses rune (Unicode code point) indices.
func callBuiltinCharCodeAt(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 2, "charCodeAt requires exactly 2 arguments: string, index", e); err != nil {
		return nil, err
	}

	str, err := assertStringArg(args[0], 1, "charCodeAt", e)
	if err != nil {
		return nil, err
	}

	idx, err := assertIntArg(args[1], 2, "charCodeAt", e)
	if err != nil {
		return nil, err
	}

	runes := []rune(str)
	if idx < 0 || idx >= len(runes) {
		return nil, newPosError(fmt.Sprintf("charCodeAt index %d out of range", idx), e.Pos())
	}

	return int(runes[idx]), nil
}

// callBuiltinFromCharCode implements the fromCharCode(code) function.
func callBuiltinFromCharCode(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "fromCharCode requires exactly 1 argument: code", e); err != nil {
		return nil, err
	}

	code, err := assertIntArg(args[0], 1, "fromCharCode", e)
	if err != nil {
		return nil, err
	}

	r := rune(code)
	if code < 0 || code > utf8.MaxRune || !utf8.ValidRune(r) {
		return nil, newPosError(fmt.Sprintf("fromCharCode invalid code point %d", code), e.Pos())
	}

	return string(r), nil
}

// callBuiltinIndexOf implements the indexOf(source, value) function.
// Supports strings (substring search) and arrays (value search).
func callBuiltinIndexOf(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 2, "indexOf requires exactly 2 arguments: source, value", e); err != nil {
		return nil, err
	}

	switch source := args[0].(type) {
	case string:
		needle, err := assertStringArg(args[1], 2, "indexOf", e)
		if err != nil {
			return nil, err
		}
		byteIdx := strings.Index(source, needle)
		if byteIdx < 0 {
			return -1, nil
		}
		return runeIndexFromByteOffset(source, byteIdx), nil
	case Array:
		target := args[1]
		for i, item := range source {
			if isEqual(item, target) {
				return i, nil
			}
		}
		return -1, nil
	default:
		return nil, newUnsupportedTypeError("indexOf", args[0], e.Pos())
	}
}

// callBuiltinLastIndexOf implements the lastIndexOf(source, value) function.
// Supports strings (substring search) and arrays (value search).
func callBuiltinLastIndexOf(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 2, "lastIndexOf requires exactly 2 arguments: source, value", e); err != nil {
		return nil, err
	}

	switch source := args[0].(type) {
	case string:
		needle, err := assertStringArg(args[1], 2, "lastIndexOf", e)
		if err != nil {
			return nil, err
		}
		byteIdx := strings.LastIndex(source, needle)
		if byteIdx < 0 {
			return -1, nil
		}
		return runeIndexFromByteOffset(source, byteIdx), nil
	case Array:
		target := args[1]
		for i := len(source) - 1; i >= 0; i-- {
			if isEqual(source[i], target) {
				return i, nil
			}
		}
		return -1, nil
	default:
		return nil, newUnsupportedTypeError("lastIndexOf", args[0], e.Pos())
	}
}

// callBuiltinToUpper implements the toUpper(string) function.
func callBuiltinToUpper(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "toUpper requires exactly 1 argument", e); err != nil {
		return nil, err
	}

	str, err := assertStringArg(args[0], 0, "toUpper", e)
	if err != nil {
		return nil, err
	}

	return strings.ToUpper(str), nil
}

// callBuiltinToLower implements the toLower(string) function.
func callBuiltinToLower(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "toLower requires exactly 1 argument", e); err != nil {
		return nil, err
	}

	str, err := assertStringArg(args[0], 0, "toLower", e)
	if err != nil {
		return nil, err
	}

	return strings.ToLower(str), nil
}

// callBuiltinAppendIfMissing implements appendIfMissing(text, suffix).
func callBuiltinAppendIfMissing(args []Value, e *ast.CallExpr) (Value, error) {
	strs, err := assertStringArgs(args, 2, "appendIfMissing", e)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(strs[0], strs[1]) {
		return strs[0], nil
	}
	return strs[0] + strs[1], nil
}

// callBuiltinPrependIfMissing implements prependIfMissing(text, prefix).
func callBuiltinPrependIfMissing(args []Value, e *ast.CallExpr) (Value, error) {
	strs, err := assertStringArgs(args, 2, "prependIfMissing", e)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(strs[0], strs[1]) {
		return strs[0], nil
	}
	return strs[1] + strs[0], nil
}
