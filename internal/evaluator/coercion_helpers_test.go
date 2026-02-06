package evaluator

import "testing"

func TestCoerceEquals(t *testing.T) {
	tests := []struct {
		name  string
		left  Value
		right Value
		want  bool
	}{
		{name: "numeric int and float", left: 2, right: 2.0, want: true},
		{name: "numeric string and number", left: "2.5", right: 2.5, want: true},
		{name: "string fallback match", left: "abc", right: "abc", want: true},
		{name: "non-equal values", left: 1, right: "2", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CoerceEquals(tt.left, tt.right); got != tt.want {
				t.Fatalf("CoerceEquals(%v, %v) = %v, want %v", tt.left, tt.right, got, tt.want)
			}
		})
	}
}

func TestNumericEquals(t *testing.T) {
	tests := []struct {
		name  string
		left  Value
		right Value
		want  bool
	}{
		{name: "same int", left: 3, right: 3, want: true},
		{name: "int and float", left: 3, right: 3.0, want: true},
		{name: "different numbers", left: 3, right: 4.0, want: false},
		{name: "non-numeric", left: "3", right: 3.0, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := numericEquals(tt.left, tt.right); got != tt.want {
				t.Fatalf("numericEquals(%v, %v) = %v, want %v", tt.left, tt.right, got, tt.want)
			}
		})
	}
}

