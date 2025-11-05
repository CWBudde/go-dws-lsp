// Package lsp implements LSP protocol handlers.
package lsp

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var (
	// serverInstance holds the global server instance
	// This is set by SetServer and accessed by handlers
	serverInstance interface{}
)

// SetServer sets the global server instance for handlers to access.
func SetServer(srv interface{}) {
	serverInstance = srv
}

// Initialize handles the LSP initialize request.
// This is the first request sent by the client and establishes the server capabilities.
func Initialize(context *glsp.Context, params *protocol.InitializeParams) (interface{}, error) {
	// Extract workspace folders from params
	// workspaceFolders := params.WorkspaceFolders

	// TODO: Store workspace folders in server state
	// TODO: Store client capabilities for feature detection

	// Build server capabilities
	changeKind := protocol.TextDocumentSyncKindIncremental
	trueVal := true
	falseVal := false

	capabilities := protocol.ServerCapabilities{
		// Text document synchronization
		TextDocumentSync: protocol.TextDocumentSyncOptions{
			OpenClose: &trueVal,
			Change:    &changeKind,
			WillSave:  &falseVal,
			Save: &protocol.SaveOptions{
				IncludeText: &falseVal,
			},
		},

		// Hover support
		HoverProvider: &[]bool{true}[0],

		// Go-to definition support
		DefinitionProvider: &[]bool{true}[0],

		// Find references support
		ReferencesProvider: &[]bool{true}[0],

		// Document symbols (outline view)
		DocumentSymbolProvider: &[]bool{true}[0],

		// Workspace symbols (global search)
		WorkspaceSymbolProvider: &[]bool{true}[0],

		// Code completion
		CompletionProvider: &protocol.CompletionOptions{
			TriggerCharacters: []string{".", ":"}, // Member access triggers
			ResolveProvider:   &[]bool{false}[0],   // Don't use lazy resolution for now
		},

		// Signature help
		SignatureHelpProvider: &protocol.SignatureHelpOptions{
			TriggerCharacters:   []string{"(", ","},
			RetriggerCharacters: []string{},
		},

		// Rename support
		RenameProvider: &protocol.RenameOptions{
			PrepareProvider: &[]bool{true}[0],
		},

		// Semantic tokens (semantic highlighting)
		SemanticTokensProvider: &protocol.SemanticTokensOptions{
			Legend: protocol.SemanticTokensLegend{
				TokenTypes: []string{
					"keyword",
					"string",
					"number",
					"comment",
					"variable",
					"parameter",
					"property",
					"function",
					"method",
					"class",
					"interface",
					"enum",
					"type",
					"operator",
				},
				TokenModifiers: []string{
					"declaration",
					"readonly",
					"static",
					"deprecated",
				},
			},
			Full: &[]bool{true}[0], // Support full document semantic tokens
		},

		// Code actions (quick fixes, refactorings)
		CodeActionProvider: &protocol.CodeActionOptions{
			CodeActionKinds: []protocol.CodeActionKind{
				protocol.CodeActionKindQuickFix,
				protocol.CodeActionKindRefactor,
			},
			ResolveProvider: &[]bool{false}[0],
		},

		// Diagnostics (we'll push these, not pull)
		// DiagnosticProvider is not set - we use publishDiagnostics
	}

	// Build and return InitializeResult
	serverVersion := "0.1.0"

	result := protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    "go-dws-lsp",
			Version: &serverVersion,
		},
	}

	return result, nil
}

// Initialized handles the initialized notification from the client.
// This is sent after the initialize response, signaling that the client is ready.
func Initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	// TODO: Trigger workspace indexing here
	// TODO: Start background tasks (if any)

	return nil
}

// Shutdown handles the shutdown request.
// The client sends this to ask the server to shut down gracefully.
func Shutdown(context *glsp.Context) error {
	// TODO: Mark server as shutting down
	// TODO: Clean up resources
	// - Flush caches
	// - Close file handles
	// - Cancel background operations

	return nil
}
