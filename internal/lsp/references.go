// Package lsp implements LSP protocol handlers.
package lsp

import (
    "log"

    "github.com/cwbudde/go-dws/pkg/ast"
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

    // Find the node at the given position
    node := analysis.FindNodeAtPosition(programAST, astLine, astColumn)
    if node == nil {
        log.Printf("No AST node found at position %d:%d for references\n", astLine, astColumn)
        return []protocol.Location{}, nil
    }

    // Determine target symbol name (best-effort)
    targetName := analysis.GetSymbolName(node)
    if targetName == "" {
        // If the node itself is an Identifier, fall back to its value
        if ident, ok := node.(*ast.Identifier); ok && ident != nil {
            targetName = ident.Value
        }
    }
    if targetName == "" {
        // Not on a symbol we recognize
        return []protocol.Location{}, nil
    }

    // Naive in-document scan for identifiers with the same name.
    // Scope and cross-document handling will be added in later tasks.
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

