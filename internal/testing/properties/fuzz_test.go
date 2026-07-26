package properties_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"infomunge/internal/preprocessor"
	"infomunge/internal/testing/exprgen"
	"infomunge/internal/testing/metrics"

	"pgregory.net/rapid"
)

// FuzzExprEval bridges rapid's structured generators into Go fuzzing so corpus
// mutation still explores syntactically valid expressions most of the time.
func FuzzExprEval(f *testing.F) {
	f.Fuzz(rapid.MakeFuzz(func(t *rapid.T) {
		tc := exprgen.SampleContext().Draw(t, "ctx")
		expr := exprgen.Expression(3, exprgen.FeatureAll).Draw(t, "expr")

		result, err, panicValue, panicStack := safeEval(expr, tc.Value)
		if panicValue != nil {
			metrics.RecordPanic()
			t.Fatalf("panic in FuzzExprEval expr=%q ctx=%#v panic=%v\n%s", expr, tc.Value, panicValue, panicStack)
		}
		if err != nil {
			return
		}

		if _, marshalErr := json.Marshal(result); marshalErr != nil {
			t.Fatalf("non-serializable result expr=%q result=%#v err=%v", expr, result, marshalErr)
		}

		typeExpr := fmt.Sprintf("typeOf(%s)", expr)
		typeResult, typeErr, panicValue, panicStack := safeEval(typeExpr, tc.Value)
		if panicValue != nil {
			metrics.RecordPanic()
			t.Fatalf("panic in typeOf expr=%q ctx=%#v panic=%v\n%s", typeExpr, tc.Value, panicValue, panicStack)
		}
		if typeErr != nil {
			t.Fatalf("typeOf failed expr=%q err=%v", typeExpr, typeErr)
		}

		typeName, ok := runtimeTypeName(typeResult)
		if !ok {
			t.Fatalf("typeOf returned non-Type value expr=%q result=%#v (%T)", typeExpr, typeResult, typeResult)
		}
		if _, ok := validTypeNames[typeName]; !ok {
			t.Fatalf("typeOf returned invalid type name expr=%q type=%q", typeExpr, typeName)
		}
	}))
}

func FuzzExprPreprocess(f *testing.F) {
	f.Fuzz(rapid.MakeFuzz(func(t *rapid.T) {
		expr := exprgen.Expression(4, exprgen.FeatureAll).Draw(t, "expr")
		opts := preprocessor.Options{AllowMultilineIfElse: true}
		preprocessor.PrepareForParsing(expr, opts) //nolint:errcheck
	}))
}

func FuzzExprDeep(f *testing.F) {
	f.Fuzz(rapid.MakeFuzz(func(t *rapid.T) {
		tc := exprgen.SampleContext().Draw(t, "ctx")
		depth := rapid.IntRange(5, 6).Draw(t, "depth")
		expr := exprgen.Expression(depth, exprgen.FeatureAll).Draw(t, "expr")

		result, err, panicValue, panicStack := safeEval(expr, tc.Value)
		if panicValue != nil {
			metrics.RecordPanic()
			t.Fatalf("panic in FuzzExprDeep expr=%q ctx=%#v panic=%v\n%s", expr, tc.Value, panicValue, panicStack)
		}
		if err != nil {
			return
		}
		if _, marshalErr := json.Marshal(result); marshalErr != nil {
			t.Fatalf("non-serializable deep result expr=%q result=%#v err=%v", expr, result, marshalErr)
		}
	}))
}
