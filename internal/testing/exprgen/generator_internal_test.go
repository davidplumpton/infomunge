package exprgen

import (
	"reflect"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

func TestDWCompatOperatorsReplacePercentWithInfixMod(t *testing.T) {
	got := (exprConfig{DWCompat: true}).filterOps([]string{"%", "+", "++", "*"})
	want := []string{"mod", "+", "*"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DW-compatible operators = %v, want %v", got, want)
	}
}

func TestDWCompatPlusPairsIncludeOnlySharedOperandCategories(t *testing.T) {
	want := []dwPlusOperandPair{
		{left: dwPlusNumber, right: dwPlusNumber},
		{left: dwPlusNumber, right: dwPlusNumericString},
		{left: dwPlusNumericString, right: dwPlusNumber},
	}
	if !reflect.DeepEqual(dwCompatiblePlusPairs, want) {
		t.Fatalf("DW-compatible plus pairs = %v, want %v", dwCompatiblePlusPairs, want)
	}
}

func TestDWCompatPlusPairsExcludeExtensionCategories(t *testing.T) {
	excluded := []struct {
		name string
		pair dwPlusOperandPair
		expr string
	}{
		{
			name: "minimized array-left string-right regression",
			pair: dwPlusOperandPair{left: dwPlusArray, right: dwPlusString},
			expr: `[""] + ""`,
		},
		{
			name: "string concatenation extension",
			pair: dwPlusOperandPair{left: dwPlusString, right: dwPlusString},
			expr: `"a" + "b"`,
		},
		{
			name: "array concatenation mismatch",
			pair: dwPlusOperandPair{left: dwPlusArray, right: dwPlusArray},
			expr: `[1] + [2]`,
		},
		{
			name: "unsupported object operands",
			pair: dwPlusOperandPair{left: dwPlusObject, right: dwPlusObject},
			expr: `{a: 1} + {b: 2}`,
		},
		{
			name: "unsupported boolean operands",
			pair: dwPlusOperandPair{left: dwPlusBoolean, right: dwPlusBoolean},
			expr: `true + false`,
		},
		{
			name: "unsupported null operands",
			pair: dwPlusOperandPair{left: dwPlusNull, right: dwPlusNull},
			expr: `null + null`,
		},
	}

	for _, tc := range excluded {
		t.Run(tc.name, func(t *testing.T) {
			for _, generated := range dwCompatiblePlusPairs {
				if generated == tc.pair {
					t.Fatalf("DW-compatible plus pairs include %s (%s)", tc.name, tc.expr)
				}
			}
		})
	}
}

func TestDWCompatBinaryPlusUsesTypedSharedOperands(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		got := filteredBinaryExpr(
			3,
			FeatureDWCompat,
			[]string{"+"},
			lambdaScope{},
			exprConfig{DWCompat: true},
		).Draw(t, "plus")
		if strings.Contains(got, "[") || strings.Contains(got, "{") ||
			strings.Contains(got, "true") || strings.Contains(got, "false") ||
			strings.Contains(got, "null") {
			t.Fatalf("DW-compatible plus generated a nonnumeric operand: %q", got)
		}
	})
}

func TestDWCompatCollectionOpsIncludeReduce(t *testing.T) {
	got := (exprConfig{DWCompat: true}).collectionOps()
	want := []string{"map", "filter", "flatMap", "reduce", "groupBy", "pluck", "mapObject"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DW-compatible collection operators = %v, want %v", got, want)
	}
}

func TestDWCompatCollectionInputsCoverComputedCollectionSources(t *testing.T) {
	sawFunctionCall := false
	sawConcatenation := false
	sawConditional := false
	for i := 0; i < 1000; i++ {
		var got string
		rapid.Check(t, func(t *rapid.T) {
			got = collectionInputExpr(t, 3, FeatureDWCompat, lambdaScope{}, exprConfig{DWCompat: true}, "map")
		})
		sawFunctionCall = sawFunctionCall || strings.HasPrefix(got, "flatten([")
		sawConcatenation = sawConcatenation || strings.Contains(got, " ++ ")
		sawConditional = sawConditional || strings.HasPrefix(got, "flatten(if (")
		if sawFunctionCall && sawConcatenation && sawConditional {
			break
		}
	}
	if !sawFunctionCall {
		t.Fatal("DW-compatible collection inputs never selected a function-call source")
	}
	if !sawConcatenation {
		t.Fatal("DW-compatible collection inputs never selected a concatenated-array source")
	}
	if !sawConditional {
		t.Fatal("DW-compatible collection inputs never selected a conditional source")
	}
}

func TestDWCompatObjectCollectionInputsUseObjectSources(t *testing.T) {
	for _, op := range []string{"pluck", "mapObject"} {
		t.Run(op, func(t *testing.T) {
			rapid.Check(t, func(t *rapid.T) {
				got := collectionInputExpr(t, 3, FeatureDWCompat, lambdaScope{}, exprConfig{DWCompat: true}, op)
				if !strings.HasPrefix(got, "{") {
					t.Fatalf("DW-compatible %s source = %q, want object expression", op, got)
				}
			})
		})
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

func TestDWCompatNestedFeaturesIncludeCollections(t *testing.T) {
	features := FeatureCollections | FeatureArrays | FeatureConditionals
	got := expressionNestedFeatures(features, exprConfig{DWCompat: true})
	want := FeatureCollections | FeatureArrays
	if got != want {
		t.Fatalf("DW-compatible nested features = %v, want %v", got, want)
	}

	got = expressionNestedFeatures(features, exprConfig{})
	want = FeatureCollections | FeatureArrays
	if got != want {
		t.Fatalf("ordinary nested features = %v, want %v", got, want)
	}
}

func TestDWCompatIndexAccessCoversPayloadAndNumericOrdinalValues(t *testing.T) {
	sawRoot := false
	sawNested := false
	sawNumber := false
	rapid.Check(t, func(t *rapid.T) {
		got := filteredIndexAccessExpr(3, FeatureDWCompat, lambdaScope{}, exprConfig{DWCompat: true}).Draw(t, "index")
		if strings.HasPrefix(got, `payload["`) {
			sawRoot = true
		}
		if strings.HasPrefix(got, `payload.`) {
			sawNested = true
		}
		if strings.HasPrefix(got, "(") {
			sawNumber = true
		}
	})
	if !sawRoot {
		t.Fatal("DW-compatible index access never selected the root payload")
	}
	if !sawNested {
		t.Fatal("DW-compatible index access never selected a nested payload value")
	}
	if !sawNumber {
		t.Fatal("DW-compatible index access never selected a numeric ordinal value")
	}
}
