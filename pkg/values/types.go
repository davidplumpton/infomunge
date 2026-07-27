package values

import (
	"reflect"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"weak"
)

// Value is the shared runtime value shape used by evaluation and format adapters.
type Value = interface{}

// TypeValue is the runtime result of typeOf. Its string representation is the
// type name, but it remains distinct from an ordinary String value.
type TypeValue string

// String returns the represented type name.
func (value TypeValue) String() string {
	return string(value)
}

// Object represents a structured data object with string keys.
type Object = map[string]Value

type objectOrder struct {
	mu          sync.RWMutex
	keys        []string
	inputOrigin bool
}

var objectOrders sync.Map

var objectOrderRegistrations atomic.Uint64

const objectOrderSweepInterval = 64

type objectOrderIdentity = weak.Pointer[byte]

// objectIdentity uses the runtime map allocation as a weak registry key. Weak
// pointers distinguish allocation generations, so a later map that reuses the
// same address cannot observe the reclaimed map's order metadata.
func objectIdentity(object Object) (objectOrderIdentity, *byte) {
	if object == nil {
		return objectOrderIdentity{}, nil
	}
	pointer := (*byte)(reflect.ValueOf(object).UnsafePointer())
	return weak.Make(pointer), pointer
}

func removeObjectOrder(identity objectOrderIdentity) {
	objectOrders.Delete(identity)
}

func installObjectOrderCleanup(object Object, identity objectOrderIdentity, pointer *byte) {
	runtime.AddCleanup(pointer, removeObjectOrder, identity)
	if objectOrderRegistrations.Add(1)%objectOrderSweepInterval == 0 {
		sweepDeadObjectOrders()
	}
	runtime.KeepAlive(object)
}

func loadOrCreateObjectOrder(object Object) *objectOrder {
	identity, pointer := objectIdentity(object)
	recordValue, loaded := objectOrders.Load(identity)
	if !loaded {
		recordValue, loaded = objectOrders.LoadOrStore(identity, &objectOrder{})
		if !loaded {
			installObjectOrderCleanup(object, identity, pointer)
		}
	}
	return recordValue.(*objectOrder)
}

func sweepDeadObjectOrders() {
	objectOrders.Range(func(key, _ any) bool {
		identity := key.(objectOrderIdentity)
		if identity.Value() == nil {
			objectOrders.Delete(identity)
		}
		return true
	})
}

// NewObject constructs an object whose insertion order is tracked.
func NewObject(capacity int) Object {
	object := make(Object, capacity)
	identity, pointer := objectIdentity(object)
	_, loaded := objectOrders.Swap(identity, &objectOrder{})
	if !loaded {
		installObjectOrderCleanup(object, identity, pointer)
	}
	return object
}

// SetObjectValue sets a field and records its first insertion position.
func SetObjectValue(object Object, key string, value Value) {
	if object == nil {
		return
	}
	_, existed := object[key]
	object[key] = value
	if existed {
		return
	}
	record := loadOrCreateObjectOrder(object)
	record.mu.Lock()
	record.keys = append(record.keys, key)
	record.mu.Unlock()
}

// MarkInputValue records that objects reachable from value originated outside
// the transformation script. DataWeave preserves a distinct missing-field
// state for input objects, while object literals created by the script yield
// ordinary null for a missing field.
func MarkInputValue(value Value) {
	visitedObjects := make(map[objectOrderIdentity]struct{})
	visitedArrays := make(map[uintptr]struct{})

	var mark func(Value)
	mark = func(current Value) {
		switch typed := current.(type) {
		case Object:
			if typed == nil {
				return
			}
			identity, _ := objectIdentity(typed)
			if _, seen := visitedObjects[identity]; seen {
				return
			}
			visitedObjects[identity] = struct{}{}

			record := loadOrCreateObjectOrder(typed)
			record.mu.Lock()
			record.inputOrigin = true
			record.mu.Unlock()
			for _, nested := range typed {
				mark(nested)
			}
		case Array:
			markInputArray(typed, visitedArrays, mark)
		case XMLMultiValue:
			markInputArray(Array(typed), visitedArrays, mark)
		}
	}

	mark(value)
}

func markInputArray(array Array, visited map[uintptr]struct{}, mark func(Value)) {
	if len(array) == 0 {
		return
	}
	identity := reflect.ValueOf(array).Pointer()
	if _, seen := visited[identity]; seen {
		return
	}
	visited[identity] = struct{}{}
	for _, nested := range array {
		mark(nested)
	}
}

// ObjectHasInputOrigin reports whether object was marked as external input.
func ObjectHasInputOrigin(object Object) bool {
	if object == nil {
		return false
	}
	identity, _ := objectIdentity(object)
	recordValue, ok := objectOrders.Load(identity)
	if !ok {
		return false
	}
	record := recordValue.(*objectOrder)
	record.mu.RLock()
	defer record.mu.RUnlock()
	return record.inputOrigin
}

// CloneObject creates a shallow ordered copy of object.
func CloneObject(object Object) Object {
	if object == nil {
		return nil
	}
	clone := NewObject(len(object))
	for _, key := range ObjectKeys(object) {
		SetObjectValue(clone, key, object[key])
	}
	return clone
}

// MergeObjects creates an ordered shallow merge. Existing keys keep their
// original positions, while keys first seen in later objects are appended.
func MergeObjects(objects ...Object) Object {
	capacity := 0
	for _, object := range objects {
		capacity += len(object)
	}
	result := NewObject(capacity)
	for _, object := range objects {
		for _, key := range ObjectKeys(object) {
			SetObjectValue(result, key, object[key])
		}
	}
	return result
}

// ObjectKeys returns tracked insertion order. Objects created outside the
// ordered constructors retain the previous deterministic alphabetical fallback.
func ObjectKeys(object Object) []string {
	if object == nil {
		return nil
	}
	identity, _ := objectIdentity(object)
	if recordValue, ok := objectOrders.Load(identity); ok {
		record := recordValue.(*objectOrder)
		record.mu.RLock()
		tracked := append([]string(nil), record.keys...)
		record.mu.RUnlock()

		keys := make([]string, 0, len(object))
		seen := make(map[string]struct{}, len(tracked))
		for _, key := range tracked {
			if _, exists := object[key]; exists {
				keys = append(keys, key)
				seen[key] = struct{}{}
			}
		}
		remaining := make([]string, 0, len(object)-len(keys))
		for key := range object {
			if _, exists := seen[key]; !exists {
				remaining = append(remaining, key)
			}
		}
		sort.Strings(remaining)
		return append(keys, remaining...)
	}

	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Array represents an ordered sequence of values.
type Array = []Value

// XMLMultiValue represents duplicate XML element values while preserving that
// they came from repeated object keys rather than a source array.
type XMLMultiValue []Value

// Namespace represents an XML namespace with an optional prefix.
// Prefix "" indicates the default namespace.
type Namespace struct {
	Prefix string
	URI    string
}
