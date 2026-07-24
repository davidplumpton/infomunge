package exprgen

import (
	"reflect"
	"testing"
)

func TestDWCompatOperatorsReplacePercentWithInfixMod(t *testing.T) {
	got := (exprConfig{DWCompat: true}).filterOps([]string{"%", "+", "++", "*"})
	want := []string{"mod", "+", "*"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DW-compatible operators = %v, want %v", got, want)
	}
}
