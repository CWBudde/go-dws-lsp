package lsp

import (
	"testing"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/analysis"
	"github.com/CWBudde/go-dws-lsp/internal/server"
)

func TestCompletion_EmptyListForValidDocument(t *testing.T) {
	// Create a test server
	srv := server.New()
	SetServer(srv)

	// Sample DWScript source code
	source := `program Test;

var x: Integer;

begin
  x := 42;
end.`

	// Add document to server
	uri := "file:///test.dws"
	program, _, err := analysis.ParseDocument(source, uri)
	if err != nil {
		t.Fatalf("Failed to parse document: %v", err)
	}

	doc := &server.Document{
		URI:        uri,
		Text:       source,
		Version:    1,
		LanguageID: "dwscript",
		Program:    program,
	}
	srv.Documents().Set(uri, doc)

	// Create completion params
	params := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      4, // Inside the begin/end block
				Character: 2,
			},
		},
	}

	// Call Completion handler
	ctx := &glsp.Context{}
	result, err := Completion(ctx, params)

	// Should return without error
	if err != nil {
		t.Fatalf("Completion returned error: %v", err)
	}

	// Should return a CompletionList (may be empty for now since we haven't implemented item collection)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Result should be a CompletionList
	completionList, ok := result.(*protocol.CompletionList)
	if !ok {
		t.Fatalf("Expected CompletionList, got %T", result)
	}

	// For now, we expect an empty list (task 9.1 just sets up the handler structure)
	if completionList == nil {
		t.Fatal("Expected non-nil CompletionList")
	}

	t.Logf("Completion returned %d items (expected 0 for now)", len(completionList.Items))
}

func TestCompletion_NonExistentDocument(t *testing.T) {
	// Create a test server
	srv := server.New()
	SetServer(srv)

	// Create completion params for non-existent document
	params := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file:///nonexistent.dws",
			},
			Position: protocol.Position{
				Line:      0,
				Character: 0,
			},
		},
	}

	// Call Completion handler
	ctx := &glsp.Context{}
	result, err := Completion(ctx, params)

	// Should return nil without error (graceful handling)
	if err != nil {
		t.Fatalf("Completion returned error: %v", err)
	}

	if result != nil {
		t.Fatalf("Expected nil result for non-existent document, got %v", result)
	}
}

func TestCompletion_DocumentWithParseErrors(t *testing.T) {
	// Create a test server
	srv := server.New()
	SetServer(srv)

	// Invalid DWScript source code (will have parse errors)
	source := `program Test;
var x Integer; // Missing colon
begin
  x := ;
end.`

	// Add document to server
	uri := "file:///test_invalid.dws"
	// Parse the document (may fail due to errors)
	program, _, _ := analysis.ParseDocument(source, uri)

	doc := &server.Document{
		URI:        uri,
		Text:       source,
		Version:    1,
		LanguageID: "dwscript",
		Program:    program, // May be nil if parsing failed
	}
	srv.Documents().Set(uri, doc)

	// Create completion params
	params := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      2,
				Character: 2,
			},
		},
	}

	// Call Completion handler
	ctx := &glsp.Context{}
	result, err := Completion(ctx, params)

	// Should return empty list without error (graceful handling)
	if err != nil {
		t.Fatalf("Completion returned error: %v", err)
	}

	// Even with parse errors, we should return an empty CompletionList
	if result == nil {
		t.Fatal("Expected non-nil result even with parse errors")
	}

	completionList, ok := result.(*protocol.CompletionList)
	if !ok {
		t.Fatalf("Expected CompletionList, got %T", result)
	}

	if len(completionList.Items) != 0 {
		t.Fatalf("Expected empty completion list for document with parse errors, got %d items", len(completionList.Items))
	}
}
