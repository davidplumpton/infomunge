package evaluator

import (
	"strings"
	"testing"
)

func TestEvaluate_Arithmetic(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected Value
	}{
		{"int addition", "2 + 3", 5},
		{"int subtraction", "10 - 4", 6},
		{"int multiplication", "3 * 4", 12},
		{"exact int division", "15 / 3", 5},
		{"fractional int division", "5 / 2", 2.5},
		{"float addition", "1.5 + 2.5", 4.0},
		{"float subtraction", "5.0 - 2.0", 3.0},
		{"float multiplication", "2.0 * 3.0", 6.0},
		{"float division", "10.0 / 4.0", 2.5},
		{"int + float", "2 + 3.5", 5.5},
		{"float + int", "2.5 + 3", 5.5},
		{"string concat", `"hello" + " world"`, "hello world"},
		{"complex expression", "(2 + 3) * 4", 20},
	}

	ctx := make(Context)
	mapping := make([]int, 100)
	for i := range mapping {
		mapping[i] = i
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

func TestEvaluate_ExactNumericSemantics(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected Value
	}{
		{"large int differs from rounded float", "9007199254740993 == 9007199254740992.0", false},
		{"large int orders after rounded float", "9007199254740993 > 9007199254740992.0", true},
		{"adding decimal zero preserves large int", "9007199254740993 + 0.0", 9007199254740993},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.expr, Context{}, nil, 0, tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Fatalf("expected %v (%T), got %v (%T)", tt.expected, tt.expected, result, result)
			}
		})
	}
}

func TestEvaluate_IntegerOverflowIsExplicit(t *testing.T) {
	tests := []string{
		"9223372036854775807 + 1",
		"-9223372036854775807 - 2",
		"9223372036854775807 * 2",
		"(-9223372036854775807 - 1) / -1",
	}

	for _, expr := range tests {
		t.Run(expr, func(t *testing.T) {
			_, err := Evaluate(expr, Context{}, nil, 0, expr)
			if err == nil {
				t.Fatal("expected overflow error")
			}
			if !strings.Contains(err.Error(), "integer overflow") {
				t.Fatalf("expected integer overflow error, got %v", err)
			}
		})
	}
}

func TestEvaluate_ArithmeticErrors(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{"division by zero int", "5 / 0"},
		{"division by zero float", "5.0 / 0.0"},
	}

	ctx := make(Context)
	mapping := make([]int, 100)
	for i := range mapping {
		mapping[i] = i
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.expr, ctx, mapping, 0, tt.expr)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestEvaluate_MathFunctions(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected Value
	}{
		{"ceil positive", "ceil(2.3)", 3.0},
		{"ceil integer", "ceil(5.0)", 5.0},
		{"floor positive", "floor(2.7)", 2.0},
		{"round up", "round(2.6)", 3.0},
		{"round down", "round(2.4)", 2.0},
		{"round half even", "round(1.5)", 2.0},
		{"sqrt", "sqrt(16.0)", 4.0},
		{"sqrt small", "sqrt(0.25)", 0.5},
		{"abs positive", "abs(5.0)", 5.0},
		{"abs zero", "abs(0.0)", 0.0},
		{"max two", "max(3, 7)", 7.0},
		{"max multiple", "max(1, 5, 3, 9, 2)", 9.0},
		{"min two", "min(3, 7)", 3.0},
		{"min multiple", "min(5, 2, 8, 1, 9)", 1.0},
		{"pow", "pow(2, 8)", 256.0},
		{"pow fractional", "pow(4, 0.5)", 2.0},
		{"pow with ints", "pow(3, 2)", 9.0},
	}

	ctx := make(Context)
	mapping := make([]int, 100)
	for i := range mapping {
		mapping[i] = i
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

func TestEvaluate_MathFunctionsErrors(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{"ceil no args", "ceil()"},
		{"ceil too many args", "ceil(1, 2)"},
		{"floor no args", "floor()"},
		{"sqrt negative", "sqrt(-1)"},
		{"max no args", "max()"},
		{"min no args", "min()"},
		{"pow one arg", "pow(2)"},
		{"pow three args", "pow(2, 3, 4)"},
	}

	ctx := make(Context)
	mapping := make([]int, 100)
	for i := range mapping {
		mapping[i] = i
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.expr, ctx, mapping, 0, tt.expr)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestEvaluate_MathEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected Value
	}{
		{"mod 10 3", "mod(10, 3)", 1.0},
		{"mod 7 2", "mod(7, 2)", 1.0},
		{"isEven 4", "isEven(4)", true},
		{"isEven 5", "isEven(5)", false},
		{"isOdd 5", "isOdd(5)", true},
		{"isOdd 4", "isOdd(4)", false},
	}

	ctx := make(Context)
	mapping := make([]int, 100)
	for i := range mapping {
		mapping[i] = i
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.expr, ctx, mapping, 0, tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluate_LogicalOperations(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected Value
	}{
		{"and true true", "true && true", true},
		{"and true false", "true && false", false},
		{"and false true", "false && true", false},
		{"and false false", "false && false", false},
		{"equality int", "5 == 5", true},
		{"equality string", `"a" == "a"`, true},
		{"inequality", "5 == 6", false},
	}

	ctx := make(Context)
	mapping := make([]int, 100)
	for i := range mapping {
		mapping[i] = i
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.expr, ctx, mapping, 0, tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluate_StringConcatWithLambdaIsDeterministic(t *testing.T) {
	ctx := make(Context)
	tests := []struct {
		name     string
		expr     string
		expected string
	}{
		{name: "string plus lambda", expr: `"prefix-" + __lambda("x", x + 1)`, expected: "prefix-<function>"},
		{name: "lambda plus string", expr: `__lambda("x", x + 1) + "-suffix"`, expected: "<function>-suffix"},
	}

	for _, tt := range tests {
		mapping := make([]int, len(tt.expr))
		for i := range mapping {
			mapping[i] = i
		}

		first, err := Evaluate(tt.expr, ctx, mapping, 0, tt.expr)
		if err != nil {
			t.Fatalf("%s: first evaluation failed: %v", tt.name, err)
		}
		second, err := Evaluate(tt.expr, ctx, mapping, 0, tt.expr)
		if err != nil {
			t.Fatalf("%s: second evaluation failed: %v", tt.name, err)
		}

		firstStr, ok := first.(string)
		if !ok {
			t.Fatalf("%s: expected string result, got %#v (%T)", tt.name, first, first)
		}
		secondStr, ok := second.(string)
		if !ok {
			t.Fatalf("%s: expected string result, got %#v (%T)", tt.name, second, second)
		}
		if firstStr != secondStr {
			t.Fatalf("%s: expected deterministic concat result, got %q and %q", tt.name, firstStr, secondStr)
		}
		if firstStr != tt.expected {
			t.Fatalf("%s: expected %q, got %q", tt.name, tt.expected, firstStr)
		}
	}
}

func TestCoerceToString_ArrayContainingLambdaIsDeterministic(t *testing.T) {
	value := Array{
		&Lambda{
			Params: []ParamDef{{Name: "x"}},
			Body:   "x + 1",
		},
		2,
	}

	first := coerceToString(value)
	second := coerceToString(value)

	if first != second {
		t.Fatalf("expected deterministic stringification, got %q and %q", first, second)
	}
	if first != "[(lambda: [x] => x + 1) 2]" {
		t.Fatalf("expected nested lambda stringification to be stable, got %q", first)
	}
}

func TestEvaluate_LambdaStringRepresentationHandlesUnaryBody(t *testing.T) {
	ctx := make(Context)
	expr := `__lambda("left, y", -6)`
	mapping := make([]int, len(expr))
	for i := range mapping {
		mapping[i] = i
	}

	result, err := Evaluate(expr, ctx, mapping, 0, expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lambda, ok := result.(*Lambda)
	if !ok {
		t.Fatalf("expected lambda result, got %#v (%T)", result, result)
	}
	if lambda.String() != "(lambda: [left, y] => -6)" {
		t.Fatalf("expected stable unary lambda body, got %q", lambda.String())
	}
}

func TestEvaluate_ComparisonErrors(t *testing.T) {
	ctx := Object{
		"str": "hello",
		"arr": Array{1, 2, 3},
	}
	mapping := make([]int, 100)
	for i := range mapping {
		mapping[i] = i
	}

	tests := []struct {
		name string
		expr string
	}{
		{"compare string with <", `"a" < "b"`},
		{"compare array with >", "arr > 5"},
		{"compare different types", `"hello" < 5`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.expr, ctx, mapping, 0, tt.expr)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}
