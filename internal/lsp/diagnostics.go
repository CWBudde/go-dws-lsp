// Package lsp implements LSP protocol handlers.
package lsp

import (
	"log"
	"sort"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// PublishDiagnostics sends diagnostic information to the client for a specific document.
// This notifies the editor about syntax errors, semantic errors, warnings, and hints.
//
// Parameters:
//   - context: The GLSP context for sending notifications
//   - uri: The document URI to publish diagnostics for
//   - diagnostics: List of diagnostics to publish
func PublishDiagnostics(context *glsp.Context, uri string, diagnostics []protocol.Diagnostic) {
	if context == nil || context.Notify == nil {
		log.Println("Warning: Cannot publish diagnostics - context or Notify is nil")
		return
	}

	// Sort diagnostics by position (line, then column) for consistent ordering
	sortDiagnostics(diagnostics)

	// Build the PublishDiagnosticsParams
	params := &protocol.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	}

	log.Printf("Publishing %d diagnostic(s) for %s", len(diagnostics), uri)

	// Send the notification to the client
	context.Notify(protocol.ServerTextDocumentPublishDiagnostics, params)
}

// sortDiagnostics sorts diagnostics by position (line first, then column).
// This ensures diagnostics are presented in a predictable order in the editor.
func sortDiagnostics(diagnostics []protocol.Diagnostic) {
	sort.Slice(diagnostics, func(i, j int) bool {
		// Compare by line first
		if diagnostics[i].Range.Start.Line != diagnostics[j].Range.Start.Line {
			return diagnostics[i].Range.Start.Line < diagnostics[j].Range.Start.Line
		}
		// If same line, compare by column
		return diagnostics[i].Range.Start.Character < diagnostics[j].Range.Start.Character
	})
}
