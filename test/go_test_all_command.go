package test

import (
	"context"
	"os"
	"os/exec"
	"time"
)

const (
	repoWideGoTestTimeout    = 5 * time.Minute
	repoWideGoTestTimeoutArg = "5m"
	repoWideGoTestRunPattern = "^$"
)

func repoWideGoTestArgs() []string {
	// This regression verifies repo-wide package discovery/compilation without
	// running intensive package tests repeatedly inside a cucumber scenario.
	return []string{"test", "./...", "-run", repoWideGoTestRunPattern, "-timeout", repoWideGoTestTimeoutArg}
}

func repoWideGoTestCommand(ctx context.Context) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "go", repoWideGoTestArgs()...)
	cmd.Dir = ".."
	cmd.Env = append(os.Environ(), "INFOMUNGE_SKIP_GODOG=1")
	return cmd
}
