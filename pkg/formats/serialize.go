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

// valueToPlainString converts a value to text/plain output.
// Strings are returned as-is; other values use compact JSON when possible.
func valueToPlainString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	bytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(bytes)
}

// valueToCSVString converts a value to a string for CSV serialization.
func valueToCSVString(v interface{}) string {
	return valueToPlainString(v)
}
