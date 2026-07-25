package evaluator

import (
	"crypto/rand"
	"fmt"
	"go/ast"
	"math"
	"math/big"
)

// callBuiltinCeil implements the ceil(number) function.
func callBuiltinCeil(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "ceil function requires exactly 1 argument", e); err != nil {
		return nil, err
	}
	num, err := normalizedBuiltinNumber(args[0], "ceil", e)
	if err != nil {
		return nil, err
	}
	if integer, ok := num.(int); ok {
		return integer, nil
	}
	return math.Ceil(num.(float64)), nil
}

// callBuiltinFloor implements the floor(number) function.
func callBuiltinFloor(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "floor function requires exactly 1 argument", e); err != nil {
		return nil, err
	}
	num, err := normalizedBuiltinNumber(args[0], "floor", e)
	if err != nil {
		return nil, err
	}
	if integer, ok := num.(int); ok {
		return integer, nil
	}
	return math.Floor(num.(float64)), nil
}

// callBuiltinRound implements the round(number) function.
func callBuiltinRound(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "round function requires exactly 1 argument", e); err != nil {
		return nil, err
	}
	num, err := normalizedBuiltinNumber(args[0], "round", e)
	if err != nil {
		return nil, err
	}
	if integer, ok := num.(int); ok {
		return integer, nil
	}
	return math.Round(num.(float64)), nil
}

// callBuiltinSqrt implements the sqrt(number) function.
func callBuiltinSqrt(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "sqrt function requires exactly 1 argument", e); err != nil {
		return nil, err
	}
	num, err := normalizedBuiltinNumber(args[0], "sqrt", e)
	if err != nil {
		return nil, err
	}
	if integer, ok := num.(int); ok {
		if integer < 0 {
			return nil, newPosError(fmt.Sprintf("sqrt: cannot take square root of negative number %v", integer), e.Pos())
		}
		root := new(big.Int).Sqrt(big.NewInt(int64(integer)))
		if new(big.Int).Mul(new(big.Int).Set(root), root).Cmp(big.NewInt(int64(integer))) == 0 {
			return int(root.Int64()), nil
		}
		if !intExactlyRepresentableAsFloat(integer) {
			return nil, newPosError("sqrt: numeric precision loss converting integer input", e.Pos())
		}
		result := math.Sqrt(float64(integer))
		if err := validateFloat(result, "sqrt"); err != nil {
			return nil, newPosError(err.Error(), e.Pos())
		}
		return result, nil
	}
	floatValue := num.(float64)
	if floatValue < 0 {
		return nil, newPosError(fmt.Sprintf("sqrt: cannot take square root of negative number %v", num), e.Pos())
	}
	result := math.Sqrt(floatValue)
	if err := validateFloat(result, "sqrt"); err != nil {
		return nil, newPosError(err.Error(), e.Pos())
	}
	return result, nil
}

// callBuiltinAbs implements the abs(number) function.
func callBuiltinAbs(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "abs function requires exactly 1 argument", e); err != nil {
		return nil, err
	}
	num, err := normalizedBuiltinNumber(args[0], "abs", e)
	if err != nil {
		return nil, err
	}
	if integer, ok := num.(int); ok {
		if integer == minInt() {
			return nil, newPosError("integer overflow during abs", e.Pos())
		}
		if integer < 0 {
			return -integer, nil
		}
		return integer, nil
	}
	return math.Abs(num.(float64)), nil
}

// callBuiltinMax implements the max(...numbers) or max(array) function.
func callBuiltinMax(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireMinArgs(args, 1, "max function requires at least 1 argument", e); err != nil {
		return nil, err
	}
	if len(args) == 1 {
		if arr, ok := args[0].(Array); ok {
			if len(arr) == 0 {
				return nil, nil
			}
			return numericExtremum(arr, "max", true, e)
		}
	}
	return numericExtremum(args, "max", true, e)
}

// callBuiltinMin implements the min(...numbers) or min(array) function.
func callBuiltinMin(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireMinArgs(args, 1, "min function requires at least 1 argument", e); err != nil {
		return nil, err
	}
	if len(args) == 1 {
		if arr, ok := args[0].(Array); ok {
			if len(arr) == 0 {
				return nil, nil
			}
			return numericExtremum(arr, "min", false, e)
		}
	}
	return numericExtremum(args, "min", false, e)
}

// callBuiltinPow implements the pow(base, exponent) function.
func callBuiltinPow(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 2, "pow function requires exactly 2 arguments: base, exponent", e); err != nil {
		return nil, err
	}
	base, err := normalizedBuiltinNumber(args[0], "pow", e)
	if err != nil {
		return nil, err
	}
	exponent, err := normalizedBuiltinNumber(args[1], "pow", e)
	if err != nil {
		return nil, err
	}

	if expFloat, ok := exponent.(float64); ok && expFloat == 0 {
		return 1.0, nil
	}
	if expInt, ok := exponent.(int); ok && expInt == 0 {
		return 1, nil
	}
	if expFloat, ok := exponent.(float64); ok && expFloat == 1 {
		return base, nil
	}
	if expInt, ok := exponent.(int); ok && expInt == 1 {
		return base, nil
	}

	baseInt, baseIsInt := base.(int)
	expInt, expIsInt := exponent.(int)
	if baseIsInt && expIsInt && expInt >= 0 {
		result, err := checkedIntPow(baseInt, expInt)
		if err != nil {
			return nil, newPosError(err.Error(), e.Pos())
		}
		return result, nil
	}
	if baseIsInt && expIsInt && expInt < 0 {
		switch baseInt {
		case 0:
			return nil, newPosError("pow resulted in infinity", e.Pos())
		case 1:
			return 1, nil
		case -1:
			if expInt%2 == 0 {
				return 1, nil
			}
			return -1, nil
		}
	}

	baseFloat, err := exactBuiltinFloat(base, "pow", e)
	if err != nil {
		return nil, err
	}
	expFloat, err := exactBuiltinFloat(exponent, "pow", e)
	if err != nil {
		return nil, err
	}
	result := math.Pow(baseFloat, expFloat)
	// Preserve the established language result for non-real powers (NaN),
	// while still rejecting range overflow to infinity.
	if math.IsInf(result, 0) {
		return nil, newPosError("pow resulted in infinity", e.Pos())
	}
	return result, nil
}

// normalizedBuiltinNumber preserves runtime integers while retaining the
// historic bool/null coercions used by the unary math builtins.
func normalizedBuiltinNumber(val Value, funcName string, e *ast.CallExpr) (Value, error) {
	switch v := val.(type) {
	case int:
		return v, nil
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, newPosError(fmt.Sprintf("%s: expected a finite number", funcName), e.Pos())
		}
		return v, nil
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

func exactBuiltinFloat(val Value, funcName string, e *ast.CallExpr) (float64, error) {
	switch number := val.(type) {
	case int:
		if !intExactlyRepresentableAsFloat(number) {
			return 0, newPosError(fmt.Sprintf("%s: numeric precision loss converting integer input", funcName), e.Pos())
		}
		return float64(number), nil
	case float64:
		return number, nil
	default:
		return 0, newPosError(fmt.Sprintf("%s: cannot convert %T to number", funcName, val), e.Pos())
	}
}

func numericExtremum(values []Value, funcName string, wantMax bool, e *ast.CallExpr) (Value, error) {
	best, err := normalizedBuiltinNumber(values[0], funcName, e)
	if err != nil {
		return nil, err
	}
	bestRat, _ := exactNumericRat(best)
	for _, candidateValue := range values[1:] {
		candidate, err := normalizedBuiltinNumber(candidateValue, funcName, e)
		if err != nil {
			return nil, err
		}
		candidateRat, _ := exactNumericRat(candidate)
		comparison := candidateRat.Cmp(bestRat)
		if (wantMax && comparison > 0) || (!wantMax && comparison < 0) {
			best = candidate
			bestRat = candidateRat
		}
	}
	return best, nil
}

func checkedIntPow(base, exponent int) (int, error) {
	result := 1
	factor := base
	for exponent > 0 {
		if exponent&1 == 1 {
			product, err := checkedIntOp(result, factor, "multiplication")
			if err != nil {
				return 0, fmt.Errorf("integer overflow during pow")
			}
			result = product.(int)
		}
		exponent >>= 1
		if exponent == 0 {
			break
		}
		square, err := checkedIntOp(factor, factor, "multiplication")
		if err != nil {
			return 0, fmt.Errorf("integer overflow during pow")
		}
		factor = square.(int)
	}
	return result, nil
}

// callBuiltinSum implements the sum(array) function.
func callBuiltinSum(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "sum requires exactly 1 argument: array", e); err != nil {
		return nil, err
	}

	if err := assertArg(args[0], beArray(), 1, "sum", e); err != nil {
		return nil, err
	}
	arr, ok := args[0].(Array)
	if !ok {
		return nil, newPosError(fmt.Sprintf("sum: expected array, got %T", args[0]), e.Pos())
	}

	return aggregateNumbers(arr, "sum", false, e)
}

// callBuiltinAvg implements the avg(array) function.
// Returns the average of all numbers in the array.
func callBuiltinAvg(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "avg requires exactly 1 argument: array", e); err != nil {
		return nil, err
	}

	if err := assertArg(args[0], beArray(), 1, "avg", e); err != nil {
		return nil, err
	}
	arr, ok := args[0].(Array)
	if !ok {
		return nil, newPosError(fmt.Sprintf("avg: expected array, got %T", args[0]), e.Pos())
	}

	if len(arr) == 0 {
		return nil, newPosError("avg: cannot calculate average of empty array", e.Pos())
	}

	return aggregateNumbers(arr, "avg", true, e)
}

func aggregateNumbers(arr Array, funcName string, average bool, e *ast.CallExpr) (Value, error) {
	total := new(big.Rat)
	hasFloat := false
	hasInexactInteger := false
	for i, item := range arr {
		number, ok := exactNumericRat(item)
		if !ok {
			return nil, newElementNotNumberError(funcName, i, item, e.Pos())
		}
		total.Add(total, number)
		switch value := item.(type) {
		case float64:
			hasFloat = true
		case int:
			hasInexactInteger = hasInexactInteger || !intExactlyRepresentableAsFloat(value)
		}
	}
	if average {
		total.Quo(total, new(big.Rat).SetInt64(int64(len(arr))))
	}

	if total.IsInt() && (!hasFloat || hasInexactInteger) {
		if total.Num().IsInt64() {
			value := total.Num().Int64()
			if int64(int(value)) == value {
				return int(value), nil
			}
		}
		return nil, newPosError(fmt.Sprintf("integer overflow during %s", funcName), e.Pos())
	}

	result, exact := total.Float64()
	if hasInexactInteger && !exact {
		return nil, newPosError(fmt.Sprintf("numeric precision loss during %s", funcName), e.Pos())
	}
	if err := validateFloat(result, funcName); err != nil {
		return nil, newPosError(err.Error(), e.Pos())
	}
	return result, nil
}

// callBuiltinIsDecimal implements the isDecimal(value) function.
func callBuiltinIsDecimal(args []Value, e *ast.CallExpr) (Value, error) {
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
func callBuiltinIsInteger(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "isInteger requires exactly 1 argument", e); err != nil {
		return nil, err
	}

	switch v := args[0].(type) {
	case float64:
		return !math.IsNaN(v) && !math.IsInf(v, 0) && math.Trunc(v) == v, nil
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
func callBuiltinIsEven(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "isEven requires exactly 1 argument", e); err != nil {
		return nil, err
	}

	switch v := args[0].(type) {
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) || math.Trunc(v) != v {
			return nil, newPosError(fmt.Sprintf("isEven expects an integer, got float %v", v), e.Pos())
		}
		return math.Mod(v, 2) == 0, nil
	case int:
		return v%2 == 0, nil
	case string:
		return nil, newPosError(fmt.Sprintf("isEven expects an integer, got string"), e.Pos())
	default:
		return nil, newPosError(fmt.Sprintf("isEven expects an integer, got %T", args[0]), e.Pos())
	}

}

// callBuiltinIsOdd implements the isOdd(value) function.
// Returns true if value is an odd integer.
func callBuiltinIsOdd(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "isOdd requires exactly 1 argument", e); err != nil {
		return nil, err
	}

	switch v := args[0].(type) {
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) || math.Trunc(v) != v {
			return nil, newPosError(fmt.Sprintf("isOdd expects an integer, got float %v", v), e.Pos())
		}
		return math.Mod(v, 2) != 0, nil
	case int:
		return v%2 != 0, nil
	case string:
		return nil, newPosError(fmt.Sprintf("isOdd expects an integer, got string"), e.Pos())
	default:
		return nil, newPosError(fmt.Sprintf("isOdd expects an integer, got %T", args[0]), e.Pos())
	}

}

// callBuiltinRandom implements the random() function.
func callBuiltinRandom(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireNoArgs(args, "random", e); err != nil {
		return nil, err
	}

	// Get a random number between 0 and 1
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("failed to generate random number: %s", err), e.Pos())
	}

	// Convert to a float between 0 and 1
	randomBits := uint64(0)
	for i := 0; i < 8; i++ {
		randomBits = (randomBits << 8) | uint64(randomBytes[i])
	}
	// Retain 53 random bits so every possible result is exactly representable
	// and strictly less than 1.
	randomFloat := float64(randomBits>>11) / float64(uint64(1)<<53)
	return randomFloat, nil
}

// callBuiltinRandomInt implements the randomInt(max) function.
// Returns a random integer from 0 to max (exclusive).
func callBuiltinRandomInt(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "randomInt requires exactly 1 argument: max", e); err != nil {
		return nil, err
	}

	max, err := truncatedRuntimeInt(args[0], "randomInt", "max", e)
	if err != nil {
		return nil, err
	}

	if max <= 0 {
		return nil, newPosError("randomInt max must be greater than 0", e.Pos())
	}

	randomValue, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return nil, newPosError(fmt.Sprintf("failed to generate random number: %s", err), e.Pos())
	}
	return int(randomValue.Int64()), nil
}

// callBuiltinTo implements the to(start, end) function.
func callBuiltinTo(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 2, "to requires exactly 2 arguments: start, end", e); err != nil {
		return nil, err
	}

	start, startRat, err := rangeBound(args[0], "start", e)
	if err != nil {
		return nil, err
	}
	_, endRat, err := rangeBound(args[1], "end", e)
	if err != nil {
		return nil, err
	}

	distance := new(big.Rat).Sub(endRat, startRat)
	distance.Abs(distance)
	steps := new(big.Int).Quo(distance.Num(), distance.Denom())
	length := new(big.Int).Add(steps, big.NewInt(1))
	if !length.IsInt64() || int64(int(length.Int64())) != length.Int64() {
		return nil, newPosError("to range length exceeds supported integer range", e.Pos())
	}

	rangeLength := int(length.Int64())
	result := make(Array, 0, rangeLength)
	ascending := startRat.Cmp(endRat) <= 0
	switch typedStart := start.(type) {
	case int:
		for offset := 0; offset < rangeLength; offset++ {
			if ascending {
				result = append(result, typedStart+offset)
			} else {
				result = append(result, typedStart-offset)
			}
		}
	case float64:
		for offset := 0; offset < rangeLength; offset++ {
			current := new(big.Rat).Set(startRat)
			if ascending {
				current.Add(current, new(big.Rat).SetInt64(int64(offset)))
			} else {
				current.Sub(current, new(big.Rat).SetInt64(int64(offset)))
			}
			value, _ := current.Float64()
			result = append(result, value)
		}
	}

	return result, nil
}

func rangeBound(value Value, argumentName string, e *ast.CallExpr) (Value, *big.Rat, error) {
	if number, ok := value.(float64); ok && (math.IsNaN(number) || math.IsInf(number, 0)) {
		return nil, nil, newPosError(fmt.Sprintf("to %s must be a finite number", argumentName), e.Pos())
	}

	number, ok := exactNumericRat(value)
	if !ok {
		return nil, nil, newPosError(fmt.Sprintf("to expects %s to be a number, got %T", argumentName, value), e.Pos())
	}
	minimum := new(big.Rat).SetInt64(int64(minInt()))
	maximum := new(big.Rat).SetInt64(int64(math.MaxInt))
	if number.Cmp(minimum) < 0 || number.Cmp(maximum) > 0 {
		return nil, nil, newPosError(fmt.Sprintf("to %s is outside the supported integer range", argumentName), e.Pos())
	}
	return value, number, nil
}

func truncatedRuntimeInt(value Value, funcName, argumentName string, e *ast.CallExpr) (int, error) {
	switch number := value.(type) {
	case int:
		return number, nil
	case float64:
		if math.IsNaN(number) || math.IsInf(number, 0) {
			return 0, newPosError(fmt.Sprintf("%s %s must be a finite number", funcName, argumentName), e.Pos())
		}
		truncated := math.Trunc(number)
		asRat := new(big.Rat).SetFloat64(truncated)
		if asRat == nil || !asRat.IsInt() || !asRat.Num().IsInt64() {
			return 0, newPosError(fmt.Sprintf("%s %s is outside the supported integer range", funcName, argumentName), e.Pos())
		}
		asInt64 := asRat.Num().Int64()
		if int64(int(asInt64)) != asInt64 {
			return 0, newPosError(fmt.Sprintf("%s %s is outside the supported integer range", funcName, argumentName), e.Pos())
		}
		return int(asInt64), nil
	default:
		return 0, newPosError(fmt.Sprintf("%s expects a number, got %T", funcName, value), e.Pos())
	}
}

// callBuiltinMod implements the mod(dividend, divisor) function.
func callBuiltinMod(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 2, "mod requires exactly 2 arguments: dividend, divisor", e); err != nil {
		return nil, err
	}

	if _, ok := exactNumericRat(args[0]); !ok {
		return nil, newPosError(fmt.Sprintf("mod expects dividend to be a number, got %T", args[0]), e.Pos())
	}
	if _, ok := exactNumericRat(args[1]); !ok {
		return nil, newPosError(fmt.Sprintf("mod expects divisor to be a number, got %T", args[1]), e.Pos())
	}
	divisor, _ := exactNumericRat(args[1])
	if divisor.Sign() == 0 {
		return nil, newPosError("mod: division by zero", e.Pos())
	}
	result, err := rem(args[0], args[1])
	if err != nil {
		return nil, newPosError(err.Error(), e.Pos())
	}
	return result, nil
}
