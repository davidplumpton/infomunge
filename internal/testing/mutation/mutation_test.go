package mutation

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"

	"infomunge/internal/preprocessor"
	"infomunge/internal/runner"
	"infomunge/internal/testing/failures"

	"pgregory.net/rapid"
)

const (
	defaultMutationSampleSize = 100
	defaultMutationChecks     = 300
)

func TestMutatedCorpusExpressions_NoPanics_AndDeterministic(t *testing.T) {
	entries := loadMutationCorpus(t)
	if len(entries) == 0 {
		t.Fatal("no corpus entries available for mutation testing")
	}

	maxMutations := 3
	checks := defaultMutationChecks
	if soakModeEnabled() {
		maxMutations = 10
		checks = 1000
	}

	setRapidChecks(t, checks)
	rapid.Check(t, func(t *rapid.T) {
		entry := entries[rapid.IntRange(0, len(entries)-1).Draw(t, "entry_idx")]
		header, baseExpr := extractHeaderAndExpr(entry.Script)
		if strings.TrimSpace(baseExpr) == "" {
			return
		}

		mutationCount := rapid.IntRange(1, maxMutations).Draw(t, "mutation_count")
		seed := rapid.Int64().Draw(t, "seed")
		rng := rand.New(rand.NewSource(seed)) //nolint:gosec
		mutatedExpr := MutateN(baseExpr, mutationCount, rng)
		script := buildScript(header, mutatedExpr)
		ctx := buildInputContext(entry.Inputs)

		firstRes, firstErr, panicVal, panicStack := safeRun(script, ctx)
		if panicVal != nil {
			recordMutationFailure("mutation_no_panic", baseExpr, mutatedExpr, seed, ctx, panicStack)
			t.Fatalf("panic while evaluating mutation (seed=%d): %v\nexpr=%q\nmutated=%q", seed, panicVal, baseExpr, mutatedExpr)
		}
		if isKnownNondeterministic(mutatedExpr) {
			return
		}

		secondRes, secondErr, panicVal, panicStack := safeRun(script, ctx)
		if panicVal != nil {
			recordMutationFailure("mutation_no_panic", baseExpr, mutatedExpr, seed, ctx, panicStack)
			t.Fatalf("panic while re-evaluating mutation (seed=%d): %v\nexpr=%q\nmutated=%q", seed, panicVal, baseExpr, mutatedExpr)
		}

		if (firstErr == nil) != (secondErr == nil) {
			recordMutationFailure("mutation_determinism", baseExpr, mutatedExpr, seed, ctx, "")
			t.Fatalf("nondeterministic success/error for mutation seed=%d mutated=%q firstErr=%v secondErr=%v", seed, mutatedExpr, firstErr, secondErr)
		}

		if firstErr != nil {
			if firstErr.Error() != secondErr.Error() {
				recordMutationFailure("mutation_determinism", baseExpr, mutatedExpr, seed, ctx, "")
				t.Fatalf("nondeterministic error text for mutation seed=%d mutated=%q firstErr=%v secondErr=%v", seed, mutatedExpr, firstErr, secondErr)
			}
			return
		}

		if !deterministicallyEqual(firstRes, secondRes) {
			recordMutationFailure("mutation_determinism", baseExpr, mutatedExpr, seed, ctx, "")
			t.Fatalf("nondeterministic result for mutation seed=%d mutated=%q first=%#v second=%#v", seed, mutatedExpr, firstRes, secondRes)
		}
	})
}

func TestMutationEngineProducesNonTrivialMutations(t *testing.T) {
	entries := loadMutationCorpus(t)
	if len(entries) == 0 {
		t.Fatal("no corpus entries available for mutation testing")
	}

	sample := entries
	if len(sample) > 50 && !soakModeEnabled() {
		sample = sample[:50]
	}

	changed := 0
	total := 0
	rng := rand.New(rand.NewSource(99)) //nolint:gosec
	for _, entry := range sample {
		_, expr := extractHeaderAndExpr(entry.Script)
		if strings.TrimSpace(expr) == "" {
			continue
		}
		total++
		mutated := MutateN(expr, 3, rng)
		if strings.TrimSpace(mutated) != strings.TrimSpace(expr) {
			changed++
		}
	}

	if total == 0 {
		t.Fatal("no valid expressions found in sampled corpus")
	}
	if changed == 0 {
		t.Fatalf("mutation engine produced no non-trivial mutations across %d expressions", total)
	}
}

func loadMutationCorpus(t *testing.T) []CorpusEntry {
	t.Helper()
	featuresDir := filepath.Join("..", "..", "..", "test", "features")
	entries, err := ExtractCorpus(featuresDir)
	if err != nil {
		t.Fatalf("ExtractCorpus failed: %v", err)
	}
	if soakModeEnabled() || len(entries) <= defaultMutationSampleSize {
		return entries
	}

	rng := rand.New(rand.NewSource(42)) //nolint:gosec
	perm := rng.Perm(len(entries))
	out := make([]CorpusEntry, 0, defaultMutationSampleSize)
	for _, idx := range perm[:defaultMutationSampleSize] {
		out = append(out, entries[idx])
	}
	return out
}

func soakModeEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("MUTATION_SOAK")))
	return v == "1" || v == "true" || v == "yes" || v == "full"
}

func extractHeaderAndExpr(script string) (string, string) {
	header, body, _ := preprocessor.ExtractHeaderAndBody(script)
	return strings.TrimSpace(header), strings.TrimSpace(body)
}

func buildScript(header, expr string) string {
	if strings.TrimSpace(header) == "" {
		return "%im 0.1\noutput application/json\n---\n" + expr
	}
	return header + "\n---\n" + expr
}

func buildInputContext(inputs map[string]string) map[string]interface{} {
	if len(inputs) == 0 {
		return nil
	}
	ctx := make(map[string]interface{}, len(inputs))
	for name, raw := range inputs {
		value := strings.TrimSpace(raw)
		if parsed, err := tryParseJSON(value); err == nil {
			ctx[name] = parsed
			continue
		}
		ctx[name] = value
	}
	return ctx
}

func tryParseJSON(raw string) (interface{}, error) {
	var out interface{}
	err := json.Unmarshal([]byte(raw), &out)
	return out, err
}

func safeRun(script string, ctx map[string]interface{}) (result interface{}, err error, panicValue interface{}, panicStack string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			panicValue = recovered
			panicStack = string(debug.Stack())
		}
	}()

	result, err = runner.RunString(script, ctx)
	return result, err, nil, ""
}

func deterministicallyEqual(a, b interface{}) bool {
	if reflect.DeepEqual(a, b) {
		return true
	}
	aj, aErr := json.Marshal(a)
	bj, bErr := json.Marshal(b)
	if aErr != nil || bErr != nil {
		return false
	}
	return string(aj) == string(bj)
}

func recordMutationFailure(property, originalExpr, mutatedExpr string, seed int64, ctx map[string]interface{}, panicStack string) {
	artifact := failures.Artifact{
		Property:            property,
		MinimizedExpression: mutatedExpr,
		OriginalExpression:  originalExpr,
		InputPayload:        ctx,
		Seed:                seed,
		PanicStack:          panicStack,
	}
	if _, _, err := failures.SaveArtifact(artifact); err != nil {
		return
	}
	_, _, _ = failures.WriteCandidateScenario(artifact)
}

func isKnownNondeterministic(expr string) bool {
	lower := strings.ToLower(expr)
	indicators := []string{
		"now(",
		"currentmilliseconds(",
		"random(",
		"uuid(",
		"randomuuid(",
	}
	for _, marker := range indicators {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func setRapidChecks(t *testing.T, checks int) {
	t.Helper()
	f := flag.Lookup("rapid.checks")
	if f == nil {
		t.Fatal("rapid.checks flag not found")
	}
	prev := f.Value.String()
	if err := f.Value.Set(fmt.Sprintf("%d", checks)); err != nil {
		t.Fatalf("set rapid.checks=%d: %v", checks, err)
	}
	t.Cleanup(func() {
		_ = f.Value.Set(prev)
	})
}
