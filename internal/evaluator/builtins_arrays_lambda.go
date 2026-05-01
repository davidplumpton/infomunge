package evaluator

import (
	"fmt"
	"go/ast"
)

// evalArrayAndLambda is a helper that extracts and validates an array and lambda
func evalArrayAndLambda(funcName string, e *ast.CallExpr, scope *Scope, depth int, minParams, maxParams int) (Array, *Lambda, error) {
	if len(e.Args) != 2 {
		return nil, nil, newPosError(fmt.Sprintf("%s requires exactly 2 arguments: array, lambda", funcName), e.Pos())
	}

	arrayVal, err := evalASTInScopeWithDepth(e.Args[0], scope, depth)
	if err != nil {
		return nil, nil, err
	}

	array, ok := arrayVal.(Array)
	if !ok {
		// Check for map to provide better error message
		if _, isMap := arrayVal.(Object); isMap {
			var suggestion string
			switch funcName {
			case "map":
				suggestion = ". Did you mean to use 'mapObject'?"
			case "filter":
				suggestion = ". Did you mean to use 'filterObject'?"
			default:
				suggestion = ". To iterate over object values, try 'values(object)'."
			}
			return nil, nil, newPosError(fmt.Sprintf("%s expects an array, got object%s", funcName, suggestion), e.Args[0].Pos())
		}
		return nil, nil, newPosError(fmt.Sprintf("%s expects an array, got %T", funcName, arrayVal), e.Args[0].Pos())
	}

	lambdaVal, err := evalASTInScopeWithDepth(e.Args[1], scope, depth)
	if err != nil {
		return nil, nil, err
	}

	lambda, ok := lambdaVal.(*Lambda)
	if !ok {
		return nil, nil, newPosError(fmt.Sprintf("%s expects a lambda function, got %T", funcName, lambdaVal), e.Args[1].Pos())
	}

	if lambda.ParamCount() < minParams || lambda.ParamCount() > maxParams {
		return nil, nil, newPosError(fmt.Sprintf("%s lambda must have %d or %d parameters, got %d", funcName, minParams, maxParams, lambda.ParamCount()), e.Args[1].Pos())
	}

	return array, lambda, nil
}

// callBuiltinFilter implements the __filter(array, lambda) function.
func callBuiltinFilter(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	array, lambda, err := evalArrayAndLambda("filter", e, scope, depth, 1, 2)
	if err != nil {
		return nil, err
	}

	result := make(Array, 0, len(array))
	err = executeLambdaOnArrayElements(array, lambda, scope, depth, func(elem Value, _ int, condVal Value) error {
		// Convert condition to boolean
		condBool, ok := condVal.(bool)
		if !ok {
			return newPosError(fmt.Sprintf("filter lambda must return a boolean, got %T", condVal), e.Args[1].Pos())
		}

		// If the condition is true, include the element
		if condBool {
			result = append(result, elem)
		}
		return nil
	})
	return result, err
}

// callBuiltinTakeWhile implements the takeWhile(array, lambda) function.
func callBuiltinTakeWhile(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	array, lambda, err := evalArrayAndLambda("takeWhile", e, scope, depth, 1, 2)
	if err != nil {
		return nil, err
	}

	result := make(Array, 0, len(array))
	for i, elem := range array {
		condVal, err := evalLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
			lambdaContext[lambda.ParamName(0)] = elem
			if lambda.ParamCount() > 1 {
				lambdaContext[lambda.ParamName(1)] = i
			}
		})
		if err != nil {
			return nil, err
		}
		condBool, ok := condVal.(bool)
		if !ok {
			return nil, newPosError(fmt.Sprintf("takeWhile lambda must return a boolean, got %T", condVal), e.Args[1].Pos())
		}
		if condBool {
			result = append(result, elem)
		} else {
			break
		}
	}
	return result, nil
}

// callBuiltinDropWhile implements the dropWhile(array, lambda) function.
func callBuiltinDropWhile(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	array, lambda, err := evalArrayAndLambda("dropWhile", e, scope, depth, 1, 2)
	if err != nil {
		return nil, err
	}

	result := make(Array, 0, len(array))
	skip := true
	for i, elem := range array {
		condVal, err := evalLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
			lambdaContext[lambda.ParamName(0)] = elem
			if lambda.ParamCount() > 1 {
				lambdaContext[lambda.ParamName(1)] = i
			}
		})
		if err != nil {
			return nil, err
		}
		condBool, ok := condVal.(bool)
		if !ok {
			return nil, newPosError(fmt.Sprintf("dropWhile lambda must return a boolean, got %T", condVal), e.Args[1].Pos())
		}
		if skip && !condBool {
			skip = false
		}
		if !skip {
			result = append(result, elem)
		}
	}
	return result, nil
}

// callBuiltinMap implements the __map(array, lambda) function.
func callBuiltinMap(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	array, lambda, err := evalArrayAndLambda("map", e, scope, depth, 1, 2)
	if err != nil {
		return nil, err
	}

	result := make(Array, 0, len(array))
	err = executeLambdaOnArrayElements(array, lambda, scope, depth, func(_ Value, _ int, mappedVal Value) error {
		result = append(result, mappedVal)
		return nil
	})
	return result, err
}

// callBuiltinFlatMap implements the __flatMap(array, lambda) function.
func callBuiltinFlatMap(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	array, lambda, err := evalArrayAndLambda("flatMap", e, scope, depth, 1, 2)
	if err != nil {
		return nil, err
	}

	result := make(Array, 0)
	err = executeLambdaOnArrayElements(array, lambda, scope, depth, func(_ Value, _ int, mappedVal Value) error {
		// If the result is an array, flatten it by one level
		if mappedArray, ok := mappedVal.(Array); ok {
			result = append(result, mappedArray...)
		} else {
			result = append(result, mappedVal)
		}
		return nil
	})
	return result, err
}
