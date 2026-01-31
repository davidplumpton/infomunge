package evaluator

import (
	"context"
	"fmt"
	"go/ast"
)

// callBuiltinToStream implements the __toStream(array) function.
func callBuiltinToStream(e *ast.CallExpr, evalCtx map[string]interface{}, depth int) (interface{}, error) {
	if len(e.Args) != 1 {
		return nil, newPosError("__toStream expects 1 argument", e.Pos())
	}

	arrayVal, err := evalASTWithDepth(e.Args[0], evalCtx, depth)
	if err != nil {
		return nil, err
	}

	array, ok := AsArray(arrayVal)
	if !ok {
		return nil, newPosError("__toStream expects an array", e.Args[0].Pos())
	}

	return NewLazyValue(func(ctx context.Context) (Value, error) {
		stream := make(chan Value)
		go func() {
			defer close(stream)
			for _, val := range array {
				select {
				case stream <- val:
				case <-ctx.Done():
					return
				}
			}
		}()
		return stream, nil
	}, GetGoContext(evalCtx)), nil
}

// callBuiltinLazyMap implements the __lazyMap(lazyStream, lambda) function.
func callBuiltinLazyMap(e *ast.CallExpr, evalCtx map[string]interface{}, depth int) (interface{}, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("__lazyMap expects 2 arguments", e.Pos())
	}

	lazyVal, err := evalASTWithDepth(e.Args[0], evalCtx, depth)
	if err != nil {
		return nil, err
	}

	lazy, ok := lazyVal.(*LazyValue)
	if !ok {
		return nil, newPosError("__lazyMap first argument must be a lazy value", e.Args[0].Pos())
	}

	lambdaVal, err := evalASTWithDepth(e.Args[1], evalCtx, depth)
	if err != nil {
		return nil, err
	}

	lambda, ok := AsLambda(lambdaVal)
	if !ok {
		return nil, newPosError("__lazyMap second argument must be a lambda", e.Args[1].Pos())
	}

	return LazyMap(lazy, lambda, evalCtx), nil
}

// callBuiltinLazyFilter implements the __lazyFilter(lazyStream, lambda) function.
func callBuiltinLazyFilter(e *ast.CallExpr, evalCtx map[string]interface{}, depth int) (interface{}, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("__lazyFilter expects 2 arguments", e.Pos())
	}

	lazyVal, err := evalASTWithDepth(e.Args[0], evalCtx, depth)
	if err != nil {
		return nil, err
	}

	lazy, ok := lazyVal.(*LazyValue)
	if !ok {
		return nil, newPosError("__lazyFilter first argument must be a lazy value", e.Args[0].Pos())
	}

	lambdaVal, err := evalASTWithDepth(e.Args[1], evalCtx, depth)
	if err != nil {
		return nil, err
	}

	lambda, ok := AsLambda(lambdaVal)
	if !ok {
		return nil, newPosError("__lazyFilter second argument must be a lambda", e.Args[1].Pos())
	}

	return LazyFilter(lazy, lambda, evalCtx), nil
}

// callBuiltinLazyReduce implements the __lazyReduce(lazyStream, lambda, initial) function.
func callBuiltinLazyReduce(e *ast.CallExpr, evalCtx map[string]interface{}, depth int) (interface{}, error) {
	if len(e.Args) != 3 {
		return nil, newPosError("__lazyReduce expects 3 arguments", e.Pos())
	}

	lazyVal, err := evalASTWithDepth(e.Args[0], evalCtx, depth)
	if err != nil {
		return nil, err
	}

	lazy, ok := lazyVal.(*LazyValue)
	if !ok {
		return nil, newPosError("__lazyReduce first argument must be a lazy value", e.Args[0].Pos())
	}

	lambdaVal, err := evalASTWithDepth(e.Args[1], evalCtx, depth)
	if err != nil {
		return nil, err
	}

	lambda, ok := AsLambda(lambdaVal)
	if !ok {
		return nil, newPosError("__lazyReduce second argument must be a lambda", e.Args[1].Pos())
	}
	if lambda.ParamCount() < 2 || lambda.ParamCount() > 3 {
		return nil, newPosError(fmt.Sprintf("lazyReduce lambda must have 2 or 3 parameters, got %d", lambda.ParamCount()), e.Args[1].Pos())
	}

	initial, err := evalASTWithDepth(e.Args[2], evalCtx, depth)
	if err != nil {
		return nil, err
	}

	return LazyReduce(lazy, lambda, initial, evalCtx), nil
}
