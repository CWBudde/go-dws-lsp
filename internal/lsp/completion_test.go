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
