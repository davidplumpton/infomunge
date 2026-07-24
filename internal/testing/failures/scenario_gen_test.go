package failures

import (
	"infomunge/internal/evaluator"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/gherkin/go/v26"
)

func TestGenerateScenarioProducesValidGherkin(t *testing.T) {
	t.Parallel()

	artifact := Artifact{
		ID:                  1,
		Property:            "no-panic",
		MinimizedExpression: "payload.a + 1",
		OriginalExpression:  "payload.a + payload.b",
		InputPayload: evaluator.Object{
			"a": float64(2),
		},
		Seed: 99,
	}

	scenario := GenerateScenario(artifact)
	if !strings.Contains(scenario, "# Property: no-panic") {
		t.Fatalf("scenario missing property comment:\n%s", scenario)
	}
	if !strings.Contains(scenario, "# Shrunk expression: payload.a + 1") {
		t.Fatalf("scenario missing shrunk expression comment:\n%s", scenario)
	}
	if !strings.Contains(scenario, "Given input payload is:") {
		t.Fatalf("scenario missing Given step:\n%s", scenario)
	}
	if !strings.Contains(scenario, "When infomunge processes:") {
		t.Fatalf("scenario missing When step:\n%s", scenario)
	}
	if !strings.Contains(scenario, "Then verify expected result or error") {
		t.Fatalf("scenario missing Then step:\n%s", scenario)
	}

	doc, err := gherkin.ParseGherkinDocument(strings.NewReader(scenario), func() string { return "id" })
	if err != nil {
		t.Fatalf("ParseGherkinDocument() error = %v\nscenario:\n%s", err, scenario)
	}
	if doc.Feature == nil {
		t.Fatalf("parsed document has nil Feature")
	}
	if len(doc.Feature.Children) == 0 {
		t.Fatalf("parsed feature has no scenarios")
	}
}

func TestGenerateAllCandidatesSkipsExisting(t *testing.T) {
	t.Parallel()

	failuresDir := t.TempDir()
	candidatesDir := t.TempDir()

	_, err := SaveArtifactToDir(failuresDir, Artifact{
		Property:            "no-panic",
		MinimizedExpression: "payload.a + 1",
		OriginalExpression:  "payload.a + payload.b",
		InputPayload:        evaluator.Object{"a": float64(1)},
		Seed:                1,
	})
	if err != nil {
		t.Fatalf("SaveArtifactToDir(first): %v", err)
	}
	_, err = SaveArtifactToDir(failuresDir, Artifact{
		Property:            "determinism",
		MinimizedExpression: "payload.b * 2",
		OriginalExpression:  "payload.b * 2 + 0",
		InputPayload:        evaluator.Object{"b": float64(2)},
		Seed:                2,
	})
	if err != nil {
		t.Fatalf("SaveArtifactToDir(second): %v", err)
	}

	existingPath, saved, err := WriteCandidateScenarioToDir(candidatesDir, Artifact{
		ID:                  1,
		Property:            "no-panic",
		MinimizedExpression: "payload.a + 1",
		InputPayload:        evaluator.Object{"a": float64(1)},
	})
	if err != nil {
		t.Fatalf("WriteCandidateScenarioToDir(existing): %v", err)
	}
	if !saved {
		t.Fatalf("expected initial candidate write to save")
	}
	if _, err := os.Stat(existingPath); err != nil {
		t.Fatalf("expected existing candidate file: %v", err)
	}

	created, err := GenerateAllCandidatesFromDirs(failuresDir, candidatesDir)
	if err != nil {
		t.Fatalf("GenerateAllCandidatesFromDirs() error = %v", err)
	}
	if len(created) != 1 {
		t.Fatalf("len(created) = %d, want 1", len(created))
	}
	if !strings.HasSuffix(created[0], "002_determinism.feature") {
		t.Fatalf("created candidate = %q, want suffix 002_determinism.feature", created[0])
	}
	if _, err := os.Stat(filepath.Join(candidatesDir, "001_no-panic.feature")); err != nil {
		t.Fatalf("expected existing candidate still present: %v", err)
	}
}

func TestSavedArtifactsGenerateDistinctNumberedCandidates(t *testing.T) {
	t.Parallel()

	failuresDir := t.TempDir()
	candidatesDir := t.TempDir()

	firstResult, err := SaveArtifactToDir(failuresDir, Artifact{
		Property:            "mutation-determinism",
		MinimizedExpression: "payload.a + 1",
		OriginalExpression:  "payload.a + 2",
		InputPayload:        evaluator.Object{"a": float64(1)},
		Seed:                1,
	})
	if err != nil {
		t.Fatalf("SaveArtifactToDir(first): %v", err)
	}
	if !firstResult.Saved {
		t.Fatal("SaveArtifactToDir(first) saved = false, want true")
	}

	duplicateResult, err := SaveArtifactToDir(failuresDir, Artifact{
		Property:            "mutation-determinism",
		MinimizedExpression: "payload.a + 1",
		OriginalExpression:  "different original expression",
		InputPayload:        evaluator.Object{"a": float64(99)},
		Seed:                99,
	})
	if err != nil {
		t.Fatalf("SaveArtifactToDir(duplicate): %v", err)
	}
	if duplicateResult.Saved {
		t.Fatal("SaveArtifactToDir(duplicate) saved = true, want fingerprint deduplication")
	}
	if duplicateResult.Artifact.ID != firstResult.Artifact.ID {
		t.Fatalf(
			"duplicate artifact ID = %d, want canonical ID %d",
			duplicateResult.Artifact.ID,
			firstResult.Artifact.ID,
		)
	}

	firstCandidate, saved, err := WriteCandidateScenarioToDir(candidatesDir, duplicateResult.Artifact)
	if err != nil {
		t.Fatalf("WriteCandidateScenarioToDir(first): %v", err)
	}
	if !saved {
		t.Fatal("WriteCandidateScenarioToDir(first) saved = false, want true")
	}

	secondResult, err := SaveArtifactToDir(failuresDir, Artifact{
		Property:            "mutation-determinism",
		MinimizedExpression: "payload.b * 2",
		OriginalExpression:  "payload.b * 3",
		InputPayload:        evaluator.Object{"b": float64(2)},
		Seed:                2,
	})
	if err != nil {
		t.Fatalf("SaveArtifactToDir(second): %v", err)
	}
	if !secondResult.Saved {
		t.Fatal("SaveArtifactToDir(second) saved = false, want true")
	}

	secondCandidate, saved, err := WriteCandidateScenarioToDir(candidatesDir, secondResult.Artifact)
	if err != nil {
		t.Fatalf("WriteCandidateScenarioToDir(second): %v", err)
	}
	if !saved {
		t.Fatal("WriteCandidateScenarioToDir(second) saved = false, want true")
	}

	if got := filepath.Base(firstCandidate); got != "001_mutation-determinism.feature" {
		t.Fatalf("first candidate = %q, want 001_mutation-determinism.feature", got)
	}
	if got := filepath.Base(secondCandidate); got != "002_mutation-determinism.feature" {
		t.Fatalf("second candidate = %q, want 002_mutation-determinism.feature", got)
	}
}
