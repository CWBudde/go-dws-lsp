// Package analysis provides scope-based completion for DWScript.
package analysis

import (
	"log"
	"strconv"
	"strings"

	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/cwbudde/go-dws/pkg/ast"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// controlStructureSnippets contains keyword completion snippets for control structures.
var controlStructureSnippets = map[string]struct {
	snippet string
	detail  string
}{
	"if": {
		snippet: "if ${1:condition} then\n\t$0\nend;",
		detail:  "if-then-end statement",
	},
	"for": {
		snippet: "for ${1:i} := ${2:0} to ${3:10} do\n\t$0\nend;",
		detail:  "for-to-do loop",
	},
	"while": {
		snippet: "while ${1:condition} do\n\t$0\nend;",
		detail:  "while-do loop",
	},
	"repeat": {
		snippet: "repeat\n\t$0\nuntil ${1:condition};",
		detail:  "repeat-until loop",
	},
	"case": {
		snippet: "case ${1:expression} of\n\t${2:value}: $0\nend;",
		detail:  "case-of statement",
	},
	"try": {
		snippet: "try\n\t$0\nexcept\n\ton E: Exception do\n\t\tRaise;\nend;",
		detail:  "try-except block",
	},
	"function": {
		snippet: "function ${1:FunctionName}(${2:params}): ${3:ReturnType};\nbegin\n\t$0\nend;",
		detail:  "function declaration",
	},
	"procedure": {
		snippet: "procedure ${1:ProcedureName}(${2:params});\nbegin\n\t$0\nend;",
		detail:  "procedure declaration",
	},
	"class": {
		snippet: "class ${1:ClassName}\nprivate\n\t$0\npublic\nend;",
		detail:  "class declaration",
	},
}

// simpleKeywordsList contains basic DWScript keywords without custom snippets.
var simpleKeywordsList = []string{
	"begin", "end", "then", "else", "do", "to", "downto", "until", "of",
	"var", "const", "type", "record", "interface", "implementation", "uses",
	"unit", "program", "except", "finally", "raise", "on", "as", "is", "in",
	"not", "and", "or", "xor", "div", "mod", "shl", "shr", "array", "set",
	"property", "read", "write", "private", "protected", "public", "published",
	"constructor", "destructor", "inherited", "nil", "true", "false", "exit",
	"break", "continue", "with",
}

// builtInTypesList contains DWScript built-in types.
var builtInTypesList = []string{
	"Integer", "Float", "String", "Boolean", "Variant",
	"TObject", "TClass", "DateTime", "Currency",
	"Byte", "Word", "Cardinal", "Int64", "UInt64",
	"Single", "Double", "Extended", "Char",
}

// builtInFunctionsMap contains DWScript built-in functions with their signatures.
var builtInFunctionsMap = map[string]string{
	"Print":          "Print(value: Variant)",
	"PrintLn":        "PrintLn(value: Variant)",
	"Length":         "Length(s: String): Integer",
	"Copy":           "Copy(s: String, index, count: Integer): String",
	"Pos":            "Pos(substr, str: String): Integer",
	"UpperCase":      "UpperCase(s: String): String",
	"LowerCase":      "LowerCase(s: String): String",
	"Trim":           "Trim(s: String): String",
	"IntToStr":       "IntToStr(value: Integer): String",
	"StrToInt":       "StrToInt(s: String): Integer",
	"FloatToStr":     "FloatToStr(value: Float): String",
	"StrToFloat":     "StrToFloat(s: String): Float",
	"Now":            "Now(): DateTime",
	"Date":           "Date(): DateTime",
	"Time":           "Time(): DateTime",
	"FormatDateTime": "FormatDateTime(format: String, dt: DateTime): String",
	"Inc":            "Inc(var x: Integer; increment: Integer = 1)",
	"Dec":            "Dec(var x: Integer; decrement: Integer = 1)",
	"Chr":            "Chr(code: Integer): Char",
	"Ord":            "Ord(ch: Char): Integer",
	"Round":          "Round(value: Float): Integer",
	"Trunc":          "Trunc(value: Float): Integer",
	"Abs":            "Abs(value: Float): Float",
	"Sqrt":           "Sqrt(value: Float): Float",
	"Sqr":            "Sqr(value: Float): Float",
	"Sin":            "Sin(angle: Float): Float",
	"Cos":            "Cos(angle: Float): Float",
	"Tan":            "Tan(angle: Float): Float",
	"Exp":            "Exp(value: Float): Float",
	"Ln":             "Ln(value: Float): Float",
	"Random":         "Random(): Float",
	"Randomize":      "Randomize()",
}

// CollectScopeCompletions gathers all completion items available in the current scope.
// This includes keywords, local variables, parameters, global symbols, and built-in functions.
// Task 9.17: Uses caching for keywords, built-ins, and global symbols.
func CollectScopeCompletions(doc *server.Document, cache *server.CompletionCache, line, character int) ([]protocol.CompletionItem, error) {
	log.Printf("CollectScopeCompletions: gathering completions at %d:%d", line, character)

	var items []protocol.CompletionItem
	var keywords, builtins, globalItems []protocol.CompletionItem

	// Try to get from cache first (task 9.17)
	var cached *server.CachedCompletionItems
	if cache != nil {
		cached = cache.GetCachedItems(doc.URI, int32(doc.Version))
	}

	if cached != nil && len(cached.Keywords) > 0 {
		log.Printf("CollectScopeCompletions: using cached keywords, builtins, and global symbols")

		keywords = cached.Keywords
		builtins = cached.Builtins
		globalItems = cached.GlobalSymbols
	} else {
		// Cache miss - compute all items
		keywords = getKeywordCompletions()
		builtins = getBuiltInCompletions()

		if doc.Program != nil && doc.Program.AST() != nil {
			globalItems = getGlobalCompletions(doc.Program.AST())
		}

		// Cache for future requests
		if cache != nil && doc.Program != nil {
			cache.SetCachedItems(doc.URI, int32(doc.Version), &server.CachedCompletionItems{
				Keywords:      keywords,
				Builtins:      builtins,
				GlobalSymbols: globalItems,
			})
			log.Printf("CollectScopeCompletions: cached completion items for %s (version %d)",
				doc.URI, doc.Version)
		}
	}

	// Add keywords
	items = append(items, keywords...)

	if doc.Program == nil || doc.Program.AST() == nil {
		log.Println("CollectScopeCompletions: no AST available, returning keywords only")
		return items, nil
	}

	program := doc.Program.AST()

	// Convert LSP position (0-based) to AST position (1-based)
	astLine := line + 1
	astColumn := character + 1

	// Task 9.9: Add local variables and parameters (not cached - varies by position)
	localItems := getLocalCompletions(program, astLine, astColumn)
	items = append(items, localItems...)

	// Add global symbols
	items = append(items, globalItems...)

	// Add built-in functions and types
	items = append(items, builtins...)

	log.Printf("CollectScopeCompletions: found %d total completion items", len(items))

	return items, nil
}

// getKeywordCompletions returns completion items for DWScript keywords.
// Task 9.15: Includes snippet support for control structures.
func getKeywordCompletions() []protocol.CompletionItem {
	items := make([]protocol.CompletionItem, 0, 40)
	kind := protocol.CompletionItemKindKeyword

	// Add control structures with snippets
	for keyword, info := range controlStructureSnippets {
		sortText := "~keyword~" + keyword
		detail := info.detail
		insertTextFormat := protocol.InsertTextFormatSnippet

		item := protocol.CompletionItem{
			Label:            keyword,
			Kind:             &kind,
			Detail:           &detail,
			InsertText:       &info.snippet,
			InsertTextFormat: &insertTextFormat,
			SortText:         &sortText,
		}
		items = append(items, item)
	}

	// Add simple keywords without snippets
	plainTextFormat := protocol.InsertTextFormatPlainText

	for _, keyword := range simpleKeywordsList {
		detail := "DWScript keyword"
		sortText := "~keyword~" + keyword
		item := protocol.CompletionItem{
			Label:            keyword,
			Kind:             &kind,
			Detail:           &detail,
			InsertText:       &keyword,
			InsertTextFormat: &plainTextFormat,
			SortText:         &sortText,
		}
		items = append(items, item)
	}

	return items
}

// getLocalCompletions returns completion items for local variables and parameters.
func getLocalCompletions(program *ast.Program, line, column int) []protocol.CompletionItem {
	items := make([]protocol.CompletionItem, 0, 10)

	// Find the enclosing function at the cursor position
	enclosingFunc := findEnclosingFunctionAt(program, line, column)
	if enclosingFunc == nil {
		// Not inside a function, no local scope
		return items
	}

	// Add function parameters
	paramKind := protocol.CompletionItemKindVariable
	plainTextFormat := protocol.InsertTextFormatPlainText

	for _, param := range enclosingFunc.Parameters {
		if param.Name == nil {
			continue
		}

		detail := "Parameter"
		if param.Type != nil {
			detail = "Parameter: " + param.Type.String()
		}

		// Use sortText to prioritize local symbols
		sortText := "0param~" + param.Name.Value

		item := protocol.CompletionItem{
			Label:            param.Name.Value,
			Kind:             &paramKind,
			Detail:           &detail,
			SortText:         &sortText,
			InsertTextFormat: &plainTextFormat,
		}
		items = append(items, item)
	}

	// Add local variables from function body
	if enclosingFunc.Body != nil {
		localVars := extractLocalVariables(enclosingFunc.Body)
		items = append(items, localVars...)
	}

	return items
}

// findEnclosingFunctionAt finds the function that contains the given position.
func findEnclosingFunctionAt(program *ast.Program, line, column int) *ast.FunctionDecl {
	var enclosingFunc *ast.FunctionDecl

	ast.Inspect(program, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		if funcDecl, ok := node.(*ast.FunctionDecl); ok {
			// Check if the cursor is within this function's range
			if isPositionInNodeRange(node, line, column) {
				enclosingFunc = funcDecl
				// Continue traversal to find inner functions if nested
			}
		}

		return true
	})

	return enclosingFunc
}

// extractLocalVariables extracts all local variable declarations from a block.
func extractLocalVariables(block *ast.BlockStatement) []protocol.CompletionItem {
	var items []protocol.CompletionItem
	kind := protocol.CompletionItemKindVariable
	plainTextFormat := protocol.InsertTextFormatPlainText

	if block == nil {
		return items
	}

	// Traverse statements in the block to find variable declarations
	for _, stmt := range block.Statements {
		if varDecl, ok := stmt.(*ast.VarDeclStatement); ok {
			for _, name := range varDecl.Names {
				detail := "Local variable"
				if varDecl.Type != nil {
					detail = "Local variable: " + varDecl.Type.String()
				}

				// Use sortText to prioritize local variables
				sortText := "0local~" + name.Value

				item := protocol.CompletionItem{
					Label:            name.Value,
					Kind:             &kind,
					Detail:           &detail,
					SortText:         &sortText,
					InsertTextFormat: &plainTextFormat,
				}
				items = append(items, item)
			}
		}
	}

	return items
}

// getGlobalCompletions returns completion items for global symbols.
func getGlobalCompletions(program *ast.Program) []protocol.CompletionItem {
	var items []protocol.CompletionItem

	// Extract top-level declarations
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.FunctionDecl:
			if s.Name == nil {
				continue
			}

			kind := protocol.CompletionItemKindFunction
			signature := buildFunctionSignature(s)
			sortText := "1global~" + s.Name.Value

			// Build snippet for function with parameters
			insertText, insertTextFormat := buildFunctionSnippet(s)

			item := protocol.CompletionItem{
				Label:            s.Name.Value,
				Kind:             &kind,
				Detail:           &signature,
				SortText:         &sortText,
				InsertText:       &insertText,
				InsertTextFormat: &insertTextFormat,
			}
			items = append(items, item)

		case *ast.ClassDecl:
			if s.Name == nil {
				continue
			}

			kind := protocol.CompletionItemKindClass
			detail := "Class"
			sortText := "1global~" + s.Name.Value
			plainTextFormat := protocol.InsertTextFormatPlainText
			item := protocol.CompletionItem{
				Label:            s.Name.Value,
				Kind:             &kind,
				Detail:           &detail,
				SortText:         &sortText,
				InsertTextFormat: &plainTextFormat,
			}
			items = append(items, item)

		case *ast.RecordDecl:
			if s.Name == nil {
				continue
			}

			kind := protocol.CompletionItemKindStruct
			detail := "Record"
			sortText := "1global~" + s.Name.Value
			plainTextFormat := protocol.InsertTextFormatPlainText
			item := protocol.CompletionItem{
				Label:            s.Name.Value,
				Kind:             &kind,
				Detail:           &detail,
				SortText:         &sortText,
				InsertTextFormat: &plainTextFormat,
			}
			items = append(items, item)

		case *ast.InterfaceDecl:
			if s.Name == nil {
				continue
			}

			kind := protocol.CompletionItemKindInterface
			detail := "Interface"
			sortText := "1global~" + s.Name.Value
			plainTextFormat := protocol.InsertTextFormatPlainText
			item := protocol.CompletionItem{
				Label:            s.Name.Value,
				Kind:             &kind,
				Detail:           &detail,
				SortText:         &sortText,
				InsertTextFormat: &plainTextFormat,
			}
			items = append(items, item)

		case *ast.VarDeclStatement:
			kind := protocol.CompletionItemKindVariable
			plainTextFormat := protocol.InsertTextFormatPlainText

			for _, name := range s.Names {
				detail := "Global variable"
				if s.Type != nil {
					detail = "Global variable: " + s.Type.String()
				}

				sortText := "1global~" + name.Value
				item := protocol.CompletionItem{
					Label:            name.Value,
					Kind:             &kind,
					Detail:           &detail,
					SortText:         &sortText,
					InsertTextFormat: &plainTextFormat,
				}
				items = append(items, item)
			}

		case *ast.ConstDecl:
			if s.Name == nil {
				continue
			}

			kind := protocol.CompletionItemKindConstant

			detail := "Constant"
			if s.Type != nil {
				detail = "Constant: " + s.Type.String()
			}

			if s.Value != nil {
				detail += " = " + s.Value.String()
			}

			sortText := "1global~" + s.Name.Value
			plainTextFormat := protocol.InsertTextFormatPlainText
			item := protocol.CompletionItem{
				Label:            s.Name.Value,
				Kind:             &kind,
				Detail:           &detail,
				SortText:         &sortText,
				InsertTextFormat: &plainTextFormat,
			}
			items = append(items, item)

		case *ast.EnumDecl:
			if s.Name == nil {
				continue
			}

			kind := protocol.CompletionItemKindEnum
			detail := "Enumeration"
			sortText := "1global~" + s.Name.Value
			plainTextFormat := protocol.InsertTextFormatPlainText
			item := protocol.CompletionItem{
				Label:            s.Name.Value,
				Kind:             &kind,
				Detail:           &detail,
				SortText:         &sortText,
				InsertTextFormat: &plainTextFormat,
			}
			items = append(items, item)

			// Also add enum values as constants
			constKind := protocol.CompletionItemKindEnumMember

			for _, enumVal := range s.Values {
				enumDetail := "Enum value: " + s.Name.Value
				enumSortText := "1global~" + enumVal.Name
				enumItem := protocol.CompletionItem{
					Label:            enumVal.Name,
					Kind:             &constKind,
					Detail:           &enumDetail,
					SortText:         &enumSortText,
					InsertTextFormat: &plainTextFormat,
				}
				items = append(items, enumItem)
			}
		}
	}

	return items
}

// buildFunctionSignature builds a function signature string for display.
func buildFunctionSignature(fn *ast.FunctionDecl) string {
	if fn.Name == nil {
		return ""
	}

	var signature strings.Builder
	signature.WriteString(fn.Name.Value + "(")

	var signatureSb451 strings.Builder

	for i, param := range fn.Parameters {
		if i > 0 {
			signatureSb451.WriteString(", ")
		}

		if param.ByRef {
			signatureSb451.WriteString("var ")
		} else if param.IsConst {
			signature.WriteString("const ")
		} else if param.IsLazy {
			signature.WriteString("lazy ")
		}

		if param.Name != nil {
			signatureSb451.WriteString(param.Name.Value)
		}

		if param.Type != nil {
			signatureSb451.WriteString(": " + param.Type.String())
		}
	}

	signature.WriteString(signatureSb451.String())

	signature.WriteString(")")

	if fn.ReturnType != nil {
		signature.WriteString(": " + fn.ReturnType.String())
	}

	return signature.String()
}

// buildFunctionSnippet builds an LSP snippet string for function insertion.
// Returns the snippet string and insertTextFormat.
// Example: "MyFunc(${1:param1}, ${2:param2})$0".
func buildFunctionSnippet(fn *ast.FunctionDecl) (string, protocol.InsertTextFormat) {
	if fn.Name == nil {
		return "", protocol.InsertTextFormatPlainText
	}

	// If function has no parameters, use plain text
	if len(fn.Parameters) == 0 {
		return fn.Name.Value + "()", protocol.InsertTextFormatPlainText
	}

	snippet := fn.Name.Value + "("

	var snippetSb496 strings.Builder

	for i, param := range fn.Parameters {
		if i > 0 {
			snippetSb496.WriteString(", ")
		}

		// Add tabstop with parameter name as placeholder
		tabstopNum := i + 1

		paramName := "param"
		if param.Name != nil {
			paramName = param.Name.Value
		}

		// Build tabstop: ${1:paramName}
		snippetSb496.WriteString("${" + strconv.Itoa(tabstopNum) + ":" + paramName + "}")
	}

	snippet += snippetSb496.String()

	snippet += ")$0" // $0 is the final cursor position

	return snippet, protocol.InsertTextFormatSnippet
}

// buildSnippetFromSignature builds a snippet from a signature string.
// Example: "Print(value: Variant)" -> "Print(${1:value})$0".
func buildSnippetFromSignature(functionName, signature string) (string, protocol.InsertTextFormat) {
	// Extract parameters from signature
	// Find the parameter list between parentheses
	startIdx := strings.Index(signature, "(")
	endIdx := strings.LastIndex(signature, ")")

	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		// No parameters or malformed signature
		return functionName + "()", protocol.InsertTextFormatPlainText
	}

	paramsStr := signature[startIdx+1 : endIdx]
	paramsStr = strings.TrimSpace(paramsStr)

	if paramsStr == "" {
		// No parameters
		return functionName + "()", protocol.InsertTextFormatPlainText
	}

	// Split parameters by comma
	params := strings.Split(paramsStr, ",")
	snippet := functionName + "("

	var snippetSb542 strings.Builder

	for i, param := range params {
		if i > 0 {
			snippetSb542.WriteString(", ")
		}

		param = strings.TrimSpace(param)

		// Extract parameter name (before colon if present)
		paramName := param
		if colonIdx := strings.Index(param, ":"); colonIdx != -1 {
			paramName = strings.TrimSpace(param[:colonIdx])
		}

		// Remove modifiers like "var", "const", "lazy"
		paramName = strings.TrimPrefix(paramName, "var ")
		paramName = strings.TrimPrefix(paramName, "const ")
		paramName = strings.TrimPrefix(paramName, "lazy ")
		paramName = strings.TrimSpace(paramName)

		if paramName == "" {
			paramName = "param" + strconv.Itoa(i+1)
		}

		// Build tabstop: ${1:paramName}
		tabstopNum := i + 1
		snippetSb542.WriteString("${" + strconv.Itoa(tabstopNum) + ":" + paramName + "}")
	}

	snippet += snippetSb542.String()

	snippet += ")$0"

	return snippet, protocol.InsertTextFormatSnippet
}

// getBuiltInCompletions returns completion items for built-in functions and types.
func getBuiltInCompletions() []protocol.CompletionItem {
	items := make([]protocol.CompletionItem, 0, 40)

	typeKind := protocol.CompletionItemKindClass
	plainTextFormat := protocol.InsertTextFormatPlainText

	for _, typeName := range builtInTypesList {
		detail := "Built-in type"
		sortText := "2builtin~" + typeName
		item := protocol.CompletionItem{
			Label:            typeName,
			Kind:             &typeKind,
			Detail:           &detail,
			SortText:         &sortText,
			InsertTextFormat: &plainTextFormat,
		}
		items = append(items, item)
	}

	funcKind := protocol.CompletionItemKindFunction

	for name, signature := range builtInFunctionsMap {
		detail := signature
		sortText := "2builtin~" + name

		// Create MarkupContent for better documentation
		doc := protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: "**Built-in function**\n\n```pascal\n" + signature + "\n```",
		}

		// Build snippet for function with parameters
		insertText, insertTextFormat := buildSnippetFromSignature(name, signature)

		item := protocol.CompletionItem{
			Label:            name,
			Kind:             &funcKind,
			Detail:           &detail,
			SortText:         &sortText,
			Documentation:    doc,
			InsertText:       &insertText,
			InsertTextFormat: &insertTextFormat,
		}
		items = append(items, item)
	}

	return items
}

// isPositionInNodeRange checks if a position is within a node's range.
func isPositionInNodeRange(node ast.Node, line, column int) bool {
	if node == nil {
		return false
	}

	start := node.Pos()
	end := node.End()

	if line < start.Line || line > end.Line {
		return false
	}

	if line == start.Line && column < start.Column {
		return false
	}

	if line == end.Line && column > end.Column {
		return false
	}

	return true
}

// FilterCompletionsByPrefix filters completion items by a prefix string.
// This is useful when the user has already typed part of an identifier.
func FilterCompletionsByPrefix(items []protocol.CompletionItem, prefix string) []protocol.CompletionItem {
	if prefix == "" {
		return items
	}

	var filtered []protocol.CompletionItem
	lowerPrefix := strings.ToLower(prefix)

	for _, item := range items {
		if strings.HasPrefix(strings.ToLower(item.Label), lowerPrefix) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}
