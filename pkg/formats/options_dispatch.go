package formats

type readOptionsHandler func(content string, options Object) (interface{}, error)
type writeOptionsHandler func(result interface{}, options Object) (string, error)

var optionMimeAliases = map[string]string{
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": "application/xlsx",
	"application/x-protobuf": "application/protobuf",
}

var readOptionsHandlers = map[string]readOptionsHandler{
	"application/flatfile": readFlatfileWithOptions,
	"application/java":     readJavaWithOptions,
	"application/xlsx":     readXLSXWithOptions,
	"application/protobuf": readProtobufWithOptions,
}

var writeOptionsHandlers = map[string]writeOptionsHandler{
	"application/flatfile": formatFlatfileWithOptions,
	"application/java":     formatJavaWithOptions,
	"application/xlsx":     formatXLSXWithOptions,
	"application/protobuf": formatProtobufWithOptions,
}

func canonicalOptionsMimeType(mimeType string) string {
	if canonical, ok := optionMimeAliases[mimeType]; ok {
		return canonical
	}
	return mimeType
}

func getReadOptionsHandler(mimeType string) (readOptionsHandler, bool) {
	h, ok := readOptionsHandlers[canonicalOptionsMimeType(mimeType)]
	return h, ok
}

func getWriteOptionsHandler(mimeType string) (writeOptionsHandler, bool) {
	h, ok := writeOptionsHandlers[canonicalOptionsMimeType(mimeType)]
	return h, ok
}
