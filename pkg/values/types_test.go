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
