package failures

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveAndLoadArtifactsRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	input := Artifact{
		Property:            "no-panic",
		MinimizedExpression: "payload.a + 1",
		OriginalExpression:  "((payload.a + payload.b) * 2) / 3",
		InputPayload: map[string]interface{}{
			"a": float64(1),
			"b": "x",
		},
		Expected: map[string]interface{}{
			"type": "Number",
		},
		Actual:     "panic",
		PanicStack: "stacktrace",
		Seed:       42,
	}

	path, saved, err := SaveArtifactToDir(dir, input)
	if err != nil {
		t.Fatalf("SaveArtifactToDir() error = %v", err)
	}
	if !saved {
		t.Fatalf("SaveArtifactToDir() saved = false, want true")
	}
	if filepath.Ext(path) != ".json" {
		t.Fatalf("artifact file extension = %q, want .json", filepath.Ext(path))
	}
	base := filepath.Base(path)
	if !strings.HasPrefix(base, "001_no-panic_") {
		t.Fatalf("artifact filename = %q, want prefix 001_no-panic_", base)
	}

	artifacts, err := LoadArtifactsFromDir(dir)
	if err != nil {
		t.Fatalf("LoadArtifactsFromDir() error = %v", err)
	}
	if len(artifacts) != 1 {
		t.Fatalf("len(artifacts) = %d, want 1", len(artifacts))
	}

	got := artifacts[0]
	if got.ID != 1 {
		t.Fatalf("artifact ID = %d, want 1", got.ID)
	}
	if got.Fingerprint != Fingerprint(input.MinimizedExpression, input.Property) {
		t.Fatalf("fingerprint mismatch: got %q", got.Fingerprint)
	}
	if got.Property != input.Property {
		t.Fatalf("property = %q, want %q", got.Property, input.Property)
	}
	if got.MinimizedExpression != input.MinimizedExpression {
		t.Fatalf("minimized expression = %q, want %q", got.MinimizedExpression, input.MinimizedExpression)
	}
	if got.OriginalExpression != input.OriginalExpression {
		t.Fatalf("original expression = %q, want %q", got.OriginalExpression, input.OriginalExpression)
	}
	if got.PanicStack != input.PanicStack {
		t.Fatalf("panic stack = %q, want %q", got.PanicStack, input.PanicStack)
	}
	if got.Seed != input.Seed {
		t.Fatalf("seed = %d, want %d", got.Seed, input.Seed)
	}

	wantInputJSON, err := json.Marshal(input.InputPayload)
	if err != nil {
		t.Fatalf("json.Marshal(input payload): %v", err)
	}
	gotInputJSON, err := json.Marshal(got.InputPayload)
	if err != nil {
		t.Fatalf("json.Marshal(got payload): %v", err)
	}
	if string(gotInputJSON) != string(wantInputJSON) {
		t.Fatalf("payload mismatch: got %s, want %s", gotInputJSON, wantInputJSON)
	}
}

func TestSaveArtifactDeduplicatesByFingerprint(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	first := Artifact{
		Property:            "no-panic",
		MinimizedExpression: "payload[0] + 1",
		OriginalExpression:  "payload[0] + payload[1] + payload[2]",
		InputPayload:        []interface{}{float64(1), float64(2)},
		Expected:            "result-or-error",
		Actual:              "panic",
		Seed:                100,
	}
	second := Artifact{
		Property:            "no-panic",
		MinimizedExpression: "payload[0] + 1",
		OriginalExpression:  "completely different source expression",
		InputPayload:        map[string]interface{}{"x": float64(9)},
		Expected:            "something else",
		Actual:              "different output",
		Seed:                999,
	}

	firstPath, firstSaved, err := SaveArtifactToDir(dir, first)
	if err != nil {
		t.Fatalf("first SaveArtifactToDir() error = %v", err)
	}
	if !firstSaved {
		t.Fatalf("first SaveArtifactToDir() saved = false, want true")
	}
	if firstPath == "" {
		t.Fatalf("first SaveArtifactToDir() path is empty")
	}

	secondPath, secondSaved, err := SaveArtifactToDir(dir, second)
	if err != nil {
		t.Fatalf("second SaveArtifactToDir() error = %v", err)
	}
	if secondSaved {
		t.Fatalf("second SaveArtifactToDir() saved = true, want false due to dedup")
	}
	if secondPath != "" {
		t.Fatalf("second SaveArtifactToDir() path = %q, want empty path for dedup skip", secondPath)
	}

	artifacts, err := LoadArtifactsFromDir(dir)
	if err != nil {
		t.Fatalf("LoadArtifactsFromDir() error = %v", err)
	}
	if len(artifacts) != 1 {
		t.Fatalf("len(artifacts) = %d, want 1 after dedup", len(artifacts))
	}
}
