package input

import (
	"os"
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

func TestParseSourceRejectsOversizedStdin(t *testing.T) {
	parser := NewParser()

	tmp, err := os.CreateTemp("", "infomunge-stdin-*")
	if err != nil {
		t.Fatalf("create temp stdin file: %v", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := tmp.WriteString(strings.Repeat("a", MaxStdinBytes+1)); err != nil {
		t.Fatalf("write temp stdin file: %v", err)
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		t.Fatalf("rewind temp stdin file: %v", err)
	}

	oldStdin := os.Stdin
	os.Stdin = tmp
	defer func() {
		os.Stdin = oldStdin
	}()

	_, err = parser.ParseSource(":json")
	if err == nil {
		t.Fatal("expected oversized stdin to fail")
	}
	if !strings.Contains(err.Error(), "stdin input exceeds maximum size") {
		t.Fatalf("unexpected error: %v", err)
	}
}
