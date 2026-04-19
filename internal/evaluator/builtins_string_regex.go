package evaluator

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"
	"strings"
)

// callBuiltinRegex implements the regex(pattern, flags) function.
// It compiles the regex and returns a Regex value.
func callBuiltinRegex(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, newPosError("regex requires 1 or 2 arguments: pattern, [flags]", e.Pos())
	}

	pattern, err := assertStringArg(args[0], 1, "regex", e)
	if err != nil {
		return nil, err
	}

	flags := ""
	if len(args) == 2 {
		f, err := assertStringArg(args[1], 2, "regex", e)
		if err != nil {
			return nil, err
		}
		flags = f
	}

	re, err := compileRegex(pattern, flags)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("invalid regex pattern: %s", err), e.Pos())
	}

	return &Regex{
		Pattern: pattern,
		Flags:   flags,
		Re:      re,
	}, nil
}

// helper to get a compiled regex from either a Regex object or a string pattern
func getCompiledRegex(arg Value, explicitFlags string, pos token.Pos) (*regexp.Regexp, error) {
	switch v := arg.(type) {
	case *Regex:
		// If explicit flags are provided with a Regex object, it's ambiguous/unsupported
		if explicitFlags != "" {
			return nil, newPosError("cannot provide flags when using a Regex object", pos)
		}
		return v.Re, nil
	case string:
		re, err := compileRegex(v, explicitFlags)
		if err != nil {
			return nil, newPosError(fmt.Sprintf("invalid regex pattern: %s", err), pos)
		}
		return re, nil
	default:
		return nil, newPosError(fmt.Sprintf("expected string or Regex, got %T", arg), pos)
	}
}

// callBuiltinStartsWith implements the startsWith(string, prefix) function.
func callBuiltinStartsWith(args []Value, e *ast.CallExpr) (Value, error) {
	strs, err := assertStringArgs(args, 2, "startsWith", e)
	if err != nil {
		return nil, err
	}

	return strings.HasPrefix(strs[0], strs[1]), nil
}

// callBuiltinEndsWith implements the endsWith(string, suffix) function.
func callBuiltinEndsWith(args []Value, e *ast.CallExpr) (Value, error) {
	strs, err := assertStringArgs(args, 2, "endsWith", e)
	if err != nil {
		return nil, err
	}

	return strings.HasSuffix(strs[0], strs[1]), nil
}

// callBuiltinContains implements the contains(string, pattern) function.
// For strings, the pattern is treated as a regex only if it looks like a regex pattern or is a Regex object.
func callBuiltinContains(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 2 {
		return nil, newPosError("contains requires exactly 2 arguments", e.Pos())
	}

	switch first := args[0].(type) {
	case string:
		// Check for Regex object
		if r, ok := args[1].(*Regex); ok {
			return r.Re.MatchString(first), nil
		}

		// Check for string pattern
		pattern, ok := args[1].(string)
		if !ok {
			return nil, newPosError(fmt.Sprintf("contains expects argument 2 to be string or Regex, got %T", args[1]), e.Pos())
		}

		// Only use regex if pattern looks like a regex
		if looksLikeRegex(pattern) {
			re, regexErr := compileRegex(pattern, "")
			if regexErr == nil {
				return re.MatchString(first), nil
			}
		}
		// Fall back to substring matching
		return strings.Contains(first, pattern), nil
	case Array:
		for _, item := range first {
			if isEqual(item, args[1]) {
				return true, nil
			}
		}
		return false, nil
	default:
		return nil, newPosError(fmt.Sprintf("contains expects a string or array as argument 1, got %T", args[0]), e.Pos())
	}
}

// callBuiltinReplace implements the replace(string, pattern, replacement) function.
// The pattern is treated as a regex only if it looks like a regex pattern or is a Regex object.
func callBuiltinReplace(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 3 {
		return nil, newPosError("replace requires exactly 3 arguments: string, pattern, replacement", e.Pos())
	}

	text, err := assertStringArg(args[0], 1, "replace", e)
	if err != nil {
		return nil, err
	}

	replacement, err := assertStringArg(args[2], 3, "replace", e)
	if err != nil {
		return nil, err
	}

	// Check if pattern is a Regex object
	if r, ok := args[1].(*Regex); ok {
		return r.Re.ReplaceAllString(text, replacement), nil
	}

	// Check if pattern is a string
	pattern, ok := args[1].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("replace expects argument 2 to be string or Regex, got %T", args[1]), e.Pos())
	}

	// Only use regex if pattern looks like a regex
	if looksLikeRegex(pattern) {
		re, regexErr := compileRegex(pattern, "")
		if regexErr == nil {
			return re.ReplaceAllString(text, replacement), nil
		}
	}

	// Fall back to literal string replacement
	return strings.ReplaceAll(text, pattern, replacement), nil
}

// callBuiltinMatch implements the match(string, regex) function.
func callBuiltinMatch(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, newPosError("match requires 2 or 3 arguments: string, regex, [flags]", e.Pos())
	}

	if err := assertArg(args[0], beString(), 1, "match", e); err != nil {
		return nil, err
	}
	text := args[0].(string)

	flags := ""
	if len(args) == 3 {
		if err := assertArg(args[2], beString(), 3, "match", e); err != nil {
			return nil, err
		}
		flags = args[2].(string)
	}

	re, err := getCompiledRegex(args[1], flags, e.Pos())
	if err != nil {
		return nil, err
	}

	// Find the first match
	matches := re.FindStringSubmatch(text)
	if matches == nil {
		return nil, nil
	}

	// Return the capture groups (excluding the full match at index 0)
	result := make(Array, len(matches)-1)
	for i := 1; i < len(matches); i++ {
		result[i-1] = matches[i]
	}
	return result, nil
}

// callBuiltinMatches implements the matches(string, regex) function.
func callBuiltinMatches(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, newPosError("matches requires 2 or 3 arguments: string, regex, [flags]", e.Pos())
	}

	text, ok := args[0].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("matches expects first argument to be a string, got %T", args[0]), e.Pos())
	}

	flags := ""
	if len(args) == 3 {
		f, ok := args[2].(string)
		if !ok {
			return nil, newPosError(fmt.Sprintf("matches expects third argument (flags) to be a string, got %T", args[2]), e.Pos())
		}
		flags = f
	}

	re, err := getCompiledRegex(args[1], flags, e.Pos())
	if err != nil {
		return nil, err
	}

	// Check if the entire string matches
	return re.MatchString(text), nil
}

// callBuiltinScan implements the scan(string, regex) function.
func callBuiltinScan(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, newPosError("scan requires 2 or 3 arguments: string, regex, [flags]", e.Pos())
	}

	text, ok := args[0].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("scan expects first argument to be a string, got %T", args[0]), e.Pos())
	}

	flags := ""
	if len(args) == 3 {
		f, ok := args[2].(string)
		if !ok {
			return nil, newPosError(fmt.Sprintf("scan expects third argument (flags) to be a string, got %T", args[2]), e.Pos())
		}
		flags = f
	}

	re, err := getCompiledRegex(args[1], flags, e.Pos())
	if err != nil {
		return nil, err
	}

	// Find all matches
	matches := re.FindAllStringSubmatch(text, -1)
	if matches == nil {
		return make(Array, 0), nil
	}

	// Convert matches to result
	result := make(Array, len(matches))
	for i, match := range matches {
		if len(match) == 1 {
			// Single match, return the string directly
			result[i] = match[0]
		} else {
			// Multiple capture groups, return as array
			groups := make(Array, len(match)-1)
			for j := 1; j < len(match); j++ {
				groups[j-1] = match[j]
			}
			result[i] = groups
		}
	}

	return result, nil
}
