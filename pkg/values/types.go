package values

import (
	"reflect"
	"sort"
	"sync"
)

// Value is the shared runtime value shape used by evaluation and format adapters.
type Value = interface{}

// Object represents a structured data object with string keys.
type Object = map[string]Value

type objectOrder struct {
	mu   sync.RWMutex
	keys []string
}

var objectOrders sync.Map

func objectIdentity(object Object) uintptr {
	if object == nil {
		return 0
	}
	return reflect.ValueOf(object).Pointer()
}

// NewObject constructs an object whose insertion order is tracked.
func NewObject(capacity int) Object {
	object := make(Object, capacity)
	RegisterObjectOrder(object, nil)
	return object
}

// RegisterObjectOrder records the known key order for an object.
func RegisterObjectOrder(object Object, keys []string) {
	identity := objectIdentity(object)
	if identity == 0 {
		return
	}
	copied := append([]string(nil), keys...)
	objectOrders.Store(identity, &objectOrder{keys: copied})
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
	identity := objectIdentity(object)
	recordValue, loaded := objectOrders.Load(identity)
	if !loaded {
		recordValue, _ = objectOrders.LoadOrStore(identity, &objectOrder{})
	}
	record := recordValue.(*objectOrder)
	record.mu.Lock()
	record.keys = append(record.keys, key)
	record.mu.Unlock()
}

// ObjectKeys returns tracked insertion order. Objects created outside the
// ordered constructors retain the previous deterministic alphabetical fallback.
func ObjectKeys(object Object) []string {
	if object == nil {
		return nil
	}
	identity := objectIdentity(object)
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
