package exprgen

import (
	"reflect"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

func TestDWCompatOperatorsReplacePercentWithInfixMod(t *testing.T) {
	got := (exprConfig{DWCompat: true}).filterOps([]string{"%", "+", "++", "*"})
	want := []string{"mod", "*"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DW-compatible operators = %v, want %v", got, want)
	}
}

func TestDWCompatCollectionOpsIncludeReduce(t *testing.T) {
	got := (exprConfig{DWCompat: true}).collectionOps()
	want := []string{"map", "filter", "flatMap", "reduce"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DW-compatible collection operators = %v, want %v", got, want)
	}
}

func TestDWCompatReduceUsesTwoParametersAndReturnsAccumulator(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		got := reduceExpr(t, "[1, 2, 3]", 1, FeatureDWCompat, lambdaScope{}, exprConfig{DWCompat: true})
		signature := strings.TrimPrefix(got, "[1, 2, 3] reduce (")
		parts := strings.SplitN(signature, ") -> ", 2)
		if len(parts) != 2 {
			t.Fatalf("DW-compatible reduce = %q, want explicit callback", got)
		}
		params := strings.Split(parts[0], ", ")
		if len(params) != 2 {
			t.Fatalf("DW-compatible reduce parameters = %q, want exactly two", parts[0])
		}
		if parts[1] != params[1] {
			t.Fatalf("DW-compatible reduce body = %q, want accumulator %q", parts[1], params[1])
		}
	})
}

func TestDWCompatNestedFeaturesExcludeCollections(t *testing.T) {
	features := FeatureCollections | FeatureArrays | FeatureConditionals
	got := expressionNestedFeatures(features, exprConfig{DWCompat: true})
	want := FeatureArrays
	if got != want {
		t.Fatalf("DW-compatible nested features = %v, want %v", got, want)
	}

	got = expressionNestedFeatures(features, exprConfig{})
	want = FeatureCollections | FeatureArrays
	if got != want {
		t.Fatalf("ordinary nested features = %v, want %v", got, want)
	}
}

func TestDWCompatIndexAccessUsesKnownObjectPayload(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		got := filteredIndexAccessExpr(3, FeatureDWCompat, lambdaScope{}, exprConfig{DWCompat: true}).Draw(t, "index")
		if !strings.HasPrefix(got, `payload["`) {
			t.Fatalf("DW-compatible index access = %q, want known object payload", got)
		}
	})
}
