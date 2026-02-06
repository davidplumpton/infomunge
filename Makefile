.PHONY: build test test-verbose test-failures test-feature test-unit playground clean

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

# Run the server playground locally at http://localhost:8080
playground:
	go run ./cmd/infomunge --server --listen :8080

# Clean build artifacts
clean:
	rm -f infomunge
	rm -f test/infomunge
	rm -f test/input.txt
