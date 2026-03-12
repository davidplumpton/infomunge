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

func TestReplaceRegexLiteralsWithMapping_NoChange(t *testing.T) {
	input := "a + b"
	result, mapping := replaceRegexLiteralsWithMapping(input)
	if result != input {
		t.Fatalf("expected %q, got %q", input, result)
	}
	if len(mapping) != len(result) {
		t.Fatalf("mapping length %d != result length %d", len(mapping), len(result))
	}
	for i, pos := range mapping {
		if pos != i {
			t.Fatalf("expected identity mapping[%d]=%d, got %d", i, i, pos)
		}
	}
}

func TestReplaceRegexLiteralsWithMapping_SimpleRegex(t *testing.T) {
	input := `/abc/`
	result, mapping := replaceRegexLiteralsWithMapping(input)
	expected := `regex("abc")`
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
	if len(mapping) != len(result) {
		t.Fatalf("mapping length %d != result length %d", len(mapping), len(result))
	}
	// All generated characters should map back to position 0 (the opening /)
	for i, pos := range mapping {
		if pos != 0 {
			t.Fatalf("expected mapping[%d]=0, got %d", i, pos)
		}
	}
}

func TestReplaceRegexLiteralsWithMapping_RegexWithFlags(t *testing.T) {
	input := `/abc/gi`
	result, mapping := replaceRegexLiteralsWithMapping(input)
	expected := `regex("abc", "gi")`
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
	if len(mapping) != len(result) {
		t.Fatalf("mapping length %d != result length %d", len(mapping), len(result))
	}
}

func TestReplaceRegexLiteralsWithMapping_PreservesContext(t *testing.T) {
	// "x" followed by regex: the prefix maps to original positions, regex maps to regex start
	input := `x + /abc/`
	result, mapping := replaceRegexLiteralsWithMapping(input)
	expected := `x + regex("abc")`
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
	if len(mapping) != len(result) {
		t.Fatalf("mapping length %d != result length %d", len(mapping), len(result))
	}
	// "x + " (4 chars) should map to original positions 0-3
	for i := 0; i < 4; i++ {
		if mapping[i] != i {
			t.Fatalf("expected mapping[%d]=%d, got %d", i, i, mapping[i])
		}
	}
	// regex replacement starts at position 4 in original (the /)
	for i := 4; i < len(mapping); i++ {
		if mapping[i] != 4 {
			t.Fatalf("expected mapping[%d]=4, got %d", i, mapping[i])
		}
	}
}

func TestReplaceRegexLiteralsWithMapping_MappingLengthMatchesResult(t *testing.T) {
	inputs := []string{
		`/[a-z]+/i`,
		`x matches /\d+/`,
		`if (/test/) true else false`,
		`"no regex here /abc/"`,
		`a / b`,
	}
	for _, input := range inputs {
		result, mapping := replaceRegexLiteralsWithMapping(input)
		if len(mapping) != len(result) {
			t.Fatalf("input %q: mapping length %d != result length %d", input, len(mapping), len(result))
		}
	}
}
