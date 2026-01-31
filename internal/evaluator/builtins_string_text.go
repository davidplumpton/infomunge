package evaluator

import (
	"fmt"
	"go/ast"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// callBuiltinIsBlank implements the isBlank(value) function.
func callBuiltinIsBlank(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 1 {
		return nil, newPosError("isBlank requires exactly 1 argument", e.Pos())
	}

	switch v := args[0].(type) {
	case nil:
		return true, nil
	case string:
		return strings.TrimSpace(v) == "", nil
	case []interface{}:
		return len(v) == 0, nil
	case map[string]interface{}:
		return len(v) == 0, nil
	default:
		return false, nil
	}
}

// callBuiltinJoinBy implements the joinBy(array, separator) function.
func callBuiltinJoinBy(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 2 {
		return nil, newPosError("joinBy requires exactly 2 arguments: array, separator", e.Pos())
	}

	if err := assertArg(args[0], beArray(), 1, "joinBy", e); err != nil {
		return nil, err
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, newPosError(fmt.Sprintf("joinBy: expected array, got %T", args[0]), e.Pos())
	}

	if err := assertArg(args[1], beString(), 2, "joinBy", e); err != nil {
		return nil, err
	}
	separator, ok := args[1].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("joinBy: expected string separator, got %T", args[1]), e.Pos())
	}

	// Convert each element to string
	strElements := make([]string, len(arr))
	for i, elem := range arr {
		strElements[i] = fmt.Sprintf("%v", elem)
	}

	return strings.Join(strElements, separator), nil
}

// callBuiltinSplitBy implements the splitBy(string, separator) function.
// The separator is treated as a regex only if it looks like a regex pattern or is a Regex object.
func callBuiltinSplitBy(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 2 {
		return nil, newPosError("splitBy requires exactly 2 arguments: string, separator", e.Pos())
	}

	text, ok := args[0].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("splitBy expects first argument to be a string, got %T", args[0]), e.Pos())
	}

	// Check for Regex object
	if r, ok := args[1].(*Regex); ok {
		parts := r.Re.Split(text, -1)
		result := make([]interface{}, len(parts))
		for i, part := range parts {
			result[i] = part
		}
		return result, nil
	}

	separator, ok := args[1].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("splitBy expects second argument to be a string or Regex, got %T", args[1]), e.Pos())
	}

	// Only use regex if separator looks like a regex pattern
	if looksLikeRegexPattern(separator) {
		re, err := regexp.Compile(separator)
		if err == nil {
			parts := re.Split(text, -1)
			result := make([]interface{}, len(parts))
			for i, part := range parts {
				result[i] = part
			}
			return result, nil
		}
	}

	// Fall back to literal string split
	parts := strings.Split(text, separator)
	result := make([]interface{}, len(parts))
	for i, part := range parts {
		result[i] = part
	}
	return result, nil
}

// looksLikeRegexPattern returns true if the pattern appears to be a regex.
// This is used to decide whether to use regex or literal string matching.
func looksLikeRegexPattern(pattern string) bool {
	return strings.ContainsAny(pattern, "*+?()[]{}|^$\\") ||
		(len(pattern) > 1 && strings.Contains(pattern, "."))
}

// callBuiltinLower implements the lower(string) function.
func callBuiltinLower(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 1 {
		return nil, newPosError("lower requires exactly 1 argument: string", e.Pos())
	}

	text, ok := args[0].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("lower expects a string, got %T", args[0]), e.Pos())
	}

	return strings.ToLower(text), nil
}

// callBuiltinUpper implements the upper(string) function.
func callBuiltinUpper(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 1 {
		return nil, newPosError("upper requires exactly 1 argument: string", e.Pos())
	}

	text, ok := args[0].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("upper expects a string, got %T", args[0]), e.Pos())
	}

	return strings.ToUpper(text), nil
}

// splitCamelCase splits a string into words based on camelCase boundaries and whitespace.
func splitCamelCase(s string) []string {
	var words []string
	var currentWord []rune

	for i, r := range s {
		if unicode.IsSpace(r) {
			if len(currentWord) > 0 {
				words = append(words, string(currentWord))
				currentWord = nil
			}
			continue
		}

		// Check for camelCase boundary: uppercase letter preceded by lowercase
		if i > 0 && unicode.IsUpper(r) {
			prevRune := rune(s[i-1])
			if unicode.IsLower(prevRune) {
				// Boundary: lowercase followed by uppercase (e.g., "storeO" in "storeOfOrigin")
				if len(currentWord) > 0 {
					words = append(words, string(currentWord))
					currentWord = nil
				}
			}
		}

		currentWord = append(currentWord, r)
	}

	if len(currentWord) > 0 {
		words = append(words, string(currentWord))
	}

	return words
}

// callBuiltinCapitalize implements the capitalize(string) function.
func callBuiltinCapitalize(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	text, err := assertStringArg(args[0], 0, "capitalize", e)
	if err != nil {
		return nil, err
	}

	// Split on camelCase boundaries and whitespace, then capitalize each word
	words := splitCamelCase(text)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " "), nil
}

// lowercaseFirst returns the string with only its first character lowercased.
func lowercaseFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// callBuiltinCamelize implements the camelize(string) function.
func callBuiltinCamelize(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	text, err := assertStringArg(args[0], 0, "camelize", e)
	if err != nil {
		return nil, err
	}

	// Convert dash/underscore separated to camelCase, preserving existing camelCase boundaries
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == '-' || r == '_'
	})
	if len(parts) == 0 {
		return text, nil
	}
	// Lowercase only the first character of the first part, preserving the rest
	result := lowercaseFirst(parts[0])
	for _, part := range parts[1:] {
		if len(part) > 0 {
			// Uppercase only the first character, preserving the rest
			result += strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return result, nil
}

// callBuiltinDasherize implements the dasherize(string) function.
func callBuiltinDasherize(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	text, err := assertStringArg(args[0], 0, "dasherize", e)
	if err != nil {
		return nil, err
	}

	// Convert camelCase to dash-separated
	var result []rune
	for i, r := range text {
		if i > 0 && unicode.IsUpper(r) && (i+1 < len(text) && unicode.IsLower(rune(text[i+1])) || unicode.IsLower(rune(text[i-1]))) {
			result = append(result, '-')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result), nil
}

// callBuiltinUnderscore implements the underscore(string) function.
func callBuiltinUnderscore(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	text, err := assertStringArg(args[0], 0, "underscore", e)
	if err != nil {
		return nil, err
	}

	// Convert camelCase to underscore separated
	var result []rune
	for i, r := range text {
		if i > 0 && unicode.IsUpper(r) && (i+1 < len(text) && unicode.IsLower(rune(text[i+1])) || unicode.IsLower(rune(text[i-1]))) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result), nil
}

// callBuiltinLeftPad implements the leftPad(string, size, padText) function.
// Pads the string on the left with padText until the string reaches the specified size.
func callBuiltinLeftPad(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 3 {
		return nil, newPosError("leftPad requires exactly 3 arguments: string, size, padText", e.Pos())
	}

	text, err := assertStringArg(args[0], 1, "leftPad", e)
	if err != nil {
		return nil, err
	}

	size, err := assertIntArg(args[1], 2, "leftPad", e)
	if err != nil {
		return nil, err
	}

	padText, err := assertStringArg(args[2], 3, "leftPad", e)
	if err != nil {
		return nil, err
	}

	if padText == "" {
		return nil, newPosError("leftPad padText cannot be empty", e.Pos())
	}

	// Count runes for proper Unicode handling
	textRunes := []rune(text)
	textLen := len(textRunes)

	if textLen >= size {
		return text, nil
	}

	// Calculate how many pad characters we need
	padNeeded := size - textLen

	// Build the padding
	var padding strings.Builder
	for padding.Len() < padNeeded*4 { // *4 to handle multi-byte runes
		padding.WriteString(padText)
		if utf8.RuneCountInString(padding.String()) >= padNeeded {
			break
		}
	}

	// Take exactly padNeeded runes from the padding
	paddingRunes := []rune(padding.String())
	if len(paddingRunes) > padNeeded {
		paddingRunes = paddingRunes[:padNeeded]
	}

	return string(paddingRunes) + text, nil
}

// callBuiltinRightPad implements the rightPad(string, size, padText) function.
// Pads the string on the right with padText until the string reaches the specified size.
func callBuiltinRightPad(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 3 {
		return nil, newPosError("rightPad requires exactly 3 arguments: string, size, padText", e.Pos())
	}

	text, err := assertStringArg(args[0], 1, "rightPad", e)
	if err != nil {
		return nil, err
	}

	size, err := assertIntArg(args[1], 2, "rightPad", e)
	if err != nil {
		return nil, err
	}

	padText, err := assertStringArg(args[2], 3, "rightPad", e)
	if err != nil {
		return nil, err
	}

	if padText == "" {
		return nil, newPosError("rightPad padText cannot be empty", e.Pos())
	}

	// Count runes for proper Unicode handling
	textRunes := []rune(text)
	textLen := len(textRunes)

	if textLen >= size {
		return text, nil
	}

	// Calculate how many pad characters we need
	padNeeded := size - textLen

	// Build the padding
	var padding strings.Builder
	for padding.Len() < padNeeded*4 { // *4 to handle multi-byte runes
		padding.WriteString(padText)
		if utf8.RuneCountInString(padding.String()) >= padNeeded {
			break
		}
	}

	// Take exactly padNeeded runes from the padding
	paddingRunes := []rune(padding.String())
	if len(paddingRunes) > padNeeded {
		paddingRunes = paddingRunes[:padNeeded]
	}

	return text + string(paddingRunes), nil
}
