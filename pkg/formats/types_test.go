package formats

import (
	"reflect"
	"testing"

	"infomunge/pkg/values"
)

func TestPublicValueAliasesUseSharedRuntimeTypes(t *testing.T) {
	formatted := Object{
		"items": Array{
			Object{"name": "alpha"},
			XMLMultiValue{"beta", "gamma"},
			Namespace{Prefix: "ns", URI: "https://example.test/ns"},
		},
	}

	var shared values.Object = formatted
	items, ok := shared["items"].(values.Array)
	if !ok {
		t.Fatalf("expected values.Array, got %T", shared["items"])
	}

	want := values.Array{
		values.Object{"name": "alpha"},
		values.XMLMultiValue{"beta", "gamma"},
		values.Namespace{Prefix: "ns", URI: "https://example.test/ns"},
	}
	if !reflect.DeepEqual(items, want) {
		t.Fatalf("shared runtime aliases changed\n got: %#v\nwant: %#v", items, want)
	}
}
