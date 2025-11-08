// Package server provides completion caching for performance optimization.
package server

import (
	"sync"
	"time"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

// CompletionCache caches completion items for performance.
// Task 9.17: Cache global symbol suggestions for performance.
type CompletionCache struct {
	// Per-document caches
	documentCaches map[string]*DocumentCompletionCache

	mu sync.RWMutex
}

// DocumentCompletionCache stores cached completion items for a specific document.
type DocumentCompletionCache struct {
	// Global symbols from this document
	globalSymbols []protocol.CompletionItem

	// Keywords cached for this document (can vary if snippets are disabled)
	keywords []protocol.CompletionItem

	// Built-in functions and types
	builtins []protocol.CompletionItem

	// Last time the cache was updated
	lastUpdate time.Time

	// Document version when cache was built
	version int32

	mu sync.RWMutex
}

// NewCompletionCache creates a new completion cache.
func NewCompletionCache() *CompletionCache {
	return &CompletionCache{
		documentCaches: make(map[string]*DocumentCompletionCache),
	}
}

// CachedCompletionItems represents a set of cached completion items.
type CachedCompletionItems struct {
	Keywords      []protocol.CompletionItem
	Builtins      []protocol.CompletionItem
	GlobalSymbols []protocol.CompletionItem
}

// GetCachedItems returns all cached completion items for a document, or nil if not cached.
func (c *CompletionCache) GetCachedItems(uri string, version int32) *CachedCompletionItems {
	c.mu.RLock()
	defer c.mu.RUnlock()

	docCache, exists := c.documentCaches[uri]
	if !exists {
		return nil
	}

	docCache.mu.RLock()
	defer docCache.mu.RUnlock()

	// Return nil if version doesn't match (document changed)
	if docCache.version != version {
		return nil
	}

	return &CachedCompletionItems{
		Keywords:      docCache.keywords,
		Builtins:      docCache.builtins,
		GlobalSymbols: docCache.globalSymbols,
	}
}

// SetCachedItems caches completion items for a document.
func (c *CompletionCache) SetCachedItems(uri string, version int32, items *CachedCompletionItems) {
	c.mu.Lock()
	defer c.mu.Unlock()

	docCache, exists := c.documentCaches[uri]
	if !exists {
		docCache = &DocumentCompletionCache{}
		c.documentCaches[uri] = docCache
	}

	docCache.mu.Lock()
	defer docCache.mu.Unlock()

	if items.Keywords != nil {
		docCache.keywords = items.Keywords
	}

	if items.Builtins != nil {
		docCache.builtins = items.Builtins
	}

	if items.GlobalSymbols != nil {
		docCache.globalSymbols = items.GlobalSymbols
	}

	docCache.version = version
	docCache.lastUpdate = time.Now()
}

// InvalidateDocument invalidates the cache for a specific document.
func (c *CompletionCache) InvalidateDocument(uri string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.documentCaches, uri)
}

// Clear clears all cached completion items.
func (c *CompletionCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.documentCaches = make(map[string]*DocumentCompletionCache)
}
