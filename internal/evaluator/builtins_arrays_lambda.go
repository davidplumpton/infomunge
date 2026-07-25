package evaluator

import (
	"fmt"
	"go/ast"
)

type collectionNullPolicy bool

const (
	rejectNullCollectionSource    collectionNullPolicy = false
	propagateNullCollectionSource collectionNullPolicy = true
)

// evalCollectionSource evaluates the source argument and applies the operation's
// explicit null policy. The handled result tells null-propagating callers to
// return language null without evaluating their selector lambda.
func evalCollectionSource(e *ast.CallExpr, scope *Scope, depth int, nullPolicy collectionNullPolicy) (Value, bool, error) {
	source, err := evalASTInScopeWithDepth(e.Args[0], scope, depth)
	if err != nil {
		return nil, false, err
	}
	if source == nil && nullPolicy == propagateNullCollectionSource {
		return nil, true, nil
	}
	return source, false, nil
}

// evalArrayAndLambda is a helper that extracts and validates an array and lambda
func evalArrayAndLambda(funcName string, e *ast.CallExpr, scope *Scope, depth int, minParams, maxParams int) (Array, *Lambda, error) {
	array, lambda, _, err := evalArrayAndLambdaWithNullPolicy(
		funcName,
		e,
		scope,
		depth,
		minParams,
		maxParams,
		rejectNullCollectionSource,
	)
	return array, lambda, err
}

func evalNullPropagatingArrayAndLambda(funcName string, e *ast.CallExpr, scope *Scope, depth int, minParams, maxParams int) (Array, *Lambda, bool, error) {
	return evalArrayAndLambdaWithNullPolicy(
		funcName,
		e,
		scope,
		depth,
		minParams,
		maxParams,
		propagateNullCollectionSource,
	)
}

func evalArrayAndLambdaWithNullPolicy(
	funcName string,
	e *ast.CallExpr,
	scope *Scope,
	depth int,
	minParams, maxParams int,
	nullPolicy collectionNullPolicy,
) (Array, *Lambda, bool, error) {
	if len(e.Args) != 2 {
		return nil, nil, false, newPosError(fmt.Sprintf("%s requires exactly 2 arguments: array, lambda", funcName), e.Pos())
	}

	arrayVal, nullHandled, err := evalCollectionSource(e, scope, depth, nullPolicy)
	if err != nil {
		return nil, nil, false, err
	}
	if nullHandled {
		return nil, nil, true, nil
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
			return nil, nil, false, newPosError(fmt.Sprintf("%s expects an array, got object%s", funcName, suggestion), e.Args[0].Pos())
		}
		return nil, nil, false, newPosError(fmt.Sprintf("%s expects an array, got %T", funcName, arrayVal), e.Args[0].Pos())
	}

	lambdaVal, err := evalASTInScopeWithDepth(e.Args[1], scope, depth)
	if err != nil {
		return nil, nil, false, err
	}

	lambda, ok := lambdaVal.(*Lambda)
	if !ok {
		return nil, nil, false, newPosError(fmt.Sprintf("%s expects a lambda function, got %T", funcName, lambdaVal), e.Args[1].Pos())
	}

	if lambda.ParamCount() < minParams || lambda.ParamCount() > maxParams {
		var requirement string
		if maxParams == minParams+1 {
			requirement = fmt.Sprintf("%d or %d", minParams, maxParams)
		} else {
			requirement = fmt.Sprintf("between %d and %d", minParams, maxParams)
		}
		return nil, nil, false, newPosError(fmt.Sprintf("%s lambda must have %s parameters, got %d", funcName, requirement, lambda.ParamCount()), e.Args[1].Pos())
	}

	return array, lambda, false, nil
}

// callBuiltinFilter implements the __filter(array, lambda) function.
func callBuiltinFilter(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	array, lambda, nullHandled, err := evalNullPropagatingArrayAndLambda("filter", e, scope, depth, 0, 2)
	if err != nil {
		return nil, err
	}
	if nullHandled {
		return nil, nil
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
	array, lambda, nullHandled, err := evalNullPropagatingArrayAndLambda("takeWhile", e, scope, depth, 0, 2)
	if err != nil {
		return nil, err
	}
	if nullHandled {
		return nil, nil
	}

	result := make(Array, 0, len(array))
	for i, elem := range array {
		condVal, err := evalLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
			bindArrayLambdaParameters(lambdaContext, lambda, elem, i)
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
	array, lambda, nullHandled, err := evalNullPropagatingArrayAndLambda("dropWhile", e, scope, depth, 0, 2)
	if err != nil {
		return nil, err
	}
	if nullHandled {
		return nil, nil
	}

	result := make(Array, 0, len(array))
	skip := true
	for i, elem := range array {
		condVal, err := evalLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
			bindArrayLambdaParameters(lambdaContext, lambda, elem, i)
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
	array, lambda, nullHandled, err := evalNullPropagatingArrayAndLambda("map", e, scope, depth, 0, 2)
	if err != nil {
		return nil, err
	}
	if nullHandled {
		return nil, nil
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
	array, lambda, nullHandled, err := evalNullPropagatingArrayAndLambda("flatMap", e, scope, depth, 0, 2)
	if err != nil {
		return nil, err
	}
	if nullHandled {
		return nil, nil
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
