package evaluator

import (
	"fmt"
	"go/token"
)

// Error message templates for consistency
const (
	MsgTypeMismatch      = "%s expects %s, got %T"
	MsgArgCountExact     = "%s requires exactly %d argument(s)"
	MsgArgCountRange     = "%s requires %d to %d arguments"
	MsgIndexOutOfRange   = "%s index %d out of range"
	MsgInvalidValue      = "invalid %s: %v"
	MsgCannotCoerce      = "cannot coerce %v to %s"
	MsgCannotCoerceType  = "cannot coerce %T to %s"
	MsgElementNotNumber  = "%s: element at index %d is not a number, got %T"
	MsgUnsupportedType   = "%s: unsupported type %T"
	MsgRegexError        = "invalid regex pattern: %s"
	MsgCyclicType        = "cyclic type definition: %s"
	MsgUnknownType       = "unknown type: %s"
	MsgLambdaParamCount  = "%s lambda must have %s parameters, got %d"
	MsgLambdaWrongReturn = "%s lambda must return %s, got %T"
)

// Helper functions for common error patterns

// newTypeMismatchError creates a type mismatch error
func newTypeMismatchError(funcName string, expected string, got interface{}, pos token.Pos) error {
	return newPosError(fmt.Sprintf(MsgTypeMismatch, funcName, expected, got), pos)
}

// newArgCountError creates an argument count error
func newArgCountError(funcName string, expected int, pos token.Pos) error {
	return newPosError(fmt.Sprintf(MsgArgCountExact, funcName, expected), pos)
}

// newArgCountRangeError creates an argument count range error
func newArgCountRangeError(funcName string, min, max int, pos token.Pos) error {
	return newPosError(fmt.Sprintf(MsgArgCountRange, funcName, min, max), pos)
}

// newIndexOutOfRangeError creates an index out of range error
func newIndexOutOfRangeError(funcName string, index int, pos token.Pos) error {
	return newPosError(fmt.Sprintf(MsgIndexOutOfRange, funcName, index), pos)
}

// newInvalidValueError creates an invalid value error
func newInvalidValueError(what string, value interface{}, pos token.Pos) error {
	return newPosError(fmt.Sprintf(MsgInvalidValue, what, value), pos)
}

// newCoercionError creates a coercion error
func newCoercionError(value interface{}, targetType string, pos token.Pos) error {
	return newPosError(fmt.Sprintf(MsgCannotCoerce, value, targetType), pos)
}

// newCoercionTypeError creates a coercion error with type info
func newCoercionTypeError(value interface{}, targetType string, pos token.Pos) error {
	return newPosError(fmt.Sprintf(MsgCannotCoerceType, value, targetType), pos)
}

// newElementNotNumberError creates an element not number error
func newElementNotNumberError(funcName string, index int, got interface{}, pos token.Pos) error {
	return newPosError(fmt.Sprintf(MsgElementNotNumber, funcName, index, got), pos)
}

// newUnsupportedTypeError creates an unsupported type error
func newUnsupportedTypeError(funcName string, got interface{}, pos token.Pos) error {
	return newPosError(fmt.Sprintf(MsgUnsupportedType, funcName, got), pos)
}

// newRegexError creates a regex error
func newRegexError(err error, pos token.Pos) error {
	return newPosError(fmt.Sprintf(MsgRegexError, err), pos)
}

// newCyclicTypeError creates a cyclic type error
func newCyclicTypeError(typeName string, pos token.Pos) error {
	return newPosError(fmt.Sprintf(MsgCyclicType, typeName), pos)
}

// newUnknownTypeError creates an unknown type error
func newUnknownTypeError(typeName string, pos token.Pos) error {
	return newPosError(fmt.Sprintf(MsgUnknownType, typeName), pos)
}

// newLambdaParamCountError creates a lambda parameter count error
func newLambdaParamCountError(funcName string, expected string, got int, pos token.Pos) error {
	return newPosError(fmt.Sprintf(MsgLambdaParamCount, funcName, expected, got), pos)
}

// newLambdaWrongReturnError creates a lambda wrong return type error
func newLambdaWrongReturnError(funcName string, expected string, got interface{}, pos token.Pos) error {
	return newPosError(fmt.Sprintf(MsgLambdaWrongReturn, funcName, expected, got), pos)
}
