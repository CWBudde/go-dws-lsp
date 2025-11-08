// Package lsp implements LSP protocol handlers.
package lsp

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/CWBudde/go-dws-lsp/internal/workspace"
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
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if ok && srv != nil {
		// Store client capabilities
		srv.SetClientCapabilities(&params.Capabilities)
		log.Printf("Client snippet support: %v\n", srv.SupportsSnippets())

		// Extract and store workspace folders from params
		if params.WorkspaceFolders != nil && len(params.WorkspaceFolders) > 0 {
			folders := make([]string, 0, len(params.WorkspaceFolders))
			for _, wf := range params.WorkspaceFolders {
				// Convert URI to path
				path := uriToPath(wf.URI)
				if path != "" {
					folders = append(folders, path)
				}
			}
			srv.SetWorkspaceFolders(folders)
			log.Printf("Stored %d workspace folders\n", len(folders))
		} else if params.RootURI != nil && *params.RootURI != "" {
			// Fallback to RootURI if WorkspaceFolders not provided
			path := uriToPath(*params.RootURI)
			if path != "" {
				srv.SetWorkspaceFolders([]string{path})
				log.Printf("Stored workspace root: %s\n", path)
			}
		}
	}

	// Build server capabilities
	changeKind := protocol.TextDocumentSyncKindIncremental
	trueVal := true
	falseVal := false

	// Get semantic tokens legend from server
	var semanticTokensProvider *protocol.SemanticTokensOptions
	if ok && srv != nil {
		legend := srv.SemanticTokensLegend()
		if legend != nil {
			semanticTokensProvider = &protocol.SemanticTokensOptions{
				Legend: legend.ToProtocolLegend(),
				Full: map[string]interface{}{
					"delta": true, // Support delta incremental updates
				},
			}
		}
	}

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
		// Uses the legend from the server instance for consistency
		SemanticTokensProvider: semanticTokensProvider,

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
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in Initialized")
		return nil
	}

	// Get workspace folders
	folders := srv.GetWorkspaceFolders()
	if len(folders) == 0 {
		log.Println("No workspace folders to index")
		return nil
	}

	// Convert folder paths to WorkspaceFolder structs
	workspaceFolders := make([]protocol.WorkspaceFolder, 0, len(folders))
	for _, folder := range folders {
		workspaceFolders = append(workspaceFolders, protocol.WorkspaceFolder{
			URI:  pathToURI(folder),
			Name: folder,
		})
	}

	// Start workspace indexing in background
	log.Printf("Starting workspace indexing for %d folders\n", len(workspaceFolders))
	workspace.IndexWorkspaceAsync(srv.WorkspaceIndex(), workspaceFolders)

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

// uriToPath converts a URI to a file system path.
func uriToPath(uri string) string {
	// Handle file:// URIs
	if strings.HasPrefix(uri, "file://") {
		path := strings.TrimPrefix(uri, "file://")
		// On Windows, URIs are like file:///C:/path, so we need to handle the leading slash
		if len(path) > 2 && path[0] == '/' && path[2] == ':' {
			path = path[1:] // Remove leading slash for Windows paths
		}
		return path
	}
	return uri
}

// pathToURI converts a file system path to a URI.
func pathToURI(path string) string {
	// Normalize path separators to forward slashes
	path = filepath.ToSlash(path)

	// On Windows, prepend an extra slash
	if len(path) > 1 && path[1] == ':' {
		return "file:///" + path
	}

	return "file://" + path
}
