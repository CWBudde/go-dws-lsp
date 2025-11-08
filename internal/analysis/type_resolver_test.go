package analysis

import (
	"testing"

	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/cwbudde/go-dws/pkg/dwscript"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

const testSimpleVarAssignmentCode = `
var x: Integer;
begin
  x := 10;
end.`

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
	typeInfo := ResolveMemberType(doc, "x", 3, 2)

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
	typeInfo := ResolveMemberType(doc, "s", 3, 2)

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
	typeInfo := ResolveMemberType(doc, "FValue", 9, 2)

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
	typeInfo := ResolveMemberType(doc, "p", 8, 2)

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
	source := testSimpleVarAssignmentCode

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
	typeInfo := ResolveMemberType(doc, "unknown", 3, 2)

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

func TestGetTypeMembers_Class(t *testing.T) {
	source := `
type
  TMyClass = class
    FValue: Integer;
    FName: String;
    property Value: Integer read FValue write FValue;
  end;

begin
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

	// Get members of TMyClass
	members, err := GetTypeMembers(doc, "TMyClass")
	if err != nil {
		t.Fatalf("GetTypeMembers returned error: %v", err)
	}

	if len(members) == 0 {
		t.Fatal("Expected members for TMyClass, got none")
	}

	// Check that we have fields and properties
	hasField := false
	hasProperty := false

	for _, member := range members {
		t.Logf("Found member: %s (kind: %d)", member.Label, *member.Kind)

		switch member.Label {
		case "FValue", "FName":
			hasField = true

			if *member.Kind != protocol.CompletionItemKindField {
				t.Errorf("Expected %s to be a Field, got kind %d", member.Label, *member.Kind)
			}
		case "GetValue", "SetValue":
			// These are implemented outside the class, so they won't appear as methods
			if *member.Kind != protocol.CompletionItemKindMethod {
				t.Errorf("Expected %s to be a Method, got kind %d", member.Label, *member.Kind)
			}
		case "Value":
			hasProperty = true

			if *member.Kind != protocol.CompletionItemKindProperty {
				t.Errorf("Expected %s to be a Property, got kind %d", member.Label, *member.Kind)
			}
		}
	}

	if !hasField {
		t.Error("Expected to find field members")
	}

	if !hasProperty {
		t.Error("Expected to find property members")
	}
}

func TestGetTypeMembers_Record(t *testing.T) {
	source := `
type
  TPoint = record
    X, Y: Integer;
    Z: Float;
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

	// Get members of TPoint
	members, err := GetTypeMembers(doc, "TPoint")
	if err != nil {
		t.Fatalf("GetTypeMembers returned error: %v", err)
	}

	if len(members) < 3 {
		t.Errorf("Expected at least 3 members for TPoint, got %d", len(members))
	}

	// Check that all are fields
	for _, member := range members {
		if *member.Kind != protocol.CompletionItemKindField {
			t.Errorf("Expected %s to be a Field, got kind %d", member.Label, *member.Kind)
		}
	}

	// Check for specific fields
	hasX := false
	hasY := false
	hasZ := false

	for _, member := range members {
		switch member.Label {
		case "X":
			hasX = true
		case "Y":
			hasY = true
		case "Z":
			hasZ = true
		}
	}

	if !hasX || !hasY || !hasZ {
		t.Errorf("Expected to find X, Y, and Z fields. Found: hasX=%v, hasY=%v, hasZ=%v", hasX, hasY, hasZ)
	}
}

func TestGetTypeMembers_BuiltInType(t *testing.T) {
	source := testSimpleVarAssignmentCode

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

	// Get members of Integer (built-in type)
	members, err := GetTypeMembers(doc, "Integer")
	if err != nil {
		t.Fatalf("GetTypeMembers returned error: %v", err)
	}

	// Built-in types currently return no members
	if len(members) != 0 {
		t.Errorf("Expected no members for built-in type Integer, got %d", len(members))
	}
}

func TestGetTypeMembers_UnknownType(t *testing.T) {
	source := testSimpleVarAssignmentCode

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

	// Get members of an unknown type
	members, err := GetTypeMembers(doc, "TUnknownType")
	if err != nil {
		t.Fatalf("GetTypeMembers returned error: %v", err)
	}

	// Unknown types should return no members
	if len(members) != 0 {
		t.Errorf("Expected no members for unknown type, got %d", len(members))
	}
}

func TestGetTypeMembers_Sorting(t *testing.T) {
	source := `
type
  TMyClass = class
    ZField: Integer;
    AField: String;
    MField: Float;
  end;

var obj: TMyClass;
begin
  obj.AField := 'test';
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

	// Get members of TMyClass
	members, err := GetTypeMembers(doc, "TMyClass")
	if err != nil {
		t.Fatalf("GetTypeMembers returned error: %v", err)
	}

	if len(members) < 3 {
		t.Errorf("Expected at least 3 members, got %d", len(members))
	}

	// Check that members are sorted alphabetically
	// Expected order: AField, MField, ZField
	if members[0].Label != "AField" {
		t.Errorf("Expected first member to be 'AField', got '%s'", members[0].Label)
	}

	if members[1].Label != "MField" {
		t.Errorf("Expected second member to be 'MField', got '%s'", members[1].Label)
	}

	if members[2].Label != "ZField" {
		t.Errorf("Expected third member to be 'ZField', got '%s'", members[2].Label)
	}
}
