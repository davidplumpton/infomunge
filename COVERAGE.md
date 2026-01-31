# Test Coverage Guide

## Overview

This project maintains test coverage metrics to ensure code quality. Current coverage by package:

- `pkg/formats`: 83.5% (excellent)
- `internal/preprocessor`: 33.4% (needs improvement)
- `internal/evaluator`: 28.1% (needs improvement)
- `internal/runner`: 0.0% (not tested)
- `cmd/infomunge`: 0.0% (integration tested)

## Running Coverage Reports

### Quick Coverage Summary

```bash
go test -cover ./...
```

### Detailed HTML Report

```bash
./coverage.sh
# or
make coverage
```

This generates `coverage/coverage.html` which you can open in a browser for detailed line-by-line coverage visualization.

### View Function Coverage

```bash
go test -coverprofile=coverage/coverage.out ./...
go tool cover -func=coverage/coverage.out
```

## Coverage Goals

We aim to increase coverage in the following areas:

### High Priority
- `internal/evaluator`: Currently 28.1%, target 70%
  - Add tests for error handling paths
  - Add tests for all built-in functions
  - Add tests for edge cases in arithmetic operations
  
- `internal/preprocessor`: Currently 33.4%, target 70%
  - Add tests for all transformation functions
  - Add tests for escape sequence handling
  - Add tests for edge cases

### Medium Priority
- `internal/runner`: Currently 0%, target 50%
  - Add integration tests for the main runner
  - Test command-line argument parsing

### Lower Priority
- `cmd/infomunge`: Currently 0%, integration tested via BDD scenarios
  - Unit tests less critical due to extensive BDD coverage

## Writing Tests

### Test Structure

Go tests follow a standard pattern:

```go
func TestFeatureName(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name:  "descriptive test case",
			input: someInput,
			want:  expectedOutput,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			// Act
			result := functionUnderTest(tt.input)
			// Assert
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.want {
				t.Errorf("got %v, want %v", result, tt.want)
			}
		})
	}
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestFeatureName ./package

# Run specific test case
go test -run TestFeatureName/test_case_name ./package
```

## CI/CD Integration

Coverage reports are generated automatically on:
- Every push to `main`
- Every pull request

Reports are uploaded as artifacts and PR comments include coverage summary.

## Tools

The project uses:
- Go's built-in `testing` package
- [Godog](https://github.com/cucumber/godog) for BDD scenarios
- [Custom assertion library](test/assertions.go) for test helpers
