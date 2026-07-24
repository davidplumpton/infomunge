package values

import (
	"reflect"
	"testing"
)

func TestObjectKeysPreservesTrackedInsertionOrder(t *testing.T) {
	object := NewObject(2)
	SetObjectValue(object, "b", 2)
	SetObjectValue(object, "a", 1)
	SetObjectValue(object, "b", 3)

	if got, want := ObjectKeys(object), []string{"b", "a"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ObjectKeys() = %v, want %v", got, want)
	}
}

func TestObjectKeysSortsUntrackedObjectsForCompatibility(t *testing.T) {
	object := Object{"b": 2, "a": 1}
	if got, want := ObjectKeys(object), []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ObjectKeys() = %v, want %v", got, want)
	}
}
