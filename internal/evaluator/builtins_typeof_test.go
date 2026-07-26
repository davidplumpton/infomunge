package evaluator

import (
	"go/ast"
	"testing"
)

func TestTypeOfReturnsDistinctTypeValue(t *testing.T) {
	call := &ast.CallExpr{Fun: &ast.Ident{Name: "typeOf"}}

	result, err := callBuiltinTypeOf([]Value{false}, call)
	if err != nil {
		t.Fatalf("callBuiltinTypeOf() error = %v", err)
	}
	typeValue, ok := result.(TypeValue)
	if !ok {
		t.Fatalf("callBuiltinTypeOf() = %#v (%T), want TypeValue", result, result)
	}
	if typeValue != TypeValue("Boolean") {
		t.Fatalf("callBuiltinTypeOf() = %q, want %q", typeValue, TypeValue("Boolean"))
	}
	if got := getTypeName(typeValue); got != "Type" {
		t.Fatalf("getTypeName(typeOf(false)) = %q, want %q", got, "Type")
	}
	if numericEquals(typeValue, "Boolean") {
		t.Fatal(`typeOf(false) unexpectedly equals the String value "Boolean"`)
	}
	if got := coerceToString(typeValue); got != "Boolean" {
		t.Fatalf("coerceToString(typeOf(false)) = %q, want %q", got, "Boolean")
	}
}

func TestEvaluateNestedTypeOf(t *testing.T) {
	expression := `typeOf(typeOf(false))`

	result, err := Evaluate(expression, Context{}, nil, 0, expression)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result != TypeValue("Type") {
		t.Fatalf("Evaluate() = %#v (%T), want TypeValue(%q)", result, result, "Type")
	}
}
