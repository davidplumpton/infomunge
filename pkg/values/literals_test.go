package values

import (
	"reflect"
	"testing"
)

func TestParseNumericLiteral(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Value
		ok    bool
	}{
		{name: "integer", input: "42", want: 42, ok: true},
		{name: "negative integer with whitespace", input: " -7 ", want: -7, ok: true},
		{name: "decimal", input: "3.5", want: 3.5, ok: true},
		{name: "exponent", input: "1e3", want: 1000.0, ok: true},
		{name: "empty", input: " ", ok: false},
		{name: "trailing text", input: "12px", ok: false},
		{name: "multiple signs", input: "--1", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseNumericLiteral(tt.input)
			if ok != tt.ok {
				t.Fatalf("ParseNumericLiteral(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ParseNumericLiteral(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseBoolLiteral(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
		ok    bool
	}{
		{name: "true", input: "true", want: true, ok: true},
		{name: "false", input: "false", want: false, ok: true},
		{name: "mixed case with whitespace", input: " TrUe ", want: true, ok: true},
		{name: "numeric truthy is rejected", input: "1", ok: false},
		{name: "prefix is rejected", input: "trueish", ok: false},
		{name: "empty", input: "", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseBoolLiteral(tt.input)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("ParseBoolLiteral(%q) = (%v, %v), want (%v, %v)", tt.input, got, ok, tt.want, tt.ok)
			}
		})
	}
}
