package lsp

import (
	"log"
	"strings"

	"github.com/CWBudde/go-dws-lsp/internal/analysis"
	"github.com/CWBudde/go-dws-lsp/internal/builtins"
	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// SignatureHelp handles textDocument/signatureHelp requests
// Shows function signatures and parameter hints during function calls.
func SignatureHelp(context *glsp.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
	uri := params.TextDocument.URI
	position := params.Position

	log.Printf("SignatureHelp request: URI=%s, Line=%d, Character=%d\n", uri, position.Line, position.Character)

	// Validate and get document
	srv, doc := validateSignatureHelpRequest(uri)
	if srv == nil || doc == nil {
		return nil, nil //nolint:nilnil // nil is valid LSP response
	}

	// Log trigger information
	logSignatureHelpTrigger(params.Context)

	// Compute signature help using implemented functions
	signatureHelp := computeSignatureHelp(doc, int(position.Line), int(position.Character), srv)

	return signatureHelp, nil
}

// computeSignatureHelp implements the core signature help logic.
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
	// Pass the temporary Program from CallContext if available
	funcSignatures, err = analysis.GetFunctionSignatures(doc, callCtx.FunctionName, line, character, srv.WorkspaceIndex(), callCtx.TempProgram)
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

// validateSignatureHelpRequest validates the request and returns server and document.
func validateSignatureHelpRequest(uri string) (*server.Server, *server.Document) {
	srv, ok := serverInstance.(*server.Server)
	if !ok || srv == nil {
		log.Println("Warning: server instance not available in SignatureHelp")
		return nil, nil
	}

	doc, exists := srv.Documents().Get(uri)
	if !exists {
		log.Printf("Document not found: %s\n", uri)
		return nil, nil
	}

	// Allow signature help even when doc.Program is nil or has no AST
	// The temporary AST parsing in DetermineCallContextWithTempAST will handle incomplete code

	return srv, doc
}

// logSignatureHelpTrigger logs information about the signature help trigger.
func logSignatureHelpTrigger(context *protocol.SignatureHelpContext) {
	if context == nil {
		log.Println("No trigger context provided")
		return
	}

	triggerKind := context.TriggerKind
	log.Printf("Signature help trigger kind: %d\n", triggerKind)

	switch triggerKind {
	case protocol.SignatureHelpTriggerKindInvoked:
		log.Println("Signature help manually invoked")
	case protocol.SignatureHelpTriggerKindTriggerCharacter:
		triggerChar := ""
		if context.TriggerCharacter != nil {
			triggerChar = *context.TriggerCharacter
		}
		log.Printf("Signature help triggered by character: '%s'\n", triggerChar)
		switch triggerChar {
		case "(":
			log.Println("Start of function call")
		case ",":
			log.Println("Moving to next parameter")
		default:
			log.Printf("Warning: Unexpected trigger character: '%s'\n", triggerChar)
		}
	case protocol.SignatureHelpTriggerKindContentChange:
		log.Println("Signature help retriggered on content change")
	default:
		log.Printf("Warning: Unknown trigger kind: %d\n", triggerKind)
	}
}

// determineActiveSignature selects the best matching signature based on parameter index
// Task 10.15: Match by parameter count - select signature where paramIndex is valid.
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
// Implements tasks 10.10, 10.11, and 10.12.
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
// Example: "function Calculate(x: Integer, y: Integer): Integer".
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
// Implements task 10.12.
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
