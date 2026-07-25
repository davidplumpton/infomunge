package evaluator

import (
	"context"
	"fmt"
	unifiederrors "infomunge/internal/errors"
)

// Stream represents a lazy stream of values.
type Stream chan Value

// StreamWithError wraps a stream with an error channel for lazy evaluation.
type StreamWithError struct {
	Stream chan Value
	Err    chan error
}

// WaitError drains the error channel and returns the first error found, if any.
func (s *StreamWithError) WaitError() error {
	if s == nil || s.Err == nil {
		return nil
	}
	for err := range s.Err {
		if err != nil {
			return err
		}
	}
	return nil
}

func streamFromValue(val Value) (chan Value, <-chan error, bool) {
	switch v := val.(type) {
	case Stream:
		return chan Value(v), nil, true
	case chan Value:
		return v, nil, true
	case *StreamWithError:
		if v == nil {
			return nil, nil, false
		}
		return v.Stream, v.Err, true
	default:
		return nil, nil, false
	}
}

func sendErrOnce(errCh chan error, err error) {
	if err == nil || errCh == nil {
		return
	}
	select {
	case errCh <- err:
	default:
	}
}

// LazyMap applies a lambda function to each element of a lazy stream, returning a new lazy stream.
func LazyMap(input *LazyValue, lambda *Lambda, evalCtx Context) *LazyValue {
	return LazyMapInScope(input, lambda, NewScope(evalCtx))
}

// LazyMapInScope applies a lambda using an explicit evaluation scope.
func LazyMapInScope(input *LazyValue, lambda *Lambda, scope *Scope) *LazyValue {
	if scope == nil {
		scope = NewScope(nil)
	}
	return NewLazyValue(func(ctx context.Context) (Value, error) {
		inputVal, err := input.GetValue()
		if err != nil {
			return nil, err
		}
		inputStream, inputErr, ok := streamFromValue(inputVal)
		if !ok {
			return nil, unifiederrors.EvalError("lazyMap input must be a stream")
		}

		output := make(chan Value)
		errCh := make(chan error, 1)
		go func() {
			defer close(output)
			defer close(errCh)
			for val := range inputStream {
				// Apply lambda to the value
				result, err := applyLambda(lambda, val, scope, 0)
				if err != nil {
					sendErrOnce(errCh, err)
					for range inputStream {
					}
					break
				}
				select {
				case output <- result:
				case <-ctx.Done():
					return
				}
			}
			if inputErr != nil {
				for err := range inputErr {
					if err != nil {
						sendErrOnce(errCh, err)
						break
					}
				}
			}
		}()
		return &StreamWithError{Stream: output, Err: errCh}, nil
	}, scope.GoContext())
}

// LazyFilter filters elements of a lazy stream using a predicate lambda, returning a new lazy stream.
func LazyFilter(input *LazyValue, lambda *Lambda, evalCtx Context) *LazyValue {
	return LazyFilterInScope(input, lambda, NewScope(evalCtx))
}

// LazyFilterInScope filters elements using an explicit evaluation scope.
func LazyFilterInScope(input *LazyValue, lambda *Lambda, scope *Scope) *LazyValue {
	if scope == nil {
		scope = NewScope(nil)
	}
	return NewLazyValue(func(ctx context.Context) (Value, error) {
		inputVal, err := input.GetValue()
		if err != nil {
			return nil, err
		}
		inputStream, inputErr, ok := streamFromValue(inputVal)
		if !ok {
			return nil, unifiederrors.EvalError("lazyFilter input must be a stream")
		}

		output := make(chan Value)
		errCh := make(chan error, 1)
		go func() {
			defer close(output)
			defer close(errCh)
			for val := range inputStream {
				// Apply predicate
				predVal, err := applyLambda(lambda, val, scope, 0)
				if err != nil {
					sendErrOnce(errCh, err)
					for range inputStream {
					}
					break
				}
				pred, ok := predVal.(bool)
				if !ok {
					sendErrOnce(errCh, fmt.Errorf("filter lambda must return a boolean, got %T", predVal))
					for range inputStream {
					}
					break
				}
				if !pred {
					continue
				}
				select {
				case output <- val:
				case <-ctx.Done():
					return
				}
			}
			if inputErr != nil {
				for err := range inputErr {
					if err != nil {
						sendErrOnce(errCh, err)
						break
					}
				}
			}
		}()
		return &StreamWithError{Stream: output, Err: errCh}, nil
	}, scope.GoContext())
}

// LazyReduceInScope aggregates a lazy stream using an explicit evaluation scope.
func LazyReduceInScope(input *LazyValue, lambda *Lambda, initial Value, scope *Scope) *LazyValue {
	if scope == nil {
		scope = NewScope(nil)
	}
	return NewLazyValue(func(ctx context.Context) (Value, error) {
		inputVal, err := input.GetValue()
		if err != nil {
			return nil, err
		}
		inputStream, inputErr, ok := streamFromValue(inputVal)
		if !ok {
			return nil, unifiederrors.EvalError("lazyReduce input must be a stream")
		}

		acc := initial
		index := 0
		for val := range inputStream {
			// Apply lambda(acc, val, index)
			result, err := applyLambdaReduce(lambda, acc, val, index, scope, 0)
			if err != nil {
				return nil, err
			}
			acc = result
			index++
		}
		if inputErr != nil {
			for err := range inputErr {
				if err != nil {
					return nil, err
				}
			}
		}
		return acc, nil
	}, scope.GoContext())
}

// applyLambda applies a lambda to a single argument.
func applyLambda(lambda *Lambda, arg Value, scope *Scope, depth int) (Value, error) {
	return evalLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
		lambdaContext[lambda.ParamName(0)] = arg
	})
}

// applyLambdaReduce applies a lambda to accumulator, current value, and optional index for reduce.
func applyLambdaReduce(lambda *Lambda, acc Value, val Value, index int, scope *Scope, depth int) (Value, error) {
	return evalLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
		lambdaContext[lambda.ParamName(0)] = acc
		lambdaContext[lambda.ParamName(1)] = val
		if lambda.ParamCount() > 2 {
			lambdaContext[lambda.ParamName(2)] = index
		}
	})
}
