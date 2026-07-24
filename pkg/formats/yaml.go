package formats

import (
	"gopkg.in/yaml.v3"
	unifiederrors "infomunge/internal/errors"
	"infomunge/pkg/values"
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
	var document yaml.Node
	err := yaml.Unmarshal([]byte(content), &document)
	if err != nil {
		return nil, unifiederrors.WrapValidationf(err, "YAML parse error: %v", err)
	}
	return decodeYAMLNode(&document)
}

func formatYAML(result interface{}) (string, error) {
	node, err := encodeYAMLNode(result)
	if err != nil {
		return "", unifiederrors.WrapValidationf(err, "YAML format error: %v", err)
	}
	bytes, err := yaml.Marshal(node)
	if err != nil {
		return "", unifiederrors.WrapValidationf(err, "YAML format error: %v", err)
	}
	return string(bytes), nil
}

func decodeYAMLNode(node *yaml.Node) (interface{}, error) {
	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 {
			return nil, nil
		}
		return decodeYAMLNode(node.Content[0])
	case yaml.MappingNode:
		object := values.NewObject(len(node.Content) / 2)
		explicitKeys := make(map[string]struct{}, len(node.Content)/2)
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i].Value
			if key == "<<" {
				merged, err := decodeYAMLNode(node.Content[i+1])
				if err != nil {
					return nil, err
				}
				if mergedObject, ok := merged.(Object); ok {
					for _, mergedKey := range values.ObjectKeys(mergedObject) {
						if _, exists := object[mergedKey]; !exists {
							values.SetObjectValue(object, mergedKey, mergedObject[mergedKey])
						}
					}
				}
				continue
			}
			if _, exists := explicitKeys[key]; exists {
				return nil, unifiederrors.ValidationErrorf("YAML parse error: duplicate key %q", key)
			}
			explicitKeys[key] = struct{}{}
			value, err := decodeYAMLNode(node.Content[i+1])
			if err != nil {
				return nil, err
			}
			values.SetObjectValue(object, key, value)
		}
		return object, nil
	case yaml.SequenceNode:
		array := make(Array, 0, len(node.Content))
		for _, child := range node.Content {
			value, err := decodeYAMLNode(child)
			if err != nil {
				return nil, err
			}
			array = append(array, value)
		}
		return array, nil
	case yaml.AliasNode:
		return decodeYAMLNode(node.Alias)
	default:
		var scalar interface{}
		if err := node.Decode(&scalar); err != nil {
			return nil, err
		}
		return scalar, nil
	}
}

func encodeYAMLNode(value interface{}) (*yaml.Node, error) {
	switch typed := value.(type) {
	case Object:
		node := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		for _, key := range values.ObjectKeys(typed) {
			valueNode, err := encodeYAMLNode(typed[key])
			if err != nil {
				return nil, err
			}
			node.Content = append(node.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
				valueNode,
			)
		}
		return node, nil
	case Array:
		node := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, item := range typed {
			itemNode, err := encodeYAMLNode(item)
			if err != nil {
				return nil, err
			}
			node.Content = append(node.Content, itemNode)
		}
		return node, nil
	default:
		node := &yaml.Node{}
		if err := node.Encode(value); err != nil {
			return nil, err
		}
		return node, nil
	}
}
