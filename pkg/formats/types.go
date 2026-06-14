package formats

import "infomunge/pkg/values"

// Object is a public compatibility alias for the shared runtime object shape.
// New internal code should import pkg/values directly.
type Object = values.Object

// Array is a public compatibility alias for the shared runtime array shape.
// New internal code should import pkg/values directly.
type Array = values.Array

// XMLMultiValue is a public compatibility alias for repeated XML values.
// New internal code should import pkg/values directly.
type XMLMultiValue = values.XMLMultiValue

// Namespace is a public compatibility alias for XML namespace values.
// New internal code should import pkg/values directly.
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
