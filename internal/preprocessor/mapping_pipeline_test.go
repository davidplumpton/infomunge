package preprocessor

import "testing"

func TestInferStageMapping_Identity(t *testing.T) {
	before := "a && b"
	after := "a && b"
	mapping := inferStageMapping(before, after)
	if len(mapping) != len(after) {
		t.Fatalf("expected mapping length %d, got %d", len(after), len(mapping))
	}
	for i := range mapping {
		if mapping[i] != i {
			t.Fatalf("expected mapping[%d]=%d, got %d", i, i, mapping[i])
		}
	}
}

func TestInferStageMapping_ResizedMiddleSegment(t *testing.T) {
	before := "a and b"
	after := "a && b"
	mapping := inferStageMapping(before, after)
	if len(mapping) != len(after) {
		t.Fatalf("expected mapping length %d, got %d", len(after), len(mapping))
	}
	if mapping[0] != 0 {
		t.Fatalf("expected first char to map to 0, got %d", mapping[0])
	}
	if mapping[len(mapping)-1] != len(before)-1 {
		t.Fatalf("expected last char to map to %d, got %d", len(before)-1, mapping[len(mapping)-1])
	}
	for i := 1; i < len(mapping); i++ {
		if mapping[i] < mapping[i-1] {
			t.Fatalf("expected monotonic mapping, got %v", mapping)
		}
	}
}

func TestComposeMappings_ComposesIntoOriginalOffsets(t *testing.T) {
	prev := []int{10, 11, 12, 13, 14, 15}
	local := []int{0, 2, 5}
	got := composeMappings(prev, local)
	want := []int{10, 12, 15}
	if len(got) != len(want) {
		t.Fatalf("expected len %d, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected got[%d]=%d, got %d", i, want[i], got[i])
		}
	}
}
