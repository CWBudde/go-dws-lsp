// Package analysis provides code analysis utilities for the LSP server.
package analysis

import (
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// FindNodeAtPosition traverses the AST to find the most specific (deepest) node
// that contains the given position. Returns nil if no node is found.
//
// The position parameters (line, col) should be 1-based to match the AST coordinate system.
// If you have LSP positions (0-based), add 1 to both line and character before calling this.
func FindNodeAtPosition(program *ast.Program, line, col int) ast.Node {
	if program == nil {
		return nil
	}

	// Create the target position for comparison
	targetPos := token.Position{
		Line:   line,
		Column: col,
	}

	var found ast.Node

	// Use ast.Inspect to traverse the tree
	ast.Inspect(program, func(node ast.Node) bool {
		if node == nil {
			return true // Continue traversal
		}

		// Get node's position range
		start := node.Pos()
		end := node.End()

		// Check if target position is within this node's range
		if !positionInRange(targetPos, start, end) {
			return false // Skip this subtree
		}

		// This node contains the position, so it's a candidate
		// Keep traversing to find more specific (deeper) nodes
		found = node

		return true
	})

	return found
}

// positionInRange checks if pos is within the range [start, end].
// Returns true if start <= pos <= end.
func positionInRange(pos, start, end token.Position) bool {
	// Check if position is before start
	if pos.Line < start.Line {
		return false
	}

	if pos.Line == start.Line && pos.Column < start.Column {
		return false
	}

	// Check if position is after end
	if pos.Line > end.Line {
		return false
	}

	if pos.Line == end.Line && pos.Column > end.Column {
		return false
	}

	return true
}

// GetSymbolName extracts a symbol name from a node, if applicable.
// Returns empty string for nodes that don't represent symbols.
func GetSymbolName(node ast.Node) string {
	if node == nil {
		return ""
	}

	switch n := node.(type) {
	case *ast.Identifier:
		return n.Value

	case *ast.FunctionDecl:
		return n.Name.Value

	case *ast.VarDeclStatement:
		// Variable declarations can have multiple names
		// Return the first one for now
		if len(n.Names) > 0 {
			return n.Names[0].Value
		}

	case *ast.ConstDecl:
		return n.Name.Value

	case *ast.ClassDecl:
		return n.Name.Value

	case *ast.RecordDecl:
		return n.Name.Value

	case *ast.InterfaceDecl:
		return n.Name.Value

	case *ast.PropertyDecl:
		return n.Name.Value

	case *ast.EnumDecl:
		return n.Name.Value

	case *ast.TypeDeclaration:
		return n.Name.Value
	}

	return ""
}
