# Task 12.20: Semantic Tokens Delta Support - Implementation Plan

## Overview
Implement incremental semantic token updates to improve performance for large files with small changes.

---

## Phase 1: Foundation & Data Structures (Essential - Start Today)

### 12.20.1: Create SemanticTokensCache data structure
- Create `internal/server/semantic_tokens_cache.go`
- Define `SemanticTokensCache` struct with:
  - Map of `documentURI + resultId -> CachedTokens`
  - RWMutex for thread safety
  - Methods: Store, Retrieve, Invalidate, Clear
- Define `CachedTokens` struct:
  - ResultId string
  - Tokens []SemanticToken (raw format)
  - Timestamp (for LRU eviction)

### 12.20.2: Implement ResultId generation
- Create function `GenerateResultId() string`
- Use UUID v4 or timestamp-based hash
- Ensure uniqueness per request

### 12.20.3: Add cache to Server struct
- Add `semanticTokensCache *SemanticTokensCache` field to Server
- Initialize in `server.New()`
- Add getter method `SemanticTokensCache()`

### 12.20.4: Implement cache management methods
- `StoreTokens(uri, resultId, tokens)`
- `RetrieveTokens(uri, resultId) (tokens, found)`
- `InvalidateDocument(uri)` - clear all results for a document
- `InvalidateResult(uri, resultId)` - clear specific result

---

## Phase 2: Delta Computation Logic (Essential - Start Today)

### 12.20.5: Create delta computation module
- Create `internal/analysis/semantic_tokens_delta.go`
- Define `SemanticTokensEdit` struct (if not in protocol)
- Define `ComputeSemanticTokensDelta` function signature

### 12.20.6: Implement token comparison algorithm
- Function: `CompareTokens(oldTokens, newTokens []SemanticToken) []Edit`
- Use diff algorithm (simple sequential scan or LCS-based)
- Identify:
  - Unchanged tokens (skip)
  - Modified tokens (delete + insert)
  - New tokens (insert)
  - Removed tokens (delete)

### 12.20.7: Implement SemanticTokensEdit generation
- Convert diff operations to LSP `SemanticTokensEdit` format
- Each edit contains:
  - `Start` - index in old encoded array (multiple of 5)
  - `DeleteCount` - how many uint32 values to delete
  - `Data` - new encoded tokens to insert
- Handle delta encoding properly

### 12.20.8: Handle edge cases
- All tokens changed (fallback to full)
- No changes (empty edit list)
- Document completely new (fallback to full)
- Only additions at end (optimize)
- Only deletions at start (optimize)

---

## Phase 3: Handler Implementation (Essential - Start Today)

### 12.20.9: Implement textDocument/semanticTokens/full/delta handler
- Create handler in `internal/lsp/semantic_tokens_handler.go`
- Function: `SemanticTokensFullDelta(context, params) (*SemanticTokensDelta, error)`
- Extract `previousResultId` from params
- Retrieve old tokens from cache

### 12.20.10: Modify full handler to support caching
- Update `SemanticTokensFull` to:
  - Generate resultId
  - Store tokens in cache before encoding
  - Return resultId in response
- Maintain backward compatibility

### 12.20.11: Handle previousResultId in delta handler
- If resultId not found in cache:
  - Log warning
  - Fallback to full tokens with new resultId
- If found:
  - Compute new tokens
  - Compute delta
  - Return delta response

### 12.20.12: Implement delta-to-full fallback
- If delta is too large (>70% of full size):
  - Return full tokens instead
  - Include new resultId
- Handle nil/empty previous tokens

---

## Phase 4: Integration & Lifecycle (Essential - Today/Tomorrow)

### 12.20.13: Advertise delta support in capabilities
- File: `internal/lsp/initialize.go`
- Modify `SemanticTokensOptions`:
  - Set `Full` to struct with `Delta: true`
  - Type: `SemanticTokensFullOptions{Delta: true}`

### 12.20.14: Register delta handler
- File: `cmd/go-dws-lsp/main.go`
- Add to protocol.Handler:
  - `TextDocumentSemanticTokensFullDelta: lsp.SemanticTokensFullDelta`

### 12.20.15: Invalidate cache on document changes
- File: `internal/lsp/text_document.go`
- In `DidChange` handler:
  - Call `srv.SemanticTokensCache().InvalidateDocument(uri)`
  - Ensure called after document update

### 12.20.16: Cleanup cache on document close
- File: `internal/lsp/text_document.go`
- In `DidClose` handler:
  - Call `srv.SemanticTokensCache().InvalidateDocument(uri)`
  - Free memory for closed documents

### 12.20.17: Implement cache size limits (Optional - Low Priority)
- Add max cache size (e.g., 50 documents)
- Implement LRU eviction
- Add metrics/logging for cache hits/misses
- **Can defer this to later**

---

## Phase 5: Testing & Validation (Tomorrow or Later)

### 12.20.18: Write unit tests for delta computation
- Test file: `internal/analysis/semantic_tokens_delta_test.go`
- Test cases:
  - No changes
  - Single token modified
  - Tokens added at end
  - Tokens removed from start
  - Multiple scattered changes
  - All tokens changed

### 12.20.19: Write integration tests for delta handler
- Test file: `internal/lsp/semantic_tokens_handler_test.go`
- Test cases:
  - First request (no previous resultId)
  - Second request with valid resultId
  - Request with expired/invalid resultId
  - Delta vs full response decision

### 12.20.20: Test cache invalidation
- Test document change invalidates cache
- Test document close clears cache
- Test multiple documents don't interfere

### 12.20.21: Test edge cases
- Empty document
- Very large documents
- Rapid successive changes
- Concurrent requests

---

## Implementation Priority

### **Must Do Today (Core Functionality):**
- Phase 1: All tasks (12.20.1 - 12.20.4)
- Phase 2: All tasks (12.20.5 - 12.20.8)
- Phase 3: All tasks (12.20.9 - 12.20.12)
- Phase 4: Tasks 12.20.13 - 12.20.16

### **Can Do Tomorrow (Polish & Testing):**
- Phase 4: Task 12.20.17 (LRU eviction - optional)
- Phase 5: All testing tasks (12.20.18 - 12.20.21)

### **Estimated Time:**
- Phase 1: 30-45 minutes
- Phase 2: 45-60 minutes
- Phase 3: 45-60 minutes
- Phase 4 (core): 30 minutes
- **Total Core Work: ~3 hours**

---

## Success Criteria

✅ Delta handler registered and advertised
✅ Cache stores and retrieves tokens correctly
✅ Delta computation produces valid edits
✅ Fallback to full tokens when delta unavailable
✅ Cache invalidated on document changes
✅ No memory leaks (cleanup on close)
✅ Manual testing shows delta responses work

---

## Risks & Mitigation

**Risk 1:** Delta computation is complex
- **Mitigation:** Start with simple sequential diff, optimize later

**Risk 2:** Memory usage from caching
- **Mitigation:** Invalidate aggressively, implement size limits

**Risk 3:** Incorrect delta encoding breaks highlighting
- **Mitigation:** Always include fallback to full tokens

**Risk 4:** Thread safety issues
- **Mitigation:** Use RWMutex, test concurrent access
