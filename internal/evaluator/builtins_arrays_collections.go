package evaluator

import (
	"fmt"
	"go/ast"
)

// callBuiltinSafeAccess implements the safeAccess(obj, key) function for optional field access
func callBuiltinSafeAccess(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 2 {
		return nil, newPosError("safeAccess requires exactly 2 arguments: object and key", e.Pos())
	}

	obj := args[0]
	key := args[1]

	if obj == nil {
		return nil, nil
	}

	// Try to index into obj with key
	switch v := obj.(type) {
	case Object:
		if strKey, ok := key.(string); ok {
			return v[strKey], nil
		}
	case Array:
		if intKey, ok := key.(int); ok {
			if intKey < 0 {
				intKey += len(v)
			}
			if intKey >= 0 && intKey < len(v) {
				return v[intKey], nil
			}
		} else if floatKey, ok := key.(float64); ok {
			intKey := int(floatKey)
			if intKey < 0 {
				intKey += len(v)
			}
			if intKey >= 0 && intKey < len(v) {
				return v[intKey], nil
			}
		}
	}

	// If cannot access, return nil
	return nil, nil
}

// callBuiltinSlice implements the slice(arrayOrString, start, end) function for slicing arrays or strings
func callBuiltinSlice(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 3 {
		return nil, newPosError("slice requires exactly 3 arguments: array or string, start, end", e.Pos())
	}

	var start int
	if startInt, ok := args[1].(int); ok {
		start = startInt
	} else if startFloat, ok := args[1].(float64); ok {
		start = int(startFloat)
	} else {
		return nil, newPosError("slice start must be a number", e.Args[1].Pos())
	}

	var end int
	if endInt, ok := args[2].(int); ok {
		end = endInt
	} else if endFloat, ok := args[2].(float64); ok {
		end = int(endFloat)
	} else {
		return nil, newPosError("slice end must be a number", e.Args[2].Pos())
	}

	if arrayVal, ok := args[0].(Array); ok {
		if start < 0 {
			start = len(arrayVal) + start
		}
		if end < 0 {
			end = len(arrayVal) + end
		}
		if start < 0 {
			start = 0
		}
		if end > len(arrayVal) {
			end = len(arrayVal)
		}
		if start > end {
			return Array{}, nil
		}
		return arrayVal[start:end], nil
	} else if strVal, ok := args[0].(string); ok {
		runes := []rune(strVal)
		if start < 0 {
			start = len(runes) + start
		}
		if end < 0 {
			end = len(runes) + end
		}
		if start < 0 {
			start = 0
		}
		if end > len(runes) {
			end = len(runes)
		}
		if start > end {
			return "", nil
		}
		return string(runes[start:end]), nil
	}

	return nil, newPosError("slice first argument must be an array or string", e.Args[0].Pos())
}

// callBuiltinRangeIndex implements DataWeave's inclusive start-to-end selector.
// Negative bounds are resolved relative to the collection, and descending
// bounds return values in reverse index order.
func callBuiltinRangeIndex(args []Value, e *ast.CallExpr) (Value, error) {
	start, ok := numericIndex(args[1])
	if !ok {
		return nil, newPosError("range index start must be a number", e.Args[1].Pos())
	}
	end, ok := numericIndex(args[2])
	if !ok {
		return nil, newPosError("range index end must be a number", e.Args[2].Pos())
	}

	switch value := args[0].(type) {
	case Array:
		start, end, ok := inclusiveRangeBounds(start, end, len(value))
		if !ok {
			return Array{}, nil
		}
		if start > end {
			result := make(Array, 0, start-end+1)
			for index := start; index >= end; index-- {
				result = append(result, value[index])
			}
			return result, nil
		}
		return value[start : end+1], nil
	case string:
		runes := []rune(value)
		start, end, ok := inclusiveRangeBounds(start, end, len(runes))
		if !ok {
			return "", nil
		}
		if start > end {
			result := make([]rune, 0, start-end+1)
			for index := start; index >= end; index-- {
				result = append(result, runes[index])
			}
			return string(result), nil
		}
		return string(runes[start : end+1]), nil
	default:
		return nil, newPosError("range index first argument must be an array or string", e.Args[0].Pos())
	}
}

func numericIndex(value Value) (int, bool) {
	switch index := value.(type) {
	case int:
		return index, true
	case float64:
		return int(index), true
	default:
		return 0, false
	}
}

func inclusiveRangeBounds(start, end, length int) (normalizedStart, normalizedEnd int, ok bool) {
	if length == 0 {
		return 0, 0, false
	}
	if start < 0 {
		start += length
	}
	if end < 0 {
		end += length
	}
	if start < 0 {
		start = 0
	}
	if start > length {
		start = length
	}
	if end < -1 {
		end = -1
	}
	if end >= length {
		end = length - 1
	}
	if start == length || end == -1 {
		return 0, 0, false
	}
	return start, end, true
}

// callBuiltinMapInternal is a helper for internal map operations
func callBuiltinMapInternal(array Array, lambda *Lambda, scope *Scope, depth int) (Value, error) {
	result := make(Array, 0, len(array))
	err := executeLambdaOnArrayElements(array, lambda, scope, depth, func(_ Value, _ int, mappedVal Value) error {
		result = append(result, mappedVal)
		return nil
	})
	return result, err
}

// callBuiltinPrepend implements the prepend(array, value) function.
func callBuiltinPrepend(args []Value, e *ast.CallExpr) (Value, error) {
	if err := validateArgCount(args, 2, "prepend", e.Pos()); err != nil {
		return nil, err
	}

	if err := assertArg(args[0], beArray(), 1, "prepend", e); err != nil {
		return nil, err
	}
	arr, ok := args[0].(Array)
	if !ok {
		return nil, newPosError(fmt.Sprintf("prepend: expected array, got %T", args[0]), e.Pos())
	}

	result := make(Array, 0, len(arr)+1)
	result = append(result, args[1])
	result = append(result, arr...)
	return result, nil
}

// callBuiltinAppend implements the append(array, value) function.
func callBuiltinAppend(args []Value, e *ast.CallExpr) (Value, error) {
	if err := validateArgCount(args, 2, "append", e.Pos()); err != nil {
		return nil, err
	}

	if err := assertArg(args[0], beArray(), 1, "append", e); err != nil {
		return nil, err
	}
	arr, ok := args[0].(Array)
	if !ok {
		return nil, newPosError(fmt.Sprintf("append: expected array, got %T", args[0]), e.Pos())
	}

	result := make(Array, 0, len(arr)+1)
	result = append(result, arr...)
	result = append(result, args[1])
	return result, nil
}

// callBuiltinFirst implements the first(array) function.
func callBuiltinFirst(args []Value, e *ast.CallExpr) (Value, error) {
	if err := validateArgCount(args, 1, "first", e.Pos()); err != nil {
		return nil, err
	}

	if err := assertArg(args[0], beArray(), 1, "first", e); err != nil {
		return nil, err
	}
	arr, ok := args[0].(Array)
	if !ok {
		return nil, newPosError(fmt.Sprintf("first: expected array, got %T", args[0]), e.Pos())
	}

	if len(arr) == 0 {
		return nil, nil
	}
	return arr[0], nil
}

// callBuiltinLast implements the last(array) function.
func callBuiltinLast(args []Value, e *ast.CallExpr) (Value, error) {
	if err := validateArgCount(args, 1, "last", e.Pos()); err != nil {
		return nil, err
	}

	if err := assertArg(args[0], beArray(), 1, "last", e); err != nil {
		return nil, err
	}
	arr, ok := args[0].(Array)
	if !ok {
		return nil, newPosError(fmt.Sprintf("last: expected array, got %T", args[0]), e.Pos())
	}

	if len(arr) == 0 {
		return nil, nil
	}
	return arr[len(arr)-1], nil
}

// callBuiltinTake implements the take(array, count) function.
// Returns the first 'count' elements of the array.
func callBuiltinTake(args []Value, e *ast.CallExpr) (Value, error) {
	if err := validateArgCount(args, 2, "take", e.Pos()); err != nil {
		return nil, err
	}

	if err := assertArg(args[0], beArray(), 1, "take", e); err != nil {
		return nil, err
	}
	arr, ok := args[0].(Array)
	if !ok {
		return nil, newPosError(fmt.Sprintf("take: expected array, got %T", args[0]), e.Pos())
	}

	count, err := toInt(args[1], "take", e)
	if err != nil {
		return nil, err
	}

	if count <= 0 {
		return Array{}, nil
	}
	if count >= len(arr) {
		return arr, nil
	}
	return arr[:count], nil
}

// callBuiltinDrop implements the drop(array, count) function.
// Returns the array with the first 'count' elements removed.
func callBuiltinDrop(args []Value, e *ast.CallExpr) (Value, error) {
	if err := validateArgCount(args, 2, "drop", e.Pos()); err != nil {
		return nil, err
	}

	if err := assertArg(args[0], beArray(), 1, "drop", e); err != nil {
		return nil, err
	}
	arr, ok := args[0].(Array)
	if !ok {
		return nil, newPosError(fmt.Sprintf("drop: expected array, got %T", args[0]), e.Pos())
	}

	count, err := toInt(args[1], "drop", e)
	if err != nil {
		return nil, err
	}

	if count <= 0 {
		return arr, nil
	}
	if count >= len(arr) {
		return Array{}, nil
	}
	return arr[count:], nil
}
