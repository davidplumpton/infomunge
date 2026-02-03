package evaluator

import (
	"testing"
)

func TestConvertJavaDatePatternToGo(t *testing.T) {
	tests := []struct {
		java string
		want string
	}{
		{"yyyy-MM-dd", "2006-01-02"},
		{"yyyy/MM/dd HH:mm:ss", "2006/01/02 15:04:05"},
		{"dd-MM-yyyy", "02-01-2006"},
		{"hh:mm a", "03:04 PM"},
		{"yyyy-MM-dd HH:mm:ss.SSSXXX", "2006-01-02 15:04:05.000Z07:00"},
		{"yyyy-MM-dd HH:mm:ss.SSXX", "2006-01-02 15:04:05.00Z0700"},
		{"yyyy-MM-dd HH:mm:ss.SX", "2006-01-02 15:04:05.0Z07"},
		{"EEE, MMM d, ''yy", "Mon, Jan 2, '06"},
		{"'Date:' yyyy-MM-dd", "Date: 2006-01-02"},
		{"yyyy.MM.dd G 'at' HH:mm:ss z", "2006.01.02 G at 15:04:05 MST"},
	}

	for _, tt := range tests {
		t.Run(tt.java, func(t *testing.T) {
			got := convertJavaDatePatternToGo(tt.java)
			if got != tt.want {
				t.Errorf("convertJavaDatePatternToGo(%q) = %q, want %q", tt.java, got, tt.want)
			}
		})
	}
}
