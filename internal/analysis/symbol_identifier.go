// Package analysis provides code analysis utilities for the LSP server.
package analysis

import (
    "github.com/cwbudde/go-dws/pkg/ast"
)

// SymbolKind categorizes the kind of symbol identified.
type SymbolKind string

const (
    SymbolKindIdentifier SymbolKind = "Identifier"
    SymbolKindFunction   SymbolKind = "Function"
    SymbolKindVariable   SymbolKind = "Variable"
    SymbolKindConstant   SymbolKind = "Constant"
    SymbolKindClass      SymbolKind = "Class"
    SymbolKindRecord     SymbolKind = "Record"
    SymbolKindEnum       SymbolKind = "Enum"
    SymbolKindProperty   SymbolKind = "Property"
    SymbolKindType       SymbolKind = "Type"
    SymbolKindInterface  SymbolKind = "Interface"
    SymbolKindUnknown    SymbolKind = "Unknown"
)

// SymbolInfo describes a symbol found at a position.
type SymbolInfo struct {
    Name string
    Kind SymbolKind
    Node ast.Node
}

// IdentifySymbolAtPosition returns the symbol information at a given 1-based position.
// If no symbol is found, it returns nil.
func IdentifySymbolAtPosition(program *ast.Program, line, col int) *SymbolInfo {
    if program == nil {
        return nil
    }

    node := FindNodeAtPosition(program, line, col)
    if node == nil {
        return nil
    }

    name := GetSymbolName(node)
    kind := classifyNodeKind(node)

    // If we couldn't classify from the parent node, and it is an Identifier,
    // use that as the name and a generic Identifier kind.
    if name == "" {
        if ident, ok := node.(*ast.Identifier); ok && ident != nil {
            name = ident.Value
            if kind == SymbolKindUnknown {
                kind = SymbolKindIdentifier
            }
        }
    }

    if name == "" {
        return nil
    }

    return &SymbolInfo{Name: name, Kind: kind, Node: node}
}

// classifyNodeKind maps AST node types to SymbolKind.
func classifyNodeKind(node ast.Node) SymbolKind {
    switch node.(type) {
    case *ast.Identifier:
        return SymbolKindIdentifier
    case *ast.FunctionDecl:
        return SymbolKindFunction
    case *ast.VarDeclStatement:
        return SymbolKindVariable
    case *ast.ConstDecl:
        return SymbolKindConstant
    case *ast.ClassDecl:
        return SymbolKindClass
    case *ast.RecordDecl:
        return SymbolKindRecord
    case *ast.EnumDecl:
        return SymbolKindEnum
    case *ast.PropertyDecl:
        return SymbolKindProperty
    case *ast.TypeDeclaration:
        return SymbolKindType
    case *ast.InterfaceDecl:
        return SymbolKindInterface
    default:
        return SymbolKindUnknown
    }
}

