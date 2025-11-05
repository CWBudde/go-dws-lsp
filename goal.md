# Implementation Plan for go-dws Language Server

To build a Go-based Language Server for DWScript (go-dws), we will follow a phased approach similar to the Delphi reference project[\[1\]](https://github.com/CWBudde/DWScript-Language-Server#:~:text=,Rename%2C%20semantic%20tokens%2C%20code%20actions). Each phase introduces a set of features and improvements, with thorough testing at every step. The focus is on Go-idiomatic design (avoiding overly OOP patterns from Delphi) and leveraging available libraries for LSP scaffolding where it makes sense.

## Phase 0: Foundation - LSP Scaffolding and Setup

**Goal:** Establish the basic language server framework: communication, lifecycle handling, and project structure.

- **Project Setup:** Initialize a new Go module (e.g. go-dws-lsp) and set up the repository structure. Plan for a simple layout using Go packages rather than the complex class hierarchy used in Delphi[\[2\]](https://github.com/CWBudde/go-dws-lsp/blob/9677dbdc95d68ec5274a98c5733dcc5e613c3e8a/README.md#L90-L95). For example, create a main.go to launch the server and separate packages/modules for LSP handlers and DWScript integration.
- **Choose LSP Framework:** Use an LSP library for Go to handle protocol boilerplate. The **GLSP** (Go Language Server Protocol SDK) by tliron is a good choice - it provides all LSP message structures and a JSON-RPC server out of the box[\[3\]](https://tamerlan.dev/how-to-build-a-language-server-with-go/#:~:text=We%20also%20don%27t%20really%20have,an%20LSP%20SDK%20for%20Golang). Add GLSP as a dependency (go get github.com/tliron/glsp) to avoid writing the low-level JSON-RPC logic from scratch (we focus on language integration). This aligns with the plan to reuse scaffolding tools where practical.
- **Basic Server Initialization:** Implement the **Initialize** and **Shutdown** request handlers using the chosen framework. With GLSP, define a protocol.Handler with at least Initialize and Shutdown populated[\[4\]](https://tamerlan.dev/how-to-build-a-language-server-with-go/#:~:text=handler%20%3D%20protocol.Handler,shutdown%2C). In Initialize, construct and return the server's capabilities:
- Advertise at least basic capabilities now, such as text document sync and maybe placeholder support for hover, completion, etc. For example, set Capabilities.TextDocumentSync to incremental mode[\[5\]](https://github.com/CWBudde/go-dws-lsp/blob/9677dbdc95d68ec5274a98c5733dcc5e613c3e8a/README.md#L78-L86), and mark features like diagnostics, hover, completion as supported (you can refine these flags as features are implemented).
- Provide a ServerInfo name and version. Use go-dws LSP name and a version string.
- For **Shutdown**, simply handle the request (e.g., set a flag to exit or cleanup).
- **Transport Layer:** Implement the server launch to use STDIN/STDOUT by default (the standard for LSP). With GLSP, you can call server.RunStdio() to start listening[\[6\]](https://jitesh117.github.io/blog/creating-a-brainrot-language-server-in-golang/#:~:text=func%20main%28%29%20,RunStdio). Also consider adding an option to run over TCP for debugging (as the Delphi LSP did support a -tcp mode[\[7\]](https://github.com/CWBudde/DWScript-Language-Server#:~:text=Transport%20Options)). This means parsing command-line flags (e.g., using flag package) to optionally call RunTCP("localhost:8765") for debugging. Ensure the server can be started in either mode.
- **Logging & Debugging:** Introduce a basic logging facility. This could be as simple as using the standard log package or GLSP's logging integration (tliron/commonlog) for adjustable verbosity[\[6\]](https://jitesh117.github.io/blog/creating-a-brainrot-language-server-in-golang/#:~:text=func%20main%28%29%20,RunStdio). Include command-line flags or environment variables to control log level (mimicking Delphi LSP's -LSPTrace=verbose option[\[8\]](https://github.com/CWBudde/DWScript-Language-Server#:~:text=Command%20Line%20Options)). Ensure that incoming requests and errors are logged for troubleshooting.
- **Go Idioms & Improvements:** Design the internal architecture to use Go's strengths:
- Use lightweight structs and functions for handlers instead of heavy classes. For example, maintain a global or package-level state (with mutex protection) for open documents and their parse results.
- Leverage concurrency carefully - the LSP requests can be handled concurrently, so prepare to guard shared data (like the documents map) with a sync.RWMutex or similar. However, GLSP's server will sequentially handle notifications by default; still, design with thread safety in mind as a best practice.
- Error handling: propagate errors instead of silent failures - e.g., if initialization fails (missing go-dws library, etc.), log and exit gracefully. Use Go's explicit error checks (no exceptions) for clarity.
- **Verify Basic Lifecycle:** Manually test that the server can start and respond to a minimal LSP client:
- Write a small Go test or script that sends an **initialize** request (with a JSON message or using GLSP's client capabilities if available) and check that an **InitializeResult** with expected capabilities is returned.
- Also test that sending an **initialized** notification (if client sends it) and a **shutdown** request triggers no errors and the process can exit.
- **Testing Strategy:** Set up a Go test suite. Even at this foundation phase, include a simple test case for initialization. This can use a fake connection or GLSP's ability to run in-process. The focus is to ensure the framework is working (e.g., a test that calls our Initialize handler function directly and asserts the returned capabilities match what we expect).

_Outcome:_ By end of Phase 0, we have a running LSP server skeleton that communicates over STDIO/TCP, correctly handles the initialize/shutdown lifecycle, and logs its activity[\[1\]](https://github.com/CWBudde/DWScript-Language-Server#:~:text=,Rename%2C%20semantic%20tokens%2C%20code%20actions). No language-specific features are implemented yet, but the foundation is ready for document management and DWScript integration.

## Phase 1: Document Synchronization and Diagnostics

**Goal:** Implement basic text document management and real-time error reporting. This corresponds to _Document Sync_ and initial workspace indexing as noted in the Delphi LSP plan[\[1\]](https://github.com/CWBudde/DWScript-Language-Server#:~:text=,Rename%2C%20semantic%20tokens%2C%20code%20actions).

- **Open/Close Document Handling:** Implement handlers for **textDocument/didOpen** and **didClose**:
- When a file is opened, store its content in an in-memory structure. Maintain a map from document URI to a struct containing at least the text (and later, parse results or AST). On open, you can also record the initial version if needed.
- On close, remove the document from the map to free memory. Also consider clearing diagnostics: send a final empty diagnostics array for that file to signal clearing of errors (some editors do this automatically, but sending an empty list is safe).
- **Change Notifications:** Implement **textDocument/didChange** to handle edits. Mark the server's sync capability as _incremental_ in Initialize so the client sends diffs[\[5\]](https://github.com/CWBudde/go-dws-lsp/blob/9677dbdc95d68ec5274a98c5733dcc5e613c3e8a/README.md#L78-L86). In the handler:
- Locate the document entry in the map and apply the changes. Each DidChangeTextDocumentParams contains one or more TextDocumentContentChangeEvent with either a full new text or a diff. For simplicity, you might start by using **full sync** (send whole text on each change) - this is easier to implement. If full sync, just replace the stored text with the new text. If incremental (preferred), implement a utility to apply the diff: the change event gives a range and new text[\[9\]](https://github.com/CWBudde/go-dws-lsp/blob/9677dbdc95d68ec5274a98c5733dcc5e613c3e8a/README.md#L78-L84). Update the in-memory text accordingly.
- Ensure thread safety if applying changes concurrently (use a mutex around the doc update).
- **Real-time Diagnostics:** After opening or changing a document, trigger compilation to provide diagnostics (errors/warnings):
- **Parsing:** Run the go-dws parser on the document's text. Use the lexer.New() and parser.New() as in the go-dws CLI[\[10\]](https://github.com/CWBudde/go-dws/blob/79ca1b3c112de54ca491e1ae2bad325a6a345576/cmd/dwscript/cmd/parse.go#L63-L71). Parse into an AST (e.g., program := p.ParseProgram()).
- Collect any syntax errors: if p.Errors() is non-empty, convert these into LSP Diagnostic objects. The go-dws parser error strings can be turned into diagnostics by extracting line/column info. The go-dws errors package provides helpers to format errors with source context[\[11\]](https://github.com/CWBudde/go-dws/blob/79ca1b3c112de54ca491e1ae2bad325a6a345576/cmd/dwscript/cmd/run.go#L90-L96)[\[12\]](https://github.com/CWBudde/go-dws/blob/79ca1b3c112de54ca491e1ae2bad325a6a345576/cmd/dwscript/cmd/run.go#L103-L111). For example, errors.FromStringErrors(p.Errors(), source, filename) yields structured errors with positions, which can then be formatted[\[13\]](https://github.com/CWBudde/go-dws/blob/79ca1b3c112de54ca491e1ae2bad325a6a345576/cmd/dwscript/cmd/run.go#L91-L99). Use those positions to fill Diagnostic.Range (start and end positions of the offending token or line). Mark severity as Error for parse errors.
- **Semantic Analysis:** If parsing succeeds, perform semantic checks (type checking, undefined identifiers, etc.). Leverage go-dws's semantic analyzer: create an analyzer via semantic.NewAnalyzer() and call analyzer.Analyze(program)[\[12\]](https://github.com/CWBudde/go-dws/blob/79ca1b3c112de54ca491e1ae2bad325a6a345576/cmd/dwscript/cmd/run.go#L103-L111). If it returns an error or if analyzer.Errors() list is non-empty, retrieve those errors. Go-dws provides analyzer.StructuredErrors() or raw error strings. Convert these to Diagnostics as well (they likely include line/col; the Delphi LSP uses structured compiler errors for rich messages[\[14\]](https://github.com/CWBudde/go-dws/blob/79ca1b3c112de54ca491e1ae2bad325a6a345576/cmd/dwscript/cmd/run.go#L115-L124)). This will catch things like type mismatches, unknown variables, etc., providing deeper diagnostics beyond parsing.
- **Publish Diagnostics:** Send the diagnostics to the client via the textDocument/publishDiagnostics notification. In GLSP, you can send notifications by calling the server's notify method (e.g., glsp.Notify(context, protocol.ServerNotifTextDocumentPublishDiagnostics, params)). Construct PublishDiagnosticsParams with the document URI and the list of Diagnostic objects. This way, as soon as a file is opened or edited, the client will underline errors in the code.
- Include **warnings** if any (the language might have warnings for unused variables or other conditions, if go-dws flags them).
- **Workspace Management:** Handle basic **workspace** events:
- If the client sends initialized (after initialize), you might scan the workspace for .dws files (**workspace indexing**). For now, implementing a full project index is optional, but we plan for it. You could iterate through the workspace folder(s) (from InitializeParams) and parse all DWScript files to build a symbol index. This index (e.g., a map of symbol names to definitions with file locations) will be useful for workspace-wide features like references and workspaceSymbol.
- At a minimum, set up data structures to hold multiple files' ASTs and symbols. For Phase 1, you might just index on-demand (e.g., only parse files when opened). In a later phase we will flesh out the **workspaceSymbol** request using this index.
- Also handle workspace/didChangeConfiguration if needed - e.g., accept any config settings (for now, there may be none, but structure is in place for future use such as toggling certain features).
- **Testing & Validation:** Adopt a test-driven approach:
- Write unit tests for the diagnostic generation: given a snippet of DWScript code, feed it into the parser/analyzer functions and assert that the returned Diagnostics match expected errors. For example, test that an input with a syntax error produces a Diagnostic with the correct message and position.
- Use **go-dws's test scripts** as a validation set. The go-dws repo has many testdata DWScript programs. You can write a test that iterates over a set of valid scripts to ensure our LSP reports zero errors for them (no false diagnostics), and also test known erroneous code to ensure the errors are caught. This leverages the existing DWScript test corpus to increase confidence in the LSP.
- Test the incremental update: simulate opening a document with valid code (expect no errors), then simulate a small edit that introduces an error (expect a diagnostic), and then another edit fixing it (diagnostic clears). Verify the publishDiagnostics outputs at each step.
- **Performance considerations:** In this phase, re-parsing the whole document on each change is acceptable (DWScript files are typically not huge, and go-dws parsing is efficient). However, note this in code comments as a potential future improvement (incremental parsing). Ensure that rapid sequence of changes doesn't flood the client with outdated diagnostics - consider debouncing rapid didChange events (e.g., a short delay before recompute) or canceling previous analysis if a new edit comes in. This can be refined later if needed.

_Outcome:_ By the end of Phase 1, the LSP server fully supports document synchronization and provides real-time **syntax and semantic diagnostics** for DWScript code. You can open a DWScript file in VSCode (or any LSP client) and see compiler errors/warnings underline as you type[\[9\]](https://github.com/CWBudde/go-dws-lsp/blob/9677dbdc95d68ec5274a98c5733dcc5e613c3e8a/README.md#L78-L84). The groundwork for managing multiple files and a workspace is laid, ready for more advanced language features.

## Phase 2: Core Language Features (Hover, Symbols, Completion, etc.)

**Goal:** Implement the main IDE features for DWScript - hover tooltips, go-to-definition, find references, code completion, signature help, and document symbols. This corresponds to the "Core Features" phase[\[15\]](https://github.com/CWBudde/DWScript-Language-Server#:~:text=,Rename%2C%20semantic%20tokens%2C%20code%20actions), bringing the language server to a usable state for code understanding and navigation.

- **Hover (textDocument/hover):** Provide type and symbol information on mouse hover.
- Determine the symbol at the hovered position. When a hover request comes with a TextDocumentPosition, retrieve the corresponding document's AST (stored from last parse).
- Identify the AST node under the position. This may involve enhancing the AST nodes with position metadata (if not already present). The go-dws parser likely records token positions; ensure these are accessible (you might modify the parser to store Line and Col on AST nodes or keep a mapping from positions to AST nodes).
- If the node is an identifier (variable, constant, function, type, etc.), gather its info:
  - For a variable: find its declaration (e.g., search the AST in the current scope for a VarDecl with the same name). Retrieve its type (the semantic analyzer has resolved types; possibly the AST or analyzer can give the variable's Type).
  - For a function: find the function definition AST (by name) and extract its signature (parameter types and return type).
  - For a class/type: get its definition, list maybe its ancestor or a short description (class name, properties count, etc.).
  - Include any documentation comments if present (future enhancement: parse DWScript doc comments).
- Construct a Hover response with a MarkupContent string summarizing the symbol. For example: **"Integer variable** x **\= 42"** or **"function** Foo(i: Integer): String\`\*\* - basically type info or signature. Format it in markdown or plaintext as needed for the client.
- If no meaningful symbol is found (e.g., hovering on whitespace or an operator), return hover with no contents (or nothing).
- **Test:** Create small code examples and verify hover returns the expected info. For instance, in code var x: Integer := 5; hovering on x should show "x: Integer". Hover on a function name should show its declaration line. Use unit tests by directly calling the hover handler with a prepared AST and position.
- **Go-to Definition (textDocument/definition):** Jump to symbol definitions.
- Given a position (usually on an identifier/reference in code), resolve which symbol it refers to. This is similar to hover: find the identifier at the position and determine its target definition.
- If the identifier is a local variable or parameter, its definition is likely in the same file (earlier in the AST). Find that VarDecl or parameter in the AST and record its location (uri and range). If it's a field or method of a class, the definition might be within the class definition (search the AST for class by name, then the member).
- If the symbol is global (like a function/procedure or a global variable, or a type), it might be in another file (if using units) or in the same file. First check the current file's AST for a matching declaration. If not found and workspace indexing is available, look up the symbol in the workspace symbol table (populated in Phase 1 or via on-demand parsing of other files). For example, if uses SomeUnit; and the identifier is defined in SomeUnit, you might parse SomeUnit.dws (if not already) to find it.
- Return the location(s) of the definition. In LSP, you can return a single Location or an array. Typically it's a single location for a unique definition. If ambiguous (e.g., function overloaded or multiple forward declarations), return an array.
- **Test:** Define symbols and references in a small multi-function file and ensure the definition handler finds the correct line. Also test cross-file: e.g., simulate a tiny "unit" file with a definition and a main file that uses it - ensure the handler finds the definition in the unit file.
- **Find References (textDocument/references):** Find all usages of a symbol.
- This feature complements go-to-definition. When a references request comes in, identify the symbol at the given position (again, find the identifier and its defining symbol). Then gather all places where this symbol is used.
- If we maintain a global index of symbol definitions and references, we can utilize it here. For a simple approach, do the following:
  - If the symbol is local (scope-limited), just search within the same AST for all occurrences of that name in relevant contexts (e.g., same function or block).
  - If the symbol is global or exported (like a unit member or global function), search all open documents' ASTs for that name. If we have parsed the whole workspace, search those ASTs too. You may create a helper that scans AST nodes and collects identifiers matching a target symbol name **and** kind (to avoid false matches where the same name is used for something else). The semantic analyzer can help here: if we had symbol IDs or memory references, that would be ideal, but since we might not, filter by scope: e.g., for a global function Foo, find all Identifier nodes Foo that are in call expressions or references, not where they are defined in another context.
  - Leverage the analyzer if possible: perhaps it can give a symbol table or we can compare resolved type info. For instance, for methods, fully qualified names might help (ClassName.MethodName).
- Create a list of Locations for each found reference (each usage position). Include the definition itself if the LSP convention is to include it (usually references exclude the definition by default; you can decide).
- **Test:** Similar to definition, set up code where a symbol is used multiple times and verify the handler returns all correct positions. Also ensure no spurious references (e.g., another variable with same name in a different function should not be included - test scope isolation).
- **Document Symbols (textDocument/documentSymbol):** Provide outline of the document's structure.
- Traverse the AST of a file to collect all top-level symbols and inner definitions for outline view. Use the AST from go-dws:
  - Include **functions/procedures** (with kind = Function/Method in LSP).
  - Include **global variables/constants** (kind = Variable or Constant).
  - Include **types** (classes, interfaces, enums if any) - kind = Class/Interface/Struct etc. For classes, consider adding child DocumentSymbol entries for their fields and methods for a hierarchical outline.
  - If the language supports nested functions or inner classes, handle those hierarchically as well.
- Map DWScript constructs to LSP SymbolKind (e.g., class -> Class, interface -> Interface, property -> Property, method -> Method, etc.). Use best-fitting kinds for things like global constants (Constant) and so on[\[5\]](https://github.com/CWBudde/go-dws-lsp/blob/9677dbdc95d68ec5274a98c5733dcc5e613c3e8a/README.md#L78-L86).
- Return the symbols either as hierarchical DocumentSymbol objects (preferred, since we can represent nesting), or as flat SymbolInformation array. Modern LSP prefers the DocumentSymbol approach when possible.
- **Test:** Write a DWScript snippet with a couple of top-level functions and a class with members. Call the documentSymbol handler and ensure it returns a structured list (the class contains its members as children, etc.) with correct names and kinds.
- **Workspace Symbols (workspace/symbol):** (Optional in this phase, or Phase 3) Provide a global search for symbols by name.
- Leverage the workspace index if built. On a query (which includes a query string), search through all known symbol definitions (across files) for matches. For simplicity, match substrings or prefix of symbol names.
- Return a list of SymbolInformation (each with name, kind, location, and containerName if applicable). This allows users to search symbols across the project.
- If full indexing was not done, this request could trigger a brute-force search: parse all workspace files not already open to find symbols. This is expensive, so better to have indexed in advance or limit to open files for now.
- Mark workspaceSymbolProvider: true in capabilities if implemented.
- **Test:** If indexing is implemented, add a test where multiple files define symbols and a search term returns the correct subset.
- **Code Completion (textDocument/completion):** Provide completion suggestions as the user types.
- Determine the context for completion:
  - If the completion is triggered by a trigger character or manually, get the text up to the cursor. Identify if we are in a context of a member access (like object.) or a normal scope.
  - The AST may not parse successfully mid-statement (because the code is incomplete where completion is invoked). To handle this, you might do a partial parse or use a simpler heuristic: for instance, use the lexical information. Alternatively, consider using the last known good AST and the current text to infer context.
  - **Basic approach:** If the triggering character is a dot (.), attempt member completion. If so, find the identifier before the dot to determine its type (for example, if user typed obj. - find the type of obj via the symbol table or analyzer results, then list members of that type). If type info is available from semantic analysis (e.g., analyzer could tell us that obj is of class TMyClass), then look up that class in AST or a class registry to get its fields/methods.
  - If no dot, provide general completions in the current scope:
  - **Keywords:** You can include language keywords (if desired) such as begin/end, if, for, etc., though editors often handle basic keywords via syntax highlighting and snippets. It's still nice to have them in suggestions.
  - **Variables in scope:** List local variables and parameters currently in scope. This requires knowing the scope of the cursor - you can determine which function or block the cursor is in by AST node positions.
  - **Globals:** Include global functions, types, and constants (from the current file and possibly imported units).
  - **Built-in functions and types:** DWScript has a set of built-in functions (e.g., PrintLn, IntToStr, etc.) and types (Integer, String, etc.). If not already in the AST, maintain a list of built-ins (go-dws might expose these, e.g., in analyze_builtins.go). Offer those as well.
  - Construct a list of CompletionItem:
  - For each suggestion, set the label (e.g., FooBar), kind (function, variable, class, keyword, etc.), and any detail (like type info for variables or signature for functions).
  - For functions, you can include snippet-style insert text (e.g., Foo(\${1:arg})) to assist with parameters, and set insertTextFormat=Snippet.
  - If additional resolution is needed (like computing documentation on item selection), you can set CompletionItem.resolveProvider = true in capabilities and implement completionItem/resolve. For now, it might be enough to provide all info in the initial response.
- **Performance:** Ensure the completion generation is fast. If needed, optimize by caching the list of global suggestions so you don't recompute them on every keystroke. You can cache after parsing: e.g., store all symbols in the AST in a list for quick filtering by prefix.
- **Test:** Simulate a completion scenario:
  - E.g., given a context where local variables alpha and beta exist, test that typing a returns alpha in the completion list.
  - Test that after a dot on an object of known class, the members of that class appear.
  - These can be unit tests where you set up a known AST (or use go-dws to parse a snippet), then call the completion handler with various cursor positions and assert the results contain expected suggestions.
- **Signature Help (textDocument/signatureHelp):** Show function parameters while calling a function.
- When the user is inside a function call and triggers signature help (often on typing ( or ,), the server should provide the function's signature and highlight the current parameter.
- Determine the call context: using the AST or a quick parse of the current line:
  - If the code is Foo(x, y| ) (cursor at |), the AST might have partially parsed Foo(...) as a call node with some arguments. If the AST is incomplete (due to missing closing parenthesis), we might need a graceful strategy:
  - Possibly temporarily insert a ) in the text and parse to get a complete AST, or traverse tokens backward to find the function identifier and count how many commas are before the cursor.
  - Identify the function being called: e.g., find the nearest ( to the left of the cursor and the identifier immediately before it as the function name. Confirm it's a call by AST or by pattern matching.
  - Find that function's definition (similar to go-to-def) to retrieve its parameters and documentation. For built-in procedures (like PrintLn), you might have a predefined signature available (e.g., PrintLn(value: Any) : Void).
- Construct a SignatureHelp response:
  - SignatureInformation for the function, including a formatted label (e.g., Foo(a: Integer, b: String): Boolean) and possibly a doc string if available.
  - Within that, provide ParameterInformation array for each param (you can include parameter name and type as the label).
  - Determine activeParameter index by counting the commas before the cursor (e.g., 0 for first param, 1 for second, etc.).
  - Set activeSignature (usually 0 unless you support function overloading with multiple signatures).
- **Test:** Write a snippet defining a function with multiple parameters, simulate a call in progress, and verify SignatureHelp returns the correct parameters and highlights the right one. For instance, MyFunc(a, b) with cursor after the comma should highlight parameter 2.
- **Improved Error Handling & Resilience:** As we add these features, ensure robust handling:
- If any handler encounters a situation it doesn't recognize (e.g., symbol not found), handle gracefully by returning an empty result or an error response (with appropriate LSP error code) rather than crashing.
- Wrap critical sections in recover blocks if panics could occur (especially if you integrate deeply with go-dws internals).
- Continue logging extensively in debug mode (e.g., log when a hover or completion request comes in, and any unexpected conditions).
- If using GLSP, it might already catch panics and return error responses - verify this behavior.
- **Testing Integration:** At this phase, consider writing integration tests or using an actual LSP client in a controlled environment:
- You could use VSCode with your server or a tool like lsp-test (if available for Go) to script a sequence: open file, request hover, etc., and verify responses. However, unit tests for each handler and manual testing in an editor might suffice.
- Test in VSCode (or another editor) manually to see the features in action: open a project with multiple DWScript files and try hovering, navigating, autocompletion. This manual run will often reveal any synchronization issues or crashes to fix.

_Outcome:_ After Phase 2, the language server supports all essential IDE features listed in the original plan - **hover, go-to definition, find references, document symbols, code completion, signature help**, etc.[\[16\]](https://github.com/CWBudde/DWScript-Language-Server#:~:text=Working%20Features%20%E2%9C%85)[\[5\]](https://github.com/CWBudde/go-dws-lsp/blob/9677dbdc95d68ec5274a98c5733dcc5e613c3e8a/README.md#L78-L86). Developers using DWScript in VSCode (or other LSP editors) can navigate and edit with a rich experience: e.g., hovering symbols shows type info, they can jump to definitions across files, get autocompletion suggestions, and see an outline of the file structure. All these features should be functioning correctly with proper synchronization with the document content.

## Phase 3: Advanced Features and Enhancements

**Goal:** Implement more advanced LSP features and refine the server with best practices. This includes refactoring support, semantic tokens for syntax highlighting, and code actions, as well as any remaining nice-to-haves. According to the roadmap, this is the phase for Rename, semantic highlighting, code actions, etc[\[17\]](https://github.com/CWBudde/DWScript-Language-Server#:~:text=indexing%20,Rename%2C%20semantic%20tokens%2C%20code%20actions).

- **Rename (textDocument/rename):** Enable symbol renaming across the code.
- When a rename request comes in, determine the symbol at the given position (same logic as definition/references). Validate that it's a symbol that can be renamed (if user tries to rename a keyword or a built-in, respond with an error or a special message).
- Find all references of that symbol (you likely have this implemented in Phase 2). This provides all locations that need to change.
- Prepare a WorkspaceEdit response:
  - For each reference location (and the definition), compute the new text (the new name provided by the request). Create a TextEdit for the range of the old name to be replaced with the new name.
  - Group TextEdits by file (each file in WorkspaceEdit.DocumentChanges or changes map).
- Make sure to handle race conditions: if the document is open and the user is actively editing, the rename might come with an old document version. Typically the LSP client includes a document version in the params; you should check if it matches the current known version to avoid applying stale renames.
- **Prepare Rename:** (Optional) Implement textDocument/prepareRename to give the client feedback on what range will be renamed. This involves identifying the symbol and returning its range and placeholder text. This can help prevent renaming if on an invalid token. Not strictly required, but a nice addition.
- **Test:** Write tests where a variable or function is renamed and ensure all occurrences in one or multiple files are updated in the edit. Also test a scenario where rename is not allowed (renaming a keyword like begin should be rejected).
- **Semantic Tokens (textDocument/semanticTokens):** Provide syntax highlighting information from the server. This can complement or replace TextMate grammars and is useful for semantic distinctions (e.g., highlighting variable vs. property differently).
- **Legend:** Define a SemanticTokensLegend and advertise it in server capabilities (SemanticTokensProvider). Decide on token types relevant for DWScript. For example: "keyword", "string", "number", "comment", "variable", "parameter", "property", "function", "class", "interface", "enum", etc., plus modifiers like "static", "deprecated".
- **Token generation:** Traverse the AST of a document to collect tokens:
  - Keywords (if desired): Since AST might not explicitly list keywords, you might rely on either the lexer or a list of reserved words. However, highlighting keywords can usually be left to a TextMate grammar. Semantic tokens are most valuable for identifiers and entities that grammars can't easily distinguish.
  - Identifiers: differentiate based on their role:
  - If an AST node is a variable declaration, add a token for the identifier with type _variable_ and a modifier _declaration_.
  - If it's a constant or enum member, token type _enumMember_ or _variable_ (depending on how you classify).
  - Function and method names: _function_ (or _method_ if you want separate type for methods).
  - Class names and type identifiers: _class_ or _type_.
  - Property names (fields in classes): _property_.
  - Interface names: _interface_.
  - Parameters: _parameter_.
  - You can also tag _this/Self_ references specially if needed.
  - Literals: number literals, string literals, boolean literals, etc. For instance, number -> _number_, string -> _string_, boolean -> _keyword_ or a separate _boolean_ type.
  - Comments: could be tagged as _comment_, though the client might color them anyway.
  - Note: Since the editor likely still does basic syntax highlighting with TextMate, our semantic tokens can either override or complement. It might be wise to focus on semantic info (identifiers) and not duplicate basic coloring of keywords and punctuation.
- **Range computation:** Each token needs a line & start character and length, plus a token type index (into the legend) and modifiers bitset. Use the positions from AST nodes:
  - Ensure all AST nodes have start/end position info. You may need to extend the parser to record end positions of constructs if needed. Alternatively, token length can be derived from the identifier's name length for identifiers.
  - For each identified token, record its \[line, startChar, length, tokenType, tokenModifiers\].
  - Sort tokens by position (the LSP requires they be delivered in increasing order).
  - Encode in the LSP relative format (delta line, delta start, etc.). If using an LSP library, it might help format the SemanticTokens data, otherwise implement the delta encoding.
- Provide implementations for textDocument/semanticTokens/full (and delta if you want incremental updates to highlighting - can be skipped initially). Every time a document changes or is opened, the client may request updated semantic tokens. Ensure to recompute after each change or parse.
- **Test:** Create a sample DWScript code string containing various constructs (variables, functions, classes) and run the semantic token generation. Verify that the output list correctly classifies each piece (e.g., variables are identified as type X, keywords as Y, etc.). This can be done with a unit test that calls our token generation function and checks a few expected entries.
- In an editor, verify that enabling semantic highlighting (and providing the legend in the client settings) results in the expected colors. You may need to supply the VSCode extension with the legend so it knows how to map token types to colors.
- **Code Actions (textDocument/codeAction):** Offer quick fixes or refactorings.
- Start by supporting **quick fixes** for diagnostics. For example:
  - If there's a compiler error "Undeclared identifier X", a possible quick fix is to declare that identifier. We can suggest a code action: _"Declare variable 'X'"_. The action would insert a line var X: &lt;Type&gt;; either at the top of the function or in the global section, depending on context. (Choosing a default type might be tricky - perhaps Integer or Variant as a fallback, or allow the user to edit after insertion).
  - If there's a missing semicolon error at a line, offer _"Insert missing ';'"_ as a fix. The action would simply insert a semicolon at the appropriate position.
  - If a variable is assigned but never used (if we had such a warning), offer to remove it or prefix with _ to mark ignored (just examples).
  - If there's a type mismatch error, possibly suggest a cast if appropriate (this is more complex and might not always be clear).
- Also consider **refactorings**:
  - _Extract to function/procedure:_ This is complex (involves selecting code and creating a new function), possibly skip at this stage unless time permits.
  - _Organize uses_: If DWScript had an import/unit mechanism, a code action to remove unused unit references or add missing ones could be useful. E.g., if a symbol from System unit is used but System not in uses, suggest adding System to the uses list.
  - _Implement interface/abstract methods:_ If a class doesn't implement some interface methods, a code action could stub them out. This requires deeper semantic knowledge (likely skip until go-dws can provide such info).
- To implement code actions:
  - For each diagnostic in the CodeActionContext, check if we recognize its message pattern. Use either error codes (if go-dws provides specific codes) or substring matching on the message (e.g., contains "Unknown identifier").
  - Create a CodeAction with kind = quickfix or appropriate kind, attach the Diagnostic as associatedDiagnostic and provide an edit that resolves the issue.
  - The edit is a WorkspaceEdit similar to rename, possibly with changes in one file (like insertion or modification).
  - Ensure the code action's title clearly states the fix.
- Mark codeActionProvider: true in capabilities (and specify codeActionKinds if limiting to quickfix).
- **Test:** Simulate a scenario: e.g., feed a file content with an undeclared variable error, call codeAction handler with the diagnostic, and verify it suggests the correct fix with an edit. Check that applying the edit would indeed fix the error (you can even apply it to the text and re-run the parser in the test to see if the error is gone).
- **Additional Testing & Quality:** Now that most features are implemented, do a thorough integration test pass:
- Run the LSP against real DWScript projects or a sample project. Use VSCode to try out all features together. Verify that no feature's implementation breaks another (for example, ensure that our document synchronization logic is robust when doing go-to-def or rename during typing).
- Pay attention to performance: the IDE should not freeze during large operations. If find-references on a big project is slow, consider optimizing by doing the search asynchronously or using the progress reporting (LSP allows sending partial results or a progress notification). This might be advanced, but mention it for future improvements.
- Memory usage: ensure that closing documents frees their stored data. If the workspace is large and we parsed many files, consider if we should release ASTs for files that are not needed or implement an LRU cache.
- **Documentation & Examples:** Update the README (or separate docs) with usage info - though the user said they will handle usage in a separate repo (for the VSCode extension). Still, documenting any important details for contributors is good (for instance, how to run tests, how the architecture is split, etc.). This is not an end-user focus, but a developer focus.
- **Go Best Practices Audit:** Refactor any parts that are not idiomatic:
- Ensure proper package naming and division. Perhaps have internal/lsp for handlers, internal/dwscript for any DWScript-specific utilities (though ideally rely on go-dws for actual parsing and analysis).
- Remove any leftover global state if not necessary; maybe use a struct to encapsulate server state (open docs, caches) and methods that operate on it, passed to handlers via closure or context. GLSP allows capturing context; we might store a pointer to our state struct in the GLSP context's user data.
- Concurrency: double-check that all shared data (documents, symbol index) are accessed safely. Possibly use sync.Map or RWMutex.
- Testing: Ensure test coverage is high for all new features. Particularly, complex logic like completion or rename should have multiple scenario tests.

_Outcome:_ Phase 3 completion means the go-dws LSP is feature-complete and robust. Advanced functionality like **Rename** and **Semantic Tokens** are implemented, enhancing refactoring and editor colorization. Code Actions provide helpful quick fixes to improve developer productivity. The language server now covers all capabilities advertised, matching or exceeding the Delphi implementation (and following best practices where the Delphi version had limitations). At this stage, one can integrate the server with a VSCode extension or other editors confidently, knowing that it provides a comprehensive LSP experience for DWScript.

By following these phases, we progressively build a powerful DWScript language server. Each phase's deliverables (foundation, document sync, core features, advanced features) align with the original plan[\[1\]](https://github.com/CWBudde/DWScript-Language-Server#:~:text=,Rename%2C%20semantic%20tokens%2C%20code%20actions) and ensure that at each step the server is testable and stable. This phased, test-driven approach will result in a Go-idiomatic, reliable LSP implementation for go-dws, ready to support rich DWScript development in any LSP-compatible editor.

**References:** The plan has been informed by the existing DWScript LSP (Delphi) feature set[\[5\]](https://github.com/CWBudde/go-dws-lsp/blob/9677dbdc95d68ec5274a98c5733dcc5e613c3e8a/README.md#L78-L86) and architecture, while leveraging Go libraries like GLSP for efficient protocol handling[\[3\]](https://tamerlan.dev/how-to-build-a-language-server-with-go/#:~:text=We%20also%20don%27t%20really%20have,an%20LSP%20SDK%20for%20Golang). We use go-dws's own compiler components for parsing and analysis[\[10\]](https://github.com/CWBudde/go-dws/blob/79ca1b3c112de54ca491e1ae2bad325a6a345576/cmd/dwscript/cmd/parse.go#L63-L71)[\[12\]](https://github.com/CWBudde/go-dws/blob/79ca1b3c112de54ca491e1ae2bad325a6a345576/cmd/dwscript/cmd/run.go#L103-L111) to ensure accuracy and consistency with DWScript's language rules. Each feature maps to standard LSP methods as listed in the DWScript LSP docs (e.g., hover, completion, diagnostics, etc.)[\[5\]](https://github.com/CWBudde/go-dws-lsp/blob/9677dbdc95d68ec5274a98c5733dcc5e613c3e8a/README.md#L78-L86). The result will be a language server that is both familiar in capability (to DWScript users) and improved in implementation through Go's simplicity and power.

[\[1\]](https://github.com/CWBudde/DWScript-Language-Server#:~:text=,Rename%2C%20semantic%20tokens%2C%20code%20actions) [\[7\]](https://github.com/CWBudde/DWScript-Language-Server#:~:text=Transport%20Options) [\[8\]](https://github.com/CWBudde/DWScript-Language-Server#:~:text=Command%20Line%20Options) [\[15\]](https://github.com/CWBudde/DWScript-Language-Server#:~:text=,Rename%2C%20semantic%20tokens%2C%20code%20actions) [\[16\]](https://github.com/CWBudde/DWScript-Language-Server#:~:text=Working%20Features%20%E2%9C%85) [\[17\]](https://github.com/CWBudde/DWScript-Language-Server#:~:text=indexing%20,Rename%2C%20semantic%20tokens%2C%20code%20actions) GitHub - CWBudde/DWScript-Language-Server: An open source implementation of a language server for DWScript

<https://github.com/CWBudde/DWScript-Language-Server>

[\[2\]](https://github.com/CWBudde/go-dws-lsp/blob/9677dbdc95d68ec5274a98c5733dcc5e613c3e8a/README.md#L90-L95) [\[5\]](https://github.com/CWBudde/go-dws-lsp/blob/9677dbdc95d68ec5274a98c5733dcc5e613c3e8a/README.md#L78-L86) [\[9\]](https://github.com/CWBudde/go-dws-lsp/blob/9677dbdc95d68ec5274a98c5733dcc5e613c3e8a/README.md#L78-L84) README.md

<https://github.com/CWBudde/go-dws-lsp/blob/9677dbdc95d68ec5274a98c5733dcc5e613c3e8a/README.md>

[\[3\]](https://tamerlan.dev/how-to-build-a-language-server-with-go/#:~:text=We%20also%20don%27t%20really%20have,an%20LSP%20SDK%20for%20Golang) [\[4\]](https://tamerlan.dev/how-to-build-a-language-server-with-go/#:~:text=handler%20%3D%20protocol.Handler,shutdown%2C) How to Build a Language Server with Go

<https://tamerlan.dev/how-to-build-a-language-server-with-go/>

[\[6\]](https://jitesh117.github.io/blog/creating-a-brainrot-language-server-in-golang/#:~:text=func%20main%28%29%20,RunStdio) Creating a Brainrot Language Server in Golang | Jitesh's Blog

<https://jitesh117.github.io/blog/creating-a-brainrot-language-server-in-golang/>

[\[10\]](https://github.com/CWBudde/go-dws/blob/79ca1b3c112de54ca491e1ae2bad325a6a345576/cmd/dwscript/cmd/parse.go#L63-L71) parse.go

<https://github.com/CWBudde/go-dws/blob/79ca1b3c112de54ca491e1ae2bad325a6a345576/cmd/dwscript/cmd/parse.go>

[\[11\]](https://github.com/CWBudde/go-dws/blob/79ca1b3c112de54ca491e1ae2bad325a6a345576/cmd/dwscript/cmd/run.go#L90-L96) [\[12\]](https://github.com/CWBudde/go-dws/blob/79ca1b3c112de54ca491e1ae2bad325a6a345576/cmd/dwscript/cmd/run.go#L103-L111) [\[13\]](https://github.com/CWBudde/go-dws/blob/79ca1b3c112de54ca491e1ae2bad325a6a345576/cmd/dwscript/cmd/run.go#L91-L99) [\[14\]](https://github.com/CWBudde/go-dws/blob/79ca1b3c112de54ca491e1ae2bad325a6a345576/cmd/dwscript/cmd/run.go#L115-L124) run.go

<https://github.com/CWBudde/go-dws/blob/79ca1b3c112de54ca491e1ae2bad325a6a345576/cmd/dwscript/cmd/run.go>