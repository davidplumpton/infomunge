package formats

import (
	"gopkg.in/yaml.v3"
	unifiederrors "infomunge/internal/errors"
)

func init() {
	RegisterReader("application/yaml", readYAML)
	// YAML writer is not currently explicitly handled in writer.go (defaults to JSON)
	// But we can add it here if needed. For now, let's stick to existing behavior or improve it.
	RegisterWriter("application/yaml", formatYAML)
	RegisterExtension(".yaml", "application/yaml")
	RegisterExtension(".yml", "application/yaml")
}

func readYAML(content string) (interface{}, error) {
	var result interface{}
	err := yaml.Unmarshal([]byte(content), &result)
	if err != nil {
		return nil, unifiederrors.WrapValidationf(err, "YAML parse error: %v", err)
	}
	return result, nil
}

func formatYAML(result interface{}) (string, error) {
	bytes, err := yaml.Marshal(result)
	if err != nil {
		return "", unifiederrors.WrapValidationf(err, "YAML format error: %v", err)
	}
	return string(bytes), nil
}
