package evaluator

import (
	"reflect"
	"testing"
)

func TestGroupByScalarSourcesUseStringSemantics(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want Value
	}{
		{
			name: "number is coerced to string",
			expr: `groupBy(1.5, __lambda("item", typeOf(item)))`,
			want: Object{"String": "1.5"},
		},
		{
			name: "integer characters are grouped",
			expr: `groupBy(101, __lambda("item", item))`,
			want: Object{"1": "11", "0": "0"},
		},
		{
			name: "boolean is coerced to string",
			expr: `groupBy(true, __lambda("item", typeOf(item)))`,
			want: Object{"String": "true"},
		},
		{
			name: "string groups characters into strings",
			expr: `groupBy("aba", __lambda("item", item))`,
			want: Object{"a": "aa", "b": "b"},
		},
		{
			name: "null propagates",
			expr: `groupBy(nil, __lambda("item", panic("callback must not run")))`,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Evaluate(tt.expr, Context{}, nil, 0, tt.expr)
			if err != nil {
				t.Fatalf("Evaluate() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Evaluate() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestGroupByStringCallbackReceivesRuneIndex(t *testing.T) {
	expr := `groupBy("åaå", __lambda("item, index", index))`

	got, err := Evaluate(expr, Context{}, nil, 0, expr)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	want := Object{"0": "å", "1": "a", "2": "å"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Evaluate() = %#v, want %#v", got, want)
	}
}
