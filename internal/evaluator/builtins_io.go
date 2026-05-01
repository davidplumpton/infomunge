package evaluator

import (
	"fmt"
	"go/ast"
	"net/url"
)

func evalBuiltinArgs(argExprs []ast.Expr, scope *Scope, depth int) ([]Value, error) {
	args := make(Array, 0, len(argExprs))
	for _, argExpr := range argExprs {
		arg, err := evalASTInScopeWithDepth(argExpr, scope, depth+1)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	return args, nil
}

// callBuiltinRead implements the read(content, mimeType[, options]) function.
func callBuiltinRead(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	args, err := evalBuiltinArgs(e.Args, scope, depth)
	if err != nil {
		return nil, err
	}
	return callBuiltinReadWithArgs(args, e, scope)
}

func callBuiltinReadWithArgs(args []Value, e *ast.CallExpr, scope *Scope) (Value, error) {
	if len(args) < 2 {
		return nil, newPosError("read function requires at least 2 arguments: content and mimeType", e.Pos())
	}
	if len(args) > 3 {
		return nil, newPosError("read function accepts at most 3 arguments: content, mimeType, and optional options object", e.Pos())
	}
	content, contentIsString := args[0].(string)
	mimeType, mimeTypeIsString := args[1].(string)
	if !contentIsString || !mimeTypeIsString {
		return nil, newPosError("read function arguments must be strings", e.Pos())
	}

	formatService, err := requireFormatService(scope)
	if err != nil {
		return nil, newPosError(err.Error(), e.Pos())
	}

	if len(args) == 2 {
		res, err := formatService.Read(content, mimeType)
		if err != nil {
			return nil, newPosError(err.Error(), e.Pos())
		}
		return res, nil
	}

	options, ok := args[2].(Object)
	if !ok {
		return nil, newPosError(fmt.Sprintf("read expects options to be an object, got %T", args[2]), e.Pos())
	}

	res, err := formatService.ReadWithOptions(content, mimeType, options)
	if err != nil {
		return nil, newPosError(err.Error(), e.Pos())
	}
	return res, nil
}

// callBuiltinReadUrl implements the readUrl(url, mimeType) function.
func callBuiltinReadUrl(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	args, err := evalBuiltinArgs(e.Args, scope, depth)
	if err != nil {
		return nil, err
	}
	return callBuiltinReadUrlWithArgs(args, e, scope)
}

func callBuiltinReadUrlWithArgs(args []Value, e *ast.CallExpr, scope *Scope) (Value, error) {
	if len(args) != 2 {
		return nil, newPosError("readUrl requires exactly 2 arguments: url and mimeType", e.Pos())
	}

	rawURL, ok := args[0].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("readUrl expects url to be a string, got %T", args[0]), e.Pos())
	}

	mimeType, ok := args[1].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("readUrl expects mimeType to be a string, got %T", args[1]), e.Pos())
	}

	// Validate URL before fetching
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("readUrl: invalid URL: %v", err), e.Pos())
	}
	if parsed.Scheme == "" && parsed.Host == "" {
		return nil, newPosError("readUrl: invalid URL: missing scheme or host", e.Pos())
	}

	urlReader, err := requireURLReadService(scope)
	if err != nil {
		return nil, newPosError(err.Error(), e.Pos())
	}

	result, err := urlReader.ReadURL(scope.GoContext(), rawURL, mimeType)
	if err != nil {
		return nil, newPosError(err.Error(), e.Pos())
	}

	return result, nil
}

// callBuiltinWrite implements the write(value, mimeType[, options]) function.
func callBuiltinWrite(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	args, err := evalBuiltinArgs(e.Args, scope, depth)
	if err != nil {
		return nil, err
	}
	return callBuiltinWriteWithArgs(args, e, scope)
}

func callBuiltinWriteWithArgs(args []Value, e *ast.CallExpr, scope *Scope) (Value, error) {
	if len(args) < 2 {
		return nil, newPosError("write requires exactly 2 arguments: value and mimeType", e.Pos())
	}
	if len(args) > 3 {
		return nil, newPosError("write accepts at most 3 arguments: value, mimeType, and optional options object", e.Pos())
	}

	mimeType, ok := args[1].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("write expects mimeType to be a string, got %T", args[1]), e.Pos())
	}

	formatService, err := requireFormatService(scope)
	if err != nil {
		return nil, newPosError(err.Error(), e.Pos())
	}

	if len(args) == 2 {
		result, err := formatService.Write(args[0], mimeType)
		if err != nil {
			return nil, newPosError(fmt.Sprintf("write error: %v", err), e.Pos())
		}
		return result, nil
	}

	options, ok := args[2].(Object)
	if !ok {
		return nil, newPosError(fmt.Sprintf("write expects options to be an object, got %T", args[2]), e.Pos())
	}

	result, err := formatService.WriteWithOptions(args[0], mimeType, options)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("write error: %v", err), e.Pos())
	}

	return result, nil
}
