package formats

import (
	unifiederrors "infomunge/internal/errors"
	"infomunge/pkg/values"
	"math"
	"strings"
)

type javaFormatOptions struct {
	structured bool
	strict     bool
	className  string
}

func init() {
	RegisterReader("application/java", readJava)
	RegisterWriter("application/java", formatJava)
	RegisterExtension(".java", "application/java")
	RegisterExtension(".ser", "application/java")
	RegisterReadOptionsHandler("application/java", readJavaWithOptions)
	RegisterWriteOptionsHandler("application/java", formatJavaWithOptions)
}

func readJava(content string) (interface{}, error) {
	return readBinary(content)
}

func formatJava(result interface{}) (string, error) {
	return formatBinary(result)
}

func readJavaWithOptions(content string, options Object) (interface{}, error) {
	parsedOptions, err := parseJavaReadOptions(options)
	if err != nil {
		return nil, err
	}
	if !parsedOptions.structured {
		return readJava(content)
	}

	parsed, err := readJSON(content)
	if err != nil {
		return nil, err
	}

	decoded, err := decodeJavaStructured(parsed, parsedOptions.strict)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

func formatJavaWithOptions(result interface{}, options Object) (string, error) {
	parsedOptions, err := parseJavaWriteOptions(options)
	if err != nil {
		return "", err
	}
	if !parsedOptions.structured {
		return formatJava(result)
	}

	className := parsedOptions.className
	if className == "" {
		inferredClass, err := inferJavaClass(result)
		if err != nil {
			return "", err
		}
		className = inferredClass
	}

	if err := validateJavaClassValue(className, result, parsedOptions.strict); err != nil {
		return "", err
	}

	encoded, err := formatJSON(Object{
		"@class": className,
		"value":  result,
	})
	if err != nil {
		return "", err
	}
	return encoded, nil
}

func parseJavaReadOptions(options Object) (javaFormatOptions, error) {
	return parseJavaOptions(options, false)
}

func parseJavaWriteOptions(options Object) (javaFormatOptions, error) {
	return parseJavaOptions(options, true)
}

func parseJavaOptions(options Object, allowClass bool) (javaFormatOptions, error) {
	parsed := javaFormatOptions{
		structured: false,
		strict:     true,
	}
	if options == nil {
		return parsed, nil
	}

	for key, raw := range options {
		switch key {
		case "structured":
			v, ok := raw.(bool)
			if !ok {
				return javaFormatOptions{}, unifiederrors.ValidationErrorf("java option structured must be a boolean, got %T", raw)
			}
			parsed.structured = v
		case "strict":
			v, ok := raw.(bool)
			if !ok {
				return javaFormatOptions{}, unifiederrors.ValidationErrorf("java option strict must be a boolean, got %T", raw)
			}
			parsed.strict = v
		case "class":
			if !allowClass {
				return javaFormatOptions{}, unifiederrors.ValidationError("java option class is only supported for write")
			}
			v, ok := raw.(string)
			if !ok || strings.TrimSpace(v) == "" {
				return javaFormatOptions{}, unifiederrors.ValidationErrorf("java option class must be a non-empty string, got %T", raw)
			}
			parsed.className = strings.TrimSpace(v)
		default:
			return javaFormatOptions{}, unifiederrors.ValidationErrorf("unsupported java option: %s", key)
		}
	}

	if parsed.className != "" && !parsed.structured {
		return javaFormatOptions{}, unifiederrors.ValidationError("java option class requires structured=true")
	}
	return parsed, nil
}

func decodeJavaStructured(value interface{}, strict bool) (interface{}, error) {
	switch v := value.(type) {
	case map[string]interface{}:
		if classRaw, hasClass := v["@class"]; hasClass {
			className, ok := classRaw.(string)
			if !ok || strings.TrimSpace(className) == "" {
				return nil, unifiederrors.ValidationErrorf("java envelope @class must be a non-empty string, got %T", classRaw)
			}
			rawValue, hasValue := v["value"]
			if !hasValue {
				return nil, unifiederrors.ValidationError("java envelope must include value")
			}
			if err := validateJavaClassValue(className, rawValue, strict); err != nil {
				return nil, err
			}
			return decodeJavaStructured(rawValue, strict)
		}

		out := values.NewObject(len(v))
		for _, key := range values.ObjectKeys(v) {
			item := v[key]
			decoded, err := decodeJavaStructured(item, strict)
			if err != nil {
				return nil, err
			}
			values.SetObjectValue(out, key, decoded)
		}
		return out, nil
	case []interface{}:
		out := make(Array, 0, len(v))
		for _, item := range v {
			decoded, err := decodeJavaStructured(item, strict)
			if err != nil {
				return nil, err
			}
			out = append(out, decoded)
		}
		return out, nil
	default:
		return value, nil
	}
}

func inferJavaClass(value interface{}) (string, error) {
	switch value.(type) {
	case map[string]interface{}:
		return "java.util.LinkedHashMap", nil
	case []interface{}:
		return "java.util.ArrayList", nil
	case string:
		return "java.lang.String", nil
	case bool:
		return "java.lang.Boolean", nil
	case int, int8, int16, int32, int64:
		return "java.lang.Long", nil
	case uint, uint8, uint16, uint32, uint64:
		return "java.lang.Long", nil
	case float32, float64:
		return "java.lang.Double", nil
	case nil:
		return "java.lang.Object", nil
	default:
		return "", unifiederrors.ValidationErrorf("java structured output does not support value type %T", value)
	}
}

func validateJavaClassValue(className string, value interface{}, strict bool) error {
	category, known := javaClassCategory(className)
	if !known {
		if strict {
			return unifiederrors.ValidationErrorf("unsupported java class %q", className)
		}
		return nil
	}

	if !matchesJavaClassCategory(category, value) {
		return unifiederrors.ValidationErrorf(
			"java class %q is incompatible with value type %T",
			className,
			value,
		)
	}
	return nil
}

func javaClassCategory(className string) (string, bool) {
	switch className {
	case "java.util.Map", "java.util.HashMap", "java.util.LinkedHashMap", "java.util.TreeMap":
		return "map", true
	case "java.util.List", "java.util.ArrayList", "java.util.LinkedList", "java.util.Vector":
		return "list", true
	case "java.lang.String", "java.lang.CharSequence", "java.lang.Character":
		return "string", true
	case "java.lang.Boolean":
		return "bool", true
	case "java.lang.Byte", "java.lang.Short", "java.lang.Integer", "java.lang.Long", "java.math.BigInteger":
		return "integral", true
	case "java.lang.Float", "java.lang.Double", "java.lang.Number", "java.math.BigDecimal":
		return "number", true
	case "java.lang.Object":
		return "object", true
	default:
		return "", false
	}
}

func matchesJavaClassCategory(category string, value interface{}) bool {
	switch category {
	case "map":
		_, ok := value.(map[string]interface{})
		return ok
	case "list":
		_, ok := value.([]interface{})
		return ok
	case "string":
		if _, ok := value.(string); !ok {
			return false
		}
		return true
	case "bool":
		_, ok := value.(bool)
		return ok
	case "integral":
		return isIntegralValue(value)
	case "number":
		return isNumericValue(value)
	case "object":
		return true
	default:
		return false
	}
}

func isIntegralValue(value interface{}) bool {
	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32:
		return math.Trunc(float64(v)) == float64(v)
	case float64:
		return math.Trunc(v) == v
	default:
		return false
	}
}

func isNumericValue(value interface{}) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return true
	default:
		return false
	}
}
