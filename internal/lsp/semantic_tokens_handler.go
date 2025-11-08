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

	// Generate a resultId for delta support
	resultID := server.GenerateResultID(params.TextDocument.URI, doc.Version)

	// Store tokens in cache for future delta requests
	if cache := srv.SemanticTokensCache(); cache != nil {
		cache.Store(params.TextDocument.URI, resultID, tokens)
		log.Printf("Cached semantic tokens with resultId: %s\n", resultID)
	}

	// Encode tokens in LSP delta format
	data := analysis.EncodeSemanticTokens(tokens)

	log.Printf("Collected %d semantic tokens for %s\n", len(tokens), params.TextDocument.URI)

	// Return semantic tokens response with resultId for delta support
	return &protocol.SemanticTokens{
		ResultID: &resultID,
		Data:     data,
	}, nil
}

// SemanticTokensFullDelta handles textDocument/semanticTokens/full/delta requests.
// It returns incremental changes to semantic tokens since the previous request.
func SemanticTokensFullDelta(context *glsp.Context, params *protocol.SemanticTokensDeltaParams) (interface{}, error) {
	log.Printf("SemanticTokensFullDelta request for: %s (previousResultId: %s)\n",
		params.TextDocument.URI, *params.PreviousResultID)

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

	// Collect new semantic tokens from AST
	newTokens, err := analysis.CollectSemanticTokens(ast, legend)
	if err != nil {
		log.Printf("Error collecting semantic tokens: %v\n", err)
		return nil, nil
	}

	// Generate new resultId
	newResultID := server.GenerateResultID(params.TextDocument.URI, doc.Version)

	// Try to retrieve old tokens from cache
	var oldTokens []analysis.SemanticToken
	cache := srv.SemanticTokensCache()
	if cache != nil && params.PreviousResultID != nil {
		if cached, found := cache.Retrieve(params.TextDocument.URI, *params.PreviousResultID); found {
			oldTokens = cached.Tokens
			log.Printf("Found cached tokens for previousResultId: %s\n", *params.PreviousResultID)
		} else {
			log.Printf("Previous resultId not found in cache: %s\n", *params.PreviousResultID)
		}
	}

	// Compute delta or fallback to full
	result := analysis.ComputeSemanticTokensDelta(oldTokens, newTokens, newResultID)

	// Store new tokens in cache for future delta requests
	if cache != nil {
		cache.Store(params.TextDocument.URI, newResultID, newTokens)
		log.Printf("Cached new semantic tokens with resultId: %s\n", newResultID)
	}

	// Return either delta or full based on computation
	if result.IsDelta {
		log.Printf("Returning delta response with %d edits\n", len(result.Delta.Edits))
		return result.Delta, nil
	} else {
		log.Printf("Returning full response as fallback\n")
		return result.Full, nil
	}
}
