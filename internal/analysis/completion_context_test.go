package analysis

import (
	"testing"

	"github.com/CWBudde/go-dws-lsp/internal/server"
)

func TestDetermineContext_General(t *testing.T) {
	source := `program Test;

var x: Integer;

begin
  x := 42;
end.`

	doc := &server.Document{
		URI:  "file:///test.dws",
		Text: source,
	}

	program, _, _ := ParseDocument(source, doc.URI)
	doc.Program = program

	// Position inside the begin/end block (line 4 is "  x := 42;", character 2 is before "x")
	ctx, err := DetermineContext(doc, 4, 2)
	if err != nil {
		t.Fatalf("DetermineContext failed: %v", err)
	}

	if ctx == nil {
		t.Fatal("Expected non-nil context")
	}

	if ctx.Type != CompletionContextGeneral {
		t.Errorf("Expected CompletionContextGeneral, got %v", ctx.Type)
	}

	if ctx.ParentIdentifier != "" {
		t.Errorf("Expected empty ParentIdentifier, got %q", ctx.ParentIdentifier)
	}
}

func TestDetermineContext_MemberAccess(t *testing.T) {
	source := `program Test;

var obj: TMyClass;

begin
  obj.
end.`

	doc := &server.Document{
		URI:  "file:///test.dws",
		Text: source,
	}

	program, _, _ := ParseDocument(source, doc.URI)
	doc.Program = program

	// Position after "obj." (line 5, character 6)
	ctx, err := DetermineContext(doc, 5, 6)
	if err != nil {
		t.Fatalf("DetermineContext failed: %v", err)
	}

	if ctx == nil {
		t.Fatal("Expected non-nil context")
	}

	if ctx.Type != CompletionContextMember {
		t.Errorf("Expected CompletionContextMember, got %v", ctx.Type)
	}

	if ctx.ParentIdentifier != "obj" {
		t.Errorf("Expected ParentIdentifier 'obj', got %q", ctx.ParentIdentifier)
	}
}

func TestDetermineContext_InsideComment(t *testing.T) {
	source := `program Test;

var x: Integer;

begin
  // This is a comment
  x := 42;
end.`

	doc := &server.Document{
		URI:  "file:///test.dws",
		Text: source,
	}

	program, _, _ := ParseDocument(source, doc.URI)
	doc.Program = program

	// Position inside the comment (line 5, character 10)
	ctx, err := DetermineContext(doc, 5, 10)
	if err != nil {
		t.Fatalf("DetermineContext failed: %v", err)
	}

	// Context should explicitly signal "none" when inside a comment
	if ctx == nil || ctx.Type != CompletionContextNone {
		t.Error("Expected CompletionContextNone when inside comment")
	}
}

func TestDetermineContext_InsideString(t *testing.T) {
	source := `program Test;

var s: String;

begin
  s := 'hello world';
end.`

	doc := &server.Document{
		URI:  "file:///test.dws",
		Text: source,
	}

	program, _, _ := ParseDocument(source, doc.URI)
	doc.Program = program

	// Position inside the string (line 5, character 12)
	ctx, err := DetermineContext(doc, 5, 12)
	if err != nil {
		t.Fatalf("DetermineContext failed: %v", err)
	}

	// Context should explicitly signal "none" when inside a string
	if ctx == nil || ctx.Type != CompletionContextNone {
		t.Error("Expected CompletionContextNone when inside string")
	}
}

func TestExtractParentIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple identifier with dot",
			input:    "myObject.",
			expected: "myObject",
		},
		{
			name:     "identifier with whitespace before dot",
			input:    "myObject  .",
			expected: "myObject",
		},
		{
			name:     "no dot",
			input:    "myObject",
			expected: "",
		},
		{
			name:     "dot but no identifier",
			input:    ".",
			expected: "",
		},
		{
			name:     "identifier with underscore",
			input:    "my_object.",
			expected: "my_object",
		},
		{
			name:     "identifier with number",
			input:    "obj123.",
			expected: "obj123",
		},
		{
			name:     "complex expression",
			input:    "var x: Integer;\n  obj.",
			expected: "obj",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractParentIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestIsInsideComment(t *testing.T) {
	tests := []struct {
		name             string
		textBeforeCursor string
		fullText         string
		line             int
		character        int
		expected         bool
	}{
		{
			name:             "single line comment",
			textBeforeCursor: "var x: Integer; // This is",
			fullText:         "var x: Integer; // This is a comment",
			line:             0,
			character:        25,
			expected:         true,
		},
		{
			name:             "not in comment",
			textBeforeCursor: "var x: Integer;",
			fullText:         "var x: Integer; // Comment later",
			line:             0,
			character:        15,
			expected:         false,
		},
		{
			name:             "inside multiline comment with parentheses",
			textBeforeCursor: "var x: Integer; (* This is\na comment",
			fullText:         "var x: Integer; (* This is\na comment *)",
			line:             1,
			character:        10,
			expected:         true,
		},
		{
			name:             "after multiline comment",
			textBeforeCursor: "var x: Integer; (* comment *) var y",
			fullText:         "var x: Integer; (* comment *) var y: String;",
			line:             0,
			character:        35,
			expected:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInsideComment(tt.textBeforeCursor, tt.fullText, tt.line, tt.character)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsInsideString(t *testing.T) {
	tests := []struct {
		name             string
		textBeforeCursor string
		expected         bool
	}{
		{
			name:             "inside string",
			textBeforeCursor: "var s: String := 'hello",
			expected:         true,
		},
		{
			name:             "after string",
			textBeforeCursor: "var s: String := 'hello';",
			expected:         false,
		},
		{
			name:             "no string",
			textBeforeCursor: "var x: Integer;",
			expected:         false,
		},
		{
			name:             "escaped quote",
			textBeforeCursor: "var s: String := 'can''t",
			expected:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInsideString(tt.textBeforeCursor)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
