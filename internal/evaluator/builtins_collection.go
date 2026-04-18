package evaluator

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
	"strings"
)

// validateArgCount checks that args has the expected count and returns an error if not.
func validateArgCount(args []interface{}, expected int, funcName string, pos token.Pos) error {
	if len(args) != expected {
		return newArgCountError(funcName, expected, pos)
	}
	return nil
}

// callBuiltinSizeOf implements the sizeOf(value) function.
func callBuiltinSizeOf(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := validateArgCount(args, 1, "sizeOf", e.Pos()); err != nil {
		return nil, err
	}
	switch v := args[0].(type) {
	case string:
		return len(v), nil
	case []interface{}:
		return len(v), nil
	case XMLMultiValue:
		return len(v), nil
	case map[string]interface{}:
		return len(v), nil
	case nil:
		return 0, nil
	default:
		return nil, newUnsupportedTypeError("sizeOf", args[0], e.Pos())
	}
}

// callBuiltinFlatten implements the flatten(array) function.
func callBuiltinFlatten(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := validateArgCount(args, 1, "flatten", e.Pos()); err != nil {
		return nil, err
	}

	if err := assertArg(args[0], beArray(), 1, "flatten", e); err != nil {
		return nil, err
	}
	arr, ok := AsArray(args[0])
	if !ok {
		return nil, newPosError(fmt.Sprintf("flatten: expected array, got %T", args[0]), e.Pos())
	}

	result := make([]interface{}, 0)
	var flattenHelper func([]interface{})
	flattenHelper = func(items []interface{}) {
		for _, item := range items {
			if nested, ok := AsArray(item); ok {
				flattenHelper(nested)
			} else {
				result = append(result, item)
			}
		}
	}

	flattenHelper(arr)
	return result, nil
}

// callBuiltinUnique implements the unique(array) function.
func callBuiltinUnique(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := validateArgCount(args, 1, "unique", e.Pos()); err != nil {
		return nil, err
	}

	if err := assertArg(args[0], beArray(), 1, "unique", e); err != nil {
		return nil, err
	}
	arr, ok := AsArray(args[0])
	if !ok {
		return nil, newPosError(fmt.Sprintf("unique: expected array, got %T", args[0]), e.Pos())
	}

	// Use a map to track seen values, including type to distinguish different types with same string representation
	seen := make(map[string]bool)
	result := make([]interface{}, 0, len(arr))

	for _, item := range arr {
		key := fmt.Sprintf("%T:%v", item, item)
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}

	return result, nil
}

// callBuiltinDistinct implements the distinct(array) function.
func callBuiltinDistinct(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := validateArgCount(args, 1, "distinct", e.Pos()); err != nil {
		return nil, err
	}

	if err := assertArg(args[0], beArray(), 1, "distinct", e); err != nil {
		return nil, err
	}
	arr, ok := AsArray(args[0])
	if !ok {
		return nil, newPosError(fmt.Sprintf("distinct: expected array, got %T", args[0]), e.Pos())
	}

	// Use a map to track seen values, including type to distinguish different types with same string representation
	seen := make(map[string]bool)
	result := make([]interface{}, 0, len(arr))

	for _, item := range arr {
		key := fmt.Sprintf("%T:%v", item, item)
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}

	return result, nil
}

// callBuiltinReverse implements the reverse(array) function.
func callBuiltinReverse(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := validateArgCount(args, 1, "reverse", e.Pos()); err != nil {
		return nil, err
	}

	if err := assertArg(args[0], beArray(), 1, "reverse", e); err != nil {
		return nil, err
	}
	arr, ok := AsArray(args[0])
	if !ok {
		return nil, newPosError(fmt.Sprintf("reverse: expected array, got %T", args[0]), e.Pos())
	}

	result := make([]interface{}, len(arr))
	for i, item := range arr {
		result[len(arr)-1-i] = item
	}
	return result, nil
}

// callBuiltinSort implements the sort(array) function.
func callBuiltinSort(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := validateArgCount(args, 1, "sort", e.Pos()); err != nil {
		return nil, err
	}

	if err := assertArg(args[0], beArray(), 1, "sort", e); err != nil {
		return nil, err
	}
	arr, ok := AsArray(args[0])
	if !ok {
		return nil, newPosError(fmt.Sprintf("sort: expected array, got %T", args[0]), e.Pos())
	}

	// Create a copy to avoid mutating the original
	result := make([]interface{}, len(arr))
	copy(result, arr)

	// Check if all elements are numbers or all are strings
	allNumbers := true
	allStrings := true

	for _, item := range result {
		switch item.(type) {
		case int, float64:
			allStrings = false
		case string:
			allNumbers = false
		default:
			return nil, newPosError(fmt.Sprintf("sort only supports arrays of numbers or strings, got %T", item), e.Pos())
		}
	}

	if allNumbers {
		sort.Slice(result, func(i, j int) bool {
			var af float64
			switch v := result[i].(type) {
			case int:
				af = float64(v)
			case float64:
				af = v
			}

			var bf float64
			switch v := result[j].(type) {
			case int:
				bf = float64(v)
			case float64:
				bf = v
			}

			return af < bf
		})
	} else if allStrings {
		sort.Slice(result, func(i, j int) bool {
			as, okA := result[i].(string)
			bs, okB := result[j].(string)
			if !okA || !okB {
				return false
			}
			return as < bs
		})
	}

	return result, nil
}

// callBuiltinJoin implements the join(array, separator) function.
func callBuiltinJoin(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := validateArgCount(args, 2, "join", e.Pos()); err != nil {
		return nil, err
	}

	if err := assertArg(args[0], beArray(), 1, "join", e); err != nil {
		return nil, err
	}
	arr, ok := AsArray(args[0])
	if !ok {
		return nil, newPosError(fmt.Sprintf("join: expected array, got %T", args[0]), e.Pos())
	}

	if err := assertArg(args[1], beString(), 2, "join", e); err != nil {
		return nil, err
	}
	sep, ok := args[1].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("join: expected string separator, got %T", args[1]), e.Pos())
	}

	parts := make([]string, len(arr))
	for i, item := range arr {
		parts[i] = fmt.Sprintf("%v", item)
	}

	return strings.Join(parts, sep), nil
}

// callBuiltinKeys implements the keys(object) function.
func callBuiltinKeys(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := validateArgCount(args, 1, "keys", e.Pos()); err != nil {
		return nil, err
	}

	if err := assertArg(args[0], beObject(), 1, "keys", e); err != nil {
		return nil, err
	}
	obj, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, newPosError(fmt.Sprintf("keys: expected object, got %T", args[0]), e.Pos())
	}

	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}

	// Sort keys alphabetically for deterministic output
	sort.Strings(keys)

	result := make([]interface{}, len(keys))
	for i, k := range keys {
		result[i] = k
	}

	return result, nil
}

// callBuiltinValues implements the values(object) function.
func callBuiltinValues(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := validateArgCount(args, 1, "values", e.Pos()); err != nil {
		return nil, err
	}

	if err := assertArg(args[0], beObject(), 1, "values", e); err != nil {
		return nil, err
	}
	obj, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, newPosError(fmt.Sprintf("values: expected object, got %T", args[0]), e.Pos())
	}

	// To maintain order, first get sorted keys
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make([]interface{}, len(keys))
	for i, k := range keys {
		result[i] = obj[k]
	}

	return result, nil
}

// callBuiltinMerge implements the merge(obj1, obj2, ...) function.
func callBuiltinMerge(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) < 1 {
		return nil, newPosError("merge requires at least 1 argument", e.Pos())
	}

	result := make(map[string]interface{})

	// Merge all objects in order
	for _, arg := range args {
		if arg == nil {
			continue
		}
		obj, ok := arg.(map[string]interface{})
		if !ok {
			return nil, newPosError(fmt.Sprintf("merge expects objects, got %T", arg), e.Pos())
		}

		for k, v := range obj {
			result[k] = v
		}
	}

	return result, nil
}

// callBuiltinWithAttrs implements the __with_attrs(value, attrs) function.
func callBuiltinWithAttrs(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 2 {
		return nil, newPosError("__with_attrs requires exactly 2 arguments: value and attributes", e.Pos())
	}

	value := args[0]
	attrsVal := args[1]

	if attrsVal == nil {
		return value, nil
	}

	attrs, ok := attrsVal.(map[string]interface{})
	if !ok {
		return nil, newPosError(fmt.Sprintf("__with_attrs expects an attributes object, got %T", attrsVal), e.Pos())
	}

	var result map[string]interface{}

	if obj, ok := value.(map[string]interface{}); ok {
		// Copy original object
		result = make(map[string]interface{}, len(obj)+len(attrs))
		for k, v := range obj {
			result[k] = v
		}
	} else {
		// Wrap non-object value in #text
		result = make(map[string]interface{}, 1+len(attrs))
		if value != nil {
			result["#text"] = value
		}
	}

	// Merge attributes, ensuring @ prefix
	for k, v := range attrs {
		attrKey := k
		if !strings.HasPrefix(k, "@") {
			attrKey = "@" + k
		}
		result[attrKey] = v
	}

	return result, nil
}

// callBuiltinUpdate implements the __update(obj1, obj2) function (~ operator).
func callBuiltinUpdate(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := validateArgCount(args, 2, "update operator (~)", e.Pos()); err != nil {
		return nil, err
	}

	left, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, newPosError(fmt.Sprintf("update operator (~) expects an object on the left, got %T", args[0]), e.Pos())
	}

	right, ok := args[1].(map[string]interface{})
	if !ok {
		return nil, newPosError(fmt.Sprintf("update operator (~) expects an object on the right, got %T", args[1]), e.Pos())
	}

	result := make(map[string]interface{})

	// Copy left object
	for k, v := range left {
		result[k] = v
	}

	// Merge right object (overwrites left values)
	for k, v := range right {
		result[k] = v
	}

	return result, nil
}

// callBuiltinTypeOf implements the typeOf(value) function.
func callBuiltinTypeOf(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 1 {
		return nil, newPosError("typeOf function requires exactly 1 argument", e.Pos())
	}
	return getTypeName(args[0]), nil
}

// callBuiltinIsType implements the __isType(value, typeName) function (is operator).
func callBuiltinIsType(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 2 {
		return nil, newPosError("is operator requires exactly 2 arguments: value is Type", e.Pos())
	}

	typeName, ok := args[1].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("is operator expects type name as string, got %T", args[1]), e.Pos())
	}

	actualType := getTypeName(args[0])
	return actualType == typeName, nil
}

// callBuiltinIsEmpty implements the isEmpty(value) function.
func callBuiltinIsEmpty(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 1 {
		return nil, newPosError("isEmpty function requires exactly 1 argument", e.Pos())
	}
	switch v := args[0].(type) {
	case nil:
		return true, nil
	case string:
		return len(v) == 0, nil
	case []interface{}:
		return len(v) == 0, nil
	case XMLMultiValue:
		return len(v) == 0, nil
	case map[string]interface{}:
		return len(v) == 0, nil
	default:
		return false, nil
	}
}

// callBuiltinDeep implements the __deep(root, fieldName) function (.. operator).
func callBuiltinDeep(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 2 {
		return nil, newPosError("recursive descent (..) requires exactly 2 arguments: root, fieldName", e.Pos())
	}

	fieldName, err := assertStringArg(args[1], 2, "__deep", e)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, 0)
	if err := deepCollect(args[0], fieldName, &result, 0, e.Pos()); err != nil {
		return nil, err
	}
	return result, nil
}

// deepCollect recursively collects all values for the given field name.
func deepCollect(node interface{}, field string, out *[]interface{}, depth int, pos token.Pos) error {
	if depth > MaxDeepDepth {
		return newPosError("deep search depth limit exceeded", pos)
	}

	switch v := node.(type) {
	case map[string]interface{}:
		// If object has the field, add it to results
		if val, ok := v[field]; ok {
			*out = append(*out, val)
		}
		// Recurse into all values
		for _, child := range v {
			if err := deepCollect(child, field, out, depth+1, pos); err != nil {
				return err
			}
		}

	case []interface{}:
		// Recurse into each array element
		for _, elem := range v {
			if err := deepCollect(elem, field, out, depth+1, pos); err != nil {
				return err
			}
		}
	}

	return nil
}

// callBuiltinObjectValues implements the __objvalues(object) function (.* operator).
func callBuiltinObjectValues(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 1 {
		return nil, newPosError("multi-value selector (.*) requires exactly 1 argument: object", e.Pos())
	}

	obj := args[0]

	switch v := obj.(type) {
	case map[string]interface{}:
		// Return all values from the object
		result := make([]interface{}, 0, len(v))
		// Use sorted keys for deterministic ordering
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			result = append(result, v[k])
		}
		return result, nil
	case []interface{}:
		// For arrays, return the array itself
		return v, nil
	case nil:
		return nil, nil
	default:
		return nil, newPosError(fmt.Sprintf("multi-value selector (.*) expects an object or array, got %T", obj), e.Pos())
	}
}

// callBuiltinMultival implements the __multival(object, "field") function (.*field selector).
// For an object: returns [obj[field]] if field exists, else []
// For an array: returns array of item[field] for each item in the array
func callBuiltinMultival(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 2 {
		return nil, newPosError("multi-value field selector (.*field) requires 2 arguments: object, field", e.Pos())
	}

	obj := args[0]
	field, ok := args[1].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("multi-value field selector (.*field) field must be a string, got %T", args[1]), e.Pos())
	}

	switch v := obj.(type) {
	case map[string]interface{}:
		// For a single object, return [obj[field]] if exists
		if val, exists := v[field]; exists {
			if arr, ok := AsArray(val); ok {
				return arr, nil
			}
			return []interface{}{val}, nil
		}
		return []interface{}{}, nil
	case []interface{}:
		// For an array, extract field from each object
		result := make([]interface{}, 0, len(v))
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if val, exists := itemMap[field]; exists {
					result = append(result, val)
				}
			}
		}
		return result, nil
	case XMLMultiValue:
		// For an array, extract field from each object
		result := make([]interface{}, 0, len(v))
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if val, exists := itemMap[field]; exists {
					result = append(result, val)
				}
			}
		}
		return result, nil
	case nil:
		return []interface{}{}, nil
	default:
		return nil, newPosError(fmt.Sprintf("multi-value field selector (.*field) expects an object or array, got %T", obj), e.Pos())
	}
}

// callBuiltinFind implements the find(source, value, [flags]) function.
// findInArray returns indices of elements matching searchValue.
func findInArray(arr []interface{}, searchValue interface{}) []interface{} {
	result := make([]interface{}, 0)
	for i, item := range arr {
		if isEqual(item, searchValue) {
			result = append(result, float64(i))
		}
	}
	return result
}

// findRegexInString finds all regex matches and returns their start/end positions.
func findRegexInString(s, pattern, flags string) ([]interface{}, bool) {
	re, err := compileRegex(pattern, flags)
	if err != nil {
		return nil, false
	}
	matches := re.FindAllStringIndex(s, -1)
	result := make([]interface{}, len(matches))
	for i, match := range matches {
		result[i] = []interface{}{float64(match[0]), float64(match[1])}
	}
	return result, true
}

// findSubstringInString finds all occurrences of a substring and returns their positions.
func findSubstringInString(s, searchStr string) []interface{} {
	result := make([]interface{}, 0)
	if len(searchStr) == 0 {
		return result
	}
	start := 0
	for {
		idx := strings.Index(s[start:], searchStr)
		if idx == -1 {
			break
		}
		result = append(result, float64(start+idx))
		start += idx + 1
	}
	return result
}

// looksLikeRegex returns true if the pattern appears to be a regex.
func looksLikeRegex(pattern string) bool {
	return strings.ContainsAny(pattern, "*+?()[]{}|^$\\") ||
		(len(pattern) > 1 && strings.Contains(pattern, "."))
}

func callBuiltinFind(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, newPosError("find requires 2 or 3 arguments: source, value, [flags]", e.Pos())
	}

	source := args[0]
	searchValue := args[1]
	flags := ""
	if len(args) == 3 {
		if err := assertArg(args[2], beString(), 3, "find", e); err != nil {
			return nil, err
		}
		flags, _ = args[2].(string)
	}

	switch s := source.(type) {
	case []interface{}:
		return findInArray(s, searchValue), nil

	case string:
		// Handle Regex object
		if r, ok := searchValue.(*Regex); ok {
			matches := r.Re.FindAllStringIndex(s, -1)
			result := make([]interface{}, len(matches))
			for i, match := range matches {
				result[i] = []interface{}{float64(match[0]), float64(match[1])}
			}
			return result, nil
		}

		searchStr, ok := searchValue.(string)
		if !ok {
			return nil, newPosError(fmt.Sprintf("find on string expects search value to be a string or Regex, got %T", searchValue), e.Pos())
		}
		if flags != "" || looksLikeRegex(searchStr) {
			if result, ok := findRegexInString(s, searchStr, flags); ok {
				return result, nil
			}
		}
		return findSubstringInString(s, searchStr), nil

	default:
		return nil, newPosError(fmt.Sprintf("find expects an array or string, got %T", source), e.Pos())
	}
}
