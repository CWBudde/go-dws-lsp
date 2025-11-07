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

// Definition handles the textDocument/definition request.
// This provides "go-to definition" functionality, allowing users to navigate
// to where a symbol is defined.
func Definition(context *glsp.Context, params *protocol.DefinitionParams) (interface{}, error) {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in Definition")
		return nil, nil
	}

	// Extract document URI and position from params
	uri := params.TextDocument.URI
	position := params.Position

	log.Printf("Definition request at %s line %d, character %d\n",
		uri, position.Line, position.Character)

	// Retrieve document from DocumentStore
	doc, exists := srv.Documents().Get(uri)
	if !exists {
		log.Printf("Document not found for definition: %s\n", uri)
		return nil, nil
	}

	// Check if document and AST are available
	if doc.Program == nil {
		log.Printf("No AST available for definition (document has parse errors): %s\n", uri)
		return nil, nil
	}

	// Get AST from Program
	programAST := doc.Program.AST()
	if programAST == nil {
		log.Printf("AST is nil for document: %s\n", uri)
		return nil, nil
	}

	// Convert LSP position (0-based, UTF-16) to AST position (1-based, UTF-8)
	// LSP uses 0-based positions, AST uses 1-based
	astLine := int(position.Line) + 1
	astColumn := int(position.Character) + 1

	// Find the AST node at this position
	node := analysis.FindNodeAtPosition(programAST, astLine, astColumn)
	if node == nil {
		log.Printf("No AST node found at position %d:%d for definition\n", astLine, astColumn)
		return nil, nil
	}

	// Identify what symbol we're on (Task 5.2)
	symbolInfo := IdentifySymbolAtPosition(node)
	if symbolInfo == nil {
		log.Printf("Position %d:%d is not on a symbol\n", astLine, astColumn)
		return nil, nil
	}

	log.Printf("Symbol at position: %s (kind: %s)", symbolInfo.Name, symbolInfo.Kind)

	// Check if we're already on a declaration
	if IsDeclaration(node) {
		log.Printf("Already on declaration of %s, returning its location", symbolInfo.Name)
		// Return the declaration's own location
		return nodeToLocation(node, uri), nil
	}

	// Use the new symbol resolver for scope-based resolution (Task 5.3)
	resolver := analysis.NewSymbolResolver(uri, programAST, token.Position{
		Line:   astLine,
		Column: astColumn,
	})

	log.Printf("Resolution scope: %s", resolver.GetResolutionScope())

	locations := resolver.ResolveSymbol(symbolInfo.Name)
	if len(locations) == 0 {
		log.Printf("No definition found for symbol %s at position %d:%d\n", symbolInfo.Name, astLine, astColumn)
		return nil, nil
	}

	// Return the first location (or all locations if we support multiple definitions)
	if len(locations) == 1 {
		return &locations[0], nil
	}

	// Multiple definitions found (e.g., overloaded functions)
	log.Printf("Found %d definitions for symbol %s", len(locations), symbolInfo.Name)
	return locations, nil // Return array of locations
}

// findDefinitionLocation finds the definition location for an AST node.
// Returns a protocol.Location pointing to where the symbol is defined, or nil if not found.
func findDefinitionLocation(node ast.Node, doc *server.Document, programAST *ast.Program, currentURI string) *protocol.Location {
	// Handle different node types
	switch n := node.(type) {
	case *ast.Identifier:
		// For identifiers, we need to find where they're declared
		return findIdentifierDefinition(n, programAST, currentURI)

	case *ast.VarDeclStatement:
		// If clicking on a variable declaration, return its own location
		return nodeToLocation(n, currentURI)

	case *ast.FunctionDecl:
		// If clicking on a function declaration, return its own location
		return nodeToLocation(n, currentURI)

	case *ast.ClassDecl:
		// If clicking on a class declaration, return its own location
		return nodeToLocation(n, currentURI)

	case *ast.ConstDecl:
		// If clicking on a constant declaration, return its own location
		return nodeToLocation(n, currentURI)

	case *ast.EnumDecl:
		// If clicking on an enum declaration, return its own location
		return nodeToLocation(n, currentURI)

	default:
		// For other node types, no definition location available
		return nil
	}
}

// findIdentifierDefinition finds the definition of an identifier.
// This searches the AST for where the identifier was declared.
func findIdentifierDefinition(ident *ast.Identifier, programAST *ast.Program, uri string) *protocol.Location {
	identName := ident.Value

	// Search for the declaration in the program
	// We'll do a simple AST traversal to find declarations matching this name

	var foundLocation *protocol.Location

	// Traverse the AST looking for declarations
	ast.Inspect(programAST, func(node ast.Node) bool {
		if foundLocation != nil {
			return false // Already found, stop searching
		}

		switch decl := node.(type) {
		case *ast.VarDeclStatement:
			// Check if any of the variable names match
			for _, name := range decl.Names {
				if name.Value == identName {
					foundLocation = nodeToLocation(decl, uri)
					return false
				}
			}

		case *ast.FunctionDecl:
			// Check function name
			if decl.Name != nil && decl.Name.Value == identName {
				foundLocation = nodeToLocation(decl, uri)
				return false
			}
			// Also check function parameters
			for _, param := range decl.Parameters {
				if param.Name != nil && param.Name.Value == identName {
					foundLocation = nodeToLocation(param.Name, uri)
					return false
				}
			}

		case *ast.ClassDecl:
			// Check class name
			if decl.Name != nil && decl.Name.Value == identName {
				foundLocation = nodeToLocation(decl, uri)
				return false
			}

		case *ast.ConstDecl:
			// Check constant name
			if decl.Name != nil && decl.Name.Value == identName {
				foundLocation = nodeToLocation(decl, uri)
				return false
			}

		case *ast.EnumDecl:
			// Check enum name
			if decl.Name != nil && decl.Name.Value == identName {
				foundLocation = nodeToLocation(decl, uri)
				return false
			}
			// Also check enum values
			for _, enumVal := range decl.Values {
				if enumVal.Name == identName {
					foundLocation = nodeToLocation(decl, uri)
					return false
				}
			}

		case *ast.FieldDecl:
			// Check field name
			if decl.Name != nil && decl.Name.Value == identName {
				foundLocation = nodeToLocation(decl, uri)
				return false
			}
		}

		return true // Continue traversal
	})

	return foundLocation
}

// nodeToLocation converts an AST node to an LSP Location.
func nodeToLocation(node ast.Node, uri string) *protocol.Location {
	pos := node.Pos()
	end := node.End()

	// Convert from 1-based AST positions to 0-based LSP positions
	return &protocol.Location{
		URI: uri,
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(pos.Line - 1),
				Character: uint32(pos.Column - 1),
			},
			End: protocol.Position{
				Line:      uint32(end.Line - 1),
				Character: uint32(end.Column - 1),
			},
		},
	}
}
