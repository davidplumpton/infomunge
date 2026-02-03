package preprocessor

import (
	"testing"
)

func TestRegexEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Division not regex",
			input:    "x / y / z",
			expected: "x / y / z",
		},
		{
			name:     "Regex in array",
			input:    "[ /abc/, /def/ ]",
			expected: "[]interface{}{ regex(\"abc\"), regex(\"def\") ,}",
		},
		{
			name:     "Regex in object",
			input:    "{ a: /abc/ }",
			expected: "map[string]interface{}{ \"a\": regex(\"abc\") ,}",
		},
		{
			name:     "Regex after comma",
			input:    "foo(a, /abc/)",
			expected: "foo(a, regex(\"abc\"))",
		},
		{
			name:     "Regex after operator",
			input:    "x == /abc/",
			expected: "x == regex(\"abc\")",
		},
		{
			name:     "Division after closing paren",
			input:    "(x + y) / z",
			expected: "(x + y) / z",
		},
		{
			name:     "Division after closing bracket",
			input:    "[1, 2, 3][0] / 2",
			expected: "[]interface{}{1, 2, 3,}[0] / 2",
		},
		{
			name:     "Regex after if",
			input:    "if /abc/ then 1 else 2",
			expected: "then(if regex(\"abc\"), 1 else 2)",
		},
		{
			name:     "Division with comments",
			input:    "x / // comment\ny",
			expected: "__seq(x /, y)",
		},
		{
			name:     "Multi-line division",
			input:    "x\n/ y",
			expected: "__seq(x, / y)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, err := PrepareForParsing(tt.input, Options{})
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("\nInput: %q\nExpected: %q\nGot: %q", tt.input, tt.expected, result)
			}
		})
	}
}
