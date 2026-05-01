package core

import "infomunge/pkg/values"

// Object represents a structured data object (map with string keys).
// The values can be nil, bool, string, number (int or float64), array, or object.
type Object = values.Object

// Array represents a sequence of values.
// The values can be nil, bool, string, number (int or float64), array, or object.
type Array = values.Array

// XMLMultiValue represents an array of values that were originally
// duplicate keys in an XML object. This helps mapObject distinguish
// between a natural array and repeated elements.
type XMLMultiValue = values.XMLMultiValue

// Namespace represents an XML namespace with an optional prefix.
// Prefix "" indicates the default namespace.
type Namespace = values.Namespace

// Result represents either a successful value of type T or an error.
// Similar to Rust's Result type, it provides a more explicit way to handle
// operations that can fail, with support for method chaining.
type Result[T any] struct {
	value T
	err   error
}

// Ok creates a successful Result containing the given value.
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value}
}

// Err creates a failed Result containing the given error.
func Err[T any](err error) Result[T] {
	return Result[T]{err: err}
}

// IsOk returns true if the Result contains a successful value.
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// IsErr returns true if the Result contains an error.
func (r Result[T]) IsErr() bool {
	return r.err != nil
}

// Unwrap returns the value and error for traditional Go error handling.
func (r Result[T]) Unwrap() (T, error) {
	return r.value, r.err
}

// Error returns the contained error, or nil if successful.
func (r Result[T]) Error() error {
	return r.err
}

// OrElse returns the value if successful, or the provided default if error.
func (r Result[T]) OrElse(defaultValue T) T {
	if r.err != nil {
		return defaultValue
	}
	return r.value
}
