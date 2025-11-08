package analysis

import (
	"fmt"
	"log"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/dwscript"
	"github.com/cwbudde/go-dws/pkg/token"
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
//   - *dwscript.Program: The compiled program (nil if compilation failed)
//   - []protocol.Diagnostic: List of syntax and semantic errors as LSP diagnostics
//   - error: Critical error that prevented parsing (e.g., engine creation failed)
func ParseDocument(text string, filename string) (*dwscript.Program, []protocol.Diagnostic, error) {
	// Create a new DWScript engine
	engine, err := dwscript.New()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create DWScript engine: %w", err)
	}

	log.Printf("Parsing document: %s (%d bytes)", filename, len(text))

	// Attempt to compile the source code
	// This will return a CompileError if there are syntax or semantic errors
	program, err := engine.Compile(text)

	var diagnostics []protocol.Diagnostic

	if err != nil {
		// Check if it's a compile error with structured errors
		if compileErr, ok := err.(*dwscript.CompileError); ok {
			log.Printf("Compilation failed for %s: %d errors", filename, len(compileErr.Errors))
			diagnostics = convertStructuredErrors(compileErr.Errors)
		} else {
			// Some other unexpected error
			return nil, nil, fmt.Errorf("unexpected error during compilation: %w", err)
		}
	} else {
		log.Printf("Compilation successful for %s", filename)
		diagnostics = []protocol.Diagnostic{}
	}

	// Perform additional validation for unsupported DWScript constructs (e.g., function overloading)
	if program != nil {
		extraDiagnostics := detectUnsupportedFunctionOverloads(program)
		if len(extraDiagnostics) > 0 {
			diagnostics = append(diagnostics, extraDiagnostics...)
		}
	}

	return program, diagnostics, nil
}

// convertStructuredErrors converts go-dws structured Error objects to LSP Diagnostic objects.
// This uses the new Phase 2 structured error format with position information already included.
func convertStructuredErrors(errors []*dwscript.Error) []protocol.Diagnostic {
	if len(errors) == 0 {
		return []protocol.Diagnostic{}
	}

	diagnostics := make([]protocol.Diagnostic, 0, len(errors))

	for _, err := range errors {
		diagnostic := convertStructuredError(err)
		diagnostics = append(diagnostics, diagnostic)
	}

	return diagnostics
}

// convertStructuredError converts a single go-dws Error to an LSP Diagnostic.
// The Error already contains structured position information (Line, Column, Length)
// using 1-based indexing, which we convert to LSP's 0-based indexing.
func convertStructuredError(err *dwscript.Error) protocol.Diagnostic {
	// Convert from 1-based (DWScript) to 0-based (LSP) indexing
	lspLine := uint32(0)
	if err.Line > 0 {
		lspLine = uint32(err.Line - 1)
	}

	lspCol := uint32(0)
	if err.Column > 0 {
		lspCol = uint32(err.Column - 1)
	}

	// Calculate end position
	length := err.Length
	if length <= 0 {
		length = 1 // Default to single character if no length specified
	}
	endCol := lspCol + uint32(length)

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

	// Map DWScript severity to LSP DiagnosticSeverity
	severity := mapSeverity(err.Severity)

	// Map diagnostic tags based on error code
	var tags []protocol.DiagnosticTag
	if err.Code != "" {
		tags = mapDiagnosticTags(err.Code)
	}

	// Create the diagnostic
	diagnostic := protocol.Diagnostic{
		Range:    diagRange,
		Severity: &severity,
		Source:   stringPtr("go-dws"),
		Message:  err.Message,
	}

	// Add error code if available
	if err.Code != "" {
		code := protocol.IntegerOrString{Value: err.Code}
		diagnostic.Code = &code
	}

	// Add tags if any
	if len(tags) > 0 {
		diagnostic.Tags = tags
	}

	return diagnostic
}

// mapSeverity maps go-dws ErrorSeverity to LSP DiagnosticSeverity
func mapSeverity(severity dwscript.ErrorSeverity) protocol.DiagnosticSeverity {
	switch severity {
	case dwscript.SeverityError:
		return protocol.DiagnosticSeverityError
	case dwscript.SeverityWarning:
		return protocol.DiagnosticSeverityWarning
	case dwscript.SeverityInfo:
		return protocol.DiagnosticSeverityInformation
	case dwscript.SeverityHint:
		return protocol.DiagnosticSeverityHint
	default:
		return protocol.DiagnosticSeverityError
	}
}

// mapDiagnosticTags maps error codes to LSP DiagnosticTag values
func mapDiagnosticTags(code string) []protocol.DiagnosticTag {
	var tags []protocol.DiagnosticTag

	switch code {
	case "W_UNUSED_VAR", "W_UNUSED_PARAM", "W_UNUSED_FUNCTION":
		// Mark unused code with Unnecessary tag
		tags = append(tags, protocol.DiagnosticTagUnnecessary)
	case "W_DEPRECATED":
		// Mark deprecated code with Deprecated tag
		tags = append(tags, protocol.DiagnosticTagDeprecated)
	}

	return tags
}

// detectUnsupportedFunctionOverloads emits diagnostics when the document declares
// multiple global functions with the same name. DWScript does not support
// function overloading, so we proactively flag this scenario even if the compiler
// succeeds.
func detectUnsupportedFunctionOverloads(program *dwscript.Program) []protocol.Diagnostic {
	if program == nil {
		return nil
	}

	root := program.AST()
	if root == nil {
		return nil
	}

	seen := make(map[string]struct{})
	diagnostics := make([]protocol.Diagnostic, 0)

	ast.Inspect(root, func(node ast.Node) bool {
		fn, ok := node.(*ast.FunctionDecl)
		if !ok || fn == nil || fn.Name == nil {
			return true
		}

		// Skip methods for now – this validation only targets global functions.
		if fn.ClassName != nil {
			return true
		}

		name := fn.Name.Value
		if _, exists := seen[name]; exists {
			diagnostics = append(diagnostics, createFunctionRedeclarationDiagnostic(name, fn.Name.Pos()))
			return true
		}

		seen[name] = struct{}{}
		return true
	})

	return diagnostics
}

func createFunctionRedeclarationDiagnostic(name string, pos token.Position) protocol.Diagnostic {
	severity := protocol.DiagnosticSeverityError

	message := fmt.Sprintf("Function '%s' is redeclared. DWScript does not support function overloading.", name)

	return protocol.Diagnostic{
		Range:    identifierRange(pos, name),
		Severity: &severity,
		Source:   stringPtr("go-dws"),
		Message:  message,
	}
}

func identifierRange(pos token.Position, identifier string) protocol.Range {
	line := uint32(0)
	if pos.Line > 0 {
		line = uint32(pos.Line - 1)
	}

	startChar := uint32(0)
	if pos.Column > 0 {
		startChar = uint32(pos.Column - 1)
	}

	tokenLength := utf16Length(identifier)
	if tokenLength <= 0 {
		tokenLength = 1
	}

	endChar := startChar + uint32(tokenLength)

	return protocol.Range{
		Start: protocol.Position{Line: line, Character: startChar},
		End:   protocol.Position{Line: line, Character: endChar},
	}
}

// stringPtr is a helper function to create a pointer to a string.
func stringPtr(s string) *string {
	return &s
}
