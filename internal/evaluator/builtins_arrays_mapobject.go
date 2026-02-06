package evaluator

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
)

// extractMapObjectResult extracts key-value pairs from the lambda result.
// Returns a map of entries and a bool indicating if the entry should be skipped.
func extractMapObjectResult(mapResult interface{}, pos token.Pos) (map[string]interface{}, bool, error) {
	switch result := mapResult.(type) {
	case []interface{}:
		if len(result) != 2 {
			return nil, false, newPosError("mapObject lambda returning array must have exactly 2 elements [key, value]", pos)
		}
		key, ok := result[0].(string)
		if !ok {
			return nil, false, newPosError(fmt.Sprintf("mapObject lambda key must be a string, got %T", result[0]), pos)
		}
		return map[string]interface{}{key: result[1]}, false, nil

	case map[string]interface{}:
		if len(result) == 0 {
			return nil, true, nil // Skip empty objects
		}
		return result, false, nil

	default:
		return nil, false, newPosError(fmt.Sprintf("mapObject lambda must return an array [key, value] or an object, got %T", mapResult), pos)
	}
}

func evalMapObjectInputs(e *ast.CallExpr, context map[string]interface{}, depth int) (map[string]interface{}, *Lambda, error) {
	if len(e.Args) != 2 {
		return nil, nil, newPosError("mapObject requires exactly 2 arguments: object, lambda", e.Pos())
	}

	objVal, err := evalASTWithDepth(e.Args[0], context, depth)
	if err != nil {
		return nil, nil, err
	}
	obj, ok := objVal.(map[string]interface{})
	if !ok {
		return nil, nil, newPosError(fmt.Sprintf("mapObject expects an object, got %T", objVal), e.Args[0].Pos())
	}

	lambdaVal, err := evalASTWithDepth(e.Args[1], context, depth)
	if err != nil {
		return nil, nil, err
	}
	lambda, ok := lambdaVal.(*Lambda)
	if !ok {
		return nil, nil, newPosError(fmt.Sprintf("mapObject expects a lambda function, got %T", lambdaVal), e.Args[1].Pos())
	}
	if lambda.ParamCount() != 2 {
		return nil, nil, newPosError(fmt.Sprintf("mapObject lambda must have exactly 2 parameters, got %d", lambda.ParamCount()), e.Args[1].Pos())
	}

	return obj, lambda, nil
}

func mapObjectLambdaContext(context map[string]interface{}, param0, param1 string, key string, value interface{}, dwOrder bool) map[string]interface{} {
	lambdaContext := copyContext(context)
	if dwOrder {
		lambdaContext[param0], lambdaContext[param1] = value, key
	} else {
		lambdaContext[param0], lambdaContext[param1] = key, value
	}
	return lambdaContext
}

func applyMapObject(obj map[string]interface{}, lambda *Lambda, context map[string]interface{}, depth int, pos token.Pos) (map[string]interface{}, error) {
	param0, param1 := lambda.ParamName(0), lambda.ParamName(1)
	dwOrder := !shouldUseLegacyObjectLambdaOrder(lambda)
	keys := sortedKeys(obj)
	result := make(map[string]interface{})

	for _, key := range keys {
		value := obj[key]

		if multi, ok := value.(XMLMultiValue); ok {
			for _, v := range multi {
				if err := applyAndMerge(v, key, result, lambda, context, param0, param1, dwOrder, depth, pos); err != nil {
					return nil, err
				}
			}
		} else {
			if err := applyAndMerge(value, key, result, lambda, context, param0, param1, dwOrder, depth, pos); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

func applyAndMerge(value interface{}, key string, result map[string]interface{}, lambda *Lambda, context map[string]interface{}, param0, param1 string, dwOrder bool, depth int, pos token.Pos) error {
	lambdaContext := mapObjectLambdaContext(context, param0, param1, key, value, dwOrder)

	mapResult, err := evalASTWithDepth(lambda.BodyAST, lambdaContext, depth+1)
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
	for entryKey, entryValue := range entries {
		if existing, ok := result[entryKey]; ok {
			switch v := existing.(type) {
			case XMLMultiValue:
				result[entryKey] = append(v, entryValue)
			default:
				result[entryKey] = XMLMultiValue{existing, entryValue}
			}
		} else {
			result[entryKey] = entryValue
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
//   - lambda: A function with exactly 2 parameters
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
func callBuiltinMapObject(e *ast.CallExpr, context map[string]interface{}, depth int) (interface{}, error) {
	obj, lambda, err := evalMapObjectInputs(e, context, depth)
	if err != nil {
		return nil, err
	}

	return applyMapObject(obj, lambda, context, depth, e.Args[1].Pos())
}

// sortedKeys returns the keys of the map in sorted order.
func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
