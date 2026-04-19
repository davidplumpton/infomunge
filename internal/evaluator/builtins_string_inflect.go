package evaluator

import (
	"fmt"
	"go/ast"
	"strings"
)

// callBuiltinPluralize implements the pluralize(string) function.
func callBuiltinPluralize(args []Value, e *ast.CallExpr) (Value, error) {
	text, err := requireOneStringArg(args, "pluralize", e)
	if err != nil {
		return nil, err
	}
	return pluralizeWord(text), nil
}

// callBuiltinSingularize implements the singularize(string) function.
func callBuiltinSingularize(args []Value, e *ast.CallExpr) (Value, error) {
	text, err := requireOneStringArg(args, "singularize", e)
	if err != nil {
		return nil, err
	}
	return singularizeWord(text), nil
}

// callBuiltinOrdinalize implements the ordinalize(number) function.
func callBuiltinOrdinalize(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "ordinalize requires exactly 1 argument", e); err != nil {
		return nil, err
	}

	num, err := assertIntArg(args[0], 1, "ordinalize", e)
	if err != nil {
		return nil, err
	}

	return formatOrdinal(num), nil
}

func pluralizeWord(word string) string {
	if word == "" {
		return word
	}

	lower := strings.ToLower(word)
	runes := []rune(word)
	lowerRunes := []rune(lower)

	if strings.HasSuffix(lower, "y") && len(lowerRunes) > 1 && !isVowel(lowerRunes[len(lowerRunes)-2]) {
		return string(runes[:len(runes)-1]) + "ies"
	}

	if strings.HasSuffix(lower, "ch") || strings.HasSuffix(lower, "sh") ||
		strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "x") || strings.HasSuffix(lower, "z") {
		return word + "es"
	}

	if strings.HasSuffix(lower, "fe") && len(runes) >= 2 {
		return string(runes[:len(runes)-2]) + "ves"
	}

	if strings.HasSuffix(lower, "f") && len(runes) >= 1 {
		return string(runes[:len(runes)-1]) + "ves"
	}

	return word + "s"
}

func singularizeWord(word string) string {
	if word == "" {
		return word
	}

	lower := strings.ToLower(word)
	runes := []rune(word)
	lowerRunes := []rune(lower)

	if strings.HasSuffix(lower, "ies") && len(runes) > 3 {
		return string(runes[:len(runes)-3]) + "y"
	}

	if strings.HasSuffix(lower, "ves") && len(runes) > 3 {
		base := string(runes[:len(runes)-3])
		baseLower := string(lowerRunes[:len(lowerRunes)-3])
		if strings.HasSuffix(baseLower, "i") {
			return base + "fe"
		}
		return base + "f"
	}

	if strings.HasSuffix(lower, "es") && len(runes) > 2 {
		baseLower := strings.ToLower(string(runes[:len(runes)-2]))
		if strings.HasSuffix(baseLower, "s") || strings.HasSuffix(baseLower, "x") || strings.HasSuffix(baseLower, "z") ||
			strings.HasSuffix(baseLower, "ch") || strings.HasSuffix(baseLower, "sh") {
			return string(runes[:len(runes)-2])
		}
	}

	if strings.HasSuffix(lower, "s") && !strings.HasSuffix(lower, "ss") && len(runes) > 1 {
		return string(runes[:len(runes)-1])
	}

	return word
}

func formatOrdinal(value int) string {
	abs := value
	if abs < 0 {
		abs = -abs
	}

	mod100 := abs % 100
	if mod100 >= 11 && mod100 <= 13 {
		return fmt.Sprintf("%dth", value)
	}

	switch abs % 10 {
	case 1:
		return fmt.Sprintf("%dst", value)
	case 2:
		return fmt.Sprintf("%dnd", value)
	case 3:
		return fmt.Sprintf("%drd", value)
	default:
		return fmt.Sprintf("%dth", value)
	}
}

func isVowel(r rune) bool {
	switch r {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	default:
		return false
	}
}
