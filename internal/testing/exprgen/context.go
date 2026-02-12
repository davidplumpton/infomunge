package exprgen

import (
	"encoding/json"
	"fmt"

	"pgregory.net/rapid"
)

var defaultContextFieldNames = []string{
	"name", "age", "active", "score", "tags",
	"address", "items", "count", "value", "label",
}

// TestContext holds a generated test context with both the Go value
// (for passing to the runner) and the JSON string (for CLI input).
type TestContext struct {
	// Value is the Go map suitable for passing as an evaluator context.
	Value map[string]interface{}
	// JSON is the JSON-encoded string of Value["payload"].
	JSON string
	// Fields lists the top-level field names available in the payload.
	Fields []string
}

// SampleContext returns a generator that produces test contexts with a
// "payload" variable containing a mix of value types.
func SampleContext() *rapid.Generator[TestContext] {
	return rapid.Custom(func(t *rapid.T) TestContext {
		payload := samplePayload(t)
		ctx := map[string]interface{}{
			"payload": payload,
		}
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}

		var fields []string
		if obj, ok := payload.(map[string]interface{}); ok {
			for k := range obj {
				fields = append(fields, k)
			}
		}

		return TestContext{
			Value:  ctx,
			JSON:   string(jsonBytes),
			Fields: fields,
		}
	})
}

// samplePayload generates a payload value: an object with mixed field types.
func samplePayload(t *rapid.T) interface{} {
	nFields := rapid.IntRange(1, 6).Draw(t, "nFields")
	obj := make(map[string]interface{}, nFields)
	for i := 0; i < nFields; i++ {
		key := fieldName(i)
		obj[key] = sampleValue(t, 2)
	}
	return obj
}

// fieldName returns a deterministic field name for a given index.
func fieldName(i int) string {
	if i < len(defaultContextFieldNames) {
		return defaultContextFieldNames[i]
	}
	return fmt.Sprintf("field%d", i)
}

// ContextShapeFields returns a copy of the top-level field names used by
// SampleContext payload generation.
func ContextShapeFields() []string {
	fields := make([]string, len(defaultContextFieldNames))
	copy(fields, defaultContextFieldNames)
	return fields
}

// sampleValue generates a random JSON-compatible value with depth control.
func sampleValue(t *rapid.T, depth int) interface{} {
	if depth <= 0 {
		return samplePrimitive(t)
	}
	// 0-4: primitive (50%), 5-6: nested object (20%), 7-8: array (20%), 9: null/empty (10%)
	kind := rapid.IntRange(0, 9).Draw(t, "valueKind")
	switch {
	case kind <= 4:
		return samplePrimitive(t)
	case kind <= 6:
		return sampleNestedObject(t, depth-1)
	case kind <= 8:
		return sampleArray(t, depth-1)
	default:
		// null or empty object/array
		choice := rapid.IntRange(0, 2).Draw(t, "emptyChoice")
		switch choice {
		case 0:
			return nil
		case 1:
			return map[string]interface{}{}
		default:
			return []interface{}{}
		}
	}
}

// samplePrimitive generates a string, number, or boolean.
func samplePrimitive(t *rapid.T) interface{} {
	kind := rapid.IntRange(0, 3).Draw(t, "primKind")
	switch kind {
	case 0: // string
		return rapid.StringOfN(
			rapid.RuneFrom([]rune("abcdefghijklmnopqrstuvwxyz ")),
			0, 15, -1,
		).Draw(t, "str")
	case 1: // integer
		return rapid.IntRange(-1000, 1000).Draw(t, "int")
	case 2: // float
		return rapid.Float64Range(-100.0, 100.0).Draw(t, "float")
	default: // boolean
		return rapid.Bool().Draw(t, "bool")
	}
}

// sampleNestedObject generates a small object (1-3 fields).
func sampleNestedObject(t *rapid.T, depth int) interface{} {
	nFields := rapid.IntRange(1, 3).Draw(t, "nestedFields")
	obj := make(map[string]interface{}, nFields)
	for i := 0; i < nFields; i++ {
		key := fieldName(i)
		obj[key] = sampleValue(t, depth)
	}
	return obj
}

// sampleArray generates an array of 0-4 elements.
func sampleArray(t *rapid.T, depth int) interface{} {
	strategy := rapid.IntRange(0, 2).Draw(t, "arrStrategy")
	switch strategy {
	case 0: // array of primitives
		n := rapid.IntRange(0, 4).Draw(t, "arrLen")
		arr := make([]interface{}, n)
		for i := range arr {
			arr[i] = samplePrimitive(t)
		}
		return arr
	case 1: // array of objects
		n := rapid.IntRange(1, 3).Draw(t, "arrLen")
		arr := make([]interface{}, n)
		for i := range arr {
			arr[i] = sampleNestedObject(t, depth)
		}
		return arr
	default: // mixed
		n := rapid.IntRange(0, 4).Draw(t, "arrLen")
		arr := make([]interface{}, n)
		for i := range arr {
			arr[i] = sampleValue(t, depth)
		}
		return arr
	}
}

// IdentGen returns a generator that produces valid variable references for the
// given test context. It picks from "payload" and "payload.field" forms.
func IdentGen(tc TestContext) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		if len(tc.Fields) == 0 {
			return "payload"
		}
		// 0: bare payload (30%), 1-2: payload.field (70%)
		kind := rapid.IntRange(0, 2).Draw(t, "identKind")
		if kind == 0 {
			return "payload"
		}
		idx := rapid.IntRange(0, len(tc.Fields)-1).Draw(t, "fieldIdx")
		return "payload." + tc.Fields[idx]
	})
}
