package test

import (
	"testing"

	"pgregory.net/rapid"
)

func TestRapidDependencySmoke(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(0, 100).Draw(t, "n")
		if n < 0 || n > 100 {
			t.Fatalf("generated value out of expected range: %d", n)
		}
	})
}
