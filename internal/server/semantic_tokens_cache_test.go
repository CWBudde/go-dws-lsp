package server

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestNewSemanticTokensCache(t *testing.T) {
	cache := NewSemanticTokensCache()
	assert.NotNil(t, cache)
	assert.Equal(t, 0, cache.Size())
}

func TestSemanticTokensCache_StoreAndRetrieve(t *testing.T) {
	cache := NewSemanticTokensCache()

	uri := protocol.URI("file:///test.dws")
	resultID := "test-result-id"
	tokens := []SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
		{Line: 1, StartChar: 0, Length: 5, TokenType: 1, Modifiers: 1},
	}

	// Store tokens
	cache.Store(uri, resultID, tokens)
	assert.Equal(t, 1, cache.Size())

	// Retrieve tokens
	retrieved, found := cache.Retrieve(uri, resultID)
	assert.True(t, found)
	assert.Len(t, retrieved.Tokens, len(tokens))
	assert.Equal(t, tokens[0].Line, retrieved.Tokens[0].Line)
	assert.Equal(t, tokens[1].TokenType, retrieved.Tokens[1].TokenType)
}

func TestSemanticTokensCache_RetrieveNotFound(t *testing.T) {
	cache := NewSemanticTokensCache()

	uri := protocol.URI("file:///test.dws")
	resultID := "nonexistent-result-id"

	retrieved, found := cache.Retrieve(uri, resultID)
	assert.False(t, found)
	assert.Nil(t, retrieved)
}

func TestSemanticTokensCache_InvalidateDocument(t *testing.T) {
	cache := NewSemanticTokensCache()

	uri := protocol.URI("file:///test.dws")
	resultID1 := "result-1"
	resultID2 := "result-2"
	tokens := []SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
	}

	// Store multiple results for the same document
	cache.Store(uri, resultID1, tokens)
	cache.Store(uri, resultID2, tokens)
	assert.Equal(t, 2, cache.Size())

	// Invalidate document
	cache.InvalidateDocument(uri)
	assert.Equal(t, 0, cache.Size())

	// Verify both results are gone
	_, found1 := cache.Retrieve(uri, resultID1)
	_, found2 := cache.Retrieve(uri, resultID2)

	assert.False(t, found1)
	assert.False(t, found2)
}

func TestSemanticTokensCache_InvalidateResult(t *testing.T) {
	cache := NewSemanticTokensCache()

	uri := protocol.URI("file:///test.dws")
	resultID1 := "result-1"
	resultID2 := "result-2"
	tokens := []SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
	}

	// Store multiple results for the same document
	cache.Store(uri, resultID1, tokens)
	cache.Store(uri, resultID2, tokens)
	assert.Equal(t, 2, cache.Size())

	// Invalidate only one result
	cache.InvalidateResult(uri, resultID1)
	assert.Equal(t, 1, cache.Size())

	// Verify first result is gone, second remains
	_, found1 := cache.Retrieve(uri, resultID1)
	_, found2 := cache.Retrieve(uri, resultID2)

	assert.False(t, found1)
	assert.True(t, found2)
}

func TestSemanticTokensCache_GetLatestResultID(t *testing.T) {
	cache := NewSemanticTokensCache()

	uri := protocol.URI("file:///test.dws")
	resultID1 := "result-1"
	resultID2 := "result-2"
	tokens := []SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
	}

	// Store results in order
	cache.Store(uri, resultID1, tokens)
	latestID := cache.GetLatestResultID(uri)
	assert.NotEmpty(t, latestID)
	assert.Equal(t, resultID1, latestID)

	cache.Store(uri, resultID2, tokens)
	latestID = cache.GetLatestResultID(uri)
	assert.NotEmpty(t, latestID)
	assert.Equal(t, resultID2, latestID)
}

func TestSemanticTokensCache_GetLatestResultID_NotFound(t *testing.T) {
	cache := NewSemanticTokensCache()

	uri := protocol.URI("file:///nonexistent.dws")
	latestID := cache.GetLatestResultID(uri)
	assert.Empty(t, latestID)
}

func TestSemanticTokensCache_Clear(t *testing.T) {
	cache := NewSemanticTokensCache()

	uri1 := protocol.URI("file:///test1.dws")
	uri2 := protocol.URI("file:///test2.dws")
	tokens := []SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
	}

	// Store multiple documents
	cache.Store(uri1, "result-1", tokens)
	cache.Store(uri2, "result-2", tokens)
	assert.Equal(t, 2, cache.Size())

	// Clear cache
	cache.Clear()
	assert.Equal(t, 0, cache.Size())
}

func TestSemanticTokensCache_MultipleDocumentsDontInterfere(t *testing.T) {
	cache := NewSemanticTokensCache()

	uri1 := protocol.URI("file:///test1.dws")
	uri2 := protocol.URI("file:///test2.dws")
	resultID := "same-result-id"

	tokens1 := []SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
	}
	tokens2 := []SemanticToken{
		{Line: 0, StartChar: 0, Length: 5, TokenType: 1, Modifiers: 1},
		{Line: 1, StartChar: 0, Length: 7, TokenType: 2, Modifiers: 0},
	}

	// Store tokens for different documents with same result ID
	cache.Store(uri1, resultID, tokens1)
	cache.Store(uri2, resultID, tokens2)
	assert.Equal(t, 2, cache.Size())

	// Retrieve tokens for each document
	retrieved1, found1 := cache.Retrieve(uri1, resultID)
	retrieved2, found2 := cache.Retrieve(uri2, resultID)

	assert.True(t, found1)
	assert.True(t, found2)
	assert.Len(t, retrieved1.Tokens, 1)
	assert.Len(t, retrieved2.Tokens, 2)

	// Invalidate one document shouldn't affect the other
	cache.InvalidateDocument(uri1)
	assert.Equal(t, 1, cache.Size())

	_, found1 = cache.Retrieve(uri1, resultID)
	_, found2 = cache.Retrieve(uri2, resultID)

	assert.False(t, found1)
	assert.True(t, found2)
}

func TestSemanticTokensCache_ConcurrentAccess(t *testing.T) {
	cache := NewSemanticTokensCache()

	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch multiple goroutines performing concurrent operations
	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()

			uri := protocol.URI(fmt.Sprintf("file:///test%d.dws", id))
			tokens := []SemanticToken{
				{Line: uint32(id), StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
			}

			for j := range numOperations {
				resultID := fmt.Sprintf("result-%d-%d", id, j)

				// Store
				cache.Store(uri, resultID, tokens)

				// Retrieve
				retrieved, found := cache.Retrieve(uri, resultID)
				assert.True(t, found)
				assert.Len(t, retrieved.Tokens, len(tokens))

				// Get latest
				latestID := cache.GetLatestResultID(uri)

				assert.True(t, found)
				assert.NotEmpty(t, latestID)

				// Invalidate result (every 10 operations)
				if j%10 == 0 && j > 0 {
					oldResultID := fmt.Sprintf("result-%d-%d", id, j-1)
					cache.InvalidateResult(uri, oldResultID)
				}
			}

			// Final invalidation
			cache.InvalidateDocument(uri)
		}(i)
	}

	wg.Wait()

	// All documents should be invalidated
	assert.Equal(t, 0, cache.Size())
}

func TestSemanticTokensCache_GenerateResultID(t *testing.T) {
	uri := protocol.URI("file:///test.dws")
	version1 := 1
	version2 := 2

	// Generate result IDs
	resultID1 := GenerateResultID(uri, version1)
	resultID2 := GenerateResultID(uri, version2)

	// Result IDs should be non-empty and different
	assert.NotEmpty(t, resultID1)
	assert.NotEmpty(t, resultID2)
	assert.NotEqual(t, resultID1, resultID2)

	// Note: Results are not deterministic due to timestamp in generation
	// So we don't test for same inputs generating same result
}

func TestSemanticTokensCache_StoreEmptyTokens(t *testing.T) {
	cache := NewSemanticTokensCache()

	uri := protocol.URI("file:///empty.dws")
	resultID := "empty-result"
	tokens := []SemanticToken{}

	// Store empty token list
	cache.Store(uri, resultID, tokens)
	assert.Equal(t, 1, cache.Size())

	// Retrieve empty token list
	retrieved, found := cache.Retrieve(uri, resultID)
	assert.True(t, found)
	assert.NotNil(t, retrieved)
	assert.Empty(t, retrieved.Tokens)
}

func TestSemanticTokensCache_StoreNilTokens(t *testing.T) {
	cache := NewSemanticTokensCache()

	uri := protocol.URI("file:///nil.dws")
	resultID := "nil-result"

	// Store nil token list
	cache.Store(uri, resultID, nil)
	assert.Equal(t, 1, cache.Size())

	// Retrieve should return cached tokens with nil token slice
	retrieved, found := cache.Retrieve(uri, resultID)
	assert.True(t, found)
	assert.NotNil(t, retrieved)
	assert.Nil(t, retrieved.Tokens)
}

func TestSemanticTokensCache_InvalidateNonexistentDocument(t *testing.T) {
	cache := NewSemanticTokensCache()

	uri := protocol.URI("file:///test.dws")
	tokens := []SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
	}

	cache.Store(uri, "result-1", tokens)
	initialSize := cache.Size()

	// Invalidating nonexistent document should not affect cache
	nonexistentURI := protocol.URI("file:///nonexistent.dws")
	cache.InvalidateDocument(nonexistentURI)

	assert.Equal(t, initialSize, cache.Size())

	// Original document should still be there
	_, found := cache.Retrieve(uri, "result-1")
	assert.True(t, found)
}

func TestSemanticTokensCache_MultipleResultsPerDocument(t *testing.T) {
	cache := NewSemanticTokensCache()

	uri := protocol.URI("file:///test.dws")
	numResults := 5

	// Store multiple results for same document
	for i := range numResults {
		resultID := fmt.Sprintf("result-%d", i)
		tokens := []SemanticToken{
			{Line: uint32(i), StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
		}
		cache.Store(uri, resultID, tokens)
	}

	assert.Equal(t, numResults, cache.Size())

	// Verify all results can be retrieved
	for i := range numResults {
		resultID := fmt.Sprintf("result-%d", i)
		retrieved, found := cache.Retrieve(uri, resultID)
		require.True(t, found)
		require.Len(t, retrieved.Tokens, 1)
		assert.Equal(t, uint32(i), retrieved.Tokens[0].Line)
	}

	// Latest result ID should be the last one stored
	latestID := cache.GetLatestResultID(uri)
	assert.NotEmpty(t, latestID)
	assert.Equal(t, "result-4", latestID)
}
