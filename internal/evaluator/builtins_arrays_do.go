package evaluator

import (
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"strconv"
	"strings"

	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/preprocessor"
)

// callBuiltinAssign implements the __assign(varName, value) function.
func callBuiltinAssign(e *ast.CallExpr, context map[string]interface{}, depth int) (interface{}, error) {
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

	// Evaluate the value expression
	value, err := evalASTWithDepth(e.Args[1], context, depth)
	if err != nil {
		return nil, err
	}

	// Update the context with the new value
	context[varName] = value

	// Return the assigned value
	return value, nil
}

// callBuiltinDo implements the __do(declarations, expression) function.
func callBuiltinDo(e *ast.CallExpr, context map[string]interface{}, depth int) (interface{}, error) {
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
	localContext := copyContext(context)

	// Parse and evaluate declarations
	if err := parseDoDeclarations(declsStr, localContext, depth); err != nil {
		return nil, newPosError(err.Error(), e.Args[0].Pos())
	}

	// Preprocess and evaluate the expression in the local context
	preparedExpr, _, err := preprocessor.PrepareForParsing(exprStr, preprocessor.Options{})
	if err != nil {
		return nil, newPosError(fmt.Sprintf("preprocessing error in do expression: %s", err), e.Args[1].Pos())
	}
	parsedExpr, err := goparser.ParseExpr(preparedExpr)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("parse error in do expression: %s", err), e.Args[1].Pos())
	}

	return evalASTWithDepth(parsedExpr, localContext, depth+1)
}

// parseVarDeclaration parses a var declaration and adds it to the context.
// Format: var name = expression
func parseVarDeclaration(line string, context map[string]interface{}, depth int) error {
	rest := strings.TrimPrefix(line, "var ")
	parts := strings.SplitN(rest, "=", 2)
	if len(parts) != 2 {
		return unifiederrors.EvalErrorf("invalid var declaration: %s", line)
	}
	varName := strings.TrimSpace(parts[0])
	exprStr := strings.TrimSpace(parts[1])

	preparedExpr, _, err := preprocessor.PrepareForParsing(exprStr, preprocessor.Options{})
	if err != nil {
		return unifiederrors.EvalErrorf("preprocessing error in var declaration '%s': %s", varName, err)
	}
	parsedExpr, err := goparser.ParseExpr(preparedExpr)
	if err != nil {
		return unifiederrors.EvalErrorf("parse error in var declaration '%s': %s", varName, err)
	}
	value, err := evalASTWithDepth(parsedExpr, context, depth+1)
	if err != nil {
		return err
	}
	context[varName] = value
	return nil
}

// parseFunDeclaration parses a fun declaration and adds it to the context.
// Format: fun name(params) = body
func parseFunDeclaration(line string, context map[string]interface{}) error {
	rest := strings.TrimPrefix(line, "fun ")

	parenIdx := strings.Index(rest, "(")
	if parenIdx == -1 {
		return unifiederrors.EvalErrorf("invalid fun declaration (missing parentheses): %s", line)
	}
	funName := strings.TrimSpace(rest[:parenIdx])

	closeParenIdx := strings.Index(rest, ")")
	if closeParenIdx == -1 {
		return unifiederrors.EvalErrorf("invalid fun declaration (unclosed parentheses): %s", line)
	}

	params := parseFunParams(rest[parenIdx+1 : closeParenIdx])

	eqIdx := strings.Index(rest[closeParenIdx:], "=")
	if eqIdx == -1 {
		return unifiederrors.EvalErrorf("invalid fun declaration (missing '='): %s", line)
	}
	bodyStr := strings.TrimSpace(rest[closeParenIdx+eqIdx+1:])

	preparedBody, _, err := preprocessor.PrepareForParsing(bodyStr, preprocessor.Options{})
	if err != nil {
		return unifiederrors.EvalErrorf("preprocessing error in fun declaration '%s': %s", funName, err)
	}
	parsedBody, err := goparser.ParseExpr(preparedBody)
	if err != nil {
		return unifiederrors.EvalErrorf("parse error in fun declaration '%s': %s", funName, err)
	}

	context[funName] = &Lambda{
		Params:  params,
		Body:    bodyStr,
		BodyAST: parsedBody,
		Env:     context,
	}
	return nil
}

// parseFunParams parses function parameters from a comma-separated string.
func parseFunParams(paramsStr string) []ParamDef {
	var params []ParamDef
	if strings.TrimSpace(paramsStr) == "" {
		return params
	}
	for _, p := range strings.Split(paramsStr, ",") {
		params = append(params, ParamDef{
			Name:         strings.TrimSpace(p),
			ExpectedKind: KindUnknown,
		})
	}
	return params
}

// parseDoDeclarations parses and evaluates declarations (var, fun) in a do block
func parseDoDeclarations(declsStr string, context map[string]interface{}, depth int) error {
	for _, line := range strings.Split(declsStr, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "var "):
			if err := parseVarDeclaration(trimmed, context, depth); err != nil {
				return err
			}
		case strings.HasPrefix(trimmed, "fun "):
			if err := parseFunDeclaration(trimmed, context); err != nil {
				return err
			}
		default:
			return unifiederrors.EvalErrorf("unknown declaration in do block: %s", line)
		}
	}
	return nil
}
