package input

import (
	"strings"
	"testing"
)

func TestParseInputsRejectsMultipleStdinBackedSources(t *testing.T) {
	parser := NewParser()

	_, err := parser.ParseInputs([]string{
		"payload=:json",
		"extra=:json",
	})
	if err == nil {
		t.Fatal("expected error for multiple stdin-backed inputs")
	}
	if !strings.Contains(err.Error(), "multiple stdin-backed inputs are not supported") {
		t.Fatalf("expected multiple-stdin error, got %v", err)
	}
}
