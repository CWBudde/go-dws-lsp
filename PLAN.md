# go-dws-lsp Implementation Plan

This document provides a detailed, phase-by-phase implementation plan for the go-dws Language Server Protocol (LSP) implementation. The plan breaks down the project into 14 distinct phases, each focusing on a specific feature or set of related features.

## Overview

The implementation is organized into the following phases:

- **Phase 0**: Foundation - LSP Scaffolding and Setup ✅
- **Phase 1**: Document Synchronization ✅
- **Phase 2**: go-dws API Enhancements for LSP Integration ✅
- **Phase 3**: Diagnostics (Syntax and Semantic Analysis) ✅ (mostly complete)
- **Phase 4**: Hover Support
- **Phase 5**: Go-to Definition
- **Phase 6**: Find References
- **Phase 7**: Document Symbols
- **Phase 8**: Workspace Symbols
- **Phase 9**: Code Completion
- **Phase 10**: Signature Help
- **Phase 11**: Rename Support
- **Phase 12**: Semantic Tokens
- **Phase 13**: Code Actions
- **Phase 14**: Testing, Quality, and Finalization

**Total Tasks**: 296 (254 original + 42 go-dws enhancements)

---

## Phase 0: Foundation - LSP Scaffolding and Setup ✅

**Status**: COMPLETE (21/21 tasks)

**Implemented:**
- Go module structure with `cmd/go-dws-lsp/` and `internal/` packages
- GLSP library integration for LSP protocol handling
- Initialize/Shutdown request handlers with full server capabilities advertised
- STDIO and TCP transport modes with command-line flags (`-tcp`, `-port`, `-log-level`, `-log-file`)
- Thread-safe `DocumentStore` with mutex-protected operations
- `Server` struct for state management with `Config` support
- Comprehensive initialize tests in `internal/lsp/initialize_test.go`

**Deferred:**
- [ ] **SymbolIndex implementation** (Phase 3) - Workspace symbol tracking
- [ ] **Shutdown cleanup** (Phase 14) - Resource cleanup and shutdown flag (low priority, server exits cleanly)

---

## Phase 1: Document Synchronization ✅

**Status**: COMPLETE (15/15 tasks)

**Implemented:**
- `textDocument/didOpen`, `didClose`, `didChange` handlers in `internal/lsp/text_document.go`
- Full and incremental sync modes with version tracking
- UTF-16 to UTF-8 position conversion utilities in `internal/document/text_edit.go`
- Empty diagnostics notification on document close
- `workspace/didChangeConfiguration` handler with dynamic config updates
- 27 comprehensive tests (10 handler tests + 17 text editing tests) - all passing

**Deferred:**
- [ ] **Trigger diagnostics on open/change** (Phase 3) - Parse and publish diagnostics when documents open or change

---

## Phase 2: go-dws API Enhancements for LSP Integration

**Goal**: Enhance the go-dws library to expose structured errors, AST access, and position metadata needed for LSP features.

**Repository**: `github.com/CWBudde/go-dws`

**Status**: MOSTLY COMPLETE (22/27 tasks = 81.5%)

**Current Version**: `v0.0.0-20251107150541-36cc51824199` (commit 36cc518)

**Why This Phase**: The current go-dws API provides string-based errors and opaque Program objects. To implement LSP features (hover, go-to-definition, completion, etc.), we need structured error information, direct AST access, and position metadata on AST nodes.

### Tasks (27)

- [x] **2.1 Create structured error types in pkg/dwscript**
  - [x] Create `pkg/dwscript/error.go` file
  - [x] Define `Error` struct with fields:
    - [x] `Message string` - The error message
    - [x] `Line int` - 1-based line number
    - [x] `Column int` - 1-based column number
    - [x] `Length int` - Length of the error span in characters
    - [x] `Severity string` - Either "error" or "warning"
    - [x] `Code string` - Optional error code (e.g., "E001", "W002")
  - [x] Implement `Error() string` method to satisfy error interface
  - [x] Add documentation explaining 1-based indexing

- [x] **2.2 Update CompileError to use structured errors**
  - [x] Change `CompileError.Errors` from `[]string` to `[]Error`
  - [x] Update `CompileError.Error()` method to format structured errors
  - [x] Ensure backwards compatibility or document breaking change
  - [x] Update all internal code that creates CompileError instances

- [x] **2.3 Update internal lexer to capture position metadata**
  - [x] Verify `internal/lexer/token.go` includes position information
  - [x] Ensure Token struct has `Line`, `Column`, `Offset` fields
  - [x] If missing, add position tracking to tokenization
  - [x] Add `Length` calculation for tokens (end - start)

- [x] **2.4 Update internal parser to capture error positions**
  - [x] Modify parser error generation to include line/column
  - [x] Change from `fmt.Sprintf()` strings to structured Error objects
  - [x] Extract position from current token when error occurs
  - [x] Calculate error span length where possible
  - [x] Update all parser error sites (syntax errors)

- [x] **2.5 Update internal semantic analyzer to capture error positions**
  - [x] Modify semantic analysis error generation
  - [x] Include position from AST node being analyzed
  - [x] Set appropriate severity (error vs warning)
  - [x] Add error codes for common semantic errors:
    - [x] "E_UNDEFINED_VAR" - Undefined variable
    - [x] "E_TYPE_MISMATCH" - Type mismatch
    - [x] "E_WRONG_ARG_COUNT" - Wrong argument count
    - [x] "W_UNUSED_VAR" - Unused variable (warning)
    - [x] "W_UNUSED_PARAM" - Unused parameter (warning)

- [x] **2.6 Add position metadata to AST node types**
  - [x] Open `internal/ast/ast.go`
  - [x] Define `Position` struct:
    - [x] `Line int` - 1-based line number
    - [x] `Column int` - 1-based column number
    - [x] `Offset int` - Byte offset from start of file
  - [x] Define `Node` interface with position methods:
    - [x] `Pos() Position` - Returns start position
    - [x] `End() Position` - Returns end position
  - [x] Document that all AST node types must implement Node interface

- [x] **2.7 Add position fields to statement AST nodes**
  - [x] Add `StartPos Position` and `EndPos Position` fields to:
    - [x] `Program`
    - [x] `BlockStatement`
    - [x] `ExpressionStatement`
    - [x] `AssignmentStatement`
    - [x] `IfStatement`
    - [x] `WhileStatement`
    - [x] `ForStatement`
    - [x] `ReturnStatement`
    - [x] `BreakStatement`
    - [x] `ContinueStatement`
  - [x] Implement `Pos()` and `End()` methods for each

- [x] **2.8 Add position fields to expression AST nodes**
  - [x] Add `StartPos Position` and `EndPos Position` fields to:
    - [x] `Identifier`
    - [x] `IntegerLiteral`
    - [x] `FloatLiteral`
    - [x] `StringLiteral`
    - [x] `BooleanLiteral`
    - [x] `BinaryExpression`
    - [x] `UnaryExpression`
    - [x] `CallExpression`
    - [x] `IndexExpression`
    - [x] `MemberExpression`
  - [x] Implement `Pos()` and `End()` methods for each

- [x] **2.9 Add position fields to declaration AST nodes**
  - [x] Add `StartPos Position` and `EndPos Position` fields to:
    - [x] `FunctionDeclaration`
    - [x] `ProcedureDeclaration`
    - [x] `VariableDeclaration`
    - [x] `ConstantDeclaration`
    - [x] `TypeDeclaration`
    - [x] `ClassDeclaration`
    - [x] `FieldDeclaration`
    - [x] `MethodDeclaration`
    - [x] `PropertyDeclaration`
  - [x] Implement `Pos()` and `End()` methods for each

- [x] **2.10 Update parser to populate position information**
  - [x] Modify parser to capture start position before parsing node
  - [x] Capture end position after parsing node
  - [x] Set `StartPos` from first token of construct
  - [x] Set `EndPos` from last token of construct
  - [x] Handle multi-line constructs correctly
  - [x] Test position accuracy with sample programs

- [x] **2.11 Export AST types as public API**
  - [x] Create `pkg/ast/` directory
  - [x] Copy AST types from `internal/ast/` to `pkg/ast/`
  - [x] Update package declaration to `package ast`
  - [x] Add comprehensive package documentation
  - [x] Export all node types (capitalize struct names if needed)
  - [x] Keep `internal/ast/` as alias to `pkg/ast/` for internal use
  - [x] OR: Make `internal/ast/` types directly accessible (less preferred)

- [x] **2.12 Add AST accessor to Program type**
  - [x] Open `pkg/dwscript/dwscript.go`
  - [x] Add method: `func (p *Program) AST() *ast.Program`
  - [x] Return the underlying parsed AST
  - [x] Add documentation explaining AST structure
  - [x] Explain that AST is read-only, modifications won't affect execution
  - [x] Add example in documentation showing AST traversal

- [x] **2.13 Add parse-only mode for LSP use cases**
  - [x] Add method to Engine: `func (e *Engine) Parse(source string) (*ast.Program, error)`
  - [x] Parse source code without semantic analysis
  - [x] Return partial AST even if syntax errors exist (best-effort)
  - [x] Return structured syntax errors only (no type checking errors)
  - [x] Document use case: "For editors/IDEs that need AST without full compilation"
  - [x] Optimize for speed (skip expensive semantic checks)

- [x] **2.14 Create visitor pattern for AST traversal** ✅
  - [x] Create `pkg/ast/visitor.go` (639 lines)
  - [x] Define `Visitor` interface:
    - [x] `Visit(node Node) (w Visitor)` - Standard Go AST walker pattern
  - [x] Implement `Walk(v Visitor, node Node)` function
  - [x] Handle all node types in Walk (64+ node types)
  - [x] Add documentation with examples
  - [x] Add `Inspect(node Node, f func(Node) bool)` helper

- [x] **2.15 Add symbol table access for semantic information** ✅
  - [x] Create `pkg/dwscript/symbols.go` (353 lines)
  - [x] Define `Symbol` struct:
    - [x] `Name string`
    - [x] `Kind string` - "variable", "function", "class", "parameter", etc.
    - [x] `Type string` - Type name
    - [x] `Position Position` - Definition location
    - [x] `Scope string` - "local", "global", "class"
  - [x] Add method: `func (p *Program) Symbols() []Symbol`
  - [x] Extract symbols from semantic analyzer's symbol table
  - [x] Include all declarations with their positions

- [x] **2.16 Add type information access** ⚠️ (Partially Complete)
  - [x] Add method: `func (p *Program) TypeAt(pos Position) (string, bool)` ✅
  - [x] Return type of expression at given position ✅
  - [x] Use semantic analyzer's type information ✅
  - [x] Return ("", false) if position doesn't map to typed expression ✅
  - [ ] Add method: `func (p *Program) DefinitionAt(pos Position) (*Position, bool)` ❌ (Deferred)
  - [ ] Return definition location for identifier at position ❌ (Deferred)

- [x] **2.17 Update error formatting for better IDE integration** ✅
  - [x] Ensure error messages are clear and concise
  - [x] Remove redundant position info from message text
  - [x] Use consistent error message format (severity at line:column: message [CODE])
  - [x] Add helper functions for position extraction
  - [x] Document error message format (in error.go)

- [x] **2.18 Write unit tests for structured errors** ✅
  - [x] Create `pkg/dwscript/error_test.go` (194 lines)
  - [x] Create `pkg/dwscript/error_format_test.go` (265 lines)
  - [x] Create `pkg/dwscript/compile_error_test.go` (192 lines)
  - [x] Test Error struct creation and formatting
  - [x] Test CompileError with multiple structured errors
  - [x] Test that positions are accurate
  - [x] Test severity levels (error vs warning)
  - [x] Test error codes

- [x] **2.19 Write unit tests for AST position metadata** ✅
  - [x] Create `pkg/ast/position_test.go` (333 lines)
  - [x] Test position on simple statements
  - [x] Test position on nested expressions
  - [x] Test position on multi-line constructs
  - [x] Test Pos() and End() methods on all node types
  - [x] Verify 1-based line numbering
  - [x] Test with Unicode/multi-byte characters

- [x] **2.20 Write unit tests for AST export** ✅
  - [x] Create `pkg/ast/ast_test.go` (373 lines)
  - [x] Test that Program.AST() returns valid AST
  - [x] Test AST traversal with visitor pattern
  - [x] Test AST structure for various programs
  - [x] Test that AST nodes have correct types
  - [x] Test accessing child nodes

- [x] **2.21 Write unit tests for Parse() mode** ✅
  - [x] Create `pkg/dwscript/parse_test.go` (342 lines)
  - [x] Test parsing valid code
  - [x] Test parsing code with syntax errors
  - [x] Verify partial AST is returned on error
  - [x] Test that structured errors are returned
  - [x] Compare Parse() vs Compile() behavior
  - [x] Measure performance difference

- [x] **2.22 Write integration tests** ✅
  - [x] Create `pkg/dwscript/integration_test.go` (587 lines)
  - [x] Test complete workflow: Parse → AST → Symbols
  - [x] Test error recovery scenarios
  - [x] Test position mapping accuracy
  - [x] Use real DWScript code samples
  - [x] Verify no regressions in existing functionality

- [x] **2.23 Update package documentation**
  - [x] Update `pkg/dwscript/doc.go` with new API
  - [x] Add examples for accessing AST
  - [x] Add examples for structured errors
  - [x] Document position coordinate system (1-based)
  - [x] Add migration guide if breaking changes
  - [x] Document LSP use case

- [x] **2.24 Update README with new capabilities**
  - [x] Add section on LSP/IDE integration
  - [x] Show example of using structured errors
  - [x] Show example of AST traversal
  - [x] Show example of symbol extraction
  - [x] Link to pkg.go.dev documentation
  - [x] Note minimum Go version if changed

- [x] **2.25 Verify backwards compatibility or version bump**
  - [x] Run all existing tests
  - [x] Check if API changes are backwards compatible
  - [x] If breaking: plan major version bump (v2.0.0)
  - [x] If compatible: plan minor version bump (v1.x.0)
  - [x] Update go.mod version if needed
  - [x] Document breaking changes in CHANGELOG

- [x] **2.26 Performance testing**
  - [x] Benchmark parsing with position tracking
  - [x] Ensure position metadata doesn't significantly slow parsing
  - [x] Target: <10% performance impact
  - [x] Benchmark Parse() vs Compile()
  - [x] Profile memory usage with AST export
  - [x] Optimize if needed

- [x] **2.27 Tag release and publish**
  - [x] Create git tag for new version
  - [x] Push tag to trigger pkg.go.dev update
  - [x] Write release notes
  - [x] Announce new LSP-friendly features
  - [x] Update go-dws-lsp dependency to new version

---

## Phase 3: Diagnostics (Syntax and Semantic Analysis)

**Goal**: Provide real-time error reporting with syntax and semantic diagnostics.

**Status**: MOSTLY COMPLETE (15/19 tasks)

**Prerequisites**: Phase 2 must be complete (structured errors and AST access available in go-dws) ✅

**Implemented:**

- Full diagnostic pipeline with structured errors from go-dws
- `ParseDocument` returns Program, diagnostics, and errors
- Document struct stores compiled Program for AST access
- `PublishDiagnostics` function in `internal/lsp/diagnostics.go`
- Diagnostics triggered on document open and change
- Severity mapping (Error, Warning, Info, Hint)
- Diagnostic tags (Unnecessary, Deprecated)
- Comprehensive test suite with 8 test functions

**Deferred:**
- [ ] **Workspace indexing** (Tasks 3.12-3.14) - Will be implemented when needed for Phase 7 (Workspace Symbols)
- [ ] **Debouncing** (Task 3.19) - Optional performance optimization, defer to Phase 14 (Testing & Quality)

### Tasks (19)

- [x] **3.1 Integrate go-dws engine for parsing and compilation** ✅
  - [x] Import `github.com/cwbudde/go-dws/pkg/dwscript` package ✅
  - [x] Update `internal/analysis/parse.go` ✅
  - [x] Implement `ParseDocument(text, filename) (*Program, []Diagnostic, error)` ✅
  - [x] Create engine instance: `engine, err := dwscript.New()` ✅
  - [x] Handle engine creation errors ✅

- [x] **3.2 Update ParseDocument to use Phase 2 structured errors** ✅
  - [x] Replace string-based error parsing with structured `dwscript.Error` types
  - [x] Access `CompileError.Errors []*Error` directly
  - [x] Use `Error.Line`, `Error.Column`, `Error.Length` for position
  - [x] Map `Error.Severity` to LSP DiagnosticSeverity
  - [x] Use `Error.Code` for diagnostic codes

- [x] **3.3 Update Document struct to store compiled Program** ✅
  - [x] Add `Program *dwscript.Program` field to Document struct in `internal/server/document_store.go`
  - [x] Store compiled program after successful compilation
  - [x] Program provides AST access via `program.AST()` method
  - [x] Keep nil program if compilation fails (for error recovery)
  - [x] Clear program on document close (via Delete)

- [x] **3.4 Convert compile errors to LSP Diagnostic objects** ✅
  - [x] Extract errors from `CompileError` via `convertStructuredErrors`
  - [x] Use structured error fields directly (no regex parsing)
  - [x] Create Diagnostic with appropriate fields
  - [x] Convert 1-based to 0-based line/column
  - [x] Set severity and source

- [x] **3.5 Simplify error conversion with structured errors** ✅
  - [x] No regex-based position extraction needed
  - [x] Directly use structured error fields
  - [x] Clean implementation in `convertStructuredError` function
  - [x] Simplified, reliable code

- [x] **3.6 Leverage semantic analysis from compilation** ✅
  - [x] `engine.Compile()` performs both syntax and semantic analysis
  - [x] Both syntax and semantic errors are in `CompileError.Errors`
  - [x] Use `Error.Code` to distinguish error types
  - [x] Semantic errors automatically included:
    - [x] Type mismatches
    - [x] Undefined variables
    - [x] Wrong argument counts
    - [x] Unused variables as warnings

- [x] **3.7 Add support for warnings** ✅
  - [x] Map `Error.Severity` using `mapSeverity` function
  - [x] Support all severity levels: Error, Warning, Info, Hint
  - [x] Add diagnostic tags via `mapDiagnosticTags`:
    - [x] `DiagnosticTag.Unnecessary` for unused variables/parameters/functions
    - [x] `DiagnosticTag.Deprecated` for deprecated constructs
  - [x] Warning level configurable via workspace settings (foundation ready)

- [x] **3.8 Implement textDocument/publishDiagnostics notification** ✅
  - [x] Create `PublishDiagnostics(ctx, uri, diagnostics)` in `internal/lsp/diagnostics.go`
  - [x] Build PublishDiagnosticsParams struct
  - [x] Call `ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, params)`
  - [x] Handle nil context gracefully
  - [x] Log diagnostics being published

- [x] **3.9 Send PublishDiagnosticsParams with URI and diagnostics list to client** ✅
  - [x] URI is properly formatted (passed through from params)
  - [x] Version tracking in Document struct
  - [x] Sort diagnostics by position (line, then column) via `sortDiagnostics`
  - [x] No artificial limit on diagnostics count

- [x] **3.10 Trigger diagnostics publishing on document open** ✅
  - [x] Call ParseDocument in DidOpen handler
  - [x] Collect all diagnostics (syntax + semantic)
  - [x] Call PublishDiagnostics with results
  - [x] Handle errors without crashing (store doc even if parse fails)

- [x] **3.11 Trigger diagnostics publishing on document change** ✅
  - [x] Call ParseDocument in DidChange handler after text update
  - [x] Re-run full analysis on each change
  - [x] Publish updated diagnostics
  - [x] Store new program in document

- [ ] **3.12 Set up workspace indexing data structures (symbol index)** (Deferred to Phase 7)
  - [ ] Create `internal/workspace/index.go`
  - [ ] Define `SymbolIndex` struct with:
    - [ ] `symbols map[string][]SymbolInfo` (name -> locations)
    - [ ] `files map[string]*FileInfo` (uri -> file metadata)
    - [ ] `mutex sync.RWMutex`
  - [ ] Define `SymbolInfo` struct: Name, Kind, Location, ContainerName
  - [ ] Implement Add, Remove, Search methods

- [ ] **3.13 Scan workspace for .dws files on initialized notification** (Deferred to Phase 7)
  - [ ] Implement `ScanWorkspace(rootURIs []string) error`
  - [ ] Use filepath.Walk to traverse directories
  - [ ] Filter files by .dws extension
  - [ ] Limit initial scan depth to avoid performance issues
  - [ ] Log progress during scan

- [ ] **3.14 Parse workspace files and build symbol index** (Deferred to Phase 7)
  - [ ] Parse each .dws file found in workspace
  - [ ] Extract top-level symbols from AST:
    - [ ] Functions/procedures
    - [ ] Global variables
    - [ ] Classes/types
    - [ ] Constants
  - [ ] Add each symbol to index with location
  - [ ] Handle parse errors gracefully (skip file, log error)
  - [ ] Run indexing in background goroutine

- [x] **3.15 Write unit tests for diagnostic generation** ✅
  - [x] Tests in `internal/analysis/parse_test.go`
  - [x] Test syntax errors:
    - [x] Missing semicolon
    - [x] Unclosed string
    - [x] Missing end keyword
  - [x] Test semantic errors:
    - [x] Undefined variable
    - [x] Type mismatch
    - [x] Wrong argument count
  - [x] Verify diagnostic positions are correct
  - [x] Verify diagnostic messages are clear

- [x] **3.16 Test structured error conversion** ✅
  - [x] Test `convertStructuredErrors` with multiple error types
  - [x] Test `convertStructuredError` for position conversion
  - [x] Test `mapSeverity` for all severity levels
  - [x] Test `mapDiagnosticTags` for warning tags
  - [x] Verify proper LSP Diagnostic structure

- [x] **3.17 Test that valid code produces no diagnostics** ✅
  - [x] Suite of valid DWScript programs
  - [x] Includes:
    - [x] Simple variable declarations
    - [x] Function definitions
    - [x] Empty program
  - [x] Assert diagnostics array is empty for each
  - [x] Assert non-nil Program returned

- [x] **3.18 Test that erroneous code produces expected diagnostics** ✅
  - [x] Suite of invalid programs with known errors
  - [x] For each error type, verify:
    - [x] Diagnostic is generated
    - [x] Correct severity level (checked in tests)
    - [x] Proper diagnostic structure
    - [x] Meaningful message
  - [x] Test multiple errors (unclosed string produces 3 errors)

- [ ] **3.19 Debouncing for rapid didChange events** (Deferred to Phase 14)
  - [ ] Implement debounce timer (e.g., 300ms delay)
  - [ ] Cancel previous timer on new didChange
  - [ ] Only run diagnostics after typing pause
  - [ ] Make debounce duration configurable
  - [ ] Ensure debounce doesn't delay diagnostics on didOpen
  - [ ] Test with rapid typing simulation

**Outcome**: Real-time syntax and semantic diagnostics are displayed in the editor as the user types, with errors and warnings properly highlighted. ✅

---

## Phase 4: Hover Support

**Goal**: Provide type and symbol information on mouse hover.

**Prerequisites**: Phase 2 and Phase 3 complete (structured errors, AST access, and diagnostics working) ✅

### Tasks (14)

- [x] **4.1 Implement textDocument/hover request handler** ✅
  - [x] Create `internal/lsp/hover.go`
  - [x] Define handler: `func Hover(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error)`
  - [x] Extract document URI and position from params
  - [x] Retrieve document from DocumentStore
  - [x] Check if document and AST are available
  - [x] Convert LSP position (UTF-16) to document position (UTF-8)
  - [x] Call helper function to get hover information
  - [x] Return Hover response or nil if no information available
  - [x] Register handler in server initialization

- [x] **4.2 Retrieve document AST for hover position** ✅
  - [x] Get document from store using URI
  - [x] Check if document has been parsed (Program exists)
  - [x] If no Program, return nil (document has errors)
  - [x] Get AST from Program: `program.AST()`
  - [x] Validate AST is not nil
  - [x] Pass AST and position to node finder

- [x] **4.3 Verify AST nodes have position metadata** ✅
  - [x] Verify all AST nodes from go-dws have `Pos()` and `End()` methods (from Phase 2) ✅
  - [x] Confirm position information is 1-based (line, column)
  - [x] Test position accuracy with sample code
  - [x] Document coordinate system (1-based in AST, 0-based in LSP)

- [x] **4.4 Implement position-to-AST-node mapping utility** ✅
  - [x] Create `internal/analysis/ast_node_finder.go`
  - [x] Implement `FindNodeAtPosition(ast *ast.Program, line, col int) ast.Node`
  - [x] Traverse AST recursively using visitor pattern
  - [x] Check if position is within node's range (Pos() to End())
  - [x] Return the most specific (deepest) node containing position
  - [x] Handle edge cases: empty files, position beyond file end
  - [x] Use `ast.Inspect()` helper from go-dws for traversal
  - [x] Write unit tests for node finding

- [x] **4.5 Identify symbol at hover position** ✅
  - [x] Check node type from FindNodeAtPosition
  - [x] Handle `*ast.Identifier` nodes (variable/function references)
  - [x] Handle `*ast.FunctionDeclaration` and `*ast.ProcedureDeclaration` nodes
  - [x] Handle `*ast.VariableDeclaration` nodes
  - [x] Handle `*ast.ClassDeclaration` and `*ast.TypeDeclaration` nodes
  - [x] Handle `*ast.MethodDeclaration` and `*ast.PropertyDeclaration` nodes
  - [x] Return symbol name and kind
  - [x] Return nil for non-symbol nodes (literals, operators, etc.)

- [x] **4.6 For variables: find declaration and retrieve type** ✅
  - [x] Create `internal/analysis/hover_info.go`
  - [x] Implement `GetVariableHoverInfo(program *dwscript.Program, identifier string, pos Position) (string, error)`
  - [x] Use `program.TypeAt(pos)` to get type information
  - [x] Search for variable declaration in AST
  - [x] Extract type from declaration node
  - [x] Handle local variables, parameters, and fields
  - [x] Format hover text: "var {name}: {type}"
  - [x] Add scope information (local vs global)

- [x] **4.7 For functions: extract signature (params, return type)** ✅
  - [x] Implement `GetFunctionHoverInfo(node *ast.FunctionDeclaration) string`
  - [x] Extract function name
  - [x] Extract parameters with types
  - [x] Extract return type
  - [x] Format signature: `function Name(param1: Type1, param2: Type2): ReturnType`
  - [x] Handle procedures (no return type)
  - [x] Handle methods (include class name)
  - [x] Format in markdown for rich display

- [x] **4.8 For classes/types: get definition and structure** ✅
  - [x] Implement `GetClassHoverInfo(node *ast.ClassDeclaration) string`
  - [x] Extract class name
  - [x] List public fields with types
  - [x] List methods with signatures
  - [x] List properties with types
  - [x] Show inheritance information if available
  - [x] Format in markdown with sections
  - [x] Handle records and interfaces similarly

- [x] **4.9 Extract documentation comments (future enhancement)** ✅
  - [x] Parse leading comments above declarations
  - [x] Extract doc comments in standard format (// or (* *))
  - [x] Include in hover response
  - [x] Format documentation text as markdown
  - [x] Handle multi-line documentation
  - [x] Note: May defer to later phase if complex

- [x] **4.10 Construct Hover response with MarkupContent** ✅
  - [x] Create `protocol.Hover` struct
  - [x] Set `Contents` field with MarkupContent
  - [x] Use `MarkupKind: protocol.Markdown` for rich formatting
  - [x] Format value with markdown syntax:
    - [x] Code blocks with ```dwscript
    - [x] Bold for keywords
    - [x] Sections for different info types
  - [x] Set `Range` field (optional) to highlight symbol

- [x] **4.11 Format hover content with type info and signatures** ✅
  - [x] Create `FormatHoverContent(symbolInfo SymbolInfo) string`
  - [x] Start with symbol kind (variable, function, class, etc.)
  - [x] Add type or signature in code block
  - [x] Add scope information if relevant
  - [x] Add documentation if available
  - [x] Use markdown formatting for readability
  - [x] Example format:
    ```markdown
    **variable** `x`
    ```dwscript
    var x: Integer
    ```
    Scope: local
    ```

- [ ] **4.12 Handle hover on non-symbol locations**
  - [ ] Return nil if FindNodeAtPosition returns nil
  - [ ] Return nil for literal nodes (numbers, strings)
  - [ ] Return nil for operators
  - [ ] Return nil for keywords
  - [ ] Return nil for comments
  - [ ] Return nil for whitespace
  - [ ] Log hover position for debugging (optional)

- [ ] **4.13 Write unit tests for hover functionality**
  - [ ] Create `internal/lsp/hover_test.go`
  - [ ] Test hover on variable declaration
  - [ ] Test hover on variable reference
  - [ ] Test hover on function declaration
  - [ ] Test hover on function call
  - [ ] Test hover on class declaration
  - [ ] Test hover on method declaration
  - [ ] Test hover on property
  - [ ] Test hover on built-in types
  - [ ] Test hover on non-symbol locations (should return nil)
  - [ ] Test hover with invalid positions
  - [ ] Test hover with missing AST (document with errors)
  - [ ] Use table-driven tests for multiple scenarios

- [ ] **4.14 Manually test hover in VSCode**
  - [ ] Open sample DWScript file in VSCode
  - [ ] Hover over variable declarations
  - [ ] Hover over variable references
  - [ ] Hover over function names
  - [ ] Hover over class names
  - [ ] Hover over method calls
  - [ ] Verify type information is displayed
  - [ ] Verify signatures are formatted correctly
  - [ ] Test with complex types (arrays, records, classes)
  - [ ] Verify hover works across different files
  - [ ] Check performance (should be instant)

**Outcome**: Hovering over symbols displays rich information including types, signatures, and documentation in a markdown-formatted popup.

**Estimated Effort**: 1-2 days

---

## Phase 5: Go-to Definition

**Goal**: Enable navigation to symbol definitions across files.

**Prerequisites**: Phase 2 and Phase 3 complete (structured errors, AST access, and diagnostics working) ✅

### Tasks (15)

- [x] **5.1 Implement textDocument/definition request handler** ✅
  - [x] Create `internal/lsp/definition.go`
  - [x] Define handler: `func Definition(context *glsp.Context, params *protocol.DefinitionParams) (interface{}, error)`
  - [x] Extract document URI and position from params
  - [x] Retrieve document from DocumentStore
  - [x] Check if document and AST are available
  - [x] Convert LSP position (UTF-16) to document position (UTF-8)
  - [x] Call helper function to find definition location
  - [x] Return Location, []Location, or nil based on results
  - [x] Register handler in server initialization

- [x] **5.2 Identify symbol at definition request position** ✅
  - [x] Reuse `FindNodeAtPosition` utility from hover implementation
  - [x] Get AST node at the requested position
  - [x] Check if node is an identifier or declaration
  - [x] Extract symbol name from node
  - [x] Determine symbol kind (variable, function, class, etc.)
  - [x] Handle member expressions (extract member name)
  - [x] Return nil if position is not on a symbol
  - [x] Log symbol identification for debugging

- [x] **5.3 Create symbol resolution framework** ✅
  - [x] Create `internal/analysis/symbol_resolver.go`
  - [x] Define `SymbolResolver` struct with document store reference
  - [x] Implement `ResolveSymbol(doc *Document, symbolName string, pos Position) ([]Location, error)`
  - [x] Define resolution strategy: local → class → global → workspace
  - [x] Return empty array if symbol not found
  - [x] Support multiple definitions (overloaded functions)
  - [x] Cache resolution results for performance (optional)

- [x] **5.4 Handle local variables/parameters (find in current file AST)** ✅
  - [x] Implement `ResolveLocalSymbol(ast *ast.Program, name string, pos Position) (*Location, error)`
  - [x] Find enclosing function or block at position
  - [x] Search function parameters for matching name
  - [x] Search local variable declarations in function body
  - [x] Check scope hierarchy (inner to outer blocks)
  - [x] Stop at first match (shadowing)
  - [x] Convert AST position to LSP Location
  - [x] Return nil if not found locally

- [ ] **5.5 Handle class fields/methods (search class definition in AST)**
  - [ ] Implement `ResolveClassMember(ast *ast.Program, className, memberName string) (*Location, error)`
  - [ ] Determine if cursor is within class context
  - [ ] Find class declaration in AST
  - [ ] Search class fields for matching name
  - [ ] Search class methods for matching name
  - [ ] Search class properties for matching name
  - [ ] Handle inherited members (search parent classes)
  - [ ] Return definition location with class URI

- [ ] **5.6 Handle global functions/variables (search current file first)**
  - [ ] Implement `ResolveGlobalSymbol(doc *Document, name string) (*Location, error)`
  - [ ] Get document AST
  - [ ] Search top-level function declarations
  - [ ] Search global variable declarations
  - [ ] Search constant declarations
  - [ ] Search type/class declarations
  - [ ] Return definition location in current file
  - [ ] Return nil if not found (will search workspace next)

- [ ] **5.7 Implement workspace symbol index for cross-file lookups**
  - [ ] Create `internal/workspace/symbol_index.go` (if not exists from Phase 3)
  - [ ] Define `SymbolIndex` struct with:
    - [ ] `symbols map[string][]SymbolLocation` (name → locations)
    - [ ] `files map[string]*FileInfo` (URI → metadata)
    - [ ] `mutex sync.RWMutex` for thread safety
  - [ ] Implement `AddSymbol(name, kind, uri string, range Range)`
  - [ ] Implement `FindSymbol(name string) []SymbolLocation`
  - [ ] Implement `RemoveFile(uri string)` for file deletions
  - [ ] Index symbols on workspace initialization

- [ ] **5.8 Search workspace symbol index for cross-file definitions**
  - [ ] Implement `ResolveWorkspaceSymbol(index *SymbolIndex, name string) ([]Location, error)`
  - [ ] Query workspace index for symbol name
  - [ ] Handle multiple matches (e.g., same name in different files)
  - [ ] Filter by symbol kind if needed (function vs variable)
  - [ ] Return all matching locations
  - [ ] Sort results by relevance (same package first)
  - [ ] Handle index unavailable gracefully (fallback to file scan)

- [ ] **5.9 Handle unit imports (parse referenced unit files on-demand)**
  - [ ] Implement `ParseImportedUnit(unitName string, workspaceRoot string) (*Document, error)`
  - [ ] Extract unit/import declarations from current file AST
  - [ ] Map unit name to file path (search workspace)
  - [ ] Parse imported file if not already in DocumentStore
  - [ ] Cache parsed unit for subsequent lookups
  - [ ] Search imported unit's AST for symbol
  - [ ] Return definition location from imported file
  - [ ] Handle circular imports gracefully

- [ ] **5.10 Return Location with URI and Range of definition**
  - [ ] Create `protocol.Location` struct for each definition
  - [ ] Set `URI` field with document URI (convert file path to URI)
  - [ ] Set `Range` field with start and end positions
  - [ ] Convert AST Position (1-based) to LSP Range (0-based)
  - [ ] Ensure Range covers the entire symbol name
  - [ ] Handle edge cases (symbol at end of file, multi-line)
  - [ ] Return single Location or array based on LSP spec

- [ ] **5.11 Handle multiple definitions (overloaded functions) - return array**
  - [ ] Check if symbol has multiple declarations (function overloading)
  - [ ] Collect all matching definitions into array
  - [ ] For each definition, create separate Location
  - [ ] Return `[]protocol.Location` instead of single Location
  - [ ] Distinguish overloads by parameter types (if available)
  - [ ] Order results by file (current file first, then imports)
  - [ ] Handle client capabilities (some clients may not support arrays)

- [ ] **5.12 Write unit tests for local symbol definitions**
  - [ ] Create `internal/lsp/definition_test.go`
  - [ ] Test go-to-definition on local variable declaration
  - [ ] Test go-to-definition on local variable reference
  - [ ] Test go-to-definition on function parameter
  - [ ] Test go-to-definition with shadowed variables (should go to nearest)
  - [ ] Test go-to-definition in nested blocks
  - [ ] Test go-to-definition on loop variables
  - [ ] Verify returned Location has correct URI and Range
  - [ ] Test with invalid positions (should return nil)

- [ ] **5.13 Write unit tests for global symbol definitions**
  - [ ] Test go-to-definition on global function declaration
  - [ ] Test go-to-definition on global function call
  - [ ] Test go-to-definition on global variable
  - [ ] Test go-to-definition on class name
  - [ ] Test go-to-definition on class field
  - [ ] Test go-to-definition on class method
  - [ ] Test go-to-definition with multiple definitions (overloads)
  - [ ] Verify array of Locations returned for overloads

- [ ] **5.14 Write unit tests for cross-file definitions (unit imports)**
  - [ ] Create test workspace with multiple .dws files
  - [ ] File A imports File B
  - [ ] Test go-to-definition from A to symbol defined in B
  - [ ] Test unit import resolution
  - [ ] Test definition in imported file
  - [ ] Verify correct URI returned (File B's URI)
  - [ ] Test with nested imports (A → B → C)
  - [ ] Test with symbol not found (should return nil)

- [ ] **5.15 Manually test go-to-definition in VSCode**
  - [ ] Open sample DWScript project in VSCode
  - [ ] Test F12 (go-to-definition) on local variable
  - [ ] Test Ctrl+Click on function name
  - [ ] Test go-to-definition on class field
  - [ ] Test go-to-definition on imported symbol
  - [ ] Verify cursor jumps to correct location
  - [ ] Test peek definition (Alt+F12)
  - [ ] Test with multiple definitions (should show picker)
  - [ ] Verify performance (should be instant)
  - [ ] Test across different files in workspace

**Outcome**: Users can jump to symbol definitions with F12 or Ctrl+Click, even across multiple files. Multiple definitions (overloaded functions) are shown in a picker menu.

**Estimated Effort**: 1-2 days

---

## Phase 6: Find References - EXPANDED

**Goal**: Find all usages of a symbol across the workspace.

**Prerequisites**: Phase 5 complete (symbol resolution working) for finding definitions first

### Tasks (15)

- [x] **6.1 Implement textDocument/references request handler**
  - [x] Create `internal/lsp/references.go`
  - [x] Define handler: `func References(context *glsp.Context, params *protocol.ReferenceParams) ([]protocol.Location, error)`
  - [x] Extract document URI, position, and context from params
  - [x] Check `params.Context.IncludeDeclaration` flag
  - [x] Retrieve document from DocumentStore
  - [x] Convert LSP position to document position
  - [x] Call helper function to find all references
  - [x] Return array of Locations (may be empty)
  - [x] Register handler in server initialization

- [ ] **6.2 Identify symbol at references request position**
  - [ ] Reuse `FindNodeAtPosition` from hover/definition
  - [ ] Get AST node at position
  - [ ] Extract symbol name and kind
  - [ ] Determine symbol scope (local, global, class member)
  - [ ] For member access, extract member name
  - [ ] Return nil if not on a symbol
  - [ ] Log symbol for debugging

- [ ] **6.3 Determine symbol scope (local vs global)**
  - [ ] Create `internal/analysis/scope_detector.go`
  - [ ] Implement `DetermineScope(ast *ast.Program, symbolName string, pos Position) (ScopeType, error)`
  - [ ] Define ScopeType enum: Local, Global, ClassMember, Parameter
  - [ ] For local: check if within function/block
  - [ ] For global: check if top-level declaration
  - [ ] For class member: check if within class
  - [ ] Return scope type and enclosing context

- [ ] **6.4 For local symbols: search within same function/block AST**
  - [ ] Implement `FindLocalReferences(ast *ast.Program, symbolName string, scopeNode ast.Node) ([]Location, error)`
  - [ ] Find enclosing function/procedure containing the symbol
  - [ ] Traverse only that function's AST
  - [ ] Use AST visitor to find all Identifier nodes
  - [ ] Match identifier name with symbol name
  - [ ] Collect positions of all matches
  - [ ] Convert AST positions to LSP Locations
  - [ ] Exclude matches in nested scopes with shadowing

- [ ] **6.5 For global symbols: search all open documents' ASTs**
  - [ ] Implement `FindGlobalReferences(symbolName string, docStore *DocumentStore) ([]Location, error)`
  - [ ] Iterate through all open documents in DocumentStore
  - [ ] For each document with valid AST:
    - [ ] Traverse AST with visitor pattern
    - [ ] Find all Identifier nodes matching symbol name
    - [ ] Check if identifier refers to global scope
    - [ ] Collect location with document URI
  - [ ] Return combined list from all documents

- [ ] **6.6 Search workspace index for references in non-open files**
  - [ ] Extend `SymbolIndex` to track references (not just definitions)
  - [ ] Implement `FindReferencesInIndex(symbolName string) ([]Location, error)`
  - [ ] Query index for all files containing symbol
  - [ ] For files not in DocumentStore:
    - [ ] Parse file on-demand
    - [ ] Search AST for references
    - [ ] Cache result for performance
  - [ ] Return all workspace references

- [ ] **6.7 Create helper to scan AST nodes for matching identifier names**
  - [ ] Implement `ScanASTForIdentifier(ast *ast.Program, name string) ([]Position, error)`
  - [ ] Use `ast.Inspect()` visitor helper
  - [ ] Visit all nodes in AST
  - [ ] Check if node is `*ast.Identifier`
  - [ ] Match `identifier.Name` with target name
  - [ ] Collect all matching positions
  - [ ] Return position array

- [ ] **6.8 Filter by scope to avoid false matches (same name, different context)**
  - [ ] Implement `FilterByScope(references []Location, targetScope Scope) []Location`
  - [ ] For each reference location:
    - [ ] Parse enclosing scope at that location
    - [ ] Check if scope matches target scope
    - [ ] Exclude if in different scope (e.g., different function's local var)
  - [ ] Handle shadowing (local var vs global var with same name)
  - [ ] Return filtered list

- [ ] **6.9 Leverage semantic analyzer for symbol resolution**
  - [ ] Use `program.Symbols()` to get semantic information
  - [ ] Match symbol by definition location, not just name
  - [ ] For each identifier, resolve to its definition
  - [ ] Only include references that resolve to target definition
  - [ ] This provides accurate filtering (no false positives)
  - [ ] Handle cases where semantic info unavailable (fallback to name matching)

- [ ] **6.10 Collect list of Locations for each reference**
  - [ ] For each found reference, create `protocol.Location`
  - [ ] Set URI to document containing reference
  - [ ] Set Range to cover identifier span
  - [ ] Convert AST Position (1-based) to LSP Range (0-based)
  - [ ] Add to results array
  - [ ] Sort by file then position

- [ ] **6.11 Include/exclude definition based on context flag**
  - [ ] Check `params.Context.IncludeDeclaration` flag
  - [ ] If true: include definition location in results
  - [ ] If false: exclude definition, only show references
  - [ ] Find definition using go-to-definition logic
  - [ ] Insert definition at beginning of results array (conventional)

- [ ] **6.12 Write unit tests for local symbol references**
  - [ ] Create `internal/lsp/references_test.go`
  - [ ] Test find references for local variable
  - [ ] Test references within same function
  - [ ] Test that references in other functions not included
  - [ ] Test with shadowed variable (only show correct scope)
  - [ ] Test with includeDeclaration true/false
  - [ ] Verify correct number of references returned
  - [ ] Verify each Location has correct URI and Range

- [ ] **6.13 Write unit tests for global symbol references**
  - [ ] Test find references for global function
  - [ ] Test references across multiple functions in same file
  - [ ] Test references in different open documents
  - [ ] Test references for class name
  - [ ] Test references for class method
  - [ ] Verify all occurrences found
  - [ ] Test performance with large files

- [ ] **6.14 Write unit tests for scope isolation (no spurious references)**
  - [ ] Create test with multiple symbols with same name
  - [ ] Local variable `x` in function A
  - [ ] Local variable `x` in function B
  - [ ] Find references for `x` in A should not include `x` in B
  - [ ] Test class field vs local variable with same name
  - [ ] Test parameter vs global variable with same name
  - [ ] Verify filtering works correctly

- [ ] **6.15 Manually test find references in VSCode**
  - [ ] Open sample DWScript project
  - [ ] Right-click on variable, select "Find All References" (Shift+F12)
  - [ ] Verify all references highlighted
  - [ ] Test on local variable (should show only in function)
  - [ ] Test on global function (should show all calls)
  - [ ] Test on class field (should show all field accesses)
  - [ ] Test across multiple files
  - [ ] Verify references panel shows correct file/line
  - [ ] Test includeDeclaration behavior

**Outcome**: Users can find all references to a symbol using Shift+F12, with proper scope filtering to avoid false positives. Results are shown in the references panel with file locations.

**Estimated Effort**: 1-2 days

---

## Phase 7: Document Symbols - EXPANDED

**Goal**: Provide outline view of document structure.

**Prerequisites**: Phase 2 complete (AST access)

### Tasks (13)

- [ ] **7.1 Implement textDocument/documentSymbol request handler**
  - [ ] Create `internal/lsp/document_symbol.go`
  - [ ] Define handler: `func DocumentSymbol(context *glsp.Context, params *protocol.DocumentSymbolParams) ([]interface{}, error)`
  - [ ] Extract document URI from params
  - [ ] Retrieve document from DocumentStore
  - [ ] Check if document has valid AST
  - [ ] Call helper to collect symbols
  - [ ] Return array of DocumentSymbol or SymbolInformation
  - [ ] Register handler in server initialization

- [ ] **7.2 Traverse document AST to collect top-level symbols**
  - [ ] Implement `CollectDocumentSymbols(ast *ast.Program) ([]protocol.DocumentSymbol, error)`
  - [ ] Visit root program node
  - [ ] Collect all top-level declarations:
    - [ ] Function and procedure declarations
    - [ ] Variable and constant declarations
    - [ ] Type, class, interface declarations
    - [ ] Unit declaration (if present)
  - [ ] For each symbol, extract name, kind, and range
  - [ ] Build hierarchical structure

- [ ] **7.3 Collect functions/procedures**
  - [ ] Visit `*ast.FunctionDeclaration` nodes
  - [ ] Extract function name from node
  - [ ] Set symbol kind to `SymbolKind.Function` or `SymbolKind.Method`
  - [ ] Set range to entire function span (from `function` to `end`)
  - [ ] Set selectionRange to function name only
  - [ ] Extract parameters and return type for detail field
  - [ ] Add to symbols array

- [ ] **7.4 Collect global variables/constants**
  - [ ] Visit `*ast.VariableDeclaration` nodes at global scope
  - [ ] Visit `*ast.ConstantDeclaration` nodes
  - [ ] Extract variable/constant name
  - [ ] Set symbol kind to `SymbolKind.Variable` or `SymbolKind.Constant`
  - [ ] Extract type information for detail field
  - [ ] Set range and selectionRange
  - [ ] Add to symbols array

- [ ] **7.5 Collect types/classes/interfaces**
  - [ ] Visit `*ast.ClassDeclaration` nodes
  - [ ] Visit `*ast.TypeDeclaration` nodes
  - [ ] Visit `*ast.InterfaceDeclaration` nodes (if supported)
  - [ ] Extract type/class name
  - [ ] Set symbol kind to `SymbolKind.Class`, `SymbolKind.Interface`, or `SymbolKind.Struct`
  - [ ] Set range to entire class definition
  - [ ] Set selectionRange to class name
  - [ ] Collect children (fields, methods, properties)

- [ ] **7.6 For classes: add child DocumentSymbol entries for fields and methods**
  - [ ] For each `*ast.ClassDeclaration`:
    - [ ] Create DocumentSymbol for class itself
    - [ ] Iterate class.Fields:
      - [ ] Create child DocumentSymbol with kind `SymbolKind.Field`
      - [ ] Add to parent's children array
    - [ ] Iterate class.Methods:
      - [ ] Create child DocumentSymbol with kind `SymbolKind.Method`
      - [ ] Include method signature in detail
      - [ ] Add to parent's children array
    - [ ] Iterate class.Properties:
      - [ ] Create child DocumentSymbol with kind `SymbolKind.Property`
      - [ ] Add to parent's children array

- [ ] **7.7 Handle nested functions and inner classes hierarchically**
  - [ ] Check for nested function declarations
  - [ ] For nested functions:
    - [ ] Create DocumentSymbol
    - [ ] Add as child of enclosing function
  - [ ] Check for inner class declarations (if supported by DWScript)
  - [ ] Build tree structure reflecting nesting
  - [ ] Ensure children array properly populated

- [ ] **7.8 Map DWScript constructs to appropriate LSP SymbolKind**
  - [ ] Define mapping function: `MapToSymbolKind(nodeType ast.NodeType) protocol.SymbolKind`
  - [ ] Mappings:
    - [ ] Function → SymbolKind.Function
    - [ ] Procedure → SymbolKind.Function
    - [ ] Method → SymbolKind.Method
    - [ ] Class → SymbolKind.Class
    - [ ] Record → SymbolKind.Struct
    - [ ] Interface → SymbolKind.Interface
    - [ ] Enum → SymbolKind.Enum
    - [ ] Variable → SymbolKind.Variable
    - [ ] Constant → SymbolKind.Constant
    - [ ] Field → SymbolKind.Field
    - [ ] Property → SymbolKind.Property

- [ ] **7.9 Return hierarchical DocumentSymbol objects (preferred over flat)**
  - [ ] Use `protocol.DocumentSymbol` struct (hierarchical)
  - [ ] Set required fields: Name, Kind, Range, SelectionRange
  - [ ] Set optional fields: Detail, Children
  - [ ] Build tree structure with parent-child relationships
  - [ ] Alternative: support flat SymbolInformation for older clients
  - [ ] Check client capabilities to choose format

- [ ] **7.10 Include symbol names, kinds, ranges, and selection ranges**
  - [ ] For each DocumentSymbol:
    - [ ] Name: the symbol identifier
    - [ ] Kind: mapped SymbolKind
    - [ ] Range: full span of symbol including body
    - [ ] SelectionRange: just the symbol name identifier
    - [ ] Detail: type signature or additional info
    - [ ] Children: nested symbols (optional)
  - [ ] Ensure ranges are 0-based (LSP format)
  - [ ] Ensure ranges are valid (end >= start)

- [ ] **7.11 Write unit tests for document symbols with functions and classes**
  - [ ] Create `internal/lsp/document_symbol_test.go`
  - [ ] Test document with functions only
  - [ ] Test document with global variables
  - [ ] Test document with class declarations
  - [ ] Test document with nested elements
  - [ ] Verify symbol count correct
  - [ ] Verify each symbol has correct kind
  - [ ] Verify hierarchical structure

- [ ] **7.12 Verify hierarchical structure (class contains members as children)**
  - [ ] Test that class DocumentSymbol has children array
  - [ ] Test that children include all fields
  - [ ] Test that children include all methods
  - [ ] Test that children include all properties
  - [ ] Verify child ranges are within parent range
  - [ ] Test nested classes (if supported)

- [ ] **7.13 Manually test document symbols outline in VSCode**
  - [ ] Open DWScript file in VSCode
  - [ ] Open Outline view (Ctrl+Shift+O)
  - [ ] Verify all functions listed
  - [ ] Verify all classes listed
  - [ ] Verify class members shown as children (indented)
  - [ ] Test clicking on symbol (should jump to definition)
  - [ ] Test search in outline (type to filter)
  - [ ] Verify icons for different symbol types
  - [ ] Test with large file (100+ symbols)

**Outcome**: The editor's outline view displays a hierarchical structure of all symbols in the document, with functions, classes, and members properly nested.

**Estimated Effort**: 1 day

---

## Phase 8: Workspace Symbols - EXPANDED

**Goal**: Enable global symbol search across the entire workspace.

**Prerequisites**: Symbol index from Phase 5 (or build index during this phase)

### Tasks (11)

- [ ] **8.1 Implement workspace/symbol request handler**
  - [ ] Create `internal/lsp/workspace_symbol.go`
  - [ ] Define handler: `func WorkspaceSymbol(context *glsp.Context, params *protocol.WorkspaceSymbolParams) ([]protocol.SymbolInformation, error)`
  - [ ] Extract query string from params
  - [ ] Access workspace symbol index
  - [ ] Call search function with query
  - [ ] Return array of SymbolInformation
  - [ ] Register handler in server initialization
  - [ ] Handle empty query (return all symbols or limit)

- [ ] **8.2 Mark workspaceSymbolProvider: true in server capabilities**
  - [ ] In initialize handler, set `capabilities.WorkspaceSymbolProvider = true`
  - [ ] Verify capability is advertised to client
  - [ ] Test that VSCode enables workspace symbol search (Ctrl+T)

- [ ] **8.3 Ensure workspace symbol index is built during initialization**
  - [ ] In initialized notification handler:
    - [ ] Get workspace folders from client
    - [ ] Call `BuildWorkspaceIndex(workspaceFolders)` in background
    - [ ] Use goroutine to avoid blocking initialization
    - [ ] Send progress notifications (optional)
  - [ ] Handle workspace with no folders (no-op)
  - [ ] Handle workspace changes (rebuild index)

- [ ] **8.4 Implement workspace indexing: scan for .dws files**
  - [ ] Implement `BuildWorkspaceIndex(roots []string) error`
  - [ ] For each workspace root:
    - [ ] Use `filepath.Walk` to traverse directories
    - [ ] Filter files by `.dws` extension
    - [ ] Skip hidden directories (`.git`, `node_modules`)
    - [ ] Collect file paths
  - [ ] Limit depth to avoid scanning too deep
  - [ ] Return list of .dws files to index

- [ ] **8.5 Parse workspace files and extract symbols**
  - [ ] For each .dws file in workspace:
    - [ ] Read file contents
    - [ ] Use `engine.Parse()` to get AST
    - [ ] Handle parse errors gracefully (log and skip file)
    - [ ] Extract top-level symbols from AST:
      - [ ] Functions/procedures
      - [ ] Classes/types
      - [ ] Global variables/constants
    - [ ] Add each symbol to index with location

- [ ] **8.6 Add symbols to index with name, kind, location, containerName**
  - [ ] For each extracted symbol:
    - [ ] Create `SymbolInformation` struct
    - [ ] Set Name to symbol identifier
    - [ ] Set Kind to appropriate SymbolKind
    - [ ] Set Location with file URI and range
    - [ ] Set ContainerName (e.g., class name for methods, file name for globals)
  - [ ] Call `symbolIndex.AddSymbol(symbolInfo)`
  - [ ] Update index statistics (total symbols)

- [ ] **8.7 Search symbol index for query string matches (substring or prefix)**
  - [ ] Implement `SearchIndex(query string) ([]SymbolInformation, error)`
  - [ ] Convert query to lowercase for case-insensitive search
  - [ ] For each symbol in index:
    - [ ] Check if symbol name contains query (substring match)
    - [ ] OR check if symbol name starts with query (prefix match)
    - [ ] Add to results if matches
  - [ ] Limit results to reasonable number (e.g., 100)
  - [ ] Sort results by relevance (exact match first, then prefix, then substring)

- [ ] **8.8 Implement fallback: parse non-open files on-demand if index not available**
  - [ ] If symbol index not built yet:
    - [ ] Fall back to on-demand search
    - [ ] Use `filepath.Walk` to find .dws files
    - [ ] Parse each file and search AST
    - [ ] Return first N matches
  - [ ] This provides basic functionality while index builds
  - [ ] Log warning that index is not ready

- [ ] **8.9 Optimize workspace symbol search performance**
  - [ ] Use map for O(1) lookup by name
  - [ ] Use trie for efficient prefix search (optional)
  - [ ] Cache search results for repeated queries
  - [ ] Limit search to first 1000 files in very large workspaces
  - [ ] Use goroutines for parallel file parsing (with limit)
  - [ ] Measure and optimize search time (target <100ms)

- [ ] **8.10 Write unit tests for workspace symbol search across multiple files**
  - [ ] Create `internal/lsp/workspace_symbol_test.go`
  - [ ] Create test workspace with multiple .dws files
  - [ ] Index the test workspace
  - [ ] Test exact name match (query = "Foo")
  - [ ] Test prefix match (query = "Get")
  - [ ] Test substring match (query = "User")
  - [ ] Test empty query
  - [ ] Test query with no matches
  - [ ] Verify all results have correct URI and range
  - [ ] Verify results include symbols from all files

- [ ] **8.11 Manually test workspace symbol search in VSCode**
  - [ ] Open DWScript workspace in VSCode
  - [ ] Press Ctrl+T to open workspace symbol search
  - [ ] Type partial symbol name
  - [ ] Verify results appear as you type
  - [ ] Test clicking on result (should open file and jump to symbol)
  - [ ] Test with common names (should show multiple results)
  - [ ] Test with unique names (should show single result)
  - [ ] Verify performance (should feel instant)
  - [ ] Test with large workspace (100+ files)

**Outcome**: Users can quickly search for symbols across the entire project using Ctrl+T. The search is fast and responsive, showing results from all workspace files.

**Estimated Effort**: 1-2 days

---

## Phase 9: Code Completion

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

## Phase 10: Signature Help

**Goal**: Show function signatures and parameter hints during function calls.

**Prerequisites**: Phase 4 complete (hover support provides foundation for signature formatting)

### Tasks (18)

- [ ] **10.1 Implement textDocument/signatureHelp request handler**
  - [ ] Create `internal/lsp/signature_help.go`
  - [ ] Define handler: `func SignatureHelp(context *glsp.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error)`
  - [ ] Extract document URI and position from params
  - [ ] Retrieve document from DocumentStore
  - [ ] Check if document and AST are available
  - [ ] Convert LSP position (UTF-16) to document position (UTF-8)
  - [ ] Call helper function to compute signature help
  - [ ] Return SignatureHelp response or nil if not in function call
  - [ ] Register handler in server initialization

- [ ] **10.2 Mark signatureHelpProvider in server capabilities**
  - [ ] In initialize handler, set `capabilities.SignatureHelpProvider` struct
  - [ ] Set `TriggerCharacters` to `["(", ","]` for auto-trigger
  - [ ] Optionally set `RetriggerCharacters` to handle edits
  - [ ] Verify capability is advertised to client
  - [ ] Test that VSCode triggers signature help on `(` and `,`

- [ ] **10.3 Determine call context from cursor position**
  - [ ] Create `internal/analysis/call_context.go`
  - [ ] Implement `DetermineCallContext(doc *Document, pos Position) (*CallContext, error)`
  - [ ] Analyze text around cursor to detect if inside function call
  - [ ] Return nil if cursor not inside parentheses
  - [ ] Extract call expression information (function name, parameter index)
  - [ ] Handle nested function calls (find innermost call)
  - [ ] Handle method calls and member function calls

- [ ] **10.4 Detect signature help triggers (opening parenthesis, comma)**
  - [ ] Check `params.Context.TriggerKind` for trigger type
  - [ ] Handle `SignatureHelpTriggerKind.Invoked` (manual Ctrl+Shift+Space)
  - [ ] Handle `SignatureHelpTriggerKind.TriggerCharacter`:
    - [ ] If `(` - start of function call
    - [ ] If `,` - moving to next parameter
  - [ ] Handle `SignatureHelpTriggerKind.ContentChange` (retrigger on typing)
  - [ ] Validate trigger character from `params.Context.TriggerCharacter`

- [ ] **10.5 Find function being called (identifier before opening parenthesis)**
  - [ ] Implement `FindFunctionAtCall(doc *Document, pos Position) (string, error)`
  - [ ] Scan backward from cursor to find opening parenthesis
  - [ ] Continue scanning backward to find function identifier
  - [ ] Handle qualified names (e.g., `object.method`)
  - [ ] Handle built-in function names
  - [ ] Return function name or error if not found
  - [ ] Log function name for debugging

- [ ] **10.6 Handle incomplete AST: temporarily insert closing parenthesis for parsing**
  - [ ] Create modified document text with `)` inserted at cursor
  - [ ] Parse modified text to get complete AST
  - [ ] Use this temporary AST for better call expression detection
  - [ ] Fallback to token-based analysis if parsing fails
  - [ ] Don't store the temporary AST (discard after analysis)
  - [ ] Test with incomplete function calls: `foo(x, `

- [ ] **10.7 Traverse tokens backward to identify function and count commas**
  - [ ] Implement `CountParameterIndex(text string, pos Position) (int, error)`
  - [ ] Scan backward from cursor position character-by-character
  - [ ] Count commas at the same parenthesis nesting level
  - [ ] Track parenthesis depth (nested calls)
  - [ ] Stop at opening parenthesis of current call
  - [ ] Return comma count as active parameter index (0-based)
  - [ ] Handle edge cases: empty parameter list, trailing comma

- [ ] **10.8 Retrieve function definition to get parameters and documentation**
  - [ ] Reuse symbol resolution from go-to-definition (Phase 5)
  - [ ] Call `ResolveSymbol(doc, functionName, pos)` to find definition
  - [ ] If found, get AST node for function declaration
  - [ ] Extract function signature from `*ast.FunctionDeclaration`
  - [ ] Get parameter names and types
  - [ ] Get return type
  - [ ] Extract documentation comments if available
  - [ ] Return nil if function not found (may be built-in)

- [ ] **10.9 Handle built-in functions with predefined signatures**
  - [ ] Create `internal/builtins/signatures.go`
  - [ ] Define signature information for built-in functions:
    - [ ] `PrintLn(text: String)`
    - [ ] `IntToStr(value: Integer): String`
    - [ ] `StrToInt(text: String): Integer`
    - [ ] `Length(str: String): Integer`
    - [ ] `Copy(str: String, index: Integer, count: Integer): String`
    - [ ] And other DWScript built-ins
  - [ ] Include parameter names and documentation
  - [ ] Check built-ins if user-defined function not found
  - [ ] Return predefined SignatureInformation

- [ ] **10.10 Construct SignatureHelp response with SignatureInformation**
  - [ ] Create `protocol.SignatureHelp` struct
  - [ ] Add one or more `protocol.SignatureInformation` to `Signatures` array
  - [ ] For each signature:
    - [ ] Set `Label` to formatted signature string
    - [ ] Set `Documentation` with function description (optional)
    - [ ] Set `Parameters` array with ParameterInformation for each param
  - [ ] Set `ActiveSignature` index (usually 0, see overloading)
  - [ ] Set `ActiveParameter` index based on comma count
  - [ ] Return SignatureHelp response

- [ ] **10.11 Format signature label (function name with parameters and return type)**
  - [ ] Implement `FormatSignature(funcDecl *ast.FunctionDeclaration) string`
  - [ ] Start with function name
  - [ ] Add opening parenthesis
  - [ ] For each parameter, add: `name: Type`
  - [ ] Separate parameters with `, `
  - [ ] Add closing parenthesis
  - [ ] If function (not procedure), add `: ReturnType`
  - [ ] Example: `function Calculate(x: Integer, y: Integer): Integer`
  - [ ] Format should match DWScript syntax

- [ ] **10.12 Provide ParameterInformation array for each parameter**
  - [ ] For each parameter in function signature:
    - [ ] Create `protocol.ParameterInformation` struct
    - [ ] Set `Label` to parameter substring in signature label (e.g., `x: Integer`)
    - [ ] OR set `Label` to [start, end] offsets in signature string
    - [ ] Set `Documentation` with parameter description (optional)
  - [ ] Add all parameters to `SignatureInformation.Parameters` array
  - [ ] Ensure parameter order matches declaration order
  - [ ] Test that VSCode highlights correct parameter

- [ ] **10.13 Determine activeParameter index by counting commas before cursor**
  - [ ] Use comma count from task 10.7
  - [ ] Active parameter is comma count (0-based index)
  - [ ] Examples:
    - [ ] `foo(|)` → activeParameter = 0
    - [ ] `foo(x|)` → activeParameter = 0
    - [ ] `foo(x, |)` → activeParameter = 1
    - [ ] `foo(x, y|)` → activeParameter = 1
  - [ ] Set `SignatureHelp.ActiveParameter` to computed index
  - [ ] Clamp to valid range if cursor beyond last parameter

- [ ] **10.14 Set activeSignature (0 unless function overloading supported)**
  - [ ] Set `SignatureHelp.ActiveSignature` to 0 by default
  - [ ] If multiple signatures exist (overloading):
    - [ ] Try to determine which signature user is calling
    - [ ] Match by parameter count (if known)
    - [ ] Match by parameter types (if available)
    - [ ] Select best matching signature
  - [ ] Update activeSignature to match selected signature
  - [ ] If cannot determine, keep as 0 (first signature)

- [ ] **10.15 Handle function overloading with multiple signatures**
  - [ ] Check if function has multiple declarations (overloads)
  - [ ] Collect all overloaded signatures
  - [ ] Create `SignatureInformation` for each overload
  - [ ] Add all to `SignatureHelp.Signatures` array
  - [ ] Order by parameter count (fewer parameters first)
  - [ ] Set activeSignature to best match (see 10.14)
  - [ ] Test with overloaded built-in functions
  - [ ] Verify VSCode shows all overloads with arrows to switch

- [ ] **10.16 Write unit tests for signature help with multi-parameter functions**
  - [ ] Create `internal/lsp/signature_help_test.go`
  - [ ] Test case: cursor after opening parenthesis
    - [ ] Code: `function foo(x: Integer, y: String); begin foo(|)`
    - [ ] Expected: activeParameter = 0, signature shown
  - [ ] Test case: cursor after first parameter
    - [ ] Code: `foo(5, |)`
    - [ ] Expected: activeParameter = 1
  - [ ] Test case: cursor in middle of parameter
    - [ ] Code: `foo(5|)`
    - [ ] Expected: activeParameter = 0
  - [ ] Test with 0-parameter, 1-parameter, and 5-parameter functions
  - [ ] Verify signature label formatting

- [ ] **10.17 Verify activeParameter highlighting at different cursor positions**
  - [ ] Test cursor positions within a call:
    - [ ] Before parameters: `foo(|x, y)`
    - [ ] After first param: `foo(x|, y)`
    - [ ] After comma: `foo(x,| y)`
    - [ ] After second param: `foo(x, y|)`
    - [ ] After all params: `foo(x, y|)`
  - [ ] For each position, verify correct activeParameter index
  - [ ] Test with nested calls: `foo(bar(|), baz())`
  - [ ] Verify innermost call is analyzed
  - [ ] Test with incomplete calls (missing closing paren)

- [ ] **10.18 Manually test signature help in VSCode during function calls**
  - [ ] Open DWScript file in VSCode
  - [ ] Type function name and `(` - verify signature popup appears
  - [ ] Verify first parameter is highlighted
  - [ ] Type parameter value and `,` - verify second parameter highlighted
  - [ ] Test with Ctrl+Shift+Space to manually trigger
  - [ ] Test with built-in functions (e.g., `PrintLn(`)
  - [ ] Test with custom functions
  - [ ] Test with overloaded functions (if supported)
  - [ ] Verify signature popup updates as you type
  - [ ] Verify popup dismisses after closing `)`

**Outcome**: When calling functions, users see parameter hints with the current parameter highlighted, making it easy to know what arguments to provide.

**Estimated Effort**: 1-2 days

---

## Phase 11: Rename Support

**Goal**: Enable symbol renaming across the codebase.

**Prerequisites**: Phase 6 complete (find references provides foundation for finding all rename locations)

### Tasks (14)

- [ ] **11.1 Implement textDocument/rename request handler**
  - [ ] Create `internal/lsp/rename.go`
  - [ ] Define handler: `func Rename(context *glsp.Context, params *protocol.RenameParams) (*protocol.WorkspaceEdit, error)`
  - [ ] Extract document URI, position, and new name from params
  - [ ] Retrieve document from DocumentStore
  - [ ] Check if document and AST are available
  - [ ] Convert LSP position to document position
  - [ ] Call helper function to compute rename edits
  - [ ] Return WorkspaceEdit or error if rename not possible
  - [ ] Register handler in server initialization

- [ ] **11.2 Mark renameProvider in server capabilities**
  - [ ] In initialize handler, set `capabilities.RenameProvider` to true or struct
  - [ ] Optionally set `PrepareProvider: true` to support textDocument/prepareRename
  - [ ] Verify capability is advertised to client
  - [ ] Test that VSCode enables rename (F2) on symbols

- [ ] **11.3 Identify symbol at rename position**
  - [ ] Reuse `FindNodeAtPosition` from hover/definition implementations
  - [ ] Get AST node at the rename position
  - [ ] Check if node is an identifier or declaration
  - [ ] Extract symbol name from node
  - [ ] Determine symbol kind (variable, function, class, etc.)
  - [ ] Return error if not on a renameable symbol
  - [ ] Log symbol for debugging

- [ ] **11.4 Validate that symbol can be renamed (reject keywords, built-ins)**
  - [ ] Create `CanRename(symbolName string, symbolKind SymbolKind) (bool, error)`
  - [ ] Reject DWScript keywords:
    - [ ] `begin`, `end`, `if`, `then`, `else`, `while`, `for`, `do`, etc.
  - [ ] Reject built-in type names:
    - [ ] `Integer`, `String`, `Float`, `Boolean`, `Variant`
  - [ ] Reject built-in function names:
    - [ ] `PrintLn`, `Length`, `Copy`, etc.
  - [ ] Return error message explaining why rename not allowed
  - [ ] Allow renaming of user-defined symbols only

- [ ] **11.5 Find all references of the symbol using references handler**
  - [ ] Reuse find references implementation from Phase 6
  - [ ] Call `FindReferences(doc, position, includeDeclaration: true)`
  - [ ] Get array of all reference locations (definition + uses)
  - [ ] Ensure all files in workspace are searched
  - [ ] Handle case where symbol not found (error)
  - [ ] Log number of references found
  - [ ] Sort locations by file URI for grouped edits

- [ ] **11.6 Prepare WorkspaceEdit with TextEdit for each reference**
  - [ ] Create `protocol.WorkspaceEdit` struct
  - [ ] Use `DocumentChanges` field (preferred over `Changes`)
  - [ ] For each reference location:
    - [ ] Create `protocol.TextEdit` struct
    - [ ] Set `Range` to the symbol's range at that location
    - [ ] Set `NewText` to the new symbol name from params
  - [ ] Group edits by document URI
  - [ ] Add all edits to WorkspaceEdit

- [ ] **11.7 Create TextEdit to replace old name with new name at each location**
  - [ ] For each reference:
    - [ ] Extract range covering the symbol identifier
    - [ ] Ensure range only covers the identifier (not surrounding whitespace)
    - [ ] Create TextEdit with Range and NewText
    - [ ] Validate that old text at range matches expected symbol name
  - [ ] Handle partial matches if symbol is qualified (e.g., `obj.field`)
  - [ ] Only replace the identifier part, not the qualifier

- [ ] **11.8 Group TextEdits by file in WorkspaceEdit.DocumentChanges**
  - [ ] Organize edits by document URI
  - [ ] For each document with changes:
    - [ ] Create `protocol.TextDocumentEdit` struct
    - [ ] Set `TextDocument` with URI and version
    - [ ] Set `Edits` array with all TextEdits for that document
  - [ ] Add all TextDocumentEdits to `WorkspaceEdit.DocumentChanges` array
  - [ ] Ensure edits are sorted by position (reverse order for application)
  - [ ] This allows atomic multi-file rename

- [ ] **11.9 Handle document version checking to avoid stale renames**
  - [ ] For each document being edited:
    - [ ] Get current version from DocumentStore
    - [ ] Set `TextDocumentIdentifier.Version` in TextDocumentEdit
  - [ ] Client will reject edit if version doesn't match
  - [ ] This prevents applying edits to stale document state
  - [ ] Handle version mismatch error gracefully
  - [ ] Log warning if versions don't match

- [ ] **11.10 Implement textDocument/prepareRename handler (optional)**
  - [ ] Create handler: `func PrepareRename(context *glsp.Context, params *protocol.PrepareRenameParams) (interface{}, error)`
  - [ ] Extract document URI and position
  - [ ] Identify symbol at position
  - [ ] Validate symbol can be renamed (call CanRename)
  - [ ] If renameable: return range and placeholder
  - [ ] If not renameable: return error with reason
  - [ ] This allows client to show error before user types new name
  - [ ] Register handler in server initialization

- [ ] **11.11 Return symbol range and placeholder text in prepareRename**
  - [ ] Create response with:
    - [ ] `Range`: the range of the symbol identifier
    - [ ] `Placeholder`: the current symbol name (pre-filled in rename dialog)
  - [ ] OR return `PrepareRenameResult` struct with range and placeholder
  - [ ] This provides better UX: rename dialog pre-filled with current name
  - [ ] Client can highlight the range before rename
  - [ ] Test in VSCode: F2 should highlight symbol and show dialog with current name

- [ ] **11.12 Write unit tests for variable/function rename**
  - [ ] Create `internal/lsp/rename_test.go`
  - [ ] Test case: rename local variable
    - [ ] Code: `var x: Integer; x := 5; PrintLn(x);`
    - [ ] Rename `x` to `y`
    - [ ] Verify 3 edits (declaration + 2 uses)
    - [ ] Verify all edits have correct ranges
  - [ ] Test case: rename function
    - [ ] Define function `foo`
    - [ ] Call `foo` multiple times
    - [ ] Rename to `bar`
    - [ ] Verify all calls updated
  - [ ] Verify WorkspaceEdit structure is correct

- [ ] **11.13 Write unit tests for rename across multiple files**
  - [ ] Create test workspace with 3 files:
    - [ ] File A defines function `GlobalFunc`
    - [ ] File B calls `GlobalFunc`
    - [ ] File C also calls `GlobalFunc`
  - [ ] Rename `GlobalFunc` to `RenamedFunc`
  - [ ] Verify WorkspaceEdit includes edits for all 3 files
  - [ ] Verify correct grouping in DocumentChanges
  - [ ] Test with classes and methods across files
  - [ ] Verify document versions are set

- [ ] **11.14 Write tests rejecting rename of keywords/built-ins**
  - [ ] Test renaming `begin` keyword - should return error
  - [ ] Test renaming `Integer` type - should return error
  - [ ] Test renaming `PrintLn` built-in - should return error
  - [ ] Verify error messages are clear and helpful
  - [ ] Test prepareRename returns error for non-renameable symbols
  - [ ] Verify client shows error dialog

- [ ] **11.15 Manually test rename operation in VSCode**
  - [ ] Open DWScript file in VSCode
  - [ ] Place cursor on local variable
  - [ ] Press F2 or right-click → Rename Symbol
  - [ ] Verify rename dialog appears with current name pre-filled
  - [ ] Type new name and press Enter
  - [ ] Verify all occurrences in file are updated
  - [ ] Test rename across multiple open files
  - [ ] Test rename in unopened files (should still work via workspace search)
  - [ ] Test rename with Ctrl+Z (undo should revert all files)
  - [ ] Verify rename preview (if supported by client)

**Outcome**: Users can rename symbols with F2, and all references across the workspace are updated automatically, with validation to prevent renaming keywords and built-ins.

**Estimated Effort**: 1-2 days

---

## Phase 12: Semantic Tokens

**Goal**: Provide semantic syntax highlighting information.

**Prerequisites**: Phase 2 complete (AST access with position metadata)

### Tasks (25)

- [ ] **12.1 Define SemanticTokensLegend with token types and modifiers**
  - [ ] Create `internal/lsp/semantic_tokens.go`
  - [ ] Define `SemanticTokensLegend` with `TokenTypes` array
  - [ ] Define `TokenModifiers` array
  - [ ] Store legend in Server struct for reuse
  - [ ] Legend must be consistent across all requests
  - [ ] Document token type and modifier indices
  - [ ] Register legend during server initialization

- [ ] **12.2 Include token types: keyword, string, number, comment, variable, parameter, property, function, class, interface, enum**
  - [ ] Define TokenTypes array with standard LSP types:
    - [ ] "namespace"
    - [ ] "type" (for classes, records)
    - [ ] "class"
    - [ ] "enum"
    - [ ] "interface"
    - [ ] "struct"
    - [ ] "typeParameter"
    - [ ] "parameter"
    - [ ] "variable"
    - [ ] "property"
    - [ ] "enumMember"
    - [ ] "function"
    - [ ] "method"
    - [ ] "keyword"
    - [ ] "string"
    - [ ] "number"
    - [ ] "comment"
  - [ ] Order matters: index is used in encoding

- [ ] **12.3 Include modifiers: static, deprecated, declaration, readonly**
  - [ ] Define TokenModifiers array:
    - [ ] "declaration" - for symbol definitions
    - [ ] "readonly" - for constants, readonly properties
    - [ ] "static" - for static/class methods and fields
    - [ ] "deprecated" - for deprecated symbols
    - [ ] "abstract" - for abstract classes/methods
    - [ ] "modification" - for assignments
    - [ ] "documentation" - for doc comments
  - [ ] Modifiers are bit flags (can combine multiple)
  - [ ] Document bit positions for encoding

- [ ] **12.4 Advertise SemanticTokensProvider in server capabilities**
  - [ ] In initialize handler, set `capabilities.SemanticTokensProvider`
  - [ ] Set `Legend` with token types and modifiers
  - [ ] Set `Full: true` to support full document tokenization
  - [ ] Optionally set `Range: true` for range requests (defer to later)
  - [ ] Optionally set `Full.Delta: true` for incremental updates (defer to later)
  - [ ] Verify capability advertised to client
  - [ ] Test that VSCode requests semantic tokens on document open

- [ ] **12.5 Implement textDocument/semanticTokens/full handler**
  - [ ] Define handler: `func SemanticTokensFull(context *glsp.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error)`
  - [ ] Extract document URI from params
  - [ ] Retrieve document from DocumentStore
  - [ ] Check if document has valid AST
  - [ ] Call helper to collect semantic tokens
  - [ ] Encode tokens in LSP delta format
  - [ ] Return SemanticTokens response with data array
  - [ ] Register handler in server initialization

- [ ] **12.6 Traverse document AST to collect semantic tokens**
  - [ ] Create `internal/analysis/semantic_tokens.go`
  - [ ] Implement `CollectSemanticTokens(ast *ast.Program, legend *Legend) ([]SemanticToken, error)`
  - [ ] Use `ast.Inspect()` to traverse all nodes
  - [ ] For each node, check type and classify token
  - [ ] Collect raw tokens with position, type, and modifiers
  - [ ] Return sorted array of tokens
  - [ ] Handle errors gracefully (skip nodes with missing positions)

- [ ] **12.7 Classify identifiers by role: variable, parameter, property, function, class, etc.**
  - [ ] For `*ast.Identifier` nodes:
    - [ ] Determine if it's a variable reference
    - [ ] Determine if it's a function call
    - [ ] Determine if it's a type reference
    - [ ] Determine if it's a property access
  - [ ] Use semantic analysis to resolve identifier role
  - [ ] Query symbol table for symbol kind
  - [ ] Map symbol kind to token type
  - [ ] Handle ambiguous cases (fallback to "variable")

- [ ] **12.8 Tag variable declarations with declaration modifier**
  - [ ] For `*ast.VariableDeclaration` nodes:
    - [ ] Classify identifier as "variable" type
    - [ ] Add "declaration" modifier
  - [ ] For `*ast.ConstantDeclaration` nodes:
    - [ ] Classify as "variable" type
    - [ ] Add "declaration" and "readonly" modifiers
  - [ ] For function parameters:
    - [ ] Classify as "parameter" type
    - [ ] Add "declaration" modifier
  - [ ] For field declarations:
    - [ ] Classify as "property" type
    - [ ] Add "declaration" modifier

- [ ] **12.9 Differentiate constants, enum members, and properties**
  - [ ] For constants:
    - [ ] Use "variable" type with "readonly" modifier
  - [ ] For enum members (if DWScript supports enums):
    - [ ] Use "enumMember" type
  - [ ] For class properties:
    - [ ] Use "property" type
    - [ ] Add "readonly" if property is read-only
  - [ ] For class fields:
    - [ ] Use "property" type (or "variable" if inside class)

- [ ] **12.10 Classify function and method names appropriately**
  - [ ] For `*ast.FunctionDeclaration` at global scope:
    - [ ] Use "function" type
    - [ ] Add "declaration" modifier
  - [ ] For `*ast.MethodDeclaration` in class:
    - [ ] Use "method" type
    - [ ] Add "declaration" modifier
    - [ ] Add "static" modifier if class method
  - [ ] For function calls (`*ast.CallExpression`):
    - [ ] Use "function" type (no declaration modifier)
  - [ ] For method calls:
    - [ ] Use "method" type

- [ ] **12.11 Classify class names, type identifiers, interface names**
  - [ ] For `*ast.ClassDeclaration`:
    - [ ] Use "class" type
    - [ ] Add "declaration" modifier
  - [ ] For `*ast.InterfaceDeclaration` (if supported):
    - [ ] Use "interface" type
    - [ ] Add "declaration" modifier
  - [ ] For `*ast.TypeDeclaration`:
    - [ ] Use "type" type (or "class"/"struct" depending on kind)
    - [ ] Add "declaration" modifier
  - [ ] For type references in variable declarations:
    - [ ] Use "class" or "type" type (no declaration modifier)

- [ ] **12.12 Tag literals (numbers, strings, booleans)**
  - [ ] For `*ast.IntegerLiteral` and `*ast.FloatLiteral`:
    - [ ] Use "number" type
    - [ ] No modifiers
  - [ ] For `*ast.StringLiteral`:
    - [ ] Use "string" type
    - [ ] No modifiers
  - [ ] For `*ast.BooleanLiteral`:
    - [ ] Use "keyword" type (true/false are keywords)
    - [ ] OR use "number" type (some languages treat booleans as numbers)
  - [ ] Note: Literals may be redundant with TextMate grammar (optional to include)

- [ ] **12.13 Optionally tag comments (may be redundant with TextMate grammar)**
  - [ ] If go-dws parser preserves comments:
    - [ ] Visit comment nodes
    - [ ] Use "comment" type
    - [ ] Optionally add "documentation" modifier for doc comments
  - [ ] If comments not in AST:
    - [ ] Skip (rely on TextMate grammar for comment highlighting)
    - [ ] This is acceptable and common practice

- [ ] **12.14 Ensure AST nodes have start/end position info for token ranges**
  - [ ] Verify all AST nodes have `Pos()` and `End()` methods (from Phase 2) ✅
  - [ ] Ensure positions are accurate (1-based in AST)
  - [ ] Convert positions to LSP format (0-based line and character)
  - [ ] Handle nodes with missing position info (skip them)
  - [ ] Test with sample programs to verify accuracy

- [ ] **12.15 Calculate token length from identifier name length**
  - [ ] For identifier nodes:
    - [ ] Get identifier name string
    - [ ] Length = `len(name)` in characters (not bytes!)
    - [ ] Handle UTF-8 multibyte characters correctly
  - [ ] For keyword nodes:
    - [ ] Length = keyword string length
  - [ ] For literals:
    - [ ] Length = literal string representation length
  - [ ] For operators:
    - [ ] Length = operator string length (e.g., `+` = 1, `<=` = 2)

- [ ] **12.16 Record [line, startChar, length, tokenType, tokenModifiers] for each token**
  - [ ] Define internal `SemanticToken` struct:
    - [ ] `Line int` (0-based)
    - [ ] `StartChar int` (0-based)
    - [ ] `Length int`
    - [ ] `TokenType int` (index in legend)
    - [ ] `TokenModifiers int` (bit flags)
  - [ ] For each AST node:
    - [ ] Extract position from `node.Pos()`
    - [ ] Convert to 0-based line and character
    - [ ] Calculate length
    - [ ] Determine token type index
    - [ ] Calculate modifier bit flags
    - [ ] Create SemanticToken and add to array

- [ ] **12.17 Sort tokens by position (required by LSP)**
  - [ ] After collecting all tokens, sort by:
    1. [ ] Line number (ascending)
    2. [ ] Start character (ascending) if same line
  - [ ] Use `sort.Slice()` with custom comparator
  - [ ] Verify no overlapping tokens (would indicate bug)
  - [ ] Verify no duplicate tokens at same position
  - [ ] LSP requires sorted tokens for delta encoding

- [ ] **12.18 Encode tokens in LSP relative format (delta encoding)**
  - [ ] Implement `EncodeSemanticTokens(tokens []SemanticToken) []uint32`
  - [ ] LSP format: flat array of uint32 values
  - [ ] Each token encoded as 5 values:
    1. [ ] Delta line (relative to previous token)
    2. [ ] Delta start char (relative to previous token if same line, else absolute)
    3. [ ] Length
    4. [ ] Token type index
    5. [ ] Token modifiers (bit flags)
  - [ ] First token is relative to (0, 0)
  - [ ] Example: `[0, 5, 3, 2, 0]` = token at line 0, char 5, length 3, type 2, no modifiers
  - [ ] Verify encoding with test cases

- [ ] **12.19 Return SemanticTokens response with encoded data**
  - [ ] Create `protocol.SemanticTokens` struct
  - [ ] Set `Data` field to encoded uint32 array
  - [ ] Optionally set `ResultId` for delta support (defer to later)
  - [ ] Return to client
  - [ ] Client decodes and applies semantic highlighting

- [ ] **12.20 Optionally implement textDocument/semanticTokens/full/delta for incremental updates**
  - [ ] Defer to later if too complex
  - [ ] Would require:
    - [ ] Storing previous token set
    - [ ] Computing diff between old and new tokens
    - [ ] Encoding edits in SemanticTokensEdit format
  - [ ] Benefit: performance improvement for large files with small changes
  - [ ] Not essential for initial implementation

- [ ] **12.21 Recompute semantic tokens after document changes**
  - [ ] When document changes (didChange event):
    - [ ] AST is reparsed
    - [ ] Client may request new semantic tokens
  - [ ] Don't proactively push semantic tokens
  - [ ] Wait for client request (pull model)
  - [ ] Ensure fresh AST is used for token computation
  - [ ] Test with rapid edits (no stale tokens)

- [ ] **12.22 Write unit tests for semantic token generation**
  - [ ] Create `internal/analysis/semantic_tokens_test.go`
  - [ ] Test case: simple variable declaration
    - [ ] Code: `var x: Integer;`
    - [ ] Expected tokens: `var` (keyword), `x` (variable, declaration), `Integer` (type)
  - [ ] Test case: function definition
    - [ ] Code: `function foo(): Integer; begin end;`
    - [ ] Expected tokens: `function` (keyword), `foo` (function, declaration), `Integer` (type), etc.
  - [ ] Test case: class with fields and methods
    - [ ] Verify class name, field names, method names classified correctly
  - [ ] Verify token positions and lengths
  - [ ] Verify modifiers applied correctly

- [ ] **12.23 Verify correct classification of various constructs (variables, functions, classes)**
  - [ ] Test with complex DWScript code samples
  - [ ] Include:
    - [ ] Global variables, local variables, parameters
    - [ ] Functions, procedures, methods
    - [ ] Classes, records, interfaces
    - [ ] Properties, fields
    - [ ] Constants, enum members
  - [ ] For each construct, verify token type and modifiers
  - [ ] Compare with expected semantic highlighting
  - [ ] Use snapshot testing for regression detection

- [ ] **12.24 Configure VSCode extension with semantic token legend**
  - [ ] If using VSCode extension for testing:
    - [ ] Ensure extension doesn't define its own semantic tokens
    - [ ] Let LSP server provide semantic tokens
    - [ ] Configure semantic token scopes mapping (optional)
  - [ ] Test that semantic highlighting appears in VSCode
  - [ ] Compare with TextMate grammar highlighting
  - [ ] Semantic tokens should enhance or override TextMate

- [ ] **12.25 Manually test semantic highlighting in VSCode**
  - [ ] Open DWScript file in VSCode with LSP active
  - [ ] Verify semantic highlighting appears
  - [ ] Variables should be colored distinctly from functions
  - [ ] Parameters should be visually distinct from variables
  - [ ] Keywords should be highlighted (may use TextMate)
  - [ ] Test with different color themes
  - [ ] Test that highlighting updates after edits
  - [ ] Compare with and without semantic tokens (toggle in VSCode settings)
  - [ ] Verify performance (no lag when opening files)

**Outcome**: Enhanced syntax highlighting based on semantic understanding, with variables, functions, parameters, and types colored appropriately based on their roles in the code.

**Estimated Effort**: 2-3 days

---

## Phase 13: Code Actions

**Goal**: Provide quick fixes and refactoring actions.

**Prerequisites**: Phase 3 complete (diagnostics provide context for quick fixes), Phase 6 complete (find references needed for some refactorings)

### Tasks (23)

- [ ] **13.1 Implement textDocument/codeAction request handler**
  - [ ] Create `internal/lsp/code_action.go`
  - [ ] Define handler: `func CodeAction(context *glsp.Context, params *protocol.CodeActionParams) ([]protocol.CodeAction, error)`
  - [ ] Extract document URI, range, and context from params
  - [ ] Get diagnostics from params.Context.Diagnostics
  - [ ] Retrieve document from DocumentStore
  - [ ] Check if document has AST available
  - [ ] Call helper functions to generate code actions
  - [ ] Return array of CodeAction (may be empty)
  - [ ] Register handler in server initialization

- [ ] **13.2 Mark codeActionProvider in server capabilities**
  - [ ] In initialize handler, set `capabilities.CodeActionProvider`
  - [ ] Use struct with `CodeActionKinds` field
  - [ ] Specify supported kinds (see next task)
  - [ ] Optionally set `ResolveProvider: true` for lazy resolution (defer to later)
  - [ ] Verify capability advertised to client
  - [ ] Test that VSCode shows lightbulb icon when code actions available

- [ ] **13.3 Specify supported codeActionKinds (quickfix, refactor, etc.)**
  - [ ] Define supported kinds:
    - [ ] `CodeActionKind.QuickFix` - for fixing diagnostics
    - [ ] `CodeActionKind.Refactor` - for refactoring actions
    - [ ] `CodeActionKind.RefactorExtract` - extract to function/variable
    - [ ] `CodeActionKind.RefactorInline` - inline variable/function
    - [ ] `CodeActionKind.Source` - source actions (organize imports)
    - [ ] `CodeActionKind.SourceOrganizeImports` - organize imports/units
  - [ ] Set `CodeActionProvider.CodeActionKinds` to array of supported kinds
  - [ ] Client may request specific kinds

- [ ] **13.4 Implement quick fix for 'Undeclared identifier' error**
  - [ ] Create `GenerateQuickFixes(diagnostic protocol.Diagnostic, doc *Document) ([]CodeAction, error)`
  - [ ] Check if diagnostic code or message indicates "undeclared identifier"
  - [ ] Extract identifier name from diagnostic message
  - [ ] Suggest code actions:
    1. [ ] Declare variable with inferred type
    2. [ ] Declare function (if identifier used as call)
  - [ ] Create CodeAction for each suggestion
  - [ ] Set `Kind` to `CodeActionKind.QuickFix`
  - [ ] Set `Title` to clear description (e.g., "Declare variable 'x'")

- [ ] **13.5 Suggest 'Declare variable X' action with default type**
  - [ ] For undeclared identifier quick fix:
  - [ ] Create CodeAction with title: `"Declare variable 'x'"`
  - [ ] Infer type if possible from usage context:
    - [ ] If assigned from integer literal: `Integer`
    - [ ] If assigned from string literal: `String`
    - [ ] Default: `Variant` if cannot infer
  - [ ] Generate declaration text: `var x: Integer;`
  - [ ] Create WorkspaceEdit with TextEdit to insert declaration
  - [ ] Attach diagnostic as `Diagnostics` field
  - [ ] Add to code actions array

- [ ] **13.6 Insert var declaration at appropriate location (function top or global)**
  - [ ] Determine insertion location based on context:
    - [ ] If inside function: insert at start of function body (after `begin`)
    - [ ] If at global scope: insert at start of file or after existing `var` block
  - [ ] Calculate insertion position (line, character)
  - [ ] Create TextEdit with:
    - [ ] Range: zero-length at insertion point
    - [ ] NewText: declaration with appropriate indentation and newline
  - [ ] Example: `\n  var x: Integer;\n`
  - [ ] Handle existing var blocks (append to block vs create new)

- [ ] **13.7 Implement quick fix for 'Missing semicolon' error**
  - [ ] Check diagnostic for "missing semicolon" or "expected ';'"
  - [ ] Extract position where semicolon expected (from diagnostic range)
  - [ ] Create CodeAction with title: `"Insert missing semicolon"`
  - [ ] Create WorkspaceEdit with TextEdit:
    - [ ] Range: zero-length at expected position
    - [ ] NewText: `;`
  - [ ] Set Kind to QuickFix
  - [ ] Attach diagnostic
  - [ ] Test that applying action resolves the diagnostic

- [ ] **13.8 Implement quick fix for unused variable warning**
  - [ ] Check diagnostic for "unused variable" or code `W_UNUSED_VAR`
  - [ ] Extract variable name from diagnostic message
  - [ ] Suggest two code actions:
    1. [ ] Remove the variable declaration
    2. [ ] Prefix variable name with underscore (convention for intentionally unused)
  - [ ] Create CodeAction for each suggestion
  - [ ] Set Kind to QuickFix

- [ ] **13.9 Suggest removing or prefixing with underscore**
  - [ ] Code action 1: "Remove unused variable 'x'"
    - [ ] Find variable declaration in AST
    - [ ] Create TextEdit to delete entire declaration statement
    - [ ] Handle formatting (remove blank line if appropriate)
  - [ ] Code action 2: "Rename to '_x'"
    - [ ] Reuse rename functionality from Phase 11
    - [ ] Create WorkspaceEdit to rename `x` to `_x`
    - [ ] This preserves declaration but indicates intentional non-use
  - [ ] Add both actions to result array

- [ ] **13.10 Implement refactoring: Organize uses/imports**
  - [ ] Create source action: "Organize units"
  - [ ] Set Kind to `CodeActionKind.SourceOrganizeImports`
  - [ ] Analyze current uses/imports clause
  - [ ] Remove unused unit references:
    - [ ] Parse unit names in uses clause
    - [ ] Check if any symbols from each unit are used
    - [ ] Remove units with no references
  - [ ] Add missing unit references:
    - [ ] Find undefined identifiers
    - [ ] Search workspace index for definitions
    - [ ] Determine which unit contains definition
    - [ ] Add unit to uses clause
  - [ ] Sort units alphabetically (optional)
  - [ ] Create WorkspaceEdit to replace uses clause

- [ ] **13.11 Remove unused unit references from uses clause**
  - [ ] Implement `FindUnusedUnits(doc *Document) ([]string, error)`
  - [ ] Extract list of imported units from AST
  - [ ] For each unit:
    - [ ] Get list of exported symbols from unit (parse unit file)
    - [ ] Search document for references to any of those symbols
    - [ ] If no references found, mark unit as unused
  - [ ] Return list of unused unit names
  - [ ] Create edit to remove unused units from uses clause

- [ ] **13.12 Add missing unit references for used symbols**
  - [ ] Implement `FindMissingUnits(doc *Document) (map[string]string, error)`
  - [ ] Collect all undefined identifier errors
  - [ ] For each undefined identifier:
    - [ ] Search workspace index for definition
    - [ ] Determine which file/unit defines the symbol
    - [ ] Extract unit name from file
    - [ ] Check if unit already in uses clause
    - [ ] If not, add to missing units map
  - [ ] Return map of symbol → unit name
  - [ ] Create edit to add missing units to uses clause

- [ ] **13.13 Consider extract to function refactoring (complex, optional)**
  - [ ] Defer to future if too complex for initial implementation
  - [ ] Would require:
    - [ ] Detecting selected code range
    - [ ] Analyzing variable usage (parameters and returns)
    - [ ] Generating function signature
    - [ ] Creating function declaration
    - [ ] Replacing selection with function call
  - [ ] This is a valuable refactoring but not essential for MVP
  - [ ] Document as future enhancement

- [ ] **13.14 Consider implement interface/abstract methods (complex, optional)**
  - [ ] Defer to future if too complex
  - [ ] Would require:
    - [ ] Detecting class implementing interface
    - [ ] Finding missing method implementations
    - [ ] Generating method stubs
    - [ ] Inserting into class body
  - [ ] Useful for DWScript if interfaces are supported
  - [ ] Document as future enhancement

- [ ] **13.15 Recognize diagnostic patterns using error codes or message matching**
  - [ ] Implement `MatchDiagnosticPattern(diagnostic protocol.Diagnostic) (PatternType, map[string]string)`
  - [ ] Match by error code if available (preferred):
    - [ ] `E_UNDEFINED_VAR` → Undeclared identifier
    - [ ] `E_MISSING_SEMICOLON` → Missing semicolon
    - [ ] `W_UNUSED_VAR` → Unused variable
  - [ ] Fallback to regex matching on message:
    - [ ] `"undeclared identifier '(.*?)'"` → extract identifier name
    - [ ] `"unused variable '(.*?)'"` → extract variable name
  - [ ] Return pattern type and extracted variables
  - [ ] Use this to route to appropriate quick fix generator

- [ ] **13.16 Create CodeAction with appropriate kind (quickfix, refactor)**
  - [ ] For each code action, create `protocol.CodeAction` struct
  - [ ] Set required fields:
    - [ ] `Title` - clear, concise description (e.g., "Declare variable 'x'")
    - [ ] `Kind` - appropriate CodeActionKind
  - [ ] Set optional fields:
    - [ ] `Diagnostics` - diagnostics this action fixes
    - [ ] `Edit` - WorkspaceEdit to apply
    - [ ] `Command` - alternative to edit (for complex actions)
    - [ ] `IsPreferred` - true if this is the recommended action
  - [ ] Add to result array

- [ ] **13.17 Attach diagnostic as associatedDiagnostic in code action**
  - [ ] For quick fixes, set `CodeAction.Diagnostics` array
  - [ ] Include the diagnostic that triggered the fix
  - [ ] Client uses this to:
    - [ ] Show fix in context of diagnostic
    - [ ] Auto-apply if "auto fix" enabled
    - [ ] Track which diagnostics are addressed
  - [ ] For refactorings (not tied to diagnostic), leave empty

- [ ] **13.18 Provide WorkspaceEdit with changes to resolve issue**
  - [ ] Create `protocol.WorkspaceEdit` for each code action
  - [ ] Use `DocumentChanges` for document edits
  - [ ] Create `TextDocumentEdit` with:
    - [ ] TextDocument identifier (URI + version)
    - [ ] Array of TextEdits
  - [ ] For multi-file refactorings, include edits for all files
  - [ ] Verify edit correctness before returning
  - [ ] Test that applying edit produces expected result

- [ ] **13.19 Ensure code action titles clearly describe the fix**
  - [ ] Follow clear naming conventions:
    - [ ] Quick fixes: Start with verb (Insert, Remove, Declare, Fix, etc.)
    - [ ] Refactorings: Use clear names (Extract to function, Rename to...)
    - [ ] Include relevant identifiers in title
  - [ ] Examples:
    - [ ] ✓ "Declare variable 'x'"
    - [ ] ✗ "Fix this"
    - [ ] ✓ "Remove unused variable 'temp'"
    - [ ] ✗ "Remove"
  - [ ] Test titles in VSCode (should be immediately understandable)

- [ ] **13.20 Write unit tests for quick fix actions**
  - [ ] Create `internal/lsp/code_action_test.go`
  - [ ] Test case: undeclared identifier quick fix
    - [ ] Code with error: `x := 5;` (x not declared)
    - [ ] Request code actions at error position
    - [ ] Verify "Declare variable 'x'" action returned
    - [ ] Verify WorkspaceEdit inserts correct declaration
  - [ ] Test case: missing semicolon quick fix
  - [ ] Test case: unused variable quick fix
  - [ ] Verify action kinds are correct
  - [ ] Verify diagnostics attached

- [ ] **13.21 Verify applying edit resolves the diagnostic**
  - [ ] For each quick fix test:
    - [ ] Apply the WorkspaceEdit to document
    - [ ] Reparse document to get new diagnostics
    - [ ] Verify the original diagnostic is gone
    - [ ] Verify no new errors introduced
  - [ ] Test with real DWScript code samples
  - [ ] Ensure fixes are syntactically correct
  - [ ] Test edge cases (fix at start/end of file, empty file)

- [ ] **13.22 Write unit tests for organize units refactoring**
  - [ ] Test case: remove unused unit
    - [ ] Code: `uses UnitA, UnitB;` but only UnitA used
    - [ ] Action: "Organize units"
    - [ ] Result: `uses UnitA;`
  - [ ] Test case: add missing unit
    - [ ] Code uses undefined symbol from known unit
    - [ ] Action: "Organize units"
    - [ ] Result: unit added to uses clause
  - [ ] Test with multiple files
  - [ ] Verify units remain functional after organizing

- [ ] **13.23 Manually test code actions in VSCode**
  - [ ] Open DWScript file with errors in VSCode
  - [ ] Place cursor on error (red squiggle)
  - [ ] Verify lightbulb icon appears in gutter
  - [ ] Click lightbulb or press Ctrl+. (Cmd+. on Mac)
  - [ ] Verify code actions menu appears with fix suggestions
  - [ ] Select a fix and verify it applies correctly
  - [ ] Test "Quick Fix" command (Ctrl+.)
  - [ ] Test "Source Action" command for organize units
  - [ ] Verify Ctrl+Z undoes the action
  - [ ] Test performance (actions appear instantly)

**Outcome**: Users receive contextual quick fixes and refactoring suggestions via the lightbulb menu, with fixes for common errors like undeclared identifiers and unused variables, plus refactoring actions like organizing units.

**Estimated Effort**: 2-3 days

---

## Phase 14: Testing, Quality, and Finalization

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
