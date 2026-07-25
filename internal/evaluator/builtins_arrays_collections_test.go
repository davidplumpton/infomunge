package evaluator

import (
	"go/ast"
	"testing"
)

func TestDistinctValuesUsesStructuralLanguageEquality(t *testing.T) {
	values := Array{
		1,
		"1",
		true,
		"true",
		nil,
		"<nil>",
		Array{1},
		Array{"1"},
		Array{1},
		Object{"a": 1},
		Object{"a": "1"},
		Object{"a": 1},
	}

	result := distinctValues(values)
	expected := Array{
		1,
		"1",
		true,
		"true",
		nil,
		"<nil>",
		Array{1},
		Array{"1"},
		Object{"a": 1},
		Object{"a": "1"},
	}

	if !numericEquals(result, expected) {
		t.Fatalf("distinctValues() = %#v, want %#v", result, expected)
	}
}

func TestDistinctValuesTreatsExactlyEquivalentNumericTypesAsEqual(t *testing.T) {
	result := distinctValues(Array{1, 1.0, 2.5, 2.5})
	expected := Array{1, 2.5}

	if !numericEquals(result, expected) {
		t.Fatalf("distinctValues() = %#v, want %#v", result, expected)
	}
	if _, ok := result[0].(int); !ok {
		t.Fatalf("distinctValues() did not preserve the first numeric representation: %#v", result[0])
	}
}

func TestBuiltinRemoveSupportsArrayAndScalarRightOperands(t *testing.T) {
	tests := []struct {
		name     string
		right    Value
		expected Array
	}{
		{
			name:     "array removes every listed value",
			right:    Array{1, 3},
			expected: Array{2, 2},
		},
		{
			name:     "scalar removes every matching value",
			right:    2,
			expected: Array{1, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := callBuiltinRemove(
				[]Value{Array{1, 2, 3, 2}, tt.right},
				&ast.CallExpr{},
			)
			if err != nil {
				t.Fatalf("callBuiltinRemove() error = %v", err)
			}
			if !numericEquals(got, tt.expected) {
				t.Fatalf("callBuiltinRemove() = %#v, want %#v", got, tt.expected)
			}
		})
	}
}

func TestSomeAndEveryShortCircuitBeforeLaterPredicateErrors(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected bool
	}{
		{
			name:     "some stops after true",
			expr:     `some([]interface{}{1, 0}, __lambda("x", x > 0 || 1 / x > 0))`,
			expected: true,
		},
		{
			name:     "every stops after false",
			expr:     `every([]interface{}{0, 1}, __lambda("x", x > 0 && 1 / (1 - x) > 0))`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.expr, Context{}, nil, 0, tt.expr)
			if err != nil {
				t.Fatalf("Evaluate() error = %v", err)
			}
			if result != tt.expected {
				t.Fatalf("Evaluate() = %#v, want %t", result, tt.expected)
			}
		})
	}
}

func TestInclusiveRangeBounds(t *testing.T) {
	tests := []struct {
		name          string
		start         int
		end           int
		length        int
		expectedStart int
		expectedEnd   int
		expectedOK    bool
	}{
		{
			name:          "ascending",
			start:         0,
			end:           1,
			length:        3,
			expectedStart: 0,
			expectedEnd:   1,
			expectedOK:    true,
		},
		{
			name:          "negative bounds",
			start:         -2,
			end:           -1,
			length:        3,
			expectedStart: 1,
			expectedEnd:   2,
			expectedOK:    true,
		},
		{
			name:          "descending",
			start:         -1,
			end:           0,
			length:        3,
			expectedStart: 2,
			expectedEnd:   0,
			expectedOK:    true,
		},
		{
			name:       "empty collection",
			start:      0,
			end:        0,
			length:     0,
			expectedOK: false,
		},
		{
			name:       "start beyond collection",
			start:      3,
			end:        0,
			length:     3,
			expectedOK: false,
		},
		{
			name:       "negative start beyond collection",
			start:      -4,
			end:        1,
			length:     3,
			expectedOK: false,
		},
		{
			name:       "end beyond collection",
			start:      1,
			end:        4,
			length:     3,
			expectedOK: false,
		},
		{
			name:       "negative end beyond collection",
			start:      1,
			end:        -4,
			length:     3,
			expectedOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, ok := inclusiveRangeBounds(tt.start, tt.end, tt.length)
			if start != tt.expectedStart || end != tt.expectedEnd || ok != tt.expectedOK {
				t.Fatalf(
					"inclusiveRangeBounds(%d, %d, %d) = (%d, %d, %t), want (%d, %d, %t)",
					tt.start,
					tt.end,
					tt.length,
					start,
					end,
					ok,
					tt.expectedStart,
					tt.expectedEnd,
					tt.expectedOK,
				)
			}
		})
	}
}
