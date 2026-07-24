package evaluator

import (
	"context"
	"fmt"
	"infomunge/pkg/values"
	"sync"
)

// Value is a generic value type that represents any value that can be evaluated in expressions.
// It can be a nil, bool, string, number (int or float64), array, object, lambda function, TypeDef, or ControlFlowSignal.
type Value = values.Value

// Context is a type alias for the evaluation context (namespace mapping).
type Context = map[string]Value

// GoContextKey is the legacy key used by old API callers to pass a
// context.Context into NewScope. Runtime code should use Scope.GoContext.
const GoContextKey = "__go_ctx__"

// Object represents a structured data object (map with string keys).
type Object = values.Object

// Array represents a sequence of values.
type Array = values.Array

// XMLMultiValue represents an array of values that were originally
// duplicate keys in an XML object. This helps mapObject distinguish
// between a natural array and repeated elements.
type XMLMultiValue = values.XMLMultiValue

// Namespace represents an XML namespace with an optional prefix.
type Namespace = values.Namespace

// ValueKind represents the discriminated type of a Value.
// This provides a type-safe way to check what kind of value we have.
type ValueKind int

const (
	KindNil ValueKind = iota
	KindBool
	KindInt
	KindFloat
	KindString
	KindArray
	KindObject
	KindLambda
	KindTypeDef
	KindNamespace
	KindControlFlow
	KindLazy
	KindRegex
	KindUnknown
)

// String returns the string representation of a ValueKind.
func (k ValueKind) String() string {
	switch k {
	case KindNil:
		return "Nil"
	case KindBool:
		return "Boolean"
	case KindInt:
		return "Int"
	case KindFloat:
		return "Float"
	case KindString:
		return "String"
	case KindArray:
		return "Array"
	case KindObject:
		return "Object"
	case KindLambda:
		return "Lambda"
	case KindTypeDef:
		return "TypeDef"
	case KindNamespace:
		return "Namespace"
	case KindControlFlow:
		return "ControlFlow"
	case KindLazy:
		return "Lazy"
	case KindRegex:
		return "Regex"
	default:
		return "Unknown"
	}
}

// LazyValue represents a value that is evaluated only when needed.
// It wraps an evaluation function and caches the result for subsequent accesses.
type LazyValue struct {
	Eval        func(ctx context.Context) (Value, error)
	ctx         context.Context
	mu          sync.Mutex
	cachedValue Value
	hasValue    bool
}

// NewLazyValue creates a new LazyValue with the given evaluation function and context.
func NewLazyValue(eval func(ctx context.Context) (Value, error), ctx context.Context) *LazyValue {
	return &LazyValue{
		Eval:     eval,
		ctx:      ctx,
		hasValue: false,
	}
}

// GetValue returns the evaluated value, evaluating it if necessary.
func (lv *LazyValue) GetValue() (Value, error) {
	lv.mu.Lock()
	defer lv.mu.Unlock()
	if !lv.hasValue {
		val, err := lv.Eval(lv.ctx)
		if err != nil {
			return nil, err
		}
		lv.cachedValue = val
		lv.hasValue = true
	}
	return lv.cachedValue, nil
}

// String returns a string representation of the LazyValue by evaluating it.
func (lv *LazyValue) String() string {
	if val, err := lv.GetValue(); err == nil {
		if _, _, ok := streamFromValue(val); ok {
			return "LazyValue(stream)"
		}
		return fmt.Sprintf("%v", val)
	}
	return "LazyValue"
}

// ResolveValue evaluates a value if it's lazy, otherwise returns it as-is.
// This is useful for functions or operators that need the actual value.
func ResolveValue(v Value) (Value, error) {
	if lv, ok := v.(*LazyValue); ok {
		return lv.GetValue()
	}
	return v, nil
}

// KindOf returns the ValueKind for a given Value.
// This enables type-safe switching on value types.
func KindOf(v Value) ValueKind {
	switch v.(type) {
	case nil:
		return KindNil
	case bool:
		return KindBool
	case int:
		return KindInt
	case float64:
		return KindFloat
	case string:
		return KindString
	case Array, XMLMultiValue:
		return KindArray
	case Object:
		return KindObject
	case *Lambda:
		return KindLambda
	case *TypeDef:
		return KindTypeDef
	case Namespace:
		return KindNamespace
	case *ControlFlowSignal:
		return KindControlFlow
	case *LazyValue:
		return KindLazy
	case *Regex:
		return KindRegex
	default:
		return KindUnknown
	}
}

// AsArray extracts an Array from a Value, returning false if not an array.
func AsArray(v Value) (Array, bool) {
	switch arr := v.(type) {
	case Array:
		return arr, true
	case XMLMultiValue:
		return Array(arr), true
	default:
		return nil, false
	}
}

// AsObject extracts an Object from a Value, returning false if not an object.
func AsObject(v Value) (Object, bool) {
	o, ok := v.(Object)
	return o, ok
}

// AsLambda extracts a Lambda from a Value, returning false if not a lambda.
func AsLambda(v Value) (*Lambda, bool) {
	l, ok := v.(*Lambda)
	return l, ok
}

// ParamDef defines a lambda parameter with optional type constraint and default value.
// This provides type-safe handling of function parameter definitions.
type ParamDef struct {
	Name         string    // Parameter name
	ExpectedKind ValueKind // Expected type constraint (KindUnknown means no constraint)
	DefaultValue Value     // Default value (nil if HasDefault is false)
	HasDefault   bool      // True if parameter has an explicit default value
}
