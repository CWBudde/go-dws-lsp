// Package lsp implements LSP protocol handlers.
package lsp

import (
	"fmt"
	"log"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// DocumentSymbol handles the textDocument/documentSymbol request.
// It returns a hierarchical list of symbols in the document for the outline view.
func DocumentSymbol(context *glsp.Context, params *protocol.DocumentSymbolParams) (any, error) {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in DocumentSymbol")
		return nil, nil
	}

	// Extract document URI from params
	uri := params.TextDocument.URI
	log.Printf("DocumentSymbol request for %s\n", uri)

	// Retrieve document from DocumentStore
	doc, exists := srv.Documents().Get(uri)
	if !exists {
		log.Printf("Document not found for document symbols: %s\n", uri)
		return nil, nil
	}

	// Check if document has valid AST
	if doc.Program == nil || doc.Program.AST() == nil {
		log.Printf("No AST available for document symbols (document has parse errors): %s\n", uri)
		return nil, nil
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
			for _, varSym := range createVariableSymbols(node) {
				symbols = append(symbols, varSym)
			}
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

// createFunctionSymbol creates a DocumentSymbol for a function declaration
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

// buildFunctionSignature builds a function signature string for the detail field
func buildFunctionSignature(fn *ast.FunctionDecl) string {
	if fn == nil || fn.Name == nil {
		return ""
	}

	sig := "function " + fn.Name.Value + "("

	// Add parameters
	for i, param := range fn.Parameters {
		if i > 0 {
			sig += ", "
		}

		if param.IsConst {
			sig += "const "
		}
		if param.IsLazy {
			sig += "lazy "
		}
		if param.ByRef {
			sig += "var "
		}

		if param.Name != nil {
			sig += param.Name.Value
			if param.Type != nil {
				sig += ": " + param.Type.Name
			}
		}

		if param.DefaultValue != nil {
			sig += " = " + param.DefaultValue.String()
		}
	}

	sig += ")"

	// Add return type
	if fn.ReturnType != nil && fn.ReturnType.Name != "" {
		sig += ": " + fn.ReturnType.Name
	}

	return sig
}

// createVariableSymbols creates DocumentSymbols for variable declarations
func createVariableSymbols(varDecl *ast.VarDeclStatement) []protocol.DocumentSymbol {
	if varDecl == nil || len(varDecl.Names) == 0 {
		return nil
	}

	var symbols []protocol.DocumentSymbol

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

// createConstSymbol creates a DocumentSymbol for a constant declaration
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

// createClassSymbol creates a DocumentSymbol for a class declaration
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
	var children []protocol.DocumentSymbol

	// Add fields
	for _, field := range classDecl.Fields {
		if field == nil || field.Name == nil {
			continue
		}

		fieldDetail := "field"
		if field.Type != nil {
			// TypeExpression is an interface, try to get type name
			if typeAnnot, ok := field.Type.(*ast.TypeAnnotation); ok && typeAnnot.Name != "" {
				fieldDetail += ": " + typeAnnot.Name
			}
		}

		fieldStart := field.Pos()
		fieldEnd := field.End()
		fieldNameStart := field.Name.Pos()
		fieldNameEnd := field.Name.End()

		children = append(children, protocol.DocumentSymbol{
			Name:   field.Name.Value,
			Kind:   protocol.SymbolKindField,
			Detail: &fieldDetail,
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      uint32(max(0, fieldStart.Line-1)),
					Character: uint32(max(0, fieldStart.Column-1)),
				},
				End: protocol.Position{
					Line:      uint32(max(0, fieldEnd.Line-1)),
					Character: uint32(max(0, fieldEnd.Column-1)),
				},
			},
			SelectionRange: protocol.Range{
				Start: protocol.Position{
					Line:      uint32(max(0, fieldNameStart.Line-1)),
					Character: uint32(max(0, fieldNameStart.Column-1)),
				},
				End: protocol.Position{
					Line:      uint32(max(0, fieldNameEnd.Line-1)),
					Character: uint32(max(0, fieldNameEnd.Column-1)),
				},
			},
		})
	}

	// Add methods
	for _, method := range classDecl.Methods {
		if method == nil || method.Name == nil {
			continue
		}

		methodDetail := buildFunctionSignature(method)

		methodStart := method.Pos()
		methodEnd := method.End()
		methodNameStart := method.Name.Pos()
		methodNameEnd := method.Name.End()

		children = append(children, protocol.DocumentSymbol{
			Name:   method.Name.Value,
			Kind:   protocol.SymbolKindMethod,
			Detail: &methodDetail,
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      uint32(max(0, methodStart.Line-1)),
					Character: uint32(max(0, methodStart.Column-1)),
				},
				End: protocol.Position{
					Line:      uint32(max(0, methodEnd.Line-1)),
					Character: uint32(max(0, methodEnd.Column-1)),
				},
			},
			SelectionRange: protocol.Range{
				Start: protocol.Position{
					Line:      uint32(max(0, methodNameStart.Line-1)),
					Character: uint32(max(0, methodNameStart.Column-1)),
				},
				End: protocol.Position{
					Line:      uint32(max(0, methodNameEnd.Line-1)),
					Character: uint32(max(0, methodNameEnd.Column-1)),
				},
			},
		})
	}

	// Add properties
	for _, prop := range classDecl.Properties {
		if prop == nil || prop.Name == nil {
			continue
		}

		propDetail := "property"
		if prop.Type != nil && prop.Type.Name != "" {
			propDetail += ": " + prop.Type.Name
		}

		propStart := prop.Pos()
		propEnd := prop.End()
		propNameStart := prop.Name.Pos()
		propNameEnd := prop.Name.End()

		children = append(children, protocol.DocumentSymbol{
			Name:   prop.Name.Value,
			Kind:   protocol.SymbolKindProperty,
			Detail: &propDetail,
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

// createRecordSymbol creates a DocumentSymbol for a record declaration
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
	var children []protocol.DocumentSymbol

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

// createEnumSymbol creates a DocumentSymbol for an enum declaration
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
	var children []protocol.DocumentSymbol

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
