.PHONY: test test-unit test-integration test-coverage build

# Run all tests
test: test-unit test-integration

# Run unit tests
test-unit:
    @echo "Running unit tests..."
    @go test -v ./tests/unit/...

# Run integration tests
test-integration:
    @echo "Running integration tests..."
    @go test -v ./tests/integration/...

# Run tests with coverage
test-coverage:
    @echo "Running tests with coverage..."
    @go test -v -coverprofile=coverage.out ./...
    @go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated at coverage.html"

# Build the API Gateway
build:
    @echo "Building API Gateway..."
    @go build -o bin/api-gateway ./cmd/main.go

# Run the API Gateway
run:
    @echo "Running API Gateway..."
    @go run ./cmd/main.go