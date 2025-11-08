package analysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CWBudde/go-dws-lsp/internal/server"
)

func TestComputeSemanticTokensDelta_NoChanges(t *testing.T) {
	tokens := []server.SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
		{Line: 0, StartChar: 4, Length: 5, TokenType: 1, Modifiers: 0},
		{Line: 1, StartChar: 0, Length: 4, TokenType: 2, Modifiers: 1},
	}

	result := ComputeSemanticTokensDelta(tokens, tokens, "new-result-id")

	// Should return delta with empty edits (no changes)
	assert.NotNil(t, result)
	assert.True(t, result.IsDelta)
	require.NotNil(t, result.Delta)
	assert.Equal(t, "new-result-id", *result.Delta.ResultId)
	assert.NotNil(t, result.Delta.Edits)
	assert.Equal(t, 0, len(result.Delta.Edits), "Expected empty edits for no changes")
}

func TestComputeSemanticTokensDelta_SingleTokenModified(t *testing.T) {
	oldTokens := []server.SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
		{Line: 0, StartChar: 4, Length: 5, TokenType: 1, Modifiers: 0},
		{Line: 1, StartChar: 0, Length: 4, TokenType: 2, Modifiers: 1},
	}

	newTokens := []server.SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
		{Line: 0, StartChar: 4, Length: 5, TokenType: 3, Modifiers: 1}, // Changed TokenType and Modifiers
		{Line: 1, StartChar: 0, Length: 4, TokenType: 2, Modifiers: 1},
	}

	result := ComputeSemanticTokensDelta(oldTokens, newTokens, "new-result-id")

	// Should return delta with edits
	assert.NotNil(t, result)
	assert.True(t, result.IsDelta)
	require.NotNil(t, result.Delta)
	assert.Equal(t, "new-result-id", *result.Delta.ResultId)
	assert.NotNil(t, result.Delta.Edits)
	assert.Greater(t, len(result.Delta.Edits), 0, "Expected edits for modified token")

	// Verify edit affects the changed values in the encoded array
	// The algorithm works on encoded values and only replaces what changed
	edit := result.Delta.Edits[0]
	// Common prefix: [0,0,3,0,0,0,4,5] = 8 values
	// Changed in old: [1,0] at indices 8-9
	// Changed in new: [3,1] at indices 8-9
	assert.Equal(t, uint32(8), edit.Start, "Edit should start where values differ")
	assert.Equal(t, uint32(2), edit.DeleteCount, "Should delete 2 changed values")
	assert.Equal(t, 2, len(edit.Data), "Should insert 2 new values")
}

func TestComputeSemanticTokensDelta_TokensAddedAtEnd(t *testing.T) {
	oldTokens := []server.SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
		{Line: 0, StartChar: 4, Length: 5, TokenType: 1, Modifiers: 0},
	}

	newTokens := []server.SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
		{Line: 0, StartChar: 4, Length: 5, TokenType: 1, Modifiers: 0},
		{Line: 1, StartChar: 0, Length: 4, TokenType: 2, Modifiers: 1}, // Added
		{Line: 2, StartChar: 0, Length: 6, TokenType: 3, Modifiers: 0}, // Added
	}

	result := ComputeSemanticTokensDelta(oldTokens, newTokens, "new-result-id")

	// Should return delta with additions
	assert.NotNil(t, result)
	assert.True(t, result.IsDelta)
	require.NotNil(t, result.Delta)
	assert.Equal(t, "new-result-id", *result.Delta.ResultId)
	assert.NotNil(t, result.Delta.Edits)
	assert.Greater(t, len(result.Delta.Edits), 0, "Expected edits for added tokens")

	// Verify edit adds tokens at the end
	edit := result.Delta.Edits[0]
	assert.Equal(t, uint32(10), edit.Start, "Edit should start after existing tokens (10 values)")
	assert.Equal(t, uint32(0), edit.DeleteCount, "Should not delete anything")
	assert.Equal(t, 10, len(edit.Data), "Should insert two tokens (10 values)")
}

func TestComputeSemanticTokensDelta_TokensRemovedFromStart(t *testing.T) {
	oldTokens := []server.SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0}, // Removed
		{Line: 0, StartChar: 4, Length: 5, TokenType: 1, Modifiers: 0}, // Removed
		{Line: 1, StartChar: 0, Length: 4, TokenType: 2, Modifiers: 1},
		{Line: 2, StartChar: 0, Length: 6, TokenType: 3, Modifiers: 0},
	}

	newTokens := []server.SemanticToken{
		{Line: 1, StartChar: 0, Length: 4, TokenType: 2, Modifiers: 1},
		{Line: 2, StartChar: 0, Length: 6, TokenType: 3, Modifiers: 0},
	}

	result := ComputeSemanticTokensDelta(oldTokens, newTokens, "new-result-id")

	// Should return delta with deletions
	assert.NotNil(t, result)
	assert.True(t, result.IsDelta)
	require.NotNil(t, result.Delta)
	assert.Equal(t, "new-result-id", *result.Delta.ResultId)
	assert.NotNil(t, result.Delta.Edits)
	assert.Greater(t, len(result.Delta.Edits), 0, "Expected edits for removed tokens")

	// Verify edit removes tokens from start
	edit := result.Delta.Edits[0]
	assert.Equal(t, uint32(0), edit.Start, "Edit should start at beginning")
	assert.Equal(t, uint32(10), edit.DeleteCount, "Should delete two tokens (10 values)")
}

func TestComputeSemanticTokensDelta_MultipleScatteredChanges(t *testing.T) {
	oldTokens := []server.SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
		{Line: 0, StartChar: 4, Length: 5, TokenType: 1, Modifiers: 0}, // Will be modified
		{Line: 1, StartChar: 0, Length: 4, TokenType: 2, Modifiers: 1},
		{Line: 2, StartChar: 0, Length: 6, TokenType: 3, Modifiers: 0}, // Will be modified
		{Line: 3, StartChar: 0, Length: 2, TokenType: 4, Modifiers: 0},
	}

	newTokens := []server.SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
		{Line: 0, StartChar: 4, Length: 7, TokenType: 5, Modifiers: 1}, // Modified
		{Line: 1, StartChar: 0, Length: 4, TokenType: 2, Modifiers: 1},
		{Line: 2, StartChar: 0, Length: 8, TokenType: 6, Modifiers: 1}, // Modified
		{Line: 3, StartChar: 0, Length: 2, TokenType: 4, Modifiers: 0},
	}

	result := ComputeSemanticTokensDelta(oldTokens, newTokens, "new-result-id")

	// Should return delta with edits
	assert.NotNil(t, result)
	assert.True(t, result.IsDelta)
	require.NotNil(t, result.Delta)
	assert.Equal(t, "new-result-id", *result.Delta.ResultId)
	assert.NotNil(t, result.Delta.Edits)
	assert.Greater(t, len(result.Delta.Edits), 0, "Expected edits for scattered changes")
}

func TestComputeSemanticTokensDelta_AllTokensChanged(t *testing.T) {
	oldTokens := []server.SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
		{Line: 0, StartChar: 4, Length: 5, TokenType: 1, Modifiers: 0},
		{Line: 1, StartChar: 0, Length: 4, TokenType: 2, Modifiers: 1},
	}

	newTokens := []server.SemanticToken{
		{Line: 0, StartChar: 0, Length: 10, TokenType: 5, Modifiers: 1},
		{Line: 1, StartChar: 0, Length: 12, TokenType: 6, Modifiers: 0},
		{Line: 2, StartChar: 0, Length: 8, TokenType: 7, Modifiers: 1},
	}

	result := ComputeSemanticTokensDelta(oldTokens, newTokens, "new-result-id")

	// Could return either delta or full fallback depending on threshold
	assert.NotNil(t, result)

	// If it's a delta, should have edits
	if result.IsDelta {
		require.NotNil(t, result.Delta)
		assert.Equal(t, "new-result-id", *result.Delta.ResultId)
		assert.NotNil(t, result.Delta.Edits)
	} else {
		// If it's full fallback, should have data
		require.NotNil(t, result.Full)
		assert.Equal(t, "new-result-id", *result.Full.ResultID)
		assert.NotNil(t, result.Full.Data)
		assert.Greater(t, len(result.Full.Data), 0)
	}
}

func TestComputeSemanticTokensDelta_EmptyOldTokens(t *testing.T) {
	oldTokens := []server.SemanticToken{}

	newTokens := []server.SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
		{Line: 0, StartChar: 4, Length: 5, TokenType: 1, Modifiers: 0},
	}

	result := ComputeSemanticTokensDelta(oldTokens, newTokens, "new-result-id")

	// Should return full fallback (no previous tokens)
	assert.NotNil(t, result)
	assert.False(t, result.IsDelta)
	require.NotNil(t, result.Full)
	assert.Equal(t, "new-result-id", *result.Full.ResultID)
	assert.NotNil(t, result.Full.Data)
	assert.Equal(t, 10, len(result.Full.Data), "Should have 2 tokens * 5 values")
}

func TestComputeSemanticTokensDelta_NilOldTokens(t *testing.T) {
	newTokens := []server.SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
		{Line: 0, StartChar: 4, Length: 5, TokenType: 1, Modifiers: 0},
	}

	result := ComputeSemanticTokensDelta(nil, newTokens, "new-result-id")

	// Should return full fallback (no previous tokens)
	assert.NotNil(t, result)
	assert.False(t, result.IsDelta)
	require.NotNil(t, result.Full)
	assert.Equal(t, "new-result-id", *result.Full.ResultID)
	assert.NotNil(t, result.Full.Data)
	assert.Equal(t, 10, len(result.Full.Data))
}

func TestComputeSemanticTokensDelta_EmptyNewTokens(t *testing.T) {
	oldTokens := []server.SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
		{Line: 0, StartChar: 4, Length: 5, TokenType: 1, Modifiers: 0},
	}

	newTokens := []server.SemanticToken{}

	result := ComputeSemanticTokensDelta(oldTokens, newTokens, "new-result-id")

	// Should return delta with deletion of all tokens
	assert.NotNil(t, result)
	assert.True(t, result.IsDelta)
	require.NotNil(t, result.Delta)
	assert.Equal(t, "new-result-id", *result.Delta.ResultId)
	assert.NotNil(t, result.Delta.Edits)
	assert.Greater(t, len(result.Delta.Edits), 0, "Expected edits for deletion")

	// Should delete all tokens
	edit := result.Delta.Edits[0]
	assert.Equal(t, uint32(0), edit.Start)
	assert.Equal(t, uint32(10), edit.DeleteCount, "Should delete all tokens (10 values)")
	assert.Equal(t, 0, len(edit.Data), "Should not insert anything")
}

func TestComputeSemanticTokensDelta_BothEmpty(t *testing.T) {
	oldTokens := []server.SemanticToken{}
	newTokens := []server.SemanticToken{}

	result := ComputeSemanticTokensDelta(oldTokens, newTokens, "new-result-id")

	// Should return full fallback with empty data
	assert.NotNil(t, result)
	assert.False(t, result.IsDelta)
	require.NotNil(t, result.Full)
	assert.Equal(t, "new-result-id", *result.Full.ResultID)
	assert.NotNil(t, result.Full.Data)
	assert.Equal(t, 0, len(result.Full.Data))
}

func TestComputeSemanticTokensDelta_BothNil(t *testing.T) {
	result := ComputeSemanticTokensDelta(nil, nil, "new-result-id")

	// Should return full fallback with empty data
	assert.NotNil(t, result)
	assert.False(t, result.IsDelta)
	require.NotNil(t, result.Full)
	assert.Equal(t, "new-result-id", *result.Full.ResultID)
	assert.NotNil(t, result.Full.Data)
	assert.Equal(t, 0, len(result.Full.Data))
}

func TestComputeSemanticTokensDelta_TokensAddedInMiddle(t *testing.T) {
	oldTokens := []server.SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
		{Line: 2, StartChar: 0, Length: 4, TokenType: 2, Modifiers: 1},
	}

	newTokens := []server.SemanticToken{
		{Line: 0, StartChar: 0, Length: 3, TokenType: 0, Modifiers: 0},
		{Line: 1, StartChar: 0, Length: 5, TokenType: 1, Modifiers: 0}, // Added in middle
		{Line: 2, StartChar: 0, Length: 4, TokenType: 2, Modifiers: 1},
	}

	result := ComputeSemanticTokensDelta(oldTokens, newTokens, "new-result-id")

	// Should return delta with edits
	assert.NotNil(t, result)
	assert.True(t, result.IsDelta)
	require.NotNil(t, result.Delta)
	assert.Equal(t, "new-result-id", *result.Delta.ResultId)
	assert.NotNil(t, result.Delta.Edits)
	assert.Greater(t, len(result.Delta.Edits), 0, "Expected edits for added token in middle")
}

func TestComputeSemanticTokensDelta_LargeDocument(t *testing.T) {
	// Create a large document with many tokens
	numTokens := 1000
	oldTokens := make([]server.SemanticToken, numTokens)
	newTokens := make([]server.SemanticToken, numTokens)

	for i := 0; i < numTokens; i++ {
		oldTokens[i] = server.SemanticToken{
			Line:       uint32(i),
			StartChar:  0,
			Length:     5,
			TokenType:  uint32(i % 10),
			Modifiers:  0,
		}
		newTokens[i] = oldTokens[i] // Same tokens
	}

	// Modify just one token in the middle
	newTokens[500].TokenType = 99

	result := ComputeSemanticTokensDelta(oldTokens, newTokens, "new-result-id")

	// Should return delta (not full fallback) since only one token changed
	assert.NotNil(t, result)
	assert.True(t, result.IsDelta, "Should use delta for small change in large document")
	require.NotNil(t, result.Delta)
	assert.Equal(t, "new-result-id", *result.Delta.ResultId)
	assert.NotNil(t, result.Delta.Edits)
}

func TestComputeSemanticTokensDelta_FallbackThreshold(t *testing.T) {
	// Create tokens where delta would be larger than threshold
	oldTokens := make([]server.SemanticToken, 100)
	newTokens := make([]server.SemanticToken, 100)

	for i := 0; i < 100; i++ {
		oldTokens[i] = server.SemanticToken{
			Line:       uint32(i),
			StartChar:  0,
			Length:     5,
			TokenType:  uint32(i % 10),
			Modifiers:  0,
		}
		// Change most tokens to trigger fallback
		newTokens[i] = server.SemanticToken{
			Line:       uint32(i),
			StartChar:  0,
			Length:     5,
			TokenType:  uint32((i + 1) % 10), // Different token type
			Modifiers:  1,                     // Different modifier
		}
	}

	result := ComputeSemanticTokensDelta(oldTokens, newTokens, "new-result-id")

	// With 100% changes, should likely fallback to full
	assert.NotNil(t, result)

	// The implementation should decide based on DeltaThreshold (0.7)
	if result.IsDelta {
		require.NotNil(t, result.Delta)
		assert.Equal(t, "new-result-id", *result.Delta.ResultId)
		assert.NotNil(t, result.Delta.Edits)
	} else {
		require.NotNil(t, result.Full)
		assert.Equal(t, "new-result-id", *result.Full.ResultID)
		assert.NotNil(t, result.Full.Data)
		assert.Greater(t, len(result.Full.Data), 0)
	}
}

func TestEncodeSemanticTokens_ConsistentWithDelta(t *testing.T) {
	// Test that encoding is consistent with delta computation
	tokens := []server.SemanticToken{
		{Line: 0, StartChar: 5, Length: 3, TokenType: 1, Modifiers: 0},
		{Line: 0, StartChar: 10, Length: 4, TokenType: 2, Modifiers: 1},
		{Line: 2, StartChar: 0, Length: 6, TokenType: 3, Modifiers: 0},
	}

	encoded := EncodeSemanticTokens(tokens)
	require.NotNil(t, encoded)
	require.Equal(t, 15, len(encoded), "Expected 3 tokens * 5 values")

	// Verify delta encoding (relative positions)
	// First token: line=0, startChar=5, length=3, tokenType=1, modifiers=0
	assert.Equal(t, uint32(0), encoded[0]) // Line delta
	assert.Equal(t, uint32(5), encoded[1]) // StartChar (absolute for first token on line)
	assert.Equal(t, uint32(3), encoded[2]) // Length
	assert.Equal(t, uint32(1), encoded[3]) // TokenType
	assert.Equal(t, uint32(0), encoded[4]) // Modifiers

	// Second token: line=0 (same line), startChar=10
	assert.Equal(t, uint32(0), encoded[5])  // Line delta (same line)
	assert.Equal(t, uint32(5), encoded[6])  // StartChar delta (10 - 5 = 5)
	assert.Equal(t, uint32(4), encoded[7])  // Length
	assert.Equal(t, uint32(2), encoded[8])  // TokenType
	assert.Equal(t, uint32(1), encoded[9])  // Modifiers

	// Third token: line=2 (delta=2), startChar=0
	assert.Equal(t, uint32(2), encoded[10])  // Line delta (2 - 0 = 2)
	assert.Equal(t, uint32(0), encoded[11])  // StartChar (absolute for first token on new line)
	assert.Equal(t, uint32(6), encoded[12])  // Length
	assert.Equal(t, uint32(3), encoded[13])  // TokenType
	assert.Equal(t, uint32(0), encoded[14])  // Modifiers
}
