package formats

import (
	unifiederrors "infomunge/internal/errors"
)

func init() {
	RegisterReader("application/octet-stream", readBinary)
	RegisterWriter("application/octet-stream", formatBinary)
	RegisterExtension(".bin", "application/octet-stream")
	RegisterExtension(".binary", "application/octet-stream")
}

func readBinary(content string) (interface{}, error) {
	return content, nil
}

func formatBinary(result interface{}) (string, error) {
	switch v := result.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return "", unifiederrors.ValidationErrorf("binary output expects string or []byte, got %T", result)
	}
}
