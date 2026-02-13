package formats

import (
	"encoding/binary"
	"fmt"
	unifiederrors "infomunge/internal/errors"
	"math"
	"slices"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

type protobufFormatOptions struct {
	structured bool
	strict     bool
	schema     protobufSchema
}

type protobufSchema struct {
	messageName string
	fields      []protobufField
	byNumber    map[int]protobufField
	byName      map[string]protobufField
}

type protobufField struct {
	number   int
	name     string
	kind     string
	repeated bool
	packed   bool
	schema   *protobufSchema
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

func parseProtobufSchemaFromDescriptor(rawDescriptorSet interface{}, rawMessage interface{}) (protobufSchema, error) {
	if rawDescriptorSet == nil {
		return protobufSchema{}, unifiederrors.ValidationError("protobuf descriptor options require descriptor set bytes")
	}
	messageName, err := parseProtobufMessageName(rawMessage)
	if err != nil {
		return protobufSchema{}, err
	}
	if messageName == "" {
		return protobufSchema{}, unifiederrors.ValidationError("protobuf descriptor options require message")
	}

	descriptorSetBytes, err := parseProtobufDescriptorSetBytes(rawDescriptorSet)
	if err != nil {
		return protobufSchema{}, err
	}

	var descriptorSet descriptorpb.FileDescriptorSet
	if err := proto.Unmarshal(descriptorSetBytes, &descriptorSet); err != nil {
		return protobufSchema{}, unifiederrors.ValidationErrorf("invalid protobuf descriptor set: %v", err)
	}

	files, err := protodesc.NewFiles(&descriptorSet)
	if err != nil {
		return protobufSchema{}, unifiederrors.ValidationErrorf("invalid protobuf descriptor set: %v", err)
	}

	messageDesc, err := findProtobufMessageDescriptor(files, messageName)
	if err != nil {
		return protobufSchema{}, err
	}

	return protobufSchemaFromMessageDescriptor(messageDesc)
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

func findProtobufMessageDescriptor(files *protoregistry.Files, messageName string) (protoreflect.MessageDescriptor, error) {
	fullName := protoreflect.FullName(messageName)
	desc, err := files.FindDescriptorByName(fullName)
	if err == nil {
		msg, ok := desc.(protoreflect.MessageDescriptor)
		if !ok {
			return nil, unifiederrors.ValidationErrorf("protobuf descriptor %q is not a message", messageName)
		}
		return msg, nil
	}

	var messages []protoreflect.MessageDescriptor
	files.RangeFiles(func(file protoreflect.FileDescriptor) bool {
		collectProtobufMessageDescriptors(file.Messages(), &messages)
		return true
	})
	var matches []protoreflect.MessageDescriptor
	for _, candidate := range messages {
		candidateName := string(candidate.FullName())
		if candidateName == messageName || strings.HasSuffix(candidateName, "."+messageName) {
			matches = append(matches, candidate)
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		return nil, unifiederrors.ValidationErrorf("protobuf message %q is ambiguous; use fully-qualified name", messageName)
	}
	return nil, unifiederrors.ValidationErrorf("protobuf message %q was not found in descriptor set", messageName)
}

func collectProtobufMessageDescriptors(messages protoreflect.MessageDescriptors, out *[]protoreflect.MessageDescriptor) {
	for i := 0; i < messages.Len(); i++ {
		message := messages.Get(i)
		*out = append(*out, message)
		collectProtobufMessageDescriptors(message.Messages(), out)
	}
}

func protobufSchemaFromMessageDescriptor(message protoreflect.MessageDescriptor) (protobufSchema, error) {
	fields := message.Fields()
	schema := protobufSchema{
		messageName: string(message.FullName()),
		fields:      make([]protobufField, 0, fields.Len()),
		byNumber:    make(map[int]protobufField, fields.Len()),
		byName:      make(map[string]protobufField, fields.Len()),
	}

	for i := 0; i < fields.Len(); i++ {
		descriptorField := fields.Get(i)
		if descriptorField.IsMap() {
			return protobufSchema{}, unifiederrors.ValidationErrorf("protobuf descriptor field %q map type is not supported", descriptorField.FullName())
		}

		kind, ok := protobufKindFromDescriptorKind(descriptorField.Kind())
		if !ok {
			return protobufSchema{}, unifiederrors.ValidationErrorf(
				"protobuf descriptor field %q has unsupported kind %q",
				descriptorField.FullName(),
				descriptorField.Kind().String(),
			)
		}

		field := protobufField{
			number:   int(descriptorField.Number()),
			name:     string(descriptorField.Name()),
			kind:     kind,
			repeated: descriptorField.Cardinality() == protoreflect.Repeated,
			packed:   descriptorField.IsPacked(),
		}
		if kind == "message" {
			nested, err := protobufSchemaFromMessageDescriptor(descriptorField.Message())
			if err != nil {
				return protobufSchema{}, err
			}
			field.schema = &nested
		}
		if field.packed && (!field.repeated || !isProtobufPackableField(field.kind)) {
			return protobufSchema{}, unifiederrors.ValidationErrorf("protobuf descriptor field %q cannot be packed", descriptorField.FullName())
		}

		schema.fields = append(schema.fields, field)
		schema.byNumber[field.number] = field
		schema.byName[field.name] = field
	}

	slices.SortFunc(schema.fields, func(a, b protobufField) int {
		if a.number < b.number {
			return -1
		}
		if a.number > b.number {
			return 1
		}
		return 0
	})

	return schema, nil
}

func protobufKindFromDescriptorKind(kind protoreflect.Kind) (string, bool) {
	switch kind {
	case protoreflect.BoolKind:
		return "bool", true
	case protoreflect.EnumKind:
		return "enum", true
	case protoreflect.Int32Kind:
		return "int32", true
	case protoreflect.Sint32Kind:
		return "sint32", true
	case protoreflect.Sfixed32Kind:
		return "sfixed32", true
	case protoreflect.Int64Kind:
		return "int64", true
	case protoreflect.Sint64Kind:
		return "sint64", true
	case protoreflect.Sfixed64Kind:
		return "sfixed64", true
	case protoreflect.Uint32Kind:
		return "uint32", true
	case protoreflect.Fixed32Kind:
		return "fixed32", true
	case protoreflect.Uint64Kind:
		return "uint64", true
	case protoreflect.Fixed64Kind:
		return "fixed64", true
	case protoreflect.FloatKind:
		return "float", true
	case protoreflect.DoubleKind:
		return "double", true
	case protoreflect.StringKind:
		return "string", true
	case protoreflect.BytesKind:
		return "bytes", true
	case protoreflect.MessageKind:
		return "message", true
	default:
		return "", false
	}
}

func parseProtobufSchema(raw interface{}, path string) (protobufSchema, error) {
	obj, ok := raw.(map[string]interface{})
	if !ok {
		return protobufSchema{}, unifiederrors.ValidationErrorf("protobuf %s must be an object, got %T", path, raw)
	}

	fieldsRaw, hasFields := obj["fields"]
	if !hasFields {
		return protobufSchema{}, unifiederrors.ValidationErrorf("protobuf %s must include fields", path)
	}
	fieldsSlice, ok := fieldsRaw.([]interface{})
	if !ok || len(fieldsSlice) == 0 {
		return protobufSchema{}, unifiederrors.ValidationErrorf("protobuf %s fields must be a non-empty array", path)
	}

	schema := protobufSchema{
		fields:   make([]protobufField, 0, len(fieldsSlice)),
		byNumber: make(map[int]protobufField, len(fieldsSlice)),
		byName:   make(map[string]protobufField, len(fieldsSlice)),
	}

	if rawMessage, ok := obj["message"]; ok {
		name, ok := rawMessage.(string)
		if !ok || strings.TrimSpace(name) == "" {
			return protobufSchema{}, unifiederrors.ValidationErrorf("protobuf %s message must be a non-empty string", path)
		}
		schema.messageName = strings.TrimSpace(name)
	}

	for idx, rawField := range fieldsSlice {
		fieldObj, ok := rawField.(map[string]interface{})
		if !ok {
			return protobufSchema{}, unifiederrors.ValidationErrorf("protobuf %s field %d must be an object, got %T", path, idx+1, rawField)
		}

		number, err := protobufPositiveInt(fieldObj["number"])
		if err != nil {
			return protobufSchema{}, unifiederrors.ValidationErrorf("protobuf %s field %d must have a positive integer number", path, idx+1)
		}

		name, ok := fieldObj["name"].(string)
		if !ok || strings.TrimSpace(name) == "" {
			return protobufSchema{}, unifiederrors.ValidationErrorf("protobuf %s field %d must have a non-empty name", path, idx+1)
		}
		name = strings.TrimSpace(name)

		kind, ok := fieldObj["type"].(string)
		if !ok || strings.TrimSpace(kind) == "" {
			return protobufSchema{}, unifiederrors.ValidationErrorf("protobuf %s field %d must have a non-empty type", path, idx+1)
		}
		kind = strings.ToLower(strings.TrimSpace(kind))
		if !isSupportedProtobufFieldType(kind) {
			return protobufSchema{}, unifiederrors.ValidationErrorf("protobuf %s field %q has unsupported type %q", path, name, kind)
		}

		repeated := protobufBool(fieldObj["repeated"], false)
		packed := protobufBool(fieldObj["packed"], false)
		field := protobufField{
			number:   number,
			name:     name,
			kind:     kind,
			repeated: repeated,
			packed:   packed,
		}
		if packed && (!repeated || !isProtobufPackableField(kind)) {
			return protobufSchema{}, unifiederrors.ValidationErrorf("protobuf %s field %q cannot be packed", path, name)
		}

		if kind == "message" {
			nestedRaw, ok := fieldObj["schema"]
			if !ok {
				return protobufSchema{}, unifiederrors.ValidationErrorf("protobuf %s field %q type message requires nested schema", path, name)
			}
			nested, err := parseProtobufSchema(nestedRaw, fmt.Sprintf("%s field %q schema", path, name))
			if err != nil {
				return protobufSchema{}, err
			}
			field.schema = &nested
		}

		if _, exists := schema.byNumber[number]; exists {
			return protobufSchema{}, unifiederrors.ValidationErrorf("protobuf %s has duplicate field number %d", path, number)
		}
		if _, exists := schema.byName[name]; exists {
			return protobufSchema{}, unifiederrors.ValidationErrorf("protobuf %s has duplicate field name %q", path, name)
		}

		schema.fields = append(schema.fields, field)
		schema.byNumber[number] = field
		schema.byName[name] = field
	}

	slices.SortFunc(schema.fields, func(a, b protobufField) int {
		if a.number < b.number {
			return -1
		}
		if a.number > b.number {
			return 1
		}
		return 0
	})

	return schema, nil
}

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

func skipUnknownProtobufField(content []byte, offset int, wireType int) (int, error) {
	switch wireType {
	case 0:
		_, n := binary.Uvarint(content[offset:])
		if n <= 0 {
			return offset, unifiederrors.ValidationErrorf("invalid protobuf varint at byte offset %d", offset)
		}
		return offset + n, nil
	case 1:
		if offset+8 > len(content) {
			return offset, unifiederrors.ValidationErrorf("invalid protobuf fixed64 at byte offset %d: truncated payload", offset)
		}
		return offset + 8, nil
	case 2:
		length, n := binary.Uvarint(content[offset:])
		if n <= 0 {
			return offset, unifiederrors.ValidationErrorf("invalid protobuf length at byte offset %d", offset)
		}
		offset += n
		if offset+int(length) > len(content) {
			return offset, unifiederrors.ValidationErrorf("invalid protobuf length-delimited field at byte offset %d: truncated payload", offset-n)
		}
		return offset + int(length), nil
	case 5:
		if offset+4 > len(content) {
			return offset, unifiederrors.ValidationErrorf("invalid protobuf fixed32 at byte offset %d: truncated payload", offset)
		}
		return offset + 4, nil
	default:
		return offset, unifiederrors.ValidationErrorf("unsupported protobuf wire type %d", wireType)
	}
}

func protobufWireTypeForField(kind string) (int, bool) {
	switch kind {
	case "int32", "int64", "uint32", "uint64", "sint32", "sint64", "bool", "enum":
		return 0, true
	case "fixed64", "sfixed64", "double":
		return 1, true
	case "string", "bytes", "message":
		return 2, true
	case "fixed32", "sfixed32", "float":
		return 5, true
	default:
		return 0, false
	}
}

func isSupportedProtobufFieldType(kind string) bool {
	_, ok := protobufWireTypeForField(kind)
	return ok
}

func isProtobufPackableField(kind string) bool {
	switch kind {
	case "int32", "int64", "uint32", "uint64", "sint32", "sint64", "bool", "enum", "fixed64", "sfixed64", "double", "fixed32", "sfixed32", "float":
		return true
	default:
		return false
	}
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

func appendVarint(out []byte, value uint64) []byte {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], value)
	return append(out, buf[:n]...)
}

func encodeZigZag32(v int32) uint32 {
	return uint32((v << 1) ^ (v >> 31))
}

func encodeZigZag64(v int64) uint64 {
	return uint64((v << 1) ^ (v >> 63))
}

func decodeZigZag32(v uint32) int32 {
	return int32((v >> 1) ^ uint32((int32(v&1)<<31)>>31))
}

func decodeZigZag64(v uint64) int64 {
	return int64((v >> 1) ^ uint64((int64(v&1)<<63)>>63))
}

func protobufPositiveInt(v interface{}) (int, error) {
	switch n := v.(type) {
	case int:
		if n <= 0 {
			return 0, fmt.Errorf("expected positive integer")
		}
		return n, nil
	case int32:
		if n <= 0 {
			return 0, fmt.Errorf("expected positive integer")
		}
		return int(n), nil
	case int64:
		if n <= 0 || n > math.MaxInt {
			return 0, fmt.Errorf("expected positive integer")
		}
		return int(n), nil
	case float64:
		if n <= 0 || math.Trunc(n) != n || n > math.MaxInt {
			return 0, fmt.Errorf("expected positive integer")
		}
		return int(n), nil
	default:
		return 0, fmt.Errorf("expected positive integer")
	}
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

func numberToFloat64(v interface{}) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case float32:
		return float64(n), nil
	case int:
		return float64(n), nil
	case int32:
		return float64(n), nil
	case int64:
		return float64(n), nil
	case uint:
		return float64(n), nil
	case uint32:
		return float64(n), nil
	case uint64:
		return float64(n), nil
	default:
		return 0, fmt.Errorf("expected number")
	}
}

func numberToInt64(v interface{}) (int64, error) {
	switch n := v.(type) {
	case int:
		return int64(n), nil
	case int32:
		return int64(n), nil
	case int64:
		return n, nil
	case uint:
		if uint64(n) > math.MaxInt64 {
			return 0, fmt.Errorf("overflow")
		}
		return int64(n), nil
	case uint32:
		return int64(n), nil
	case uint64:
		if n > math.MaxInt64 {
			return 0, fmt.Errorf("overflow")
		}
		return int64(n), nil
	case float64:
		if math.Trunc(n) != n || n < math.MinInt64 || n > math.MaxInt64 {
			return 0, fmt.Errorf("not an int64")
		}
		return int64(n), nil
	default:
		return 0, fmt.Errorf("expected integer")
	}
}

func numberToUint64(v interface{}) (uint64, error) {
	switch n := v.(type) {
	case int:
		if n < 0 {
			return 0, fmt.Errorf("negative value")
		}
		return uint64(n), nil
	case int32:
		if n < 0 {
			return 0, fmt.Errorf("negative value")
		}
		return uint64(n), nil
	case int64:
		if n < 0 {
			return 0, fmt.Errorf("negative value")
		}
		return uint64(n), nil
	case uint:
		return uint64(n), nil
	case uint32:
		return uint64(n), nil
	case uint64:
		return n, nil
	case float64:
		if math.Trunc(n) != n || n < 0 || n > math.MaxUint64 {
			return 0, fmt.Errorf("not a uint64")
		}
		return uint64(n), nil
	default:
		return 0, fmt.Errorf("expected non-negative integer")
	}
}
