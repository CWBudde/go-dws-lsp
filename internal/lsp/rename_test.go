package lsp

import (
	"testing"

	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// TestCanRenameSymbol_Keywords tests that DWScript keywords cannot be renamed.
// This test validates task 11.4 implementation.
func TestCanRenameSymbol_Keywords(t *testing.T) {
	tests := []struct {
		name       string
		symbolName string
		canRename  bool
		reason     string
	}{
		{
			name:       "reject 'begin' keyword",
			symbolName: "begin",
			canRename:  false,
			reason:     "cannot rename DWScript keyword",
		},
		{
			name:       "reject 'end' keyword",
			symbolName: "end",
			canRename:  false,
			reason:     "cannot rename DWScript keyword",
		},
		{
			name:       "reject 'if' keyword",
			symbolName: "if",
			canRename:  false,
			reason:     "cannot rename DWScript keyword",
		},
		{
			name:       "reject 'then' keyword",
			symbolName: "then",
			canRename:  false,
			reason:     "cannot rename DWScript keyword",
		},
		{
			name:       "reject 'else' keyword",
			symbolName: "else",
			canRename:  false,
			reason:     "cannot rename DWScript keyword",
		},
		{
			name:       "reject 'while' keyword",
			symbolName: "while",
			canRename:  false,
			reason:     "cannot rename DWScript keyword",
		},
		{
			name:       "reject 'for' keyword",
			symbolName: "for",
			canRename:  false,
			reason:     "cannot rename DWScript keyword",
		},
		{
			name:       "reject 'do' keyword",
			symbolName: "do",
			canRename:  false,
			reason:     "cannot rename DWScript keyword",
		},
		{
			name:       "reject 'var' keyword",
			symbolName: "var",
			canRename:  false,
			reason:     "cannot rename DWScript keyword",
		},
		{
			name:       "reject 'function' keyword",
			symbolName: "function",
			canRename:  false,
			reason:     "cannot rename DWScript keyword",
		},
		{
			name:       "reject 'procedure' keyword",
			symbolName: "procedure",
			canRename:  false,
			reason:     "cannot rename DWScript keyword",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canRename, reason := canRenameSymbol(tt.symbolName)

			if canRename != tt.canRename {
				t.Errorf("expected canRename=%v, got %v", tt.canRename, canRename)
			}

			if !canRename && reason != tt.reason {
				t.Errorf("expected reason '%s', got '%s'", tt.reason, reason)
			}
		})
	}
}

// TestCanRenameSymbol_BuiltInTypes tests that built-in types cannot be renamed.
// This test validates task 11.4 implementation.
func TestCanRenameSymbol_BuiltInTypes(t *testing.T) {
	tests := []struct {
		name       string
		symbolName string
		canRename  bool
		reason     string
	}{
		{
			name:       "reject 'Integer' type",
			symbolName: "Integer",
			canRename:  false,
			reason:     "cannot rename built-in type",
		},
		{
			name:       "reject 'String' type",
			symbolName: "String",
			canRename:  false,
			reason:     "cannot rename built-in type",
		},
		{
			name:       "reject 'Float' type",
			symbolName: "Float",
			canRename:  false,
			reason:     "cannot rename built-in type",
		},
		{
			name:       "reject 'Boolean' type",
			symbolName: "Boolean",
			canRename:  false,
			reason:     "cannot rename built-in type",
		},
		{
			name:       "reject 'Variant' type",
			symbolName: "Variant",
			canRename:  false,
			reason:     "cannot rename built-in type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canRename, reason := canRenameSymbol(tt.symbolName)

			if canRename != tt.canRename {
				t.Errorf("expected canRename=%v, got %v", tt.canRename, canRename)
			}

			if !canRename && reason != tt.reason {
				t.Errorf("expected reason '%s', got '%s'", tt.reason, reason)
			}
		})
	}
}

// TestCanRenameSymbol_BuiltInFunctions tests that built-in functions cannot be renamed.
// This test validates task 11.4 implementation.
func TestCanRenameSymbol_BuiltInFunctions(t *testing.T) {
	tests := []struct {
		name       string
		symbolName string
		canRename  bool
		reason     string
	}{
		{
			name:       "reject 'PrintLn' function",
			symbolName: "PrintLn",
			canRename:  false,
			reason:     "cannot rename built-in function",
		},
		{
			name:       "reject 'Length' function",
			symbolName: "Length",
			canRename:  false,
			reason:     "cannot rename built-in function",
		},
		{
			name:       "reject 'Copy' function",
			symbolName: "Copy",
			canRename:  false,
			reason:     "cannot rename built-in function",
		},
		{
			name:       "reject 'IntToStr' function",
			symbolName: "IntToStr",
			canRename:  false,
			reason:     "cannot rename built-in function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canRename, reason := canRenameSymbol(tt.symbolName)

			if canRename != tt.canRename {
				t.Errorf("expected canRename=%v, got %v", tt.canRename, canRename)
			}

			if !canRename && reason != tt.reason {
				t.Errorf("expected reason '%s', got '%s'", tt.reason, reason)
			}
		})
	}
}

// TestCanRenameSymbol_UserDefined tests that user-defined symbols can be renamed.
// This test validates task 11.4 implementation.
func TestCanRenameSymbol_UserDefined(t *testing.T) {
	tests := []struct {
		name       string
		symbolName string
		canRename  bool
	}{
		{
			name:       "allow renaming 'myVariable'",
			symbolName: "myVariable",
			canRename:  true,
		},
		{
			name:       "allow renaming 'MyFunction'",
			symbolName: "MyFunction",
			canRename:  true,
		},
		{
			name:       "allow renaming 'TMyClass'",
			symbolName: "TMyClass",
			canRename:  true,
		},
		{
			name:       "allow renaming 'x'",
			symbolName: "x",
			canRename:  true,
		},
		{
			name:       "allow renaming 'counter'",
			symbolName: "counter",
			canRename:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canRename, reason := canRenameSymbol(tt.symbolName)

			if canRename != tt.canRename {
				t.Errorf("expected canRename=%v, got %v", tt.canRename, canRename)
			}

			if canRename && reason != "" {
				t.Errorf("expected empty reason for valid symbol, got '%s'", reason)
			}
		})
	}
}

// TestBuildWorkspaceEdit tests that WorkspaceEdit is built correctly from locations.
// This test validates tasks 11.6, 11.7, and 11.8 implementation.
func TestBuildWorkspaceEdit(t *testing.T) {
	// Create a mock document store
	docStore := server.NewDocumentStore()

	tests := []struct {
		name                 string
		locations            []protocol.Location
		newName              string
		expectedDocCount     int
		expectedTotalEdits   int
		expectedFirstURI     protocol.DocumentUri
		expectedFirstNewText string
	}{
		{
			name: "single location rename",
			locations: []protocol.Location{
				{
					URI: "file:///test.dws",
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 4},
						End:   protocol.Position{Line: 1, Character: 5},
					},
				},
			},
			newName:              "y",
			expectedDocCount:     1,
			expectedTotalEdits:   1,
			expectedFirstURI:     "file:///test.dws",
			expectedFirstNewText: "y",
		},
		{
			name: "multiple locations same file",
			locations: []protocol.Location{
				{
					URI: "file:///test.dws",
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 4},
						End:   protocol.Position{Line: 1, Character: 5},
					},
				},
				{
					URI: "file:///test.dws",
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 10},
						End:   protocol.Position{Line: 3, Character: 11},
					},
				},
				{
					URI: "file:///test.dws",
					Range: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 15},
						End:   protocol.Position{Line: 5, Character: 16},
					},
				},
			},
			newName:              "newName",
			expectedDocCount:     1,
			expectedTotalEdits:   3,
			expectedFirstURI:     "file:///test.dws",
			expectedFirstNewText: "newName",
		},
		{
			name: "multiple locations across files",
			locations: []protocol.Location{
				{
					URI: "file:///a.dws",
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 4},
						End:   protocol.Position{Line: 1, Character: 10},
					},
				},
				{
					URI: "file:///b.dws",
					Range: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 8},
						End:   protocol.Position{Line: 5, Character: 14},
					},
				},
				{
					URI: "file:///c.dws",
					Range: protocol.Range{
						Start: protocol.Position{Line: 10, Character: 12},
						End:   protocol.Position{Line: 10, Character: 18},
					},
				},
			},
			newName:              "RenamedFunc",
			expectedDocCount:     3,
			expectedTotalEdits:   3,
			expectedFirstNewText: "RenamedFunc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edit := buildWorkspaceEdit(tt.locations, tt.newName, docStore)

			if edit == nil {
				t.Fatal("expected WorkspaceEdit, got nil")
			}

			// Verify DocumentChanges exists
			if edit.DocumentChanges == nil {
				t.Fatal("expected DocumentChanges to be set")
			}

			// Verify document count
			if len(edit.DocumentChanges) != tt.expectedDocCount {
				t.Errorf("expected %d documents, got %d", tt.expectedDocCount, len(edit.DocumentChanges))
			}

			// Count total edits across all documents
			totalEdits := 0
			for _, change := range edit.DocumentChanges {
				if textDocEdit, ok := change.(protocol.TextDocumentEdit); ok {
					totalEdits += len(textDocEdit.Edits)

					// Verify each edit has the correct NewText
					for _, editInterface := range textDocEdit.Edits {
						if textEdit, ok := editInterface.(protocol.TextEdit); ok {
							if textEdit.NewText != tt.expectedFirstNewText {
								t.Errorf("expected NewText '%s', got '%s'", tt.expectedFirstNewText, textEdit.NewText)
							}
						}
					}
				}
			}

			if totalEdits != tt.expectedTotalEdits {
				t.Errorf("expected %d total edits, got %d", tt.expectedTotalEdits, totalEdits)
			}

			// Verify first document URI if specified
			if tt.expectedFirstURI != "" && len(edit.DocumentChanges) > 0 {
				if textDocEdit, ok := edit.DocumentChanges[0].(protocol.TextDocumentEdit); ok {
					if textDocEdit.TextDocument.URI != tt.expectedFirstURI {
						t.Errorf("expected first URI '%s', got '%s'", tt.expectedFirstURI, textDocEdit.TextDocument.URI)
					}
				}
			}
		})
	}
}

// TestBuildWorkspaceEdit_EmptyLocations tests that empty locations produce an empty edit.
func TestBuildWorkspaceEdit_EmptyLocations(t *testing.T) {
	docStore := server.NewDocumentStore()
	locations := []protocol.Location{}
	newName := "newName"

	edit := buildWorkspaceEdit(locations, newName, docStore)

	if edit == nil {
		t.Fatal("expected WorkspaceEdit, got nil")
	}

	// DocumentChanges should be an empty slice when there are no locations
	if len(edit.DocumentChanges) != 0 {
		t.Errorf("expected 0 documents, got %d", len(edit.DocumentChanges))
	}
}

// TestConvertToEdits tests that TextEdits are converted to interface{} correctly.
func TestConvertToEdits(t *testing.T) {
	textEdits := []protocol.TextEdit{
		{
			Range: protocol.Range{
				Start: protocol.Position{Line: 1, Character: 4},
				End:   protocol.Position{Line: 1, Character: 5},
			},
			NewText: "y",
		},
		{
			Range: protocol.Range{
				Start: protocol.Position{Line: 3, Character: 10},
				End:   protocol.Position{Line: 3, Character: 11},
			},
			NewText: "y",
		},
	}

	result := convertToEdits(textEdits)

	if len(result) != len(textEdits) {
		t.Errorf("expected %d edits, got %d", len(textEdits), len(result))
	}

	for i, edit := range result {
		if textEdit, ok := edit.(protocol.TextEdit); ok {
			if textEdit.NewText != textEdits[i].NewText {
				t.Errorf("edit[%d]: expected NewText '%s', got '%s'", i, textEdits[i].NewText, textEdit.NewText)
			}
			if textEdit.Range.Start.Line != textEdits[i].Range.Start.Line {
				t.Errorf("edit[%d]: expected Line %d, got %d", i, textEdits[i].Range.Start.Line, textEdit.Range.Start.Line)
			}
		} else {
			t.Errorf("edit[%d]: expected protocol.TextEdit type", i)
		}
	}
}
