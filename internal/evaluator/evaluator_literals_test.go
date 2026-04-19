package evaluator

import (
	"testing"
)

func TestEvaluate_BasicLiterals(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected Value
	}{
		{"integer", "42", 42},
		{"float", "3.14", 3.14},
		{"string", `"hello"`, "hello"},
		{"empty string", `""`, ""},
		{"true", "true", true},
		{"false", "false", false},
		{"nil", "nil", nil},
	}

	ctx := make(Context)
	mapping := []int{0}
	for i := 1; i < 100; i++ {
		mapping = append(mapping, i)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.expr, ctx, mapping, 0, tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v (%T), got %v (%T)", tt.expected, tt.expected, result, result)
			}
		})
	}
}

func TestEvaluate_CompositeLiterals(t *testing.T) {
	ctx := make(Context)
	mapping := make([]int, 200)
	for i := range mapping {
		mapping[i] = i
	}

	t.Run("map literal", func(t *testing.T) {
		result, err := Evaluate(`map[string]interface{}{"a": 1}`, ctx, mapping, 0, `map[string]interface{}{"a": 1}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		m, ok := result.(Object)
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}
		if m["a"] != 1 {
			t.Errorf("expected m[a]=1, got %v", m["a"])
		}
	})

	t.Run("array literal", func(t *testing.T) {
		result, err := Evaluate(`[]interface{}{1, 2, 3}`, ctx, mapping, 0, `[]interface{}{1, 2, 3}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		arr, ok := result.(Array)
		if !ok {
			t.Fatalf("expected array, got %T", result)
		}
		if len(arr) != 3 {
			t.Errorf("expected len=3, got %d", len(arr))
		}
	})
}

func TestEvaluate_NilSafety(t *testing.T) {
	ctx := Object{
		"nullVal": nil,
	}
	mapping := make([]int, 100)
	for i := range mapping {
		mapping[i] = i
	}

	t.Run("nil index access returns nil", func(t *testing.T) {
		result, err := Evaluate("nullVal[0]", ctx, mapping, 0, "nullVal[0]")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})
}
