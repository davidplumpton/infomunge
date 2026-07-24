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
	TypeIO       ErrorType = "IOError"
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

// Error represents a unified error with position and cause information.
type Error struct {
	Type     ErrorType
	Message  string
	Position *Position
	Cause    error
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

func newError(errType ErrorType, message string) *Error {
	return &Error{
		Type:    errType,
		Message: message,
	}
}

func newErrorf(errType ErrorType, format string, args ...interface{}) *Error {
	return &Error{
		Type:    errType,
		Message: fmt.Sprintf(format, args...),
	}
}

func wrapError(err error, errType ErrorType, message string) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Cause:   err,
	}
}

func wrapErrorf(err error, errType ErrorType, format string, args ...interface{}) *Error {
	return &Error{
		Type:    errType,
		Message: fmt.Sprintf(format, args...),
		Cause:   err,
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

func (e *Error) withPos(pos token.Pos, fset *token.FileSet) *Error {
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
