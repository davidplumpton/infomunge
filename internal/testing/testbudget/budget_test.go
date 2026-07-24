package testbudget

import "testing"

func TestDiffChecks_DefaultIsBoundedForExternalProcesses(t *testing.T) {
	t.Setenv("INTENSIVE_TEST_SOAK", "0")

	if got := DiffChecks(); got != defaultDiffChecks {
		t.Fatalf("DiffChecks() = %d, want bounded default %d", got, defaultDiffChecks)
	}
}

func TestDiffChecks_SoakUsesLargerExplicitBudget(t *testing.T) {
	t.Setenv("INTENSIVE_TEST_SOAK", "1")

	got := DiffChecks()
	if got != soakDiffChecks {
		t.Fatalf("DiffChecks() = %d, want soak budget %d", got, soakDiffChecks)
	}
	if got <= defaultDiffChecks {
		t.Fatalf("soak differential budget %d must exceed default budget %d", got, defaultDiffChecks)
	}
}
