package stringutils

import "testing"

func TestIsEscapedAt(t *testing.T) {
	tests := []struct {
		name string
		s    string
		pos  int
		want bool
	}{
		{"no backslash", `abc"`, 3, false},
		{"single backslash", `ab\"`, 3, true},
		{"double backslash", `ab\\"`, 4, false},
		{"triple backslash", `ab\\\"`, 5, true},
		{"at start", `"abc`, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEscapedAt(tt.s, tt.pos); got != tt.want {
				t.Fatalf("IsEscapedAt(%q, %d) = %v, want %v", tt.s, tt.pos, got, tt.want)
			}
		})
	}
}

func TestIsEscapedRuneAt(t *testing.T) {
	tests := []struct {
		name  string
		runes []rune
		pos   int
		want  bool
	}{
		{"no backslash", []rune(`abc"`), 3, false},
		{"single backslash", []rune(`ab\"`), 3, true},
		{"double backslash", []rune(`ab\\"`), 4, false},
		{"triple backslash", []rune(`ab\\\"`), 5, true},
		{"at start", []rune(`"abc`), 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEscapedRuneAt(tt.runes, tt.pos); got != tt.want {
				t.Fatalf("IsEscapedRuneAt(%q, %d) = %v, want %v", string(tt.runes), tt.pos, got, tt.want)
			}
		})
	}
}
