// Package lsp implements LSP protocol handlers.
package lsp

import (
	"errors"
	"fmt"
	"log"

	"github.com/CWBudde/go-dws-lsp/internal/analysis"
	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// DWScript keywords that cannot be renamed.
var dwscriptKeywords = map[string]bool{
	// Control structures
	"if": true, "then": true, "else": true, "begin": true, "end": true,
	"for": true, "to": true, "downto": true, "do": true,
	"while": true, "repeat": true, "until": true,
	"case": true, "of": true,
	"try": true, "except": true, "finally": true, "raise": true, "on": true,

	// Declarations
	"var": true, "const": true, "type": true,
	"function": true, "procedure": true,
	"class": true, "record": true, "interface": true,
	"property": true, "read": true, "write": true,
	"constructor": true, "destructor": true,

	// Visibility modifiers
	"private": true, "protected": true, "public": true, "published": true,

	// Other keywords
	"as": true, "is": true, "in": true, "inherited": true,
	"not": true, "and": true, "or": true, "xor": true,
	"div": true, "mod": true, "shl": true, "shr": true,
	"array": true, "set": true,
	"unit": true, "program": true, "uses": true,
	"implementation": true, "with": true,

	// Control flow
	"exit": true, "break": true, "continue": true,

	// Literals
	"nil": true, "true": true, "false": true,
}

// Built-in type names that cannot be renamed.
var builtInTypes = map[string]bool{
	"Integer": true, "Float": true, "String": true, "Boolean": true, "Variant": true,
	"TObject": true, "TClass": true, "DateTime": true, "Currency": true,
	"Byte": true, "Word": true, "Cardinal": true, "Int64": true, "UInt64": true,
	"Single": true, "Double": true, "Extended": true, "Char": true,
}

// Built-in function names that cannot be renamed.
var builtInFunctions = map[string]bool{
	"Print": true, "PrintLn": true, "Length": true, "Copy": true, "Pos": true,
	"UpperCase": true, "LowerCase": true, "Trim": true,
	"IntToStr": true, "StrToInt": true, "FloatToStr": true, "StrToFloat": true,
	"Now": true, "Date": true, "Time": true, "FormatDateTime": true,
	"Inc": true, "Dec": true, "Chr": true, "Ord": true,
	"Round": true, "Trunc": true, "Abs": true, "Sqrt": true, "Sqr": true,
	"Sin": true, "Cos": true, "Tan": true, "Exp": true, "Ln": true,
	"Random": true, "Randomize": true,
}

// Rename handles the textDocument/rename request.
// It performs a symbol rename across all references in the workspace.
//
// This implementation (Phase 11.1-11.2) provides the foundation for rename support.
// Future tasks will refine validation, cross-file edits, and edge cases per PLAN.md (Phases 11.3+).
func Rename(context *glsp.Context, params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in Rename")
		return nil, errors.New("server instance not available")
	}

	// Extract request details
	uri := params.TextDocument.URI
	position := params.Position
	newName := params.NewName

	log.Printf("Rename request at %s line %d, character %d (newName=%s)\n",
		uri, position.Line, position.Character, newName)

	// Retrieve document from store
	doc, exists := srv.Documents().Get(uri)
	if !exists {
		log.Printf("Document not found for rename: %s\n", uri)
		return nil, fmt.Errorf("document not found: %s", uri)
	}

	// Ensure we have a parsed program/AST
	if doc.Program == nil || doc.Program.AST() == nil {
		log.Printf("No AST available for rename (document has parse errors): %s\n", uri)
		return nil, errors.New("cannot rename in document with parse errors")
	}

	// Convert LSP (0-based) to AST (1-based)
	astLine := int(position.Line) + 1
	astColumn := int(position.Character) + 1

	programAST := doc.Program.AST()

	// Identify symbol at position (name + kind)
	sym := analysis.IdentifySymbolAtPosition(programAST, astLine, astColumn)
	if sym == nil || sym.Name == "" {
		log.Printf("No symbol at position %d:%d for rename\n", astLine, astColumn)
		return nil, errors.New("no symbol found at cursor position")
	}

	oldName := sym.Name
	log.Printf("Rename target symbol: %s (kind=%s) -> %s", oldName, sym.Kind, newName)

	// Validate that the symbol can be renamed
	if canRename, reason := canRenameSymbol(oldName); !canRename {
		log.Printf("Cannot rename symbol %s: %s\n", oldName, reason)
		return nil, fmt.Errorf("cannot rename '%s': %s", oldName, reason)
	}

	// Validate the new name is valid
	if newName == "" {
		return nil, errors.New("new name cannot be empty")
	}

	// Find all references to the symbol (including declaration)
	// Reuse the References handler logic with includeDeclaration=true
	refParams := &protocol.ReferenceParams{
		TextDocumentPositionParams: params.TextDocumentPositionParams,
		Context: protocol.ReferenceContext{
			IncludeDeclaration: true,
		},
	}

	locations, err := References(context, refParams)
	if err != nil {
		log.Printf("Error finding references for rename: %v\n", err)
		return nil, fmt.Errorf("failed to find references: %w", err)
	}

	if len(locations) == 0 {
		log.Printf("No references found for symbol %s\n", oldName)
		return nil, fmt.Errorf("no references found for symbol '%s'", oldName)
	}

	log.Printf("Found %d references for rename\n", len(locations))

	// Build WorkspaceEdit with all the text edits
	workspaceEdit := buildWorkspaceEdit(locations, newName, srv.Documents())

	return workspaceEdit, nil
}

// PrepareRename handles the textDocument/prepareRename request.
// It validates whether a symbol can be renamed and returns the range and placeholder text.
//
// This handler is called before the actual rename to provide early feedback to the user.
func PrepareRename(context *glsp.Context, params *protocol.PrepareRenameParams) (any, error) {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in PrepareRename")
		return nil, errors.New("server instance not available")
	}

	// Extract request details
	uri := params.TextDocument.URI
	position := params.Position

	log.Printf("PrepareRename request at %s line %d, character %d\n",
		uri, position.Line, position.Character)

	// Retrieve document from store
	doc, exists := srv.Documents().Get(uri)
	if !exists {
		log.Printf("Document not found for prepareRename: %s\n", uri)
		return nil, fmt.Errorf("document not found: %s", uri)
	}

	// Ensure we have a parsed program/AST
	if doc.Program == nil || doc.Program.AST() == nil {
		log.Printf("No AST available for prepareRename (document has parse errors): %s\n", uri)
		return nil, errors.New("cannot rename in document with parse errors")
	}

	// Convert LSP (0-based) to AST (1-based)
	astLine := int(position.Line) + 1
	astColumn := int(position.Character) + 1

	programAST := doc.Program.AST()

	// Identify symbol at position
	sym := analysis.IdentifySymbolAtPosition(programAST, astLine, astColumn)
	if sym == nil || sym.Name == "" {
		log.Printf("No symbol at position %d:%d for prepareRename\n", astLine, astColumn)
		return nil, errors.New("no symbol found at cursor position")
	}

	symbolName := sym.Name
	log.Printf("PrepareRename target symbol: %s (kind=%s)", symbolName, sym.Kind)

	// Validate that the symbol can be renamed
	if canRename, reason := canRenameSymbol(symbolName); !canRename {
		log.Printf("Cannot rename symbol %s: %s\n", symbolName, reason)
		return nil, fmt.Errorf("cannot rename '%s': %s", symbolName, reason)
	}

	// Get the range of the symbol
	// Find the node at the position to get its range
	node := analysis.FindNodeAtPosition(programAST, astLine, astColumn)
	if node == nil {
		log.Printf("No AST node at position %d:%d for prepareRename\n", astLine, astColumn)
		return nil, errors.New("no symbol found at cursor position")
	}

	// Get the identifier node
	var identNode *ast.Identifier

	switch n := node.(type) {
	case *ast.Identifier:
		identNode = n
	default:
		// Try to find an identifier within the node
		ast.Inspect(node, func(child ast.Node) bool {
			if ident, ok := child.(*ast.Identifier); ok {
				if ident.Value == symbolName {
					identNode = ident
					return false
				}
			}

			return true
		})
	}

	if identNode == nil {
		log.Printf("No identifier node found for symbol %s\n", symbolName)
		return nil, errors.New("cannot determine symbol range")
	}

	// Convert AST positions (1-based) to LSP positions (0-based)
	start := identNode.Pos()
	end := identNode.End()

	symbolRange := protocol.Range{
		Start: protocol.Position{
			Line:      uint32(max(0, start.Line-1)),
			Character: uint32(max(0, start.Column-1)),
		},
		End: protocol.Position{
			Line:      uint32(max(0, end.Line-1)),
			Character: uint32(max(0, end.Column-1)),
		},
	}

	// Return range with placeholder
	// According to LSP spec, we can return either a Range or a RangeWithPlaceholder
	result := map[string]any{
		"range":       symbolRange,
		"placeholder": symbolName,
	}

	log.Printf("PrepareRename successful for symbol %s at range %v\n", symbolName, symbolRange)

	return result, nil
}

// canRenameSymbol checks whether a symbol can be renamed.
// It rejects DWScript keywords, built-in types, and built-in functions.
// Returns (true, "") if the symbol can be renamed, or (false, reason) if not.
func canRenameSymbol(symbolName string) (bool, string) {
	// Check if it's a keyword
	if dwscriptKeywords[symbolName] {
		return false, "cannot rename DWScript keyword"
	}

	// Check if it's a built-in type
	if builtInTypes[symbolName] {
		return false, "cannot rename built-in type"
	}

	// Check if it's a built-in function
	if builtInFunctions[symbolName] {
		return false, "cannot rename built-in function"
	}

	// Symbol can be renamed
	return true, ""
}

// buildWorkspaceEdit creates a WorkspaceEdit from a list of locations.
// It groups edits by document URI and creates TextDocumentEdit entries.
func buildWorkspaceEdit(locations []protocol.Location, newName string, docs *server.DocumentStore) *protocol.WorkspaceEdit {
	// Group locations by document URI
	editsByURI := make(map[protocol.DocumentUri][]protocol.TextEdit)

	for _, loc := range locations {
		// Create a TextEdit for this location
		edit := protocol.TextEdit{
			Range:   loc.Range,
			NewText: newName,
		}

		editsByURI[loc.URI] = append(editsByURI[loc.URI], edit)
	}

	// Build DocumentChanges (preferred over Changes for versioned edits)
	var documentChanges []any

	for uri, edits := range editsByURI {
		// Get document version from DocumentStore
		var version *int32

		if doc, exists := docs.Get(uri); exists {
			v := int32(doc.Version)
			version = &v
		}

		// Create TextDocumentEdit
		textDocEdit := protocol.TextDocumentEdit{
			TextDocument: protocol.OptionalVersionedTextDocumentIdentifier{
				TextDocumentIdentifier: protocol.TextDocumentIdentifier{
					URI: uri,
				},
				Version: version,
			},
			Edits: convertToEdits(edits),
		}

		documentChanges = append(documentChanges, textDocEdit)
	}

	workspaceEdit := &protocol.WorkspaceEdit{
		DocumentChanges: documentChanges,
	}

	log.Printf("Built WorkspaceEdit with %d document(s) and %d total edit(s)\n",
		len(documentChanges), len(locations))

	return workspaceEdit
}

// convertToEdits converts []protocol.TextEdit to []interface{} as required by the protocol.
func convertToEdits(textEdits []protocol.TextEdit) []any {
	edits := make([]any, len(textEdits))
	for i, edit := range textEdits {
		edits[i] = edit
	}

	return edits
}
