# Justfile for go-dws-lsp
# Provides automated testing and development workflows

# Default recipe - show available commands
default:
    @just --list

# Install development dependencies (formatters and linters)
setup-deps:
    @echo "Installing development dependencies..."
    @echo "Installing gofumpt (Go formatter)..."
    go install mvdan.cc/gofumpt@latest
    @echo "Installing gci (Go import formatter)..."
    go install github.com/daixiang0/gci@latest
    @echo "Installing shfmt (Shell formatter)..."
    go install mvdan.cc/sh/v3/cmd/shfmt@latest
    @echo "Installing golangci-lint (Go linter)..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0
    @echo "Development dependencies installation complete!"
    @echo "Note: Ensure $(go env GOPATH)/bin is in your PATH for Go-based tools"
    @echo ""
    @echo "Optional dependencies (install manually if needed):"
    @echo "  - treefmt: https://github.com/numtide/treefmt/releases"
    @echo "  - prettier: npm install -g prettier"
    @echo "  - shellcheck: https://github.com/koalaman/shellcheck"

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

# Format code using treefmt (falls back to gofumpt+gci if treefmt not available)
fmt:
    #!/usr/bin/env sh
    if command -v treefmt >/dev/null 2>&1; then \
        treefmt --allow-missing-formatter; \
    else \
        echo "treefmt not found, using gofumpt + gci fallback..."; \
        gofumpt -w .; \
        gci write --skip-generated -s standard -s default .; \
    fi

# Run linter
lint:
    golangci-lint run --config ./.golangci.toml --timeout 2m

# Run linter (with fix)
lint-fix:
    golangci-lint run --config ./.golangci.toml --timeout 2m --fix

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

# Check if code is formatted (for CI)
check-fmt:
    #!/usr/bin/env sh
    if command -v treefmt >/dev/null 2>&1; then \
        treefmt --allow-missing-formatter --fail-on-change; \
    else \
        echo "treefmt not found, using gofumpt check fallback..."; \
        test -z "$(gofumpt -l .)" || (echo "Files need formatting:" && gofumpt -l . && exit 1); \
    fi

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
