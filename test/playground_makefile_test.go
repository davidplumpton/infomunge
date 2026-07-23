package test

import "testing"

func TestMakefilePlaygroundTargetUsesExplicitLoopbackAddress(t *testing.T) {
	targets, err := readMakefileTargets("../Makefile")
	if err != nil {
		t.Fatal(err)
	}

	const expected = "go run ./cmd/infomunge --server --listen 127.0.0.1:8080"
	commands := targets["playground"]
	if len(commands) != 1 || commands[0] != expected {
		t.Fatalf(
			"Makefile playground target must use the explicit unauthenticated loopback command %q; got %v",
			expected,
			commands,
		)
	}
}
