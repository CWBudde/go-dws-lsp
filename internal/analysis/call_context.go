package analysis

import (
	"log"
	"strings"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"

	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// CallContext holds information about a function call at the cursor position
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
// Returns nil if the cursor is not inside a function call
func DetermineCallContext(doc *server.Document, line, character int) (*CallContext, error) {
	if doc.Program == nil {
		return nil, nil
	}

	programAST := doc.Program.AST()
	if programAST == nil {
		return nil, nil
	}

	// Convert LSP position (0-based) to AST position (1-based)
	astLine := line + 1
	astColumn := character + 1

	log.Printf("DetermineCallContext: line=%d, character=%d (AST: %d:%d)\n", line, character, astLine, astColumn)

	// Find enclosing call expression using AST traversal
	callNode := findEnclosingCallExpression(programAST, astLine, astColumn)
	if callNode == nil {
		log.Printf("No enclosing call expression found\n")
		return nil, nil
	}

	// Extract function name based on node type
	functionName, objectExpr := extractFunctionName(callNode)
	if functionName == "" {
		log.Printf("Could not extract function name from call node\n")
		return nil, nil
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
	}, nil
}

// findEnclosingCallExpression traverses the AST to find the innermost CallExpression
// or MethodCallExpression that contains the given position
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
// to determine which parameter the cursor is on (0-based index)
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

	// Find the opening parenthesis in the text
	// We need to scan from the start of the call to find '('
	var textBeforeCursor strings.Builder

	// Get all text from start of document up to cursor
	for i := 0; i < line; i++ {
		textBeforeCursor.WriteString(lines[i])
		textBeforeCursor.WriteString("\n")
	}
	textBeforeCursor.WriteString(currentLine[:character])

	textBefore := textBeforeCursor.String()

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
		if r == ')' {
			parenDepth++
		} else if r == '(' {
			if parenDepth == 0 {
				// Found the opening parenthesis of our call
				foundOpenParen = true
				break
			}
			parenDepth--
		} else if r == ']' {
			bracketDepth++
		} else if r == '[' {
			if bracketDepth > 0 {
				bracketDepth--
			}
		} else if r == ',' {
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

	log.Printf("Counted %d commas before cursor\n", commaCount)

	// Parameter index is the number of commas we found
	// (0 commas = first parameter, 1 comma = second parameter, etc.)
	return commaCount
}

// extractFunctionName extracts the function or method name from a call expression node
// Returns the name and optionally the object expression (for method calls)
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
