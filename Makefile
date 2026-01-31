.PHONY: build test test-verbose test-failures test-feature test-unit playground clean

# Build the infomunge binary
build:
	go build -o infomunge ./cmd/infomunge

# Run all tests with real-time progress output
test:
	go test -v ./test -timeout 10m

# Run tests with pretty format for detailed step output
test-verbose:
	GODOG_FORMAT=pretty go test -v ./test -timeout 10m

# Run tests with minimal output (failed steps + periodic pass counts)
test-failures:
	GODOG_FORMAT=failures go test ./test -timeout 10m

# Run a specific feature file (usage: make test-feature FEATURE=simple_types)
test-feature:
	GODOG_PATHS=features/$(FEATURE).feature go test -v ./test -run TestFeatures

# Run unit tests only (exclude cucumber)
test-unit:
	go test -v ./test -run 'Test[^F]'

# Run the server playground locally at http://localhost:8080
playground:
	go run ./cmd/infomunge --server --listen :8080

# Clean build artifacts
clean:
	rm -f infomunge
	rm -f test/infomunge
	rm -f test/input.txt
