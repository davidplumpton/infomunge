package formats

import (
	"encoding/binary"
	unifiederrors "infomunge/internal/errors"
	"math"
	"slices"
)

func encodeProtobufMessage(value interface{}, schema protobufSchema, strict bool) ([]byte, error) {
	obj, ok := value.(map[string]interface{})
	if !ok {
		return nil, unifiederrors.ValidationErrorf("protobuf structured write expects object input, got %T", value)
	}

	if strict {
		for key := range obj {
			if _, exists := schema.byName[key]; !exists {
				return nil, unifiederrors.ValidationErrorf("protobuf input contains unknown field %q", key)
			}
		}
	}

	var out []byte
	for _, field := range schema.fields {
		rawValue, present := obj[field.name]
		if !present || rawValue == nil {
			continue
		}

		if field.kind == "map" {
			mapValue, ok := rawValue.(map[string]interface{})
			if !ok {
				return nil, unifiederrors.ValidationErrorf("protobuf field %q is map and expects an object, got %T", field.name, rawValue)
			}
			keys := make([]string, 0, len(mapValue))
			for key := range mapValue {
				keys = append(keys, key)
			}
			slices.Sort(keys)
			for _, key := range keys {
				encoded, err := encodeProtobufMapEntry(field, key, mapValue[key], strict)
				if err != nil {
					return nil, err
				}
				out = append(out, encoded...)
			}
			continue
		}

		if field.repeated {
			values, ok := rawValue.([]interface{})
			if !ok {
				return nil, unifiederrors.ValidationErrorf("protobuf field %q is repeated and expects an array, got %T", field.name, rawValue)
			}
			if field.packed && isProtobufPackableField(field.kind) {
				encoded, err := encodePackedProtobufField(field, values, strict)
				if err != nil {
					return nil, err
				}
				out = append(out, encoded...)
				continue
			}
			for _, item := range values {
				encoded, err := encodeProtobufField(field, item, strict)
				if err != nil {
					return nil, err
				}
				out = append(out, encoded...)
			}
			continue
		}

		encoded, err := encodeProtobufField(field, rawValue, strict)
		if err != nil {
			return nil, err
		}
		out = append(out, encoded...)
	}

	return out, nil
}

func encodeProtobufMapEntry(field protobufField, key string, value interface{}, strict bool) ([]byte, error) {
	typedKey, err := protobufParseMapKey(field.mapKeyKind, key)
	if err != nil {
		return nil, unifiederrors.ValidationErrorf("protobuf field %q map key %q is invalid for key type %q", field.name, key, field.mapKeyKind)
	}

	keyField := protobufField{
		number: 1,
		name:   field.name + ".key",
		kind:   field.mapKeyKind,
	}
	valueField := protobufField{
		number: 2,
		name:   field.name + ".value",
		kind:   field.mapValueKind,
		schema: field.mapValueSchema,
	}

	if value == nil {
		value = protobufDefaultValueForKind(valueField)
	}

	entry := make([]byte, 0, 24)
	keyEncoded, err := encodeProtobufField(keyField, typedKey, strict)
	if err != nil {
		return nil, err
	}
	entry = append(entry, keyEncoded...)

	valueEncoded, err := encodeProtobufField(valueField, value, strict)
	if err != nil {
		return nil, err
	}
	entry = append(entry, valueEncoded...)

	tag := uint64(field.number<<3 | 2)
	out := appendVarint(nil, tag)
	out = appendVarint(out, uint64(len(entry)))
	out = append(out, entry...)
	return out, nil
}

func encodePackedProtobufField(field protobufField, values []interface{}, strict bool) ([]byte, error) {
	if len(values) == 0 {
		return nil, nil
	}

	payload := make([]byte, 0, len(values)*4)
	for _, value := range values {
		encodedValue, err := encodeProtobufFieldValue(field, value, strict)
		if err != nil {
			return nil, err
		}
		payload = append(payload, encodedValue...)
	}

	tag := uint64(field.number<<3 | 2)
	out := appendVarint(nil, tag)
	out = appendVarint(out, uint64(len(payload)))
	out = append(out, payload...)
	return out, nil
}

func encodeProtobufField(field protobufField, value interface{}, strict bool) ([]byte, error) {
	wireType, ok := protobufWireTypeForField(field.kind)
	if !ok {
		return nil, unifiederrors.ValidationErrorf("protobuf field %q has unsupported type %q", field.name, field.kind)
	}
	tag := uint64(field.number<<3 | wireType)

	out := make([]byte, 0, 16)
	out = appendVarint(out, tag)

	encodedValue, err := encodeProtobufFieldValue(field, value, strict)
	if err != nil {
		return nil, err
	}
	out = append(out, encodedValue...)

	return out, nil
}

func encodeProtobufFieldValue(field protobufField, value interface{}, strict bool) ([]byte, error) {
	switch field.kind {
	case "int32":
		n, err := numberToInt64(value)
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q expects int32-compatible number, got %T", field.name, value)
		}
		if n < math.MinInt32 || n > math.MaxInt32 {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q int32 value out of range: %d", field.name, n)
		}
		return appendVarint(nil, uint64(uint32(int32(n)))), nil
	case "int64":
		n, err := numberToInt64(value)
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q expects int64-compatible number, got %T", field.name, value)
		}
		return appendVarint(nil, uint64(n)), nil
	case "uint32":
		n, err := numberToUint64(value)
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q expects uint32-compatible number, got %T", field.name, value)
		}
		if n > math.MaxUint32 {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q uint32 value out of range: %d", field.name, n)
		}
		return appendVarint(nil, n), nil
	case "uint64":
		n, err := numberToUint64(value)
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q expects uint64-compatible number, got %T", field.name, value)
		}
		return appendVarint(nil, n), nil
	case "sint32":
		n, err := numberToInt64(value)
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q expects sint32-compatible number, got %T", field.name, value)
		}
		if n < math.MinInt32 || n > math.MaxInt32 {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q sint32 value out of range: %d", field.name, n)
		}
		return appendVarint(nil, uint64(encodeZigZag32(int32(n)))), nil
	case "sint64":
		n, err := numberToInt64(value)
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q expects sint64-compatible number, got %T", field.name, value)
		}
		return appendVarint(nil, encodeZigZag64(n)), nil
	case "bool":
		v, ok := value.(bool)
		if !ok {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q expects bool, got %T", field.name, value)
		}
		if v {
			return []byte{1}, nil
		}
		return []byte{0}, nil
	case "enum":
		n, err := numberToInt64(value)
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q expects enum numeric value, got %T", field.name, value)
		}
		return appendVarint(nil, uint64(n)), nil
	case "double":
		f, err := numberToFloat64(value)
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q expects double-compatible number, got %T", field.name, value)
		}
		out := make([]byte, 8)
		binary.LittleEndian.PutUint64(out, math.Float64bits(f))
		return out, nil
	case "float":
		f, err := numberToFloat64(value)
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q expects float-compatible number, got %T", field.name, value)
		}
		if f < -math.MaxFloat32 || f > math.MaxFloat32 {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q float value out of range: %v", field.name, f)
		}
		out := make([]byte, 4)
		binary.LittleEndian.PutUint32(out, math.Float32bits(float32(f)))
		return out, nil
	case "fixed64":
		n, err := numberToUint64(value)
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q expects fixed64-compatible number, got %T", field.name, value)
		}
		out := make([]byte, 8)
		binary.LittleEndian.PutUint64(out, n)
		return out, nil
	case "sfixed64":
		n, err := numberToInt64(value)
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q expects sfixed64-compatible number, got %T", field.name, value)
		}
		out := make([]byte, 8)
		binary.LittleEndian.PutUint64(out, uint64(n))
		return out, nil
	case "fixed32":
		n, err := numberToUint64(value)
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q expects fixed32-compatible number, got %T", field.name, value)
		}
		if n > math.MaxUint32 {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q fixed32 value out of range: %d", field.name, n)
		}
		out := make([]byte, 4)
		binary.LittleEndian.PutUint32(out, uint32(n))
		return out, nil
	case "sfixed32":
		n, err := numberToInt64(value)
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q expects sfixed32-compatible number, got %T", field.name, value)
		}
		if n < math.MinInt32 || n > math.MaxInt32 {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q sfixed32 value out of range: %d", field.name, n)
		}
		out := make([]byte, 4)
		binary.LittleEndian.PutUint32(out, uint32(int32(n)))
		return out, nil
	case "string":
		s, ok := value.(string)
		if !ok {
			return nil, unifiederrors.ValidationErrorf("protobuf field %q expects string, got %T", field.name, value)
		}
		out := appendVarint(nil, uint64(len(s)))
		out = append(out, []byte(s)...)
		return out, nil
	case "bytes":
		var b []byte
		switch v := value.(type) {
		case string:
			b = []byte(v)
		case []byte:
			b = v
		default:
			return nil, unifiederrors.ValidationErrorf("protobuf field %q expects string or []byte, got %T", field.name, value)
		}
		out := appendVarint(nil, uint64(len(b)))
		out = append(out, b...)
		return out, nil
	case "message":
		if field.schema == nil {
			return nil, unifiederrors.ValidationErrorf("protobuf message field %q is missing nested schema", field.name)
		}
		encoded, err := encodeProtobufMessage(value, *field.schema, strict)
		if err != nil {
			return nil, err
		}
		out := appendVarint(nil, uint64(len(encoded)))
		out = append(out, encoded...)
		return out, nil
	default:
		return nil, unifiederrors.ValidationErrorf("protobuf field %q has unsupported type %q", field.name, field.kind)
	}
}
