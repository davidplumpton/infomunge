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

// TestDifferential_InfomungeVsDataWeave generates DW-compatible expressions,
// evaluates them in both infomunge and the DataWeave CLI, and compares results
// using StructuralCompare. Mismatches are saved as failure artifacts.
func TestDifferential_InfomungeVsDataWeave(t *testing.T) {
	if !DWAvailable() {
		t.Skip("dw CLI not available on PATH")
	}

	setRapidChecks(t, testbudget.DiffChecks())

	rapid.Check(t, func(t *rapid.T) {
		tc := exprgen.SampleContext().Draw(t, "ctx")
		expr := exprgen.DWCompatExpression(3).Draw(t, "expr")

		script := exprgen.WrapDWScript(expr)

		// Evaluate in infomunge.
		imResult, imErr := safeEvalInfomunge(script, tc.Value)
		if imErr != nil {
			if strings.Contains(strings.ToLower(imErr.Error()), "panic:") {
				metrics.RecordPanic()
			}
			// Expression errors (type errors, div-by-zero, etc.) are not
			// differential mismatches — skip.
			return
		}

		// Evaluate in DataWeave.
		dwResult, dwErr := DWEval(script, map[string]string{"payload": tc.JSON})
		if dwErr != nil {
			// DW may be stricter or have different type coercion; skip DW errors.
			return
		}

		// Compare results structurally.
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
