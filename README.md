# go-dws-lsp

A Language Server Protocol (LSP) implementation for DWScript, written in Delphi. This LSP server provides language features such as syntax highlighting, diagnostics, code completion, and more for DWScript code in editors that support LSP.

## Overview

DWScript is a powerful Object Pascal-based scripting language featuring strong static typing, object-oriented programming, design by contract, and comprehensive built-in functions. This LSP server enables rich editing experiences for DWScript in modern code editors.

This implementation is built in Delphi and aims to provide full LSP compliance for DWScript development.

## Features

- **Text Synchronization**: Handles document open, change, and close events
- **Diagnostics**: Provides syntax and semantic error reporting
- **Language Features**: Supports hover information, completion, signature help, etc.
- **Workspace Management**: Handles workspace folders and configuration
- **Window Management**: Manages show message and log message requests
- **Telemetry**: Basic telemetry support for usage analytics

## Installation

### Prerequisites

- Delphi (RAD Studio or compatible IDE)
- DWScript runtime or go-dws (the Go implementation of DWScript)

### Building

1. Clone the repository:

   ```bash
   git clone https://github.com/CWBudde/go-dws-lsp.git
   cd go-dws-lsp
   ```

2. Open the project in Delphi:
   - Load `dwsc.dproj` in RAD Studio

3. Build the project:
   - Compile the LSP server executable

### Running

The LSP server communicates via JSON-RPC over stdin/stdout. It's designed to be launched by an LSP-compatible editor.

## Usage

### Editor Integration

To use this LSP server with your editor, configure it to launch the compiled `dwsc.exe` as the language server for DWScript files (`.dws` extension).

#### Example: VS Code

Add the following to your VS Code settings:

```json
{
  "languages": [
    {
      "id": "dwscript",
      "extensions": [".dws"],
      "aliases": ["DWScript"]
    }
  ],
  "languageServers": {
    "dwscript": {
      "command": "path/to/dwsc.exe",
      "args": []
    }
  }
}
```

Note: VS Code's LSP support may require additional configuration or extensions. This is a basic example.

### Supported LSP Capabilities

- `textDocumentSync`: Incremental text synchronization
- `diagnostics`: Publishing diagnostics
- `hover`: Hover information
- `completion`: Code completion
- `signatureHelp`: Function signature help
- `definition`: Go to definition
- `references`: Find references
- `documentSymbol`: Document symbols
- `workspaceSymbol`: Workspace symbols

## Project Structure

- `dwsc.dpr`: Main project file
- `dwsc.LanguageServer.pas`: Core LSP server implementation
- `dwsc.Classes.*.pas`: LSP protocol classes (BaseProtocol, Capabilities, Client, etc.)
- `dwsc.RequestManager.pas`: Handles LSP requests and responses
- `dwsc.IO.*.pas`: Input/Output handling (Pipe, Socket)
- `dwsc.Utils.pas`: Utility functions
- `reference/`: Reference DWScript source files

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

### Development Setup

- Ensure you have Delphi installed
- Familiarize yourself with the LSP specification: <https://microsoft.github.io/language-server-protocol/>
- Test changes with a compatible editor

## License

To be determined. This project is related to DWScript and will respect the original DWScript license.

## Related Projects

- [DWScript](https://github.com/EricGrange/DWScript) - The original Delphi implementation
- [go-dws](https://github.com/CWBudde/go-dws) - Go implementation of DWScript

## Contact

- GitHub Issues: [Report bugs or request features](https://github.com/CWBudde/go-dws-lsp/issues)

---

**Status**: 🚧 Work in Progress - This LSP implementation is under active development and may not support all DWScript features yet.
