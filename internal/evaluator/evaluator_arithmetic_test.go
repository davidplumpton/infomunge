package evaluator

import (
	"regexp"
	"strings"
	"testing"

	"infomunge/pkg/values"
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
		{"int modulo", "5 % 2", 1},
		{"negative int modulo", "-5 % 2", -1},
		{"float modulo", "5.5 % 2", 1.5},
		{"mixed modulo", "10 % 4.0", 2.0},
		{"int + float", "2 + 3.5", 5.5},
		{"float + int", "2.5 + 3", 5.5},
		{"int + numeric string", `2 + "3"`, 5},
		{"numeric string + int", `"2" + 3`, 5},
		{"float + numeric string", `0.08003611321817175 + "1"`, 1.0800361132181717},
		{"numeric string + float", `"1" + 0.08003611321817175`, 1.0800361132181717},
		{"negate numeric string", `-"2.5"`, -2.5},
		{"numeric string minus int", `"2" - 1`, 1},
		{"int minus numeric string", `1 - "2"`, -1},
		{"numeric string times int", `"2" * 3`, 6},
		{"int times numeric string", `3 * "2"`, 6},
		{"numeric string divided by int", `"6" / 2`, 3},
		{"int divided by numeric string", `6 / "2"`, 3},
		{"numeric string percent modulo", `"7" % 3`, 1},
		{"numeric string keyword modulo implementation", `mod("7", 3)`, 1},
		{"string concat", `"hello" + " world"`, "hello world"},
		{"nonnumeric string + number", `"Count: " + 42`, "Count: 42"},
		{"number + nonnumeric string", `42 + " items"`, "42 items"},
		{"complex expression", "(2 + 3) * 4", 20},
		{"modulo precedence", "2 + 5 % 3 * 4", 10},
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

func TestArithmeticNumericStringsPreserveExactIntegers(t *testing.T) {
	tests := []struct {
		name      string
		operation func(Value, Value) (Value, error)
		left      Value
		right     Value
		want      string
	}{
		{
			name:      "addition",
			operation: add,
			left:      "9223372036854775808",
			right:     1,
			want:      "9223372036854775809",
		},
		{
			name:      "subtraction",
			operation: sub,
			left:      "9223372036854775808",
			right:     1,
			want:      "9223372036854775807",
		},
		{
			name:      "multiplication",
			operation: mul,
			left:      "9223372036854775808",
			right:     2,
			want:      "18446744073709551616",
		},
		{
			name:      "division",
			operation: quo,
			left:      "18446744073709551616",
			right:     2,
			want:      "9223372036854775808",
		},
		{
			name:      "modulo",
			operation: rem,
			left:      "18446744073709551617",
			right:     2,
			want:      "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.operation(tt.left, tt.right)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotText := coerceToString(got); gotText != tt.want {
				t.Fatalf("result = %s (%T), want %s", gotText, got, tt.want)
			}
		})
	}

	got, err := negate("9223372036854775808")
	if err != nil {
		t.Fatalf("negate returned unexpected error: %v", err)
	}
	if gotText := coerceToString(got); gotText != "-9223372036854775808" {
		t.Fatalf("negate result = %s (%T), want -9223372036854775808", gotText, got)
	}
}

func TestEvaluate_ObjectSubtractionUsesScalarKeys(t *testing.T) {
	tests := []struct {
		name     string
		source   Object
		right    Value
		expected Object
	}{
		{
			name:     "string key",
			source:   Object{"remove": 1, "keep": 2},
			right:    "remove",
			expected: Object{"keep": 2},
		},
		{
			name:     "integer key",
			source:   Object{"1": "remove", "keep": 2},
			right:    1,
			expected: Object{"keep": 2},
		},
		{
			name:     "integral float key",
			source:   Object{"0": "remove", "keep": 2},
			right:    0.0,
			expected: Object{"keep": 2},
		},
		{
			name:     "decimal key",
			source:   Object{"1.5": "remove", "keep": 2},
			right:    1.5,
			expected: Object{"keep": 2},
		},
		{
			name:     "boolean key",
			source:   Object{"true": "remove", "keep": 2},
			right:    true,
			expected: Object{"keep": 2},
		},
		{
			name:     "regex key",
			source:   Object{"remove": 1, "keep": 2},
			right:    &Regex{Pattern: "remove", Re: regexp.MustCompile("remove")},
			expected: Object{"keep": 2},
		},
		{
			name:     "namespace key",
			source:   Object{"https://example.com/ns": "remove", "keep": 2},
			right:    Namespace{Prefix: "ns", URI: "https://example.com/ns"},
			expected: Object{"keep": 2},
		},
		{
			name:     "type key",
			source:   Object{"Currency": "remove", "keep": 2},
			right:    &TypeDef{Name: "Currency", BaseType: "String"},
			expected: Object{"keep": 2},
		},
		{
			name:     "binary key",
			source:   Object{"remove": 1, "keep": 2},
			right:    []byte("remove"),
			expected: Object{"keep": 2},
		},
		{
			name:     "missing coerced key",
			source:   Object{"keep": 2},
			right:    false,
			expected: Object{"keep": 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := values.CloneObject(tt.source)
			result, err := sub(tt.source, tt.right)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !isEqual(result, tt.expected) {
				t.Fatalf("expected %#v, got %#v", tt.expected, result)
			}
			if !isEqual(tt.source, original) {
				t.Fatal("object subtraction mutated the source object")
			}
		})
	}
}

func TestArraySubtractionRemovesStructurallyEqualValues(t *testing.T) {
	tests := []struct {
		name     string
		source   Array
		right    Value
		expected Array
	}{
		{
			name:     "repeated scalar values",
			source:   Array{1, 2, 1},
			right:    1,
			expected: Array{2},
		},
		{
			name:     "absent value",
			source:   Array{1, 2},
			right:    97,
			expected: Array{1, 2},
		},
		{
			name:     "null values",
			source:   Array{1, nil, 2, nil},
			right:    nil,
			expected: Array{1, 2},
		},
		{
			name:     "nested arrays",
			source:   Array{Array{1}, Array{2}, Array{1}},
			right:    Array{1},
			expected: Array{Array{2}},
		},
		{
			name: "nested objects",
			source: Array{
				Object{"id": 1, "nested": Array{"x"}},
				Object{"id": 2},
				Object{"nested": Array{"x"}, "id": 1},
			},
			right:    Object{"nested": Array{"x"}, "id": 1},
			expected: Array{Object{"id": 2}},
		},
		{
			name:     "negative absent value",
			source:   Array{1, 2},
			right:    -97,
			expected: Array{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := append(Array(nil), tt.source...)
			result, err := sub(tt.source, tt.right)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !isEqual(result, tt.expected) {
				t.Fatalf("expected %#v, got %#v", tt.expected, result)
			}
			if !isEqual(tt.source, original) {
				t.Fatal("array subtraction mutated the source array")
			}
		})
	}
}

func TestObjectSubtractionRejectsNonKeyOperands(t *testing.T) {
	tests := []struct {
		name  string
		right Value
	}{
		{name: "null", right: nil},
		{name: "array", right: Array{"remove"}},
		{name: "object", right: Object{"remove": true}},
		{name: "function", right: &Lambda{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := sub(Object{"remove": 1}, tt.right); err == nil {
				t.Fatal("expected object subtraction to reject the right operand")
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
		{"large int modulo decimal preserves precision", "9007199254740993 % 2.0", 1},
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
		{"modulo by zero int", "5 % 0"},
		{"modulo by zero float", "5.0 % 0.0"},
		{"modulo rejects non-number", "true % 2"},
		{"negation rejects nonnumeric string", `-"two"`},
		{"subtraction rejects nonnumeric string", `"two" - 1`},
		{"multiplication rejects nonnumeric string", `"two" * 1`},
		{"division rejects nonnumeric string", `"two" / 1`},
		{"modulo rejects nonnumeric string", `"two" % 1`},
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
		{"max two", "max(3, 7)", 7},
		{"max multiple", "max(1, 5, 3, 9, 2)", 9},
		{"min two", "min(3, 7)", 3},
		{"min multiple", "min(5, 2, 8, 1, 9)", 1},
		{"pow", "pow(2, 8)", 256},
		{"pow fractional", "pow(4, 0.5)", 2.0},
		{"pow with ints", "pow(3, 2)", 9},
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
		{"mod 10 3", "mod(10, 3)", 1},
		{"mod 7 2", "mod(7, 2)", 1},
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

func TestEvaluate_LogicalOperationsShortCircuit(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected Value
	}{
		{"false AND skips right operand", "false && (1 / 0 == 0)", false},
		{"true OR skips right operand", "true || (1 / 0 == 0)", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.expr, Context{}, nil, 0, tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluate_LogicalOperationsEvaluateRequiredOperands(t *testing.T) {
	tests := []struct {
		name        string
		expr        string
		errorSubstr string
	}{
		{"true AND evaluates right operand", "true && (1 / 0 == 0)", "division by zero"},
		{"false OR evaluates right operand", "false || (1 / 0 == 0)", "division by zero"},
		{"AND validates evaluated left operand", "1 && missing", "logical AND requires booleans"},
		{"OR validates evaluated right operand", "false || 1", "logical OR requires booleans"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.expr, Context{}, nil, 0, tt.expr)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errorSubstr) {
				t.Fatalf("expected error containing %q, got %v", tt.errorSubstr, err)
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

func TestEvaluate_RelationalComparisonsCoerceNumericStrings(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want bool
	}{
		{name: "int less than numeric string", expr: `1 < "2"`, want: true},
		{name: "numeric string greater than int", expr: `"2" > 1`, want: true},
		{name: "int not less than numeric string", expr: `3 < "2"`, want: false},
		{name: "int less than or equal to numeric string", expr: `3 <= "3"`, want: true},
		{name: "numeric string greater than or equal to int", expr: `"3" >= 3`, want: true},
		{name: "float less than numeric string", expr: `1.5 < "2.25"`, want: true},
		{
			name: "big integer numeric string preserves exact ordering",
			expr: `"9223372036854775808" > 9223372036854775807`,
			want: true,
		},
		{
			name: "big integers preserve exact ordering in either operand order",
			expr: `9223372036854775808 < "9223372036854775809"`,
			want: true,
		},
		{
			name: "integer string and float compare without rounding the integer",
			expr: `"9007199254740993" > 9007199254740992.0`,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := make([]int, len(tt.expr))
			for i := range mapping {
				mapping[i] = i
			}

			got, err := Evaluate(tt.expr, Context{}, mapping, 0, tt.expr)
			if err != nil {
				t.Fatalf("Evaluate(%q) returned an unexpected error: %v", tt.expr, err)
			}
			if got != tt.want {
				t.Fatalf("Evaluate(%q) returned %#v, want %v", tt.expr, got, tt.want)
			}
		})
	}
}

func TestEvaluate_RelationalComparisonRejectsNonnumericString(t *testing.T) {
	expr := `1 < "abc"`
	mapping := make([]int, len(expr))
	for i := range mapping {
		mapping[i] = i
	}

	_, err := Evaluate(expr, Context{}, mapping, 0, expr)
	if err == nil {
		t.Fatal("Evaluate returned nil error for a nonnumeric string comparison")
	}
	if got, want := err.Error(), "EvalError: cannot compare int and string with <"; got != want {
		t.Fatalf("Evaluate returned error %q, want %q", got, want)
	}
}
