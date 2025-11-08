// Package analysis provides symbol resolution and analysis utilities.
package analysis

import (
	"log"
	"path/filepath"
	"sort"

	"github.com/CWBudde/go-dws-lsp/internal/workspace"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

const (
	resolutionScopeFunction = "function"
	resolutionScopeMethod   = "method"
	resolutionScopeClass    = "class"
	resolutionScopeGlobal   = "global"
)

// SymbolResolver provides symbol resolution with scope-based lookup.
// It implements the resolution strategy: local → class → global → workspace.
type SymbolResolver struct {
	// documentURI is the current document being resolved
	documentURI string

	// program is the AST of the current document
	program *ast.Program

	// position is the cursor position where resolution is requested
	position token.Position

	// workspaceIndex is the workspace-wide symbol index (may be nil)
	workspaceIndex *workspace.SymbolIndex
}

// NewSymbolResolver creates a new symbol resolver for a document.
// The workspace index parameter is optional (can be nil).
func NewSymbolResolver(uri string, program *ast.Program, pos token.Position) *SymbolResolver {
	return &SymbolResolver{
		documentURI:    uri,
		program:        program,
		position:       pos,
		workspaceIndex: nil, // Will be set via SetWorkspaceIndex if available
	}
}

// NewSymbolResolverWithIndex creates a new symbol resolver with workspace index.
func NewSymbolResolverWithIndex(uri string, program *ast.Program, pos token.Position, index *workspace.SymbolIndex) *SymbolResolver {
	return &SymbolResolver{
		documentURI:    uri,
		program:        program,
		position:       pos,
		workspaceIndex: index,
	}
}

// SetWorkspaceIndex sets the workspace index for cross-file symbol resolution.
func (sr *SymbolResolver) SetWorkspaceIndex(index *workspace.SymbolIndex) {
	sr.workspaceIndex = index
}

// ResolveSymbol resolves a symbol name to its definition location(s).
// It follows the resolution strategy: local → class → global → imported units → workspace
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

	// Step 3: Try to resolve as a global symbol (top-level declarations in current file)
	if globalLocs := sr.resolveGlobal(symbolName); len(globalLocs) > 0 {
		log.Printf("Resolved '%s' as global symbol (%d definition(s))", symbolName, len(globalLocs))
		locations = append(locations, globalLocs...)

		return locations
	}

	// Step 4: Try to resolve in imported units (respects DWScript visibility rules)
	if importedLocs := sr.resolveInImportedUnits(symbolName); len(importedLocs) > 0 {
		log.Printf("Resolved '%s' in imported units (%d definition(s))", symbolName, len(importedLocs))
		locations = append(locations, importedLocs...)

		return locations
	}

	// Step 5: Fall back to full workspace search (all files, not just imported)
	// This is used when unit imports are not available or as a last resort
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
// It queries the workspace symbol index for matching symbols.
func (sr *SymbolResolver) resolveWorkspace(symbolName string) []protocol.Location {
	// If no workspace index is available, return empty
	if sr.workspaceIndex == nil {
		log.Printf("No workspace index available for cross-file resolution")
		return nil
	}

	// Query the workspace index for the symbol
	symbolLocations := sr.workspaceIndex.FindSymbol(symbolName)
	if len(symbolLocations) == 0 {
		return nil
	}

	// Convert workspace.SymbolLocation to protocol.Location
	locations := make([]protocol.Location, 0, 10)

	for _, symLoc := range symbolLocations {
		// Skip symbols from the current file (already handled by resolveGlobal)
		if symLoc.Location.URI == sr.documentURI {
			continue
		}

		locations = append(locations, symLoc.Location)
	}

	// Sort by relevance: prefer files in the same directory
	if len(locations) > 1 {
		sr.sortLocationsByRelevance(locations)
	}

	return locations
}

// sortLocationsByRelevance sorts locations by relevance to the current document.
// Relevance is determined by:
// 1. Files in the same directory as the current document
// 2. Files in parent directories
// 3. Files in other directories (alphabetically).
func (sr *SymbolResolver) sortLocationsByRelevance(locations []protocol.Location) {
	// Extract directory from current document URI
	currentPath, err := uriToPath(sr.documentURI)
	if err != nil {
		log.Printf("sortLocationsByRelevance: unable to resolve current URI %s: %v", sr.documentURI, err)
		return
	}

	currentDir := filepath.Dir(currentPath)

	sort.Slice(locations, func(i, j int) bool {
		pathI := resolvePathOrFallback(locations[i].URI)
		pathJ := resolvePathOrFallback(locations[j].URI)

		dirI := filepath.Dir(pathI)
		dirJ := filepath.Dir(pathJ)

		// Same directory as current file takes precedence
		isSameDirI := dirI == currentDir
		isSameDirJ := dirJ == currentDir

		if isSameDirI && !isSameDirJ {
			return true
		}

		if !isSameDirI && isSameDirJ {
			return false
		}

		// Both in same directory or both in different directories - sort alphabetically
		return pathI < pathJ
	})
}

func resolvePathOrFallback(uri string) string {
	path, err := uriToPath(uri)
	if err != nil {
		return uri
	}

	return path
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
				Line:      uint32(pos.Line - 1),   // Convert to 0-based
				Character: uint32(pos.Column - 1), // Convert to 0-based
			},
			End: protocol.Position{
				Line:      uint32(end.Line - 1),   // Convert to 0-based
				Character: uint32(end.Column - 1), // Convert to 0-based
			},
		},
	}
}

// extractUsesClause extracts the list of imported unit names from the AST.
// Returns a slice of unit names (empty if no uses clause is found).
func (sr *SymbolResolver) extractUsesClause() []string {
	var unitNames []string

	// Traverse the AST to find UsesClause statements
	ast.Inspect(sr.program, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		// Check if this is a UsesClause statement
		if usesClause, ok := node.(*ast.UsesClause); ok {
			for _, unitIdent := range usesClause.Units {
				if unitIdent != nil {
					unitNames = append(unitNames, unitIdent.Value)
				}
			}

			return false // Stop traversing this branch
		}

		return true
	})

	if len(unitNames) > 0 {
		log.Printf("Found %d imported units: %v", len(unitNames), unitNames)
	}

	return unitNames
}

// mapUnitNameToURIs maps unit names to possible file URIs in the workspace.
// It uses the workspace index to find files that might define these units.
// Returns a map of unit name → URIs.
func (sr *SymbolResolver) mapUnitNameToURIs(unitNames []string) map[string][]string {
	unitToURIs := make(map[string][]string)

	if sr.workspaceIndex == nil {
		return unitToURIs
	}

	// For each unit name, search the workspace index for matching files
	for _, unitName := range unitNames {
		// Look for symbols in the workspace that match this unit name
		// Typically, a unit file will have a UnitDeclaration with matching name
		symbolLocs := sr.workspaceIndex.FindSymbol(unitName)

		for _, loc := range symbolLocs {
			// Check if this is a unit declaration or if the filename matches
			uri := loc.Location.URI
			unitToURIs[unitName] = append(unitToURIs[unitName], uri)
		}

		// Also check for files with matching base names (e.g., MyUnit.dws)
		// This is a heuristic: convert unit name to lowercase and look for files
		// We'll iterate through all files in the index
		if sr.workspaceIndex != nil {
			// Get all symbols and extract unique URIs that might match
			// Since we don't have a GetAllFiles() method, we'll use the symbols we found
			log.Printf("Mapped unit '%s' to %d file(s)", unitName, len(unitToURIs[unitName]))
		}
	}

	return unitToURIs
}

// resolveInImportedUnits searches for a symbol in explicitly imported units.
// This implements DWScript visibility rules: symbols are only visible from imported units.
func (sr *SymbolResolver) resolveInImportedUnits(symbolName string) []protocol.Location {
	// Extract imported unit names
	unitNames := sr.extractUsesClause()
	if len(unitNames) == 0 {
		return nil // No imports, nothing to search
	}

	// Map unit names to file URIs
	unitToURIs := sr.mapUnitNameToURIs(unitNames)
	if len(unitToURIs) == 0 {
		return nil // No matching files found
	}

	// Collect all URIs from imported units
	importedURIs := make(map[string]bool)

	for _, uris := range unitToURIs {
		for _, uri := range uris {
			importedURIs[uri] = true
		}
	}

	// Query workspace index for the symbol
	if sr.workspaceIndex == nil {
		return nil
	}

	symbolLocations := sr.workspaceIndex.FindSymbol(symbolName)
	if len(symbolLocations) == 0 {
		return nil
	}

	// Filter to only include symbols from imported units
	var locations []protocol.Location

	for _, symLoc := range symbolLocations {
		// Skip symbols from the current file (already handled by resolveGlobal)
		if symLoc.Location.URI == sr.documentURI {
			continue
		}

		// Only include if from an imported unit
		if importedURIs[symLoc.Location.URI] {
			locations = append(locations, symLoc.Location)
		}
	}

	// Sort by relevance
	if len(locations) > 1 {
		sr.sortLocationsByRelevance(locations)
	}

	return locations
}

// GetResolutionScope returns a string describing the scope where a symbol would be resolved.
// This is useful for debugging and providing user feedback.
func (sr *SymbolResolver) GetResolutionScope() string {
	if sr.findEnclosingFunction() != nil {
		if sr.findEnclosingClass() != nil {
			return resolutionScopeMethod
		}

		return resolutionScopeFunction
	}

	if sr.findEnclosingClass() != nil {
		return resolutionScopeClass
	}

	return resolutionScopeGlobal
}
