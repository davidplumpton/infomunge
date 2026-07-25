package preprocessor

import "testing"

func TestPrepareForParsingPreservesNestedCollectionLambdas(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expected         string
		nestedTransform  string
		implicitRewrites bool
	}{
		{
			name:             "implicit outer and constant implicit inner",
			input:            `[[1]] map ($ map 1)`,
			expected:         `__map([]interface{}{[]interface{}{1,},}, __lambda("__arg", (__map(__arg, __lambda("__arg", 1)))))`,
			nestedTransform:  "replaceMapOperator",
			implicitRewrites: true,
		},
		{
			name:            "explicit outer and ungrouped explicit inner",
			input:           `[[1]] map ((x) -> x map (y) -> y)`,
			expected:        `__map([]interface{}{[]interface{}{1,},}, (__lambda("x",__map(x, __lambda("y", y)))))`,
			nestedTransform: "replaceMapOperator",
		},
		{
			name:            "explicit outer and grouped explicit inner",
			input:           `[[1]] map ((x) -> (x map (y) -> y))`,
			expected:        `__map([]interface{}{[]interface{}{1,},}, (__lambda("x", (__map(x, __lambda("y", y))))))`,
			nestedTransform: "replaceMapOperator",
		},
		{
			name:             "explicit outer and grouped implicit filter",
			input:            `[[1, 2], [3]] map ((x) -> x filter ($ > 1))`,
			expected:         `__map([]interface{}{[]interface{}{1, 2,}, []interface{}{3,},}, (__lambda("x",__filter(x, __lambda("__arg", (__arg > 1))))))`,
			nestedTransform:  "replaceFilterOperator",
			implicitRewrites: true,
		},
		{
			name:             "implicit outer and grouped explicit flatMap",
			input:            `[[1], [2]] map ($ flatMap ((y) -> [y, y]))`,
			expected:         `__map([]interface{}{[]interface{}{1,}, []interface{}{2,},}, __lambda("__arg", (__flatMap(__arg, (__lambda("y", []interface{}{y, y,}))))))`,
			nestedTransform:  "replaceFlatMapOperator",
			implicitRewrites: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var trace []TransformTraceEntry
			result, mapping, err := PrepareForParsing(tt.input, Options{
				TraceTransforms: func(entry TransformTraceEntry) {
					trace = append(trace, entry)
				},
			})
			if err != nil {
				t.Fatalf("PrepareForParsing returned error: %v", err)
			}
			if result != tt.expected {
				t.Fatalf("result = %q, want %q", result, tt.expected)
			}
			if len(mapping) != len(result) {
				t.Fatalf("mapping length = %d, want %d", len(mapping), len(result))
			}
			if _, ok := changedTraceEntry(trace, "replaceArrowFunctions"); !ok {
				t.Fatalf("trace did not include an arrow-function rewrite")
			}
			if _, ok := changedTraceEntry(trace, tt.nestedTransform); !ok {
				t.Fatalf("trace did not include nested %s rewrite", tt.nestedTransform)
			}
			if tt.implicitRewrites {
				if _, ok := changedTraceEntry(trace, "replaceImplicitLambdas"); !ok {
					t.Fatalf("trace did not include an implicit-lambda rewrite")
				}
			}
		})
	}
}
