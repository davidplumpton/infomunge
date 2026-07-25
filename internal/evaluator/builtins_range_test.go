package evaluator

import (
	"go/ast"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestRangePreservesExactRuntimeIntegers(t *testing.T) {
	const expr = "__range(4)"
	result, err := Evaluate(expr, Context{}, nil, 0, expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := Array{0, 1, 2, 3}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %#v, got %#v", expected, result)
	}
	for index, value := range result.(Array) {
		if _, ok := value.(int); !ok {
			t.Fatalf("expected element %d to be an exact runtime int, got %#v (%T)", index, value, value)
		}
	}
}

func TestExactRangeEndValidatesNumericBoundaries(t *testing.T) {
	call := &ast.CallExpr{Fun: &ast.Ident{Name: "range"}}

	type testCase struct {
		name        string
		value       Value
		expected    int
		errorSubstr string
	}
	tests := []testCase{
		{"minimum runtime integer", minInt(), minInt(), ""},
		{"fractional float", 3.5, 0, "numeric precision loss"},
		{"positive infinity", math.Inf(1), 0, "finite number"},
		{"negative infinity", math.Inf(-1), 0, "finite number"},
		{"not a number", math.NaN(), 0, "finite number"},
		{"float beyond int64 maximum", float64(math.MaxInt64), 0, "outside the supported integer range"},
		{"non-number", "5", 0, "range expects a number"},
	}
	if strconv.IntSize == 64 {
		aboveExactFloatBoundary := int(uint64(1)<<53) + 1
		tests = append(tests,
			testCase{"integer immediately above 2^53", aboveExactFloatBoundary, aboveExactFloatBoundary, ""},
			testCase{"exact float at 2^53", float64(uint64(1) << 53), int(uint64(1) << 53), ""},
			testCase{"maximum runtime integer", int(math.MaxInt64), int(math.MaxInt64), ""},
			testCase{"minimum int64 float", float64(math.MinInt64), int(math.MinInt64), ""},
		)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := exactRangeEnd(tt.value, call)
			if tt.errorSubstr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Fatalf("expected %d, got %d", tt.expected, result)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error containing %q, got result %d", tt.errorSubstr, result)
			}
			if !strings.Contains(err.Error(), tt.errorSubstr) {
				t.Fatalf("expected error containing %q, got %v", tt.errorSubstr, err)
			}
		})
	}
}
