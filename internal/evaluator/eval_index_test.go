package evaluator

import (
	"fmt"
	"go/token"
	"strings"
	"testing"

	"infomunge/pkg/values"
)

func TestEvalObjectOrdinalIndexResolvesNegativeIndexesFromEnd(t *testing.T) {
	object := values.NewObject(3)
	values.SetObjectValue(object, "second", 2)
	values.SetObjectValue(object, "first", 1)
	values.SetObjectValue(object, "third", 3)

	tests := []struct {
		name  string
		index int
		want  Value
	}{
		{name: "last", index: -1, want: 3},
		{name: "middle", index: -2, want: 1},
		{name: "first", index: -3, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evalObjectOrdinalIndex(object, tt.index, token.Pos(1))
			if err != nil {
				t.Fatalf("evalObjectOrdinalIndex(%d) returned an unexpected error: %v", tt.index, err)
			}
			if got != tt.want {
				t.Fatalf("evalObjectOrdinalIndex(%d) = %v, want %v", tt.index, got, tt.want)
			}
		})
	}
}

func TestEvalObjectOrdinalIndexRejectsIndexesOutsideBothBounds(t *testing.T) {
	object := values.NewObject(2)
	values.SetObjectValue(object, "first", 1)
	values.SetObjectValue(object, "second", 2)

	for _, index := range []int{-3, 2} {
		t.Run(fmt.Sprintf("index_%d", index), func(t *testing.T) {
			_, err := evalObjectOrdinalIndex(object, index, token.Pos(1))
			if err == nil {
				t.Fatalf("evalObjectOrdinalIndex(%d) returned nil error", index)
			}
			want := fmt.Sprintf("object index out of bounds: %d (object has 2 keys)", index)
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("evalObjectOrdinalIndex(%d) error %q does not contain %q", index, err, want)
			}
		})
	}
}
