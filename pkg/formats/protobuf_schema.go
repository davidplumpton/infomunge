package formats

import (
	"fmt"
	unifiederrors "infomunge/internal/errors"
	"slices"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

type protobufSchema struct {
	messageName string
	fields      []protobufField
	byNumber    map[int]protobufField
	byName      map[string]protobufField
}

type protobufField struct {
	number         int
	name           string
	kind           string
	repeated       bool
	packed         bool
	schema         *protobufSchema
	mapKeyKind     string
	mapValueKind   string
	mapValueSchema *protobufSchema
}

// parseProtobufSchemaFromDescriptor builds a schema from a serialized FileDescriptorSet.
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
			mapKeyKind, mapValueKind, mapValueSchema, err := protobufMapKindsFromDescriptorField(descriptorField)
			if err != nil {
				return protobufSchema{}, err
			}
			field := protobufField{
				number:         int(descriptorField.Number()),
				name:           string(descriptorField.Name()),
				kind:           "map",
				mapKeyKind:     mapKeyKind,
				mapValueKind:   mapValueKind,
				mapValueSchema: mapValueSchema,
			}
			schema.fields = append(schema.fields, field)
			schema.byNumber[field.number] = field
			schema.byName[field.name] = field
			continue
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

func protobufMapKindsFromDescriptorField(descriptorField protoreflect.FieldDescriptor) (string, string, *protobufSchema, error) {
	entry := descriptorField.Message()
	if entry == nil {
		return "", "", nil, unifiederrors.ValidationErrorf("protobuf descriptor field %q map entry descriptor is missing", descriptorField.FullName())
	}

	keyDescriptor := entry.Fields().ByName("key")
	valueDescriptor := entry.Fields().ByName("value")
	if keyDescriptor == nil || valueDescriptor == nil {
		return "", "", nil, unifiederrors.ValidationErrorf("protobuf descriptor field %q map entry must define key and value fields", descriptorField.FullName())
	}

	keyKind, ok := protobufKindFromDescriptorKind(keyDescriptor.Kind())
	if !ok {
		return "", "", nil, unifiederrors.ValidationErrorf("protobuf descriptor field %q map key has unsupported kind %q", descriptorField.FullName(), keyDescriptor.Kind().String())
	}
	if !protobufMapKeyKindAllowed(keyKind) {
		return "", "", nil, unifiederrors.ValidationErrorf("protobuf descriptor field %q map key kind %q is not supported", descriptorField.FullName(), keyKind)
	}

	valueKind, ok := protobufKindFromDescriptorKind(valueDescriptor.Kind())
	if !ok {
		return "", "", nil, unifiederrors.ValidationErrorf("protobuf descriptor field %q map value has unsupported kind %q", descriptorField.FullName(), valueDescriptor.Kind().String())
	}
	var valueSchema *protobufSchema
	if valueKind == "message" {
		nested, err := protobufSchemaFromMessageDescriptor(valueDescriptor.Message())
		if err != nil {
			return "", "", nil, err
		}
		valueSchema = &nested
	}

	return keyKind, valueKind, valueSchema, nil
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

// parseProtobufSchema builds a schema from an inline literal object.
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
