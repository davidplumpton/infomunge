package formats

import formatcore "infomunge/pkg/formats/core"

// Reader is a function that parses content into an interface{} representation.
// Readers should return one of: Object, Array, string, float64, bool, or nil.
// Use ReadAsObject or ReadAsArray for type-safe access with Result types.
type Reader = formatcore.Reader

// Writer is a function that serializes an interface{} into a string.
// Writers accept any interface{} value and produce a string representation.
// Use FormatResult for type-safe access with Result types.
type Writer = formatcore.Writer

// ObjectReader is a typed reader that returns an Object directly.
// Use this for formats that always produce objects (e.g., XML, YAML with single document).
type ObjectReader = formatcore.ObjectReader

// ArrayReader is a typed reader that returns an Array directly.
// Use this for formats that always produce arrays (e.g., CSV).
type ArrayReader = formatcore.ArrayReader

var registry = formatcore.NewRegistry()

// Registration normally happens during init() in format-specific files. The
// underlying registry is synchronized to support tests and optional runtime
// extension safely; callers that need deterministic behavior should finish
// registration before serving concurrent read traffic.

// RegisterReader registers a reader for a specific MIME type.
func RegisterReader(mimeType string, r Reader) {
	registry.RegisterReader(mimeType, r)
}

// RegisterWriter registers a writer for a specific MIME type.
func RegisterWriter(mimeType string, w Writer) {
	registry.RegisterWriter(mimeType, w)
}

// RegisterObjectReader registers a typed ObjectReader for a MIME type.
// ObjectReaders guarantee the output is an Object, enabling type-safe access.
func RegisterObjectReader(mimeType string, r ObjectReader) {
	registry.RegisterObjectReader(mimeType, r)
}

// RegisterArrayReader registers a typed ArrayReader for a MIME type.
// ArrayReaders guarantee the output is an Array, enabling type-safe access.
func RegisterArrayReader(mimeType string, r ArrayReader) {
	registry.RegisterArrayReader(mimeType, r)
}

// RegisterExtension registers a file extension (including dot, e.g., ".json") to a MIME type.
func RegisterExtension(ext, mimeType string) {
	registry.RegisterExtension(ext, mimeType)
}

// DetectMimeType attempts to determine the MIME type from a filename's extension.
// Returns empty string if unknown.
func DetectMimeType(filename string) string {
	return registry.DetectMimeType(filename)
}

// MimeTypeForFormat returns the MIME type for a format name (e.g., "json" -> "application/json").
// Returns empty string if unknown.
func MimeTypeForFormat(format string) string {
	return registry.MimeTypeForFormat(format)
}

// GetReader returns the reader registered for the given MIME type.
func GetReader(mimeType string) (Reader, error) {
	return registry.GetReader(mimeType)
}

// GetWriter returns the writer registered for the given MIME type.
func GetWriter(mimeType string) (Writer, error) {
	return registry.GetWriter(mimeType)
}

// GetObjectReader returns the typed ObjectReader for a MIME type, if registered.
// Returns nil if no ObjectReader is registered (even if a generic Reader exists).
func GetObjectReader(mimeType string) (ObjectReader, bool) {
	return registry.GetObjectReader(mimeType)
}

// GetArrayReader returns the typed ArrayReader for a MIME type, if registered.
// Returns nil if no ArrayReader is registered (even if a generic Reader exists).
func GetArrayReader(mimeType string) (ArrayReader, bool) {
	return registry.GetArrayReader(mimeType)
}

// HasTypedReader returns true if a typed reader (ObjectReader or ArrayReader) is registered.
func HasTypedReader(mimeType string) bool {
	return registry.HasTypedReader(mimeType)
}
