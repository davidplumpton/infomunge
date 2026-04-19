package mutation

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"infomunge/internal/evaluator"
	"infomunge/internal/preprocessor"
	"infomunge/internal/runner"
	"infomunge/internal/testing/determinism"
	"infomunge/internal/testing/failures"
	"infomunge/internal/testing/metrics"
	"infomunge/internal/testing/testbudget"

	"pgregory.net/rapid"
)

func TestMutatedCorpusExpressions_NoPanics_AndDeterministic(t *testing.T) {
	entries := loadMutationCorpus(t)
	if len(entries) == 0 {
		t.Fatal("no corpus entries available for mutation testing")
	}

	maxMutations := testbudget.MutationMaxMutations()
	setRapidChecks(t, testbudget.MutationChecks())
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
		baseScript := buildScript(header, baseExpr)
		ctx := buildInputContext(entry.Inputs)
		detectedAt := time.Now().UTC()

		firstRes, firstErr, panicVal, panicStack := safeRun(script, ctx)
		if panicVal != nil {
			metrics.RecordPanic()
			recordMutationFailure("mutation_no_panic", baseExpr, mutatedExpr, seed, ctx, panicStack, detectedAt)
			t.Fatalf("panic while evaluating mutation (seed=%d): %v\nexpr=%q\nmutated=%q", seed, panicVal, baseExpr, mutatedExpr)
		}
		if isKnownNondeterministic(mutatedExpr) {
			return
		}

		secondRes, secondErr, panicVal, panicStack := safeRun(script, ctx)
		if panicVal != nil {
			metrics.RecordPanic()
			recordMutationFailure("mutation_no_panic", baseExpr, mutatedExpr, seed, ctx, panicStack, detectedAt)
			t.Fatalf("panic while re-evaluating mutation (seed=%d): %v\nexpr=%q\nmutated=%q", seed, panicVal, baseExpr, mutatedExpr)
		}

		if (firstErr == nil) != (secondErr == nil) {
			recordMutationFailure("mutation_determinism", baseExpr, mutatedExpr, seed, ctx, "", detectedAt)
			t.Fatalf("nondeterministic success/error for mutation seed=%d mutated=%q firstErr=%v secondErr=%v", seed, mutatedExpr, firstErr, secondErr)
		}

		if firstErr != nil {
			if !determinism.EqualErrors(firstErr, secondErr) {
				recordMutationFailure("mutation_determinism", baseExpr, mutatedExpr, seed, ctx, "", detectedAt)
				t.Fatalf("nondeterministic error text for mutation seed=%d mutated=%q firstErr=%v secondErr=%v", seed, mutatedExpr, firstErr, secondErr)
			}
			return
		}

		if !determinism.Equal(firstRes, secondRes) {
			recordMutationFailure("mutation_determinism", baseExpr, mutatedExpr, seed, ctx, "", detectedAt)
			t.Fatalf("nondeterministic result for mutation seed=%d mutated=%q first=%#v second=%#v", seed, mutatedExpr, firstRes, secondRes)
		}

		baseRes, baseErr, basePanic, baseStack := safeRun(baseScript, ctx)
		if basePanic != nil {
			metrics.RecordPanic()
			recordMutationFailure("mutation_no_panic", baseExpr, baseExpr, seed, ctx, baseStack, detectedAt)
			t.Fatalf("panic while evaluating base expression (seed=%d): %v\nexpr=%q", seed, basePanic, baseExpr)
		}
		if strings.TrimSpace(mutatedExpr) != strings.TrimSpace(baseExpr) && !isKnownNondeterministic(baseExpr) {
			killed := mutationChangesBehavior(baseRes, baseErr, firstRes, firstErr)
			metrics.RecordMutationOutcome(killed)
		}
	})
}

func TestMutationEngineProducesNonTrivialMutations(t *testing.T) {
	entries := loadMutationCorpus(t)
	if len(entries) == 0 {
		t.Fatal("no corpus entries available for mutation testing")
	}

	sample := entries
	sampleSize := testbudget.MutationSampleSize()
	if sampleSize > 0 && len(sample) > sampleSize {
		sample = sample[:sampleSize]
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
	sampleSize := testbudget.MutationSampleSize()
	if sampleSize == 0 || len(entries) <= sampleSize {
		return entries
	}

	rng := rand.New(rand.NewSource(42)) //nolint:gosec
	perm := rng.Perm(len(entries))
	out := make([]CorpusEntry, 0, sampleSize)
	for _, idx := range perm[:sampleSize] {
		out = append(out, entries[idx])
	}
	return out
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

func buildInputContext(inputs map[string]string) evaluator.Context {
	if len(inputs) == 0 {
		return nil
	}
	ctx := make(evaluator.Context, len(inputs))
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

func tryParseJSON(raw string) (evaluator.Value, error) {
	var out interface{}
	err := json.Unmarshal([]byte(raw), &out)
	return out, err
}

func safeRun(script string, ctx evaluator.Context) (result evaluator.Value, err error, panicValue interface{}, panicStack string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			panicValue = recovered
			panicStack = string(debug.Stack())
		}
	}()

	result, err = runner.RunString(script, ctx)
	return result, err, nil, ""
}

func recordMutationFailure(property, originalExpr, mutatedExpr string, seed int64, ctx evaluator.Context, panicStack string, detectedAt time.Time) {
	minimizedAt := time.Now().UTC()
	artifact := failures.Artifact{
		Property:            property,
		MinimizedExpression: mutatedExpr,
		OriginalExpression:  originalExpr,
		InputPayload:        ctx,
		Seed:                seed,
		PanicStack:          panicStack,
		DetectedAt:          detectedAt.Format(time.RFC3339),
		MinimizedAt:         minimizedAt.Format(time.RFC3339),
	}
	if _, _, err := failures.SaveArtifact(artifact); err != nil {
		return
	}
	_, _, _ = failures.WriteCandidateScenario(artifact)
}

func mutationChangesBehavior(baseRes interface{}, baseErr error, mutatedRes interface{}, mutatedErr error) bool {
	if (baseErr == nil) != (mutatedErr == nil) {
		return true
	}
	if baseErr != nil && mutatedErr != nil {
		return !determinism.EqualErrors(baseErr, mutatedErr)
	}
	return !determinism.Equal(baseRes, mutatedRes)
}

func TestMutationChangesBehavior_TreatsNaNAsDeterministic(t *testing.T) {
	if mutationChangesBehavior(math.NaN(), nil, math.NaN(), nil) {
		t.Fatal("expected NaN to be treated as deterministic in mutation comparator")
	}
}

func TestMain(m *testing.M) {
	code := m.Run()
	metrics.ReportAndPersist("mutation", metrics.Options{EnableCoverage: true})
	os.Exit(code)
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
