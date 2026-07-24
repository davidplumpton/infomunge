package values

import (
	"strconv"
	"strings"
)

// ParseNumericLiteral parses a string as int or float64 with strict semantics.
func ParseNumericLiteral(s string) (Value, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, false
	}

	if !strings.ContainsAny(s, ".eE") {
		if iv, err := strconv.Atoi(s); err == nil {
			return iv, true
		}
	}

	if fv, err := strconv.ParseFloat(s, 64); err == nil {
		return fv, true
	}

	return nil, false
}

// ParseBoolLiteral parses "true" or "false", case-insensitively.
func ParseBoolLiteral(s string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true":
		return true, true
	case "false":
		return false, true
	default:
		return false, false
	}
}
