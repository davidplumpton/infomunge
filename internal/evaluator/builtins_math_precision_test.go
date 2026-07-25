package evaluator

import (
	"reflect"
	"strings"
	"testing"
)

func TestNumericBuiltinsPreserveExactIntegers(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		context  Context
		expected Value
	}{
		{"ceil", "ceil(9007199254740993)", nil, 9007199254740993},
		{"floor", "floor(9007199254740993)", nil, 9007199254740993},
		{"round", "round(9007199254740993)", nil, 9007199254740993},
		{"perfect square root", "sqrt(9007199515875289)", nil, 94906267},
		{"absolute value", "abs(9007199254740993)", nil, 9007199254740993},
		{"maximum", "max(payload)", Context{"payload": Array{9007199254740992, 9007199254740993}}, 9007199254740993},
		{"minimum", "min(payload)", Context{"payload": Array{9007199254740994, 9007199254740993}}, 9007199254740993},
		{"power identity", "pow(9007199254740993, 1)", nil, 9007199254740993},
		{"sum", "sum(payload)", Context{"payload": Array{9007199254740993, 1}}, 9007199254740994},
		{"sum mixed with exact zero", "sum(payload)", Context{"payload": Array{9007199254740993, 0.0}}, 9007199254740993},
		{"average", "avg(payload)", Context{"payload": Array{9007199254740993, 9007199254740993}}, 9007199254740993},
		{"modulo", "mod(9007199254740993, 2)", nil, 1},
		{"integer predicate at int64 boundary", "isInteger(9223372036854775808.0)", nil, true},
		{"even predicate at int64 boundary", "isEven(9223372036854775808.0)", nil, true},
		{"odd predicate at int64 boundary", "isOdd(9223372036854775808.0)", nil, false},
		{"range endpoint", "to(9223372036854775807, 9223372036854775807)", nil, Array{9223372036854775807}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.expr, tt.context, nil, 0, tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Fatalf("expected %#v (%T), got %#v (%T)", tt.expected, tt.expected, result, result)
			}
		})
	}
}

func TestNumericBuiltinsRejectPrecisionLossAndRangeOverflow(t *testing.T) {
	tests := []struct {
		name        string
		expr        string
		context     Context
		errorSubstr string
	}{
		{"absolute value overflow", "abs(-9223372036854775807 - 1)", nil, "integer overflow during abs"},
		{"sum overflow", "sum(payload)", Context{"payload": Array{9223372036854775807, 1}}, "integer overflow during sum"},
		{"average precision loss", "avg(payload)", Context{"payload": Array{9007199254740992, 9007199254740993}}, "numeric precision loss during avg"},
		{"power overflow", "pow(9223372036854775807, 2)", nil, "integer overflow during pow"},
		{"square root input precision", "sqrt(9007199254740993)", nil, "numeric precision loss"},
		{"random bound overflow", "randomInt(9223372036854775808.0)", nil, "outside the supported integer range"},
		{"range start overflow", "to(9223372036854775808.0, 1)", nil, "outside the supported integer range"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.expr, tt.context, nil, 0, tt.expr)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.errorSubstr) {
				t.Fatalf("expected error containing %q, got %v", tt.errorSubstr, err)
			}
		})
	}
}

func TestRandomIntReturnsAnExactRuntimeInteger(t *testing.T) {
	const expr = "randomInt(9007199254740993)"
	result, err := Evaluate(expr, Context{}, nil, 0, expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	integer, ok := result.(int)
	if !ok {
		t.Fatalf("expected int result, got %#v (%T)", result, result)
	}
	if integer < 0 || integer >= 9007199254740993 {
		t.Fatalf("randomInt result %d is outside the requested bound", integer)
	}
}

func TestToPreservesFractionalBounds(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected Array
	}{
		{"ascending", "to(1.9, 3.1)", Array{1.9, 2.9}},
		{"descending", "to(3.1, 1.9)", Array{3.1, 2.1}},
		{"ascending exact endpoint", "to(1.1, 3.1)", Array{1.1, 2.1, 3.1}},
		{"descending exact endpoint", "to(3.1, 1.1)", Array{3.1, 2.1, 1.1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.expr, Context{}, nil, 0, tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Fatalf("expected %#v, got %#v", tt.expected, result)
			}
		})
	}
}
