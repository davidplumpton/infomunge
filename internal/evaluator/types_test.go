package evaluator

import (
	"go/token"
	"math"
	"reflect"
	"strings"
	"testing"

	"infomunge/pkg/values"
)

func TestCoerceToTypePreservesExactStructuralValues(t *testing.T) {
	t.Run("Array", func(t *testing.T) {
		original := Array{1, 2}

		got, err := coerceToType(original, "Array", nil, token.NoPos)
		if err != nil {
			t.Fatalf("coerceToType() error = %v", err)
		}
		gotArray := got.(Array)
		gotArray[0] = 99
		if original[0] != 99 {
			t.Fatal("coerceToType() copied the array instead of returning the original value")
		}
	})

	t.Run("Object", func(t *testing.T) {
		original := values.NewObject(2)
		values.SetObjectValue(original, "second", 2)
		values.SetObjectValue(original, "first", 1)

		got, err := coerceToType(original, "Object", nil, token.NoPos)
		if err != nil {
			t.Fatalf("coerceToType() error = %v", err)
		}
		gotObject := got.(Object)
		values.SetObjectValue(gotObject, "third", 3)
		if _, ok := original["third"]; !ok {
			t.Fatal("coerceToType() copied the object instead of returning the original value")
		}
		if gotKeys, want := values.ObjectKeys(gotObject), []string{"second", "first", "third"}; !reflect.DeepEqual(gotKeys, want) {
			t.Fatalf("ObjectKeys() = %v, want %v", gotKeys, want)
		}
	})

	t.Run("Function", func(t *testing.T) {
		original := &Lambda{}

		got, err := coerceToType(original, "Function", nil, token.NoPos)
		if err != nil {
			t.Fatalf("coerceToType() error = %v", err)
		}
		if got != original {
			t.Fatal("coerceToType() did not return the original function")
		}
	})

	t.Run("Regex", func(t *testing.T) {
		original := &Regex{Pattern: "abc"}

		got, err := coerceToType(original, "Regex", nil, token.NoPos)
		if err != nil {
			t.Fatalf("coerceToType() error = %v", err)
		}
		if got != original {
			t.Fatal("coerceToType() did not return the original regex")
		}
	})
}

func TestCoerceToTypeRejectsMismatchedStructuralValuesAsKnownTypes(t *testing.T) {
	for _, typeName := range []string{"Array", "Object", "Function", "Regex"} {
		t.Run(typeName, func(t *testing.T) {
			_, err := coerceToType("not a match", typeName, nil, token.NoPos)
			if err == nil {
				t.Fatal("coerceToType() error = nil, want coercion error")
			}
			if !strings.Contains(err.Error(), "cannot coerce") {
				t.Fatalf("coerceToType() error = %q, want coercion error", err)
			}
			if strings.Contains(err.Error(), "unknown type") {
				t.Fatalf("coerceToType() error = %q, built-in type was treated as unknown", err)
			}
		})
	}
}

func TestCoerceToStringUsesLanguageNullSpellingRecursively(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		want  string
	}{
		{
			name:  "array",
			value: Array{nil, Array{nil}},
			want:  "[null [null]]",
		},
		{
			name:  "XML multi-value",
			value: XMLMultiValue{nil},
			want:  "[null]",
		},
		{
			name: "object",
			value: Object{
				"value":  nil,
				"nested": Array{nil},
			},
			want: "map[nested:[null] value:null]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := coerceToString(tt.value)
			if got != tt.want {
				t.Fatalf("coerceToString() = %q, want %q", got, tt.want)
			}
			if strings.Contains(got, "<nil>") {
				t.Fatalf("coerceToString() leaked Go null spelling in %q", got)
			}
		})
	}
}

func TestCoerceToStringUsesDataWeaveNumberFormatting(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		want  string
	}{
		{name: "integer", value: 123, want: "123"},
		{name: "ordinary decimal", value: 12.5, want: "12.5"},
		{name: "small plain decimal", value: 1e-6, want: "0.000001"},
		{name: "small exponent", value: 1e-7, want: "1E-7"},
		{name: "positive exponent", value: 1e10, want: "1E+10"},
		{name: "negative zero", value: math.Copysign(0, -1), want: "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := coerceToString(tt.value); got != tt.want {
				t.Fatalf("coerceToString() = %q, want %q", got, tt.want)
			}
		})
	}
}
