// Package util provides common utility functions used across the LSP server.
package util

import (
	"github.com/cwbudde/go-dws/pkg/ast"
)

// GetTypeName extracts the type name from a TypeExpression.
// TypeExpression is an interface that can be implemented by TypeAnnotation,
// FunctionPointerTypeNode, or ArrayTypeNode.
func GetTypeName(typeExpr ast.TypeExpression) string {
	if typeExpr == nil {
		return ""
	}

	// Try to type assert to *TypeAnnotation, which has a Name field
	if typeAnnotation, ok := typeExpr.(*ast.TypeAnnotation); ok {
		return typeAnnotation.Name
	}

	// For other types (FunctionPointerTypeNode, ArrayTypeNode), use String()
	return typeExpr.String()
}
