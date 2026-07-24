package formats

import (
	"bytes"
	"encoding/json"
	"fmt"
	unifiederrors "infomunge/internal/errors"
	"infomunge/pkg/values"
	"io"
	"strings"
)

func init() {
	RegisterReader("application/json", readJSON)
	RegisterWriter("application/json", formatJSON)
	RegisterExtension(".json", "application/json")
}

func readJSON(content string) (interface{}, error) {
	decoder := json.NewDecoder(strings.NewReader(content))
	result, err := decodeJSONValue(decoder)
	if err != nil {
		return nil, unifiederrors.WrapValidationf(err, "JSON parse error: %v", err)
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("unexpected trailing JSON content")
		}
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
		encoded, err := marshalOrderedJSON(v)
		if err != nil {
			return "", err
		}
		return string(encoded), nil
	}
}

func decodeJSONValue(decoder *json.Decoder) (interface{}, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	delim, isDelim := token.(json.Delim)
	if !isDelim {
		return token, nil
	}
	switch delim {
	case '{':
		object := values.NewObject(0)
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			key := keyToken.(string)
			value, err := decodeJSONValue(decoder)
			if err != nil {
				return nil, err
			}
			values.SetObjectValue(object, key, value)
		}
		_, err := decoder.Token()
		return object, err
	case '[':
		array := make(Array, 0)
		for decoder.More() {
			value, err := decodeJSONValue(decoder)
			if err != nil {
				return nil, err
			}
			array = append(array, value)
		}
		_, err := decoder.Token()
		return array, err
	default:
		return nil, fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
}

func marshalOrderedJSON(value interface{}) ([]byte, error) {
	switch typed := value.(type) {
	case Object:
		var buf bytes.Buffer
		buf.WriteByte('{')
		for index, key := range values.ObjectKeys(typed) {
			if index > 0 {
				buf.WriteByte(',')
			}
			encodedKey, _ := json.Marshal(key)
			encodedValue, err := marshalOrderedJSON(typed[key])
			if err != nil {
				return nil, err
			}
			buf.Write(encodedKey)
			buf.WriteByte(':')
			buf.Write(encodedValue)
		}
		buf.WriteByte('}')
		return buf.Bytes(), nil
	case Array:
		if typed == nil {
			return []byte("null"), nil
		}
		var buf bytes.Buffer
		buf.WriteByte('[')
		for index, item := range typed {
			if index > 0 {
				buf.WriteByte(',')
			}
			encodedItem, err := marshalOrderedJSON(item)
			if err != nil {
				return nil, err
			}
			buf.Write(encodedItem)
		}
		buf.WriteByte(']')
		return buf.Bytes(), nil
	default:
		return json.Marshal(value)
	}
}
