package analysis

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/pkg/dwscript"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// ParseDocument parses DWScript source code and returns diagnostics.
// It uses the go-dws engine to compile the code and extract syntax/semantic errors.
//
// Parameters:
//   - text: The source code text to parse
//   - filename: The filename for error reporting (typically the URI)
//
// Returns:
//   - []protocol.Diagnostic: List of syntax and semantic errors as LSP diagnostics
//   - error: Critical error that prevented parsing (e.g., engine creation failed)
func ParseDocument(text string, filename string) ([]protocol.Diagnostic, error) {
	// Create a new DWScript engine
	engine, err := dwscript.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create DWScript engine: %w", err)
	}

	log.Printf("Parsing document: %s (%d bytes)", filename, len(text))

	// Attempt to compile the source code
	// This will return a CompileError if there are syntax or semantic errors
	_, err = engine.Compile(text)

	var diagnostics []protocol.Diagnostic

	if err != nil {
		// Check if it's a compile error
		if compileErr, ok := err.(*dwscript.CompileError); ok {
			log.Printf("Compilation failed for %s: %d errors", filename, len(compileErr.Errors))
			diagnostics = convertCompileErrors(compileErr.Errors, text)
		} else {
			// Some other unexpected error
			return nil, fmt.Errorf("unexpected error during compilation: %w", err)
		}
	} else {
		log.Printf("Compilation successful for %s", filename)
		diagnostics = []protocol.Diagnostic{}
	}

	return diagnostics, nil
}

// convertCompileErrors converts DWScript compile error messages to LSP Diagnostic objects.
// It parses error messages to extract position information.
func convertCompileErrors(errorMessages []string, source string) []protocol.Diagnostic {
	if len(errorMessages) == 0 {
		return []protocol.Diagnostic{}
	}

	diagnostics := make([]protocol.Diagnostic, 0, len(errorMessages))

	for _, errMsg := range errorMessages {
		diagnostic := parseErrorMessage(errMsg, source)
		diagnostics = append(diagnostics, diagnostic)
	}

	return diagnostics
}

// parseErrorMessage attempts to extract position information from an error message
// and create an LSP Diagnostic.
//
// DWScript error messages typically follow patterns like:
//   - "Syntax Error: <message> [line X]"
//   - "Error at line X, col Y: <message>"
//   - "<message> (X,Y)"
func parseErrorMessage(errMsg string, source string) protocol.Diagnostic {
	// Try to extract line and column information using various patterns

	// Pattern 1: [line X] at the end
	lineRegex1 := regexp.MustCompile(`\[line (\d+)\]`)
	if matches := lineRegex1.FindStringSubmatch(errMsg); len(matches) > 1 {
		line, _ := strconv.Atoi(matches[1])
		return createDiagnostic(line, 0, 0, errMsg)
	}

	// Pattern 2: "line X, col Y" or "line X:Y"
	lineColRegex := regexp.MustCompile(`line (\d+)[,:]\s*col(?:umn)?\s*(\d+)`)
	if matches := lineColRegex.FindStringSubmatch(errMsg); len(matches) > 2 {
		line, _ := strconv.Atoi(matches[1])
		col, _ := strconv.Atoi(matches[2])
		return createDiagnostic(line, col, 0, errMsg)
	}

	// Pattern 3: (line,col) format
	parenRegex := regexp.MustCompile(`\((\d+),(\d+)\)`)
	if matches := parenRegex.FindStringSubmatch(errMsg); len(matches) > 2 {
		line, _ := strconv.Atoi(matches[1])
		col, _ := strconv.Atoi(matches[2])
		return createDiagnostic(line, col, 0, errMsg)
	}

	// Pattern 4: Just "line X" anywhere in the message
	lineOnlyRegex := regexp.MustCompile(`line (\d+)`)
	if matches := lineOnlyRegex.FindStringSubmatch(errMsg); len(matches) > 1 {
		line, _ := strconv.Atoi(matches[1])
		return createDiagnostic(line, 0, 0, errMsg)
	}

	// If no position info found, create diagnostic at line 0
	return createDiagnostic(0, 0, 0, errMsg)
}

// createDiagnostic creates an LSP Diagnostic from position information and message.
// It converts from 1-based (DWScript) to 0-based (LSP) line/column indexing.
func createDiagnostic(line, col, length int, message string) protocol.Diagnostic {
	// Convert from 1-based to 0-based indexing
	lspLine := uint32(0)
	if line > 0 {
		lspLine = uint32(line - 1)
	}

	lspCol := uint32(0)
	if col > 0 {
		lspCol = uint32(col - 1)
	}

	// Default length if not specified
	if length <= 0 {
		length = 1
	}

	endCol := lspCol + uint32(length)

	// Clean up the message by removing position information
	cleanMessage := cleanErrorMessage(message)

	// Create the diagnostic range
	diagRange := protocol.Range{
		Start: protocol.Position{
			Line:      lspLine,
			Character: lspCol,
		},
		End: protocol.Position{
			Line:      lspLine,
			Character: endCol,
		},
	}

	// Determine severity based on message content
	severity := protocol.DiagnosticSeverityError
	if strings.Contains(strings.ToLower(message), "warning") {
		severity = protocol.DiagnosticSeverityWarning
	}

	// Create the diagnostic
	diagnostic := protocol.Diagnostic{
		Range:    diagRange,
		Severity: &severity,
		Source:   stringPtr("go-dws"),
		Message:  cleanMessage,
	}

	return diagnostic
}

// cleanErrorMessage removes position information from the error message
// to avoid redundancy (position is shown separately in the IDE).
func cleanErrorMessage(message string) string {
	// Remove [line X] suffix
	message = regexp.MustCompile(`\s*\[line \d+\]\s*$`).ReplaceAllString(message, "")

	// Remove (line,col) patterns
	message = regexp.MustCompile(`\s*\(\d+,\d+\)\s*`).ReplaceAllString(message, " ")

	// Remove "line X, col Y:" prefix
	message = regexp.MustCompile(`^line \d+[,:]\s*col(?:umn)?\s*\d+:\s*`).ReplaceAllString(message, "")

	// Remove "Error at line X, col Y:" prefix
	message = regexp.MustCompile(`^Error at line \d+[,:]\s*col(?:umn)?\s*\d+:\s*`).ReplaceAllString(message, "")

	return strings.TrimSpace(message)
}

// stringPtr is a helper function to create a pointer to a string.
func stringPtr(s string) *string {
	return &s
}
