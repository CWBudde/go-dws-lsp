package lsp

import (
	"testing"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/server"
)

func TestInitialize(t *testing.T) {
	// Create test parameters
	clientName := "test-client"
	clientVersion := "1.0.0"
	rootURI := "file:///test/workspace"

	params := &protocol.InitializeParams{
		ProcessID: nil,
		ClientInfo: &struct {
			Name    string  `json:"name"`
			Version *string `json:"version,omitempty"`
		}{
			Name:    clientName,
			Version: &clientVersion,
		},
		RootURI: &rootURI,
		Capabilities: protocol.ClientCapabilities{
			// Add basic client capabilities
		},
	}

	// Create a mock context (GLSP context is complex, so we'll use nil for now)
	// In a real scenario, you'd create a proper mock
	ctx := &glsp.Context{}

	// Call Initialize handler
	result, err := Initialize(ctx, params)
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	// Assert result is not nil
	if result == nil {
		t.Fatal("Initialize returned nil result")
	}

	// Type assert to InitializeResult
	initResult, ok := result.(protocol.InitializeResult)
	if !ok {
		t.Fatalf("Initialize returned wrong type: %T", result)
	}

	// Verify ServerInfo
	if initResult.ServerInfo == nil {
		t.Error("ServerInfo is nil")
	} else {
		if initResult.ServerInfo.Name != "go-dws-lsp" {
			t.Errorf("ServerInfo.Name = %q, want %q", initResult.ServerInfo.Name, "go-dws-lsp")
		}
		if initResult.ServerInfo.Version == nil {
			t.Error("ServerInfo.Version is nil")
		} else if *initResult.ServerInfo.Version != "0.1.0" {
			t.Errorf("ServerInfo.Version = %q, want %q", *initResult.ServerInfo.Version, "0.1.0")
		}
	}

	// Verify Capabilities
	caps := initResult.Capabilities

	// Test TextDocumentSync
	if syncOpts, ok := caps.TextDocumentSync.(protocol.TextDocumentSyncOptions); ok {
		if syncOpts.OpenClose == nil || !*syncOpts.OpenClose {
			t.Error("TextDocumentSync.OpenClose should be true")
		}
		if syncOpts.Change == nil || *syncOpts.Change != protocol.TextDocumentSyncKindIncremental {
			t.Error("TextDocumentSync.Change should be Incremental")
		}
	} else {
		t.Errorf("TextDocumentSync has wrong type: %T", caps.TextDocumentSync)
	}

	// Test HoverProvider
	if caps.HoverProvider == nil {
		t.Error("HoverProvider should be set")
	}

	// Test DefinitionProvider
	if caps.DefinitionProvider == nil {
		t.Error("DefinitionProvider should be set")
	}

	// Test ReferencesProvider
	if caps.ReferencesProvider == nil {
		t.Error("ReferencesProvider should be set")
	}

	// Test DocumentSymbolProvider
	if caps.DocumentSymbolProvider == nil {
		t.Error("DocumentSymbolProvider should be set")
	}

	// Test WorkspaceSymbolProvider
	if caps.WorkspaceSymbolProvider == nil {
		t.Error("WorkspaceSymbolProvider should be set")
	}

	// Test CompletionProvider
	if caps.CompletionProvider == nil {
		t.Error("CompletionProvider should be set")
	} else {
		triggers := caps.CompletionProvider.TriggerCharacters
		if len(triggers) == 0 {
			t.Error("CompletionProvider should have trigger characters")
		}
		// Check for expected triggers
		hasDot := false
		for _, trigger := range triggers {
			if trigger == "." {
				hasDot = true
				break
			}
		}
		if !hasDot {
			t.Error("CompletionProvider should have '.' as trigger character")
		}
	}

	// Test SignatureHelpProvider
	if caps.SignatureHelpProvider == nil {
		t.Error("SignatureHelpProvider should be set")
	} else {
		triggers := caps.SignatureHelpProvider.TriggerCharacters
		if len(triggers) == 0 {
			t.Error("SignatureHelpProvider should have trigger characters")
		}
		// Check for expected triggers
		hasParen := false
		for _, trigger := range triggers {
			if trigger == "(" {
				hasParen = true
				break
			}
		}
		if !hasParen {
			t.Error("SignatureHelpProvider should have '(' as trigger character")
		}
	}

	// Test RenameProvider
	if caps.RenameProvider == nil {
		t.Error("RenameProvider should be set")
	}

	// Test SemanticTokensProvider
	if caps.SemanticTokensProvider == nil {
		t.Error("SemanticTokensProvider should be set")
	} else {
		// Type assert to SemanticTokensOptions
		if semOpts, ok := caps.SemanticTokensProvider.(*protocol.SemanticTokensOptions); ok {
			legend := semOpts.Legend
			if len(legend.TokenTypes) == 0 {
				t.Error("SemanticTokensProvider should have token types in legend")
			}
			if len(legend.TokenModifiers) == 0 {
				t.Error("SemanticTokensProvider should have token modifiers in legend")
			}
		} else {
			t.Errorf("SemanticTokensProvider has wrong type: %T", caps.SemanticTokensProvider)
		}
	}

	// Test CodeActionProvider
	if caps.CodeActionProvider == nil {
		t.Error("CodeActionProvider should be set")
	}
}

func TestInitialized(t *testing.T) {
	// Create test parameters
	params := &protocol.InitializedParams{}

	// Create mock context
	ctx := &glsp.Context{}

	// Call Initialized handler
	err := Initialized(ctx, params)
	if err != nil {
		t.Fatalf("Initialized returned error: %v", err)
	}

	// Initialized should succeed without error
	// Currently it's a no-op, so just verify it doesn't crash
}

func TestShutdown(t *testing.T) {
	// Create mock context
	ctx := &glsp.Context{}

	// Create a server instance and set it
	srv := server.New()
	SetServer(srv)

	// Add some data to caches and stores
	doc := &server.Document{
		URI:     "file:///test.dws",
		Text:    "var x = 1;",
		Version: 1,
	}
	srv.Documents().Set(doc.URI, doc)

	// Add some cached items
	srv.CompletionCache().SetCachedItems(doc.URI, 1, &server.CachedCompletionItems{
		Keywords: []protocol.CompletionItem{{Label: "var"}},
	})

	// Verify data exists before shutdown
	if _, ok := srv.Documents().Get(doc.URI); !ok {
		t.Fatal("Document should exist before shutdown")
	}

	// Call Shutdown handler
	err := Shutdown(ctx)
	if err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}

	// Verify shutdown flag is set
	if !srv.IsShuttingDown() {
		t.Error("Server should be marked as shutting down")
	}

	// Verify resources are cleaned up
	if len(srv.Documents().List()) != 0 {
		t.Error("Document store should be cleared after shutdown")
	}

	if srv.CompletionCache().GetCachedItems(doc.URI, 1) != nil {
		t.Error("Completion cache should be cleared after shutdown")
	}

	if srv.SemanticTokensCache().Size() != 0 {
		t.Error("Semantic tokens cache should be cleared after shutdown")
	}
}

