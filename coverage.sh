#!/bin/bash
# Coverage script for infomunge

set -e

OUTPUT_DIR="coverage"
COVERAGE_FILE="${OUTPUT_DIR}/coverage.out"
HTML_FILE="${OUTPUT_DIR}/coverage.html"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Run tests with coverage
echo "Running tests with coverage..."
go test -v -coverprofile="$COVERAGE_FILE" ./...

# Generate HTML report
echo "Generating HTML coverage report..."
go tool cover -html="$COVERAGE_FILE" -o "$HTML_FILE"

# Display coverage summary
echo ""
echo "Coverage Summary:"
echo "================"
go tool cover -func="$COVERAGE_FILE" | grep total

echo ""
echo "HTML report generated: $HTML_FILE"
echo "To view: open $HTML_FILE"
