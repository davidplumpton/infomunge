package evaluator

import (
	"fmt"
	"go/token"
	"infomunge/pkg/values"
	"strings"
)

// evalArrayIndex handles indexing into arrays (and XMLMultiValue).
func evalArrayIndex(arr Array, idx Value, pos token.Pos) (Value, error) {
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
// presence (?) and assert (!) selectors and field extraction. DataWeave array
// selectors collect matching keys from immediate object elements, skip
// non-matching elements, and return null when no key matches.
func evalArrayStringIndex(arr Array, key string, pos token.Pos) (Value, error) {
	if strings.HasSuffix(key, "?") {
		return evalArrayPresenceSelector(arr, strings.TrimSuffix(key, "?")), nil
	}
	if strings.HasSuffix(key, "!") {
		return evalArrayAssertSelector(arr, strings.TrimSuffix(key, "!"), pos)
	}
	result := make(Array, 0, len(arr))
	for _, item := range arr {
		if itemMap, ok := item.(Object); ok {
			if val, exists := getObjectValue(itemMap, key); exists {
				result = append(result, val)
			}
		}
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}

// evalArrayPresenceSelector reports whether any immediate object element has
// the given key.
func evalArrayPresenceSelector(arr Array, key string) bool {
	for _, item := range arr {
		if itemMap, ok := item.(Object); ok {
			_, exists := getObjectValue(itemMap, key)
			if exists {
				return true
			}
		}
	}
	return false
}

// evalArrayAssertSelector extracts the given key from each array item,
// returning an error when no immediate object element has the key.
func evalArrayAssertSelector(arr Array, key string, pos token.Pos) (Value, error) {
	result := make(Array, 0, len(arr))
	for _, item := range arr {
		itemMap, ok := item.(Object)
		if !ok {
			continue
		}
		val, exists := getObjectValue(itemMap, key)
		if exists {
			result = append(result, val)
		}
	}
	if len(result) == 0 {
		return nil, newPosError(fmt.Sprintf("assert selector failed: missing key %q", key), pos)
	}
	return result, nil
}

// evalStringIndex handles indexing into strings.
func evalStringIndex(s string, idx Value, pos token.Pos) (Value, error) {
	switch i := idx.(type) {
	case int:
		runes := []rune(s)
		if i < 0 {
			i += len(runes)
		}
		if i < 0 || i >= len(runes) {
			return nil, newPosError(fmt.Sprintf("string index out of bounds: %d", i), pos)
		}
		return string(runes[i]), nil
	default:
		return nil, newPosError(fmt.Sprintf("string index must be an integer, got %T", idx), pos)
	}
}

// evalObjectIndex handles indexing into maps/objects, including special keys
// (#, @), presence/assert selectors, and ordinal indexing.
func evalObjectIndex(obj Object, idx Value, pos token.Pos) (Value, error) {
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
func evalObjectStringIndex(obj Object, key string, pos token.Pos) (Value, error) {
	if key == "#" {
		return extractNamespaceValue(obj), nil
	}
	if key == "@" {
		attrs := values.NewObject(0)
		for _, k := range values.ObjectKeys(obj) {
			if strings.HasPrefix(k, "@") {
				values.SetObjectValue(attrs, k, obj[k])
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

// evalObjectOrdinalIndex accesses an object by insertion position.
func evalObjectOrdinalIndex(obj Object, i int, pos token.Pos) (Value, error) {
	keys := values.ObjectKeys(obj)
	ordinal := i
	if ordinal < 0 {
		ordinal += len(keys)
	}
	if ordinal < 0 || ordinal >= len(keys) {
		return nil, newPosError(fmt.Sprintf("object index out of bounds: %d (object has %d keys)", i, len(keys)), pos)
	}
	return obj[keys[ordinal]], nil
}
