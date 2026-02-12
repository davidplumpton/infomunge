package mutation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractCorpusSupportsTemplateAndNamedInputs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	featurePath := filepath.Join(dir, "template.feature")
	content := `Feature: Template extraction

  Scenario: Extract generated template scenario
    Given input payload is:
      """
      {"name":"Alice"}
      """
    And input profile is:
      """
      {"active":true}
      """
    When infomunge processes:
      """
      %im 0.1
      output application/json
      ---
      payload.name
      """
    Then the output should be:
      """
      "Alice"
      """
`
	if err := os.WriteFile(featurePath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	entries, err := ExtractCorpus(dir)
	if err != nil {
		t.Fatalf("ExtractCorpus() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}

	entry := entries[0]
	if entry.ScenarioName != "Extract generated template scenario" {
		t.Fatalf("ScenarioName = %q", entry.ScenarioName)
	}
	if !strings.Contains(entry.Script, "payload.name") {
		t.Fatalf("Script = %q, expected payload.name", entry.Script)
	}
	if got := entry.Inputs["payload"]; !strings.Contains(got, "Alice") {
		t.Fatalf("Inputs[payload] = %q, expected Alice payload", got)
	}
	if got := entry.Inputs["profile"]; !strings.Contains(got, "active") {
		t.Fatalf("Inputs[profile] = %q, expected profile input", got)
	}
	if strings.TrimSpace(entry.ExpectedOutput) != `"Alice"` {
		t.Fatalf("ExpectedOutput = %q", entry.ExpectedOutput)
	}
}

func TestExtractCorpusRealFeaturesSanity(t *testing.T) {
	t.Parallel()

	featuresDir := filepath.Join("..", "..", "..", "test", "features")
	entries, err := ExtractCorpus(featuresDir)
	if err != nil {
		t.Fatalf("ExtractCorpus() error = %v", err)
	}

	if len(entries) < 200 {
		t.Fatalf("len(entries) = %d, expected at least 200", len(entries))
	}

	for _, entry := range entries {
		if strings.TrimSpace(entry.Script) == "" {
			t.Fatalf("entry from %s scenario %q has empty script", entry.SourceFile, entry.ScenarioName)
		}
		if strings.TrimSpace(entry.ExpectedOutput) == "" {
			t.Fatalf("entry from %s scenario %q has empty expected output", entry.SourceFile, entry.ScenarioName)
		}
	}
}
