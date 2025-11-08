package workspace

import (
	"strings"
	"testing"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

const (
	testURI  = "file:///test.dws"
	testURI1 = "file:///test1.dws"
	testURI2 = "file:///test2.dws"
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

	uri := testURI
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

	uri := testURI
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
	uri1 := testURI1
	uri2 := testURI2

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

	uri := testURI
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

	uri1 := testURI1
	uri2 := testURI2

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

	uri1 := testURI1
	uri2 := testURI2

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

	uri := testURI
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

	uri1 := testURI1
	uri2 := testURI2

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

	uri := testURI
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

	uri := testURI
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

	uri := testURI
	symbolRange := protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: 0, Character: 10},
	}

	// Test concurrent access (basic smoke test)
	// In a real scenario, this would be more thorough
	done := make(chan bool)

	// Concurrent writes
	go func() {
		for range 10 {
			index.AddSymbol("Func1", protocol.SymbolKindFunction, uri, symbolRange, "", "")
		}

		done <- true
	}()

	// Concurrent reads
	go func() {
		for range 10 {
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

func TestSymbolIndex_Search_BasicMatching(t *testing.T) {
	index := NewSymbolIndex()

	uri := testURI
	symbolRange := protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: 0, Character: 10},
	}

	// Add test symbols
	index.AddSymbol("testFunc", protocol.SymbolKindFunction, uri, symbolRange, "", "")
	index.AddSymbol("MyTest", protocol.SymbolKindClass, uri, symbolRange, "", "")
	index.AddSymbol("helper", protocol.SymbolKindFunction, uri, symbolRange, "", "")

	// Test substring match
	results := index.Search("test", 100)
	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'test', got %d", len(results))
	}

	// Test no match
	results = index.Search("xyz", 100)
	if len(results) != 0 {
		t.Errorf("Expected 0 results for 'xyz', got %d", len(results))
	}

	// Test empty query
	results = index.Search("", 100)
	if len(results) != 3 {
		t.Errorf("Expected 3 results for empty query, got %d", len(results))
	}
}

func TestSymbolIndex_Search_RelevanceSorting(t *testing.T) {
	index := NewSymbolIndex()

	uri := testURI
	symbolRange := protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: 0, Character: 10},
	}

	// Add symbols with different match types for query "test"
	index.AddSymbol("test", protocol.SymbolKindFunction, uri, symbolRange, "", "")        // Exact match
	index.AddSymbol("TEST", protocol.SymbolKindConstant, uri, symbolRange, "", "")        // Exact match (case insensitive)
	index.AddSymbol("testFunc", protocol.SymbolKindFunction, uri, symbolRange, "", "")    // Prefix match
	index.AddSymbol("TestClass", protocol.SymbolKindClass, uri, symbolRange, "", "")      // Prefix match (case insensitive)
	index.AddSymbol("myTest", protocol.SymbolKindClass, uri, symbolRange, "", "")         // Substring match
	index.AddSymbol("aTestHelper", protocol.SymbolKindFunction, uri, symbolRange, "", "") // Substring match

	// Search for "test"
	results := index.Search("test", 100)

	if len(results) != 6 {
		t.Fatalf("Expected 6 results, got %d", len(results))
	}

	// Verify sorting: exact matches first
	exactCount := 0
	prefixCount := 0
	substringCount := 0
	lastMatchType := matchExact

	for i, result := range results {
		nameLower := strings.ToLower(result.Name)

		var currentMatchType matchType
		switch {
		case nameLower == "test":
			currentMatchType = matchExact
			exactCount++
		case strings.HasPrefix(nameLower, "test"):
			currentMatchType = matchPrefix
			prefixCount++
		default:
			currentMatchType = matchSubstring
			substringCount++
		}

		// Verify match types are in order
		if currentMatchType < lastMatchType {
			t.Errorf("Result %d (%s) has better match type than previous result", i, result.Name)
		}

		lastMatchType = currentMatchType
	}

	// Verify we have the expected counts
	if exactCount != 2 {
		t.Errorf("Expected 2 exact matches, got %d", exactCount)
	}

	if prefixCount != 2 {
		t.Errorf("Expected 2 prefix matches, got %d", prefixCount)
	}

	if substringCount != 2 {
		t.Errorf("Expected 2 substring matches, got %d", substringCount)
	}

	// First two results should be exact matches ("test" and "TEST")
	if exactCount > 0 {
		firstExactMatch := strings.ToLower(results[0].Name)
		if firstExactMatch != "test" {
			t.Errorf("First result should be an exact match, got: %s", results[0].Name)
		}
	}
}

func TestSymbolIndex_Search_MaxResults(t *testing.T) {
	index := NewSymbolIndex()

	uri := testURI
	symbolRange := protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: 0, Character: 10},
	}

	// Add many symbols
	for i := range 20 {
		index.AddSymbol("func"+string(rune('A'+i)), protocol.SymbolKindFunction, uri, symbolRange, "", "")
	}

	// Test max results limit
	results := index.Search("func", 5)
	if len(results) != 5 {
		t.Errorf("Expected 5 results (max limit), got %d", len(results))
	}

	// Test with higher limit
	results = index.Search("func", 15)
	if len(results) != 15 {
		t.Errorf("Expected 15 results, got %d", len(results))
	}

	// Test with no limit (0 means unlimited)
	results = index.Search("func", 0)
	if len(results) != 20 {
		t.Errorf("Expected 20 results (no limit), got %d", len(results))
	}
}

func TestSymbolIndex_Search_CaseInsensitive(t *testing.T) {
	index := NewSymbolIndex()

	uri := testURI
	symbolRange := protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: 0, Character: 10},
	}

	index.AddSymbol("TestFunction", protocol.SymbolKindFunction, uri, symbolRange, "", "")
	index.AddSymbol("UPPERCASE", protocol.SymbolKindClass, uri, symbolRange, "", "")
	index.AddSymbol("lowercase", protocol.SymbolKindVariable, uri, symbolRange, "", "")

	// Test different case variations
	testCases := []struct {
		query         string
		expectedCount int
	}{
		{"test", 1},
		{"TEST", 1},
		{"Test", 1},
		{"upper", 1},
		{"UPPER", 1},
		{"lower", 1},
		{"LOWER", 1},
	}

	for _, tc := range testCases {
		results := index.Search(tc.query, 100)
		if len(results) != tc.expectedCount {
			t.Errorf("Query %q: expected %d results, got %d", tc.query, tc.expectedCount, len(results))
		}
	}
}

func TestSymbolIndex_Search_PrefixVsSubstring(t *testing.T) {
	index := NewSymbolIndex()

	uri := testURI
	symbolRange := protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: 0, Character: 10},
	}

	// Add symbols that demonstrate prefix vs substring matching
	index.AddSymbol("getUserName", protocol.SymbolKindFunction, uri, symbolRange, "", "") // Prefix "get"
	index.AddSymbol("getPassword", protocol.SymbolKindFunction, uri, symbolRange, "", "") // Prefix "get"
	index.AddSymbol("targetPath", protocol.SymbolKindVariable, uri, symbolRange, "", "")  // Substring "get"
	index.AddSymbol("budgetData", protocol.SymbolKindVariable, uri, symbolRange, "", "")  // Substring "get"

	results := index.Search("get", 100)

	if len(results) != 4 {
		t.Fatalf("Expected 4 results, got %d", len(results))
	}

	// First two should be prefix matches (getUserName, getPassword)
	// Last two should be substring matches (targetPath, budgetData)
	for i := range 2 {
		if !strings.HasPrefix(strings.ToLower(results[i].Name), "get") {
			t.Errorf("Result %d (%s) should be a prefix match", i, results[i].Name)
		}
	}

	for i := 2; i < 4; i++ {
		nameLower := strings.ToLower(results[i].Name)
		if strings.HasPrefix(nameLower, "get") {
			t.Errorf("Result %d (%s) should be a substring match, not prefix", i, results[i].Name)
		}

		if !strings.Contains(nameLower, "get") {
			t.Errorf("Result %d (%s) should contain 'get'", i, results[i].Name)
		}
	}
}
