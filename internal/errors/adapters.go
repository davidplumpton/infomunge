package errors

import (
	"fmt"
	"go/token"
)

// EvaluatorAdapter provides compatibility with the existing evaluator error patterns
type EvaluatorAdapter struct {
	fset *token.FileSet
}

// NewEvaluatorAdapter creates a new evaluator adapter
func NewEvaluatorAdapter(fset *token.FileSet) *EvaluatorAdapter {
	return &EvaluatorAdapter{fset: fset}
}

// NewPosError creates a new positional error (replaces evaluator.newPosError)
func (a *EvaluatorAdapter) NewPosError(msg string, pos token.Pos) error {
	return EvalPositional(msg, pos, a.fset)
}

// NewPosErrorf creates a new positional error with formatted message
func (a *EvaluatorAdapter) NewPosErrorf(pos token.Pos, format string, args ...interface{}) error {
	return EvalPositionalf(pos, a.fset, format, args...)
}

// UserErrorAdapter provides compatibility with the existing UserError type
type UserErrorAdapter struct {
	Message  string
	Kind     string
	Location string
	pos      token.Pos
	fset     *token.FileSet
}

// NewUserErrorAdapter creates a new user error adapter
func NewUserErrorAdapter(message, kind, location string, pos token.Pos, fset *token.FileSet) *UserErrorAdapter {
	return &UserErrorAdapter{
		Message:  message,
		Kind:     kind,
		Location: location,
		pos:      pos,
		fset:     fset,
	}
}

func (e *UserErrorAdapter) Error() string {
	return e.Message
}

func (e *UserErrorAdapter) Pos() token.Pos {
	return e.pos
}

// ToError converts to the unified error type
func (e *UserErrorAdapter) ToError() *Error {
	err := UserError(e.Message).WithPos(e.pos, e.fset)
	if e.Kind != "" {
		err.WithContext("kind", e.Kind)
	}
	if e.Location != "" {
		err.WithContext("location", e.Location)
	}
	return err
}

// ErrorContextAdapter provides compatibility with the existing ErrorContext
type ErrorContextAdapter struct {
	ExprStr    string
	Mapping    []int
	BodyOffset int
	FullRaw    string
	fset       *token.FileSet
}

// NewErrorContextAdapter creates a new error context adapter
func NewErrorContextAdapter(exprStr string, mapping []int, bodyOffset int, fullRaw string, fset *token.FileSet) *ErrorContextAdapter {
	return &ErrorContextAdapter{
		ExprStr:    exprStr,
		Mapping:    mapping,
		BodyOffset: bodyOffset,
		FullRaw:    fullRaw,
		fset:       fset,
	}
}

// FormatEvalError converts a posError to a user-friendly error (replaces ErrorContext.FormatEvalError)
func (ec *ErrorContextAdapter) FormatEvalError(err error) error {
	if posErr, ok := err.(*Error); ok && posErr.Position != nil {
		// Already has position info, just add more context
		posErr.WithContext("exprStr", ec.ExprStr)
		posErr.WithContext("bodyOffset", ec.BodyOffset)
		return posErr
	}

	// If it's an old-style posError, convert it
	if posErr, ok := err.(interface{ Pos() token.Pos }); ok {
		converted := EvalError(err.Error()).WithPos(posErr.Pos(), ec.fset)
		converted.WithContext("exprStr", ec.ExprStr)
		converted.WithContext("bodyOffset", ec.BodyOffset)
		return converted
	}

	// Fallback for generic errors
	return EvalError(err.Error())
}

// FormatParseError converts a parser error to a user-friendly error (replaces ErrorContext.FormatParseError)
func (ec *ErrorContextAdapter) FormatParseError(err error, exprStr string) error {
	parseErr := ParseError(err.Error())
	parseErr.WithContext("exprStr", exprStr)
	parseErr.WithContext("bodyOffset", ec.BodyOffset)
	return parseErr
}

// Legacy compatibility functions for gradual migration

// IsPosError checks if an error is a positional error (legacy compatibility)
func IsPosError(err error) bool {
	_, ok := err.(*Error)
	return ok
}

// IsUserError checks if an error is a user error (legacy compatibility)
func IsUserError(err error) bool {
	if unifiedErr, ok := err.(*Error); ok {
		return unifiedErr.IsUserError()
	}
	return false
}

// GetErrorPosition extracts position from any error type (legacy compatibility)
func GetErrorPosition(err error) (token.Pos, bool) {
	if unifiedErr, ok := err.(*Error); ok && unifiedErr.Position != nil {
		// Convert back to token.Pos for compatibility
		if unifiedErr.Position.Filename != "" {
			// We can't easily convert back to token.Pos without the FileSet
			return token.NoPos, false
		}
		return token.Pos(unifiedErr.Position.Offset), true
	}

	// Check for legacy posError
	if posErr, ok := err.(interface{ Pos() token.Pos }); ok {
		return posErr.Pos(), true
	}

	return token.NoPos, false
}

// WrapForRunner wraps errors for the runner package (replaces common runner patterns)
func WrapForRunner(err error, context string) *Error {
	if err == nil {
		return nil
	}

	// If it's already a unified error, just add context
	if unifiedErr, ok := err.(*Error); ok {
		return unifiedErr.WithContext("runner_context", context)
	}

	// Common runner error patterns
	switch context {
	case "file_read":
		return WrapIO(err, "error reading file")
	case "invalid_function":
		return WrapParse(err, "invalid function body")
	case "invalid_declaration":
		return WrapParse(err, "invalid declaration")
	default:
		return WrapRuntime(err, fmt.Sprintf("runner error in %s", context))
	}
}
