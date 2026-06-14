package evaluator

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	declparser "infomunge/internal/declarations"
	unifiederrors "infomunge/internal/errors"
)

// callBuiltinAssign implements the __assign(varName, value) function.
func callBuiltinAssign(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("assignment requires exactly 2 arguments: variable name and value", e.Pos())
	}

	// First argument must be a string literal (variable name)
	varNameExpr, ok := e.Args[0].(*ast.BasicLit)
	if !ok || varNameExpr.Kind != token.STRING {
		return nil, newPosError("assignment variable name must be a string literal", e.Args[0].Pos())
	}

	varName, err := strconv.Unquote(varNameExpr.Value)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("invalid variable name: %s", err), e.Args[0].Pos())
	}
	if IsReservedBindingName(varName) {
		return nil, newPosError(fmt.Sprintf("%q is reserved for runtime metadata", varName), e.Args[0].Pos())
	}

	// Evaluate the value expression
	value, err := evalASTInScopeWithDepth(e.Args[1], scope, depth)
	if err != nil {
		return nil, err
	}

	// Update the context with the new value
	scope.Vars[varName] = value

	// Return the assigned value
	return value, nil
}

// callBuiltinDo implements the __do(declarations, expression) function.
func callBuiltinDo(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("do block requires exactly 2 arguments: declarations and expression", e.Pos())
	}

	// Both arguments must be string literals
	declsExpr, ok := e.Args[0].(*ast.BasicLit)
	if !ok || declsExpr.Kind != token.STRING {
		return nil, newPosError("do block declarations must be a string literal", e.Args[0].Pos())
	}
	exprExpr, ok := e.Args[1].(*ast.BasicLit)
	if !ok || exprExpr.Kind != token.STRING {
		return nil, newPosError("do block expression must be a string literal", e.Args[1].Pos())
	}

	declsStr, err := strconv.Unquote(declsExpr.Value)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("invalid declarations string: %s", err), e.Args[0].Pos())
	}
	exprStr, err := strconv.Unquote(exprExpr.Value)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("invalid expression string: %s", err), e.Args[1].Pos())
	}

	// Create a new local context by copying the parent context
	localScope := scope.Copy()

	// Parse and evaluate declarations
	if err := parseDoDeclarations(declsStr, localScope, depth); err != nil {
		return nil, newPosError(err.Error(), e.Args[0].Pos())
	}

	parsedExpr, err := compileExpressionInScope(scope, exprStr)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("compile error in do expression: %s", err), e.Args[1].Pos())
	}

	return evalASTInScopeWithDepth(parsedExpr, localScope, depth+1)
}

func bindDoVarDeclaration(decl *declparser.VarDeclaration, scope *Scope, depth int) error {
	if decl == nil {
		return unifiederrors.InternalError("internal error: missing do block variable declaration")
	}
	if IsReservedBindingName(decl.Name) {
		return unifiederrors.EvalErrorf("%q is reserved for runtime metadata", decl.Name)
	}

	parsedExpr, err := compileExpressionInScope(scope, decl.Expression)
	if err != nil {
		return unifiederrors.EvalErrorf("compile error in var declaration '%s': %s", decl.Name, err)
	}
	value, err := evalASTInScopeWithDepth(parsedExpr, scope, depth+1)
	if err != nil {
		return err
	}
	scope.Vars[decl.Name] = value
	return nil
}

func bindDoFunctionDeclaration(decl *declparser.FunctionDeclaration, scope *Scope) error {
	if decl == nil {
		return unifiederrors.InternalError("internal error: missing do block function declaration")
	}
	if IsReservedBindingName(decl.Name) {
		return unifiederrors.EvalErrorf("%q is reserved for runtime metadata", decl.Name)
	}

	parsedBody, err := compileExpressionInScope(scope, decl.Body)
	if err != nil {
		return unifiederrors.EvalErrorf("compile error in fun declaration '%s': %s", decl.Name, err)
	}

	scope.Vars[decl.Name] = &Lambda{
		Params:  paramDeclarationsToParamDefs(decl.Params),
		Body:    decl.Body,
		BodyAST: parsedBody,
		Env:     scope.Vars,
	}
	return nil
}

func paramDeclarationsToParamDefs(params []declparser.ParamDeclaration) []ParamDef {
	defs := make([]ParamDef, 0, len(params))
	for _, param := range params {
		defs = append(defs, ParamDef{
			Name:         param.Name,
			ExpectedKind: KindUnknown,
		})
	}
	return defs
}

// parseDoDeclarations parses and evaluates declarations (var, fun) in a do block
func parseDoDeclarations(declsStr string, scope *Scope, depth int) error {
	lines := strings.Split(declsStr, "\n")
	for i := 0; i < len(lines); {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			i++
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "var "):
			decl, consumed, err := declparser.ParseVarDeclarationFromLines(lines, i)
			if err != nil {
				return err
			}
			if err := bindDoVarDeclaration(decl, scope, depth); err != nil {
				return err
			}
			i += consumed
		case strings.HasPrefix(trimmed, "fun "):
			decl, consumed, err := declparser.ParseFunctionDeclarationFromLines(lines, i)
			if err != nil {
				return err
			}
			if err := bindDoFunctionDeclaration(decl, scope); err != nil {
				return err
			}
			i += consumed
		default:
			return unifiederrors.EvalErrorf("unknown declaration in do block: %s", lines[i])
		}
	}
	return nil
}
