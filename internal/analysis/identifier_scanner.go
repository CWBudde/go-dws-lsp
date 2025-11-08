package analysis

import (
	"errors"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// ScanASTForIdentifier walks the AST and returns 1-based positions for all identifiers
// that match the provided name. It is a simple helper that future reference search
// tasks can re-use for quick scans before more expensive filtering.
func ScanASTForIdentifier(program *ast.Program, name string) ([]Position, error) {
	if program == nil {
		return nil, errors.New("analysis: program AST is nil")
	}

	if name == "" {
		return nil, errors.New("analysis: identifier name is empty")
	}

	var positions []Position

	ast.Inspect(program, func(node ast.Node) bool {
		ident, ok := node.(*ast.Identifier)
		if !ok || ident == nil {
			return true
		}

		if ident.Value == name {
			pos := ident.Pos()
			positions = append(positions, Position{Line: pos.Line, Column: pos.Column})
		}

		return true
	})

	return positions, nil
}
