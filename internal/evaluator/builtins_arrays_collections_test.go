package evaluator

import (
	"go/ast"
	"testing"
)

func TestBuiltinDeepReturnsNullWithoutMatches(t *testing.T) {
	got, err := callBuiltinDeep(
		[]Value{Array{Object{"other": 1}, 2}, "score"},
		&ast.CallExpr{Fun: &ast.Ident{Name: "__deep"}},
	)
	if err != nil {
		t.Fatalf("callBuiltinDeep() error = %v", err)
	}
	if got != nil {
		t.Fatalf("callBuiltinDeep() = %#v, want nil", got)
	}
}

func TestBuiltinDeepTraversesNestedArraysAndObjectsInOrder(t *testing.T) {
	got, err := callBuiltinDeep(
		[]Value{
			Array{
				Object{"score": 1},
				Object{"child": Object{"score": 2}},
				Array{Object{"score": 3}},
			},
			"score",
		},
		&ast.CallExpr{Fun: &ast.Ident{Name: "__deep"}},
	)
	if err != nil {
		t.Fatalf("callBuiltinDeep() error = %v", err)
	}
	want := Array{1, 2, 3}
	if !numericEquals(got, want) {
		t.Fatalf("callBuiltinDeep() = %#v, want %#v", got, want)
	}
}

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

func TestNullArrayHelpersMatchDataWeaveIdentitiesWithoutEvaluatingCallbacks(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected Value
	}{
		{name: "takeWhile propagates null", expr: `takeWhile(nil, 1 / 0)`, expected: nil},
		{name: "dropWhile propagates null", expr: `dropWhile(nil, 1 / 0)`, expected: nil},
		{name: "some uses false identity", expr: `some(nil, 1 / 0)`, expected: false},
		{name: "every uses true identity", expr: `every(nil, 1 / 0)`, expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.expr, Context{}, nil, 0, tt.expr)
			if err != nil {
				t.Fatalf("Evaluate() error = %v", err)
			}
			if result != tt.expected {
				t.Fatalf("Evaluate() = %#v, want %#v", result, tt.expected)
			}
		})
	}
}

func TestNullArrayValueHelpersPropagateNull(t *testing.T) {
	tests := []string{
		`slice(nil, 0, 1)`,
		`take(nil, 1)`,
		`drop(nil, 1)`,
	}

	for _, expr := range tests {
		t.Run(expr, func(t *testing.T) {
			result, err := Evaluate(expr, Context{}, nil, 0, expr)
			if err != nil {
				t.Fatalf("Evaluate() error = %v", err)
			}
			if result != nil {
				t.Fatalf("Evaluate() = %#v, want nil", result)
			}
		})
	}
}

func TestBuiltinKeysPropagatesNull(t *testing.T) {
	call := &ast.CallExpr{Fun: &ast.Ident{Name: "keys"}}

	result, err := callBuiltinKeys([]Value{nil}, call)
	if err != nil {
		t.Fatalf("callBuiltinKeys() error = %v", err)
	}
	if result != nil {
		t.Fatalf("callBuiltinKeys() = %#v, want nil", result)
	}
}

func TestBuiltinArraysSliceClampsDataWeaveBounds(t *testing.T) {
	tests := []struct {
		name     string
		input    Array
		start    int
		end      int
		expected Array
	}{
		{
			name:     "negative start clamps to first element",
			input:    Array{1, 2, 3, 4},
			start:    -2,
			end:      2,
			expected: Array{1, 2},
		},
		{
			name:     "negative end returns empty",
			input:    Array{1, 2, 3, 4},
			start:    1,
			end:      -1,
			expected: Array{},
		},
		{
			name:     "oversized end clamps after last element",
			input:    Array{1, 2, 3, 4},
			start:    1,
			end:      99,
			expected: Array{2, 3, 4},
		},
		{
			name:     "reversed bounds return empty",
			input:    Array{1, 2, 3, 4},
			start:    3,
			end:      1,
			expected: Array{},
		},
		{
			name:     "start after last element returns empty",
			input:    Array{1, 2, 3, 4},
			start:    4,
			end:      5,
			expected: Array{},
		},
		{
			name:     "empty input returns empty",
			input:    Array{},
			start:    -1,
			end:      1,
			expected: Array{},
		},
	}

	call := &ast.CallExpr{Fun: &ast.Ident{Name: "__arraysSlice"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := callBuiltinArraysSlice(
				[]Value{tt.input, tt.start, tt.end},
				call,
			)
			if err != nil {
				t.Fatalf("callBuiltinArraysSlice() error = %v", err)
			}
			if !numericEquals(result, tt.expected) {
				t.Fatalf("callBuiltinArraysSlice() = %#v, want %#v", result, tt.expected)
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
