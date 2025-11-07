package analysis

import (
	"testing"

	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/cwbudde/go-dws/pkg/dwscript"
)

func TestResolveMemberType_LocalVariable(t *testing.T) {
	source := `
var x: Integer;
begin
  x := 5;
end.`

	// Parse the source
	engine, err := dwscript.New()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	// Create a document
	doc := &server.Document{
		URI:     "file:///test.dws",
		Text:    source,
		Program: program,
	}

	// Resolve type of 'x' at the position where the cursor would be
	// Line 3 (0-based), character 2 (after 'x')
	typeInfo, err := ResolveMemberType(doc, "x", 3, 2)

	if err != nil {
		t.Fatalf("ResolveMemberType returned error: %v", err)
	}

	if typeInfo == nil {
		t.Fatal("Expected typeInfo to be non-nil, got nil")
	}

	if typeInfo.TypeName != "Integer" {
		t.Errorf("Expected type 'Integer', got '%s'", typeInfo.TypeName)
	}

	if !typeInfo.IsBuiltIn {
		t.Error("Expected IsBuiltIn to be true for Integer")
	}
}

func TestResolveMemberType_Parameter(t *testing.T) {
	source := `
function Test(s: String);
begin
  PrintLn(s);
end;`

	engine, err := dwscript.New()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	doc := &server.Document{
		URI:     "file:///test.dws",
		Text:    source,
		Program: program,
	}

	// Resolve type of parameter 's'
	typeInfo, err := ResolveMemberType(doc, "s", 3, 2)

	if err != nil {
		t.Fatalf("ResolveMemberType returned error: %v", err)
	}

	if typeInfo == nil {
		t.Fatal("Expected typeInfo to be non-nil, got nil")
	}

	if typeInfo.TypeName != "String" {
		t.Errorf("Expected type 'String', got '%s'", typeInfo.TypeName)
	}

	if !typeInfo.IsBuiltIn {
		t.Error("Expected IsBuiltIn to be true for String")
	}
}

func TestResolveMemberType_ClassField(t *testing.T) {
	source := `
type
  TMyClass = class
    FValue: Float;
    procedure Test;
  end;

procedure TMyClass.Test;
begin
  FValue := 3.14;
end;`

	engine, err := dwscript.New()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	doc := &server.Document{
		URI:     "file:///test.dws",
		Text:    source,
		Program: program,
	}

	// Resolve type of field 'FValue'
	typeInfo, err := ResolveMemberType(doc, "FValue", 9, 2)

	if err != nil {
		t.Fatalf("ResolveMemberType returned error: %v", err)
	}

	if typeInfo == nil {
		t.Fatal("Expected typeInfo to be non-nil, got nil")
	}

	if typeInfo.TypeName != "Float" {
		t.Errorf("Expected type 'Float', got '%s'", typeInfo.TypeName)
	}

	if !typeInfo.IsBuiltIn {
		t.Error("Expected IsBuiltIn to be true for Float")
	}
}

func TestResolveMemberType_UserDefinedType(t *testing.T) {
	source := `
type
  TPoint = record
    X, Y: Integer;
  end;

var p: TPoint;
begin
  p.X := 10;
end.`

	engine, err := dwscript.New()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	doc := &server.Document{
		URI:     "file:///test.dws",
		Text:    source,
		Program: program,
	}

	// Resolve type of 'p'
	typeInfo, err := ResolveMemberType(doc, "p", 8, 2)

	if err != nil {
		t.Fatalf("ResolveMemberType returned error: %v", err)
	}

	if typeInfo == nil {
		t.Fatal("Expected typeInfo to be non-nil, got nil")
	}

	if typeInfo.TypeName != "TPoint" {
		t.Errorf("Expected type 'TPoint', got '%s'", typeInfo.TypeName)
	}

	if typeInfo.IsBuiltIn {
		t.Error("Expected IsBuiltIn to be false for user-defined type TPoint")
	}
}

func TestResolveMemberType_UnknownIdentifier(t *testing.T) {
	source := `
var x: Integer;
begin
  x := 10;
end.`

	engine, err := dwscript.New()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	doc := &server.Document{
		URI:     "file:///test.dws",
		Text:    source,
		Program: program,
	}

	// Try to resolve a non-existent identifier
	typeInfo, err := ResolveMemberType(doc, "unknown", 3, 2)

	if err != nil {
		t.Fatalf("ResolveMemberType returned error: %v", err)
	}

	// Should return nil for unknown identifiers
	if typeInfo != nil {
		t.Errorf("Expected typeInfo to be nil for unknown identifier, got type '%s'", typeInfo.TypeName)
	}
}

func TestIsBuiltInType(t *testing.T) {
	tests := []struct {
		typeName  string
		isBuiltIn bool
	}{
		{"Integer", true},
		{"String", true},
		{"Float", true},
		{"Boolean", true},
		{"TObject", true},
		{"DateTime", true},
		{"TMyClass", false},
		{"TPoint", false},
		{"CustomType", false},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			result := isBuiltInType(tt.typeName)
			if result != tt.isBuiltIn {
				t.Errorf("isBuiltInType(%s) = %v, want %v", tt.typeName, result, tt.isBuiltIn)
			}
		})
	}
}
