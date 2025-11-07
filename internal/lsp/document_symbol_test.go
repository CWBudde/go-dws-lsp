package lsp

import (
	"testing"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// TestDocumentSymbol_Functions tests that function declarations are returned as symbols
func TestDocumentSymbol_Functions(t *testing.T) {
	srv := server.New()
	SetServer(srv)

	uri := "file:///test_functions.dws"
	content := `function Add(x, y: Integer): Integer;
begin
  Result := x + y;
end;

procedure PrintMessage(msg: String);
begin
  PrintLn(msg);
end;`

	// Open document
	err := DidOpen(&glsp.Context{}, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  uri,
			Text: content,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request document symbols
	result, err := DocumentSymbol(&glsp.Context{}, &protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	})
	if err != nil {
		t.Fatalf("DocumentSymbol failed: %v", err)
	}

	symbols, ok := result.([]protocol.DocumentSymbol)
	if !ok {
		t.Fatalf("Expected []protocol.DocumentSymbol, got %T", result)
	}

	// Should have 2 symbols: Add and PrintMessage
	if len(symbols) != 2 {
		t.Fatalf("Expected 2 symbols, got %d", len(symbols))
	}

	// Check first symbol (Add)
	if symbols[0].Name != "Add" {
		t.Errorf("Expected first symbol name 'Add', got '%s'", symbols[0].Name)
	}
	if symbols[0].Kind != protocol.SymbolKindFunction {
		t.Errorf("Expected symbol kind Function, got %v", symbols[0].Kind)
	}
	if symbols[0].Detail == nil || *symbols[0].Detail == "" {
		t.Errorf("Expected non-empty detail for function Add")
	}

	// Check second symbol (PrintMessage)
	if symbols[1].Name != "PrintMessage" {
		t.Errorf("Expected second symbol name 'PrintMessage', got '%s'", symbols[1].Name)
	}
	if symbols[1].Kind != protocol.SymbolKindFunction {
		t.Errorf("Expected symbol kind Function, got %v", symbols[1].Kind)
	}
}

// TestDocumentSymbol_Variables tests that variable declarations are returned as symbols
func TestDocumentSymbol_Variables(t *testing.T) {
	srv := server.New()
	SetServer(srv)

	uri := "file:///test_variables.dws"
	content := `var x: Integer := 10;
var y, z: String;`

	err := DidOpen(&glsp.Context{}, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  uri,
			Text: content,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	result, err := DocumentSymbol(&glsp.Context{}, &protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	})
	if err != nil {
		t.Fatalf("DocumentSymbol failed: %v", err)
	}

	symbols, ok := result.([]protocol.DocumentSymbol)
	if !ok {
		t.Fatalf("Expected []protocol.DocumentSymbol, got %T", result)
	}

	// Should have 3 symbols: x, y, z
	if len(symbols) != 3 {
		t.Fatalf("Expected 3 variable symbols, got %d", len(symbols))
	}

	// Check symbol names
	expectedNames := []string{"x", "y", "z"}
	for i, expected := range expectedNames {
		if symbols[i].Name != expected {
			t.Errorf("Expected symbol %d name '%s', got '%s'", i, expected, symbols[i].Name)
		}
		if symbols[i].Kind != protocol.SymbolKindVariable {
			t.Errorf("Expected symbol %d kind Variable, got %v", i, symbols[i].Kind)
		}
	}
}

// TestDocumentSymbol_Constants tests that constant declarations are returned as symbols
func TestDocumentSymbol_Constants(t *testing.T) {
	srv := server.New()
	SetServer(srv)

	uri := "file:///test_constants.dws"
	content := `const PI: Float = 3.14159;
const MAX_SIZE = 100;`

	err := DidOpen(&glsp.Context{}, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  uri,
			Text: content,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	result, err := DocumentSymbol(&glsp.Context{}, &protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	})
	if err != nil {
		t.Fatalf("DocumentSymbol failed: %v", err)
	}

	symbols, ok := result.([]protocol.DocumentSymbol)
	if !ok {
		t.Fatalf("Expected []protocol.DocumentSymbol, got %T", result)
	}

	// Should have 2 symbols: PI and MAX_SIZE
	if len(symbols) != 2 {
		t.Fatalf("Expected 2 constant symbols, got %d", len(symbols))
	}

	// Check first symbol (PI)
	if symbols[0].Name != "PI" {
		t.Errorf("Expected first symbol name 'PI', got '%s'", symbols[0].Name)
	}
	if symbols[0].Kind != protocol.SymbolKindConstant {
		t.Errorf("Expected symbol kind Constant, got %v", symbols[0].Kind)
	}

	// Check second symbol (MAX_SIZE)
	if symbols[1].Name != "MAX_SIZE" {
		t.Errorf("Expected second symbol name 'MAX_SIZE', got '%s'", symbols[1].Name)
	}
	if symbols[1].Kind != protocol.SymbolKindConstant {
		t.Errorf("Expected symbol kind Constant, got %v", symbols[1].Kind)
	}
}

// TestDocumentSymbol_Class tests that class declarations with members are returned as hierarchical symbols
func TestDocumentSymbol_Class(t *testing.T) {
	srv := server.New()
	SetServer(srv)

	uri := "file:///test_class.dws"
	content := `type TPerson = class
  FName: String;
  FAge: Integer;

  function GetName: String;
  begin
    Result := FName;
  end;

  procedure SetAge(value: Integer);
  begin
    FAge := value;
  end;
end;`

	err := DidOpen(&glsp.Context{}, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  uri,
			Text: content,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	result, err := DocumentSymbol(&glsp.Context{}, &protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	})
	if err != nil {
		t.Fatalf("DocumentSymbol failed: %v", err)
	}

	symbols, ok := result.([]protocol.DocumentSymbol)
	if !ok {
		t.Fatalf("Expected []protocol.DocumentSymbol, got %T", result)
	}

	// Should have 1 top-level symbol: TPerson
	if len(symbols) != 1 {
		t.Fatalf("Expected 1 class symbol, got %d", len(symbols))
	}

	classSymbol := symbols[0]
	if classSymbol.Name != "TPerson" {
		t.Errorf("Expected class name 'TPerson', got '%s'", classSymbol.Name)
	}
	if classSymbol.Kind != protocol.SymbolKindClass {
		t.Errorf("Expected symbol kind Class, got %v", classSymbol.Kind)
	}

	// Check children (should have 4: FName, FAge, GetName, SetAge)
	if len(classSymbol.Children) != 4 {
		t.Fatalf("Expected 4 class members, got %d", len(classSymbol.Children))
	}

	// Check field FName
	if classSymbol.Children[0].Name != "FName" {
		t.Errorf("Expected first child name 'FName', got '%s'", classSymbol.Children[0].Name)
	}
	if classSymbol.Children[0].Kind != protocol.SymbolKindField {
		t.Errorf("Expected first child kind Field, got %v", classSymbol.Children[0].Kind)
	}

	// Check field FAge
	if classSymbol.Children[1].Name != "FAge" {
		t.Errorf("Expected second child name 'FAge', got '%s'", classSymbol.Children[1].Name)
	}
	if classSymbol.Children[1].Kind != protocol.SymbolKindField {
		t.Errorf("Expected second child kind Field, got %v", classSymbol.Children[1].Kind)
	}

	// Check method GetName
	if classSymbol.Children[2].Name != "GetName" {
		t.Errorf("Expected third child name 'GetName', got '%s'", classSymbol.Children[2].Name)
	}
	if classSymbol.Children[2].Kind != protocol.SymbolKindMethod {
		t.Errorf("Expected third child kind Method, got %v", classSymbol.Children[2].Kind)
	}

	// Check method SetAge
	if classSymbol.Children[3].Name != "SetAge" {
		t.Errorf("Expected fourth child name 'SetAge', got '%s'", classSymbol.Children[3].Name)
	}
	if classSymbol.Children[3].Kind != protocol.SymbolKindMethod {
		t.Errorf("Expected fourth child kind Method, got %v", classSymbol.Children[3].Kind)
	}
}

// TestDocumentSymbol_Record tests that record declarations are returned as symbols
func TestDocumentSymbol_Record(t *testing.T) {
	srv := server.New()
	SetServer(srv)

	uri := "file:///test_record.dws"
	content := `type TPoint = record
  X: Integer;
  Y: Integer;
end;`

	err := DidOpen(&glsp.Context{}, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  uri,
			Text: content,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	result, err := DocumentSymbol(&glsp.Context{}, &protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	})
	if err != nil {
		t.Fatalf("DocumentSymbol failed: %v", err)
	}

	symbols, ok := result.([]protocol.DocumentSymbol)
	if !ok {
		t.Fatalf("Expected []protocol.DocumentSymbol, got %T", result)
	}

	// Should have 1 symbol: TPoint
	if len(symbols) != 1 {
		t.Fatalf("Expected 1 record symbol, got %d", len(symbols))
	}

	recordSymbol := symbols[0]
	if recordSymbol.Name != "TPoint" {
		t.Errorf("Expected record name 'TPoint', got '%s'", recordSymbol.Name)
	}
	if recordSymbol.Kind != protocol.SymbolKindStruct {
		t.Errorf("Expected symbol kind Struct, got %v", recordSymbol.Kind)
	}

	// Note: Fields in records might not be parsed as children depending on AST structure
	// This is a basic test to ensure the record itself is detected
}

// TestDocumentSymbol_Enum tests that enum declarations are returned as symbols
func TestDocumentSymbol_Enum(t *testing.T) {
	srv := server.New()
	SetServer(srv)

	uri := "file:///test_enum.dws"
	content := `type TColor = (Red, Green, Blue);`

	err := DidOpen(&glsp.Context{}, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  uri,
			Text: content,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	result, err := DocumentSymbol(&glsp.Context{}, &protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	})
	if err != nil {
		t.Fatalf("DocumentSymbol failed: %v", err)
	}

	symbols, ok := result.([]protocol.DocumentSymbol)
	if !ok {
		t.Fatalf("Expected []protocol.DocumentSymbol, got %T", result)
	}

	// Should have 1 symbol: TColor
	if len(symbols) != 1 {
		t.Fatalf("Expected 1 enum symbol, got %d", len(symbols))
	}

	enumSymbol := symbols[0]
	if enumSymbol.Name != "TColor" {
		t.Errorf("Expected enum name 'TColor', got '%s'", enumSymbol.Name)
	}
	if enumSymbol.Kind != protocol.SymbolKindEnum {
		t.Errorf("Expected symbol kind Enum, got %v", enumSymbol.Kind)
	}

	// Check enum values as children
	if len(enumSymbol.Children) != 3 {
		t.Fatalf("Expected 3 enum members, got %d", len(enumSymbol.Children))
	}

	expectedMembers := []string{"Red", "Green", "Blue"}
	for i, expected := range expectedMembers {
		if enumSymbol.Children[i].Name != expected {
			t.Errorf("Expected enum member %d name '%s', got '%s'", i, expected, enumSymbol.Children[i].Name)
		}
		if enumSymbol.Children[i].Kind != protocol.SymbolKindEnumMember {
			t.Errorf("Expected enum member %d kind EnumMember, got %v", i, enumSymbol.Children[i].Kind)
		}
	}
}

// TestDocumentSymbol_Mixed tests a document with multiple symbol types
func TestDocumentSymbol_Mixed(t *testing.T) {
	srv := server.New()
	SetServer(srv)

	uri := "file:///test_mixed.dws"
	content := `const MAX = 100;
var counter: Integer;

function Increment: Integer;
begin
  counter := counter + 1;
  Result := counter;
end;

type TStatus = (Ready, Running, Stopped);`

	err := DidOpen(&glsp.Context{}, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  uri,
			Text: content,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	result, err := DocumentSymbol(&glsp.Context{}, &protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	})
	if err != nil {
		t.Fatalf("DocumentSymbol failed: %v", err)
	}

	symbols, ok := result.([]protocol.DocumentSymbol)
	if !ok {
		t.Fatalf("Expected []protocol.DocumentSymbol, got %T", result)
	}

	// Should have 4 symbols: MAX (const), counter (var), Increment (function), TStatus (enum)
	if len(symbols) != 4 {
		t.Fatalf("Expected 4 symbols, got %d", len(symbols))
	}

	// Verify symbol types
	expectedTypes := []protocol.SymbolKind{
		protocol.SymbolKindConstant,
		protocol.SymbolKindVariable,
		protocol.SymbolKindFunction,
		protocol.SymbolKindEnum,
	}

	for i, expected := range expectedTypes {
		if symbols[i].Kind != expected {
			t.Errorf("Expected symbol %d kind %v, got %v", i, expected, symbols[i].Kind)
		}
	}
}

// TestDocumentSymbol_EmptyDocument tests that an empty document returns no symbols
func TestDocumentSymbol_EmptyDocument(t *testing.T) {
	srv := server.New()
	SetServer(srv)

	uri := "file:///test_empty.dws"
	content := ``

	err := DidOpen(&glsp.Context{}, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  uri,
			Text: content,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	result, err := DocumentSymbol(&glsp.Context{}, &protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	})
	if err != nil {
		t.Fatalf("DocumentSymbol failed: %v", err)
	}

	symbols, ok := result.([]protocol.DocumentSymbol)
	if !ok {
		t.Fatalf("Expected []protocol.DocumentSymbol, got %T", result)
	}

	// Should have no symbols
	if len(symbols) != 0 {
		t.Errorf("Expected 0 symbols for empty document, got %d", len(symbols))
	}
}

// TestDocumentSymbol_DocumentNotFound tests handling of non-existent document
func TestDocumentSymbol_DocumentNotFound(t *testing.T) {
	srv := server.New()
	SetServer(srv)

	uri := "file:///nonexistent.dws"

	result, err := DocumentSymbol(&glsp.Context{}, &protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	})
	if err != nil {
		t.Fatalf("DocumentSymbol failed: %v", err)
	}

	// Should return nil for non-existent document
	if result != nil {
		t.Errorf("Expected nil for non-existent document, got %v", result)
	}
}
