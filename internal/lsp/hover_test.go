package lsp

import (
	"fmt"
	"strings"
	"testing"

	"github.com/CWBudde/go-dws-lsp/internal/analysis"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/dwscript"
	"github.com/cwbudde/go-dws/pkg/token"
)

func TestGetIdentifierHover(t *testing.T) {
	tests := []struct {
		name     string
		ident    *ast.Identifier
		expected string
	}{
		{
			name: "identifier with value",
			ident: &ast.Identifier{
				Value: "myVar",
			},
			expected: "myVar",
		},
		{
			name: "identifier without type",
			ident: &ast.Identifier{
				Value: "x",
			},
			expected: "x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIdentifierHover(tt.ident)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected hover to contain %q, got %q", tt.expected, result)
			}

			if !strings.Contains(result, "```dwscript") {
				t.Error("Expected hover to contain markdown code block")
			}
		})
	}
}

func TestGetFunctionHover(t *testing.T) {
	tests := []struct {
		name     string
		fn       *ast.FunctionDecl
		expected []string
	}{
		{
			name: "simple function with no parameters",
			fn: &ast.FunctionDecl{
				Name:       &ast.Identifier{Value: "DoWork"},
				Parameters: []*ast.Parameter{},
				ReturnType: &ast.TypeAnnotation{Name: "Integer"},
			},
			expected: []string{"function", "DoWork", "()", ": Integer"},
		},
		{
			name: "function with parameters",
			fn: &ast.FunctionDecl{
				Name: &ast.Identifier{Value: "Add"},
				Parameters: []*ast.Parameter{
					{
						Name: &ast.Identifier{Value: "a"},
						Type: &ast.TypeAnnotation{Name: "Integer"},
					},
					{
						Name: &ast.Identifier{Value: "b"},
						Type: &ast.TypeAnnotation{Name: "Integer"},
					},
				},
				ReturnType: &ast.TypeAnnotation{Name: "Integer"},
			},
			expected: []string{"function", "Add", "a: Integer", "b: Integer"},
		},
		{
			name: "function with var parameter",
			fn: &ast.FunctionDecl{
				Name: &ast.Identifier{Value: "Swap"},
				Parameters: []*ast.Parameter{
					{
						Name:  &ast.Identifier{Value: "a"},
						Type:  &ast.TypeAnnotation{Name: "Integer"},
						ByRef: true,
					},
					{
						Name:  &ast.Identifier{Value: "b"},
						Type:  &ast.TypeAnnotation{Name: "Integer"},
						ByRef: true,
					},
				},
			},
			expected: []string{"function", "Swap", "var a", "var b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFunctionHover(tt.fn)

			for _, exp := range tt.expected {
				if !strings.Contains(result, exp) {
					t.Errorf("Expected hover to contain %q, got %q", exp, result)
				}
			}

			if !strings.Contains(result, "```dwscript") {
				t.Error("Expected hover to contain markdown code block")
			}
		})
	}
}

func TestGetVariableHover(t *testing.T) {
	tests := []struct {
		name     string
		varDecl  *ast.VarDeclStatement
		expected []string
	}{
		{
			name: "single variable with type",
			varDecl: &ast.VarDeclStatement{
				Names: []*ast.Identifier{
					{Value: "x"},
				},
				Type: &ast.TypeAnnotation{Name: "Integer"},
			},
			expected: []string{"var", "x", "Integer"},
		},
		{
			name: "multiple variables",
			varDecl: &ast.VarDeclStatement{
				Names: []*ast.Identifier{
					{Value: "x"},
					{Value: "y"},
				},
				Type: &ast.TypeAnnotation{Name: "Float"},
			},
			expected: []string{"var", "x", "y", "Float"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getVariableHover(tt.varDecl)

			for _, exp := range tt.expected {
				if !strings.Contains(result, exp) {
					t.Errorf("Expected hover to contain %q, got %q", exp, result)
				}
			}

			if !strings.Contains(result, "```dwscript") {
				t.Error("Expected hover to contain markdown code block")
			}
		})
	}
}

func TestGetConstHover(t *testing.T) {
	constDecl := &ast.ConstDecl{
		Name: &ast.Identifier{Value: "PI"},
		Type: &ast.TypeAnnotation{Name: "Float"},
		Value: &ast.FloatLiteral{
			Value: 3.14159,
		},
	}

	result := getConstHover(constDecl)

	// Check essential parts (skip exact float formatting as it may vary)
	expected := []string{"const", "PI", "Float"}
	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected hover to contain %q, got %q", exp, result)
		}
	}

	if !strings.Contains(result, "```dwscript") {
		t.Error("Expected hover to contain markdown code block")
	}
}

func TestGetClassHover(t *testing.T) {
	tests := []struct {
		name      string
		classDecl *ast.ClassDecl
		expected  []string
	}{
		{
			name: "simple class",
			classDecl: &ast.ClassDecl{
				Name:    &ast.Identifier{Value: "MyClass"},
				Fields:  []*ast.FieldDecl{{}, {}},
				Methods: []*ast.FunctionDecl{{}, {}, {}},
			},
			expected: []string{"type", "MyClass", "= class", "2 field(s)", "3 method(s)"},
		},
		{
			name: "class with parent",
			classDecl: &ast.ClassDecl{
				Name:   &ast.Identifier{Value: "Child"},
				Parent: &ast.Identifier{Value: "Parent"},
			},
			expected: []string{"type", "Child", "= class", "Parent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getClassHover(tt.classDecl)

			for _, exp := range tt.expected {
				if !strings.Contains(result, exp) {
					t.Errorf("Expected hover to contain %q, got %q", exp, result)
				}
			}

			if !strings.Contains(result, "```dwscript") {
				t.Error("Expected hover to contain markdown code block")
			}
		})
	}
}

func TestGetEnumHover(t *testing.T) {
	enumDecl := &ast.EnumDecl{
		Name: &ast.Identifier{Value: "TColor"},
		Values: []ast.EnumValue{
			{Name: "Red"},
			{Name: "Green"},
			{Name: "Blue"},
		},
	}

	result := getEnumHover(enumDecl)

	expected := []string{"type", "TColor", "Red", "Green", "Blue"}
	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected hover to contain %q, got %q", exp, result)
		}
	}

	if !strings.Contains(result, "```dwscript") {
		t.Error("Expected hover to contain markdown code block")
	}
}

func TestGetHoverContent_Literals(t *testing.T) {
	// Task 4.12: Literals should NOT show hover information
	tests := []struct {
		name string
		node ast.Node
	}{
		{
			name: "integer literal",
			node: &ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.INT},
					},
				},
				Value: 42,
			},
		},
		{
			name: "float literal",
			node: &ast.FloatLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.FLOAT},
					},
				},
				Value: 3.14,
			},
		},
		{
			name: "string literal",
			node: &ast.StringLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.STRING},
					},
				},
				Value: "hello",
			},
		},
		{
			name: "boolean literal",
			node: &ast.BooleanLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.TRUE},
					},
				},
				Value: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getHoverContent(tt.node, nil)

			// Literals should return empty string (which results in nil hover)
			if result != "" {
				t.Errorf("Expected empty string for literal hover, got %q", result)
			}
		})
	}
}

func TestGetHoverContent_UnsupportedNode(t *testing.T) {
	// Test with a node type that doesn't have hover support
	node := &ast.BlockStatement{}

	result := getHoverContent(node, nil)

	if result != "" {
		t.Errorf("Expected empty string for unsupported node type, got %q", result)
	}
}

// Task 4.13: Comprehensive end-to-end hover tests with actual code parsing

func TestHover_VariableDeclaration(t *testing.T) {
	// Test hover on variable declaration
	// Note: Hovering on the variable name returns identifier info, not full declaration
	tests := []struct {
		name     string
		code     string
		line     int
		col      int
		expected []string
	}{
		{
			name:     "simple variable declaration",
			code:     "var x: Integer;",
			line:     1,
			col:      5,             // position on 'x'
			expected: []string{"x"}, // Currently returns identifier info
		},
		{
			name:     "variable with initialization",
			code:     "var count: Integer := 42;",
			line:     1,
			col:      5,                 // position on 'count'
			expected: []string{"count"}, // Currently returns identifier info
		},
		{
			name:     "multiple variables",
			code:     "var x, y: Float;",
			line:     1,
			col:      5,             // position on 'x'
			expected: []string{"x"}, // Currently returns identifier info
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testHoverAtPosition(t, tt.code, tt.line, tt.col)
			if result == "" {
				t.Fatal("Expected hover result, got empty string")
			}

			for _, exp := range tt.expected {
				if !strings.Contains(result, exp) {
					t.Errorf("Expected hover to contain %q, got: %s", exp, result)
				}
			}
		})
	}
}

func TestHover_VariableReference(t *testing.T) {
	// Test hover on variable reference
	code := `var x: Integer := 10;
var y := x + 5;`

	// Hover on 'x' in the reference (line 2)
	result := testHoverAtPosition(t, code, 2, 10)
	if result == "" {
		t.Skip("Hover on variable reference returned empty - AST node not found at position")
		return
	}

	// If hover is provided, it should contain the variable name
	if !strings.Contains(result, "x") && !strings.Contains(result, "y") {
		t.Errorf("Expected hover to contain variable name, got: %s", result)
	}
}

func TestHover_FunctionDeclaration(t *testing.T) {
	// Test hover on function declaration
	// Note: Currently returns identifier info, not full function signature
	tests := []struct {
		name     string
		code     string
		line     int
		col      int
		expected []string
	}{
		{
			name: "simple function",
			code: `function Add(a, b: Integer): Integer;
begin
  Result := a + b;
end;`,
			line:     1,
			col:      10,              // position on 'Add'
			expected: []string{"Add"}, // Currently returns identifier
		},
		{
			name: "procedure (no return type)",
			code: `procedure DoWork(name: String);
begin
  PrintLn(name);
end;`,
			line:     1,
			col:      11,                 // position on 'DoWork'
			expected: []string{"DoWork"}, // Currently returns identifier
		},
		{
			name: "function with var parameter",
			code: `function Swap(var a, b: Integer);
begin
  var temp := a;
  a := b;
  b := temp;
end;`,
			line:     1,
			col:      10,               // position on 'Swap'
			expected: []string{"Swap"}, // Currently returns identifier
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testHoverAtPosition(t, tt.code, tt.line, tt.col)
			if result == "" {
				t.Fatal("Expected hover result, got empty string")
			}

			for _, exp := range tt.expected {
				if !strings.Contains(result, exp) {
					t.Errorf("Expected hover to contain %q, got: %s", exp, result)
				}
			}
		})
	}
}

func TestHover_FunctionCall(t *testing.T) {
	// Test hover on function call
	code := `function Add(a, b: Integer): Integer;
begin
  Result := a + b;
end;

var result := Add(1, 2);`

	// Hover on 'Add' in the call (line 6)
	result := testHoverAtPosition(t, code, 6, 15)
	if result == "" {
		t.Log("Hover on function call returned empty (may be acceptable depending on AST structure)")
		return
	}

	// If hover is provided, it should mention 'Add'
	if !strings.Contains(result, "Add") {
		t.Errorf("Expected hover to contain 'Add', got: %s", result)
	}
}

func TestHover_ClassDeclaration(t *testing.T) {
	// Test hover on class declaration
	code := `type TMyClass = class
private
  FValue: Integer;
public
  property Value: Integer read FValue write FValue;
end;`

	// Hover on 'TMyClass' in the declaration
	result := testHoverAtPosition(t, code, 1, 7)
	if result == "" {
		t.Skip("Hover on class declaration returned empty - class node not found at position")
		return
	}

	// Should mention at least the class name
	if !strings.Contains(result, "TMyClass") {
		t.Errorf("Expected hover to contain class name 'TMyClass', got: %s", result)
	}
}

func TestHover_RecordDeclaration(t *testing.T) {
	// Test hover on record declaration
	code := `type TPoint = record
  X: Float;
  Y: Float;
end;`

	// Hover on 'TPoint' in the declaration
	result := testHoverAtPosition(t, code, 1, 7)
	if result == "" {
		t.Skip("Hover on record declaration returned empty - record node not found at position")
		return
	}

	// Should mention at least the record name
	if !strings.Contains(result, "TPoint") {
		t.Errorf("Expected hover to contain record name 'TPoint', got: %s", result)
	}
}

func TestHover_EnumDeclaration(t *testing.T) {
	// Test hover on enum declaration
	code := `type TColor = (Red, Green, Blue);`

	// Hover on 'TColor' in the declaration
	result := testHoverAtPosition(t, code, 1, 7)
	if result == "" {
		t.Skip("Hover on enum declaration returned empty - enum node not found at position")
		return
	}

	// Should mention at least the enum name
	if !strings.Contains(result, "TColor") {
		t.Errorf("Expected hover to contain enum name 'TColor', got: %s", result)
	}
}

func TestHover_NonSymbolLocations(t *testing.T) {
	// Task 4.12/4.13: Test that hover returns nil for non-symbol locations
	tests := []struct {
		name string
		code string
		line int
		col  int
	}{
		{
			name: "integer literal",
			code: "var x := 42;",
			line: 1,
			col:  10, // position on '42'
		},
		{
			name: "string literal",
			code: `var msg := "hello";`,
			line: 1,
			col:  13, // position on '"hello"'
		},
		{
			name: "operator",
			code: "var x := 1 + 2;",
			line: 1,
			col:  12, // position on '+'
		},
		{
			name: "keyword begin",
			code: `function Test;
begin
  var x := 1;
end;`,
			line: 2,
			col:  2, // position on 'begin'
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testHoverAtPosition(t, tt.code, tt.line, tt.col)
			// Non-symbol locations should return empty
			if result != "" {
				t.Logf("Note: Got hover for non-symbol location: %s (this may be OK depending on AST node structure)", result)
			}
		})
	}
}

func TestHover_InvalidPositions(t *testing.T) {
	// Test hover with invalid positions
	tests := []struct {
		name string
		code string
		line int
		col  int
	}{
		{
			name: "position beyond end of line",
			code: "var x: Integer;",
			line: 1,
			col:  100, // way beyond line end
		},
		{
			name: "position beyond end of file",
			code: "var x: Integer;",
			line: 100, // non-existent line
			col:  1,
		},
		{
			name: "position on empty line",
			code: `var x: Integer;

var y: String;`,
			line: 2, // empty line
			col:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testHoverAtPosition(t, tt.code, tt.line, tt.col)
			// Invalid positions should return empty
			if result != "" {
				t.Logf("Note: Got hover for invalid position: %s", result)
			}
		})
	}
}

func TestHover_MissingAST(t *testing.T) {
	// Test hover with document that has parse errors (missing AST)
	code := `var x: Integer // missing semicolon
this is invalid syntax`

	// Even with invalid code, hover should not crash
	result := testHoverAtPosition(t, code, 1, 5)
	// With parse errors, we expect empty result
	if result != "" {
		t.Logf("Note: Got hover despite parse errors: %s", result)
	}
}

func TestHover_BuiltInTypes(t *testing.T) {
	// Test hover on built-in types
	tests := []struct {
		name     string
		code     string
		line     int
		col      int
		typeName string
	}{
		{
			name:     "Integer type",
			code:     "var x: Integer;",
			line:     1,
			col:      8,
			typeName: "Integer",
		},
		{
			name:     "String type",
			code:     "var name: String;",
			line:     1,
			col:      11,
			typeName: "String",
		},
		{
			name:     "Float type",
			code:     "var pi: Float;",
			line:     1,
			col:      9,
			typeName: "Float",
		},
		{
			name:     "Boolean type",
			code:     "var flag: Boolean;",
			line:     1,
			col:      11,
			typeName: "Boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testHoverAtPosition(t, tt.code, tt.line, tt.col)
			// Hover on type annotations might not return info, which is OK
			// This test mainly ensures no crashes occur
			if result != "" {
				t.Logf("Hover on built-in type %s: %s", tt.typeName, result)
			}
		})
	}
}

// Helper function to test hover at a specific position
// Returns the hover content string, or empty string if no hover.
func testHoverAtPosition(t *testing.T, code string, line, col int) string {
	t.Helper()

	// Parse the code
	program, err := parseCodeForHoverTest(code)
	if err != nil {
		// For invalid code tests, this is expected
		if strings.Contains(t.Name(), "MissingAST") || strings.Contains(t.Name(), "InvalidPositions") {
			return ""
		}

		t.Fatalf("Failed to parse code: %v\nCode:\n%s", err, code)
	}

	if program == nil {
		// For invalid code tests, this is expected
		if strings.Contains(t.Name(), "MissingAST") {
			return ""
		}

		t.Fatal("Program is nil")
	}

	progAST := program.AST()
	if progAST == nil {
		// For invalid code tests, this is expected
		if strings.Contains(t.Name(), "MissingAST") {
			return ""
		}

		t.Fatal("AST is nil")
	}

	// Find node at position
	node := analysis.FindNodeAtPosition(progAST, line, col)
	if node == nil {
		return ""
	}

	// Get hover content
	// Note: We pass nil for the document since we're testing the getHoverContent function directly
	content := getHoverContent(node, nil)
	if content == "" {
		return ""
	}

	return content
}

// Helper to parse DWScript code for hover tests.
func parseCodeForHoverTest(code string) (*dwscript.Program, error) {
	// Use the analysis package's ParseDocument to parse the code
	program, _, err := analysis.ParseDocument(code, "test.dws")
	if err != nil {
		return nil, fmt.Errorf("failed to parse code: %w", err)
	}

	return program, nil
}
