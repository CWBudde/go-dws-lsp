package analysis

import (
	"testing"

	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/cwbudde/go-dws/pkg/dwscript"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestCollectScopeCompletions_Keywords(t *testing.T) {
	source := `
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

	// Get completions at line 1 (inside begin/end block)
	items, err := CollectScopeCompletions(doc, nil, 1, 0)
	if err != nil {
		t.Fatalf("CollectScopeCompletions returned error: %v", err)
	}

	// Should include keywords
	hasKeyword := false

	for _, item := range items {
		if item.Label == "if" || item.Label == "while" || item.Label == "for" {
			if *item.Kind == protocol.CompletionItemKindKeyword {
				hasKeyword = true
				break
			}
		}
	}

	if !hasKeyword {
		t.Error("Expected to find keyword completions")
	}
}

func TestCollectScopeCompletions_LocalVariables(t *testing.T) {
	source := `
function TestFunc(): Integer;
var x, y: Integer;
var s: String;
begin
  // cursor here
  Result := x + y;
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

	// Get completions inside the function (line 5, where comment is)
	items, err := CollectScopeCompletions(doc, nil, 5, 2)
	if err != nil {
		t.Fatalf("CollectScopeCompletions returned error: %v", err)
	}

	// Should include local variables x, y, s
	foundX := false
	foundY := false
	foundS := false

	for _, item := range items {
		if item.Label == "x" {
			foundX = true

			if *item.Kind != protocol.CompletionItemKindVariable {
				t.Errorf("Expected 'x' to be a Variable, got kind %d", *item.Kind)
			}
		}

		if item.Label == "y" {
			foundY = true
		}

		if item.Label == "s" {
			foundS = true
		}
	}

	if !foundX || !foundY || !foundS {
		t.Errorf("Expected to find local variables x, y, s. Found: x=%v, y=%v, s=%v", foundX, foundY, foundS)
	}
}

func TestCollectScopeCompletions_Parameters(t *testing.T) {
	source := `
function Add(a, b: Integer): Integer;
begin
  // cursor here
  Result := a + b;
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

	// Get completions inside the function
	items, err := CollectScopeCompletions(doc, nil, 3, 2)
	if err != nil {
		t.Fatalf("CollectScopeCompletions returned error: %v", err)
	}

	// Should include parameters a and b
	foundA := false
	foundB := false

	for _, item := range items {
		if item.Label == "a" {
			foundA = true

			if *item.Kind != protocol.CompletionItemKindVariable {
				t.Errorf("Expected 'a' to be a Variable, got kind %d", *item.Kind)
			}
		}

		if item.Label == "b" {
			foundB = true
		}
	}

	if !foundA || !foundB {
		t.Errorf("Expected to find parameters a and b. Found: a=%v, b=%v", foundA, foundB)
	}
}

func TestCollectScopeCompletions_GlobalFunctions(t *testing.T) {
	source := `
function MyGlobalFunc(): Integer;
begin
  Result := 42;
end;

procedure MyGlobalProc;
begin
end;

begin
  // cursor here
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

	// Get completions in main begin/end block
	items, err := CollectScopeCompletions(doc, nil, 11, 2)
	if err != nil {
		t.Fatalf("CollectScopeCompletions returned error: %v", err)
	}

	// Should include global functions
	foundFunc := false
	foundProc := false

	for _, item := range items {
		if item.Label == "MyGlobalFunc" {
			foundFunc = true

			if *item.Kind != protocol.CompletionItemKindFunction {
				t.Errorf("Expected 'MyGlobalFunc' to be a Function, got kind %d", *item.Kind)
			}
		}

		if item.Label == "MyGlobalProc" {
			foundProc = true
		}
	}

	if !foundFunc || !foundProc {
		t.Errorf("Expected to find global functions. Found: MyGlobalFunc=%v, MyGlobalProc=%v", foundFunc, foundProc)
	}
}

func TestCollectScopeCompletions_GlobalTypes(t *testing.T) {
	source := `
type TMyClass = class
  FValue: Integer;
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

	// Get completions
	items, err := CollectScopeCompletions(doc, nil, 6, 0)
	if err != nil {
		t.Fatalf("CollectScopeCompletions returned error: %v", err)
	}

	// Should include global type
	foundClass := false

	for _, item := range items {
		if item.Label == "TMyClass" {
			foundClass = true

			if *item.Kind != protocol.CompletionItemKindClass {
				t.Errorf("Expected 'TMyClass' to be a Class, got kind %d", *item.Kind)
			}

			break
		}
	}

	if !foundClass {
		t.Error("Expected to find global type TMyClass")
	}
}

func TestCollectScopeCompletions_BuiltInFunctions(t *testing.T) {
	source := `
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

	// Get completions
	items, err := CollectScopeCompletions(doc, nil, 1, 0)
	if err != nil {
		t.Fatalf("CollectScopeCompletions returned error: %v", err)
	}

	// Should include built-in functions
	foundPrintLn := false
	foundLength := false
	foundIntToStr := false

	for _, item := range items {
		if item.Label == "PrintLn" {
			foundPrintLn = true

			if *item.Kind != protocol.CompletionItemKindFunction {
				t.Errorf("Expected 'PrintLn' to be a Function, got kind %d", *item.Kind)
			}
		}

		if item.Label == "Length" {
			foundLength = true
		}

		if item.Label == "IntToStr" {
			foundIntToStr = true
		}
	}

	if !foundPrintLn || !foundLength || !foundIntToStr {
		t.Errorf("Expected to find built-in functions. Found: PrintLn=%v, Length=%v, IntToStr=%v",
			foundPrintLn, foundLength, foundIntToStr)
	}
}

func TestCollectScopeCompletions_BuiltInTypes(t *testing.T) {
	source := `
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

	// Get completions
	items, err := CollectScopeCompletions(doc, nil, 1, 0)
	if err != nil {
		t.Fatalf("CollectScopeCompletions returned error: %v", err)
	}

	// Should include built-in types
	foundInteger := false
	foundString := false
	foundBoolean := false

	for _, item := range items {
		if item.Label == "Integer" {
			foundInteger = true
		}

		if item.Label == "String" {
			foundString = true
		}

		if item.Label == "Boolean" {
			foundBoolean = true
		}
	}

	if !foundInteger || !foundString || !foundBoolean {
		t.Errorf("Expected to find built-in types. Found: Integer=%v, String=%v, Boolean=%v",
			foundInteger, foundString, foundBoolean)
	}
}

func TestFilterCompletionsByPrefix(t *testing.T) {
	items := []protocol.CompletionItem{
		{Label: "Apple"},
		{Label: "Banana"},
		{Label: "Apricot"},
		{Label: "Cherry"},
		{Label: "Avocado"},
	}

	// Filter with prefix "ap"
	filtered := FilterCompletionsByPrefix(items, "ap")

	if len(filtered) != 2 {
		t.Errorf("Expected 2 items with prefix 'ap', got %d", len(filtered))
	}

	// Check that we have Apple and Apricot (case-insensitive)
	foundApple := false
	foundApricot := false

	for _, item := range filtered {
		if item.Label == "Apple" {
			foundApple = true
		}

		if item.Label == "Apricot" {
			foundApricot = true
		}
	}

	if !foundApple || !foundApricot {
		t.Errorf("Expected to find Apple and Apricot. Found: Apple=%v, Apricot=%v", foundApple, foundApricot)
	}

	// Filter with empty prefix should return all items
	filteredEmpty := FilterCompletionsByPrefix(items, "")
	if len(filteredEmpty) != len(items) {
		t.Errorf("Expected %d items with empty prefix, got %d", len(items), len(filteredEmpty))
	}
}

func TestCollectScopeCompletions_NoAST(t *testing.T) {
	// Document without AST
	doc := &server.Document{
		URI:     "file:///test.dws",
		Text:    "begin end.",
		Program: nil,
	}

	// Should still return keywords
	items, err := CollectScopeCompletions(doc, nil, 0, 0)
	if err != nil {
		t.Fatalf("CollectScopeCompletions returned error: %v", err)
	}

	// Should have at least some keywords
	if len(items) == 0 {
		t.Error("Expected to get keyword completions even without AST")
	}

	// Verify we have keywords
	hasKeyword := false

	for _, item := range items {
		if *item.Kind == protocol.CompletionItemKindKeyword {
			hasKeyword = true
			break
		}
	}

	if !hasKeyword {
		t.Error("Expected to find keywords even without AST")
	}
}
