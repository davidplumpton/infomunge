# Testing

InfoMunge uses Go unit tests plus Godog feature tests.

## Unit Tests

```bash
INFOMUNGE_SKIP_GODOG=1 go test ./... -timeout 5m
```

This skips Godog because feature tests have their own command. The `tmp/` scratch area is a nested Go module, so ad hoc helpers there do not participate in repo-root package discovery. Run scratch helpers from inside `tmp/` when needed.

The bounded repo-wide command skips the mutation corpus soak by default. Run the mutation soak explicitly when needed:

```bash
INTENSIVE_TEST_SOAK=1 go test -v ./internal/testing/mutation -run TestMutatedCorpusExpressions_NoPanics_AndDeterministic -timeout 30m
```

When `dw` is available, the bounded suite generates 50 differential cases and
requires at least 20% to reach structural comparison after evaluator errors are
classified. Each DataWeave evaluation starts an external CLI process, so this
budget is deliberately smaller than the in-process property-test budgets. Run
the larger 500-case budget explicitly:

```bash
make test-differential-soak
```

Targeted packages:

- `internal/evaluator/*_test.go`
- `internal/preprocessor/*_test.go`
- `pkg/formats/*_test.go`

## Intensive Expression Testing

The intensive-testing system exercises source expressions through the full
preprocess, parse, evaluate, and format pipeline. Its implementation is under
`internal/testing`:

- `exprgen` builds bounded, source-level expressions and payload contexts.
- `properties` checks no-panic, determinism, algebraic, structural, type, and
  JSON round-trip properties.
- `mutation` extracts scripts from the Godog corpus and applies reproducible
  source mutations.
- `differential` compares the compatible expression subset with the DataWeave
  CLI using normalized, path-aware structural comparison. It skips when `dw`
  is not on `PATH`.
- `failures` deduplicates minimized findings and produces reviewable candidate
  Godog scenarios.
- `metrics` reports failure, shrinking, mutation, promotion, panic, and optional
  coverage measurements.

Run the bounded property, mutation, and differential packages:

```bash
make test-intensive
```

Run each Go fuzz target for its short local budget:

```bash
make test-fuzz
```

Run the extended property/mutation/differential budgets and fuzz targets:

```bash
make test-soak
```

To run only the extended differential budget:

```bash
make test-differential-soak
```

The soak target is intentionally a long-running local workflow. The regular
repository suite does not enable it.

### Failure Review

Findings are stored beneath `tmp/intensive-testing/` regardless of the package
working directory:

- `failures/` contains sequential JSON artifacts deduplicated by property and
  minimized-expression fingerprint.
- `candidates/` contains generated `.feature` scenario drafts for human review.
- `report.json` contains the latest metrics and history.

Candidate scenarios are starting points, not automatic regressions. Confirm the
finding, replace the placeholder assertion with the exact expected result or
error, move the scenario into the appropriate file under `test/features`, and
run that focused feature before committing it.

## Feature Tests (Godog)

```bash
go test -v ./test -timeout 5m
```

Feature files live in `test/features/*.feature`.

## Cucumber Coverage

Use the Godog suite with a shared coverage profile over runtime-critical packages:

```bash
go test -v ./test -run TestFeatures -count=1 -timeout 5m \
  -coverprofile=tmp/cucumber.cover \
  -coverpkg=./internal/runner,./internal/preprocessor,./internal/evaluator,./pkg/formats
```

Inspect overall coverage:

```bash
go tool cover -func=tmp/cucumber.cover | tail -n 1
```

Inspect package-level statement coverage for the tracked runtime packages:

```bash
awk 'NR==1{next} { split($1,a,":"); file=a[1]; n=$2; c=$3; pkg="";
if (file ~ /infomunge\/internal\/evaluator\//) pkg="internal/evaluator";
else if (file ~ /infomunge\/internal\/preprocessor\//) pkg="internal/preprocessor";
else if (file ~ /infomunge\/internal\/runner\//) pkg="internal/runner";
else if (file ~ /infomunge\/pkg\/formats\//) pkg="pkg/formats";
if (pkg!="") { total[pkg]+=n; if (c>0) covered[pkg]+=n; } }
END { for (p in total) printf "%s %.1f%% (%d/%d)\n", p, (covered[p]/total[p])*100, covered[p], total[p]; }' tmp/cucumber.cover | sort
```

Baseline from the full Godog suite on **2026-07-24**:
- `internal/evaluator`: `58.4%` (`2923/5005`)
- `internal/preprocessor`: `78.5%` (`2082/2653`)
- `internal/runner`: `73.3%` (`519/708`)
- `pkg/formats`: `70.0%` (`1809/2585`)
