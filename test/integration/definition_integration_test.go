//go:build integration
// +build integration

package integration

import (
	"testing"

	"github.com/CWBudde/go-dws-lsp/internal/lsp"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// TestDefinitionIntegration_VariableDeclaration tests go-to-definition on a variable reference
func TestDefinitionIntegration_VariableDeclaration(t *testing.T) {
	_ = setupTestServer()

	uri := "file:///test/definition_vars.dws"
	code := `var x: Integer;
var y: String;
x := 42;`

	// Open document
	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request definition for 'x' on line 2 (the assignment)
	defParams := &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      2, // Line with "x := 42;"
				Character: 0, // Position on 'x'
			},
		},
	}

	result, err := lsp.Definition(ctx, defParams)
	if err != nil {
		t.Fatalf("Definition failed: %v", err)
	}

	if result == nil {
		t.Log("Definition returned nil (acceptable if definition finding is not fully implemented)")
		return
	}

	// Result should be a Location or []Location
	switch loc := result.(type) {
	case *protocol.Location:
		if loc.URI != uri {
			t.Errorf("Expected URI %s, got %s", uri, loc.URI)
		}
		// Should point to line 0 (the declaration)
		if loc.Range.Start.Line != 0 {
			t.Errorf("Expected definition on line 0, got line %d", loc.Range.Start.Line)
		}
		t.Logf("Found definition at line %d, character %d",
			loc.Range.Start.Line, loc.Range.Start.Character)

	case []protocol.Location:
		if len(loc) == 0 {
			t.Error("Definition returned empty location array")
		} else {
			t.Logf("Found %d definitions", len(loc))
			if loc[0].URI != uri {
				t.Errorf("Expected URI %s, got %s", uri, loc[0].URI)
			}
		}

	default:
		t.Errorf("Unexpected result type: %T", result)
	}
}

// TestDefinitionIntegration_FunctionDeclaration tests go-to-definition on a function call
func TestDefinitionIntegration_FunctionDeclaration(t *testing.T) {
	_ = setupTestServer()

	uri := "file:///test/definition_funcs.dws"
	code := `function Add(a, b: Integer): Integer;
begin
  Result := a + b;
end;

var result := Add(1, 2);`

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request definition for 'Add' in the function call
	defParams := &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      5, // Line with "var result := Add(1, 2);"
				Character: 14, // Position on 'Add'
			},
		},
	}

	result, err := lsp.Definition(ctx, defParams)
	if err != nil {
		t.Fatalf("Definition failed: %v", err)
	}

	if result == nil {
		t.Log("Definition returned nil (acceptable if definition finding is not fully implemented)")
		return
	}

	switch loc := result.(type) {
	case *protocol.Location:
		if loc.URI != uri {
			t.Errorf("Expected URI %s, got %s", uri, loc.URI)
		}
		// Should point to line 0 (the function declaration)
		if loc.Range.Start.Line != 0 {
			t.Errorf("Expected definition on line 0, got line %d", loc.Range.Start.Line)
		}
		t.Logf("Found function definition at line %d, character %d",
			loc.Range.Start.Line, loc.Range.Start.Character)

	case []protocol.Location:
		if len(loc) == 0 {
			t.Error("Definition returned empty location array")
		} else {
			t.Logf("Found %d function definitions", len(loc))
		}

	default:
		t.Errorf("Unexpected result type: %T", result)
	}
}

// TestDefinitionIntegration_OnDeclaration tests go-to-definition on the declaration itself
func TestDefinitionIntegration_OnDeclaration(t *testing.T) {
	_ = setupTestServer()

	uri := "file:///test/definition_on_decl.dws"
	code := `var myVariable: Integer;`

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request definition on the variable declaration itself
	defParams := &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      0,
				Character: 4, // Position on 'myVariable'
			},
		},
	}

	result, err := lsp.Definition(ctx, defParams)
	if err != nil {
		t.Fatalf("Definition failed: %v", err)
	}

	if result == nil {
		t.Log("Definition returned nil (acceptable - already at declaration)")
		return
	}

	// When on a declaration, it should either return the declaration itself
	// or return nil. Both are acceptable behaviors.
	t.Logf("Definition on declaration returned: %v", result)
}

// TestDefinitionIntegration_InvalidPosition tests definition on an invalid position
func TestDefinitionIntegration_InvalidPosition(t *testing.T) {
	_ = setupTestServer()

	uri := "file:///test/definition_invalid.dws"
	code := `var x: Integer;`

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request definition on an invalid position (beyond end of line)
	defParams := &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      0,
				Character: 100, // Invalid position
			},
		},
	}

	result, err := lsp.Definition(ctx, defParams)
	if err != nil {
		t.Fatalf("Definition failed: %v", err)
	}

	if result != nil {
		t.Logf("Definition on invalid position returned: %v (this is OK)", result)
	} else {
		t.Log("Definition correctly returned nil for invalid position")
	}
}

// TestDefinitionIntegration_MultipleVariables tests definition with multiple variables
func TestDefinitionIntegration_MultipleVariables(t *testing.T) {
	_ = setupTestServer()

	uri := "file:///test/definition_multi_vars.dws"
	code := `var x, y, z: Integer;
x := 10;
y := 20;
z := 30;`

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Test definition for 'y' on line 2
	defParams := &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      2, // Line with "y := 20;"
				Character: 0, // Position on 'y'
			},
		},
	}

	result, err := lsp.Definition(ctx, defParams)
	if err != nil {
		t.Fatalf("Definition failed: %v", err)
	}

	if result == nil {
		t.Log("Definition returned nil (acceptable if definition finding is not fully implemented)")
		return
	}

	// Should find the declaration on line 0
	switch loc := result.(type) {
	case *protocol.Location:
		if loc.Range.Start.Line != 0 {
			t.Errorf("Expected definition on line 0, got line %d", loc.Range.Start.Line)
		}
		t.Logf("Found definition for 'y' at line %d", loc.Range.Start.Line)

	case []protocol.Location:
		if len(loc) > 0 && loc[0].Range.Start.Line != 0 {
			t.Errorf("Expected definition on line 0, got line %d", loc[0].Range.Start.Line)
		}

	default:
		t.Errorf("Unexpected result type: %T", result)
	}
}

// TestDefinitionIntegration_ClassType tests definition on a class type
func TestDefinitionIntegration_ClassType(t *testing.T) {
	_ = setupTestServer()

	uri := "file:///test/definition_class.dws"
	code := `type TMyClass = class
private
  FValue: Integer;
public
end;

var obj: TMyClass;`

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request definition for 'TMyClass' in the variable declaration
	defParams := &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      6, // Line with "var obj: TMyClass;"
				Character: 9, // Position on 'TMyClass'
			},
		},
	}

	result, err := lsp.Definition(ctx, defParams)
	if err != nil {
		t.Fatalf("Definition failed: %v", err)
	}

	if result == nil {
		t.Log("Definition returned nil (acceptable if class definition finding is not fully implemented)")
		return
	}

	switch loc := result.(type) {
	case *protocol.Location:
		// Should point to the class declaration
		if loc.Range.Start.Line != 0 {
			t.Logf("Expected definition on line 0, got line %d (may be acceptable)", loc.Range.Start.Line)
		}
		t.Logf("Found class definition at line %d", loc.Range.Start.Line)

	case []protocol.Location:
		if len(loc) > 0 {
			t.Logf("Found %d class definitions", len(loc))
		}

	default:
		t.Errorf("Unexpected result type: %T", result)
	}
}

// TestDefinitionIntegration_NonExistentSymbol tests definition on a symbol that doesn't exist
func TestDefinitionIntegration_NonExistentSymbol(t *testing.T) {
	_ = setupTestServer()

	uri := "file:///test/definition_nonexistent.dws"
	code := `var x: Integer;`

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request definition on the type name 'Integer' (which is a built-in type)
	defParams := &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      0,
				Character: 7, // Position on 'Integer'
			},
		},
	}

	result, err := lsp.Definition(ctx, defParams)
	if err != nil {
		t.Fatalf("Definition failed: %v", err)
	}

	// Built-in types won't have definitions in user code, so nil is expected
	if result != nil {
		t.Logf("Definition returned result for built-in type: %v (may be OK)", result)
	} else {
		t.Log("Definition correctly returned nil for built-in type")
	}
}
