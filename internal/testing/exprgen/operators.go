package exprgen

import (
	"fmt"
	"strings"

	"pgregory.net/rapid"
)

// binaryOps is the minimal set of binary operators for phase-1 generation.
var binaryOps = []string{"+", "-", "*", "/", "==", "!=", "<", ">", "&&", "||"}

// BinaryOp returns a generator that picks a random binary operator.
func BinaryOp() *rapid.Generator[string] {
	return rapid.SampledFrom(binaryOps)
}

// UnaryOp returns a generator that picks a random unary operator (! or -).
func UnaryOp() *rapid.Generator[string] {
	return rapid.SampledFrom([]string{"!", "-"})
}

// Expr returns a generator of syntactically valid infomunge expression strings
// with a maximum recursion depth. At depth 0, only literals are produced.
// At higher depths, binary/unary operations, parenthesized expressions, and
// array literals may be generated.
func Expr(maxDepth int) *rapid.Generator[string] {
	return exprAtDepth(maxDepth)
}

func exprAtDepth(depth int) *rapid.Generator[string] {
	if depth <= 0 {
		return Literal()
	}
	return rapid.Custom(func(t *rapid.T) string {
		// Weight literals more heavily to keep expressions manageable.
		// 0-3: literal (40%), 4-5: binary (20%), 6: unary (10%),
		// 7: parens (10%), 8-9: array (20%)
		kind := rapid.IntRange(0, 9).Draw(t, "exprKind")
		switch {
		case kind <= 3:
			return Literal().Draw(t, "lit")
		case kind <= 5:
			return binaryExpr(depth).Draw(t, "binary")
		case kind == 6:
			return unaryExpr(depth).Draw(t, "unary")
		case kind == 7:
			return parenExpr(depth).Draw(t, "paren")
		default:
			return arrayExpr(depth).Draw(t, "array")
		}
	})
}

// binaryExpr generates "left op right" with children at depth-1.
func binaryExpr(depth int) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		left := exprAtDepth(depth - 1).Draw(t, "left")
		op := BinaryOp().Draw(t, "op")
		right := exprAtDepth(depth - 1).Draw(t, "right")
		return fmt.Sprintf("%s %s %s", left, op, right)
	})
}

// unaryExpr generates "op(expr)" — the operand is always parenthesized to
// avoid ambiguous parses like "!-3" or "--x".
func unaryExpr(depth int) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		op := UnaryOp().Draw(t, "op")
		operand := exprAtDepth(depth - 1).Draw(t, "operand")
		return fmt.Sprintf("%s(%s)", op, operand)
	})
}

// parenExpr wraps a sub-expression in parentheses.
func parenExpr(depth int) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		inner := exprAtDepth(depth - 1).Draw(t, "inner")
		return fmt.Sprintf("(%s)", inner)
	})
}

// arrayExpr generates an array literal with 0-5 elements.
func arrayExpr(depth int) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		n := rapid.IntRange(0, 5).Draw(t, "len")
		if n == 0 {
			return "[]"
		}
		elems := make([]string, n)
		for i := range elems {
			elems[i] = exprAtDepth(depth - 1).Draw(t, fmt.Sprintf("elem%d", i))
		}
		return "[" + strings.Join(elems, ", ") + "]"
	})
}
