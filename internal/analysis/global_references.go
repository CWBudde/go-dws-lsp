package analysis

import (
	"log"

	"github.com/cwbudde/go-dws/pkg/ast"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// FindGlobalReferences scans all open documents for references to the given symbol name.
// This best-effort implementation traverses each document's AST and collects locations
// of identifiers that match the target name. Future tasks (6.8+, 6.9) will refine scope
// filtering and semantic resolution to avoid false positives.
func FindGlobalReferences(symbolName string, docStore *server.DocumentStore) []protocol.Location {
	if symbolName == "" || docStore == nil {
		return nil
	}

	uris := docStore.List()
	var locations []protocol.Location

	for _, uri := range uris {
		doc, ok := docStore.Get(uri)
		if !ok || doc == nil || doc.Program == nil {
			continue
		}

		programAST := doc.Program.AST()
		if programAST == nil {
			continue
		}

		log.Printf("Scanning %s for global references to %s", uri, symbolName)

		ast.Inspect(programAST, func(node ast.Node) bool {
			ident, ok := node.(*ast.Identifier)
			if !ok || ident == nil {
				return true
			}
			if ident.Value != symbolName {
				return true
			}

			start := ident.Pos()
			end := ident.End()

			loc := protocol.Location{
				URI: uri,
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      uint32(max(0, start.Line-1)),
						Character: uint32(max(0, start.Column-1)),
					},
					End: protocol.Position{
						Line:      uint32(max(0, end.Line-1)),
						Character: uint32(max(0, end.Column-1)),
					},
				},
			}

			locations = append(locations, loc)
			return true
		})
	}

	return locations
}
