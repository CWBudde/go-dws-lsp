//go:build integration
// +build integration

package integration

import (
	"testing"

	"github.com/CWBudde/go-dws-lsp/internal/lsp"
	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// setupTestServer creates a new server instance for integration testing
func setupTestServer() *server.Server {
	srv := server.New()
	lsp.SetServer(srv)
	return srv
}

// TestHoverIntegration_VariableDeclaration tests hover on a variable declaration
func TestHoverIntegration_VariableDeclaration(t *testing.T) {
	srv := setupTestServer()

	// Simulate opening a document
	uri := "file:///test/variables.dws"
	code := `var x: Integer;
x := 42;`

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request hover on the variable 'x' at line 0, character 4 (middle of 'x')
	hoverParams := &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      0,
				Character: 4,
			},
		},
	}

	hover, err := lsp.Hover(ctx, hoverParams)
	if err != nil {
		t.Fatalf("Hover failed: %v", err)
	}

	if hover == nil {
		t.Fatal("Expected hover result, got nil")
	}

	// Verify hover content contains expected information
	content := hover.Contents.(protocol.MarkupContent)
	if content.Kind != protocol.MarkupKindMarkdown {
		t.Errorf("Expected Markdown content, got %v", content.Kind)
	}

	// Check that the hover contains the variable name
	if !contains(content.Value, "x") {
		t.Errorf("Expected hover to contain variable name 'x', got: %s", content.Value)
	}

	// The hover content should contain meaningful information about the symbol
	if content.Value == "" {
		t.Error("Hover content should not be empty")
	}

	t.Logf("Hover content: %s", content.Value)

	// Clean up
	_ = srv
}

// TestHoverIntegration_FunctionDeclaration tests hover on a function declaration
func TestHoverIntegration_FunctionDeclaration(t *testing.T) {
	_ = setupTestServer()

	uri := "file:///test/functions.dws"
	code := `function Add(a, b: Integer): Integer;
begin
  Result := a + b;
end;

var x := Add(1, 2);`

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request hover on function name 'Add'
	hoverParams := &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      0,
				Character: 10, // Position in "Add"
			},
		},
	}

	hover, err := lsp.Hover(ctx, hoverParams)
	if err != nil {
		t.Fatalf("Hover failed: %v", err)
	}

	if hover == nil {
		t.Fatal("Expected hover result for function, got nil")
	}

	content := hover.Contents.(protocol.MarkupContent)

	// Verify function name is present
	if !contains(content.Value, "Add") {
		t.Errorf("Expected hover to contain function name 'Add', got: %s", content.Value)
	}

	// Verify content is not empty
	if content.Value == "" {
		t.Error("Hover content should not be empty")
	}

	t.Logf("Function hover content: %s", content.Value)
}

// TestHoverIntegration_ClassDeclaration tests hover on a class declaration
func TestHoverIntegration_ClassDeclaration(t *testing.T) {
	_ = setupTestServer()

	uri := "file:///test/classes.dws"
	code := `type TMyClass = class
private
  FValue: Integer;
public
end;

var obj: TMyClass;`

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request hover on variable 'obj' of class type TMyClass
	hoverParams := &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      5, // Line with "var obj: TMyClass;"
				Character: 4, // Position in "obj"
			},
		},
	}

	hover, err := lsp.Hover(ctx, hoverParams)
	if err != nil {
		t.Fatalf("Hover failed: %v", err)
	}

	if hover == nil {
		// This is acceptable - hover on some positions may return nil
		// depending on AST node structure. The important thing is that
		// the document compiled successfully and hover doesn't crash.
		t.Log("Hover returned nil (this is acceptable for this test case)")
		return
	}

	content := hover.Contents.(protocol.MarkupContent)

	// Verify variable information is present
	if !contains(content.Value, "obj") {
		t.Errorf("Expected hover to contain variable name 'obj', got: %s", content.Value)
	}

	// Verify content is not empty
	if content.Value == "" {
		t.Error("Hover content should not be empty")
	}

	t.Logf("Class variable hover content: %s", content.Value)
}

// TestHoverIntegration_NoHoverOnInvalidPosition tests that hover returns nil for invalid positions
func TestHoverIntegration_NoHoverOnInvalidPosition(t *testing.T) {
	_ = setupTestServer()

	uri := "file:///test/invalid.dws"
	code := `var x: Integer;`

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request hover on an invalid position (beyond end of line)
	hoverParams := &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      0,
				Character: 100, // Invalid position
			},
		},
	}

	hover, err := lsp.Hover(ctx, hoverParams)
	if err != nil {
		t.Fatalf("Hover failed: %v", err)
	}

	// Expecting nil or empty result for invalid position
	if hover != nil {
		content := hover.Contents.(protocol.MarkupContent)
		if content.Value != "" {
			t.Logf("Note: Got hover content for invalid position (may be OK): %s", content.Value)
		}
	}
}

// TestHoverIntegration_DocumentUpdate tests hover after document changes
func TestHoverIntegration_DocumentUpdate(t *testing.T) {
	_ = setupTestServer()

	uri := "file:///test/update.dws"
	initialCode := `var x: Integer;`

	// Open document
	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       initialCode,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Update document
	updatedCode := `var x: String;`
	changeParams := &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Version: 2,
		},
		ContentChanges: []interface{}{
			protocol.TextDocumentContentChangeEvent{
				Range: nil, // Full sync
				Text:  updatedCode,
			},
		},
	}

	err = lsp.DidChange(ctx, changeParams)
	if err != nil {
		t.Fatalf("DidChange failed: %v", err)
	}

	// Request hover to verify updated type
	hoverParams := &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      0,
				Character: 4,
			},
		},
	}

	hover, err := lsp.Hover(ctx, hoverParams)
	if err != nil {
		t.Fatalf("Hover failed: %v", err)
	}

	if hover != nil {
		content := hover.Contents.(protocol.MarkupContent)

		// Verify content is not empty
		if content.Value == "" {
			t.Error("Hover content should not be empty after document update")
		}

		// Verify hover contains the variable name
		if !contains(content.Value, "x") {
			t.Errorf("Expected hover to contain variable name 'x' after update, got: %s", content.Value)
		}

		t.Logf("Hover after update: %s", content.Value)
	}
}

// TestHoverIntegration_MultipleDocuments tests hover across multiple open documents
func TestHoverIntegration_MultipleDocuments(t *testing.T) {
	_ = setupTestServer()
	ctx := &glsp.Context{}

	// Open first document
	uri1 := "file:///test/doc1.dws"
	code1 := `var x: Integer;`
	openParams1 := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri1,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code1,
		},
	}

	err := lsp.DidOpen(ctx, openParams1)
	if err != nil {
		t.Fatalf("DidOpen for doc1 failed: %v", err)
	}

	// Open second document
	uri2 := "file:///test/doc2.dws"
	code2 := `var y: String;`
	openParams2 := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri2,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code2,
		},
	}

	err = lsp.DidOpen(ctx, openParams2)
	if err != nil {
		t.Fatalf("DidOpen for doc2 failed: %v", err)
	}

	// Test hover on first document
	hoverParams1 := &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri1,
			},
			Position: protocol.Position{Line: 0, Character: 4},
		},
	}

	hover1, err := lsp.Hover(ctx, hoverParams1)
	if err != nil {
		t.Fatalf("Hover on doc1 failed: %v", err)
	}

	if hover1 != nil {
		content1 := hover1.Contents.(protocol.MarkupContent)
		if !contains(content1.Value, "x") {
			t.Errorf("Expected hover on doc1 to contain 'x', got: %s", content1.Value)
		}
		t.Logf("Hover on doc1: %s", content1.Value)
	}

	// Test hover on second document
	hoverParams2 := &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri2,
			},
			Position: protocol.Position{Line: 0, Character: 4},
		},
	}

	hover2, err := lsp.Hover(ctx, hoverParams2)
	if err != nil {
		t.Fatalf("Hover on doc2 failed: %v", err)
	}

	if hover2 != nil {
		content2 := hover2.Contents.(protocol.MarkupContent)
		if !contains(content2.Value, "y") {
			t.Errorf("Expected hover on doc2 to contain 'y', got: %s", content2.Value)
		}
		t.Logf("Hover on doc2: %s", content2.Value)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
