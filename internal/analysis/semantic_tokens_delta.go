// Package analysis provides semantic token delta computation for incremental updates.
package analysis

import (
	"log"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

// DeltaThreshold defines the threshold for using delta vs full response.
// If the delta size is more than this percentage of full size, return full instead.
const DeltaThreshold = 0.7 // 70%

// SemanticTokensDeltaResult wraps either a delta or full response.
type SemanticTokensDeltaResult struct {
	IsDelta bool                              // true if delta, false if full
	Delta   *protocol.SemanticTokensDelta    // set if IsDelta == true
	Full    *protocol.SemanticTokens         // set if IsDelta == false
}

// ComputeSemanticTokensDelta computes the delta between old and new tokens.
// If the delta is too large or oldTokens is nil, it returns a full response instead.
func ComputeSemanticTokensDelta(oldTokens, newTokens []SemanticToken, newResultID string) *SemanticTokensDeltaResult {
	// If no old tokens, return full
	if oldTokens == nil || len(oldTokens) == 0 {
		log.Println("No old tokens, returning full semantic tokens")
		return &SemanticTokensDeltaResult{
			IsDelta: false,
			Full: &protocol.SemanticTokens{
				ResultID: &newResultID,
				Data:     EncodeSemanticTokens(newTokens),
			},
		}
	}

	// If no new tokens but had old tokens, it's a delete-all delta
	if newTokens == nil || len(newTokens) == 0 {
		log.Println("No new tokens, returning delete-all delta")
		edits := []protocol.SemanticTokensEdit{
			{
				Start:       0,
				DeleteCount: uint32(len(oldTokens) * 5), // Each token is 5 uint32 values
				Data:        []uint32{},
			},
		}
		return &SemanticTokensDeltaResult{
			IsDelta: true,
			Delta: &protocol.SemanticTokensDelta{
				ResultID: &newResultID,
				Edits:    edits,
			},
		}
	}

	// Compute edits
	edits := computeEdits(oldTokens, newTokens)

	// Check if delta is worth it
	oldEncoded := EncodeSemanticTokens(oldTokens)
	newEncoded := EncodeSemanticTokens(newTokens)
	deltaSize := calculateDeltaSize(edits)
	fullSize := len(newEncoded)

	// If delta is too large, return full instead
	if float64(deltaSize) > float64(fullSize)*DeltaThreshold {
		log.Printf("Delta too large (%d vs %d), returning full semantic tokens", deltaSize, fullSize)
		return &SemanticTokensDeltaResult{
			IsDelta: false,
			Full: &protocol.SemanticTokens{
				ResultID: &newResultID,
				Data:     newEncoded,
			},
		}
	}

	log.Printf("Returning delta with %d edits (delta size: %d, full size: %d)", len(edits), deltaSize, fullSize)
	return &SemanticTokensDeltaResult{
		IsDelta: true,
		Delta: &protocol.SemanticTokensDelta{
			ResultID: &newResultID,
			Edits:    edits,
		},
	}
}

// computeEdits computes the edit operations to transform oldTokens into newTokens.
// This uses a simple sequential scan algorithm for efficiency.
func computeEdits(oldTokens, newTokens []SemanticToken) []protocol.SemanticTokensEdit {
	var edits []protocol.SemanticTokensEdit

	// Encode both token sets for comparison
	oldEncoded := EncodeSemanticTokens(oldTokens)
	newEncoded := EncodeSemanticTokens(newTokens)

	// Find the common prefix (unchanged tokens at the start)
	commonPrefixLen := 0
	maxPrefix := min(len(oldEncoded), len(newEncoded))
	for commonPrefixLen < maxPrefix && oldEncoded[commonPrefixLen] == newEncoded[commonPrefixLen] {
		commonPrefixLen++
	}

	// Find the common suffix (unchanged tokens at the end)
	commonSuffixLen := 0
	oldSuffixStart := len(oldEncoded)
	newSuffixStart := len(newEncoded)
	for commonSuffixLen < len(oldEncoded)-commonPrefixLen &&
		commonSuffixLen < len(newEncoded)-commonPrefixLen &&
		oldEncoded[oldSuffixStart-1-commonSuffixLen] == newEncoded[newSuffixStart-1-commonSuffixLen] {
		commonSuffixLen++
	}

	// If everything is the same, return empty edits
	if commonPrefixLen+commonSuffixLen >= max(len(oldEncoded), len(newEncoded)) {
		log.Println("No changes detected in semantic tokens")
		return edits
	}

	// Calculate the changed region
	oldChangedStart := commonPrefixLen
	oldChangedEnd := len(oldEncoded) - commonSuffixLen
	newChangedStart := commonPrefixLen
	newChangedEnd := len(newEncoded) - commonSuffixLen

	deleteCount := oldChangedEnd - oldChangedStart
	insertData := newEncoded[newChangedStart:newChangedEnd]

	// Create a single edit that replaces the changed region
	edit := protocol.SemanticTokensEdit{
		Start:       uint32(oldChangedStart),
		DeleteCount: uint32(deleteCount),
		Data:        insertData,
	}

	edits = append(edits, edit)
	return edits
}

// calculateDeltaSize estimates the size of the delta response in uint32 values.
func calculateDeltaSize(edits []protocol.SemanticTokensEdit) int {
	size := 0
	for _, edit := range edits {
		// Each edit has overhead (start + deleteCount) plus the data
		size += 2 + len(edit.Data) // Simplified size calculation
	}
	return size
}

// tokensEqual checks if two tokens are equal in all fields.
func tokensEqual(a, b SemanticToken) bool {
	return a.Line == b.Line &&
		a.StartChar == b.StartChar &&
		a.Length == b.Length &&
		a.TokenType == b.TokenType &&
		a.Modifiers == b.Modifiers
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
