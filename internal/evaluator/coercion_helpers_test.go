package evaluator

import "testing"

func TestCoerceEquals(t *testing.T) {
	tests := []struct {
		name  string
		left  Value
		right Value
		want  bool
	}{
		{name: "numeric int and float", left: 2, right: 2.0, want: true},
		{name: "numeric string and number", left: "2.5", right: 2.5, want: true},
		{name: "string fallback match", left: "abc", right: "abc", want: true},
		{name: "non-equal values", left: 1, right: "2", want: false},
		{
			name:  "recursive array coercion",
			left:  Array{"1", Object{"nested": "2.5"}},
			right: Array{1, Object{"nested": 2.5}},
			want:  true,
		},
		{
			name:  "recursive object coercion ignores field order",
			left:  Object{"first": "1", "second": Array{"true"}},
			right: Object{"second": Array{true}, "first": 1.0},
			want:  true,
		},
		{
			name:  "XML multi-value coercion",
			left:  XMLMultiValue{"1", Object{"value": "2"}},
			right: XMLMultiValue{1, Object{"value": 2}},
			want:  true,
		},
		{
			name:  "large integer string remains exact",
			left:  "9007199254740993",
			right: 9007199254740992,
			want:  false,
		},
		{name: "nil does not equal Go display artifact", left: nil, right: "<nil>", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CoerceEquals(tt.left, tt.right); got != tt.want {
				t.Fatalf("CoerceEquals(%v, %v) = %v, want %v", tt.left, tt.right, got, tt.want)
			}
		})
	}
}

func TestNumericEquals(t *testing.T) {
	tests := []struct {
		name  string
		left  Value
		right Value
		want  bool
	}{
		{name: "same int", left: 3, right: 3, want: true},
		{name: "int and float", left: 3, right: 3.0, want: true},
		{name: "different numbers", left: 3, right: 4.0, want: false},
		{name: "non-numeric", left: "3", right: 3.0, want: false},
		{
			name:  "nested numeric arrays",
			left:  Array{1, Array{2}},
			right: Array{1.0, Array{2.0}},
			want:  true,
		},
		{
			name:  "nested numeric objects ignore field order",
			left:  Object{"first": 1, "second": Object{"value": 2}},
			right: Object{"second": Object{"value": 2.0}, "first": 1.0},
			want:  true,
		},
		{
			name:  "XML multi-values compare recursively",
			left:  XMLMultiValue{1, Object{"value": 2}},
			right: XMLMultiValue{1.0, Object{"value": 2.0}},
			want:  true,
		},
		{
			name:  "different XML multi-values",
			left:  XMLMultiValue{1, 2},
			right: XMLMultiValue{1, 3},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := numericEquals(tt.left, tt.right); got != tt.want {
				t.Fatalf("numericEquals(%v, %v) = %v, want %v", tt.left, tt.right, got, tt.want)
			}
		})
	}
}

func TestNumericEqualsHandlesCyclicCollections(t *testing.T) {
	leftObject := Object{}
	leftObject["self"] = leftObject
	rightObject := Object{}
	rightObject["self"] = rightObject
	if !numericEquals(leftObject, rightObject) {
		t.Fatal("numericEquals rejected equivalent cyclic objects")
	}

	leftArray := make(Array, 1)
	leftArray[0] = leftArray
	rightArray := make(Array, 1)
	rightArray[0] = rightArray
	if !numericEquals(leftArray, rightArray) {
		t.Fatal("numericEquals rejected equivalent cyclic arrays")
	}
}

func TestAssertionEqualityUsesExactRecursiveComparison(t *testing.T) {
	if !isEqual(
		Object{"value": Array{1}},
		Object{"value": Array{1.0}},
	) {
		t.Fatal("isEqual rejected recursively equivalent numeric values")
	}
	if isEqual(9007199254740993, 9007199254740992.0) {
		t.Fatal("isEqual collapsed distinct integers above float precision")
	}
}
