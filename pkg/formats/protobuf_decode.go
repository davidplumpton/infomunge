package formats

import (
	"encoding/binary"
	unifiederrors "infomunge/internal/errors"
	"math"
)

func decodeProtobufMessage(content []byte, schema protobufSchema, strict bool) (Object, error) {
	out := make(Object)
	offset := 0
	seen := make(map[int]bool)

	for offset < len(content) {
		tag, n := binary.Uvarint(content[offset:])
		if n <= 0 {
			return nil, unifiederrors.ValidationErrorf("invalid protobuf tag at byte offset %d", offset)
		}
		offset += n

		fieldNumber := int(tag >> 3)
		if fieldNumber <= 0 {
			return nil, unifiederrors.ValidationErrorf("invalid protobuf field number at byte offset %d", offset-n)
		}
		wireType := int(tag & 0x7)

		field, known := schema.byNumber[fieldNumber]
		if !known {
			next, err := skipUnknownProtobufField(content, offset, wireType)
			if err != nil {
				return nil, err
			}
			offset = next
			if strict {
				return nil, unifiederrors.ValidationErrorf("protobuf payload contains unknown field number %d", fieldNumber)
			}
			continue
		}

		value, next, err := decodeProtobufField(content, offset, field, wireType, strict)
		if err != nil {
			return nil, err
		}
		offset = next

		if field.kind == "map" {
			entry, ok := value.(protobufMapEntry)
			if !ok {
				return nil, unifiederrors.ValidationErrorf("protobuf map field %q decode produced unexpected type %T", field.name, value)
			}
			existing, _ := out[field.name].(map[string]interface{})
			if existing == nil {
				existing = make(map[string]interface{})
			}
			existing[entry.key] = entry.value
			out[field.name] = existing
			continue
		}

		if field.repeated {
			existing, _ := out[field.name].([]interface{})
			if packedValues, ok := value.([]interface{}); ok {
				out[field.name] = append(existing, packedValues...)
			} else {
				out[field.name] = append(existing, value)
			}
			continue
		}

		if seen[field.number] && strict {
			return nil, unifiederrors.ValidationErrorf("protobuf payload repeats non-repeated field %q", field.name)
		}
		out[field.name] = value
		seen[field.number] = true
	}

	return out, nil
}

func decodeProtobufField(content []byte, offset int, field protobufField, wireType int, strict bool) (interface{}, int, error) {
	if field.kind == "map" {
		if wireType != 2 {
			return nil, offset, unifiederrors.ValidationErrorf("protobuf map field %q expects wire type 2, got %d", field.name, wireType)
		}
		length, n := binary.Uvarint(content[offset:])
		if n <= 0 {
			return nil, offset, unifiederrors.ValidationErrorf("invalid protobuf length-delimited size for field %q", field.name)
		}
		offset += n
		if int(length) < 0 || offset+int(length) > len(content) {
			return nil, offset, unifiederrors.ValidationErrorf("invalid protobuf length-delimited value for field %q: truncated payload", field.name)
		}
		raw := content[offset : offset+int(length)]
		offset += int(length)

		entry, err := decodeProtobufMapEntry(raw, field, strict)
		if err != nil {
			return nil, offset, err
		}
		return entry, offset, nil
	}

	expectedWireType, ok := protobufWireTypeForField(field.kind)
	if !ok {
		return nil, offset, unifiederrors.ValidationErrorf("protobuf field %q has unsupported type %q", field.name, field.kind)
	}
	if field.repeated && isProtobufPackableField(field.kind) && wireType == 2 {
		values, next, err := decodePackedProtobufField(content, offset, field)
		if err != nil {
			return nil, offset, err
		}
		return values, next, nil
	}
	if wireType != expectedWireType {
		return nil, offset, unifiederrors.ValidationErrorf(
			"protobuf field %q expects wire type %d, got %d",
			field.name,
			expectedWireType,
			wireType,
		)
	}

	switch wireType {
	case 0:
		raw, n := binary.Uvarint(content[offset:])
		if n <= 0 {
			return nil, offset, unifiederrors.ValidationErrorf("invalid protobuf varint for field %q", field.name)
		}
		offset += n
		value, err := decodeProtobufVarint(field.kind, raw)
		if err != nil {
			return nil, offset, err
		}
		return value, offset, nil
	case 1:
		if offset+8 > len(content) {
			return nil, offset, unifiederrors.ValidationErrorf("invalid protobuf fixed64 for field %q: truncated payload", field.name)
		}
		raw := content[offset : offset+8]
		offset += 8
		value, err := decodeProtobufFixed64(field.kind, raw)
		if err != nil {
			return nil, offset, err
		}
		return value, offset, nil
	case 2:
		length, n := binary.Uvarint(content[offset:])
		if n <= 0 {
			return nil, offset, unifiederrors.ValidationErrorf("invalid protobuf length-delimited size for field %q", field.name)
		}
		offset += n
		if int(length) < 0 || offset+int(length) > len(content) {
			return nil, offset, unifiederrors.ValidationErrorf("invalid protobuf length-delimited value for field %q: truncated payload", field.name)
		}
		raw := content[offset : offset+int(length)]
		offset += int(length)
		value, err := decodeProtobufLengthDelimited(field, raw, strict)
		if err != nil {
			return nil, offset, err
		}
		return value, offset, nil
	case 5:
		if offset+4 > len(content) {
			return nil, offset, unifiederrors.ValidationErrorf("invalid protobuf fixed32 for field %q: truncated payload", field.name)
		}
		raw := content[offset : offset+4]
		offset += 4
		value, err := decodeProtobufFixed32(field.kind, raw)
		if err != nil {
			return nil, offset, err
		}
		return value, offset, nil
	default:
		return nil, offset, unifiederrors.ValidationErrorf("protobuf field %q has unsupported wire type %d", field.name, wireType)
	}
}

func decodeProtobufVarint(kind string, value uint64) (interface{}, error) {
	switch kind {
	case "int32":
		return float64(int32(value)), nil
	case "int64":
		return float64(int64(value)), nil
	case "uint32":
		return float64(uint32(value)), nil
	case "uint64":
		if value > math.MaxInt64 {
			return nil, unifiederrors.ValidationErrorf("protobuf uint64 value %d exceeds supported range", value)
		}
		return float64(value), nil
	case "sint32":
		return float64(decodeZigZag32(uint32(value))), nil
	case "sint64":
		return float64(decodeZigZag64(value)), nil
	case "bool":
		if value == 0 {
			return false, nil
		}
		if value == 1 {
			return true, nil
		}
		return nil, unifiederrors.ValidationErrorf("protobuf bool expects 0 or 1, got %d", value)
	case "enum":
		return float64(int64(value)), nil
	default:
		return nil, unifiederrors.ValidationErrorf("protobuf varint type %q is not supported", kind)
	}
}

func decodeProtobufFixed64(kind string, raw []byte) (interface{}, error) {
	bits64 := binary.LittleEndian.Uint64(raw)
	switch kind {
	case "double":
		return math.Float64frombits(bits64), nil
	case "fixed64":
		if bits64 > math.MaxInt64 {
			return nil, unifiederrors.ValidationErrorf("protobuf fixed64 value %d exceeds supported range", bits64)
		}
		return float64(bits64), nil
	case "sfixed64":
		return float64(int64(bits64)), nil
	default:
		return nil, unifiederrors.ValidationErrorf("protobuf fixed64 type %q is not supported", kind)
	}
}

func decodeProtobufFixed32(kind string, raw []byte) (interface{}, error) {
	bits32 := binary.LittleEndian.Uint32(raw)
	switch kind {
	case "float":
		return float64(math.Float32frombits(bits32)), nil
	case "fixed32":
		return float64(bits32), nil
	case "sfixed32":
		return float64(int32(bits32)), nil
	default:
		return nil, unifiederrors.ValidationErrorf("protobuf fixed32 type %q is not supported", kind)
	}
}

func decodeProtobufLengthDelimited(field protobufField, raw []byte, strict bool) (interface{}, error) {
	switch field.kind {
	case "string":
		return string(raw), nil
	case "bytes":
		return string(raw), nil
	case "message":
		if field.schema == nil {
			return nil, unifiederrors.ValidationErrorf("protobuf message field %q is missing nested schema", field.name)
		}
		return decodeProtobufMessage(raw, *field.schema, strict)
	default:
		return nil, unifiederrors.ValidationErrorf("protobuf length-delimited type %q is not supported", field.kind)
	}
}

type protobufMapEntry struct {
	key   string
	value interface{}
}

func decodeProtobufMapEntry(raw []byte, field protobufField, strict bool) (protobufMapEntry, error) {
	keyField := protobufField{number: 1, name: field.name + ".key", kind: field.mapKeyKind}
	valueField := protobufField{number: 2, name: field.name + ".value", kind: field.mapValueKind, schema: field.mapValueSchema}

	offset := 0
	key := protobufDefaultMapKey(field.mapKeyKind)
	value := protobufDefaultValueForKind(valueField)
	keySeen := false
	valueSeen := false

	for offset < len(raw) {
		tag, n := binary.Uvarint(raw[offset:])
		if n <= 0 {
			return protobufMapEntry{}, unifiederrors.ValidationErrorf("invalid protobuf map entry tag for field %q", field.name)
		}
		offset += n
		fieldNumber := int(tag >> 3)
		wireType := int(tag & 0x7)

		switch fieldNumber {
		case 1:
			decoded, next, err := decodeProtobufField(raw, offset, keyField, wireType, strict)
			if err != nil {
				return protobufMapEntry{}, err
			}
			offset = next
			keyAsString, err := protobufMapKeyToString(field.mapKeyKind, decoded)
			if err != nil {
				return protobufMapEntry{}, err
			}
			key = keyAsString
			keySeen = true
		case 2:
			decoded, next, err := decodeProtobufField(raw, offset, valueField, wireType, strict)
			if err != nil {
				return protobufMapEntry{}, err
			}
			offset = next
			value = decoded
			valueSeen = true
		default:
			next, err := skipUnknownProtobufField(raw, offset, wireType)
			if err != nil {
				return protobufMapEntry{}, err
			}
			offset = next
			if strict {
				return protobufMapEntry{}, unifiederrors.ValidationErrorf("protobuf map field %q contains unknown entry field number %d", field.name, fieldNumber)
			}
		}
	}

	// Defaults follow protobuf map entry semantics when key/value are omitted.
	if !keySeen {
		key = protobufDefaultMapKey(field.mapKeyKind)
	}
	if !valueSeen {
		value = protobufDefaultValueForKind(valueField)
	}

	return protobufMapEntry{key: key, value: value}, nil
}

func decodePackedProtobufField(content []byte, offset int, field protobufField) ([]interface{}, int, error) {
	length, n := binary.Uvarint(content[offset:])
	if n <= 0 {
		return nil, offset, unifiederrors.ValidationErrorf("invalid protobuf packed length for field %q", field.name)
	}
	offset += n
	if int(length) < 0 || offset+int(length) > len(content) {
		return nil, offset, unifiederrors.ValidationErrorf("invalid protobuf packed value for field %q: truncated payload", field.name)
	}
	packed := content[offset : offset+int(length)]
	offset += int(length)

	values := make([]interface{}, 0, 4)
	packedOffset := 0
	for packedOffset < len(packed) {
		value, next, err := decodePackedProtobufElement(packed, packedOffset, field)
		if err != nil {
			return nil, offset, err
		}
		packedOffset = next
		values = append(values, value)
	}

	return values, offset, nil
}

func decodePackedProtobufElement(content []byte, offset int, field protobufField) (interface{}, int, error) {
	wireType, _ := protobufWireTypeForField(field.kind)
	switch wireType {
	case 0:
		raw, n := binary.Uvarint(content[offset:])
		if n <= 0 {
			return nil, offset, unifiederrors.ValidationErrorf("invalid protobuf packed varint for field %q", field.name)
		}
		value, err := decodeProtobufVarint(field.kind, raw)
		if err != nil {
			return nil, offset, err
		}
		return value, offset + n, nil
	case 1:
		if offset+8 > len(content) {
			return nil, offset, unifiederrors.ValidationErrorf("invalid protobuf packed fixed64 for field %q: truncated payload", field.name)
		}
		value, err := decodeProtobufFixed64(field.kind, content[offset:offset+8])
		if err != nil {
			return nil, offset, err
		}
		return value, offset + 8, nil
	case 5:
		if offset+4 > len(content) {
			return nil, offset, unifiederrors.ValidationErrorf("invalid protobuf packed fixed32 for field %q: truncated payload", field.name)
		}
		value, err := decodeProtobufFixed32(field.kind, content[offset:offset+4])
		if err != nil {
			return nil, offset, err
		}
		return value, offset + 4, nil
	default:
		return nil, offset, unifiederrors.ValidationErrorf("protobuf field %q type %q cannot be packed", field.name, field.kind)
	}
}
