package evaluator

import (
	"strings"
	"testing"
)

func TestSetValueAtPath_OutOfBoundsNonTerminalIndexReturnsError(t *testing.T) {
	value := Array{
		Array{1.0},
	}
	path := []selectorSegment{
		{index: 5, isIndex: true},
		{index: 0, isIndex: true},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("setValueAtPath should not panic, got %v", r)
		}
	}()

	_, err := setValueAtPath(value, path, 99.0, 0)
	if err == nil {
		t.Fatalf("expected out-of-bounds error, got nil")
	}
	if !strings.Contains(err.Error(), "index out of bounds: 5") {
		t.Fatalf("expected index out-of-bounds error, got %v", err)
	}
}
