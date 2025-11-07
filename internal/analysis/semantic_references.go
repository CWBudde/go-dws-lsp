// Package analysis provides semantic analysis utilities.
package analysis

import (
	"log"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/dwscript"
	"github.com/cwbudde/go-dws/pkg/token"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// FindSymbolDefinitionAtPosition finds the symbol definition at the given position
// using both AST and semantic information. Returns nil if no symbol is found at that position.
func FindSymbolDefinitionAtPosition(program *dwscript.Program, pos token.Position) *dwscript.Symbol {
	if program == nil {
		return nil
	}

	programAST := program.AST()
	if programAST == nil {
		return nil
	}

	// First, find the identifier at this position in the AST
	identSym := IdentifySymbolAtPosition(programAST, pos.Line, pos.Column)
	if identSym == nil || identSym.Name == "" {
		log.Printf("No identifier at position %d:%d", pos.Line, pos.Column)
		return nil
	}

	// Now look up this symbol in the semantic symbol table
	symbols := program.Symbols()
	if len(symbols) == 0 {
		log.Println("No symbols available from semantic analyzer")
		return nil
	}

	// Find the semantic symbol that matches this identifier
	// We match by name and kind
	for i := range symbols {
		sym := &symbols[i]
		if sym.Name == identSym.Name {
			// For now, we match by name only since position info from semantic analyzer
			// is not always reliable (often returns 0:0)
			log.Printf("Found semantic symbol: %s (kind=%s, scope=%s)",
				sym.Name, sym.Kind, sym.Scope)
			return sym
		}
	}

	log.Printf("No semantic symbol found for '%s' at position %d:%d", identSym.Name, pos.Line, pos.Column)
	return nil
}

// ResolveIdentifierToDefinition attempts to resolve an identifier at a given position
// to its symbol definition using the semantic analyzer and symbol resolver.
// Returns the definition position if found, nil otherwise.
func ResolveIdentifierToDefinition(
	program *dwscript.Program,
	uri string,
	identifierName string,
	identifierPos token.Position,
) *token.Position {
	if program == nil {
		return nil
	}

	// Use the symbol resolver to find where this identifier is defined
	resolver := NewSymbolResolver(uri, program.AST(), identifierPos)
	locations := resolver.ResolveSymbol(identifierName)

	if len(locations) == 0 {
		return nil
	}

	// Convert the first location to a token.Position
	// (we take the first one as it's the most relevant based on scope rules)
	loc := locations[0]
	defPos := token.Position{
		Line:   int(loc.Range.Start.Line) + 1,   // Convert from 0-based to 1-based
		Column: int(loc.Range.Start.Character) + 1, // Convert from 0-based to 1-based
	}

	return &defPos
}

// FindSemanticReferences finds all references to a symbol using semantic analysis.
// It matches identifiers by their resolved definition location, not just by name.
// This provides accurate filtering and avoids false positives.
//
// Parameters:
//   - program: The compiled program with semantic information
//   - targetName: The name of the symbol we're finding references for
//   - targetPos: The position of the symbol we're finding references for
//   - uri: The document URI
//
// Returns a list of locations where the symbol is referenced.
func FindSemanticReferences(
	program *dwscript.Program,
	targetName string,
	targetPos token.Position,
	uri string,
) []protocol.Range {
	if program == nil {
		return nil
	}

	programAST := program.AST()
	if programAST == nil {
		return nil
	}

	// First, resolve the target position to its definition location
	// This handles cases where the cursor is on a reference, not the definition
	targetDefPos := ResolveIdentifierToDefinition(program, uri, targetName, targetPos)
	if targetDefPos == nil {
		log.Printf("Could not resolve target '%s' at %d:%d to definition", targetName, targetPos.Line, targetPos.Column)
		return nil
	}

	var ranges []protocol.Range
	log.Printf("Finding semantic references for %s defined at %d:%d",
		targetName, targetDefPos.Line, targetDefPos.Column)

	// Traverse the AST to find all identifiers
	ast.Inspect(programAST, func(n ast.Node) bool {
		ident, ok := n.(*ast.Identifier)
		if !ok || ident == nil {
			return true
		}

		// Quick filter: only consider identifiers with matching name
		if ident.Value != targetName {
			return true
		}

		identPos := ident.Pos()

		// Try to resolve this identifier to its definition
		defPos := ResolveIdentifierToDefinition(program, uri, ident.Value, identPos)
		if defPos == nil {
			// Could not resolve - skip this identifier
			// (it might be in an error state or unparseable context)
			return true
		}

		// Check if the resolved definition matches our target's definition
		if defPos.Line == targetDefPos.Line && defPos.Column == targetDefPos.Column {
			// This identifier resolves to our target symbol - it's a reference!
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

			ranges = append(ranges, lspRange)
			log.Printf("  Found reference at %d:%d", identPos.Line, identPos.Column)
		}

		return true
	})

	log.Printf("Found %d semantic references for %s", len(ranges), targetName)
	return ranges
}
