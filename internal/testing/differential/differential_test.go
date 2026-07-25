package differential

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"testing"

	"infomunge/internal/evaluator"
	"infomunge/internal/runner"
	"infomunge/internal/testing/exprgen"
	"infomunge/internal/testing/failures"
	"infomunge/internal/testing/metrics"
	"infomunge/internal/testing/testbudget"

	"pgregory.net/rapid"
)

const minimumComparisonPercent = 20

type differentialOutcomes struct {
	generated           int
	infomungeErrors     int
	dataWeaveErrors     int
	structuralChecks    int
	firstInfomungeError string
	firstDataWeaveError string
}

// TestDifferential_InfomungeVsDataWeave generates DW-compatible expressions,
// evaluates them in both infomunge and the DataWeave CLI, and compares results
// using StructuralCompare. Mismatches are saved as failure artifacts.
func TestDifferential_InfomungeVsDataWeave(t *testing.T) {
	if !DWAvailable() {
		t.Skip("dw CLI not available on PATH")
	}

	checks := testbudget.DiffChecks()
	setRapidChecks(t, checks)
	var outcomes differentialOutcomes

	rapid.Check(t, func(t *rapid.T) {
		outcomes.generated++
		tc := exprgen.SampleContext().Draw(t, "ctx")
		expr := exprgen.DWCompatExpression(3).Draw(t, "expr")

		script := exprgen.WrapDWScript(expr)

		// Evaluate in infomunge.
		imResult, imErr := safeEvalInfomunge(script, tc.Value)
		if imErr != nil {
			if strings.Contains(strings.ToLower(imErr.Error()), "panic:") {
				metrics.RecordPanic()
			}
			outcomes.infomungeErrors++
			if outcomes.firstInfomungeError == "" {
				outcomes.firstInfomungeError = imErr.Error()
			}
			// Expression errors (type errors, div-by-zero, etc.) are not
			// differential mismatches — skip.
			return
		}

		// Evaluate in DataWeave.
		dwResult, dwErr := DWEval(script, map[string]string{"payload": tc.JSON})
		if dwErr != nil {
			outcomes.dataWeaveErrors++
			if outcomes.firstDataWeaveError == "" {
				outcomes.firstDataWeaveError = dwErr.Error()
			}
			// DW may be stricter or have different type coercion; skip DW errors.
			return
		}

		// Compare results structurally.
		outcomes.structuralChecks++
		if err := StructuralCompare(imResult, dwResult); err != nil {
			artifact := failures.Artifact{
				Property:            "differential-dw",
				MinimizedExpression: expr,
				OriginalExpression:  script,
				InputPayload:        tc.Value["payload"],
				Expected:            dwResult,
				Actual:              imResult,
			}
			if _, saveErr := failures.SaveArtifact(artifact); saveErr != nil {
				t.Logf("failed to save artifact: %v", saveErr)
			}

			t.Fatalf(
				"differential mismatch\nexpr: %s\npayload: %s\ninfomunge: %#v\ndw:       %#v\ndiff: %v",
				expr, tc.JSON, imResult, dwResult, err,
			)
		}
	})

	t.Logf("differential outcomes: %s", outcomes)
	if outcomes.firstInfomungeError != "" {
		t.Logf("first infomunge skip: %s", outcomes.firstInfomungeError)
	}
	if outcomes.firstDataWeaveError != "" {
		t.Logf("first DataWeave skip: %s", outcomes.firstDataWeaveError)
	}
	if err := outcomes.validate(checks); err != nil {
		t.Fatal(err)
	}
}

func (outcomes differentialOutcomes) String() string {
	return fmt.Sprintf(
		"generated=%d compared=%d infomunge_errors=%d dataweave_errors=%d",
		outcomes.generated,
		outcomes.structuralChecks,
		outcomes.infomungeErrors,
		outcomes.dataWeaveErrors,
	)
}

func (outcomes differentialOutcomes) validate(expectedChecks int) error {
	if outcomes.generated < expectedChecks {
		return fmt.Errorf(
			"insufficient differential generation: %s; expected at least %d generated checks",
			outcomes,
			expectedChecks,
		)
	}

	accounted := outcomes.infomungeErrors + outcomes.dataWeaveErrors + outcomes.structuralChecks
	if accounted != outcomes.generated {
		return fmt.Errorf(
			"incomplete differential outcome accounting: %s; accounted for %d cases",
			outcomes,
			accounted,
		)
	}

	minimumComparisons := outcomes.generated * minimumComparisonPercent / 100
	if minimumComparisons < 1 {
		minimumComparisons = 1
	}
	if outcomes.structuralChecks >= minimumComparisons {
		return nil
	}
	return fmt.Errorf(
		"insufficient differential coverage: %s; need at least %d structural comparisons (%d%% of %d generated checks)",
		outcomes,
		minimumComparisons,
		minimumComparisonPercent,
		outcomes.generated,
	)
}

func TestDifferentialOutcomes_RejectsEveryGeneratedCaseSkipped(t *testing.T) {
	outcomes := differentialOutcomes{
		generated:       50,
		infomungeErrors: 25,
		dataWeaveErrors: 25,
	}

	err := outcomes.validate(50)
	if err == nil {
		t.Fatal("expected an all-skipped differential run to fail")
	}
	if !strings.Contains(err.Error(), "compared=0") {
		t.Fatalf("expected outcome counts in error, got: %v", err)
	}
}

func TestDifferentialOutcomes_RejectsUnaccountedGeneratedCase(t *testing.T) {
	outcomes := differentialOutcomes{
		generated:        50,
		infomungeErrors:  10,
		dataWeaveErrors:  10,
		structuralChecks: 29,
	}

	err := outcomes.validate(50)
	if err == nil {
		t.Fatal("expected incomplete outcome accounting to fail")
	}
	if !strings.Contains(err.Error(), "accounted for 49 cases") {
		t.Fatalf("expected accounting details in error, got: %v", err)
	}
}

// safeEvalInfomunge evaluates a script in infomunge with panic recovery.
func safeEvalInfomunge(script string, ctx evaluator.Context) (result evaluator.Value, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("panic: %v\n%s", recovered, debug.Stack())
		}
	}()
	execution, err := runner.ExecuteString(context.Background(), script, ctx, runner.RunnerOptions{})
	if err != nil {
		return nil, err
	}
	resolved, err := execution.Resolved()
	if err != nil {
		return nil, err
	}
	return resolved.Value, nil
}

func setRapidChecks(t *testing.T, checks int) {
	t.Helper()
	f := flag.Lookup("rapid.checks")
	if f == nil {
		t.Fatal("rapid.checks flag not found")
	}
	previous := f.Value.String()
	if err := f.Value.Set(fmt.Sprintf("%d", checks)); err != nil {
		t.Fatalf("set rapid.checks=%d: %v", checks, err)
	}
	t.Cleanup(func() {
		_ = f.Value.Set(previous)
	})
}

func TestMain(m *testing.M) {
	code := m.Run()
	metrics.ReportAndPersist("differential", metrics.Options{EnableCoverage: false})
	os.Exit(code)
}
