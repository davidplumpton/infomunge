package evaluator

import (
	"fmt"
	"go/ast"
)

// ASTVisitor defines the visitor interface for evaluating AST nodes.
// Each Visit* method corresponds to a specific AST node type.
// All Visit* methods return (result Value, error).
type ASTVisitor interface {
	// VisitBasicLit visits a basic literal (string, number, etc.)
	VisitBasicLit(expr *ast.BasicLit) (Value, error)

	// VisitIdent visits an identifier (variable reference)
	VisitIdent(expr *ast.Ident) (Value, error)

	// VisitBinaryExpr visits a binary expression (a + b, a && b, etc.)
	VisitBinaryExpr(expr *ast.BinaryExpr) (Value, error)

	// VisitUnaryExpr visits a unary expression (e.g., -x, !x)
	VisitUnaryExpr(expr *ast.UnaryExpr) (Value, error)

	// VisitCompositeLit visits a composite literal (map or array)
	VisitCompositeLit(expr *ast.CompositeLit) (Value, error)

	// VisitParenExpr visits a parenthesized expression
	VisitParenExpr(expr *ast.ParenExpr) (Value, error)

	// VisitIndexExpr visits an index expression (a[b])
	VisitIndexExpr(expr *ast.IndexExpr) (Value, error)

	// VisitCallExpr visits a function call expression
	VisitCallExpr(expr *ast.CallExpr) (Value, error)
}

// DefaultVisitor provides a concrete implementation of ASTVisitor.
// It uses recursion depth tracking to prevent stack overflow attacks.
type DefaultVisitor struct {
	scope *Scope
	depth int
}

// NewDefaultVisitorForScope creates a visitor with explicit runtime scope.
func NewDefaultVisitorForScope(scope *Scope, depth int) *DefaultVisitor {
	if scope == nil {
		scope = NewScope(nil)
	}
	scope.ensure()
	return &DefaultVisitor{
		scope: scope,
		depth: depth,
	}
}

// Visit dispatches an AST node to the appropriate Visit* method based on its type.
// This is the entry point for the visitor pattern.
func (v *DefaultVisitor) Visit(expr ast.Expr) (Value, error) {
	if v.depth > MaxEvalDepth {
		return nil, newPosError("expression evaluation depth limit exceeded (possible infinite recursion)", expr.Pos())
	}

	switch e := expr.(type) {
	case *ast.BasicLit:
		return v.VisitBasicLit(e)
	case *ast.Ident:
		return v.VisitIdent(e)
	case *ast.BinaryExpr:
		return v.VisitBinaryExpr(e)
	case *ast.UnaryExpr:
		return v.VisitUnaryExpr(e)
	case *ast.CompositeLit:
		return v.VisitCompositeLit(e)
	case *ast.ParenExpr:
		return v.VisitParenExpr(e)
	case *ast.IndexExpr:
		return v.VisitIndexExpr(e)
	case *ast.CallExpr:
		return v.VisitCallExpr(e)
	}
	return nil, newPosError(fmt.Sprintf("unsupported expression type: %T", expr), expr.Pos())
}

// VisitBasicLit evaluates a basic literal.
func (v *DefaultVisitor) VisitBasicLit(expr *ast.BasicLit) (Value, error) {
	return evalBasicLit(expr)
}

// VisitIdent evaluates an identifier (variable reference).
func (v *DefaultVisitor) VisitIdent(expr *ast.Ident) (Value, error) {
	return evalIdent(expr, v.scope.Vars)
}

// VisitBinaryExpr evaluates a binary expression by evaluating both operands
// and applying the operator.
func (v *DefaultVisitor) VisitBinaryExpr(expr *ast.BinaryExpr) (Value, error) {
	childVisitor := NewDefaultVisitorForScope(v.scope, v.depth+1)

	left, err := childVisitor.Visit(expr.X)
	if err != nil {
		return nil, err
	}

	right, err := childVisitor.Visit(expr.Y)
	if err != nil {
		return nil, err
	}

	return evalBinaryOp(left, right, expr.Op, expr.Pos())
}

// VisitUnaryExpr evaluates a unary expression (e.g., -x, !x).
func (v *DefaultVisitor) VisitUnaryExpr(expr *ast.UnaryExpr) (Value, error) {
	childVisitor := NewDefaultVisitorForScope(v.scope, v.depth+1)

	val, err := childVisitor.Visit(expr.X)
	if err != nil {
		return nil, err
	}

	return evalUnaryOp(val, expr.Op, expr.Pos())
}

// VisitCompositeLit evaluates a composite literal (map or array).
func (v *DefaultVisitor) VisitCompositeLit(expr *ast.CompositeLit) (Value, error) {
	childVisitor := NewDefaultVisitorForScope(v.scope, v.depth+1)

	switch expr.Type.(type) {
	case *ast.MapType:
		res := make(Object)
		for _, elt := range expr.Elts {
			switch node := elt.(type) {
			case *ast.KeyValueExpr:
				key, err := childVisitor.Visit(node.Key)
				if err != nil {
					return nil, err
				}
				val, err := childVisitor.Visit(node.Value)
				if err != nil {
					return nil, err
				}
				keyStr := fmt.Sprintf("%v", key)
				if existing, exists := res[keyStr]; exists {
					switch v := existing.(type) {
					case XMLMultiValue:
						res[keyStr] = append(v, val)
					default:
						res[keyStr] = XMLMultiValue{existing, val}
					}
				} else {
					res[keyStr] = val
				}
			default:
				// Handle dynamic object merging: {(payload)} or similar expressions
				// These must evaluate to Object or Array of Objects
				val, err := childVisitor.Visit(node)
				if err != nil {
					return nil, err
				}

				if obj, ok := val.(Object); ok {
					for k, v := range obj {
						if existing, exists := res[k]; exists {
							// Handle duplicate keys by converting to an array of values
							switch ex := existing.(type) {
							case XMLMultiValue:
								res[k] = append(ex, v)
							default:
								res[k] = XMLMultiValue{existing, v}
							}
						} else {
							res[k] = v
						}
					}
				} else if arr, ok := val.(Array); ok {
					for _, item := range arr {
						if itemObj, ok := item.(Object); ok {
							for k, v := range itemObj {
								if existing, exists := res[k]; exists {
									// Handle duplicate keys by converting to an array of values
									switch ex := existing.(type) {
									case XMLMultiValue:
										res[k] = append(ex, v)
									default:
										res[k] = XMLMultiValue{existing, v}
									}
								} else {
									res[k] = v
								}
							}
						}
					}
				}
			}
		}
		return res, nil
	case *ast.ArrayType:
		res := make(Array, 0, len(expr.Elts))
		for _, elt := range expr.Elts {
			val, err := childVisitor.Visit(elt)
			if err != nil {
				return nil, err
			}
			res = append(res, val)
		}
		return res, nil
	}
	return nil, newPosError(fmt.Sprintf("unsupported composite type: %T", expr.Type), expr.Pos())
}

// VisitParenExpr evaluates a parenthesized expression by visiting its inner expression.
func (v *DefaultVisitor) VisitParenExpr(expr *ast.ParenExpr) (Value, error) {
	childVisitor := NewDefaultVisitorForScope(v.scope, v.depth+1)
	return childVisitor.Visit(expr.X)
}

// VisitIndexExpr evaluates an index expression (a[b]).
func (v *DefaultVisitor) VisitIndexExpr(expr *ast.IndexExpr) (Value, error) {
	childVisitor := NewDefaultVisitorForScope(v.scope, v.depth+1)

	obj, err := childVisitor.Visit(expr.X)
	if err != nil {
		return nil, err
	}
	idx, err := childVisitor.Visit(expr.Index)
	if err != nil {
		return nil, err
	}

	return evalIndex(obj, idx, expr.Pos())
}

// VisitCallExpr evaluates a function call expression.
func (v *DefaultVisitor) VisitCallExpr(expr *ast.CallExpr) (Value, error) {
	// Increase depth for the function call
	childVisitor := NewDefaultVisitorForScope(v.scope, v.depth+1)
	return evalCallExprWithVisitor(expr, childVisitor)
}

// evalCallExprWithVisitor evaluates a call expression using the visitor pattern.
func evalCallExprWithVisitor(e *ast.CallExpr, visitor *DefaultVisitor) (Value, error) {
	fun, ok := e.Fun.(*ast.Ident)
	if !ok {
		return nil, newPosError("only simple function calls are supported", e.Fun.Pos())
	}

	// Handle special functions that need unevaluated arguments
	if handler, ok := GetBuiltinSpecial(fun.Name); ok {
		return handler(e, visitor.scope, visitor.depth)
	}

	// Helper to evaluate arguments using the visitor
	evalArgs := func() ([]Value, error) {
		args := make([]Value, 0, len(e.Args))
		for _, argExpr := range e.Args {
			arg, err := visitor.Visit(argExpr)
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
		return args, nil
	}

	// Dispatch to specific function handler
	if handler, ok := GetBuiltinFunction(fun.Name); ok {
		args, err := evalArgs()
		if err != nil {
			return nil, err
		}
		return handler(args, e)
	}

	// Check if it's a user-defined function in the context
	if userFn, exists := visitor.scope.Vars[fun.Name]; exists {
		if lambda, ok := userFn.(*Lambda); ok {
			args, err := evalArgs()
			if err != nil {
				return nil, err
			}
			return callUserDefinedFunctionWithVisitor(lambda, args, e, visitor.scope, visitor.depth)
		}
	}

	return nil, newPosError(fmt.Sprintf("undefined function: %s", fun.Name), fun.Pos())
}

// callUserDefinedFunctionWithVisitor calls a user-defined function using lexical scope.
func callUserDefinedFunctionWithVisitor(lambda *Lambda, args []Value, e *ast.CallExpr, caller *Scope, depth int) (Value, error) {
	if len(args) != lambda.ParamCount() {
		return nil, newPosError(fmt.Sprintf("function expects %d arguments, got %d", lambda.ParamCount(), len(args)), e.Pos())
	}

	fnScope := newLambdaInvocationScope(lambda, caller)
	for i, param := range lambda.Params {
		fnScope.Vars[param.Name] = args[i]
	}
	return evalASTInScopeWithDepth(lambda.BodyAST, fnScope, depth+1)
}
