package evaluator

import (
	"go/ast"
	"regexp"
	"testing"
)

func TestFindSubstringInStringUsesCharacterOffsets(t *testing.T) {
	got := findSubstringInString("éaé🙂é", "é")
	want := Array{float64(0), float64(2), float64(4)}

	if !numericEquals(got, want) {
		t.Fatalf("findSubstringInString() = %#v, want %#v", got, want)
	}
}

func TestFindRegexInStringUsesCharacterSpans(t *testing.T) {
	got, ok := findRegexInString("éa🙂é", "é|🙂", "")
	if !ok {
		t.Fatal("findRegexInString() did not compile a valid pattern")
	}
	want := Array{
		Array{float64(0), float64(1)},
		Array{float64(2), float64(3)},
		Array{float64(3), float64(4)},
	}

	if !numericEquals(got, want) {
		t.Fatalf("findRegexInString() = %#v, want %#v", got, want)
	}
}

func TestBuiltinFindRegexValueUsesCharacterSpans(t *testing.T) {
	search := &Regex{
		Pattern: "é|🙂",
		Re:      regexp.MustCompile("é|🙂"),
	}
	call := &ast.CallExpr{Fun: &ast.Ident{Name: "find"}}

	got, err := callBuiltinFind([]Value{"éa🙂é", search}, call)
	if err != nil {
		t.Fatalf("callBuiltinFind() error = %v", err)
	}
	want := Array{
		Array{float64(0), float64(1)},
		Array{float64(2), float64(3)},
		Array{float64(3), float64(4)},
	}

	if !numericEquals(got, want) {
		t.Fatalf("callBuiltinFind() = %#v, want %#v", got, want)
	}
}
