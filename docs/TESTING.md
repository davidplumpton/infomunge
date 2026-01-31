# Testing

InfoMunge uses Go unit tests plus Godog feature tests.

## Unit Tests

```bash
go test ./...
```

Targeted packages:
- `internal/evaluator/*_test.go`
- `internal/preprocessor/*_test.go`
- `pkg/formats/*_test.go`

## Feature Tests (Godog)

```bash
go test -v ./test
```

Feature files live in `test/features/*.feature`.
