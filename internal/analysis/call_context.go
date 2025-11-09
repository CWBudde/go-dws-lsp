package analysis

import (
	"log"
	"strings"

	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/dwscript"
	"github.com/cwbudde/go-dws/pkg/token"
)

// CallContext holds information about a function call at the cursor position.
type CallContext struct {
	// The call expression node (CallExpression or MethodCallExpression)
	CallNode ast.Node

	// Function or method name
	FunctionName string

	// For method calls, the object expression (nil for regular function calls)
	ObjectExpr ast.Expression

	// Current parameter index (0-based)
	ParameterIndex int

	// Whether cursor is inside the parentheses of a call
	IsInsideCall bool
}

// DetermineCallContext analyzes the cursor position to determine if it's inside a function call
// Returns nil if the cursor is not inside a function call.
func DetermineCallContext(doc *server.Document, line, character int) *CallContext {
	if doc.Program == nil {
		return nil
	}

	programAST := doc.Program.AST()
	if programAST == nil {
		return nil
	}

	// Convert LSP position (0-based) to AST position (1-based)
	astLine := line + 1
	astColumn := character + 1

	log.Printf("DetermineCallContext: line=%d, character=%d (AST: %d:%d)\n", line, character, astLine, astColumn)

	// Find enclosing call expression using AST traversal
	callNode := findEnclosingCallExpression(programAST, astLine, astColumn)
	if callNode == nil {
		log.Printf("No enclosing call expression found\n")
		return nil
	}

	// Extract function name based on node type
	functionName, objectExpr := extractFunctionName(callNode)
	if functionName == "" {
		log.Printf("Could not extract function name from call node\n")
		return nil
	}

	log.Printf("Found call expression: function=%s\n", functionName)

	// Use text-based analysis to find parameter index
	// This handles incomplete AST during typing
	paramIndex := findParameterIndex(doc.Text, line, character, callNode)

	log.Printf("Parameter index: %d\n", paramIndex)

	return &CallContext{
		CallNode:       callNode,
		FunctionName:   functionName,
		ObjectExpr:     objectExpr,
		ParameterIndex: paramIndex,
		IsInsideCall:   true,
	}
}

// findEnclosingCallExpression traverses the AST to find the innermost CallExpression
// or MethodCallExpression that contains the given position.
func findEnclosingCallExpression(program *ast.Program, line, col int) ast.Node {
	var enclosingCall ast.Node

	// Track depth to find the deepest (innermost) call expression
	var deepestDepth int
	currentDepth := 0

	ast.Inspect(program, func(n ast.Node) bool {
		if n == nil {
			currentDepth--
			return false
		}

		currentDepth++

		// Check if this node contains our position
		nodeStart := n.Pos()
		nodeEnd := n.End()

		if !positionInRange(token.Position{Line: line, Column: col}, nodeStart, nodeEnd) {
			currentDepth--
			return false // Skip this branch
		}

		// Check if this is a call expression
		switch node := n.(type) {
		case *ast.CallExpression:
			// For CallExpression, check if cursor is within the parentheses
			// by checking if it's after the function name
			if currentDepth > deepestDepth {
				enclosingCall = node
				deepestDepth = currentDepth
				log.Printf("Found CallExpression at depth %d: %v\n", currentDepth, node.Function)
			}

		case *ast.MethodCallExpression:
			// For MethodCallExpression, check if cursor is within the parentheses
			if currentDepth > deepestDepth {
				enclosingCall = node
				deepestDepth = currentDepth
				log.Printf("Found MethodCallExpression at depth %d: %s\n", currentDepth, node.Method.Value)
			}
		}

		return true // Continue traversing
	})

	return enclosingCall
}

// findParameterIndex counts commas from the opening parenthesis to the cursor
// to determine which parameter the cursor is on (0-based index).
func findParameterIndex(text string, line, character int, callNode ast.Node) int {
	// Convert positions to text offset
	lines := strings.Split(text, "\n")
	if line < 0 || line >= len(lines) {
		return 0
	}

	currentLine := lines[line]
	if character < 0 || character > len(currentLine) {
		return 0
	}

	textBefore := buildTextBeforeCursor(lines, line, currentLine, character)
	commaCount := countCommasBeforeCursor(textBefore)

	log.Printf("Counted %d commas before cursor\n", commaCount)

	// Parameter index is the number of commas we found
	// (0 commas = first parameter, 1 comma = second parameter, etc.)
	return commaCount
}

// buildTextBeforeCursor constructs the text content from start of document to cursor position.
func buildTextBeforeCursor(lines []string, line int, currentLine string, character int) string {
	var textBeforeCursor strings.Builder

	// Get all text from start of document up to cursor
	for i := range line {
		textBeforeCursor.WriteString(lines[i])
		textBeforeCursor.WriteString("\n")
	}

	textBeforeCursor.WriteString(currentLine[:character])
	return textBeforeCursor.String()
}

// countCommasBeforeCursor scans backward from cursor to count parameter-separating commas.
func countCommasBeforeCursor(textBefore string) int {
	// Scan backward from cursor to find the opening parenthesis of this call
	// Track nesting level to handle nested calls
	parenDepth := 0
	bracketDepth := 0
	inString := false
	var stringChar rune
	commaCount := 0

	// Convert text to runes for proper character handling
	runes := []rune(textBefore)

	// Scan backward
	foundOpenParen := false

	for i := len(runes) - 1; i >= 0; i-- {
		r := runes[i]

		// Handle string literals
		if r == '"' || r == '\'' {
			// Check if not escaped
			if i == 0 || runes[i-1] != '\\' {
				if inString && r == stringChar {
					inString = false
				} else if !inString {
					inString = true
					stringChar = r
				}
			}

			continue
		}

		if inString {
			continue
		}

		// Handle parentheses
		switch r {
		case ')':
			parenDepth++
		case '(':
			if parenDepth == 0 {
				// Found the opening parenthesis of our call
				foundOpenParen = true
				break
			}

			parenDepth--
		case ']':
			bracketDepth++
		case '[':
			if bracketDepth > 0 {
				bracketDepth--
			}
		case ',':
			// Only count commas at the same nesting level
			if parenDepth == 0 && bracketDepth == 0 {
				commaCount++
			}
		}
	}

	if !foundOpenParen {
		log.Printf("Warning: Could not find opening parenthesis for call\n")
		return 0
	}

	return commaCount
}

// extractFunctionName extracts the function or method name from a call expression node
// Returns the name and optionally the object expression (for method calls).
func extractFunctionName(callNode ast.Node) (string, ast.Expression) {
	switch node := callNode.(type) {
	case *ast.CallExpression:
		// Extract function name from the Function expression
		switch funcExpr := node.Function.(type) {
		case *ast.Identifier:
			return funcExpr.Value, nil
		case *ast.MemberAccessExpression:
			// Handle qualified names like "object.property"
			// For now, just return the member name
			if funcExpr.Member != nil {
				return funcExpr.Member.Value, funcExpr.Object
			}
		}

	case *ast.MethodCallExpression:
		// Method call has explicit Method field
		if node.Method != nil {
			return node.Method.Value, node.Object
		}
	}

	return "", nil
}

// FindFunctionAtCall scans backward from the cursor position to find the function being called
// This is a text-based approach that works even when the AST is incomplete
// Returns the function name or an error if not found.
func FindFunctionAtCall(doc *server.Document, line, character int) (string, error) {
	if doc.Text == "" {
		return "", nil
	}

	lines := strings.Split(doc.Text, "\n")
	if line < 0 || line >= len(lines) {
		return "", nil
	}

	currentLine := lines[line]
	if character < 0 || character > len(currentLine) {
		return "", nil
	}

	textBefore := buildTextBeforeCursor(lines, line, currentLine, character)
	runes := []rune(textBefore)

	openParenIndex := findOpeningParenthesis(runes)
	if openParenIndex == -1 {
		log.Printf("FindFunctionAtCall: No opening parenthesis found\n")
		return "", nil
	}

	functionName := extractFunctionNameFromRunes(runes, openParenIndex)
	if functionName == "" {
		log.Printf("FindFunctionAtCall: No function identifier found\n")
		return "", nil
	}

	log.Printf("FindFunctionAtCall: Found function '%s'\n", functionName)
	return functionName, nil
}

// findOpeningParenthesis scans backward to locate the opening parenthesis of a function call.
func findOpeningParenthesis(runes []rune) int {
	parenDepth := 0
	inString := false
	var stringChar rune

	for index := len(runes) - 1; index >= 0; index-- {
		r := runes[index]

		// Handle string literals
		if r == '"' || r == '\'' {
			if index == 0 || runes[index-1] != '\\' {
				if inString && r == stringChar {
					inString = false
				} else if !inString {
					inString = true
					stringChar = r
				}
			}

			continue
		}

		if inString {
			continue
		}

		// Handle parentheses
		if r == ')' {
			parenDepth++
		} else if r == '(' {
			if parenDepth == 0 {
				return index
			}

			parenDepth--
		}
	}

	return -1
}

// extractFunctionNameFromRunes extracts the function name before the opening parenthesis.
func extractFunctionNameFromRunes(runes []rune, openParenIndex int) string {
	// Skip whitespace before the opening parenthesis
	index := openParenIndex - 1
	for index >= 0 && isWhitespace(runes[index]) {
		index--
	}

	if index < 0 {
		return ""
	}

	// Collect the function name (may include dots for qualified names)
	var functionNameRunes []rune

	for index >= 0 {
		r := runes[index]
		if isIdentifierChar(r) || r == '.' {
			functionNameRunes = append([]rune{r}, functionNameRunes...)
			index--
		} else {
			break
		}
	}

	return string(functionNameRunes)
}

// isWhitespace checks if a rune is whitespace.
func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// isIdentifierChar checks if a rune can be part of an identifier.
func isIdentifierChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_'
}

// ParseWithTemporaryClosingParen handles incomplete AST by temporarily inserting a closing parenthesis
// This helps parse incomplete function calls like `foo(x, ` to get a complete AST
// Returns the temporary AST or nil if parsing fails
// The temporary AST should be discarded after use and not stored.
func ParseWithTemporaryClosingParen(text string, line, character int) *ast.Program {
	if text == "" {
		return nil
	}

	lines := strings.Split(text, "\n")
	if line < 0 || line >= len(lines) {
		return nil
	}

	currentLine := lines[line]
	if character < 0 || character > len(currentLine) {
		return nil
	}

	// Create modified text with `)` inserted at cursor position
	var modifiedText strings.Builder

	// Add all lines before the cursor line
	for i := range line {
		modifiedText.WriteString(lines[i])
		modifiedText.WriteString("\n")
	}

	// Add the current line up to cursor, then insert `)`, then rest of line
	modifiedText.WriteString(currentLine[:character])
	modifiedText.WriteString(")")
	modifiedText.WriteString(currentLine[character:])

	// Add remaining lines
	for i := line + 1; i < len(lines); i++ {
		modifiedText.WriteString("\n")
		modifiedText.WriteString(lines[i])
	}

	modifiedSource := modifiedText.String()

	log.Printf("ParseWithTemporaryClosingParen: Attempting to parse modified source\n")

	// Parse the modified text using dwscript engine
	engine, err := dwscript.New()
	if err != nil {
		log.Printf("ParseWithTemporaryClosingParen: Failed to create engine: %v\n", err)
		return nil
	}

	tempProgram, err := engine.Compile(modifiedSource)
	if err != nil {
		log.Printf("ParseWithTemporaryClosingParen: Failed to compile modified source: %v\n", err)
		return nil
	}

	if tempProgram == nil {
		log.Printf("ParseWithTemporaryClosingParen: Compilation returned nil program\n")
		return nil
	}

	tempAST := tempProgram.AST()
	if tempAST == nil {
		log.Printf("ParseWithTemporaryClosingParen: AST is nil\n")
		return nil
	}

	log.Printf("ParseWithTemporaryClosingParen: Successfully parsed modified source\n")

	// Return the temporary AST (caller should discard it after use)
	return tempAST
}

// DetermineCallContextWithTempAST uses a temporary AST with closing paren inserted
// This is useful for incomplete function calls during typing
// Falls back to token-based analysis if temporary parsing fails.
func DetermineCallContextWithTempAST(doc *server.Document, line, character int) (*CallContext, error) {
	buildFallbackContext := func() (*CallContext, error) {
		functionName, err := FindFunctionAtCall(doc, line, character)
		if err != nil || functionName == "" {
			return nil, err
		}

		return &CallContext{
			CallNode:       nil,
			FunctionName:   functionName,
			ObjectExpr:     nil,
			ParameterIndex: findParameterIndexFromText(doc.Text, line, character),
			IsInsideCall:   true,
		}, nil
	}

	// First try the normal approach with the actual AST
	ctx := DetermineCallContext(doc, line, character)
	if ctx != nil {
		return ctx, nil
	}

	log.Printf("DetermineCallContextWithTempAST: Normal approach failed, trying with temporary AST\n")

	// Try parsing with a temporary closing parenthesis
	tempAST := ParseWithTemporaryClosingParen(doc.Text, line, character)
	if tempAST == nil {
		log.Printf("DetermineCallContextWithTempAST: Temporary parsing failed, falling back to token-based analysis\n")
		return buildFallbackContext()
	}

	// Find enclosing call expression in the temporary AST
	astLine := line + 1
	astColumn := character + 1
	callNode := findEnclosingCallExpression(tempAST, astLine, astColumn)

	if callNode == nil {
		log.Printf("DetermineCallContextWithTempAST: No call expression found in temporary AST\n")
		return buildFallbackContext()
	}

	// Extract function name
	functionName, objectExpr := extractFunctionName(callNode)
	if functionName == "" {
		log.Printf("DetermineCallContextWithTempAST: Could not extract function name\n")
		return buildFallbackContext()
	}

	log.Printf("DetermineCallContextWithTempAST: Found call expression: function=%s\n", functionName)

	// Use text-based analysis for parameter index (more reliable than AST during typing)
	paramIndex := findParameterIndex(doc.Text, line, character, callNode)

	// Discard the temporary AST (don't store it)
	// Go's garbage collector will clean it up

	return &CallContext{
		CallNode:       callNode,
		FunctionName:   functionName,
		ObjectExpr:     objectExpr,
		ParameterIndex: paramIndex,
		IsInsideCall:   true,
	}, nil
}

// CountParameterIndex traverses tokens backward to count commas and determine parameter index
// This implements task 10.7 - it scans backward from cursor position character-by-character
// to count commas at the same parenthesis nesting level and returns the active parameter index (0-based).
func CountParameterIndex(text string, line, character int) (int, error) {
	if text == "" {
		return 0, nil
	}

	lines := strings.Split(text, "\n")
	if line < 0 || line >= len(lines) {
		return 0, nil
	}

	currentLine := lines[line]
	if character < 0 || character > len(currentLine) {
		return 0, nil
	}

	textBefore := buildTextBeforeCursor(lines, line, currentLine, character)
	commaCount := countCommasBeforeCursor(textBefore)

	log.Printf("CountParameterIndex: Found %d commas, parameter index = %d\n", commaCount, commaCount)

	// Return comma count as active parameter index (0-based)
	// 0 commas = first parameter (index 0)
	// 1 comma = second parameter (index 1), etc.
	return commaCount, nil
}

// findParameterIndexFromText is a simplified version that just counts commas.
func findParameterIndexFromText(text string, line, character int) int {
	lines := strings.Split(text, "\n")
	if line < 0 || line >= len(lines) {
		return 0
	}

	currentLine := lines[line]
	if character < 0 || character > len(currentLine) {
		return 0
	}

	textBefore := buildTextBeforeCursor(lines, line, currentLine, character)
	return countCommasBeforeCursor(textBefore)
}
