package analysis

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/token"
)

// TestPositionCoordinates verifies that token.Position uses 1-based coordinates
// as documented. This is important for LSP integration since LSP uses 0-based positions.
func TestPositionCoordinates(t *testing.T) {
	// Create a position at the start of file
	pos := token.Position{
		Line:   1, // 1-indexed: first line is 1
		Column: 1, // 1-indexed: first column is 1
		Offset: 0, // 0-indexed: first byte is 0
	}

	if pos.Line != 1 {
		t.Errorf("Expected line to be 1-indexed, got Line=%d", pos.Line)
	}

	if pos.Column != 1 {
		t.Errorf("Expected column to be 1-indexed, got Column=%d", pos.Column)
	}

	if pos.Offset != 0 {
		t.Errorf("Expected offset to be 0-indexed, got Offset=%d", pos.Offset)
	}

	// Test position validity
	if !pos.IsValid() {
		t.Error("Position with Line > 0 should be valid")
	}

	invalidPos := token.Position{Line: 0, Column: 0}
	if invalidPos.IsValid() {
		t.Error("Position with Line = 0 should be invalid")
	}
}

// TestLSPToASTPositionConversion documents the coordinate system conversion
// between LSP (0-based) and AST (1-based) positions.
func TestLSPToASTPositionConversion(t *testing.T) {
	tests := []struct {
		name        string
		lspLine     int // 0-based
		lspChar     int // 0-based
		expectedAST token.Position
	}{
		{
			name:    "Start of file",
			lspLine: 0,
			lspChar: 0,
			expectedAST: token.Position{
				Line:   1, // LSP 0 -> AST 1
				Column: 1, // LSP 0 -> AST 1
			},
		},
		{
			name:    "Second line, first character",
			lspLine: 1,
			lspChar: 0,
			expectedAST: token.Position{
				Line:   2, // LSP 1 -> AST 2
				Column: 1, // LSP 0 -> AST 1
			},
		},
		{
			name:    "Third line, fifth character",
			lspLine: 2,
			lspChar: 4,
			expectedAST: token.Position{
				Line:   3, // LSP 2 -> AST 3
				Column: 5, // LSP 4 -> AST 5
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert LSP (0-based) to AST (1-based)
			astLine := tt.lspLine + 1
			astColumn := tt.lspChar + 1

			if astLine != tt.expectedAST.Line {
				t.Errorf("Expected AST line %d, got %d", tt.expectedAST.Line, astLine)
			}

			if astColumn != tt.expectedAST.Column {
				t.Errorf("Expected AST column %d, got %d", tt.expectedAST.Column, astColumn)
			}
		})
	}
}
