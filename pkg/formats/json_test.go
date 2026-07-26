package formats

import (
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestReadJSONPreservesExactIntegersAndDecimalPolicy(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1

	type testCase struct {
		name string
		text string
		want interface{}
	}
	tests := []testCase{
		{name: "maximum runtime integer", text: strconv.Itoa(maxInt), want: maxInt},
		{name: "minimum runtime integer", text: strconv.Itoa(minInt), want: minInt},
		{name: "decimal remains float", text: "0.1", want: float64(0.1)},
		{name: "exponent remains float", text: "1e3", want: float64(1000)},
	}
	if strconv.IntSize >= 64 {
		belowFloatBoundary, _ := strconv.ParseInt("9007199254740991", 10, 64)
		aboveFloatBoundary, _ := strconv.ParseInt("9007199254740993", 10, 64)
		tests = append(tests,
			testCase{name: "below exact float boundary", text: "9007199254740991", want: int(belowFloatBoundary)},
			testCase{name: "above exact float boundary", text: "9007199254740993", want: int(aboveFloatBoundary)},
		)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Read(test.text, "application/json")
			if err != nil {
				t.Fatalf("Read(%q) error = %v", test.text, err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("Read(%q) = %#v (%T), want %#v (%T)", test.text, got, got, test.want, test.want)
			}
		})
	}
}

func TestReadJSONRejectsIntegersOutsideRuntimeRange(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	one := big.NewInt(1)
	aboveMax := new(big.Int).Add(big.NewInt(int64(maxInt)), one).String()
	belowMin := new(big.Int).Sub(big.NewInt(int64(minInt)), one).String()

	for _, text := range []string{aboveMax, belowMin} {
		_, err := Read(text, "application/json")
		if err == nil {
			t.Fatalf("Read(%q) error = nil, want integer range error", text)
		}
		if !strings.Contains(err.Error(), "outside the supported numeric range") {
			t.Fatalf("Read(%q) error = %q, want numeric range context", text, err)
		}
	}
}

func TestReadJSONRejectsNonFiniteDecimalAndTrailingContent(t *testing.T) {
	_, err := Read("1e400", "application/json")
	if err == nil {
		t.Fatal("Read(non-finite decimal) error = nil, want numeric range error")
	}
	if !strings.Contains(err.Error(), "outside the supported numeric range") {
		t.Fatalf("Read(non-finite decimal) error = %q, want numeric range context", err)
	}

	_, err = Read("1 2", "application/json")
	if err == nil {
		t.Fatal("Read(trailing content) error = nil, want trailing-content error")
	}
	if !strings.Contains(err.Error(), "unexpected trailing JSON content") {
		t.Fatalf("Read(trailing content) error = %q, want trailing-content context", err)
	}
}

func TestReadNDJSONUsesExactJSONNumberPolicy(t *testing.T) {
	content := "0.1\n"
	want := Array{float64(0.1)}
	if strconv.IntSize >= 64 {
		belowFloatBoundary, _ := strconv.ParseInt("9007199254740991", 10, 64)
		aboveFloatBoundary, _ := strconv.ParseInt("9007199254740993", 10, 64)
		content = "9007199254740991\n9007199254740993\n" + content
		want = Array{int(belowFloatBoundary), int(aboveFloatBoundary), float64(0.1)}
	}

	got, err := Read(content, "application/x-ndjson")
	if err != nil {
		t.Fatalf("Read(NDJSON) error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Read(NDJSON) = %#v, want %#v", got, want)
	}

	maxInt := int(^uint(0) >> 1)
	aboveMax := new(big.Int).Add(big.NewInt(int64(maxInt)), big.NewInt(1)).String()
	_, err = Read(aboveMax+"\n", "application/x-ndjson")
	if err == nil {
		t.Fatal("Read(out-of-range NDJSON) error = nil, want integer range error")
	}
	if !strings.Contains(err.Error(), "NDJSON parse error on line 1") {
		t.Fatalf("Read(out-of-range NDJSON) error = %q, want line context", err)
	}
}
