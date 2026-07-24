package errors

import (
	"go/token"
)

// ParseError creates a parse error
func ParseError(message string) *Error {
	return newError(TypeParse, message)
}

// ParseErrorf creates a parse error with formatted message
func ParseErrorf(format string, args ...interface{}) *Error {
	return newErrorf(TypeParse, format, args...)
}

// EvalError creates an evaluation error
func EvalError(message string) *Error {
	return newError(TypeEval, message)
}

// EvalErrorf creates an evaluation error with formatted message
func EvalErrorf(format string, args ...interface{}) *Error {
	return newErrorf(TypeEval, format, args...)
}

// IOErrorf creates an I/O error with formatted message
func IOErrorf(format string, args ...interface{}) *Error {
	return newErrorf(TypeIO, format, args...)
}

// ValidationError creates a validation error
func ValidationError(message string) *Error {
	return newError(TypeValidate, message)
}

// ValidationErrorf creates a validation error with formatted message
func ValidationErrorf(format string, args ...interface{}) *Error {
	return newErrorf(TypeValidate, format, args...)
}

// InternalError creates an internal error
func InternalError(message string) *Error {
	return newError(TypeInternal, message)
}

// EvalPositional creates a positional evaluation error
func EvalPositional(message string, pos token.Pos, fset *token.FileSet) *Error {
	return EvalError(message).withPos(pos, fset)
}

// Wrap helpers

// WrapParse wraps an error as a parse error
func WrapParse(err error, message string) *Error {
	return wrapError(err, TypeParse, message)
}

// WrapParsef wraps an error as a parse error with formatted message
func WrapParsef(err error, format string, args ...interface{}) *Error {
	return wrapErrorf(err, TypeParse, format, args...)
}

// WrapIOf wraps an error as an I/O error with formatted message
func WrapIOf(err error, format string, args ...interface{}) *Error {
	return wrapErrorf(err, TypeIO, format, args...)
}

// WrapValidationf wraps an error as a validation error with formatted message
func WrapValidationf(err error, format string, args ...interface{}) *Error {
	return wrapErrorf(err, TypeValidate, format, args...)
}
