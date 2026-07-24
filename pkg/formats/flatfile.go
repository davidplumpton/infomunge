package formats

import (
	"fmt"
	unifiederrors "infomunge/internal/errors"
	"infomunge/pkg/values"
	"math"
	"strconv"
	"strings"
)

type flatfileSchema struct {
	fields       []flatfileField
	totalLength  int
	singleRecord bool
}

type flatfileField struct {
	name   string
	length int
	align  string
	pad    rune
	trim   bool
	kind   string
}

func init() {
	RegisterReader("application/flatfile", readFlatfile)
	RegisterWriter("application/flatfile", formatFlatfile)
	RegisterExtension(".flatfile", "application/flatfile")
	RegisterExtension(".ffd", "application/flatfile")
	RegisterReadOptionsHandler("application/flatfile", readFlatfileWithOptions)
	RegisterWriteOptionsHandler("application/flatfile", formatFlatfileWithOptions)
}

func readFlatfile(content string) (interface{}, error) {
	return readBinary(content)
}

func formatFlatfile(result interface{}) (string, error) {
	return formatBinary(result)
}

func readFlatfileWithOptions(content string, options Object) (interface{}, error) {
	schema, err := parseFlatfileSchema(options)
	if err != nil {
		return nil, err
	}

	lines := splitFlatfileLines(content)
	records := make(Array, 0, len(lines))
	for i, line := range lines {
		if len(line) != schema.totalLength {
			return nil, unifiederrors.ValidationErrorf(
				"flatfile record %d has length %d, expected %d",
				i+1,
				len(line),
				schema.totalLength,
			)
		}

		record := values.NewObject(len(schema.fields))
		offset := 0
		for _, field := range schema.fields {
			segment := line[offset : offset+field.length]
			offset += field.length

			valueText := decodeFlatfileSegment(segment, field)
			value, err := coerceFlatfileValue(valueText, field.kind)
			if err != nil {
				return nil, unifiederrors.ValidationErrorf(
					"flatfile record %d field %q: %v",
					i+1,
					field.name,
					err,
				)
			}
			values.SetObjectValue(record, field.name, value)
		}
		records = append(records, record)
	}

	if schema.singleRecord {
		if len(records) != 1 {
			return nil, unifiederrors.ValidationErrorf("flatfile expected exactly 1 record, got %d", len(records))
		}
		return records[0], nil
	}

	return records, nil
}

func formatFlatfileWithOptions(result interface{}, options Object) (string, error) {
	schema, err := parseFlatfileSchema(options)
	if err != nil {
		return "", err
	}

	records, err := normalizeFlatfileRecords(result)
	if err != nil {
		return "", err
	}

	if schema.singleRecord && len(records) != 1 {
		return "", unifiederrors.ValidationErrorf("flatfile schema expects a single record, got %d", len(records))
	}

	lines := make([]string, 0, len(records))
	for i, record := range records {
		var b strings.Builder
		for _, field := range schema.fields {
			formatted, err := stringifyFlatfileValue(record[field.name], field.kind)
			if err != nil {
				return "", unifiederrors.ValidationErrorf(
					"flatfile record %d field %q: %v",
					i+1,
					field.name,
					err,
				)
			}
			if len(formatted) > field.length {
				return "", unifiederrors.ValidationErrorf(
					"flatfile record %d field %q value length %d exceeds field length %d",
					i+1,
					field.name,
					len(formatted),
					field.length,
				)
			}
			b.WriteString(padFlatfileField(formatted, field))
		}
		lines = append(lines, b.String())
	}

	return strings.Join(lines, "\n"), nil
}

func normalizeFlatfileRecords(result interface{}) ([]Object, error) {
	switch v := result.(type) {
	case Object:
		return []Object{v}, nil
	case Array:
		out := make([]Object, 0, len(v))
		for i, item := range v {
			obj, ok := item.(map[string]interface{})
			if !ok {
				return nil, unifiederrors.ValidationErrorf(
					"flatfile output expects object records, item %d is %T",
					i+1,
					item,
				)
			}
			out = append(out, obj)
		}
		return out, nil
	default:
		return nil, unifiederrors.ValidationErrorf("flatfile output expects object or array of objects, got %T", result)
	}
}

func parseFlatfileSchema(options Object) (flatfileSchema, error) {
	if len(options) == 0 {
		return flatfileSchema{}, unifiederrors.ValidationError("flatfile options must include a non-empty schema.fields array")
	}

	schemaSource := options
	if rawSchema, ok := options["schema"]; ok {
		schemaObj, ok := rawSchema.(map[string]interface{})
		if !ok {
			return flatfileSchema{}, unifiederrors.ValidationErrorf("flatfile options.schema must be an object, got %T", rawSchema)
		}
		schemaSource = schemaObj
	}

	fieldsRaw, ok := schemaSource["fields"]
	if !ok {
		return flatfileSchema{}, unifiederrors.ValidationError("flatfile schema must include fields")
	}

	fieldsSlice, ok := fieldsRaw.([]interface{})
	if !ok || len(fieldsSlice) == 0 {
		return flatfileSchema{}, unifiederrors.ValidationError("flatfile schema fields must be a non-empty array")
	}

	schema := flatfileSchema{
		fields:       make([]flatfileField, 0, len(fieldsSlice)),
		singleRecord: flatfileBool(schemaSource["singleRecord"], false),
	}

	for i, rawField := range fieldsSlice {
		fieldObj, ok := rawField.(map[string]interface{})
		if !ok {
			return flatfileSchema{}, unifiederrors.ValidationErrorf("flatfile schema field %d must be an object, got %T", i+1, rawField)
		}

		name, ok := fieldObj["name"].(string)
		if !ok || strings.TrimSpace(name) == "" {
			return flatfileSchema{}, unifiederrors.ValidationErrorf("flatfile schema field %d must have a non-empty name", i+1)
		}

		length, err := flatfileInt(fieldObj["length"])
		if err != nil || length <= 0 {
			return flatfileSchema{}, unifiederrors.ValidationErrorf("flatfile schema field %q must have a positive integer length", name)
		}

		align := strings.ToLower(strings.TrimSpace(flatfileString(fieldObj["align"], "left")))
		if align != "left" && align != "right" {
			return flatfileSchema{}, unifiederrors.ValidationErrorf("flatfile schema field %q align must be left or right", name)
		}

		pad := ' '
		if rawPad, ok := fieldObj["pad"]; ok {
			padStr, ok := rawPad.(string)
			if !ok || len(padStr) != 1 {
				return flatfileSchema{}, unifiederrors.ValidationErrorf("flatfile schema field %q pad must be a single character string", name)
			}
			pad = rune(padStr[0])
		}

		kind := strings.ToLower(strings.TrimSpace(flatfileString(fieldObj["type"], "string")))
		if kind == "" {
			kind = "string"
		}
		switch kind {
		case "string", "number", "int", "integer", "bool", "boolean":
		default:
			return flatfileSchema{}, unifiederrors.ValidationErrorf("flatfile schema field %q has unsupported type %q", name, kind)
		}

		field := flatfileField{
			name:   name,
			length: length,
			align:  align,
			pad:    pad,
			trim:   flatfileBool(fieldObj["trim"], true),
			kind:   kind,
		}

		schema.fields = append(schema.fields, field)
		schema.totalLength += length
	}

	return schema, nil
}

func splitFlatfileLines(content string) []string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func decodeFlatfileSegment(segment string, field flatfileField) string {
	if !field.trim {
		return segment
	}
	switch field.align {
	case "right":
		trimmed := strings.TrimLeft(segment, string(field.pad))
		return strings.TrimSpace(trimmed)
	default:
		trimmed := strings.TrimRight(segment, string(field.pad))
		return strings.TrimSpace(trimmed)
	}
}

func padFlatfileField(value string, field flatfileField) string {
	if len(value) >= field.length {
		return value
	}
	padding := strings.Repeat(string(field.pad), field.length-len(value))
	if field.align == "right" {
		return padding + value
	}
	return value + padding
}

func coerceFlatfileValue(raw string, kind string) (interface{}, error) {
	switch kind {
	case "int", "integer":
		if raw == "" {
			return float64(0), nil
		}
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("expected integer, got %q", raw)
		}
		return float64(n), nil
	case "number":
		if raw == "" {
			return float64(0), nil
		}
		n, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("expected number, got %q", raw)
		}
		return n, nil
	case "bool", "boolean":
		if raw == "" {
			return false, nil
		}
		switch strings.ToLower(raw) {
		case "true", "1", "y", "yes":
			return true, nil
		case "false", "0", "n", "no":
			return false, nil
		default:
			return nil, fmt.Errorf("expected boolean, got %q", raw)
		}
	default:
		return raw, nil
	}
}

func stringifyFlatfileValue(value interface{}, kind string) (string, error) {
	if value == nil {
		return "", nil
	}
	switch kind {
	case "int", "integer":
		n, err := flatfileNumeric(value)
		if err != nil {
			return "", err
		}
		if math.Trunc(n) != n {
			return "", fmt.Errorf("expected integer-compatible value, got %v", value)
		}
		return strconv.FormatInt(int64(n), 10), nil
	case "number":
		n, err := flatfileNumeric(value)
		if err != nil {
			return "", err
		}
		return strconv.FormatFloat(n, 'f', -1, 64), nil
	case "bool", "boolean":
		switch v := value.(type) {
		case bool:
			if v {
				return "true", nil
			}
			return "false", nil
		case string:
			lower := strings.ToLower(strings.TrimSpace(v))
			switch lower {
			case "true", "false":
				return lower, nil
			}
		}
		return "", fmt.Errorf("expected boolean-compatible value, got %T", value)
	default:
		return fmt.Sprint(value), nil
	}
}

func flatfileNumeric(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		n, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return 0, fmt.Errorf("expected numeric-compatible value, got %q", v)
		}
		return n, nil
	default:
		return 0, fmt.Errorf("expected numeric-compatible value, got %T", value)
	}
}

func flatfileBool(raw interface{}, defaultValue bool) bool {
	if raw == nil {
		return defaultValue
	}
	b, ok := raw.(bool)
	if !ok {
		return defaultValue
	}
	return b
}

func flatfileString(raw interface{}, defaultValue string) string {
	if raw == nil {
		return defaultValue
	}
	s, ok := raw.(string)
	if !ok {
		return defaultValue
	}
	return s
}

func flatfileInt(raw interface{}) (int, error) {
	switch v := raw.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		if math.Trunc(v) != v {
			return 0, fmt.Errorf("not an integer")
		}
		return int(v), nil
	case float32:
		if math.Trunc(float64(v)) != float64(v) {
			return 0, fmt.Errorf("not an integer")
		}
		return int(v), nil
	default:
		return 0, fmt.Errorf("unsupported numeric type %T", raw)
	}
}
