// Package lsp implements LSP protocol handlers.
package lsp

import (
	"log"
	"regexp"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// CodeAction handles the textDocument/codeAction request.
// This provides quick fixes and refactoring actions for diagnostics and code.
func CodeAction(context *glsp.Context, params *protocol.CodeActionParams) (any, error) {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in CodeAction")
		return nil, nil
	}

	// Extract document URI, range, and context from params
	uri := params.TextDocument.URI
	selectedRange := params.Range
	actionContext := params.Context

	log.Printf("CodeAction request at %s range (%d:%d)-(%d:%d)\n",
		uri,
		selectedRange.Start.Line, selectedRange.Start.Character,
		selectedRange.End.Line, selectedRange.End.Character)

	// Get diagnostics from params.Context.Diagnostics
	diagnostics := actionContext.Diagnostics
	log.Printf("CodeAction context has %d diagnostics\n", len(diagnostics))

	// Retrieve document from DocumentStore
	doc, exists := srv.Documents().Get(uri)
	if !exists {
		log.Printf("Document not found for code action: %s\n", uri)
		return nil, nil
	}

	// Check if document has AST available
	if doc.Program == nil {
		log.Printf("No AST available for code action (document has parse errors): %s\n", uri)
		// Even without AST, we can still provide some code actions based on diagnostics
		// For now, return empty array
		return []protocol.CodeAction{}, nil
	}

	// Get AST from Program
	programAST := doc.Program.AST()
	if programAST == nil {
		log.Printf("AST is nil for document: %s\n", uri)
		return []protocol.CodeAction{}, nil
	}

	// Call helper functions to generate code actions
	var actions []protocol.CodeAction

	// Generate quick fixes from diagnostics
	for _, diagnostic := range diagnostics {
		quickFixes, err := GenerateQuickFixes(diagnostic, doc, uri)
		if err != nil {
			log.Printf("Error generating quick fixes: %v\n", err)
			continue
		}
		actions = append(actions, quickFixes...)
	}

	// TODO: Generate code actions based on:
	// 2. Code context (refactoring actions)
	// 3. Selected range (extract method, etc.)

	log.Printf("Returning %d code actions\n", len(actions))
	return actions, nil
}

// GenerateQuickFixes generates quick fix code actions for a diagnostic.
func GenerateQuickFixes(diagnostic protocol.Diagnostic, doc *server.Document, uri string) ([]protocol.CodeAction, error) {
	var actions []protocol.CodeAction

	// Check if diagnostic is for undeclared identifier
	if isUndeclaredIdentifier(diagnostic) {
		identifierName := extractIdentifierName(diagnostic)
		if identifierName != "" {
			log.Printf("Generating quick fixes for undeclared identifier: %s\n", identifierName)

			// Create "Declare variable" quick fix
			action := createDeclareVariableAction(diagnostic, identifierName, uri)
			if action != nil {
				actions = append(actions, *action)
			}
		}
	}

	return actions, nil
}

// isUndeclaredIdentifier checks if a diagnostic indicates an undeclared identifier error.
func isUndeclaredIdentifier(diagnostic protocol.Diagnostic) bool {
	message := strings.ToLower(diagnostic.Message)

	// Check for common undeclared identifier patterns
	patterns := []string{
		"undeclared identifier",
		"unknown identifier",
		"identifier not found",
		"undefined identifier",
		"unknown symbol",
		"undeclared",
	}

	for _, pattern := range patterns {
		if strings.Contains(message, pattern) {
			return true
		}
	}

	// Check error code if available
	if diagnostic.Code != nil {
		code := diagnostic.Code.Value
		if code == "E_UNDECLARED" || code == "E_UNKNOWN_IDENTIFIER" {
			return true
		}
	}

	return false
}

// extractIdentifierName extracts the identifier name from a diagnostic message.
// It looks for patterns like "undeclared identifier 'x'" or "unknown identifier: x"
func extractIdentifierName(diagnostic protocol.Diagnostic) string {
	message := diagnostic.Message

	// Try various regex patterns to extract the identifier
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`['"]([a-zA-Z_][a-zA-Z0-9_]*)['"]`),           // 'identifier' or "identifier"
		regexp.MustCompile(`identifier:\s*([a-zA-Z_][a-zA-Z0-9_]*)`),     // identifier: name
		regexp.MustCompile(`identifier\s+([a-zA-Z_][a-zA-Z0-9_]*)`),      // identifier name
		regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\s+not\s+found`),   // name not found
		regexp.MustCompile(`unknown\s+([a-zA-Z_][a-zA-Z0-9_]*)`),         // unknown name
	}

	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(message)
		if len(matches) >= 2 {
			return matches[1]
		}
	}

	return ""
}

// createDeclareVariableAction creates a quick fix action to declare an undeclared variable.
func createDeclareVariableAction(diagnostic protocol.Diagnostic, identifierName string, uri string) *protocol.CodeAction {
	title := "Declare variable '" + identifierName + "'"

	// For now, we'll create a placeholder action
	// Tasks 13.5 and 13.6 will implement the actual WorkspaceEdit
	// to insert the variable declaration at the appropriate location

	action := protocol.CodeAction{
		Title: title,
		Kind:  stringPtr(string(protocol.CodeActionKindQuickFix)),
		Diagnostics: []protocol.Diagnostic{diagnostic},
		// TODO (Task 13.5): Add WorkspaceEdit to insert variable declaration
		// TODO (Task 13.6): Determine insertion location (function top or global)
	}

	log.Printf("Created quick fix: %s\n", title)
	return &action
}

// stringPtr is a helper function to create a pointer to a string.
func stringPtr(s string) *string {
	return &s
}
