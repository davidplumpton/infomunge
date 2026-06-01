package test

import "testing"

func TestOutputShouldBeValidJSONWithNumberRequiresExactInteger(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
		wantErr  bool
	}{
		{
			name:     "integer matches",
			output:   "42",
			expected: "42",
		},
		{
			name:     "decimal integer matches",
			output:   "42.0",
			expected: "42",
		},
		{
			name:     "fractional number fails",
			output:   "42.9",
			expected: "42",
			wantErr:  true,
		},
		{
			name:     "negative decimal integer matches",
			output:   "-3.0",
			expected: "-3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &testContext{lastOutput: tt.output}

			err := tc.theOutputShouldBeValidJSONWithNumber(tt.expected)
			if (err != nil) != tt.wantErr {
				t.Fatalf("theOutputShouldBeValidJSONWithNumber() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
