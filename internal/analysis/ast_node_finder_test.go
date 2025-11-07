package analysis

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

func TestPositionInRange(t *testing.T) {
	tests := []struct {
		name     string
		pos      token.Position
		start    token.Position
		end      token.Position
		expected bool
	}{
		{
			name:     "Position at start",
			pos:      token.Position{Line: 1, Column: 1},
			start:    token.Position{Line: 1, Column: 1},
			end:      token.Position{Line: 1, Column: 10},
			expected: true,
		},
		{
			name:     "Position at end",
			pos:      token.Position{Line: 1, Column: 10},
			start:    token.Position{Line: 1, Column: 1},
			end:      token.Position{Line: 1, Column: 10},
			expected: true,
		},
		{
			name:     "Position in middle",
			pos:      token.Position{Line: 1, Column: 5},
			start:    token.Position{Line: 1, Column: 1},
			end:      token.Position{Line: 1, Column: 10},
			expected: true,
		},
		{
			name:     "Position before start",
			pos:      token.Position{Line: 1, Column: 0},
			start:    token.Position{Line: 1, Column: 1},
			end:      token.Position{Line: 1, Column: 10},
			expected: false,
		},
		{
			name:     "Position after end",
			pos:      token.Position{Line: 1, Column: 11},
			start:    token.Position{Line: 1, Column: 1},
			end:      token.Position{Line: 1, Column: 10},
			expected: false,
		},
		{
			name:     "Multi-line range - position on first line",
			pos:      token.Position{Line: 1, Column: 5},
			start:    token.Position{Line: 1, Column: 1},
			end:      token.Position{Line: 3, Column: 10},
			expected: true,
		},
		{
			name:     "Multi-line range - position on middle line",
			pos:      token.Position{Line: 2, Column: 5},
			start:    token.Position{Line: 1, Column: 1},
			end:      token.Position{Line: 3, Column: 10},
			expected: true,
		},
		{
			name:     "Multi-line range - position on last line",
			pos:      token.Position{Line: 3, Column: 5},
			start:    token.Position{Line: 1, Column: 1},
			end:      token.Position{Line: 3, Column: 10},
			expected: true,
		},
		{
			name:     "Position on line before range",
			pos:      token.Position{Line: 0, Column: 5},
			start:    token.Position{Line: 1, Column: 1},
			end:      token.Position{Line: 3, Column: 10},
			expected: false,
		},
		{
			name:     "Position on line after range",
			pos:      token.Position{Line: 4, Column: 5},
			start:    token.Position{Line: 1, Column: 1},
			end:      token.Position{Line: 3, Column: 10},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := positionInRange(tt.pos, tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("positionInRange(%v, %v, %v) = %v, want %v",
					tt.pos, tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

func TestGetSymbolName(t *testing.T) {
	tests := []struct {
		name     string
		node     ast.Node
		expected string
	}{
		{
			name: "Identifier",
			node: &ast.Identifier{
				Value: "myVar",
			},
			expected: "myVar",
		},
		{
			name: "Function declaration",
			node: &ast.FunctionDecl{
				Name: &ast.Identifier{Value: "myFunction"},
			},
			expected: "myFunction",
		},
		{
			name: "Variable declaration",
			node: &ast.VarDeclStatement{
				Names: []*ast.Identifier{
					{Value: "x"},
					{Value: "y"},
				},
			},
			expected: "x", // Returns first name
		},
		{
			name: "Const declaration",
			node: &ast.ConstDecl{
				Name: &ast.Identifier{Value: "MAX_SIZE"},
			},
			expected: "MAX_SIZE",
		},
		{
			name: "Class declaration",
			node: &ast.ClassDecl{
				Name: &ast.Identifier{Value: "MyClass"},
			},
			expected: "MyClass",
		},
		{
			name:     "Nil node",
			node:     nil,
			expected: "",
		},
		{
			name: "Integer literal (not a symbol)",
			node: &ast.IntegerLiteral{
				Value: 42,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSymbolName(tt.node)
			if result != tt.expected {
				t.Errorf("GetSymbolName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFindNodeAtPosition(t *testing.T) {
	// Create a simple program for testing
	// var x: Integer = 42;
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Token: token.Token{
					Pos: token.Position{Line: 1, Column: 1},
				},
				Names: []*ast.Identifier{
					{
						Token: token.Token{
							Pos:     token.Position{Line: 1, Column: 5},
							Literal: "x",
						},
						Value:  "x",
						EndPos: token.Position{Line: 1, Column: 6},
					},
				},
				Type: &ast.TypeAnnotation{
					Token: token.Token{
						Pos:     token.Position{Line: 1, Column: 8},
						Literal: "Integer",
					},
					Name:   "Integer",
					EndPos: token.Position{Line: 1, Column: 15},
				},
				Value: &ast.IntegerLiteral{
					Token: token.Token{
						Type:    token.INT,
						Literal: "42",
						Pos:     token.Position{Line: 1, Column: 18},
					},
					Value:  42,
					EndPos: token.Position{Line: 1, Column: 20},
				},
				EndPos: token.Position{Line: 1, Column: 21},
			},
		},
	}

	tests := []struct {
		name         string
		line         int
		col          int
		expectFound  bool
		expectSymbol string
	}{
		{
			name:         "Position on variable name",
			line:         1,
			col:          5,
			expectFound:  true,
			expectSymbol: "x",
		},
		{
			name:        "Position on integer literal",
			line:        1,
			col:         18,
			expectFound: true,
			// IntegerLiteral doesn't have a symbol name
		},
		{
			name:        "Position outside any node",
			line:        10,
			col:         1,
			expectFound: false,
		},
		{
			name:        "Position before first node",
			line:        1,
			col:         0, // Column 0 is invalid (1-indexed)
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := FindNodeAtPosition(program, tt.line, tt.col)

			if tt.expectFound && node == nil {
				t.Errorf("Expected to find a node at %d:%d, but got nil", tt.line, tt.col)
				return
			}

			if !tt.expectFound && node != nil {
				t.Errorf("Expected no node at %d:%d, but found %T", tt.line, tt.col, node)
				return
			}

			if tt.expectFound && tt.expectSymbol != "" {
				symbolName := GetSymbolName(node)
				if symbolName != tt.expectSymbol {
					t.Errorf("Expected symbol name %q, got %q", tt.expectSymbol, symbolName)
				}
			}
		})
	}
}

func TestFindNodeAtPosition_Nil(t *testing.T) {
	// Test with nil program
	node := FindNodeAtPosition(nil, 1, 1)
	if node != nil {
		t.Errorf("Expected nil for nil program, got %T", node)
	}
}
