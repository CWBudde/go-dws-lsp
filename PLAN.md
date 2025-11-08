# go-dws-lsp Implementation Plan

This document tracks the implementation status of the go-dws Language Server Protocol (LSP) server.

---

## Phase 1: Document Synchronization ✅

**Achievements**: Full document lifecycle handlers (didOpen, didClose, didChange) with incremental sync support. UTF-16/UTF-8 position conversion. Configuration change handling. Version tracking and diagnostics integration. 27 tests passing.

---

## Phase 2: go-dws API Enhancements for LSP Integration ✅

**Achievements**: go-dws library enhanced with structured errors (line/column/severity), AST position metadata on all 64+ node types, public AST API with visitor pattern, symbol table access, type information queries, and parse-only mode. Released as v0.3.1. Full test coverage and performance benchmarks.

---

## Phase 3: Diagnostics ✅

**Achievements**: textDocument/publishDiagnostics handler with syntax and semantic error reporting. Real-time diagnostics on document open/save/change. Error classification with severity levels. Integration with go-dws structured errors. Comprehensive test coverage including multi-error scenarios.

---

## Phase 4: Hover Support ✅

**Achievements**: textDocument/hover handler with type information, documentation, and symbol details. Hover on variables, functions, parameters, class members, built-in types. Markdown formatting with syntax highlighting. Position-aware AST traversal. 17 tests passing covering all DWScript constructs.

---

## Phase 5: Go-to Definition ✅

**Achievements**: textDocument/definition handler navigating to symbol declarations. Support for local/global variables, functions, classes, parameters across files. Workspace symbol indexing for fast lookups. URI conversion and cross-file navigation. 15 comprehensive tests including multi-file scenarios.

---

## Phase 6: Find References ✅

**Achievements**: textDocument/references handler finding all symbol usages. Workspace-wide reference search with includeDeclaration support. Symbol index for non-open files. Deduplication and location merging. Multi-file reference tests. Performance optimized with concurrent workspace scanning.

---

## Phase 7: Document Symbols ✅

**Achievements**: textDocument/documentSymbol handler with hierarchical symbol tree. Support for functions, classes, variables, constants, enums, records. Proper parent-child relationships and symbol kinds. Selection/detail ranges. 13 tests covering nested structures and edge cases.

---

## Phase 8: Workspace Symbols ✅

**Achievements**: workspace/symbol handler for global symbol search. Query filtering with fuzzy/substring/prefix matching. Background workspace indexing with file watching. Symbol deduplication across workspace. Performance optimized for large workspaces (100ms target). 8 comprehensive tests.

---

## Phase 9: Code Completion ✅

**Achievements**: textDocument/completion handler with context-aware suggestions. Keyword completions with snippets. Global/local symbols, built-in types/functions. Member access (dot notation) for classes/records. Function/method parameters. Import/unit name completion. CompletionItem details and documentation. Caching for performance. 33 tests passing covering all completion scenarios.

---

## Phase 10: Signature Help ✅

**Achievements**: textDocument/signatureHelp handler with function/method signatures. Active parameter highlighting based on cursor position. Parameter documentation and type information. Support for overloaded methods and built-in functions. Proper handling of nested calls and complex expressions. 11 tests including edge cases.

---

## Phase 11: Rename Support ✅

**Achievements**: textDocument/rename and textDocument/prepareRename handlers. Workspace-wide symbol renaming with validation. Multi-file edits via WorkspaceEdit. Rejection of keyword/built-in renames. Version tracking for document changes. 14 tests including cross-file renames and validation scenarios.

---

## Phase 12: Semantic Tokens ✅

**Achievements**: textDocument/semanticTokens/full and /range handlers. Semantic token legend with 11 token types and 5 modifiers. Classification of variables, functions, classes, parameters, properties, keywords. Declaration and readonly modifiers. Delta support with resultId. Encoding to LSP integer array format. 15 tests covering all construct types and edge cases.

---

## Phase 13: Code Actions ✅

**Achievements**: textDocument/codeAction handler with QuickFix and Source actions. "Declare variable" quick fix with type inference. "Declare function" quick fix with parameter type inference from call site. "Remove unused variable" and "Prefix with underscore" actions. "Organize units" source action (add missing, remove unused, sort alphabetically). Smart insertion locations. 30+ tests covering all action types.

---

## Phase 14: Testing, Quality, and Finalization

**Goal**: Ensure robustness, performance, and code quality before release.

### Tasks (20)

- [x] **14.1 Collect all deferred or yet open tasks from this roadmap**
  - [x] Review all previous phases
  - [x] Identify any incomplete or deferred tasks
  - [x] Create a consolidated list of open tasks
  - [x] Prioritize tasks based on impact and effort
  - [x] Check, if - by any chance - the task is already done (maybe we forgot to mark it as done)
  - [x] Collect tasks to Phase 14 "Remaining Tasks" list (based on impact and priority, renumber!)
  - [x] Mark manual test tasks as done (do not consider these)
  - [x] Mark tasks, which have been listed under Phase 14 as done in the original Phase
    - Add a note that it is deferred to Phase 14.
  - **Results**:
    - Updated go-dws to latest commit (v0.3.1-0.20251108175356-790bc43db6be)
    - Found 2 genuinely incomplete tasks, added to Phase 14
    - Marked 3 tasks as complete that were already implemented
    - All manual testing tasks properly ignored
    - Phase 14 now contains all remaining work with priorities

- [ ] **14.2 Run comprehensive integration tests against real DWScript projects**
  - [ ] Identify or create sample DWScript projects for testing:
    - [ ] Small project (single file, ~100 LOC)
    - [ ] Medium project (5-10 files, ~1000 LOC)
    - [ ] Large project (50+ files, ~10000 LOC)
  - [x] Create `test/integration/` directory - DONE
  - [x] Write integration test suite framework - DONE
    - [x] Setup helper functions
    - [x] Server lifecycle management
  - [ ] Integration tests for all LSP operations (26 tests exist, 7 LSP features missing):
    - [x] Initialize/Shutdown workflow (2 tests) - lsp_integration_test.go
    - [x] Document lifecycle (didOpen/didChange/didClose) (4 tests) - lsp_integration_test.go
    - [x] Diagnostics (1 test) - lsp_integration_test.go
    - [x] Hover (6 tests) - hover_integration_test.go
    - [x] Go to Definition (7 tests) - definition_integration_test.go
    - [x] Signature Help (7 tests) - signature_help_integration_test.go
    - [ ] Find References (0 tests) - MISSING
    - [ ] Document Symbols (0 tests) - MISSING
    - [ ] Workspace Symbols (0 tests) - MISSING
    - [ ] Code Completion (0 tests) - MISSING
    - [ ] Rename (0 tests) - MISSING
    - [ ] Semantic Tokens (0 tests) - MISSING
    - [ ] Code Actions (0 tests) - MISSING
  - [x] Run tests with `-race` flag - Available via `go test -race -tags=integration ./test/integration/...`
  - [ ] Document any issues found and verify fixes

- [ ] **14.3 Test all features together in VSCode with sample projects**
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

- [ ] **14.4 Verify no feature breaks another (document sync during go-to-def, etc.)**
  - [ ] Test document synchronization while:
    - [ ] Hover requests are in-flight
    - [ ] Completion is triggered
    - [ ] Find references is running
  - [ ] Test concurrent operations:
    - [ ] Edit in File A while finding references in File B
    - [ ] Trigger completion while diagnostics computing
  - [ ] Verify state consistency after each operation
  - [ ] Check for deadlocks or race conditions

- [ ] **14.5 Performance testing: ensure no IDE freezing during operations**
  - [ ] Use `pprof` to profile CPU usage
  - [ ] Test performance of critical operations:
    - [ ] Document open (target: <100ms)
    - [ ] Completion (target: <50ms)
    - [ ] Hover (target: <50ms)
    - [ ] Find references (target: <500ms for 1000 files)
    - [ ] Workspace symbol search (target: <200ms)
  - [ ] Identify bottlenecks and optimize
  - [ ] Test with large files (>5000 LOC)
  - [ ] Verify memory usage stays reasonable

- [ ] **14.6 Memory profiling to detect leaks**
  - [ ] Use `pprof` to profile memory allocations
  - [ ] Test long-running sessions (1+ hours)
  - [ ] Monitor for growing heap size
  - [ ] Check for goroutine leaks
  - [ ] Verify caches don't grow unbounded
  - [ ] Add memory limits/eviction policies if needed

- [ ] **14.7 Stress test: open 100+ files rapidly**
  - [ ] Create test with 100+ DWScript files
  - [ ] Rapidly open/close files
  - [ ] Verify server remains responsive
  - [ ] Check for resource exhaustion
  - [ ] Monitor CPU and memory usage
  - [ ] Verify no file handle leaks

- [ ] **14.8 Benchmark critical paths (completion, hover, etc.)**
  - [ ] Create benchmark suite in `internal/lsp/benchmark_test.go`
  - [ ] Benchmark each LSP operation:
    - [ ] `BenchmarkCompletion`
    - [ ] `BenchmarkHover`
    - [ ] `BenchmarkDefinition`
    - [ ] `BenchmarkReferences`
    - [ ] `BenchmarkDiagnostics`
  - [ ] Run benchmarks with various file sizes
  - [ ] Document baseline performance
  - [ ] Set performance regression tests

- [ ] **14.9 Add logging at appropriate levels (debug, info, error)**
  - [ ] Review all log statements for appropriate levels
  - [ ] Add debug logging for key decision points
  - [ ] Add info logging for major operations
  - [ ] Ensure errors are logged with context
  - [ ] Add request/response IDs for tracing
  - [ ] Test that log levels can be changed dynamically

- [ ] **14.10 Ensure error messages are helpful to users**
  - [ ] Review all error messages
  - [ ] Add context to error messages
  - [ ] Suggest fixes where possible
  - [ ] Avoid internal/technical jargon
  - [ ] Test error messages in VSCode
  - [ ] Document common errors in README

- [ ] **14.11 Handle edge cases gracefully (empty files, huge files, invalid syntax)**
  - [ ] Test with empty files
  - [ ] Test with files >10,000 LOC
  - [ ] Test with binary files
  - [ ] Test with invalid UTF-8
  - [ ] Test with deeply nested structures
  - [ ] Verify graceful degradation

- [ ] **14.12 Add timeout handling for long operations**
  - [ ] Identify operations that could timeout:
    - [ ] Workspace indexing
    - [ ] Find references
    - [ ] Rename across many files
  - [ ] Add context.WithTimeout to long operations
  - [ ] Return partial results when timeout occurs
  - [ ] Show progress notifications for long operations
  - [ ] Allow cancellation of in-progress operations

- [ ] **14.13 Test with malformed LSP requests**
  - [ ] Send requests with missing required fields
  - [ ] Send requests with wrong types
  - [ ] Send requests out of sequence (e.g., hover before initialize)
  - [ ] Verify server doesn't crash
  - [ ] Verify appropriate error responses
  - [ ] Test recovery after malformed requests

- [ ] **14.14 Verify thread safety with concurrent requests**
  - [ ] Run tests with `-race` flag
  - [ ] Send concurrent requests from multiple clients
  - [ ] Verify no data races
  - [ ] Verify no deadlocks
  - [ ] Test concurrent document edits
  - [ ] Stress test with 100+ concurrent operations

- [ ] **14.15 Document known limitations and future work**
  - [ ] Create LIMITATIONS.md
  - [ ] Document features not yet implemented
  - [ ] Document known bugs/issues
  - [ ] List DWScript features with partial support
  - [ ] Note performance limitations
  - [ ] Link to GitHub issues for tracking

- [ ] **14.16 Write user-facing README with setup instructions**
  - [ ] Installation instructions
  - [ ] VSCode configuration guide
  - [ ] Building from source
  - [ ] Command-line options
  - [ ] Troubleshooting section
  - [ ] Feature list with examples
  - [ ] Contributing guidelines

- [ ] **14.17 Create CONTRIBUTING.md for future contributors**
  - [ ] Code structure overview
  - [ ] Development setup
  - [ ] Running tests
  - [ ] Code style guidelines
  - [ ] Pull request process
  - [ ] Issue reporting guidelines

- [ ] **14.18 Add architecture documentation in docs/**
  - [ ] Architecture overview diagram
  - [ ] Component descriptions
  - [ ] Request flow diagrams
  - [ ] State management explanation
  - [ ] Extension points for new features
  - [ ] Performance considerations

- [ ] **14.19 Prepare release checklist**
  - [ ] All tests passing
  - [ ] No known critical bugs
  - [ ] Documentation complete
  - [ ] Performance benchmarks meet targets
  - [ ] Memory leaks addressed
  - [ ] CHANGELOG.md up to date
  - [ ] Version numbers updated
  - [ ] Release notes drafted

- [ ] **14.20 Tag v1.0.0 release**
  - [ ] Complete all tasks above
  - [ ] Final code review
  - [ ] Create git tag
  - [ ] Push to GitHub
  - [ ] Create GitHub release with notes
  - [ ] Announce on relevant channels

**Outcome**: Production-ready LSP server with comprehensive test coverage, documentation, and performance validation.

**Estimated Effort**: 1-2 weeks

---

## Phase 15: Code Quality and Linting Fixes

**Goal**: Address all golangci-lint issues to improve code quality, maintainability, and performance.

### Categories

- [ ] **15.1 Error Handling (err113 - 26 issues)**
  - [ ] Replace dynamic errors with wrapped static errors in identifier_scanner.go (2 issues)
  - [ ] Replace dynamic errors with wrapped static errors in path_utils.go (2 issues)
  - [ ] Replace dynamic errors with wrapped static errors in text_edit.go (7 issues)
  - [ ] Replace dynamic errors with wrapped static errors in rename.go (15 issues)

- [ ] **15.2 Duplicate Code (dupl - 22 issues)**
  - [ ] Remove duplicate test code in text_document_test.go (lines 294-351 duplicate of 234-292)
  - [ ] Address remaining 21 duplicate code blocks across the codebase

- [ ] **15.3 Duplicate Words (dupword - 2 issues)**
  - [ ] Fix duplicate "class" in semantic_tokens_test.go
  - [ ] Fix duplicate "end;" in document_symbol_test.go

- [ ] **15.4 Forbidden Functions (forbidigo - 1 issue)**
  - [ ] Replace fmt.Printf with proper logging in main.go

- [ ] **15.5 Function Ordering (funcorder - 13 issues)**
  - [ ] Reorder unexported methods in symbol_resolver.go to appear after exported methods

- [ ] **15.6 Function Length (funlen - 24 issues)**
  - [ ] Break down main function in main.go (77 lines > 60)
  - [ ] Break down TestPositionInRange in ast_node_finder_test.go (88 lines > 60)
  - [ ] Break down TestGetSymbolName in ast_node_finder_test.go (64 lines > 60)
  - [ ] Break down TestFindNodeAtPosition in ast_node_finder_test.go (93 lines > 60)
  - [ ] Break down TestCountParameterIndex in call_context_test.go (90 lines > 60)
  - [ ] Break down TestFindFunctionAtCall in call_context_test.go (68 lines > 60)
  - [ ] Break down TestDetermineCallContextWithTempAST in call_context_test.go (91 lines > 60)
  - [ ] Break down TestParseDocument_SyntaxErrors in parse_test.go (63 lines > 60)
  - [ ] Break down TestConvertStructuredErrors in parse_test.go (67 lines > 60)
  - [ ] Break down getKeywordCompletions in scope_completion.go (88 lines > 60)
  - [ ] Break down getBuiltInCompletions in scope_completion.go (86 lines > 60)
  - [ ] Break down FindSemanticReferences in semantic_references.go (68 lines > 60)
  - [ ] Break down GetFunctionSignatures in signature_info.go (64 lines > 60)
  - [ ] Break down TestExtractCallArguments in code_action_test.go (63 lines > 60)
  - [ ] Break down TestCompletion_TriggerCharacterDot in completion_test.go (66 lines > 60)
  - [ ] Break down TestCompletion_MemberAccessOnRecord in completion_test.go (83 lines > 60)
  - [ ] Break down TestFindIdentifierDefinition_MultipleScopes in definition_test.go (76 lines > 60)
  - [ ] Break down createRecordSymbol in document_symbol.go (88 lines > 60)
  - [ ] Break down createEnumSymbol in document_symbol.go (64 lines > 60)
  - [ ] Break down TestGetFunctionHover in hover_test.go (68 lines > 60)
  - [ ] Break down TestSortLocationsByFileAndPosition in references_test.go (106 lines > 60)
  - [ ] Break down TestGlobalReferences_AcrossMultipleFunctions in references_test.go (65 lines > 60)
  - [ ] Break down TestGlobalReferences_VerifySorting in references_test.go (74 lines > 60)
  - [ ] Break down TestCanRenameSymbol_Keywords in rename_test.go (87 lines > 60)

- [ ] **15.7 Global Variables (gochecknoglobals - 5 issues)**
  - [ ] Address builtinSignatures global variable in builtins/signatures.go
  - [ ] Address serverInstance global variable in lsp/initialize.go
  - [ ] Address dwscriptKeywords global variable in lsp/rename.go
  - [ ] Address builtInTypes global variable in lsp/rename.go
  - [ ] Address builtInFunctions global variable in lsp/rename.go

- [ ] **15.8 Cognitive Complexity (gocognit - 15 issues)**
  - [ ] Reduce complexity of findParameterIndex in call_context.go (38 > 30)
  - [ ] Reduce complexity of FindFunctionAtCall in call_context.go (38 > 30)
  - [ ] Reduce complexity of CountParameterIndex in call_context.go (39 > 30)
  - [ ] Reduce complexity of findScopeAtPosition in completion_context.go (39 > 30)
  - [ ] Reduce complexity of FindLocalReferences in local_references.go (43 > 30)
  - [ ] Reduce complexity of getGlobalCompletions in scope_completion.go (37 > 30)
  - [ ] Reduce complexity of visit method in semantic_tokens.go (63 > 30)
  - [ ] Reduce complexity of checkStatementForSymbol in symbol_resolver.go (38 > 30)
  - [ ] Reduce complexity of extractClassMembers in type_resolver.go (35 > 30)
  - [ ] Reduce complexity of organizeUsesClause in code_action.go (40 > 30)
  - [ ] Reduce complexity of getUsedIdentifiers in code_action.go (31 > 30)
  - [ ] Reduce complexity of Completion in completion.go (44 > 30)
  - [ ] Reduce complexity of findIdentifierDefinition in definition.go (46 > 30)
  - [ ] Reduce complexity of TestInitialize in initialize_test.go (39 > 30)
  - [ ] Reduce complexity of TestBuildWorkspaceEdit in rename_test.go (51 > 30)

- [ ] **15.9 Code Improvements (gocritic - 4 issues)**
  - [ ] Rewrite if-else chains to switch statements in call_context.go (3 issues)
  - [ ] Replace 'else {if cond {}}' with 'else if cond {}' in call_context_test.go

- [ ] **15.10 Documentation (godoclint - 11 issues)**
  - [ ] Fix multiple godoc comments in analysis package files
  - [ ] Fix multiple godoc comments in lsp package files
  - [ ] Fix multiple godoc comments in server package files
  - [ ] Fix multiple godoc comments in workspace package files

- [ ] **15.11 Module Directives (gomoddirectives - 1 issue)**
  - [ ] Remove or fix replacement directive in go.mod

- [ ] **15.12 Security (gosec - 6 issues)**
  - [ ] Fix integer overflow conversions (G115) in scope_completion.go, rename.go, indexer.go

- [ ] **15.13 Line Length (lll - 19 issues)**
  - [ ] Break long lines in various files (max 120 characters)
  - [ ] Address long function signatures and variable declarations
  - [ ] Address long log statements and comments

- [ ] **15.14 Nested If Statements (nestif - 8 issues)**
  - [ ] Reduce nesting complexity in code_action.go (2 issues)
  - [ ] Reduce nesting complexity in completion.go (2 issues)
  - [ ] Reduce nesting complexity in initialize.go (1 issue)
  - [ ] Reduce nesting complexity in initialize_test.go (1 issue)
  - [ ] Reduce nesting complexity in workspace.go (1 issue)
  - [ ] Reduce nesting complexity in call_context_test.go (1 issue)

- [ ] **15.15 Error Handling (nilerr - 1 issue)**
  - [ ] Fix nil error return in indexer.go

- [ ] **15.16 Return Values (nilnil - 3 issues)**
  - [ ] Fix nil error returns with invalid values in call_context.go

- [ ] **15.17 Named Returns (nonamedreturns - 1 issue)**
  - [ ] Remove named return in text_edit.go

- [ ] **15.18 Parallel Tests (paralleltest - 50 issues)**
  - [ ] Add t.Parallel() calls to all test functions missing them
  - [ ] Address parallel test issues across all test files

- [x] **15.19 Pre-allocation (prealloc - 15 issues)**
  - [x] Pre-allocate slices in various functions across codebase

- [ ] **15.20 Code Quality (revive - 14 issues)**
  - [ ] Fix unused parameters across various functions
  - [ ] Fix empty blocks and variable declarations
  - [ ] Fix indent-error-flow and var-naming issues

- [ ] **15.21 Static Analysis (staticcheck - 4 issues)**
  - [ ] Fix nil checks and type comparisons

- [ ] **15.22 Test Packages (testpackage - 9 issues)**
  - [ ] Rename test packages from `analysis` to `analysis_test` etc.

- [ ] **15.23 Unnecessary Conversions (unconvert - 3 issues)**
  - [ ] Remove unnecessary type conversions

- [ ] **15.24 Unused Parameters (unparam - 2 issues)**
  - [ ] Remove or use unused function parameters

- [ ] **15.25 Variable Names (varnamelen - 30 issues)**
  - [ ] Improve variable naming across codebase (avoid single-letter names)

- [ ] **15.26 Error Wrapping (wrapcheck - 3 issues)**
  - [ ] Properly wrap external package errors

**Outcome**: Clean, maintainable codebase with zero lint issues, improved code quality and consistency.

**Estimated Effort**: 2-3 weeks

---

## Summary

**Total Phases**: 15 (13 complete, 1 in progress)

**Completed Features**:

- Full LSP lifecycle (initialize, shutdown, configuration)
- Document synchronization (open, close, incremental edits)
- Real-time diagnostics (syntax and semantic errors)
- Navigation (hover, go-to-definition, find references)
- Symbols (document and workspace-wide search)
- Code intelligence (completion, signature help)
- Refactoring (rename, code actions)
- Semantic highlighting (token classification)
- Quick fixes (declare variable/function, organize imports)

**Test Coverage**: 200+ unit tests, all passing

**Performance**: Optimized for IDE use with caching and incremental updates

**Next Steps**: Phase 15 - Code quality improvements and linting fixes
