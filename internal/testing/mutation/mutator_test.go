package mutation

import (
	"math/rand"
	"testing"

	"infomunge/internal/testing/exprgen"
)

func TestOperatorSwapChangesExpression(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	input := "1 + 2 == 3"
	got := OperatorSwap(input, rng)
	if got == input {
		t.Fatalf("OperatorSwap did not change expression: %s", DebugDescribeMutation(input, got))
	}
	if !exprgen.IsValid(got) {
		t.Fatalf("OperatorSwap produced invalid expression: %s", got)
	}
}

func TestLiteralPerturbChangesExpression(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	input := `"abc" default true`
	got := LiteralPerturb(input, rng)
	if got == input {
		t.Fatalf("LiteralPerturb did not change expression: %s", DebugDescribeMutation(input, got))
	}
	if !exprgen.IsValid(got) {
		t.Fatalf("LiteralPerturb produced invalid expression: %s", got)
	}
}

func TestSubtreeSwapChangesExpression(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	a := "(1 + 2) * (3 + 4)"
	b := "(9 - 8) / (7 - 6)"
	got := SubtreeSwap(a, b, rng)
	if got == a {
		t.Fatalf("SubtreeSwap did not change expression: %s", DebugDescribeMutation(a, got))
	}
	if !exprgen.IsValid(got) {
		t.Fatalf("SubtreeSwap produced invalid expression: %s", got)
	}
}

func TestArgManipulateChangesExpression(t *testing.T) {
	rng := rand.New(rand.NewSource(4))
	input := `contains("abc", "a")`
	got := ArgManipulate(input, rng)
	if got == input {
		t.Fatalf("ArgManipulate did not change expression: %s", DebugDescribeMutation(input, got))
	}
	if !exprgen.IsValid(got) {
		t.Fatalf("ArgManipulate produced invalid expression: %s", got)
	}
}

func TestNestWrapChangesExpression(t *testing.T) {
	rng := rand.New(rand.NewSource(5))
	input := "payload.age + 1"
	got := NestWrap(input, rng)
	if got == input {
		t.Fatalf("NestWrap did not change expression: %s", DebugDescribeMutation(input, got))
	}
	if !exprgen.IsValid(got) {
		t.Fatalf("NestWrap produced invalid expression: %s", got)
	}
}

func TestParenMutateChangesExpression(t *testing.T) {
	rng := rand.New(rand.NewSource(6))
	input := "1 + 2 * 3"
	got := ParenMutate(input, rng)
	if got == input {
		t.Fatalf("ParenMutate did not change expression: %s", DebugDescribeMutation(input, got))
	}
	if !exprgen.IsValid(got) {
		t.Fatalf("ParenMutate produced invalid expression: %s", got)
	}
}

func TestFunctionSwapChangesExpression(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	input := `sizeOf("abc")`
	got := FunctionSwap(input, rng)
	if got == input {
		t.Fatalf("FunctionSwap did not change expression: %s", DebugDescribeMutation(input, got))
	}
	if !exprgen.IsValid(got) {
		t.Fatalf("FunctionSwap produced invalid expression: %s", got)
	}
}

func TestMutateChangesExpression(t *testing.T) {
	rng := rand.New(rand.NewSource(8))
	input := `contains("abc", "a") && (1 + 2 > 0)`
	got := Mutate(input, rng)
	if got == input {
		t.Fatalf("Mutate did not change expression: %s", DebugDescribeMutation(input, got))
	}
}

func TestMutateNChangesExpression(t *testing.T) {
	rng := rand.New(rand.NewSource(9))
	input := `sizeOf([1, 2, 3])`
	got := MutateN(input, 3, rng)
	if got == input {
		t.Fatalf("MutateN did not change expression: %s", DebugDescribeMutation(input, got))
	}
	if !exprgen.IsValid(got) {
		t.Fatalf("MutateN produced invalid expression: %s", got)
	}
}
