# Integration Tests

This directory contains integration tests for the go-dws-lsp Language Server.

## Overview

Integration tests verify end-to-end functionality of the LSP server by simulating real client-server interactions. These tests cover:

- **Hover Support**: Testing textDocument/hover requests
- **Document Lifecycle**: Testing didOpen, didChange, didClose notifications
- **Initialization**: Testing initialize/initialized workflow
- **Diagnostics**: Testing error reporting and validation
- **Concurrent Operations**: Testing multiple documents and operations

## Running Integration Tests

### Prerequisites

Install all required dependencies:

```bash
just install-deps
```

Or manually:

```bash
go mod download
```

### Run Integration Tests

Using justfile (recommended):

```bash
just test-integration
```

Using go test directly:

```bash
go test -v -tags=integration ./test/integration/...
```

### Run All Tests (Unit + Integration)

```bash
just test-all
```

## Test Structure

### Test Files

- `hover_integration_test.go` - Tests for hover functionality
- `lsp_integration_test.go` - Tests for general LSP protocol operations

### Test Categories

**Hover Tests:**

- Variable declaration hover
- Function declaration hover
- Class declaration hover
- Invalid position handling
- Document updates
- Multiple documents

**LSP Protocol Tests:**

- Initialize workflow
- Document lifecycle (open/change/close)
- Diagnostics on didOpen
- Incremental document changes
- Concurrent document operations
- Shutdown workflow

## Writing New Integration Tests

1. Add the build tag at the top of your test file:

   ```go
   //go:build integration
   // +build integration
   ```

2. Use the `setupTestServer()` helper to create a test server instance:

   ```go
   srv := setupTestServer()
   ```

3. Simulate LSP client operations using the protocol types:

   ```go
   openParams := &protocol.DidOpenTextDocumentParams{
       TextDocument: protocol.TextDocumentItem{
           URI:        "file:///test/example.dws",
           LanguageID: "dwscript",
           Version:    1,
           Text:       "var x: Integer;",
       },
   }
   ```

4. Verify expected behavior using standard Go testing assertions

## Test Data

Integration tests use inline code snippets for simplicity. For more complex scenarios, consider adding test fixture files in a `testdata` directory.

## Continuous Integration

Integration tests are run as part of the CI pipeline defined in `.github/workflows/ci.yml`.

## Troubleshooting

**Tests failing due to missing dependencies:**

- Run `just install-deps` to install all required dependencies

**Tests failing with "undefined: lsp.SetServer":**

- Ensure you're running from the project root directory
- Verify all imports are correct

**Tests timing out:**

- Check if the LSP server is hanging on a specific operation
- Use `go test -v -timeout 30s` to increase timeout

## Performance Considerations

Integration tests are slower than unit tests because they test the entire system. Consider:

- Running integration tests separately from unit tests
- Using parallel test execution where appropriate
- Mocking external dependencies when possible

## Contributing

When adding new LSP features, please add corresponding integration tests to verify end-to-end functionality.
