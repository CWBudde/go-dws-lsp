// Package lsp provides symbol identification utilities for LSP operations.
package lsp

import (
	"fmt"
	"log"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// SymbolKind represents the kind of symbol found at a position.
type SymbolKind string

const (
	SymbolKindVariable  SymbolKind = "variable"
	SymbolKindFunction  SymbolKind = "function"
	SymbolKindParameter SymbolKind = "parameter"
	SymbolKindClass     SymbolKind = "class"
	SymbolKindField     SymbolKind = "field"
	SymbolKindMethod    SymbolKind = "method"
	SymbolKindProperty  SymbolKind = "property"
	SymbolKindConstant  SymbolKind = "constant"
	SymbolKindEnum      SymbolKind = "enum"
	SymbolKindEnumValue SymbolKind = "enumValue"
	SymbolKindType      SymbolKind = "type"
	SymbolKindInterface SymbolKind = "interface"
	SymbolKindUnknown   SymbolKind = "unknown"
)

// SymbolInfo contains information about a symbol at a cursor position.
type SymbolInfo struct {
	Name string     // The symbol name
	Kind SymbolKind // The kind of symbol
	Node ast.Node   // The AST node
}

// IdentifySymbolAtPosition identifies what symbol is at the given AST node.
// This extracts the symbol name, determines its kind, and prepares it for
// definition lookup.
func IdentifySymbolAtPosition(node ast.Node) *SymbolInfo {
	if node == nil {
		return nil
	}

	var info *SymbolInfo

	switch n := node.(type) {
	case *ast.Identifier:
		// An identifier reference - we need to find what it refers to
		info = &SymbolInfo{
			Name: n.Value,
			Kind: SymbolKindUnknown, // Will be determined by context
			Node: n,
		}
		log.Printf("Identified identifier: %s", n.Value)

	case *ast.VarDeclStatement:
		// Variable declaration - extract the first name
		if len(n.Names) > 0 {
			info = &SymbolInfo{
				Name: n.Names[0].Value,
				Kind: SymbolKindVariable,
				Node: n,
			}
			log.Printf("Identified variable declaration: %s", n.Names[0].Value)
		}

	case *ast.FunctionDecl:
		// Function declaration
		if n.Name != nil {
			info = &SymbolInfo{
				Name: n.Name.Value,
				Kind: SymbolKindFunction,
				Node: n,
			}
			log.Printf("Identified function declaration: %s", n.Name.Value)
		}

	case *ast.ClassDecl:
		// Class declaration
		if n.Name != nil {
			info = &SymbolInfo{
				Name: n.Name.Value,
				Kind: SymbolKindClass,
				Node: n,
			}
			log.Printf("Identified class declaration: %s", n.Name.Value)
		}

	case *ast.ConstDecl:
		// Constant declaration
		if n.Name != nil {
			info = &SymbolInfo{
				Name: n.Name.Value,
				Kind: SymbolKindConstant,
				Node: n,
			}
			log.Printf("Identified constant declaration: %s", n.Name.Value)
		}

	case *ast.EnumDecl:
		// Enum declaration
		if n.Name != nil {
			info = &SymbolInfo{
				Name: n.Name.Value,
				Kind: SymbolKindEnum,
				Node: n,
			}
			log.Printf("Identified enum declaration: %s", n.Name.Value)
		}

	case *ast.FieldDecl:
		// Field declaration
		if n.Name != nil {
			info = &SymbolInfo{
				Name: n.Name.Value,
				Kind: SymbolKindField,
				Node: n,
			}
			log.Printf("Identified field declaration: %s", n.Name.Value)
		}

	case *ast.MemberAccessExpression:
		// Member access (e.g., obj.field or obj.method)
		if n.Member != nil {
			info = &SymbolInfo{
				Name: n.Member.Value,
				Kind: SymbolKindUnknown, // Could be field, method, or property
				Node: n,
			}
			log.Printf("Identified member access expression: %s", n.Member.Value)
		}

	case *ast.CallExpression:
		// Function/method call - extract the function name
		if funcIdent, ok := n.Function.(*ast.Identifier); ok {
			info = &SymbolInfo{
				Name: funcIdent.Value,
				Kind: SymbolKindFunction,
				Node: n,
			}
			log.Printf("Identified function call: %s", funcIdent.Value)
		} else if memberExpr, ok := n.Function.(*ast.MemberAccessExpression); ok {
			// Method call (e.g., obj.method())
			if memberExpr.Member != nil {
				info = &SymbolInfo{
					Name: memberExpr.Member.Value,
					Kind: SymbolKindMethod,
					Node: n,
				}
				log.Printf("Identified method call: %s", memberExpr.Member.Value)
			}
		}

	default:
		// Not a symbol we can identify
		log.Printf("Node at position is not a recognized symbol: %T", node)
		return nil
	}

	if info != nil {
		log.Printf("Symbol identified - Name: %s, Kind: %s", info.Name, info.Kind)
	}

	return info
}

// ExtractSymbolName extracts the symbol name from a node, if possible.
// Returns empty string if the node doesn't represent a named symbol.
func ExtractSymbolName(node ast.Node) string {
	if node == nil {
		return ""
	}

	switch n := node.(type) {
	case *ast.Identifier:
		return n.Value

	case *ast.VarDeclStatement:
		if len(n.Names) > 0 {
			return n.Names[0].Value
		}

	case *ast.FunctionDecl:
		if n.Name != nil {
			return n.Name.Value
		}

	case *ast.ClassDecl:
		if n.Name != nil {
			return n.Name.Value
		}

	case *ast.ConstDecl:
		if n.Name != nil {
			return n.Name.Value
		}

	case *ast.EnumDecl:
		if n.Name != nil {
			return n.Name.Value
		}

	case *ast.FieldDecl:
		if n.Name != nil {
			return n.Name.Value
		}

	case *ast.MemberAccessExpression:
		if n.Member != nil {
			return n.Member.Value
		}

	case *ast.CallExpression:
		if funcIdent, ok := n.Function.(*ast.Identifier); ok {
			return funcIdent.Value
		} else if memberExpr, ok := n.Function.(*ast.MemberAccessExpression); ok {
			if memberExpr.Member != nil {
				return memberExpr.Member.Value
			}
		}
	}

	return ""
}

// GetSymbolContext returns contextual information about where a symbol appears.
// This can help determine scope and resolution strategy.
func GetSymbolContext(node ast.Node, program *ast.Program) string {
	if node == nil || program == nil {
		return "unknown"
	}

	// Determine context based on node type and position
	switch node.(type) {
	case *ast.MemberAccessExpression:
		return "member"
	case *ast.CallExpression:
		return "call"
	default:
		return "reference"
	}
}

// IsDeclaration checks if a node represents a declaration (as opposed to a reference).
func IsDeclaration(node ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.(type) {
	case *ast.VarDeclStatement,
		*ast.FunctionDecl,
		*ast.ClassDecl,
		*ast.ConstDecl,
		*ast.EnumDecl,
		*ast.FieldDecl:
		return true
	default:
		return false
	}
}

// String returns a string representation of SymbolInfo for debugging.
func (si *SymbolInfo) String() string {
	if si == nil {
		return "<nil>"
	}

	return fmt.Sprintf("Symbol{Name: %s, Kind: %s}", si.Name, si.Kind)
}
