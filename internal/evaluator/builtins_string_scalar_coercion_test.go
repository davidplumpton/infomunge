package evaluator

import (
	"go/ast"
	"regexp"
	"testing"
)

func TestUnaryStringBuiltinsCoerceScalarsAndPropagateNull(t *testing.T) {
	tests := []struct {
		name    string
		builtin func([]Value, *ast.CallExpr) (Value, error)
		input   Value
		want    Value
	}{
		{name: "lower number", builtin: callBuiltinLower, input: 123, want: "123"},
		{name: "lower boolean", builtin: callBuiltinLower, input: true, want: "true"},
		{name: "lower null", builtin: callBuiltinLower, input: nil, want: nil},
		{name: "upper number", builtin: callBuiltinUpper, input: -12.5, want: "-12.5"},
		{name: "upper boolean", builtin: callBuiltinUpper, input: false, want: "FALSE"},
		{name: "upper null", builtin: callBuiltinUpper, input: nil, want: nil},
		{name: "trim number", builtin: callBuiltinTrim, input: 12.5, want: "12.5"},
		{name: "trim exponent", builtin: callBuiltinTrim, input: 1e10, want: "1E+10"},
		{name: "trim boolean", builtin: callBuiltinTrim, input: true, want: "true"},
		{name: "trim null", builtin: callBuiltinTrim, input: nil, want: nil},
	}

	call := &ast.CallExpr{Fun: &ast.Ident{Name: "stringBuiltin"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.builtin([]Value{tt.input}, call)
			if err != nil {
				t.Fatalf("builtin returned an error: %v", err)
			}
			if !isEqual(got, tt.want) {
				t.Fatalf("builtin returned %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestStringPredicatesCoerceScalarsAndHandleNullSources(t *testing.T) {
	numberPattern := &Regex{Pattern: "23", Re: regexp.MustCompile("23")}
	typePattern := &Regex{Pattern: "umb", Re: regexp.MustCompile("umb")}
	tests := []struct {
		name    string
		builtin func([]Value, *ast.CallExpr) (Value, error)
		args    []Value
		want    bool
	}{
		{name: "startsWith numbers", builtin: callBuiltinStartsWith, args: []Value{12345, 123}, want: true},
		{name: "startsWith booleans", builtin: callBuiltinStartsWith, args: []Value{true, true}, want: true},
		{name: "startsWith null source", builtin: callBuiltinStartsWith, args: []Value{nil, nil}, want: false},
		{name: "endsWith numbers", builtin: callBuiltinEndsWith, args: []Value{12345, 45}, want: true},
		{name: "endsWith booleans", builtin: callBuiltinEndsWith, args: []Value{false, false}, want: true},
		{name: "endsWith null source", builtin: callBuiltinEndsWith, args: []Value{nil, nil}, want: false},
		{name: "contains numbers", builtin: callBuiltinContains, args: []Value{12345, 234}, want: true},
		{name: "contains exponent marker", builtin: callBuiltinContains, args: []Value{1e10, "E"}, want: true},
		{name: "contains booleans", builtin: callBuiltinContains, args: []Value{true, true}, want: true},
		{name: "contains scalar regex", builtin: callBuiltinContains, args: []Value{12345, numberPattern}, want: true},
		{name: "contains type source", builtin: callBuiltinContains, args: []Value{TypeValue("Number"), "umb"}, want: true},
		{name: "contains type source with regex", builtin: callBuiltinContains, args: []Value{TypeValue("Number"), typePattern}, want: true},
		{name: "contains type search text", builtin: callBuiltinContains, args: []Value{"value Number", TypeValue("Number")}, want: true},
		{name: "contains null source", builtin: callBuiltinContains, args: []Value{nil, nil}, want: false},
		{name: "contains null source ignores regex", builtin: callBuiltinContains, args: []Value{nil, numberPattern}, want: false},
		{name: "contains array preserves element semantics", builtin: callBuiltinContains, args: []Value{Array{1, nil}, nil}, want: true},
		{name: "contains array preserves type value semantics", builtin: callBuiltinContains, args: []Value{Array{TypeValue("Number")}, TypeValue("Number")}, want: true},
	}

	call := &ast.CallExpr{Fun: &ast.Ident{Name: "stringPredicate"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.builtin(tt.args, call)
			if err != nil {
				t.Fatalf("builtin returned an error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("builtin returned %#v, want %t", got, tt.want)
			}
		})
	}
}

func TestStringBuiltinScalarCoercionRejectsCollectionsAndNullSearchText(t *testing.T) {
	tests := []struct {
		name    string
		builtin func([]Value, *ast.CallExpr) (Value, error)
		args    []Value
	}{
		{name: "lower array", builtin: callBuiltinLower, args: []Value{Array{1}}},
		{name: "upper object", builtin: callBuiltinUpper, args: []Value{Object{"a": 1}}},
		{name: "trim array", builtin: callBuiltinTrim, args: []Value{Array{true}}},
		{name: "startsWith array source", builtin: callBuiltinStartsWith, args: []Value{Array{1}, "["}},
		{name: "startsWith null prefix", builtin: callBuiltinStartsWith, args: []Value{"null", nil}},
		{name: "endsWith object source", builtin: callBuiltinEndsWith, args: []Value{Object{"a": 1}, "1}"}},
		{name: "endsWith null suffix", builtin: callBuiltinEndsWith, args: []Value{"null", nil}},
		{name: "contains null search text", builtin: callBuiltinContains, args: []Value{"null", nil}},
		{name: "contains type source with null search text", builtin: callBuiltinContains, args: []Value{TypeValue("Null"), nil}},
		{name: "contains regex source", builtin: callBuiltinContains, args: []Value{&Regex{Pattern: "a", Re: regexp.MustCompile("a")}, "a"}},
		{name: "contains object source", builtin: callBuiltinContains, args: []Value{Object{"a": 1}, "a"}},
	}

	call := &ast.CallExpr{Fun: &ast.Ident{Name: "stringBuiltin"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.builtin(tt.args, call); err == nil {
				t.Fatal("builtin unexpectedly accepted unsupported arguments")
			}
		})
	}
}
