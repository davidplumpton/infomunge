package evaluator

import "testing"

func TestEmptyAggregationOperationsReturnNull(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{"max", "max([]interface{}{})"},
		{"min", "min([]interface{}{})"},
		{"maxBy", `maxBy([]interface{}{}, __lambda("x", x))`},
		{"minBy", `minBy([]interface{}{}, __lambda("x", x))`},
		{"reduce without initial value", `__reduce([]interface{}{}, __lambda("x, acc", acc + x))`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.expr, Context{}, nil, 0, tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != nil {
				t.Fatalf("expected null, got %#v (%T)", result, result)
			}
		})
	}
}
