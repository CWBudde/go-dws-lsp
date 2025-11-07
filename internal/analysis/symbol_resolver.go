// Package analysis provides symbol resolution and analysis utilities.
package analysis

import (
	"log"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// SymbolResolver provides symbol resolution with scope-based lookup.
// It implements the resolution strategy: local → class → global → workspace
type SymbolResolver struct {
	// documentURI is the current document being resolved
	documentURI string

	// program is the AST of the current document
	program *ast.Program

	// position is the cursor position where resolution is requested
	position token.Position
}

// NewSymbolResolver creates a new symbol resolver for a document.
func NewSymbolResolver(uri string, program *ast.Program, pos token.Position) *SymbolResolver {
	return &SymbolResolver{
		documentURI: uri,
		program:     program,
		position:    pos,
	}
}

// ResolveSymbol resolves a symbol name to its definition location(s).
// It follows the resolution strategy: local → class → global → workspace
// Returns a slice of locations (may be empty if not found).
func (sr *SymbolResolver) ResolveSymbol(symbolName string) []protocol.Location {
	log.Printf("Resolving symbol '%s' at position %d:%d", symbolName, sr.position.Line, sr.position.Column)

	if sr.program == nil {
		log.Println("Cannot resolve symbol: program AST is nil")
		return nil
	}

	var locations []protocol.Location

	// Step 1: Try to resolve as a local symbol (variables, parameters in current scope)
	if localLoc := sr.resolveLocal(symbolName); localLoc != nil {
		log.Printf("Resolved '%s' as local symbol", symbolName)
		locations = append(locations, *localLoc)
		return locations // Local symbols take precedence, stop here
	}

	// Step 2: Try to resolve as a class member (if we're inside a class)
	if classMemberLoc := sr.resolveClassMember(symbolName); classMemberLoc != nil {
		log.Printf("Resolved '%s' as class member", symbolName)
		locations = append(locations, *classMemberLoc)
		return locations // Class members take precedence over globals
	}

	// Step 3: Try to resolve as a global symbol (top-level declarations)
	if globalLocs := sr.resolveGlobal(symbolName); len(globalLocs) > 0 {
		log.Printf("Resolved '%s' as global symbol (%d definition(s))", symbolName, len(globalLocs))
		locations = append(locations, globalLocs...)
		return locations
	}

	// Step 4: Try to resolve in workspace (other files)
	// NOTE: This will be implemented in later tasks when we have workspace indexing
	// For now, we just return what we found (if anything)
	if workspaceLocs := sr.resolveWorkspace(symbolName); len(workspaceLocs) > 0 {
		log.Printf("Resolved '%s' in workspace (%d definition(s))", symbolName, len(workspaceLocs))
		locations = append(locations, workspaceLocs...)
		return locations
	}

	log.Printf("Could not resolve symbol '%s'", symbolName)
	return locations // Empty if not found
}

// resolveLocal attempts to resolve a symbol in the local scope.
// This includes function parameters and local variables in the current function/block.
func (sr *SymbolResolver) resolveLocal(symbolName string) *protocol.Location {
	// Find the enclosing function at the cursor position
	enclosingFunc := sr.findEnclosingFunction()
	if enclosingFunc == nil {
		// Not inside a function, no local scope
		return nil
	}

	// Check function parameters
	for _, param := range enclosingFunc.Parameters {
		if param.Name != nil && param.Name.Value == symbolName {
			return sr.nodeToLocation(param.Name)
		}
	}

	// Check local variable declarations in function body
	if enclosingFunc.Body != nil {
		if localVar := sr.findLocalVariable(enclosingFunc.Body, symbolName); localVar != nil {
			return localVar
		}
	}

	return nil
}

// resolveClassMember attempts to resolve a symbol as a class member.
// This is used when the cursor is inside a class method.
func (sr *SymbolResolver) resolveClassMember(symbolName string) *protocol.Location {
	// First, try to find the enclosing class at the cursor position
	enclosingClass := sr.findEnclosingClass()

	// If not directly inside a class, check if we're in a method implementation
	if enclosingClass == nil {
		enclosingFunc := sr.findEnclosingFunction()
		if enclosingFunc != nil && enclosingFunc.ClassName != nil {
			// We're in a method implementation (function TClassName.MethodName)
			enclosingClass = sr.findClassByName(enclosingFunc.ClassName.Value)
		}
	}

	if enclosingClass == nil {
		// Not inside a class or method
		return nil
	}

	// Check class fields
	for _, field := range enclosingClass.Fields {
		if field.Name != nil && field.Name.Value == symbolName {
			return sr.nodeToLocation(field.Name)
		}
	}

	// Check class methods
	for _, method := range enclosingClass.Methods {
		if method.Name != nil && method.Name.Value == symbolName {
			return sr.nodeToLocation(method.Name)
		}
	}

	// Check class properties
	for _, prop := range enclosingClass.Properties {
		if prop.Name != nil && prop.Name.Value == symbolName {
			return sr.nodeToLocation(prop.Name)
		}
	}

	// Check parent class members (inheritance)
	if enclosingClass.Parent != nil {
		if parentLoc := sr.resolveInheritedMember(enclosingClass.Parent.Value, symbolName); parentLoc != nil {
			return parentLoc
		}
	}

	return nil
}

// resolveInheritedMember searches for a symbol in parent class hierarchy.
// It recursively searches parent classes for the given symbol.
func (sr *SymbolResolver) resolveInheritedMember(parentClassName string, symbolName string) *protocol.Location {
	// Find the parent class declaration in the program
	parentClass := sr.findClassByName(parentClassName)
	if parentClass == nil {
		log.Printf("Parent class '%s' not found in current file", parentClassName)
		return nil
	}

	// Check parent class fields
	for _, field := range parentClass.Fields {
		if field.Name != nil && field.Name.Value == symbolName {
			return sr.nodeToLocation(field.Name)
		}
	}

	// Check parent class methods
	for _, method := range parentClass.Methods {
		if method.Name != nil && method.Name.Value == symbolName {
			return sr.nodeToLocation(method.Name)
		}
	}

	// Check parent class properties
	for _, prop := range parentClass.Properties {
		if prop.Name != nil && prop.Name.Value == symbolName {
			return sr.nodeToLocation(prop.Name)
		}
	}

	// Recursively check the parent's parent (grandparent)
	if parentClass.Parent != nil {
		return sr.resolveInheritedMember(parentClass.Parent.Value, symbolName)
	}

	return nil
}

// findClassByName finds a class declaration by name in the program.
func (sr *SymbolResolver) findClassByName(className string) *ast.ClassDecl {
	for _, stmt := range sr.program.Statements {
		if classDecl, ok := stmt.(*ast.ClassDecl); ok {
			if classDecl.Name != nil && classDecl.Name.Value == className {
				return classDecl
			}
		}
	}
	return nil
}

// resolveGlobal attempts to resolve a symbol at the global (file) level.
// This includes top-level functions, classes, constants, and variables.
func (sr *SymbolResolver) resolveGlobal(symbolName string) []protocol.Location {
	var locations []protocol.Location

	// Traverse all top-level statements
	for _, stmt := range sr.program.Statements {
		if loc := sr.checkStatementForSymbol(stmt, symbolName); loc != nil {
			locations = append(locations, *loc)
		}
	}

	return locations
}

// resolveWorkspace attempts to resolve a symbol in other files in the workspace.
// This is a placeholder for future workspace-wide symbol indexing.
func (sr *SymbolResolver) resolveWorkspace(symbolName string) []protocol.Location {
	// TODO: Implement workspace-wide symbol resolution
	// This will require:
	// - A workspace symbol index
	// - Cross-file import/reference tracking
	// - Concurrent file parsing
	// For now, return empty
	return nil
}

// findEnclosingFunction finds the function that contains the cursor position.
func (sr *SymbolResolver) findEnclosingFunction() *ast.FunctionDecl {
	var enclosingFunc *ast.FunctionDecl

	ast.Inspect(sr.program, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		// Check if this is a function declaration
		if funcDecl, ok := node.(*ast.FunctionDecl); ok {
			// Check if the cursor is within this function's range
			if sr.isPositionInRange(funcDecl) {
				enclosingFunc = funcDecl
				// Continue traversal to find inner functions (if nested)
			}
		}

		return true
	})

	return enclosingFunc
}

// findEnclosingClass finds the class that contains the cursor position.
func (sr *SymbolResolver) findEnclosingClass() *ast.ClassDecl {
	var enclosingClass *ast.ClassDecl

	ast.Inspect(sr.program, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		// Check if this is a class declaration
		if classDecl, ok := node.(*ast.ClassDecl); ok {
			// Check if the cursor is within this class's range
			if sr.isPositionInRange(classDecl) {
				enclosingClass = classDecl
				return false // Found the class, stop traversal
			}
		}

		return true
	})

	return enclosingClass
}

// findLocalVariable searches for a local variable declaration in a block.
func (sr *SymbolResolver) findLocalVariable(block *ast.BlockStatement, symbolName string) *protocol.Location {
	if block == nil {
		return nil
	}

	// Search through statements in the block
	for _, stmt := range block.Statements {
		// Check if it's a variable declaration
		if varDecl, ok := stmt.(*ast.VarDeclStatement); ok {
			for _, name := range varDecl.Names {
				if name.Value == symbolName {
					// Only return if the declaration is before the cursor position
					if sr.isBeforeCursor(name) {
						return sr.nodeToLocation(name)
					}
				}
			}
		}

		// TODO: Handle nested blocks (if/while/for statements)
		// This would require recursive traversal of block statements
	}

	return nil
}

// checkStatementForSymbol checks if a statement declares the given symbol.
func (sr *SymbolResolver) checkStatementForSymbol(stmt ast.Statement, symbolName string) *protocol.Location {
	switch s := stmt.(type) {
	case *ast.VarDeclStatement:
		for _, name := range s.Names {
			if name.Value == symbolName {
				return sr.nodeToLocation(name)
			}
		}

	case *ast.FunctionDecl:
		if s.Name != nil && s.Name.Value == symbolName {
			return sr.nodeToLocation(s.Name)
		}

	case *ast.ClassDecl:
		if s.Name != nil && s.Name.Value == symbolName {
			return sr.nodeToLocation(s.Name)
		}

	case *ast.ConstDecl:
		if s.Name != nil && s.Name.Value == symbolName {
			return sr.nodeToLocation(s.Name)
		}

	case *ast.EnumDecl:
		if s.Name != nil && s.Name.Value == symbolName {
			return sr.nodeToLocation(s.Name)
		}
		// Also check enum values
		for _, enumVal := range s.Values {
			if enumVal.Name == symbolName {
				return sr.nodeToLocation(s) // Return enum declaration location
			}
		}

	case *ast.RecordDecl:
		if s.Name != nil && s.Name.Value == symbolName {
			return sr.nodeToLocation(s.Name)
		}

	case *ast.InterfaceDecl:
		if s.Name != nil && s.Name.Value == symbolName {
			return sr.nodeToLocation(s.Name)
		}

	case *ast.ArrayDecl:
		if s.Name != nil && s.Name.Value == symbolName {
			return sr.nodeToLocation(s.Name)
		}

	case *ast.SetDecl:
		if s.Name != nil && s.Name.Value == symbolName {
			return sr.nodeToLocation(s.Name)
		}

	case *ast.HelperDecl:
		if s.Name != nil && s.Name.Value == symbolName {
			return sr.nodeToLocation(s.Name)
		}
	}

	return nil
}

// isPositionInRange checks if the resolver's cursor position is within a node's range.
func (sr *SymbolResolver) isPositionInRange(node ast.Node) bool {
	start := node.Pos()
	end := node.End()

	// Check if cursor is within the range
	if sr.position.Line < start.Line || sr.position.Line > end.Line {
		return false
	}

	if sr.position.Line == start.Line && sr.position.Column < start.Column {
		return false
	}

	if sr.position.Line == end.Line && sr.position.Column > end.Column {
		return false
	}

	return true
}

// isBeforeCursor checks if a node is before the cursor position.
// This is used to ensure we only consider declarations that come before the reference.
func (sr *SymbolResolver) isBeforeCursor(node ast.Node) bool {
	pos := node.Pos()

	if pos.Line < sr.position.Line {
		return true
	}

	if pos.Line == sr.position.Line && pos.Column <= sr.position.Column {
		return true
	}

	return false
}

// nodeToLocation converts an AST node to an LSP Location.
func (sr *SymbolResolver) nodeToLocation(node ast.Node) *protocol.Location {
	if node == nil {
		return nil
	}

	pos := node.Pos()
	end := node.End()

	return &protocol.Location{
		URI: sr.documentURI,
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(pos.Line - 1),    // Convert to 0-based
				Character: uint32(pos.Column - 1),  // Convert to 0-based
			},
			End: protocol.Position{
				Line:      uint32(end.Line - 1),    // Convert to 0-based
				Character: uint32(end.Column - 1),  // Convert to 0-based
			},
		},
	}
}

// GetResolutionScope returns a string describing the scope where a symbol would be resolved.
// This is useful for debugging and providing user feedback.
func (sr *SymbolResolver) GetResolutionScope() string {
	if sr.findEnclosingFunction() != nil {
		if sr.findEnclosingClass() != nil {
			return "method"
		}
		return "function"
	}

	if sr.findEnclosingClass() != nil {
		return "class"
	}

	return "global"
}
