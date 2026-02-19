package evaluator

import (
	"crypto/rand"
	"fmt"
	"go/ast"
	"math"
)

// callBuiltinCeil implements the ceil(number) function.
func callBuiltinCeil(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 1, "ceil function requires exactly 1 argument", e); err != nil {
		return nil, err
	}
	num, err := toNumber(args[0], "ceil", e)
	if err != nil {
		return nil, err
	}
	return math.Ceil(num), nil
}

// callBuiltinFloor implements the floor(number) function.
func callBuiltinFloor(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 1, "floor function requires exactly 1 argument", e); err != nil {
		return nil, err
	}
	num, err := toNumber(args[0], "floor", e)
	if err != nil {
		return nil, err
	}
	return math.Floor(num), nil
}

// callBuiltinRound implements the round(number) function.
func callBuiltinRound(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 1, "round function requires exactly 1 argument", e); err != nil {
		return nil, err
	}
	num, err := toNumber(args[0], "round", e)
	if err != nil {
		return nil, err
	}
	return math.Round(num), nil
}

// callBuiltinSqrt implements the sqrt(number) function.
func callBuiltinSqrt(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 1, "sqrt function requires exactly 1 argument", e); err != nil {
		return nil, err
	}
	num, err := toNumber(args[0], "sqrt", e)
	if err != nil {
		return nil, err
	}
	if num < 0 {
		return nil, newPosError(fmt.Sprintf("sqrt: cannot take square root of negative number %v", num), e.Pos())
	}
	return math.Sqrt(num), nil
}

// callBuiltinAbs implements the abs(number) function.
func callBuiltinAbs(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 1, "abs function requires exactly 1 argument", e); err != nil {
		return nil, err
	}
	num, err := toNumber(args[0], "abs", e)
	if err != nil {
		return nil, err
	}
	return math.Abs(num), nil
}

// callBuiltinMax implements the max(...numbers) or max(array) function.
func callBuiltinMax(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireMinArgs(args, 1, "max function requires at least 1 argument", e); err != nil {
		return nil, err
	}
	if len(args) == 1 {
		if arr, ok := args[0].([]interface{}); ok {
			// max(array)
			if len(arr) == 0 {
				return nil, newPosError("max function requires non-empty array", e.Pos())
			}
			maxVal := math.Inf(-1)
			for _, item := range arr {
				num, err := toNumber(item, "max", e)
				if err != nil {
					return nil, err
				}
				if num > maxVal {
					maxVal = num
				}
			}
			return maxVal, nil
		}
	}
	// max(...numbers)
	maxVal := math.Inf(-1)
	for _, arg := range args {
		num, err := toNumber(arg, "max", e)
		if err != nil {
			return nil, err
		}
		if num > maxVal {
			maxVal = num
		}
	}
	return maxVal, nil
}

// callBuiltinMin implements the min(...numbers) or min(array) function.
func callBuiltinMin(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireMinArgs(args, 1, "min function requires at least 1 argument", e); err != nil {
		return nil, err
	}
	if len(args) == 1 {
		if arr, ok := args[0].([]interface{}); ok {
			// min(array)
			if len(arr) == 0 {
				return nil, newPosError("min function requires non-empty array", e.Pos())
			}
			minVal := math.Inf(1)
			for _, item := range arr {
				num, err := toNumber(item, "min", e)
				if err != nil {
					return nil, err
				}
				if num < minVal {
					minVal = num
				}
			}
			return minVal, nil
		}
	}
	// min(...numbers)
	minVal := math.Inf(1)
	for _, arg := range args {
		num, err := toNumber(arg, "min", e)
		if err != nil {
			return nil, err
		}
		if num < minVal {
			minVal = num
		}
	}
	return minVal, nil
}

// callBuiltinPow implements the pow(base, exponent) function.
func callBuiltinPow(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 2, "pow function requires exactly 2 arguments: base, exponent", e); err != nil {
		return nil, err
	}
	base, err := toNumber(args[0], "pow", e)
	if err != nil {
		return nil, err
	}
	exp, err := toNumber(args[1], "pow", e)
	if err != nil {
		return nil, err
	}
	return math.Pow(base, exp), nil
}

// toNumber converts a value to a float64 number for math operations.
func toNumber(val interface{}, funcName string, e *ast.CallExpr) (float64, error) {
	if num, ok := ToFloat(val); ok {
		return num, nil
	}
	switch v := val.(type) {
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case nil:
		return 0, nil
	default:
		return 0, newPosError(fmt.Sprintf("%s: cannot convert %T to number", funcName, val), e.Pos())
	}
}

// callBuiltinSum implements the sum(array) function.
func callBuiltinSum(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 1, "sum requires exactly 1 argument: array", e); err != nil {
		return nil, err
	}

	if err := assertArg(args[0], beArray(), 1, "sum", e); err != nil {
		return nil, err
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, newPosError(fmt.Sprintf("sum: expected array, got %T", args[0]), e.Pos())
	}

	var sum float64
	for i, item := range arr {
		num, ok := ToFloat(item)
		if !ok {
			return nil, newElementNotNumberError("sum", i, item, e.Pos())
		}
		sum += num
	}

	return sum, nil
}

// callBuiltinAvg implements the avg(array) function.
// Returns the average of all numbers in the array.
func callBuiltinAvg(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 1, "avg requires exactly 1 argument: array", e); err != nil {
		return nil, err
	}

	if err := assertArg(args[0], beArray(), 1, "avg", e); err != nil {
		return nil, err
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, newPosError(fmt.Sprintf("avg: expected array, got %T", args[0]), e.Pos())
	}

	if len(arr) == 0 {
		return nil, newPosError("avg: cannot calculate average of empty array", e.Pos())
	}

	var sum float64
	for i, item := range arr {
		num, ok := ToFloat(item)
		if !ok {
			return nil, newElementNotNumberError("avg", i, item, e.Pos())
		}
		sum += num
	}

	return sum / float64(len(arr)), nil
}

// callBuiltinIsDecimal implements the isDecimal(value) function.
func callBuiltinIsDecimal(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 1, "isDecimal requires exactly 1 argument", e); err != nil {
		return nil, err
	}

	switch args[0].(type) {
	case float64:
		// It's a float64 type
		return true, nil
	case int:
		// It's an integer type
		return false, nil
	case string, nil:
		return nil, newPosError(fmt.Sprintf("isDecimal expects a number"), e.Pos())
	default:
		return nil, newPosError(fmt.Sprintf("isDecimal expects a number"), e.Pos())
	}
}

// callBuiltinIsInteger implements the isInteger(value) function.
// Returns true if value is an integer (no fractional part).
func callBuiltinIsInteger(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 1, "isInteger requires exactly 1 argument", e); err != nil {
		return nil, err
	}

	switch v := args[0].(type) {
	case float64:
		// Check if the float is actually an integer (no fractional part)
		return v == float64(int64(v)), nil
	case int:
		return true, nil
	case string, nil:
		return nil, newPosError(fmt.Sprintf("isInteger expects a number, got %T", args[0]), e.Pos())
	default:
		return nil, newPosError(fmt.Sprintf("isInteger expects a number, got %T", args[0]), e.Pos())
	}
}

// callBuiltinIsEven implements the isEven(value) function.
// Returns true if value is an even integer.
func callBuiltinIsEven(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 1, "isEven requires exactly 1 argument", e); err != nil {
		return nil, err
	}

	var num int64
	switch v := args[0].(type) {
	case float64:
		if v != float64(int64(v)) {
			return nil, newPosError(fmt.Sprintf("isEven expects an integer, got float %v", v), e.Pos())
		}
		num = int64(v)
	case int:
		num = int64(v)
	case string:
		return nil, newPosError(fmt.Sprintf("isEven expects an integer, got string"), e.Pos())
	default:
		return nil, newPosError(fmt.Sprintf("isEven expects an integer, got %T", args[0]), e.Pos())
	}

	return num%ParityDivisor == 0, nil
}

// callBuiltinIsOdd implements the isOdd(value) function.
// Returns true if value is an odd integer.
func callBuiltinIsOdd(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 1, "isOdd requires exactly 1 argument", e); err != nil {
		return nil, err
	}

	var num int64
	switch v := args[0].(type) {
	case float64:
		if v != float64(int64(v)) {
			return nil, newPosError(fmt.Sprintf("isOdd expects an integer, got float %v", v), e.Pos())
		}
		num = int64(v)
	case int:
		num = int64(v)
	case string:
		return nil, newPosError(fmt.Sprintf("isOdd expects an integer, got string"), e.Pos())
	default:
		return nil, newPosError(fmt.Sprintf("isOdd expects an integer, got %T", args[0]), e.Pos())
	}

	return num%ParityDivisor != 0, nil
}

// callBuiltinRandom implements the random() function.
func callBuiltinRandom(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 0 {
		return nil, newPosError("random takes no arguments", e.Pos())
	}

	// Get a random number between 0 and 1
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("failed to generate random number: %s", err), e.Pos())
	}

	// Convert to a float between 0 and 1
	randomInt := int64(0)
	for i := 0; i < 8; i++ {
		randomInt = (randomInt << 8) | int64(randomBytes[i])
	}
	// Ensure non-negative and in range [0, 1)
	randomFloat := float64(randomInt&0x7FFFFFFFFFFFFFFF) / float64(0x7FFFFFFFFFFFFFFF)
	return randomFloat, nil
}

// callBuiltinRandomInt implements the randomInt(max) function.
// Returns a random integer from 0 to max (exclusive).
func callBuiltinRandomInt(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 1, "randomInt requires exactly 1 argument: max", e); err != nil {
		return nil, err
	}

	var max int64
	switch v := args[0].(type) {
	case float64:
		max = int64(v)
	case int:
		max = int64(v)
	default:
		return nil, newPosError(fmt.Sprintf("randomInt expects a number, got %T", args[0]), e.Pos())
	}

	if max <= 0 {
		return nil, newPosError("randomInt max must be greater than 0", e.Pos())
	}

	// Generate a random integer
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("failed to generate random number: %s", err), e.Pos())
	}

	randomInt := int64(0)
	for i := 0; i < 8; i++ {
		randomInt = (randomInt << 8) | int64(randomBytes[i])
	}

	// Ensure non-negative and apply modulo
	result := (randomInt & 0x7FFFFFFFFFFFFFFF) % max
	return float64(result), nil
}

// callBuiltinTo implements the to(start, end) function.
func callBuiltinTo(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 2, "to requires exactly 2 arguments: start, end", e); err != nil {
		return nil, err
	}

	var start, end int64
	switch v := args[0].(type) {
	case float64:
		start = int64(v)
	case int:
		start = int64(v)
	default:
		return nil, newPosError(fmt.Sprintf("to expects start to be a number, got %T", args[0]), e.Pos())
	}

	switch v := args[1].(type) {
	case float64:
		end = int64(v)
	case int:
		end = int64(v)
	default:
		return nil, newPosError(fmt.Sprintf("to expects end to be a number, got %T", args[1]), e.Pos())
	}

	// Build the range
	var result []interface{}
	if start <= end {
		for i := start; i <= end; i++ {
			result = append(result, float64(i))
		}
	} else {
		for i := start; i >= end; i-- {
			result = append(result, float64(i))
		}
	}

	return result, nil
}

// callBuiltinMod implements the mod(dividend, divisor) function.
func callBuiltinMod(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 2, "mod requires exactly 2 arguments: dividend, divisor", e); err != nil {
		return nil, err
	}

	var dividend, divisor float64
	switch v := args[0].(type) {
	case float64:
		dividend = v
	case int:
		dividend = float64(v)
	default:
		return nil, newPosError(fmt.Sprintf("mod expects dividend to be a number, got %T", args[0]), e.Pos())
	}

	switch v := args[1].(type) {
	case float64:
		divisor = v
	case int:
		divisor = float64(v)
	default:
		return nil, newPosError(fmt.Sprintf("mod expects divisor to be a number, got %T", args[1]), e.Pos())
	}

	if divisor == 0 {
		return nil, newPosError("mod: division by zero", e.Pos())
	}

	return math.Mod(dividend, divisor), nil
}
