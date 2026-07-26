package evaluator

import (
	"reflect"
	"strings"
	"testing"

	"infomunge/pkg/values"
)

func TestSetValueAtPath_OutOfBoundsNonTerminalIndexReturnsError(t *testing.T) {
	value := Array{
		Array{1.0},
	}
	path := []selectorSegment{
		{index: 5, isIndex: true},
		{index: 0, isIndex: true},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("setValueAtPath should not panic, got %v", r)
		}
	}()

	_, err := setValueAtPath(value, path, 99.0, 0)
	if err == nil {
		t.Fatalf("expected out-of-bounds error, got nil")
	}
	if !strings.Contains(err.Error(), "index out of bounds: 5") {
		t.Fatalf("expected index out-of-bounds error, got %v", err)
	}
}

func TestMergeEvaluatedUpdate_UsesAncestorAndFirstIdenticalSelector(t *testing.T) {
	path := func(selector string) []selectorSegment {
		t.Helper()
		parsed, err := parseSelectorPath(selector)
		if err != nil {
			t.Fatalf("parse selector %q: %v", selector, err)
		}
		return parsed
	}

	tests := []struct {
		name       string
		selectors  []string
		values     []string
		wantValues []string
	}{
		{
			name:       "ancestor before descendant",
			selectors:  []string{".a", ".a.b"},
			values:     []string{"ancestor", "descendant"},
			wantValues: []string{"ancestor"},
		},
		{
			name:       "descendant before ancestor",
			selectors:  []string{".a.b", ".a"},
			values:     []string{"descendant", "ancestor"},
			wantValues: []string{"ancestor"},
		},
		{
			name:       "first identical selector",
			selectors:  []string{".a", ".a"},
			values:     []string{"first", "second"},
			wantValues: []string{"first"},
		},
		{
			name:       "disjoint selectors",
			selectors:  []string{".a", ".b"},
			values:     []string{"left", "right"},
			wantValues: []string{"left", "right"},
		},
		{
			name:       "array ancestor replaces object descendant",
			selectors:  []string{".items[0].value", ".items[0]"},
			values:     []string{"value", "item"},
			wantValues: []string{"item"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var updates []evaluatedUpdate
			for i, selector := range test.selectors {
				updates = mergeEvaluatedUpdate(updates, evaluatedUpdate{
					path:  path(selector),
					value: test.values[i],
				})
			}

			gotValues := make([]string, len(updates))
			for i, update := range updates {
				gotValues[i] = update.value.(string)
			}
			if !reflect.DeepEqual(gotValues, test.wantValues) {
				t.Fatalf("merged values = %v, want %v", gotValues, test.wantValues)
			}
		})
	}
}

func TestApplyUpdateCases_PreservesObjectOrderAndDeepCopyIsolation(t *testing.T) {
	nested := values.NewObject(1)
	values.SetObjectValue(nested, "value", 1)
	input := values.NewObject(3)
	values.SetObjectValue(input, "second", 2)
	values.SetObjectValue(input, "nested", nested)
	values.SetObjectValue(input, "first", 1)

	result, err := applyUpdateCases(
		input,
		"case whole at .nested -> whole",
		NewScope(nil),
		0,
	)
	if err != nil {
		t.Fatalf("apply update cases: %v", err)
	}

	resultObject := result.(Object)
	if got, want := values.ObjectKeys(resultObject), []string{"second", "nested", "first"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("result keys = %v, want %v", got, want)
	}

	resultNested := resultObject["nested"].(Object)
	resultNested["value"] = 99
	if got := input["nested"].(Object)["value"]; got != 1 {
		t.Fatalf("source nested value = %v after result mutation, want 1", got)
	}
}
