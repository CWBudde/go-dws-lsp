package analysis

import (
	"testing"

	"github.com/CWBudde/go-dws-lsp/internal/server"
)

var countParameterIndexTests = []struct {
	name      string
	code      string
	line      int
	character int
	expected  int
}{
	{
		name:      "cursor after opening parenthesis",
		code:      "foo(",
		line:      0,
		character: 4,
		expected:  0,
	},
	{
		name:      "cursor in first parameter",
		code:      "foo(x",
		line:      0,
		character: 5,
		expected:  0,
	},
	{
		name:      "cursor after comma",
		code:      "foo(x, ",
		line:      0,
		character: 7,
		expected:  1,
	},
	{
		name:      "cursor in second parameter",
		code:      "foo(x, y",
		line:      0,
		character: 8,
		expected:  1,
	},
	{
		name:      "cursor after second comma",
		code:      "foo(x, y, ",
		line:      0,
		character: 10,
		expected:  2,
	},
	{
		name:      "nested function call - outer",
		code:      "foo(bar(",
		line:      0,
		character: 8,
		expected:  0, // Inside bar()
	},
	{
		name:      "nested function call - after inner",
		code:      "foo(bar(), ",
		line:      0,
		character: 11,
		expected:  1,
	},
	{
		name:      "string with comma inside",
		code:      `foo("x, y", `,
		line:      0,
		character: 12,
		expected:  1, // Comma inside string shouldn't count
	},
	{
		name:      "empty parameter list",
		code:      "foo()",
		line:      0,
		character: 4,
		expected:  0,
	},
	{
		name:      "five parameters",
		code:      "foo(a, b, c, d, e",
		line:      0,
		character: 17,
		expected:  4,
	},
}

// TestCountParameterIndex tests parameter index counting at various positions.
func TestCountParameterIndex(t *testing.T) {
	for _, tt := range countParameterIndexTests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CountParameterIndex(tt.code, tt.line, tt.character)
			if err != nil {
				t.Fatalf("CountParameterIndex failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected parameter index %d, got %d for code: %q at position %d:%d",
					tt.expected, result, tt.code, tt.line, tt.character)
			}
		})
	}
}

var findFunctionAtCallTests = []struct {
	name      string
	code      string
	line      int
	character int
	expected  string
}{
	{
		name:      "simple function call",
		code:      "PrintLn(",
		line:      0,
		character: 8,
		expected:  "PrintLn",
	},
	{
		name:      "function with parameters",
		code:      "foo(x, y",
		line:      0,
		character: 8,
		expected:  "foo",
	},
	{
		name:      "qualified name",
		code:      "obj.Method(",
		line:      0,
		character: 11,
		expected:  "obj.Method",
	},
	{
		name:      "function with whitespace",
		code:      "Calculate  (x",
		line:      0,
		character: 13,
		expected:  "Calculate",
	},
	{
		name:      "nested call - inner function",
		code:      "foo(bar(",
		line:      0,
		character: 8,
		expected:  "bar",
	},
	{
		name:      "multiline call",
		code:      "foo(\n  x",
		line:      1,
		character: 3,
		expected:  "foo",
	},
}

// TestFindFunctionAtCall tests finding function names from cursor position.
func TestFindFunctionAtCall(t *testing.T) {
	for _, tt := range findFunctionAtCallTests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &server.Document{
				Text: tt.code,
			}

			result, err := FindFunctionAtCall(doc, tt.line, tt.character)
			if err != nil {
				t.Fatalf("FindFunctionAtCall failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected function name %q, got %q for code: %q at position %d:%d",
					tt.expected, result, tt.code, tt.line, tt.character)
			}
		})
	}
}

// TestFindParameterIndexFromText tests the simplified parameter counting.
func TestFindParameterIndexFromText(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		line      int
		character int
		expected  int
	}{
		{
			name:      "no commas",
			text:      "foo(x",
			line:      0,
			character: 5,
			expected:  0,
		},
		{
			name:      "one comma",
			text:      "foo(x, y",
			line:      0,
			character: 8,
			expected:  1,
		},
		{
			name:      "two commas",
			text:      "foo(x, y, z",
			line:      0,
			character: 11,
			expected:  2,
		},
		{
			name:      "nested parens",
			text:      "foo(bar(a, b), c",
			line:      0,
			character: 16,
			expected:  1, // Should count outer comma only
		},
		{
			name:      "string with comma",
			text:      `foo("a, b", c`,
			line:      0,
			character: 13,
			expected:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findParameterIndexFromText(tt.text, tt.line, tt.character)

			if result != tt.expected {
				t.Errorf("Expected parameter index %d, got %d for text: %q at position %d:%d",
					tt.expected, result, tt.text, tt.line, tt.character)
			}
		})
	}
}

// TestDetermineCallContextWithTempAST tests call context determination with temporary AST.
var determineCallContextTests = []struct {
	name             string
	code             string
	line             int
	character        int
	expectContext    bool
	expectedFunction string
	expectedParamIdx int
}{
	{
		name: "complete function call",
		code: `function foo(x: Integer, y: String);
begin
end;

begin
  foo(42, 'test');
end.`,
		line:             5,
		character:        7, // After '('
		expectContext:    true,
		expectedFunction: "foo",
		expectedParamIdx: 0,
	},
	{
		name: "incomplete function call",
		code: `function bar(a: Integer);
begin
end;

begin
  bar(
end.`,
		line:             5,
		character:        6, // After '('
		expectContext:    true,
		expectedFunction: "bar",
		expectedParamIdx: 0,
	},
	{
		name: "built-in function call",
		code: `begin
  PrintLn('Hello',
end.`,
		line:             1,
		character:        18, // After comma
		expectContext:    true,
		expectedFunction: "PrintLn",
		expectedParamIdx: 1,
	},
}

func TestDetermineCallContextWithTempAST(t *testing.T) {
	for _, tt := range determineCallContextTests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &server.Document{
				Text: tt.code,
				URI:  "file:///test.dws",
			}

			// Parse the document
			program, _, err := ParseDocument(tt.code, "test.dws")
			if err != nil {
				t.Logf("Warning: ParseDocument failed: %v (expected for incomplete code)", err)
			}

			doc.Program = program

			ctx, err := DetermineCallContextWithTempAST(doc, tt.line, tt.character)
			if err != nil {
				t.Fatalf("DetermineCallContextWithTempAST failed: %v", err)
			}

			if tt.expectContext {
				if ctx == nil {
					t.Fatal("Expected call context, got nil")
				}

				if ctx.FunctionName != tt.expectedFunction {
					t.Errorf("Expected function name %q, got %q", tt.expectedFunction, ctx.FunctionName)
				}

				if ctx.ParameterIndex != tt.expectedParamIdx {
					t.Errorf("Expected parameter index %d, got %d", tt.expectedParamIdx, ctx.ParameterIndex)
				}

				if !ctx.IsInsideCall {
					t.Error("Expected IsInsideCall to be true")
				}
			} else if ctx != nil {
				t.Errorf("Expected no call context, got context for function %q", ctx.FunctionName)
			}
		})
	}
}
