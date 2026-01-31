package errors

import (
	"go/token"
)

// ParseError creates a parse error
func ParseError(message string) *Error {
	return New(TypeParse, message)
}

// ParseErrorf creates a parse error with formatted message
func ParseErrorf(format string, args ...interface{}) *Error {
	return Newf(TypeParse, format, args...)
}

// EvalError creates an evaluation error
func EvalError(message string) *Error {
	return New(TypeEval, message)
}

// EvalErrorf creates an evaluation error with formatted message
func EvalErrorf(format string, args ...interface{}) *Error {
	return Newf(TypeEval, format, args...)
}

// RuntimeError creates a runtime error
func RuntimeError(message string) *Error {
	return New(TypeRuntime, message)
}

// RuntimeErrorf creates a runtime error with formatted message
func RuntimeErrorf(format string, args ...interface{}) *Error {
	return Newf(TypeRuntime, format, args...)
}

// IOError creates an I/O error
func IOError(message string) *Error {
	return New(TypeIO, message)
}

// IOErrorf creates an I/O error with formatted message
func IOErrorf(format string, args ...interface{}) *Error {
	return Newf(TypeIO, format, args...)
}

// ValidationError creates a validation error
func ValidationError(message string) *Error {
	return New(TypeValidate, message)
}

// ValidationErrorf creates a validation error with formatted message
func ValidationErrorf(format string, args ...interface{}) *Error {
	return Newf(TypeValidate, format, args...)
}

// InternalError creates an internal error
func InternalError(message string) *Error {
	return New(TypeInternal, message)
}

// InternalErrorf creates an internal error with formatted message
func InternalErrorf(format string, args ...interface{}) *Error {
	return Newf(TypeInternal, format, args...)
}

// UserError creates a user-thrown error
func UserError(message string) *Error {
	return New(TypeUser, message).WithUserError()
}

// UserErrorf creates a user-thrown error with formatted message
func UserErrorf(format string, args ...interface{}) *Error {
	return Newf(TypeUser, format, args...).WithUserError()
}

// Positional helpers for common patterns

// NewPositional creates a position-aware error
func NewPositional(errType ErrorType, message string, pos token.Pos, fset *token.FileSet) *Error {
	return New(errType, message).WithPos(pos, fset)
}

// NewPositionalf creates a position-aware error with formatted message
func NewPositionalf(errType ErrorType, pos token.Pos, fset *token.FileSet, format string, args ...interface{}) *Error {
	return Newf(errType, format, args...).WithPos(pos, fset)
}

// ParsePositional creates a positional parse error
func ParsePositional(message string, pos token.Pos, fset *token.FileSet) *Error {
	return ParseError(message).WithPos(pos, fset)
}

// ParsePositionalf creates a positional parse error with formatted message
func ParsePositionalf(pos token.Pos, fset *token.FileSet, format string, args ...interface{}) *Error {
	return ParseErrorf(format, args...).WithPos(pos, fset)
}

// EvalPositional creates a positional evaluation error
func EvalPositional(message string, pos token.Pos, fset *token.FileSet) *Error {
	return EvalError(message).WithPos(pos, fset)
}

// EvalPositionalf creates a positional evaluation error with formatted message
func EvalPositionalf(pos token.Pos, fset *token.FileSet, format string, args ...interface{}) *Error {
	return EvalErrorf(format, args...).WithPos(pos, fset)
}

// RuntimePositional creates a positional runtime error
func RuntimePositional(message string, pos token.Pos, fset *token.FileSet) *Error {
	return RuntimeError(message).WithPos(pos, fset)
}

// RuntimePositionalf creates a positional runtime error with formatted message
func RuntimePositionalf(pos token.Pos, fset *token.FileSet, format string, args ...interface{}) *Error {
	return RuntimeErrorf(format, args...).WithPos(pos, fset)
}

// ValidationPositional creates a positional validation error
func ValidationPositional(message string, pos token.Pos, fset *token.FileSet) *Error {
	return ValidationError(message).WithPos(pos, fset)
}

// ValidationPositionalf creates a positional validation error with formatted message
func ValidationPositionalf(pos token.Pos, fset *token.FileSet, format string, args ...interface{}) *Error {
	return ValidationErrorf(format, args...).WithPos(pos, fset)
}

// UserPositional creates a positional user error
func UserPositional(message string, pos token.Pos, fset *token.FileSet) *Error {
	return UserError(message).WithPos(pos, fset)
}

// UserPositionalf creates a positional user error with formatted message
func UserPositionalf(pos token.Pos, fset *token.FileSet, format string, args ...interface{}) *Error {
	return UserErrorf(format, args...).WithPos(pos, fset)
}

// Wrap helpers

// WrapParse wraps an error as a parse error
func WrapParse(err error, message string) *Error {
	return Wrap(err, TypeParse, message)
}

// WrapParsef wraps an error as a parse error with formatted message
func WrapParsef(err error, format string, args ...interface{}) *Error {
	return Wrapf(err, TypeParse, format, args...)
}

// WrapEval wraps an error as an evaluation error
func WrapEval(err error, message string) *Error {
	return Wrap(err, TypeEval, message)
}

// WrapEvalf wraps an error as an evaluation error with formatted message
func WrapEvalf(err error, format string, args ...interface{}) *Error {
	return Wrapf(err, TypeEval, format, args...)
}

// WrapRuntime wraps an error as a runtime error
func WrapRuntime(err error, message string) *Error {
	return Wrap(err, TypeRuntime, message)
}

// WrapRuntimef wraps an error as a runtime error with formatted message
func WrapRuntimef(err error, format string, args ...interface{}) *Error {
	return Wrapf(err, TypeRuntime, format, args...)
}

// WrapIO wraps an error as an I/O error
func WrapIO(err error, message string) *Error {
	return Wrap(err, TypeIO, message)
}

// WrapIOf wraps an error as an I/O error with formatted message
func WrapIOf(err error, format string, args ...interface{}) *Error {
	return Wrapf(err, TypeIO, format, args...)
}

// WrapValidation wraps an error as a validation error
func WrapValidation(err error, message string) *Error {
	return Wrap(err, TypeValidate, message)
}

// WrapValidationf wraps an error as a validation error with formatted message
func WrapValidationf(err error, format string, args ...interface{}) *Error {
	return Wrapf(err, TypeValidate, format, args...)
}

// WrapInternal wraps an error as an internal error
func WrapInternal(err error, message string) *Error {
	return Wrap(err, TypeInternal, message)
}

// WrapInternalf wraps an error as an internal error with formatted message
func WrapInternalf(err error, format string, args ...interface{}) *Error {
	return Wrapf(err, TypeInternal, format, args...)
}
