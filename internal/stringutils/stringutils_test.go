package stringutils

import (
	"testing"
)

func TestReplaceOperatorsOutsideStrings(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		replacements []OperatorReplacement
		expected     string
	}{
		{
			name:  "simple replacement",
			input: "a and b",
			replacements: []OperatorReplacement{
				{Pattern: " and ", Replacement: []rune(" && ")},
			},
			expected: "a && b",
		},
		{
			name:  "multiple replacements",
			input: "a and b or c",
			replacements: []OperatorReplacement{
				{Pattern: " and ", Replacement: []rune(" && ")},
				{Pattern: " or ", Replacement: []rune(" || ")},
			},
			expected: "a && b || c",
		},
		{
			name:  "no replacement inside string",
			input: `"a and b" or c`,
			replacements: []OperatorReplacement{
				{Pattern: " and ", Replacement: []rune(" && ")},
				{Pattern: " or ", Replacement: []rune(" || ")},
			},
			expected: `"a and b" || c`,
		},
		{
			name:  "escaped quote in string",
			input: `"a \"and\" b" and c`,
			replacements: []OperatorReplacement{
				{Pattern: " and ", Replacement: []rune(" && ")},
			},
			expected: `"a \"and\" b" && c`,
		},
		{
			name:  "no matches",
			input: "a + b",
			replacements: []OperatorReplacement{
				{Pattern: " and ", Replacement: []rune(" && ")},
			},
			expected: "a + b",
		},
		{
			name:         "empty input",
			input:        "",
			replacements: []OperatorReplacement{{Pattern: " and ", Replacement: []rune(" && ")}},
			expected:     "",
		},
		{
			name:  "pattern at start",
			input: " and b",
			replacements: []OperatorReplacement{
				{Pattern: " and ", Replacement: []rune(" && ")},
			},
			expected: " && b",
		},
		{
			name:  "pattern at end",
			input: "a and ",
			replacements: []OperatorReplacement{
				{Pattern: " and ", Replacement: []rune(" && ")},
			},
			expected: "a && ",
		},
		{
			name:  "multiple occurrences",
			input: "a and b and c and d",
			replacements: []OperatorReplacement{
				{Pattern: " and ", Replacement: []rune(" && ")},
			},
			expected: "a && b && c && d",
		},
		{
			name:  "string with multiple quotes",
			input: `"first" and "second" and "third"`,
			replacements: []OperatorReplacement{
				{Pattern: " and ", Replacement: []rune(" && ")},
			},
			expected: `"first" && "second" && "third"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceOperatorsOutsideStrings(tt.input, tt.replacements)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestReplaceBinaryOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		cfg      BinaryOperatorConfig
		expected string
	}{
		{
			name:  "simple default operator",
			input: "x default 0",
			cfg: BinaryOperatorConfig{
				Operator: " default ",
				FuncName: "__default",
			},
			expected: "__default(x, 0)",
		},
		{
			name:  "operator with function call left",
			input: "getValue() default 0",
			cfg: BinaryOperatorConfig{
				Operator: " default ",
				FuncName: "__default",
			},
			expected: "__default(getValue(), 0)",
		},
		{
			name:  "operator with nested expression",
			input: "obj.field default fallback",
			cfg: BinaryOperatorConfig{
				Operator: " default ",
				FuncName: "__default",
			},
			expected: "__default(obj.field, fallback)",
		},
		{
			name:  "operator not replaced inside string",
			input: `"x default y"`,
			cfg: BinaryOperatorConfig{
				Operator: " default ",
				FuncName: "__default",
			},
			expected: `"x default y"`,
		},
		{
			name:  "right operand with parentheses",
			input: "x default (a + b)",
			cfg: BinaryOperatorConfig{
				Operator: " default ",
				FuncName: "__default",
			},
			expected: "__default(x, (a + b))",
		},
		{
			name:  "right operand stops at comma",
			input: "x default 0, y",
			cfg: BinaryOperatorConfig{
				Operator: " default ",
				FuncName: "__default",
			},
			expected: "__default(x, 0), y",
		},
		{
			name:  "right operand stops at closing bracket",
			input: "(x default 0)",
			cfg: BinaryOperatorConfig{
				Operator: " default ",
				FuncName: "__default",
			},
			expected: "(__default(x, 0))",
		},
		{
			name:  "chained operators",
			input: "a default b default c",
			cfg: BinaryOperatorConfig{
				Operator: " default ",
				FuncName: "__default",
			},
			// Single pass: first default is replaced, second becomes part of right operand
			expected: "__default(a, b default c)",
		},
		{
			name:  "with right stop operator",
			input: "x default y and z",
			cfg: BinaryOperatorConfig{
				Operator:     " default ",
				FuncName:     "__default",
				RightStopOps: []string{" and "},
			},
			expected: "__default(x, y) and z",
		},
		{
			name:  "with extra stops",
			input: "a ~ b ~ c",
			cfg: BinaryOperatorConfig{
				Operator:     " ~ ",
				FuncName:     "__update",
				ExtraStops:   []rune{'~'},
				RightStopOps: []string{" ~ "},
			},
			expected: "__update(__update(a, b), c)",
		},
		{
			name:  "left operand with array access",
			input: "arr[0] default 0",
			cfg: BinaryOperatorConfig{
				Operator: " default ",
				FuncName: "__default",
			},
			expected: "__default(arr[0], 0)",
		},
		{
			name:  "complex nested expression",
			input: "foo(bar[x]) default baz()",
			cfg: BinaryOperatorConfig{
				Operator: " default ",
				FuncName: "__default",
			},
			expected: "__default(foo(bar[x]), baz())",
		},
		{
			name:  "operator at start with no left operand",
			input: " default x",
			cfg: BinaryOperatorConfig{
				Operator: " default ",
				FuncName: "__default",
			},
			expected: " default x",
		},
		{
			name:  "string in right operand",
			input: `x default "fallback"`,
			cfg: BinaryOperatorConfig{
				Operator: " default ",
				FuncName: "__default",
			},
			expected: `__default(x, "fallback")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceBinaryOperator(tt.input, tt.cfg)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestFindLeftOperandStart(t *testing.T) {
	tests := []struct {
		name       string
		input      []rune
		extraStops []rune
		expected   int
	}{
		{
			name:     "simple identifier",
			input:    []rune("foo"),
			expected: 0,
		},
		{
			name:     "after operator",
			input:    []rune("a + b"),
			expected: 3,
		},
		{
			name:     "after comma",
			input:    []rune("a, b"),
			expected: 2, // Returns position right after comma, caller trims whitespace
		},
		{
			name:     "with parentheses",
			input:    []rune("a + (b + c)"),
			expected: 3,
		},
		{
			name:     "function call",
			input:    []rune("foo()"),
			expected: 0,
		},
		{
			name:     "nested function call",
			input:    []rune("a + foo(bar())"),
			expected: 3,
		},
		{
			name:     "array access",
			input:    []rune("a + arr[0]"),
			expected: 3,
		},
		{
			name:     "nested brackets",
			input:    []rune("arr[obj[key]]"),
			expected: 0,
		},
		{
			name:     "trailing whitespace",
			input:    []rune("foo   "),
			expected: 0,
		},
		{
			name:     "empty input",
			input:    []rune(""),
			expected: 0,
		},
		{
			name:     "only whitespace",
			input:    []rune("   "),
			expected: 0,
		},
		{
			name:     "after equals",
			input:    []rune("x = y"),
			expected: 3,
		},
		{
			name:     "after colon",
			input:    []rune("key: value"),
			expected: 4, // Returns position right after colon, caller trims whitespace
		},
		{
			name:       "with extra stop",
			input:      []rune("a ~ b"),
			extraStops: []rune{'~'},
			expected:   3,
		},
		{
			name:     "string literal",
			input:    []rune(`a + "hello"`),
			expected: 3,
		},
		{
			name:     "string with escaped quote",
			input:    []rune(`a + "say \"hi\""`),
			expected: 3,
		},
		{
			name:     "after opening paren",
			input:    []rune("(x"),
			expected: 1,
		},
		{
			name:     "complex expression",
			input:    []rune("foo(a, b) + bar[x]"),
			expected: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindLeftOperandStart(tt.input, tt.extraStops)
			if got != tt.expected {
				t.Errorf("expected %d, got %d (input: %q)", tt.expected, got, string(tt.input))
			}
		})
	}
}

func TestFindLeftOperandStartWithStops(t *testing.T) {
	tests := []struct {
		name     string
		input    []rune
		stops    []rune
		expected int
	}{
		{
			name:     "minimal stops - includes operators in operand",
			input:    []rune("a + b"),
			stops:    MinimalStops,
			expected: 0, // '+' is not a stop, so whole expression is the operand
		},
		{
			name:     "minimal stops - stops at comma",
			input:    []rune("a, b + c"),
			stops:    MinimalStops,
			expected: 2,
		},
		{
			name:     "minimal stops - stops at semicolon",
			input:    []rune("a; b"),
			stops:    MinimalStops,
			expected: 2,
		},
		{
			name:     "minimal stops - stops at opening bracket",
			input:    []rune("(a + b"),
			stops:    MinimalStops,
			expected: 1,
		},
		{
			name:     "default stops - stops at operator",
			input:    []rune("a + b"),
			stops:    DefaultOperatorStops,
			expected: 3,
		},
		{
			name:     "custom stops",
			input:    []rune("a | b"),
			stops:    []rune{'|'},
			expected: 3,
		},
		{
			name:     "respects nesting with minimal stops",
			input:    []rune("foo(a + b)"),
			stops:    MinimalStops,
			expected: 0,
		},
		{
			name:     "handles strings with minimal stops",
			input:    []rune(`"a, b" + c`),
			stops:    MinimalStops,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindLeftOperandStartWithStops(tt.input, tt.stops)
			if got != tt.expected {
				t.Errorf("expected %d, got %d (input: %q)", tt.expected, got, string(tt.input))
			}
		})
	}
}
