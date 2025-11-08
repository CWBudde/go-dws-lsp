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

			// Create "Declare variable" quick fix
			action := createDeclareVariableAction(diagnostic, identifierName, uri, doc)
			if action != nil {
				actions = append(actions, *action)
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

// extractVariableName extracts the variable name from an unused variable diagnostic message.
// It looks for patterns like "unused variable 'x'" or "variable x not used"
func extractVariableName(diagnostic protocol.Diagnostic) string {
	message := diagnostic.Message

	// Try various regex patterns to extract the variable name
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`['\"]([a-zA-Z_][a-zA-Z0-9_]*)['\"]`),                // 'varname' or "varname"
		regexp.MustCompile(`variable\s+([a-zA-Z_][a-zA-Z0-9_]*)`),               // variable name
		regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*)\s+(?:not used|unused)`),    // name not used/unused
		regexp.MustCompile(`unused:\s*([a-zA-Z_][a-zA-Z0-9_]*)`),                // unused: name
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
// - Default → Variant
func inferTypeFromContext(diagnostic protocol.Diagnostic, identifierName string, doc *server.Document) string {
	// For now, we'll use a simple heuristic by looking at the line of code
	// More sophisticated analysis would examine the AST

	// Get the document text
	if doc.Text == "" {
		return "Variant" // Default type if no text available
	}

	// Get the line where the error occurred
	lines := strings.Split(doc.Text, "\n")
	if int(diagnostic.Range.Start.Line) >= len(lines) {
		return "Variant"
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
	return "Variant"
}

// findInsertionLocation determines where to insert a variable declaration.
// Returns the position and indentation string.
// Task 13.6: Insert at function top (after begin) or global scope (after var block)
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
		if lowerLine == "begin" || strings.HasPrefix(lowerLine, "begin ") || strings.HasPrefix(lowerLine, "begin;") {
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
// For task 13.10, this implements basic alphabetical sorting of units.
func createOrganizeUnitsAction(doc *server.Document, uri string) *protocol.CodeAction {
	title := "Organize units"

	if doc.Text == "" {
		log.Println("Cannot organize units: document text is empty")
		return nil
	}

	// Extract and organize the uses clause
	edit := organizeUsesClause(doc.Text, uri)
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

// organizeUsesClause extracts the uses clause, sorts the units, and creates a WorkspaceEdit.
// Returns nil if no uses clause is found or no changes are needed.
func organizeUsesClause(text string, uri string) *protocol.WorkspaceEdit {
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

	if usesStart == -1 || usesEnd == -1 {
		return nil
	}

	// Extract the uses clause text
	usesText := ""
	for i := usesStart; i <= usesEnd; i++ {
		usesText += lines[i]
		if i < usesEnd {
			usesText += "\n"
		}
	}

	// Parse unit names from the uses clause
	units := parseUnitsFromUsesClause(usesText)
	if len(units) == 0 {
		return nil
	}

	// Check if already sorted
	sortedUnits := make([]string, len(units))
	copy(sortedUnits, units)
	sortUnits(sortedUnits)

	// If already sorted, no changes needed
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

	log.Printf("Organized %d units in uses clause\n", len(units))
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
	parts := strings.Split(text, ",")
	for _, part := range parts {
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
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
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
