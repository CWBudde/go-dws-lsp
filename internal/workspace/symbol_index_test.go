package workspace

import (
	"testing"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestSymbolIndex_NewSymbolIndex(t *testing.T) {
	index := NewSymbolIndex()

	if index == nil {
		t.Fatal("NewSymbolIndex returned nil")
	}

	if index.GetFileCount() != 0 {
		t.Errorf("Expected 0 files, got %d", index.GetFileCount())
	}

	if index.GetSymbolCount() != 0 {
		t.Errorf("Expected 0 symbols, got %d", index.GetSymbolCount())
	}
}

func TestSymbolIndex_AddSymbol(t *testing.T) {
	index := NewSymbolIndex()

	uri := "file:///test.dws"
	symbolRange := protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: 0, Character: 10},
	}

	index.AddSymbol("TestFunc", protocol.SymbolKindFunction, uri, symbolRange, "", "function TestFunc()")

	if index.GetSymbolCount() != 1 {
		t.Errorf("Expected 1 symbol, got %d", index.GetSymbolCount())
	}

	if index.GetFileCount() != 1 {
		t.Errorf("Expected 1 file, got %d", index.GetFileCount())
	}

	if index.GetTotalLocationCount() != 1 {
		t.Errorf("Expected 1 location, got %d", index.GetTotalLocationCount())
	}
}

func TestSymbolIndex_FindSymbol(t *testing.T) {
	index := NewSymbolIndex()

	uri := "file:///test.dws"
	symbolRange := protocol.Range{
		Start: protocol.Position{Line: 5, Character: 9},
		End:   protocol.Position{Line: 5, Character: 17},
	}

	index.AddSymbol("TestFunc", protocol.SymbolKindFunction, uri, symbolRange, "", "function TestFunc(): Integer")

	locations := index.FindSymbol("TestFunc")

	if len(locations) != 1 {
		t.Fatalf("Expected 1 location, got %d", len(locations))
	}

	loc := locations[0]
	if loc.Name != "TestFunc" {
		t.Errorf("Expected name 'TestFunc', got '%s'", loc.Name)
	}

	if loc.Kind != protocol.SymbolKindFunction {
		t.Errorf("Expected kind Function, got %v", loc.Kind)
	}

	if loc.Location.URI != uri {
		t.Errorf("Expected URI '%s', got '%s'", uri, loc.Location.URI)
	}

	if loc.Location.Range.Start.Line != 5 {
		t.Errorf("Expected start line 5, got %d", loc.Location.Range.Start.Line)
	}
}

func TestSymbolIndex_FindSymbol_NotFound(t *testing.T) {
	index := NewSymbolIndex()

	locations := index.FindSymbol("NonExistent")

	if locations != nil {
		t.Errorf("Expected nil for non-existent symbol, got %v", locations)
	}
}

func TestSymbolIndex_FindSymbol_MultipleLocations(t *testing.T) {
	index := NewSymbolIndex()

	// Add same symbol name in different files (function overloading or same name in different units)
	uri1 := "file:///test1.dws"
	uri2 := "file:///test2.dws"

	range1 := protocol.Range{
		Start: protocol.Position{Line: 1, Character: 0},
		End:   protocol.Position{Line: 1, Character: 10},
	}

	range2 := protocol.Range{
		Start: protocol.Position{Line: 5, Character: 0},
		End:   protocol.Position{Line: 5, Character: 10},
	}

	index.AddSymbol("Helper", protocol.SymbolKindFunction, uri1, range1, "", "function Helper(): String")
	index.AddSymbol("Helper", protocol.SymbolKindFunction, uri2, range2, "", "function Helper(): Integer")

	locations := index.FindSymbol("Helper")

	if len(locations) != 2 {
		t.Fatalf("Expected 2 locations, got %d", len(locations))
	}

	// Verify both locations are present
	foundUri1 := false
	foundUri2 := false

	for _, loc := range locations {
		if loc.Location.URI == uri1 {
			foundUri1 = true
		}
		if loc.Location.URI == uri2 {
			foundUri2 = true
		}
	}

	if !foundUri1 || !foundUri2 {
		t.Error("Expected locations from both files")
	}
}

func TestSymbolIndex_RemoveFile(t *testing.T) {
	index := NewSymbolIndex()

	uri := "file:///test.dws"
	symbolRange := protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: 0, Character: 10},
	}

	// Add multiple symbols from the same file
	index.AddSymbol("Func1", protocol.SymbolKindFunction, uri, symbolRange, "", "")
	index.AddSymbol("Func2", protocol.SymbolKindFunction, uri, symbolRange, "", "")
	index.AddSymbol("MyClass", protocol.SymbolKindClass, uri, symbolRange, "", "")

	if index.GetSymbolCount() != 3 {
		t.Errorf("Expected 3 symbols, got %d", index.GetSymbolCount())
	}

	// Remove the file
	index.RemoveFile(uri)

	if index.GetSymbolCount() != 0 {
		t.Errorf("Expected 0 symbols after removing file, got %d", index.GetSymbolCount())
	}

	if index.GetFileCount() != 0 {
		t.Errorf("Expected 0 files after removing file, got %d", index.GetFileCount())
	}

	// Verify symbols are gone
	if locs := index.FindSymbol("Func1"); locs != nil {
		t.Error("Expected Func1 to be removed")
	}
}

func TestSymbolIndex_RemoveFile_KeepsOtherFiles(t *testing.T) {
	index := NewSymbolIndex()

	uri1 := "file:///test1.dws"
	uri2 := "file:///test2.dws"

	symbolRange := protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: 0, Character: 10},
	}

	// Add symbols from two files
	index.AddSymbol("Func1", protocol.SymbolKindFunction, uri1, symbolRange, "", "")
	index.AddSymbol("Func2", protocol.SymbolKindFunction, uri2, symbolRange, "", "")

	// Remove only one file
	index.RemoveFile(uri1)

	if index.GetFileCount() != 1 {
		t.Errorf("Expected 1 file remaining, got %d", index.GetFileCount())
	}

	// Verify Func1 is gone
	if locs := index.FindSymbol("Func1"); locs != nil {
		t.Error("Expected Func1 to be removed")
	}

	// Verify Func2 still exists
	locs := index.FindSymbol("Func2")
	if locs == nil || len(locs) != 1 {
		t.Error("Expected Func2 to still exist")
	}
}

func TestSymbolIndex_RemoveFile_WithSharedSymbolNames(t *testing.T) {
	index := NewSymbolIndex()

	uri1 := "file:///test1.dws"
	uri2 := "file:///test2.dws"

	symbolRange := protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: 0, Character: 10},
	}

	// Add same symbol name in different files
	index.AddSymbol("Helper", protocol.SymbolKindFunction, uri1, symbolRange, "", "")
	index.AddSymbol("Helper", protocol.SymbolKindFunction, uri2, symbolRange, "", "")

	// Remove one file
	index.RemoveFile(uri1)

	// Verify symbol still exists (from uri2)
	locations := index.FindSymbol("Helper")
	if locations == nil || len(locations) != 1 {
		t.Fatalf("Expected 1 location remaining, got %v", locations)
	}

	if locations[0].Location.URI != uri2 {
		t.Errorf("Expected location from uri2, got %s", locations[0].Location.URI)
	}
}

func TestSymbolIndex_FindSymbolsByKind(t *testing.T) {
	index := NewSymbolIndex()

	uri := "file:///test.dws"
	symbolRange := protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: 0, Character: 10},
	}

	// Add symbols of different kinds
	index.AddSymbol("Func1", protocol.SymbolKindFunction, uri, symbolRange, "", "")
	index.AddSymbol("Func2", protocol.SymbolKindFunction, uri, symbolRange, "", "")
	index.AddSymbol("MyClass", protocol.SymbolKindClass, uri, symbolRange, "", "")
	index.AddSymbol("myVar", protocol.SymbolKindVariable, uri, symbolRange, "", "")

	// Find all functions
	functions := index.FindSymbolsByKind(protocol.SymbolKindFunction)

	if len(functions) != 2 {
		t.Errorf("Expected 2 functions, got %d", len(functions))
	}

	// Find all classes
	classes := index.FindSymbolsByKind(protocol.SymbolKindClass)

	if len(classes) != 1 {
		t.Errorf("Expected 1 class, got %d", len(classes))
	}
}

func TestSymbolIndex_FindSymbolsInFile(t *testing.T) {
	index := NewSymbolIndex()

	uri1 := "file:///test1.dws"
	uri2 := "file:///test2.dws"

	symbolRange := protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: 0, Character: 10},
	}

	// Add symbols to different files
	index.AddSymbol("Func1", protocol.SymbolKindFunction, uri1, symbolRange, "", "")
	index.AddSymbol("Func2", protocol.SymbolKindFunction, uri1, symbolRange, "", "")
	index.AddSymbol("Func3", protocol.SymbolKindFunction, uri2, symbolRange, "", "")

	// Find symbols in file 1
	symbols := index.FindSymbolsInFile(uri1)

	if len(symbols) != 2 {
		t.Errorf("Expected 2 symbols in file1, got %d", len(symbols))
	}

	// Find symbols in file 2
	symbols = index.FindSymbolsInFile(uri2)

	if len(symbols) != 1 {
		t.Errorf("Expected 1 symbol in file2, got %d", len(symbols))
	}
}

func TestSymbolIndex_ContainerName(t *testing.T) {
	index := NewSymbolIndex()

	uri := "file:///test.dws"
	symbolRange := protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: 0, Character: 10},
	}

	// Add a method with container name
	index.AddSymbol("DoSomething", protocol.SymbolKindMethod, uri, symbolRange, "TMyClass", "function DoSomething(): Integer")

	locations := index.FindSymbol("DoSomething")

	if len(locations) != 1 {
		t.Fatalf("Expected 1 location, got %d", len(locations))
	}

	if locations[0].ContainerName != "TMyClass" {
		t.Errorf("Expected container name 'TMyClass', got '%s'", locations[0].ContainerName)
	}

	if locations[0].Detail != "function DoSomething(): Integer" {
		t.Errorf("Expected detail to be preserved, got '%s'", locations[0].Detail)
	}
}

func TestSymbolIndex_Clear(t *testing.T) {
	index := NewSymbolIndex()

	uri := "file:///test.dws"
	symbolRange := protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: 0, Character: 10},
	}

	// Add some symbols
	index.AddSymbol("Func1", protocol.SymbolKindFunction, uri, symbolRange, "", "")
	index.AddSymbol("Func2", protocol.SymbolKindFunction, uri, symbolRange, "", "")

	// Clear the index
	index.Clear()

	if index.GetSymbolCount() != 0 {
		t.Errorf("Expected 0 symbols after clear, got %d", index.GetSymbolCount())
	}

	if index.GetFileCount() != 0 {
		t.Errorf("Expected 0 files after clear, got %d", index.GetFileCount())
	}
}

func TestSymbolIndex_ThreadSafety(t *testing.T) {
	index := NewSymbolIndex()

	uri := "file:///test.dws"
	symbolRange := protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: 0, Character: 10},
	}

	// Test concurrent access (basic smoke test)
	// In a real scenario, this would be more thorough
	done := make(chan bool)

	// Concurrent writes
	go func() {
		for i := 0; i < 10; i++ {
			index.AddSymbol("Func1", protocol.SymbolKindFunction, uri, symbolRange, "", "")
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 10; i++ {
			index.FindSymbol("Func1")
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Just verify it didn't crash
	if index == nil {
		t.Error("Index should not be nil after concurrent access")
	}
}
