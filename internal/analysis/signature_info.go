package analysis

import (
	"log"

	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/CWBudde/go-dws-lsp/internal/workspace"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// FunctionSignature holds information about a function's signature
type FunctionSignature struct {
	// Function name
	Name string

	// Parameters with names and types
	Parameters []ParameterInfo

	// Return type (empty string for procedures)
	ReturnType string

	// Documentation comment if available
	Documentation string

	// Whether this is a method (has a receiver/object)
	IsMethod bool

	// For methods, the class/type name
	ClassName string
}

// ParameterInfo holds information about a single parameter
type ParameterInfo struct {
	Name         string
	Type         string
	DefaultValue string // Optional default value
	IsOptional   bool
}

// GetFunctionSignature retrieves function definition to get parameters and documentation
// This implements task 10.8 - it reuses symbol resolution from go-to-definition (Phase 5)
func GetFunctionSignature(doc *server.Document, functionName string, line, character int, workspaceIndex *workspace.SymbolIndex) (*FunctionSignature, error) {
	if doc.Program == nil {
		log.Printf("GetFunctionSignature: No program available\n")
		return nil, nil
	}

	programAST := doc.Program.AST()
	if programAST == nil {
		log.Printf("GetFunctionSignature: No AST available\n")
		return nil, nil
	}

	// Convert to AST position (1-based)
	astLine := line + 1
	astColumn := character + 1
	pos := token.Position{Line: astLine, Column: astColumn}

	log.Printf("GetFunctionSignature: Looking for function '%s' at %d:%d\n", functionName, astLine, astColumn)

	// Use SymbolResolver to find the function definition
	var resolver *SymbolResolver
	if workspaceIndex != nil {
		resolver = NewSymbolResolverWithIndex(doc.URI, programAST, pos, workspaceIndex)
	} else {
		resolver = NewSymbolResolver(doc.URI, programAST, pos)
	}

	locations := resolver.ResolveSymbol(functionName)
	if len(locations) == 0 {
		log.Printf("GetFunctionSignature: Function '%s' not found (may be built-in)\n", functionName)
		return nil, nil
	}

	// Get the first location (functions should have a single definition)
	location := locations[0]
	log.Printf("GetFunctionSignature: Found definition at %s:%d:%d\n", location.URI, location.Range.Start.Line, location.Range.Start.Character)

	// Find the AST node at the definition location
	// Convert back to 1-based for AST
	defLine := int(location.Range.Start.Line) + 1
	defColumn := int(location.Range.Start.Character) + 1
	defPos := token.Position{Line: defLine, Column: defColumn}

	// Find the function declaration node
	funcDecl := findFunctionDeclarationAtPosition(programAST, defPos)
	if funcDecl == nil {
		log.Printf("GetFunctionSignature: Could not find function declaration AST node\n")
		return nil, nil
	}

	// Extract signature information from the function declaration
	signature := extractSignatureFromDeclaration(funcDecl)
	if signature != nil {
		signature.Name = functionName
		log.Printf("GetFunctionSignature: Extracted signature for '%s' with %d parameters\n", functionName, len(signature.Parameters))
	}

	return signature, nil
}

// findFunctionDeclarationAtPosition finds a function or method declaration at the given position
func findFunctionDeclarationAtPosition(program *ast.Program, pos token.Position) *ast.FunctionDecl {
	var funcDecl *ast.FunctionDecl

	ast.Inspect(program, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		// Check if this node contains the position
		if !positionInRange(pos, n.Pos(), n.End()) {
			return false
		}

		// Check if this is a function declaration
		if fd, ok := n.(*ast.FunctionDecl); ok {
			funcDecl = fd
			// Continue searching for more specific match
		}

		return true
	})

	return funcDecl
}

// extractSignatureFromDeclaration extracts signature information from a function declaration
func extractSignatureFromDeclaration(funcDecl *ast.FunctionDecl) *FunctionSignature {
	if funcDecl == nil {
		return nil
	}

	signature := &FunctionSignature{
		Parameters: []ParameterInfo{},
	}

	// Extract parameters from FunctionDecl
	// In the go-dws AST, Parameters is a slice of Parameter structs
	for _, param := range funcDecl.Parameters {
		// Get parameter type
		paramType := ""
		if param.Type != nil && param.Type.Name != "" {
			paramType = param.Type.Name
		}

		// Get parameter name
		paramName := ""
		if param.Name != nil {
			paramName = param.Name.Value
		}

		paramInfo := ParameterInfo{
			Name: paramName,
			Type: paramType,
		}

		// Check for default value
		if param.DefaultValue != nil {
			paramInfo.IsOptional = true
			paramInfo.DefaultValue = param.DefaultValue.String()
		}

		signature.Parameters = append(signature.Parameters, paramInfo)
	}

	// Extract return type
	if funcDecl.ReturnType != nil && funcDecl.ReturnType.Name != "" {
		signature.ReturnType = funcDecl.ReturnType.Name
	}

	// Extract documentation from leading comments
	// TODO: Implement documentation extraction from comments
	// This would require access to the token stream or comment nodes
	signature.Documentation = ""

	return signature
}

// Note: extractTypeName is not needed since we get type names directly from
// param.Type.Name and funcDecl.ReturnType.Name in the go-dws AST
