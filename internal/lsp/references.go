// Package lsp implements LSP protocol handlers.
package lsp

import (
	"log"
	"sort"

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

	log.Printf("References request at %s line %d, character %d (includeDeclaration=%t)\n",
		uri, position.Line, position.Character, includeDecl)

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

	// Find the definition location (Task 6.11) to support includeDeclaration flag
	defLocation := resolveSymbolDefinition(uri, targetName, astLine, astColumn, programAST)
	if defLocation != nil {
		log.Printf("Found definition at %s line %d, character %d",
			defLocation.URI, defLocation.Range.Start.Line, defLocation.Range.Start.Character)
	}

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
		// Apply includeDeclaration flag (task 6.11)
		locations = applyIncludeDeclaration(locations, defLocation, includeDecl)
		// Sort by file then position (task 6.10)
		sortLocationsByFileAndPosition(locations)
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
		// Apply includeDeclaration flag (task 6.11)
		locations = applyIncludeDeclaration(locations, defLocation, includeDecl)
		// Sort by file then position (task 6.10)
		sortLocationsByFileAndPosition(locations)
		return locations, nil
	}

	// For global symbols, search across all open documents
	if scope != nil && scope.Type == analysis.ScopeGlobal {
		openLocations := analysis.FindGlobalReferences(targetName, srv.Documents())
		indexLocations := []protocol.Location{}
		if srv.Symbols() != nil {
			indexLocations = srv.Symbols().FindReferences(targetName, srv.Documents())
		}

		combined := append([]protocol.Location{}, openLocations...)
		combined = append(combined, indexLocations...)

		filtered := analysis.FilterByScope(combined, srv.Documents(), targetName, scope)
		// Apply includeDeclaration flag (task 6.11)
		filtered = applyIncludeDeclaration(filtered, defLocation, includeDecl)
		// Sort by file then position (task 6.10)
		sortLocationsByFileAndPosition(filtered)
		return filtered, nil
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

	filtered := analysis.FilterByScope(locations, srv.Documents(), targetName, scope)
	// Apply includeDeclaration flag (task 6.11)
	filtered = applyIncludeDeclaration(filtered, defLocation, includeDecl)
	// Sort by file then position (task 6.10)
	sortLocationsByFileAndPosition(filtered)
	return filtered, nil
}

// max returns the larger of a or b.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// sortLocationsByFileAndPosition sorts locations by file (URI) then by position (line, then character).
// This ensures consistent ordering of reference results as specified in task 6.10.
func sortLocationsByFileAndPosition(locations []protocol.Location) {
	sort.Slice(locations, func(i, j int) bool {
		// First, compare by URI (file)
		if locations[i].URI != locations[j].URI {
			return locations[i].URI < locations[j].URI
		}
		// Same file, compare by line
		if locations[i].Range.Start.Line != locations[j].Range.Start.Line {
			return locations[i].Range.Start.Line < locations[j].Range.Start.Line
		}
		// Same line, compare by character
		return locations[i].Range.Start.Character < locations[j].Range.Start.Character
	})
}

// resolveSymbolDefinition finds the definition location for a symbol at the given position.
// This reuses the go-to-definition logic to locate where the symbol is defined.
// Returns nil if no definition is found.
func resolveSymbolDefinition(uri string, symbolName string, astLine, astColumn int, programAST *ast.Program) *protocol.Location {
	// Use the symbol resolver to find the definition (same logic as Definition handler)
	resolver := analysis.NewSymbolResolver(uri, programAST, token.Position{
		Line:   astLine,
		Column: astColumn,
	})

	locations := resolver.ResolveSymbol(symbolName)
	if len(locations) == 0 {
		return nil
	}

	// Return the first location (most relevant based on scope rules)
	return &locations[0]
}

// applyIncludeDeclaration adds or removes the definition location from the results
// based on the includeDeclaration flag.
// If includeDeclaration is true, the definition is inserted at the beginning (conventional).
// If false, the definition is removed from the results if present.
func applyIncludeDeclaration(locations []protocol.Location, defLocation *protocol.Location, includeDeclaration bool) []protocol.Location {
	if defLocation == nil {
		// No definition found, return locations as-is
		return locations
	}

	// Check if the definition is already in the results
	defIndex := -1
	for i, loc := range locations {
		if loc.URI == defLocation.URI &&
			loc.Range.Start.Line == defLocation.Range.Start.Line &&
			loc.Range.Start.Character == defLocation.Range.Start.Character {
			defIndex = i
			break
		}
	}

	if includeDeclaration {
		// If definition should be included but isn't in the list, add it at the beginning
		if defIndex == -1 {
			result := make([]protocol.Location, 0, len(locations)+1)
			result = append(result, *defLocation)
			result = append(result, locations...)
			return result
		}
		// Definition is already in the list
		// Move it to the beginning if it's not already there
		if defIndex > 0 {
			result := make([]protocol.Location, len(locations))
			result[0] = locations[defIndex]
			copy(result[1:defIndex+1], locations[0:defIndex])
			copy(result[defIndex+1:], locations[defIndex+1:])
			return result
		}
		// Definition is already at the beginning
		return locations
	}

	// includeDeclaration is false: remove definition if present
	if defIndex >= 0 {
		result := make([]protocol.Location, 0, len(locations)-1)
		result = append(result, locations[:defIndex]...)
		result = append(result, locations[defIndex+1:]...)
		return result
	}

	// Definition not in list, return as-is
	return locations
}
