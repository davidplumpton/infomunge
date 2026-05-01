package evaluator

import (
	"context"
	"fmt"
	"go/ast"
)

// callBuiltinToStream implements the __toStream(array) function.
func callBuiltinToStream(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) != 1 {
		return nil, newPosError("__toStream expects 1 argument", e.Pos())
	}

	arrayVal, err := evalASTInScopeWithDepth(e.Args[0], scope, depth)
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
	}, scope.GoContext()), nil
}

// callBuiltinLazyMap implements the __lazyMap(lazyStream, lambda) function.
func callBuiltinLazyMap(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("__lazyMap expects 2 arguments", e.Pos())
	}

	lazyVal, err := evalASTInScopeWithDepth(e.Args[0], scope, depth)
	if err != nil {
		return nil, err
	}

	lazy, ok := lazyVal.(*LazyValue)
	if !ok {
		return nil, newPosError("__lazyMap first argument must be a lazy value", e.Args[0].Pos())
	}

	lambdaVal, err := evalASTInScopeWithDepth(e.Args[1], scope, depth)
	if err != nil {
		return nil, err
	}

	lambda, ok := AsLambda(lambdaVal)
	if !ok {
		return nil, newPosError("__lazyMap second argument must be a lambda", e.Args[1].Pos())
	}

	return LazyMapInScope(lazy, lambda, scope), nil
}

// callBuiltinLazyFilter implements the __lazyFilter(lazyStream, lambda) function.
func callBuiltinLazyFilter(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("__lazyFilter expects 2 arguments", e.Pos())
	}

	lazyVal, err := evalASTInScopeWithDepth(e.Args[0], scope, depth)
	if err != nil {
		return nil, err
	}

	lazy, ok := lazyVal.(*LazyValue)
	if !ok {
		return nil, newPosError("__lazyFilter first argument must be a lazy value", e.Args[0].Pos())
	}

	lambdaVal, err := evalASTInScopeWithDepth(e.Args[1], scope, depth)
	if err != nil {
		return nil, err
	}

	lambda, ok := AsLambda(lambdaVal)
	if !ok {
		return nil, newPosError("__lazyFilter second argument must be a lambda", e.Args[1].Pos())
	}

	return LazyFilterInScope(lazy, lambda, scope), nil
}

// callBuiltinLazyReduce implements the __lazyReduce(lazyStream, lambda, initial) function.
func callBuiltinLazyReduce(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) != 3 {
		return nil, newPosError("__lazyReduce expects 3 arguments", e.Pos())
	}

	lazyVal, err := evalASTInScopeWithDepth(e.Args[0], scope, depth)
	if err != nil {
		return nil, err
	}

	lazy, ok := lazyVal.(*LazyValue)
	if !ok {
		return nil, newPosError("__lazyReduce first argument must be a lazy value", e.Args[0].Pos())
	}

	lambdaVal, err := evalASTInScopeWithDepth(e.Args[1], scope, depth)
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

	initial, err := evalASTInScopeWithDepth(e.Args[2], scope, depth)
	if err != nil {
		return nil, err
	}

	return LazyReduceInScope(lazy, lambda, initial, scope), nil
}
