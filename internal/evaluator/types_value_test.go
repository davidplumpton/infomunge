package evaluator

import (
	"context"
	"testing"
)

func TestKindOf(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected ValueKind
	}{
		{"nil", nil, KindNil},
		{"bool true", true, KindBool},
		{"bool false", false, KindBool},
		{"int", 42, KindInt},
		{"float", 3.14, KindFloat},
		{"string", "hello", KindString},
		{"array", Array{1, 2, 3}, KindArray},
		{"object", Object{"key": "value"}, KindObject},
		{"lambda", &Lambda{}, KindLambda},
		{"type value", TypeValue("Boolean"), KindType},
		{"typedef", &TypeDef{Name: "Test"}, KindTypeDef},
		{"lazy", &LazyValue{}, KindLazy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := KindOf(tt.value)
			if got != tt.expected {
				t.Errorf("KindOf(%v) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

func TestLazyValue(t *testing.T) {
	evalCount := 0
	lv := &LazyValue{
		Eval: func(ctx context.Context) (Value, error) {
			evalCount++
			return 42, nil
		},
		ctx: context.Background(),
	}

	if lv.hasValue {
		t.Error("LazyValue should not have value initially")
	}

	val, err := lv.GetValue()
	if err != nil {
		t.Errorf("GetValue() error: %v", err)
	}
	if val != 42 {
		t.Errorf("GetValue() = %v, want 42", val)
	}
	if evalCount != 1 {
		t.Errorf("evalCount = %d, want 1", evalCount)
	}
	if !lv.hasValue {
		t.Error("LazyValue should have value after GetValue()")
	}

	// Second call should return cached value
	val, err = lv.GetValue()
	if err != nil {
		t.Errorf("GetValue() error: %v", err)
	}
	if val != 42 {
		t.Errorf("GetValue() = %v, want 42", val)
	}
	if evalCount != 1 {
		t.Errorf("evalCount = %d, want 1 (should be cached)", evalCount)
	}
}

func TestLazyValueStringDoesNotDrainStream(t *testing.T) {
	stream := make(chan Value, 1)
	stream <- 42
	close(stream)

	lv := &LazyValue{
		Eval: func(ctx context.Context) (Value, error) {
			return stream, nil
		},
		ctx: context.Background(),
	}

	got := lv.String()
	if got != "LazyValue(stream)" {
		t.Errorf("String() = %q, want %q", got, "LazyValue(stream)")
	}
	if len(stream) != 1 {
		t.Errorf("String() drained stream, len = %d, want 1", len(stream))
	}
}

func TestLazyValueStringUsesLanguageNullSpelling(t *testing.T) {
	lv := &LazyValue{
		Eval: func(ctx context.Context) (Value, error) {
			return nil, nil
		},
		ctx: context.Background(),
	}

	if got := lv.String(); got != "null" {
		t.Errorf("String() = %q, want %q", got, "null")
	}
}

func TestResolveValue(t *testing.T) {
	// Test immediate value
	val, err := ResolveValue(10)
	if err != nil {
		t.Errorf("ResolveValue(10) error: %v", err)
	}
	if val != 10 {
		t.Errorf("ResolveValue(10) = %v, want 10", val)
	}

	// Test lazy value
	lv := &LazyValue{
		Eval: func(ctx context.Context) (Value, error) {
			return "lazy", nil
		},
		ctx: context.Background(),
	}
	val, err = ResolveValue(lv)
	if err != nil {
		t.Errorf("ResolveValue(lazy) error: %v", err)
	}
	if val != "lazy" {
		t.Errorf("ResolveValue(lazy) = %v, want \"lazy\"", val)
	}
}

func TestValueKindString(t *testing.T) {
	tests := []struct {
		kind     ValueKind
		expected string
	}{
		{KindNil, "Nil"},
		{KindBool, "Boolean"},
		{KindInt, "Int"},
		{KindFloat, "Float"},
		{KindString, "String"},
		{KindArray, "Array"},
		{KindObject, "Object"},
		{KindLambda, "Lambda"},
		{KindType, "Type"},
		{KindTypeDef, "TypeDef"},
		{KindControlFlow, "ControlFlow"},
		{KindLazy, "Lazy"},
		{KindUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.expected {
				t.Errorf("%v.String() = %v, want %v", tt.kind, got, tt.expected)
			}
		})
	}
}

func TestAsArray(t *testing.T) {
	arr, ok := AsArray(Array{1, 2, 3})
	if !ok || len(arr) != 3 {
		t.Errorf("AsArray([1,2,3]) = %v, %v, want [1,2,3], true", arr, ok)
	}

	_, ok = AsArray("hello")
	if ok {
		t.Error("AsArray(\"hello\") should return false")
	}
}

func TestAsObject(t *testing.T) {
	obj, ok := AsObject(Object{"key": "value"})
	if !ok || obj["key"] != "value" {
		t.Errorf("AsObject({key:value}) = %v, %v, want {key:value}, true", obj, ok)
	}

	_, ok = AsObject("hello")
	if ok {
		t.Error("AsObject(\"hello\") should return false")
	}
}

func TestAsLambda(t *testing.T) {
	lambda := &Lambda{Params: []ParamDef{{Name: "x", ExpectedKind: KindUnknown}}}
	l, ok := AsLambda(lambda)
	if !ok || l != lambda {
		t.Error("AsLambda should return the lambda and true")
	}

	_, ok = AsLambda("hello")
	if ok {
		t.Error("AsLambda(\"hello\") should return false")
	}
}
