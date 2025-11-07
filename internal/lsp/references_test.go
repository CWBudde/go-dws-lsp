package lsp

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/dwscript"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/analysis"
	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// TestSortLocationsByFileAndPosition verifies that locations are sorted
// correctly by file (URI) then by position (line, then character).
// This test validates task 6.10 implementation.
func TestSortLocationsByFileAndPosition(t *testing.T) {
	tests := []struct {
		name     string
		input    []protocol.Location
		expected []protocol.Location
	}{
		{
			name: "sort by line within same file",
			input: []protocol.Location{
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 10, Character: 5}}},
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 3}}},
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 8, Character: 0}}},
			},
			expected: []protocol.Location{
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 3}}},
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 8, Character: 0}}},
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 10, Character: 5}}},
			},
		},
		{
			name: "sort by character within same line",
			input: []protocol.Location{
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 10}}},
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 3}}},
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 7}}},
			},
			expected: []protocol.Location{
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 3}}},
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 7}}},
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 10}}},
			},
		},
		{
			name: "sort by file (URI) first",
			input: []protocol.Location{
				{URI: "file:///c.dws", Range: protocol.Range{Start: protocol.Position{Line: 1, Character: 0}}},
				{URI: "file:///a.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 0}}},
				{URI: "file:///b.dws", Range: protocol.Range{Start: protocol.Position{Line: 3, Character: 0}}},
			},
			expected: []protocol.Location{
				{URI: "file:///a.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 0}}},
				{URI: "file:///b.dws", Range: protocol.Range{Start: protocol.Position{Line: 3, Character: 0}}},
				{URI: "file:///c.dws", Range: protocol.Range{Start: protocol.Position{Line: 1, Character: 0}}},
			},
		},
		{
			name: "sort across multiple files with multiple positions",
			input: []protocol.Location{
				{URI: "file:///b.dws", Range: protocol.Range{Start: protocol.Position{Line: 10, Character: 5}}},
				{URI: "file:///a.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 3}}},
				{URI: "file:///b.dws", Range: protocol.Range{Start: protocol.Position{Line: 3, Character: 0}}},
				{URI: "file:///a.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 1}}},
				{URI: "file:///c.dws", Range: protocol.Range{Start: protocol.Position{Line: 1, Character: 0}}},
			},
			expected: []protocol.Location{
				{URI: "file:///a.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 1}}},
				{URI: "file:///a.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 3}}},
				{URI: "file:///b.dws", Range: protocol.Range{Start: protocol.Position{Line: 3, Character: 0}}},
				{URI: "file:///b.dws", Range: protocol.Range{Start: protocol.Position{Line: 10, Character: 5}}},
				{URI: "file:///c.dws", Range: protocol.Range{Start: protocol.Position{Line: 1, Character: 0}}},
			},
		},
		{
			name:     "empty slice",
			input:    []protocol.Location{},
			expected: []protocol.Location{},
		},
		{
			name: "single location",
			input: []protocol.Location{
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 3}}},
			},
			expected: []protocol.Location{
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 3}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of the input to avoid modifying the test data
			locations := make([]protocol.Location, len(tt.input))
			copy(locations, tt.input)

			// Sort the locations
			sortLocationsByFileAndPosition(locations)

			// Verify the result matches expected
			if len(locations) != len(tt.expected) {
				t.Fatalf("expected %d locations, got %d", len(tt.expected), len(locations))
			}

			for i := range locations {
				if locations[i].URI != tt.expected[i].URI {
					t.Errorf("location[%d]: expected URI %s, got %s",
						i, tt.expected[i].URI, locations[i].URI)
				}
				if locations[i].Range.Start.Line != tt.expected[i].Range.Start.Line {
					t.Errorf("location[%d]: expected line %d, got %d",
						i, tt.expected[i].Range.Start.Line, locations[i].Range.Start.Line)
				}
				if locations[i].Range.Start.Character != tt.expected[i].Range.Start.Character {
					t.Errorf("location[%d]: expected character %d, got %d",
						i, tt.expected[i].Range.Start.Character, locations[i].Range.Start.Character)
				}
			}
		})
	}
}

// TestApplyIncludeDeclaration verifies that the includeDeclaration flag works correctly.
// This test validates task 6.11 implementation.
func TestApplyIncludeDeclaration(t *testing.T) {
	defLocation := &protocol.Location{
		URI: "file:///test.dws",
		Range: protocol.Range{
			Start: protocol.Position{Line: 1, Character: 4},
			End:   protocol.Position{Line: 1, Character: 10},
		},
	}

	tests := []struct {
		name               string
		locations          []protocol.Location
		defLocation        *protocol.Location
		includeDeclaration bool
		expectedCount      int
		expectedFirst      *protocol.Location // If non-nil, check first element
	}{
		{
			name: "include declaration - not in list",
			locations: []protocol.Location{
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 3, Character: 5}}},
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 8}}},
			},
			defLocation:        defLocation,
			includeDeclaration: true,
			expectedCount:      3,
			expectedFirst:      defLocation,
		},
		{
			name: "include declaration - already in list",
			locations: []protocol.Location{
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 1, Character: 4}}},
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 3, Character: 5}}},
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 8}}},
			},
			defLocation:        defLocation,
			includeDeclaration: true,
			expectedCount:      3,
			expectedFirst:      defLocation,
		},
		{
			name: "include declaration - in middle of list",
			locations: []protocol.Location{
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 0, Character: 0}}},
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 1, Character: 4}}},
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 8}}},
			},
			defLocation:        defLocation,
			includeDeclaration: true,
			expectedCount:      3,
			expectedFirst:      defLocation,
		},
		{
			name: "exclude declaration - not in list",
			locations: []protocol.Location{
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 3, Character: 5}}},
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 8}}},
			},
			defLocation:        defLocation,
			includeDeclaration: false,
			expectedCount:      2,
		},
		{
			name: "exclude declaration - in list",
			locations: []protocol.Location{
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 1, Character: 4}}},
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 3, Character: 5}}},
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 8}}},
			},
			defLocation:        defLocation,
			includeDeclaration: false,
			expectedCount:      2,
		},
		{
			name: "exclude declaration - in middle of list",
			locations: []protocol.Location{
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 0, Character: 0}}},
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 1, Character: 4}}},
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 5, Character: 8}}},
			},
			defLocation:        defLocation,
			includeDeclaration: false,
			expectedCount:      2,
		},
		{
			name: "nil definition location - include",
			locations: []protocol.Location{
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 3, Character: 5}}},
			},
			defLocation:        nil,
			includeDeclaration: true,
			expectedCount:      1,
		},
		{
			name: "nil definition location - exclude",
			locations: []protocol.Location{
				{URI: "file:///test.dws", Range: protocol.Range{Start: protocol.Position{Line: 3, Character: 5}}},
			},
			defLocation:        nil,
			includeDeclaration: false,
			expectedCount:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of the input to avoid modifying the test data
			locations := make([]protocol.Location, len(tt.locations))
			copy(locations, tt.locations)

			// Apply includeDeclaration logic
			result := applyIncludeDeclaration(locations, tt.defLocation, tt.includeDeclaration)

			// Verify the result count
			if len(result) != tt.expectedCount {
				t.Errorf("expected %d locations, got %d", tt.expectedCount, len(result))
			}

			// If expectedFirst is specified, verify the first element
			if tt.expectedFirst != nil && len(result) > 0 {
				first := result[0]
				if first.URI != tt.expectedFirst.URI ||
					first.Range.Start.Line != tt.expectedFirst.Range.Start.Line ||
					first.Range.Start.Character != tt.expectedFirst.Range.Start.Character {
					t.Errorf("expected first location to be %+v, got %+v", tt.expectedFirst, first)
				}
			}

			// Verify definition is not in list when includeDeclaration is false
			if !tt.includeDeclaration && tt.defLocation != nil {
				for i, loc := range result {
					if loc.URI == tt.defLocation.URI &&
						loc.Range.Start.Line == tt.defLocation.Range.Start.Line &&
						loc.Range.Start.Character == tt.defLocation.Range.Start.Character {
						t.Errorf("definition should not be in results when includeDeclaration=false, found at index %d", i)
					}
				}
			}
		})
	}
}


// Helper function to parse DWScript code for testing
func parseCodeForReferenceTest(t *testing.T, code string) (*dwscript.Program, error) {
	t.Helper()
	program, _, err := analysis.ParseDocument(code, "test.dws")
	if err != nil {
		return nil, err
	}
	return program, nil
}

// TestLocalReferences_LocalVariable tests finding references for a local variable.
// This validates task 6.12 requirements.
func TestLocalReferences_LocalVariable(t *testing.T) {
	source := `procedure TestProc;
var
  myVar: Integer;
begin
  myVar := 10;
  myVar := myVar + 5;
end;`

	// Set up server and document
	srv := server.New()
	SetServer(srv)

	uri := "file:///test.dws"
	params := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       source,
		},
	}

	// Open the document
	err := DidOpen(nil, params)
	if err != nil {
		t.Fatalf("DidOpen returned error: %v", err)
	}

	// Test finding references for myVar (at line 5, the first usage: "myVar := 10;")
	refParams := &protocol.ReferenceParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 4, Character: 4}, // 0-based: line 5 "  myVar := 10;"
		},
		Context: protocol.ReferenceContext{
			IncludeDeclaration: true,
		},
	}

	locations, err := References(nil, refParams)
	if err != nil {
		t.Fatalf("References returned error: %v", err)
	}

	// Should find multiple references (declaration + usages)
	if len(locations) == 0 {
		t.Error("Expected to find references, got 0")
	}

	t.Logf("Found %d references for myVar", len(locations))
	for i, loc := range locations {
		t.Logf("  [%d] line %d, char %d", i, loc.Range.Start.Line, loc.Range.Start.Character)
	}

	// Verify all locations have correct URI
	for i, loc := range locations {
		if loc.URI != uri {
			t.Errorf("Location[%d] has wrong URI: expected %s, got %s", i, uri, loc.URI)
		}
	}
}

// TestLocalReferences_WithinSameFunction tests that references are limited to the same function.
func TestLocalReferences_WithinSameFunction(t *testing.T) {
	source := `procedure FuncA;
var x: Integer;
begin
  x := 1;
  x := x + 2;
end;

procedure FuncB;
var x: Integer;
begin
  x := 10;
  x := x + 20;
end;`

	// Set up server and document
	srv := server.New()
	SetServer(srv)

	uri := "file:///test.dws"
	params := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       source,
		},
	}

	// Open the document
	err := DidOpen(nil, params)
	if err != nil {
		t.Fatalf("DidOpen returned error: %v", err)
	}

	// Test finding references for x in FuncA (line 4, first usage: "  x := 1;")
	refParams := &protocol.ReferenceParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 3, Character: 4}, // 0-based: line 4, "  x := 1;"
		},
		Context: protocol.ReferenceContext{
			IncludeDeclaration: true,
		},
	}

	locations, err := References(nil, refParams)
	if err != nil {
		t.Fatalf("References returned error: %v", err)
	}

	// Should find references only in FuncA
	if len(locations) == 0 {
		t.Error("Expected to find references in FuncA, got 0")
	}

	t.Logf("Found %d references for x in FuncA", len(locations))

	// Verify no references are from FuncB (lines 8-12, 0-based: 7-11)
	for i, loc := range locations {
		line := loc.Range.Start.Line
		if line >= 7 && line <= 11 {
			t.Errorf("Location[%d] at line %d should not be included (it's in FuncB)", i, line)
		}
	}
}

// TestLocalReferences_IncludeDeclarationFlag tests the includeDeclaration flag.
func TestLocalReferences_IncludeDeclarationFlag(t *testing.T) {
	source := `procedure TestProc;
var myVar: Integer;
begin
  myVar := 10;
  myVar := myVar + 5;
end;`

	// Set up server and document
	srv := server.New()
	SetServer(srv)

	uri := "file:///test.dws"
	params := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       source,
		},
	}

	// Open the document
	err := DidOpen(nil, params)
	if err != nil {
		t.Fatalf("DidOpen returned error: %v", err)
	}

	// Test 1: includeDeclaration = true
	paramsWithDecl := &protocol.ReferenceParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 3, Character: 4}, // 0-based: line 4, "  myVar := 10;"
		},
		Context: protocol.ReferenceContext{
			IncludeDeclaration: true,
		},
	}

	locationsWithDecl, err := References(nil, paramsWithDecl)
	if err != nil {
		t.Fatalf("References (with declaration) returned error: %v", err)
	}

	// Test 2: includeDeclaration = false
	paramsWithoutDecl := &protocol.ReferenceParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 3, Character: 4}, // Same position as above
		},
		Context: protocol.ReferenceContext{
			IncludeDeclaration: false,
		},
	}

	locationsWithoutDecl, err := References(nil, paramsWithoutDecl)
	if err != nil {
		t.Fatalf("References (without declaration) returned error: %v", err)
	}

	t.Logf("With declaration: %d locations", len(locationsWithDecl))
	t.Logf("Without declaration: %d locations", len(locationsWithoutDecl))

	// With declaration should have more or equal results than without
	if len(locationsWithoutDecl) > len(locationsWithDecl) {
		t.Errorf("includeDeclaration=false should not have more results than includeDeclaration=true")
	}
}
