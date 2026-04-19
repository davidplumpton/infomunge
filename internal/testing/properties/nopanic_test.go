package properties_test

import (
	"flag"
	"fmt"
	"runtime/debug"
	"testing"

	"infomunge/internal/evaluator"
	"infomunge/internal/preprocessor"
	"infomunge/internal/testing/determinism"
	"infomunge/internal/testing/exprgen"
	"infomunge/internal/testing/metrics"
	"infomunge/internal/testing/testbudget"

	"pgregory.net/rapid"
)

var validTypeNames = map[string]struct{}{
	"Null":      {},
	"Boolean":   {},
	"Number":    {},
	"String":    {},
	"Array":     {},
	"Object":    {},
	"Function":  {},
	"Regex":     {},
	"Type":      {},
	"Namespace": {},
}

func TestEvaluate_NoPanics_Deterministic_AndTypeConsistent(t *testing.T) {
	setRapidChecks(t, testbudget.NoPanicChecks())

	rapid.Check(t, func(t *rapid.T) {
		tc := exprgen.SampleContext().Draw(t, "ctx")
		expr := drawExpression(t, tc)

		firstResult, firstErr, panicValue, panicStack := safeEval(expr, tc.Value)
		if panicValue != nil {
			metrics.RecordPanic()
			t.Fatalf("panic while evaluating expression\nexpr: %q\nctx: %#v\npanic: %v\nstack:\n%s", expr, tc.Value, panicValue, panicStack)
		}

		secondResult, secondErr, panicValue, panicStack := safeEval(expr, tc.Value)
		if panicValue != nil {
			metrics.RecordPanic()
			t.Fatalf("panic while re-evaluating expression\nexpr: %q\nctx: %#v\npanic: %v\nstack:\n%s", expr, tc.Value, panicValue, panicStack)
		}

		if (firstErr == nil) != (secondErr == nil) {
			t.Fatalf("nondeterministic success/error\nexpr: %q\nctx: %#v\nfirst err: %v\nsecond err: %v", expr, tc.Value, firstErr, secondErr)
		}

		if firstErr != nil {
			if !determinism.EqualErrors(firstErr, secondErr) {
				t.Fatalf("nondeterministic error message\nexpr: %q\nctx: %#v\nfirst err: %v\nsecond err: %v", expr, tc.Value, firstErr, secondErr)
			}
			return
		}

		if !determinism.Equal(firstResult, secondResult) {
			t.Fatalf("nondeterministic result\nexpr: %q\nctx: %#v\nfirst:  %#v\nsecond: %#v", expr, tc.Value, firstResult, secondResult)
		}

		typeExpr := fmt.Sprintf("typeOf(%s)", expr)
		typeResult, typeErr, panicValue, panicStack := safeEval(typeExpr, tc.Value)
		if panicValue != nil {
			metrics.RecordPanic()
			t.Fatalf("panic while evaluating typeOf\nexpr: %q\nctx: %#v\npanic: %v\nstack:\n%s", typeExpr, tc.Value, panicValue, panicStack)
		}
		if typeErr != nil {
			t.Fatalf("typeOf evaluation failed\nexpr: %q\nctx: %#v\nerr: %v", typeExpr, tc.Value, typeErr)
		}

		typeName, ok := typeResult.(string)
		if !ok {
			t.Fatalf("typeOf returned non-string\nexpr: %q\nctx: %#v\nresult: %#v (%T)", typeExpr, tc.Value, typeResult, typeResult)
		}
		if _, ok := validTypeNames[typeName]; !ok {
			t.Fatalf("typeOf returned invalid type name\nexpr: %q\nctx: %#v\ntype: %q", typeExpr, tc.Value, typeName)
		}
	})
}

func drawExpression(t *rapid.T, tc exprgen.TestContext) string {
	switch rapid.IntRange(0, 3).Draw(t, "expr_kind") {
	case 0:
		return exprgen.IdentGen(tc).Draw(t, "ident_expr")
	case 1:
		base := exprgen.Expression(3, exprgen.FeatureAll).Draw(t, "base_expr")
		ident := exprgen.IdentGen(tc).Draw(t, "ident_binary")
		return fmt.Sprintf("%s + %s", ident, base)
	default:
		return exprgen.Expression(3, exprgen.FeatureAll).Draw(t, "random_expr")
	}
}

func TestEvaluate_DeterministicLambdaResults(t *testing.T) {
	ctx := evaluator.Object{
		"payload": evaluator.Object{
			"active":  true,
			"address": 7,
			"age":     true,
			"name":    evaluator.Array{-230},
			"score":   false,
			"tags": evaluator.Object{
				"age": "c",
				"name": evaluator.Object{
					"active": true,
					"age":    "dwho",
					"name":   false,
				},
			},
		},
	}
	expr := "(left, current) -> -1.7976931348623157e+308"

	firstResult, firstErr := evalWithContext(expr, ctx)
	if firstErr != nil {
		t.Fatalf("first eval failed: %v", firstErr)
	}
	secondResult, secondErr := evalWithContext(expr, ctx)
	if secondErr != nil {
		t.Fatalf("second eval failed: %v", secondErr)
	}

	if _, ok := firstResult.(*evaluator.Lambda); !ok {
		t.Fatalf("expected lambda result, got %#v (%T)", firstResult, firstResult)
	}
	if _, ok := secondResult.(*evaluator.Lambda); !ok {
		t.Fatalf("expected lambda result, got %#v (%T)", secondResult, secondResult)
	}
	if !determinism.Equal(firstResult, secondResult) {
		t.Fatalf("lambda results should be treated as deterministic\nfirst: %#v\nsecond: %#v", firstResult, secondResult)
	}
}

func evalWithContext(expr string, ctx evaluator.Context) (evaluator.Value, error) {
	prepared, mapping, err := preprocessor.PrepareForParsing(expr, preprocessor.Options{})
	if err != nil {
		return nil, err
	}
	return evaluator.Evaluate(prepared, ctx, mapping, 0, expr)
}

func safeEval(expr string, ctx evaluator.Context) (result evaluator.Value, err error, panicValue interface{}, panicStack string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			panicValue = recovered
			panicStack = string(debug.Stack())
		}
	}()

	result, err = evalWithContext(expr, ctx)
	return result, err, nil, ""
}

func setRapidChecks(t *testing.T, checks int) {
	t.Helper()
	f := flag.Lookup("rapid.checks")
	if f == nil {
		t.Fatal("rapid.checks flag not found")
	}

	previous := f.Value.String()
	if err := f.Value.Set(fmt.Sprintf("%d", checks)); err != nil {
		t.Fatalf("set rapid.checks=%d: %v", checks, err)
	}
	t.Cleanup(func() {
		_ = f.Value.Set(previous)
	})
}
