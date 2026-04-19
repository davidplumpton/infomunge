package evaluator

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"
)

// callBuiltinToRadix implements toRadix(number, radix).
func callBuiltinToRadix(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 2 {
		return nil, newPosError("toRadix requires exactly 2 arguments: number, radix", e.Pos())
	}

	number, err := toInt(args[0], "toRadix", e)
	if err != nil {
		return nil, err
	}
	radix, err := toInt(args[1], "toRadix", e)
	if err != nil {
		return nil, err
	}

	if radix < 2 || radix > 36 {
		return nil, newPosError("toRadix: radix must be between 2 and 36", e.Pos())
	}

	return strings.ToUpper(strconv.FormatInt(int64(number), radix)), nil
}

// callBuiltinFromRadix implements fromRadix(text, radix).
func callBuiltinFromRadix(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 2 {
		return nil, newPosError("fromRadix requires exactly 2 arguments: text, radix", e.Pos())
	}

	text, err := assertStringArg(args[0], 1, "fromRadix", e)
	if err != nil {
		return nil, err
	}
	radix, err := toInt(args[1], "fromRadix", e)
	if err != nil {
		return nil, err
	}

	if radix < 2 || radix > 36 {
		return nil, newPosError("fromRadix: radix must be between 2 and 36", e.Pos())
	}

	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, newPosError("fromRadix: value cannot be empty", e.Pos())
	}

	n, parseErr := strconv.ParseInt(trimmed, radix, 64)
	if parseErr != nil {
		return nil, newPosError(fmt.Sprintf("fromRadix: invalid value for radix %d", radix), e.Pos())
	}

	return int(n), nil
}

// callBuiltinToBinary implements toBinary(number).
func callBuiltinToBinary(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 1 {
		return nil, newPosError("toBinary requires exactly 1 argument: number", e.Pos())
	}
	return callBuiltinToRadix(Array{args[0], 2}, e)
}

// callBuiltinFromBinary implements fromBinary(text).
func callBuiltinFromBinary(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 1 {
		return nil, newPosError("fromBinary requires exactly 1 argument: text", e.Pos())
	}
	return callBuiltinFromRadix(Array{args[0], 2}, e)
}
