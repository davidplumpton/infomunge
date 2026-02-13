package formats

import (
	"encoding/json"
	"strings"

	unifiederrors "infomunge/internal/errors"
)

func init() {
	RegisterReader("application/x-ndjson", readNDJSON)
	RegisterWriter("application/x-ndjson", formatNDJSON)
	RegisterReader("application/ndjson", readNDJSON)
	RegisterWriter("application/ndjson", formatNDJSON)
	RegisterExtension(".ndjson", "application/x-ndjson")
	RegisterExtension(".jsonl", "application/x-ndjson")
}

func readNDJSON(content string) (interface{}, error) {
	lines := strings.Split(content, "\n")
	var result Array
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var val interface{}
		if err := json.Unmarshal([]byte(line), &val); err != nil {
			return nil, unifiederrors.WrapValidationf(err, "NDJSON parse error on line %d: %v", i+1, err)
		}
		result = append(result, val)
	}
	if result == nil {
		return Array{}, nil
	}
	return result, nil
}

func formatNDJSON(result interface{}) (string, error) {
	items, ok := result.(Array)
	if !ok {
		return "", unifiederrors.ValidationErrorf("NDJSON output expects an array, got %T", result)
	}
	lines := make([]string, 0, len(items))
	for _, item := range items {
		b, err := json.Marshal(item)
		if err != nil {
			return "", unifiederrors.WrapValidationf(err, "NDJSON marshal error: %v", err)
		}
		lines = append(lines, string(b))
	}
	return strings.Join(lines, "\n"), nil
}
