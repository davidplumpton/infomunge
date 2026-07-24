package evaluator

import (
	"context"
	"errors"
	"go/ast"
	"sync/atomic"
	"testing"
)

type cancelAfterChecksContext struct {
	context.Context
	checks      atomic.Int64
	cancelAfter int64
	cancel      context.CancelFunc
}

func (c *cancelAfterChecksContext) Err() error {
	if c.checks.Add(1) >= c.cancelAfter {
		c.cancel()
	}
	return c.Context.Err()
}

func TestCollectionLambdaStopsWhenContextIsCanceled(t *testing.T) {
	const elementCount = 10_000
	array := make(Array, elementCount)
	for i := range array {
		array[i] = i
	}

	baseCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	goCtx := &cancelAfterChecksContext{
		Context:     baseCtx,
		cancelAfter: 25,
		cancel:      cancel,
	}
	scope := NewScope(nil).WithGoContext(goCtx)
	lambda := &Lambda{
		Params:  []ParamDef{{Name: "item"}},
		BodyAST: &ast.Ident{Name: "item"},
	}

	_, err := callBuiltinMapInternal(array, lambda, scope, 0)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("callBuiltinMapInternal() error = %v, want context.Canceled", err)
	}
	if checks := goCtx.checks.Load(); checks >= elementCount {
		t.Fatalf("context was checked %d times; collection evaluation did not stop promptly", checks)
	}
}
