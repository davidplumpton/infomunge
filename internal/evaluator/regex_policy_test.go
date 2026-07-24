package evaluator

import "testing"

func TestLooksLikeRegexPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    bool
	}{
		{name: "plain text", pattern: "literal", want: false},
		{name: "single dot remains literal", pattern: ".", want: false},
		{name: "dot in pattern", pattern: "a.b", want: true},
		{name: "quantifier", pattern: "a+", want: true},
		{name: "character class", pattern: "[a-z]", want: true},
		{name: "group", pattern: "(cat|dog)", want: true},
		{name: "anchor", pattern: "^start", want: true},
		{name: "escaped class", pattern: `\d`, want: true},
		{name: "empty", pattern: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := looksLikeRegexPattern(tt.pattern); got != tt.want {
				t.Fatalf("looksLikeRegexPattern(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		})
	}
}
