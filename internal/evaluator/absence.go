package evaluator

// absentValue is the evaluator-only representation of a field that is missing
// from an external input object. It stays distinct from language null so a
// surrounding default can handle failures caused by that missing field without
// swallowing the same failure when it was caused by an explicit null.
type absentValue struct {
	deferredErr error
}

func newAbsentValue() Value {
	return &absentValue{}
}

func asAbsentValue(value Value) (*absentValue, bool) {
	absent, ok := value.(*absentValue)
	return absent, ok
}

// evalAbsentAware materializes direct absence as null for an operation. A
// successful operation keeps its result (for example, missing == null is true);
// a nil result or error retains the absence provenance so default can handle
// it. An error already deferred by an inner expression takes precedence.
func evalAbsentAware(values []Value, evaluate func([]Value) (Value, error)) (Value, error) {
	materialized := make([]Value, len(values))
	var firstAbsent *absentValue
	for i, value := range values {
		absent, ok := asAbsentValue(value)
		if !ok {
			materialized[i] = value
			continue
		}
		if firstAbsent == nil {
			firstAbsent = absent
		}
		if absent.deferredErr != nil {
			return absent, nil
		}
		materialized[i] = nil
	}

	result, err := evaluate(materialized)
	if firstAbsent == nil {
		return result, err
	}
	if err != nil {
		return &absentValue{deferredErr: err}, nil
	}
	if result == nil {
		return firstAbsent, nil
	}
	return result, nil
}

func evalBuiltinAbsentAware(name string, values []Value, evaluate func([]Value) (Value, error)) (Value, error) {
	if builtinConsumesAbsentAsNull(name) {
		return evalAbsentAware(values, evaluate)
	}
	for _, value := range values {
		if absent, ok := asAbsentValue(value); ok {
			return absent, nil
		}
	}
	return evaluate(values)
}

// These builtins define meaningful behavior for language null rather than
// merely accepting it through broad coercion. Missing input values therefore
// participate in that behavior instead of bypassing the call.
func builtinConsumesAbsentAsNull(name string) bool {
	switch name {
	case "__isType", "contains", "endsWith", "every", "isBlank", "isEmpty",
		"some", "startsWith", "typeOf":
		return true
	default:
		return false
	}
}

// finalizeAbsentValue resolves absence at a semantic boundary. A bare missing
// selector becomes language null, while a composed expression replays the
// operation error that default would otherwise have intercepted.
func finalizeAbsentValue(value Value) (Value, error) {
	absent, ok := asAbsentValue(value)
	if !ok {
		return value, nil
	}
	if absent.deferredErr != nil {
		return nil, absent.deferredErr
	}
	return nil, nil
}

func evalCollectionLambdaWithBindingsAtDepth(lambda *Lambda, caller *Scope, depth int, bind func(Context)) (Value, error) {
	value, err := evalLambdaWithBindingsAtDepth(lambda, caller, depth, bind)
	if err != nil {
		return nil, err
	}
	return finalizeAbsentValue(value)
}
