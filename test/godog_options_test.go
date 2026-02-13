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

func TestResolveGodogConcurrency_DefaultsToFour(t *testing.T) {
	t.Setenv("GODOG_CONCURRENCY", "")

	if got := resolveGodogConcurrency(); got != 4 {
		t.Fatalf("expected default concurrency %d, got %d", 4, got)
	}
}

func TestResolveGodogConcurrency_UsesEnvironmentOverride(t *testing.T) {
	t.Setenv("GODOG_CONCURRENCY", "2")

	if got := resolveGodogConcurrency(); got != 2 {
		t.Fatalf("expected env override concurrency %d, got %d", 2, got)
	}
}

func TestResolveGodogConcurrency_InvalidValuesFallBackToDefault(t *testing.T) {
	t.Setenv("GODOG_CONCURRENCY", "0")
	if got := resolveGodogConcurrency(); got != 4 {
		t.Fatalf("expected fallback concurrency %d, got %d", 4, got)
	}

	t.Setenv("GODOG_CONCURRENCY", "invalid")
	if got := resolveGodogConcurrency(); got != 4 {
		t.Fatalf("expected fallback concurrency %d, got %d", 4, got)
	}
}
