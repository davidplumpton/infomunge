package formats

import (
	unifiederrors "infomunge/internal/errors"
)

// Format formats the result based on the provided mimeType using the registered writers.
// Returns an error if mimeType is empty or result cannot be formatted.
func Format(result interface{}, mimeType string) (string, error) {
	return FormatWithOptions(result, mimeType, nil)
}

// FormatWithOptions formats result with optional format-specific options.
func FormatWithOptions(result interface{}, mimeType string, options Object) (string, error) {
	if mimeType == "" {
		return "", unifiederrors.ValidationError("mimeType cannot be empty")
	}
	if result == nil {
		return "null", nil
	}

	if options != nil {
		switch mimeType {
		case "application/flatfile":
			return formatFlatfileWithOptions(result, options)
		case "application/java":
			return formatJavaWithOptions(result, options)
		default:
			if len(options) > 0 {
				return "", unifiederrors.ValidationErrorf("write options are not supported for mimeType: %s", mimeType)
			}
		}
	}

	writer, err := GetWriter(mimeType)
	if err != nil {
		return "", err
	}

	return writer(result)
}

// FormatXMLWithNamespaces formats result as XML with declared namespaces.
// declaredNs maps namespace prefixes to URIs (e.g., {"ns0": "http://www.abc.com"}).
func FormatXMLWithNamespaces(result interface{}, declaredNs map[string]string) (string, error) {
	opts := XMLOutputOptions{
		DeclaredNamespaces: declaredNs,
		WriteDeclaration:   true,
	}
	return FormatXMLWithOptions(result, opts)
}

// FormatXMLWithOptions formats result as XML with custom options.
func FormatXMLWithOptions(result interface{}, opts XMLOutputOptions) (string, error) {
	if result == nil {
		return "", unifiederrors.ValidationError("result cannot be nil")
	}

	return formatXMLWithOptions(result, opts)
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
