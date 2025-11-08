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

## Summary

**Total Phases**: 15 (14 complete, 1 in progress)

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

**Next Steps**: Phase 15 - Testing, quality assurance, and documentation
