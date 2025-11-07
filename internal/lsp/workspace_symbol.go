// Package lsp implements LSP protocol handlers.
package lsp

import (
	"log"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// WorkspaceSymbol handles the workspace/symbol request.
// It returns symbols across the entire workspace that match the query string.
func WorkspaceSymbol(context *glsp.Context, params *protocol.WorkspaceSymbolParams) ([]protocol.SymbolInformation, error) {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in WorkspaceSymbol")
		return nil, nil
	}

	// Extract query string from params
	query := params.Query
	log.Printf("WorkspaceSymbol request with query: %q\n", query)

	// Access workspace symbol index
	index := srv.WorkspaceIndex()
	if index == nil {
		log.Println("Warning: workspace index not available")
		return nil, nil
	}

	// Search for symbols matching the query
	// Limit to 500 results to avoid overwhelming the client
	const maxResults = 500
	symbolLocations := index.Search(query, maxResults)

	log.Printf("Found %d workspace symbols matching query %q\n", len(symbolLocations), query)

	// Convert to SymbolInformation array
	var symbols []protocol.SymbolInformation
	for _, symLoc := range symbolLocations {
		// Build SymbolInformation
		symbolInfo := protocol.SymbolInformation{
			Name:     symLoc.Name,
			Kind:     symLoc.Kind,
			Location: symLoc.Location,
		}

		// Add container name if available
		if symLoc.ContainerName != "" {
			symbolInfo.ContainerName = &symLoc.ContainerName
		}

		symbols = append(symbols, symbolInfo)
	}

	return symbols, nil
}
