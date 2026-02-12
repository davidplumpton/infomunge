package properties_test

import (
	"encoding/json"
	"fmt"
	"infomunge/internal/testing/exprgen"
	"infomunge/pkg/formats"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

const structuralChecks = 500

func TestStructural_SizeOfMapEqualsOriginalLength(t *testing.T) {
	setRapidChecks(t, structuralChecks)

	rapid.Check(t, func(t *rapid.T) {
		arr := drawArrayLiteral(t, "arr")
		lambdaBody := rapid.SampledFrom([]string{
			"x",
			"typeOf(x)",
			"x default null",
			"[x]",
			"isEmpty([x])",
			"sizeOf([x])",
		}).Draw(t, "map_lambda")

		left, err := evalExpr(fmt.Sprintf("sizeOf(%s map (x) -> %s)", arr, lambdaBody))
		if err != nil {
			t.Fatalf("map expression failed: %v", err)
		}
		right, err := evalExpr(fmt.Sprintf("sizeOf(%s)", arr))
		if err != nil {
			t.Fatalf("array size expression failed: %v", err)
		}

		if !reflect.DeepEqual(left, right) {
			t.Fatalf("sizeOf(arr map f) != sizeOf(arr): left=%#v right=%#v arr=%s", left, right, arr)
		}
	})
}

func TestStructural_SizeOfFilterIsBounded(t *testing.T) {
	setRapidChecks(t, structuralChecks)

	rapid.Check(t, func(t *rapid.T) {
		arr := drawArrayLiteral(t, "arr")
		predicate := rapid.SampledFrom([]string{
			"true",
			"false",
			"typeOf(x) == typeOf(x)",
			"isEmpty([x])",
			"!(false)",
		}).Draw(t, "filter_lambda")

		filteredSizeRaw, err := evalExpr(fmt.Sprintf("sizeOf(%s filter (x) -> %s)", arr, predicate))
		if err != nil {
			t.Fatalf("filter expression failed: %v", err)
		}
		origSizeRaw, err := evalExpr(fmt.Sprintf("sizeOf(%s)", arr))
		if err != nil {
			t.Fatalf("array size expression failed: %v", err)
		}

		filteredSize, ok := toInt(filteredSizeRaw)
		if !ok {
			t.Fatalf("filtered size was not numeric: %#v", filteredSizeRaw)
		}
		origSize, ok := toInt(origSizeRaw)
		if !ok {
			t.Fatalf("original size was not numeric: %#v", origSizeRaw)
		}
		if filteredSize > origSize {
			t.Fatalf("sizeOf(arr filter f) > sizeOf(arr): filtered=%d original=%d arr=%s", filteredSize, origSize, arr)
		}
	})
}

func TestStructural_FlattenSizeLowerBound(t *testing.T) {
	setRapidChecks(t, structuralChecks)

	rapid.Check(t, func(t *rapid.T) {
		arr := drawArrayOfArraysLiteral(t, "arr_of_arr")
		flatSizeRaw, err := evalExpr(fmt.Sprintf("sizeOf(flatten(%s))", arr))
		if err != nil {
			t.Fatalf("flatten size expression failed: %v", err)
		}
		origSizeRaw, err := evalExpr(fmt.Sprintf("sizeOf(%s)", arr))
		if err != nil {
			t.Fatalf("array size expression failed: %v", err)
		}

		flatSize, ok := toInt(flatSizeRaw)
		if !ok {
			t.Fatalf("flatten size was not numeric: %#v", flatSizeRaw)
		}
		origSize, ok := toInt(origSizeRaw)
		if !ok {
			t.Fatalf("original size was not numeric: %#v", origSizeRaw)
		}
		if flatSize < origSize {
			t.Fatalf("sizeOf(flatten(arr)) < sizeOf(arr): flattened=%d original=%d arr=%s", flatSize, origSize, arr)
		}
	})
}

func TestStructural_KeysValuesSameSize(t *testing.T) {
	setRapidChecks(t, structuralChecks)

	rapid.Check(t, func(t *rapid.T) {
		obj := drawObjectLiteral(t, "obj")
		keysSizeRaw, err := evalExpr(fmt.Sprintf("sizeOf(keys(%s))", obj))
		if err != nil {
			t.Fatalf("keys size expression failed: %v", err)
		}
		valuesSizeRaw, err := evalExpr(fmt.Sprintf("sizeOf(values(%s))", obj))
		if err != nil {
			t.Fatalf("values size expression failed: %v", err)
		}

		if !reflect.DeepEqual(keysSizeRaw, valuesSizeRaw) {
			t.Fatalf("sizeOf(keys(obj)) != sizeOf(values(obj)): keys=%#v values=%#v obj=%s", keysSizeRaw, valuesSizeRaw, obj)
		}
	})
}

func TestRoundTrip_JSONSafeValues(t *testing.T) {
	setRapidChecks(t, structuralChecks)

	rapid.Check(t, func(t *rapid.T) {
		valueExpr := drawJSONSafeLiteral(t, "json_safe")
		original, err := evalExpr(valueExpr)
		if err != nil {
			t.Fatalf("original expression failed: %v", err)
		}

		roundTripExpr := fmt.Sprintf(`read(write(%s, "application/json"), "application/json")`, valueExpr)
		roundTripped, err := evalExpr(roundTripExpr)
		if err != nil {
			t.Fatalf("round-trip expression failed: %v", err)
		}

		origJSON, err := canonicalJSON(original)
		if err != nil {
			t.Fatalf("marshal original failed: %v", err)
		}
		roundJSON, err := canonicalJSON(roundTripped)
		if err != nil {
			t.Fatalf("marshal round-trip failed: %v", err)
		}
		if origJSON != roundJSON {
			t.Fatalf("round-trip mismatch:\nexpr: %s\norig:  %s\nround: %s", valueExpr, origJSON, roundJSON)
		}
	})
}

func TestStability_ByteIdenticalAcrossRepeatedEvaluation(t *testing.T) {
	setRapidChecks(t, structuralChecks)

	rapid.Check(t, func(t *rapid.T) {
		tc := exprgen.SampleContext().Draw(t, "ctx")
		expr := drawExpression(t, tc)
		if strings.Contains(expr, "->") {
			return
		}
		exprTypeRaw, err := evalWithContext(fmt.Sprintf("typeOf(%s)", expr), tc.Value)
		if err != nil {
			return
		}
		exprType, ok := exprTypeRaw.(string)
		if !ok {
			return
		}
		if exprType == "Function" {
			return
		}

		first, err := evalWithContext(expr, tc.Value)
		if err != nil {
			return
		}
		second, err := evalWithContext(expr, tc.Value)
		if err != nil {
			return
		}
		third, err := evalWithContext(expr, tc.Value)
		if err != nil {
			return
		}

		firstBytes, err := formats.Format(first, "application/json")
		if err != nil {
			t.Fatalf("format first result failed: %v", err)
		}
		secondBytes, err := formats.Format(second, "application/json")
		if err != nil {
			t.Fatalf("format second result failed: %v", err)
		}
		thirdBytes, err := formats.Format(third, "application/json")
		if err != nil {
			t.Fatalf("format third result failed: %v", err)
		}

		if firstBytes != secondBytes || secondBytes != thirdBytes {
			t.Fatalf("non-byte-identical output:\nexpr: %q\nfirst:  %s\nsecond: %s\nthird:  %s", expr, firstBytes, secondBytes, thirdBytes)
		}
	})
}

func canonicalJSON(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}

func drawObjectLiteral(t *rapid.T, label string) string {
	return drawObjectLiteralAtDepth(t, label, 1)
}

func drawObjectLiteralAtDepth(t *rapid.T, label string, depth int) string {
	nFields := rapid.IntRange(0, 6).Draw(t, label+"_len")
	if nFields == 0 {
		return "{}"
	}

	fields := make([]string, 0, nFields)
	seen := map[string]struct{}{}
	for i := 0; i < nFields; i++ {
		key := rapid.StringMatching(`[A-Za-z_][A-Za-z0-9_]{0,10}`).Draw(t, fmt.Sprintf("%s_key_%d", label, i))
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		value := drawJSONSafeLiteralAtDepth(t, fmt.Sprintf("%s_val_%d", label, i), depth-1)
		fields = append(fields, fmt.Sprintf("%s: %s", strconv.Quote(key), value))
	}
	if len(fields) == 0 {
		return "{}"
	}
	return "{" + joinComma(fields) + "}"
}

func drawArrayOfArraysLiteral(t *rapid.T, label string) string {
	n := rapid.IntRange(0, 6).Draw(t, label+"_outer_len")
	if n == 0 {
		return "[]"
	}

	items := make([]string, n)
	for i := 0; i < n; i++ {
		items[i] = drawNonEmptyArrayLiteral(t, fmt.Sprintf("%s_inner_%d", label, i))
	}
	return "[" + joinComma(items) + "]"
}

func drawNonEmptyArrayLiteral(t *rapid.T, label string) string {
	length := rapid.IntRange(1, 6).Draw(t, label+"_len")
	items := make([]string, length)
	for i := 0; i < length; i++ {
		itemLabel := fmt.Sprintf("%s_item_%d", label, i)
		switch rapid.IntRange(0, 4).Draw(t, itemLabel+"_kind") {
		case 0:
			items[i] = strconv.Itoa(rapid.IntRange(-1000, 1000).Draw(t, itemLabel+"_int"))
		case 1:
			items[i] = drawSmallFloatLiteral(t, itemLabel+"_float")
		case 2:
			items[i] = safeStringLiteral(t, itemLabel+"_string")
		case 3:
			items[i] = exprgen.BoolLiteral().Draw(t, itemLabel+"_bool")
		default:
			items[i] = exprgen.NullLiteral().Draw(t, itemLabel+"_null")
		}
	}
	return "[" + joinComma(items) + "]"
}

func drawJSONSafeLiteral(t *rapid.T, label string) string {
	return drawJSONSafeLiteralAtDepth(t, label, 2)
}

func drawJSONSafeLiteralAtDepth(t *rapid.T, label string, depth int) string {
	if depth <= 0 {
		return drawAnyLiteral(t, label+"_leaf")
	}

	switch rapid.IntRange(0, 4).Draw(t, label+"_kind") {
	case 0:
		return drawAnyLiteral(t, label+"_scalar")
	case 1:
		return drawArrayLiteral(t, label+"_array")
	case 2:
		return drawObjectLiteralAtDepth(t, label+"_object", depth-1)
	case 3:
		return "null"
	default:
		return drawAnyLiteral(t, label+"_fallback")
	}
}
