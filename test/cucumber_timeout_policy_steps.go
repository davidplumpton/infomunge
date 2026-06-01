package test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func (tc *testContext) theMakefileCucumberTargetsShouldUseAFiveMinuteGoTestTimeout() error {
	targets, err := readMakefileTargets("../Makefile")
	if err != nil {
		return err
	}

	expected := map[string]string{
		"test":          "GODOG_FORMAT=failures go test -v ./test -timeout 5m",
		"test-verbose":  "GODOG_FORMAT=progress go test -v ./test -timeout 5m",
		"test-failures": "GODOG_FORMAT=failures go test -v ./test -timeout 5m",
		"test-feature":  "GODOG_PATHS=features/$(FEATURE).feature go test -v ./test -run TestFeatures -timeout 5m",
	}
	for target, command := range expected {
		if !containsString(targets[target], command) {
			return fmt.Errorf("Makefile target %q should include %q, got %v", target, command, targets[target])
		}
	}
	return nil
}

func (tc *testContext) theRepoWideGoTestRegressionStepShouldUseAFiveMinuteGoTestTimeout() error {
	if repoWideGoTestTimeout != 5*time.Minute {
		return fmt.Errorf("repo-wide go test timeout = %s, want 5m0s", repoWideGoTestTimeout)
	}
	if !hasTimeoutArg(repoWideGoTestArgs(), repoWideGoTestTimeoutArg) {
		return fmt.Errorf("repo-wide go test args should include -timeout %s, got %v", repoWideGoTestTimeoutArg, repoWideGoTestArgs())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := repoWideGoTestCommand(ctx)
	if cmd.Dir != ".." {
		return fmt.Errorf("repo-wide go test command dir = %q, want %q", cmd.Dir, "..")
	}
	if !containsString(cmd.Env, "INFOMUNGE_SKIP_GODOG=1") {
		return fmt.Errorf("repo-wide go test command should preserve INFOMUNGE_SKIP_GODOG=1")
	}
	return nil
}

func (tc *testContext) theTestingDocsShouldShowCucumberCommandsWithAFiveMinuteGoTestTimeout() error {
	expected := []string{
		"go test -v ./test -timeout 5m",
		"timeout 5m go test -v ./test -run TestFeatures -count=1 -timeout 5m",
	}
	for _, path := range []string{"../README.md", "../docs/TESTING.md"} {
		body, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		text := string(body)
		for _, command := range expected {
			if !strings.Contains(text, command) {
				return fmt.Errorf("%s should document %q", path, command)
			}
		}
	}
	return nil
}

func (tc *testContext) theTestingDocsShouldShowBoundedRepoWidePackageTestCommands() error {
	expected := []string{
		"INFOMUNGE_SKIP_GODOG=1 go test ./... -timeout 5m",
		"INTENSIVE_TEST_SOAK=1 go test -v ./internal/testing/mutation -run TestMutatedCorpusExpressions_NoPanics_AndDeterministic -timeout 30m",
	}
	for _, path := range []string{"../README.md", "../docs/TESTING.md"} {
		body, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		text := string(body)
		for _, command := range expected {
			if !strings.Contains(text, command) {
				return fmt.Errorf("%s should document %q", path, command)
			}
		}
	}
	return nil
}

func (tc *testContext) theDefaultMutationCorpusTestShouldBeSkippedOutsideSoakMode() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "test", "-v", "./internal/testing/mutation", "-run", "TestMutatedCorpusExpressions_NoPanics_AndDeterministic", "-count=1", "-timeout", "5m")
	cmd.Dir = ".."
	cmd.Env = append(withoutEnv(os.Environ(), "INTENSIVE_TEST_SOAK"), "INTENSIVE_TEST_SOAK=0")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String() + stderr.String()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("default mutation corpus test timed out, output: %s", output)
		}
		return fmt.Errorf("default mutation corpus test failed: %v, output: %s", err, output)
	}
	if !strings.Contains(output, "skipping mutation corpus soak; set INTENSIVE_TEST_SOAK=1 to run") {
		return fmt.Errorf("default mutation corpus test should skip outside soak mode, output: %s", output)
	}
	return nil
}

func readMakefileTargets(path string) (map[string][]string, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	targets := make(map[string][]string)
	currentTarget := ""
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, "\t") {
			if currentTarget != "" {
				targets[currentTarget] = append(targets[currentTarget], strings.TrimSpace(line))
			}
			continue
		}
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ".") {
			currentTarget = ""
			continue
		}
		if before, _, ok := strings.Cut(line, ":"); ok {
			currentTarget = strings.TrimSpace(before)
		}
	}
	return targets, nil
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func hasTimeoutArg(args []string, expected string) bool {
	for i := 0; i+1 < len(args); i++ {
		arg := args[i]
		if arg == "-timeout" && args[i+1] == expected {
			return true
		}
	}
	return false
}

func withoutEnv(env []string, name string) []string {
	prefix := name + "="
	out := make([]string, 0, len(env))
	for _, value := range env {
		if !strings.HasPrefix(value, prefix) {
			out = append(out, value)
		}
	}
	return out
}
