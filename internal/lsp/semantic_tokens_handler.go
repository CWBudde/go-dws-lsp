// Package lsp implements semantic tokens LSP handler.
package lsp

import (
	"log"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/analysis"
	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// SemanticTokensFull handles textDocument/semanticTokens/full requests.
// It returns semantic highlighting information for the entire document.
func SemanticTokensFull(context *glsp.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	log.Printf("SemanticTokensFull request for: %s\n", params.TextDocument.URI)

	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Error: server instance not available")
		return nil, nil
	}

	// Get document from store
	doc, ok := srv.Documents().Get(string(params.TextDocument.URI))
	if !ok || doc == nil {
		log.Printf("Document not found: %s\n", params.TextDocument.URI)
		return nil, nil
	}

	// Check if document has a valid program
	program := doc.Program
	if program == nil {
		log.Printf("Document has no program: %s\n", params.TextDocument.URI)
		return nil, nil
	}

	// Get the AST
	ast := program.AST()
	if ast == nil {
		log.Printf("Document AST is nil: %s\n", params.TextDocument.URI)
		return nil, nil
	}

	// Get the semantic tokens legend
	legend := srv.SemanticTokensLegend()
	if legend == nil {
		log.Println("Error: semantic tokens legend not available")
		return nil, nil
	}

	// Collect semantic tokens from AST
	tokens, err := analysis.CollectSemanticTokens(ast, legend)
	if err != nil {
		log.Printf("Error collecting semantic tokens: %v\n", err)
		return nil, nil
	}

	// Encode tokens in LSP delta format
	data := analysis.EncodeSemanticTokens(tokens)

	log.Printf("Collected %d semantic tokens for %s\n", len(tokens), params.TextDocument.URI)

	// Return semantic tokens response
	return &protocol.SemanticTokens{
		Data: data,
	}, nil
}
