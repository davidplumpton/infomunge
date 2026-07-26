package evaluator

import (
	"go/ast"
	"math"
	"strconv"
	"testing"
)

func TestSizeOfNumbersUsesDataWeaveRendering(t *testing.T) {
	call := &ast.CallExpr{Fun: &ast.Ident{Name: "sizeOf"}}
	testCases := []struct {
		name  string
		value Value
		want  int
	}{
		{name: "zero integer", value: 0, want: 1},
		{name: "positive integer", value: 211, want: 3},
		{name: "negative integer", value: -211, want: 4},
		{name: "maximum integer without float conversion", value: math.MaxInt, want: len(strconv.Itoa(math.MaxInt))},
		{name: "minimum integer without float conversion", value: math.MinInt, want: len(strconv.Itoa(math.MinInt))},
		{name: "ordinary decimal", value: 123.45, want: 6},
		{name: "negative decimal", value: -123.45, want: 7},
		{name: "small plain decimal", value: 1e-6, want: 8},
		{name: "small exponent", value: 1e-7, want: 4},
		{name: "positive exponent", value: 1e10, want: 5},
		{name: "maximum float", value: math.MaxFloat64, want: 23},
		{name: "negative maximum float", value: -math.MaxFloat64, want: 24},
		{name: "smallest nonzero float", value: math.SmallestNonzeroFloat64, want: 6},
		{name: "negative zero", value: math.Copysign(0, -1), want: 1},
		{name: "type value", value: TypeValue("Boolean"), want: 7},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := callBuiltinSizeOf([]Value{testCase.value}, call)
			if err != nil {
				t.Fatalf("callBuiltinSizeOf() error = %v", err)
			}
			if got != testCase.want {
				t.Fatalf("callBuiltinSizeOf(%v) = %v, want %d", testCase.value, got, testCase.want)
			}
		})
	}
}
