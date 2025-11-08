package server

import (
	"sync"

	"github.com/cwbudde/go-dws/pkg/dwscript"
)

// Document represents an open document in the workspace.
type Document struct {
	URI        string
	Text       string
	Version    int
	LanguageID string

	// Program is the compiled DWScript program (nil if compilation failed).
	// This is stored after successful compilation and provides access to the AST.
	Program *dwscript.Program
}

// DocumentStore manages all open documents.
type DocumentStore struct {
	documents map[string]*Document
	mu        sync.RWMutex
}

// NewDocumentStore creates a new document store.
func NewDocumentStore() *DocumentStore {
	return &DocumentStore{
		documents: make(map[string]*Document),
	}
}

// Set stores or updates a document.
func (ds *DocumentStore) Set(uri string, doc *Document) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.documents[uri] = doc
}

// Get retrieves a document by URI.
func (ds *DocumentStore) Get(uri string) (*Document, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	doc, ok := ds.documents[uri]

	return doc, ok
}

// Delete removes a document from the store.
func (ds *DocumentStore) Delete(uri string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	delete(ds.documents, uri)
}

// List returns all document URIs.
func (ds *DocumentStore) List() []string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	uris := make([]string, 0, len(ds.documents))
	for uri := range ds.documents {
		uris = append(uris, uri)
	}

	return uris
}

// Clear removes all documents from the store.
func (ds *DocumentStore) Clear() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.documents = make(map[string]*Document)
}
