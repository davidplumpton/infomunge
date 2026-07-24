package sourcemap

import "testing"

func TestClampIndex(t *testing.T) {
	tests := []struct {
		name   string
		pos    int
		length int
		want   int
	}{
		{name: "empty length", pos: 4, length: 0, want: 0},
		{name: "negative length", pos: 4, length: -1, want: 0},
		{name: "below range", pos: -1, length: 3, want: 0},
		{name: "inside range", pos: 1, length: 3, want: 1},
		{name: "at upper bound", pos: 3, length: 3, want: 2},
		{name: "above upper bound", pos: 9, length: 3, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClampIndex(tt.pos, tt.length); got != tt.want {
				t.Fatalf("ClampIndex(%d, %d) = %d, want %d", tt.pos, tt.length, got, tt.want)
			}
		})
	}
}
