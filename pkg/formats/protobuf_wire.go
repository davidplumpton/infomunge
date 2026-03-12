package formats

import (
	"encoding/binary"
	"fmt"
	unifiederrors "infomunge/internal/errors"
	"math"
	"strconv"
	"strings"
)

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

func isProtobufPackableField(kind string) bool {
	switch kind {
	case "int32", "int64", "uint32", "uint64", "sint32", "sint64", "bool", "enum", "fixed64", "sfixed64", "double", "fixed32", "sfixed32", "float":
		return true
	default:
		return false
	}
}

func protobufMapKeyKindAllowed(kind string) bool {
	switch kind {
	case "bool", "int32", "int64", "uint32", "uint64", "sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64", "string":
		return true
	default:
		return false
	}
}

func protobufDefaultMapKey(kind string) string {
	switch kind {
	case "bool":
		return "false"
	case "string":
		return ""
	default:
		return "0"
	}
}

func protobufDefaultValueForKind(field protobufField) interface{} {
	switch field.kind {
	case "bool":
		return false
	case "string", "bytes":
		return ""
	case "message":
		return nil
	case "float", "double", "int32", "int64", "uint32", "uint64", "sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64", "enum":
		return float64(0)
	default:
		return nil
	}
}

func protobufMapKeyToString(kind string, value interface{}) (string, error) {
	switch kind {
	case "string":
		s, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("expected string key, got %T", value)
		}
		return s, nil
	case "bool":
		b, ok := value.(bool)
		if !ok {
			return "", fmt.Errorf("expected bool key, got %T", value)
		}
		if b {
			return "true", nil
		}
		return "false", nil
	default:
		n, err := numberToInt64(value)
		if err != nil {
			u, uErr := numberToUint64(value)
			if uErr != nil {
				return "", fmt.Errorf("expected numeric key, got %T", value)
			}
			return strconv.FormatUint(u, 10), nil
		}
		return strconv.FormatInt(n, 10), nil
	}
}

func protobufParseMapKey(kind string, key string) (interface{}, error) {
	switch kind {
	case "string":
		return key, nil
	case "bool":
		switch strings.TrimSpace(strings.ToLower(key)) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return nil, fmt.Errorf("expected bool key")
		}
	case "int32", "int64", "sint32", "sint64", "sfixed32", "sfixed64":
		n, err := strconv.ParseInt(strings.TrimSpace(key), 10, 64)
		if err != nil {
			return nil, err
		}
		return float64(n), nil
	case "uint32", "uint64", "fixed32", "fixed64":
		n, err := strconv.ParseUint(strings.TrimSpace(key), 10, 64)
		if err != nil {
			return nil, err
		}
		return float64(n), nil
	default:
		return nil, fmt.Errorf("unsupported map key kind %q", kind)
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
