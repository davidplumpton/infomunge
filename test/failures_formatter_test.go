package test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"
)

func TestGodogHarnessFailsOnUndefinedStep(t *testing.T) {
	t.Setenv("GODOG_FORMAT", failuresFormatName)
	t.Setenv("GODOG_CONCURRENCY", "1")

	featurePath := filepath.Join(t.TempDir(), "undefined.feature")
	feature := []byte(`Feature: Strict step enforcement
  Scenario: Undefined steps fail the suite
    Given this step has deliberately not been registered
`)
	if err := os.WriteFile(featurePath, feature, 0o600); err != nil {
		t.Fatalf("write test feature: %v", err)
	}

	var out bytes.Buffer
	opts := newGodogOptions([]string{featurePath}, &out)
	suite := godog.TestSuite{
		ScenarioInitializer: func(*godog.ScenarioContext) {},
		Options:             opts,
	}

	if status := suite.Run(); status == 0 {
		t.Fatalf("strict Godog suite returned status 0 for an undefined step:\n%s", out.String())
	}
	assertContainsAll(t, out.String(),
		"Non-passing steps:",
		"undefined: Undefined steps fail the suite",
		"this step has deliberately not been registered",
	)
}

func TestFailuresFormatterReportsUndefinedStep(t *testing.T) {
	var out bytes.Buffer
	formatter := &failuresFormatter{out: &out, passInterval: defaultPassInterval}

	formatter.Undefined(testPickle("Undefined scenario", "features/example.feature"), testStep("an unregistered step"), nil)
	formatter.Summary()

	got := out.String()
	assertContainsAll(t, got,
		"Non-passing steps:",
		"1) undefined: Undefined scenario (features/example.feature)",
		"   an unregistered step",
	)
	if strings.Contains(got, "error:") {
		t.Fatalf("undefined step should not report an unavailable error, got:\n%s", got)
	}
}

func TestFailuresFormatterReportsAmbiguousStepError(t *testing.T) {
	var out bytes.Buffer
	formatter := &failuresFormatter{out: &out, passInterval: defaultPassInterval}

	formatter.Ambiguous(
		testPickle("Ambiguous scenario", "features/ambiguous.feature"),
		testStep("I match two definitions"),
		nil,
		errors.New("ambiguous step definitions"),
	)
	formatter.Summary()

	assertContainsAll(t, out.String(),
		"Non-passing steps:",
		"1) ambiguous: Ambiguous scenario (features/ambiguous.feature)",
		"   I match two definitions",
		"   error: ambiguous step definitions",
	)
}

func TestFailuresFormatterReportsSkippedAndPendingSteps(t *testing.T) {
	var out bytes.Buffer
	formatter := &failuresFormatter{out: &out, passInterval: defaultPassInterval}

	formatter.Skipped(testPickle("Skipped scenario", "features/skipped.feature"), testStep("a skipped step"), nil)
	formatter.Pending(testPickle("Pending scenario", "features/pending.feature"), testStep("a pending step"), nil)
	formatter.Summary()

	assertContainsAll(t, out.String(),
		"1) skipped: Skipped scenario (features/skipped.feature)",
		"   a skipped step",
		"2) pending: Pending scenario (features/pending.feature)",
		"   a pending step",
	)
}

func TestFailuresFormatterPassingSummaryRemainsConcise(t *testing.T) {
	var out bytes.Buffer
	formatter := &failuresFormatter{out: &out, passInterval: defaultPassInterval}

	formatter.Passed(testPickle("Passing scenario", "features/passing.feature"), testStep("a passing step"), nil)
	formatter.Summary()

	if got := out.String(); got != "passed steps: 1\n" {
		t.Fatalf("passing summary output = %q, want concise pass count", got)
	}
}

func testPickle(name, uri string) *messages.Pickle {
	return &messages.Pickle{Name: name, Uri: uri}
}

func testStep(text string) *messages.PickleStep {
	return &messages.PickleStep{Text: text}
}

func assertContainsAll(t *testing.T, got string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
}
