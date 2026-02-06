package test

import "testing"

func TestResolveGodogFormat_DefaultsToFailures(t *testing.T) {
	t.Setenv("GODOG_FORMAT", "")

	if got := resolveGodogFormat(); got != failuresFormatName {
		t.Fatalf("expected default format %q, got %q", failuresFormatName, got)
	}
}

func TestResolveGodogFormat_UsesEnvironmentOverride(t *testing.T) {
	t.Setenv("GODOG_FORMAT", "pretty")

	if got := resolveGodogFormat(); got != "pretty" {
		t.Fatalf("expected env override format %q, got %q", "pretty", got)
	}
}
