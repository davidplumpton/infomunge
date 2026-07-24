package evaluator

import (
	"go/ast"
	"go/token"
	"reflect"
	"strings"
	"testing"
)

func TestCompareValuesPreservesLargeIntegerPrecision(t *testing.T) {
	t.Parallel()

	cmp, err := compareValues(9007199254740993, 9007199254740992)
	if err != nil {
		t.Fatalf("compareValues returned an unexpected error: %v", err)
	}
	if cmp != 1 {
		t.Fatalf("compareValues returned %d, want 1", cmp)
	}
}

func TestSortPreservesLargeIntegerPrecision(t *testing.T) {
	t.Parallel()

	call := &ast.CallExpr{Fun: &ast.Ident{Name: "sort", NamePos: token.Pos(1)}}
	got, err := callBuiltinSort(
		[]Value{Array{9007199254740993, 9007199254740992}},
		call,
	)
	if err != nil {
		t.Fatalf("sort returned an unexpected error: %v", err)
	}

	want := Array{9007199254740992, 9007199254740993}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sort returned %#v, want %#v", got, want)
	}
}

func TestSortRejectsMixedComparableTypes(t *testing.T) {
	t.Parallel()

	call := &ast.CallExpr{Fun: &ast.Ident{Name: "sort", NamePos: token.Pos(1)}}
	_, err := callBuiltinSort([]Value{Array{1, "2"}}, call)
	if err == nil {
		t.Fatal("sort returned nil error for mixed number and string values")
	}
	if !strings.Contains(err.Error(), "element types cannot be mixed") {
		t.Fatalf("sort error %q does not explain the homogeneity requirement", err)
	}
}
