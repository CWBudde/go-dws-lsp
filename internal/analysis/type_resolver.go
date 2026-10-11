// Package analysis provides type resolution utilities for code completion.
package analysis

import (
	"log"
	"strconv"
	"strings"

	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

const (
	snippetFinalCursorPosition = ")$0" // $0 is the final cursor position
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
func ResolveMemberType(doc *server.Document, identifier string, line, character int) *TypeInfo {
	log.Printf("ResolveMemberType: resolving type of '%s' at %d:%d", identifier, line, character)

	if doc.Program == nil || doc.Program.AST() == nil {
		log.Println("ResolveMemberType: no AST available")
		return nil
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
		return nil
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

	return typeInfo
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
					typeInfo = typeExpressionToTypeInfo(varDecl.Type)
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
					typeInfo = typeExpressionToTypeInfo(param.Type)
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

// GetTypeMembers retrieves all members (fields, methods, properties) of a type
// and returns them as CompletionItems suitable for member access completion.
func GetTypeMembers(doc *server.Document, typeName string) ([]protocol.CompletionItem, error) {
	log.Printf("GetTypeMembers: retrieving members for type '%s'", typeName)

	if doc.Program == nil || doc.Program.AST() == nil {
		log.Println("GetTypeMembers: no AST available")
		return nil, nil
	}

	var items []protocol.CompletionItem

	// For built-in types, we currently don't provide members
	// In a full implementation, we might provide built-in methods/properties
	if isBuiltInType(typeName) {
		log.Printf("GetTypeMembers: '%s' is a built-in type (no members available)", typeName)
		return items, nil
	}

	// Search for the type definition in the AST
	program := doc.Program.AST()

	// Try to find as a class
	if classMembers := extractClassMembers(program, typeName); len(classMembers) > 0 {
		items = append(items, classMembers...)
	}

	// Try to find as a record
	if recordMembers := extractRecordMembers(program, typeName); len(recordMembers) > 0 {
		items = append(items, recordMembers...)
	}

	// Sort items alphabetically by label
	sortCompletionItems(items)

	log.Printf("GetTypeMembers: found %d members for type '%s'", len(items), typeName)

	return items, nil
}

// extractClassMembers extracts all members from a class declaration.
func extractClassMembers(program *ast.Program, className string) []protocol.CompletionItem {
	items := make([]protocol.CompletionItem, 0, 30)

	// Find the class declaration
	var classDecl *ast.ClassDecl

	ast.Inspect(program, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		if class, ok := node.(*ast.ClassDecl); ok {
			if class.Name != nil && class.Name.Value == className {
				classDecl = class
				return false // Stop searching
			}
		}

		return true
	})

	if classDecl == nil {
		return items
	}

	// Extract fields
	plainTextFormat := protocol.InsertTextFormatPlainText

	for _, field := range classDecl.Fields {
		if field.Name == nil {
			continue
		}

		kind := protocol.CompletionItemKindField
		sortText := "0field~" + field.Name.Value

		item := protocol.CompletionItem{
			Label:            field.Name.Value,
			Kind:             &kind,
			SortText:         &sortText,
			InsertTextFormat: &plainTextFormat,
		}

		// Add type information in detail
		if field.Type != nil {
			detail := field.Type.String()
			item.Detail = &detail
		}

		// Add documentation with MarkupContent
		docValue := "**Field**"
		if field.IsClassVar {
			docValue = "**Class variable** (static field)"
		}

		if field.Type != nil {
			docValue += "\n\n```pascal\n" + field.Name.Value + ": " + field.Type.String() + "\n```"
		}

		doc := protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: docValue,
		}
		item.Documentation = doc

		items = append(items, item)
	}

	// Extract methods
	for _, method := range classDecl.Methods {
		if method.Name == nil {
			continue
		}

		kind := protocol.CompletionItemKindMethod
		sortText := "1method~" + method.Name.Value

		// Build method signature
		signature := buildMethodSignature(method)

		// Build snippet for method with parameters
		insertText, insertTextFormat := buildMethodSnippet(method)

		// Add documentation with MarkupContent
		docValue := "**Method**"
		if method.IsClassMethod {
			docValue = "**Class method** (static method)"
		}

		docValue += "\n\n```pascal\n" + signature + "\n```"
		doc := protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: docValue,
		}

		item := protocol.CompletionItem{
			Label:            method.Name.Value,
			Kind:             &kind,
			SortText:         &sortText,
			Detail:           &signature,
			Documentation:    doc,
			InsertText:       &insertText,
			InsertTextFormat: &insertTextFormat,
		}

		items = append(items, item)
	}

	// Extract properties
	for _, prop := range classDecl.Properties {
		if prop.Name == nil {
			continue
		}

		kind := protocol.CompletionItemKindProperty
		sortText := "2property~" + prop.Name.Value

		item := protocol.CompletionItem{
			Label:            prop.Name.Value,
			Kind:             &kind,
			SortText:         &sortText,
			InsertTextFormat: &plainTextFormat,
		}

		// Add type information in detail
		if prop.Type != nil {
			detail := prop.Type.String()
			item.Detail = &detail
		}

		// Add documentation about read/write access with MarkupContent
		docValue := "**Property**"

		accessMode := ""
		if prop.ReadSpec != nil && prop.WriteSpec == nil {
			accessMode = " (read-only)"
		} else if prop.ReadSpec == nil && prop.WriteSpec != nil {
			accessMode = " (write-only)"
		}

		docValue += accessMode
		if prop.Type != nil {
			docValue += "\n\n```pascal\nproperty " + prop.Name.Value + ": " + prop.Type.String() + "\n```"
		}

		doc := protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: docValue,
		}
		item.Documentation = doc

		items = append(items, item)
	}

	return items
}

// extractRecordMembers extracts all fields from a record declaration.
func extractRecordMembers(program *ast.Program, recordName string) []protocol.CompletionItem {
	items := make([]protocol.CompletionItem, 0, 10)

	// Find the record declaration
	var recordDecl *ast.RecordDecl

	ast.Inspect(program, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		if record, ok := node.(*ast.RecordDecl); ok {
			if record.Name != nil && record.Name.Value == recordName {
				recordDecl = record
				return false // Stop searching
			}
		}

		return true
	})

	if recordDecl == nil {
		return items
	}

	// Extract fields
	plainTextFormat := protocol.InsertTextFormatPlainText

	for _, field := range recordDecl.Fields {
		if field.Name == nil {
			continue
		}

		kind := protocol.CompletionItemKindField
		sortText := "0field~" + field.Name.Value

		item := protocol.CompletionItem{
			Label:            field.Name.Value,
			Kind:             &kind,
			SortText:         &sortText,
			InsertTextFormat: &plainTextFormat,
		}

		// Add type information in detail
		if field.Type != nil {
			detail := field.Type.String()
			item.Detail = &detail
		}

		// Add documentation with MarkupContent
		docValue := "**Record field**"
		if field.Type != nil {
			docValue += "\n\n```pascal\n" + field.Name.Value + ": " + field.Type.String() + "\n```"
		}

		doc := protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: docValue,
		}
		item.Documentation = doc

		items = append(items, item)
	}

	return items
}

// buildMethodSignature builds a method signature string for display.
func buildMethodSignature(method *ast.FunctionDecl) string {
	// Start with method name
	var signature strings.Builder
	signature.WriteString(method.Name.Value)

	// Add parameters
	signature.WriteString("(")
	var signatureSb520 strings.Builder

	for i, param := range method.Parameters {
		if i > 0 {
			signatureSb520.WriteString(", ")
		}

		// Add parameter modifiers
		if param.ByRef {
			signatureSb520.WriteString("var ")
		} else if param.IsConst {
			signature.WriteString("const ")
		} else if param.IsLazy {
			signature.WriteString("lazy ")
		}

		// Add parameter name and type
		if param.Name != nil {
			signatureSb520.WriteString(param.Name.Value)
		}

		if param.Type != nil {
			signatureSb520.WriteString(": " + param.Type.String())
		}

		// Add default value if present
		if param.DefaultValue != nil {
			signatureSb520.WriteString(" = " + param.DefaultValue.String())
		}
	}

	signature.WriteString(signatureSb520.String())

	signature.WriteString(")")

	// Add return type
	if method.ReturnType != nil {
		signature.WriteString(": " + method.ReturnType.String())
	}

	return signature.String()
}

// buildMethodSnippet builds an LSP snippet string for method insertion.
// Returns the snippet string and insertTextFormat.
// Example: "MyMethod(${1:param1}, ${2:param2})$0".
func buildMethodSnippet(method *ast.FunctionDecl) (string, protocol.InsertTextFormat) {
	if method.Name == nil {
		return "", protocol.InsertTextFormatPlainText
	}

	// If method has no parameters, use plain text
	if len(method.Parameters) == 0 {
		return method.Name.Value + "()", protocol.InsertTextFormatPlainText
	}

	snippet := method.Name.Value + "("

	var snippetSb572 strings.Builder

	for i, param := range method.Parameters {
		if i > 0 {
			snippetSb572.WriteString(", ")
		}

		// Add tabstop with parameter name as placeholder
		tabstopNum := i + 1

		paramName := "param"
		if param.Name != nil {
			paramName = param.Name.Value
		}

		// Build tabstop: ${1:paramName}
		snippetSb572.WriteString("${" + strconv.Itoa(tabstopNum) + ":" + paramName + "}")
	}

	snippet += snippetSb572.String()

	snippet += snippetFinalCursorPosition

	return snippet, protocol.InsertTextFormatSnippet
}

// sortCompletionItems sorts completion items alphabetically by label.
func sortCompletionItems(items []protocol.CompletionItem) {
	// Use a simple bubble sort since the list is typically small
	for i := range items {
		for j := i + 1; j < len(items); j++ {
			if items[i].Label > items[j].Label {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}
