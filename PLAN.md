# go-dws-lsp Implementation Plan

This document provides a detailed, phase-by-phase implementation plan for the go-dws Language Server Protocol (LSP) implementation. The plan breaks down the project into 13 distinct phases, each focusing on a specific feature or set of related features.

## Overview

The implementation is organized into the following phases:

- **Phase 0**: Foundation - LSP Scaffolding and Setup
- **Phase 1**: Document Synchronization
- **Phase 2**: Diagnostics (Syntax and Semantic Analysis)
- **Phase 3**: Hover Support
- **Phase 4**: Go-to Definition
- **Phase 5**: Find References
- **Phase 6**: Document Symbols
- **Phase 7**: Workspace Symbols
- **Phase 8**: Code Completion
- **Phase 9**: Signature Help
- **Phase 10**: Rename Support
- **Phase 11**: Semantic Tokens
- **Phase 12**: Code Actions
- **Phase 13**: Testing, Quality, and Finalization

**Total Tasks**: 254

---

## Phase 0: Foundation - LSP Scaffolding and Setup

**Goal**: Establish the basic language server framework: communication, lifecycle handling, and project structure.

### Tasks (21)

- [x] **Initialize Go module (go-dws-lsp) and repository structure**
  - [x] Run `go mod init github.com/CWBudde/go-dws-lsp`
  - [x] Create directory structure:
    - [x] `cmd/go-dws-lsp/` for main executable
    - [x] `internal/lsp/` for LSP protocol handlers
    - [x] `internal/server/` for server state management
    - [x] `internal/document/` for document management
    - [x] `internal/analysis/` for DWScript integration
    - [ ] `pkg/protocol/` for LSP types (if extending GLSP)
  - [x] Create `.gitignore` with Go-specific entries
  - [x] Initialize git repository with initial commit

- [x] **Create main.go to launch the server**
  - [x] Create `cmd/go-dws-lsp/main.go`
  - [x] Implement command-line flag parsing structure
  - [x] Add version constant/variable
  - [x] Implement basic main() function skeleton

- [x] **Set up separate packages for LSP handlers and DWScript integration**
  - [x] Create `internal/lsp/handlers.go` for handler registration
  - [x] Create `internal/analysis/analyzer.go` for DWScript wrapper
  - [x] Define clear package boundaries and interfaces
  - [x] Document package responsibilities in package comments

- [x] **Add GLSP dependency (github.com/tliron/glsp) for LSP protocol handling**
  - [x] Run `go get github.com/tliron/glsp`
  - [x] Run `go get github.com/tliron/commonlog`
  - [x] Verify dependencies compile correctly
  - [x] Study GLSP examples to understand usage patterns

- [x] **Implement Initialize request handler with server capabilities**
  - [x] Create `internal/lsp/initialize.go`
  - [x] Define handler function: `func Initialize(context *glsp.Context, params *protocol.InitializeParams) (interface{}, error)`
  - [x] Extract workspace folders from params
  - [x] Store initialization info in server state
  - [x] Build and return InitializeResult

- [x] **Advertise text document sync, diagnostics, hover, completion capabilities in Initialize**
  - [x] Set `TextDocumentSyncKind` to `Incremental`
  - [x] Enable `TextDocumentSyncOptions` with openClose, change, willSave flags
  - [x] Mark `HoverProvider: true`
  - [x] Mark `DefinitionProvider: true`
  - [x] Mark `ReferencesProvider: true`
  - [x] Mark `DocumentSymbolProvider: true`
  - [x] Set `CompletionProvider` with trigger characters (`.`, etc.)
  - [x] Mark `SignatureHelpProvider` with trigger characters (`(`, `,`)

- [x] **Provide ServerInfo name and version in Initialize response**
  - [x] Set `serverInfo.name = "go-dws-lsp"`
  - [x] Set `serverInfo.version` from build constant
  - [x] Include in InitializeResult

- [x] **Implement Shutdown request handler**
  - [x] Create handler: `func Shutdown(context *glsp.Context) error`
  - [ ] Set shutdown flag in server state
  - [ ] Clean up resources (close files, flush caches)
  - [x] Return nil on success

- [x] **Implement STDIN/STDOUT transport layer using server.RunStdio()**
  - [x] Create GLSP server instance
  - [x] Register all handlers with server
  - [x] Call `server.RunStdio()` for stdio mode
  - [x] Handle errors and exit codes

- [x] **Add TCP transport option (-tcp flag) for debugging**
  - [x] Add `-tcp` flag (default: false)
  - [x] Add `-port` flag (default: 8765)
  - [x] Implement TCP server mode using `server.RunTCP()`
  - [x] Log connection info when TCP mode active

- [x] **Parse command-line flags for transport mode selection**
  - [x] Use `flag` package for CLI parsing
  - [x] Add `-tcp` boolean flag
  - [x] Add `-port` int flag
  - [x] Add `-log-level` string flag (debug, info, warn, error)
  - [x] Parse flags before server initialization

- [x] **Set up logging facility with adjustable verbosity (LSPTrace flag)**
  - [x] Initialize commonlog logger
  - [x] Create log level mapping (debug, info, warn, error)
  - [x] Configure log output destination (stderr or file)
  - [x] Add structured logging helpers

- [x] **Add command-line flags or environment variables for log level control**
  - [x] Implement `-log-level` flag handling
  - [ ] Support `LSP_LOG_LEVEL` environment variable
  - [x] Support `-log-file` flag for output redirection
  - [x] Default to error-level logging in production

- [x] **Design lightweight struct-based architecture instead of heavy classes**
  - [x] Define `Server` struct to hold server state
  - [x] Define `DocumentStore` for managing open documents
  - [ ] Define `SymbolIndex` for workspace symbols
  - [x] Use composition over inheritance
  - [x] Keep structs focused on data, functions separate

- [x] **Set up global/package-level state with mutex protection for open documents**
  - [x] Create `DocumentStore` struct with `sync.RWMutex`
  - [x] Implement `documents map[string]*Document` field
  - [x] Add methods: `Get()`, `Set()`, `Delete()`, `List()`
  - [x] Ensure all access goes through mutex-protected methods

- [x] **Implement thread-safe access patterns using sync.RWMutex**
  - [x] Use `RLock()`/`RUnlock()` for read operations
  - [x] Use `Lock()`/`Unlock()` for write operations
  - [x] Document locking requirements in comments
  - [x] Avoid holding locks during long operations

- [x] **Implement explicit error handling throughout (Go idioms)**
  - [x] Return errors from all functions that can fail
  - [x] Use `errors.New()` or `fmt.Errorf()` for error creation
  - [x] Check errors immediately: `if err != nil { return err }`
  - [x] Log errors before returning them to caller
  - [x] Never panic in production code paths

- [x] **Write test for initialize request/response cycle**
  - [x] Create `internal/lsp/initialize_test.go`
  - [x] Mock GLSP context
  - [x] Create test InitializeParams
  - [x] Call Initialize handler
  - [x] Assert InitializeResult contains expected capabilities
  - [x] Verify serverInfo is populated correctly

- [x] **Write test for shutdown request handling**
  - [x] Create test that calls Shutdown handler
  - [ ] Verify server state is marked as shutting down
  - [x] Verify no errors returned
  - [ ] Verify resources are cleaned up

- [x] **Manually verify server lifecycle with minimal LSP client**
  - [x] Write simple test script that sends initialize JSON-RPC message
  - [x] Verify InitializeResult is received
  - [ ] Send initialized notification
  - [ ] Send shutdown request
  - [ ] Send exit notification
  - [x] Verify process exits cleanly

**Outcome**: A running LSP server skeleton that communicates over STDIO/TCP, correctly handles the initialize/shutdown lifecycle, and logs its activity.

---

## Phase 1: Document Synchronization

**Goal**: Implement basic text document management for real-time synchronization with the editor.

### Tasks (15)

- [ ] **Implement textDocument/didOpen handler**
  - [ ] Create `internal/lsp/text_document.go`
  - [ ] Define handler: `func DidOpen(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error`
  - [ ] Extract URI, text, languageId, version from params
  - [ ] Create Document struct instance with content
  - [ ] Store document in DocumentStore
  - [ ] Log document open event
  - [ ] Trigger initial parse and diagnostics

- [ ] **Store document content in in-memory map (URI to struct with text and metadata)**
  - [ ] Define `Document` struct with fields: URI, Text, Version, LanguageID, AST, Diagnostics
  - [ ] Implement `DocumentStore.Set(uri string, doc *Document)`
  - [ ] Use map[string]*Document for storage
  - [ ] Ensure mutex protection during insertion

- [ ] **Record document version on open**
  - [ ] Store `version` field from DidOpenTextDocumentParams
  - [ ] Use version for conflict detection in later operations
  - [ ] Increment local version counter on changes

- [ ] **Implement textDocument/didClose handler**
  - [ ] Define handler: `func DidClose(context *glsp.Context, params *protocol.DidCloseTextDocumentParams) error`
  - [ ] Extract URI from params
  - [ ] Remove document from DocumentStore
  - [ ] Log document close event

- [ ] **Remove document from map on close to free memory**
  - [ ] Implement `DocumentStore.Delete(uri string)`
  - [ ] Clear AST references to allow garbage collection
  - [ ] Acquire write lock before deletion
  - [ ] Release lock after deletion

- [ ] **Send empty diagnostics array on document close**
  - [ ] Create empty PublishDiagnosticsParams for URI
  - [ ] Send notification via GLSP context
  - [ ] Verify client clears error markers

- [ ] **Implement textDocument/didChange handler**
  - [ ] Define handler: `func DidChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error`
  - [ ] Extract URI and version from params
  - [ ] Retrieve document from store
  - [ ] Iterate through contentChanges array
  - [ ] Apply each change to document text
  - [ ] Update document version
  - [ ] Trigger re-parse and diagnostics

- [ ] **Mark TextDocumentSync capability as incremental in Initialize**
  - [ ] Set `TextDocumentSyncKind = Incremental` in capabilities
  - [ ] Configure `TextDocumentSyncOptions.Change = Incremental`
  - [ ] Enable `OpenClose = true`

- [ ] **Implement full sync mode (replace whole text on change)**
  - [ ] Check if `contentChange.Range == nil` (indicates full sync)
  - [ ] Replace entire document text with `contentChange.Text`
  - [ ] Update version number
  - [ ] Mark document as modified

- [ ] **Implement incremental sync mode (apply text diffs)**
  - [ ] Check if `contentChange.Range != nil` (indicates incremental)
  - [ ] Extract Range (start line/char, end line/char) and new text
  - [ ] Calculate byte offset from line/character positions
  - [ ] Apply text replacement at specified range
  - [ ] Adjust document length tracking

- [ ] **Implement utility to apply TextDocumentContentChangeEvent diffs**
  - [ ] Create `internal/document/text_edit.go`
  - [ ] Implement `ApplyContentChange(text string, change ContentChangeEvent) (string, error)`
  - [ ] Convert LSP Position (line, character) to byte offset
  - [ ] Handle multi-line text correctly
  - [ ] Handle UTF-8/UTF-16 encoding issues (LSP uses UTF-16 code units)
  - [ ] Return updated text or error

- [ ] **Ensure thread safety when updating document content**
  - [ ] Acquire write lock before any document mutation
  - [ ] Hold lock for minimal duration (just the update)
  - [ ] Use defer to ensure lock release
  - [ ] Test concurrent didChange events

- [ ] **Handle workspace/didChangeConfiguration notification**
  - [ ] Define handler: `func DidChangeConfiguration(context *glsp.Context, params *protocol.DidChangeConfigurationParams) error`
  - [ ] Parse settings from params.Settings (JSON)
  - [ ] Update server configuration state
  - [ ] Log configuration changes

- [ ] **Set up data structures for configuration settings**
  - [ ] Define `Config` struct with fields: MaxProblems, Trace, etc.
  - [ ] Add `Config` to Server struct
  - [ ] Implement default configuration values
  - [ ] Support dynamic configuration updates

- [ ] **Write unit tests for document open/close/change handlers**
  - [ ] Test didOpen: verify document stored correctly
  - [ ] Test didClose: verify document removed
  - [ ] Test didChange full sync: verify text replaced
  - [ ] Test didChange incremental: verify diff applied
  - [ ] Test version tracking across operations
  - [ ] Test concurrent access (race detector enabled)

- [ ] **Test incremental update: open valid code, introduce error, fix error**
  - [ ] Create integration test with valid DWScript code
  - [ ] Simulate didOpen
  - [ ] Verify no diagnostics published
  - [ ] Simulate didChange introducing syntax error
  - [ ] Verify diagnostic published
  - [ ] Simulate didChange fixing error
  - [ ] Verify diagnostics cleared

**Outcome**: The LSP server fully supports document synchronization, properly tracking open documents and their changes in memory.

---

## Phase 2: Diagnostics (Syntax and Semantic Analysis)

**Goal**: Provide real-time error reporting with syntax and semantic diagnostics.

### Tasks (27)

- [ ] **Integrate go-dws lexer.New() for tokenization**
  - [ ] Import `github.com/CWBudde/go-dws/pkg/lexer` package
  - [ ] Create `internal/analysis/parse.go`
  - [ ] Implement `ParseDocument(text string, filename string) (*ast.Program, []Diagnostic, error)`
  - [ ] Create lexer instance: `l := lexer.New(text)`
  - [ ] Verify lexer initialization succeeds

- [ ] **Integrate go-dws parser.New() for AST generation**
  - [ ] Import `github.com/CWBudde/go-dws/pkg/parser` package
  - [ ] Create parser from lexer: `p := parser.New(l)`
  - [ ] Handle parser creation errors

- [ ] **Parse document text into AST on open/change using p.ParseProgram()**
  - [ ] Call `program := p.ParseProgram()`
  - [ ] Check if program is nil (parsing failed completely)
  - [ ] Handle parser panics with recover() if needed
  - [ ] Log parse duration for performance monitoring

- [ ] **Store AST with document in in-memory map**
  - [ ] Add `AST *ast.Program` field to Document struct
  - [ ] Store parsed AST in document after successful parse
  - [ ] Keep previous AST if parsing fails (for error recovery)
  - [ ] Clear AST on document close

- [ ] **Collect syntax errors from p.Errors() after parsing**
  - [ ] Call `syntaxErrors := p.Errors()` after ParseProgram
  - [ ] Check if error slice is empty
  - [ ] Iterate through each error string
  - [ ] Log total syntax error count

- [ ] **Convert parser errors to LSP Diagnostic objects with line/column info**
  - [ ] Create `convertSyntaxError(err string, source string) protocol.Diagnostic`
  - [ ] Parse error string to extract line/column info
  - [ ] Create Diagnostic with:
    - [ ] `severity = DiagnosticSeverity.Error`
    - [ ] `source = "go-dws-parser"`
    - [ ] `message` = error text
    - [ ] `range` = calculated from position

- [ ] **Use errors.FromStringErrors() to format errors with source context**
  - [ ] Import `github.com/CWBudde/go-dws/pkg/errors` package
  - [ ] Call `structuredErrors := errors.FromStringErrors(p.Errors(), source, filename)`
  - [ ] Extract position information from structured errors
  - [ ] Use structured error's Line, Column, Length fields

- [ ] **Extract position info from errors and fill Diagnostic.Range**
  - [ ] Convert 1-based line numbers to 0-based (LSP uses 0-based)
  - [ ] Convert column numbers appropriately
  - [ ] Create Range with Start and End positions
  - [ ] Handle errors without position (use line 0, col 0)
  - [ ] Calculate End position from error length or use single character

- [ ] **Mark syntax errors with severity Error**
  - [ ] Set `diagnostic.Severity = protocol.DiagnosticSeverityError`
  - [ ] Use consistent severity levels
  - [ ] Add related information if available

- [ ] **Create semantic.NewAnalyzer() for semantic analysis**
  - [ ] Import `github.com/CWBudde/go-dws/pkg/semantic` package
  - [ ] Create `internal/analysis/semantic.go`
  - [ ] Implement `AnalyzeDocument(program *ast.Program) ([]Diagnostic, error)`
  - [ ] Create analyzer: `analyzer := semantic.NewAnalyzer()`
  - [ ] Configure analyzer options if available

- [ ] **Call analyzer.Analyze(program) on successfully parsed AST**
  - [ ] Skip analysis if program is nil
  - [ ] Call `err := analyzer.Analyze(program)`
  - [ ] Handle analysis errors gracefully
  - [ ] Log analysis duration

- [ ] **Retrieve semantic errors from analyzer.Errors() or analyzer.StructuredErrors()**
  - [ ] Try `analyzer.StructuredErrors()` first if available
  - [ ] Fall back to `analyzer.Errors()` if structured not available
  - [ ] Collect all semantic error messages
  - [ ] Distinguish between errors and warnings

- [ ] **Convert semantic errors to LSP Diagnostic objects (type mismatches, unknown variables)**
  - [ ] Create `convertSemanticError()` function
  - [ ] Map go-dws error types to LSP diagnostics
  - [ ] Set appropriate severity:
    - [ ] Type errors: Error
    - [ ] Unknown identifiers: Error
    - [ ] Unused items: Warning
  - [ ] Include error code if available

- [ ] **Include warnings in diagnostics if available (unused variables)**
  - [ ] Set severity = DiagnosticSeverityWarning for warnings
  - [ ] Add diagnostic tags (Unnecessary, Deprecated) where appropriate
  - [ ] Allow configuration to enable/disable warnings

- [ ] **Implement textDocument/publishDiagnostics notification**
  - [ ] Create `PublishDiagnostics(ctx *glsp.Context, uri string, diagnostics []protocol.Diagnostic) error`
  - [ ] Build PublishDiagnosticsParams struct
  - [ ] Call `ctx.Notify(protocol.ServerNotificationTextDocumentPublishDiagnostics, params)`
  - [ ] Handle notification errors
  - [ ] Log diagnostics being published

- [ ] **Send PublishDiagnosticsParams with URI and diagnostics list to client**
  - [ ] Ensure URI is properly formatted (file:// scheme)
  - [ ] Include version if available
  - [ ] Sort diagnostics by position (line, then column)
  - [ ] Limit diagnostics count if configured (e.g., max 100)

- [ ] **Trigger diagnostics publishing on document open**
  - [ ] Call ParseDocument in didOpen handler
  - [ ] Collect all diagnostics (syntax + semantic)
  - [ ] Call PublishDiagnostics with results
  - [ ] Handle errors without crashing

- [ ] **Trigger diagnostics publishing on document change**
  - [ ] Call ParseDocument in didChange handler after text update
  - [ ] Re-run full analysis on each change
  - [ ] Publish updated diagnostics
  - [ ] Consider caching previous diagnostics

- [ ] **Set up workspace indexing data structures (symbol index)**
  - [ ] Create `internal/workspace/index.go`
  - [ ] Define `SymbolIndex` struct with:
    - [ ] `symbols map[string][]SymbolInfo` (name -> locations)
    - [ ] `files map[string]*FileInfo` (uri -> file metadata)
    - [ ] `mutex sync.RWMutex`
  - [ ] Define `SymbolInfo` struct: Name, Kind, Location, ContainerName
  - [ ] Implement Add, Remove, Search methods

- [ ] **Scan workspace for .dws files on initialized notification**
  - [ ] Implement `ScanWorkspace(rootURIs []string) error`
  - [ ] Use filepath.Walk to traverse directories
  - [ ] Filter files by .dws extension
  - [ ] Limit initial scan depth to avoid performance issues
  - [ ] Log progress during scan

- [ ] **Parse workspace files and build symbol index (name to definition map)**
  - [ ] Parse each .dws file found in workspace
  - [ ] Extract top-level symbols from AST:
    - [ ] Functions/procedures
    - [ ] Global variables
    - [ ] Classes/types
    - [ ] Constants
  - [ ] Add each symbol to index with location
  - [ ] Handle parse errors gracefully (skip file, log error)
  - [ ] Run indexing in background goroutine

- [ ] **Write unit tests for diagnostic generation with known error snippets**
  - [ ] Create `internal/analysis/diagnostics_test.go`
  - [ ] Test syntax errors:
    - [ ] Missing semicolon
    - [ ] Unclosed string
    - [ ] Invalid token
  - [ ] Test semantic errors:
    - [ ] Undefined variable
    - [ ] Type mismatch
    - [ ] Wrong argument count
  - [ ] Verify diagnostic positions are correct
  - [ ] Verify diagnostic messages are clear

- [ ] **Validate diagnostics using go-dws testdata scripts (zero false positives)**
  - [ ] Load test files from go-dws repository testdata/
  - [ ] Parse each valid test file
  - [ ] Assert zero diagnostics for valid code
  - [ ] Run all existing go-dws tests
  - [ ] Report any false positives found

- [ ] **Test that valid code produces no diagnostics**
  - [ ] Create suite of valid DWScript programs
  - [ ] Include:
    - [ ] Simple variable declarations
    - [ ] Function definitions
    - [ ] Class definitions
    - [ ] Control flow statements
  - [ ] Assert diagnostics array is empty for each

- [ ] **Test that erroneous code produces expected diagnostics**
  - [ ] Create suite of invalid programs with known errors
  - [ ] For each error type, verify:
    - [ ] Diagnostic is generated
    - [ ] Correct severity level
    - [ ] Correct position
    - [ ] Meaningful message
  - [ ] Test multiple errors in one file

- [ ] **Consider debouncing rapid didChange events to avoid diagnostic floods**
  - [ ] Implement debounce timer (e.g., 300ms delay)
  - [ ] Cancel previous timer on new didChange
  - [ ] Only run diagnostics after typing pause
  - [ ] Make debounce duration configurable
  - [ ] Ensure debounce doesn't delay diagnostics on didOpen
  - [ ] Test with rapid typing simulation

**Outcome**: Real-time syntax and semantic diagnostics are displayed in the editor as the user types, with errors and warnings properly highlighted.

---

## Phase 3: Hover Support

**Goal**: Provide type and symbol information on mouse hover.

### Tasks (14)

- [ ] Implement textDocument/hover request handler
- [ ] Retrieve document AST for hover position
- [ ] Enhance AST nodes with position metadata (line, column)
- [ ] Implement position-to-AST-node mapping utility
- [ ] Identify symbol at hover position (identifier/variable/function/class)
- [ ] For variables: find declaration and retrieve type information
- [ ] For functions: find definition and extract signature (params, return type)
- [ ] For classes/types: get definition and structure information
- [ ] Extract documentation comments if present (future enhancement)
- [ ] Construct Hover response with MarkupContent (markdown format)
- [ ] Format hover content with type info and signatures
- [ ] Handle hover on non-symbol locations (return empty/null)
- [ ] Write unit tests for hover on variables, functions, and classes/types
- [ ] Manually test hover in VSCode with sample DWScript files

**Outcome**: Hovering over symbols displays rich information including types, signatures, and documentation.

---

## Phase 4: Go-to Definition

**Goal**: Enable navigation to symbol definitions across files.

### Tasks (13)

- [ ] Implement textDocument/definition request handler
- [ ] Identify symbol at definition request position
- [ ] Resolve symbol to its definition location
- [ ] Handle local variables/parameters (find in current file AST)
- [ ] Handle class fields/methods (search class definition in AST)
- [ ] Handle global functions/variables (search current file first)
- [ ] Search workspace symbol index for cross-file definitions
- [ ] Handle unit imports (parse referenced unit files on-demand)
- [ ] Return Location with URI and Range of definition
- [ ] Handle multiple definitions (overloaded functions) - return array
- [ ] Write unit tests for local symbol definitions
- [ ] Write unit tests for global symbol definitions
- [ ] Write unit tests for cross-file definitions (unit imports)
- [ ] Manually test go-to-definition in VSCode

**Outcome**: Users can jump to symbol definitions with F12 or Ctrl+Click, even across multiple files.

---

## Phase 5: Find References

**Goal**: Find all usages of a symbol across the workspace.

### Tasks (14)

- [ ] Implement textDocument/references request handler
- [ ] Identify symbol at references request position
- [ ] Determine symbol scope (local vs global)
- [ ] For local symbols: search within same function/block AST
- [ ] For global symbols: search all open documents' ASTs
- [ ] Search workspace index for references in non-open files
- [ ] Create helper to scan AST nodes for matching identifier names
- [ ] Filter by scope to avoid false matches (same name, different context)
- [ ] Leverage semantic analyzer for symbol resolution if available
- [ ] Collect list of Locations for each reference
- [ ] Decide whether to include definition in references list
- [ ] Write unit tests for local symbol references
- [ ] Write unit tests for global symbol references
- [ ] Write unit tests for scope isolation (no spurious references)
- [ ] Manually test find references in VSCode

**Outcome**: Users can find all references to a symbol using Shift+F12, with proper scope filtering to avoid false positives.

---

## Phase 6: Document Symbols

**Goal**: Provide outline view of document structure.

### Tasks (11)

- [ ] Implement textDocument/documentSymbol request handler
- [ ] Traverse document AST to collect top-level symbols
- [ ] Collect functions/procedures (SymbolKind: Function/Method)
- [ ] Collect global variables/constants (SymbolKind: Variable/Constant)
- [ ] Collect types/classes/interfaces (SymbolKind: Class/Interface/Struct)
- [ ] For classes: add child DocumentSymbol entries for fields and methods
- [ ] Handle nested functions and inner classes hierarchically
- [ ] Map DWScript constructs to appropriate LSP SymbolKind
- [ ] Return hierarchical DocumentSymbol objects (preferred over flat)
- [ ] Include symbol names, kinds, ranges, and selection ranges
- [ ] Write unit tests for document symbols with functions and classes
- [ ] Verify hierarchical structure (class contains members as children)
- [ ] Manually test document symbols outline in VSCode

**Outcome**: The editor's outline view displays a hierarchical structure of all symbols in the document.

---

## Phase 7: Workspace Symbols

**Goal**: Enable global symbol search across the entire workspace.

### Tasks (9)

- [ ] Implement workspace/symbol request handler
- [ ] Mark workspaceSymbolProvider: true in server capabilities
- [ ] Leverage workspace symbol index built during initialization
- [ ] Search symbol index for query string matches (substring or prefix)
- [ ] Return list of SymbolInformation with name, kind, location, containerName
- [ ] Implement fallback: parse non-open files on-demand if index not available
- [ ] Optimize workspace symbol search performance (caching, indexing)
- [ ] Write unit tests for workspace symbol search across multiple files
- [ ] Manually test workspace symbol search in VSCode

**Outcome**: Users can quickly search for symbols across the entire project using Ctrl+T.

---

## Phase 8: Code Completion

**Goal**: Provide intelligent code completion suggestions.

### Tasks (27)

- [ ] **Implement textDocument/completion request handler**
  - [ ] Create `internal/lsp/completion.go`
  - [ ] Define handler: `func Completion(context *glsp.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error)`
  - [ ] Extract document URI and position from params
  - [ ] Retrieve document from store
  - [ ] Check if document and AST are available
  - [ ] Determine completion context (see below)
  - [ ] Collect completion items based on context
  - [ ] Return CompletionList with items

- [ ] **Determine completion context from cursor position**
  - [ ] Create `internal/analysis/completion_context.go`
  - [ ] Implement `DetermineContext(doc *Document, pos Position) (*CompletionContext, error)`
  - [ ] Analyze text before cursor position
  - [ ] Identify if inside a comment (skip completion)
  - [ ] Identify if inside a string literal (skip completion)
  - [ ] Check for member access pattern (identifier followed by dot)
  - [ ] Determine current scope from AST
  - [ ] Return context struct with: Type (general/member/keyword), Scope, ParentType

- [ ] **Detect trigger characters (dot for member access)**
  - [ ] Check `params.Context.TriggerKind == CompletionTriggerKindTriggerCharacter`
  - [ ] Check `params.Context.TriggerCharacter == "."`
  - [ ] Extract identifier before the dot
  - [ ] Set context type to MemberAccess
  - [ ] Store parent identifier for type resolution

- [ ] **Handle member access completion (object.): determine object type**
  - [ ] Create `ResolveMemberType(doc *Document, identifier string, pos Position) (Type, error)`
  - [ ] Search for identifier declaration in current scope
  - [ ] If local variable: get type from declaration
  - [ ] If parameter: get type from function signature
  - [ ] If field: get type from class definition
  - [ ] Query semantic analyzer for type information
  - [ ] Return resolved type or error if unknown

- [ ] **Retrieve type information from semantic analyzer**
  - [ ] Add `GetSymbolType(symbol string, position Position) (Type, error)` to analyzer
  - [ ] Use analyzer's symbol table to lookup type
  - [ ] Handle built-in types (Integer, String, Float, Boolean, etc.)
  - [ ] Handle user-defined types (classes, records)
  - [ ] Return type structure with methods and fields

- [ ] **List members (fields/methods) of determined type/class**
  - [ ] Create `GetTypeMembers(typeName string) ([]CompletionItem, error)`
  - [ ] Search AST for class/record definition
  - [ ] Extract all fields and their types
  - [ ] Extract all methods and their signatures
  - [ ] Extract all properties (getters/setters)
  - [ ] Create CompletionItem for each member:
    - [ ] Fields: kind = Field, detail = type
    - [ ] Methods: kind = Method, detail = signature
    - [ ] Properties: kind = Property, detail = type
  - [ ] Sort members alphabetically
  - [ ] Return member list

- [ ] **Handle general scope completion (no dot): provide keywords, variables, globals**
  - [ ] Create `CollectScopeCompletions(doc *Document, pos Position) ([]CompletionItem, error)`
  - [ ] Initialize empty items slice
  - [ ] Add keywords (if at statement start)
  - [ ] Add local variables and parameters
  - [ ] Add global symbols
  - [ ] Add built-in functions
  - [ ] Filter by prefix if user has typed partial identifier
  - [ ] Return combined list

- [ ] **Include language keywords in completion suggestions**
  - [ ] Define keyword list: begin, end, if, then, else, while, for, do, var, const, function, procedure, class, etc.
  - [ ] Create CompletionItems for each keyword:
    - [ ] kind = Keyword
    - [ ] detail = "DWScript keyword"
    - [ ] insertText = keyword
  - [ ] Only include if at appropriate position (e.g., statement start)
  - [ ] Optionally provide snippets for complex keywords (if-then-else, for-do)

- [ ] **List local variables and parameters in current scope**
  - [ ] Implement `FindEnclosingScope(ast *ast.Program, pos Position) (*ast.Scope, error)`
  - [ ] Traverse AST to find the function/block containing position
  - [ ] Extract variable declarations from that scope
  - [ ] Extract function parameters if in function body
  - [ ] For each variable/parameter, create CompletionItem:
    - [ ] kind = Variable or Parameter
    - [ ] label = name
    - [ ] detail = type (if available)
  - [ ] Return list of items

- [ ] **Determine current scope from cursor position in AST**
  - [ ] Create `FindNodeAtPosition(ast *ast.Program, pos Position) (ast.Node, error)`
  - [ ] Traverse AST recursively
  - [ ] Check if position is within node's range
  - [ ] Return the deepest (most specific) node containing position
  - [ ] From node, determine enclosing function, class, or global scope
  - [ ] Build scope chain (nested scopes)

- [ ] **Include global functions, types, and constants**
  - [ ] Extract top-level function declarations from AST
  - [ ] Extract global variable/constant declarations
  - [ ] Extract type/class definitions
  - [ ] For each, create CompletionItem:
    - [ ] Functions: kind = Function, detail = signature
    - [ ] Constants: kind = Constant, detail = type and value
    - [ ] Types: kind = Class/Interface/Struct
  - [ ] Include symbols from workspace index (other files)

- [ ] **Include built-in functions and types from DWScript**
  - [ ] Create `internal/builtins/builtins.go`
  - [ ] Define list of built-in functions:
    - [ ] PrintLn, Print, Length, Copy, Pos, IntToStr, StrToInt, etc.
  - [ ] Define list of built-in types:
    - [ ] Integer, Float, String, Boolean, Variant, etc.
  - [ ] For each built-in, create CompletionItem with:
    - [ ] kind = Function or Class
    - [ ] detail = signature or description
    - [ ] documentation = usage info (MarkupContent)
  - [ ] Return built-in items

- [ ] **Construct CompletionItem list with label, kind, detail**
  - [ ] Create CompletionItem struct for each suggestion
  - [ ] Set required fields:
    - [ ] label = display name
    - [ ] kind = appropriate SymbolKind
  - [ ] Set optional fields:
    - [ ] detail = type or signature summary
    - [ ] documentation = longer description (optional)
    - [ ] sortText = for custom ordering (optional)
    - [ ] filterText = for filtering (usually same as label)
  - [ ] Add all items to CompletionList

- [ ] **For functions: provide snippet-style insert text with parameters**
  - [ ] Parse function signature to extract parameters
  - [ ] Build snippet string: `FunctionName($1:param1, $2:param2)$0`
  - [ ] Use LSP snippet syntax with tabstops
  - [ ] Set `insertTextFormat = InsertTextFormat.Snippet`
  - [ ] Set `insertText = snippet string`
  - [ ] Example: `"WriteLine(${1:text})$0"`

- [ ] **Set insertTextFormat to Snippet where appropriate**
  - [ ] For functions with parameters: use Snippet
  - [ ] For control structures (if-then, for-do): use Snippet
  - [ ] For simple identifiers: use PlainText
  - [ ] Ensure editor supports snippets (check client capabilities)

- [ ] **Optionally implement completionItem/resolve for lazy resolution**
  - [ ] Mark `CompletionProvider.ResolveProvider = true` in capabilities
  - [ ] Implement resolve handler: `func CompletionResolve(context *glsp.Context, item *protocol.CompletionItem) (*protocol.CompletionItem, error)`
  - [ ] Use item.Data to store deferred resolution info
  - [ ] In resolve, add documentation, additional edits, etc.
  - [ ] This improves performance by deferring expensive computation

- [ ] **Cache global symbol suggestions for performance**
  - [ ] Create `CompletionCache` struct with:
    - [ ] `globalSymbols []CompletionItem`
    - [ ] `builtins []CompletionItem`
    - [ ] `keywords []CompletionItem`
    - [ ] `lastUpdate time.Time`
  - [ ] Rebuild cache when workspace changes
  - [ ] Use cached items for quick response
  - [ ] Invalidate cache on file changes

- [ ] **Optimize completion generation for fast response**
  - [ ] Target <100ms response time
  - [ ] Use cached data where possible
  - [ ] Limit completion list size (e.g., max 100 items)
  - [ ] Use goroutines for parallel symbol lookup (if safe)
  - [ ] Implement prefix filtering early to reduce processing
  - [ ] Profile and optimize hot paths

- [ ] **Write unit tests for variable name completion**
  - [ ] Create `internal/lsp/completion_test.go`
  - [ ] Test case: typing partial variable name
    - [ ] Setup: code with variables `alpha`, `beta`, `alphabet`
    - [ ] Input: cursor after `alp`
    - [ ] Expected: `alpha` and `alphabet` in results
    - [ ] Verify: `beta` not in results
  - [ ] Test case: parameter completion in function
  - [ ] Test case: local variable shadowing global

- [ ] **Write unit tests for member access completion**
  - [ ] Test case: member access on class instance
    - [ ] Setup: class with fields `Name`, `Age`, method `GetInfo()`
    - [ ] Input: `person.` (cursor after dot)
    - [ ] Expected: `Name`, `Age`, `GetInfo` in results
  - [ ] Test case: chained member access (`obj.field.`)
  - [ ] Test case: member access on built-in type
  - [ ] Verify completion includes correct kinds (Field, Method)

- [ ] **Write unit tests for keyword and built-in completion**
  - [ ] Test case: keyword completion at statement start
    - [ ] Input: cursor at beginning of line in function
    - [ ] Expected: `if`, `while`, `for`, `var`, etc. in results
  - [ ] Test case: built-in function completion
    - [ ] Expected: `PrintLn`, `IntToStr`, `Length`, etc.
  - [ ] Verify keywords not suggested in inappropriate contexts

- [ ] **Manually test completion in VSCode during typing**
  - [ ] Open DWScript file in VSCode with LSP active
  - [ ] Test auto-trigger (typing identifier prefix)
  - [ ] Test manual trigger (Ctrl+Space)
  - [ ] Test dot-trigger for member access
  - [ ] Verify completion list appearance and ordering
  - [ ] Test snippet expansion with Tab
  - [ ] Test filtering as you continue typing
  - [ ] Verify performance (no lag)

**Outcome**: As users type, they receive context-aware completion suggestions including keywords, variables, functions, and members.

---

## Phase 9: Signature Help

**Goal**: Show function signatures and parameters during function calls.

### Tasks (18)

- [ ] Implement textDocument/signatureHelp request handler
- [ ] Determine call context from cursor position
- [ ] Detect signature help triggers (opening parenthesis, comma)
- [ ] Find function being called (identifier before opening parenthesis)
- [ ] Handle incomplete AST: temporarily insert closing parenthesis for parsing
- [ ] Traverse tokens backward to identify function and count commas
- [ ] Retrieve function definition to get parameters and documentation
- [ ] Handle built-in functions with predefined signatures
- [ ] Construct SignatureHelp response with SignatureInformation
- [ ] Format signature label (function name with parameters and return type)
- [ ] Provide ParameterInformation array for each parameter
- [ ] Determine activeParameter index by counting commas before cursor
- [ ] Set activeSignature (0 unless function overloading supported)
- [ ] Handle function overloading with multiple signatures
- [ ] Write unit tests for signature help with multi-parameter functions
- [ ] Verify activeParameter highlighting at different cursor positions
- [ ] Manually test signature help in VSCode during function calls

**Outcome**: When calling functions, users see parameter hints with the current parameter highlighted.

---

## Phase 10: Rename Support

**Goal**: Enable symbol renaming across the codebase.

### Tasks (13)

- [ ] Implement textDocument/rename request handler
- [ ] Identify symbol at rename position
- [ ] Validate that symbol can be renamed (reject keywords, built-ins)
- [ ] Find all references of the symbol using references handler
- [ ] Prepare WorkspaceEdit with TextEdit for each reference
- [ ] Create TextEdit to replace old name with new name at each location
- [ ] Group TextEdits by file in WorkspaceEdit.DocumentChanges
- [ ] Handle document version checking to avoid stale renames
- [ ] Implement textDocument/prepareRename handler (optional)
- [ ] Return symbol range and placeholder text in prepareRename
- [ ] Write unit tests for variable/function rename
- [ ] Write unit tests for rename across multiple files
- [ ] Write tests rejecting rename of keywords/built-ins
- [ ] Manually test rename operation in VSCode

**Outcome**: Users can rename symbols with F2, and all references across the workspace are updated automatically.

---

## Phase 11: Semantic Tokens

**Goal**: Provide semantic syntax highlighting information.

### Tasks (28)

- [ ] Define SemanticTokensLegend with token types and modifiers
- [ ] Include token types: keyword, string, number, comment, variable, parameter, property, function, class, interface, enum
- [ ] Include modifiers: static, deprecated, declaration, readonly
- [ ] Advertise SemanticTokensProvider in server capabilities
- [ ] Implement textDocument/semanticTokens/full handler
- [ ] Traverse document AST to collect semantic tokens
- [ ] Classify identifiers by role: variable, parameter, property, function, class, etc.
- [ ] Tag variable declarations with declaration modifier
- [ ] Differentiate constants, enum members, and properties
- [ ] Classify function and method names appropriately
- [ ] Classify class names, type identifiers, interface names
- [ ] Tag literals (numbers, strings, booleans)
- [ ] Optionally tag comments (may be redundant with TextMate grammar)
- [ ] Ensure AST nodes have start/end position info for token ranges
- [ ] Extend parser to record end positions if needed
- [ ] Calculate token length from identifier name length
- [ ] Record [line, startChar, length, tokenType, tokenModifiers] for each token
- [ ] Sort tokens by position (required by LSP)
- [ ] Encode tokens in LSP relative format (delta encoding)
- [ ] Optionally implement textDocument/semanticTokens/full/delta for incremental updates
- [ ] Recompute semantic tokens after document changes
- [ ] Write unit tests for semantic token generation
- [ ] Verify correct classification of various constructs (variables, functions, classes)
- [ ] Configure VSCode extension with semantic token legend
- [ ] Manually test semantic highlighting in VSCode

**Outcome**: Enhanced syntax highlighting based on semantic understanding, with variables, functions, and types colored appropriately.

---

## Phase 12: Code Actions

**Goal**: Provide quick fixes and refactoring actions.

### Tasks (24)

- [ ] Implement textDocument/codeAction request handler
- [ ] Mark codeActionProvider: true in server capabilities
- [ ] Specify supported codeActionKinds (quickfix, refactor, etc.)
- [ ] Implement quick fix for 'Undeclared identifier' error
- [ ] Suggest 'Declare variable X' action with default type
- [ ] Insert var declaration at appropriate location (function top or global)
- [ ] Implement quick fix for 'Missing semicolon' error
- [ ] Suggest 'Insert missing semicolon' action
- [ ] Implement quick fix for unused variable warning
- [ ] Suggest removing or prefixing with underscore
- [ ] Implement refactoring: Organize uses/imports
- [ ] Remove unused unit references from uses clause
- [ ] Add missing unit references for used symbols
- [ ] Consider extract to function refactoring (complex, optional)
- [ ] Consider implement interface/abstract methods (complex, optional)
- [ ] Recognize diagnostic patterns using error codes or message matching
- [ ] Create CodeAction with appropriate kind (quickfix, refactor)
- [ ] Attach diagnostic as associatedDiagnostic in code action
- [ ] Provide WorkspaceEdit with changes to resolve issue
- [ ] Ensure code action titles clearly describe the fix
- [ ] Write unit tests for quick fix actions
- [ ] Verify applying edit resolves the diagnostic
- [ ] Manually test code actions in VSCode

**Outcome**: Users receive contextual quick fixes and refactoring suggestions via the lightbulb menu.

---

## Phase 13: Testing, Quality, and Finalization

**Goal**: Ensure robustness, performance, and code quality before release.

### Tasks (20)

- [ ] **Run comprehensive integration tests against real DWScript projects**
  - [ ] Identify or create sample DWScript projects for testing:
    - [ ] Small project (single file, ~100 LOC)
    - [ ] Medium project (5-10 files, ~1000 LOC)
    - [ ] Large project (50+ files, ~10000 LOC)
  - [ ] Create `test/integration/` directory
  - [ ] Write integration test suite that:
    - [ ] Starts LSP server
    - [ ] Opens project files
    - [ ] Executes all LSP operations
    - [ ] Verifies expected responses
  - [ ] Run tests with `-race` flag to detect race conditions
  - [ ] Document any issues found and verify fixes

- [ ] **Test all features together in VSCode with sample projects**
  - [ ] Install language server in VSCode (via manual configuration or extension)
  - [ ] Open each sample project
  - [ ] Test feature interactions:
    - [ ] Edit file → verify diagnostics update
    - [ ] Go-to-definition → verify file opens correctly
    - [ ] Rename symbol → verify all files updated
    - [ ] Complete code → verify suggestions accurate
  - [ ] Test rapid editing scenarios
  - [ ] Test opening/closing multiple files
  - [ ] Monitor for crashes or hangs

- [ ] **Verify no feature breaks another (document sync during go-to-def, etc.)**
  - [ ] Test document synchronization while:
    - [ ] Hover requests are in-flight
    - [ ] Completion is triggered
    - [ ] Find references is running
  - [ ] Test concurrent operations:
    - [ ] Edit in File A while finding references in File B
    - [ ] Trigger completion while diagnostics computing
  - [ ] Verify state consistency after each operation
  - [ ] Check for deadlocks or race conditions

- [ ] **Performance testing: ensure no IDE freezing during operations**
  - [ ] Use `pprof` to profile CPU usage
  - [ ] Test performance of critical operations:
    - [ ] didChange on large file (10000+ LOC)
    - [ ] Find references in large workspace
    - [ ] Completion with large symbol table
  - [ ] Measure latency:
    - [ ] Hover: <50ms target
    - [ ] Completion: <100ms target
    - [ ] Diagnostics: <500ms target
  - [ ] Identify and fix performance bottlenecks
  - [ ] Add performance regression tests

- [ ] **Optimize find-references for large projects (async, partial results)**
  - [ ] Implement workspace search in batches
  - [ ] Send partial results as they're found (streaming)
  - [ ] Use LSP progress notifications:
    - [ ] `window/workDoneProgress/create`
    - [ ] `$/progress` updates
  - [ ] Allow cancellation via `$/cancelRequest`
  - [ ] Test with large workspace (1000+ files)
  - [ ] Verify UI remains responsive

- [ ] **Implement progress reporting for long operations (optional)**
  - [ ] Identify operations that may take >1 second:
    - [ ] Workspace indexing
    - [ ] Find all references
    - [ ] Rename in large workspace
  - [ ] Implement `WorkDoneProgress` support
  - [ ] Create progress token and notify client
  - [ ] Update progress with percentage or message
  - [ ] Complete progress on finish
  - [ ] Test in VSCode (progress indicator shows)

- [ ] **Memory management: ensure closing documents frees data**
  - [ ] Run memory profiler: `go test -memprofile`
  - [ ] Test document lifecycle:
    - [ ] Open 100 documents
    - [ ] Check memory usage
    - [ ] Close all documents
    - [ ] Verify memory released (GC runs)
  - [ ] Check for memory leaks:
    - [ ] AST nodes retained after close
    - [ ] Goroutines not terminated
    - [ ] Caches not invalidated
  - [ ] Add finalizers or explicit cleanup if needed

- [ ] **Consider LRU cache for workspace file ASTs**
  - [ ] Implement LRU cache with size limit (e.g., 50 files)
  - [ ] When cache full, evict least recently used AST
  - [ ] Keep open documents always in cache
  - [ ] Re-parse on cache miss
  - [ ] Measure cache hit rate (aim for >80%)
  - [ ] Test with large workspace
  - [ ] Compare memory usage with/without LRU

- [ ] **Audit code for Go best practices and idioms**
  - [ ] Run `go vet` on all packages
  - [ ] Run `golint` or `staticcheck`
  - [ ] Review and fix all warnings
  - [ ] Check error handling:
    - [ ] All errors checked
    - [ ] No silent failures
    - [ ] Errors properly wrapped with context
  - [ ] Check for proper use of:
    - [ ] defer (cleanup)
    - [ ] context (cancellation)
    - [ ] channels (communication)
  - [ ] Review exported API for clarity

- [ ] **Ensure proper package naming and division (internal/lsp, internal/dwscript)**
  - [ ] Verify package structure:
    - [ ] `cmd/go-dws-lsp/` - main executable
    - [ ] `internal/lsp/` - LSP handlers
    - [ ] `internal/server/` - server state
    - [ ] `internal/document/` - document management
    - [ ] `internal/analysis/` - DWScript integration
    - [ ] `internal/workspace/` - workspace indexing
    - [ ] `internal/builtins/` - built-in symbols
  - [ ] Ensure no cyclic dependencies
  - [ ] Use `internal/` to hide implementation details
  - [ ] Document package responsibilities

- [ ] **Refactor to remove unnecessary global state**
  - [ ] Identify all package-level variables
  - [ ] Move state to `Server` struct where appropriate
  - [ ] Pass dependencies explicitly (dependency injection)
  - [ ] Keep only truly global things (constants, loggers)
  - [ ] Update tests to use struct-based state

- [ ] **Use struct to encapsulate server state (docs, caches)**
  - [ ] Define comprehensive `Server` struct:
    - [ ] `documents *DocumentStore`
    - [ ] `index *SymbolIndex`
    - [ ] `config *Config`
    - [ ] `logger *log.Logger`
    - [ ] `glspServer *glsp.Server`
  - [ ] Pass server instance to all handlers
  - [ ] Store server in GLSP context user data
  - [ ] Use methods on Server for operations

- [ ] **Double-check concurrency safety for all shared data**
  - [ ] Review all mutex usage:
    - [ ] Locks acquired in consistent order (avoid deadlock)
    - [ ] Locks held for minimal time
    - [ ] No locks held during I/O or RPC
  - [ ] Check for data races:
    - [ ] Run all tests with `-race`
    - [ ] Fix any reported races
  - [ ] Document locking strategy in comments
  - [ ] Consider using `sync.Map` for high-contention maps

- [ ] **Consider sync.Map or RWMutex for document/symbol access**
  - [ ] Profile lock contention
  - [ ] If high read contention: use `sync.RWMutex`
  - [ ] If high write contention: consider `sync.Map`
  - [ ] Benchmark different approaches
  - [ ] Choose best option for access patterns
  - [ ] Document choice and reasoning

- [ ] **Ensure high test coverage for all features**
  - [ ] Run `go test -cover ./...`
  - [ ] Target >80% code coverage
  - [ ] Focus on critical paths:
    - [ ] All LSP handlers
    - [ ] Parser/analyzer integration
    - [ ] Symbol resolution
  - [ ] Add tests for edge cases:
    - [ ] Empty files
    - [ ] Very large files
    - [ ] Malformed input
    - [ ] Unicode/UTF-8 handling
  - [ ] Use table-driven tests for multiple scenarios

- [ ] **Add scenario tests for complex features (completion, rename)**
  - [ ] Create `test/scenarios/` directory
  - [ ] Write end-to-end scenario tests:
    - [ ] Complete a class member access
    - [ ] Rename a variable used in multiple files
    - [ ] Go-to-definition across units
    - [ ] Find all references to overloaded function
  - [ ] Use realistic DWScript code samples
  - [ ] Verify complete LSP interaction flow
  - [ ] Automate with test harness

- [ ] **Document architecture and contribution guidelines**
  - [ ] Create `ARCHITECTURE.md`:
    - [ ] High-level design overview
    - [ ] Package responsibilities
    - [ ] Data flow diagrams
    - [ ] Key design decisions
  - [ ] Create `CONTRIBUTING.md`:
    - [ ] How to set up dev environment
    - [ ] How to run tests
    - [ ] Code style guidelines
    - [ ] Pull request process
  - [ ] Add code comments to complex algorithms
  - [ ] Document LSP protocol mapping

- [ ] **Update README with build/test instructions**
  - [ ] Add "Building" section:
    - [ ] Prerequisites (Go version)
    - [ ] Clone and dependencies
    - [ ] Build command: `go build ./cmd/go-dws-lsp`
  - [ ] Add "Testing" section:
    - [ ] Unit tests: `go test ./...`
    - [ ] Integration tests: `go test ./test/integration`
    - [ ] Coverage: `go test -cover ./...`
  - [ ] Add "Usage" section:
    - [ ] How to run server
    - [ ] Command-line flags
    - [ ] VSCode integration (link to extension repo)

- [ ] **Document important implementation details for contributors**
  - [ ] Document position encoding (UTF-16 vs UTF-8)
  - [ ] Document AST traversal patterns
  - [ ] Document symbol resolution strategy
  - [ ] Document caching strategy
  - [ ] Add example code for common tasks:
    - [ ] Adding a new LSP handler
    - [ ] Extending the semantic analyzer
    - [ ] Adding completion items
  - [ ] Document debugging techniques (TCP mode, logging)

- [ ] **Final manual testing pass in VSCode with all features**
  - [ ] Create comprehensive test checklist
  - [ ] Test each feature systematically:
    - [ ] ✓ Diagnostics (syntax and semantic)
    - [ ] ✓ Hover (variables, functions, types)
    - [ ] ✓ Go-to-definition (local and cross-file)
    - [ ] ✓ Find references
    - [ ] ✓ Document symbols (outline)
    - [ ] ✓ Workspace symbols (Ctrl+T)
    - [ ] ✓ Completion (identifiers and members)
    - [ ] ✓ Signature help
    - [ ] ✓ Rename
    - [ ] ✓ Semantic highlighting
    - [ ] ✓ Code actions
  - [ ] Test error scenarios (network loss, bad files, etc.)
  - [ ] Verify graceful degradation on errors

- [ ] **Verify feature-completeness against original plan**
  - [ ] Review goal.md and PLAN.md
  - [ ] Check off all implemented features
  - [ ] Document any deferred features (future work)
  - [ ] Ensure all "Outcome" goals are met
  - [ ] Prepare release notes with feature list
  - [ ] Tag release version (v1.0.0)
  - [ ] Celebrate! 🎉

**Outcome**: A production-ready, well-tested, performant language server with comprehensive documentation.

---

## Summary

This plan provides a systematic approach to building a complete LSP implementation for DWScript in Go. Each phase builds upon the previous one, ensuring steady progress with testable milestones. By following this plan, we will create a robust, idiomatic Go implementation that provides an excellent development experience for DWScript users.
