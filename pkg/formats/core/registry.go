package core

import (
	unifiederrors "infomunge/internal/errors"
	"infomunge/pkg/values"
	"strings"
	"sync"
)

// Reader is a function that parses content into an interface{} representation.
// Readers should return one of: values.Object, values.Array, string, float64,
// bool, or nil.
type Reader func(content string) (interface{}, error)

// Writer is a function that serializes an interface{} into a string.
type Writer func(result interface{}) (string, error)

// ObjectReader is a typed reader that returns an object directly.
type ObjectReader func(content string) (values.Object, error)

// ArrayReader is a typed reader that returns an array directly.
type ArrayReader func(content string) (values.Array, error)

// ReadOptionsHandler parses content using format-specific options.
type ReadOptionsHandler func(content string, options values.Object) (interface{}, error)

// WriteOptionsHandler serializes a value using format-specific options.
type WriteOptionsHandler func(result interface{}, options values.Object) (string, error)

// Registry owns format readers, writers, extension lookup, and option handlers.
type Registry struct {
	mu sync.RWMutex

	readers       map[string]Reader
	writers       map[string]Writer
	objectReaders map[string]ObjectReader
	arrayReaders  map[string]ArrayReader
	extensions    map[string]string

	optionMimeAliases    map[string]string
	readOptionsHandlers  map[string]ReadOptionsHandler
	writeOptionsHandlers map[string]WriteOptionsHandler
}

// NewRegistry creates an empty format registry.
func NewRegistry() *Registry {
	return &Registry{
		readers:              make(map[string]Reader),
		writers:              make(map[string]Writer),
		objectReaders:        make(map[string]ObjectReader),
		arrayReaders:         make(map[string]ArrayReader),
		extensions:           make(map[string]string),
		optionMimeAliases:    make(map[string]string),
		readOptionsHandlers:  make(map[string]ReadOptionsHandler),
		writeOptionsHandlers: make(map[string]WriteOptionsHandler),
	}
}

// RegisterReader registers a reader for a specific MIME type.
func (r *Registry) RegisterReader(mimeType string, reader Reader) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.readers[mimeType] = reader
}

// RegisterWriter registers a writer for a specific MIME type.
func (r *Registry) RegisterWriter(mimeType string, writer Writer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.writers[mimeType] = writer
}

// RegisterObjectReader registers a typed ObjectReader for a MIME type.
func (r *Registry) RegisterObjectReader(mimeType string, reader ObjectReader) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.objectReaders[mimeType] = reader
	r.readers[mimeType] = func(content string) (interface{}, error) {
		return reader(content)
	}
}

// RegisterArrayReader registers a typed ArrayReader for a MIME type.
func (r *Registry) RegisterArrayReader(mimeType string, reader ArrayReader) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.arrayReaders[mimeType] = reader
	r.readers[mimeType] = func(content string) (interface{}, error) {
		return reader(content)
	}
}

// RegisterExtension registers a file extension (including dot, e.g., ".json") to a MIME type.
func (r *Registry) RegisterExtension(ext, mimeType string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.extensions[strings.ToLower(ext)] = mimeType
}

// RegisterOptionsAlias maps an alternate MIME type to another option-handler MIME type.
func (r *Registry) RegisterOptionsAlias(alias, canonical string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.optionMimeAliases[alias] = canonical
}

// RegisterReadOptionsHandler registers a read options handler for a MIME type.
func (r *Registry) RegisterReadOptionsHandler(mimeType string, handler ReadOptionsHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.readOptionsHandlers[mimeType] = handler
}

// RegisterWriteOptionsHandler registers a write options handler for a MIME type.
func (r *Registry) RegisterWriteOptionsHandler(mimeType string, handler WriteOptionsHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.writeOptionsHandlers[mimeType] = handler
}

// DetectMimeType attempts to determine the MIME type from a filename's extension.
// Returns empty string if unknown.
func (r *Registry) DetectMimeType(filename string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lastDot := strings.LastIndex(filename, ".")
	if lastDot == -1 {
		return ""
	}

	lastSep := strings.LastIndexAny(filename, "/\\")
	if lastDot < lastSep {
		return ""
	}

	ext := strings.ToLower(filename[lastDot:])
	if mime, ok := r.extensions[ext]; ok {
		return mime
	}
	return ""
}

// MimeTypeForFormat returns the MIME type for a format name (e.g., "json" -> "application/json").
// Returns empty string if unknown.
func (r *Registry) MimeTypeForFormat(format string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ext := "." + strings.ToLower(format)
	if mime, ok := r.extensions[ext]; ok {
		return mime
	}
	return ""
}

// GetReader returns the reader registered for the given MIME type.
func (r *Registry) GetReader(mimeType string) (Reader, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	reader, ok := r.readers[mimeType]
	if !ok {
		return nil, unifiederrors.ValidationErrorf("unsupported input mimeType: %s", mimeType)
	}
	return reader, nil
}

// GetWriter returns the writer registered for the given MIME type.
func (r *Registry) GetWriter(mimeType string) (Writer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	writer, ok := r.writers[mimeType]
	if !ok {
		return nil, unifiederrors.ValidationErrorf("unsupported output mimeType: %s", mimeType)
	}
	return writer, nil
}

// GetObjectReader returns the typed ObjectReader for a MIME type, if registered.
func (r *Registry) GetObjectReader(mimeType string) (ObjectReader, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	reader, ok := r.objectReaders[mimeType]
	return reader, ok
}

// GetArrayReader returns the typed ArrayReader for a MIME type, if registered.
func (r *Registry) GetArrayReader(mimeType string) (ArrayReader, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	reader, ok := r.arrayReaders[mimeType]
	return reader, ok
}

// HasTypedReader returns true if a typed reader (ObjectReader or ArrayReader) is registered.
func (r *Registry) HasTypedReader(mimeType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, hasObj := r.objectReaders[mimeType]
	_, hasArr := r.arrayReaders[mimeType]
	return hasObj || hasArr
}

// GetReadOptionsHandler returns a read options handler for the MIME type, following aliases.
func (r *Registry) GetReadOptionsHandler(mimeType string) (ReadOptionsHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, ok := r.readOptionsHandlers[r.canonicalOptionsMimeTypeLocked(mimeType)]
	return handler, ok
}

// GetWriteOptionsHandler returns a write options handler for the MIME type, following aliases.
func (r *Registry) GetWriteOptionsHandler(mimeType string) (WriteOptionsHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, ok := r.writeOptionsHandlers[r.canonicalOptionsMimeTypeLocked(mimeType)]
	return handler, ok
}

func (r *Registry) canonicalOptionsMimeTypeLocked(mimeType string) string {
	seen := make(map[string]struct{})
	for {
		canonical, ok := r.optionMimeAliases[mimeType]
		if !ok || canonical == mimeType {
			return mimeType
		}
		if _, exists := seen[mimeType]; exists {
			return mimeType
		}
		seen[mimeType] = struct{}{}
		mimeType = canonical
	}
}
