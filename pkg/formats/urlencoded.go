package formats

import (
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"

	unifiederrors "infomunge/internal/errors"
)

func init() {
	RegisterReader("application/x-www-form-urlencoded", readURLEncoded)
	RegisterWriter("application/x-www-form-urlencoded", formatURLEncoded)
	RegisterExtension(".urlencoded", "application/x-www-form-urlencoded")
}

func readURLEncoded(content string) (interface{}, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return make(Object), nil
	}
	values, err := url.ParseQuery(content)
	if err != nil {
		return nil, unifiederrors.WrapValidationf(err, "urlencoded parse error: %v", err)
	}
	result := make(Object)
	for key, vals := range values {
		if len(vals) == 1 {
			result[key] = vals[0]
		} else {
			arr := make(Array, len(vals))
			for i, v := range vals {
				arr[i] = v
			}
			result[key] = arr
		}
	}
	return result, nil
}

func formatURLEncoded(result interface{}) (string, error) {
	obj, ok := result.(Object)
	if !ok {
		return "", unifiederrors.ValidationErrorf("urlencoded output expects an object, got %T", result)
	}
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v := obj[k]
		switch val := v.(type) {
		case Array:
			for _, item := range val {
				parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(formatURLValue(item)))
			}
		default:
			parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(formatURLValue(v)))
		}
	}
	return strings.Join(parts, "&"), nil
}

func formatURLValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return fmt.Sprintf("%d", val)
	case float64:
		if val == math.Trunc(val) && !math.IsInf(val, 0) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}
