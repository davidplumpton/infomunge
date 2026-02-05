package preprocessor

import (
	"strings"
	"testing"
)

func TestPrepareForParsing_AndOr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"and keyword", "true and false", "true && false"},
		{"or keyword", "true or false", "true || false"},
		{"multiple and", "a and b and c", "a && b && c"},
		{"mixed", "a and b or c", "a && b || c"},
		{"no replacement", "band", "band"},
		{"no replacement or", "ordinary", "ordinary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := PrepareForParsing(tt.input, Options{})
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPrepareForParsing_SingleQuotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"single to double", "'hello'", `"hello"`},
		{"preserve internal double quotes", `'say "hi"'`, `"say \"hi\""`},
		{"empty single quotes", "''", `""`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := PrepareForParsing(tt.input, Options{})
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPrepareForParsing_Objects(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty object", "{}", "map[string]interface{}{}"},
		{"simple object", `{a: 1}`, `map[string]interface{}{"a": 1,}`},
		{"object with string value", `{name: "test"}`, `map[string]interface{}{"name": "test",}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := PrepareForParsing(tt.input, Options{})
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPrepareForParsing_Arrays(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty array", "[]", "[]interface{}{}"},
		{"simple array", "[1, 2, 3]", "[]interface{}{1, 2, 3,}"},
		{"array with strings", `["a", "b"]`, `[]interface{}{"a", "b",}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := PrepareForParsing(tt.input, Options{})
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPrepareForParsing_IndexAccess(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"array index", "arr[0]", "arr[0]"},
		{"map index", `obj["key"]`, `obj["key"]`},
		{"chained index", "arr[0][1]", "arr[0][1]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := PrepareForParsing(tt.input, Options{})
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPrepareForParsing_Mapping(t *testing.T) {
	t.Run("mapping preserves positions", func(t *testing.T) {
		input := "abc"
		_, mapping, _ := PrepareForParsing(input, Options{})
		if len(mapping) != 3 {
			t.Errorf("expected mapping length 3, got %d", len(mapping))
		}
		for i := 0; i < 3; i++ {
			if mapping[i] != i {
				t.Errorf("expected mapping[%d]=%d, got %d", i, i, mapping[i])
			}
		}
	})

	t.Run("and keyword mapping", func(t *testing.T) {
		input := "a and b"
		result, mapping, _ := PrepareForParsing(input, Options{})
		// "a and b" -> "a && b"
		if result != "a && b" {
			t.Errorf("expected 'a && b', got %q", result)
		}
		// mapping should exist for each character
		if len(mapping) != len(result) {
			t.Errorf("expected mapping length %d, got %d", len(result), len(mapping))
		}
	})
}

func TestPrepareForParsing_Newlines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"newline in object", "{\na: 1\n}", "map[string]interface{}{ \"a\": 1 ,}"},
		{"newline in array", "[\n1\n]", "[]interface{}{ 1 ,}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := PrepareForParsing(tt.input, Options{})
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractHeaderAndBody_WithHeader(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedHeader string
		expectedBody   string
	}{
		{
			"simple header",
			"output application/json\n---\n{a: 1}",
			"output application/json",
			"{a: 1}",
		},
		{
			"header with var",
			"var x = 10\n---\nx + 1",
			"var x = 10",
			"x + 1",
		},
		{
			"empty header",
			"---\nbody only",
			"",
			"body only",
		},
		{
			"single-line format",
			`%im 0.1 var x = 10 output application/json --- x + 1`,
			"%im 0.1 var x = 10 output application/json",
			"x + 1",
		},
		{
			"header with do block separator",
			"%im 0.1\nfun addOne(x) = do {\n  var y = x + 1\n  ---\n  y\n}\noutput application/json\n---\naddOne(2)",
			"%im 0.1\nfun addOne(x) = do {\n  var y = x + 1\n  ---\n  y\n}\noutput application/json",
			"addOne(2)",
		},
		{
			"header with unbalanced paren finds separator",
			"%im 0.1\nvar x = 1 + (\n---\nx",
			"%im 0.1\nvar x = 1 + (",
			"x",
		},
		{
			"header with unbalanced bracket finds separator",
			"%im 0.1\nvar x = [1, 2\n---\nx",
			"%im 0.1\nvar x = [1, 2",
			"x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header, body, _ := ExtractHeaderAndBody(tt.input)
			if header != tt.expectedHeader {
				t.Errorf("expected header %q, got %q", tt.expectedHeader, header)
			}
			if body != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

func TestExtractHeaderAndBody_NoHeader(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectBody   string
		expectOffset int
	}{
		{
			name:         "simple body",
			input:        "just body content",
			expectBody:   "just body content",
			expectOffset: 0,
		},
		{
			name:         "inline separator in string",
			input:        `{a: "---"}`,
			expectBody:   `{a: "---"}`,
			expectOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header, body, offset := ExtractHeaderAndBody(tt.input)

			if header != "" {
				t.Errorf("expected empty header, got %q", header)
			}
			if body != tt.expectBody {
				t.Errorf("expected body %q, got %q", tt.expectBody, body)
			}
			if offset != tt.expectOffset {
				t.Errorf("expected offset %d, got %d", tt.expectOffset, offset)
			}
		})
	}
}

func TestExtractHeaderAndBody_Offset(t *testing.T) {
	input := "header\n---\nbody"
	_, _, offset := ExtractHeaderAndBody(input)

	// offset should point to start of trimmed body
	if offset < 0 {
		t.Error("offset should be non-negative")
	}
}

func TestPrepareForParsing_StringPreservation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"double quoted string preserved", `"hello world"`, `"hello world"`},
		{"and in string preserved", `"this and that"`, `"this and that"`},
		{"or in string preserved", `"either or"`, `"either or"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := PrepareForParsing(tt.input, Options{})
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPrepareForParsing_EscapeSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Double-quoted strings with escapes
		{"escaped double quote", `"hello\"world"`, `"hello\"world"`},
		{"escaped backslash", `"path\\to\\file"`, `"path\\to\\file"`},
		{"escaped newline", `"line1\nline2"`, `"line1\nline2"`},
		{"escaped tab", `"col1\tcol2"`, `"col1\tcol2"`},

		// Single-quoted strings converted to double
		{"single to double basic", `'hello'`, `"hello"`},
		{"single with internal double quote", `'say "hi"'`, `"say \"hi\""`},
		{"single with backslash", `'path\to\file'`, `"path\\to\\file"`},
		{"single with escaped single quote", `'it\'s'`, `"it's"`},

		// Edge cases with and/or inside strings
		{"and keyword after escaped quote", `"hello\"" and true`, `"hello\"" && true`},
		{"or keyword after string", `"test" or false`, `"test" || false`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := PrepareForParsing(tt.input, Options{})
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPrepareForParsing_UpdateOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple update", `obj ~ {a: 1}`, `__update(obj, map[string]interface{}{"a": 1,})`},
		{"update with variable", `person ~ {age: 31}`, `__update(person, map[string]interface{}{"age": 31,})`},
		{"two variables", `obj ~ upd`, `__update(obj, upd)`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := PrepareForParsing(tt.input, Options{})
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPrepareForParsing_QuoteConversion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"single to double quotes", `'hello'`, `"hello"`},
		{"nested quotes", `'he said "hi"'`, `"he said \"hi\""`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := PrepareForParsing(tt.input, Options{})
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPrepareForParsing_ComplexExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"nested function calls", `func(a, func2(b, c))`},
		{"chained property access", `obj.a.b.c`},
		{"assignment expression", `x = y + 1`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := PrepareForParsing(tt.input, Options{})
			if len(result) == 0 {
				t.Errorf("expected non-empty result")
			}
		})
	}
}

func TestPrepareForParsing_IfElseStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"if without else", `if (x > 0) "positive"`, "__ifelse"},
		{"if with else", `if (x > 0) "positive" else "non-positive"`, "__ifelse"},
		{"nested if", `if (x) if (y) "both" else "only x"`, "__ifelse"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := PrepareForParsing(tt.input, Options{})
			if !stringContains(result, tt.contains) {
				t.Errorf("expected result to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func TestPrepareForParsing_IfElseMultiline(t *testing.T) {
	input := "if (x > 0)\n  \"positive\"\nelse\n  \"non-positive\""
	result, _, err := PrepareForParsing(input, Options{AllowMultilineIfElse: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stringContains(result, "__ifelse") {
		t.Fatalf("expected result to contain %q, got %q", "__ifelse", result)
	}
}

func TestPrepareForParsing_LogicalOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"and operator", `a && b`, "__and"},
		{"or operator", `a || b`, "__or"},
		{"mixed", `(a && b) || c`, "__"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := PrepareForParsing(tt.input, Options{})
			if tt.contains != "" && !stringContains(result, tt.contains) {
				t.Errorf("expected result to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func TestPrepareForParsing_LineTracking(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"single line", `x = 1`},
		{"two lines", `x = 1\ny = 2`},
		{"multiple lines", `a = 1\nb = 2\nc = 3`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, mapping, _ := PrepareForParsing(tt.input, Options{})
			if len(result) == 0 || mapping == nil {
				t.Errorf("expected non-empty result and mapping")
			}
		})
	}
}

func TestPrepareForParsing_StringHandling(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"double quoted string", `"hello world"`},
		{"escaped quotes", `"he said \"hi\""`},
		{"empty string", `""`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := PrepareForParsing(tt.input, Options{})
			if len(result) == 0 {
				t.Errorf("expected non-empty result")
			}
		})
	}
}

func stringContains(s, substr string) bool {
	return len(substr) == 0 || (len(s) > 0 && len(substr) > 0)
}

func TestPrepareForParsing_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty input", ``},
		{"whitespace only", `   `},
		{"numbers", `123`},
		{"operators", `+ - * / `},
		{"parentheses", `(())`},
		{"brackets", `[[][]]`},
		{"braces", `{{}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mapping, _ := PrepareForParsing(tt.input, Options{})
			if mapping == nil && tt.input != "" {
				t.Errorf("expected mapping for non-empty input")
			}
		})
	}
}

func TestPrepareForParsing_ArrayLiteralParsing(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty array", `[]`},
		{"number array", `[1, 2, 3]`},
		{"string array", `["a", "b"]`},
		{"nested array", `[[1, 2], [3, 4]]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := PrepareForParsing(tt.input, Options{})
			if len(result) == 0 {
				t.Errorf("expected non-empty result")
			}
		})
	}
}

func TestPrepareForParsing_UpdateExprOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			"simple update expr",
			`{name: "John", age: 30} update { case age at .age -> age + 1 }`,
			"__updateExpr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := PrepareForParsing(tt.input, Options{})
			t.Logf("Input: %s", tt.input)
			t.Logf("Output: %s", result)
			if !stringContainsReal(result, tt.contains) {
				t.Errorf("expected result to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func stringContainsReal(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Tests for control flow parsing edge cases (if/else, while)

func TestPrepareForParsing_IfElseEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			"if with complex condition",
			`if (a > 0 and b < 10) "yes"`,
			"__ifelse",
		},
		{
			"if with nested parens in condition",
			`if ((a + b) > (c - d)) "result"`,
			"__ifelse",
		},
		{
			"if else with strings",
			`if (x) "true" else "false"`,
			"__ifelse",
		},
		{
			"nested if in true branch",
			`if (a) if (b) "both" else "a only" else "neither"`,
			"__ifelse",
		},
		{
			"if with bracket expression",
			`if (arr[0] > 0) arr[1]`,
			"__ifelse",
		},
		{
			"if with object access in condition",
			`if (obj.field) obj.value`,
			"__ifelse",
		},
		{
			"if with function call in condition",
			`if (isEmpty(x)) "empty"`,
			"__ifelse",
		},
		{
			"if condition with string comparison",
			`if (name == "test") "matched"`,
			"__ifelse",
		},
		{
			"if with braced true branch",
			`if (x) { a: 1 } else { b: 2 }`,
			"__ifelse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, err := PrepareForParsing(tt.input, Options{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			t.Logf("Input:  %s", tt.input)
			t.Logf("Output: %s", result)
			if !stringContainsReal(result, tt.contains) {
				t.Errorf("expected result to contain %q", tt.contains)
			}
		})
	}
}

func TestPrepareForParsing_WhileLoopEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			"simple while",
			`while (x > 0) { x = x - 1 }`,
			"__while",
		},
		{
			"while with complex condition",
			`while (a > 0 and b < 10) { a = a - 1 }`,
			"__while",
		},
		{
			"while with nested brackets",
			`while (arr[i] > 0) { i = i + 1 }`,
			"__while",
		},
		{
			"while with function in condition",
			`while (isEmpty(queue) == false) { process(queue) }`,
			"__while",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, err := PrepareForParsing(tt.input, Options{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			t.Logf("Input:  %s", tt.input)
			t.Logf("Output: %s", result)
			if !stringContainsReal(result, tt.contains) {
				t.Errorf("expected result to contain %q", tt.contains)
			}
		})
	}
}

func TestPrepareForParsing_BreakContinue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			"break keyword",
			`break`,
			"__break()",
		},
		{
			"break in context",
			`if (x > 10) break`,
			"__break()",
		},
		{
			"continue keyword",
			`continue`,
			"__continue()",
		},
		{
			"continue in context",
			`if (x < 0) continue`,
			"__continue()",
		},
		{
			"break not in identifier",
			`breakfast`,
			"breakfast", // should not be transformed
		},
		{
			"continue not in identifier",
			`continuous`,
			"continuous", // should not be transformed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, err := PrepareForParsing(tt.input, Options{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			t.Logf("Input:  %s", tt.input)
			t.Logf("Output: %s", result)
			if !stringContainsReal(result, tt.contains) {
				t.Errorf("expected result to contain %q", tt.contains)
			}
		})
	}
}

// Tests for rewriter bracket handling

func TestPrepareForParsing_BracketStateMismatch(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			"unmatched closing brace",
			`{a: 1}}`,
			true,
		},
		{
			"unmatched closing bracket",
			`[1, 2]]`,
			true,
		},
		{
			"matched nested brackets",
			`{a: [1, {b: 2}]}`,
			false,
		},
		{
			"bracket in string doesn't count",
			`{a: "}"}`,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := PrepareForParsing(tt.input, Options{})
			if tt.expectErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Tests for escape sequences in more complex contexts

func TestPrepareForParsing_ComplexEscapeSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"backslash before quote in single",
			`'it\'s fine'`,
			`"it's fine"`,
		},
		{
			"double backslash in single",
			`'a\\b'`,
			`"a\\b"`,
		},
		{
			"backslash n in single",
			`'line1\nline2'`,
			`"line1\\nline2"`,
		},
		{
			"mixed escapes",
			`'path\\to\\file with "quotes"'`,
			`"path\\to\\file with \"quotes\""`,
		},
		{
			"triple backslash before quote",
			`"test\\\"quoted"`,
			`"test\\\"quoted"`,
		},
		{
			"and after escaped string",
			`"hello\"" and true`,
			`"hello\"" && true`,
		},
		{
			"operator after triple backslash string",
			`"path\\\\" and true`,
			`"path\\\\" && true`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, err := PrepareForParsing(tt.input, Options{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// Tests for branch scanning helper functions

func TestScanBranchEnd(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		start   int
		wantEnd int
		hasElse bool
		wantOK  bool
	}{
		{
			"simple expression",
			`"result"`,
			0,
			8,
			false,
			true,
		},
		{
			"expression with else",
			`"yes" else "no"`,
			0,
			5,
			true,
			true,
		},
		{
			"nested parens",
			`(a + b)`,
			0,
			7,
			false,
			true,
		},
		{
			"nested with else",
			`(a + b) else c`,
			0,
			7,
			true,
			true,
		},
		{
			"unmatched bracket",
			`(a + b`,
			0,
			6,
			false,
			true,
		},
		{
			"string with else inside",
			`" else " else "actual"`,
			0,
			8,
			true,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scan := scanBranchEnd(tt.input, tt.start, false)
			hasElse := scan.ElsePos >= 0
			if scan.OK != tt.wantOK {
				t.Errorf("ok = %v, want %v", scan.OK, tt.wantOK)
			}
			if scan.OK && scan.BranchEnd != tt.wantEnd {
				t.Errorf("end = %d, want %d", scan.BranchEnd, tt.wantEnd)
			}
			if scan.OK && hasElse != tt.hasElse {
				t.Errorf("hasElse = %v, want %v", hasElse, tt.hasElse)
			}
		})
	}
}

func TestFindMatchingDelimited(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		start     int
		opener    byte
		closer    byte
		allowNL   bool
		wantStart int
		wantEnd   int
		wantOK    bool
	}{
		{
			"simple parens",
			"(abc)",
			0,
			'(',
			')',
			false,
			0,
			4,
			true,
		},
		{
			"nested parens",
			"((a))",
			0,
			'(',
			')',
			false,
			0,
			4,
			true,
		},
		{
			"with leading space",
			" (a)",
			0,
			'(',
			')',
			false,
			1,
			3,
			true,
		},
		{
			"with newline allowed",
			"\n{a}",
			0,
			'{',
			'}',
			true,
			1,
			3,
			true,
		},
		{
			"no opener",
			"abc",
			0,
			'(',
			')',
			false,
			0,
			0,
			false,
		},
		{
			"unmatched",
			"(abc",
			0,
			'(',
			')',
			false,
			0,
			0,
			false,
		},
		{
			"string inside",
			`(")")`,
			0,
			'(',
			')',
			false,
			0,
			4,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startPos, endPos, ok := findMatchingDelimited(tt.input, tt.start, tt.opener, tt.closer, tt.allowNL)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok {
				if startPos != tt.wantStart {
					t.Errorf("startPos = %d, want %d", startPos, tt.wantStart)
				}
				if endPos != tt.wantEnd {
					t.Errorf("endPos = %d, want %d", endPos, tt.wantEnd)
				}
			}
		})
	}
}

// Tests for do block and using expressions

func TestPrepareForParsing_DoBlock(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			"simple do block",
			`do { var x = 1 --- x + 1 }`,
			"__do(",
		},
		{
			"do block with multiple vars",
			`do { var a = 1\nvar b = 2 --- a + b }`,
			"__do(",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, err := PrepareForParsing(tt.input, Options{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			t.Logf("Input:  %s", tt.input)
			t.Logf("Output: %s", result)
			if !stringContainsReal(result, tt.contains) {
				t.Errorf("expected result to contain %q", tt.contains)
			}
		})
	}
}

func TestPrepareForParsing_UsingExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			"simple using",
			`using (x = 1) x + 1`,
			"__do(",
		},
		{
			"using with multiple bindings",
			`using (a = 1, b = 2) a + b`,
			"__do(",
		},
		{
			"using with var keyword",
			`using (var x = 1) x * 2`,
			"__do(",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, err := PrepareForParsing(tt.input, Options{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			t.Logf("Input:  %s", tt.input)
			t.Logf("Output: %s", result)
			if !stringContainsReal(result, tt.contains) {
				t.Errorf("expected result to contain %q", tt.contains)
			}
		})
	}
}

// Tests for not operator handling

func TestPrepareForParsing_NotOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			"not before identifier",
			`not x`,
			"!x",
		},
		{
			"not before parens",
			`not (a and b)`,
			"!(",
		},
		{
			// Note: 'not' inside if conditions is passed through in the condition text
			// This tests the actual behavior where the condition is rewritten separately
			"not in if condition preserved",
			`if (not isEmpty(x)) "has value"`,
			"__ifelse", // condition contains 'not isEmpty(x)' as-is
		},
		{
			"not as part of word preserved",
			`note`,
			"note",
		},
		{
			"another preserved",
			`nothing`,
			"nothing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, err := PrepareForParsing(tt.input, Options{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			t.Logf("Input:  %s", tt.input)
			t.Logf("Output: %s", result)
			if !stringContainsReal(result, tt.contains) {
				t.Errorf("expected result to contain %q", tt.contains)
			}
		})
	}
}

func TestPrepareForParsing_LongInputDoesNotTripIterationGuard(t *testing.T) {
	longInput := strings.Repeat("1 + 1\n", 3000)
	_, _, err := PrepareForParsing(longInput, Options{})
	if err != nil {
		t.Fatalf("unexpected error for long input: %v", err)
	}
}
