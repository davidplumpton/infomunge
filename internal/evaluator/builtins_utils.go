package evaluator

import (
	"fmt"
	"go/ast"
	"regexp"
	"strings"
)

// requireExactArgs validates exact argument count and returns a positioned error.
func requireExactArgs(args []interface{}, expected int, errMsg string, e *ast.CallExpr) error {
	if len(args) != expected {
		return newPosError(errMsg, e.Pos())
	}
	return nil
}

// requireMinArgs validates minimum argument count and returns a positioned error.
func requireMinArgs(args []interface{}, min int, errMsg string, e *ast.CallExpr) error {
	if len(args) < min {
		return newPosError(errMsg, e.Pos())
	}
	return nil
}

// assertStringArg validates that an argument is a string.
func assertStringArg(val interface{}, argIndex int, funcName string, e *ast.CallExpr) (string, error) {
	str, ok := val.(string)
	if !ok {
		msg := fmt.Sprintf("%s expects a string", funcName)
		if argIndex > 0 {
			msg = fmt.Sprintf("%s expects a string as argument %d, got %T", funcName, argIndex, val)
		}
		return "", newPosError(msg, e.Pos())
	}
	return str, nil
}

// assertStringArgs validates that multiple arguments are strings.
func assertStringArgs(args []interface{}, count int, funcName string, e *ast.CallExpr) ([]string, error) {
	if len(args) != count {
		return nil, newPosError(fmt.Sprintf("%s requires exactly %d argument(s)", funcName, count), e.Pos())
	}

	result := make([]string, count)
	for i, arg := range args {
		if str, err := assertStringArg(arg, i+1, funcName, e); err != nil {
			return nil, err
		} else {
			result[i] = str
		}
	}
	return result, nil
}

// assertIntArg validates that an argument is an integer.
func assertIntArg(val interface{}, argIndex int, funcName string, e *ast.CallExpr) (int, error) {
	i, ok := val.(int)
	if !ok {
		return 0, newPosError(fmt.Sprintf("%s expects an integer as argument %d, got %T", funcName, argIndex, val), e.Pos())
	}
	return i, nil
}

// assertArg validates an argument using a matcher.
func assertArg(val interface{}, matcher Matcher, argIndex int, funcName string, e *ast.CallExpr) error {
	result := matcher(val)
	if !result.Success {
		msg := fmt.Sprintf("%s: argument %d %s", funcName, argIndex, result.Message)
		if argIndex == 0 {
			msg = fmt.Sprintf("%s: %s", funcName, result.Message)
		}
		return newPosError(msg, e.Pos())
	}
	return nil
}

// compileRegex compiles a regex pattern with optional flags.
func compileRegex(pattern string, flags string) (*regexp.Regexp, error) {
	prefix := ""
	if strings.Contains(flags, "i") {
		prefix += "(?i)"
	}
	if strings.Contains(flags, "m") {
		prefix += "(?m)"
	}
	if strings.Contains(flags, "s") {
		prefix += "(?s)"
	}
	return regexp.Compile(prefix + pattern)
}
