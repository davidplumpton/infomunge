package determinism

import (
	"math"
	"testing"

	"infomunge/internal/evaluator"
)

func TestEqualTreatsNaNAsDeterministic(t *testing.T) {
	if !Equal(math.NaN(), math.NaN()) {
		t.Fatal("expected NaN and NaN to be deterministic equals")
	}

	first := evaluator.Object{
		"value": evaluator.Array{math.NaN()},
	}
	second := evaluator.Object{
		"value": evaluator.Array{math.NaN()},
	}
	if !Equal(first, second) {
		t.Fatal("expected nested NaN values to be deterministic equals")
	}
}

func TestEqualTreatsEquivalentLambdasAsDeterministic(t *testing.T) {
	first := &evaluator.Lambda{
		Params: []evaluator.ParamDef{
			{Name: "x", ExpectedKind: evaluator.KindUnknown},
		},
		Body: "unstable pointer representation 1",
	}
	second := &evaluator.Lambda{
		Params: []evaluator.ParamDef{
			{Name: "x", ExpectedKind: evaluator.KindUnknown},
		},
		Body: "unstable pointer representation 2",
	}

	if !Equal(first, second) {
		t.Fatal("expected lambdas with equivalent semantic params to be deterministic equals")
	}
}

func TestEqualRejectsDifferentLambdas(t *testing.T) {
	first := &evaluator.Lambda{
		Params: []evaluator.ParamDef{
			{Name: "x", ExpectedKind: evaluator.KindUnknown},
		},
	}
	second := &evaluator.Lambda{
		Params: []evaluator.ParamDef{
			{Name: "y", ExpectedKind: evaluator.KindUnknown},
		},
	}

	if Equal(first, second) {
		t.Fatal("expected lambdas with different params to not be deterministic equals")
	}
}
