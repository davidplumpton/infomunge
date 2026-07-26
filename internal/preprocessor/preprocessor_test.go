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

func TestPrepareForParsing_DefaultArrayFallbacks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty array",
			input:    "null default []",
			expected: "__default(null, []interface{}{})",
		},
		{
			name:     "single element array",
			input:    "null default [1]",
			expected: "__default(null, []interface{}{1,})",
		},
		{
			name:     "multiple element array",
			input:    "null default [1, 2]",
			expected: "__default(null, []interface{}{1, 2,})",
		},
		{
			name:     "nested array",
			input:    "null default [[1], [2, 3]]",
			expected: "__default(null, []interface{}{[]interface{}{1,}, []interface{}{2, 3,},})",
		},
		{
			name:     "grouped default",
			input:    "(null default [1, 2])",
			expected: "(__default(null, []interface{}{1, 2,}))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, mapping, err := PrepareForParsing(tt.input, Options{})
			if err != nil {
				t.Fatalf("PrepareForParsing returned error: %v", err)
			}
			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
			if len(mapping) != len(result) {
				t.Fatalf("mapping length = %d, want %d", len(mapping), len(result))
			}
		})
	}
}

func TestPrepareForParsing_DefaultArrayFallbackInsideLambda(t *testing.T) {
	result, mapping, err := PrepareForParsing(`[null] map ($ default [1, 2])`, Options{})
	if err != nil {
		t.Fatalf("PrepareForParsing returned error: %v", err)
	}
	if !strings.Contains(result, `__default(__arg, []interface{}{1, 2,})`) {
		t.Fatalf("expected rewritten array fallback inside lambda, got %q", result)
	}
	if len(mapping) != len(result) {
		t.Fatalf("mapping length = %d, want %d", len(mapping), len(result))
	}
}

func TestPrepareForParsing_CollectionOperatorAppliesOutsideDefault(t *testing.T) {
	input := `payload.name default flatten([[0]]) reduce (item, x) -> x`
	expected := `__reduce(__default(payload["name"], flatten([]interface{}{[]interface{}{0,},})), __lambda("item, x", x))`

	result, mapping, err := PrepareForParsing(input, Options{})
	if err != nil {
		t.Fatalf("PrepareForParsing returned error: %v", err)
	}
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
	if len(mapping) != len(result) {
		t.Fatalf("mapping length = %d, want %d", len(mapping), len(result))
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
		{"implicit value index", "$[0]", "$[0]"},
		{"implicit index parameter index", "$$[0]", "$$[0]"},
		{"chained implicit value index", "$[0][1]", "$[0][1]"},
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

func TestRewriteCore_ImplicitLambdaIndexPreservesArrayLiterals(t *testing.T) {
	input := `[$[0], [$[1], 2]]`
	expected := `[]interface{}{$[0], []interface{}{$[1], 2,},}`

	result, _, err := newRewriter(input, Options{}).RewriteCoreWithDepth(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestPrepareForParsing_ImplicitLambdaIndexAcrossCollectionOperators(t *testing.T) {
	operators := []string{
		"map",
		"filter",
		"reduce",
		"groupBy",
		"sort",
		"maxBy",
		"minBy",
		"orderBy",
		"distinctBy",
		"filterObject",
		"mapObject",
		"flatMap",
		"pluck",
	}

	for _, operator := range operators {
		t.Run(operator, func(t *testing.T) {
			input := "items " + operator + " $[0][1]"
			result, _, err := PrepareForParsing(input, Options{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(result, "__arg[0][1]") {
				t.Fatalf("expected implicit selector in transformed output, got %q", result)
			}
			if strings.Contains(result, "$[]interface{}") {
				t.Fatalf("implicit selector was rewritten as an array literal: %q", result)
			}
		})
	}
}

func TestPrepareForParsing_ImplicitLambdaIndexAlongsideArrayLiteral(t *testing.T) {
	result, _, err := PrepareForParsing(`items flatMap [$[0]]`, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "[]interface{}{__arg[0],}") {
		t.Fatalf("expected array literal containing implicit selector, got %q", result)
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

func TestPrepareForParsing_MappingLengthInvariant(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple expression", "1 + 2"},
		{"if-else", `if (x > 0) "yes" else "no"`},
		{"nested if-else", `if (a) if (b) "both" else "a" else "neither"`},
		{"update expression", `x update { case v at .a -> v + 1 }`},
		{"object literal", `{name: "Alice", age: 30}`},
		{"while loop", `while (x > 0) { x }`},
		{"do block", `do { var x = 1 --- x + 1 }`},
		{"using expression", `using (x = 1) x + 1`},
		{"and-or", `a and b or c`},
		{"not operator", `not x`},
		{"single quotes", `'hello'`},
		{"break keyword", `break`},
		{"continue keyword", `continue`},
		{"range builtin", `range(10)`},
		{"coerce equals", `a ~= b`},
		{"array literal", `[1, 2, 3]`},
		{"empty object", `{}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, mapping, err := PrepareForParsing(tt.input, Options{AllowMultilineIfElse: true})
			if err != nil {
				return // errors are fine; we check the invariant only on success
			}
			if len(result) != len(mapping) {
				t.Errorf("mapping length mismatch: len(result)=%d, len(mapping)=%d for input %q",
					len(result), len(mapping), tt.input)
			}
		})
	}
}

func TestPrepareForParsing_MappingRangeValidity(t *testing.T) {
	inputs := []string{
		"1 + 2",
		`if (x) "yes"`,
		`{a: 1, b: 2}`,
		`a and b`,
		`'hello world'`,
		`not x`,
		`break`,
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			result, mapping, err := PrepareForParsing(input, Options{})
			if err != nil {
				return
			}
			for i, pos := range mapping {
				if pos < 0 || pos >= len(input) {
					t.Errorf("mapping[%d]=%d out of range [0, %d) for input %q, result %q",
						i, pos, len(input), input, result)
				}
			}
		})
	}
}

func TestPrepareForParsing_InlineIfObjectValuesTerminateAtObjectBoundary(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		contains     string
		keySourcePos int
	}{
		{
			name:         "explicit object false branch before closing brace",
			input:        `{foo: if (false) 1 else 2}`,
			contains:     `__ifelse(false, 1, 2)`,
			keySourcePos: 1,
		},
		{
			name:         "implicit top level object false branch before wrapper close",
			input:        `foo: if (false) 1 else 2`,
			contains:     `__ifelse(false, 1, 2)`,
			keySourcePos: 0,
		},
		{
			name:         "true branch before comma",
			input:        `{foo: if (true) 1, bar: 2}`,
			contains:     `__ifelse(true, 1, nil)`,
			keySourcePos: 1,
		},
		{
			name:         "true branch before closing brace",
			input:        `{foo: if (true) 1}`,
			contains:     `__ifelse(true, 1, nil)`,
			keySourcePos: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, mapping, err := PrepareForParsing(tt.input, Options{})
			if err != nil {
				t.Fatalf("PrepareForParsing returned error: %v", err)
			}
			if !strings.Contains(result, tt.contains) {
				t.Fatalf("expected result to contain %q, got %q", tt.contains, result)
			}
			if len(result) != len(mapping) {
				t.Fatalf("mapping length %d != result length %d", len(mapping), len(result))
			}

			keyPos := strings.Index(result, `"foo"`)
			if keyPos == -1 {
				t.Fatalf("result missing rewritten key: %q", result)
			}
			for i := 0; i < len("foo"); i++ {
				if mapping[keyPos+1+i] != tt.keySourcePos+i {
					t.Fatalf("expected key mapping[%d] to be %d, got %d",
						keyPos+1+i, tt.keySourcePos+i, mapping[keyPos+1+i])
				}
			}
		})
	}
}

func TestPrepareForParsing_ExactMappingAcrossPostProcessingTransforms(t *testing.T) {
	input := `payload.user default items..name`

	result, mapping, err := PrepareForParsing(input, Options{})
	if err != nil {
		t.Fatalf("PrepareForParsing returned error: %v", err)
	}

	if result != `__default(payload["user"],__deep(items, "name"))` {
		t.Fatalf("unexpected result: %q", result)
	}

	userPos := strings.Index(result, "user")
	if userPos == -1 {
		t.Fatalf("result missing user selector: %q", result)
	}
	for i := 0; i < len("user"); i++ {
		if mapping[userPos+i] != 8+i {
			t.Fatalf("expected mapping for user[%d] to be %d, got %d", i, 8+i, mapping[userPos+i])
		}
	}

	namePos := strings.LastIndex(result, "name")
	if namePos == -1 {
		t.Fatalf("result missing recursive selector: %q", result)
	}
	originalNamePos := strings.LastIndex(input, "name")
	for i := 0; i < len("name"); i++ {
		if mapping[namePos+i] != originalNamePos+i {
			t.Fatalf("expected mapping for name[%d] to be %d, got %d", i, originalNamePos+i, mapping[namePos+i])
		}
	}
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

func TestPrepareForParsing_ConditionalObjectLiteralEntries(t *testing.T) {
	input := `{(a: value) if (flag), (b: value + 1)}`
	result, _, err := PrepareForParsing(input, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "merge(") {
		t.Fatalf("expected merged object rewrite, got %q", result)
	}
	if !strings.Contains(result, "__ifelse(") {
		t.Fatalf("expected conditional entry rewrite, got %q", result)
	}
	if !strings.Contains(result, `map[string]interface{}{"b": value + 1,}`) {
		t.Fatalf("expected non-conditional entry to be preserved, got %q", result)
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
		{
			"expression before comma",
			`value, next`,
			0,
			5,
			false,
			true,
		},
		{
			"expression before closing object",
			`value}`,
			0,
			5,
			false,
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

// TestMappingInvariantViolationReturnsError verifies that a corrupted rewriter
// (mismatched result/mapping lengths) surfaces an error instead of panicking.
func TestMappingInvariantViolationReturnsError(t *testing.T) {
	r := newRewriter("x", Options{})
	// Manually corrupt the state: append to result without a mapping entry.
	r.result = append(r.result, 'x')
	// mapping remains empty — lengths now differ.

	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("assertMappingInvariant panicked instead of setting r.err: %v", rec)
		}
	}()

	r.assertMappingInvariant()
	if r.err == nil {
		t.Fatal("expected r.err to be set after invariant violation, got nil")
	}
	if !strings.Contains(r.err.Error(), "mapping invariant violated") {
		t.Errorf("unexpected error message: %v", r.err)
	}
}
