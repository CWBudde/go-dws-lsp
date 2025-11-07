// Package lsp implements LSP protocol handlers.
package lsp

import (
	"fmt"
	"log"
	"strings"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/analysis"
	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// Hover handles the textDocument/hover request.
// This provides type and symbol information when the user hovers over code.
func Hover(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in Hover")
		return nil, nil
	}

	// Extract document URI and position from params
	uri := params.TextDocument.URI
	position := params.Position

	log.Printf("Hover request at %s line %d, character %d\n",
		uri, position.Line, position.Character)

	// Retrieve document from DocumentStore
	doc, exists := srv.Documents().Get(uri)
	if !exists {
		log.Printf("Document not found for hover: %s\n", uri)
		return nil, nil
	}

	// Check if document and AST are available
	if doc.Program == nil {
		log.Printf("No AST available for hover (document has parse errors): %s\n", uri)
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
		log.Printf("No AST node found at position %d:%d\n", astLine, astColumn)
		return nil, nil
	}

	// Get hover information based on node type
	hoverContent := getHoverContent(node, doc)
	if hoverContent == "" {
		// No hover information available for this node
		return nil, nil
	}

	hover := &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: hoverContent,
		},
	}

	return hover, nil
}

// getHoverContent extracts hover information from an AST node
func getHoverContent(node ast.Node, doc *server.Document) string {
	switch n := node.(type) {
	case *ast.Identifier:
		return getIdentifierHover(n)

	case *ast.FunctionDecl:
		return getFunctionHover(n)

	case *ast.VarDeclStatement:
		return getVariableHover(n)

	case *ast.ConstDecl:
		return getConstHover(n)

	case *ast.ClassDecl:
		return getClassHover(n)

	case *ast.RecordDecl:
		return getRecordHover(n)

	case *ast.EnumDecl:
		return getEnumHover(n)

	case *ast.IntegerLiteral:
		return fmt.Sprintf("```dwscript\n%d\n```\n(Integer literal)", n.Value)

	case *ast.FloatLiteral:
		return fmt.Sprintf("```dwscript\n%f\n```\n(Float literal)", n.Value)

	case *ast.StringLiteral:
		return fmt.Sprintf("```dwscript\n%q\n```\n(String literal)", n.Value)

	case *ast.BooleanLiteral:
		return fmt.Sprintf("```dwscript\n%t\n```\n(Boolean literal)", n.Value)

	default:
		// No hover info for this node type
		return ""
	}
}

// getIdentifierHover returns hover info for an identifier
func getIdentifierHover(ident *ast.Identifier) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("```dwscript\n%s\n```", ident.Value))

	if ident.Type != nil && ident.Type.Name != "" {
		parts = append(parts, fmt.Sprintf("Type: `%s`", ident.Type.Name))
	} else {
		parts = append(parts, "(identifier)")
	}

	return strings.Join(parts, "\n\n")
}

// getFunctionHover returns hover info for a function declaration
func getFunctionHover(fn *ast.FunctionDecl) string {
	var sig strings.Builder

	sig.WriteString("function ")
	sig.WriteString(fn.Name.Value)
	sig.WriteString("(")

	// Add parameters
	for i, param := range fn.Parameters {
		if i > 0 {
			sig.WriteString(", ")
		}

		if param.IsConst {
			sig.WriteString("const ")
		}
		if param.IsLazy {
			sig.WriteString("lazy ")
		}
		if param.ByRef {
			sig.WriteString("var ")
		}

		if param.Name != nil {
			sig.WriteString(param.Name.Value)
			sig.WriteString(": ")
		}

		if param.Type != nil {
			sig.WriteString(param.Type.Name)
		}

		if param.DefaultValue != nil {
			sig.WriteString(" = ")
			sig.WriteString(param.DefaultValue.String())
		}
	}

	sig.WriteString(")")

	// Add return type
	if fn.ReturnType != nil && fn.ReturnType.Name != "" {
		sig.WriteString(": ")
		sig.WriteString(fn.ReturnType.Name)
	}

	return fmt.Sprintf("```dwscript\n%s\n```", sig.String())
}

// getVariableHover returns hover info for a variable declaration
func getVariableHover(varDecl *ast.VarDeclStatement) string {
	if len(varDecl.Names) == 0 {
		return ""
	}

	var parts []string
	for _, name := range varDecl.Names {
		var decl strings.Builder
		decl.WriteString("var ")
		decl.WriteString(name.Value)

		if varDecl.Type != nil && varDecl.Type.Name != "" {
			decl.WriteString(": ")
			decl.WriteString(varDecl.Type.Name)
		}

		parts = append(parts, decl.String())
	}

	return fmt.Sprintf("```dwscript\n%s\n```", strings.Join(parts, "\n"))
}

// getConstHover returns hover info for a constant declaration
func getConstHover(constDecl *ast.ConstDecl) string {
	var sig strings.Builder
	sig.WriteString("const ")
	sig.WriteString(constDecl.Name.Value)

	if constDecl.Type != nil && constDecl.Type.Name != "" {
		sig.WriteString(": ")
		sig.WriteString(constDecl.Type.Name)
	}

	if constDecl.Value != nil {
		sig.WriteString(" = ")
		sig.WriteString(constDecl.Value.String())
	}

	return fmt.Sprintf("```dwscript\n%s\n```", sig.String())
}

// getClassHover returns hover info for a class declaration
func getClassHover(classDecl *ast.ClassDecl) string {
	var sig strings.Builder
	sig.WriteString("type ")
	sig.WriteString(classDecl.Name.Value)
	sig.WriteString(" = class")

	if classDecl.Parent != nil {
		sig.WriteString("(")
		sig.WriteString(classDecl.Parent.Value)
		sig.WriteString(")")
	}

	// Add summary of members
	var info []string
	if len(classDecl.Fields) > 0 {
		info = append(info, fmt.Sprintf("%d field(s)", len(classDecl.Fields)))
	}
	if len(classDecl.Methods) > 0 {
		info = append(info, fmt.Sprintf("%d method(s)", len(classDecl.Methods)))
	}
	if len(classDecl.Properties) > 0 {
		info = append(info, fmt.Sprintf("%d property(ies)", len(classDecl.Properties)))
	}

	result := fmt.Sprintf("```dwscript\n%s\n```", sig.String())

	if len(info) > 0 {
		result += "\n\n" + strings.Join(info, ", ")
	}

	return result
}

// getRecordHover returns hover info for a record declaration
func getRecordHover(recordDecl *ast.RecordDecl) string {
	var sig strings.Builder
	sig.WriteString("type ")
	sig.WriteString(recordDecl.Name.Value)
	sig.WriteString(" = record")

	// Add summary of fields
	var info string
	if len(recordDecl.Properties) > 0 {
		info = fmt.Sprintf("\n\n%d field(s)", len(recordDecl.Properties))
	}

	return fmt.Sprintf("```dwscript\n%s\n```%s", sig.String(), info)
}

// getEnumHover returns hover info for an enum declaration
func getEnumHover(enumDecl *ast.EnumDecl) string {
	var sig strings.Builder
	sig.WriteString("type ")
	sig.WriteString(enumDecl.Name.Value)
	sig.WriteString(" = (")

	// Add enum values
	for i, value := range enumDecl.Values {
		if i > 0 {
			sig.WriteString(", ")
		}
		sig.WriteString(value.Name)
	}

	sig.WriteString(")")

	return fmt.Sprintf("```dwscript\n%s\n```", sig.String())
}
