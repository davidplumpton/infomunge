package formats

import (
	unifiederrors "infomunge/internal/errors"
)

// Read parses the content based on the provided mimeType using the registered readers.
func Read(content, mimeType string) (interface{}, error) {
	return ReadWithOptions(content, mimeType, nil)
}

// ReadWithOptions parses content with optional format-specific options.
func ReadWithOptions(content, mimeType string, options Object) (interface{}, error) {
	if mimeType == "" {
		return nil, unifiederrors.ValidationError("mimeType cannot be empty")
	}
	if content == "" {
		return nil, nil
	}

	if options != nil {
		switch mimeType {
		case "application/flatfile":
			return readFlatfileWithOptions(content, options)
		default:
			if len(options) > 0 {
				return nil, unifiederrors.ValidationErrorf("read options are not supported for mimeType: %s", mimeType)
			}
		}
	}

	reader, err := GetReader(mimeType)
	if err != nil {
		return nil, err
	}

	return reader(content)
}

// ReadAsObject parses content and returns a Result[Object].
// Returns an error result if the content cannot be parsed or is not an Object.
func ReadAsObject(content, mimeType string) Result[Object] {
	result, err := Read(content, mimeType)
	if err != nil {
		return Err[Object](err)
	}
	if result == nil {
		return Ok[Object](nil)
	}
	obj, ok := result.(Object)
	if !ok {
		return Err[Object](unifiederrors.ValidationErrorf("expected Object, got %T", result))
	}
	return Ok(obj)
}

// ReadAsArray parses content and returns a Result[Array].
// Returns an error result if the content cannot be parsed or is not an Array.
func ReadAsArray(content, mimeType string) Result[Array] {
	result, err := Read(content, mimeType)
	if err != nil {
		return Err[Array](err)
	}
	if result == nil {
		return Ok[Array](nil)
	}
	arr, ok := result.(Array)
	if !ok {
		return Err[Array](unifiederrors.ValidationErrorf("expected Array, got %T", result))
	}
	return Ok(arr)
}
