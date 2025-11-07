package lsp

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

func TestGetIdentifierHover(t *testing.T) {
	tests := []struct {
		name     string
		ident    *ast.Identifier
		expected string
	}{
		{
			name: "identifier with type",
			ident: &ast.Identifier{
				Value: "myVar",
				Type: &ast.TypeAnnotation{
					Name: "Integer",
				},
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
		name     string
		classDecl *ast.ClassDecl
		expected []string
	}{
		{
			name: "simple class",
			classDecl: &ast.ClassDecl{
				Name:   &ast.Identifier{Value: "MyClass"},
				Fields: []*ast.FieldDecl{{}, {}},
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
	tests := []struct {
		name     string
		node     ast.Node
		expected string
	}{
		{
			name: "integer literal",
			node: &ast.IntegerLiteral{
				Value: 42,
				Token: token.Token{Type: token.INT},
			},
			expected: "42",
		},
		{
			name: "float literal",
			node: &ast.FloatLiteral{
				Value: 3.14,
				Token: token.Token{Type: token.FLOAT},
			},
			expected: "3.14",
		},
		{
			name: "string literal",
			node: &ast.StringLiteral{
				Value: "hello",
				Token: token.Token{Type: token.STRING},
			},
			expected: "hello",
		},
		{
			name: "boolean literal",
			node: &ast.BooleanLiteral{
				Value: true,
				Token: token.Token{Type: token.TRUE},
			},
			expected: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getHoverContent(tt.node, nil)

			if result == "" {
				t.Error("Expected hover content, got empty string")
			}

			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected hover to contain %q, got %q", tt.expected, result)
			}

			if !strings.Contains(result, "```dwscript") {
				t.Error("Expected hover to contain markdown code block")
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
