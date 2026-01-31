package evaluator

import (
	"context"
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

func BenchmarkLazyValueEvaluation(b *testing.B) {
	lv := NewLazyValue(func(ctx context.Context) (Value, error) {
		return 42, nil
	}, context.Background())

	for i := 0; i < b.N; i++ {
		_, _ = lv.GetValue()
	}
}
