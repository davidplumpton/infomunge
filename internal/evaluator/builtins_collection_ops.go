package evaluator

import (
	"fmt"
	"go/ast"
	"infomunge/pkg/values"
	"math/big"
	"sort"
	"strings"
	"unicode/utf8"
)

// executeLambdaOnArrayElements iterates over every array element, evaluates a lambda,
// and calls a callback with the element, index, and evaluated value.
type lambdaElementCallback func(elem Value, index int, value Value) error

func executeLambdaOnArrayElements(
	array Array,
	lambda *Lambda,
	scope *Scope,
	depth int,
	callback lambdaElementCallback,
) error {
	return executeLambdaOnArrayElementsUntil(array, lambda, scope, depth, func(elem Value, index int, value Value) (bool, error) {
		return false, callback(elem, index, value)
	})
}

// executeLambdaOnArrayElementsUntil is the short-circuiting variant of
// executeLambdaOnArrayElements. The callback's boolean result stops iteration
// when true.
type lambdaElementUntilCallback func(elem Value, index int, value Value) (bool, error)

func bindArrayLambdaParameters(lambdaContext Context, lambda *Lambda, elem Value, index int) {
	if lambda.ParamCount() > 0 {
		lambdaContext[lambda.ParamName(0)] = elem
	}
	if lambda.ParamCount() > 1 {
		lambdaContext[lambda.ParamName(1)] = index
	}
}

func executeLambdaOnArrayElementsUntil(
	array Array,
	lambda *Lambda,
	scope *Scope,
	depth int,
	callback lambdaElementUntilCallback,
) error {
	for i, elem := range array {
		value, err := evalCollectionLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
			bindArrayLambdaParameters(lambdaContext, lambda, elem, i)
		})
		if err != nil {
			return err
		}

		stop, err := callback(elem, i, value)
		if err != nil {
			return err
		}
		if stop {
			return nil
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
		if len(result) == 0 {
			return nil, nil
		}
		return result, nil
	}

	obj, ok := AsObject(source)
	if !ok {
		return nil, newPosError(
			fmt.Sprintf("selector filter expects an array or object, got %T", source),
			e.Args[0].Pos(),
		)
	}

	keys := values.ObjectKeys(obj)

	result := values.NewObject(len(obj))
	for idx, key := range keys {
		val := obj[key]
		condVal, err := evalCollectionLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
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

	if len(result) == 0 {
		return nil, nil
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
			return utf8.RuneCountInString(v), nil
		case []byte:
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
	array Array,
	lambda *Lambda,
	predicate func(new, current Value) (bool, error),
	scope *Scope,
	depth int,
	e *ast.CallExpr,
) (Value, error) {
	if len(array) == 0 {
		return nil, nil
	}

	var extremumElement Value = array[0]
	var extremumValue Value

	for i, elem := range array {
		value, err := evalCollectionLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
			bindArrayLambdaParameters(lambdaContext, lambda, elem, i)
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
	array, lambda, err := evalArrayAndLambda("maxBy", e, scope, depth, 0, 2)
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

	return findExtremumByLambda(array, lambda, predicate, scope, depth, e)
}

// callBuiltinMinBy implements the __minBy(array, lambda) function.
func callBuiltinMinBy(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	array, lambda, err := evalArrayAndLambda("minBy", e, scope, depth, 0, 2)
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

	return findExtremumByLambda(array, lambda, predicate, scope, depth, e)
}

// callBuiltinOrderBy implements the __orderBy(array, lambda) function.
func callBuiltinOrderBy(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	array, lambda, nullHandled, err := evalNullPropagatingArrayAndLambda("orderBy", e, scope, depth, 0, 2)
	if err != nil {
		return nil, err
	}
	if nullHandled {
		return nil, nil
	}

	type elementWithKey struct {
		element Value
		key     Value
		index   int
	}

	elements := make([]elementWithKey, len(array))
	for i, elem := range array {
		key, err := evalCollectionLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
			bindArrayLambdaParameters(lambdaContext, lambda, elem, i)
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
	array, lambda, nullHandled, err := evalNullPropagatingArrayAndLambda("distinctBy", e, scope, depth, 0, 2)
	if err != nil {
		return nil, err
	}
	if nullHandled {
		return nil, nil
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
	array, lambda, nullHandled, err := evalNullPropagatingArrayAndLambda("some", e, scope, depth, 0, 2)
	if err != nil {
		return nil, err
	}
	if nullHandled {
		return false, nil
	}

	found := false
	err = executeLambdaOnArrayElementsUntil(array, lambda, scope, depth, func(_ Value, _ int, result Value) (bool, error) {
		boolResult, ok := result.(bool)
		if !ok {
			return false, newLambdaWrongReturnError("some", "a boolean", result, e.Args[1].Pos())
		}
		if boolResult {
			found = true
		}
		return boolResult, nil
	})
	if err != nil {
		return nil, err
	}
	return found, nil
}

// callBuiltinEvery implements the every(array, lambda) function.
func callBuiltinEvery(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	array, lambda, nullHandled, err := evalNullPropagatingArrayAndLambda("every", e, scope, depth, 0, 2)
	if err != nil {
		return nil, err
	}
	if nullHandled {
		return true, nil
	}

	allTrue := true
	err = executeLambdaOnArrayElementsUntil(array, lambda, scope, depth, func(_ Value, _ int, result Value) (bool, error) {
		boolResult, ok := result.(bool)
		if !ok {
			return false, newLambdaWrongReturnError("every", "a boolean", result, e.Args[1].Pos())
		}
		if !boolResult {
			allTrue = false
		}
		return !boolResult, nil
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
	objVal, nullHandled, err := evalCollectionSource(e, scope, depth, propagateNullCollectionSource)
	if err != nil {
		return nil, err
	}
	if nullHandled {
		return nil, nil
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

	if lambda.ParamCount() > 3 {
		return nil, newPosError(fmt.Sprintf("filterObject lambda must have between 0 and 3 parameters, got %d", lambda.ParamCount()), e.Args[1].Pos())
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
	condVal, err := evalCollectionLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
		bindObjectLambdaParameters(lambdaContext, lambda, key, value, index, isDataWeaveOrder)
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

// callBuiltinGroupBy implements the __groupBy(source, lambda) function.
func callBuiltinGroupBy(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("groupBy requires exactly 2 arguments: source, lambda", e.Pos())
	}

	source, nullHandled, err := evalCollectionSource(e, scope, depth, propagateNullCollectionSource)
	if err != nil {
		return nil, err
	}
	if nullHandled {
		return nil, nil
	}

	var (
		arraySource  Array
		stringSource string
		isString     bool
	)
	switch typed := source.(type) {
	case Array:
		arraySource = typed
	case string, int, float64, *big.Int, bool:
		stringSource = coerceToString(typed)
		isString = true
	case Object:
		return nil, newPosError("groupBy expects an array or string, got object", e.Args[0].Pos())
	default:
		return nil, newPosError(fmt.Sprintf("groupBy expects an array or string, got %T", source), e.Args[0].Pos())
	}

	lambda, err := evalCollectionLambda("groupBy", e, scope, depth, 0, 2)
	if err != nil {
		return nil, err
	}

	if isString {
		return groupStringBy(stringSource, lambda, scope, depth, e)
	}
	return groupArrayBy(arraySource, lambda, scope, depth, e)
}

func groupArrayBy(array Array, lambda *Lambda, scope *Scope, depth int, e *ast.CallExpr) (Value, error) {
	result := values.NewObject(0)

	err := executeLambdaOnArrayElements(array, lambda, scope, depth, func(elem Value, _ int, key Value) error {
		keyStr := coerceToString(key)

		if _, exists := result[keyStr]; !exists {
			values.SetObjectValue(result, keyStr, make(Array, 0))
		}

		groupVal, ok := result[keyStr].(Array)
		if !ok {
			return newPosError("groupBy: internal error - unexpected type for group", e.Fun.Pos())
		}
		values.SetObjectValue(result, keyStr, append(groupVal, elem))
		return nil
	})
	return result, err
}

func groupStringBy(source string, lambda *Lambda, scope *Scope, depth int, e *ast.CallExpr) (Value, error) {
	result := values.NewObject(0)

	for index, char := range []rune(source) {
		elem := string(char)
		key, err := evalCollectionLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
			bindArrayLambdaParameters(lambdaContext, lambda, elem, index)
		})
		if err != nil {
			return nil, err
		}

		keyStr := coerceToString(key)
		group, ok := result[keyStr]
		if !ok {
			values.SetObjectValue(result, keyStr, elem)
			continue
		}

		groupString, ok := group.(string)
		if !ok {
			return nil, newPosError("groupBy: internal error - unexpected type for group", e.Fun.Pos())
		}
		values.SetObjectValue(result, keyStr, groupString+elem)
	}

	return result, nil
}

// callBuiltinPluck implements the __pluck(source, selector) function.
func callBuiltinPluck(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("pluck requires exactly 2 arguments: source, selector", e.Pos())
	}

	sourceVal, nullHandled, err := evalCollectionSource(e, scope, depth, propagateNullCollectionSource)
	if err != nil {
		return nil, err
	}
	if nullHandled {
		return nil, nil
	}

	selectorVal, err := evalASTInScopeWithDepth(e.Args[1], scope, depth)
	if err != nil {
		return nil, err
	}

	// Case 1: Source is an Object
	if obj, ok := sourceVal.(Object); ok {
		if lambda, ok := selectorVal.(*Lambda); ok {
			if lambda.ParamCount() > 3 {
				return nil, newPosError(fmt.Sprintf("pluck lambda must have between 0 and 3 parameters, got %d", lambda.ParamCount()), e.Args[1].Pos())
			}
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
			if lambda.ParamCount() > 2 {
				return nil, newPosError(fmt.Sprintf("pluck lambda must have between 0 and 2 parameters for an array, got %d", lambda.ParamCount()), e.Args[1].Pos())
			}
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
	return evalCollectionLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
		bindObjectLambdaParameters(lambdaContext, lambda, key, value, index, isDataWeaveOrder)
	})
}

func bindObjectLambdaParameters(lambdaContext Context, lambda *Lambda, key string, value Value, index int, isDataWeaveOrder bool) {
	if lambda.ParamCount() == 0 {
		return
	}

	first, second := Value(value), Value(key)
	if !isDataWeaveOrder {
		first, second = key, value
	}
	lambdaContext[lambda.ParamName(0)] = first
	if lambda.ParamCount() > 1 {
		lambdaContext[lambda.ParamName(1)] = second
	}
	if lambda.ParamCount() > 2 {
		lambdaContext[lambda.ParamName(2)] = index
	}
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
