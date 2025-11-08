//go:build integration
// +build integration

package integration

import (
	"testing"

	"github.com/CWBudde/go-dws-lsp/internal/lsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp"
)

// TestSignatureHelp_AfterOpeningParenthesis tests signature help immediately after opening parenthesis
func TestSignatureHelp_AfterOpeningParenthesis(t *testing.T) {
	srv := setupTestServer()

	uri := "file:///test/signature_test.dws"
	code := `function Calculate(x, y: Integer): Integer;
begin
  Result := x + y;
end;

var z: Integer;
z := Calculate(0, 0);
`

	// Open the document
	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request signature help after opening parenthesis
	// Line 6 is: z := Calculate(0, 0);
	// Character 16 is right after the opening paren
	sigParams := &protocol.SignatureHelpParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 6, Character: 16}, // Inside "Calculate(...)" after the opening paren
		},
	}

	sigHelp, err := lsp.SignatureHelp(ctx, sigParams)
	if err != nil {
		t.Fatalf("SignatureHelp failed: %v", err)
	}

	if sigHelp == nil {
		t.Fatal("Expected signature help, got nil")
	}

	// Verify we got signature information
	if len(sigHelp.Signatures) == 0 {
		t.Fatal("Expected at least one signature")
	}

	sig := sigHelp.Signatures[0]

	// Verify signature label contains function name and parameters
	// Note: The signature extractor reads parameters individually, so x and y are separate params
	if !contains(sig.Label, "Calculate") || !contains(sig.Label, "Integer") {
		t.Errorf("Expected signature to contain 'Calculate' and 'Integer', got %q", sig.Label)
	}

	// Verify activeParameter is 0 (first parameter)
	if sigHelp.ActiveParameter == nil {
		t.Fatal("Expected activeParameter to be set")
	}
	if *sigHelp.ActiveParameter != 0 {
		t.Errorf("Expected activeParameter 0, got %d", *sigHelp.ActiveParameter)
	}

	// Verify parameters array (should have at least the parameters)
	if len(sig.Parameters) < 1 {
		t.Errorf("Expected at least 1 parameter, got %d", len(sig.Parameters))
	}

	t.Logf("Signature: %s", sig.Label)
	t.Logf("Active parameter: %d", *sigHelp.ActiveParameter)

	_ = srv
}

// TestSignatureHelp_AfterComma tests signature help after typing a comma
func TestSignatureHelp_AfterComma(t *testing.T) {
	srv := setupTestServer()

	uri := "file:///test/signature_comma.dws"
	code := `function Add(a: Integer, b: Integer): Integer;
begin
  Result := a + b;
end;

begin
  Add(5,
end.`

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request signature help after comma
	sigParams := &protocol.SignatureHelpParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 6, Character: 9}, // After "Add(5, "
		},
	}

	sigHelp, err := lsp.SignatureHelp(ctx, sigParams)
	if err != nil {
		t.Fatalf("SignatureHelp failed: %v", err)
	}

	if sigHelp == nil {
		t.Fatal("Expected signature help, got nil")
	}

	// Verify activeParameter is 1 (second parameter)
	if sigHelp.ActiveParameter == nil {
		t.Fatal("Expected activeParameter to be set")
	}
	if *sigHelp.ActiveParameter != 1 {
		t.Errorf("Expected activeParameter 1, got %d", *sigHelp.ActiveParameter)
	}

	t.Logf("Active parameter after comma: %d", *sigHelp.ActiveParameter)

	_ = srv
}

// TestSignatureHelp_BuiltinFunction tests signature help for built-in functions
func TestSignatureHelp_BuiltinFunction(t *testing.T) {
	srv := setupTestServer()

	uri := "file:///test/builtin.dws"
	code := `begin
  PrintLn(
end.`

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request signature help for built-in function
	sigParams := &protocol.SignatureHelpParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 1, Character: 10}, // After "PrintLn("
		},
	}

	sigHelp, err := lsp.SignatureHelp(ctx, sigParams)
	if err != nil {
		t.Fatalf("SignatureHelp failed: %v", err)
	}

	if sigHelp == nil {
		t.Fatal("Expected signature help for built-in function, got nil")
	}

	// Verify signature for PrintLn
	if len(sigHelp.Signatures) == 0 {
		t.Fatal("Expected signature for PrintLn")
	}

	sig := sigHelp.Signatures[0]
	if sig.Label != "procedure PrintLn(text: String)" {
		t.Errorf("Expected PrintLn signature, got %q", sig.Label)
	}

	// Verify documentation is present
	if sig.Documentation != nil {
		doc := sig.Documentation.(protocol.MarkupContent)
		t.Logf("Documentation: %s", doc.Value)
	}

	t.Logf("Built-in function signature: %s", sig.Label)

	_ = srv
}

// TestSignatureHelp_MultipleParameters tests various cursor positions with multiple parameters
func TestSignatureHelp_MultipleParameters(t *testing.T) {
	srv := setupTestServer()

	uri := "file:///test/multi_param.dws"
	code := `function Test(a: Integer, b: String, c: Float): Boolean;
begin
  Result := True;
end;

begin
  Test(1, 'hello', 3.14
end.`

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	testCases := []struct {
		name              string
		line              uint32
		character         uint32
		expectedParamIdx  uint32
		description       string
	}{
		{
			name:             "before first parameter",
			line:             6,
			character:        7, // After "Test("
			expectedParamIdx: 0,
			description:      "cursor right after opening paren",
		},
		{
			name:             "in first parameter",
			line:             6,
			character:        8, // At "Test(1"
			expectedParamIdx: 0,
			description:      "cursor in middle of first parameter",
		},
		{
			name:             "after first comma",
			line:             6,
			character:        11, // After "Test(1, "
			expectedParamIdx: 1,
			description:      "cursor after first comma",
		},
		{
			name:             "in second parameter",
			line:             6,
			character:        15, // At "Test(1, 'hello"
			expectedParamIdx: 1,
			description:      "cursor in middle of second parameter",
		},
		{
			name:             "after second comma",
			line:             6,
			character:        20, // After "Test(1, 'hello', "
			expectedParamIdx: 2,
			description:      "cursor after second comma",
		},
		{
			name:             "in third parameter",
			line:             6,
			character:        24, // At "Test(1, 'hello', 3.14"
			expectedParamIdx: 2,
			description:      "cursor in third parameter",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sigParams := &protocol.SignatureHelpParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: uri},
					Position:     protocol.Position{Line: tc.line, Character: tc.character},
				},
			}

			sigHelp, err := lsp.SignatureHelp(ctx, sigParams)
			if err != nil {
				t.Fatalf("SignatureHelp failed: %v", err)
			}

			if sigHelp == nil {
				t.Fatalf("Expected signature help at position %d:%d (%s), got nil",
					tc.line, tc.character, tc.description)
			}

			if sigHelp.ActiveParameter == nil {
				t.Fatal("Expected activeParameter to be set")
			}

			if *sigHelp.ActiveParameter != tc.expectedParamIdx {
				t.Errorf("%s: expected activeParameter %d, got %d",
					tc.description, tc.expectedParamIdx, *sigHelp.ActiveParameter)
			}

			t.Logf("%s: activeParameter = %d ✓", tc.description, *sigHelp.ActiveParameter)
		})
	}

	_ = srv
}

// TestSignatureHelp_NestedCalls tests signature help with nested function calls
func TestSignatureHelp_NestedCalls(t *testing.T) {
	srv := setupTestServer()

	uri := "file:///test/nested.dws"
	code := `function Outer(x: Integer): Integer;
begin
  Result := x;
end;

function Inner(y: String): String;
begin
  Result := y;
end;

begin
  Outer(Inner(
end.`

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request signature help for the inner function call
	sigParams := &protocol.SignatureHelpParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 11, Character: 14}, // After "Inner("
		},
	}

	sigHelp, err := lsp.SignatureHelp(ctx, sigParams)
	if err != nil {
		t.Fatalf("SignatureHelp failed: %v", err)
	}

	if sigHelp == nil {
		t.Fatal("Expected signature help for nested call, got nil")
	}

	// Should show signature for the innermost function (Inner)
	sig := sigHelp.Signatures[0]
	if !contains(sig.Label, "Inner") {
		t.Errorf("Expected signature for Inner function, got %q", sig.Label)
	}

	t.Logf("Nested call signature: %s", sig.Label)

	_ = srv
}

// TestSignatureHelp_ZeroParameters tests signature help for functions with no parameters
func TestSignatureHelp_ZeroParameters(t *testing.T) {
	srv := setupTestServer()

	uri := "file:///test/no_params.dws"
	code := `function GetValue(): Integer;
begin
  Result := 42;
end;

begin
  GetValue(
end.`

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	sigParams := &protocol.SignatureHelpParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 6, Character: 11}, // After "GetValue("
		},
	}

	sigHelp, err := lsp.SignatureHelp(ctx, sigParams)
	if err != nil {
		t.Fatalf("SignatureHelp failed: %v", err)
	}

	if sigHelp == nil {
		t.Fatal("Expected signature help, got nil")
	}

	sig := sigHelp.Signatures[0]

	// Verify no parameters in signature
	if len(sig.Parameters) != 0 {
		t.Errorf("Expected 0 parameters, got %d", len(sig.Parameters))
	}

	// Verify the label shows empty parameter list
	expectedLabel := "function GetValue(): Integer"
	if sig.Label != expectedLabel {
		t.Errorf("Expected label %q, got %q", expectedLabel, sig.Label)
	}

	t.Logf("Zero-parameter function signature: %s", sig.Label)

	_ = srv
}

// TestSignatureHelp_IncompleteCall tests signature help with incomplete function calls
func TestSignatureHelp_IncompleteCall(t *testing.T) {
	srv := setupTestServer()

	uri := "file:///test/incomplete.dws"
	code := `function Process(data: String, count: Integer);
begin
end;

begin
  Process('test',
end.`

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "dwscript",
			Version:    1,
			Text:       code,
		},
	}

	ctx := &glsp.Context{}
	err := lsp.DidOpen(ctx, openParams)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	sigParams := &protocol.SignatureHelpParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 5, Character: 18}, // After comma
		},
	}

	sigHelp, err := lsp.SignatureHelp(ctx, sigParams)
	if err != nil {
		t.Fatalf("SignatureHelp failed: %v", err)
	}

	if sigHelp == nil {
		t.Fatal("Expected signature help for incomplete call, got nil")
	}

	// Should show second parameter as active
	if sigHelp.ActiveParameter == nil {
		t.Fatal("Expected activeParameter to be set")
	}
	if *sigHelp.ActiveParameter != 1 {
		t.Errorf("Expected activeParameter 1, got %d", *sigHelp.ActiveParameter)
	}

	t.Logf("Incomplete call handled correctly, active parameter: %d", *sigHelp.ActiveParameter)

	_ = srv
}
