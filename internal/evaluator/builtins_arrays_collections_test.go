package evaluator

import "testing"

func TestInclusiveRangeBounds(t *testing.T) {
	tests := []struct {
		name          string
		start         int
		end           int
		length        int
		expectedStart int
		expectedEnd   int
		expectedOK    bool
	}{
		{
			name:          "ascending",
			start:         0,
			end:           1,
			length:        3,
			expectedStart: 0,
			expectedEnd:   1,
			expectedOK:    true,
		},
		{
			name:          "negative bounds",
			start:         -2,
			end:           -1,
			length:        3,
			expectedStart: 1,
			expectedEnd:   2,
			expectedOK:    true,
		},
		{
			name:          "descending",
			start:         -1,
			end:           0,
			length:        3,
			expectedStart: 2,
			expectedEnd:   0,
			expectedOK:    true,
		},
		{
			name:       "empty collection",
			start:      0,
			end:        0,
			length:     0,
			expectedOK: false,
		},
		{
			name:       "start beyond collection",
			start:      3,
			end:        0,
			length:     3,
			expectedOK: false,
		},
		{
			name:          "end beyond collection is clamped",
			start:         1,
			end:           4,
			length:        3,
			expectedStart: 1,
			expectedEnd:   2,
			expectedOK:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, ok := inclusiveRangeBounds(tt.start, tt.end, tt.length)
			if start != tt.expectedStart || end != tt.expectedEnd || ok != tt.expectedOK {
				t.Fatalf(
					"inclusiveRangeBounds(%d, %d, %d) = (%d, %d, %t), want (%d, %d, %t)",
					tt.start,
					tt.end,
					tt.length,
					start,
					end,
					ok,
					tt.expectedStart,
					tt.expectedEnd,
					tt.expectedOK,
				)
			}
		})
	}
}
