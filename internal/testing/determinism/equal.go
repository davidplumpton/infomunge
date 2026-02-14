package determinism

import (
	"math"
	"reflect"

	"infomunge/internal/evaluator"
)

// Equal reports whether two evaluation results are deterministic equivalents.
func Equal(first, second interface{}) bool {
	return equalValues(first, second)
}

func equalValues(first, second interface{}) bool {
	if first == nil || second == nil {
		return first == second
	}

	if reflect.DeepEqual(first, second) {
		return true
	}

	switch left := first.(type) {
	case float64:
		right, ok := second.(float64)
		if !ok {
			return false
		}
		if math.IsNaN(left) && math.IsNaN(right) {
			return true
		}
		return left == right
	case float32:
		right, ok := second.(float32)
		if !ok {
			return false
		}
		if math.IsNaN(float64(left)) && math.IsNaN(float64(right)) {
			return true
		}
		return left == right
	case []interface{}:
		right, ok := second.([]interface{})
		if !ok || len(left) != len(right) {
			return false
		}
		for i := range left {
			if !equalValues(left[i], right[i]) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		right, ok := second.(map[string]interface{})
		if !ok || len(left) != len(right) {
			return false
		}
		for key, leftValue := range left {
			rightValue, exists := right[key]
			if !exists || !equalValues(leftValue, rightValue) {
				return false
			}
		}
		return true
	case *evaluator.Lambda:
		right, ok := second.(*evaluator.Lambda)
		if !ok {
			return false
		}
		return equalLambdas(left, right)
	default:
		return false
	}
}

func equalLambdas(first, second *evaluator.Lambda) bool {
	if len(first.Params) != len(second.Params) {
		return false
	}

	for i := range first.Params {
		left := first.Params[i]
		right := second.Params[i]
		if left.Name != right.Name || left.ExpectedKind != right.ExpectedKind || left.HasDefault != right.HasDefault {
			return false
		}
		if left.HasDefault && !equalValues(left.DefaultValue, right.DefaultValue) {
			return false
		}
	}

	return true
}
