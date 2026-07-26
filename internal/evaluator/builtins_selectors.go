package evaluator

import (
	"fmt"
	"go/ast"
)

type selectorMode uint8

const (
	selectorModePresence selectorMode = iota
	selectorModeAssert
)

type selectorOperation struct {
	mode selectorMode
	key  string
}

func callBuiltinPresenceSelector(args []Value, e *ast.CallExpr) (Value, error) {
	return newSelectorOperation(args, selectorModePresence, e)
}

func callBuiltinAssertSelector(args []Value, e *ast.CallExpr) (Value, error) {
	return newSelectorOperation(args, selectorModeAssert, e)
}

func newSelectorOperation(args []Value, mode selectorMode, e *ast.CallExpr) (Value, error) {
	key, ok := args[0].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("selector name must be a string, got %T", args[0]), e.Pos())
	}
	return selectorOperation{mode: mode, key: key}, nil
}
