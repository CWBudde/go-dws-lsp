package lsp

import (
	"log"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/CWBudde/go-dws-lsp/internal/analysis"
	"github.com/CWBudde/go-dws-lsp/internal/builtins"
	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// SignatureHelp handles textDocument/signatureHelp requests
// Shows function signatures and parameter hints during function calls
func SignatureHelp(context *glsp.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
	// Get server instance
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in SignatureHelp")
		return nil, nil
	}

	// Extract document URI and position from params
	uri := params.TextDocument.URI
	position := params.Position

	log.Printf("SignatureHelp request: URI=%s, Line=%d, Character=%d\n", uri, position.Line, position.Character)

	// Retrieve document from DocumentStore
	doc, exists := srv.Documents().Get(uri)
	if !exists {
		log.Printf("Document not found: %s\n", uri)
		return nil, nil
	}

	// Check if document and AST are available
	if doc.Program == nil {
		log.Printf("No program available for document: %s\n", uri)
		return nil, nil
	}

	programAST := doc.Program.AST()
	if programAST == nil {
		log.Printf("No AST available for document: %s\n", uri)
		return nil, nil
	}

	// Convert LSP position (0-based, UTF-16) to document position (1-based, UTF-8)
	astLine := int(position.Line) + 1
	astColumn := int(position.Character) + 1

	log.Printf("Converted position: line=%d, column=%d\n", astLine, astColumn)

	// Task 10.4: Detect signature help triggers
	if params.Context != nil {
		triggerKind := params.Context.TriggerKind
		log.Printf("Signature help trigger kind: %d\n", triggerKind)

		// Handle different trigger types
		switch triggerKind {
		case protocol.SignatureHelpTriggerKindInvoked:
			// Manual invocation (Ctrl+Shift+Space)
			log.Println("Signature help manually invoked")

		case protocol.SignatureHelpTriggerKindTriggerCharacter:
			// Triggered by typing a character
			triggerChar := ""
			if params.Context.TriggerCharacter != nil {
				triggerChar = *params.Context.TriggerCharacter
			}

			log.Printf("Signature help triggered by character: '%s'\n", triggerChar)

			// Validate trigger character
			if triggerChar == "(" {
				log.Println("Start of function call")
			} else if triggerChar == "," {
				log.Println("Moving to next parameter")
			} else {
				log.Printf("Warning: Unexpected trigger character: '%s'\n", triggerChar)
			}

		case protocol.SignatureHelpTriggerKindContentChange:
			// Retrigger on typing (content change)
			log.Println("Signature help retriggered on content change")

		default:
			log.Printf("Warning: Unknown trigger kind: %d\n", triggerKind)
		}
	} else {
		log.Println("No trigger context provided")
	}

	// Compute signature help using implemented functions
	signatureHelp := computeSignatureHelp(doc, int(position.Line), int(position.Character), srv)
	return signatureHelp, nil
}

// computeSignatureHelp implements the core signature help logic
func computeSignatureHelp(doc *server.Document, line, character int, srv *server.Server) *protocol.SignatureHelp {
	// Task 10.3 & 10.6: Determine call context (with incomplete AST support)
	callCtx, err := analysis.DetermineCallContextWithTempAST(doc, line, character)
	if err != nil || callCtx == nil {
		log.Printf("computeSignatureHelp: No call context found\n")
		return nil
	}

	log.Printf("computeSignatureHelp: Function='%s', ParameterIndex=%d\n", callCtx.FunctionName, callCtx.ParameterIndex)

	// Tasks 10.8, 10.9 & 10.15: Retrieve function signatures (user-defined or built-in, supports overloading)
	var funcSignatures []*analysis.FunctionSignature

	// First try to get user-defined function signatures (may have multiple overloads)
	funcSignatures, err = analysis.GetFunctionSignatures(doc, callCtx.FunctionName, line, character, srv.WorkspaceIndex())
	if err != nil {
		log.Printf("computeSignatureHelp: Error getting function signatures: %v\n", err)
	}

	// If not found, check built-in functions
	if len(funcSignatures) == 0 {
		builtinSig := builtins.GetBuiltinSignature(callCtx.FunctionName)
		if builtinSig != nil {
			log.Printf("computeSignatureHelp: Found built-in function '%s'\n", callCtx.FunctionName)
			funcSignatures = []*analysis.FunctionSignature{builtinSig}
		}
	}

	if len(funcSignatures) == 0 {
		log.Printf("computeSignatureHelp: Function '%s' not found\n", callCtx.FunctionName)
		return nil
	}

	log.Printf("computeSignatureHelp: Found %d signature(s) for '%s'\n", len(funcSignatures), callCtx.FunctionName)

	// Tasks 10.10-10.12 & 10.15: Construct SignatureHelp response with all signatures
	var signatureInfos []protocol.SignatureInformation
	for _, sig := range funcSignatures {
		sigInfo := buildSignatureInformation(sig)
		if sigInfo != nil {
			signatureInfos = append(signatureInfos, *sigInfo)
		}
	}

	if len(signatureInfos) == 0 {
		return nil
	}

	// Task 10.14 & 10.15: Determine activeSignature based on parameter count
	activeSignature := determineActiveSignature(funcSignatures, callCtx.ParameterIndex)
	activeParameter := uint32(callCtx.ParameterIndex)

	// Clamp activeParameter to valid range for the active signature
	if int(activeSignature) < len(funcSignatures) {
		activeSig := funcSignatures[activeSignature]
		if int(activeParameter) >= len(activeSig.Parameters) {
			if len(activeSig.Parameters) > 0 {
				activeParameter = uint32(len(activeSig.Parameters) - 1)
			} else {
				activeParameter = 0
			}
		}
	}

	signatureHelp := &protocol.SignatureHelp{
		Signatures:      signatureInfos,
		ActiveSignature: &activeSignature,
		ActiveParameter: &activeParameter,
	}

	log.Printf("computeSignatureHelp: Returning %d signature(s), active=%d, activeParam=%d\n",
		len(signatureInfos), activeSignature, activeParameter)

	return signatureHelp
}

// determineActiveSignature selects the best matching signature based on parameter index
// Task 10.15: Match by parameter count - select signature where paramIndex is valid
func determineActiveSignature(signatures []*analysis.FunctionSignature, paramIndex int) uint32 {
	if len(signatures) == 0 {
		return 0
	}

	// If only one signature, return it
	if len(signatures) == 1 {
		return 0
	}

	// Try to find a signature where the current parameter index is valid
	// Prefer signatures with exact or slightly more parameters than the current index
	for i, sig := range signatures {
		if paramIndex < len(sig.Parameters) {
			// This signature can accommodate the current parameter
			return uint32(i)
		}
	}

	// If no signature can accommodate the parameter index,
	// return the signature with the most parameters
	return uint32(len(signatures) - 1)
}

// buildSignatureInformation constructs a SignatureInformation from a FunctionSignature
// Implements tasks 10.10, 10.11, and 10.12
func buildSignatureInformation(funcSig *analysis.FunctionSignature) *protocol.SignatureInformation {
	if funcSig == nil {
		return nil
	}

	// Task 10.11: Format signature label
	label := formatSignatureLabel(funcSig)

	// Task 10.12: Build parameter information array
	parameters := buildParameterInformation(funcSig, label)

	// Create SignatureInformation
	sigInfo := &protocol.SignatureInformation{
		Label:      label,
		Parameters: parameters,
	}

	// Add documentation if available
	if funcSig.Documentation != "" {
		doc := protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: funcSig.Documentation,
		}
		sigInfo.Documentation = &doc
	}

	return sigInfo
}

// formatSignatureLabel formats a function signature label
// Implements task 10.11
// Example: "function Calculate(x: Integer, y: Integer): Integer"
func formatSignatureLabel(funcSig *analysis.FunctionSignature) string {
	var sb strings.Builder

	// Start with "function" or "procedure"
	if funcSig.ReturnType != "" {
		sb.WriteString("function ")
	} else {
		sb.WriteString("procedure ")
	}

	// Add function name
	sb.WriteString(funcSig.Name)

	// Add opening parenthesis
	sb.WriteString("(")

	// Add parameters
	for i, param := range funcSig.Parameters {
		if i > 0 {
			sb.WriteString(", ")
		}

		// Format parameter: name: Type
		sb.WriteString(param.Name)
		if param.Type != "" {
			sb.WriteString(": ")
			sb.WriteString(param.Type)
		}

		// Add default value if present
		if param.IsOptional && param.DefaultValue != "" {
			sb.WriteString(" = ")
			sb.WriteString(param.DefaultValue)
		}
	}

	// Add closing parenthesis
	sb.WriteString(")")

	// Add return type if it's a function (not a procedure)
	if funcSig.ReturnType != "" {
		sb.WriteString(": ")
		sb.WriteString(funcSig.ReturnType)
	}

	return sb.String()
}

// buildParameterInformation builds an array of ParameterInformation
// Implements task 10.12
func buildParameterInformation(funcSig *analysis.FunctionSignature, label string) []protocol.ParameterInformation {
	if len(funcSig.Parameters) == 0 {
		return nil
	}

	parameters := make([]protocol.ParameterInformation, 0, len(funcSig.Parameters))

	// Find each parameter substring in the label
	// We need to locate "name: Type" for each parameter
	for _, param := range funcSig.Parameters {
		// Build the parameter label substring
		paramLabel := param.Name
		if param.Type != "" {
			paramLabel += ": " + param.Type
		}

		// Find the position of this parameter in the full label
		// We use the substring as the label (LSP supports both string and [start, end])
		paramInfo := protocol.ParameterInformation{
			Label: paramLabel,
		}

		// Could add parameter documentation here if we extract it from comments
		// For now, leave it empty

		parameters = append(parameters, paramInfo)
	}

	return parameters
}
