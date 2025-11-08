// Package lsp implements LSP protocol handlers.
package lsp

import (
	"log"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/CWBudde/go-dws-lsp/internal/workspace"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

const (
	defaultTypeVariant = "Variant"
	keywordBegin       = "begin"
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

	// Generate source actions (refactoring actions)
	sourceActions := GenerateSourceActions(doc, uri, actionContext)
	actions = append(actions, sourceActions...)

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

			// Check if the identifier is used as a function call
			if isFunctionCall(identifierName, diagnostic, doc) {
				// Create "Declare function" quick fix
				action := createDeclareFunctionAction(diagnostic, identifierName, uri, doc)
				if action != nil {
					actions = append(actions, *action)
				}
			} else {
				// Create "Declare variable" quick fix
				action := createDeclareVariableAction(diagnostic, identifierName, uri, doc)
				if action != nil {
					actions = append(actions, *action)
				}
			}
		}
	}

	// Check if diagnostic is for missing semicolon
	if isMissingSemicolon(diagnostic) {
		log.Printf("Generating quick fix for missing semicolon at line %d\n", diagnostic.Range.Start.Line)

		// Create "Insert missing semicolon" quick fix
		action := createInsertSemicolonAction(diagnostic, uri)
		if action != nil {
			actions = append(actions, *action)
		}
	}

	// Check if diagnostic is for unused variable
	if isUnusedVariable(diagnostic) {
		variableName := extractVariableName(diagnostic)
		if variableName != "" {
			log.Printf("Generating quick fixes for unused variable: %s\n", variableName)

			// Create "Remove unused variable" quick fix
			removeAction := createRemoveVariableAction(diagnostic, variableName, uri, doc)
			if removeAction != nil {
				actions = append(actions, *removeAction)
			}

			// Create "Prefix with underscore" quick fix
			prefixAction := createPrefixUnderscoreAction(diagnostic, variableName, uri, doc)
			if prefixAction != nil {
				actions = append(actions, *prefixAction)
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

// isMissingSemicolon checks if a diagnostic indicates a missing semicolon error.
func isMissingSemicolon(diagnostic protocol.Diagnostic) bool {
	message := strings.ToLower(diagnostic.Message)

	// Check for common missing semicolon patterns
	patterns := []string{
		"missing semicolon",
		"expected ';'",
		"expected semicolon",
		"; expected",
		"semicolon expected",
	}

	for _, pattern := range patterns {
		if strings.Contains(message, pattern) {
			return true
		}
	}

	// Check error code if available
	if diagnostic.Code != nil {
		code := diagnostic.Code.Value
		if code == "E_MISSING_SEMICOLON" || code == "E_SEMICOLON_EXPECTED" {
			return true
		}
	}

	return false
}

// isUnusedVariable checks if a diagnostic indicates an unused variable warning.
func isUnusedVariable(diagnostic protocol.Diagnostic) bool {
	message := strings.ToLower(diagnostic.Message)

	// Check for common unused variable patterns
	patterns := []string{
		"unused variable",
		"variable not used",
		"variable declared but not used",
		"unused local variable",
		"variable is declared but never used",
	}

	for _, pattern := range patterns {
		if strings.Contains(message, pattern) {
			return true
		}
	}

	// Check error code if available
	if diagnostic.Code != nil {
		code := diagnostic.Code.Value
		if code == "W_UNUSED_VAR" || code == "W_UNUSED_VARIABLE" {
			return true
		}
	}

	return false
}

// extractIdentifierName extracts the identifier name from a diagnostic message.
// It looks for patterns like "undeclared identifier 'x'" or "unknown identifier: x".
func extractIdentifierName(diagnostic protocol.Diagnostic) string {
	message := diagnostic.Message

	// Try various regex patterns to extract the identifier
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`['"]([a-zA-Z_][a-zA-Z0-9_]*)['"]`),         // 'identifier' or "identifier"
		regexp.MustCompile(`identifier:\s*([a-zA-Z_][a-zA-Z0-9_]*)`),   // identifier: name
		regexp.MustCompile(`identifier\s+([a-zA-Z_][a-zA-Z0-9_]*)`),    // identifier name
		regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\s+not\s+found`), // name not found
		regexp.MustCompile(`unknown\s+([a-zA-Z_][a-zA-Z0-9_]*)`),       // unknown name
	}

	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(message)
		if len(matches) >= 2 {
			return matches[1]
		}
	}

	return ""
}

// extractVariableName extracts the variable name from an unused variable diagnostic message.
// It looks for patterns like "unused variable 'x'" or "variable x not used".
func extractVariableName(diagnostic protocol.Diagnostic) string {
	message := diagnostic.Message

	// Try various regex patterns to extract the variable name
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`['\"]([a-zA-Z_][a-zA-Z0-9_]*)['\"]`),             // 'varname' or "varname"
		regexp.MustCompile(`variable\s+([a-zA-Z_][a-zA-Z0-9_]*)`),            // variable name
		regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*)\s+(?:not used|unused)`), // name not used/unused
		regexp.MustCompile(`unused:\s*([a-zA-Z_][a-zA-Z0-9_]*)`),             // unused: name
	}

	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(message)
		if len(matches) >= 2 {
			return matches[1]
		}
	}

	return ""
}

// isFunctionCall checks if an undeclared identifier is used as a function call.
// It examines the document text and AST to determine if the identifier is followed by parentheses.
func isFunctionCall(identifierName string, diagnostic protocol.Diagnostic, doc *server.Document) bool {
	if doc.Text == "" {
		return false
	}

	// Get the line where the error occurred
	lines := strings.Split(doc.Text, "\n")

	lineNum := int(diagnostic.Range.Start.Line)
	if lineNum >= len(lines) {
		return false
	}

	line := lines[lineNum]

	// Look for the identifier followed by an opening parenthesis
	// Pattern: identifier(
	pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(identifierName) + `\s*\(`)

	return pattern.MatchString(line)
}

// extractCallArguments extracts the argument list from a function call in the source text.
// Returns a slice of argument expressions (as strings) and the number of arguments.
func extractCallArguments(identifierName string, diagnostic protocol.Diagnostic, doc *server.Document) []string {
	if doc.Text == "" {
		return nil
	}

	lines := strings.Split(doc.Text, "\n")

	lineNum := int(diagnostic.Range.Start.Line)
	if lineNum >= len(lines) {
		return nil
	}

	line := lines[lineNum]

	// Find the function call pattern: identifier(args)
	// This is a simplified approach that looks for the parentheses
	pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(identifierName) + `\s*\((.*?)\)`)
	matches := pattern.FindStringSubmatch(line)

	if len(matches) < 2 {
		// No arguments found, return empty slice
		return []string{}
	}

	argsText := matches[1]
	if strings.TrimSpace(argsText) == "" {
		return []string{}
	}

	// Split by comma (simplified - doesn't handle nested calls)
	args := strings.Split(argsText, ",")

	result := make([]string, 0, len(args))
	for _, arg := range args {
		result = append(result, strings.TrimSpace(arg))
	}

	return result
}

// inferParameterType infers a parameter type from an argument expression.
func inferParameterType(argExpr string) string {
	argExpr = strings.TrimSpace(argExpr)

	// Check for integer literal
	if matched, _ := regexp.MatchString(`^-?\d+$`, argExpr); matched {
		return "Integer"
	}

	// Check for float literal
	if matched, _ := regexp.MatchString(`^-?\d+\.\d+$`, argExpr); matched {
		return "Float"
	}

	// Check for string literal
	if strings.HasPrefix(argExpr, "'") || strings.HasPrefix(argExpr, "\"") {
		return "String"
	}

	// Check for boolean literal
	lowerArg := strings.ToLower(argExpr)
	if lowerArg == "true" || lowerArg == "false" {
		return "Boolean"
	}

	// Default to Variant for complex expressions
	return defaultTypeVariant
}

// generateFunctionSignature generates a function signature with inferred parameter types.
func generateFunctionSignature(functionName string, args []string) string {
	var params []string

	for i, arg := range args {
		paramType := inferParameterType(arg)

		paramName := "arg" + string(rune('0'+i))
		if i < 26 {
			// Use letters for first 26 parameters
			paramName = string(rune('a' + i))
		}

		params = append(params, paramName+": "+paramType)
	}

	paramsStr := strings.Join(params, "; ")
	if paramsStr != "" {
		return "function " + functionName + "(" + paramsStr + "): " + defaultTypeVariant + ";"
	}

	return "function " + functionName + "(): " + defaultTypeVariant + ";"
}

// createDeclareFunctionAction creates a quick fix action to declare an undeclared function.
func createDeclareFunctionAction(diagnostic protocol.Diagnostic, identifierName string, uri string, doc *server.Document) *protocol.CodeAction {
	title := "Declare function '" + identifierName + "'"

	// Extract call arguments to infer parameter types
	args := extractCallArguments(identifierName, diagnostic, doc)
	log.Printf("Function call %s has %d arguments\n", identifierName, len(args))

	// Generate function signature
	functionSignature := generateFunctionSignature(identifierName, args)

	// Find the appropriate insertion location
	insertPosition, indentation := findFunctionInsertionLocation(diagnostic, doc)

	// Generate the function declaration with proper indentation
	functionDecl := indentation + functionSignature + "\n" +
		indentation + "begin\n" +
		indentation + "  // TODO: Implement " + identifierName + "\n" +
		indentation + "  Result := nil;\n" +
		indentation + "end;\n"

	textEdit := protocol.TextEdit{
		Range: protocol.Range{
			Start: insertPosition,
			End:   insertPosition,
		},
		NewText: functionDecl + "\n",
	}

	// Create WorkspaceEdit
	changes := make(map[string][]protocol.TextEdit)
	changes[uri] = []protocol.TextEdit{textEdit}

	workspaceEdit := protocol.WorkspaceEdit{
		Changes: changes,
	}

	action := protocol.CodeAction{
		Title:       title,
		Kind:        stringPtr(string(protocol.CodeActionKindQuickFix)),
		Diagnostics: []protocol.Diagnostic{diagnostic},
		Edit:        &workspaceEdit,
	}

	log.Printf("Created quick fix: %s with signature: %s at line %d\n", title, functionSignature, insertPosition.Line)

	return &action
}

// findFunctionInsertionLocation determines where to insert a function declaration.
// Returns the position and indentation string.
// Functions should be inserted at the top level (global scope) or in the implementation section.
func findFunctionInsertionLocation(diagnostic protocol.Diagnostic, doc *server.Document) (protocol.Position, string) {
	if doc.Text == "" {
		return protocol.Position{Line: 0, Character: 0}, ""
	}

	lines := strings.Split(doc.Text, "\n")
	errorLine := int(diagnostic.Range.Start.Line)

	// Strategy:
	// 1. Look for "implementation" section and insert there
	// 2. Otherwise, look for the last function declaration and insert after it
	// 3. Otherwise, insert after var declarations or at the end of the file

	// Find implementation section
	implLine := findImplementationSection(lines)
	if implLine >= 0 {
		// Insert after implementation keyword
		insertLine := implLine + 1
		return protocol.Position{Line: uint32(insertLine), Character: 0}, ""
	}

	// Find last function declaration before error line
	lastFuncLine := findLastFunctionDeclaration(lines, errorLine)
	if lastFuncLine >= 0 {
		// Find the end of that function
		funcEndLine := findFunctionEnd(lines, lastFuncLine)
		if funcEndLine >= 0 {
			insertLine := funcEndLine + 1
			return protocol.Position{Line: uint32(insertLine), Character: 0}, ""
		}
	}

	// Insert after var declarations or at a reasonable position
	lastVarLine := findLastGlobalVarDeclaration(lines, len(lines))
	if lastVarLine >= 0 {
		insertLine := lastVarLine + 1
		return protocol.Position{Line: uint32(insertLine), Character: 0}, ""
	}

	// Default to after program header
	insertAfterLine := findProgramHeader(lines)

	return protocol.Position{Line: uint32(insertAfterLine + 1), Character: 0}, ""
}

// findImplementationSection finds the "implementation" keyword line.
// Returns the line number, or -1 if not found.
func findImplementationSection(lines []string) int {
	for i, line := range lines {
		lowerLine := strings.TrimSpace(strings.ToLower(line))
		if lowerLine == "implementation" {
			return i
		}
	}

	return -1
}

// findLastFunctionDeclaration finds the last function/procedure declaration before a given line.
// Returns the line number, or -1 if not found.
func findLastFunctionDeclaration(lines []string, beforeLine int) int {
	lastFunc := -1

	for i := 0; i < beforeLine && i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		lowerLine := strings.ToLower(line)

		if strings.HasPrefix(lowerLine, "function ") || strings.HasPrefix(lowerLine, "procedure ") {
			lastFunc = i
		}
	}

	return lastFunc
}

// findFunctionEnd finds the "end;" that closes a function starting at startLine.
// Returns the line number of the "end;", or -1 if not found.
func findFunctionEnd(lines []string, startLine int) int {
	depth := 0
	inFunction := false

	for i := startLine; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		lowerLine := strings.ToLower(line)

		// Track begin/end depth
		if lowerLine == "begin" || strings.HasPrefix(lowerLine, "begin ") {
			depth++
			inFunction = true
		} else if lowerLine == "end;" || strings.HasPrefix(lowerLine, "end;") {
			if inFunction {
				depth--
				if depth == 0 {
					return i
				}
			}
		}
	}

	return -1
}

// createDeclareVariableAction creates a quick fix action to declare an undeclared variable.
func createDeclareVariableAction(diagnostic protocol.Diagnostic, identifierName string, uri string, doc *server.Document) *protocol.CodeAction {
	title := "Declare variable '" + identifierName + "'"

	// Infer the type for the variable (Task 13.5)
	varType := inferTypeFromContext(diagnostic, identifierName, doc)

	// Find the appropriate insertion location (Task 13.6)
	insertPosition, indentation := findInsertionLocation(diagnostic, doc)

	// Generate the declaration text with proper indentation
	declarationText := indentation + generateVariableDeclaration(identifierName, varType)

	textEdit := protocol.TextEdit{
		Range: protocol.Range{
			Start: insertPosition,
			End:   insertPosition,
		},
		NewText: declarationText + "\n",
	}

	// Create WorkspaceEdit
	changes := make(map[string][]protocol.TextEdit)
	changes[uri] = []protocol.TextEdit{textEdit}

	workspaceEdit := protocol.WorkspaceEdit{
		Changes: changes,
	}

	action := protocol.CodeAction{
		Title:       title,
		Kind:        stringPtr(string(protocol.CodeActionKindQuickFix)),
		Diagnostics: []protocol.Diagnostic{diagnostic},
		Edit:        &workspaceEdit,
	}

	log.Printf("Created quick fix: %s (type: %s) at line %d\n", title, varType, insertPosition.Line)

	return &action
}

// createInsertSemicolonAction creates a quick fix action to insert a missing semicolon.
func createInsertSemicolonAction(diagnostic protocol.Diagnostic, uri string) *protocol.CodeAction {
	title := "Insert missing semicolon"

	// The semicolon should be inserted at the end of the diagnostic range
	// This is typically where the parser expected the semicolon to be
	insertPosition := diagnostic.Range.End

	// Create a zero-length range at the insertion point
	textEdit := protocol.TextEdit{
		Range: protocol.Range{
			Start: insertPosition,
			End:   insertPosition,
		},
		NewText: ";",
	}

	// Create WorkspaceEdit
	changes := make(map[string][]protocol.TextEdit)
	changes[uri] = []protocol.TextEdit{textEdit}

	workspaceEdit := protocol.WorkspaceEdit{
		Changes: changes,
	}

	action := protocol.CodeAction{
		Title:       title,
		Kind:        stringPtr(string(protocol.CodeActionKindQuickFix)),
		Diagnostics: []protocol.Diagnostic{diagnostic},
		Edit:        &workspaceEdit,
	}

	log.Printf("Created quick fix: %s at line %d, column %d\n", title, insertPosition.Line, insertPosition.Character)

	return &action
}

// createRemoveVariableAction creates a quick fix action to remove an unused variable declaration.
func createRemoveVariableAction(diagnostic protocol.Diagnostic, variableName string, uri string, doc *server.Document) *protocol.CodeAction {
	title := "Remove unused variable '" + variableName + "'"

	// Find the variable declaration line
	if doc.Text == "" {
		log.Println("Cannot remove variable: document text is empty")
		return nil
	}

	lines := strings.Split(doc.Text, "\n")
	varLine := int(diagnostic.Range.Start.Line)

	if varLine >= len(lines) {
		log.Printf("Variable line %d out of bounds\n", varLine)
		return nil
	}

	// Create a range that covers the entire line including newline
	deleteRange := protocol.Range{
		Start: protocol.Position{Line: uint32(varLine), Character: 0},
		End:   protocol.Position{Line: uint32(varLine + 1), Character: 0},
	}

	// If this is the last line, adjust the range
	if varLine == len(lines)-1 {
		deleteRange.End.Line = uint32(varLine)
		deleteRange.End.Character = uint32(len(lines[varLine]))
	}

	textEdit := protocol.TextEdit{
		Range:   deleteRange,
		NewText: "",
	}

	// Create WorkspaceEdit
	changes := make(map[string][]protocol.TextEdit)
	changes[uri] = []protocol.TextEdit{textEdit}

	workspaceEdit := protocol.WorkspaceEdit{
		Changes: changes,
	}

	action := protocol.CodeAction{
		Title:       title,
		Kind:        stringPtr(string(protocol.CodeActionKindQuickFix)),
		Diagnostics: []protocol.Diagnostic{diagnostic},
		Edit:        &workspaceEdit,
	}

	log.Printf("Created quick fix: %s at line %d\n", title, varLine)

	return &action
}

// createPrefixUnderscoreAction creates a quick fix action to prefix an unused variable with underscore.
func createPrefixUnderscoreAction(diagnostic protocol.Diagnostic, variableName string, uri string, doc *server.Document) *protocol.CodeAction {
	newName := "_" + variableName
	title := "Rename to '" + newName + "'"

	// Find all occurrences of the variable and rename them
	// This is similar to the rename functionality, but simplified for this quick fix
	if doc.Text == "" {
		log.Println("Cannot rename variable: document text is empty")
		return nil
	}

	// For now, we'll do a simple text-based replacement
	// A more sophisticated approach would use the AST to find all references
	var edits []protocol.TextEdit

	lines := strings.Split(doc.Text, "\n")
	varPattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(variableName) + `\b`)

	for lineNum, line := range lines {
		matches := varPattern.FindAllStringIndex(line, -1)
		for _, match := range matches {
			startChar := match[0]
			endChar := match[1]

			edit := protocol.TextEdit{
				Range: protocol.Range{
					Start: protocol.Position{Line: uint32(lineNum), Character: uint32(startChar)},
					End:   protocol.Position{Line: uint32(lineNum), Character: uint32(endChar)},
				},
				NewText: newName,
			}
			edits = append(edits, edit)
		}
	}

	if len(edits) == 0 {
		log.Printf("No occurrences of variable '%s' found\n", variableName)
		return nil
	}

	// Create WorkspaceEdit
	changes := make(map[string][]protocol.TextEdit)
	changes[uri] = edits

	workspaceEdit := protocol.WorkspaceEdit{
		Changes: changes,
	}

	action := protocol.CodeAction{
		Title:       title,
		Kind:        stringPtr(string(protocol.CodeActionKindQuickFix)),
		Diagnostics: []protocol.Diagnostic{diagnostic},
		Edit:        &workspaceEdit,
	}

	log.Printf("Created quick fix: %s (%d edits)\n", title, len(edits))

	return &action
}

// inferTypeFromContext attempts to infer the type of a variable from its usage context.
// For Task 13.5, this provides basic type inference:
// - Integer literals → Integer
// - String literals → String
// - Default → Variant.
func inferTypeFromContext(diagnostic protocol.Diagnostic, identifierName string, doc *server.Document) string {
	// For now, we'll use a simple heuristic by looking at the line of code
	// More sophisticated analysis would examine the AST

	// Get the document text
	if doc.Text == "" {
		return defaultTypeVariant // Default type if no text available
	}

	// Get the line where the error occurred
	lines := strings.Split(doc.Text, "\n")
	if int(diagnostic.Range.Start.Line) >= len(lines) {
		return defaultTypeVariant
	}

	line := lines[diagnostic.Range.Start.Line]

	// Simple pattern matching for common cases
	// Look for assignment patterns like: x := value or x = value
	assignmentPattern := regexp.MustCompile(identifierName + `\s*:?=\s*(.+?)[;,\n]`)
	matches := assignmentPattern.FindStringSubmatch(line)

	if len(matches) >= 2 {
		value := strings.TrimSpace(matches[1])

		// Check for integer literal
		if matched, _ := regexp.MatchString(`^-?\d+$`, value); matched {
			return "Integer"
		}

		// Check for float literal
		if matched, _ := regexp.MatchString(`^-?\d+\.\d+$`, value); matched {
			return "Float"
		}

		// Check for string literal
		if strings.HasPrefix(value, "'") || strings.HasPrefix(value, "\"") {
			return "String"
		}

		// Check for boolean literal
		lowerValue := strings.ToLower(value)
		if lowerValue == "true" || lowerValue == "false" {
			return "Boolean"
		}
	}

	// Default to Variant if we can't infer the type
	return defaultTypeVariant
}

// findInsertionLocation determines where to insert a variable declaration.
// Returns the position and indentation string.
// Task 13.6: Insert at function top (after begin) or global scope (after var block).
func findInsertionLocation(diagnostic protocol.Diagnostic, doc *server.Document) (protocol.Position, string) {
	if doc.Text == "" {
		// Default to beginning of file if no text
		return protocol.Position{Line: 0, Character: 0}, ""
	}

	lines := strings.Split(doc.Text, "\n")
	errorLine := int(diagnostic.Range.Start.Line)

	// Look backwards from error line to find context
	// Strategy:
	// 1. Find if we're inside a function (look for "begin" keyword)
	// 2. If yes, insert after the "begin" line
	// 3. If no, look for global "var" declarations and insert after them
	// 4. Otherwise, insert at beginning of file

	// Look for "begin" keyword indicating function body
	functionBeginLine := findFunctionBegin(lines, errorLine)
	if functionBeginLine >= 0 {
		// We're inside a function, insert after the begin line
		insertLine := functionBeginLine + 1
		indentation := detectIndentation(lines, insertLine)

		return protocol.Position{Line: uint32(insertLine), Character: 0}, indentation
	}

	// Not in a function, look for global var declarations
	lastVarLine := findLastGlobalVarDeclaration(lines, errorLine)
	if lastVarLine >= 0 {
		// Insert after the last var declaration
		insertLine := lastVarLine + 1
		return protocol.Position{Line: uint32(insertLine), Character: 0}, ""
	}

	// No var block found, insert at beginning of file
	// Look for program header/uses clause and insert after
	insertAfterLine := findProgramHeader(lines)

	return protocol.Position{Line: uint32(insertAfterLine), Character: 0}, ""
}

// findFunctionBegin looks backwards from errorLine to find a "begin" keyword.
// Returns the line number of the begin, or -1 if not found.
func findFunctionBegin(lines []string, errorLine int) int {
	// Look backwards up to 50 lines
	maxLookback := 50
	for i := errorLine - 1; i >= 0 && i >= errorLine-maxLookback; i-- {
		line := strings.TrimSpace(lines[i])
		lowerLine := strings.ToLower(line)

		// Check if this line has "begin"
		if lowerLine == keywordBegin || strings.HasPrefix(lowerLine, keywordBegin+" ") || strings.HasPrefix(lowerLine, keywordBegin+";") {
			return i
		}

		// Stop if we hit "end" or other block terminators
		if lowerLine == "end;" || lowerLine == "end" || strings.HasPrefix(lowerLine, "end;") {
			break
		}
	}

	return -1
}

// findLastGlobalVarDeclaration finds the last global "var" declaration before errorLine.
// Returns the line number, or -1 if not found.
func findLastGlobalVarDeclaration(lines []string, errorLine int) int {
	lastVarLine := -1

	// Look from start of file to error line
	for i := 0; i < errorLine && i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		lowerLine := strings.ToLower(line)

		// Check for "var" keyword at the start of line (global var)
		if strings.HasPrefix(lowerLine, "var ") {
			lastVarLine = i
		}

		// Stop if we hit "begin" (start of code)
		if lowerLine == "begin" || strings.HasPrefix(lowerLine, "begin ") {
			break
		}
	}

	return lastVarLine
}

// findProgramHeader finds where to insert after program header/uses clauses.
// Returns the line number to insert after.
func findProgramHeader(lines []string) int {
	insertAfter := 0

	for i := 0; i < len(lines) && i < 20; i++ {
		line := strings.TrimSpace(lines[i])
		lowerLine := strings.ToLower(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Check for program/unit/uses declarations
		if strings.HasPrefix(lowerLine, "program ") ||
			strings.HasPrefix(lowerLine, "unit ") ||
			strings.HasPrefix(lowerLine, "uses ") {
			insertAfter = i + 1
			continue
		}

		// If we hit actual code, stop
		if strings.HasPrefix(lowerLine, "var ") ||
			strings.HasPrefix(lowerLine, "function ") ||
			strings.HasPrefix(lowerLine, "procedure ") ||
			strings.HasPrefix(lowerLine, "begin") {
			break
		}
	}

	return insertAfter
}

// detectIndentation detects the indentation used around the given line.
// Returns a string of spaces (typically 2 or 4 spaces).
func detectIndentation(lines []string, nearLine int) string {
	// Look at nearby lines to detect indentation
	for i := nearLine; i < len(lines) && i < nearLine+5; i++ {
		if i >= len(lines) {
			break
		}

		line := lines[i]
		if len(line) > 0 && line[0] == ' ' {
			// Count leading spaces
			spaces := 0
			for j := 0; j < len(line) && line[j] == ' '; j++ {
				spaces++
			}

			if spaces > 0 {
				return strings.Repeat(" ", spaces)
			}
		}
	}

	// Default to 2 spaces
	return "  "
}

// generateVariableDeclaration generates the variable declaration text.
func generateVariableDeclaration(identifierName string, varType string) string {
	return "var " + identifierName + ": " + varType + ";"
}

// stringPtr is a helper function to create a pointer to a string.
func stringPtr(s string) *string {
	return &s
}

// GenerateSourceActions generates source code actions (refactorings) for a document.
// These are actions that aren't tied to specific diagnostics but operate on the entire source.
func GenerateSourceActions(doc *server.Document, uri string, context protocol.CodeActionContext) []protocol.CodeAction {
	var actions []protocol.CodeAction

	// Check if source actions are requested
	// If only specific kinds are requested, check if source actions are included
	if context.Only != nil && len(context.Only) > 0 {
		hasSourceKind := false

		for _, kind := range context.Only {
			if kind == protocol.CodeActionKindSource ||
				kind == protocol.CodeActionKindSourceOrganizeImports ||
				strings.HasPrefix(string(kind), string(protocol.CodeActionKindSource)) {
				hasSourceKind = true
				break
			}
		}

		if !hasSourceKind {
			return actions
		}
	}

	// Create "Organize units" source action
	organizeAction := createOrganizeUnitsAction(doc, uri)
	if organizeAction != nil {
		actions = append(actions, *organizeAction)
	}

	return actions
}

// createOrganizeUnitsAction creates a source action to organize the uses clause.
// For task 13.10, this implements:
// - Removing unused unit references
// - Adding missing unit references for undeclared identifiers
// - Sorting units alphabetically.
func createOrganizeUnitsAction(doc *server.Document, uri string) *protocol.CodeAction {
	title := "Organize units"

	if doc.Text == "" {
		log.Println("Cannot organize units: document text is empty")
		return nil
	}

	// Get server instance for workspace index access
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available for organize units")
		// Fall back to basic sorting
		edit := organizeUsesClause(doc.Text, uri, nil, doc)
		if edit == nil {
			return nil
		}

		action := protocol.CodeAction{
			Title: title,
			Kind:  stringPtr(string(protocol.CodeActionKindSourceOrganizeImports)),
			Edit:  edit,
		}

		return &action
	}

	// Extract and organize the uses clause with full functionality
	edit := organizeUsesClause(doc.Text, uri, srv.WorkspaceIndex(), doc)
	if edit == nil {
		log.Println("No uses clause found or no changes needed")
		return nil
	}

	action := protocol.CodeAction{
		Title: title,
		Kind:  stringPtr(string(protocol.CodeActionKindSourceOrganizeImports)),
		Edit:  edit,
	}

	log.Printf("Created source action: %s\n", title)

	return &action
}

// organizeUsesClause extracts the uses clause, removes unused units, adds missing units,
// sorts them, and creates a WorkspaceEdit.
// Returns nil if no uses clause is found or no changes are needed.
func organizeUsesClause(text string, uri string, workspaceIndex *workspace.SymbolIndex, doc *server.Document) *protocol.WorkspaceEdit {
	lines := strings.Split(text, "\n")

	// Find the uses clause
	usesStart := -1
	usesEnd := -1

	for i, line := range lines {
		lowerLine := strings.TrimSpace(strings.ToLower(line))

		// Look for "uses" keyword
		if strings.HasPrefix(lowerLine, "uses ") || lowerLine == "uses" {
			usesStart = i

			// Find the end of the uses clause (ends with semicolon)
			for j := i; j < len(lines); j++ {
				if strings.Contains(lines[j], ";") {
					usesEnd = j
					break
				}
			}

			break
		}
	}

	// If no uses clause exists, check if we need to add one
	if usesStart == -1 || usesEnd == -1 {
		// Try to find missing units and create a new uses clause if needed
		if workspaceIndex != nil && doc != nil {
			missingUnits := findMissingUnits(text, workspaceIndex, doc)
			if len(missingUnits) > 0 {
				// Create a new uses clause
				insertLine := findUsesInsertionPoint(lines)

				sortUnits(missingUnits)
				newUsesClause := formatUsesClause(missingUnits)

				textEdit := protocol.TextEdit{
					Range: protocol.Range{
						Start: protocol.Position{Line: uint32(insertLine), Character: 0},
						End:   protocol.Position{Line: uint32(insertLine), Character: 0},
					},
					NewText: newUsesClause + "\n",
				}

				changes := make(map[string][]protocol.TextEdit)
				changes[uri] = []protocol.TextEdit{textEdit}

				workspaceEdit := protocol.WorkspaceEdit{
					Changes: changes,
				}

				log.Printf("Created new uses clause with %d units\n", len(missingUnits))

				return &workspaceEdit
			}
		}

		return nil
	}

	// Extract the uses clause text
	usesText := ""

	var usesTextSb1095 strings.Builder
	for i := usesStart; i <= usesEnd; i++ {
		usesTextSb1095.WriteString(lines[i])

		if i < usesEnd {
			usesTextSb1095.WriteString("\n")
		}
	}

	usesText += usesTextSb1095.String()

	// Parse unit names from the uses clause
	units := parseUnitsFromUsesClause(usesText)
	if len(units) == 0 {
		return nil
	}

	// Create a set of units to keep
	finalUnits := make([]string, 0, len(units))
	finalUnits = append(finalUnits, units...)

	// Remove unused units if workspace index is available and populated
	if workspaceIndex != nil && workspaceIndex.GetSymbolCount() > 0 && doc != nil {
		unusedUnits := findUnusedUnits(units, text, workspaceIndex, doc)
		if len(unusedUnits) > 0 {
			log.Printf("Found %d unused units: %v\n", len(unusedUnits), unusedUnits)
			// Filter out unused units
			filteredUnits := make([]string, 0, len(finalUnits))
			for _, unit := range finalUnits {
				if !containsString(unusedUnits, unit) {
					filteredUnits = append(filteredUnits, unit)
				}
			}

			finalUnits = filteredUnits
		}

		// Add missing units
		missingUnits := findMissingUnits(text, workspaceIndex, doc)
		for _, unit := range missingUnits {
			if !containsString(finalUnits, unit) {
				finalUnits = append(finalUnits, unit)
				log.Printf("Adding missing unit: %s\n", unit)
			}
		}
	}

	// Sort units alphabetically
	sortedUnits := make([]string, len(finalUnits))
	copy(sortedUnits, finalUnits)
	sortUnits(sortedUnits)

	// If no changes needed, return nil
	if unitsEqual(units, sortedUnits) {
		log.Println("Units already organized")
		return nil
	}

	// Generate new uses clause
	newUsesClause := formatUsesClause(sortedUnits)

	// Create text edit to replace the uses clause
	textEdit := protocol.TextEdit{
		Range: protocol.Range{
			Start: protocol.Position{Line: uint32(usesStart), Character: 0},
			End:   protocol.Position{Line: uint32(usesEnd), Character: uint32(len(lines[usesEnd]))},
		},
		NewText: newUsesClause,
	}

	changes := make(map[string][]protocol.TextEdit)
	changes[uri] = []protocol.TextEdit{textEdit}

	workspaceEdit := protocol.WorkspaceEdit{
		Changes: changes,
	}

	log.Printf("Organized %d units in uses clause (from %d original)\n", len(sortedUnits), len(units))

	return &workspaceEdit
}

// parseUnitsFromUsesClause extracts unit names from a uses clause.
func parseUnitsFromUsesClause(usesText string) []string {
	var units []string

	// Remove "uses" keyword and semicolon
	text := usesText
	text = regexp.MustCompile(`(?i)\buses\b`).ReplaceAllString(text, "")
	text = strings.ReplaceAll(text, ";", "")
	text = strings.TrimSpace(text)

	// Split by comma
	parts := strings.SplitSeq(text, ",")
	for part := range parts {
		unit := strings.TrimSpace(part)
		if unit != "" {
			units = append(units, unit)
		}
	}

	return units
}

// sortUnits sorts a slice of unit names alphabetically (case-insensitive).
func sortUnits(units []string) {
	// Simple bubble sort for small lists
	n := len(units)
	for i := range n - 1 {
		for j := range n - i - 1 {
			if strings.ToLower(units[j]) > strings.ToLower(units[j+1]) {
				units[j], units[j+1] = units[j+1], units[j]
			}
		}
	}
}

// unitsEqual checks if two unit slices are equal.
func unitsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// formatUsesClause formats a uses clause from a list of sorted units.
func formatUsesClause(units []string) string {
	if len(units) == 0 {
		return ""
	}

	// Simple format: uses unit1, unit2, unit3;
	return "uses " + strings.Join(units, ", ") + ";"
}

// findUnusedUnits identifies units in the uses clause that are not actually referenced in the document.
// For task 13.10 and 13.11: Remove unused unit references.
func findUnusedUnits(units []string, text string, workspaceIndex *workspace.SymbolIndex, doc *server.Document) []string {
	var unused []string

	// For each unit, check if any symbols from that unit are used
	for _, unit := range units {
		if !isUnitUsed(unit, text, workspaceIndex, doc) {
			unused = append(unused, unit)
		}
	}

	return unused
}

// isUnitUsed checks if a unit is actually used in the document by checking if any symbols
// from that unit are referenced.
func isUnitUsed(unitName string, text string, workspaceIndex *workspace.SymbolIndex, doc *server.Document) bool {
	// Get all symbols from the workspace index
	// We need to find symbols that are defined in files matching the unit name

	// Extract all identifiers used in the current document
	identifiers := extractIdentifiersFromText(text)
	if len(identifiers) == 0 {
		return false
	}

	// For each identifier, check if it's defined in this unit
	for identifier := range identifiers {
		locations := workspaceIndex.FindSymbol(identifier)
		for _, loc := range locations {
			// Extract unit name from the file URI
			if getUnitNameFromURI(loc.Location.URI) == unitName {
				// This identifier is from this unit and is used
				return true
			}
		}
	}

	return false
}

// findMissingUnits identifies symbols that are used but not declared and their defining units.
// For task 13.10 and 13.12: Add missing unit references for used symbols.
func findMissingUnits(text string, workspaceIndex *workspace.SymbolIndex, doc *server.Document) []string {
	var missing []string
	seenUnits := make(map[string]bool)

	// Get current units from uses clause (if any)
	currentUnits := parseUnitsFromUsesClause(text)

	currentUnitsMap := make(map[string]bool)
	for _, unit := range currentUnits {
		currentUnitsMap[unit] = true
	}

	// Find all undeclared identifiers by looking at the AST
	if doc.Program == nil || doc.Program.AST() == nil {
		return missing
	}

	undeclaredIdentifiers := findUndeclaredIdentifiers(doc.Program.AST(), doc)
	if len(undeclaredIdentifiers) == 0 {
		return missing
	}

	log.Printf("Found %d potentially undeclared identifiers\n", len(undeclaredIdentifiers))

	// For each undeclared identifier, search for its definition in the workspace
	for identifier := range undeclaredIdentifiers {
		locations := workspaceIndex.FindSymbol(identifier)
		if len(locations) == 0 {
			continue
		}

		// Get the unit name from the first location
		unitName := getUnitNameFromURI(locations[0].Location.URI)
		if unitName == "" {
			continue
		}

		// Skip if already in current units or already seen
		if currentUnitsMap[unitName] || seenUnits[unitName] {
			continue
		}

		// Skip if it's the current file itself
		if getUnitNameFromURI(doc.URI) == unitName {
			continue
		}

		missing = append(missing, unitName)
		seenUnits[unitName] = true
		log.Printf("Found missing unit %s for identifier %s\n", unitName, identifier)
	}

	return missing
}

// extractIdentifiersFromText extracts all identifiers used in the document text.
// This is a simple regex-based approach.
func extractIdentifiersFromText(text string) map[string]bool {
	identifiers := make(map[string]bool)

	// Match identifier pattern: word characters starting with a letter
	pattern := regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`)
	matches := pattern.FindAllString(text, -1)

	for _, match := range matches {
		// Skip keywords
		lower := strings.ToLower(match)
		if isKeyword(lower) {
			continue
		}

		identifiers[match] = true
	}

	return identifiers
}

// findUndeclaredIdentifiers finds identifiers that are used but not declared in the current scope.
// This is a simplified implementation that looks for identifier expressions.
func findUndeclaredIdentifiers(programAST *ast.Program, doc *server.Document) map[string]bool {
	undeclared := make(map[string]bool)

	// Get all declared symbols in the current file
	declared := getDeclaredSymbols(programAST)

	// Traverse the AST to find all used identifiers
	used := getUsedIdentifiers(programAST)

	// Find identifiers that are used but not declared
	for identifier := range used {
		if !declared[identifier] && !isBuiltinIdentifier(identifier) {
			undeclared[identifier] = true
		}
	}

	return undeclared
}

// getDeclaredSymbols extracts all symbols declared in the program.
func getDeclaredSymbols(programAST *ast.Program) map[string]bool {
	declared := make(map[string]bool)

	if programAST == nil {
		return declared
	}

	for _, stmt := range programAST.Statements {
		if stmt == nil {
			continue
		}

		switch node := stmt.(type) {
		case *ast.FunctionDecl:
			if node.Name != nil {
				declared[node.Name.Value] = true
			}
		case *ast.VarDeclStatement:
			for _, name := range node.Names {
				if name != nil {
					declared[name.Value] = true
				}
			}
		case *ast.ConstDecl:
			if node.Name != nil {
				declared[node.Name.Value] = true
			}
		case *ast.ClassDecl:
			if node.Name != nil {
				declared[node.Name.Value] = true
			}
		case *ast.RecordDecl:
			if node.Name != nil {
				declared[node.Name.Value] = true
			}
		case *ast.EnumDecl:
			if node.Name != nil {
				declared[node.Name.Value] = true
			}
		}
	}

	return declared
}

// getUsedIdentifiers traverses the AST and collects all identifiers that are used.
// This is a simplified implementation that extracts identifiers from basic expressions.
func getUsedIdentifiers(programAST *ast.Program) map[string]bool {
	used := make(map[string]bool)

	if programAST == nil {
		return used
	}

	// Traverse all statements and collect identifiers
	var traverse func(node ast.Node)
	traverse = func(node ast.Node) {
		if node == nil {
			return
		}

		switch n := node.(type) {
		case *ast.Identifier:
			used[n.Value] = true

		case *ast.Program:
			for _, stmt := range n.Statements {
				traverse(stmt)
			}

		case *ast.FunctionDecl:
			if n.Body != nil {
				traverse(n.Body)
			}
			// Traverse parameters
			for _, param := range n.Parameters {
				if param.DefaultValue != nil {
					traverse(param.DefaultValue)
				}
			}

		case *ast.BlockStatement:
			for _, stmt := range n.Statements {
				traverse(stmt)
			}

		case *ast.ExpressionStatement:
			traverse(n.Expression)

		case *ast.BinaryExpression:
			traverse(n.Left)
			traverse(n.Right)

		case *ast.CallExpression:
			traverse(n.Function)

			for _, arg := range n.Arguments {
				traverse(arg)
			}

		case *ast.IndexExpression:
			traverse(n.Left)
			traverse(n.Index)

		case *ast.IfStatement:
			traverse(n.Condition)
			traverse(n.Consequence)

			if n.Alternative != nil {
				traverse(n.Alternative)
			}

		case *ast.WhileStatement:
			traverse(n.Condition)
			traverse(n.Body)

		case *ast.RepeatStatement:
			traverse(n.Condition)
			traverse(n.Body)

		case *ast.ForStatement:
			traverse(n.Body)

		case *ast.CaseStatement:
			traverse(n.Expression)

		case *ast.VarDeclStatement:
			if n.Value != nil {
				traverse(n.Value)
			}
		}
	}

	for _, stmt := range programAST.Statements {
		traverse(stmt)
	}

	return used
}

// getUnitNameFromURI extracts the unit name from a file URI.
// For example, "file:///path/to/MyUnit.dws" returns "MyUnit".
func getUnitNameFromURI(uri string) string {
	// Extract filename from URI
	path := uri
	if after, ok := strings.CutPrefix(uri, "file://"); ok {
		path = after
		// On Windows, remove leading slash
		if len(path) > 2 && path[0] == '/' && path[2] == ':' {
			path = path[1:]
		}
	}

	// Get base filename without extension
	filename := filepath.Base(path)

	ext := filepath.Ext(filename)
	if ext != "" {
		filename = filename[:len(filename)-len(ext)]
	}

	return filename
}

// findUsesInsertionPoint finds the best location to insert a new uses clause.
// Returns the line number where the uses clause should be inserted.
func findUsesInsertionPoint(lines []string) int {
	// Look for program/unit declaration and insert after it
	for i, line := range lines {
		lowerLine := strings.TrimSpace(strings.ToLower(line))
		if strings.HasPrefix(lowerLine, "program ") || strings.HasPrefix(lowerLine, "unit ") {
			return i + 1
		}
	}

	// If no program/unit declaration, insert at the beginning
	return 0
}

// containsString checks if a string slice contains a specific string.
func containsString(slice []string, str string) bool {
	return slices.Contains(slice, str)
}

// isKeyword checks if a string is a DWScript keyword.
func isKeyword(word string) bool {
	keywords := map[string]bool{
		"and": true, "array": true, "as": true, "begin": true, "case": true,
		"class": true, "const": true, "constructor": true, "destructor": true,
		"div": true, "do": true, "downto": true, "else": true, "end": true,
		"except": true, "exit": true, "false": true, "finally": true, "for": true,
		"function": true, "if": true, "implementation": true, "in": true,
		"inherited": true, "interface": true, "is": true, "lazy": true, "mod": true,
		"new": true, "nil": true, "not": true, "of": true, "or": true,
		"procedure": true, "program": true, "property": true, "raise": true,
		"record": true, "repeat": true, "result": true, "self": true, "set": true,
		"shl": true, "shr": true, "then": true, "to": true, "true": true,
		"try": true, "type": true, "unit": true, "until": true, "uses": true,
		"var": true, "while": true, "with": true, "xor": true,
	}

	return keywords[word]
}

// isBuiltinIdentifier checks if an identifier is a built-in type or function.
func isBuiltinIdentifier(identifier string) bool {
	builtins := map[string]bool{
		"Integer": true, "String": true, "Float": true, "Boolean": true,
		"Variant": true, "TObject": true, "TClass": true,
		"WriteLn": true, "Write": true, "ReadLn": true, "Read": true,
		"Length": true, "Copy": true, "Pos": true, "Delete": true,
		"Insert": true, "Chr": true, "Ord": true, "Inc": true, "Dec": true,
		"High": true, "Low": true, "SetLength": true,
	}

	return builtins[identifier]
}
