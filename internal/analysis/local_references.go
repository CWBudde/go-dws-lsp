package analysis

import (
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// FindLocalReferences scans within the given scope node (typically an enclosing
// function/block) to find Identifier occurrences matching symbolName.
// It returns LSP ranges for each match. Matches in nested blocks that declare
// the same name (shadowing) are excluded.
func FindLocalReferences(program *ast.Program, symbolName string, scopeNode ast.Node) []protocol.Range {
	if program == nil || scopeNode == nil || symbolName == "" {
		return nil
	}

	// Determine traversal root
	var root ast.Node

	switch n := scopeNode.(type) {
	case *ast.FunctionDecl:
		root = n.Body
	case *ast.BlockStatement:
		root = n
	default:
		root = scopeNode
	}

	if root == nil {
		return nil
	}

	// Track nested block spans that shadow the symbol via local declaration
	type blockSpan struct {
		start token.Position
		end   token.Position
	}
	var blockStack []blockSpan
	var shadowBlocks []blockSpan

	// First pass: collect blocks that declare a variable with the same name
	ast.Inspect(root, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		// Do not descend into nested function declarations
		if fn, ok := node.(*ast.FunctionDecl); ok {
			// Skip nested functions entirely from shadow analysis
			// (they form independent scopes)
			_ = fn
			return false
		}

		// Maintain block stack
		if blk, ok := node.(*ast.BlockStatement); ok {
			// Enter block
			blockStack = append(blockStack, blockSpan{start: blk.Pos(), end: blk.End()})
			// Continue to inspect inside this block
			return true
		}

		// On leaving a block, ast.Inspect does not give explicit exit events,
		// so we handle stack trimming by checking containment when we move up.
		// We cannot reliably detect exits, so we keep all blocks encountered.

		// Record blocks that declare the same name
		if vd, ok := node.(*ast.VarDeclStatement); ok {
			// If this var declaration contains the same name, mark the current (innermost) block as a shadowing block
			for _, name := range vd.Names {
				if name.Value == symbolName {
					if len(blockStack) > 0 {
						b := blockStack[len(blockStack)-1]
						shadowBlocks = append(shadowBlocks, b)
					}

					break
				}
			}
		}

		return true
	})

	// Helper to know if a position is within a shadowed block
	inShadow := func(pos token.Position) bool {
		for _, sb := range shadowBlocks {
			if positionInRange(pos, sb.start, sb.end) {
				return true
			}
		}

		return false
	}

	// Second pass: collect identifiers matching the name, excluding shadowed blocks
	var ranges []protocol.Range

	ast.Inspect(root, func(node ast.Node) bool {
		if node == nil {
			return false
		}
		// Skip nested functions entirely
		if _, ok := node.(*ast.FunctionDecl); ok {
			return false
		}
		// Match identifiers
		if ident, ok := node.(*ast.Identifier); ok && ident != nil {
			if ident.Value == symbolName {
				pos := ident.Pos()
				if inShadow(pos) {
					return true
				}

				end := ident.End()
				ranges = append(ranges, protocol.Range{
					Start: protocol.Position{Line: uint32(max(0, pos.Line-1)), Character: uint32(max(0, pos.Column-1))},
					End:   protocol.Position{Line: uint32(max(0, end.Line-1)), Character: uint32(max(0, end.Column-1))},
				})
			}
		}

		return true
	})

	return ranges
}
