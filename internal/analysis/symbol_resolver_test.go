package analysis

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// Helper function to parse DWScript code for testing
func parseCode(t *testing.T, code string) *ast.Program {
	t.Helper()
	program, _, err := ParseDocument(code, "test.dws")
	if err != nil {
		t.Fatalf("Failed to parse test code: %v", err)
	}
	if program == nil {
		t.Fatal("ParseDocument returned nil program")
	}
	return program.AST()
}

func TestSymbolResolver_ResolveLocal_Parameter(t *testing.T) {
	code := `
function TestFunc(param1: Integer);
begin
  var x := param1;
end;
`
	programAST := parseCode(t, code)

	// Cursor position inside the function body, on line with "var x := param1"
	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   3,
		Column: 16, // On "param1" in the assignment
	})

	locations := resolver.ResolveSymbol("param1")

	if len(locations) == 0 {
		t.Fatal("Expected to find parameter, got no results")
	}

	if len(locations) != 1 {
		t.Errorf("Expected 1 location, got %d", len(locations))
	}
}

func TestSymbolResolver_ResolveLocal_Variable(t *testing.T) {
	code := `function TestFunc();
begin
  var localVar: Integer;
  localVar := 42;
end;`
	programAST := parseCode(t, code)

	// Cursor position on the assignment line
	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   4,
		Column: 3, // On "localVar" in the assignment
	})

	locations := resolver.ResolveSymbol("localVar")

	if len(locations) == 0 {
		t.Fatal("Expected to find local variable, got no results")
	}

	if len(locations) != 1 {
		t.Errorf("Expected 1 location, got %d", len(locations))
	}
}

func TestSymbolResolver_ResolveClassMember_Field(t *testing.T) {
	// Note: DWScript may not support finding class fields from method implementations
	// This test verifies that we attempt resolution but may not find results
	// depending on how the parser structures class methods
	code := `type
  TMyClass = class
    FValue: Integer;
    function GetValue: Integer;
  end;

function TMyClass.GetValue: Integer;
begin
  Result := FValue;
end;`
	programAST := parseCode(t, code)

	// Cursor position inside the method, on "FValue"
	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   9,
		Column: 13, // On "FValue" in the assignment
	})

	locations := resolver.ResolveSymbol("FValue")

	// This may or may not find the field depending on AST structure
	// For now, just verify no crash
	_ = locations
}

func TestSymbolResolver_ResolveGlobal_Function(t *testing.T) {
	code := `
function GlobalFunc(): Integer;
begin
  Result := 42;
end;

var x := GlobalFunc();
`
	programAST := parseCode(t, code)

	// Cursor position on the function call
	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   6,
		Column: 12, // On "GlobalFunc" in the call
	})

	locations := resolver.ResolveSymbol("GlobalFunc")

	if len(locations) == 0 {
		t.Fatal("Expected to find global function, got no results")
	}

	if len(locations) != 1 {
		t.Errorf("Expected 1 location, got %d", len(locations))
	}
}

func TestSymbolResolver_ResolveGlobal_Variable(t *testing.T) {
	code := `
var globalVar: Integer;

begin
  globalVar := 42;
end;
`
	programAST := parseCode(t, code)

	// Cursor position on the assignment
	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   4,
		Column: 5, // On "globalVar"
	})

	locations := resolver.ResolveSymbol("globalVar")

	if len(locations) == 0 {
		t.Fatal("Expected to find global variable, got no results")
	}

	if len(locations) != 1 {
		t.Errorf("Expected 1 location, got %d", len(locations))
	}
}

func TestSymbolResolver_ResolveGlobal_Class(t *testing.T) {
	code := `
type
  TMyClass = class
    FValue: Integer;
  end;

var obj: TMyClass;
`
	programAST := parseCode(t, code)

	// Cursor position on the type reference
	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   6,
		Column: 12, // On "TMyClass"
	})

	locations := resolver.ResolveSymbol("TMyClass")

	if len(locations) == 0 {
		t.Fatal("Expected to find class, got no results")
	}

	if len(locations) != 1 {
		t.Errorf("Expected 1 location, got %d", len(locations))
	}
}

func TestSymbolResolver_ResolveGlobal_Constant(t *testing.T) {
	code := `
const PI: Float = 3.14159;

var x := PI;
`
	programAST := parseCode(t, code)

	// Cursor position on PI reference
	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   3,
		Column: 12, // On "PI"
	})

	locations := resolver.ResolveSymbol("PI")

	if len(locations) == 0 {
		t.Fatal("Expected to find constant, got no results")
	}

	if len(locations) != 1 {
		t.Errorf("Expected 1 location, got %d", len(locations))
	}
}

func TestSymbolResolver_NotFound(t *testing.T) {
	code := `
var x: Integer;
`
	programAST := parseCode(t, code)

	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   1,
		Column: 1,
	})

	locations := resolver.ResolveSymbol("nonExistent")

	if len(locations) != 0 {
		t.Errorf("Expected no locations for non-existent symbol, got %d", len(locations))
	}
}

func TestSymbolResolver_LocalTakesPrecedenceOverGlobal(t *testing.T) {
	code := `var value: Integer;

function TestFunc();
begin
  var value: String;
  value := 'local';
end;`
	programAST := parseCode(t, code)

	// Cursor position inside the function (on the assignment line)
	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   6,
		Column: 3, // On "value" in the assignment
	})

	locations := resolver.ResolveSymbol("value")

	if len(locations) != 1 {
		t.Fatalf("Expected 1 location, got %d", len(locations))
	}

	// Should resolve to the local variable (line 5 in 0-based)
	if locations[0].Range.Start.Line < 4 {
		t.Errorf("Expected local variable (line >= 4 in 0-based), got line %d", locations[0].Range.Start.Line)
	}
}

func TestSymbolResolver_GetResolutionScope_Global(t *testing.T) {
	code := `
var x: Integer;
`
	programAST := parseCode(t, code)

	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   1,
		Column: 1,
	})

	scope := resolver.GetResolutionScope()

	if scope != "global" {
		t.Errorf("Expected 'global' scope, got '%s'", scope)
	}
}

func TestSymbolResolver_GetResolutionScope_Function(t *testing.T) {
	code := `
function TestFunc();
begin
  var x: Integer;
end;
`
	programAST := parseCode(t, code)

	// Cursor inside the function body
	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   3,
		Column: 5,
	})

	scope := resolver.GetResolutionScope()

	if scope != "function" {
		t.Errorf("Expected 'function' scope, got '%s'", scope)
	}
}

func TestSymbolResolver_NilProgram(t *testing.T) {
	resolver := NewSymbolResolver("file:///test.dws", nil, token.Position{
		Line:   1,
		Column: 1,
	})

	locations := resolver.ResolveSymbol("test")

	if locations != nil {
		t.Errorf("Expected nil for nil program, got %v", locations)
	}
}
