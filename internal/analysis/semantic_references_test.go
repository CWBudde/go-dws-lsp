package analysis

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/dwscript"
	"github.com/cwbudde/go-dws/pkg/token"
)

// Helper function to compile DWScript code with semantic info for testing.
func compileCode(t *testing.T, code string) *dwscript.Program {
	t.Helper()

	program, compileMsgs, err := ParseDocument(code, "test.dws")
	if err != nil {
		t.Fatalf("Failed to compile test code: %v", err)
	}

	if program == nil {
		if compileMsgs != nil && len(compileMsgs) > 0 {
			t.Logf("Compilation errors:")

			for _, msg := range compileMsgs {
				t.Logf("  - %s", msg.Message)
			}
		}

		t.Fatal("ParseDocument returned nil program")
	}

	return program
}

func TestFindSymbolDefinitionAtPosition_Function(t *testing.T) {
	code := `function MyFunc(): Integer;
begin
  Result := 42;
end;`
	program := compileCode(t, code)

	// Position at the function name definition (line 1, "MyFunc")
	pos := token.Position{Line: 1, Column: 10}
	symbol := FindSymbolDefinitionAtPosition(program, pos)

	if symbol == nil {
		t.Fatal("Expected to find function symbol, got nil")
	}

	if symbol.Name != "MyFunc" {
		t.Errorf("Expected symbol name 'MyFunc', got '%s'", symbol.Name)
	}

	if symbol.Kind != "function" {
		t.Errorf("Expected symbol kind 'function', got '%s'", symbol.Kind)
	}
}

func TestFindSymbolDefinitionAtPosition_Variable(t *testing.T) {
	code := `var globalVar: Integer;
begin
  globalVar := 42;
end.`
	program := compileCode(t, code)

	// Position at the variable name definition (line 1, "globalVar")
	pos := token.Position{Line: 1, Column: 5}
	symbol := FindSymbolDefinitionAtPosition(program, pos)

	if symbol == nil {
		t.Fatal("Expected to find variable symbol, got nil")
	}

	if symbol.Name != "globalVar" {
		t.Errorf("Expected symbol name 'globalVar', got '%s'", symbol.Name)
	}

	if symbol.Kind != "variable" {
		t.Errorf("Expected symbol kind 'variable', got '%s'", symbol.Kind)
	}
}

func TestFindSymbolDefinitionAtPosition_NoSymbol(t *testing.T) {
	code := `
var x: Integer;`
	program := compileCode(t, code)

	// Position at a location with no symbol (line 2, beyond the code)
	pos := token.Position{Line: 10, Column: 1}
	symbol := FindSymbolDefinitionAtPosition(program, pos)

	if symbol != nil {
		t.Errorf("Expected no symbol at empty position, got '%s'", symbol.Name)
	}
}

func TestFindSemanticReferences_LocalVariable(t *testing.T) {
	code := `function TestFunc(): Integer;
var x: Integer;
begin
  x := 10;
  x := x + 1;
  Result := x;
end;`
	program := compileCode(t, code)

	// Position at the variable definition 'x' (line 2, "x")
	symbolPos := token.Position{Line: 2, Column: 5}

	// Find all references
	ranges := FindSemanticReferences(program, "x", symbolPos, "file:///test.dws")

	// We expect to find 'x' at:
	// Line 3: var x: Integer;  (definition)
	// Line 5: x := 10;  (reference)
	// Line 6: x := x + 1;  (2 references)
	// Line 7: Result := x;  (reference)
	// Total: 5 occurrences
	expectedMin := 4 // At least the uses (not counting definition if not included)

	if len(ranges) < expectedMin {
		t.Errorf("Expected at least %d references, got %d", expectedMin, len(ranges))

		for i, r := range ranges {
			t.Logf("  Reference %d: line %d, char %d", i+1, r.Start.Line, r.Start.Character)
		}
	}
}

func TestFindSemanticReferences_GlobalFunction(t *testing.T) {
	code := "function MyFunction(): Integer;\nbegin\n  Result := 42;\nend;\n\nbegin\n  var result := MyFunction();\n  end."
	program := compileCode(t, code)

	// Position at the function definition 'MyFunction' (line 1)
	symbolPos := token.Position{Line: 1, Column: 10}

	// Find all references
	ranges := FindSemanticReferences(program, "MyFunction", symbolPos, "file:///test.dws")

	// We expect to find 'MyFunction' at:
	// Line 2: function MyFunction()  (definition)
	// Line 8: MyFunction()  (call)
	// Line 9: MyFunction()  (call)
	// Total: 3 occurrences
	expectedMin := 2 // At least the two calls

	if len(ranges) < expectedMin {
		t.Errorf("Expected at least %d references, got %d", expectedMin, len(ranges))

		for i, r := range ranges {
			t.Logf("  Reference %d: line %d, char %d", i+1, r.Start.Line, r.Start.Character)
		}
	}
}

func TestFindSemanticReferences_NoFalsePositives(t *testing.T) {
	code := `function FuncA();
var x: Integer;
begin
  x := 10;
end;

function FuncB();
var x: Integer;
begin
  x := 20;
end;`
	program := compileCode(t, code)

	// Position at 'x' in FuncA (line 2)
	symbolPos := token.Position{Line: 2, Column: 5}

	// Find all references
	ranges := FindSemanticReferences(program, "x", symbolPos, "file:///test.dws")

	// We expect to find only the 'x' from FuncA:
	// Line 3: var x: Integer;  (definition)
	// Line 5: x := 10;  (reference)
	// Total: 2 occurrences
	// The 'x' in FuncB (lines 9 and 11) should NOT be included

	if len(ranges) > 2 {
		t.Errorf("Expected at most 2 references (no false positives from FuncB), got %d", len(ranges))

		for i, r := range ranges {
			t.Logf("  Reference %d: line %d, char %d", i+1, r.Start.Line, r.Start.Character)
		}
	}

	// Verify that all found references are in the correct line range (3-5, not 9-11)
	for _, r := range ranges {
		line := int(r.Start.Line) + 1 // Convert from 0-based to 1-based
		if line > 6 {
			t.Errorf("Found reference at line %d, which is in FuncB (should only be in FuncA)", line)
		}
	}
}

func TestResolveIdentifierToDefinition(t *testing.T) {
	code := `
function TestFunc(): Integer;
var localVar: Integer;
begin
  localVar := 42;
  Result := localVar;
end;
`
	program := compileCode(t, code)

	// Position at a reference to 'localVar' (line 6, "localVar" in Result := localVar)
	identPos := token.Position{Line: 6, Column: 13}

	defPos := ResolveIdentifierToDefinition(program, "file:///test.dws", "localVar", identPos)

	if defPos == nil {
		t.Fatal("Expected to resolve identifier to definition, got nil")
	}

	// The definition should be on line 3 (var localVar: Integer)
	expectedLine := 3
	if defPos.Line != expectedLine {
		t.Errorf("Expected definition at line %d, got line %d", expectedLine, defPos.Line)
	}
}

func TestResolveIdentifierToDefinition_NoResolution(t *testing.T) {
	code := `
var x: Integer;
begin
  x := 42;
end.
`
	program := compileCode(t, code)

	// Try to resolve a non-existent identifier
	identPos := token.Position{Line: 4, Column: 3}

	defPos := ResolveIdentifierToDefinition(program, "file:///test.dws", "nonExistent", identPos)

	if defPos != nil {
		t.Errorf("Expected no resolution for non-existent identifier, got position %d:%d", defPos.Line, defPos.Column)
	}
}
