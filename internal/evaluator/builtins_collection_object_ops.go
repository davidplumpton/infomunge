package evaluator

import (
	"fmt"
	"go/ast"
	"math"
	"strings"

	"infomunge/pkg/values"
)

// callBuiltinObjectToArray implements the objectToArray(object) function.
func callBuiltinObjectToArray(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 1 {
		return nil, newPosError("objectToArray requires exactly 1 argument: object", e.Pos())
	}

	if err := assertArg(args[0], beObject(), 1, "objectToArray", e); err != nil {
		return nil, err
	}
	obj, ok := args[0].(Object)
	if !ok {
		return nil, newPosError(fmt.Sprintf("objectToArray: expected object, got %T", args[0]), e.Pos())
	}

	keys := values.ObjectKeys(obj)

	result := make(Array, len(keys))
	for i, k := range keys {
		result[i] = newObjectEntry(k, obj[k])
	}

	return result, nil
}

// callBuiltinArrayToObject implements the arrayToObject(array) function.
func callBuiltinArrayToObject(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 1 {
		return nil, newPosError("arrayToObject requires exactly 1 argument: array", e.Pos())
	}

	if err := assertArg(args[0], beArray(), 1, "arrayToObject", e); err != nil {
		return nil, err
	}
	arr, ok := args[0].(Array)
	if !ok {
		return nil, newPosError(fmt.Sprintf("arrayToObject: expected array, got %T", args[0]), e.Pos())
	}

	result := values.NewObject(len(arr))
	for i, item := range arr {
		entry, ok := item.(Object)
		if !ok {
			return nil, newPosError(fmt.Sprintf("arrayToObject: expected object at index %d, got %T", i, item), e.Pos())
		}

		keyValue, ok := entry["key"]
		if !ok {
			return nil, newPosError(fmt.Sprintf("arrayToObject: missing key at index %d", i), e.Pos())
		}
		key, ok := keyValue.(string)
		if !ok {
			return nil, newPosError(fmt.Sprintf("arrayToObject: key at index %d must be string, got %T", i, keyValue), e.Pos())
		}

		value, ok := entry["value"]
		if !ok {
			return nil, newPosError(fmt.Sprintf("arrayToObject: missing value at index %d", i), e.Pos())
		}

		values.SetObjectValue(result, key, value)
	}

	return result, nil
}

// callBuiltinEntriesOf implements the entriesOf(object) function.
func callBuiltinEntriesOf(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 1 {
		return nil, newPosError("entriesOf requires exactly 1 argument: object", e.Pos())
	}

	if err := assertArg(args[0], beObject(), 1, "entriesOf", e); err != nil {
		return nil, err
	}
	obj, ok := args[0].(Object)
	if !ok {
		return nil, newPosError(fmt.Sprintf("entriesOf: expected object, got %T", args[0]), e.Pos())
	}

	uniqueKeys := values.ObjectKeys(obj)

	// Create result array of {key, value} pairs, expanding XMLMultiValue
	result := make(Array, 0)
	for _, k := range uniqueKeys {
		val := obj[k]
		if multi, ok := val.(XMLMultiValue); ok {
			for _, v := range multi {
				result = append(result, newObjectEntry(k, v))
			}
		} else {
			result = append(result, newObjectEntry(k, val))
		}
	}

	return result, nil
}

func newObjectEntry(key string, value Value) Object {
	entry := values.NewObject(2)
	values.SetObjectValue(entry, "key", key)
	values.SetObjectValue(entry, "value", value)
	return entry
}

// callBuiltinKeysOf implements the keysOf(object) function.
func callBuiltinKeysOf(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 1 {
		return nil, newPosError("keysOf requires exactly 1 argument: object", e.Pos())
	}

	if err := assertArg(args[0], beObject(), 1, "keysOf", e); err != nil {
		return nil, err
	}
	obj, ok := args[0].(Object)
	if !ok {
		return nil, newPosError(fmt.Sprintf("keysOf: expected object, got %T", args[0]), e.Pos())
	}

	uniqueKeys := values.ObjectKeys(obj)

	// Convert to interface array, expanding XMLMultiValue
	result := make(Array, 0)
	for _, k := range uniqueKeys {
		val := obj[k]
		if multi, ok := val.(XMLMultiValue); ok {
			for range multi {
				result = append(result, k)
			}
		} else {
			result = append(result, k)
		}
	}

	return result, nil
}

// callBuiltinValuesOf implements the valuesOf(object) function.
func callBuiltinValuesOf(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 1 {
		return nil, newPosError("valuesOf requires exactly 1 argument: object", e.Pos())
	}

	if err := assertArg(args[0], beObject(), 1, "valuesOf", e); err != nil {
		return nil, err
	}
	obj, ok := args[0].(Object)
	if !ok {
		return nil, newPosError(fmt.Sprintf("valuesOf: expected object, got %T", args[0]), e.Pos())
	}

	uniqueKeys := values.ObjectKeys(obj)

	// Create result array with values in key order, expanding XMLMultiValue
	result := make(Array, 0)
	for _, k := range uniqueKeys {
		val := obj[k]
		if multi, ok := val.(XMLMultiValue); ok {
			for _, v := range multi {
				result = append(result, v)
			}
		} else {
			result = append(result, val)
		}
	}

	return result, nil
}

// callBuiltinNamesOf implements the namesOf(object) function.
func callBuiltinNamesOf(args []Value, e *ast.CallExpr) (Value, error) {
	return callBuiltinKeysOf(args, e)
}

// callBuiltinZip implements the zip(array1, array2) function.
func callBuiltinZip(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 2 {
		return nil, newPosError("zip requires exactly 2 arguments: array1, array2", e.Pos())
	}

	arr1, ok := args[0].(Array)
	if !ok {
		return nil, newPosError(fmt.Sprintf("zip expects first argument to be an array, got %T", args[0]), e.Pos())
	}

	arr2, ok := args[1].(Array)
	if !ok {
		return nil, newPosError(fmt.Sprintf("zip expects second argument to be an array, got %T", args[1]), e.Pos())
	}

	// Use the length of the shorter array
	minLen := len(arr1)
	if len(arr2) < minLen {
		minLen = len(arr2)
	}

	result := make(Array, minLen)
	for i := 0; i < minLen; i++ {
		pair := Array{arr1[i], arr2[i]}
		result[i] = pair
	}

	return result, nil
}

// callBuiltinUnzip implements the unzip(array) function.
func callBuiltinUnzip(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 1 {
		return nil, newPosError("unzip requires exactly 1 argument: array", e.Pos())
	}

	if err := assertArg(args[0], beArray(), 1, "unzip", e); err != nil {
		return nil, err
	}
	arr, ok := args[0].(Array)
	if !ok {
		return nil, newPosError(fmt.Sprintf("unzip: expected array, got %T", args[0]), e.Pos())
	}

	arr1 := make(Array, len(arr))
	arr2 := make(Array, len(arr))

	for i, item := range arr {
		// Each item should be an array with 2 elements
		pair, ok := item.(Array)
		if !ok || len(pair) != 2 {
			return nil, newPosError("unzip expects array of pairs [key, value]", e.Pos())
		}
		arr1[i] = pair[0]
		arr2[i] = pair[1]
	}

	// Return as array of two arrays
	return Array{arr1, arr2}, nil
}

// callBuiltinRange implements the range(end) function.
func callBuiltinRange(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 1 {
		return nil, newPosError("range requires exactly 1 argument: end", e.Pos())
	}

	end, err := exactRangeEnd(args[0], e)
	if err != nil {
		return nil, err
	}
	if end < 0 {
		return nil, newPosError("range end must be non-negative", e.Pos())
	}

	result := make(Array, end)
	for i := 0; i < end; i++ {
		result[i] = i
	}

	return result, nil
}

func exactRangeEnd(value Value, e *ast.CallExpr) (int, error) {
	if number, ok := value.(float64); ok {
		if math.IsNaN(number) || math.IsInf(number, 0) {
			return 0, newPosError("range end must be a finite number", e.Pos())
		}
	}

	number, ok := exactNumericRat(value)
	if !ok {
		return 0, newPosError(fmt.Sprintf("range expects a number, got %T", value), e.Pos())
	}
	if !number.IsInt() {
		return 0, newPosError("range end causes numeric precision loss: expected an integer", e.Pos())
	}
	if !number.Num().IsInt64() {
		return 0, newPosError("range end is outside the supported integer range", e.Pos())
	}

	end64 := number.Num().Int64()
	end := int(end64)
	if int64(end) != end64 {
		return 0, newPosError("range end is outside the supported integer range", e.Pos())
	}
	return end, nil
}

// callBuiltinConcat implements the __concat(value1, value2, ...) function (++ operator).
func callBuiltinConcat(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) < 2 {
		return nil, newPosError("concat requires at least 2 arguments", e.Pos())
	}

	// Check if we're concatenating strings (string + non-string scalars allowed).
	hasString := false
	hasArray := false
	hasObject := false
	for _, arg := range args {
		switch arg.(type) {
		case string:
			hasString = true
		case Array, XMLMultiValue:
			hasArray = true
		case Object:
			hasObject = true
		}
	}

	if hasString {
		if hasArray || hasObject {
			return nil, newPosError("cannot concatenate mixed types", e.Pos())
		}
		var result strings.Builder
		for _, arg := range args {
			result.WriteString(coerceToString(arg))
		}
		return result.String(), nil
	}

	// Check if we're concatenating arrays
	allArrays := true
	arrays := make([]Array, 0, len(args))
	for _, arg := range args {
		arr, ok := AsArray(arg)
		if !ok {
			allArrays = false
			break
		}
		arrays = append(arrays, arr)
	}

	if allArrays {
		var totalLen int
		for _, arr := range arrays {
			totalLen += len(arr)
		}
		result := make(Array, 0, totalLen)
		for _, arr := range arrays {
			result = append(result, arr...)
		}
		return result, nil
	}

	// Check if we're concatenating objects
	allObjects := true
	for _, arg := range args {
		if _, ok := arg.(Object); !ok {
			allObjects = false
			break
		}
	}

	if allObjects {
		objects := make([]Object, 0, len(args))
		for _, arg := range args {
			obj, ok := arg.(Object)
			if !ok {
				return nil, newPosError("cannot concatenate mixed types", e.Pos())
			}
			objects = append(objects, obj)
		}
		return values.MergeObjects(objects...), nil
	}

	return nil, newPosError("cannot concatenate mixed types", e.Pos())
}

// callBuiltinRemove implements the __remove function (-- operator).
// Supports: array -- array (remove the right-hand values)
//
//	array -- value (remove a single value, InfoMunge extension)
//	object -- [keys] (remove keys from object)
func callBuiltinRemove(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 2 {
		return nil, newPosError("remove requires exactly 2 arguments", e.Pos())
	}

	// Case 1: object -- [keys] (remove multiple keys from object)
	if obj, ok := args[0].(Object); ok {
		keysToRemove, ok := AsArray(args[1])
		if !ok {
			return nil, newPosError(fmt.Sprintf("-- operator with object expects array of keys as second argument, got %T", args[1]), e.Pos())
		}

		// Build a set of keys to remove
		removeSet := make(map[string]bool, len(keysToRemove))
		for _, key := range keysToRemove {
			if keyStr, ok := key.(string); ok {
				removeSet[keyStr] = true
			}
		}

		result := values.CloneObject(obj)
		for key := range removeSet {
			delete(result, key)
		}
		return result, nil
	}

	// Case 2: array -- array/value
	array, ok := AsArray(args[0])
	if !ok {
		return nil, newPosError(fmt.Sprintf("-- operator expects first argument to be an array or object, got %T", args[0]), e.Pos())
	}

	valuesToRemove, removeMany := AsArray(args[1])
	if !removeMany {
		valuesToRemove = Array{args[1]}
	}
	result := make(Array, 0, len(array))

	for _, item := range array {
		remove := false
		for _, valueToRemove := range valuesToRemove {
			if isEqual(item, valueToRemove) {
				remove = true
				break
			}
		}
		if !remove {
			result = append(result, item)
		}
	}

	return result, nil
}

// callBuiltinWith implements the with(value, replacement) function.
func callBuiltinWith(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 2 {
		return nil, newPosError("with requires exactly 2 arguments: value and replacement", e.Pos())
	}

	// Simply return the replacement value
	return args[1], nil
}

// callBuiltinXsiType implements the xsiType(typeName) function.
func callBuiltinXsiType(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 1 {
		return nil, newPosError("xsiType requires exactly 1 argument: typeName", e.Pos())
	}

	typeName, ok := args[0].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("xsiType expects typeName to be a string, got %T", args[0]), e.Pos())
	}

	// Return an object with xsi:type property
	result := values.NewObject(1)
	values.SetObjectValue(result, "xsi:type", typeName)
	return result, nil
}
