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

func TestDWCompatCollectionOpsExcludeReduce(t *testing.T) {
	got := (exprConfig{DWCompat: true}).collectionOps()
	want := []string{"map", "filter", "flatMap"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DW-compatible collection operators = %v, want %v", got, want)
	}
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
