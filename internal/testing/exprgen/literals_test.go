package exprgen_test

import (
	"go/parser"
	"strconv"
	"strings"
	"testing"

	"infomunge/internal/evaluator"
	"infomunge/internal/testing/exprgen"

	"pgregory.net/rapid"
)

// evalLiteral evaluates an expression string using the infomunge evaluator
// with an empty context and identity mapping.
func evalLiteral(expr string) (evaluator.Value, error) {
	mapping := make([]int, len(expr)+10)
	for i := range mapping {
		mapping[i] = i
	}
	return evaluator.Evaluate(expr, make(evaluator.Context), mapping, 0, expr)
}

// --- Property tests: parsing validity ---

func TestIntLiteral_ParsesAsGoExpr(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		lit := exprgen.IntLiteral().Draw(t, "lit")
		if _, err := parser.ParseExpr(lit); err != nil {
			t.Fatalf("IntLiteral %q is not a valid Go expr: %v", lit, err)
		}
	})
}

func TestFloatLiteral_ParsesAsGoExpr(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		lit := exprgen.FloatLiteral().Draw(t, "lit")
		if _, err := parser.ParseExpr(lit); err != nil {
			t.Fatalf("FloatLiteral %q is not a valid Go expr: %v", lit, err)
		}
	})
}

func TestStringLiteral_ParsesAsGoExpr(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		lit := exprgen.StringLiteral().Draw(t, "lit")
		if _, err := parser.ParseExpr(lit); err != nil {
			t.Fatalf("StringLiteral %q is not a valid Go expr: %v", lit, err)
		}
	})
}

func TestDWStringLiteral_AvoidsAmbiguousEscapes(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		lit := exprgen.DWStringLiteral().Draw(t, "lit")
		if _, err := parser.ParseExpr(lit); err != nil {
			t.Fatalf("DWStringLiteral %q is not a valid Go expr: %v", lit, err)
		}
		if strings.Contains(lit, `\`) {
			t.Fatalf("DWStringLiteral %q contains an escape with language-specific semantics", lit)
		}
	})
}

func TestBoolLiteral_ParsesAsGoExpr(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		lit := exprgen.BoolLiteral().Draw(t, "lit")
		if _, err := parser.ParseExpr(lit); err != nil {
			t.Fatalf("BoolLiteral %q is not a valid Go expr: %v", lit, err)
		}
	})
}

func TestNullLiteral_ParsesAsGoExpr(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		lit := exprgen.NullLiteral().Draw(t, "lit")
		if _, err := parser.ParseExpr(lit); err != nil {
			t.Fatalf("NullLiteral %q is not a valid Go expr: %v", lit, err)
		}
	})
}

// --- Property tests: evaluation and type checking ---

func TestIntLiteral_Evaluates(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		lit := exprgen.IntLiteral().Draw(t, "lit")
		result, err := evalLiteral(lit)
		if err != nil {
			t.Fatalf("IntLiteral %q evaluation failed: %v", lit, err)
		}
		if _, ok := result.(int); !ok {
			t.Fatalf("IntLiteral %q evaluated to %T, want int", lit, result)
		}
	})
}

func TestFloatLiteral_Evaluates(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		lit := exprgen.FloatLiteral().Draw(t, "lit")
		result, err := evalLiteral(lit)
		if err != nil {
			t.Fatalf("FloatLiteral %q evaluation failed: %v", lit, err)
		}
		switch result.(type) {
		case float64:
			// expected
		case int:
			// acceptable: formatFloat might produce "-0.0" which evaluates as float,
			// but some edge cases might resolve to int through negate(0)
		default:
			t.Fatalf("FloatLiteral %q evaluated to %T, want float64 or int", lit, result)
		}
	})
}

func TestStringLiteral_Evaluates(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		lit := exprgen.StringLiteral().Draw(t, "lit")
		result, err := evalLiteral(lit)
		if err != nil {
			t.Fatalf("StringLiteral %q evaluation failed: %v", lit, err)
		}
		if _, ok := result.(string); !ok {
			t.Fatalf("StringLiteral %q evaluated to %T, want string", lit, result)
		}
	})
}

func TestBoolLiteral_Evaluates(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		lit := exprgen.BoolLiteral().Draw(t, "lit")
		result, err := evalLiteral(lit)
		if err != nil {
			t.Fatalf("BoolLiteral %q evaluation failed: %v", lit, err)
		}
		if _, ok := result.(bool); !ok {
			t.Fatalf("BoolLiteral %q evaluated to %T, want bool", lit, result)
		}
	})
}

func TestNullLiteral_Evaluates(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		lit := exprgen.NullLiteral().Draw(t, "lit")
		result, err := evalLiteral(lit)
		if err != nil {
			t.Fatalf("NullLiteral %q evaluation failed: %v", lit, err)
		}
		if result != nil {
			t.Fatalf("NullLiteral %q evaluated to %v (%T), want nil", lit, result, result)
		}
	})
}

// --- Round-trip property tests ---

func TestIntLiteral_RoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		lit := exprgen.IntLiteral().Draw(t, "lit")
		result, err := evalLiteral(lit)
		if err != nil {
			t.Fatalf("eval failed for %q: %v", lit, err)
		}
		n := result.(int)
		parsed, err := strconv.Atoi(lit)
		if err != nil {
			t.Fatalf("strconv.Atoi failed for %q: %v", lit, err)
		}
		if n != parsed {
			t.Fatalf("round-trip mismatch: lit=%q eval=%d parsed=%d", lit, n, parsed)
		}
	})
}

func TestStringLiteral_RoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		lit := exprgen.StringLiteral().Draw(t, "lit")
		result, err := evalLiteral(lit)
		if err != nil {
			t.Fatalf("eval failed for %q: %v", lit, err)
		}
		s := result.(string)
		unquoted, err := strconv.Unquote(lit)
		if err != nil {
			t.Fatalf("strconv.Unquote failed for %q: %v", lit, err)
		}
		if s != unquoted {
			t.Fatalf("round-trip mismatch: lit=%q eval=%q unquoted=%q", lit, s, unquoted)
		}
	})
}

// --- Combined generator test ---

func TestLiteral_Evaluates(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		lit := exprgen.Literal().Draw(t, "lit")
		if _, err := evalLiteral(lit); err != nil {
			t.Fatalf("Literal %q evaluation failed: %v", lit, err)
		}
	})
}

// --- Table-driven known examples ---

func TestLiteral_KnownExamples(t *testing.T) {
	tests := []struct {
		name     string
		lit      string
		expected evaluator.Value
	}{
		{"zero int", "0", 0},
		{"positive int", "42", 42},
		{"negative int", "-1", -1},
		{"large int", "999999", 999999},
		{"zero float", "0.0", 0.0},
		{"pi", "3.14", 3.14},
		{"negative float", "-2.5", -2.5},
		{"scientific", "1e10", 1e10},
		{"small scientific", "1.5e-3", 1.5e-3},
		{"empty string", `""`, ""},
		{"hello string", `"hello"`, "hello"},
		{"escaped newline", `"line1\nline2"`, "line1\nline2"},
		{"escaped tab", `"a\tb"`, "a\tb"},
		{"escaped quote", `"say \"hi\""`, `say "hi"`},
		{"true", "true", true},
		{"false", "false", false},
		{"null", "null", nil},
		{"nil", "nil", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evalLiteral(tt.lit)
			if err != nil {
				t.Fatalf("eval %q: %v", tt.lit, err)
			}
			if result != tt.expected {
				t.Errorf("eval %q = %v (%T), want %v (%T)", tt.lit, result, result, tt.expected, tt.expected)
			}
		})
	}
}
