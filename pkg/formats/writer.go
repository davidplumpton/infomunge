package formats

import (
	unifiederrors "infomunge/internal/errors"
)

// Format formats the result based on the provided mimeType using the registered writers.
// Returns an error if mimeType is empty or result cannot be formatted.
func Format(result interface{}, mimeType string) (string, error) {
	if mimeType == "" {
		return "", unifiederrors.ValidationError("mimeType cannot be empty")
	}
	if result == nil {
		return "null", nil
	}

	if err := validateFormatInput(result); err != nil {
		return "", unifiederrors.ValidationErrorf("invalid input for formatting: %v", err)
	}

	writer, err := GetWriter(mimeType)
	if err != nil {
		// Default to JSON if no specific writer is registered
		jsonWriter, jsonErr := GetWriter("application/json")
		if jsonErr == nil {
			return jsonWriter(result)
		}
		return "", err
	}

	return writer(result)
}

// validateFormatInput performs basic validation on input before formatting.
// Checks for null pointers and basic type compatibility.
func validateFormatInput(result interface{}) error {
	if result == nil {
		return unifiederrors.ValidationError("input cannot be nil")
	}
	return nil
}

// FormatXMLWithNamespaces formats result as XML with declared namespaces.
// declaredNs maps namespace prefixes to URIs (e.g., {"ns0": "http://www.abc.com"}).
func FormatXMLWithNamespaces(result interface{}, declaredNs map[string]string) (string, error) {
	if result == nil {
		return "", unifiederrors.ValidationError("result cannot be nil")
	}

	if err := validateFormatInput(result); err != nil {
		return "", unifiederrors.ValidationErrorf("invalid input for formatting: %v", err)
	}

	return formatXMLWithNamespaces(result, declaredNs)
}

// FormatResult formats the result and returns a Result[string].
// Provides a Result-based API for formatting operations.
func FormatResult(result interface{}, mimeType string) Result[string] {
	str, err := Format(result, mimeType)
	if err != nil {
		return Err[string](err)
	}
	return Ok(str)
}
