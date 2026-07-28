package evaluator

import (
	"reflect"
	"strings"
	"testing"
)

func TestFilterSelectorReturnsNullWithoutCollectionMatches(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{
			name: "array",
			expr: `__filter_selector([]interface{}{1, 2}, __lambda("value, index", value > 9))`,
		},
		{
			name: "object",
			expr: `__filter_selector(map[string]interface{}{"a": 1}, __lambda("value, index", value > 9))`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Evaluate(test.expr, Context{}, nil, 0, test.expr)
			if err != nil {
				t.Fatalf("Evaluate(%q) error = %v", test.expr, err)
			}
			if got != nil {
				t.Fatalf("Evaluate(%q) = %#v, want nil", test.expr, got)
			}
		})
	}
}

func TestFilterSelectorPreservesSuccessfulCollectionMatches(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want Value
	}{
		{
			name: "array",
			expr: `__filter_selector([]interface{}{1, 2, 3}, __lambda("value, index", value > 1))`,
			want: Array{2, 3},
		},
		{
			name: "object",
			expr: `__filter_selector(map[string]interface{}{"a": 1, "b": 2}, __lambda("value, index", value > 1))`,
			want: Object{"b": 2},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Evaluate(test.expr, Context{}, nil, 0, test.expr)
			if err != nil {
				t.Fatalf("Evaluate(%q) error = %v", test.expr, err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("Evaluate(%q) = %#v, want %#v", test.expr, got, test.want)
			}
		})
	}
}

func TestFilterSelectorRejectsUnsupportedSources(t *testing.T) {
	tests := []struct {
		name string
		expr string
		kind string
	}{
		{name: "string", expr: `"abc"`, kind: "string"},
		{name: "number", expr: `123`, kind: "int"},
		{name: "null", expr: `nil`, kind: "<nil>"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr := `__filter_selector(` + test.expr + `, __lambda("value, index", true))`
			_, err := Evaluate(expr, Context{}, nil, 0, expr)
			if err == nil {
				t.Fatalf("Evaluate(%q) error = nil, want source type error", expr)
			}
			want := "selector filter expects an array or object, got " + test.kind
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("Evaluate(%q) error = %q, want %q", expr, err, want)
			}
		})
	}
}
