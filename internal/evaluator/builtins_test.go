package evaluator

import (
	"reflect"
	"testing"
)

func TestEvaluate_ObjectBuiltins(t *testing.T) {
	ctx := Object{
		"obj": Object{"a": 1, "b": 2},
	}
	mapping := make([]int, 500)
	for i := range mapping {
		mapping[i] = i
	}

	tests := []struct {
		name      string
		expr      string
		expected  interface{}
		expectErr bool
	}{
		{
			"mapObject simple",
			`mapObject(obj, __lambda("v, k", []interface{}{k + "!", v * 10}))`,
			Object{"a!": 10, "b!": 20},
			false,
		},
		{
			"filterObject simple",
			`filterObject(obj, __lambda("k, v", v > 1))`,
			Object{"b": 2},
			false,
		},
		{
			"pluck on object",
			`__pluck(obj, __lambda("v, k", k))`,
			Array{"a", "b"},
			false,
		},
		{
			"reduce sum",
			`__reduce([]interface{}{1, 2, 3, 4}, __lambda("acc, v", acc + v))`,
			10,
			false,
		},
		{
			"flatten simple",
			`flatten([]interface{}{[]interface{}{1, 2}, []interface{}{3, 4}} )`,
			Array{1, 2, 3, 4},
			false,
		},
		{
			"assert success",
			`assert(10, beNumber())`,
			10,
			false,
		},
		{
			"assertThat success",
			`assertThat("hello", beString())`,
			"hello",
			false,
		},
		{
			"assert with message success",
			`assert(true, beBoolean(), "must be bool")`,
			true,
			false,
		},
		{
			"assert failure",
			`assert(10, beString())`,
			nil,
			true,
		},
		{
			"assert failure with message",
			`assert(10, beString(), "should be string")`,
			nil,
			true,
		},
		{
			"lastIndexOf string",
			`lastIndexOf("hello", "l")`,
			3,
			false,
		},
		{
			"lastIndexOf array",
			`lastIndexOf([]interface{}{1, 2, 1}, 1)`,
			2,
			false,
		},
		{
			"lastIndexOf not found",
			`lastIndexOf("hello", "z")`,
			-1,
			false,
		},
		{
			"log returns value",
			`log("hello")`,
			"hello",
			false,
		},
		{
			"logDebug returns value",
			`logDebug("hello")`,
			"hello",
			false,
		},
		{
			"logInfo returns value",
			`logInfo("hello")`,
			"hello",
			false,
		},
		{
			"logWarn returns value",
			`logWarn("hello")`,
			"hello",
			false,
		},
		{
			"logError returns value",
			`logError("hello")`,
			"hello",
			false,
		},
		{
			"logWith returns value",
			`logWith("hello", "Processing")`,
			"hello",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.expr, ctx, mapping, 0, tt.expr)
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v (%T), got %v (%T)", tt.expected, tt.expected, result, result)
			}
		})
	}
}
