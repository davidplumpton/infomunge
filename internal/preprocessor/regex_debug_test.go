package preprocessor

import (
	"testing"
)

func TestRegexPreprocessing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`/test/`, `regex("test")`},
		{`/[a-z]+/`, `regex("[a-z]+")`},
		// match operator converts to match function call
		{`"test" match(/[a-z]+/)`, `match("test", regex("[a-z]+"))`},
		// matches operator converts to matches function call
		{`"test" matches(/[a-z]+/)`, `matches("test", regex("[a-z]+"))`},
		// scan supports both infix and method-call syntax
		{`"test1" scan /[0-9]+/`, `scan("test1", regex("[0-9]+"))`},
		{`"test1" scan(/[0-9]+/)`, `scan("test1", regex("[0-9]+"))`},
		// contains with method call syntax
		{`"test" contains(/e/)`, `contains("test", regex("e"))`},
		// Emoji handling
		{`"Hello 👋 World 🌍"`, `"Hello 👋 World 🌍"`},
	}

	for _, tt := range tests {
		result, _, err := PrepareForParsing(tt.input, Options{})
		if err != nil {
			t.Errorf("Error for %q: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("Input: %q\nExpected: %q\nGot: %q", tt.input, tt.expected, result)
		}
	}
}
