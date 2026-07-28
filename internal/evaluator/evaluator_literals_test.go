package evaluator

import (
	"math/big"
	"strconv"
	"strings"
	"testing"

	unifiederrors "infomunge/internal/errors"
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

func TestEvaluate_MinimumSignedIntegerLiteral(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("acceptance value is specific to 64-bit int builds")
	}

	const expr = "-9223372036854775808"
	result, err := Evaluate(expr, Context{}, nil, 0, expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != minInt() {
		t.Fatalf("expected %d (%T), got %v (%T)", minInt(), minInt(), result, result)
	}
}

func TestEvaluate_MinimumSignedIntegerNestedNegation(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("acceptance value is specific to 64-bit int builds")
	}

	tests := []struct {
		name string
		expr string
		want string
	}{
		{
			name: "double negation produces exact positive magnitude",
			expr: "-(-9223372036854775808)",
			want: "9223372036854775808",
		},
		{
			name: "triple negation produces minimum integer",
			expr: "-(-(-9223372036854775808))",
			want: "-9223372036854775808",
		},
		{
			name: "four negations produce exact positive magnitude",
			expr: "-(-(-(-9223372036854775808)))",
			want: "9223372036854775808",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.expr, Context{}, nil, 0, tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			switch value := result.(type) {
			case int:
				if strconv.Itoa(value) != tt.want {
					t.Fatalf("expected %s, got %d", tt.want, value)
				}
			case *big.Int:
				if value.String() != tt.want {
					t.Fatalf("expected %s, got %s", tt.want, value)
				}
			default:
				t.Fatalf("expected exact integer result %s, got %v (%T)", tt.want, result, result)
			}
		})
	}
}

func TestEvaluate_MinimumSignedIntegerLiteralRangeSafety(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("acceptance values are specific to 64-bit int builds")
	}

	tests := []struct {
		name string
		expr string
		want string
	}{
		{
			name: "computed minimum negation remains overflow",
			expr: "-(-9223372036854775807 - 1)",
			want: "integer overflow during negation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.expr, Context{}, nil, 0, tt.expr)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestEvaluate_OutOfRangeIntegerLiteralsRemainExact(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("acceptance values are specific to 64-bit int builds")
	}

	tests := []struct {
		name string
		expr string
		want string
	}{
		{name: "literal", expr: "9223372036854775808", want: "9223372036854775808"},
		{name: "addition", expr: "9223372036854775808 + 1", want: "9223372036854775809"},
		{name: "subtraction normalizes in-range result", expr: "9223372036854775808 - 1", want: "9223372036854775807"},
		{name: "multiplication", expr: "9223372036854775808 * 2", want: "18446744073709551616"},
		{name: "division", expr: "9223372036854775808 / 2", want: "4611686018427387904"},
		{name: "modulo", expr: "9223372036854775808 % 3", want: "2"},
		{name: "negative literal", expr: "-9223372036854775809", want: "-9223372036854775809"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.expr, Context{}, nil, 0, tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			switch value := result.(type) {
			case int:
				if strconv.Itoa(value) != tt.want {
					t.Fatalf("expected %s, got %d", tt.want, value)
				}
			case *big.Int:
				if value.String() != tt.want {
					t.Fatalf("expected %s, got %s", tt.want, value)
				}
			default:
				t.Fatalf("expected exact integer result %s, got %v (%T)", tt.want, result, result)
			}
		})
	}

	result, err := Evaluate("9223372036854775808 > 9223372036854775807", Context{}, nil, 0, "")
	if err != nil {
		t.Fatalf("comparison returned an error: %v", err)
	}
	if result != true {
		t.Fatalf("expected out-of-range literal to compare greater, got %v", result)
	}

	classificationTests := []struct {
		expr string
		want Value
	}{
		{expr: "typeOf(9223372036854775808)", want: TypeValue("Number")},
		{expr: "sizeOf(9223372036854775808)", want: 19},
		{expr: "isInteger(9223372036854775808)", want: true},
		{expr: "isEven(9223372036854775808)", want: true},
	}
	for _, tt := range classificationTests {
		result, err := Evaluate(tt.expr, Context{}, nil, 0, tt.expr)
		if err != nil {
			t.Fatalf("%s returned an error: %v", tt.expr, err)
		}
		if result != tt.want {
			t.Fatalf("%s: expected %v (%T), got %v (%T)", tt.expr, tt.want, tt.want, result, result)
		}
	}
}

func TestEvaluate_CompositeLiterals(t *testing.T) {
	ctx := make(Context)
	mapping := make([]int, 200)
	for i := range mapping {
		mapping[i] = i
	}

	t.Run("map literal", func(t *testing.T) {
		result, err := Evaluate(`map[string]interface{}{a: 1}`, ctx, mapping, 0, `map[string]interface{}{a: 1}`)
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

	t.Run("dynamic map key", func(t *testing.T) {
		ctx := Context{"key": "answer"}
		expr := `map[string]interface{}{(key): 42}`
		result, err := Evaluate(expr, ctx, mapping, 0, expr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		m, ok := result.(Object)
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}
		if m["answer"] != 42 {
			t.Errorf("expected m[answer]=42, got %v", m["answer"])
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

func TestEvaluate_IdentifierResolution(t *testing.T) {
	t.Run("declared variable", func(t *testing.T) {
		result, err := Evaluate("answer", Context{"answer": 42}, []int{0}, 0, "answer")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != 42 {
			t.Fatalf("expected 42, got %v", result)
		}
	})

	t.Run("builtin call", func(t *testing.T) {
		expr := `sizeOf([]interface{}{1, 2})`
		mapping := make([]int, len(expr))
		for i := range mapping {
			mapping[i] = i
		}
		result, err := Evaluate(expr, nil, mapping, 0, expr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != 2 {
			t.Fatalf("expected 2, got %v", result)
		}
	})

	for _, expr := range []string{
		"missing",
		"missing + 1",
		`missing == "missing"`,
	} {
		t.Run("unresolved reference in "+expr, func(t *testing.T) {
			mapping := make([]int, len(expr))
			for i := range mapping {
				mapping[i] = i
			}
			_, err := Evaluate(expr, nil, mapping, 0, expr)
			if err == nil {
				t.Fatal("expected unresolved reference error")
			}
			if !strings.Contains(err.Error(), "unresolved reference: missing") {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(err.Error(), "1:1") {
				t.Fatalf("expected positional error, got: %v", err)
			}
		})
	}
}

func TestEvaluate_BuiltinErrorUsesMappedGeneratedPosition(t *testing.T) {
	expr := "\n  sqrt(-1)"
	mapping := make([]int, len(expr))
	for i := range mapping {
		mapping[i] = i
	}

	_, err := Evaluate(expr, nil, mapping, 0, expr)
	if err == nil {
		t.Fatal("expected sqrt error")
	}
	if got, want := err.Error(), "2:3: sqrt: cannot take square root of negative number -1"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}

	rawErr := newPosError("representative builtin failure", 4)
	typedErr, ok := rawErr.(*unifiederrors.Error)
	if !ok {
		t.Fatalf("newPosError() type = %T, want *errors.Error", rawErr)
	}
	if typedErr.Type != unifiederrors.TypeEval {
		t.Fatalf("newPosError().Type = %q, want %q", typedErr.Type, unifiederrors.TypeEval)
	}
	if typedErr.Message != "representative builtin failure" {
		t.Fatalf("newPosError().Message = %q, want %q", typedErr.Message, "representative builtin failure")
	}
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
