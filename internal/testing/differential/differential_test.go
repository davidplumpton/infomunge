package differential

import (
	"context"
	"errors"
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

type infomungeOnlyErrorAllowance struct {
	issueID        string
	errorSubstring string
}

// infomungeOnlyErrorAllowlist is deliberately limited to known, ticketed
// compatibility gaps. Every matching case is still counted and saved as an
// artifact. Remove each entry when its issue is closed.
var infomungeOnlyErrorAllowlist []infomungeOnlyErrorAllowance

type differentialOutcome int

const (
	outcomeBothErrors differentialOutcome = iota
	outcomeInfomungeOnlyError
	outcomeDataWeaveOnlyError
	outcomeBothSuccess
)

type differentialOutcomes struct {
	generated                      int
	bothErrors                     int
	infomungeOnlyErrors            int
	allowlistedInfomungeOnlyErrors int
	dataWeaveOnlyErrors            int
	bothSuccess                    int
	artifactSaveErrors             int
	firstInfomungeError            string
	firstDataWeaveError            string
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
		tc := exprgen.SampleContext().Draw(t, "ctx")
		expr := exprgen.DWCompatExpression(3).Draw(t, "expr")

		script := exprgen.WrapDWScript(expr)

		imResult, imErr := safeEvalInfomunge(script, tc.Value)
		if imErr != nil && strings.Contains(strings.ToLower(imErr.Error()), "panic:") {
			metrics.RecordPanic()
		}

		dwResult, dwErr := DWEval(script, map[string]string{"payload": tc.JSON})
		switch outcomes.record(imErr, dwErr) {
		case outcomeBothErrors:
			return
		case outcomeInfomungeOnlyError:
			artifact := failures.Artifact{
				Property:            "differential-infomunge-only-error",
				MinimizedExpression: expr,
				OriginalExpression:  script,
				InputPayload:        tc.Value["payload"],
				Expected:            dwResult,
				Actual:              imErr.Error(),
			}
			if _, saveErr := failures.SaveArtifact(artifact); saveErr != nil {
				outcomes.artifactSaveErrors++
				t.Logf("failed to save artifact: %v", saveErr)
			}
			if issueID, allowed := allowlistedInfomungeOnlyError(imErr); allowed {
				outcomes.allowlistedInfomungeOnlyErrors++
				t.Logf("allowlisted InfoMunge-only error tracked by %s: %v", issueID, imErr)
			}
			return
		case outcomeDataWeaveOnlyError:
			return
		case outcomeBothSuccess:
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
		}
	})

	t.Logf("differential outcomes: %s", outcomes)
	if outcomes.firstInfomungeError != "" {
		t.Logf("first InfoMunge error: %s", outcomes.firstInfomungeError)
	}
	if outcomes.firstDataWeaveError != "" {
		t.Logf("first DataWeave error: %s", outcomes.firstDataWeaveError)
	}
	if err := outcomes.validate(checks); err != nil {
		t.Fatal(err)
	}
}

func TestDifferential_ObjectSubtractionMatchesDataWeave(t *testing.T) {
	if !DWAvailable() {
		t.Skip("dw CLI not available on PATH")
	}

	supported := []string{
		`{"remove": 1, "keep": 2} - "remove"`,
		`{"0": 1, "keep": 2} - 0.0`,
		`{"1.5": 1, "keep": 2} - 1.5`,
		`{"true": 1, "keep": 2} - true`,
		`{"keep": 2} - false`,
	}
	for _, expr := range supported {
		t.Run("supported "+expr, func(t *testing.T) {
			script := exprgen.WrapDWScript(expr)
			imResult, imErr := safeEvalInfomunge(script, evaluator.Context{})
			if imErr != nil {
				t.Fatalf("InfoMunge returned an error: %v", imErr)
			}
			dwResult, dwErr := DWEval(script, nil)
			if dwErr != nil {
				t.Fatalf("DataWeave returned an error: %v", dwErr)
			}
			if err := StructuralCompare(imResult, dwResult); err != nil {
				t.Fatalf("object subtraction differs: %v", err)
			}
		})
	}

	rejected := []string{
		`{"a": 1} - null`,
		`{"a": 1} - ["a"]`,
		`{"a": 1} - {"a": 1}`,
	}
	for _, expr := range rejected {
		t.Run("rejected "+expr, func(t *testing.T) {
			script := exprgen.WrapDWScript(expr)
			if _, err := safeEvalInfomunge(script, evaluator.Context{}); err == nil {
				t.Fatal("InfoMunge unexpectedly accepted the operand")
			}
			if _, err := DWEval(script, nil); err == nil {
				t.Fatal("DataWeave unexpectedly accepted the operand")
			}
		})
	}
}

func TestDifferential_NestedTypeOfMatchesDataWeave(t *testing.T) {
	if !DWAvailable() {
		t.Skip("dw CLI not available on PATH")
	}

	script := exprgen.WrapDWScript(`typeOf(typeOf(false))`)
	imResult, imErr := safeEvalInfomunge(script, evaluator.Context{})
	if imErr != nil {
		t.Fatalf("InfoMunge returned an error: %v", imErr)
	}
	dwResult, dwErr := DWEval(script, nil)
	if dwErr != nil {
		t.Fatalf("DataWeave returned an error: %v", dwErr)
	}
	if err := StructuralCompare(imResult, dwResult); err != nil {
		t.Fatalf("nested typeOf differs: %v", err)
	}
}

func TestDifferential_SizeOfBooleansMatchesDataWeave(t *testing.T) {
	if !DWAvailable() {
		t.Skip("dw CLI not available on PATH")
	}

	script := exprgen.WrapDWScript(`[sizeOf(true), sizeOf(false)]`)
	imResult, imErr := safeEvalInfomunge(script, evaluator.Context{})
	if imErr != nil {
		t.Fatalf("InfoMunge returned an error: %v", imErr)
	}
	dwResult, dwErr := DWEval(script, nil)
	if dwErr != nil {
		t.Fatalf("DataWeave returned an error: %v", dwErr)
	}
	if err := StructuralCompare(imResult, dwResult); err != nil {
		t.Fatalf("boolean sizeOf differs: %v", err)
	}
}

func (outcomes *differentialOutcomes) record(imErr, dwErr error) differentialOutcome {
	outcomes.generated++
	if imErr != nil && outcomes.firstInfomungeError == "" {
		outcomes.firstInfomungeError = imErr.Error()
	}
	if dwErr != nil && outcomes.firstDataWeaveError == "" {
		outcomes.firstDataWeaveError = dwErr.Error()
	}

	switch {
	case imErr != nil && dwErr != nil:
		outcomes.bothErrors++
		return outcomeBothErrors
	case imErr != nil:
		outcomes.infomungeOnlyErrors++
		return outcomeInfomungeOnlyError
	case dwErr != nil:
		outcomes.dataWeaveOnlyErrors++
		return outcomeDataWeaveOnlyError
	default:
		outcomes.bothSuccess++
		return outcomeBothSuccess
	}
}

func allowlistedInfomungeOnlyError(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	message := err.Error()
	for _, allowance := range infomungeOnlyErrorAllowlist {
		if strings.Contains(message, allowance.errorSubstring) {
			return allowance.issueID, true
		}
	}
	return "", false
}

func (outcomes differentialOutcomes) String() string {
	return fmt.Sprintf(
		"generated=%d both_errors=%d infomunge_only_errors=%d allowlisted_infomunge_only_errors=%d dataweave_only_errors=%d both_success=%d artifact_save_errors=%d",
		outcomes.generated,
		outcomes.bothErrors,
		outcomes.infomungeOnlyErrors,
		outcomes.allowlistedInfomungeOnlyErrors,
		outcomes.dataWeaveOnlyErrors,
		outcomes.bothSuccess,
		outcomes.artifactSaveErrors,
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

	accounted := outcomes.bothErrors + outcomes.infomungeOnlyErrors + outcomes.dataWeaveOnlyErrors + outcomes.bothSuccess
	if accounted != outcomes.generated {
		return fmt.Errorf(
			"incomplete differential outcome accounting: %s; accounted for %d cases",
			outcomes,
			accounted,
		)
	}

	if outcomes.allowlistedInfomungeOnlyErrors > outcomes.infomungeOnlyErrors {
		return fmt.Errorf(
			"invalid differential allowlist accounting: %s",
			outcomes,
		)
	}
	if outcomes.artifactSaveErrors > 0 {
		return fmt.Errorf(
			"failed to save %d differential failure artifact(s): %s",
			outcomes.artifactSaveErrors,
			outcomes,
		)
	}

	unallowlistedInfomungeOnlyErrors := outcomes.infomungeOnlyErrors - outcomes.allowlistedInfomungeOnlyErrors
	if unallowlistedInfomungeOnlyErrors > 0 {
		return fmt.Errorf(
			"differential mismatch: %s; %d case(s) failed only in InfoMunge; see saved failure artifacts",
			outcomes,
			unallowlistedInfomungeOnlyErrors,
		)
	}

	minimumComparisons := outcomes.generated * minimumComparisonPercent / 100
	if minimumComparisons < 1 {
		minimumComparisons = 1
	}
	if outcomes.bothSuccess >= minimumComparisons {
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

func TestDifferentialOutcomes_RecordsEveryRuntimeCombination(t *testing.T) {
	imErr := errors.New("infomunge failed")
	dwErr := errors.New("DataWeave failed")
	testCases := []struct {
		name    string
		imErr   error
		dwErr   error
		outcome differentialOutcome
	}{
		{name: "both error", imErr: imErr, dwErr: dwErr, outcome: outcomeBothErrors},
		{name: "only InfoMunge errors", imErr: imErr, outcome: outcomeInfomungeOnlyError},
		{name: "only DataWeave errors", dwErr: dwErr, outcome: outcomeDataWeaveOnlyError},
		{name: "both succeed", outcome: outcomeBothSuccess},
	}

	var outcomes differentialOutcomes
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := outcomes.record(tc.imErr, tc.dwErr); got != tc.outcome {
				t.Fatalf("record() = %v, want %v", got, tc.outcome)
			}
		})
	}

	if outcomes.generated != 4 ||
		outcomes.bothErrors != 1 ||
		outcomes.infomungeOnlyErrors != 1 ||
		outcomes.dataWeaveOnlyErrors != 1 ||
		outcomes.bothSuccess != 1 {
		t.Fatalf("unexpected outcome accounting: %s", outcomes)
	}
}

func TestDifferentialOutcomes_RejectsEveryGeneratedCaseSkipped(t *testing.T) {
	outcomes := differentialOutcomes{
		generated:           50,
		bothErrors:          25,
		dataWeaveOnlyErrors: 25,
	}

	err := outcomes.validate(50)
	if err == nil {
		t.Fatal("expected an all-skipped differential run to fail")
	}
	if !strings.Contains(err.Error(), "both_success=0") {
		t.Fatalf("expected outcome counts in error, got: %v", err)
	}
}

func TestDifferentialOutcomes_RejectsUnaccountedGeneratedCase(t *testing.T) {
	outcomes := differentialOutcomes{
		generated:           50,
		bothErrors:          10,
		dataWeaveOnlyErrors: 10,
		bothSuccess:         29,
	}

	err := outcomes.validate(50)
	if err == nil {
		t.Fatal("expected incomplete outcome accounting to fail")
	}
	if !strings.Contains(err.Error(), "accounted for 49 cases") {
		t.Fatalf("expected accounting details in error, got: %v", err)
	}
}

func TestDifferentialOutcomes_RejectsInfomungeOnlyErrors(t *testing.T) {
	outcomes := differentialOutcomes{
		generated:           50,
		infomungeOnlyErrors: 1,
		bothSuccess:         49,
	}

	err := outcomes.validate(50)
	if err == nil {
		t.Fatal("expected an InfoMunge-only error to fail the differential gate")
	}
	if !strings.Contains(err.Error(), "failed only in InfoMunge") {
		t.Fatalf("expected one-sided failure details, got: %v", err)
	}
}

func TestDifferentialOutcomes_AcceptsTicketedInfomungeOnlyErrors(t *testing.T) {
	outcomes := differentialOutcomes{
		generated:                      50,
		infomungeOnlyErrors:            2,
		allowlistedInfomungeOnlyErrors: 2,
		bothSuccess:                    48,
	}

	if err := outcomes.validate(50); err != nil {
		t.Fatalf("expected ticketed compatibility gaps to remain visible but allowed, got: %v", err)
	}
}

func TestDifferentialOutcomes_RejectsMissingOneSidedArtifacts(t *testing.T) {
	outcomes := differentialOutcomes{
		generated:                      50,
		infomungeOnlyErrors:            1,
		allowlistedInfomungeOnlyErrors: 1,
		bothSuccess:                    49,
		artifactSaveErrors:             1,
	}

	err := outcomes.validate(50)
	if err == nil {
		t.Fatal("expected an artifact persistence failure to fail the differential gate")
	}
	if !strings.Contains(err.Error(), "failed to save 1 differential failure artifact") {
		t.Fatalf("expected artifact persistence details, got: %v", err)
	}
}

func TestDifferentialOutcomes_RejectsImpossibleAllowlistAccounting(t *testing.T) {
	outcomes := differentialOutcomes{
		generated:                      50,
		allowlistedInfomungeOnlyErrors: 1,
		bothSuccess:                    50,
	}

	err := outcomes.validate(50)
	if err == nil {
		t.Fatal("expected impossible allowlist accounting to fail")
	}
	if !strings.Contains(err.Error(), "invalid differential allowlist accounting") {
		t.Fatalf("expected allowlist accounting details, got: %v", err)
	}
}

func TestDifferentialOutcomes_AcceptsAccountedRunWithSufficientComparisons(t *testing.T) {
	outcomes := differentialOutcomes{
		generated:           50,
		bothErrors:          20,
		dataWeaveOnlyErrors: 20,
		bothSuccess:         10,
	}

	if err := outcomes.validate(50); err != nil {
		t.Fatalf("expected sufficient differential coverage, got: %v", err)
	}
}

func TestAllowlistedInfomungeOnlyError_RejectsClosedCompatibilityGap(t *testing.T) {
	testCases := []struct {
		message string
		issueID string
		allowed bool
	}{
		{message: "4:1: sizeOf: unsupported type int", allowed: false},
		{message: "4:1: lower expects a string", allowed: false},
		{message: "4:1: unexpected new evaluator failure", allowed: false},
	}

	for _, tc := range testCases {
		t.Run(tc.message, func(t *testing.T) {
			issueID, allowed := allowlistedInfomungeOnlyError(errors.New(tc.message))
			if issueID != tc.issueID || allowed != tc.allowed {
				t.Fatalf(
					"allowlistedInfomungeOnlyError() = (%q, %t), want (%q, %t)",
					issueID,
					allowed,
					tc.issueID,
					tc.allowed,
				)
			}
		})
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
