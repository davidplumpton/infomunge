package test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

const successfulCLIScript = `%im 0.1
output application/json
---
1
`

func TestExpectCLIFailureRequiresRunErrorAndNonZeroExitCode(t *testing.T) {
	tests := []struct {
		name         string
		runErr       error
		lastExitCode int
		wantErr      bool
	}{
		{
			name:         "accepts run error with non-zero exit code",
			runErr:       errors.New("exit status 1"),
			lastExitCode: 1,
		},
		{
			name:         "rejects successful run",
			lastExitCode: 0,
			wantErr:      true,
		},
		{
			name:         "rejects failed run with zero exit code",
			runErr:       errors.New("failed without exit code"),
			lastExitCode: 0,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &testContext{lastExitCode: tt.lastExitCode}

			err := tc.expectCLIFailure(tt.runErr)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expectCLIFailure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAndItFailsStepsRejectSuccessfulCLI(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*testing.T, *testContext)
		run   func(*testContext) error
	}{
		{
			name: "file argument",
			run: func(tc *testContext) error {
				return tc.iRunTheApplicationWithAndItFails("input.txt")
			},
		},
		{
			name: "raw arguments",
			run: func(tc *testContext) error {
				return tc.iRunTheApplicationWithArgumentsAndItFails("-f input.txt")
			},
		},
		{
			name: "this content with lazy",
			setup: func(t *testing.T, tc *testContext) {
				tc.inputContent = successfulCLIScript
			},
			run: func(tc *testContext) error {
				return tc.iRunTheApplicationWithThisContentUsingLazyAndItFails()
			},
		},
		{
			name: "this content with inputs",
			setup: func(t *testing.T, tc *testContext) {
				tc.inputContent = successfulCLIScript
			},
			run: func(tc *testContext) error {
				return tc.iRunTheApplicationWithThisContentAndInputsAndItFails(&godog.DocString{})
			},
		},
		{
			name: "this content with stdin-backed inputs",
			setup: func(t *testing.T, tc *testContext) {
				tc.inputContent = successfulCLIScript
			},
			run: func(tc *testContext) error {
				return tc.iRunTheApplicationWithThisContentAndStdinBackedInputsAndItFails(&godog.DocString{})
			},
		},
		{
			name: "this content",
			setup: func(t *testing.T, tc *testContext) {
				tc.inputContent = successfulCLIScript
			},
			run: func(tc *testContext) error {
				return tc.iRunTheApplicationAndItFails()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newContextWithSuccessfulCLI()
			defer cleanupScenarioWorkspace(t, tc)
			if tt.setup != nil {
				tt.setup(t, tc)
			}

			err := tt.run(tc)
			if err == nil {
				t.Fatal("expected failure step to reject a successful CLI run")
			}
			if !strings.Contains(err.Error(), "expected application to fail") {
				t.Fatalf("expected success rejection error, got: %v", err)
			}
		})
	}
}

func newContextWithSuccessfulCLI() *testContext {
	tc := &testContext{}
	succeed := func() error {
		tc.lastStdout = "1\n"
		tc.lastStderr = ""
		tc.lastOutput = tc.lastStdout
		tc.lastExitCode = 0
		return nil
	}
	tc.runCLIOverride = func(args ...string) error {
		return succeed()
	}
	tc.runCLIWithStdinOverride = func(stdinContent string, args ...string) error {
		return succeed()
	}
	return tc
}

func cleanupScenarioWorkspace(t *testing.T, tc *testContext) {
	t.Helper()
	if tc.workDir == "" {
		return
	}
	if err := os.RemoveAll(tc.workDir); err != nil {
		t.Fatalf("RemoveAll(%q) error = %v", tc.workDir, err)
	}
	tc.workDir = ""
}
