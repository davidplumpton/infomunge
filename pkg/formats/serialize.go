package formats

import (
	"encoding/json"
	"fmt"
)

// marshalToJSON attempts to marshal a value to JSON format string,
// falling back to fmt.Sprintf if marshaling fails.
func marshalToJSON(v interface{}) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(bytes)
}

// valueToCSVString converts a value to a string for CSV serialization.
// For strings, returns the value as-is. For other types, uses JSON marshaling
// and falls back to fmt.Sprintf if marshaling fails.
func valueToCSVString(v interface{}) string {
	// For strings, return as-is without JSON marshaling (which would add quotes)
	if s, ok := v.(string); ok {
		return s
	}
	// For other types, try JSON marshaling
	bytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(bytes)
}
