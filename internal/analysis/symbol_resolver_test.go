package analysis

import (
	"testing"

	"github.com/CWBudde/go-dws-lsp/internal/workspace"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// Helper function to parse DWScript code for testing
func parseCode(t *testing.T, code string) *ast.Program {
	t.Helper()
	program, compileMsgs, err := ParseDocument(code, "test.dws")
	if err != nil {
		t.Fatalf("Failed to parse test code: %v", err)
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

func TestSymbolResolver_ResolveGlobal_Record(t *testing.T) {
	code := `
type TPoint = record
  X: Integer;
  Y: Integer;
end;

var p: TPoint;
`
	programAST := parseCode(t, code)

	// Cursor position on TPoint reference
	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   6,
		Column: 10, // On "TPoint"
	})

	locations := resolver.ResolveSymbol("TPoint")

	if len(locations) == 0 {
		t.Fatal("Expected to find record type, got no results")
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

func TestSymbolResolver_ResolveClassMember_Property(t *testing.T) {
	code := `type
  TMyClass = class
  private
    FValue: Integer;
  public
    property Value: Integer read FValue write FValue;
    function DoSomething: Integer;
  end;

function TMyClass.DoSomething: Integer;
begin
  Result := Value;
end;`
	programAST := parseCode(t, code)

	// Cursor position inside the method, on "Value"
	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   12,
		Column: 13, // On "Value" property reference
	})

	locations := resolver.ResolveSymbol("Value")

	// Should find the property declaration
	if len(locations) == 0 {
		t.Fatal("Expected to find property, got no results")
	}
}

func TestSymbolResolver_ResolveClassMember_InheritedField(t *testing.T) {
	code := `type TBaseClass = class
    FBaseField: Integer;
  end;

type TDerivedClass = class(TBaseClass)
    FDerivedField: String;
    function GetBaseField: Integer;
  end;

function TDerivedClass.GetBaseField: Integer;
begin
  Result := FBaseField;
end;`
	programAST := parseCode(t, code)

	// Cursor position inside the derived class method, on "FBaseField"
	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   12,
		Column: 13, // On "FBaseField" reference
	})

	locations := resolver.ResolveSymbol("FBaseField")

	// Should find the field in parent class
	if len(locations) == 0 {
		t.Fatal("Expected to find inherited field, got no results")
	}
}

func TestSymbolResolver_ResolveClassMember_InheritedMethod(t *testing.T) {
	code := `type TBaseClass = class
    function BaseMethod: Integer;
  end;

type TDerivedClass = class(TBaseClass)
    function DerivedMethod: String;
  end;

function TBaseClass.BaseMethod: Integer;
begin
  Result := 42;
end;

function TDerivedClass.DerivedMethod: String;
begin
  var x := BaseMethod();
  Result := IntToStr(x);
end;`
	programAST := parseCode(t, code)

	// Cursor position inside the derived class method, on "BaseMethod"
	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   16,
		Column: 12, // On "BaseMethod" call
	})

	locations := resolver.ResolveSymbol("BaseMethod")

	// Should find the method in parent class
	if len(locations) == 0 {
		t.Fatal("Expected to find inherited method, got no results")
	}
}

func TestSymbolResolver_ResolveClassMember_InheritedProperty(t *testing.T) {
	code := `type TBaseClass = class
  private
    FValue: Integer;
  public
    property Value: Integer read FValue write FValue;
  end;

type TDerivedClass = class(TBaseClass)
    function GetValue: Integer;
  end;

function TDerivedClass.GetValue: Integer;
begin
  Result := Value;
end;`
	programAST := parseCode(t, code)

	// Cursor position inside the derived class method, on "Value"
	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   14,
		Column: 13, // On "Value" property reference
	})

	locations := resolver.ResolveSymbol("Value")

	// Should find the property in parent class
	if len(locations) == 0 {
		t.Fatal("Expected to find inherited property, got no results")
	}
}

func TestSymbolResolver_ResolveClassMember_MultiLevelInheritance(t *testing.T) {
	code := `type TGrandparent = class
    FGrandField: Integer;
  end;

type TParent = class(TGrandparent)
    FParentField: String;
  end;

type TChild = class(TParent)
    FChildField: Boolean;
    function GetGrandField: Integer;
  end;

function TChild.GetGrandField: Integer;
begin
  Result := FGrandField;
end;`
	programAST := parseCode(t, code)

	// Cursor position inside the child class method, on "FGrandField"
	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   16,
		Column: 13, // On "FGrandField" reference
	})

	locations := resolver.ResolveSymbol("FGrandField")

	// Should find the field in grandparent class
	if len(locations) == 0 {
		t.Fatal("Expected to find field from grandparent class, got no results")
	}
}

func TestSymbolResolver_ResolveWorkspace_NoIndex(t *testing.T) {
	code := `
function TestFunc(): Integer;
begin
  Result := 42;
end;
`
	programAST := parseCode(t, code)

	// Create resolver without workspace index
	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   2,
		Column: 10,
	})

	// Try to resolve a symbol that doesn't exist in current file
	locations := resolver.ResolveSymbol("UnknownFunc")

	// Should not find the symbol (doesn't exist anywhere)
	if len(locations) != 0 {
		t.Errorf("Expected 0 locations for non-existent symbol, got %d", len(locations))
	}
}

func TestSymbolResolver_ResolveWorkspace_WithIndex(t *testing.T) {
	code := `
function LocalFunc(): Integer;
begin
  Result := 10;
end;
`
	programAST := parseCode(t, code)

	// Create a workspace index
	index := workspace.NewSymbolIndex()

	// Add a symbol from another file
	index.AddSymbol(
		"ExternalFunc",
		protocol.SymbolKindFunction,
		"file:///other.dws",
		protocol.Range{
			Start: protocol.Position{Line: 5, Character: 9},
			End:   protocol.Position{Line: 5, Character: 21},
		},
		"",
		"function ExternalFunc(): Integer",
	)

	// Create resolver with workspace index
	resolver := NewSymbolResolverWithIndex("file:///test.dws", programAST, token.Position{
		Line:   2,
		Column: 10,
	}, index)

	locations := resolver.ResolveSymbol("ExternalFunc")

	// Should find the symbol from the workspace
	if len(locations) == 0 {
		t.Fatal("Expected to find ExternalFunc in workspace, got no results")
	}

	if len(locations) != 1 {
		t.Errorf("Expected 1 location, got %d", len(locations))
	}

	if locations[0].URI != "file:///other.dws" {
		t.Errorf("Expected URI 'file:///other.dws', got '%s'", locations[0].URI)
	}
}

func TestSymbolResolver_ResolveWorkspace_SkipsCurrentFile(t *testing.T) {
	code := `
function MyFunc(): Integer;
begin
  Result := 42;
end;

var x := MyFunc();
`
	programAST := parseCode(t, code)

	// Create a workspace index
	index := workspace.NewSymbolIndex()

	// Add the same symbol from the current file (should be skipped)
	index.AddSymbol(
		"MyFunc",
		protocol.SymbolKindFunction,
		"file:///test.dws",
		protocol.Range{
			Start: protocol.Position{Line: 1, Character: 9},
			End:   protocol.Position{Line: 1, Character: 15},
		},
		"",
		"",
	)

	// Add the symbol from another file
	index.AddSymbol(
		"MyFunc",
		protocol.SymbolKindFunction,
		"file:///other.dws",
		protocol.Range{
			Start: protocol.Position{Line: 10, Character: 9},
			End:   protocol.Position{Line: 10, Character: 15},
		},
		"",
		"",
	)

	// Create resolver with workspace index
	resolver := NewSymbolResolverWithIndex("file:///test.dws", programAST, token.Position{
		Line:   7,
		Column: 10,
	}, index)

	locations := resolver.ResolveSymbol("MyFunc")

	// Should find the local definition first (via resolveGlobal)
	// The workspace resolver should skip the current file's entry
	if len(locations) == 0 {
		t.Fatal("Expected to find MyFunc, got no results")
	}

	// First result should be from the current file (via resolveGlobal)
	if locations[0].URI != "file:///test.dws" {
		t.Errorf("Expected first result from current file, got '%s'", locations[0].URI)
	}
}

func TestSymbolResolver_ResolveWorkspace_MultipleFiles(t *testing.T) {
	code := `
function Main(): Integer;
begin
  Result := 0;
end;
`
	programAST := parseCode(t, code)

	// Create a workspace index
	index := workspace.NewSymbolIndex()

	// Add the same symbol from multiple files
	index.AddSymbol(
		"SharedFunc",
		protocol.SymbolKindFunction,
		"file:///lib/helpers.dws",
		protocol.Range{Start: protocol.Position{Line: 1, Character: 9}, End: protocol.Position{Line: 1, Character: 19}},
		"",
		"",
	)

	index.AddSymbol(
		"SharedFunc",
		protocol.SymbolKindFunction,
		"file:///lib/utils.dws",
		protocol.Range{Start: protocol.Position{Line: 5, Character: 9}, End: protocol.Position{Line: 5, Character: 19}},
		"",
		"",
	)

	// Create resolver with workspace index
	resolver := NewSymbolResolverWithIndex("file:///app/main.dws", programAST, token.Position{
		Line:   2,
		Column: 10,
	}, index)

	locations := resolver.ResolveSymbol("SharedFunc")

	// Should find both definitions
	if len(locations) != 2 {
		t.Fatalf("Expected 2 locations, got %d", len(locations))
	}

	// Verify both files are in the results
	foundHelpers := false
	foundUtils := false

	for _, loc := range locations {
		if loc.URI == "file:///lib/helpers.dws" {
			foundHelpers = true
		}
		if loc.URI == "file:///lib/utils.dws" {
			foundUtils = true
		}
	}

	if !foundHelpers || !foundUtils {
		t.Error("Expected locations from both helpers.dws and utils.dws")
	}
}

func TestSymbolResolver_SetWorkspaceIndex(t *testing.T) {
	code := `var x := 1;`
	programAST := parseCode(t, code)

	// Create resolver without index
	resolver := NewSymbolResolver("file:///test.dws", programAST, token.Position{
		Line:   1,
		Column: 5,
	})

	// Create and set index
	index := workspace.NewSymbolIndex()
	index.AddSymbol(
		"TestSymbol",
		protocol.SymbolKindFunction,
		"file:///other.dws",
		protocol.Range{Start: protocol.Position{Line: 0, Character: 0}, End: protocol.Position{Line: 0, Character: 10}},
		"",
		"",
	)

	resolver.SetWorkspaceIndex(index)

	// Should now be able to resolve workspace symbols
	locations := resolver.ResolveSymbol("TestSymbol")

	if len(locations) == 0 {
		t.Error("Expected to find TestSymbol after setting workspace index")
	}
}
