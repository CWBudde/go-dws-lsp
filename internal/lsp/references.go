// Package lsp implements LSP protocol handlers.
package lsp

import (
	"log"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/analysis"
	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// References handles the textDocument/references request.
// It returns locations of all references to the symbol at the given position.
//
// Note: This initial implementation performs a simple in-document identifier
// name match. Future tasks will refine this to incorporate scope and
// cross-document analysis per PLAN.md (Phases 6.2+).
func References(context *glsp.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in References")
		return []protocol.Location{}, nil
	}

	// Extract request details
	uri := params.TextDocument.URI
	position := params.Position
	includeDecl := params.Context.IncludeDeclaration
	_ = includeDecl // Will be used in future tasks for filtering

	log.Printf("References request at %s line %d, character %d\n",
		uri, position.Line, position.Character)

	// Retrieve document from store
	doc, exists := srv.Documents().Get(uri)
	if !exists {
		log.Printf("Document not found for references: %s\n", uri)
		return []protocol.Location{}, nil
	}

	// Ensure we have a parsed program/AST
	if doc.Program == nil || doc.Program.AST() == nil {
		log.Printf("No AST available for references (document has parse errors): %s\n", uri)
		return []protocol.Location{}, nil
	}

	// Convert LSP (0-based) to AST (1-based)
	astLine := int(position.Line) + 1
	astColumn := int(position.Character) + 1

	programAST := doc.Program.AST()

	// Identify symbol at position (name + kind)
	sym := analysis.IdentifySymbolAtPosition(programAST, astLine, astColumn)
	if sym == nil || sym.Name == "" {
		log.Printf("No symbol at position %d:%d for references\n", astLine, astColumn)
		return []protocol.Location{}, nil
	}
	targetName := sym.Name
	log.Printf("References target symbol: %s (kind=%s)", sym.Name, sym.Kind)

	// Try to use semantic analyzer for accurate symbol resolution (Task 6.9)
	// Convert LSP position to token.Position (1-based)
	tokenPos := token.Position{Line: astLine, Column: astColumn}

	// Try to find references using semantic analysis (matching by definition location)
	ranges := analysis.FindSemanticReferences(doc.Program, targetName, tokenPos, uri)

	if len(ranges) > 0 {
		// Semantic analysis succeeded
		log.Printf("Using semantic analysis for references: found %d references", len(ranges))
		locations := make([]protocol.Location, 0, len(ranges))
		for _, r := range ranges {
			locations = append(locations, protocol.Location{URI: uri, Range: r})
		}
		return locations, nil
	}

	// Fallback: semantic info not available, use name-based matching with scope filtering
	log.Printf("Semantic info unavailable, falling back to name-based matching")

	// Determine basic scope context for logging and future filtering
	scope := analysis.DetermineScope(programAST, sym.Name, analysis.Position{Line: astLine, Column: astColumn})
	if scope != nil {
		log.Printf("References scope: %s\n", scope.Type)
	}

	// If scope is local or parameter, limit search to the same function/block
	if scope != nil && (scope.Type == analysis.ScopeLocal || scope.Type == analysis.ScopeParameter) && scope.Function != nil {
		ranges := analysis.FindLocalReferences(programAST, targetName, scope.Function)
		locations := make([]protocol.Location, 0, len(ranges))
		for _, r := range ranges {
			locations = append(locations, protocol.Location{URI: uri, Range: r})
		}
		return locations, nil
	}

	// For global symbols, search across all open documents
	if scope != nil && scope.Type == analysis.ScopeGlobal {
		openLocations := analysis.FindGlobalReferences(targetName, srv.Documents())
		indexLocations := []protocol.Location{}
		if srv.Symbols() != nil {
			indexLocations = srv.Symbols().FindReferences(targetName, srv.Documents())
		}

		if len(indexLocations) == 0 {
			return openLocations, nil
		}

		combined := make([]protocol.Location, 0, len(openLocations)+len(indexLocations))
		combined = append(combined, openLocations...)
		combined = append(combined, indexLocations...)
		return combined, nil
	}

	// Otherwise, do a naive in-document scan for identifiers with the same name.
	// Cross-document and scope-aware filtering will be added in later tasks.
	var locations []protocol.Location

	ast.Inspect(programAST, func(n ast.Node) bool {
		ident, ok := n.(*ast.Identifier)
		if !ok || ident == nil {
			return true
		}
		if ident.Value != targetName {
			return true
		}

		// Convert AST positions (1-based) to LSP positions (0-based)
		start := ident.Pos()
		end := ident.End()

		lspRange := protocol.Range{
			Start: protocol.Position{
				Line:      uint32(max(0, start.Line-1)),
				Character: uint32(max(0, start.Column-1)),
			},
			End: protocol.Position{
				Line:      uint32(max(0, end.Line-1)),
				Character: uint32(max(0, end.Column-1)),
			},
		}

		locations = append(locations, protocol.Location{
			URI:   uri,
			Range: lspRange,
		})

		return true
	})

	return locations, nil
}

// max returns the larger of a or b.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
