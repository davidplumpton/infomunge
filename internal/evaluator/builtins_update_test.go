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

func TestResolveUpdateArrayIndex(t *testing.T) {
	tests := []struct {
		name      string
		index     int
		length    int
		wantIndex int
		wantOK    bool
	}{
		{name: "positive", index: 1, length: 3, wantIndex: 1, wantOK: true},
		{name: "last", index: -1, length: 3, wantIndex: 2, wantOK: true},
		{name: "negative length boundary", index: -3, length: 3, wantIndex: 0, wantOK: true},
		{name: "below negative length", index: -4, length: 3, wantIndex: -1, wantOK: false},
		{name: "positive length boundary", index: 3, length: 3, wantIndex: 3, wantOK: false},
		{name: "empty array", index: -1, length: 0, wantIndex: -1, wantOK: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotIndex, gotOK := resolveUpdateArrayIndex(test.index, test.length)
			if gotIndex != test.wantIndex || gotOK != test.wantOK {
				t.Fatalf(
					"resolveUpdateArrayIndex(%d, %d) = (%d, %t), want (%d, %t)",
					test.index,
					test.length,
					gotIndex,
					gotOK,
					test.wantIndex,
					test.wantOK,
				)
			}
		})
	}
}

func TestParseSelectorPathAndUpsert(t *testing.T) {
	tests := []struct {
		name       string
		selector   string
		wantUpsert bool
		wantPath   []selectorSegment
	}{
		{
			name:     "ordinary selector",
			selector: ".profile.name",
			wantPath: []selectorSegment{
				{fieldName: "profile", index: -1},
				{fieldName: "name", index: -1},
			},
		},
		{
			name:       "upsert selector",
			selector:   ".profile.name!",
			wantUpsert: true,
			wantPath: []selectorSegment{
				{fieldName: "profile", index: -1},
				{fieldName: "name", index: -1},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotPath, gotUpsert, err := parseSelectorPathAndUpsert(test.selector)
			if err != nil {
				t.Fatalf("parse selector: %v", err)
			}
			if gotUpsert != test.wantUpsert {
				t.Fatalf("upsert = %t, want %t", gotUpsert, test.wantUpsert)
			}
			if !reflect.DeepEqual(gotPath, test.wantPath) {
				t.Fatalf("path = %#v, want %#v", gotPath, test.wantPath)
			}
		})
	}
}

func TestApplyUpdateCases_UpsertsMissingObjectFields(t *testing.T) {
	tests := []struct {
		name     string
		input    Object
		cases    string
		wantKeys []string
		want     Value
	}{
		{
			name:     "missing terminal field binds null",
			input:    Object{"a": 1},
			cases:    "case missing at .missing! -> missing",
			wantKeys: []string{"a", "missing"},
			want: Object{
				"a":       1,
				"missing": nil,
			},
		},
		{
			name:     "missing parent objects are created",
			input:    Object{"a": 1},
			cases:    "case name at .profile.name! -> \"John\"",
			wantKeys: []string{"a", "profile"},
			want: Object{
				"a": 1,
				"profile": Object{
					"name": "John",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := applyUpdateCases(test.input, test.cases, NewScope(nil), 0)
			if err != nil {
				t.Fatalf("apply update cases: %v", err)
			}
			gotObject := got.(Object)
			if !reflect.DeepEqual(gotObject, test.want) {
				t.Fatalf("result = %#v, want %#v", gotObject, test.want)
			}
			if !reflect.DeepEqual(values.ObjectKeys(gotObject), test.wantKeys) {
				t.Fatalf("keys = %v, want %v", values.ObjectKeys(gotObject), test.wantKeys)
			}
		})
	}
}

func TestSetValueAtPath_ResolvesNegativeNonTerminalIndexes(t *testing.T) {
	value := Array{
		Array{1, 2},
	}
	path := []selectorSegment{
		{index: -1, isIndex: true},
		{index: -1, isIndex: true},
	}

	got, err := setValueAtPath(value, path, 9, 0)
	if err != nil {
		t.Fatalf("setValueAtPath returned error: %v", err)
	}
	want := Array{
		Array{1, 9},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("setValueAtPath result = %#v, want %#v", got, want)
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

func TestApplyUpdateCases_PreservesInputOriginForNestedCase(t *testing.T) {
	user := values.NewObject(0)
	input := values.NewObject(1)
	values.SetObjectValue(input, "user", user)
	values.MarkInputValue(input)

	got, err := applyUpdateCases(
		input,
		"case user at .user -> __default(user[\"missing\"] + 1, 5)",
		NewScope(nil),
		0,
	)
	if err != nil {
		t.Fatalf("apply update cases: %v", err)
	}

	want := Object{"user": 5}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("result = %#v, want %#v", got, want)
	}
	if !values.ObjectHasInputOrigin(got.(Object)) {
		t.Fatal("updated result lost input-origin metadata")
	}
}

func TestApplyUpdateCases_PreservesInputOriginThroughCollectionCallback(t *testing.T) {
	first := values.NewObject(0)
	second := values.NewObject(0)
	input := values.NewObject(1)
	values.SetObjectValue(input, "users", Array{first, second})
	values.MarkInputValue(input)

	got, err := applyUpdateCases(
		input,
		"case users at .users -> __map(users, __lambda(\"user\", __default(user[\"missing\"] + 1, 5)))",
		NewScope(nil),
		0,
	)
	if err != nil {
		t.Fatalf("apply update cases: %v", err)
	}

	want := Object{"users": Array{5, 5}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("result = %#v, want %#v", got, want)
	}
	if !values.ObjectHasInputOrigin(got.(Object)) {
		t.Fatal("updated collection result lost input-origin metadata")
	}
}
