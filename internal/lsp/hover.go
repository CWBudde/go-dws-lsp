// Package lsp implements LSP protocol handlers.
package lsp

import (
	"log"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// Hover handles the textDocument/hover request.
// This provides type and symbol information when the user hovers over code.
func Hover(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in Hover")
		return nil, nil
	}

	// Extract document URI and position from params
	uri := params.TextDocument.URI
	position := params.Position

	log.Printf("Hover request at %s line %d, character %d\n",
		uri, position.Line, position.Character)

	// Retrieve document from DocumentStore
	doc, exists := srv.Documents().Get(uri)
	if !exists {
		log.Printf("Document not found for hover: %s\n", uri)
		return nil, nil
	}

	// Check if document and AST are available
	if doc.Program == nil {
		log.Printf("No AST available for hover (document has parse errors): %s\n", uri)
		return nil, nil
	}

	// Get AST from Program
	ast := doc.Program.AST()
	if ast == nil {
		log.Printf("AST is nil for document: %s\n", uri)
		return nil, nil
	}

	// TODO: Convert LSP position (0-based, UTF-16) to AST position (1-based, UTF-8)
	// For now, just return a placeholder hover response to test the infrastructure

	// Create a simple placeholder hover response
	hoverText := "Hover support is being implemented"

	hover := &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: hoverText,
		},
	}

	return hover, nil
}
