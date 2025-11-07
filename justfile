# Justfile for go-dws-lsp
# Provides automated testing and development workflows

# Default recipe - show available commands
default:
    @just --list

# Install all required dependencies
install-deps:
    @echo "Installing Go module dependencies..."
    go mod download
    @echo "Installing testing tools..."
    go install golang.org/x/tools/gopls@latest
    @echo "Installing test coverage tools..."
    go install github.com/axw/gocov/gocov@latest
    go install github.com/AlekSi/gocov-xml@latest
    @echo "All dependencies installed successfully!"

# Build the LSP server
build:
    @echo "Building go-dws-lsp..."
    go build -o bin/go-dws-lsp ./cmd/go-dws-lsp
    @echo "Build complete: bin/go-dws-lsp"

# Run all unit tests
test:
    @echo "Running unit tests..."
    go test -v ./...

# Run unit tests with coverage
test-coverage:
    @echo "Running tests with coverage..."
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Run unit tests with race detection
test-race:
    @echo "Running tests with race detection..."
    go test -v -race ./...

# Run integration tests
test-integration:
    @echo "Running integration tests..."
    go test -v -tags=integration ./test/integration/...

# Run all tests (unit + integration)
test-all: test test-integration

# Run tests in watch mode (requires entr)
test-watch:
    @echo "Watching for changes and running tests..."
    find . -name "*.go" | entr -c just test

# Run linters and static analysis
lint:
    @echo "Running go vet..."
    go vet ./...
    @echo "Running go fmt check..."
    @test -z "$(gofmt -l .)" || (echo "Files need formatting:" && gofmt -l . && exit 1)
    @echo "All checks passed!"

# Format code
fmt:
    @echo "Formatting code..."
    go fmt ./...
    @echo "Code formatted!"

# Run benchmarks
bench:
    @echo "Running benchmarks..."
    go test -bench=. -benchmem ./...

# Clean build artifacts and test cache
clean:
    @echo "Cleaning build artifacts..."
    rm -rf bin/
    rm -f coverage.out coverage.html
    go clean -testcache
    @echo "Clean complete!"

# Check for security vulnerabilities
security:
    @echo "Checking for security vulnerabilities..."
    go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth

# Run all checks before commit (format, lint, test)
pre-commit: fmt lint test
    @echo "All pre-commit checks passed!"

# Development: build and run the LSP server
dev: build
    @echo "Starting LSP server..."
    ./bin/go-dws-lsp

# CI: Run all checks suitable for continuous integration
ci: lint test-race test-coverage
    @echo "CI checks complete!"

# Show test coverage statistics
coverage-stats:
    @echo "Generating coverage statistics..."
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out | tail -1
