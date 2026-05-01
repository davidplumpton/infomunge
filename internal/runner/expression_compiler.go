package runner

import (
	"go/ast"
	goparser "go/parser"
	"strings"

	"infomunge/internal/preprocessor"
)

type expressionCompiler struct{}

func (expressionCompiler) CompileExpression(source string) (ast.Expr, error) {
	parseable, _, err := prepareExpressionForParsing(source)
	if err != nil {
		return nil, err
	}
	return goparser.ParseExpr(parseable)
}

func prepareExpressionForParsing(source string) (string, []int, error) {
	return preprocessor.PrepareForParsing(source, preprocessingOptionsForExpression(source))
}

func preprocessingOptionsForExpression(source string) preprocessor.Options {
	opts := preprocessor.Options{}
	if strings.ContainsAny(source, "\n\r") {
		opts.AllowMultilineIfElse = true
	}
	return opts
}
