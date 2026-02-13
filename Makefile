.PHONY: build test test-verbose test-failures test-feature test-unit test-intensive test-fuzz test-soak playground clean

# Build the infomunge binary
build:
	go build -o infomunge ./cmd/infomunge

# Run all tests (failed steps + periodic pass counts)
test:
	GODOG_FORMAT=failures go test -v ./test -timeout 10m

# Run tests with full progress output (use to find hanging/timeout tests)
test-verbose:
	GODOG_FORMAT=progress go test -v ./test -timeout 10m

# Run tests with minimal output (failed steps + periodic pass counts)
test-failures:
	GODOG_FORMAT=failures go test -v ./test -timeout 10m

# Run a specific feature file (usage: make test-feature FEATURE=simple_types)
test-feature:
	GODOG_PATHS=features/$(FEATURE).feature go test -v ./test -run TestFeatures

# Run unit tests only (exclude cucumber)
test-unit:
	go test ./test -run 'Test[^F]'

# Run intensive property/mutation tests (CI budget, ~30s)
test-intensive:
	go test ./internal/testing/properties/ ./internal/testing/mutation/ ./internal/testing/differential/ -timeout 2m

# Run fuzz tests (CI budget, ~30s each target)
test-fuzz:
	go test ./internal/testing/properties/ -fuzz=FuzzExprEval -fuzztime=30s -timeout 2m
	go test ./internal/testing/properties/ -fuzz=FuzzExprPreprocess -fuzztime=30s -timeout 2m
	go test ./internal/testing/properties/ -fuzz=FuzzExprDeep -fuzztime=30s -timeout 2m
	go test ./internal/preprocessor/ -fuzz=FuzzPrepareForParsing -fuzztime=30s -timeout 2m
	go test ./internal/preprocessor/ -fuzz=FuzzExtractHeaderAndBody -fuzztime=30s -timeout 2m

# Run soak tests locally (full budget, INTENSIVE_TEST_SOAK=1)
test-soak:
	INTENSIVE_TEST_SOAK=1 go test -v ./internal/testing/properties/ ./internal/testing/mutation/ ./internal/testing/differential/ -timeout 60m
	go test ./internal/testing/properties/ -fuzz=FuzzExprEval -fuzztime=30m -timeout 45m
	go test ./internal/testing/properties/ -fuzz=FuzzExprPreprocess -fuzztime=30m -timeout 45m
	go test ./internal/testing/properties/ -fuzz=FuzzExprDeep -fuzztime=30m -timeout 45m
	go test ./internal/preprocessor/ -fuzz=FuzzPrepareForParsing -fuzztime=30m -timeout 45m
	go test ./internal/preprocessor/ -fuzz=FuzzExtractHeaderAndBody -fuzztime=30m -timeout 45m

# Run the server playground locally at http://localhost:8080
playground:
	go run ./cmd/infomunge --server --listen :8080

# Clean build artifacts
clean:
	rm -f infomunge
	rm -f test/infomunge
	rm -f test/input.txt
