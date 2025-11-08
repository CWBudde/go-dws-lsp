package lsp

import (
	"testing"

	"github.com/CWBudde/go-dws-lsp/internal/analysis"
	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// parseCodeToDocument is a helper function to compile DWScript code for testing.
func parseCodeToDocument(t *testing.T, code string, uri string) *server.Document {
	t.Helper()

	program, _, err := analysis.ParseDocument(code, "test.dws")
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	doc := &server.Document{
		URI:        uri,
		Text:       code,
		Version:    1,
		LanguageID: "dwscript",
		Program:    program,
	}

	return doc
}

// setupTestServer creates a test server instance for handler testing.
func setupTestServer(t *testing.T) *server.Server {
	t.Helper()

	srv := server.New()
	SetServer(srv)

	return srv
}

func TestSemanticTokensFull_FirstRequest(t *testing.T) {
	srv := setupTestServer(t)

	// Create a test document with some DWScript code
	code := `var x: Integer = 42;
var y: String = "hello";`

	uri := protocol.URI("file:///test.dws")
	doc := parseCodeToDocument(t, code, string(uri))
	uri = protocol.URI("file:///test.dws")

	// Add document to server store
	srv.Documents().Set(string(uri), doc)

	// Create request params
	params := &protocol.SemanticTokensParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri,
		},
	}

	// Call handler
	result, err := SemanticTokensFull(nil, params)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.ResultID, "ResultID should be set for delta support")
	assert.NotNil(t, result.Data)
	assert.NotEmpty(t, result.Data, "Should have semantic tokens")

	// Verify tokens are multiples of 5 (LSP delta encoding)
	assert.Equal(t, 0, len(result.Data)%5, "Token data should be multiple of 5")

	// Verify resultID is cached
	cache := srv.SemanticTokensCache()
	require.NotNil(t, cache)
	_, found := cache.Retrieve(uri, *result.ResultID)
	assert.True(t, found, "Tokens should be cached with resultID")
}

func TestSemanticTokensFull_EmptyDocument(t *testing.T) {
	srv := setupTestServer(t)

	code := ""
	uri := protocol.URI("file:///test.dws")
	doc := parseCodeToDocument(t, code, string(uri))
	uri = protocol.URI("file:///empty.dws")

	srv.Documents().Set(string(uri), doc)

	params := &protocol.SemanticTokensParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri,
		},
	}

	result, err := SemanticTokensFull(nil, params)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.ResultID)
	assert.NotNil(t, result.Data)
	// Empty document might have 0 tokens
	assert.Equal(t, 0, len(result.Data)%5)
}

func TestSemanticTokensFull_DocumentNotFound(t *testing.T) {
	_ = setupTestServer(t)

	uri := protocol.URI("file:///nonexistent.dws")

	params := &protocol.SemanticTokensParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri,
		},
	}

	result, err := SemanticTokensFull(nil, params)

	require.NoError(t, err)
	assert.Nil(t, result, "Should return nil for nonexistent document")
}

func TestSemanticTokensFullDelta_ValidPreviousResultID(t *testing.T) {
	srv := setupTestServer(t)

	// Create initial document
	code := `var x: Integer = 42;`
	uri := protocol.URI("file:///test.dws")
	doc := parseCodeToDocument(t, code, string(uri))
	uri = protocol.URI("file:///test.dws")

	srv.Documents().Set(string(uri), doc)

	// First request to get initial resultID
	fullParams := &protocol.SemanticTokensParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri,
		},
	}

	fullResult, err := SemanticTokensFull(nil, fullParams)
	require.NoError(t, err)
	require.NotNil(t, fullResult)
	require.NotNil(t, fullResult.ResultID)

	previousResultID := *fullResult.ResultID

	// Modify document slightly
	code2 := `var x: Integer = 43;` // Changed value
	doc2 := parseCodeToDocument(t, code2, string(uri))
	doc2.Version = 2

	srv.Documents().Set(string(uri), doc2)

	// Delta request with previous resultID
	deltaParams := &protocol.SemanticTokensDeltaParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri,
		},
		PreviousResultID: previousResultID,
	}

	deltaResult, err := SemanticTokensFullDelta(nil, deltaParams)

	require.NoError(t, err)
	require.NotNil(t, deltaResult)

	// Result should be either SemanticTokensDelta or SemanticTokens
	switch result := deltaResult.(type) {
	case *protocol.SemanticTokensDelta:
		assert.NotNil(t, result.ResultId, "Delta response should have resultID")
		assert.NotNil(t, result.Edits, "Delta response should have edits")
		assert.NotEqual(t, previousResultID, *result.ResultId, "ResultID should be updated")
	case *protocol.SemanticTokens:
		assert.NotNil(t, result.ResultID, "Full fallback should have resultID")
		assert.NotNil(t, result.Data, "Full fallback should have data")
		assert.NotEqual(t, previousResultID, *result.ResultID, "ResultID should be updated")
	default:
		t.Fatalf("Unexpected result type: %T", deltaResult)
	}
}

func TestSemanticTokensFullDelta_InvalidPreviousResultID(t *testing.T) {
	srv := setupTestServer(t)

	code := `var x: Integer = 42;`
	uri := protocol.URI("file:///test.dws")
	doc := parseCodeToDocument(t, code, string(uri))
	uri = protocol.URI("file:///test.dws")

	srv.Documents().Set(string(uri), doc)

	// Use invalid previous resultID
	invalidResultID := "invalid-result-id-12345"

	deltaParams := &protocol.SemanticTokensDeltaParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri,
		},
		PreviousResultID: invalidResultID,
	}

	deltaResult, err := SemanticTokensFullDelta(nil, deltaParams)

	require.NoError(t, err)
	require.NotNil(t, deltaResult)

	// Should fallback to full response when previous resultID not found
	fullResult, ok := deltaResult.(*protocol.SemanticTokens)
	require.True(t, ok, "Should return full tokens when previous resultID not found")
	assert.NotNil(t, fullResult.ResultID)
	assert.NotNil(t, fullResult.Data)
	assert.NotEmpty(t, fullResult.Data)
}

func TestSemanticTokensFullDelta_NoPreviousResultID(t *testing.T) {
	srv := setupTestServer(t)

	code := `var x: Integer = 42;`
	uri := protocol.URI("file:///test.dws")
	doc := parseCodeToDocument(t, code, string(uri))
	uri = protocol.URI("file:///test.dws")

	srv.Documents().Set(string(uri), doc)

	// Delta request without previous resultID
	deltaParams := &protocol.SemanticTokensDeltaParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri,
		},
		PreviousResultID: "",
	}

	deltaResult, err := SemanticTokensFullDelta(nil, deltaParams)

	require.NoError(t, err)
	require.NotNil(t, deltaResult)

	// Should fallback to full response when no previous resultID
	fullResult, ok := deltaResult.(*protocol.SemanticTokens)
	require.True(t, ok, "Should return full tokens when no previous resultID")
	assert.NotNil(t, fullResult.ResultID)
	assert.NotNil(t, fullResult.Data)
}

func TestSemanticTokensFullDelta_NoChanges(t *testing.T) {
	srv := setupTestServer(t)

	code := `var x: Integer = 42;`
	uri := protocol.URI("file:///test.dws")
	doc := parseCodeToDocument(t, code, string(uri))
	uri = protocol.URI("file:///test.dws")

	srv.Documents().Set(string(uri), doc)

	// First request
	fullParams := &protocol.SemanticTokensParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri,
		},
	}

	fullResult, err := SemanticTokensFull(nil, fullParams)
	require.NoError(t, err)
	require.NotNil(t, fullResult)
	previousResultID := *fullResult.ResultID

	// Request again with same document (no changes)
	deltaParams := &protocol.SemanticTokensDeltaParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri,
		},
		PreviousResultID: previousResultID,
	}

	deltaResult, err := SemanticTokensFullDelta(nil, deltaParams)
	require.NoError(t, err)
	require.NotNil(t, deltaResult)

	// Should return delta with empty edits
	delta, ok := deltaResult.(*protocol.SemanticTokensDelta)
	require.True(t, ok, "Should return delta for no changes")
	assert.NotNil(t, delta.ResultId)
	assert.NotNil(t, delta.Edits)
	assert.Empty(t, delta.Edits, "Should have empty edits for no changes")
}

func TestSemanticTokensFullDelta_LargeChanges(t *testing.T) {
	srv := setupTestServer(t)

	// Initial document
	code1 := `var x: Integer = 1;
var y: Integer = 2;`

	uri := protocol.URI("file:///test.dws")
	doc1 := parseCodeToDocument(t, code1, string(uri))

	srv.Documents().Set(string(uri), doc1)

	// Get initial resultID
	fullParams := &protocol.SemanticTokensParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri,
		},
	}

	fullResult, err := SemanticTokensFull(nil, fullParams)
	require.NoError(t, err)
	require.NotNil(t, fullResult)
	previousResultID := *fullResult.ResultID

	// Completely different document
	code2 := `function Test(): String;
begin
  Result := "completely different";
end;

var a, b, c: String;
var d, e, f: Integer;
var g, h, i: Boolean;`

	doc2 := parseCodeToDocument(t, code2, string(uri))
	doc2.Version = 2

	srv.Documents().Set(string(uri), doc2)

	// Delta request
	deltaParams := &protocol.SemanticTokensDeltaParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri,
		},
		PreviousResultID: previousResultID,
	}

	deltaResult, err := SemanticTokensFullDelta(nil, deltaParams)
	require.NoError(t, err)
	require.NotNil(t, deltaResult)

	// Could be either delta or full depending on threshold
	switch result := deltaResult.(type) {
	case *protocol.SemanticTokensDelta:
		assert.NotNil(t, result.ResultId)
		assert.NotNil(t, result.Edits)
	case *protocol.SemanticTokens:
		assert.NotNil(t, result.ResultID)
		assert.NotNil(t, result.Data)
	default:
		t.Fatalf("Unexpected result type: %T", deltaResult)
	}
}

func TestSemanticTokens_CacheInvalidationOnChange(t *testing.T) {
	srv := setupTestServer(t)

	code := `var x: Integer = 42;`
	uri := protocol.URI("file:///test.dws")
	doc := parseCodeToDocument(t, code, string(uri))
	uri = protocol.URI("file:///test.dws")

	srv.Documents().Set(string(uri), doc)

	// Get initial tokens
	fullParams := &protocol.SemanticTokensParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri,
		},
	}

	fullResult, err := SemanticTokensFull(nil, fullParams)
	require.NoError(t, err)
	require.NotNil(t, fullResult)
	require.NotNil(t, fullResult.ResultID)

	resultID := *fullResult.ResultID

	// Verify cache has the tokens
	cache := srv.SemanticTokensCache()
	_, found := cache.Retrieve(uri, resultID)
	assert.True(t, found, "Tokens should be cached")

	// Simulate document change (normally handled by DidChange handler)
	cache.InvalidateDocument(uri)

	// Verify cache is invalidated
	_, found = cache.Retrieve(uri, resultID)
	assert.False(t, found, "Tokens should be invalidated after document change")
}

func TestSemanticTokens_MultipleDocuments(t *testing.T) {
	srv := setupTestServer(t)

	// Create two documents
	code1 := `var x: Integer = 1;`
	code2 := `var y: String = "test";`

	uri1 := protocol.URI("file:///test1.dws")
	uri2 := protocol.URI("file:///test2.dws")

	doc1 := parseCodeToDocument(t, code1, string(uri1))
	doc2 := parseCodeToDocument(t, code2, string(uri2))

	srv.Documents().Set(string(uri1), doc1)
	srv.Documents().Set(string(uri2), doc2)

	// Get tokens for both documents
	result1, err1 := SemanticTokensFull(nil, &protocol.SemanticTokensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri1},
	})
	result2, err2 := SemanticTokensFull(nil, &protocol.SemanticTokensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri2},
	})

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NotNil(t, result1)
	require.NotNil(t, result2)
	require.NotNil(t, result1.ResultID)
	require.NotNil(t, result2.ResultID)

	// ResultIDs should be different
	assert.NotEqual(t, *result1.ResultID, *result2.ResultID)

	// Invalidating one document shouldn't affect the other
	cache := srv.SemanticTokensCache()
	cache.InvalidateDocument(uri1)

	_, found1 := cache.Retrieve(uri1, *result1.ResultID)
	_, found2 := cache.Retrieve(uri2, *result2.ResultID)

	assert.False(t, found1, "Document 1 should be invalidated")
	assert.True(t, found2, "Document 2 should still be cached")
}

func TestSemanticTokensFullDelta_DocumentNotFound(t *testing.T) {
	_ = setupTestServer(t)

	uri := protocol.URI("file:///nonexistent.dws")
	resultID := "some-result-id"

	params := &protocol.SemanticTokensDeltaParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri,
		},
		PreviousResultID: resultID,
	}

	result, err := SemanticTokensFullDelta(nil, params)

	require.NoError(t, err)
	assert.Nil(t, result, "Should return nil for nonexistent document")
}

func TestSemanticTokens_ComplexDocument(t *testing.T) {
	srv := setupTestServer(t)

	// Complex DWScript code with various token types
	code := `// Comment
type TMyClass = class
  private
    FField: Integer;
  public
    constructor Create;
    function GetValue: Integer;
    procedure SetValue(const AValue: Integer);
end;

var
  gGlobal: String = "global";

function Calculate(x, y: Integer): Integer;
begin
  Result := x + y;
end;

constructor TMyClass.Create;
begin
  FField := 0;
end;

function TMyClass.GetValue: Integer;
begin
  Result := FField;
end;

procedure TMyClass.SetValue(const AValue: Integer);
begin
  FField := AValue;
end;`

	uri := protocol.URI("file:///test.dws")
	doc := parseCodeToDocument(t, code, string(uri))
	uri = protocol.URI("file:///complex.dws")

	srv.Documents().Set(string(uri), doc)

	// Get semantic tokens
	params := &protocol.SemanticTokensParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri,
		},
	}

	result, err := SemanticTokensFull(nil, params)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.ResultID)
	assert.NotNil(t, result.Data)
	assert.NotEmpty(t, result.Data, "Complex document should have many tokens")
	assert.Equal(t, 0, len(result.Data)%5, "Token data should be multiple of 5")

	// Verify cache
	cache := srv.SemanticTokensCache()
	cached, found := cache.Retrieve(uri, *result.ResultID)
	assert.True(t, found)
	assert.NotNil(t, cached)
}
