package lsp

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

func TestNodeToLocation(t *testing.T) {
	// Create a test node with position information
	ident := &ast.Identifier{
		Value: "testVar",
		Token: token.Token{
			Pos: token.Position{Line: 5, Column: 10},
		},
	}

	// Set the end position (normally done by parser)
	// For testing, we'll create a simple node with known positions
	uri := "file:///test/file.dws"

	location := nodeToLocation(ident, uri)

	if location == nil {
		t.Fatal("Expected location, got nil")
	}

	if location.URI != uri {
		t.Errorf("Expected URI %s, got %s", uri, location.URI)
	}

	// Check that position was converted from 1-based to 0-based
	if location.Range.Start.Line != 4 {
		t.Errorf("Expected line 4 (0-based), got %d", location.Range.Start.Line)
	}

	if location.Range.Start.Character != 9 {
		t.Errorf("Expected character 9 (0-based), got %d", location.Range.Start.Character)
	}
}

func TestFindDefinitionLocation_VarDecl(t *testing.T) {
	// Test finding definition for a variable declaration
	varDecl := &ast.VarDeclStatement{
		Names: []*ast.Identifier{
			{
				Value: "x",
				Token: token.Token{Pos: token.Position{Line: 1, Column: 5}},
			},
		},
		Type: &ast.TypeAnnotation{Name: "Integer"},
	}

	uri := "file:///test/vars.dws"
	programAST := &ast.Program{
		Statements: []ast.Statement{varDecl},
	}

	location := findDefinitionLocation(varDecl, nil, programAST, uri)

	if location == nil {
		t.Fatal("Expected location for variable declaration, got nil")
	}

	if location.URI != uri {
		t.Errorf("Expected URI %s, got %s", uri, location.URI)
	}
}

func TestFindDefinitionLocation_FunctionDecl(t *testing.T) {
	// Test finding definition for a function declaration
	funcDecl := &ast.FunctionDecl{
		Name: &ast.Identifier{
			Value: "TestFunc",
			Token: token.Token{Pos: token.Position{Line: 10, Column: 10}},
		},
		Parameters: []*ast.Parameter{},
		Body:       &ast.BlockStatement{},
	}

	uri := "file:///test/funcs.dws"
	programAST := &ast.Program{
		Statements: []ast.Statement{funcDecl},
	}

	location := findDefinitionLocation(funcDecl, nil, programAST, uri)

	if location == nil {
		t.Fatal("Expected location for function declaration, got nil")
	}

	if location.URI != uri {
		t.Errorf("Expected URI %s, got %s", uri, location.URI)
	}
}

func TestFindIdentifierDefinition_Variable(t *testing.T) {
	// Create a program with a variable declaration
	varName := &ast.Identifier{
		Value: "myVar",
		Token: token.Token{Pos: token.Position{Line: 1, Column: 5}},
	}

	varDecl := &ast.VarDeclStatement{
		Token: token.Token{Pos: token.Position{Line: 1, Column: 1}},
		Names: []*ast.Identifier{varName},
		Type:  &ast.TypeAnnotation{Name: "Integer"},
	}

	programAST := &ast.Program{
		Statements: []ast.Statement{varDecl},
	}

	// Create an identifier reference to search for
	identRef := &ast.Identifier{
		Value: "myVar",
		Token: token.Token{Pos: token.Position{Line: 5, Column: 10}},
	}

	uri := "file:///test/test.dws"
	location := findIdentifierDefinition(identRef, programAST, uri)

	if location == nil {
		t.Fatal("Expected to find variable definition, got nil")
	}

	if location.URI != uri {
		t.Errorf("Expected URI %s, got %s", uri, location.URI)
	}

	// The location should point to the declaration (line 0 in 0-based indexing)
	if location.Range.Start.Line != 0 {
		t.Errorf("Expected line 0, got %d", location.Range.Start.Line)
	}
}

func TestFindIdentifierDefinition_Function(t *testing.T) {
	// Create a program with a function declaration
	funcName := &ast.Identifier{
		Value: "MyFunction",
		Token: token.Token{Pos: token.Position{Line: 1, Column: 10}},
	}

	funcDecl := &ast.FunctionDecl{
		Name:       funcName,
		Parameters: []*ast.Parameter{},
		Body:       &ast.BlockStatement{},
	}

	programAST := &ast.Program{
		Statements: []ast.Statement{funcDecl},
	}

	// Create an identifier reference to the function
	identRef := &ast.Identifier{
		Value: "MyFunction",
		Token: token.Token{Pos: token.Position{Line: 10, Column: 5}},
	}

	uri := "file:///test/test.dws"
	location := findIdentifierDefinition(identRef, programAST, uri)

	if location == nil {
		t.Fatal("Expected to find function definition, got nil")
	}

	if location.URI != uri {
		t.Errorf("Expected URI %s, got %s", uri, location.URI)
	}
}

func TestFindIdentifierDefinition_FunctionParameter(t *testing.T) {
	// Create a program with a function that has parameters
	paramName := &ast.Identifier{
		Value: "param1",
		Token: token.Token{Pos: token.Position{Line: 1, Column: 20}},
	}

	funcDecl := &ast.FunctionDecl{
		Name: &ast.Identifier{
			Value: "TestFunc",
			Token: token.Token{Pos: token.Position{Line: 1, Column: 10}},
		},
		Parameters: []*ast.Parameter{
			{
				Name: paramName,
				Type: &ast.TypeAnnotation{Name: "Integer"},
			},
		},
		Body: &ast.BlockStatement{},
	}

	programAST := &ast.Program{
		Statements: []ast.Statement{funcDecl},
	}

	// Create an identifier reference to the parameter
	identRef := &ast.Identifier{
		Value: "param1",
		Token: token.Token{Pos: token.Position{Line: 3, Column: 5}},
	}

	uri := "file:///test/test.dws"
	location := findIdentifierDefinition(identRef, programAST, uri)

	if location == nil {
		t.Fatal("Expected to find parameter definition, got nil")
	}

	if location.URI != uri {
		t.Errorf("Expected URI %s, got %s", uri, location.URI)
	}
}

func TestFindIdentifierDefinition_Class(t *testing.T) {
	// Create a program with a class declaration
	className := &ast.Identifier{
		Value: "MyClass",
		Token: token.Token{Pos: token.Position{Line: 1, Column: 6}},
	}

	classDecl := &ast.ClassDecl{
		Name:    className,
		Fields:  []*ast.FieldDecl{},
		Methods: []*ast.FunctionDecl{},
	}

	programAST := &ast.Program{
		Statements: []ast.Statement{classDecl},
	}

	// Create an identifier reference to the class
	identRef := &ast.Identifier{
		Value: "MyClass",
		Token: token.Token{Pos: token.Position{Line: 10, Column: 10}},
	}

	uri := "file:///test/test.dws"
	location := findIdentifierDefinition(identRef, programAST, uri)

	if location == nil {
		t.Fatal("Expected to find class definition, got nil")
	}

	if location.URI != uri {
		t.Errorf("Expected URI %s, got %s", uri, location.URI)
	}
}

func TestFindIdentifierDefinition_NotFound(t *testing.T) {
	// Create a program without any declarations
	programAST := &ast.Program{
		Statements: []ast.Statement{},
	}

	// Try to find a non-existent identifier
	identRef := &ast.Identifier{
		Value: "nonExistent",
		Token: token.Token{Pos: token.Position{Line: 5, Column: 10}},
	}

	uri := "file:///test/test.dws"
	location := findIdentifierDefinition(identRef, programAST, uri)

	if location != nil {
		t.Errorf("Expected nil for non-existent identifier, got location: %v", location)
	}
}

func TestFindIdentifierDefinition_Constant(t *testing.T) {
	// Create a program with a constant declaration
	constName := &ast.Identifier{
		Value: "PI",
		Token: token.Token{Pos: token.Position{Line: 1, Column: 7}},
	}

	constDecl := &ast.ConstDecl{
		Name:  constName,
		Type:  &ast.TypeAnnotation{Name: "Float"},
		Value: &ast.FloatLiteral{Value: 3.14159},
	}

	programAST := &ast.Program{
		Statements: []ast.Statement{constDecl},
	}

	// Create an identifier reference to the constant
	identRef := &ast.Identifier{
		Value: "PI",
		Token: token.Token{Pos: token.Position{Line: 10, Column: 5}},
	}

	uri := "file:///test/test.dws"
	location := findIdentifierDefinition(identRef, programAST, uri)

	if location == nil {
		t.Fatal("Expected to find constant definition, got nil")
	}

	if location.URI != uri {
		t.Errorf("Expected URI %s, got %s", uri, location.URI)
	}
}

func TestFindIdentifierDefinition_Enum(t *testing.T) {
	// Create a program with an enum declaration
	enumName := &ast.Identifier{
		Value: "TColor",
		Token: token.Token{Pos: token.Position{Line: 1, Column: 6}},
	}

	enumDecl := &ast.EnumDecl{
		Name: enumName,
		Values: []ast.EnumValue{
			{Name: "Red"},
			{Name: "Green"},
			{Name: "Blue"},
		},
	}

	programAST := &ast.Program{
		Statements: []ast.Statement{enumDecl},
	}

	// Create an identifier reference to the enum
	identRef := &ast.Identifier{
		Value: "TColor",
		Token: token.Token{Pos: token.Position{Line: 10, Column: 5}},
	}

	uri := "file:///test/test.dws"
	location := findIdentifierDefinition(identRef, programAST, uri)

	if location == nil {
		t.Fatal("Expected to find enum definition, got nil")
	}

	if location.URI != uri {
		t.Errorf("Expected URI %s, got %s", uri, location.URI)
	}
}

func TestFindIdentifierDefinition_EnumValue(t *testing.T) {
	// Create a program with an enum that has values
	enumDecl := &ast.EnumDecl{
		Name: &ast.Identifier{
			Value: "TColor",
			Token: token.Token{Pos: token.Position{Line: 1, Column: 6}},
		},
		Values: []ast.EnumValue{
			{Name: "Red"},
			{Name: "Green"},
			{Name: "Blue"},
		},
	}

	programAST := &ast.Program{
		Statements: []ast.Statement{enumDecl},
	}

	// Create an identifier reference to an enum value
	identRef := &ast.Identifier{
		Value: "Red",
		Token: token.Token{Pos: token.Position{Line: 10, Column: 5}},
	}

	uri := "file:///test/test.dws"
	location := findIdentifierDefinition(identRef, programAST, uri)

	if location == nil {
		t.Fatal("Expected to find enum value definition, got nil")
	}

	if location.URI != uri {
		t.Errorf("Expected URI %s, got %s", uri, location.URI)
	}
}

func TestFindDefinitionLocation_UnsupportedNode(t *testing.T) {
	// Test with a node type that doesn't have definition support
	node := &ast.BlockStatement{}

	programAST := &ast.Program{
		Statements: []ast.Statement{},
	}

	uri := "file:///test/test.dws"
	location := findDefinitionLocation(node, nil, programAST, uri)

	if location != nil {
		t.Errorf("Expected nil for unsupported node type, got location: %v", location)
	}
}
