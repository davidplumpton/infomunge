package evaluator

import (
	"errors"
	"go/ast"
	"reflect"
	"testing"

	"infomunge/pkg/values"
)

func TestTryResultPreservesServiceFieldOrder(t *testing.T) {
	tests := []struct {
		name string
		got  Object
		want []string
	}{
		{name: "success", got: newTryResult(true, "result", 42), want: []string{"success", "result"}},
		{name: "failure", got: newTryResult(false, "error", Object{}), want: []string{"success", "error"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := values.ObjectKeys(tt.got); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ObjectKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildErrorObjectPreservesDiagnosticFieldOrder(t *testing.T) {
	got := buildErrorObject(errors.New("boom"))
	want := []string{"kind", "message", "location", "stack"}
	if keys := values.ObjectKeys(got); !reflect.DeepEqual(keys, want) {
		t.Fatalf("ObjectKeys() = %v, want %v", keys, want)
	}
}

func TestParseURIPreservesResultAndQueryFieldOrder(t *testing.T) {
	call := &ast.CallExpr{Fun: &ast.Ident{Name: "parseURI"}}
	value, err := callBuiltinParseURI([]Value{"https://user:pass@example.com:8443/a/b?x=1&x=2&msg=hello#top"}, call)
	if err != nil {
		t.Fatalf("callBuiltinParseURI() error = %v", err)
	}

	result, ok := value.(Object)
	if !ok {
		t.Fatalf("callBuiltinParseURI() value = %T, want Object", value)
	}
	if got, want := values.ObjectKeys(result), []string{"scheme", "host", "path", "fragment", "query", "user", "password", "port"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("result ObjectKeys() = %v, want %v", got, want)
	}

	query, ok := result["query"].(Object)
	if !ok {
		t.Fatalf("query value = %T, want Object", result["query"])
	}
	if got, want := values.ObjectKeys(query), []string{"msg", "x"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("query ObjectKeys() = %v, want %v", got, want)
	}
}
