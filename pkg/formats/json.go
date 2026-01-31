package formats

import (
	"encoding/json"
	"fmt"
	unifiederrors "infomunge/internal/errors"
)

func init() {
	RegisterReader("application/json", readJSON)
	RegisterWriter("application/json", formatJSON)
	RegisterExtension(".json", "application/json")
}

func readJSON(content string) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal([]byte(content), &result)
	if err != nil {
		return nil, unifiederrors.WrapValidationf(err, "JSON parse error: %v", err)
	}
	return result, nil
}

func formatJSON(result interface{}) (string, error) {
	switch v := result.(type) {
	case string:
		return marshalToJSON(v), nil
	case int, float64, bool:
		return fmt.Sprintf("%v", v), nil
	default:
		return marshalToJSON(v), nil
	}
}
