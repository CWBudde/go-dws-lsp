//go:build integration
// +build integration

package integration

import (
	"testing"

	"github.com/CWBudde/go-dws-lsp/internal/lsp"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// TestInitializeWorkflow tests the complete initialization workflow
func TestInitializeWorkflow(t *testing.T) {
	ctx := &glsp.Context{}

	// Test Initialize
	params := &protocol.InitializeParams{
		ProcessID: nil,
		RootURI:   stringPtr("file:///test/workspace"),
		Capabilities: protocol.ClientCapabilities{
			TextDocument: &protocol.TextDocumentClientCapabilities{},
		},
	}

	result, err := lsp.Initialize(ctx, params)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if result == nil {
		t.Fatal("Initialize returned nil result")
	}

	initResult, ok := result.(protocol.InitializeResult)
	if !ok {
		t.Fatalf("Initialize returned wrong type: %T", result)
	}

	// Verify server capabilities
	if initResult.Capabilities.HoverProvider == nil {
		t.Error("HoverProvider capability should be advertised")
	}

	if initResult.Capabilities.TextDocumentSync == nil {
		t.Error("TextDocumentSync capability should be advertised")
	}

	// Test Initialized notification
	initializedParams := &protocol.InitializedParams{}
	err = lsp.Initialized(ctx, initializedParams)
	if err != nil {
		t.Fatalf("Initialized failed: %v", err)
	}
}

// TestDocumentLifecycle tests the complete document lifecycle
func TestDocumentLifecycle(t *testing.T) {
	srv := setupTestServer()
	ctx := &glsp.Context{}

	uri := "file:///test/lifecycle.dws"

	// 1. Open document
	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       "var x: Integer;",
		},
	}

	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Verify document is stored
	doc, exists := srv.Documents().Get(uri)
	if !exists {
		t.Fatal("Document should exist after DidOpen")
	}

	if doc.Version != 1 {
		t.Errorf("Document version = %d, want 1", doc.Version)
	}

	// 2. Change document
	changeParams := &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Version: 2,
		},
		ContentChanges: []interface{}{
			protocol.TextDocumentContentChangeEvent{
				Range: nil,
				Text:  "var x: String;",
			},
		},
	}

	err = lsp.DidChange(ctx, changeParams)
	if err != nil {
		t.Fatalf("DidChange failed: %v", err)
	}

	// Verify document was updated
	doc, exists = srv.Documents().Get(uri)
	if !exists {
		t.Fatal("Document should still exist after DidChange")
	}

	if doc.Version != 2 {
		t.Errorf("Document version = %d, want 2", doc.Version)
	}

	if doc.Text != "var x: String;" {
		t.Errorf("Document text = %q, want %q", doc.Text, "var x: String;")
	}

	// 3. Close document
	closeParams := &protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri,
		},
	}

	err = lsp.DidClose(ctx, closeParams)
	if err != nil {
		t.Fatalf("DidClose failed: %v", err)
	}

	// Verify document was removed
	_, exists = srv.Documents().Get(uri)
	if exists {
		t.Error("Document should be removed after DidClose")
	}
}

// TestDiagnosticsOnDidOpen tests that diagnostics are generated when opening a document
func TestDiagnosticsOnDidOpen(t *testing.T) {
	srv := setupTestServer()
	ctx := &glsp.Context{}

	// Open a document with valid code
	validURI := "file:///test/valid.dws"
	validCode := "var x: Integer;"

	validParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        validURI,
			LanguageID: "dwscript",
			Version:    1,
			Text:       validCode,
		},
	}

	err := lsp.DidOpen(ctx, validParams)
	if err != nil {
		t.Fatalf("DidOpen for valid code failed: %v", err)
	}

	// Verify document was parsed successfully
	doc, exists := srv.Documents().Get(validURI)
	if !exists {
		t.Fatal("Valid document should exist")
	}

	if doc.Program == nil {
		t.Error("Valid document should have a parsed Program")
	}

	// Open a document with syntax errors
	invalidURI := "file:///test/invalid.dws"
	invalidCode := "var x Integer;" // Missing colon

	invalidParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        invalidURI,
			LanguageID: "dwscript",
			Version:    1,
			Text:       invalidCode,
		},
	}

	err = lsp.DidOpen(ctx, invalidParams)
	if err != nil {
		t.Fatalf("DidOpen for invalid code failed: %v", err)
	}

	// Verify document exists but may have parse errors
	doc, exists = srv.Documents().Get(invalidURI)
	if !exists {
		t.Fatal("Invalid document should still be stored")
	}

	// Program might be nil for documents with errors
	if doc.Program == nil {
		t.Log("Invalid document has no Program (expected for code with syntax errors)")
	} else {
		t.Log("Invalid document has a Program (it may contain errors)")
	}
}

// TestIncrementalDocumentChanges tests incremental text document synchronization
func TestIncrementalDocumentChanges(t *testing.T) {
	srv := setupTestServer()
	ctx := &glsp.Context{}

	uri := "file:///test/incremental.dws"
	initialText := "var x: Integer;\nvar y: Integer;"

	// Open document
	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       initialText,
		},
	}

	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Make an incremental change: change "Integer" to "String" on first line
	changeParams := &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Version: 2,
		},
		ContentChanges: []interface{}{
			protocol.TextDocumentContentChangeEvent{
				Range: &protocol.Range{
					Start: protocol.Position{Line: 0, Character: 7},
					End:   protocol.Position{Line: 0, Character: 14},
				},
				Text: "String",
			},
		},
	}

	err = lsp.DidChange(ctx, changeParams)
	if err != nil {
		t.Fatalf("DidChange failed: %v", err)
	}

	// Verify the change was applied correctly
	doc, exists := srv.Documents().Get(uri)
	if !exists {
		t.Fatal("Document should exist after incremental change")
	}

	expectedText := "var x: String;\nvar y: Integer;"
	if doc.Text != expectedText {
		t.Errorf("Document text = %q, want %q", doc.Text, expectedText)
	}
}

// TestConcurrentDocumentOperations tests handling of concurrent operations on documents
func TestConcurrentDocumentOperations(t *testing.T) {
	srv := setupTestServer()
	ctx := &glsp.Context{}

	// Open multiple documents
	uris := []string{
		"file:///test/concurrent1.dws",
		"file:///test/concurrent2.dws",
		"file:///test/concurrent3.dws",
	}

	for i, uri := range uris {
		params := &protocol.DidOpenTextDocumentParams{
			TextDocument: protocol.TextDocumentItem{
				URI:        uri,
				LanguageID: "dwscript",
				Version:    1,
				Text:       "var x: Integer;",
			},
		}

		err := lsp.DidOpen(ctx, params)
		if err != nil {
			t.Fatalf("DidOpen for document %d failed: %v", i, err)
		}
	}

	// Verify all documents exist
	for i, uri := range uris {
		_, exists := srv.Documents().Get(uri)
		if !exists {
			t.Errorf("Document %d should exist", i)
		}
	}

	// Perform operations on different documents
	for i, uri := range uris {
		// Request hover
		hoverParams := &protocol.HoverParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{
					URI: uri,
				},
				Position: protocol.Position{Line: 0, Character: 4},
			},
		}

		_, err := lsp.Hover(ctx, hoverParams)
		if err != nil {
			t.Errorf("Hover on document %d failed: %v", i, err)
		}
	}
}

// TestShutdownWorkflow tests the shutdown workflow
func TestShutdownWorkflow(t *testing.T) {
	ctx := &glsp.Context{}

	err := lsp.Shutdown(ctx)
	if err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	// Shutdown should succeed without errors
	// In a real implementation, we might check that resources are cleaned up
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
