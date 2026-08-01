package values

import (
	"reflect"
	"runtime"
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

func TestCloneObjectPreservesOrderAndDoesNotAliasTopLevelMap(t *testing.T) {
	object := NewObject(2)
	SetObjectValue(object, "b", 2)
	SetObjectValue(object, "a", 1)

	clone := CloneObject(object)
	SetObjectValue(clone, "b", 3)
	SetObjectValue(clone, "c", 4)

	if got, want := ObjectKeys(clone), []string{"b", "a", "c"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("clone keys = %v, want %v", got, want)
	}
	if got := object["b"]; got != 2 {
		t.Fatalf("source b = %v, want 2", got)
	}
	if _, exists := object["c"]; exists {
		t.Fatal("source unexpectedly contains clone-only key c")
	}
}

func TestCloneObjectPreservesAlphabeticalFallbackOrder(t *testing.T) {
	clone := CloneObject(Object{"b": 2, "a": 1})
	if got, want := ObjectKeys(clone), []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("clone keys = %v, want %v", got, want)
	}
}

func TestCloneObjectPreservesInputOriginPerObject(t *testing.T) {
	source := NewObject(1)
	MarkInputValue(source)
	scriptObject := NewObject(0)
	SetObjectValue(source, "script", scriptObject)

	clone := CloneObject(source)

	if !ObjectHasInputOrigin(clone) {
		t.Fatal("clone object was not marked as input")
	}
	if ObjectHasInputOrigin(clone["script"].(Object)) {
		t.Fatal("clone incorrectly marked a newly reachable script object as input")
	}
}

func TestMergeObjectsKeepsExistingPositionsAndAppendsNewKeys(t *testing.T) {
	left := NewObject(2)
	SetObjectValue(left, "b", 2)
	SetObjectValue(left, "a", 1)
	right := NewObject(2)
	SetObjectValue(right, "b", 20)
	SetObjectValue(right, "c", 3)

	merged := MergeObjects(left, right)

	if got, want := ObjectKeys(merged), []string{"b", "a", "c"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("merged keys = %v, want %v", got, want)
	}
	if got := merged["b"]; got != 20 {
		t.Fatalf("merged b = %v, want 20", got)
	}
}

func TestMarkInputValueMarksReachableObjectsOnly(t *testing.T) {
	nested := NewObject(1)
	SetObjectValue(nested, "value", 1)
	root := NewObject(2)
	SetObjectValue(root, "nested", nested)
	SetObjectValue(root, "items", Array{nested})
	scriptObject := NewObject(0)

	MarkInputValue(root)

	if !ObjectHasInputOrigin(root) {
		t.Fatal("root object was not marked as input")
	}
	if !ObjectHasInputOrigin(nested) {
		t.Fatal("nested object was not marked as input")
	}
	if ObjectHasInputOrigin(scriptObject) {
		t.Fatal("unreachable script object was marked as input")
	}
}

func TestObjectOrderIdentityExpiresWithObject(t *testing.T) {
	identity := unreachableObjectOrderIdentity()

	for range 10 {
		runtime.GC()
		if identity.Value() == nil {
			return
		}
		runtime.Gosched()
	}
	t.Fatal("object order identity still references an unreachable map")
}

func TestObjectOrderRegistryReclaimsHighChurnMetadata(t *testing.T) {
	runtime.GC()
	sweepDeadObjectOrders()
	baseline := objectOrderRegistrySize()

	survivor := NewObject(2)
	SetObjectValue(survivor, "b", 2)
	SetObjectValue(survivor, "a", 1)

	const (
		cycles          = 10
		objectsPerCycle = 1_000
	)
	for range cycles {
		createUnreachableOrderedObjects(objectsPerCycle)
		runtime.GC()
		sweepDeadObjectOrders()

		if got, wantMaximum := objectOrderRegistrySize(), baseline+1; got > wantMaximum {
			t.Fatalf("registry contains %d entries after collection, want at most %d", got, wantMaximum)
		}
		if got, want := ObjectKeys(survivor), []string{"b", "a"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("live object order after registry churn = %v, want %v", got, want)
		}
	}
	runtime.KeepAlive(survivor)
}

func unreachableObjectOrderIdentity() objectOrderIdentity {
	object := NewObject(1)
	SetObjectValue(object, "key", "value")
	identity, _ := objectIdentity(object)
	return identity
}

func createUnreachableOrderedObjects(count int) {
	for i := range count {
		object := NewObject(2)
		SetObjectValue(object, "second", i)
		SetObjectValue(object, "first", i)
	}
}

func objectOrderRegistrySize() int {
	size := 0
	objectOrders.Range(func(_, _ any) bool {
		size++
		return true
	})
	return size
}
