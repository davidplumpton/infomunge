package properties_test

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"testing"

	"infomunge/internal/runner"
	"infomunge/internal/testing/exprgen"

	"pgregory.net/rapid"
)

const algebraicChecks = 500

func TestArithmeticIdentities(t *testing.T) {
	setRapidChecks(t, algebraicChecks)

	rapid.Check(t, func(t *rapid.T) {
		x := drawSmallNumberLiteral(t, "x")
		assertExpressionTrue(t, fmt.Sprintf("(%s + 0) == %s", x, x))
		assertExpressionTrue(t, fmt.Sprintf("(%s * 1) == %s", x, x))
		assertExpressionTrue(t, fmt.Sprintf("(%s * 0) == 0", x))
		assertExpressionTrue(t, fmt.Sprintf("(%s ** 1) == %s", x, x))
	})
}

func TestCommutativity(t *testing.T) {
	setRapidChecks(t, algebraicChecks)

	rapid.Check(t, func(t *rapid.T) {
		a := drawSmallNumberLiteral(t, "a")
		b := drawSmallNumberLiteral(t, "b")
		assertExpressionTrue(t, fmt.Sprintf("(%s + %s) == (%s + %s)", a, b, b, a))
		assertExpressionTrue(t, fmt.Sprintf("(%s * %s) == (%s * %s)", a, b, b, a))
	})
}

func TestAssociativity(t *testing.T) {
	setRapidChecks(t, algebraicChecks)

	rapid.Check(t, func(t *rapid.T) {
		a := drawSmallFloatLiteral(t, "a")
		b := drawSmallFloatLiteral(t, "b")
		c := drawSmallFloatLiteral(t, "c")

		leftRaw, err := evalExpr(fmt.Sprintf("(%s + %s) + %s", a, b, c))
		if err != nil {
			t.Fatalf("left expression failed: %v", err)
		}
		rightRaw, err := evalExpr(fmt.Sprintf("%s + (%s + %s)", a, b, c))
		if err != nil {
			t.Fatalf("right expression failed: %v", err)
		}

		left, ok := toFloat64(leftRaw)
		if !ok {
			t.Fatalf("left expression was not numeric: %#v (%T)", leftRaw, leftRaw)
		}
		right, ok := toFloat64(rightRaw)
		if !ok {
			t.Fatalf("right expression was not numeric: %#v (%T)", rightRaw, rightRaw)
		}

		if math.Abs(left-right) > 1e-9 {
			t.Fatalf("associativity violated: left=%v right=%v a=%s b=%s c=%s", left, right, a, b, c)
		}
	})
}

func TestDefaultOperatorIdentities(t *testing.T) {
	setRapidChecks(t, algebraicChecks)

	rapid.Check(t, func(t *rapid.T) {
		x := drawAnyLiteral(t, "x")
		y := drawAnyLiteral(t, "y")
		nonNullX := drawNonNullLiteral(t, "nonnull_x")

		assertExpressionTrue(t, fmt.Sprintf("(%s default %s) == %s", x, x, x))
		assertExpressionTrue(t, fmt.Sprintf("(%s default %s) == %s", nonNullX, y, nonNullX))
		assertExpressionTrue(t, fmt.Sprintf("(null default %s) == %s", y, y))
	})
}

func TestBooleanLogicIdentities(t *testing.T) {
	setRapidChecks(t, algebraicChecks)

	rapid.Check(t, func(t *rapid.T) {
		x := rapid.SampledFrom([]string{"true", "false"}).Draw(t, "x")
		assertExpressionTrue(t, fmt.Sprintf("!(!(%s)) == %s", x, x))
		assertExpressionTrue(t, fmt.Sprintf("(%s && true) == %s", x, x))
		assertExpressionTrue(t, fmt.Sprintf("(%s || false) == %s", x, x))
	})
}

func TestCollectionIdentityProperties(t *testing.T) {
	setRapidChecks(t, algebraicChecks)

	rapid.Check(t, func(t *rapid.T) {
		arr := drawArrayLiteral(t, "arr")
		mapExpr := fmt.Sprintf("%s map (x) -> x", arr)
		filterExpr := fmt.Sprintf("%s filter (x) -> true", arr)

		mapResult, err := evalExpr(mapExpr)
		if err != nil {
			t.Fatalf("map identity failed to evaluate: %v", err)
		}
		filterResult, err := evalExpr(filterExpr)
		if err != nil {
			t.Fatalf("filter identity failed to evaluate: %v", err)
		}
		arrValue, err := evalExpr(arr)
		if err != nil {
			t.Fatalf("array literal failed to evaluate: %v", err)
		}

		if !reflect.DeepEqual(mapResult, arrValue) {
			t.Fatalf("map identity violated: expr=%q got=%#v want=%#v", mapExpr, mapResult, arrValue)
		}
		if !reflect.DeepEqual(filterResult, arrValue) {
			t.Fatalf("filter identity violated: expr=%q got=%#v want=%#v", filterExpr, filterResult, arrValue)
		}
	})
}

func TestConcatenationIdentities(t *testing.T) {
	setRapidChecks(t, algebraicChecks)

	rapid.Check(t, func(t *rapid.T) {
		arr := drawArrayLiteral(t, "arr")
		leftExpr := fmt.Sprintf("[] ++ %s", arr)
		rightExpr := fmt.Sprintf("%s ++ []", arr)

		leftResult, err := evalExpr(leftExpr)
		if err != nil {
			t.Fatalf("left concat identity failed: %v", err)
		}
		rightResult, err := evalExpr(rightExpr)
		if err != nil {
			t.Fatalf("right concat identity failed: %v", err)
		}
		arrValue, err := evalExpr(arr)
		if err != nil {
			t.Fatalf("array literal failed to evaluate: %v", err)
		}

		if !reflect.DeepEqual(leftResult, arrValue) {
			t.Fatalf("left concat identity violated: expr=%q got=%#v want=%#v", leftExpr, leftResult, arrValue)
		}
		if !reflect.DeepEqual(rightResult, arrValue) {
			t.Fatalf("right concat identity violated: expr=%q got=%#v want=%#v", rightExpr, rightResult, arrValue)
		}
	})
}

func TestStringOperationProperties(t *testing.T) {
	setRapidChecks(t, algebraicChecks)

	rapid.Check(t, func(t *rapid.T) {
		s := strconv.Quote(rapid.StringMatching(`[A-Za-z0-9 ]{0,40}`).Draw(t, "s"))
		leftRaw, err := evalExpr(fmt.Sprintf("sizeOf(upper(lower(%s)))", s))
		if err != nil {
			t.Fatalf("sizeOf(upper(lower(s))) evaluation failed: %v", err)
		}
		rightRaw, err := evalExpr(fmt.Sprintf("sizeOf(%s)", s))
		if err != nil {
			t.Fatalf("sizeOf(s) evaluation failed: %v", err)
		}
		nonNegativeRaw, err := evalExpr(fmt.Sprintf("sizeOf(%s) >= 0", s))
		if err != nil {
			t.Fatalf("sizeOf(s) >= 0 evaluation failed: %v", err)
		}

		left, leftOK := toFloat64(leftRaw)
		right, rightOK := toFloat64(rightRaw)
		if !leftOK || !rightOK {
			t.Fatalf("sizeOf results must be numeric, got left=%#v right=%#v", leftRaw, rightRaw)
		}
		if left != right {
			t.Fatalf("length changed after upper/lower: left=%v right=%v s=%s", left, right, s)
		}
		assertIsTrue(t, nonNegativeRaw, "sizeOf(s) must be non-negative")
	})
}

func evalExpr(expr string) (interface{}, error) {
	return runner.RunString(exprgen.WrapScript(expr), nil)
}

func assertExpressionTrue(t *rapid.T, expr string) {
	t.Helper()
	result, err := evalExpr(expr)
	if err != nil {
		t.Fatalf("expression failed: %q err=%v", expr, err)
	}
	assertIsTrue(t, result, fmt.Sprintf("expected true for %q", expr))
}

func assertIsTrue(t *rapid.T, value interface{}, msg string) {
	t.Helper()
	b, ok := value.(bool)
	if !ok || !b {
		t.Fatalf("%s: %#v (%T)", msg, value, value)
	}
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case float64:
		return n, true
	default:
		return 0, false
	}
}

func drawSmallNumberLiteral(t *rapid.T, label string) string {
	if rapid.Bool().Draw(t, label+"_is_int") {
		return strconv.Itoa(rapid.IntRange(-1000, 1000).Draw(t, label+"_int"))
	}
	return drawSmallFloatLiteral(t, label+"_float")
}

func drawSmallFloatLiteral(t *rapid.T, label string) string {
	v := rapid.Float64Range(-1000.0, 1000.0).Draw(t, label)
	if math.IsNaN(v) || math.IsInf(v, 0) {
		v = 0
	}
	return strconv.FormatFloat(v, 'f', 6, 64)
}

func drawNonNullLiteral(t *rapid.T, label string) string {
	switch rapid.IntRange(0, 3).Draw(t, label+"_kind") {
	case 0:
		return strconv.Itoa(rapid.IntRange(-1000, 1000).Draw(t, label+"_int"))
	case 1:
		return drawSmallFloatLiteral(t, label+"_float")
	case 2:
		return exprgen.StringLiteral().Draw(t, label+"_string")
	default:
		return exprgen.BoolLiteral().Draw(t, label+"_bool")
	}
}

func drawAnyLiteral(t *rapid.T, label string) string {
	switch rapid.IntRange(0, 4).Draw(t, label+"_kind") {
	case 0:
		return strconv.Itoa(rapid.IntRange(-1000, 1000).Draw(t, label+"_int"))
	case 1:
		return drawSmallFloatLiteral(t, label+"_float")
	case 2:
		return exprgen.StringLiteral().Draw(t, label+"_string")
	case 3:
		return exprgen.BoolLiteral().Draw(t, label+"_bool")
	default:
		return exprgen.NullLiteral().Draw(t, label+"_null")
	}
}

func drawArrayLiteral(t *rapid.T, label string) string {
	length := rapid.IntRange(0, 6).Draw(t, label+"_len")
	if length == 0 {
		return "[]"
	}

	items := make([]string, length)
	for i := 0; i < length; i++ {
		itemLabel := fmt.Sprintf("%s_item_%d", label, i)
		switch rapid.IntRange(0, 4).Draw(t, itemLabel+"_kind") {
		case 0:
			items[i] = strconv.Itoa(rapid.IntRange(-1000, 1000).Draw(t, itemLabel+"_int"))
		case 1:
			items[i] = drawSmallFloatLiteral(t, itemLabel+"_float")
		case 2:
			items[i] = exprgen.StringLiteral().Draw(t, itemLabel+"_string")
		case 3:
			items[i] = exprgen.BoolLiteral().Draw(t, itemLabel+"_bool")
		default:
			items[i] = exprgen.NullLiteral().Draw(t, itemLabel+"_null")
		}
	}

	return "[" + joinComma(items) + "]"
}

func joinComma(items []string) string {
	if len(items) == 0 {
		return ""
	}
	result := items[0]
	for i := 1; i < len(items); i++ {
		result += ", " + items[i]
	}
	return result
}
