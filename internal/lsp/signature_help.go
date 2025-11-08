package lsp

import (
	"log"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// SignatureHelp handles textDocument/signatureHelp requests
// Shows function signatures and parameter hints during function calls
func SignatureHelp(context *glsp.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in SignatureHelp")
		return nil, nil
	}

	// Extract document URI and position from params
	uri := params.TextDocument.URI
	position := params.Position

	log.Printf("SignatureHelp request: URI=%s, Line=%d, Character=%d\n", uri, position.Line, position.Character)

	// Retrieve document from DocumentStore
	doc, exists := srv.Documents().Get(uri)
	if !exists {
		log.Printf("Document not found: %s\n", uri)
		return nil, nil
	}

	// Check if document and AST are available
	if doc.Program == nil {
		log.Printf("No program available for document: %s\n", uri)
		return nil, nil
	}

	programAST := doc.Program.AST()
	if programAST == nil {
		log.Printf("No AST available for document: %s\n", uri)
		return nil, nil
	}

	// Convert LSP position (0-based, UTF-16) to document position (1-based, UTF-8)
	astLine := int(position.Line) + 1
	astColumn := int(position.Character) + 1

	log.Printf("Converted position: line=%d, column=%d\n", astLine, astColumn)

	// TODO: Call helper function to compute signature help (will be implemented in tasks 10.3-10.8)
	// For now, return nil to indicate no signature help available
	// This will be replaced with actual implementation that:
	// - Determines call context from cursor position (task 10.3)
	// - Detects signature help triggers (task 10.4)
	// - Finds function being called (task 10.5)
	// - Handles incomplete AST (task 10.6)
	// - Traverses tokens to count commas (task 10.7)
	// - Retrieves function definition (task 10.8)
	// - Handles built-in functions (task 10.9)
	// - Constructs SignatureHelp response (task 10.10)

	return nil, nil
}
