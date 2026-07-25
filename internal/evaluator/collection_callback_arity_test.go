package evaluator

import (
	"go/ast"
	"go/token"
	"reflect"
	"strings"
	"testing"
)

func TestArrayCollectionCallbacksAllowZeroParameters(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want Value
	}{
		{name: "map", expr: `__map([]interface{}{1, 2}, __lambda("", 9))`, want: Array{9, 9}},
		{name: "filter", expr: `__filter([]interface{}{1, 2}, __lambda("", true))`, want: Array{1, 2}},
		{name: "flatMap", expr: `__flatMap([]interface{}{1, 2}, __lambda("", []interface{}{9}))`, want: Array{9, 9}},
		{name: "orderBy", expr: `orderBy([]interface{}{2, 1}, __lambda("", 0))`, want: Array{2, 1}},
		{name: "distinctBy", expr: `distinctBy([]interface{}{1, 2}, __lambda("", 0))`, want: Array{1}},
		{name: "groupBy", expr: `groupBy([]interface{}{1, 2}, __lambda("", "all"))`, want: Object{"all": Array{1, 2}}},
		{name: "maxBy", expr: `maxBy([]interface{}{1, 2}, __lambda("", 0))`, want: 1},
		{name: "minBy", expr: `minBy([]interface{}{1, 2}, __lambda("", 0))`, want: 1},
		{name: "takeWhile", expr: `takeWhile([]interface{}{1, 2}, __lambda("", false))`, want: Array{}},
		{name: "dropWhile", expr: `dropWhile([]interface{}{1, 2}, __lambda("", false))`, want: Array{1, 2}},
		{name: "some", expr: `some([]interface{}{1, 2}, __lambda("", true))`, want: true},
		{name: "every", expr: `every([]interface{}{1, 2}, __lambda("", false))`, want: false},
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

func TestObjectCollectionCallbacksAllowParameterPrefixes(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want Value
	}{
		{
			name: "mapObject zero parameters",
			expr: `mapObject(map[string]interface{}{"a": 1}, __lambda("", []interface{}{"fixed", 9}))`,
			want: Object{"fixed": 9},
		},
		{
			name: "mapObject one parameter receives value",
			expr: `mapObject(map[string]interface{}{"a": 1}, __lambda("v", []interface{}{"value", v}))`,
			want: Object{"value": 1},
		},
		{
			name: "mapObject three parameters receive index",
			expr: `mapObject(map[string]interface{}{"a": 1, "b": 2}, __lambda("v, k, i", []interface{}{k, i}))`,
			want: Object{"a": 0, "b": 1},
		},
		{
			name: "filterObject zero parameters",
			expr: `filterObject(map[string]interface{}{"a": 1, "b": 2}, __lambda("", true))`,
			want: Object{"a": 1, "b": 2},
		},
		{
			name: "filterObject one parameter receives value",
			expr: `filterObject(map[string]interface{}{"a": 1, "b": 2}, __lambda("v", v > 1))`,
			want: Object{"b": 2},
		},
		{
			name: "filterObject three parameters receive index",
			expr: `filterObject(map[string]interface{}{"a": 1, "b": 2}, __lambda("v, k, i", i == 1))`,
			want: Object{"b": 2},
		},
		{
			name: "pluck zero parameters",
			expr: `__pluck(map[string]interface{}{"a": 1, "b": 2}, __lambda("", 9))`,
			want: Array{9, 9},
		},
		{
			name: "pluck one parameter receives value",
			expr: `__pluck(map[string]interface{}{"a": 1, "b": 2}, __lambda("v", v))`,
			want: Array{1, 2},
		},
		{
			name: "pluck three parameters receive index",
			expr: `__pluck(map[string]interface{}{"a": 1, "b": 2}, __lambda("v, k, i", i))`,
			want: Array{0, 1},
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

func TestCollectionCallbackArityDoesNotRelaxDirectCalls(t *testing.T) {
	lambda := &Lambda{
		Params:  []ParamDef{{Name: "x", ExpectedKind: KindUnknown}},
		BodyAST: &ast.Ident{Name: "x"},
	}

	_, err := invokeUserLambda(lambda, nil, token.NoPos, NewScope(Context{}), 0)
	if err == nil {
		t.Fatal("Evaluate() error = nil, want exact-arity failure")
	}
	if !strings.Contains(err.Error(), "function expects 1 arguments, got 0") {
		t.Fatalf("Evaluate() error = %q, want exact-arity failure", err)
	}
}
