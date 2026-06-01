# Testing

InfoMunge uses Go unit tests plus Godog feature tests.

## Unit Tests

```bash
go test ./...
```

The `tmp/` scratch area is a nested Go module, so ad hoc helpers there do not participate in repo-root package discovery. Run scratch helpers from inside `tmp/` when needed.

Targeted packages:
- `internal/evaluator/*_test.go`
- `internal/preprocessor/*_test.go`
- `pkg/formats/*_test.go`

## Feature Tests (Godog)

```bash
go test -v ./test -timeout 5m
```

Feature files live in `test/features/*.feature`.

## Cucumber Coverage

Use the Godog suite with a shared coverage profile over runtime-critical packages:

```bash
timeout 5m go test -v ./test -run TestFeatures -count=1 -timeout 5m \
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

Baseline from the full Godog suite on **2026-02-14**:
- `internal/evaluator`: `35.5%` (`1734/4882`)
- `internal/preprocessor`: `74.3%` (`1798/2419`)
- `internal/runner`: `36.7%` (`327/892`)
- `pkg/formats`: `62.8%` (`1635/2605`)
