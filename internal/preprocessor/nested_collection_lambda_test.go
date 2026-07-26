package preprocessor

import (
	"strings"
	"testing"
)

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
			name:            "ungrouped explicit outer and ungrouped explicit inner",
			input:           `[1] map (i, x) -> [] flatMap (x0) -> [true]`,
			expected:        `__map([]interface{}{1,}, __lambda("i, x",__flatMap([]interface{}{}, __lambda("x0", []interface{}{true,}))))`,
			nestedTransform: "replaceFlatMapOperator",
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

func TestPrepareForParsingPreservesComputedCollectionSourcesInExplicitLambdaBodies(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expected        string
		arrowContains   string
		nestedTransform string
	}{
		{
			name:            "function call followed by reduce",
			input:           `[1,2] map (x) -> flatten([[x, x + 1]]) reduce (v,a) -> v + a`,
			expected:        `__map([]interface{}{1,2,}, __lambda("x",__reduce(flatten([]interface{}{[]interface{}{x, x + 1,},}), __lambda("v,a", v + a))))`,
			arrowContains:   `__lambda("x", flatten([]interface{}{[]interface{}{x, x + 1,},}) reduce __lambda("v,a", v + a))`,
			nestedTransform: "replaceReduceOperator",
		},
		{
			name:            "concatenated arrays followed by map",
			input:           `[1,2] map (x) -> [x] ++ [x + 1] map (v) -> v * 2`,
			expected:        `__map([]interface{}{1,2,}, __lambda("x",__map(__concat([]interface{}{x,}, []interface{}{x + 1,}), __lambda("v", v * 2))))`,
			arrowContains:   `__lambda("x", []interface{}{x,} ++ []interface{}{x + 1,} map __lambda("v", v * 2))`,
			nestedTransform: "replaceMapOperator",
		},
		{
			name:            "object literal followed by pluck",
			input:           `[1,2] map (x) -> {a:x,b:x+1} pluck $`,
			expected:        `__map([]interface{}{1,2,}, __lambda("x",__pluck(map[string]interface{}{"a":x,"b":x+1,}, __lambda("__arg", __arg))))`,
			arrowContains:   `__lambda("x", map[string]interface{}{"a":x,"b":x+1,} pluck __lambda("__arg", __arg))`,
			nestedTransform: "replacePluckOperator",
		},
		{
			name:            "conditional call source followed by filter",
			input:           `[1,2] map (x) -> flatten(if (x > 1) [[x]] else [[x + 1]]) filter (v) -> v > 0`,
			expected:        `__map([]interface{}{1,2,}, __lambda("x",__filter(flatten(__ifelse(x > 1, []interface{}{[]interface{}{x,},}, []interface{}{[]interface{}{x + 1,},})), __lambda("v", v > 0))))`,
			arrowContains:   `__lambda("x", flatten(__ifelse(x > 1, []interface{}{[]interface{}{x,},}, []interface{}{[]interface{}{x + 1,},})) filter __lambda("v", v > 0))`,
			nestedTransform: "replaceFilterOperator",
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

			arrowEntry, ok := changedTraceEntry(trace, "replaceArrowFunctions")
			if !ok {
				t.Fatal("trace did not include an arrow-function rewrite")
			}
			if !strings.Contains(arrowEntry.After, tt.arrowContains) {
				t.Fatalf("arrow rewrite did not keep nested collection in body:\n%s", arrowEntry.After)
			}
			if _, ok := changedTraceEntry(trace, tt.nestedTransform); !ok {
				t.Fatalf("trace did not include nested %s rewrite", tt.nestedTransform)
			}
		})
	}
}

func TestPrepareForParsingPreservesIdentifierCollectionSourcesInBoundedExplicitLambdaBodies(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "collection callback map",
			input:    `[0,0] map (ignored) -> xs map sizeOf($)`,
			expected: `__map([]interface{}{0,0,}, __lambda("ignored",__map(xs, __lambda("__arg", sizeOf(__arg)))))`,
		},
		{
			name:     "function argument filter",
			input:    `apply([1,2], (xs) -> xs filter ($ > 1))`,
			expected: `apply([]interface{}{1,2,}, __lambda("xs",__filter(xs, __lambda("__arg", (__arg > 1)))))`,
		},
		{
			name:     "then callback reduce",
			input:    `values then (xs) -> xs reduce (item, acc = 0) -> acc + item`,
			expected: `then(values, __lambda("xs",__reduce(xs, __lambda("item, acc = 0", acc + item))))`,
		},
		{
			name:     "onNull callback mapObject",
			input:    `null onNull () -> xs mapObject (v, k) -> {(k): v}`,
			expected: `onNull(null, __lambda("",mapObject(xs, __lambda("v, k", map[string]interface{}{(k): v,}))))`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, mapping, err := PrepareForParsing(tt.input, Options{})
			if err != nil {
				t.Fatalf("PrepareForParsing returned error: %v", err)
			}
			if result != tt.expected {
				t.Fatalf("result = %q, want %q", result, tt.expected)
			}
			if len(mapping) != len(result) {
				t.Fatalf("mapping length = %d, want %d", len(mapping), len(result))
			}
		})
	}
}

func TestCollectionSourceOwnsOperatorPreservesOuterChainingCases(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		operator string
		want     bool
	}{
		{"array map", `[]interface{}{x,}`, " map value", true},
		{"array does not own object operator", `[]interface{}{x,}`, " pluck value", false},
		{"object pluck", `map[string]interface{}{"a":x,}`, " pluck value", true},
		{"object mapObject", `map[string]interface{}{"a":x,}`, " mapObject value", true},
		{"object does not own array operator", `map[string]interface{}{"a":x,}`, " map value", false},
		{"complete function call", `flatten([]interface{}{x,})`, " reduce value", true},
		{"concatenated arrays", `[]interface{}{x,} ++ []interface{}{x + 1,}`, " groupBy value", true},
		{"scalar body remains outer", `x + 1`, " filter value", false},
		{"predicate body remains outer", `x > 1`, " map value", false},
		{"completed grouped body remains outer", `(flatten([]interface{}{x,}))`, " reduce value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectionSourceOwnsOperator(tt.body, tt.operator, 0)
			if got != tt.want {
				t.Fatalf("collectionSourceOwnsOperator(%q, %q) = %v, want %v", tt.body, tt.operator, got, tt.want)
			}
		})
	}
}
