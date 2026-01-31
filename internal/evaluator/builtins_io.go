package evaluator

import (
	"fmt"
	"go/ast"
	"infomunge/pkg/formats"
	"io"
	"net/http"
)

// callBuiltinRead implements the read(content, mimeType) function.
func callBuiltinRead(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) < 2 {
		return nil, newPosError("read function requires at least 2 arguments: content and mimeType", e.Pos())
	}
	content, contentIsString := args[0].(string)
	mimeType, mimeTypeIsString := args[1].(string)
	if !contentIsString || !mimeTypeIsString {
		return nil, newPosError("read function arguments must be strings", e.Pos())
	}
	res, err := formats.Read(content, mimeType)
	if err != nil {
		return nil, newPosError(err.Error(), e.Pos())
	}
	return res, nil
}

// callBuiltinReadUrl implements the readUrl(url, mimeType) function.
func callBuiltinReadUrl(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 2 {
		return nil, newPosError("readUrl requires exactly 2 arguments: url and mimeType", e.Pos())
	}

	url, ok := args[0].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("readUrl expects url to be a string, got %T", args[0]), e.Pos())
	}

	mimeType, ok := args[1].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("readUrl expects mimeType to be a string, got %T", args[1]), e.Pos())
	}

	// Fetch the URL
	resp, err := http.Get(url)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("readUrl: failed to fetch URL: %v", err), e.Pos())
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("readUrl: failed to read response: %v", err), e.Pos())
	}

	// Parse the content
	result, err := formats.Read(string(body), mimeType)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("readUrl: failed to parse content: %v", err), e.Pos())
	}

	return result, nil
}

// callBuiltinWrite implements the write(value, mimeType) function.
func callBuiltinWrite(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) != 2 {
		return nil, newPosError("write requires exactly 2 arguments: value and mimeType", e.Pos())
	}

	mimeType, ok := args[1].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("write expects mimeType to be a string, got %T", args[1]), e.Pos())
	}

	result, err := formats.Format(args[0], mimeType)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("write error: %v", err), e.Pos())
	}

	return result, nil
}
