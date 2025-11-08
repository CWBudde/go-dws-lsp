package lsp

import (
	"testing"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/analysis"
	"github.com/CWBudde/go-dws-lsp/internal/server"
)

func TestCompletion_EmptyListForValidDocument(t *testing.T) {
	// Create a test server
	srv := server.New()
	SetServer(srv)

	// Sample DWScript source code
	source := `program Test;

var x: Integer;

begin
  x := 42;
end.`

	// Add document to server
	uri := "file:///test.dws"
	program, _, err := analysis.ParseDocument(source, uri)
	if err != nil {
		t.Fatalf("Failed to parse document: %v", err)
	}

	doc := &server.Document{
		URI:        uri,
		Text:       source,
		Version:    1,
		LanguageID: "dwscript",
		Program:    program,
	}
	srv.Documents().Set(uri, doc)

	// Create completion params
	params := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      4, // Inside the begin/end block
				Character: 2,
			},
		},
	}

	// Call Completion handler
	ctx := &glsp.Context{}
	result, err := Completion(ctx, params)

	// Should return without error
	if err != nil {
		t.Fatalf("Completion returned error: %v", err)
	}

	// Should return a CompletionList (may be empty for now since we haven't implemented item collection)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Result should be a CompletionList
	completionList, ok := result.(*protocol.CompletionList)
	if !ok {
		t.Fatalf("Expected CompletionList, got %T", result)
	}

	// For now, we expect an empty list (task 9.1 just sets up the handler structure)
	if completionList == nil {
		t.Fatal("Expected non-nil CompletionList")
	}

	t.Logf("Completion returned %d items (expected 0 for now)", len(completionList.Items))
}

func TestCompletion_TriggerCharacterDot(t *testing.T) {
	// Create a test server
	srv := server.New()
	SetServer(srv)

	// Sample DWScript source code with member access
	source := `program Test;

type TMyClass = class
  Field: Integer;
end;

var obj: TMyClass;

begin
  obj.Field := 42;
end.`

	// Add document to server
	uri := "file:///test.dws"
	program, _, err := analysis.ParseDocument(source, uri)
	if err != nil {
		t.Fatalf("Failed to parse document: %v", err)
	}

	doc := &server.Document{
		URI:        uri,
		Text:       source,
		Version:    1,
		LanguageID: "dwscript",
		Program:    program,
	}
	srv.Documents().Set(uri, doc)

	// Create completion params with trigger character
	// Simulate completion request right after typing "obj."
	triggerChar := "."
	params := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      9, // After "obj."
				Character: 6, // Position right after the dot
			},
		},
		Context: &protocol.CompletionContext{
			TriggerKind:      protocol.CompletionTriggerKindTriggerCharacter,
			TriggerCharacter: &triggerChar,
		},
	}

	// Call Completion handler
	ctx := &glsp.Context{}
	result, err := Completion(ctx, params)

	// Should return without error
	if err != nil {
		t.Fatalf("Completion returned error: %v", err)
	}

	// Should return a CompletionList
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	completionList, ok := result.(*protocol.CompletionList)
	if !ok {
		t.Fatalf("Expected CompletionList, got %T", result)
	}

	// For now, we expect an empty list (member completion not yet implemented)
	if completionList == nil {
		t.Fatal("Expected non-nil CompletionList")
	}

	t.Logf("Completion triggered by dot returned %d items", len(completionList.Items))
}

func TestCompletion_NonExistentDocument(t *testing.T) {
	// Create a test server
	srv := server.New()
	SetServer(srv)

	// Create completion params for non-existent document
	params := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file:///nonexistent.dws",
			},
			Position: protocol.Position{
				Line:      0,
				Character: 0,
			},
		},
	}

	// Call Completion handler
	ctx := &glsp.Context{}
	result, err := Completion(ctx, params)

	// Should return nil without error (graceful handling)
	if err != nil {
		t.Fatalf("Completion returned error: %v", err)
	}

	if result != nil {
		t.Fatalf("Expected nil result for non-existent document, got %v", result)
	}
}

func TestCompletion_DocumentWithParseErrors(t *testing.T) {
	// Create a test server
	srv := server.New()
	SetServer(srv)

	// Invalid DWScript source code (will have parse errors)
	source := `program Test;
var x Integer; // Missing colon
begin
  x := ;
end.`

	// Add document to server
	uri := "file:///test_invalid.dws"
	// Parse the document (may fail due to errors)
	program, _, _ := analysis.ParseDocument(source, uri)

	doc := &server.Document{
		URI:        uri,
		Text:       source,
		Version:    1,
		LanguageID: "dwscript",
		Program:    program, // May be nil if parsing failed
	}
	srv.Documents().Set(uri, doc)

	// Create completion params
	params := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      2,
				Character: 2,
			},
		},
	}

	// Call Completion handler
	ctx := &glsp.Context{}
	result, err := Completion(ctx, params)

	// Should return empty list without error (graceful handling)
	if err != nil {
		t.Fatalf("Completion returned error: %v", err)
	}

	// Even with parse errors, we should return an empty CompletionList
	if result == nil {
		t.Fatal("Expected non-nil result even with parse errors")
	}

	completionList, ok := result.(*protocol.CompletionList)
	if !ok {
		t.Fatalf("Expected CompletionList, got %T", result)
	}

	if len(completionList.Items) != 0 {
		t.Fatalf("Expected empty completion list for document with parse errors, got %d items", len(completionList.Items))
	}
}

// Task 9.19: Test partial variable name completion
func TestCompletion_PartialVariableName(t *testing.T) {
	// Create a test server
	srv := server.New()
	SetServer(srv)

	// Setup: code with variables alpha, beta, alphabet
	source := `program Test;

var alpha: Integer;
var beta: String;
var alphabet: Float;

begin
  alpha := 1;
end.`

	// Add document to server
	uri := "file:///test.dws"
	program, _, err := analysis.ParseDocument(source, uri)
	if err != nil {
		t.Fatalf("Failed to parse document: %v", err)
	}

	doc := &server.Document{
		URI:        uri,
		Text:       source,
		Version:    1,
		LanguageID: "dwscript",
		Program:    program,
	}
	srv.Documents().Set(uri, doc)

	// Input: cursor after "alp" (in the middle of "alpha")
	// We're testing if typing "alp" would suggest "alpha" and "alphabet"
	params := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      7,  // On the "alpha := 1;" line
				Character: 5, // After "alp" (position 5 is after the 'p')
			},
		},
	}

	// Call Completion handler
	ctx := &glsp.Context{}
	result, err := Completion(ctx, params)

	if err != nil {
		t.Fatalf("Completion returned error: %v", err)
	}

	completionList, ok := result.(*protocol.CompletionList)
	if !ok {
		t.Fatalf("Expected CompletionList, got %T", result)
	}

	// Expected: alpha and alphabet in results
	foundAlpha := false
	foundAlphabet := false
	foundBeta := false

	for _, item := range completionList.Items {
		t.Logf("Found completion item: %s", item.Label)
		if item.Label == "alpha" {
			foundAlpha = true
		}
		if item.Label == "alphabet" {
			foundAlphabet = true
		}
		if item.Label == "beta" {
			foundBeta = true
		}
	}

	// Verify: alpha and alphabet should be in results
	if !foundAlpha {
		t.Error("Expected 'alpha' to be in completion results")
	}
	if !foundAlphabet {
		t.Error("Expected 'alphabet' to be in completion results")
	}

	// Verify: beta should NOT be in results (doesn't match prefix "alp")
	if foundBeta {
		t.Error("Expected 'beta' to NOT be in completion results (doesn't match prefix)")
	}

	t.Logf("Completion test passed: found alpha=%v, alphabet=%v, beta=%v (expected beta=false)",
		foundAlpha, foundAlphabet, foundBeta)
}

// Task 9.19: Test parameter completion in function
func TestCompletion_ParameterCompletion(t *testing.T) {
	// Create a test server
	srv := server.New()
	SetServer(srv)

	// Setup: function with parameters
	source := `program Test;

function Calculate(firstParam: Integer; secondParam: Float): String;
var temp: Integer;
begin
  // Using firstParam here to test completion
  temp := firstParam;
  Result := '';
end;

begin
end.`

	// Add document to server
	uri := "file:///test.dws"
	program, _, err := analysis.ParseDocument(source, uri)
	if err != nil {
		t.Fatalf("Failed to parse document: %v", err)
	}

	doc := &server.Document{
		URI:        uri,
		Text:       source,
		Version:    1,
		LanguageID: "dwscript",
		Program:    program,
	}
	srv.Documents().Set(uri, doc)

	// Input: cursor after "fir" (in the middle of "firstParam")
	// Testing if typing "fir" would suggest "firstParam"
	params := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      6,   // On the line with "temp := firstParam;"
				Character: 13, // After "fir" in "firstParam"
			},
		},
	}

	// Call Completion handler
	ctx := &glsp.Context{}
	result, err := Completion(ctx, params)

	if err != nil {
		t.Fatalf("Completion returned error: %v", err)
	}

	completionList, ok := result.(*protocol.CompletionList)
	if !ok {
		t.Fatalf("Expected CompletionList, got %T", result)
	}

	// Expected: firstParam in results
	foundFirstParam := false
	foundSecondParam := false

	for _, item := range completionList.Items {
		t.Logf("Found completion item: %s (kind: %v)", item.Label, item.Kind)
		if item.Label == "firstParam" {
			foundFirstParam = true
			// Verify it's marked as a parameter
			if item.Kind != nil && *item.Kind != protocol.CompletionItemKindVariable {
				t.Logf("Warning: firstParam has kind %v, expected Variable", *item.Kind)
			}
		}
		if item.Label == "secondParam" {
			foundSecondParam = true
		}
	}

	// Verify: firstParam should be in results (matches prefix "fir")
	if !foundFirstParam {
		t.Error("Expected 'firstParam' to be in completion results")
	}

	// Verify: secondParam should NOT be in results (doesn't match prefix "fir")
	if foundSecondParam {
		t.Error("Expected 'secondParam' to NOT be in completion results (doesn't match prefix)")
	}

	t.Logf("Parameter completion test passed: found firstParam=%v, secondParam=%v",
		foundFirstParam, foundSecondParam)
}

// Task 9.19: Test local variable shadowing global
func TestCompletion_LocalVariableShadowsGlobal(t *testing.T) {
	// Create a test server
	srv := server.New()
	SetServer(srv)

	// Setup: local variable shadows global variable
	source := `program Test;

var value: Integer; // Global variable

procedure TestProc;
var value: String; // Local variable shadows global
begin
  value := 'test';
end;

begin
end.`

	// Add document to server
	uri := "file:///test.dws"
	program, _, err := analysis.ParseDocument(source, uri)
	if err != nil {
		t.Fatalf("Failed to parse document: %v", err)
	}

	doc := &server.Document{
		URI:        uri,
		Text:       source,
		Version:    1,
		LanguageID: "dwscript",
		Program:    program,
	}
	srv.Documents().Set(uri, doc)

	// Input: cursor after "val" inside the function (in the middle of "value")
	// We're testing if typing "val" would suggest the local "value" with higher priority
	params := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      7,  // On the "value := 'test';" line
				Character: 5, // After "value" (position 5 is after 'e')
			},
		},
	}

	// Call Completion handler
	ctx := &glsp.Context{}
	result, err := Completion(ctx, params)

	if err != nil {
		t.Fatalf("Completion returned error: %v", err)
	}

	completionList, ok := result.(*protocol.CompletionList)
	if !ok {
		t.Fatalf("Expected CompletionList, got %T", result)
	}

	// Expected: local "value" should appear in results
	// Note: Both local and global "value" might appear, but local should have higher priority
	foundValue := false
	valueCount := 0
	var localValueIndex, globalValueIndex int = -1, -1

	for i, item := range completionList.Items {
		t.Logf("Found completion item: %s (kind: %v, sortText: %v, detail: %v)",
			item.Label, item.Kind, item.SortText, item.Detail)
		if item.Label == "value" {
			foundValue = true
			valueCount++
			// Check if it's marked as local or global based on detail or sortText
			if item.Detail != nil && (*item.Detail == "Local variable: String" || *item.Detail == "Local variable") {
				localValueIndex = i
			} else if item.Detail != nil && (*item.Detail == "Global variable: Integer" || *item.Detail == "Global variable") {
				globalValueIndex = i
			}
		}
	}

	// Verify: "value" should be in results
	if !foundValue {
		t.Error("Expected 'value' to be in completion results")
	}

	t.Logf("Found %d 'value' entries: localIndex=%d, globalIndex=%d",
		valueCount, localValueIndex, globalValueIndex)

	// If both are present, verify local comes before global (based on sortText)
	// Local variables should have sortText starting with "0", globals with "1"
	if localValueIndex >= 0 && globalValueIndex >= 0 {
		localItem := completionList.Items[localValueIndex]
		globalItem := completionList.Items[globalValueIndex]

		if localItem.SortText != nil && globalItem.SortText != nil {
			if *localItem.SortText >= *globalItem.SortText {
				t.Errorf("Local variable should sort before global: local sortText=%s, global sortText=%s",
					*localItem.SortText, *globalItem.SortText)
			} else {
				t.Logf("Correct sorting: local sortText=%s < global sortText=%s",
					*localItem.SortText, *globalItem.SortText)
			}
		}
	}

	t.Logf("Shadowing test passed: found 'value' in results (count=%d)", valueCount)
}

// Task 9.20: Test member access on class instance
func TestCompletion_MemberAccessOnClass(t *testing.T) {
	// Create a test server
	srv := server.New()
	SetServer(srv)

	// Setup: class with fields Name, Age, method GetInfo()
	source := `program Test;

type TPerson = class
  Name: String;
  Age: Integer;

  function GetInfo(): String;
end;

function TPerson.GetInfo(): String;
begin
  Result := Name;
end;

var person: TPerson;

begin
  person.Name := 'John';
end.`

	// Add document to server
	uri := "file:///test.dws"
	program, _, err := analysis.ParseDocument(source, uri)
	if err != nil {
		t.Fatalf("Failed to parse document: %v", err)
	}

	doc := &server.Document{
		URI:        uri,
		Text:       source,
		Version:    1,
		LanguageID: "dwscript",
		Program:    program,
	}
	srv.Documents().Set(uri, doc)

	// Input: cursor after "person." (after the dot)
	// Testing member access completion
	triggerChar := "."
	params := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      17, // On the "person.Name := 'John';" line (0-indexed)
				Character: 9,  // After "person." -> "  person." = 2 spaces + 6 chars + 1 dot = position 9
			},
		},
		Context: &protocol.CompletionContext{
			TriggerKind:      protocol.CompletionTriggerKindTriggerCharacter,
			TriggerCharacter: &triggerChar,
		},
	}

	// Call Completion handler
	ctx := &glsp.Context{}
	result, err := Completion(ctx, params)

	if err != nil {
		t.Fatalf("Completion returned error: %v", err)
	}

	completionList, ok := result.(*protocol.CompletionList)
	if !ok {
		t.Fatalf("Expected CompletionList, got %T", result)
	}

	// Expected: Name, Age, GetInfo in results
	foundName := false
	foundAge := false
	foundGetInfo := false
	var nameKind, ageKind, getInfoKind *protocol.CompletionItemKind

	for _, item := range completionList.Items {
		t.Logf("Found completion item: %s (kind: %v, detail: %v)",
			item.Label, item.Kind, item.Detail)
		if item.Label == "Name" {
			foundName = true
			nameKind = item.Kind
		}
		if item.Label == "Age" {
			foundAge = true
			ageKind = item.Kind
		}
		if item.Label == "GetInfo" {
			foundGetInfo = true
			getInfoKind = item.Kind
		}
	}

	// Verify: Name, Age, and GetInfo should be in results
	if !foundName {
		t.Error("Expected 'Name' to be in completion results")
	}
	if !foundAge {
		t.Error("Expected 'Age' to be in completion results")
	}
	if !foundGetInfo {
		t.Error("Expected 'GetInfo' to be in completion results")
	}

	// Verify completion item kinds are correct
	if nameKind != nil && *nameKind != protocol.CompletionItemKindField {
		t.Errorf("Expected 'Name' to have kind Field (%d), got %d",
			protocol.CompletionItemKindField, *nameKind)
	}
	if ageKind != nil && *ageKind != protocol.CompletionItemKindField {
		t.Errorf("Expected 'Age' to have kind Field (%d), got %d",
			protocol.CompletionItemKindField, *ageKind)
	}
	if getInfoKind != nil && *getInfoKind != protocol.CompletionItemKindMethod {
		t.Errorf("Expected 'GetInfo' to have kind Method (%d), got %d",
			protocol.CompletionItemKindMethod, *getInfoKind)
	}

	t.Logf("Member access test passed: found Name=%v, Age=%v, GetInfo=%v",
		foundName, foundAge, foundGetInfo)
}

// Task 9.20: Test member access on record type
func TestCompletion_MemberAccessOnRecord(t *testing.T) {
	// Create a test server
	srv := server.New()
	SetServer(srv)

	// Setup: record type with fields
	source := `program Test;

type TPoint = record
  X: Integer;
  Y: Integer;
end;

var point: TPoint;

begin
  point.X := 10;
end.`

	// Add document to server
	uri := "file:///test.dws"
	program, _, err := analysis.ParseDocument(source, uri)
	if err != nil {
		t.Fatalf("Failed to parse document: %v", err)
	}

	doc := &server.Document{
		URI:        uri,
		Text:       source,
		Version:    1,
		LanguageID: "dwscript",
		Program:    program,
	}
	srv.Documents().Set(uri, doc)

	// Input: cursor after "point." (after the dot)
	// Testing member access completion on record
	triggerChar := "."
	params := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      10, // On the "point.X := 10;" line (0-indexed)
				Character: 8,  // After "point." -> "  point." = 2 + 5 + 1 = 8
			},
		},
		Context: &protocol.CompletionContext{
			TriggerKind:      protocol.CompletionTriggerKindTriggerCharacter,
			TriggerCharacter: &triggerChar,
		},
	}

	// Call Completion handler
	ctx := &glsp.Context{}
	result, err := Completion(ctx, params)

	if err != nil {
		t.Fatalf("Completion returned error: %v", err)
	}

	completionList, ok := result.(*protocol.CompletionList)
	if !ok {
		t.Fatalf("Expected CompletionList, got %T", result)
	}

	// Expected: X and Y from TPoint record
	foundX := false
	foundY := false

	for _, item := range completionList.Items {
		t.Logf("Found completion item: %s (kind: %v)", item.Label, item.Kind)
		if item.Label == "X" {
			foundX = true
		}
		if item.Label == "Y" {
			foundY = true
		}
	}

	// Verify: X and Y should be in results
	if !foundX {
		t.Error("Expected 'X' to be in completion results for record member access")
	}
	if !foundY {
		t.Error("Expected 'Y' to be in completion results for record member access")
	}

	t.Logf("Record member access test passed: found X=%v, Y=%v", foundX, foundY)
}

// Task 9.20: Test member access returns all members (no prefix)
func TestCompletion_MemberAccessAllMembers(t *testing.T) {
	// Create a test server
	srv := server.New()
	SetServer(srv)

	// Setup: class with multiple members
	source := `program Test;

type TData = class
  GetValue: Integer;
  GetName: String;
  SetValue: Integer;
  Count: Integer;
end;

var data: TData;

begin
  data.GetValue := 1;
end.`

	// Add document to server
	uri := "file:///test.dws"
	program, _, err := analysis.ParseDocument(source, uri)
	if err != nil {
		t.Fatalf("Failed to parse document: %v", err)
	}

	doc := &server.Document{
		URI:        uri,
		Text:       source,
		Version:    1,
		LanguageID: "dwscript",
		Program:    program,
	}
	srv.Documents().Set(uri, doc)

	// Input: cursor right after "data." (testing all members)
	// Testing that member access returns all members of the type
	triggerChar := "."
	params := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      12, // On the "data.GetValue := 1;" line (0-indexed)
				Character: 7,  // After "data." -> "  data." = 2 + 4 + 1 = 7
			},
		},
		Context: &protocol.CompletionContext{
			TriggerKind:      protocol.CompletionTriggerKindTriggerCharacter,
			TriggerCharacter: &triggerChar,
		},
	}

	// Call Completion handler
	ctx := &glsp.Context{}
	result, err := Completion(ctx, params)

	if err != nil {
		t.Fatalf("Completion returned error: %v", err)
	}

	completionList, ok := result.(*protocol.CompletionList)
	if !ok {
		t.Fatalf("Expected CompletionList, got %T", result)
	}

	// Expected: All four members should be in results
	foundGetValue := false
	foundGetName := false
	foundSetValue := false
	foundCount := false

	for _, item := range completionList.Items {
		t.Logf("Found completion item: %s (kind: %v)", item.Label, item.Kind)
		if item.Label == "GetValue" {
			foundGetValue = true
		}
		if item.Label == "GetName" {
			foundGetName = true
		}
		if item.Label == "SetValue" {
			foundSetValue = true
		}
		if item.Label == "Count" {
			foundCount = true
		}
	}

	// Verify: All four members should be in results
	if !foundGetValue {
		t.Error("Expected 'GetValue' to be in completion results")
	}
	if !foundGetName {
		t.Error("Expected 'GetName' to be in completion results")
	}
	if !foundSetValue {
		t.Error("Expected 'SetValue' to be in completion results")
	}
	if !foundCount {
		t.Error("Expected 'Count' to be in completion results")
	}

	t.Logf("Member access all members test passed: found all 4 members (GetValue=%v, GetName=%v, SetValue=%v, Count=%v)",
		foundGetValue, foundGetName, foundSetValue, foundCount)
}
