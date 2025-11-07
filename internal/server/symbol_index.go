package server

import (
	"sync"

	"github.com/cwbudde/go-dws/pkg/ast"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type SymbolIndex struct {
	mu         sync.RWMutex
	references map[string]map[string][]protocol.Range
}

func NewSymbolIndex() *SymbolIndex {
	return &SymbolIndex{
		references: make(map[string]map[string][]protocol.Range),
	}
}

func (si *SymbolIndex) UpdateDocument(doc *Document) {
	if doc == nil {
		return
	}
	if doc.Program == nil || doc.Program.AST() == nil {
		si.RemoveDocument(doc.URI)
		return
	}

	ranges := collectReferences(doc.Program.AST())

	si.mu.Lock()
	defer si.mu.Unlock()

	// remove existing entries for doc
	for symbol, uris := range si.references {
		if _, ok := uris[doc.URI]; ok {
			delete(uris, doc.URI)
			if len(uris) == 0 {
				delete(si.references, symbol)
			}
		}
	}

	for symbol, list := range ranges {
		if len(list) == 0 {
			continue
		}
		if _, ok := si.references[symbol]; !ok {
			si.references[symbol] = make(map[string][]protocol.Range)
		}
		si.references[symbol][doc.URI] = list
	}
}

func (si *SymbolIndex) RemoveDocument(uri string) {
	if uri == "" {
		return
	}
	si.mu.Lock()
	defer si.mu.Unlock()
	for symbol, uris := range si.references {
		if _, ok := uris[uri]; ok {
			delete(uris, uri)
			if len(uris) == 0 {
				delete(si.references, symbol)
			}
		}
	}
}

func (si *SymbolIndex) FindReferences(symbolName string, docStore *DocumentStore) []protocol.Location {
	if symbolName == "" {
		return nil
	}

	openDocs := make(map[string]struct{})
	if docStore != nil {
		for _, uri := range docStore.List() {
			openDocs[uri] = struct{}{}
		}
	}

	si.mu.RLock()
	perURI, ok := si.references[symbolName]
	if !ok {
		si.mu.RUnlock()
		return nil
	}

	copied := make(map[string][]protocol.Range, len(perURI))
	for uri, ranges := range perURI {
		rangesCopy := make([]protocol.Range, len(ranges))
		copy(rangesCopy, ranges)
		copied[uri] = rangesCopy
	}
	si.mu.RUnlock()

	var locations []protocol.Location
	for uri, ranges := range copied {
		if _, open := openDocs[uri]; open {
			continue
		}
		for _, r := range ranges {
			locations = append(locations, protocol.Location{URI: uri, Range: r})
		}
	}

	return locations
}

func collectReferences(root ast.Node) map[string][]protocol.Range {
	result := make(map[string][]protocol.Range)
	if root == nil {
		return result
	}

	ast.Inspect(root, func(node ast.Node) bool {
		ident, ok := node.(*ast.Identifier)
		if !ok || ident == nil {
			return true
		}
		start := ident.Pos()
		end := ident.End()
		rng := protocol.Range{
			Start: protocol.Position{Line: uint32(maxZero(start.Line - 1)), Character: uint32(maxZero(start.Column - 1))},
			End:   protocol.Position{Line: uint32(maxZero(end.Line - 1)), Character: uint32(maxZero(end.Column - 1))},
		}
		result[ident.Value] = append(result[ident.Value], rng)
		return true
	})

	return result
}

func maxZero(val int) int {
	if val < 0 {
		return 0
	}
	return val
}
