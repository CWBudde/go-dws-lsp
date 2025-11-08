// Package lsp implements LSP protocol handlers.
package lsp

import (
	"log"

	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// DidChangeConfiguration handles workspace configuration changes from the client.
// This notification is sent from the client to the server when the client's configuration changes.
func DidChangeConfiguration(context *glsp.Context, params *protocol.DidChangeConfigurationParams) error {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in DidChangeConfiguration")
		return nil
	}

	// Parse settings from params.Settings
	// The Settings field is typically a JSON object
	// For DWScript LSP, we expect settings like:
	// {
	//   "go-dws-lsp": {
	//     "maxProblems": 100,
	//     "trace": "off"
	//   }
	// }

	if params.Settings != nil {
		// Try to parse settings as a map
		if settingsMap, ok := params.Settings.(map[string]any); ok {
			// Look for our namespace
			if dwsSettings, ok := settingsMap["go-dws-lsp"].(map[string]any); ok {
				// Update maxProblems if present
				if maxProblems, ok := dwsSettings["maxProblems"].(float64); ok {
					srv.UpdateConfig(func(cfg *server.Config) {
						cfg.MaxProblems = int(maxProblems)
					})
					log.Printf("Configuration updated: maxProblems = %d\n", int(maxProblems))
				}

				// Update trace level if present
				if trace, ok := dwsSettings["trace"].(string); ok {
					srv.UpdateConfig(func(cfg *server.Config) {
						cfg.Trace = trace
					})
					log.Printf("Configuration updated: trace = %s\n", trace)
				}
			}
		}
	}

	return nil
}

// DidChangeWorkspaceFolders handles changes to workspace folders.
// This notification is sent when workspace folders are added or removed.
func DidChangeWorkspaceFolders(context *glsp.Context, params *protocol.DidChangeWorkspaceFoldersParams) error {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in DidChangeWorkspaceFolders")
		return nil
	}

	// Handle added workspace folders
	for _, folder := range params.Event.Added {
		log.Printf("Workspace folder added: %s (%s)\n", folder.Name, folder.URI)
		// TODO: Trigger indexing for new workspace folder
	}

	// Handle removed workspace folders
	for _, folder := range params.Event.Removed {
		log.Printf("Workspace folder removed: %s (%s)\n", folder.Name, folder.URI)
		// TODO: Clear index entries for removed workspace folder
	}

	_ = srv // Placeholder to avoid unused variable error

	return nil
}
