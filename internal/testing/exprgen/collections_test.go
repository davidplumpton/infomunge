package exprgen_test

import (
	"strings"
	"testing"

	"infomunge/internal/testing/exprgen"

	"pgregory.net/rapid"
)

func TestCollectionOpGen_SyntacticallyValid(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		expr := exprgen.CollectionOpGen(3, exprgen.FeatureAll).Draw(t, "expr")
		if !exprgen.IsValid(expr) {
			t.Fatalf("CollectionOpGen produced syntactically invalid expression: %q", expr)
		}
	})
}

func TestLambdaGen_SyntacticallyValid(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		expr := exprgen.LambdaGen(3, exprgen.FeatureAll).Draw(t, "expr")
		if !exprgen.IsValid(expr) {
			t.Fatalf("LambdaGen produced syntactically invalid expression: %q", expr)
		}
	})
}

func TestExpression_CollectionAndLambdaFormsAppear(t *testing.T) {
	sawCollection := false
	sawLambda := false
	sawImplicit := false

	features := exprgen.FeatureCollections | exprgen.FeatureLambdas | exprgen.FeatureImplicitLambda | exprgen.FeatureArrays | exprgen.FeatureArithmetic | exprgen.FeatureComparison | exprgen.FeatureDotAccess | exprgen.FeatureParens
	for i := 0; i < 1000; i++ {
		var expr string
		rapid.Check(t, func(t *rapid.T) {
			expr = exprgen.Expression(4, features).Draw(t, "expr")
		})
		if !exprgen.IsValid(expr) {
			t.Fatalf("generated expression is invalid: %q", expr)
		}
		if strings.Contains(expr, " map ") || strings.Contains(expr, " filter ") || strings.Contains(expr, " reduce ") || strings.Contains(expr, " flatMap ") {
			sawCollection = true
		}
		if strings.Contains(expr, "->") {
			sawLambda = true
		}
		if strings.Contains(expr, "$") {
			sawImplicit = true
		}
		if sawCollection && sawLambda && sawImplicit {
			break
		}
	}

	if !sawCollection {
		t.Fatal("1000 draws never produced a collection operator")
	}
	if !sawLambda {
		t.Fatal("1000 draws never produced an explicit lambda")
	}
	if !sawImplicit {
		t.Fatal("1000 draws never produced an implicit lambda reference ($ or $$)")
	}
}
