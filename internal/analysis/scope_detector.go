// Package analysis provides scope detection utilities.
package analysis

import (
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// Position represents a 1-based position (line, column) in source.
type Position struct {
	Line   int
	Column int
}

// ScopeType classifies the scope of a symbol.
type ScopeType int

const (
	ScopeUnknown ScopeType = iota
	ScopeLocal
	ScopeGlobal
	ScopeClassMember
	ScopeParameter
)

func (s ScopeType) String() string {
	switch s {
	case ScopeLocal:
		return "Local"
	case ScopeGlobal:
		return "Global"
	case ScopeClassMember:
		return "ClassMember"
	case ScopeParameter:
		return "Parameter"
	default:
		return "Unknown"
	}
}

// ScopeInfo describes the detected scope and enclosing constructs.
type ScopeInfo struct {
	Type     ScopeType
	Function *ast.FunctionDecl
	Class    *ast.ClassDecl
}

// DetermineScope determines the scope of a symbol at the given position.
// It returns basic categorization (Local, Global, ClassMember, Parameter) and
// the most specific enclosing function and/or class if present.
func DetermineScope(program *ast.Program, symbolName string, pos Position) *ScopeInfo {
	if program == nil {
		return &ScopeInfo{Type: ScopeUnknown}
	}

	// Build a token.Position for comparisons
	target := token.Position{Line: pos.Line, Column: pos.Column}

	var (
		bestFunc      *ast.FunctionDecl
		bestClass     *ast.ClassDecl
		bestFuncSpan  = spanLarge()
		bestClassSpan = spanLarge()
	)

	// Collect the most specific enclosing function and class that contain the position
	ast.Inspect(program, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		start, end := n.Pos(), n.End()
		if !positionInRange(target, start, end) {
			return false
		}

		switch typed := n.(type) {
		case *ast.FunctionDecl:
			if moreSpecific(start, end, bestFuncSpan.start, bestFuncSpan.end) {
				bestFunc = typed
				bestFuncSpan = span{start: start, end: end}
			}
		case *ast.ClassDecl:
			if moreSpecific(start, end, bestClassSpan.start, bestClassSpan.end) {
				bestClass = typed
				bestClassSpan = span{start: start, end: end}
			}
		}

		// Continue searching deeper for more specific matches
		return true
	})

	// Decide scope type
	// Parameter takes precedence if inside a function and matches a parameter name
	if bestFunc != nil {
		for _, p := range bestFunc.Parameters {
			if p.Name != nil && p.Name.Value == symbolName {
				return &ScopeInfo{Type: ScopeParameter, Function: bestFunc, Class: bestClass}
			}
		}
		// Otherwise it's a local (variable or local reference) within the function
		return &ScopeInfo{Type: ScopeLocal, Function: bestFunc, Class: bestClass}
	}

	// If inside a class but not inside a method, it's a class member context
	if bestClass != nil {
		return &ScopeInfo{Type: ScopeClassMember, Function: nil, Class: bestClass}
	}

	// Otherwise it's global (top-level)
	return &ScopeInfo{Type: ScopeGlobal, Function: nil, Class: nil}
}

// Helper span structure used to choose the most specific enclosing node
type span struct {
	start token.Position
	end   token.Position
}

func spanLarge() span {
	// Large sentinel span to be replaced by any real node
	return span{start: token.Position{Line: 0, Column: 0}, end: token.Position{Line: 1<<30 - 1, Column: 1<<30 - 1}}
}

// moreSpecific returns true if span (aStart,aEnd) is strictly more specific (smaller) than (bStart,bEnd)
func moreSpecific(aStart, aEnd, bStart, bEnd token.Position) bool {
	aLen := spanLength(aStart, aEnd)
	bLen := spanLength(bStart, bEnd)
	if aLen < bLen {
		return true
	}
	if aLen > bLen {
		return false
	}
	// If equal length, prefer the one that starts later (deeper)
	if aStart.Line > bStart.Line {
		return true
	}
	if aStart.Line < bStart.Line {
		return false
	}
	return aStart.Column > bStart.Column
}

func spanLength(start, end token.Position) int {
	// Rough span measure: line distance weighted + column distance
	lineDelta := end.Line - start.Line
	colDelta := end.Column - start.Column
	if lineDelta < 0 {
		lineDelta = 0
	}
	if colDelta < 0 {
		colDelta = 0
	}
	return lineDelta*100000 + colDelta
}
