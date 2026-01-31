package evaluator

import (
	"fmt"
	"go/ast"
)

// callNoArgMatcher creates a wrapper for matchers that take no arguments
func callNoArgMatcher(name string, matcherFunc func() Matcher) func([]interface{}, *ast.CallExpr) (interface{}, error) {
	return func(args []interface{}, e *ast.CallExpr) (interface{}, error) {
		if len(args) != 0 {
			return nil, newPosError(fmt.Sprintf("%s() takes no arguments", name), e.Pos())
		}
		return matcherFunc(), nil
	}
}

// callBeArray creates a type matcher for arrays
func callBeArray(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	return callNoArgMatcher("beArray", beArray)(args, e)
}

// callBeObject creates a type matcher for objects
func callBeObject(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	return callNoArgMatcher("beObject", beObject)(args, e)
}

// callBeString creates a type matcher for strings
func callBeString(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	return callNoArgMatcher("beString", beString)(args, e)
}

// callBeNumber creates a type matcher for numbers
func callBeNumber(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	return callNoArgMatcher("beNumber", beNumber)(args, e)
}

// callBeBoolean creates a type matcher for booleans
func callBeBoolean(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	return callNoArgMatcher("beBoolean", beBoolean)(args, e)
}

// callBeNull creates a type matcher for null
func callBeNull(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	return callNoArgMatcher("beNull", beNull)(args, e)
}

// callBeEmpty creates a matcher for empty values
func callBeEmpty(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	return callNoArgMatcher("beEmpty", beEmpty)(args, e)
}

// callBeBlank creates a matcher for blank strings
func callBeBlank(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	return callNoArgMatcher("beBlank", beBlank)(args, e)
}

// callEqualTo creates an equality matcher
func callEqualTo(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 1 {
		return nil, newPosError("equalTo() requires exactly 1 argument", e.Pos())
	}
	return equalTo(args[0]), nil
}

// callBeGreaterThan creates a greater-than matcher
func callBeGreaterThan(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, newPosError("beGreaterThan() requires 1 or 2 arguments", e.Pos())
	}
	inclusive := false
	if len(args) == 2 {
		inc, ok := args[1].(bool)
		if !ok {
			return nil, newPosError("beGreaterThan() second argument must be boolean", e.Pos())
		}
		inclusive = inc
	}
	return beGreaterThan(args[0], inclusive), nil
}

// callBeLowerThan creates a less-than matcher
func callBeLowerThan(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, newPosError("beLowerThan() requires 1 or 2 arguments", e.Pos())
	}
	inclusive := false
	if len(args) == 2 {
		inc, ok := args[1].(bool)
		if !ok {
			return nil, newPosError("beLowerThan() second argument must be boolean", e.Pos())
		}
		inclusive = inc
	}
	return beLowerThan(args[0], inclusive), nil
}

// callBeOneOf creates a one-of matcher
func callBeOneOf(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 1 {
		return nil, newPosError("beOneOf() requires exactly 1 argument (array of options)", e.Pos())
	}
	options, ok := args[0].([]interface{})
	if !ok {
		return nil, newPosError("beOneOf() argument must be an array", e.Pos())
	}
	return beOneOf(options), nil
}

// callStringMatcher creates a wrapper for string argument matchers
func callStringMatcher(name string, matcherFunc func(string) Matcher) func([]interface{}, *ast.CallExpr) (interface{}, error) {
	return func(args []interface{}, e *ast.CallExpr) (interface{}, error) {
		if len(args) != 1 {
			return nil, newPosError(fmt.Sprintf("%s() requires exactly 1 argument", name), e.Pos())
		}
		str, ok := args[0].(string)
		if !ok {
			return nil, newPosError(fmt.Sprintf("%s() argument must be a string", name), e.Pos())
		}
		return matcherFunc(str), nil
	}
}

// callContainStr creates a string contains matcher
func callContainStr(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	return callStringMatcher("containStr", containString)(args, e)
}

// callContainVal creates an array contains matcher
func callContainVal(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 1 {
		return nil, newPosError("containVal() requires exactly 1 argument", e.Pos())
	}
	return containValue(args[0]), nil
}

// callStartWith creates a string prefix matcher
func callStartWith(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	return callStringMatcher("startWith", startWith)(args, e)
}

// callEndWith creates a string suffix matcher
func callEndWith(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	return callStringMatcher("endWith", endWith)(args, e)
}

// callHaveSize creates a size matcher
func callHaveSize(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 1 {
		return nil, newPosError("haveSize() requires exactly 1 argument", e.Pos())
	}
	size, ok := args[0].(float64)
	if !ok {
		sizeInt, ok := args[0].(int)
		if !ok {
			return nil, newPosError("haveSize() argument must be a number", e.Pos())
		}
		size = float64(sizeInt)
	}
	return haveSize(int(size)), nil
}

// callHaveKey creates a key matcher
func callHaveKey(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 1 {
		return nil, newPosError("haveKey() requires exactly 1 argument", e.Pos())
	}
	key, ok := args[0].(string)
	if !ok {
		return nil, newPosError("haveKey() argument must be a string", e.Pos())
	}
	return haveKey(key), nil
}

// callHaveValue creates a value matcher
func callHaveValue(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 1 {
		return nil, newPosError("haveValue() requires exactly 1 argument", e.Pos())
	}
	return haveValue(args[0]), nil
}

// callNotBeNull creates a not-null matcher
func callNotBeNull(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 0 {
		return nil, newPosError("notBeNull() takes no arguments", e.Pos())
	}
	return notBeNull(), nil
}

// callMust implements the must(value, matcher(s)) function
// It can take 2 arguments (value, matcher) or (value, array of matchers)
func callMust(e *ast.CallExpr, context map[string]interface{}, depth int) (interface{}, error) {
	if len(e.Args) < 2 || len(e.Args) > 2 {
		return nil, newPosError("must() requires exactly 2 arguments: value and matcher(s)", e.Pos())
	}

	// Evaluate the value
	value, err := evalASTWithDepth(e.Args[0], context, depth)
	if err != nil {
		return nil, err
	}

	// Evaluate the matcher(s)
	matcherArg, err := evalASTWithDepth(e.Args[1], context, depth)
	if err != nil {
		return nil, err
	}

	// Handle single matcher or array of matchers
	var matchers []Matcher
	switch m := matcherArg.(type) {
	case Matcher:
		matchers = []Matcher{m}
	case []interface{}:
		// Array of matchers
		for i, item := range m {
			matcher, ok := item.(Matcher)
			if !ok {
				return nil, newPosError(fmt.Sprintf("matcher at index %d is not a valid matcher", i), e.Args[1].Pos())
			}
			matchers = append(matchers, matcher)
		}
	default:
		return nil, newPosError(fmt.Sprintf("expected matcher or array of matchers, got %T", matcherArg), e.Args[1].Pos())
	}

	// Apply all matchers
	for _, matcher := range matchers {
		result := matcher(value)
		if !result.Success {
			return nil, newPosError(fmt.Sprintf("assertion failed: %s", result.Message), e.Pos())
		}
	}

	// Return the Matched result on success
	return Matched, nil
}

// callEachItemMatcher implements the eachItem(matcher) function
func callEachItemMatcher(e *ast.CallExpr, context map[string]interface{}, depth int) (interface{}, error) {
	if len(e.Args) != 1 {
		return nil, newPosError("eachItem() requires exactly 1 argument (matcher)", e.Pos())
	}

	// Evaluate the matcher
	matcherArg, err := evalASTWithDepth(e.Args[0], context, depth)
	if err != nil {
		return nil, err
	}

	matcher, ok := matcherArg.(Matcher)
	if !ok {
		return nil, newPosError(fmt.Sprintf("expected matcher, got %T", matcherArg), e.Args[0].Pos())
	}

	return eachItem(matcher), nil
}

// callHaveItemMatcher implements the haveItem(matcher) function
func callHaveItemMatcher(e *ast.CallExpr, context map[string]interface{}, depth int) (interface{}, error) {
	if len(e.Args) != 1 {
		return nil, newPosError("haveItem() requires exactly 1 argument (matcher)", e.Pos())
	}

	// Evaluate the matcher
	matcherArg, err := evalASTWithDepth(e.Args[0], context, depth)
	if err != nil {
		return nil, err
	}

	matcher, ok := matcherArg.(Matcher)
	if !ok {
		return nil, newPosError(fmt.Sprintf("expected matcher, got %T", matcherArg), e.Args[0].Pos())
	}

	return haveItem(matcher), nil
}

// callAnyOfMatcher implements the anyOf(matchers) function
func callAnyOfMatcher(e *ast.CallExpr, context map[string]interface{}, depth int) (interface{}, error) {
	if len(e.Args) != 1 {
		return nil, newPosError("anyOf() requires exactly 1 argument (array of matchers)", e.Pos())
	}

	// Evaluate the matchers array
	matchersArg, err := evalASTWithDepth(e.Args[0], context, depth)
	if err != nil {
		return nil, err
	}

	matchersArray, ok := matchersArg.([]interface{})
	if !ok {
		return nil, newPosError(fmt.Sprintf("expected array of matchers, got %T", matchersArg), e.Args[0].Pos())
	}

	var matchers []Matcher
	for i, item := range matchersArray {
		matcher, ok := item.(Matcher)
		if !ok {
			return nil, newPosError(fmt.Sprintf("matcher at index %d is not a valid matcher", i), e.Args[0].Pos())
		}
		matchers = append(matchers, matcher)
	}

	return anyOf(matchers), nil
}

// callNotBeMatcher implements the notBe(matcher) function
func callNotBeMatcher(e *ast.CallExpr, context map[string]interface{}, depth int) (interface{}, error) {
	if len(e.Args) != 1 {
		return nil, newPosError("notBe() requires exactly 1 argument (matcher)", e.Pos())
	}

	// Evaluate the matcher
	matcherArg, err := evalASTWithDepth(e.Args[0], context, depth)
	if err != nil {
		return nil, err
	}

	matcher, ok := matcherArg.(Matcher)
	if !ok {
		return nil, newPosError(fmt.Sprintf("expected matcher, got %T", matcherArg), e.Args[0].Pos())
	}

	return notBe(matcher), nil
}
