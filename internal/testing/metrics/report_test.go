package metrics

import (
	"testing"
	"time"

	"infomunge/internal/testing/failures"
)

func TestCountUniqueFailures(t *testing.T) {
	t.Parallel()

	artifacts := []failures.Artifact{
		{Property: "p1", MinimizedExpression: "a", Fingerprint: failures.Fingerprint("a", "p1")},
		{Property: "p1", MinimizedExpression: "a", Fingerprint: failures.Fingerprint("a", "p1")},
		{Property: "p2", MinimizedExpression: "b", Fingerprint: failures.Fingerprint("b", "p2")},
	}

	got := countUniqueFailures(artifacts)
	if got != 2 {
		t.Fatalf("countUniqueFailures() = %d, want 2", got)
	}
}

func TestAverageShrinkRatio(t *testing.T) {
	t.Parallel()

	artifacts := []failures.Artifact{
		{OriginalExpression: "abcdefghij", MinimizedExpression: "abcde"},
		{OriginalExpression: "12345678", MinimizedExpression: "12"},
	}

	got := averageShrinkRatio(artifacts)
	if got == nil {
		t.Fatal("averageShrinkRatio() = nil, want value")
	}
	// (5/10 + 2/8) / 2 = 0.375
	if diff := *got - 0.375; diff < -1e-9 || diff > 1e-9 {
		t.Fatalf("averageShrinkRatio() = %.6f, want 0.375", *got)
	}
}

func TestAverageTimeToMinimalRepro(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 2, 13, 6, 0, 0, 0, time.UTC)
	artifacts := []failures.Artifact{
		{
			DetectedAt:  start.Format(time.RFC3339),
			MinimizedAt: start.Add(10 * time.Second).Format(time.RFC3339),
		},
		{
			DetectedAt:  start.Format(time.RFC3339),
			MinimizedAt: start.Add(20 * time.Second).Format(time.RFC3339),
		},
	}

	got := averageTimeToMinimalRepro(artifacts)
	if got == nil {
		t.Fatal("averageTimeToMinimalRepro() = nil, want value")
	}
	if *got != 15 {
		t.Fatalf("averageTimeToMinimalRepro() = %.2f, want 15", *got)
	}
}

func TestCollectSnapshotIncludesMutationCounters(t *testing.T) {
	panicCount.Store(0)
	mutationKilled.Store(0)
	mutationSurvived.Store(0)

	RecordPanic()
	RecordMutationOutcome(true)
	RecordMutationOutcome(false)

	got := collectSnapshot(Options{EnableCoverage: false})
	if got.PanicCount != 1 {
		t.Fatalf("PanicCount = %d, want 1", got.PanicCount)
	}
	if got.MutationKilled != 1 {
		t.Fatalf("MutationKilled = %d, want 1", got.MutationKilled)
	}
	if got.MutationSurvived != 1 {
		t.Fatalf("MutationSurvived = %d, want 1", got.MutationSurvived)
	}
	if got.MutationKillRate == nil || *got.MutationKillRate != 0.5 {
		t.Fatalf("MutationKillRate = %#v, want 0.5", got.MutationKillRate)
	}
}
