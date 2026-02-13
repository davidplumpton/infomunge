// Package testbudget provides CI vs soak test budget configuration controlled
// by the INTENSIVE_TEST_SOAK environment variable. When the variable is unset
// (or "0"), budgets target a short CI run (< 60 seconds total). When set to "1"
// (or "true"/"yes"), budgets expand for thorough local/nightly soak runs.
package testbudget

import (
	"os"
	"strings"
)

// SoakEnabled returns true when INTENSIVE_TEST_SOAK is set.
func SoakEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("INTENSIVE_TEST_SOAK")))
	return v == "1" || v == "true" || v == "yes"
}

// PropertyChecks returns the rapid iteration count for algebraic property tests.
func PropertyChecks() int {
	if SoakEnabled() {
		return 50000
	}
	return 500
}

// NoPanicChecks returns the rapid iteration count for no-panic property tests.
func NoPanicChecks() int {
	if SoakEnabled() {
		return 50000
	}
	return 500
}

// StructuralChecks returns the rapid iteration count for structural/round-trip tests.
func StructuralChecks() int {
	if SoakEnabled() {
		return 50000
	}
	return 500
}

// DiffChecks returns the rapid iteration count for differential comparison tests.
func DiffChecks() int {
	if SoakEnabled() {
		return 50000
	}
	return 500
}

// MutationSampleSize returns the number of corpus entries to sample.
// Returns 0 when in soak mode, meaning use the full corpus.
func MutationSampleSize() int {
	if SoakEnabled() {
		return 0
	}
	return 50
}

// MutationMaxMutations returns the maximum mutations to apply per expression.
func MutationMaxMutations() int {
	if SoakEnabled() {
		return 10
	}
	return 3
}

// MutationChecks returns the rapid iteration count for mutation tests.
func MutationChecks() int {
	if SoakEnabled() {
		return 1000
	}
	return 300
}
