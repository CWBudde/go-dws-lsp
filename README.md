# go-dws-lsp

A Language Server Protocol (LSP) implementation for DWScript, written in Go. This LSP server provides rich language features for DWScript in any editor that supports LSP.

## Overview

DWScript is a powerful Object Pascal-based scripting language with strong static typing, object-oriented programming, and comprehensive built-in functions. This LSP server enables modern IDE features for DWScript development in editors like VSCode, Vim, Emacs, and others.

**Status**: ✅ Feature Complete - Phase 14/14 complete, testing and documentation in progress (Phase 15)

## Features

### ✅ Implemented Features

- **Document Synchronization**: Full and incremental sync with version tracking
- **Real-time Diagnostics**: Syntax and semantic error reporting with severity levels
- **Hover Information**: Type information, documentation, and symbol details
- **Go to Definition**: Navigate to symbol declarations across files
- **Find References**: Find all symbol usages workspace-wide
- **Document Symbols**: Hierarchical symbol tree with outline support
- **Workspace Symbols**: Global symbol search with fuzzy/substring/prefix matching
- **Code Completion**: Context-aware suggestions for keywords, symbols, members, and imports
- **Signature Help**: Function/method signatures with active parameter highlighting
- **Rename Support**: Workspace-wide symbol renaming with validation
- **Semantic Tokens**: Semantic syntax highlighting with 11 token types and 5 modifiers
- **Code Actions**: Quick fixes and refactoring actions:
  - "Declare variable" with type inference
  - "Declare function" with parameter type inference from call site
  - "Remove unused variable" and "Prefix with underscore"
  - "Organize units" (add missing, remove unused, sort)

### 📊 Test Coverage

- **200+ unit tests** covering all LSP features
- **All tests passing** with race condition detection
- **Performance optimized** with caching and incremental updates

## Installation

### Prerequisites

- Go 1.21 or later
- [go-dws](https://github.com/CWBudde/go-dws) v0.3.1+

### Building from Source

```bash
# Clone the repository
git clone https://github.com/CWBudde/go-dws-lsp.git
cd go-dws-lsp

# Build the LSP server
go build -o go-dws-lsp cmd/go-dws-lsp/main.go

# Optional: Install to $GOPATH/bin
go install ./cmd/go-dws-lsp
```

### Binary Installation

Pre-built binaries will be available with the v1.0.0 release.

## Usage

### Command-line Options

```bash
# Run via STDIO (default mode for editors)
go-dws-lsp

# Run as TCP server (useful for debugging)
go-dws-lsp -tcp -port 4389

# Set log level
go-dws-lsp -log-level debug -log-file lsp.log
```

**Options:**

- `-tcp`: Enable TCP server mode (default: false, uses STDIO)
- `-port <num>`: TCP port to listen on (default: 4389)
- `-log-level <level>`: Logging level: off, error, warn, info, debug (default: info)
- `-log-file <path>`: Log file path (default: stderr)

### Editor Integration

#### VSCode

Create or edit `.vscode/settings.json` in your workspace:

```json
{
  "dwscript.languageServer": {
    "command": "go-dws-lsp",
    "args": [],
    "filetypes": ["dws", "pas"]
  }
}
```

Or use the generic LSP client extension and configure:

```json
{
  "genericLanguageServer.servers": {
    "dwscript": {
      "command": "/path/to/go-dws-lsp",
      "args": [],
      "rootPatterns": ["*.dws"],
      "filetypes": ["dws"]
    }
  }
}
```

#### Vim/Neovim (with coc.nvim)

Add to `coc-settings.json`:

```json
{
  "languageserver": {
    "dwscript": {
      "command": "go-dws-lsp",
      "filetypes": ["dws", "pas"],
      "rootPatterns": ["*.dws"]
    }
  }
}
```

#### Emacs (with lsp-mode)

```elisp
(add-to-list 'lsp-language-id-configuration '(dwscript-mode . "dwscript"))

(lsp-register-client
 (make-lsp-client :new-connection (lsp-stdio-connection "go-dws-lsp")
                  :major-modes '(dwscript-mode)
                  :server-id 'go-dws-lsp))
```

## Supported LSP Capabilities

### Text Document Capabilities

- ✅ `textDocument/didOpen`
- ✅ `textDocument/didChange` (incremental sync)
- ✅ `textDocument/didClose`
- ✅ `textDocument/publishDiagnostics`
- ✅ `textDocument/hover`
- ✅ `textDocument/completion`
- ✅ `textDocument/signatureHelp`
- ✅ `textDocument/definition`
- ✅ `textDocument/references`
- ✅ `textDocument/documentSymbol`
- ✅ `textDocument/codeAction`
- ✅ `textDocument/rename`
- ✅ `textDocument/prepareRename`
- ✅ `textDocument/semanticTokens/full`
- ✅ `textDocument/semanticTokens/range`
- ✅ `textDocument/semanticTokens/full/delta`

### Workspace Capabilities

- ✅ `workspace/symbol`
- ✅ `workspace/didChangeConfiguration`
- ✅ `workspace/didChangeWatchedFiles`

### Server Capabilities

- ✅ `initialize` / `initialized`
- ✅ `shutdown` / `exit`

## Project Structure

```
go-dws-lsp/
├── cmd/go-dws-lsp/       # Main entry point
├── internal/
│   ├── analysis/         # AST analysis and semantic tokens
│   ├── builtins/         # DWScript built-in definitions
│   ├── document/         # Text document utilities
│   ├── lsp/              # LSP handlers (hover, completion, etc.)
│   ├── server/           # Server state management
│   └── workspace/        # Workspace indexing and symbols
├── test/                 # Integration tests
└── PLAN.md              # Implementation roadmap
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with race detection
go test -race ./...

# Run specific package tests
go test ./internal/lsp -v

# Run with coverage
go test -cover ./...
```

### Debugging

Run in TCP mode for easier debugging:

```bash
go-dws-lsp -tcp -port 4389 -log-level debug -log-file debug.log
```

Then connect your editor's LSP client to `localhost:4389`.

### Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`go test ./...`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

See [PLAN.md](PLAN.md) for the implementation roadmap and current status.

## Performance

The LSP server is optimized for interactive use:

- **Document open**: <100ms
- **Completion**: <50ms
- **Hover**: <50ms
- **Find references**: <500ms (1000 files)
- **Workspace symbols**: <200ms

Caching and incremental updates ensure responsive editing even in large workspaces.

## Dependencies

- [github.com/CWBudde/go-dws](https://github.com/CWBudde/go-dws) - DWScript implementation in Go
- [github.com/tliron/glsp](https://github.com/tliron/glsp) - Go LSP library

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Related Projects

- [DWScript](https://github.com/EricGrange/DWScript) - The original Delphi implementation
- [go-dws](https://github.com/CWBudde/go-dws) - Go implementation of DWScript

## Roadmap

**Current Phase**: Phase 15 - Testing, Quality, and Finalization

- ✅ Phase 0-14: All core LSP features complete
- 🚧 Phase 15: Integration testing, performance validation, documentation
- 🎯 v1.0.0 Release: Planned after Phase 15 completion

See [PLAN.md](PLAN.md) for detailed progress and task tracking.

## Contact

- **Issues**: [GitHub Issues](https://github.com/CWBudde/go-dws-lsp/issues)
- **Discussions**: [GitHub Discussions](https://github.com/CWBudde/go-dws-lsp/discussions)

---

**Maintained by**: [Christian Budde](https://github.com/CWBudde)
