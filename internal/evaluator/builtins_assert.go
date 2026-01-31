package evaluator

import (
	"fmt"
	"go/ast"
)

// callBuiltinAssert implements the assert(value, matcher, [message]) function.
func callBuiltinAssert(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, newPosError("assert expects 2 or 3 arguments: value, matcher, [message]", e.Pos())
	}

	value := args[0]
	matcher, ok := args[1].(Matcher)
	if !ok {
		return nil, newPosError(fmt.Sprintf("assert expects second argument to be a matcher, got %T", args[1]), e.Pos())
	}

	customMessage := ""
	if len(args) == 3 {
		msg, ok := args[2].(string)
		if !ok {
			return nil, newPosError(fmt.Sprintf("assert expects third argument (message) to be a string, got %T", args[2]), e.Pos())
		}
		customMessage = msg
	}

	result := matcher(value)
	if !result.Success {
		msg := result.Message
		if customMessage != "" {
			msg = fmt.Sprintf("%s: %s", customMessage, result.Message)
		}
		return nil, newPosError(fmt.Sprintf("assertion failed: %s", msg), e.Pos())
	}

	return value, nil
}

// callBuiltinAssertThat is an alias for assert for better readability in some contexts.
func callBuiltinAssertThat(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	return callBuiltinAssert(args, e)
}
