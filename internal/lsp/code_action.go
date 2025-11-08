// Package lsp implements LSP protocol handlers.
package lsp

import (
	"log"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// CodeAction handles the textDocument/codeAction request.
// This provides quick fixes and refactoring actions for diagnostics and code.
func CodeAction(context *glsp.Context, params *protocol.CodeActionParams) (any, error) {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in CodeAction")
		return nil, nil
	}

	// Extract document URI, range, and context from params
	uri := params.TextDocument.URI
	selectedRange := params.Range
	actionContext := params.Context

	log.Printf("CodeAction request at %s range (%d:%d)-(%d:%d)\n",
		uri,
		selectedRange.Start.Line, selectedRange.Start.Character,
		selectedRange.End.Line, selectedRange.End.Character)

	// Get diagnostics from params.Context.Diagnostics
	diagnostics := actionContext.Diagnostics
	log.Printf("CodeAction context has %d diagnostics\n", len(diagnostics))

	// Retrieve document from DocumentStore
	doc, exists := srv.Documents().Get(uri)
	if !exists {
		log.Printf("Document not found for code action: %s\n", uri)
		return nil, nil
	}

	// Check if document has AST available
	if doc.Program == nil {
		log.Printf("No AST available for code action (document has parse errors): %s\n", uri)
		// Even without AST, we can still provide some code actions based on diagnostics
		// For now, return empty array
		return []protocol.CodeAction{}, nil
	}

	// Get AST from Program
	programAST := doc.Program.AST()
	if programAST == nil {
		log.Printf("AST is nil for document: %s\n", uri)
		return []protocol.CodeAction{}, nil
	}

	// Call helper functions to generate code actions
	var actions []protocol.CodeAction

	// TODO: Generate code actions based on:
	// 1. Diagnostics (quick fixes)
	// 2. Code context (refactoring actions)
	// 3. Selected range (extract method, etc.)

	// For now, return empty array (will be populated in future tasks)
	log.Printf("Returning %d code actions\n", len(actions))
	return actions, nil
}
