.PHONY: build test test-verbose test-failures test-feature test-unit test-intensive test-fuzz test-differential-soak test-soak playground playground-wasm playground-wasm-serve clean

# Build the infomunge binary
build:
	go build -o infomunge ./cmd/infomunge

# Run all tests (failed steps + periodic pass counts)
test test-failures:
	GODOG_FORMAT=failures go test -v ./test -timeout 5m

# Run tests with full progress output (use to find hanging/timeout tests)
test-verbose:
	GODOG_FORMAT=progress go test -v ./test -timeout 5m

# Run a specific feature file (usage: make test-feature FEATURE=simple_types)
test-feature:
	GODOG_PATHS=features/$(FEATURE).feature go test -v ./test -run TestFeatures -timeout 5m

# Run unit tests only (exclude cucumber)
test-unit:
	INFOMUNGE_SKIP_GODOG=1 go test ./... -timeout 5m

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

# Run the larger external DataWeave comparison budget.
test-differential-soak:
	INTENSIVE_TEST_SOAK=1 go test -v ./internal/testing/differential/ -timeout 15m

# Run soak tests locally (full budget, INTENSIVE_TEST_SOAK=1)
test-soak:
	INTENSIVE_TEST_SOAK=1 go test -v ./internal/testing/properties/ ./internal/testing/mutation/ ./internal/testing/differential/ -timeout 60m
	go test ./internal/testing/properties/ -fuzz=FuzzExprEval -fuzztime=30m -timeout 45m
	go test ./internal/testing/properties/ -fuzz=FuzzExprPreprocess -fuzztime=30m -timeout 45m
	go test ./internal/testing/properties/ -fuzz=FuzzExprDeep -fuzztime=30m -timeout 45m
	go test ./internal/preprocessor/ -fuzz=FuzzPrepareForParsing -fuzztime=30m -timeout 45m
	go test ./internal/preprocessor/ -fuzz=FuzzExtractHeaderAndBody -fuzztime=30m -timeout 45m

# Run the server playground locally at http://127.0.0.1:8080
playground:
	go run ./cmd/infomunge --server --listen 127.0.0.1:8080

# Build standalone WebAssembly playground artifacts.
playground-wasm:
	GOOS=js GOARCH=wasm go build -o docs/playground/infomunge.wasm ./cmd/infomunge-wasm
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" docs/playground/wasm_exec.js

# Build and serve the standalone WebAssembly playground at http://127.0.0.1:8081.
playground-wasm-serve: playground-wasm
	go run ./cmd/infomunge-playground --listen 127.0.0.1:8081

# Clean build artifacts
clean:
	rm -f infomunge
	rm -f test/infomunge
	rm -f test/input.txt
