package test

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestMakefileTestUnitCoversRepositoryUnitPackages(t *testing.T) {
	if err := validateMakefileTestUnitTarget(); err != nil {
		t.Fatal(err)
	}
}

func (tc *testContext) theMakefileUnitTargetShouldCoverRepresentativeRepositoryUnitPackages() error {
	return validateMakefileTestUnitTarget()
}

func validateMakefileTestUnitTarget() error {
	targets, err := readMakefileTargets("../Makefile")
	if err != nil {
		return err
	}

	const expectedCommand = "INFOMUNGE_SKIP_GODOG=1 go test ./... -timeout 5m"
	commands := targets["test-unit"]
	if !containsString(commands, expectedCommand) {
		return fmt.Errorf("Makefile target %q should include %q, got %v", "test-unit", expectedCommand, commands)
	}
	for _, command := range commands {
		if strings.Contains(command, " -run ") {
			return fmt.Errorf("Makefile target %q should not filter tests by name, got %q", "test-unit", command)
		}
	}

	cmd := exec.Command("go", "list", "./...")
	cmd.Dir = ".."
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("list packages matched by test-unit: %w", err)
	}

	discovered := strings.Fields(string(output))
	for _, representative := range []string{
		"infomunge/internal/evaluator",
		"infomunge/pkg/formats",
		"infomunge/test",
	} {
		if !containsString(discovered, representative) {
			return fmt.Errorf("test-unit package pattern should discover representative package %q; got %v", representative, discovered)
		}
	}
	return nil
}
