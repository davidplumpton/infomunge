package exprgen_test

import (
	"regexp"
	"strings"
	"testing"

	"infomunge/internal/preprocessor"
	"infomunge/internal/testing/exprgen"

	"pgregory.net/rapid"
)

var objectFieldPattern = regexp.MustCompile(`"[A-Za-z_][A-Za-z0-9_]*"\s*:`)

// --- Bounded depth: generated expressions are reasonable size ---

func TestExpression_BoundedDepth(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		depth := rapid.IntRange(0, 4).Draw(t, "depth")
		expr := exprgen.Expression(depth, exprgen.FeatureAll).Draw(t, "expr")
		// Sanity: expressions should be non-empty.
		if len(expr) == 0 {
			t.Fatal("Expression generated empty string")
		}
	})
}

// --- Syntactically valid: pass through preprocessor without panic ---

func TestExpression_SyntacticallyValid(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		expr := exprgen.Expression(3, exprgen.FeatureAll).Draw(t, "expr")
		// Must not panic. Errors are acceptable (type mismatches at eval time),
		// but preprocessing should not panic.
		preprocessor.PrepareForParsing(expr, preprocessor.Options{})
	})
}

func TestExpression_WrappedSyntacticallyValid(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		expr := exprgen.Expression(2, exprgen.FeatureAll).Draw(t, "expr")
		script := exprgen.WrapScript(expr)
		if !strings.HasPrefix(script, "%im 0.1\n") {
			t.Fatalf("WrapScript did not produce expected header, got: %q", script[:40])
		}
	})
}

// --- Feature subset restricts output ---

func TestExpression_ArithmeticOnly(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		expr := exprgen.Expression(3, exprgen.FeatureArithmetic).Draw(t, "expr")
		// Should not contain comparison or logical operators.
		for _, op := range []string{" == ", " != ", " && ", " || "} {
			if strings.Contains(expr, op) {
				t.Fatalf("FeatureArithmetic expression %q contains forbidden operator %q", expr, op)
			}
		}
		// Should not contain array brackets (outside of strings).
		if containsArrayLiteral(expr) {
			t.Fatalf("FeatureArithmetic expression %q contains array literal", expr)
		}
	})
}

func TestExpression_ComparisonOnly(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		expr := exprgen.Expression(3, exprgen.FeatureComparison).Draw(t, "expr")
		// Should not contain arithmetic ops (except - which is also unary in literals).
		for _, op := range []string{" + ", " * ", " / "} {
			if strings.Contains(expr, op) {
				t.Fatalf("FeatureComparison expression %q contains forbidden operator %q", expr, op)
			}
		}
		// Should not contain logical operators.
		for _, op := range []string{" && ", " || "} {
			if strings.Contains(expr, op) {
				t.Fatalf("FeatureComparison expression %q contains forbidden operator %q", expr, op)
			}
		}
	})
}

func TestExpression_NoArraysWithoutFeature(t *testing.T) {
	features := exprgen.FeatureArithmetic | exprgen.FeatureComparison | exprgen.FeatureLogical | exprgen.FeatureObjects | exprgen.FeatureUnary | exprgen.FeatureParens
	rapid.Check(t, func(t *rapid.T) {
		expr := exprgen.Expression(3, features).Draw(t, "expr")
		if containsArrayLiteral(expr) {
			t.Fatalf("Expression without FeatureArrays %q contains array literal", expr)
		}
	})
}

// --- WrapScript ---

func TestWrapScript(t *testing.T) {
	script := exprgen.WrapScript("1 + 2")
	expected := "%im 0.1\noutput application/json\n---\n1 + 2"
	if script != expected {
		t.Fatalf("WrapScript mismatch:\n  got:  %q\n  want: %q", script, expected)
	}
}

// --- IsValid ---

func TestIsValid_GoodExpr(t *testing.T) {
	if !exprgen.IsValid("1 + 2") {
		t.Fatal("IsValid returned false for valid expression")
	}
}

// --- Feature variety: enabled features do produce their forms ---

func TestExpression_ArithmeticProducesOps(t *testing.T) {
	sawOp := false
	for i := 0; i < 500; i++ {
		var expr string
		rapid.Check(t, func(t *rapid.T) {
			expr = exprgen.Expression(3, exprgen.FeatureArithmetic).Draw(t, "expr")
		})
		for _, op := range []string{" + ", " - ", " * ", " / "} {
			if strings.Contains(expr, op) {
				sawOp = true
				break
			}
		}
		if sawOp {
			break
		}
	}
	if !sawOp {
		t.Error("500 draws with FeatureArithmetic never produced an arithmetic operator")
	}
}

func TestExpression_ArraysProducesBrackets(t *testing.T) {
	sawArray := false
	for i := 0; i < 500; i++ {
		var expr string
		rapid.Check(t, func(t *rapid.T) {
			expr = exprgen.Expression(3, exprgen.FeatureArrays).Draw(t, "expr")
		})
		if containsArrayLiteral(expr) {
			sawArray = true
			break
		}
	}
	if !sawArray {
		t.Error("500 draws with FeatureArrays never produced an array literal")
	}
}

func TestExpression_ObjectsProduceBraces(t *testing.T) {
	sawObject := false
	for i := 0; i < 500; i++ {
		var expr string
		rapid.Check(t, func(t *rapid.T) {
			expr = exprgen.Expression(3, exprgen.FeatureObjects).Draw(t, "expr")
		})
		if strings.HasPrefix(strings.TrimSpace(expr), "{") {
			if strings.TrimSpace(expr) != "{}" && !objectFieldPattern.MatchString(expr) {
				t.Fatalf("object literal fields must be valid identifier-like keys, got %q", expr)
			}
			sawObject = true
			break
		}
	}
	if !sawObject {
		t.Error("500 draws with FeatureObjects never produced an object literal")
	}
}

func TestExpression_DotAccessProducesFieldSelectors(t *testing.T) {
	sawDot := false
	for i := 0; i < 500; i++ {
		var expr string
		rapid.Check(t, func(t *rapid.T) {
			expr = exprgen.Expression(3, exprgen.FeatureDotAccess).Draw(t, "expr")
		})
		if !exprgen.IsValid(expr) {
			t.Fatalf("FeatureDotAccess produced syntactically invalid expression: %q", expr)
		}
		if strings.Contains(expr, "payload.") {
			sawDot = true
			break
		}
	}
	if !sawDot {
		t.Fatal("500 draws with FeatureDotAccess never produced payload field access")
	}
}

func TestExpression_IndexAccessSyntacticallyValid(t *testing.T) {
	sawIndex := false
	for i := 0; i < 500; i++ {
		var expr string
		rapid.Check(t, func(t *rapid.T) {
			expr = exprgen.Expression(3, exprgen.FeatureIndexAccess).Draw(t, "expr")
		})
		if !exprgen.IsValid(expr) {
			t.Fatalf("FeatureIndexAccess produced syntactically invalid expression: %q", expr)
		}
		if strings.Contains(expr, "[") && strings.Contains(expr, "]") {
			sawIndex = true
			break
		}
	}
	if !sawIndex {
		t.Fatal("500 draws with FeatureIndexAccess never produced index brackets")
	}
}

func TestExpression_RangeIndexSyntacticallyValid(t *testing.T) {
	sawRange := false
	for i := 0; i < 500; i++ {
		var expr string
		rapid.Check(t, func(t *rapid.T) {
			expr = exprgen.Expression(3, exprgen.FeatureRangeIndex).Draw(t, "expr")
		})
		if !exprgen.IsValid(expr) {
			t.Fatalf("FeatureRangeIndex produced syntactically invalid expression: %q", expr)
		}
		if strings.Contains(expr, " to ") {
			sawRange = true
			break
		}
	}
	if !sawRange {
		t.Fatal("500 draws with FeatureRangeIndex never produced a range index")
	}
}

// containsArrayLiteral checks if the expression contains [ or ] outside of strings.
func containsArrayLiteral(expr string) bool {
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
		if ch == '[' || ch == ']' {
			return true
		}
	}
	return false
}
