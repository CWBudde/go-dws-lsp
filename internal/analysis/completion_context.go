// Package analysis provides code analysis utilities for the LSP server.
package analysis

import (
	"strings"
	"unicode"

	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// CompletionContextType represents the type of completion context.
type CompletionContextType int

const (
	// CompletionContextGeneral represents general code completion (identifiers, keywords, etc.)
	CompletionContextGeneral CompletionContextType = iota

	// CompletionContextMember represents member access completion (after a dot).
	CompletionContextMember

	// CompletionContextKeyword represents keyword-specific completion.
	CompletionContextKeyword
)

// CompletionContext holds information about the completion request context.
type CompletionContext struct {
	// Type is the type of completion context
	Type CompletionContextType

	// Scope is the AST node representing the current scope
	Scope ast.Node

	// ParentIdentifier is the identifier before a dot (for member access)
	ParentIdentifier string

	// Line is the 0-based line number (LSP convention)
	Line int

	// Character is the 0-based character offset (LSP convention)
	Character int

	// Prefix is the partial identifier the user has typed (for filtering)
	// Task 9.18: Used for early prefix filtering to reduce processing
	Prefix string
}

// DetermineContext analyzes the document and position to determine the completion context.
func DetermineContext(doc *server.Document, line, character int) (*CompletionContext, error) {
	ctx := &CompletionContext{
		Type:      CompletionContextGeneral,
		Line:      line,
		Character: character,
	}

	// Get the text before the cursor position
	textBeforeCursor := getTextBeforeCursor(doc.Text, line, character)

	// Check if we're inside a comment
	if isInsideComment(textBeforeCursor, doc.Text, line, character) {
		// Return nil to indicate no completion should be provided
		return nil, nil
	}

	// Check if we're inside a string literal
	if isInsideString(textBeforeCursor) {
		// Return nil to indicate no completion should be provided
		return nil, nil
	}

	// Check for member access pattern (identifier followed by dot)
	if parentIdent := extractParentIdentifier(textBeforeCursor); parentIdent != "" {
		ctx.Type = CompletionContextMember
		ctx.ParentIdentifier = parentIdent
		// For member access, extract the prefix after the dot
		ctx.Prefix = extractPartialIdentifier(textBeforeCursor)
	} else {
		// For general completion, extract the partial identifier being typed
		ctx.Prefix = extractPartialIdentifier(textBeforeCursor)
	}

	// Determine current scope from AST
	if doc.Program != nil && doc.Program.AST() != nil {
		// Convert LSP position (0-based) to AST position (1-based)
		astLine := line + 1
		astColumn := character + 1
		ctx.Scope = findScopeAtPosition(doc.Program.AST(), astLine, astColumn)
	}

	return ctx, nil
}

// getTextBeforeCursor extracts the text from the beginning of the document up to the cursor position.
func getTextBeforeCursor(text string, line, character int) string {
	lines := strings.Split(text, "\n")

	if line >= len(lines) {
		return text
	}

	// Get all lines before the cursor line
	beforeLines := make([]string, 0, line)
	for i := range line {
		beforeLines = append(beforeLines, lines[i])
	}

	// Get the portion of the current line up to the cursor
	currentLine := lines[line]
	if character > len(currentLine) {
		character = len(currentLine)
	}

	beforeLines = append(beforeLines, currentLine[:character])

	return strings.Join(beforeLines, "\n")
}

// isInsideComment checks if the cursor position is inside a comment.
func isInsideComment(textBeforeCursor, fullText string, line, character int) bool {
	// Check for single-line comment
	lines := strings.Split(textBeforeCursor, "\n")
	if len(lines) > 0 {
		lastLine := lines[len(lines)-1]
		// Check if there's a // comment before the cursor
		if commentIdx := strings.Index(lastLine, "//"); commentIdx != -1 && commentIdx < len(lastLine) {
			return true
		}
	}

	// Check for multi-line comment (* ... *) or { ... }
	// Count opening and closing comment markers
	openParen := strings.Count(textBeforeCursor, "(*")
	closeParen := strings.Count(textBeforeCursor, "*)")
	openBrace := strings.Count(textBeforeCursor, "{")
	closeBrace := strings.Count(textBeforeCursor, "}")

	// If there are more opening than closing markers, we're inside a comment
	if openParen > closeParen || openBrace > closeBrace {
		return true
	}

	return false
}

// isInsideString checks if the cursor position is inside a string literal.
func isInsideString(textBeforeCursor string) bool {
	// Count single quotes that are not escaped
	// In DWScript, strings use single quotes
	quoteCount := 0

	for i := 0; i < len(textBeforeCursor); i++ {
		if textBeforeCursor[i] == '\'' {
			// Check if it's escaped (doubled single quote in DWScript)
			if i+1 < len(textBeforeCursor) && textBeforeCursor[i+1] == '\'' {
				i++ // Skip the next quote
				continue
			}

			quoteCount++
		}
	}

	// If odd number of quotes, we're inside a string
	return quoteCount%2 == 1
}

// extractParentIdentifier extracts the identifier before a dot (for member access).
// For example, if textBeforeCursor ends with "myObject.", it returns "myObject".
func extractParentIdentifier(textBeforeCursor string) string {
	// Check if the text ends with a dot (with possible trailing whitespace)
	trimmed := strings.TrimRight(textBeforeCursor, " \t\r\n")
	if !strings.HasSuffix(trimmed, ".") {
		return ""
	}

	// Remove the trailing dot
	withoutDot := strings.TrimSuffix(trimmed, ".")

	// Trim whitespace before the dot
	withoutDot = strings.TrimRight(withoutDot, " \t\r\n")

	// Extract the identifier before the dot
	// Work backwards to find the start of the identifier
	i := len(withoutDot) - 1
	for i >= 0 {
		ch := rune(withoutDot[i])
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			break
		}

		i--
	}

	identifier := withoutDot[i+1:]
	if identifier == "" {
		return ""
	}

	return identifier
}

// findScopeAtPosition finds the most specific scope (AST node) that contains the given position.
func findScopeAtPosition(node ast.Node, line, column int) ast.Node {
	if node == nil {
		return nil
	}

	// Check if the position is within this node's range
	if !isPositionInNode(node, line, column) {
		return nil
	}

	// Try to find a more specific child node that contains the position
	// This is a simplified implementation - a full implementation would traverse
	// all child nodes recursively
	switch n := node.(type) {
	case *ast.Program:
		// Program has Statements, not Declarations
		for _, stmt := range n.Statements {
			if childScope := findScopeAtPosition(stmt, line, column); childScope != nil {
				return childScope
			}
		}
		// If no more specific scope found, return the program
		return node

	case *ast.FunctionDecl:
		if n.Body != nil {
			if childScope := findScopeAtPosition(n.Body, line, column); childScope != nil {
				return childScope
			}
		}

		return node

	case *ast.ClassDecl:
		// Check methods within the class
		for _, method := range n.Methods {
			if childScope := findScopeAtPosition(method, line, column); childScope != nil {
				return childScope
			}
		}
		// Check constructor and destructor
		if n.Constructor != nil {
			if childScope := findScopeAtPosition(n.Constructor, line, column); childScope != nil {
				return childScope
			}
		}

		if n.Destructor != nil {
			if childScope := findScopeAtPosition(n.Destructor, line, column); childScope != nil {
				return childScope
			}
		}

		return node

	case *ast.BlockStatement:
		for _, stmt := range n.Statements {
			if childScope := findScopeAtPosition(stmt, line, column); childScope != nil {
				return childScope
			}
		}

		return node

	default:
		// For other node types, this is the most specific scope we can find
		return node
	}
}

// isPositionInNode checks if the given position is within the node's range.
func isPositionInNode(node ast.Node, line, column int) bool {
	if node == nil {
		return false
	}

	pos := node.Pos()
	end := node.End()

	// Check if position is within the node's range
	if line < pos.Line || line > end.Line {
		return false
	}

	if line == pos.Line && column < pos.Column {
		return false
	}

	if line == end.Line && column > end.Column {
		return false
	}

	return true
}

// extractPartialIdentifier extracts the partial identifier being typed at the cursor position.
// Task 9.18: Used for early prefix filtering to reduce processing.
// Example: "var myVa" -> "myVa"
// Example: "person." -> "" (no prefix after dot)
// Example: "person.Na" -> "Na".
func extractPartialIdentifier(textBeforeCursor string) string {
	// Get the text on the current line only (faster than processing full text)
	lines := strings.Split(textBeforeCursor, "\n")
	if len(lines) == 0 {
		return ""
	}

	currentLine := lines[len(lines)-1]

	// Work backwards from the end to find the start of the identifier
	i := len(currentLine) - 1

	// Skip trailing whitespace
	for i >= 0 && unicode.IsSpace(rune(currentLine[i])) {
		i--
	}

	// If we hit a dot, return empty (we're right after a dot for member access)
	if i >= 0 && currentLine[i] == '.' {
		return ""
	}

	// Find the start of the identifier
	end := i + 1
	for i >= 0 {
		ch := rune(currentLine[i])
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			break
		}

		i--
	}

	start := i + 1
	if start >= end {
		return ""
	}

	return currentLine[start:end]
}
