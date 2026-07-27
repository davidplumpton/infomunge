package evaluator

import (
	"fmt"
	"go/ast"
	"go/token"
	"infomunge/pkg/values"
)

// extractMapObjectResult extracts key-value pairs from the lambda result.
// Returns a map of entries and a bool indicating if the entry should be skipped.
func extractMapObjectResult(mapResult Value, pos token.Pos) (Object, bool, error) {
	switch result := mapResult.(type) {
	case Array:
		if len(result) != 2 {
			return nil, false, newPosError("mapObject lambda returning array must have exactly 2 elements [key, value]", pos)
		}
		key, ok := result[0].(string)
		if !ok {
			return nil, false, newPosError(fmt.Sprintf("mapObject lambda key must be a string, got %T", result[0]), pos)
		}
		entry := values.NewObject(1)
		values.SetObjectValue(entry, key, result[1])
		return entry, false, nil

	case Object:
		if len(result) == 0 {
			return nil, true, nil // Skip empty objects
		}
		return result, false, nil

	default:
		return nil, false, newPosError(fmt.Sprintf("mapObject lambda must return an array [key, value] or an object, got %T", mapResult), pos)
	}
}

func evalMapObjectInputs(e *ast.CallExpr, scope *Scope, depth int) (Object, *Lambda, bool, error) {
	if len(e.Args) != 2 {
		return nil, nil, false, newPosError("mapObject requires exactly 2 arguments: object, lambda", e.Pos())
	}

	objVal, nullHandled, err := evalCollectionSource(e, scope, depth, propagateNullCollectionSource)
	if err != nil {
		return nil, nil, false, err
	}
	if nullHandled {
		return nil, nil, true, nil
	}
	obj, ok := objVal.(Object)
	if !ok {
		return nil, nil, false, newPosError(fmt.Sprintf("mapObject expects an object, got %T", objVal), e.Args[0].Pos())
	}

	lambdaVal, err := evalASTInScopeWithDepth(e.Args[1], scope, depth)
	if err != nil {
		return nil, nil, false, err
	}
	lambda, ok := lambdaVal.(*Lambda)
	if !ok {
		return nil, nil, false, newPosError(fmt.Sprintf("mapObject expects a lambda function, got %T", lambdaVal), e.Args[1].Pos())
	}
	if lambda.ParamCount() > 3 {
		return nil, nil, false, newPosError(fmt.Sprintf("mapObject lambda must have between 0 and 3 parameters, got %d", lambda.ParamCount()), e.Args[1].Pos())
	}

	return obj, lambda, false, nil
}

func applyMapObject(obj Object, lambda *Lambda, scope *Scope, depth int, pos token.Pos) (Object, error) {
	dwOrder := !shouldUseLegacyObjectLambdaOrder(lambda)
	keys := sortedKeys(obj)
	result := values.NewObject(len(obj))
	index := 0

	for _, key := range keys {
		value := obj[key]

		if multi, ok := value.(XMLMultiValue); ok {
			for _, v := range multi {
				if err := applyAndMerge(v, key, index, result, lambda, scope, dwOrder, depth, pos); err != nil {
					return nil, err
				}
				index++
			}
		} else {
			if err := applyAndMerge(value, key, index, result, lambda, scope, dwOrder, depth, pos); err != nil {
				return nil, err
			}
			index++
		}
	}

	return result, nil
}

func applyAndMerge(value Value, key string, index int, result Object, lambda *Lambda, scope *Scope, dwOrder bool, depth int, pos token.Pos) error {
	mapResult, err := evalCollectionLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
		bindObjectLambdaParameters(lambdaContext, lambda, key, value, index, dwOrder)
	})
	if err != nil {
		return err
	}

	entries, skip, err := extractMapObjectResult(mapResult, pos)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}
	for _, entryKey := range values.ObjectKeys(entries) {
		entryValue := entries[entryKey]
		if existing, ok := result[entryKey]; ok {
			switch v := existing.(type) {
			case XMLMultiValue:
				values.SetObjectValue(result, entryKey, append(v, entryValue))
			default:
				values.SetObjectValue(result, entryKey, XMLMultiValue{existing, entryValue})
			}
		} else {
			values.SetObjectValue(result, entryKey, entryValue)
		}
	}
	return nil
}

// callBuiltinMapObject implements the __mapObject(object, lambda) function.
//
// Transforms each key-value pair in an object using a lambda function.
//
// Arguments:
//   - object: The object to transform
//   - lambda: A function with up to 3 parameters
//
// Lambda parameter order detection:
//   - Legacy InfoMunge style (key, value): only for params named "key"/"value" or "k"/"v"
//   - DataWeave style (value, key): all other parameter names
//
// Lambda return value:
//   - Should return an object with a single key-value pair: {"newKey": newValue}
//   - The returned key-value pairs are merged into the result object
//   - Return nil or empty object to skip the entry
//
// Example (DataWeave style):
//
//	{"a": 1, "b": 2} mapObject (value, key) -> {(upper(key)): value * 2}
//	// Returns: {"A": 2, "B": 4}
func callBuiltinMapObject(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	obj, lambda, nullHandled, err := evalMapObjectInputs(e, scope, depth)
	if err != nil {
		return nil, err
	}
	if nullHandled {
		return nil, nil
	}

	return applyMapObject(obj, lambda, scope, depth, e.Args[1].Pos())
}

// sortedKeys returns keys in stable object order.
func sortedKeys(m Object) []string {
	return values.ObjectKeys(m)
}
