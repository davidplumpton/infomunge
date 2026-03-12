package formats

import (
	unifiederrors "infomunge/internal/errors"
	"strings"
)

type protobufFormatOptions struct {
	structured bool
	strict     bool
	schema     protobufSchema
}

func readProtobufWithOptions(content string, options Object) (interface{}, error) {
	parsed, err := parseProtobufReadOptions(options)
	if err != nil {
		return nil, err
	}
	if !parsed.structured {
		return readBinary(content)
	}

	return decodeProtobufMessage([]byte(content), parsed.schema, parsed.strict)
}

func formatProtobufWithOptions(result interface{}, options Object) (string, error) {
	parsed, err := parseProtobufWriteOptions(options)
	if err != nil {
		return "", err
	}
	if !parsed.structured {
		return formatBinary(result)
	}

	bytes, err := encodeProtobufMessage(result, parsed.schema, parsed.strict)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func parseProtobufReadOptions(options Object) (protobufFormatOptions, error) {
	return parseProtobufOptions(options, false)
}

func parseProtobufWriteOptions(options Object) (protobufFormatOptions, error) {
	return parseProtobufOptions(options, true)
}

func parseProtobufOptions(options Object, allowWriteOnly bool) (protobufFormatOptions, error) {
	parsed := protobufFormatOptions{
		structured: false,
		strict:     true,
	}
	if options == nil {
		return parsed, nil
	}

	var rawSchema interface{}
	var rawDescriptor interface{}
	var rawDescriptorSet interface{}
	var rawMessage interface{}
	for key, raw := range options {
		switch key {
		case "structured":
			v, ok := raw.(bool)
			if !ok {
				return protobufFormatOptions{}, unifiederrors.ValidationErrorf("protobuf option structured must be a boolean, got %T", raw)
			}
			parsed.structured = v
		case "strict":
			v, ok := raw.(bool)
			if !ok {
				return protobufFormatOptions{}, unifiederrors.ValidationErrorf("protobuf option strict must be a boolean, got %T", raw)
			}
			parsed.strict = v
		case "schema":
			rawSchema = raw
		case "descriptor":
			rawDescriptor = raw
		case "descriptorSet":
			rawDescriptorSet = raw
		case "message":
			rawMessage = raw
		default:
			if !allowWriteOnly || key != "preserveUnknown" {
				return protobufFormatOptions{}, unifiederrors.ValidationErrorf("unsupported protobuf option: %s", key)
			}
		}
	}

	if rawDescriptor != nil {
		if rawDescriptorSet != nil {
			return protobufFormatOptions{}, unifiederrors.ValidationError("protobuf options descriptor and descriptorSet are mutually exclusive")
		}
		if descriptorObj, ok := rawDescriptor.(map[string]interface{}); ok {
			rawSet, hasSet := descriptorObj["set"]
			if !hasSet {
				rawSet, hasSet = descriptorObj["descriptorSet"]
			}
			if !hasSet {
				return protobufFormatOptions{}, unifiederrors.ValidationError("protobuf option descriptor must include set")
			}
			rawDescriptorSet = rawSet
			if rawDescriptorMessage, hasMessage := descriptorObj["message"]; hasMessage {
				if rawMessage != nil {
					return protobufFormatOptions{}, unifiederrors.ValidationError("protobuf message cannot be set in both descriptor and top-level message option")
				}
				rawMessage = rawDescriptorMessage
			}
		} else {
			rawDescriptorSet = rawDescriptor
		}
	}

	if parsed.structured {
		if rawSchema != nil && (rawDescriptorSet != nil || rawMessage != nil) {
			return protobufFormatOptions{}, unifiederrors.ValidationError("protobuf structured mode options schema and descriptor are mutually exclusive")
		}
		if rawSchema != nil {
			schema, err := parseProtobufSchema(rawSchema, "schema")
			if err != nil {
				return protobufFormatOptions{}, err
			}
			parsed.schema = schema
		} else if rawDescriptorSet != nil || rawMessage != nil {
			schema, err := parseProtobufSchemaFromDescriptor(rawDescriptorSet, rawMessage)
			if err != nil {
				return protobufFormatOptions{}, err
			}
			parsed.schema = schema
		} else {
			return protobufFormatOptions{}, unifiederrors.ValidationError("protobuf structured mode requires schema or descriptor")
		}
	}

	if !parsed.structured && rawSchema != nil {
		return protobufFormatOptions{}, unifiederrors.ValidationError("protobuf option schema requires structured=true")
	}
	if !parsed.structured && rawDescriptor != nil {
		return protobufFormatOptions{}, unifiederrors.ValidationError("protobuf option descriptor requires structured=true")
	}
	if !parsed.structured && rawDescriptorSet != nil {
		return protobufFormatOptions{}, unifiederrors.ValidationError("protobuf option descriptorSet requires structured=true")
	}
	if !parsed.structured && rawMessage != nil {
		return protobufFormatOptions{}, unifiederrors.ValidationError("protobuf option message requires structured=true")
	}

	return parsed, nil
}

func protobufBool(v interface{}, defaultValue bool) bool {
	if v == nil {
		return defaultValue
	}
	b, ok := v.(bool)
	if !ok {
		return defaultValue
	}
	return b
}

func isSupportedProtobufFieldType(kind string) bool {
	_, ok := protobufWireTypeForField(kind)
	return ok
}

func parseProtobufMessageName(raw interface{}) (string, error) {
	if raw == nil {
		return "", nil
	}
	name, ok := raw.(string)
	if !ok {
		return "", unifiederrors.ValidationErrorf("protobuf message must be a non-empty string, got %T", raw)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", unifiederrors.ValidationError("protobuf message must be a non-empty string")
	}
	return name, nil
}

func parseProtobufDescriptorSetBytes(raw interface{}) ([]byte, error) {
	switch v := raw.(type) {
	case string:
		if v == "" {
			return nil, unifiederrors.ValidationError("protobuf descriptor set must not be empty")
		}
		return []byte(v), nil
	case []byte:
		if len(v) == 0 {
			return nil, unifiederrors.ValidationError("protobuf descriptor set must not be empty")
		}
		return v, nil
	default:
		return nil, unifiederrors.ValidationErrorf("protobuf descriptor set must be string or []byte, got %T", raw)
	}
}
