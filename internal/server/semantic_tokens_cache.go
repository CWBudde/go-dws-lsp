// Package server provides semantic tokens caching for incremental updates.
package server

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

// CachedTokens represents a cached set of semantic tokens for a document.
type CachedTokens struct {
	ResultID  string          // Unique identifier for this token set
	Tokens    []SemanticToken // Raw tokens (not encoded)
	Timestamp time.Time       // When this cache entry was created
}

// SemanticTokensCache manages cached semantic tokens for delta computation.
// It stores the previous token sets per document to enable incremental updates.
type SemanticTokensCache struct {
	// cache maps "documentURI:resultId" to cached tokens
	cache map[string]*CachedTokens

	// latestResultId maps documentURI to the most recent resultId
	// This is used to quickly check if a resultId is the latest
	latestResultID map[protocol.DocumentUri]string

	// mu protects concurrent access to the cache
	mu sync.RWMutex
}

// NewSemanticTokensCache creates a new semantic tokens cache.
func NewSemanticTokensCache() *SemanticTokensCache {
	return &SemanticTokensCache{
		cache:          make(map[string]*CachedTokens),
		latestResultID: make(map[protocol.DocumentUri]string),
	}
}

// GenerateResultID generates a unique result identifier based on timestamp and document URI.
// The resultId is used by the client to request delta updates.
func GenerateResultID(uri protocol.DocumentUri, version int) string {
	// Create a hash based on URI, version, and timestamp for uniqueness
	hash := sha256.New()
	hash.Write([]byte(uri))
	fmt.Fprintf(hash, ":%d:%d", version, time.Now().UnixNano())

	return hex.EncodeToString(hash.Sum(nil))[:16] // Use first 16 hex chars
}

// Store saves a set of semantic tokens in the cache with the given resultId.
func (c *SemanticTokensCache) Store(uri protocol.DocumentUri, resultID string, tokens []SemanticToken) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.makeKey(uri, resultID)
	c.cache[key] = &CachedTokens{
		ResultID:  resultID,
		Tokens:    tokens,
		Timestamp: time.Now(),
	}

	// Update latest resultId for this document
	c.latestResultID[uri] = resultID
}

// Retrieve fetches cached tokens for a specific document and resultId.
// Returns the cached tokens and true if found, or nil and false if not found.
func (c *SemanticTokensCache) Retrieve(uri protocol.DocumentUri, resultID string) (*CachedTokens, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.makeKey(uri, resultID)
	cached, found := c.cache[key]

	return cached, found
}

// InvalidateDocument removes all cached tokens for a specific document.
// This should be called when a document is changed or closed.
func (c *SemanticTokensCache) InvalidateDocument(uri protocol.DocumentUri) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove all cache entries for this document
	for key := range c.cache {
		// Keys are in format "uri:resultId", check if they start with this URI
		if len(key) > len(uri) && key[:len(uri)] == uri && key[len(uri)] == ':' {
			delete(c.cache, key)
		}
	}

	// Remove latest resultId tracking
	delete(c.latestResultID, uri)
}

// InvalidateResult removes a specific cached result by URI and resultId.
func (c *SemanticTokensCache) InvalidateResult(uri protocol.DocumentUri, resultID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.makeKey(uri, resultID)
	delete(c.cache, key)

	// If this was the latest result, clear the latest tracking
	if c.latestResultID[uri] == resultID {
		delete(c.latestResultID, uri)
	}
}

// GetLatestResultID returns the most recent resultId for a document.
// Returns empty string if no cached result exists.
func (c *SemanticTokensCache) GetLatestResultID(uri protocol.DocumentUri) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.latestResultID[uri]
}

// Clear removes all cached tokens from the cache.
func (c *SemanticTokensCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*CachedTokens)
	c.latestResultID = make(map[protocol.DocumentUri]string)
}

// Size returns the number of cached token sets.
func (c *SemanticTokensCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}

// makeKey creates a cache key from URI and resultId.
func (c *SemanticTokensCache) makeKey(uri protocol.DocumentUri, resultID string) string {
	return fmt.Sprintf("%s:%s", uri, resultID)
}
