// Package lsp implements LSP protocol handlers.
package lsp

import (
	"fmt"
	"log"
	"strings"

	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// DocumentSymbol handles the textDocument/documentSymbol request.
// It returns a hierarchical list of symbols in the document for the outline view.
func DocumentSymbol(context *glsp.Context, params *protocol.DocumentSymbolParams) (any, error) {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in DocumentSymbol")
		return []protocol.DocumentSymbol{}, nil
	}

	// Extract document URI from params
	uri := params.TextDocument.URI
	log.Printf("DocumentSymbol request for %s\n", uri)

	// Retrieve document from DocumentStore
	doc, exists := srv.Documents().Get(uri)
	if !exists {
		log.Printf("Document not found for document symbols: %s\n", uri)
		return []protocol.DocumentSymbol{}, nil
	}

	// Check if document has valid AST
	if doc.Program == nil || doc.Program.AST() == nil {
		log.Printf("No AST available for document symbols (document has parse errors): %s\n", uri)
		return []protocol.DocumentSymbol{}, nil
	}

	programAST := doc.Program.AST()

	// Collect symbols from the AST
	symbols := collectDocumentSymbols(programAST)

	log.Printf("Found %d top-level symbols in %s\n", len(symbols), uri)

	return symbols, nil
}

// collectDocumentSymbols traverses the AST and collects all top-level symbols
// including functions, variables, constants, classes, etc.
func collectDocumentSymbols(program *ast.Program) []protocol.DocumentSymbol {
	if program == nil {
		return nil
	}

	var symbols []protocol.DocumentSymbol

	// Traverse top-level statements
	for _, stmt := range program.Statements {
		if stmt == nil {
			continue
		}

		var sym *protocol.DocumentSymbol

		switch node := stmt.(type) {
		case *ast.FunctionDecl:
			sym = createFunctionSymbol(node)

		case *ast.VarDeclStatement:
			// For variable declarations, create a symbol for each variable
			symbols = append(symbols, createVariableSymbols(node)...)

			continue

		case *ast.ConstDecl:
			sym = createConstSymbol(node)

		case *ast.ClassDecl:
			sym = createClassSymbol(node)

		case *ast.RecordDecl:
			sym = createRecordSymbol(node)

		case *ast.EnumDecl:
			sym = createEnumSymbol(node)

		default:
			// Skip other statement types (expressions, control flow, etc.)
			continue
		}

		if sym != nil {
			symbols = append(symbols, *sym)
		}
	}

	return symbols
}

// createFunctionSymbol creates a DocumentSymbol for a function declaration.
func createFunctionSymbol(fn *ast.FunctionDecl) *protocol.DocumentSymbol {
	if fn == nil || fn.Name == nil {
		return nil
	}

	// Build function signature for detail
	detail := buildFunctionSignature(fn)

	// Determine the range (entire function span)
	start := fn.Pos()
	end := fn.End()

	// Selection range is just the function name
	nameStart := fn.Name.Pos()
	nameEnd := fn.Name.End()

	return &protocol.DocumentSymbol{
		Name:   fn.Name.Value,
		Kind:   protocol.SymbolKindFunction,
		Detail: &detail,
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
		SelectionRange: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(max(0, nameStart.Line-1)),
				Character: uint32(max(0, nameStart.Column-1)),
			},
			End: protocol.Position{
				Line:      uint32(max(0, nameEnd.Line-1)),
				Character: uint32(max(0, nameEnd.Column-1)),
			},
		},
	}
}

// buildFunctionSignature builds a function signature string for the detail field.
func buildFunctionSignature(fn *ast.FunctionDecl) string {
	if fn == nil || fn.Name == nil {
		return ""
	}

	sig := "function " + fn.Name.Value + "("

	// Add parameters
	var sigSb157 strings.Builder

	for i, param := range fn.Parameters {
		if i > 0 {
			sigSb157.WriteString(", ")
		}

		if param.IsConst {
			sigSb157.WriteString("const ")
		}

		if param.IsLazy {
			sigSb157.WriteString("lazy ")
		}

		if param.ByRef {
			sigSb157.WriteString("var ")
		}

		if param.Name != nil {
			sigSb157.WriteString(param.Name.Value)

			if param.Type != nil {
				sigSb157.WriteString(": " + param.Type.Name)
			}
		}

		if param.DefaultValue != nil {
			sigSb157.WriteString(" = " + param.DefaultValue.String())
		}
	}

	sig += sigSb157.String()

	sig += ")"

	// Add return type
	if fn.ReturnType != nil && fn.ReturnType.Name != "" {
		sig += ": " + fn.ReturnType.Name
	}

	return sig
}

// createVariableSymbols creates DocumentSymbols for variable declarations.
func createVariableSymbols(varDecl *ast.VarDeclStatement) []protocol.DocumentSymbol {
	if varDecl == nil || len(varDecl.Names) == 0 {
		return nil
	}

	symbols := make([]protocol.DocumentSymbol, 0, len(varDecl.Names))

	for _, name := range varDecl.Names {
		if name == nil {
			continue
		}

		detail := "var"
		if varDecl.Type != nil && varDecl.Type.Name != "" {
			detail += ": " + varDecl.Type.Name
		}

		// For variables, the range is the entire declaration statement
		start := varDecl.Pos()
		end := varDecl.End()

		// Selection range is just the variable name
		nameStart := name.Pos()
		nameEnd := name.End()

		symbols = append(symbols, protocol.DocumentSymbol{
			Name:   name.Value,
			Kind:   protocol.SymbolKindVariable,
			Detail: &detail,
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
			SelectionRange: protocol.Range{
				Start: protocol.Position{
					Line:      uint32(max(0, nameStart.Line-1)),
					Character: uint32(max(0, nameStart.Column-1)),
				},
				End: protocol.Position{
					Line:      uint32(max(0, nameEnd.Line-1)),
					Character: uint32(max(0, nameEnd.Column-1)),
				},
			},
		})
	}

	return symbols
}

// createConstSymbol creates a DocumentSymbol for a constant declaration.
func createConstSymbol(constDecl *ast.ConstDecl) *protocol.DocumentSymbol {
	if constDecl == nil || constDecl.Name == nil {
		return nil
	}

	detail := "const"
	if constDecl.Type != nil && constDecl.Type.Name != "" {
		detail += ": " + constDecl.Type.Name
	}

	if constDecl.Value != nil {
		detail += " = " + constDecl.Value.String()
	}

	start := constDecl.Pos()
	end := constDecl.End()

	nameStart := constDecl.Name.Pos()
	nameEnd := constDecl.Name.End()

	return &protocol.DocumentSymbol{
		Name:   constDecl.Name.Value,
		Kind:   protocol.SymbolKindConstant,
		Detail: &detail,
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
		SelectionRange: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(max(0, nameStart.Line-1)),
				Character: uint32(max(0, nameStart.Column-1)),
			},
			End: protocol.Position{
				Line:      uint32(max(0, nameEnd.Line-1)),
				Character: uint32(max(0, nameEnd.Column-1)),
			},
		},
	}
}

// createClassSymbol creates a DocumentSymbol for a class declaration.
func createClassSymbol(classDecl *ast.ClassDecl) *protocol.DocumentSymbol {
	if classDecl == nil || classDecl.Name == nil {
		return nil
	}

	detail := "class"
	if classDecl.Parent != nil {
		detail += "(" + classDecl.Parent.Value + ")"
	}

	start := classDecl.Pos()
	end := classDecl.End()

	nameStart := classDecl.Name.Pos()
	nameEnd := classDecl.Name.End()

	symbol := &protocol.DocumentSymbol{
		Name:   classDecl.Name.Value,
		Kind:   protocol.SymbolKindClass,
		Detail: &detail,
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
		SelectionRange: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(max(0, nameStart.Line-1)),
				Character: uint32(max(0, nameStart.Column-1)),
			},
			End: protocol.Position{
				Line:      uint32(max(0, nameEnd.Line-1)),
				Character: uint32(max(0, nameEnd.Column-1)),
			},
		},
	}

	// Add class members as children (fields, methods, properties)
	children := make([]protocol.DocumentSymbol, 0, 20)
	addClassFields(&children, classDecl.Fields)
	addClassMethods(&children, classDecl.Methods)
	addClassProperties(&children, classDecl.Properties)

	if len(children) > 0 {
		symbol.Children = children
	}

	return symbol
}

// addClassFields adds field symbols to the children list.
func addClassFields(children *[]protocol.DocumentSymbol, fields []*ast.FieldDecl) {
	for _, field := range fields {
		if field == nil || field.Name == nil {
			continue
		}

		fieldDetail := "field"
		if field.Type != nil {
			if typeAnnot, ok := field.Type.(*ast.TypeAnnotation); ok && typeAnnot.Name != "" {
				fieldDetail += ": " + typeAnnot.Name
			}
		}

		*children = append(*children, createDocSymbol(
			field.Name.Value,
			protocol.SymbolKindField,
			fieldDetail,
			field,
			field.Name,
		))
	}
}

// addClassMethods adds method symbols to the children list.
func addClassMethods(children *[]protocol.DocumentSymbol, methods []*ast.FunctionDecl) {
	for _, method := range methods {
		if method == nil || method.Name == nil {
			continue
		}

		methodDetail := buildFunctionSignature(method)
		*children = append(*children, createDocSymbol(
			method.Name.Value,
			protocol.SymbolKindMethod,
			methodDetail,
			method,
			method.Name,
		))
	}
}

// addClassProperties adds property symbols to the children list.
func addClassProperties(children *[]protocol.DocumentSymbol, properties []*ast.PropertyDecl) {
	for _, prop := range properties {
		if prop == nil || prop.Name == nil {
			continue
		}

		propDetail := "property"
		if prop.Type != nil && prop.Type.Name != "" {
			propDetail += ": " + prop.Type.Name
		}

		*children = append(*children, createDocSymbol(
			prop.Name.Value,
			protocol.SymbolKindProperty,
			propDetail,
			prop,
			prop.Name,
		))
	}
}

// createDocSymbol creates a DocumentSymbol from an AST node with ranges.
func createDocSymbol(name string, kind protocol.SymbolKind, detail string, node ast.Node, nameNode ast.Node) protocol.DocumentSymbol {
	nodeStart := node.Pos()
	nodeEnd := node.End()
	nameStart := nameNode.Pos()
	nameEnd := nameNode.End()

	return protocol.DocumentSymbol{
		Name:   name,
		Kind:   kind,
		Detail: &detail,
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(max(0, nodeStart.Line-1)),
				Character: uint32(max(0, nodeStart.Column-1)),
			},
			End: protocol.Position{
				Line:      uint32(max(0, nodeEnd.Line-1)),
				Character: uint32(max(0, nodeEnd.Column-1)),
			},
		},
		SelectionRange: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(max(0, nameStart.Line-1)),
				Character: uint32(max(0, nameStart.Column-1)),
			},
			End: protocol.Position{
				Line:      uint32(max(0, nameEnd.Line-1)),
				Character: uint32(max(0, nameEnd.Column-1)),
			},
		},
	}
}

// createRecordSymbol creates a DocumentSymbol for a record declaration.
func createRecordSymbol(recordDecl *ast.RecordDecl) *protocol.DocumentSymbol {
	if recordDecl == nil || recordDecl.Name == nil {
		return nil
	}

	detail := "record"

	start := recordDecl.Pos()
	end := recordDecl.End()

	nameStart := recordDecl.Name.Pos()
	nameEnd := recordDecl.Name.End()

	symbol := &protocol.DocumentSymbol{
		Name:   recordDecl.Name.Value,
		Kind:   protocol.SymbolKindStruct,
		Detail: &detail,
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
		SelectionRange: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(max(0, nameStart.Line-1)),
				Character: uint32(max(0, nameStart.Column-1)),
			},
			End: protocol.Position{
				Line:      uint32(max(0, nameEnd.Line-1)),
				Character: uint32(max(0, nameEnd.Column-1)),
			},
		},
	}

	// Add record fields as children
	children := make([]protocol.DocumentSymbol, 0, len(recordDecl.Properties))

	for i := range recordDecl.Properties {
		prop := &recordDecl.Properties[i]
		if prop.Name == nil {
			continue
		}

		fieldDetail := "field"
		if prop.Type != nil && prop.Type.Name != "" {
			fieldDetail += ": " + prop.Type.Name
		}

		// Use the name's position since RecordPropertyDecl doesn't have Pos()
		propStart := prop.Name.Pos()
		propEnd := prop.End()
		propNameStart := prop.Name.Pos()
		propNameEnd := prop.Name.End()

		children = append(children, protocol.DocumentSymbol{
			Name:   prop.Name.Value,
			Kind:   protocol.SymbolKindField,
			Detail: &fieldDetail,
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      uint32(max(0, propStart.Line-1)),
					Character: uint32(max(0, propStart.Column-1)),
				},
				End: protocol.Position{
					Line:      uint32(max(0, propEnd.Line-1)),
					Character: uint32(max(0, propEnd.Column-1)),
				},
			},
			SelectionRange: protocol.Range{
				Start: protocol.Position{
					Line:      uint32(max(0, propNameStart.Line-1)),
					Character: uint32(max(0, propNameStart.Column-1)),
				},
				End: protocol.Position{
					Line:      uint32(max(0, propNameEnd.Line-1)),
					Character: uint32(max(0, propNameEnd.Column-1)),
				},
			},
		})
	}

	if len(children) > 0 {
		symbol.Children = children
	}

	return symbol
}

// createEnumSymbol creates a DocumentSymbol for an enum declaration.
func createEnumSymbol(enumDecl *ast.EnumDecl) *protocol.DocumentSymbol {
	if enumDecl == nil || enumDecl.Name == nil {
		return nil
	}

	detail := "enum"

	start := enumDecl.Pos()
	end := enumDecl.End()

	nameStart := enumDecl.Name.Pos()
	nameEnd := enumDecl.Name.End()

	symbol := &protocol.DocumentSymbol{
		Name:   enumDecl.Name.Value,
		Kind:   protocol.SymbolKindEnum,
		Detail: &detail,
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
		SelectionRange: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(max(0, nameStart.Line-1)),
				Character: uint32(max(0, nameStart.Column-1)),
			},
			End: protocol.Position{
				Line:      uint32(max(0, nameEnd.Line-1)),
				Character: uint32(max(0, nameEnd.Column-1)),
			},
		},
	}

	// Add enum values as children
	children := make([]protocol.DocumentSymbol, 0, len(enumDecl.Values))

	for _, value := range enumDecl.Values {
		if value.Name == "" {
			continue
		}

		valueDetail := "enum member"
		if value.Value != nil {
			// Format the integer value
			valueDetail += fmt.Sprintf(" = %d", *value.Value)
		}

		// For enum values, we don't have precise position information
		// So we use the enum's position range as an approximation
		children = append(children, protocol.DocumentSymbol{
			Name:           value.Name,
			Kind:           protocol.SymbolKindEnumMember,
			Detail:         &valueDetail,
			Range:          symbol.Range,
			SelectionRange: symbol.Range,
		})
	}

	if len(children) > 0 {
		symbol.Children = children
	}

	return symbol
}
