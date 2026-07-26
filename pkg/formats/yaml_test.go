package formats

import (
	stderrors "errors"
	unifiederrors "infomunge/internal/errors"
	"reflect"
	"strings"
	"testing"
)

func TestReadYAMLRejectsRecursiveAliases(t *testing.T) {
	_, err := Read("root: &root\n  self: *root\n", "application/yaml")
	assertYAMLValidationError(t, err, "recursive alias cycle")
}

func TestReadYAMLRejectsTrailingDocuments(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "second populated document",
			content: "---\nfirst: 1\n---\nsecond: 2\n",
		},
		{
			name:    "second empty document",
			content: "---\nfirst: 1\n---\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Read(tt.content, "application/yaml")
			assertYAMLValidationError(t, err, "multiple documents are not supported")
		})
	}
}

func TestReadYAMLPreservesNonCyclicAliasesAndMerges(t *testing.T) {
	result, err := Read(
		"defaults: &defaults\n  name: default\n  version: 1\nfirst: *defaults\nsecond: *defaults\nproduction:\n  <<: *defaults\n  name: production\n",
		"application/yaml",
	)
	if err != nil {
		t.Fatalf("Read(yaml) error = %v", err)
	}

	want := Object{
		"defaults": Object{"name": "default", "version": 1},
		"first":    Object{"name": "default", "version": 1},
		"second":   Object{"name": "default", "version": 1},
		"production": Object{
			"name":    "production",
			"version": 1,
		},
	}
	if !reflect.DeepEqual(result, want) {
		t.Fatalf("Read(yaml) = %#v, want %#v", result, want)
	}
}

func assertYAMLValidationError(t *testing.T, err error, messageFragment string) {
	t.Helper()
	if err == nil {
		t.Fatal("Read(yaml) error = nil, want validation error")
	}
	var typedErr *unifiederrors.Error
	if !stderrors.As(err, &typedErr) {
		t.Fatalf("Read(yaml) error type = %T, want *errors.Error", err)
	}
	if typedErr.Type != unifiederrors.TypeValidate {
		t.Fatalf("Read(yaml) error category = %s, want %s", typedErr.Type, unifiederrors.TypeValidate)
	}
	if !strings.Contains(typedErr.Message, messageFragment) {
		t.Fatalf("Read(yaml) error = %q, want it to contain %q", typedErr.Message, messageFragment)
	}
}
