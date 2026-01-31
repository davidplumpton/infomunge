package formats

import (
	unifiederrors "infomunge/internal/errors"
	"strings"
	"sync"
)

// Reader is a function that parses content into an interface{} representation.
// Readers should return one of: Object, Array, string, float64, bool, or nil.
// Use ReadAsObject or ReadAsArray for type-safe access with Result types.
type Reader func(content string) (interface{}, error)

// Writer is a function that serializes an interface{} into a string.
// Writers accept any interface{} value and produce a string representation.
// Use FormatResult for type-safe access with Result types.
type Writer func(result interface{}) (string, error)

// ObjectReader is a typed reader that returns an Object directly.
// Use this for formats that always produce objects (e.g., XML, YAML with single document).
type ObjectReader func(content string) (Object, error)

// ArrayReader is a typed reader that returns an Array directly.
// Use this for formats that always produce arrays (e.g., CSV).
type ArrayReader func(content string) (Array, error)

var (
	readers       = make(map[string]Reader)
	writers       = make(map[string]Writer)
	objectReaders = make(map[string]ObjectReader)
	arrayReaders  = make(map[string]ArrayReader)
	extensions    = make(map[string]string)
	registryMu    sync.RWMutex
)

// RegisterReader registers a reader for a specific MIME type.
func RegisterReader(mimeType string, r Reader) {
	registryMu.Lock()
	defer registryMu.Unlock()
	readers[mimeType] = r
}

// RegisterWriter registers a writer for a specific MIME type.
func RegisterWriter(mimeType string, w Writer) {
	registryMu.Lock()
	defer registryMu.Unlock()
	writers[mimeType] = w
}

// RegisterObjectReader registers a typed ObjectReader for a MIME type.
// ObjectReaders guarantee the output is an Object, enabling type-safe access.
func RegisterObjectReader(mimeType string, r ObjectReader) {
	registryMu.Lock()
	defer registryMu.Unlock()
	objectReaders[mimeType] = r
	// Also register as a generic Reader for compatibility
	readers[mimeType] = func(content string) (interface{}, error) {
		return r(content)
	}
}

// RegisterArrayReader registers a typed ArrayReader for a MIME type.
// ArrayReaders guarantee the output is an Array, enabling type-safe access.
func RegisterArrayReader(mimeType string, r ArrayReader) {
	registryMu.Lock()
	defer registryMu.Unlock()
	arrayReaders[mimeType] = r
	// Also register as a generic Reader for compatibility
	readers[mimeType] = func(content string) (interface{}, error) {
		return r(content)
	}
}

// RegisterExtension registers a file extension (including dot, e.g., ".json") to a MIME type.
func RegisterExtension(ext, mimeType string) {
	registryMu.Lock()
	defer registryMu.Unlock()
	extensions[strings.ToLower(ext)] = mimeType
}

// DetectMimeType attempts to determine the MIME type from a filename's extension.
// Returns empty string if unknown.
func DetectMimeType(filename string) string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	// Find the last dot
	lastDot := strings.LastIndex(filename, ".")
	if lastDot == -1 {
		return ""
	}

	// Check if the dot is part of the path (after the last separator)
	// We do a simple check for forward slash and backslash
	lastSep := strings.LastIndexAny(filename, "/\\")
	if lastDot < lastSep {
		return ""
	}

	ext := strings.ToLower(filename[lastDot:])
	if mime, ok := extensions[ext]; ok {
		return mime
	}
	return ""
}

// MimeTypeForFormat returns the MIME type for a format name (e.g., "json" -> "application/json").
// Returns empty string if unknown.
func MimeTypeForFormat(format string) string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	ext := "." + strings.ToLower(format)
	if mime, ok := extensions[ext]; ok {
		return mime
	}
	return ""
}

// GetReader returns the reader registered for the given MIME type.
func GetReader(mimeType string) (Reader, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	r, ok := readers[mimeType]
	if !ok {
		return nil, unifiederrors.ValidationErrorf("unsupported input mimeType: %s", mimeType)
	}
	return r, nil
}

// GetWriter returns the writer registered for the given MIME type.
func GetWriter(mimeType string) (Writer, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	w, ok := writers[mimeType]
	if !ok {
		// Default to JSON if not found? No, better be explicit or handle in Format.
		return nil, unifiederrors.ValidationErrorf("unsupported output mimeType: %s", mimeType)
	}
	return w, nil
}

// GetObjectReader returns the typed ObjectReader for a MIME type, if registered.
// Returns nil if no ObjectReader is registered (even if a generic Reader exists).
func GetObjectReader(mimeType string) (ObjectReader, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	r, ok := objectReaders[mimeType]
	return r, ok
}

// GetArrayReader returns the typed ArrayReader for a MIME type, if registered.
// Returns nil if no ArrayReader is registered (even if a generic Reader exists).
func GetArrayReader(mimeType string) (ArrayReader, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	r, ok := arrayReaders[mimeType]
	return r, ok
}

// HasTypedReader returns true if a typed reader (ObjectReader or ArrayReader) is registered.
func HasTypedReader(mimeType string) bool {
	registryMu.RLock()
	defer registryMu.RUnlock()
	_, hasObj := objectReaders[mimeType]
	_, hasArr := arrayReaders[mimeType]
	return hasObj || hasArr
}
