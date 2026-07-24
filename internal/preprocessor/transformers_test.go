package preprocessor

import (
	"infomunge/internal/stringutils"
	"testing"
)

func TestReplaceImplicitLambdas(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"map with $", "payload map $ + 1", "payload map (__arg) -> __arg + 1"},
		{"filter with $", "payload filter $ > 10", "payload filter (__arg) -> __arg > 10"},
		{"map with $ and $$", "payload map $ + $$", "payload map (__arg, __idx) -> __arg + __idx"},
		// Note: brace wrapping is done by wrapImplicitObjectLiteralBodies, not replaceImplicitLambdas
		{"map with implicit object body", "payload map user: $.name", "payload map (__arg) -> user: __arg.name"},
		{"nested map", "payload map ($ map $$)", "payload map (__arg, __idx) -> (__arg map __idx)"},
		{"already explicit", "payload map (item) -> item + 1", "payload map (item) -> item + 1"},
		{"explicit map then implicit filter", "arr map (x) -> (x + 1) filter $ > 1", "arr map (x) -> (x + 1) filter (__arg) -> __arg > 1"},
		{"reduce with paren no space", "data reduce($ + $$)", "data reduce (__arg, __idx) -> (__arg + __idx)"},
		{"reduce with paren complex", `data reduce(($$ splitBy ":")[0])`, `data reduce (__arg, __idx) -> ((__idx splitBy ":")[0])`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceImplicitLambdas(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestReplaceDotNotation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple dot", "obj.field", `obj["field"]`},
		{"chained dot", "obj.a.b", `obj["a"]["b"]`},
		{"dot with @", "obj.@attr", `obj["@attr"]`},
		{"dot with #", "obj.#", `obj["#"]`},
		{"dot assert", "obj.field!", `obj["field!"]`},
		{"dot optional", "obj.field?", `obj["field?"]`},
		{"no replacement in string", `"obj.field"`, `"obj.field"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := replaceDotNotationWithMapping(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestReplaceRecursiveDescent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple deep", "obj..field", `__deep(obj, "field")`},
		{"objvalues", "obj.*", `__objvalues(obj)`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := replaceRecursiveDescentWithMapping(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestReplaceFilterSelectors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple filter selector", "items[?($ > 1)]", `__filter_selector(items, __lambda("__arg, __idx", __arg > 1))`},
		{"filter selector with index", "items[?($$ > 0)]", `__filter_selector(items, __lambda("__arg, __idx", __idx > 0))`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceFilterSelectors(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestReplaceMetadataSelectors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple metadata", "payload.^size", `__metadata(payload, "size")`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceMetadataSelectors(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestReplaceStringInterpolation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple interpolation", `"Hello $(name)"`, `("Hello " + (name))`},
		{"multiple interpolation", `"$(a) + $(b) = $(a + b)"`, `((a) + " + " + (b) + " = " + (a + b))`},
		{"no interpolation", `"hello"`, `"hello"`},
		{"escaped dollar", `"$(a) \$ (b)"`, `((a) + " \$ (b)")`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceStringInterpolation(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

// Tests for bracket/string state tracking in scanner

func TestScannerBracketStateTracking(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		checkDepth int // position at which to check depth
		wantDepth  int
	}{
		{"single paren", "(a)", 1, 1},
		{"nested parens", "((a))", 2, 2},
		{"mixed brackets", "([{a}])", 3, 3},
		{"after close", "(a)b", 3, 0},
		{"bracket in string ignored", `("(")`, 4, 1}, // depth stays 1, inner ( is in string
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := stringutils.NewExpressionScanner(tt.input)
			for i := 0; i < tt.checkDepth && sc.Pos() < len(tt.input); i++ {
				sc.Next()
			}
			if sc.Depth() != tt.wantDepth {
				t.Errorf("at pos %d: expected depth %d, got %d", tt.checkDepth, tt.wantDepth, sc.Depth())
			}
		})
	}
}

func TestScannerStringStateTracking(t *testing.T) {
	// Note: checkPos is the number of Next() calls, so we process chars 0 through checkPos-1.
	tests := []struct {
		name        string
		input       string
		checkPos    int // number of Next() calls
		wantInStr   bool
		description string
	}{
		{"not in string initially", `abc`, 1, false, "no string"},
		{"inside double quote", `"abc"`, 2, true, "inside string"},
		{"after string closes", `"a"b`, 4, false, "after closing quote at pos 2"},
		{"escaped quote stays in string", `"a\"b"`, 4, true, "escaped quote not closing"},
		{"double backslash closes string", `"a\\"`, 5, false, "even backslashes should not escape quote"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := stringutils.NewExpressionScanner(tt.input)
			for i := 0; i < tt.checkPos && sc.Pos() < len(tt.input); i++ {
				sc.Next()
			}
			if sc.IsInString() != tt.wantInStr {
				t.Errorf("%s: at pos %d expected inString=%v, got %v",
					tt.description, tt.checkPos, tt.wantInStr, sc.IsInString())
			}
		})
	}
}

func TestScannerFindMatchingCloseBracket(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		startPos int
		want     int
	}{
		{"simple parens", "(abc)", 0, 4},
		{"nested parens", "((a))", 0, 4},
		{"inner nested", "((a))", 1, 3},
		{"bracket in string", `(")")`, 0, 4},
		{"unmatched", "(abc", 0, -1},
		{"square brackets", "[a, b]", 0, 5},
		{"curly braces", "{a: 1}", 0, 5},
		{"mixed nested", "([{x}])", 0, 6},
		{"complex string content", `("bracket (in) string")`, 0, 22},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := stringutils.NewExpressionScanner(tt.input)
			got := sc.FindMatchingCloseBracket(tt.startPos)
			if got != tt.want {
				t.Errorf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

// Tests for escape sequence handling

func TestIsEscapedAt(t *testing.T) {
	tests := []struct {
		name string
		s    string
		pos  int
		want bool
	}{
		{"no backslash", `abc"`, 3, false},
		{"single backslash", `ab\"`, 3, true},
		{"double backslash", `ab\\"`, 4, false},
		{"triple backslash", `ab\\\"`, 5, true},
		{"quad backslash", `ab\\\\"`, 6, false},
		{"at start", `"a`, 0, false},
		{"backslash at start", `\"a`, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isEscapedAt(tt.s, tt.pos)
			if got != tt.want {
				t.Errorf("isEscapedAt(%q, %d) = %v, want %v", tt.s, tt.pos, got, tt.want)
			}
		})
	}
}

func TestCopyStringLiteral(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		startPos   int
		wantEnd    int
		wantResult string
	}{
		{"simple string", `"hello"`, 0, 6, `"hello"`},
		{"with escaped quote", `"he\"llo"`, 0, 8, `"he\"llo"`},
		{"with escaped backslash", `"path\\file"`, 0, 11, `"path\\file"`},
		{"double backslash before quote closes", `"a\\"b"`, 0, 4, `"a\\"`},
		{"empty string", `""`, 0, 1, `""`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runes := []rune(tt.input)
			endPos, result := copyStringLiteral(runes, tt.startPos, nil)
			if endPos != tt.wantEnd {
				t.Errorf("endPos = %d, want %d", endPos, tt.wantEnd)
			}
			if string(result) != tt.wantResult {
				t.Errorf("result = %q, want %q", string(result), tt.wantResult)
			}
		})
	}
}

// Tests for arrow function parsing

func TestReplaceArrowFunctions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple arrow", "(x) -> x + 1", `__lambda("x", x + 1)`},
		{"multiple params", "(a, b) -> a + b", `__lambda("a, b", a + b)`},
		{"nested parens in body", "(x) -> (x + 1) * 2", `__lambda("x", (x + 1) * 2)`},
		{"string in body", `(x) -> "hello"`, `__lambda("x", "hello")`},
		{"no arrow function", "(a + b)", "(a + b)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceArrowFunctions(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

// Tests for collection operators

func TestReplaceCollectionOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		opFunc   func(string) string
	}{
		// Note: These functions operate on raw input before array literal transformation
		{"map basic", `arr map __lambda("x", x + 1)`, `__map(arr, __lambda("x", x + 1))`, replaceMapOperator},
		{"filter basic", `arr filter __lambda("x", x > 0)`, `__filter(arr, __lambda("x", x > 0))`, replaceFilterOperator},
		{"reduce basic", `arr reduce __lambda("acc", acc + 1)`, `__reduce(arr, __lambda("acc", acc + 1))`, replaceReduceOperator},
		{"groupBy basic", `arr groupBy __lambda("x", x.key)`, `__groupBy(arr, __lambda("x", x.key))`, replaceGroupByOperator},
		{"flatMap basic", `arr flatMap __lambda("x", x)`, `__flatMap(arr, __lambda("x", x))`, replaceFlatMapOperator},
		// Array literals are not transformed by these functions - they expect raw input
		{"map with literal", `[1, 2] map __lambda("x", x)`, `__map([1, 2], __lambda("x", x))`, replaceMapOperator},
		{"map stops before next collection operator", `arr map __lambda("x", (x + 1)) filter __lambda("x", x > 1)`, `__map(arr, __lambda("x", (x + 1))) filter __lambda("x", x > 1)`, replaceMapOperator},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.opFunc(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestIsCollectionOperatorHelpers(t *testing.T) {
	input := " map value filter rest"
	runes := []rune(input)

	if !isCollectionOperatorAt(input, 1) {
		t.Fatalf("expected collection operator at byte pos 1")
	}
	if !isCollectionOperatorAtRunes(runes, 1) {
		t.Fatalf("expected collection operator at rune pos 1")
	}
	if !isCollectionOperatorWithSpacesAt(input, 0) {
		t.Fatalf("expected space-delimited collection operator at byte pos 0")
	}
	if isCollectionOperatorAt(input, 0) {
		t.Fatalf("did not expect collection operator at leading space")
	}
}

// Tests for case statement parsing

func TestParseCaseItems(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLen  int
		contains string
	}{
		{"single case", `1 -> "one"`, 1, `"pattern": "1"`},
		{"two cases", `1 -> "one", 2 -> "two"`, 2, `"pattern": "2"`},
		{"case with else", `case 1 -> "one", else -> "other"`, 2, `"pattern": "else"`},
		{"case with expression", `n if n > 0 -> "positive"`, 1, `"pattern": "n if n > 0"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := parseCaseItems(tt.input)
			if len(items) != tt.wantLen {
				t.Errorf("expected %d items, got %d: %v", tt.wantLen, len(items), items)
			}
			found := false
			for _, item := range items {
				if len(item) > 0 && containsSubstr(item, tt.contains) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected items to contain %q, got %v", tt.contains, items)
			}
		})
	}
}

func containsSubstr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Tests for operator replacements

func TestReplaceAsOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple as", "x as String", `__coerce(x, "String")`},
		{"as with expression", "(a + b) as Number", `__coerce((a + b), "Number")`},
		{"as in string ignored", `"x as y"`, `"x as y"`},
		{"as with optional type", "x as String?", `__coerce(x, "String?")`},
		{"as stays on immediate rhs operand", "value default other as String", `value default __coerce(other, "String")`},
		{"as keeps config object", `payload as String map[string]interface{}{"format": "yyyy"}`, `__coerce(payload, "String", map[string]interface{}{"format": "yyyy"})`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := replaceAsOperatorWithMapping(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestReplaceIsOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple is", "x is String", `__isType(x, "String")`},
		{"is with expression", "(a) is Number", `__isType((a), "Number")`},
		{"is in string ignored", `"x is y"`, `"x is y"`},
		{"is stays on immediate boolean operand", `left and right is String`, `left and __isType(right, "String")`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := replaceIsOperatorWithMapping(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestReplacePipeToFunctionOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple pipe", "x | trim", "trim(x)"},
		{"pipe to upper", `"hello" | upper`, `upper("hello")`},
		{"pipe to function call unchanged", "x | func(y)", "x | func(y)"},
		{"pipe in string ignored", `"a | b"`, `"a | b"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replacePipeToFunctionOperator(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

// Tests for module call replacement

func TestReplaceModuleCall(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple module call", "mod::func(x)", `__modcall("mod", "func", x)`},
		{"module call no args", "mod::func()", `__modcall("mod", "func")`},
		{"module call in string ignored", `"mod::func"`, `"mod::func"`},
		{"incomplete module", "mod::", "mod::"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceModuleCall(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

// Tests for array range indexing

func TestReplaceArrayRangeIndexing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple range", "arr[0 to 2]", "__rangeIndex(arr, 0, 2)"},
		{"range with variable start", "arr[start to 5]", "__rangeIndex(arr, start, 5)"},
		{"no range", "arr[0]", "arr[0]"},
		{"range with variable end", "arr[0 to end]", "__rangeIndex(arr, 0, end)"},
		{"range with computed end", "arr[0 to sizeOf(arr) - 2]", "__rangeIndex(arr, 0, sizeOf(arr) - 2)"},
		{"negative bounds", "arr[-2 to -1]", "__rangeIndex(arr, -2, -1)"},
		{"multiple independent ranges", `["hello"[0 to 0], "world"[1 to 2]]`, `[__rangeIndex("hello", 0, 0), __rangeIndex("world", 1, 2)]`},
		{"chained ranges", `arr[0 to 3][1 to 2]`, `__rangeIndex(__rangeIndex(arr, 0, 3), 1, 2)`},
		// Range inside function call arguments (bug q022)
		{"range inside function call", `typeOf(payload["name"][0 to 0])`, `typeOf(__rangeIndex(payload["name"], 0, 0))`},
		{"range inside function call with dot notation", `typeOf(payload.name[0 to 0])`, `typeOf(__rangeIndex(payload.name, 0, 0))`},
		{"range inside nested brackets", `foo(bar["x"][1 to 3])`, `foo(__rangeIndex(bar["x"], 1, 3))`},
		{"nested bracket in end expression", `arr[0 to indexes[1]]`, `__rangeIndex(arr, 0, indexes[1])`},
		{"range text inside string", `arr[0 to sizeOf("x[y to z]")]`, `__rangeIndex(arr, 0, sizeOf("x[y to z]"))`},
		{"array range expression is not indexing", `[-2 to -1]`, `[-2 to -1]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceArrayRangeIndexing(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

// Tests for extractInterpolationExpr with nested structures

func TestExtractInterpolationExpr(t *testing.T) {
	tests := []struct {
		name    string
		content string
		pos     int
		expr    string
		endPos  int
	}{
		{"simple", "name)", 0, "name", 5},
		{"with parens", "func(x))", 0, "func(x)", 8},
		{"nested parens", "func(a, (b + c)))", 0, "func(a, (b + c))", 17},
		{"string in expr", `"a") + b`, 0, `"a"`, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, endPos := extractInterpolationExpr(tt.content, tt.pos)
			if expr != tt.expr {
				t.Errorf("expr = %q, want %q", expr, tt.expr)
			}
			if endPos != tt.endPos {
				t.Errorf("endPos = %d, want %d", endPos, tt.endPos)
			}
		})
	}
}

// Tests for multi-statement sequences

func TestReplaceMultiStatementSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"single line", "x + 1", "x + 1"},
		{"two lines", "x = 1\ny = 2", "__seq(x = 1, y = 2)"},
		{"three lines", "a\nb\nc", "__seq(a, b, c)"},
		{"with empty lines", "a\n\nb", "__seq(a, b)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceMultiStatementSequences(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

// Tests for binary operators using stringutils

func TestReplaceDefaultOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple default", "x default 0", "__default(x, 0)"},
		{"default with string", `x default "none"`, `__default(x, "none")`},
		{"default in string ignored", `"a default b"`, `"a default b"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := replaceConfiguredBinaryOperatorWithMapping(tt.input, binaryOpDefault)
			if err != nil {
				t.Fatalf("replaceConfiguredBinaryOperatorWithMapping() error = %v", err)
			}
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestReplaceConcatenateOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Array literals are not transformed by this function - it operates on raw input
		{"array concat raw", "[1] ++ [2]", "__concat([1], [2])"},
		{"string concat", `"a" ++ "b"`, `__concat("a", "b")`},
		{"variable concat", "arr1 ++ arr2", "__concat(arr1, arr2)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := replaceConfiguredBinaryOperatorWithMapping(tt.input, binaryOpConcatenate)
			if err != nil {
				t.Fatalf("replaceConfiguredBinaryOperatorWithMapping() error = %v", err)
			}
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestReplaceContainsOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"array contains", "arr contains x", "contains(arr, x)"},
		{"string contains", `"hello" contains "ell"`, `contains("hello", "ell")`},
		{"contains in string ignored", `"a contains b"`, `"a contains b"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := replaceConfiguredBinaryOperatorWithMapping(tt.input, binaryOpContains)
			if err != nil {
				t.Fatalf("replaceConfiguredBinaryOperatorWithMapping() error = %v", err)
			}
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestReplaceModOperatorUsesDataWeavePrecedence(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"division before mod", "8 / 4 mod 3", "mod(8 / 4, 3)"},
		{"multiplication before mod", "8 * 4 mod 3", "mod(8 * 4, 3)"},
		{"percent before mod", "20 % 6 mod 4", "mod(20 % 6, 4)"},
		{"subtraction before mod", "5 - 2 mod 2", "mod(5 - 2, 2)"},
		{"multiplication after mod", "20 mod 6 * 4", "mod(20, 6 * 4)"},
		{"division after mod", "20 mod 6 / 2", "mod(20, 6 / 2)"},
		{"percent after mod", "20 mod 6 % 4", "mod(20, 6 % 4)"},
		{"repeated mod", "20 mod 6 mod 4", "mod(mod(20, 6), 4)"},
		{"additive expression before mod", "2 + 8 * 4 mod 3", "mod(2 + 8 * 4, 3)"},
		{"unary sign in chain", "8 * -4 mod 3", "mod(8 * -4, 3)"},
		{"unary sign on right", "8 mod -3 * 2", "mod(8, -3 * 2)"},
		{"grouped left operand", "(8 / 4) mod 3", "mod((8 / 4), 3)"},
		{"grouped right operand", "8 / (4 mod 3)", "8 / (mod(4, 3))"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := replaceConfiguredBinaryOperatorWithMapping(tt.input, binaryOpMod)
			if err != nil {
				t.Fatalf("replaceConfiguredBinaryOperatorWithMapping() error = %v", err)
			}
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestReplaceConfiguredBinaryOperator_MissingConfigReturnsError(t *testing.T) {
	input := "x default 0"
	result, _, err := replaceConfiguredBinaryOperatorWithMapping(input, "missing-config-key")
	if err == nil {
		t.Fatal("expected missing binary operator config error, got nil")
	}
	if result != input {
		t.Fatalf("expected original input %q on error, got %q", input, result)
	}
}
