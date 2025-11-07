// Package analysis provides type resolution utilities for code completion.
package analysis

import (
	"log"

	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// TypeInfo represents type information for a symbol.
type TypeInfo struct {
	// TypeName is the name of the type (e.g., "Integer", "String", "TMyClass")
	TypeName string

	// IsBuiltIn indicates if this is a built-in type
	IsBuiltIn bool

	// Declaration is the AST node where the type is declared (for user-defined types)
	Declaration ast.Node
}

// ResolveMemberType resolves the type of an identifier at a given position.
// This is used for member access completion (e.g., "myObject.").
// Returns the type information or nil if the type cannot be determined.
func ResolveMemberType(doc *server.Document, identifier string, line, character int) (*TypeInfo, error) {
	log.Printf("ResolveMemberType: resolving type of '%s' at %d:%d", identifier, line, character)

	if doc.Program == nil || doc.Program.AST() == nil {
		log.Println("ResolveMemberType: no AST available")
		return nil, nil
	}

	// Convert LSP position (0-based) to AST position (1-based)
	astLine := line + 1
	astColumn := character + 1
	pos := token.Position{
		Line:   astLine,
		Column: astColumn,
	}

	// Create a symbol resolver to find the identifier declaration
	resolver := NewSymbolResolver(doc.URI, doc.Program.AST(), pos)

	// Resolve the identifier to its declaration
	locations := resolver.ResolveSymbol(identifier)
	if len(locations) == 0 {
		log.Printf("ResolveMemberType: could not find declaration for '%s'", identifier)
		return nil, nil
	}

	// Get the first location (most relevant)
	// Note: For now we only consider the first result, but in the future
	// we might want to handle multiple definitions (e.g., overloaded functions)
	location := locations[0]

	// Now we need to find the actual declaration node and extract its type
	// We'll search the AST for a node at the resolved location
	typeInfo := extractTypeFromLocation(doc.Program.AST(), identifier, location)

	if typeInfo != nil {
		log.Printf("ResolveMemberType: resolved '%s' to type '%s'", identifier, typeInfo.TypeName)
	} else {
		log.Printf("ResolveMemberType: could not determine type for '%s'", identifier)
	}

	return typeInfo, nil
}

// extractTypeFromLocation finds the declaration node at the given location
// and extracts type information from it.
func extractTypeFromLocation(program *ast.Program, identifier string, location protocol.Location) *TypeInfo {
	// Convert LSP position (0-based) to AST position (1-based)
	astLine := int(location.Range.Start.Line) + 1
	astColumn := int(location.Range.Start.Character) + 1

	// Search the AST for the declaration at this position
	var foundType *TypeInfo

	ast.Inspect(program, func(node ast.Node) bool {
		if node == nil || foundType != nil {
			return false
		}

		// Check if this node is at the target position
		pos := node.Pos()
		if pos.Line != astLine || pos.Column != astColumn {
			return true // Continue searching
		}

		// We found a node at the target position
		// Now try to extract type information based on the node type
		foundType = extractTypeFromNode(program, node, identifier)

		return false // Stop searching
	})

	return foundType
}

// extractTypeFromNode extracts type information from an AST node.
// The node should be the identifier or declaration node.
func extractTypeFromNode(program *ast.Program, node ast.Node, identifier string) *TypeInfo {
	// If the node is an identifier, we need to find the parent declaration
	// For now, we'll search the program for declarations matching the identifier

	// Try to find the declaration in various contexts
	if typeInfo := findVariableType(program, identifier); typeInfo != nil {
		return typeInfo
	}

	if typeInfo := findParameterType(program, identifier); typeInfo != nil {
		return typeInfo
	}

	if typeInfo := findFieldType(program, identifier); typeInfo != nil {
		return typeInfo
	}

	return nil
}

// findVariableType searches for a variable declaration and returns its type.
func findVariableType(program *ast.Program, varName string) *TypeInfo {
	var typeInfo *TypeInfo

	ast.Inspect(program, func(node ast.Node) bool {
		if node == nil || typeInfo != nil {
			return false
		}

		// Check for variable declarations
		if varDecl, ok := node.(*ast.VarDeclStatement); ok {
			// Check if any of the declared names match
			for _, name := range varDecl.Names {
				if name.Value == varName && varDecl.Type != nil {
					typeInfo = typeAnnotationToTypeInfo(varDecl.Type)
					return false
				}
			}
		}

		return true
	})

	return typeInfo
}

// findParameterType searches for a function parameter and returns its type.
func findParameterType(program *ast.Program, paramName string) *TypeInfo {
	var typeInfo *TypeInfo

	ast.Inspect(program, func(node ast.Node) bool {
		if node == nil || typeInfo != nil {
			return false
		}

		// Check function declarations
		if funcDecl, ok := node.(*ast.FunctionDecl); ok {
			for _, param := range funcDecl.Parameters {
				if param.Name != nil && param.Name.Value == paramName && param.Type != nil {
					typeInfo = typeAnnotationToTypeInfo(param.Type)
					return false
				}
			}
		}

		return true
	})

	return typeInfo
}

// findFieldType searches for a class field and returns its type.
func findFieldType(program *ast.Program, fieldName string) *TypeInfo {
	var typeInfo *TypeInfo

	ast.Inspect(program, func(node ast.Node) bool {
		if node == nil || typeInfo != nil {
			return false
		}

		// Check class declarations
		if classDecl, ok := node.(*ast.ClassDecl); ok {
			for _, field := range classDecl.Fields {
				if field.Name != nil && field.Name.Value == fieldName {
					typeInfo = typeExpressionToTypeInfo(field.Type)
					return false
				}
			}
		}

		return true
	})

	return typeInfo
}

// typeAnnotationToTypeInfo converts a TypeAnnotation to TypeInfo.
func typeAnnotationToTypeInfo(typeAnnotation *ast.TypeAnnotation) *TypeInfo {
	if typeAnnotation == nil {
		return nil
	}

	return &TypeInfo{
		TypeName:  typeAnnotation.Name,
		IsBuiltIn: isBuiltInType(typeAnnotation.Name),
	}
}

// typeExpressionToTypeInfo converts a TypeExpression to TypeInfo.
func typeExpressionToTypeInfo(typeExpr ast.TypeExpression) *TypeInfo {
	if typeExpr == nil {
		return nil
	}

	// Try to convert to TypeAnnotation (most common case)
	if typeAnnotation, ok := typeExpr.(*ast.TypeAnnotation); ok {
		return typeAnnotationToTypeInfo(typeAnnotation)
	}

	// For other type expressions, we'll need to determine the type name
	// For now, return the string representation
	return &TypeInfo{
		TypeName:  typeExpr.String(),
		IsBuiltIn: false,
	}
}

// isBuiltInType checks if a type name is a built-in DWScript type.
func isBuiltInType(typeName string) bool {
	builtInTypes := map[string]bool{
		"Integer":  true,
		"Float":    true,
		"String":   true,
		"Boolean":  true,
		"Variant":  true,
		"TObject":  true,
		"TClass":   true,
		"DateTime": true,
		"Currency": true,
		"Byte":     true,
		"Word":     true,
		"Cardinal": true,
		"Int64":    true,
		"UInt64":   true,
		"Single":   true,
		"Double":   true,
		"Extended": true,
		"Char":     true,
	}

	return builtInTypes[typeName]
}
