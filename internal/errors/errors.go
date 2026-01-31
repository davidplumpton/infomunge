package errors

import (
	"fmt"
	"go/token"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	TypeParse    ErrorType = "ParseError"
	TypeEval     ErrorType = "EvalError"
	TypeRuntime  ErrorType = "RuntimeError"
	TypeIO       ErrorType = "IOError"
	TypeUser     ErrorType = "UserError"
	TypeValidate ErrorType = "ValidationError"
	TypeInternal ErrorType = "InternalError"
)

// Position represents source code position information
type Position struct {
	Filename string
	Offset   int
	Line     int
	Column   int
}

// Error represents a unified error with position, type, and context information
type Error struct {
	Type      ErrorType
	Message   string
	Position  *Position
	Cause     error
	Context   map[string]interface{}
	UserError bool // true for errors explicitly thrown by users
}

func (e *Error) Error() string {
	if e.Position != nil {
		return fmt.Sprintf("%s at %s:%d:%d: %s", e.Type, e.Position.Filename, e.Position.Line, e.Position.Column, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target error type
func (e *Error) Is(target error) bool {
	if t, ok := target.(*Error); ok {
		return e.Type == t.Type
	}
	return false
}

// New creates a new error with the specified type and message
func New(errType ErrorType, message string) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// Newf creates a new error with formatted message
func Newf(errType ErrorType, format string, args ...interface{}) *Error {
	return &Error{
		Type:    errType,
		Message: fmt.Sprintf(format, args...),
		Context: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errType ErrorType, message string) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Cause:   err,
		Context: make(map[string]interface{}),
	}
}

// Wrapf wraps an existing error with formatted message
func Wrapf(err error, errType ErrorType, format string, args ...interface{}) *Error {
	return &Error{
		Type:    errType,
		Message: fmt.Sprintf(format, args...),
		Cause:   err,
		Context: make(map[string]interface{}),
	}
}

// WithPosition adds position information to an error
func (e *Error) WithPosition(pos token.Position) *Error {
	e.Position = &Position{
		Filename: pos.Filename,
		Offset:   pos.Offset,
		Line:     pos.Line,
		Column:   pos.Column,
	}
	return e
}

// WithPos adds token.Pos position information to an error
func (e *Error) WithPos(pos token.Pos, fset *token.FileSet) *Error {
	if fset != nil {
		position := fset.Position(pos)
		e.Position = &Position{
			Filename: position.Filename,
			Offset:   position.Offset,
			Line:     position.Line,
			Column:   position.Column,
		}
	}
	return e
}

// WithContext adds context key-value pairs to an error
func (e *Error) WithContext(key string, value interface{}) *Error {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithUserError marks this as a user-thrown error
func (e *Error) WithUserError() *Error {
	e.UserError = true
	return e
}

// IsUserError returns true if this error was explicitly thrown by a user
func (e *Error) IsUserError() bool {
	return e.UserError
}

// GetType returns the error type
func (e *Error) GetType() ErrorType {
	return e.Type
}

// GetPosition returns the position information
func (e *Error) GetPosition() *Position {
	return e.Position
}

// GetContext returns the context map
func (e *Error) GetContext() map[string]interface{} {
	return e.Context
}

// Builder provides a fluent interface for building errors
type Builder struct {
	error *Error
}

// NewBuilder creates a new error builder
func NewBuilder(errType ErrorType, message string) *Builder {
	return &Builder{
		error: New(errType, message),
	}
}

// NewBuilderf creates a new error builder with formatted message
func NewBuilderf(errType ErrorType, format string, args ...interface{}) *Builder {
	return &Builder{
		error: Newf(errType, format, args...),
	}
}

// Wrap wraps an existing error
func (b *Builder) Wrap(err error) *Builder {
	b.error.Cause = err
	return b
}

// WithPosition adds position information
func (b *Builder) WithPosition(pos token.Position) *Builder {
	b.error.WithPosition(pos)
	return b
}

// WithPos adds token.Pos position information
func (b *Builder) WithPos(pos token.Pos, fset *token.FileSet) *Builder {
	b.error.WithPos(pos, fset)
	return b
}

// WithContext adds context key-value pairs
func (b *Builder) WithContext(key string, value interface{}) *Builder {
	b.error.WithContext(key, value)
	return b
}

// WithUserError marks as user error
func (b *Builder) WithUserError() *Builder {
	b.error.WithUserError()
	return b
}

// Build returns the final error
func (b *Builder) Build() *Error {
	return b.error
}
