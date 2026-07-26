package evaluator

import (
	"fmt"
	"go/token"
	"reflect"
	"strings"
	"testing"

	"infomunge/pkg/values"
)

func TestEvalArrayStringIndexCollectsOnlyMatchingImmediateFields(t *testing.T) {
	input := Array{
		Object{"score": 1},
		Object{"other": 2},
		3,
		Object{"score": nil},
	}

	got, err := evalArrayStringIndex(input, "score", token.Pos(1))
	if err != nil {
		t.Fatalf("evalArrayStringIndex() returned an unexpected error: %v", err)
	}
	want := Array{1, nil}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("evalArrayStringIndex() = %#v, want %#v", got, want)
	}
}

func TestEvalArrayStringIndexReturnsNullWithoutMatches(t *testing.T) {
	tests := []struct {
		name  string
		input Array
	}{
		{name: "primitive elements", input: Array{-80.45, 2}},
		{name: "objects missing field", input: Array{Object{"other": 1}}},
		{name: "empty array", input: Array{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evalArrayStringIndex(tt.input, "score", token.Pos(1))
			if err != nil {
				t.Fatalf("evalArrayStringIndex() returned an unexpected error: %v", err)
			}
			if got != nil {
				t.Fatalf("evalArrayStringIndex() = %#v, want nil", got)
			}
		})
	}
}

func TestEvalArrayStringIndexTreatsSelectorSuffixesAsLiteralKeys(t *testing.T) {
	input := Array{
		Object{"score?": 7},
		Object{"score!": 8},
	}

	for _, tt := range []struct {
		key  string
		want Array
	}{
		{key: "score?", want: Array{7}},
		{key: "score!", want: Array{8}},
	} {
		got, err := evalArrayStringIndex(input, tt.key, token.Pos(1))
		if err != nil {
			t.Fatalf("evalArrayStringIndex(%q) returned an unexpected error: %v", tt.key, err)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("evalArrayStringIndex(%q) = %#v, want %#v", tt.key, got, tt.want)
		}
	}
}

func TestEvalArrayPresenceSelectorReportsAnyMatchingField(t *testing.T) {
	if !evalArrayPresenceSelector(Array{Object{"score": 1}, Object{"other": 2}, 3}, "score") {
		t.Fatal("evalArrayPresenceSelector() = false, want true")
	}
	if evalArrayPresenceSelector(Array{Object{"other": 2}, 3}, "score") {
		t.Fatal("evalArrayPresenceSelector() = true, want false")
	}
}

func TestEvalArrayAssertSelectorRequiresAtLeastOneMatchingField(t *testing.T) {
	got, err := evalArrayAssertSelector(
		Array{Object{"score": 1}, Object{"other": 2}, 3, Object{"score": 4}},
		"score",
		token.Pos(1),
	)
	if err != nil {
		t.Fatalf("evalArrayAssertSelector() returned an unexpected error: %v", err)
	}
	want := Array{1, 4}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("evalArrayAssertSelector() = %#v, want %#v", got, want)
	}

	_, err = evalArrayAssertSelector(Array{Object{"other": 2}, 3}, "score", token.Pos(1))
	if err == nil {
		t.Fatal("evalArrayAssertSelector() returned nil error without a matching field")
	}
	if !strings.Contains(err.Error(), `assert selector failed: missing key "score"`) {
		t.Fatalf("evalArrayAssertSelector() error = %q, want missing-key context", err)
	}
}

func TestEvalIndexUsesExplicitSelectorOperations(t *testing.T) {
	array := Array{Object{"score": 7}}
	object := Object{"score": 8}

	got, err := evalArrayIndex(array, selectorOperation{mode: selectorModePresence, key: "score"}, token.Pos(1))
	if err != nil {
		t.Fatalf("evalArrayIndex() returned an unexpected error: %v", err)
	}
	if got != true {
		t.Fatalf("evalArrayIndex() = %#v, want true", got)
	}

	got, err = evalObjectIndex(object, selectorOperation{mode: selectorModeAssert, key: "score"}, token.Pos(1))
	if err != nil {
		t.Fatalf("evalObjectIndex() returned an unexpected error: %v", err)
	}
	if got != 8 {
		t.Fatalf("evalObjectIndex() = %#v, want 8", got)
	}
}

func TestEvalObjectStringIndexTreatsSelectorSuffixesAsLiteralKeys(t *testing.T) {
	object := Object{"score?": 7, "score!": 8}

	for _, tt := range []struct {
		key  string
		want Value
	}{
		{key: "score?", want: 7},
		{key: "score!", want: 8},
	} {
		got, err := evalObjectStringIndex(object, tt.key, token.Pos(1))
		if err != nil {
			t.Fatalf("evalObjectStringIndex(%q) returned an unexpected error: %v", tt.key, err)
		}
		if got != tt.want {
			t.Fatalf("evalObjectStringIndex(%q) = %#v, want %#v", tt.key, got, tt.want)
		}
	}
}

func TestEvalObjectOrdinalIndexResolvesNegativeIndexesFromEnd(t *testing.T) {
	object := values.NewObject(3)
	values.SetObjectValue(object, "second", 2)
	values.SetObjectValue(object, "first", 1)
	values.SetObjectValue(object, "third", 3)

	tests := []struct {
		name  string
		index int
		want  Value
	}{
		{name: "last", index: -1, want: 3},
		{name: "middle", index: -2, want: 1},
		{name: "first", index: -3, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evalObjectOrdinalIndex(object, tt.index, token.Pos(1))
			if err != nil {
				t.Fatalf("evalObjectOrdinalIndex(%d) returned an unexpected error: %v", tt.index, err)
			}
			if got != tt.want {
				t.Fatalf("evalObjectOrdinalIndex(%d) = %v, want %v", tt.index, got, tt.want)
			}
		})
	}
}

func TestEvalObjectOrdinalIndexRejectsIndexesOutsideBothBounds(t *testing.T) {
	object := values.NewObject(2)
	values.SetObjectValue(object, "first", 1)
	values.SetObjectValue(object, "second", 2)

	for _, index := range []int{-3, 2} {
		t.Run(fmt.Sprintf("index_%d", index), func(t *testing.T) {
			_, err := evalObjectOrdinalIndex(object, index, token.Pos(1))
			if err == nil {
				t.Fatalf("evalObjectOrdinalIndex(%d) returned nil error", index)
			}
			want := fmt.Sprintf("object index out of bounds: %d (object has 2 keys)", index)
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("evalObjectOrdinalIndex(%d) error %q does not contain %q", index, err, want)
			}
		})
	}
}
