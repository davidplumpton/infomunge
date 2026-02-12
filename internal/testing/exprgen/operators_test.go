package exprgen_test

import (
	"fmt"
	"strings"
	"testing"

	"infomunge/internal/evaluator"
	"infomunge/internal/preprocessor"
	"infomunge/internal/testing/exprgen"

	"pgregory.net/rapid"
)

// evalExpr evaluates an infomunge expression by running it through the
// preprocessor first (to handle array/object literals), then evaluating.
func evalExpr(expr string) (interface{}, error) {
	prepared, mapping, err := preprocessor.PrepareForParsing(expr, preprocessor.Options{})
	if err != nil {
		return nil, err
	}
	return evaluator.Evaluate(prepared, make(map[string]interface{}), mapping, 0, expr)
}

// --- Operator generators ---

func TestBinaryOp_ValidOps(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		op := exprgen.BinaryOp().Draw(t, "op")
		valid := map[string]bool{
			"+": true, "-": true, "*": true, "/": true,
			"%": true, "++": true, "**": true,
			"==": true, "!=": true, "<": true, ">": true, "<=": true, ">=": true, "~=": true,
			"&&": true, "||": true,
		}
		if !valid[op] {
			t.Fatalf("BinaryOp produced unexpected operator: %q", op)
		}
	})
}

func TestBinaryOp_WeightedDistribution(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		counts := map[string]int{}
		for i := 0; i < 4000; i++ {
			op := exprgen.BinaryOp().Draw(t, fmt.Sprintf("op%d", i))
			counts[op]++
		}
		if counts["+"] <= counts["**"] {
			t.Fatalf("expected + to appear more often than **, got +=%d **=%d", counts["+"], counts["**"])
		}
		if counts["=="] <= counts["~="] {
			t.Fatalf("expected == to appear more often than ~=, got ==%d ~=%d", counts["=="], counts["~="])
		}
		if counts["&&"] <= counts["~="] {
			t.Fatalf("expected && to appear more often than ~=, got &&=%d ~=%d", counts["&&"], counts["~="])
		}
	})
}

func TestUnaryOp_ValidOps(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		op := exprgen.UnaryOp().Draw(t, "op")
		if op != "!" && op != "-" {
			t.Fatalf("UnaryOp produced unexpected operator: %q", op)
		}
	})
}

// --- Expr at depth 0 should produce only literals ---

func TestExpr_Depth0_IsLiteral(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		expr := exprgen.Expr(0).Draw(t, "expr")
		// Depth 0 produces literals which should evaluate cleanly.
		if _, err := evalExpr(expr); err != nil {
			t.Fatalf("Expr(0) %q evaluation failed: %v", expr, err)
		}
	})
}

// --- Expr at various depths: no panics (evaluation validity) ---
// Expressions may produce evaluation errors (division by zero, type mismatches)
// but must never panic and must produce balanced brackets.

func TestExpr_Depth1_DoesNotPanic(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		expr := exprgen.Expr(1).Draw(t, "expr")
		evalExpr(expr) //nolint:errcheck
	})
}

func TestExpr_Depth2_DoesNotPanic(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		expr := exprgen.Expr(2).Draw(t, "expr")
		evalExpr(expr) //nolint:errcheck
	})
}

func TestExpr_Depth3_DoesNotPanic(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		expr := exprgen.Expr(3).Draw(t, "expr")
		evalExpr(expr) //nolint:errcheck
	})
}

func TestExpr_Depth4_DoesNotPanic(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		expr := exprgen.Expr(4).Draw(t, "expr")
		evalExpr(expr) //nolint:errcheck
	})
}

// --- Balanced brackets property ---

func TestExpr_BalancedBrackets(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		expr := exprgen.Expr(3).Draw(t, "expr")
		parens := 0
		brackets := 0
		inString := false
		escaped := false
		for _, ch := range expr {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' && inString {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = !inString
				continue
			}
			if inString {
				continue
			}
			switch ch {
			case '(':
				parens++
			case ')':
				parens--
			case '[':
				brackets++
			case ']':
				brackets--
			}
			if parens < 0 || brackets < 0 {
				t.Fatalf("Expr %q has unbalanced brackets (parens=%d, brackets=%d)", expr, parens, brackets)
			}
		}
		if parens != 0 || brackets != 0 {
			t.Fatalf("Expr %q has unbalanced brackets (parens=%d, brackets=%d)", expr, parens, brackets)
		}
	})
}

// --- Expr produces variety (not just literals) at depth >= 2 ---

func TestExpr_ProducesVariety(t *testing.T) {
	sawBinaryOp := false
	sawUnary := false
	sawArray := false
	sawParen := false

	for i := 0; i < 1000; i++ {
		var expr string
		rapid.Check(t, func(t *rapid.T) {
			expr = exprgen.Expr(3).Draw(t, "expr")
		})
		for _, op := range []string{" + ", " - ", " * ", " / ", " % ", " ++ ", " ** ", " == ", " != ", " < ", " > ", " <= ", " >= ", " ~= ", " && ", " || "} {
			if strings.Contains(expr, op) {
				sawBinaryOp = true
				break
			}
		}
		if strings.HasPrefix(expr, "!(") || strings.HasPrefix(expr, "-(") {
			sawUnary = true
		}
		if strings.HasPrefix(expr, "[") {
			sawArray = true
		}
		if strings.HasPrefix(expr, "(") && !strings.HasPrefix(expr, "-(") && !strings.HasPrefix(expr, "!(") {
			sawParen = true
		}
		if sawBinaryOp && sawUnary && sawArray && sawParen {
			break
		}
	}

	if !sawBinaryOp {
		t.Error("1000 draws at depth 3 never produced a binary operation")
	}
	if !sawUnary {
		t.Error("1000 draws at depth 3 never produced a unary expression")
	}
	if !sawArray {
		t.Error("1000 draws at depth 3 never produced an array literal")
	}
	if !sawParen {
		t.Error("1000 draws at depth 3 never produced a parenthesized expression")
	}
}

// --- Known examples: evaluate through infomunge ---

func TestExpr_KnownExamples(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected interface{}
	}{
		{"simple add", "1 + 2", 3},
		{"nested ops", "(1 + 2) * 3", 9},
		{"comparison", "1 == 2", false},
		{"logical", "true && false", false},
		{"negation", "-(1)", -1},
		{"not", "!(true)", false},
		{"empty array", "[]", []interface{}{}},
		{"complex", "(1 + 2) * -(3) > 0 && !(false)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evalExpr(tt.expr)
			if err != nil {
				t.Fatalf("eval %q: %v", tt.expr, err)
			}
			// For slices, just check type (deep equality is complex for interface{})
			if _, ok := tt.expected.([]interface{}); ok {
				if _, ok2 := result.([]interface{}); !ok2 {
					t.Fatalf("eval %q = %v (%T), want slice", tt.expr, result, result)
				}
				return
			}
			if result != tt.expected {
				t.Errorf("eval %q = %v (%T), want %v (%T)", tt.expr, result, result, tt.expected, tt.expected)
			}
		})
	}
}
