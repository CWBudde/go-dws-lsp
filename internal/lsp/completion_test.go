package lsp

import (
	"strings"
	"testing"

	"github.com/CWBudde/go-dws-lsp/internal/analysis"
	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

const testURI = "file:///test.dws"

const (
	testKeywordIf    = "if"
	testKeywordWhile = "while"
	testKeywordFor   = "for"
	testKeywordVar   = "var"
	testKeywordBegin = "begin"
	testTypeInteger  = "Integer"
	testTypeString   = "String"
	testTypeBoolean  = "Boolean"
	testTypeFloat    = "Float"
)

// setupCompletionTestServer creates and initializes a new test server for completion tests.
func setupCompletionTestServer() *server.Server {
	srv := server.New()
	SetServer(srv)
	return srv
}

// createAndAddTestDocument parses DWScript source and adds it to the server.
func createAndAddTestDocument(t *testing.T, srv *server.Server, source, uri string) *server.Document {
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
	return doc
}

// createCompletionParams creates completion request parameters.
func createCompletionParams(uri string, line, character uint32, triggerChar *string) *protocol.CompletionParams {
	params := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      line,
				Character: character,
			},
		},
	}

	if triggerChar != nil {
		params.Context = &protocol.CompletionContext{
			TriggerKind:      protocol.CompletionTriggerKindTriggerCharacter,
			TriggerCharacter: triggerChar,
		}
	}

	return params
}

// callCompletion calls the Completion handler and returns the result.
func callCompletion(t *testing.T, params *protocol.CompletionParams) *protocol.CompletionList {
	ctx := &glsp.Context{}
	result, err := Completion(ctx, params)
	if err != nil {
		t.Fatalf("Completion returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	completionList, ok := result.(*protocol.CompletionList)
	if !ok {
		t.Fatalf("Expected CompletionList, got %T", result)
	}

	return completionList
}

// findCompletionItem searches for a completion item by label.
func findCompletionItem(items []protocol.CompletionItem, label string) *protocol.CompletionItem {
	for i := range items {
		if items[i].Label == label {
			return &items[i]
		}
	}
	return nil
}

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
	uri := testURI

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
	srv := setupCompletionTestServer()

	source := `program Test;

type TMyClass = class
  Field: Integer;
end;

var obj: TMyClass;

begin
  obj.Field := 42;
end.`

	createAndAddTestDocument(t, srv, source, testURI)

	// Simulate completion request right after typing "obj."
	triggerChar := "."
	params := createCompletionParams(testURI, 9, 6, &triggerChar)
	completionList := callCompletion(t, params)

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
	// Should return an empty completion list without error (graceful handling)
	if err != nil {
		t.Fatalf("Completion returned error: %v", err)
	}

	list, ok := result.(*protocol.CompletionList)
	if !ok {
		t.Fatalf("Expected CompletionList for non-existent document, got %T", result)
	}

	if len(list.Items) != 0 || list.IsIncomplete {
		t.Fatalf("Expected empty completion list for non-existent document, got %+v", list)
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

// Task 9.19: Test partial variable name completion.
func TestCompletion_PartialVariableName(t *testing.T) {
	srv := setupCompletionTestServer()

	source := `program Test;

var alpha: Integer;
var beta: String;
var alphabet: Float;

begin
  alpha := 1;
end.`

	createAndAddTestDocument(t, srv, source, testURI)

	// Input: cursor after "alp" (in the middle of "alpha")
	params := createCompletionParams(testURI, 7, 5, nil)
	completionList := callCompletion(t, params)

	// Expected: alpha and alphabet in results
	foundAlpha := findCompletionItem(completionList.Items, "alpha")
	foundAlphabet := findCompletionItem(completionList.Items, "alphabet")
	foundBeta := findCompletionItem(completionList.Items, "beta")

	if foundAlpha == nil {
		t.Error("Expected 'alpha' to be in completion results")
	}

	if foundAlphabet == nil {
		t.Error("Expected 'alphabet' to be in completion results")
	}

	// Verify: beta should NOT be in results (doesn't match prefix "alp")
	if foundBeta != nil {
		t.Error("Expected 'beta' to NOT be in completion results (doesn't match prefix)")
	}

	t.Logf("Completion test passed: found alpha=%v, alphabet=%v, beta=%v (expected beta=nil)",
		foundAlpha != nil, foundAlphabet != nil, foundBeta != nil)
}

// Task 9.19: Test parameter completion in function.
func TestCompletion_ParameterCompletion(t *testing.T) {
	srv := setupCompletionTestServer()

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

	createAndAddTestDocument(t, srv, source, testURI)

	// Input: cursor after "fir" (testing "firstParam" completion)
	params := createCompletionParams(testURI, 6, 13, nil)
	completionList := callCompletion(t, params)

	// Expected: firstParam in results, secondParam should NOT be
	foundFirstParam := findCompletionItem(completionList.Items, "firstParam")
	foundSecondParam := findCompletionItem(completionList.Items, "secondParam")

	if foundFirstParam == nil {
		t.Error("Expected 'firstParam' to be in completion results")
	} else if foundFirstParam.Kind != nil && *foundFirstParam.Kind != protocol.CompletionItemKindVariable {
		t.Logf("Warning: firstParam has kind %v, expected Variable", *foundFirstParam.Kind)
	}

	if foundSecondParam != nil {
		t.Error("Expected 'secondParam' to NOT be in completion results (doesn't match prefix)")
	}

	t.Logf("Parameter completion test passed: found firstParam=%v, secondParam=%v",
		foundFirstParam != nil, foundSecondParam != nil)
}

// Task 9.19: Test local variable shadowing global.
func TestCompletion_LocalVariableShadowsGlobal(t *testing.T) {
	srv := setupCompletionTestServer()

	source := `program Test;

var value: Integer; // Global variable

procedure TestProc;
var value: String; // Local variable shadows global
begin
  value := 'test';
end;

begin
end.`

	createAndAddTestDocument(t, srv, source, testURI)

	// Input: cursor after "val" inside the function
	params := createCompletionParams(testURI, 7, 5, nil)
	completionList := callCompletion(t, params)

	verifyLocalShadowsGlobal(t, completionList.Items)
}

// verifyLocalShadowsGlobal checks that local variables appear before global ones in completion results.
func verifyLocalShadowsGlobal(t *testing.T, items []protocol.CompletionItem) {
	localValueIndex, globalValueIndex := -1, -1
	valueCount := 0

	for i, item := range items {
		if item.Label == "value" {
			valueCount++
			// Check if it's marked as local or global based on detail
			if item.Detail != nil && (strings.Contains(*item.Detail, "Local variable") || strings.Contains(*item.Detail, "String")) {
				localValueIndex = i
			} else if item.Detail != nil && (strings.Contains(*item.Detail, "Global variable") || strings.Contains(*item.Detail, "Integer")) {
				globalValueIndex = i
			}
		}
	}

	if valueCount == 0 {
		t.Error("Expected 'value' to be in completion results")
		return
	}

	t.Logf("Found %d 'value' entries: localIndex=%d, globalIndex=%d", valueCount, localValueIndex, globalValueIndex)

	// If both are present, verify local comes before global (based on sortText)
	if localValueIndex >= 0 && globalValueIndex >= 0 {
		localItem := items[localValueIndex]
		globalItem := items[globalValueIndex]

		if localItem.SortText != nil && globalItem.SortText != nil {
			if *localItem.SortText >= *globalItem.SortText {
				t.Errorf("Local variable should sort before global: local sortText=%s, global sortText=%s",
					*localItem.SortText, *globalItem.SortText)
			}
		}
	}
}

// Task 9.20: Test member access on class instance.
func TestCompletion_MemberAccessOnClass(t *testing.T) {
	srv := setupCompletionTestServer()
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
	createAndAddTestDocument(t, srv, source, testURI)
	triggerChar := "."
	params := createCompletionParams(testURI, 17, 9, &triggerChar)
	completionList := callCompletion(t, params)
	verifyClassMemberCompletion(t, completionList.Items)
}

// verifyClassMemberCompletion checks that Name, Age, and GetInfo are in completion results with correct kinds.
func verifyClassMemberCompletion(t *testing.T, items []protocol.CompletionItem) {
	foundName := false
	foundAge := false
	foundGetInfo := false
	var nameKind, ageKind, getInfoKind *protocol.CompletionItemKind

	for _, item := range items {
		t.Logf("Found completion item: %s (kind: %v, detail: %v)",
			item.Label, item.Kind, item.Detail)

		switch item.Label {
		case "Name":
			foundName = true
			nameKind = item.Kind
		case "Age":
			foundAge = true
			ageKind = item.Kind
		case "GetInfo":
			foundGetInfo = true
			getInfoKind = item.Kind
		}
	}

	if !foundName {
		t.Error("Expected 'Name' to be in completion results")
	}
	if !foundAge {
		t.Error("Expected 'Age' to be in completion results")
	}
	if !foundGetInfo {
		t.Error("Expected 'GetInfo' to be in completion results")
	}

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

// Task 9.20: Test member access on record type.
func TestCompletion_MemberAccessOnRecord(t *testing.T) {
	srv := setupCompletionTestServer()
	source := `program Test;

type TPoint = record
  X: Integer;
  Y: Integer;
end;

var point: TPoint;

begin
  point.X := 10;
end.`
	createAndAddTestDocument(t, srv, source, testURI)
	triggerChar := "."
	params := createCompletionParams(testURI, 10, 8, &triggerChar)
	completionList := callCompletion(t, params)
	verifyRecordMemberCompletion(t, completionList.Items)
}

// verifyRecordMemberCompletion checks that X and Y record fields are in completion results.
func verifyRecordMemberCompletion(t *testing.T, items []protocol.CompletionItem) {
	foundX := false
	foundY := false

	for _, item := range items {
		t.Logf("Found completion item: %s (kind: %v)", item.Label, item.Kind)
		if item.Label == "X" {
			foundX = true
		}
		if item.Label == "Y" {
			foundY = true
		}
	}

	if !foundX {
		t.Error("Expected 'X' to be in completion results for record member access")
	}
	if !foundY {
		t.Error("Expected 'Y' to be in completion results for record member access")
	}

	t.Logf("Record member access test passed: found X=%v, Y=%v", foundX, foundY)
}

// Task 9.20: Test member access returns all members (no prefix).
func TestCompletion_MemberAccessAllMembers(t *testing.T) {
	srv := setupCompletionTestServer()
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
	createAndAddTestDocument(t, srv, source, testURI)
	triggerChar := "."
	params := createCompletionParams(testURI, 12, 7, &triggerChar)
	completionList := callCompletion(t, params)
	verifyAllMembersPresent(t, completionList.Items)
}

// verifyAllMembersPresent checks that all four class members are in completion results.
func verifyAllMembersPresent(t *testing.T, items []protocol.CompletionItem) {
	expectedMembers := []string{"GetValue", "GetName", "SetValue", "Count"}
	foundMembers := make(map[string]bool)

	for _, item := range items {
		t.Logf("Found completion item: %s (kind: %v)", item.Label, item.Kind)
		for _, expected := range expectedMembers {
			if item.Label == expected {
				foundMembers[expected] = true
			}
		}
	}

	for _, member := range expectedMembers {
		if !foundMembers[member] {
			t.Errorf("Expected '%s' to be in completion results", member)
		}
	}

	t.Logf("Member access all members test passed: found all 4 members")
}

// Task 9.21: Test keyword completion at statement start.
func TestCompletion_KeywordsAtStatementStart(t *testing.T) {
	srv := setupCompletionTestServer()
	source := `program Test;

function DoSomething(): Integer;
begin
  Result := 0;
end;

begin
end.`
	createAndAddTestDocument(t, srv, source, testURI)
	params := createCompletionParams(testURI, 4, 2, nil)
	completionList := callCompletion(t, params)
	verifyKeywordCompletion(t, completionList.Items)
}

// verifyKeywordCompletion checks that expected keywords are in completion results.
func verifyKeywordCompletion(t *testing.T, items []protocol.CompletionItem) {
	expectedKeywords := []string{testKeywordIf, testKeywordWhile, testKeywordFor}
	foundKeywords := make(map[string]bool)
	keywordCount := 0

	for _, item := range items {
		if item.Kind != nil && *item.Kind == protocol.CompletionItemKindKeyword {
			keywordCount++
			for _, expected := range expectedKeywords {
				if item.Label == expected {
					foundKeywords[expected] = true
				}
			}
		}
	}

	for _, keyword := range expectedKeywords {
		if !foundKeywords[keyword] {
			t.Errorf("Expected '%s' keyword to be in completion results", keyword)
		}
	}

	if keywordCount < 10 {
		t.Errorf("Expected at least 10 keywords, found %d", keywordCount)
	}

	t.Logf("Keyword completion test passed: found %d keywords", keywordCount)
}

// Task 9.21: Test built-in function completion.
func TestCompletion_BuiltInFunctions(t *testing.T) {
	srv := setupCompletionTestServer()
	source := `program Test;

begin
  PrintLn('test');
end.`
	createAndAddTestDocument(t, srv, source, testURI)
	params := createCompletionParams(testURI, 3, 2, nil)
	completionList := callCompletion(t, params)
	verifyBuiltInFunctionCompletion(t, completionList.Items)
}

// verifyBuiltInFunctionCompletion checks that expected built-in functions are in completion results.
func verifyBuiltInFunctionCompletion(t *testing.T, items []protocol.CompletionItem) {
	expectedFuncs := []string{"PrintLn", "IntToStr", "Length"}
	foundFuncs := make(map[string]bool)
	builtinFuncCount := 0

	for _, item := range items {
		if item.Kind != nil && *item.Kind == protocol.CompletionItemKindFunction {
			if item.Detail != nil {
				detail := *item.Detail
				if strings.Contains(detail, "(") && strings.Contains(detail, ")") {
					builtinFuncCount++
					for _, expected := range expectedFuncs {
						if item.Label == expected {
							foundFuncs[expected] = true
						}
					}
				}
			}
		}
	}

	for _, fn := range expectedFuncs {
		if !foundFuncs[fn] {
			t.Errorf("Expected '%s' built-in function to be in completion results", fn)
		}
	}

	if builtinFuncCount < 4 {
		t.Errorf("Expected at least 4 built-in functions, found %d", builtinFuncCount)
	}

	t.Logf("Built-in function completion test passed: found %d built-ins", builtinFuncCount)
}

// Task 9.21: Test built-in types completion.
func TestCompletion_BuiltInTypes(t *testing.T) {
	srv := setupCompletionTestServer()
	source := `program Test;

var x: Integer;

begin
end.`
	createAndAddTestDocument(t, srv, source, testURI)
	params := createCompletionParams(testURI, 4, 0, nil)
	completionList := callCompletion(t, params)
	verifyBuiltInTypeCompletion(t, completionList.Items)
}

// verifyBuiltInTypeCompletion checks that expected built-in types are in completion results.
func verifyBuiltInTypeCompletion(t *testing.T, items []protocol.CompletionItem) {
	expectedTypes := []string{testTypeInteger, testTypeString, testTypeBoolean}
	foundTypes := make(map[string]bool)
	builtinTypeCount := 0

	for _, item := range items {
		if item.Kind != nil && *item.Kind == protocol.CompletionItemKindClass {
			builtinTypeCount++
			for _, expected := range expectedTypes {
				if item.Label == expected {
					foundTypes[expected] = true
				}
			}
		}
	}

	for _, typ := range expectedTypes {
		if !foundTypes[typ] {
			t.Errorf("Expected '%s' built-in type to be in completion results", typ)
		}
	}

	if builtinTypeCount < 3 {
		t.Errorf("Expected at least 3 built-in types, found %d", builtinTypeCount)
	}

	t.Logf("Built-in type completion test passed: found %d built-in types", builtinTypeCount)
}
