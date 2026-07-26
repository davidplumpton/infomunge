package preprocessor

import (
	"fmt"
	"strings"
	"testing"
)

func TestPrepareForParsingKeepsUpdateInsideCollectionLambdaBodies(t *testing.T) {
	tests := []struct {
		operator string
		function string
	}{
		{operator: "map", function: "__map"},
		{operator: "filter", function: "__filter"},
		{operator: "reduce", function: "__reduce"},
		{operator: "flatMap", function: "__flatMap"},
		{operator: "groupBy", function: "__groupBy"},
		{operator: "pluck", function: "__pluck"},
		{operator: "sort", function: "orderBy"},
		{operator: "orderBy", function: "orderBy"},
		{operator: "maxBy", function: "maxBy"},
		{operator: "minBy", function: "minBy"},
		{operator: "distinctBy", function: "distinctBy"},
		{operator: "filterObject", function: "filterObject"},
		{operator: "mapObject", function: "mapObject"},
	}

	for _, tt := range tests {
		t.Run(tt.operator, func(t *testing.T) {
			input := fmt.Sprintf(
				`source %s (item) -> item update { case value at .a -> value + 1 }`,
				tt.operator,
			)
			var trace []TransformTraceEntry
			result, mapping, err := PrepareForParsing(input, Options{
				TraceTransforms: func(entry TransformTraceEntry) {
					trace = append(trace, entry)
				},
			})
			if err != nil {
				t.Fatalf("PrepareForParsing returned error: %v", err)
			}
			if len(mapping) != len(result) {
				t.Fatalf("mapping length = %d, want %d", len(mapping), len(result))
			}

			want := fmt.Sprintf(
				`%s(source, __lambda("item", __updateExpr(item, "case value at .a -> value + 1")))`,
				tt.function,
			)
			if result != want {
				t.Fatalf("result = %q, want %q", result, want)
			}

			coreEntry, ok := changedTraceEntry(trace, "rewriteCoreSyntax")
			if !ok {
				t.Fatal("trace did not include a core rewrite")
			}
			if !strings.Contains(
				coreEntry.After,
				fmt.Sprintf(`source %s (item) -> __updateExpr(item,`, tt.operator),
			) {
				t.Fatalf("core rewrite moved update outside %s callback: %s", tt.operator, coreEntry.After)
			}
		})
	}
}

func TestPrepareForParsingKeepsUpdateInsideImplicitCollectionLambda(t *testing.T) {
	input := `source map $ update { case value at .a -> value + 1 }`
	result, mapping, err := PrepareForParsing(input, Options{})
	if err != nil {
		t.Fatalf("PrepareForParsing returned error: %v", err)
	}

	want := `__map(source, __lambda("__arg", __updateExpr(__arg, "case value at .a -> value + 1")))`
	if result != want {
		t.Fatalf("result = %q, want %q", result, want)
	}
	if len(mapping) != len(result) {
		t.Fatalf("mapping length = %d, want %d", len(mapping), len(result))
	}
}

func TestPrepareForParsingPreservesParenthesizedOuterUpdate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "grouped collection",
			input: `(source map (item) -> item) update { case value at .a -> value + 1 }`,
			want:  `__updateExpr((__map(source, __lambda("item", item))), "case value at .a -> value + 1")`,
		},
		{
			name:  "grouped callback",
			input: `source map ((item) -> item) update { case value at .a -> value + 1 }`,
			want:  `__updateExpr(__map(source, (__lambda("item", item))), "case value at .a -> value + 1")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, mapping, err := PrepareForParsing(tt.input, Options{})
			if err != nil {
				t.Fatalf("PrepareForParsing returned error: %v", err)
			}
			if result != tt.want {
				t.Fatalf("result = %q, want %q", result, tt.want)
			}
			if len(mapping) != len(result) {
				t.Fatalf("mapping length = %d, want %d", len(mapping), len(result))
			}
		})
	}
}
