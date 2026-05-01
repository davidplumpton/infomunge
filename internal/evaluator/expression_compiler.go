package evaluator

import (
	"go/ast"
	goparser "go/parser"
)

// ExpressionCompiler compiles source text that nested runtime constructs keep
// unevaluated until execution time.
type ExpressionCompiler interface {
	CompileExpression(source string) (ast.Expr, error)
}

type parseExpressionCompiler struct{}

func (parseExpressionCompiler) CompileExpression(source string) (ast.Expr, error) {
	return goparser.ParseExpr(source)
}

func compileExpressionInScope(scope *Scope, source string) (ast.Expr, error) {
	if scope == nil {
		scope = NewScope(nil)
	}
	return scope.ExpressionCompiler().CompileExpression(source)
}
