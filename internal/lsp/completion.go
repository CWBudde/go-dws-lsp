// Package lsp implements LSP protocol handlers.
package lsp

import (
	"log"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/analysis"
	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// Completion handles the textDocument/completion request.
// This provides intelligent code completion suggestions.
func Completion(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in Completion")
		return nil, nil
	}

	// Extract document URI and position from params
	uri := params.TextDocument.URI
	position := params.Position

	log.Printf("Completion request at %s line %d, character %d\n",
		uri, position.Line, position.Character)

	// Retrieve document from DocumentStore
	doc, exists := srv.Documents().Get(uri)
	if !exists {
		log.Printf("Document not found for completion: %s\n", uri)
		return nil, nil
	}

	// Check if document and AST are available
	if doc.Program == nil {
		log.Printf("No AST available for completion (document has parse errors): %s\n", uri)
		// Return empty completion list instead of nil to indicate completion is supported
		return &protocol.CompletionList{
			IsIncomplete: false,
			Items:        []protocol.CompletionItem{},
		}, nil
	}

	// Get AST from Program
	programAST := doc.Program.AST()
	if programAST == nil {
		log.Printf("AST is nil for document: %s\n", uri)
		return &protocol.CompletionList{
			IsIncomplete: false,
			Items:        []protocol.CompletionItem{},
		}, nil
	}

	// Task 9.2: Determine completion context from cursor position
	completionContext, err := analysis.DetermineContext(doc, int(position.Line), int(position.Character))
	if err != nil {
		log.Printf("Error determining completion context: %v\n", err)
		return &protocol.CompletionList{
			IsIncomplete: false,
			Items:        []protocol.CompletionItem{},
		}, nil
	}

	// If context is nil, we're in a location where completion shouldn't be provided
	// (e.g., inside a comment or string)
	if completionContext == nil {
		log.Println("Completion suppressed (inside comment or string)")
		return &protocol.CompletionList{
			IsIncomplete: false,
			Items:        []protocol.CompletionItem{},
		}, nil
	}

	// Task 9.3: Detect trigger characters (dot for member access)
	if params.Context != nil {
		// Check if completion was triggered by a trigger character
		if params.Context.TriggerKind == protocol.CompletionTriggerKindTriggerCharacter {
			// Check if the trigger character is a dot
			if params.Context.TriggerCharacter != nil && *params.Context.TriggerCharacter == "." {
				log.Println("Completion triggered by dot (member access)")
				// The context type should already be set to Member by DetermineContext
				// But we can verify and override if needed
				if completionContext.Type != analysis.CompletionContextMember {
					log.Printf("Warning: trigger character is dot but context type is %d, overriding to Member",
						completionContext.Type)
					completionContext.Type = analysis.CompletionContextMember
				}
			} else if params.Context.TriggerCharacter != nil {
				log.Printf("Completion triggered by character: %s\n", *params.Context.TriggerCharacter)
			}
		}
	}

	// Log the completion context for debugging
	log.Printf("Completion context: type=%d, parent=%s\n",
		completionContext.Type, completionContext.ParentIdentifier)

	// Task 9.4: Handle member access completion
	var completionList *protocol.CompletionList

	if completionContext.Type == analysis.CompletionContextMember {
		// Member access completion: resolve the type of the parent identifier
		if completionContext.ParentIdentifier != "" {
			log.Printf("Resolving type of parent identifier: %s", completionContext.ParentIdentifier)

			typeInfo, err := analysis.ResolveMemberType(doc, completionContext.ParentIdentifier,
				int(position.Line), int(position.Character))

			if err != nil {
				log.Printf("Error resolving member type: %v", err)
			} else if typeInfo != nil {
				log.Printf("Resolved parent type: %s (built-in: %v)", typeInfo.TypeName, typeInfo.IsBuiltIn)
				// TODO: Task 9.5-9.6 - Get members of the resolved type
				// For now, we just log the type
			} else {
				log.Printf("Could not determine type of '%s'", completionContext.ParentIdentifier)
			}
		}

		// Return empty list for now - will be populated in tasks 9.5-9.6
		completionList = &protocol.CompletionList{
			IsIncomplete: false,
			Items:        []protocol.CompletionItem{},
		}
	} else {
		// TODO: Task 9.7+ - Handle general scope completion
		completionList = &protocol.CompletionList{
			IsIncomplete: false,
			Items:        []protocol.CompletionItem{},
		}
	}

	log.Printf("Returning %d completion items\n", len(completionList.Items))

	return completionList, nil
}
