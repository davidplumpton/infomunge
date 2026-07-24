package evaluator

import (
	"fmt"
	"go/ast"
	"infomunge/pkg/values"
	"sort"
	"strings"
)

// executeLambdaOnArrayElements iterates over an array, evaluates a lambda on each element,
// and calls a callback with the element, index, and evaluated value.
// The callback receives (element, index, evaluatedValue) and should return an error or nil.
type lambdaElementCallback func(elem Value, index int, value Value) error

func executeLambdaOnArrayElements(
	array Array,
	lambda *Lambda,
	scope *Scope,
	depth int,
	callback lambdaElementCallback,
) error {
	for i, elem := range array {
		value, err := evalLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
			lambdaContext[lambda.ParamName(0)] = elem
			if lambda.ParamCount() > 1 {
				lambdaContext[lambda.ParamName(1)] = i
			}
		})
		if err != nil {
			return err
		}

		if err := callback(elem, i, value); err != nil {
			return err
		}
	}
	return nil
}

// callBuiltinFilterSelector implements DataWeave-style selector filter: source[?(expr)].
func callBuiltinFilterSelector(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("selector filter requires exactly 2 arguments: source, lambda", e.Pos())
	}

	source, err := evalASTInScopeWithDepth(e.Args[0], scope, depth)
	if err != nil {
		return nil, err
	}
	lambdaVal, err := evalASTInScopeWithDepth(e.Args[1], scope, depth)
	if err != nil {
		return nil, err
	}
	lambda, ok := lambdaVal.(*Lambda)
	if !ok {
		return nil, newPosError(fmt.Sprintf("selector filter expects a lambda function, got %T", lambdaVal), e.Args[1].Pos())
	}
	if lambda.ParamCount() < 1 || lambda.ParamCount() > 2 {
		return nil, newPosError(fmt.Sprintf("selector filter lambda must have 1 or 2 parameters, got %d", lambda.ParamCount()), e.Args[1].Pos())
	}

	if arr, ok := AsArray(source); ok {
		result := make(Array, 0, len(arr))
		err := executeLambdaOnArrayElements(arr, lambda, scope, depth, func(elem Value, _ int, cond Value) error {
			condBool, ok := cond.(bool)
			if !ok {
				return newPosError(fmt.Sprintf("selector filter lambda must return a boolean, got %T", cond), e.Args[1].Pos())
			}
			if condBool {
				result = append(result, elem)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	obj, ok := AsObject(source)
	if !ok {
		if source == nil {
			return nil, nil
		}
		return nil, nil
	}

	keys := values.ObjectKeys(obj)

	result := values.NewObject(len(obj))
	for idx, key := range keys {
		val := obj[key]
		condVal, err := evalLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
			lambdaContext[lambda.ParamName(0)] = val
			if lambda.ParamCount() > 1 {
				lambdaContext[lambda.ParamName(1)] = idx
			}
		})
		if err != nil {
			return nil, err
		}
		condBool, ok := condVal.(bool)
		if !ok {
			return nil, newPosError(fmt.Sprintf("selector filter lambda must return a boolean, got %T", condVal), e.Args[1].Pos())
		}
		if condBool {
			values.SetObjectValue(result, key, val)
		}
	}

	return result, nil
}

// callBuiltinMetadata implements DataWeave-style metadata selector: value.^meta.
func callBuiltinMetadata(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 2 {
		return nil, newPosError("metadata selector requires exactly 2 arguments: value, metadata", e.Pos())
	}

	metaName, ok := args[1].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("metadata selector name must be a string, got %T", args[1]), e.Pos())
	}

	value := args[0]
	switch metaName {
	case "class", "type":
		return getTypeName(value), nil
	case "size":
		switch v := value.(type) {
		case string:
			return len(v), nil
		case Array:
			return len(v), nil
		case XMLMultiValue:
			return len(v), nil
		case Object:
			return len(v), nil
		default:
			return nil, nil
		}
	case "attributes":
		obj, ok := value.(Object)
		if !ok {
			return nil, nil
		}
		attrs := values.NewObject(0)
		for _, k := range values.ObjectKeys(obj) {
			if strings.HasPrefix(k, "@") && k != "@xmlns" {
				values.SetObjectValue(attrs, k, obj[k])
			}
		}
		return attrs, nil
	case "namespaces":
		obj, ok := value.(Object)
		if !ok {
			return nil, nil
		}
		return obj["@xmlns"], nil
	case "text":
		obj, ok := value.(Object)
		if !ok {
			return nil, nil
		}
		return obj["#text"], nil
	default:
		return nil, nil
	}
}

// findExtremumByLambda evaluates a lambda on each array element and returns the element
// that satisfies the comparison predicate. The predicate receives (newValue, currentValue)
// and returns true if newValue should replace currentValue.
func findExtremumByLambda(
	funcName string,
	array Array,
	lambda *Lambda,
	predicate func(new, current Value) (bool, error),
	scope *Scope,
	depth int,
	e *ast.CallExpr,
) (Value, error) {
	if len(array) == 0 {
		return nil, newPosError(fmt.Sprintf("%s: cannot find extremum of empty array", funcName), e.Args[0].Pos())
	}

	var extremumElement Value = array[0]
	var extremumValue Value

	for i, elem := range array {
		value, err := evalLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
			lambdaContext[lambda.ParamName(0)] = elem
			if lambda.ParamCount() > 1 {
				lambdaContext[lambda.ParamName(1)] = i
			}
		})
		if err != nil {
			return nil, err
		}

		if i == 0 {
			extremumValue = value
		} else {
			// Use the predicate to determine if this value should replace the current extremum
			shouldUpdate, err := predicate(value, extremumValue)
			if err != nil {
				return nil, newPosError(err.Error(), e.Pos())
			}
			if shouldUpdate {
				extremumElement = elem
				extremumValue = value
			}
		}
	}

	return extremumElement, nil
}

// callBuiltinMaxBy implements the __maxBy(array, lambda) function.
func callBuiltinMaxBy(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	array, lambda, err := evalArrayAndLambda("maxBy", e, scope, depth, 1, 2)
	if err != nil {
		return nil, err
	}

	// Predicate: newValue should replace currentValue if it's greater
	predicate := func(newValue, currentValue Value) (bool, error) {
		cmp, err := compareValues(newValue, currentValue)
		if err != nil {
			return false, err
		}
		return cmp > 0, nil
	}

	return findExtremumByLambda("maxBy", array, lambda, predicate, scope, depth, e)
}

// callBuiltinMinBy implements the __minBy(array, lambda) function.
func callBuiltinMinBy(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	array, lambda, err := evalArrayAndLambda("minBy", e, scope, depth, 1, 2)
	if err != nil {
		return nil, err
	}

	// Predicate: newValue should replace currentValue if it's less
	predicate := func(newValue, currentValue Value) (bool, error) {
		cmp, err := compareValues(newValue, currentValue)
		if err != nil {
			return false, err
		}
		return cmp < 0, nil
	}

	return findExtremumByLambda("minBy", array, lambda, predicate, scope, depth, e)
}

// callBuiltinOrderBy implements the __orderBy(array, lambda) function.
func callBuiltinOrderBy(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	array, lambda, err := evalArrayAndLambda("orderBy", e, scope, depth, 1, 2)
	if err != nil {
		return nil, err
	}

	type elementWithKey struct {
		element Value
		key     Value
		index   int
	}

	elements := make([]elementWithKey, len(array))
	for i, elem := range array {
		key, err := evalLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
			lambdaContext[lambda.ParamName(0)] = elem
			if lambda.ParamCount() > 1 {
				lambdaContext[lambda.ParamName(1)] = i
			}
		})
		if err != nil {
			return nil, err
		}

		elements[i] = elementWithKey{
			element: elem,
			key:     key,
			index:   i,
		}
	}

	// Sort the elements by their keys. sort callbacks cannot return errors, so
	// retain the first comparison failure and report it after sorting stops.
	var comparisonErr error
	sort.SliceStable(elements, func(i, j int) bool {
		if comparisonErr != nil {
			return false
		}
		cmp, err := compareValues(elements[i].key, elements[j].key)
		if err != nil {
			comparisonErr = fmt.Errorf(
				"cannot compare keys at indexes %d and %d: %w",
				elements[i].index,
				elements[j].index,
				err,
			)
			return false
		}
		return cmp < 0
	})
	if comparisonErr != nil {
		return nil, &posError{
			msg: fmt.Sprintf("orderBy: %s", comparisonErr),
			pos: e.Args[0].Pos(),
		}
	}

	// Extract sorted elements
	result := make(Array, len(elements))
	for i, ek := range elements {
		result[i] = ek.element
	}

	return result, nil
}

// callBuiltinDistinctBy implements the __distinctBy(array, lambda) function.
func callBuiltinDistinctBy(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	array, lambda, err := evalArrayAndLambda("distinctBy", e, scope, depth, 1, 2)
	if err != nil {
		return nil, err
	}

	seenValues := make(Array, 0, len(array))
	result := make(Array, 0, len(array))

	err = executeLambdaOnArrayElements(array, lambda, scope, depth, func(elem Value, _ int, key Value) error {
		if !containsEqualValue(seenValues, key) {
			seenValues = append(seenValues, key)
			result = append(result, elem)
		}
		return nil
	})
	return result, err
}

// callBuiltinSome implements the some(array, lambda) function.
func callBuiltinSome(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	array, lambda, err := evalArrayAndLambda("some", e, scope, depth, 1, 2)
	if err != nil {
		return nil, err
	}

	found := false
	err = executeLambdaOnArrayElements(array, lambda, scope, depth, func(_ Value, _ int, result Value) error {
		boolResult, ok := result.(bool)
		if !ok {
			return newLambdaWrongReturnError("some", "a boolean", result, e.Args[1].Pos())
		}
		if boolResult {
			found = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return found, nil
}

// callBuiltinEvery implements the every(array, lambda) function.
func callBuiltinEvery(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	array, lambda, err := evalArrayAndLambda("every", e, scope, depth, 1, 2)
	if err != nil {
		return nil, err
	}

	allTrue := true
	err = executeLambdaOnArrayElements(array, lambda, scope, depth, func(_ Value, _ int, result Value) error {
		boolResult, ok := result.(bool)
		if !ok {
			return newLambdaWrongReturnError("every", "a boolean", result, e.Args[1].Pos())
		}
		if !boolResult {
			allTrue = false
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return allTrue, nil
}

// callBuiltinFilterObject implements the __filterObject(object, lambda) function.
func callBuiltinFilterObject(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("filterObject requires exactly 2 arguments: object, lambda", e.Pos())
	}

	// Evaluate the object argument
	objVal, err := evalASTInScopeWithDepth(e.Args[0], scope, depth)
	if err != nil {
		return nil, err
	}

	// Check that the first argument is an object
	obj, ok := objVal.(Object)
	if !ok {
		return nil, newPosError(fmt.Sprintf("filterObject expects an object, got %T", objVal), e.Args[0].Pos())
	}

	// Evaluate the lambda argument to get the lambda function
	lambdaVal, err := evalASTInScopeWithDepth(e.Args[1], scope, depth)
	if err != nil {
		return nil, err
	}

	lambda, ok := lambdaVal.(*Lambda)
	if !ok {
		return nil, newPosError(fmt.Sprintf("filterObject expects a lambda function, got %T", lambdaVal), e.Args[1].Pos())
	}

	// The lambda should have 2 parameters (key, value) or 3 (key, value, index)
	if lambda.ParamCount() < 2 || lambda.ParamCount() > 3 {
		return nil, newPosError(fmt.Sprintf("filterObject lambda must have 2 or 3 parameters, got %d", lambda.ParamCount()), e.Args[1].Pos())
	}

	result := values.NewObject(len(obj))
	index := 0

	// Default to DataWeave order (value, key) unless lambda explicitly uses
	// legacy names (key, value) or (k, v).
	isDataWeaveOrder := !shouldUseLegacyObjectLambdaOrder(lambda)

	keys := values.ObjectKeys(obj)

	for _, key := range keys {
		value := obj[key]
		if multi, ok := value.(XMLMultiValue); ok {
			for _, v := range multi {
				passed, err := evaluateFilterObjectCond(key, v, index, isDataWeaveOrder, lambda, scope, depth, e)
				if err != nil {
					return nil, err
				}
				if passed {
					mergeIntoResult(result, key, v)
				}
				index++
			}
		} else {
			passed, err := evaluateFilterObjectCond(key, value, index, isDataWeaveOrder, lambda, scope, depth, e)
			if err != nil {
				return nil, err
			}
			if passed {
				mergeIntoResult(result, key, value)
			}
			index++
		}
	}

	return result, nil
}

func evaluateFilterObjectCond(key string, value Value, index int, isDataWeaveOrder bool, lambda *Lambda, scope *Scope, depth int, e *ast.CallExpr) (bool, error) {
	condVal, err := evalLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
		if isDataWeaveOrder {
			lambdaContext[lambda.ParamName(0)] = value
			lambdaContext[lambda.ParamName(1)] = key
			if lambda.ParamCount() > 2 {
				lambdaContext[lambda.ParamName(2)] = index
			}
			return
		}
		lambdaContext[lambda.ParamName(0)] = key
		lambdaContext[lambda.ParamName(1)] = value
		if lambda.ParamCount() > 2 {
			lambdaContext[lambda.ParamName(2)] = index
		}
	})
	if err != nil {
		return false, err
	}

	// Convert condition to boolean
	condBool, ok := condVal.(bool)
	if !ok {
		return false, newPosError(fmt.Sprintf("filterObject lambda must return a boolean, got %T", condVal), e.Args[1].Pos())
	}
	return condBool, nil
}

func mergeIntoResult(result Object, key string, value Value) {
	if existing, ok := result[key]; ok {
		switch v := existing.(type) {
		case XMLMultiValue:
			values.SetObjectValue(result, key, append(v, value))
		default:
			values.SetObjectValue(result, key, XMLMultiValue{existing, value})
		}
	} else {
		values.SetObjectValue(result, key, value)
	}
}

// callBuiltinGroupBy implements the __groupBy(array, lambda) function.
func callBuiltinGroupBy(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	array, lambda, err := evalArrayAndLambda("groupBy", e, scope, depth, 1, 2)
	if err != nil {
		return nil, err
	}

	result := values.NewObject(0)

	err = executeLambdaOnArrayElements(array, lambda, scope, depth, func(elem Value, _ int, key Value) error {
		// Convert key to string for use as object key
		keyStr := fmt.Sprintf("%v", key)

		// Get or create the group for this key
		if _, exists := result[keyStr]; !exists {
			values.SetObjectValue(result, keyStr, make(Array, 0))
		}

		// Append element to the group with proper type checking
		groupVal, ok := result[keyStr].(Array)
		if !ok {
			return newPosError("groupBy: internal error - unexpected type for group", e.Fun.Pos())
		}
		values.SetObjectValue(result, keyStr, append(groupVal, elem))
		return nil
	})
	return result, err
}

// callBuiltinPluck implements the __pluck(source, selector) function.
func callBuiltinPluck(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("pluck requires exactly 2 arguments: source, selector", e.Pos())
	}

	sourceVal, err := evalASTInScopeWithDepth(e.Args[0], scope, depth)
	if err != nil {
		return nil, err
	}

	selectorVal, err := evalASTInScopeWithDepth(e.Args[1], scope, depth)
	if err != nil {
		return nil, err
	}

	// Case 1: Source is an Object
	if obj, ok := sourceVal.(Object); ok {
		if lambda, ok := selectorVal.(*Lambda); ok {
			// Object + Lambda (DataWeave style pluck)
			return pluckObject(obj, lambda, scope, depth, e)
		}
		return nil, newPosError(fmt.Sprintf("pluck on object expects a lambda function, got %T", selectorVal), e.Args[1].Pos())
	}

	// Case 2: Source is an Array
	if arr, ok := sourceVal.(Array); ok {
		if keyStr, ok := selectorVal.(string); ok {
			// Array + String (existing field extraction logic)
			return pluckArray(arr, keyStr, e)
		}
		if lambda, ok := selectorVal.(*Lambda); ok {
			// Pluck on array with lambda acts like map
			return callBuiltinMapInternal(arr, lambda, scope, depth)
		}
		return nil, newPosError(fmt.Sprintf("pluck on array expects a string or lambda function, got %T", selectorVal), e.Args[1].Pos())
	}

	return nil, newPosError(fmt.Sprintf("pluck expects an array or object, got %T", sourceVal), e.Args[0].Pos())
}

func pluckObject(obj Object, lambda *Lambda, scope *Scope, depth int, e *ast.CallExpr) (Value, error) {
	// Default to DataWeave order (value, key) unless lambda explicitly uses
	// legacy names (key, value) or (k, v).
	isDataWeaveOrder := !shouldUseLegacyObjectLambdaOrder(lambda)

	keys := values.ObjectKeys(obj)

	result := make(Array, 0, len(obj))
	index := 0
	for _, k := range keys {
		value := obj[k]
		if multi, ok := value.(XMLMultiValue); ok {
			for _, v := range multi {
				res, err := evaluatePluckLambda(k, v, index, isDataWeaveOrder, lambda, scope, depth)
				if err != nil {
					return nil, err
				}
				result = append(result, res)
				index++
			}
		} else {
			res, err := evaluatePluckLambda(k, value, index, isDataWeaveOrder, lambda, scope, depth)
			if err != nil {
				return nil, err
			}
			result = append(result, res)
			index++
		}
	}
	return result, nil
}

func evaluatePluckLambda(key string, value Value, index int, isDataWeaveOrder bool, lambda *Lambda, scope *Scope, depth int) (Value, error) {
	return evalLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
		if isDataWeaveOrder {
			lambdaContext[lambda.ParamName(0)] = value
			if lambda.ParamCount() > 1 {
				lambdaContext[lambda.ParamName(1)] = key
			}
			if lambda.ParamCount() > 2 {
				lambdaContext[lambda.ParamName(2)] = index
			}
			return
		}
		lambdaContext[lambda.ParamName(0)] = key
		if lambda.ParamCount() > 1 {
			lambdaContext[lambda.ParamName(1)] = value
		}
		if lambda.ParamCount() > 2 {
			lambdaContext[lambda.ParamName(2)] = index
		}
	})
}

// pluckArray extracts nested properties from array elements using dot notation
func pluckArray(arr Array, keyStr string, e *ast.CallExpr) (Value, error) {
	// Split the key for nested property access
	keyParts := strings.Split(keyStr, ".")

	result := make(Array, 0, len(arr))

	for _, item := range arr {
		// Navigate nested properties
		current := item
		for _, part := range keyParts {
			if current == nil {
				current = nil
				break
			}

			// Try to access as map
			if m, ok := current.(Object); ok {
				current = m[part]
			} else {
				// Property doesn't exist or current is not a map
				current = nil
				break
			}
		}

		result = append(result, current)
	}

	return result, nil
}
