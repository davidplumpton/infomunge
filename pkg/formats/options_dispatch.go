package formats

import formatcore "infomunge/pkg/formats/core"

// ReadOptionsHandler parses content using format-specific options.
type ReadOptionsHandler = formatcore.ReadOptionsHandler

// WriteOptionsHandler serializes a value using format-specific options.
type WriteOptionsHandler = formatcore.WriteOptionsHandler

// RegisterOptionsAlias maps an alternate MIME type to another option-handler MIME type.
func RegisterOptionsAlias(alias, canonical string) {
	registry.RegisterOptionsAlias(alias, canonical)
}

// RegisterReadOptionsHandler registers a read options handler for a MIME type.
func RegisterReadOptionsHandler(mimeType string, handler ReadOptionsHandler) {
	registry.RegisterReadOptionsHandler(mimeType, handler)
}

// RegisterWriteOptionsHandler registers a write options handler for a MIME type.
func RegisterWriteOptionsHandler(mimeType string, handler WriteOptionsHandler) {
	registry.RegisterWriteOptionsHandler(mimeType, handler)
}

func getReadOptionsHandler(mimeType string) (ReadOptionsHandler, bool) {
	return registry.GetReadOptionsHandler(mimeType)
}

func getWriteOptionsHandler(mimeType string) (WriteOptionsHandler, bool) {
	return registry.GetWriteOptionsHandler(mimeType)
}
