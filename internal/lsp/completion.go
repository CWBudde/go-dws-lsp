// Package lsp implements LSP protocol handlers.
package lsp

import (
	"log"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// Completion handles the textDocument/completion request.
// This provides intelligent code completion suggestions.
func Completion(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in Completion")
		return nil, nil
	}

	// Extract document URI and position from params
	uri := params.TextDocument.URI
	position := params.Position

	log.Printf("Completion request at %s line %d, character %d\n",
		uri, position.Line, position.Character)

	// Retrieve document from DocumentStore
	doc, exists := srv.Documents().Get(uri)
	if !exists {
		log.Printf("Document not found for completion: %s\n", uri)
		return nil, nil
	}

	// Check if document and AST are available
	if doc.Program == nil {
		log.Printf("No AST available for completion (document has parse errors): %s\n", uri)
		// Return empty completion list instead of nil to indicate completion is supported
		return &protocol.CompletionList{
			IsIncomplete: false,
			Items:        []protocol.CompletionItem{},
		}, nil
	}

	// Get AST from Program
	programAST := doc.Program.AST()
	if programAST == nil {
		log.Printf("AST is nil for document: %s\n", uri)
		return &protocol.CompletionList{
			IsIncomplete: false,
			Items:        []protocol.CompletionItem{},
		}, nil
	}

	// TODO: Task 9.2 - Determine completion context from cursor position
	// TODO: Task 9.3 - Detect trigger characters (dot for member access)
	// TODO: Task 9.4-9.6 - Handle member access completion
	// TODO: Task 9.7+ - Collect completion items based on context

	// For now, return an empty completion list
	// This will be enhanced in subsequent tasks (9.2+)
	completionList := &protocol.CompletionList{
		IsIncomplete: false,
		Items:        []protocol.CompletionItem{},
	}

	log.Printf("Returning %d completion items\n", len(completionList.Items))

	return completionList, nil
}
