package evaluator

import (
	"fmt"
	"unicode/utf8"

	unifiederrors "infomunge/internal/errors"
)

// MatcherResult represents the result of an assertion
type MatcherResult struct {
	Success bool
	Message string
}

// Matcher represents a matcher function that validates a value
type Matcher func(Value) *MatcherResult

// Matched is a variable that represents a successful match
var Matched = &MatcherResult{Success: true, Message: ""}

// assertion functions return Matcher objects that can be applied to values

// typeError creates a failed MatcherResult for type mismatch
func typeError(expected string, value Value) *MatcherResult {
	return &MatcherResult{Success: false, Message: fmt.Sprintf("expected %s, got %T", expected, value)}
}

// typeMatcherFactory creates a type matcher that validates via type assertion
func typeMatcherFactory(typeName string, checker func(Value) bool) Matcher {
	return func(value Value) *MatcherResult {
		if checker(value) {
			return Matched
		}
		return typeError(typeName, value)
	}
}

// beArray validates that a given value is of type Array
func beArray() Matcher {
	return typeMatcherFactory("array", func(v Value) bool {
		_, ok := v.(Array)
		return ok
	})
}

// beObject validates that a given value is of type Object
func beObject() Matcher {
	return typeMatcherFactory("object", func(v Value) bool {
		_, ok := v.(Object)
		return ok
	})
}

// beString validates that a given value is of type String
func beString() Matcher {
	return typeMatcherFactory("string", func(v Value) bool {
		_, ok := v.(string)
		return ok
	})
}

// beNumber validates that a given value is of type Number
func beNumber() Matcher {
	return typeMatcherFactory("number", func(v Value) bool {
		switch v.(type) {
		case float64, int:
			return true
		default:
			return false
		}
	})
}

// beBoolean validates that a given value is of type Boolean
func beBoolean() Matcher {
	return typeMatcherFactory("boolean", func(v Value) bool {
		_, ok := v.(bool)
		return ok
	})
}

// beNull validates that a given value is of type Null
func beNull() Matcher {
	return typeMatcherFactory("null", func(v Value) bool {
		return v == nil
	})
}

// beEmpty validates that the value (String, Object or Array) is empty
func beEmpty() Matcher {
	return func(value Value) *MatcherResult {
		switch v := value.(type) {
		case string:
			if len(v) == 0 {
				return Matched
			}
			return &MatcherResult{Success: false, Message: fmt.Sprintf("expected empty string, got %q", v)}
		case Array:
			if len(v) == 0 {
				return Matched
			}
			return &MatcherResult{Success: false, Message: fmt.Sprintf("expected empty array, got array with %d items", len(v))}
		case Object:
			if len(v) == 0 {
				return Matched
			}
			return &MatcherResult{Success: false, Message: fmt.Sprintf("expected empty object, got object with %d keys", len(v))}
		default:
			return &MatcherResult{Success: false, Message: fmt.Sprintf("expected string, array, or object, got %T", value)}
		}
	}
}

// beBlank validates that the String value is blank (empty or whitespace)
func beBlank() Matcher {
	return func(value Value) *MatcherResult {
		str, ok := value.(string)
		if !ok {
			return &MatcherResult{Success: false, Message: fmt.Sprintf("expected string, got %T", value)}
		}
		// Check if string is empty or only whitespace
		for _, ch := range str {
			if ch > ' ' {
				return &MatcherResult{Success: false, Message: fmt.Sprintf("expected blank string, got %q", str)}
			}
		}
		return Matched
	}
}

// equalTo validates that a value is equal to another one
func equalTo(expected Value) Matcher {
	return func(value Value) *MatcherResult {
		if isEqual(value, expected) {
			return Matched
		}
		return &MatcherResult{Success: false, Message: fmt.Sprintf("expected %v, got %v", expected, value)}
	}
}

// beGreaterThan validates that the asserted Comparable value is greater than the given one
func beGreaterThan(threshold Value, inclusive bool) Matcher {
	return func(value Value) *MatcherResult {
		cmp, err := compareValues(value, threshold)
		if err != nil {
			return &MatcherResult{Success: false, Message: err.Error()}
		}

		if inclusive {
			if cmp >= 0 {
				return Matched
			}
			return &MatcherResult{Success: false, Message: fmt.Sprintf("expected value >= %v, got %v", threshold, value)}
		}

		if cmp > 0 {
			return Matched
		}
		return &MatcherResult{Success: false, Message: fmt.Sprintf("expected value > %v, got %v", threshold, value)}
	}
}

// beLowerThan validates that the asserted Comparable value is lower than the given one
func beLowerThan(threshold Value, inclusive bool) Matcher {
	return func(value Value) *MatcherResult {
		cmp, err := compareValues(value, threshold)
		if err != nil {
			return &MatcherResult{Success: false, Message: err.Error()}
		}

		if inclusive {
			if cmp <= 0 {
				return Matched
			}
			return &MatcherResult{Success: false, Message: fmt.Sprintf("expected value <= %v, got %v", threshold, value)}
		}

		if cmp < 0 {
			return Matched
		}
		return &MatcherResult{Success: false, Message: fmt.Sprintf("expected value < %v, got %v", threshold, value)}
	}
}

// beOneOf validates that the value is contained in the given Array
func beOneOf(options Array) Matcher {
	return func(value Value) *MatcherResult {
		for _, opt := range options {
			if isEqual(value, opt) {
				return Matched
			}
		}
		return &MatcherResult{Success: false, Message: fmt.Sprintf("expected one of %v, got %v", options, value)}
	}
}

// contain validates that the asserted String contains the given String
func containString(substr string) Matcher {
	return func(value Value) *MatcherResult {
		str, ok := value.(string)
		if !ok {
			return &MatcherResult{Success: false, Message: fmt.Sprintf("expected string, got %T", value)}
		}
		for i := 0; i <= len(str)-len(substr); i++ {
			if str[i:i+len(substr)] == substr {
				return Matched
			}
		}
		return &MatcherResult{Success: false, Message: fmt.Sprintf("expected string to contain %q, got %q", substr, str)}
	}
}

// containValue validates that the asserted Array contains the given value
func containValue(needle Value) Matcher {
	return func(value Value) *MatcherResult {
		arr, ok := value.(Array)
		if !ok {
			return &MatcherResult{Success: false, Message: fmt.Sprintf("expected array, got %T", value)}
		}
		for _, item := range arr {
			if isEqual(item, needle) {
				return Matched
			}
		}
		return &MatcherResult{Success: false, Message: fmt.Sprintf("expected array to contain %v", needle)}
	}
}

// startWith validates that the asserted String starts with the given String
func startWith(prefix string) Matcher {
	return func(value Value) *MatcherResult {
		str, ok := value.(string)
		if !ok {
			return &MatcherResult{Success: false, Message: fmt.Sprintf("expected string, got %T", value)}
		}
		if len(str) >= len(prefix) && str[:len(prefix)] == prefix {
			return Matched
		}
		return &MatcherResult{Success: false, Message: fmt.Sprintf("expected string to start with %q, got %q", prefix, str)}
	}
}

// endWith validates that the asserted String ends with the given String
func endWith(suffix string) Matcher {
	return func(value Value) *MatcherResult {
		str, ok := value.(string)
		if !ok {
			return &MatcherResult{Success: false, Message: fmt.Sprintf("expected string, got %T", value)}
		}
		if len(str) >= len(suffix) && str[len(str)-len(suffix):] == suffix {
			return Matched
		}
		return &MatcherResult{Success: false, Message: fmt.Sprintf("expected string to end with %q, got %q", suffix, str)}
	}
}

// haveSize validates that the string/binary/array/object has the given size.
func haveSize(expectedSize int) Matcher {
	return func(value Value) *MatcherResult {
		var size int
		switch v := value.(type) {
		case string:
			size = utf8.RuneCountInString(v)
		case []byte:
			size = len(v)
		case Array:
			size = len(v)
		case Object:
			size = len(v)
		default:
			return &MatcherResult{Success: false, Message: fmt.Sprintf("expected string, binary, array, or object, got %T", value)}
		}

		if size == expectedSize {
			return Matched
		}
		return &MatcherResult{Success: false, Message: fmt.Sprintf("expected size %d, got %d", expectedSize, size)}
	}
}

// haveKey validates that the Object has the given key
func haveKey(key string) Matcher {
	return func(value Value) *MatcherResult {
		obj, ok := value.(Object)
		if !ok {
			return &MatcherResult{Success: false, Message: fmt.Sprintf("expected object, got %T", value)}
		}
		if _, exists := obj[key]; exists {
			return Matched
		}
		return &MatcherResult{Success: false, Message: fmt.Sprintf("expected object to have key %q", key)}
	}
}

// haveValue validates that the Object has the given value
func haveValue(expectedVal Value) Matcher {
	return func(value Value) *MatcherResult {
		obj, ok := value.(Object)
		if !ok {
			return &MatcherResult{Success: false, Message: fmt.Sprintf("expected object, got %T", value)}
		}
		for _, val := range obj {
			if isEqual(val, expectedVal) {
				return Matched
			}
		}
		return &MatcherResult{Success: false, Message: fmt.Sprintf("expected object to have value %v", expectedVal)}
	}
}

// eachItem validates that each item of the array satisfies the given matcher
func eachItem(matcher Matcher) Matcher {
	return func(value Value) *MatcherResult {
		arr, ok := value.(Array)
		if !ok {
			return &MatcherResult{Success: false, Message: fmt.Sprintf("expected array, got %T", value)}
		}
		for i, item := range arr {
			result := matcher(item)
			if !result.Success {
				return &MatcherResult{Success: false, Message: fmt.Sprintf("item at index %d: %s", i, result.Message)}
			}
		}
		return Matched
	}
}

// haveItem validates that at least one item of the array satisfies the given matcher
func haveItem(matcher Matcher) Matcher {
	return func(value Value) *MatcherResult {
		arr, ok := value.(Array)
		if !ok {
			return &MatcherResult{Success: false, Message: fmt.Sprintf("expected array, got %T", value)}
		}
		for _, item := range arr {
			result := matcher(item)
			if result.Success {
				return Matched
			}
		}
		return &MatcherResult{Success: false, Message: "no items matched the given matcher"}
	}
}

// anyOf validates that the value satisfies at least one of the given matchers
func anyOf(matchers []Matcher) Matcher {
	return func(value Value) *MatcherResult {
		for _, matcher := range matchers {
			result := matcher(value)
			if result.Success {
				return Matched
			}
		}
		return &MatcherResult{Success: false, Message: "value did not satisfy any of the matchers"}
	}
}

// notBe validates that the value doesn't satisfy the given matcher
func notBe(matcher Matcher) Matcher {
	return func(value Value) *MatcherResult {
		result := matcher(value)
		if !result.Success {
			return Matched
		}
		return &MatcherResult{Success: false, Message: fmt.Sprintf("expected matcher to fail, but it passed")}
	}
}

// notBeNull validates that a given value isn't of type Null
func notBeNull() Matcher {
	return func(value Value) *MatcherResult {
		if value != nil {
			return Matched
		}
		return &MatcherResult{Success: false, Message: "expected non-null value, got null"}
	}
}

// Helper functions

func isEqual(a, b Value) bool {
	return numericEquals(a, b)
}

func compareValues(a, b Value) (int, error) {
	// Compare the exact numeric values represented by ints and float64s. This
	// must stay aligned with relational operators so adjacent integers above
	// 2^53 do not collapse when an int is converted to float64.
	aNum, aIsNum := exactNumericRat(a)
	bNum, bIsNum := exactNumericRat(b)
	if aIsNum && bIsNum {
		return aNum.Cmp(bNum), nil
	}

	// Try string comparison
	aStr, aIsStr := a.(string)
	bStr, bIsStr := b.(string)
	if aIsStr && bIsStr {
		if aStr < bStr {
			return -1, nil
		}
		if aStr > bStr {
			return 1, nil
		}
		return 0, nil
	}

	return 0, unifiederrors.EvalErrorf("cannot compare values: %T and %T", a, b)
}
