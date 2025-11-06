// Package lsp implements LSP protocol handlers.
package lsp

import (
	"log"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/document"
	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// DidOpen handles the textDocument/didOpen notification.
// This is sent when a document is opened in the editor.
func DidOpen(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in DidOpen")
		return nil
	}

	// Extract document information
	uri := params.TextDocument.URI
	text := params.TextDocument.Text
	languageID := params.TextDocument.LanguageID
	version := int(params.TextDocument.Version)

	// Create document struct
	doc := &server.Document{
		URI:        uri,
		Text:       text,
		Version:    version,
		LanguageID: languageID,
	}

	// Store document in DocumentStore
	srv.Documents().Set(uri, doc)

	log.Printf("Document opened: %s (version %d, language %s, %d bytes)\n",
		uri, version, languageID, len(text))

	// TODO Phase 2: Parse document and publish diagnostics
	// For now, we just store the document content

	return nil
}

// DidClose handles the textDocument/didClose notification.
// This is sent when a document is closed in the editor.
func DidClose(context *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in DidClose")
		return nil
	}

	// Extract URI
	uri := params.TextDocument.URI

	// Remove document from store
	srv.Documents().Delete(uri)

	log.Printf("Document closed: %s\n", uri)

	// Send empty diagnostics to clear error markers in the editor
	// Only send notification if context is properly initialized (not in tests)
	if context != nil && context.Notify != nil {
		diagnosticsParams := &protocol.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: []protocol.Diagnostic{},
		}
		context.Notify(protocol.ServerTextDocumentPublishDiagnostics, diagnosticsParams)
	}

	return nil
}

// DidChange handles the textDocument/didChange notification.
// This is sent when a document's content changes in the editor.
// It supports both full and incremental sync modes.
func DidChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in DidChange")
		return nil
	}

	// Extract URI and version
	uri := params.TextDocument.URI
	version := int(params.TextDocument.Version)

	// Retrieve document from store
	doc, exists := srv.Documents().Get(uri)
	if !exists {
		log.Printf("Warning: Document not found for didChange: %s\n", uri)
		return nil
	}

	// Apply all content changes
	newText := doc.Text
	for i, changeInterface := range params.ContentChanges {
		// Type assert to TextDocumentContentChangeEvent
		change, ok := changeInterface.(protocol.TextDocumentContentChangeEvent)
		if !ok {
			log.Printf("Warning: Invalid content change type at index %d for %s\n", i, uri)
			continue
		}

		if change.Range == nil {
			// Full sync mode: replace entire document text
			newText = change.Text
			log.Printf("Document changed (full sync): %s (version %d, change %d/%d)\n",
				uri, version, i+1, len(params.ContentChanges))
		} else {
			// Incremental sync mode: apply diff
			updatedText, err := document.ApplyContentChange(newText, change)
			if err != nil {
				log.Printf("Error applying incremental change to %s: %v\n", uri, err)
				// Continue with unchanged text to avoid corruption
				continue
			}
			newText = updatedText
			log.Printf("Document changed (incremental): %s (version %d, change %d/%d)\n",
				uri, version, i+1, len(params.ContentChanges))
		}
	}

	// Update document in store
	updatedDoc := &server.Document{
		URI:        uri,
		Text:       newText,
		Version:    version,
		LanguageID: doc.LanguageID,
	}
	srv.Documents().Set(uri, updatedDoc)

	// TODO Phase 2: Re-parse document and publish diagnostics

	return nil
}
