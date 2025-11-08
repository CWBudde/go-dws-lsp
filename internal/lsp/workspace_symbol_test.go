package lsp

import (
	"testing"

	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/CWBudde/go-dws-lsp/internal/workspace"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestWorkspaceSymbol_EmptyIndex(t *testing.T) {
	// Setup
	srv := server.New()
	SetServer(srv)

	// Test with empty index
	result, err := WorkspaceSymbol(nil, &protocol.WorkspaceSymbolParams{
		Query: "test",
	})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Empty index returns empty array ([]protocol.SymbolInformation{})
	// which is correct behavior
	if len(result) != 0 {
		t.Errorf("Expected 0 symbols, got: %d", len(result))
	}
}

func TestWorkspaceSymbol_WithSymbols(t *testing.T) {
	// Setup
	srv := server.New()
	SetServer(srv)

	index := srv.WorkspaceIndex()

	// Add some test symbols
	index.AddSymbol("testFunc", protocol.SymbolKindFunction, "file:///test.dws",
		protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
			End:   protocol.Position{Line: 0, Character: 8},
		}, "", "function testFunc(): String")

	index.AddSymbol("MyClass", protocol.SymbolKindClass, "file:///test.dws",
		protocol.Range{
			Start: protocol.Position{Line: 10, Character: 0},
			End:   protocol.Position{Line: 20, Character: 3},
		}, "", "class MyClass")

	index.AddSymbol("globalVar", protocol.SymbolKindVariable, "file:///test.dws",
		protocol.Range{
			Start: protocol.Position{Line: 5, Character: 0},
			End:   protocol.Position{Line: 5, Character: 9},
		}, "", "var: String")

	// Test query matching "test"
	result, err := WorkspaceSymbol(nil, &protocol.WorkspaceSymbolParams{
		Query: "test",
	})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 symbol matching 'test', got: %d", len(result))
	}

	if len(result) > 0 && result[0].Name != "testFunc" {
		t.Errorf("Expected symbol name 'testFunc', got: %s", result[0].Name)
	}

	// Test query matching "Class"
	result, err = WorkspaceSymbol(nil, &protocol.WorkspaceSymbolParams{
		Query: "Class",
	})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 symbol matching 'Class', got: %d", len(result))
	}

	if len(result) > 0 && result[0].Name != "MyClass" {
		t.Errorf("Expected symbol name 'MyClass', got: %s", result[0].Name)
	}
}

func TestWorkspaceSymbol_CaseInsensitive(t *testing.T) {
	// Setup
	srv := server.New()
	SetServer(srv)

	index := srv.WorkspaceIndex()

	// Add test symbol
	index.AddSymbol("TestFunction", protocol.SymbolKindFunction, "file:///test.dws",
		protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
			End:   protocol.Position{Line: 0, Character: 12},
		}, "", "function TestFunction(): Integer")

	// Test case-insensitive search
	testCases := []string{"test", "TEST", "Test", "testfunction", "TESTFUNCTION"}
	for _, query := range testCases {
		result, err := WorkspaceSymbol(nil, &protocol.WorkspaceSymbolParams{
			Query: query,
		})
		if err != nil {
			t.Errorf("Expected no error for query %q, got: %v", query, err)
		}

		if len(result) != 1 {
			t.Errorf("Expected 1 symbol for query %q, got: %d", query, len(result))
		}

		if len(result) > 0 && result[0].Name != "TestFunction" {
			t.Errorf("Expected symbol name 'TestFunction' for query %q, got: %s", query, result[0].Name)
		}
	}
}

func TestWorkspaceSymbol_EmptyQuery(t *testing.T) {
	// Setup
	srv := server.New()
	SetServer(srv)

	index := srv.WorkspaceIndex()

	// Add some test symbols
	index.AddSymbol("func1", protocol.SymbolKindFunction, "file:///test.dws",
		protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
			End:   protocol.Position{Line: 0, Character: 5},
		}, "", "function func1()")

	index.AddSymbol("func2", protocol.SymbolKindFunction, "file:///test.dws",
		protocol.Range{
			Start: protocol.Position{Line: 10, Character: 0},
			End:   protocol.Position{Line: 10, Character: 5},
		}, "", "function func2()")

	index.AddSymbol("var1", protocol.SymbolKindVariable, "file:///test.dws",
		protocol.Range{
			Start: protocol.Position{Line: 5, Character: 0},
			End:   protocol.Position{Line: 5, Character: 4},
		}, "", "var: Integer")

	// Test empty query (should return all symbols)
	result, err := WorkspaceSymbol(nil, &protocol.WorkspaceSymbolParams{
		Query: "",
	})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 symbols for empty query, got: %d", len(result))
	}
}

func TestWorkspaceSymbol_WithContainerName(t *testing.T) {
	// Setup
	srv := server.New()
	SetServer(srv)

	index := srv.WorkspaceIndex()

	// Add symbol with container name
	index.AddSymbol("myMethod", protocol.SymbolKindMethod, "file:///test.dws",
		protocol.Range{
			Start: protocol.Position{Line: 15, Character: 2},
			End:   protocol.Position{Line: 15, Character: 10},
		}, "MyClass", "function myMethod()")

	// Test query
	result, err := WorkspaceSymbol(nil, &protocol.WorkspaceSymbolParams{
		Query: "myMethod",
	})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 symbol, got: %d", len(result))
		return
	}

	if result[0].Name != "myMethod" {
		t.Errorf("Expected symbol name 'myMethod', got: %s", result[0].Name)
	}

	if result[0].ContainerName == nil || *result[0].ContainerName != "MyClass" {
		if result[0].ContainerName == nil {
			t.Error("Expected container name 'MyClass', got nil")
		} else {
			t.Errorf("Expected container name 'MyClass', got: %s", *result[0].ContainerName)
		}
	}
}

func TestWorkspaceSymbol_MultipleFiles(t *testing.T) {
	// Setup
	srv := server.New()
	SetServer(srv)

	index := srv.WorkspaceIndex()

	// Add symbols from different files
	index.AddSymbol("func1", protocol.SymbolKindFunction, "file:///file1.dws",
		protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
			End:   protocol.Position{Line: 0, Character: 5},
		}, "", "function func1()")

	index.AddSymbol("func2", protocol.SymbolKindFunction, "file:///file2.dws",
		protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
			End:   protocol.Position{Line: 0, Character: 5},
		}, "", "function func2()")

	index.AddSymbol("func3", protocol.SymbolKindFunction, "file:///file3.dws",
		protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
			End:   protocol.Position{Line: 0, Character: 5},
		}, "", "function func3()")

	// Test query matching all functions
	result, err := WorkspaceSymbol(nil, &protocol.WorkspaceSymbolParams{
		Query: "func",
	})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 symbols matching 'func', got: %d", len(result))
	}

	// Verify URIs are different
	uris := make(map[string]bool)
	for _, sym := range result {
		uris[sym.Location.URI] = true
	}

	if len(uris) != 3 {
		t.Errorf("Expected symbols from 3 different files, got: %d", len(uris))
	}
}

func TestWorkspaceIndex_Search(t *testing.T) {
	index := workspace.NewSymbolIndex()

	// Add test symbols
	index.AddSymbol("testFunc", protocol.SymbolKindFunction, "file:///test.dws",
		protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
			End:   protocol.Position{Line: 0, Character: 8},
		}, "", "function testFunc(): String")

	index.AddSymbol("MyClass", protocol.SymbolKindClass, "file:///test.dws",
		protocol.Range{
			Start: protocol.Position{Line: 10, Character: 0},
			End:   protocol.Position{Line: 20, Character: 3},
		}, "", "class MyClass")

	index.AddSymbol("globalVar", protocol.SymbolKindVariable, "file:///test.dws",
		protocol.Range{
			Start: protocol.Position{Line: 5, Character: 0},
			End:   protocol.Position{Line: 5, Character: 9},
		}, "", "var: String")

	// Test search
	results := index.Search("test", 100)
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'test', got: %d", len(results))
	}

	results = index.Search("My", 100)
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'My', got: %d", len(results))
	}

	results = index.Search("", 100)
	if len(results) != 3 {
		t.Errorf("Expected 3 results for empty query, got: %d", len(results))
	}

	// Test max results limit
	results = index.Search("", 2)
	if len(results) != 2 {
		t.Errorf("Expected 2 results (max limit), got: %d", len(results))
	}
}
