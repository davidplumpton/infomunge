package evaluator

import (
	"fmt"
	"go/token"
	"sort"
	"strings"
)

// evalArrayIndex handles indexing into arrays (and XMLMultiValue).
func evalArrayIndex(arr []interface{}, idx Value, pos token.Pos) (Value, error) {
	switch i := idx.(type) {
	case int:
		if i < 0 {
			i += len(arr)
		}
		if i < 0 || i >= len(arr) {
			return nil, newPosError(fmt.Sprintf("array index out of bounds: %d", i), pos)
		}
		return arr[i], nil
	case string:
		return evalArrayStringIndex(arr, i, pos)
	default:
		return nil, newPosError(fmt.Sprintf("array index must be an integer or string, got %T", idx), pos)
	}
}

// evalArrayStringIndex handles string-based indexing into arrays, including
// presence (?) and assert (!) selectors and field extraction.
func evalArrayStringIndex(arr []interface{}, key string, pos token.Pos) (Value, error) {
	if strings.HasSuffix(key, "?") {
		return evalArrayPresenceSelector(arr, strings.TrimSuffix(key, "?")), nil
	}
	if strings.HasSuffix(key, "!") {
		return evalArrayAssertSelector(arr, strings.TrimSuffix(key, "!"), pos)
	}
	// DataWeave compatibility: applying a string index to an array extracts that field from all elements
	result := make([]interface{}, 0, len(arr))
	for _, item := range arr {
		if itemMap, ok := item.(map[string]interface{}); ok {
			val, _ := getObjectValue(itemMap, key)
			result = append(result, val)
		} else {
			result = append(result, nil)
		}
	}
	return result, nil
}

// evalArrayPresenceSelector returns a boolean for each array item indicating
// whether the given key exists.
func evalArrayPresenceSelector(arr []interface{}, key string) []interface{} {
	result := make([]interface{}, 0, len(arr))
	for _, item := range arr {
		if itemMap, ok := item.(map[string]interface{}); ok {
			_, exists := getObjectValue(itemMap, key)
			result = append(result, exists)
		} else {
			result = append(result, false)
		}
	}
	return result
}

// evalArrayAssertSelector extracts the given key from each array item,
// returning an error if any item is not an object or is missing the key.
func evalArrayAssertSelector(arr []interface{}, key string, pos token.Pos) ([]interface{}, error) {
	result := make([]interface{}, 0, len(arr))
	for idx, item := range arr {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, newPosError(fmt.Sprintf("assert selector failed: expected object at index %d", idx), pos)
		}
		val, exists := getObjectValue(itemMap, key)
		if !exists {
			return nil, newPosError(fmt.Sprintf("assert selector failed: missing key %q at index %d", key, idx), pos)
		}
		result = append(result, val)
	}
	return result, nil
}

// evalStringIndex handles indexing into strings.
func evalStringIndex(s string, idx Value, pos token.Pos) (Value, error) {
	switch i := idx.(type) {
	case int:
		if i < 0 {
			i += len(s)
		}
		if i < 0 || i >= len(s) {
			return nil, newPosError(fmt.Sprintf("string index out of bounds: %d", i), pos)
		}
		return s[i : i+1], nil
	default:
		return nil, newPosError(fmt.Sprintf("string index must be an integer, got %T", idx), pos)
	}
}

// evalObjectIndex handles indexing into maps/objects, including special keys
// (#, @), presence/assert selectors, and ordinal indexing.
func evalObjectIndex(obj map[string]interface{}, idx Value, pos token.Pos) (Value, error) {
	switch i := idx.(type) {
	case string:
		return evalObjectStringIndex(obj, i, pos)
	case int:
		return evalObjectOrdinalIndex(obj, i, pos)
	default:
		return nil, newPosError(fmt.Sprintf("map key must be a string or int, got %T", idx), pos)
	}
}

// evalObjectStringIndex handles string-based object access including special
// keys (#, @) and presence/assert selectors.
func evalObjectStringIndex(obj map[string]interface{}, key string, pos token.Pos) (Value, error) {
	if key == "#" {
		return extractNamespaceValue(obj), nil
	}
	if key == "@" {
		attrs := make(map[string]interface{})
		for k, val := range obj {
			if strings.HasPrefix(k, "@") {
				attrs[k] = val
			}
		}
		return attrs, nil
	}
	if strings.HasSuffix(key, "?") {
		trimmed := strings.TrimSuffix(key, "?")
		_, exists := getObjectValue(obj, trimmed)
		return exists, nil
	}
	if strings.HasSuffix(key, "!") {
		trimmed := strings.TrimSuffix(key, "!")
		val, exists := getObjectValue(obj, trimmed)
		if !exists {
			return nil, newPosError(fmt.Sprintf("assert selector failed: missing key %q", trimmed), pos)
		}
		return val, nil
	}
	val, _ := getObjectValue(obj, key)
	return val, nil
}

// evalObjectOrdinalIndex accesses an object by sorted key position.
func evalObjectOrdinalIndex(obj map[string]interface{}, i int, pos token.Pos) (Value, error) {
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if i < 0 || i >= len(keys) {
		return nil, newPosError(fmt.Sprintf("object index out of bounds: %d (object has %d keys)", i, len(keys)), pos)
	}
	return obj[keys[i]], nil
}
