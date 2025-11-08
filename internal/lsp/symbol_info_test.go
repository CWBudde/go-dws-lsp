package lsp

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

func TestIdentifySymbolAtPosition_Identifier(t *testing.T) {
	ident := &ast.Identifier{
		Value: "myVar",
		Token: token.Token{Pos: token.Position{Line: 5, Column: 10}},
	}

	info := IdentifySymbolAtPosition(ident)

	if info == nil {
		t.Fatal("Expected symbol info for identifier, got nil")
	}

	if info.Name != "myVar" {
		t.Errorf("Expected name 'myVar', got '%s'", info.Name)
	}

	if info.Kind != SymbolKindUnknown {
		t.Errorf("Expected kind %s, got %s", SymbolKindUnknown, info.Kind)
	}
}

func TestIdentifySymbolAtPosition_VarDecl(t *testing.T) {
	varDecl := &ast.VarDeclStatement{
		Names: []*ast.Identifier{
			{Value: "x", Token: token.Token{Pos: token.Position{Line: 1, Column: 5}}},
		},
		Type: &ast.TypeAnnotation{Name: "Integer"},
	}

	info := IdentifySymbolAtPosition(varDecl)

	if info == nil {
		t.Fatal("Expected symbol info for variable declaration, got nil")
	}

	if info.Name != "x" {
		t.Errorf("Expected name 'x', got '%s'", info.Name)
	}

	if info.Kind != SymbolKindVariable {
		t.Errorf("Expected kind %s, got %s", SymbolKindVariable, info.Kind)
	}
}

func TestIdentifySymbolAtPosition_FunctionDecl(t *testing.T) {
	funcDecl := &ast.FunctionDecl{
		Name: &ast.Identifier{
			Value: "MyFunc",
			Token: token.Token{Pos: token.Position{Line: 1, Column: 10}},
		},
		Parameters: []*ast.Parameter{},
		Body:       &ast.BlockStatement{},
	}

	info := IdentifySymbolAtPosition(funcDecl)

	if info == nil {
		t.Fatal("Expected symbol info for function declaration, got nil")
	}

	if info.Name != "MyFunc" {
		t.Errorf("Expected name 'MyFunc', got '%s'", info.Name)
	}

	if info.Kind != SymbolKindFunction {
		t.Errorf("Expected kind %s, got %s", SymbolKindFunction, info.Kind)
	}
}

func TestIdentifySymbolAtPosition_ClassDecl(t *testing.T) {
	classDecl := &ast.ClassDecl{
		Name: &ast.Identifier{
			Value: "MyClass",
			Token: token.Token{Pos: token.Position{Line: 1, Column: 6}},
		},
		Fields:  []*ast.FieldDecl{},
		Methods: []*ast.FunctionDecl{},
	}

	info := IdentifySymbolAtPosition(classDecl)

	if info == nil {
		t.Fatal("Expected symbol info for class declaration, got nil")
	}

	if info.Name != "MyClass" {
		t.Errorf("Expected name 'MyClass', got '%s'", info.Name)
	}

	if info.Kind != SymbolKindClass {
		t.Errorf("Expected kind %s, got %s", SymbolKindClass, info.Kind)
	}
}

func TestIdentifySymbolAtPosition_ConstDecl(t *testing.T) {
	constDecl := &ast.ConstDecl{
		Name: &ast.Identifier{
			Value: "PI",
			Token: token.Token{Pos: token.Position{Line: 1, Column: 7}},
		},
		Type:  &ast.TypeAnnotation{Name: "Float"},
		Value: &ast.FloatLiteral{Value: 3.14159},
	}

	info := IdentifySymbolAtPosition(constDecl)

	if info == nil {
		t.Fatal("Expected symbol info for constant declaration, got nil")
	}

	if info.Name != "PI" {
		t.Errorf("Expected name 'PI', got '%s'", info.Name)
	}

	if info.Kind != SymbolKindConstant {
		t.Errorf("Expected kind %s, got %s", SymbolKindConstant, info.Kind)
	}
}

func TestIdentifySymbolAtPosition_EnumDecl(t *testing.T) {
	enumDecl := &ast.EnumDecl{
		Name: &ast.Identifier{
			Value: "TColor",
			Token: token.Token{Pos: token.Position{Line: 1, Column: 6}},
		},
		Values: []ast.EnumValue{
			{Name: "Red"},
			{Name: "Green"},
		},
	}

	info := IdentifySymbolAtPosition(enumDecl)

	if info == nil {
		t.Fatal("Expected symbol info for enum declaration, got nil")
	}

	if info.Name != "TColor" {
		t.Errorf("Expected name 'TColor', got '%s'", info.Name)
	}

	if info.Kind != SymbolKindEnum {
		t.Errorf("Expected kind %s, got %s", SymbolKindEnum, info.Kind)
	}
}

func TestIdentifySymbolAtPosition_FieldDecl(t *testing.T) {
	fieldDecl := &ast.FieldDecl{
		Name: &ast.Identifier{
			Value: "FValue",
			Token: token.Token{Pos: token.Position{Line: 3, Column: 5}},
		},
		Type: &ast.TypeAnnotation{Name: "Integer"},
	}

	info := IdentifySymbolAtPosition(fieldDecl)

	if info == nil {
		t.Fatal("Expected symbol info for field declaration, got nil")
	}

	if info.Name != "FValue" {
		t.Errorf("Expected name 'FValue', got '%s'", info.Name)
	}

	if info.Kind != SymbolKindField {
		t.Errorf("Expected kind %s, got %s", SymbolKindField, info.Kind)
	}
}

func TestIdentifySymbolAtPosition_MemberExpression(t *testing.T) {
	memberExpr := &ast.MemberAccessExpression{
		Object: &ast.Identifier{
			Value: "obj",
			Token: token.Token{Pos: token.Position{Line: 5, Column: 1}},
		},
		Member: &ast.Identifier{
			Value: "field",
			Token: token.Token{Pos: token.Position{Line: 5, Column: 5}},
		},
	}

	info := IdentifySymbolAtPosition(memberExpr)

	if info == nil {
		t.Fatal("Expected symbol info for member expression, got nil")
	}

	if info.Name != "field" {
		t.Errorf("Expected name 'field', got '%s'", info.Name)
	}

	// Kind is unknown until we resolve the member
	if info.Kind != SymbolKindUnknown {
		t.Errorf("Expected kind %s, got %s", SymbolKindUnknown, info.Kind)
	}
}

func TestIdentifySymbolAtPosition_CallExpression_Function(t *testing.T) {
	callExpr := &ast.CallExpression{
		Function: &ast.Identifier{
			Value: "DoWork",
			Token: token.Token{Pos: token.Position{Line: 10, Column: 1}},
		},
		Arguments: []ast.Expression{},
	}

	info := IdentifySymbolAtPosition(callExpr)

	if info == nil {
		t.Fatal("Expected symbol info for function call, got nil")
	}

	if info.Name != "DoWork" {
		t.Errorf("Expected name 'DoWork', got '%s'", info.Name)
	}

	if info.Kind != SymbolKindFunction {
		t.Errorf("Expected kind %s, got %s", SymbolKindFunction, info.Kind)
	}
}

func TestIdentifySymbolAtPosition_CallExpression_Method(t *testing.T) {
	callExpr := &ast.CallExpression{
		Function: &ast.MemberAccessExpression{
			Object: &ast.Identifier{
				Value: "obj",
				Token: token.Token{Pos: token.Position{Line: 10, Column: 1}},
			},
			Member: &ast.Identifier{
				Value: "DoWork",
				Token: token.Token{Pos: token.Position{Line: 10, Column: 5}},
			},
		},
		Arguments: []ast.Expression{},
	}

	info := IdentifySymbolAtPosition(callExpr)

	if info == nil {
		t.Fatal("Expected symbol info for method call, got nil")
	}

	if info.Name != "DoWork" {
		t.Errorf("Expected name 'DoWork', got '%s'", info.Name)
	}

	if info.Kind != SymbolKindMethod {
		t.Errorf("Expected kind %s, got %s", SymbolKindMethod, info.Kind)
	}
}

func TestIdentifySymbolAtPosition_Nil(t *testing.T) {
	info := IdentifySymbolAtPosition(nil)

	if info != nil {
		t.Errorf("Expected nil for nil node, got %v", info)
	}
}

func TestIdentifySymbolAtPosition_UnsupportedNode(t *testing.T) {
	// BlockStatement is not a symbol
	block := &ast.BlockStatement{}

	info := IdentifySymbolAtPosition(block)

	if info != nil {
		t.Errorf("Expected nil for unsupported node type, got %v", info)
	}
}

func TestExtractSymbolName_Identifier(t *testing.T) {
	ident := &ast.Identifier{
		Value: "testName",
		Token: token.Token{Pos: token.Position{Line: 1, Column: 1}},
	}

	name := ExtractSymbolName(ident)

	if name != "testName" {
		t.Errorf("Expected 'testName', got '%s'", name)
	}
}

func TestExtractSymbolName_VarDecl(t *testing.T) {
	varDecl := &ast.VarDeclStatement{
		Names: []*ast.Identifier{
			{Value: "x", Token: token.Token{Pos: token.Position{Line: 1, Column: 5}}},
			{Value: "y", Token: token.Token{Pos: token.Position{Line: 1, Column: 8}}},
		},
	}

	name := ExtractSymbolName(varDecl)

	// Should return first name
	if name != "x" {
		t.Errorf("Expected 'x', got '%s'", name)
	}
}

func TestExtractSymbolName_Nil(t *testing.T) {
	name := ExtractSymbolName(nil)

	if name != "" {
		t.Errorf("Expected empty string for nil node, got '%s'", name)
	}
}

func TestIsDeclaration_VarDecl(t *testing.T) {
	varDecl := &ast.VarDeclStatement{
		Names: []*ast.Identifier{
			{Value: "x", Token: token.Token{Pos: token.Position{Line: 1, Column: 5}}},
		},
	}

	if !IsDeclaration(varDecl) {
		t.Error("VarDeclStatement should be considered a declaration")
	}
}

func TestIsDeclaration_FunctionDecl(t *testing.T) {
	funcDecl := &ast.FunctionDecl{
		Name:       &ast.Identifier{Value: "Test"},
		Parameters: []*ast.Parameter{},
		Body:       &ast.BlockStatement{},
	}

	if !IsDeclaration(funcDecl) {
		t.Error("FunctionDecl should be considered a declaration")
	}
}

func TestIsDeclaration_Identifier(t *testing.T) {
	ident := &ast.Identifier{
		Value: "x",
		Token: token.Token{Pos: token.Position{Line: 5, Column: 10}},
	}

	if IsDeclaration(ident) {
		t.Error("Identifier should not be considered a declaration")
	}
}

func TestIsDeclaration_Nil(t *testing.T) {
	if IsDeclaration(nil) {
		t.Error("Nil should not be considered a declaration")
	}
}

func TestSymbolInfo_String(t *testing.T) {
	info := &SymbolInfo{
		Name: "testVar",
		Kind: SymbolKindVariable,
	}

	str := info.String()

	if str != "Symbol{Name: testVar, Kind: variable}" {
		t.Errorf("Unexpected string representation: %s", str)
	}
}

func TestSymbolInfo_String_Nil(t *testing.T) {
	var info *SymbolInfo = nil

	str := info.String()

	if str != "<nil>" {
		t.Errorf("Expected '<nil>', got '%s'", str)
	}
}

func TestGetSymbolContext_MemberExpression(t *testing.T) {
	memberExpr := &ast.MemberAccessExpression{
		Object: &ast.Identifier{Value: "obj"},
		Member: &ast.Identifier{Value: "field"},
	}

	context := GetSymbolContext(memberExpr, &ast.Program{})

	if context != "member" {
		t.Errorf("Expected 'member' context, got '%s'", context)
	}
}

func TestGetSymbolContext_CallExpression(t *testing.T) {
	callExpr := &ast.CallExpression{
		Function:  &ast.Identifier{Value: "func"},
		Arguments: []ast.Expression{},
	}

	context := GetSymbolContext(callExpr, &ast.Program{})

	if context != "call" {
		t.Errorf("Expected 'call' context, got '%s'", context)
	}
}

func TestGetSymbolContext_Reference(t *testing.T) {
	ident := &ast.Identifier{Value: "x"}

	context := GetSymbolContext(ident, &ast.Program{})

	if context != "reference" {
		t.Errorf("Expected 'reference' context, got '%s'", context)
	}
}

func TestGetSymbolContext_Nil(t *testing.T) {
	context := GetSymbolContext(nil, nil)

	if context != "unknown" {
		t.Errorf("Expected 'unknown' context for nil, got '%s'", context)
	}
}
