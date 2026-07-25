package evaluator

import (
	"context"
	"errors"
	"go/parser"
	"testing"
)

func TestNewLazyValue(t *testing.T) {
	lv := NewLazyValue(func(ctx context.Context) (Value, error) {
		return "test", nil
	}, context.Background())

	val, err := lv.GetValue()
	if err != nil {
		t.Fatalf("GetValue error: %v", err)
	}
	if val != "test" {
		t.Errorf("Expected 'test', got %v", val)
	}
}

func TestLazyMapType(t *testing.T) {
	streamLazy := NewLazyValue(func(ctx context.Context) (Value, error) {
		return make(chan Value), nil
	}, context.Background())

	lambda := &Lambda{} // Dummy lambda

	result := LazyMap(streamLazy, lambda, Context{})

	if result == nil {
		t.Error("LazyMap should return a LazyValue")
	}
}

func TestLazyMapPropagatesErrors(t *testing.T) {
	streamLazy := NewLazyValue(func(ctx context.Context) (Value, error) {
		stream := make(chan Value, 1)
		stream <- 1
		close(stream)
		return stream, nil
	}, context.Background())

	expr, err := parser.ParseExpr("unknownFunc()")
	if err != nil {
		t.Fatalf("ParseExpr error: %v", err)
	}
	lambda := &Lambda{
		Params:  []ParamDef{{Name: "x"}},
		Body:    "unknownFunc()",
		BodyAST: expr,
	}

	result := LazyMap(streamLazy, lambda, Context{})
	val, err := result.GetValue()
	if err != nil {
		t.Fatalf("GetValue error: %v", err)
	}
	streamResult, ok := val.(*StreamWithError)
	if !ok {
		t.Fatalf("expected StreamWithError, got %T", val)
	}
	for range streamResult.Stream {
	}
	if err := streamResult.WaitError(); err == nil {
		t.Fatal("expected error from lazy map")
	}
}

func TestLazyFilterPropagatesErrors(t *testing.T) {
	streamLazy := NewLazyValue(func(ctx context.Context) (Value, error) {
		stream := make(chan Value, 1)
		stream <- 1
		close(stream)
		return stream, nil
	}, context.Background())

	expr, err := parser.ParseExpr("unknownFunc()")
	if err != nil {
		t.Fatalf("ParseExpr error: %v", err)
	}
	lambda := &Lambda{
		Params:  []ParamDef{{Name: "x"}},
		Body:    "unknownFunc()",
		BodyAST: expr,
	}

	result := LazyFilter(streamLazy, lambda, Context{})
	val, err := result.GetValue()
	if err != nil {
		t.Fatalf("GetValue error: %v", err)
	}
	streamResult, ok := val.(*StreamWithError)
	if !ok {
		t.Fatalf("expected StreamWithError, got %T", val)
	}
	for range streamResult.Stream {
	}
	if err := streamResult.WaitError(); err == nil {
		t.Fatal("expected error from lazy filter")
	}
}

func TestLazyReduceBindsZeroBasedIndex(t *testing.T) {
	streamLazy := NewLazyValue(func(ctx context.Context) (Value, error) {
		stream := make(chan Value, 3)
		stream <- 10
		stream <- 20
		stream <- 30
		close(stream)
		return stream, nil
	}, context.Background())

	expr, err := parser.ParseExpr("acc + idx")
	if err != nil {
		t.Fatalf("ParseExpr error: %v", err)
	}
	lambda := &Lambda{
		Params:  []ParamDef{{Name: "acc"}, {Name: "value"}, {Name: "idx"}},
		Body:    "acc + idx",
		BodyAST: expr,
	}

	result := LazyReduceInScope(streamLazy, lambda, 0, NewScope(nil))
	got, err := result.GetValue()
	if err != nil {
		t.Fatalf("GetValue error: %v", err)
	}
	if got != 3 {
		t.Errorf("expected zero-based indexes to sum to 3, got %v", got)
	}
}

func TestLazyReducePropagatesStreamError(t *testing.T) {
	wantErr := errors.New("stream failed")
	streamLazy := NewLazyValue(func(ctx context.Context) (Value, error) {
		stream := make(chan Value, 1)
		errCh := make(chan error, 1)
		stream <- 10
		close(stream)
		errCh <- wantErr
		close(errCh)
		return &StreamWithError{Stream: stream, Err: errCh}, nil
	}, context.Background())

	expr, err := parser.ParseExpr("acc + value")
	if err != nil {
		t.Fatalf("ParseExpr error: %v", err)
	}
	lambda := &Lambda{
		Params:  []ParamDef{{Name: "acc"}, {Name: "value"}},
		Body:    "acc + value",
		BodyAST: expr,
	}

	result := LazyReduceInScope(streamLazy, lambda, 0, NewScope(nil))
	_, err = result.GetValue()
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected stream error %v, got %v", wantErr, err)
	}
}

func BenchmarkLazyValueEvaluation(b *testing.B) {
	lv := NewLazyValue(func(ctx context.Context) (Value, error) {
		return 42, nil
	}, context.Background())

	for i := 0; i < b.N; i++ {
		_, _ = lv.GetValue()
	}
}
