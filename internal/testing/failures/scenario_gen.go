package failures

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"infomunge/internal/testing/teststore"
)

// GenerateScenario produces a candidate .feature snippet for a failure artifact.
func GenerateScenario(a Artifact) string {
	payload := "null"
	if data, err := json.MarshalIndent(a.InputPayload, "", "  "); err == nil {
		payload = string(data)
	}

	expr := strings.TrimSpace(a.MinimizedExpression)
	if expr == "" {
		expr = "/* TODO: fill minimized expression */"
	}

	script := "%im 0.1\noutput application/json\n---\n" + expr
	scenarioName := fmt.Sprintf(
		"Auto-generated failure %03d (%s)",
		max(1, a.ID),
		sanitizeLabel(a.Property),
	)

	return fmt.Sprintf(
		`# Auto-generated from intensive testing failure %03d
# Property: %s
# Seed: %d
# Shrunk expression: %s
Feature: Auto-generated intensive-testing failures

  Scenario: %s
    Given input payload is:
      """
%s
      """
    When infomunge processes:
      """
%s
      """
    Then verify expected result or error for property "%s"
`,
		max(1, a.ID),
		coalesce(a.Property, "unknown"),
		a.Seed,
		coalesce(expr, "unknown"),
		scenarioName,
		indentLines(payload, 6),
		indentLines(script, 6),
		coalesce(a.Property, "unknown"),
	)
}

// WriteCandidateScenario writes an auto-generated .feature file.
// If a candidate for the artifact already exists, it is skipped.
func WriteCandidateScenario(a Artifact) (string, bool, error) {
	dir, err := teststore.Path("candidates")
	if err != nil {
		return "", false, fmt.Errorf("resolve candidates dir: %w", err)
	}
	return WriteCandidateScenarioToDir(dir, a)
}

// WriteCandidateScenarioToDir writes an auto-generated .feature file to dir.
func WriteCandidateScenarioToDir(dir string, a Artifact) (string, bool, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", false, fmt.Errorf("create candidates dir: %w", err)
	}

	id := max(1, a.ID)
	name := fmt.Sprintf("%03d_%s.feature", id, sanitizeLabel(a.Property))
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); err == nil {
		return path, false, nil
	} else if !os.IsNotExist(err) {
		return "", false, fmt.Errorf("stat candidate: %w", err)
	}

	content := GenerateScenario(a)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", false, fmt.Errorf("write candidate: %w", err)
	}

	return path, true, nil
}

// GenerateAllCandidates creates candidate scenarios for artifacts that do not
// already have matching candidate files.
func GenerateAllCandidates() ([]string, error) {
	root, err := teststore.Root()
	if err != nil {
		return nil, fmt.Errorf("resolve intensive-testing dir: %w", err)
	}
	return GenerateAllCandidatesFromDirs(
		filepath.Join(root, "failures"),
		filepath.Join(root, "candidates"),
	)
}

// GenerateAllCandidatesFromDirs is the directory-configurable variant used by tests.
func GenerateAllCandidatesFromDirs(failuresDir, candidatesDir string) ([]string, error) {
	artifacts, err := LoadArtifactsFromDir(failuresDir)
	if err != nil {
		return nil, err
	}

	created := make([]string, 0, len(artifacts))
	for _, a := range artifacts {
		path, saved, err := WriteCandidateScenarioToDir(candidatesDir, a)
		if err != nil {
			return nil, err
		}
		if saved {
			created = append(created, path)
		}
	}
	return created, nil
}

func indentLines(s string, spaces int) string {
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}

func coalesce(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}
