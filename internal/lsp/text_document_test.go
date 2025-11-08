package lsp

import (
	"testing"

	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

const (
	testDocumentURI    = "file:///test/document.dws"
	testVarDeclaration = "var x: Integer;"
)

func TestDidOpen(t *testing.T) {
	// Create a new server instance for testing
	srv := server.New()
	SetServer(srv)

	// Create test document parameters
	uri := testDocumentURI
	text := "var x: Integer;\nx := 42;"
	languageID := "dwscript"
	version := int32(1)

	params := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: languageID,
			Version:    version,
			Text:       text,
		},
	}

	// Create mock context
	ctx := &glsp.Context{}

	// Call DidOpen handler
	err := DidOpen(ctx, params)
	if err != nil {
		t.Fatalf("DidOpen returned error: %v", err)
	}

	// Verify document was stored
	doc, exists := srv.Documents().Get(uri)
	if !exists {
		t.Fatal("Document was not stored in DocumentStore")
	}

	// Verify document fields
	if doc.URI != uri {
		t.Errorf("Document URI = %q, want %q", doc.URI, uri)
	}

	if doc.Text != text {
		t.Errorf("Document Text = %q, want %q", doc.Text, text)
	}

	if doc.Version != int(version) {
		t.Errorf("Document Version = %d, want %d", doc.Version, int(version))
	}

	if doc.LanguageID != languageID {
		t.Errorf("Document LanguageID = %q, want %q", doc.LanguageID, languageID)
	}
}

func TestDidClose(t *testing.T) {
	// Create a new server instance for testing
	srv := server.New()
	SetServer(srv)

	// First, open a document
	uri := testDocumentURI
	doc := &server.Document{
		URI:        uri,
		Text:       testVarDeclaration,
		Version:    1,
		LanguageID: "dwscript",
	}
	srv.Documents().Set(uri, doc)

	// Verify document exists
	_, exists := srv.Documents().Get(uri)
	if !exists {
		t.Fatal("Document should exist before DidClose")
	}

	// Create close parameters
	params := &protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri,
		},
	}

	// Create mock context
	ctx := &glsp.Context{}

	// Call DidClose handler
	err := DidClose(ctx, params)
	if err != nil {
		t.Fatalf("DidClose returned error: %v", err)
	}

	// Verify document was removed
	_, exists = srv.Documents().Get(uri)
	if exists {
		t.Error("Document should be removed from DocumentStore after DidClose")
	}
}

func TestDidChange_FullSync(t *testing.T) {
	// Create a new server instance for testing
	srv := server.New()
	SetServer(srv)

	// First, open a document
	uri := testDocumentURI
	originalText := testVarDeclaration
	doc := &server.Document{
		URI:        uri,
		Text:       originalText,
		Version:    1,
		LanguageID: "dwscript",
	}
	srv.Documents().Set(uri, doc)

	// Create change parameters (full sync - no range)
	newText := "var x: Integer;\nx := 42;\nPrintLn(x);"
	newVersion := int32(2)

	params := &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Version: newVersion,
		},
		ContentChanges: []any{
			protocol.TextDocumentContentChangeEvent{
				Range: nil, // nil range means full sync
				Text:  newText,
			},
		},
	}

	// Create mock context
	ctx := &glsp.Context{}

	// Call DidChange handler
	err := DidChange(ctx, params)
	if err != nil {
		t.Fatalf("DidChange returned error: %v", err)
	}

	// Verify document was updated
	updatedDoc, exists := srv.Documents().Get(uri)
	if !exists {
		t.Fatal("Document should still exist after DidChange")
	}

	// Verify text was replaced
	if updatedDoc.Text != newText {
		t.Errorf("Document Text = %q, want %q", updatedDoc.Text, newText)
	}

	// Verify version was updated
	if updatedDoc.Version != int(newVersion) {
		t.Errorf("Document Version = %d, want %d", updatedDoc.Version, int(newVersion))
	}
}

func TestDidChange_IncrementalSync_SingleLine(t *testing.T) {
	// Create a new server instance for testing
	srv := server.New()
	SetServer(srv)

	// First, open a document
	uri := testDocumentURI
	originalText := testVarDeclaration
	doc := &server.Document{
		URI:        uri,
		Text:       originalText,
		Version:    1,
		LanguageID: "dwscript",
	}
	srv.Documents().Set(uri, doc)

	// Create incremental change parameters
	// Change "Integer" to "String" (positions 7-14, replace with "String")
	newVersion := int32(2)

	params := &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Version: newVersion,
		},
		ContentChanges: []any{
			protocol.TextDocumentContentChangeEvent{
				Range: &protocol.Range{
					Start: protocol.Position{Line: 0, Character: 7},
					End:   protocol.Position{Line: 0, Character: 14},
				},
				Text: "String",
			},
		},
	}

	// Create mock context
	ctx := &glsp.Context{}

	// Call DidChange handler
	err := DidChange(ctx, params)
	if err != nil {
		t.Fatalf("DidChange returned error: %v", err)
	}

	// Verify document was updated
	updatedDoc, exists := srv.Documents().Get(uri)
	if !exists {
		t.Fatal("Document should still exist after DidChange")
	}

	// Verify text was correctly updated
	expectedText := "var x: String;"
	if updatedDoc.Text != expectedText {
		t.Errorf("Document Text = %q, want %q", updatedDoc.Text, expectedText)
	}

	// Verify version was updated
	if updatedDoc.Version != int(newVersion) {
		t.Errorf("Document Version = %d, want %d", updatedDoc.Version, int(newVersion))
	}
}

func TestDidChange_IncrementalSync_MultiLine(t *testing.T) {
	// Create a new server instance for testing
	srv := server.New()
	SetServer(srv)

	// First, open a document with multiple lines
	uri := testDocumentURI
	originalText := "var x: Integer;\nvar y: String;\nPrintLn(x);"
	doc := &server.Document{
		URI:        uri,
		Text:       originalText,
		Version:    1,
		LanguageID: "dwscript",
	}
	srv.Documents().Set(uri, doc)

	// Create incremental change that spans multiple lines
	// Delete "var y: String;\n" (entire second line)
	newVersion := int32(2)

	params := &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Version: newVersion,
		},
		ContentChanges: []any{
			protocol.TextDocumentContentChangeEvent{
				Range: &protocol.Range{
					Start: protocol.Position{Line: 1, Character: 0},
					End:   protocol.Position{Line: 2, Character: 0},
				},
				Text: "",
			},
		},
	}

	// Create mock context
	ctx := &glsp.Context{}

	// Call DidChange handler
	err := DidChange(ctx, params)
	if err != nil {
		t.Fatalf("DidChange returned error: %v", err)
	}

	// Verify document was updated
	updatedDoc, exists := srv.Documents().Get(uri)
	if !exists {
		t.Fatal("Document should still exist after DidChange")
	}

	// Verify text was correctly updated
	expectedText := "var x: Integer;\nPrintLn(x);"
	if updatedDoc.Text != expectedText {
		t.Errorf("Document Text = %q, want %q", updatedDoc.Text, expectedText)
	}
}

func TestDidChange_IncrementalSync_Insertion(t *testing.T) {
	// Create a new server instance for testing
	srv := server.New()
	SetServer(srv)

	// First, open a document
	uri := testDocumentURI
	originalText := "var x: Integer;\nPrintLn(x);"
	doc := &server.Document{
		URI:        uri,
		Text:       originalText,
		Version:    1,
		LanguageID: "dwscript",
	}
	srv.Documents().Set(uri, doc)

	// Insert a new line between the two existing lines
	newVersion := int32(2)

	params := &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Version: newVersion,
		},
		ContentChanges: []any{
			protocol.TextDocumentContentChangeEvent{
				Range: &protocol.Range{
					Start: protocol.Position{Line: 0, Character: 15}, // End of first line
					End:   protocol.Position{Line: 0, Character: 15},
				},
				Text: "\nx := 42;",
			},
		},
	}

	// Create mock context
	ctx := &glsp.Context{}

	// Call DidChange handler
	err := DidChange(ctx, params)
	if err != nil {
		t.Fatalf("DidChange returned error: %v", err)
	}

	// Verify document was updated
	updatedDoc, exists := srv.Documents().Get(uri)
	if !exists {
		t.Fatal("Document should still exist after DidChange")
	}

	// Verify text was correctly updated
	expectedText := "var x: Integer;\nx := 42;\nPrintLn(x);"
	if updatedDoc.Text != expectedText {
		t.Errorf("Document Text = %q, want %q", updatedDoc.Text, expectedText)
	}
}

func TestDidChange_VersionTracking(t *testing.T) {
	// Create a new server instance for testing
	srv := server.New()
	SetServer(srv)

	// Open a document
	uri := testDocumentURI
	originalText := testVarDeclaration
	doc := &server.Document{
		URI:        uri,
		Text:       originalText,
		Version:    1,
		LanguageID: "dwscript",
	}
	srv.Documents().Set(uri, doc)

	// Apply multiple changes and verify version tracking
	versions := []int32{2, 3, 4, 5}
	ctx := &glsp.Context{}

	for _, version := range versions {
		params := &protocol.DidChangeTextDocumentParams{
			TextDocument: protocol.VersionedTextDocumentIdentifier{
				TextDocumentIdentifier: protocol.TextDocumentIdentifier{
					URI: uri,
				},
				Version: version,
			},
			ContentChanges: []any{
				protocol.TextDocumentContentChangeEvent{
					Range: nil,
					Text:  "var x: Integer; // version " + string(rune(version)),
				},
			},
		}

		err := DidChange(ctx, params)
		if err != nil {
			t.Fatalf("DidChange returned error on version %d: %v", version, err)
		}

		// Verify version was updated
		updatedDoc, exists := srv.Documents().Get(uri)
		if !exists {
			t.Fatalf("Document should exist after change to version %d", version)
		}

		if updatedDoc.Version != int(version) {
			t.Errorf("After change, Document Version = %d, want %d", updatedDoc.Version, int(version))
		}
	}
}

func TestDidChange_MultipleChanges(t *testing.T) {
	// Create a new server instance for testing
	srv := server.New()
	SetServer(srv)

	// Open a document
	uri := testDocumentURI
	originalText := "var x: Integer;\nvar y: Integer;"
	doc := &server.Document{
		URI:        uri,
		Text:       originalText,
		Version:    1,
		LanguageID: "dwscript",
	}
	srv.Documents().Set(uri, doc)

	// Apply multiple changes in a single DidChange notification
	newVersion := int32(2)

	params := &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Version: newVersion,
		},
		ContentChanges: []any{
			// Change 1: Replace "Integer" with "String" on line 1
			protocol.TextDocumentContentChangeEvent{
				Range: &protocol.Range{
					Start: protocol.Position{Line: 0, Character: 7},
					End:   protocol.Position{Line: 0, Character: 14},
				},
				Text: "String",
			},
			// Change 2: Replace "Integer" with "Float" on line 2
			protocol.TextDocumentContentChangeEvent{
				Range: &protocol.Range{
					Start: protocol.Position{Line: 1, Character: 7},
					End:   protocol.Position{Line: 1, Character: 14},
				},
				Text: "Float",
			},
		},
	}

	// Create mock context
	ctx := &glsp.Context{}

	// Call DidChange handler
	err := DidChange(ctx, params)
	if err != nil {
		t.Fatalf("DidChange returned error: %v", err)
	}

	// Verify document was updated
	updatedDoc, exists := srv.Documents().Get(uri)
	if !exists {
		t.Fatal("Document should still exist after DidChange")
	}

	// Verify both changes were applied
	// Note: The second change's positions refer to the ORIGINAL text,
	// but our implementation applies them sequentially to the updated text
	// This test verifies our implementation handles this correctly
	expectedText := "var x: String;\nvar y: Float;"
	if updatedDoc.Text != expectedText {
		t.Errorf("Document Text = %q, want %q", updatedDoc.Text, expectedText)
	}
}

func TestDidOpen_NonexistentServer(t *testing.T) {
	// Set server to nil to test error handling
	SetServer(nil)

	params := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        testDocumentURI,
			LanguageID: "dwscript",
			Version:    1,
			Text:       "var x: Integer;",
		},
	}

	ctx := &glsp.Context{}

	// Should not crash or return error, just log warning
	err := DidOpen(ctx, params)
	if err != nil {
		t.Errorf("DidOpen should not return error when server is nil, got: %v", err)
	}
}

func TestDidChange_NonexistentDocument(t *testing.T) {
	// Create a new server instance
	srv := server.New()
	SetServer(srv)

	// Try to change a document that wasn't opened
	uri := "file:///test/nonexistent.dws"
	params := &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Version: 1,
		},
		ContentChanges: []any{
			protocol.TextDocumentContentChangeEvent{
				Range: nil,
				Text:  "var x: Integer;",
			},
		},
	}

	ctx := &glsp.Context{}

	// Should not crash or return error, just log warning
	err := DidChange(ctx, params)
	if err != nil {
		t.Errorf("DidChange should not return error for nonexistent document, got: %v", err)
	}

	// Document should still not exist
	_, exists := srv.Documents().Get(uri)
	if exists {
		t.Error("Nonexistent document should not be created by DidChange")
	}
}
