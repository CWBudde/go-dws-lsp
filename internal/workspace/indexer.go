// Package workspace provides workspace-wide symbol indexing and management.
package workspace

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/CWBudde/go-dws-lsp/internal/util"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/dwscript"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// Indexer handles workspace file indexing.
type Indexer struct {
	index     *SymbolIndex
	maxDepth  int
	maxFiles  int
	fileCount int
}

// NewIndexer creates a new workspace indexer.
func NewIndexer(index *SymbolIndex) *Indexer {
	return &Indexer{
		index:    index,
		maxDepth: 10,    // Maximum directory depth
		maxFiles: 10000, // Maximum files to index
	}
}



// BuildWorkspaceIndex scans workspace folders and indexes all .dws files.
// This runs in the background and doesn't block the caller.
func (idx *Indexer) BuildWorkspaceIndex(workspaceFolders []protocol.WorkspaceFolder) {
	if len(workspaceFolders) == 0 {
		log.Println("No workspace folders to index")
		return
	}

	log.Printf("Starting workspace indexing for %d folders\n", len(workspaceFolders))

	for _, folder := range workspaceFolders {
		// Convert URI to file path
		path := uriToPath(folder.URI)
		if path == "" {
			log.Printf("Warning: Could not convert URI to path: %s\n", folder.URI)
			continue
		}

		log.Printf("Indexing workspace folder: %s\n", path)
		idx.indexDirectory(path, 0)
	}

	log.Printf("Workspace indexing complete. Indexed %d files, %d symbols\n",
		idx.fileCount, idx.index.GetTotalLocationCount())
}

// indexDirectory recursively indexes a directory.
func (idx *Indexer) indexDirectory(dirPath string, depth int) {
	// Check depth limit
	if depth > idx.maxDepth {
		return
	}

	// Check file count limit
	if idx.fileCount >= idx.maxFiles {
		return
	}

	// Read directory entries
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		// Silently skip directories we can't read (permissions, etc.)
		return
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())

		// Skip hidden files and directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Skip common directories to ignore
		if entry.IsDir() {
			switch entry.Name() {
			case "node_modules", "vendor", "bin", "obj", "dist", "build", "out", "__pycache__":
				continue
			}

			// Recursively index subdirectory
			idx.indexDirectory(fullPath, depth+1)

			continue
		}

		// Check if it's a .dws file
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".dws") {
			continue
		}

		// Index the file
		idx.indexFile(fullPath)
	}
}

// indexFile parses a file and adds its symbols to the index.
func (idx *Indexer) indexFile(filePath string) {
	// Check file count limit
	if idx.fileCount >= idx.maxFiles {
		return
	}

	// Read file contents
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Warning: Could not read file %s: %v\n", filePath, err)
		return
	}

	// Create DWScript engine and compile the file
	engine, err := dwscript.New()
	if err != nil {
		log.Printf("Warning: Could not create engine: %v\n", err)
		return
	}

	prog, err := engine.Compile(string(content))
	if err != nil {
		// Log parse errors but don't fail the indexing
		// (many files may have errors during development)
		return
	}

	if prog == nil || prog.AST() == nil {
		return
	}

	// Convert file path to URI
	uri := pathToURI(filePath)

	// Extract and index symbols
	idx.extractSymbols(uri, prog.AST())

	idx.fileCount++
	if idx.fileCount%100 == 0 {
		log.Printf("Indexed %d files so far...\n", idx.fileCount)
	}
}

// extractSymbols extracts top-level symbols from an AST.
func (idx *Indexer) extractSymbols(uri string, programAST *ast.Program) {
	if programAST == nil {
		return
	}

	// Traverse top-level statements
	for _, stmt := range programAST.Statements {
		if stmt == nil {
			continue
		}

		switch node := stmt.(type) {
		case *ast.FunctionDecl:
			idx.addFunctionSymbol(uri, node, "")

		case *ast.VarDeclStatement:
			idx.addVariableSymbols(uri, node, "")

		case *ast.ConstDecl:
			idx.addConstSymbol(uri, node, "")

		case *ast.ClassDecl:
			idx.addClassSymbol(uri, node)

		case *ast.RecordDecl:
			idx.addRecordSymbol(uri, node)

		case *ast.EnumDecl:
			idx.addEnumSymbol(uri, node)
		}
	}
}

// addFunctionSymbol adds a function symbol to the index.
func (idx *Indexer) addFunctionSymbol(uri string, fn *ast.FunctionDecl, containerName string) {
	if fn == nil || fn.Name == nil {
		return
	}

	// Build function signature
	detail := buildFunctionSignature(fn)

	// Get range
	start := fn.Pos()
	end := fn.End()

	symbolRange := protocol.Range{
		Start: protocol.Position{
			Line:      uint32(max(0, start.Line-1)),
			Character: uint32(max(0, start.Column-1)),
		},
		End: protocol.Position{
			Line:      uint32(max(0, end.Line-1)),
			Character: uint32(max(0, end.Column-1)),
		},
	}

	// Determine kind (function or method)
	kind := protocol.SymbolKindFunction
	if containerName != "" {
		kind = protocol.SymbolKindMethod
	}

	idx.index.AddSymbol(fn.Name.Value, kind, uri, symbolRange, containerName, detail)
}

// addVariableSymbols adds variable symbols to the index.
func (idx *Indexer) addVariableSymbols(uri string, varDecl *ast.VarDeclStatement, containerName string) {
	if varDecl == nil || len(varDecl.Names) == 0 {
		return
	}

	for _, name := range varDecl.Names {
		if name == nil {
			continue
		}

		detail := "var"
		if typeName := util.GetTypeName(varDecl.Type); typeName != "" {
			detail += ": " + typeName
		}

		start := varDecl.Pos()
		end := varDecl.End()

		symbolRange := protocol.Range{
			Start: protocol.Position{
				Line:      uint32(max(0, start.Line-1)),
				Character: uint32(max(0, start.Column-1)),
			},
			End: protocol.Position{
				Line:      uint32(max(0, end.Line-1)),
				Character: uint32(max(0, end.Column-1)),
			},
		}

		idx.index.AddSymbol(name.Value, protocol.SymbolKindVariable, uri, symbolRange, containerName, detail)
	}
}

// addConstSymbol adds a constant symbol to the index.
func (idx *Indexer) addConstSymbol(uri string, constDecl *ast.ConstDecl, containerName string) {
	if constDecl == nil || constDecl.Name == nil {
		return
	}

	detail := "const"
	if typeName := util.GetTypeName(constDecl.Type); typeName != "" {
		detail += ": " + typeName
	}

	if constDecl.Value != nil {
		detail += " = " + constDecl.Value.String()
	}

	start := constDecl.Pos()
	end := constDecl.End()

	symbolRange := protocol.Range{
		Start: protocol.Position{
			Line:      uint32(max(0, start.Line-1)),
			Character: uint32(max(0, start.Column-1)),
		},
		End: protocol.Position{
			Line:      uint32(max(0, end.Line-1)),
			Character: uint32(max(0, end.Column-1)),
		},
	}

	idx.index.AddSymbol(constDecl.Name.Value, protocol.SymbolKindConstant, uri, symbolRange, containerName, detail)
}

// addClassSymbol adds a class symbol to the index (including its members).
func (idx *Indexer) addClassSymbol(uri string, classDecl *ast.ClassDecl) {
	if classDecl == nil || classDecl.Name == nil {
		return
	}

	detail := "class"
	if classDecl.Parent != nil {
		detail += "(" + classDecl.Parent.Value + ")"
	}

	start := classDecl.Pos()
	end := classDecl.End()

	symbolRange := protocol.Range{
		Start: protocol.Position{
			Line:      uint32(max(0, start.Line-1)),
			Character: uint32(max(0, start.Column-1)),
		},
		End: protocol.Position{
			Line:      uint32(max(0, end.Line-1)),
			Character: uint32(max(0, end.Column-1)),
		},
	}

	className := classDecl.Name.Value
	idx.index.AddSymbol(className, protocol.SymbolKindClass, uri, symbolRange, "", detail)

	// Add class methods with class name as container
	for _, method := range classDecl.Methods {
		if method != nil {
			idx.addFunctionSymbol(uri, method, className)
		}
	}

	// Add class fields
	for _, field := range classDecl.Fields {
		if field == nil || field.Name == nil {
			continue
		}

		fieldDetail := "field"

		if field.Type != nil {
			if typeAnnot, ok := field.Type.(*ast.TypeAnnotation); ok && typeAnnot.Name != "" {
				fieldDetail += ": " + typeAnnot.Name
			}
		}

		fieldStart := field.Pos()
		fieldEnd := field.End()

		fieldRange := protocol.Range{
			Start: protocol.Position{
				Line:      uint32(max(0, fieldStart.Line-1)),
				Character: uint32(max(0, fieldStart.Column-1)),
			},
			End: protocol.Position{
				Line:      uint32(max(0, fieldEnd.Line-1)),
				Character: uint32(max(0, fieldEnd.Column-1)),
			},
		}

		idx.index.AddSymbol(field.Name.Value, protocol.SymbolKindField, uri, fieldRange, className, fieldDetail)
	}

	// Add class properties
	for _, prop := range classDecl.Properties {
		if prop == nil || prop.Name == nil {
			continue
		}

		propDetail := "property"
		if typeName := util.GetTypeName(prop.Type); typeName != "" {
			propDetail += ": " + typeName
		}

		propStart := prop.Pos()
		propEnd := prop.End()

		propRange := protocol.Range{
			Start: protocol.Position{
				Line:      uint32(max(0, propStart.Line-1)),
				Character: uint32(max(0, propStart.Column-1)),
			},
			End: protocol.Position{
				Line:      uint32(max(0, propEnd.Line-1)),
				Character: uint32(max(0, propEnd.Column-1)),
			},
		}

		idx.index.AddSymbol(prop.Name.Value, protocol.SymbolKindProperty, uri, propRange, className, propDetail)
	}
}

// addRecordSymbol adds a record symbol to the index.
func (idx *Indexer) addRecordSymbol(uri string, recordDecl *ast.RecordDecl) {
	if recordDecl == nil || recordDecl.Name == nil {
		return
	}

	detail := "record"

	start := recordDecl.Pos()
	end := recordDecl.End()

	symbolRange := protocol.Range{
		Start: protocol.Position{
			Line:      uint32(max(0, start.Line-1)),
			Character: uint32(max(0, start.Column-1)),
		},
		End: protocol.Position{
			Line:      uint32(max(0, end.Line-1)),
			Character: uint32(max(0, end.Column-1)),
		},
	}

	recordName := recordDecl.Name.Value
	idx.index.AddSymbol(recordName, protocol.SymbolKindStruct, uri, symbolRange, "", detail)

	// Add record properties as fields
	for i := range recordDecl.Properties {
		prop := &recordDecl.Properties[i]
		if prop.Name == nil {
			continue
		}

		propDetail := "field"
		if typeName := util.GetTypeName(prop.Type); typeName != "" {
			propDetail += ": " + typeName
		}

		propStart := prop.Name.Pos()
		propEnd := prop.End()

		propRange := protocol.Range{
			Start: protocol.Position{
				Line:      uint32(max(0, propStart.Line-1)),
				Character: uint32(max(0, propStart.Column-1)),
			},
			End: protocol.Position{
				Line:      uint32(max(0, propEnd.Line-1)),
				Character: uint32(max(0, propEnd.Column-1)),
			},
		}

		idx.index.AddSymbol(prop.Name.Value, protocol.SymbolKindField, uri, propRange, recordName, propDetail)
	}
}

// addEnumSymbol adds an enum symbol to the index.
func (idx *Indexer) addEnumSymbol(uri string, enumDecl *ast.EnumDecl) {
	if enumDecl == nil || enumDecl.Name == nil {
		return
	}

	detail := "enum"

	start := enumDecl.Pos()
	end := enumDecl.End()

	symbolRange := protocol.Range{
		Start: protocol.Position{
			Line:      uint32(max(0, start.Line-1)),
			Character: uint32(max(0, start.Column-1)),
		},
		End: protocol.Position{
			Line:      uint32(max(0, end.Line-1)),
			Character: uint32(max(0, end.Column-1)),
		},
	}

	enumName := enumDecl.Name.Value
	idx.index.AddSymbol(enumName, protocol.SymbolKindEnum, uri, symbolRange, "", detail)

	// Add enum values as enum members
	for _, value := range enumDecl.Values {
		if value.Name == "" {
			continue
		}

		valueDetail := "enum member"
		if value.Value != nil {
			valueDetail += " = " + string(rune(*value.Value))
		}

		// Enum values don't have precise position information
		// Use the enum's position as an approximation
		idx.index.AddSymbol(value.Name, protocol.SymbolKindEnumMember, uri, symbolRange, enumName, valueDetail)
	}
}

// buildFunctionSignature builds a function signature string.
func buildFunctionSignature(fn *ast.FunctionDecl) string {
	if fn == nil || fn.Name == nil {
		return ""
	}

	sig := "function " + fn.Name.Value + "("

	var sigSb483 strings.Builder

	for i, param := range fn.Parameters {
		if i > 0 {
			sigSb483.WriteString(", ")
		}

		if param.IsConst {
			sigSb483.WriteString("const ")
		}

		if param.IsLazy {
			sigSb483.WriteString("lazy ")
		}

		if param.ByRef {
			sigSb483.WriteString("var ")
		}

		if param.Name != nil {
			sigSb483.WriteString(param.Name.Value)

			if typeName := util.GetTypeName(param.Type); typeName != "" {
				sigSb483.WriteString(": " + typeName)
			}
		}

		if param.DefaultValue != nil {
			sigSb483.WriteString(" = " + param.DefaultValue.String())
		}
	}

	sig += sigSb483.String()

	sig += ")"

	if returnType := util.GetTypeName(fn.ReturnType); returnType != "" {
		sig += ": " + returnType
	}

	return sig
}

// uriToPath converts a URI to a file system path.
func uriToPath(uri string) string {
	// Handle file:// URIs
	if after, ok := strings.CutPrefix(uri, "file://"); ok {
		path := after
		// On Windows, URIs are like file:///C:/path, so we need to handle the leading slash
		if len(path) > 2 && path[0] == '/' && path[2] == ':' {
			path = path[1:] // Remove leading slash for Windows paths
		}

		return path
	}

	return uri
}

// pathToURI converts a file system path to a URI.
func pathToURI(path string) string {
	// Normalize path separators to forward slashes
	path = filepath.ToSlash(path)

	// On Windows, prepend an extra slash
	if len(path) > 1 && path[1] == ':' {
		return "file:///" + path
	}

	return "file://" + path
}

// IndexWorkspace is a helper function that creates an indexer and builds the workspace index.
func IndexWorkspace(index *SymbolIndex, workspaceFolders []protocol.WorkspaceFolder) {
	indexer := NewIndexer(index)
	indexer.BuildWorkspaceIndex(workspaceFolders)
}

// IndexWorkspaceAsync runs workspace indexing in a background goroutine.
func IndexWorkspaceAsync(index *SymbolIndex, workspaceFolders []protocol.WorkspaceFolder) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic in workspace indexing: %v\n", r)
			}
		}()

		IndexWorkspace(index, workspaceFolders)
	}()
}

// FallbackSearch performs on-demand symbol search when the index is not available.
// It walks through workspace folders, parses .dws files, and searches for symbols matching the query.
// This provides basic functionality while the index is being built.
func FallbackSearch(workspaceFolders []string, query string, maxResults int) []SymbolLocation {
	log.Printf("Warning: Symbol index not ready, using fallback search for query %q\n", query)

	if maxResults <= 0 {
		maxResults = 100 // Default limit
	}

	queryLower := strings.ToLower(query)
	var results []SymbolLocation
	filesSearched := 0
	maxFilesToSearch := 50 // Limit files to avoid blocking too long

	// Search each workspace folder
	for _, folder := range workspaceFolders {
		if len(results) >= maxResults {
			break
		}

		// Walk through the folder
		err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err // Return walk errors
			}

			// Skip directories
			if info.IsDir() {
				// Skip hidden directories and common build directories
				name := filepath.Base(path)
				if strings.HasPrefix(name, ".") {
					return filepath.SkipDir
				}

				switch name {
				case "node_modules", "vendor", "bin", "obj", "dist", "build", "out", "__pycache__":
					return filepath.SkipDir
				}

				return nil
			}

			// Only process .dws files
			if !strings.HasSuffix(strings.ToLower(path), ".dws") {
				return nil
			}

			// Check limits
			if filesSearched >= maxFilesToSearch {
				return filepath.SkipAll
			}

			if len(results) >= maxResults {
				return filepath.SkipAll
			}

			filesSearched++

			// Parse file and search for symbols
			fileResults := searchFileForSymbols(path, queryLower)
			results = append(results, fileResults...)

			return nil
		})
		if err != nil {
			log.Printf("Warning: Error walking workspace folder %s: %v\n", folder, err)
		}
	}

	log.Printf("Fallback search found %d results from %d files\n", len(results), filesSearched)

	// Limit results
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results
}

// searchFileForSymbols parses a single file and returns symbols matching the query.
func searchFileForSymbols(filePath string, queryLower string) []SymbolLocation {
	// Read file contents
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	// Parse the file
	engine, err := dwscript.New()
	if err != nil {
		return nil
	}

	prog, err := engine.Compile(string(content))
	if err != nil {
		return nil
	}

	if prog == nil || prog.AST() == nil {
		return nil
	}

	// Convert file path to URI
	uri := pathToURI(filePath)

	// Extract all symbols from the file
	var allSymbols []SymbolLocation
	extractSymbolsForSearch(uri, prog.AST(), &allSymbols)

	// Filter symbols by query
	var matchingSymbols []SymbolLocation

	for _, sym := range allSymbols {
		nameLower := strings.ToLower(sym.Name)

		// Check if symbol matches query
		if queryLower == "" || strings.Contains(nameLower, queryLower) {
			matchingSymbols = append(matchingSymbols, sym)
		}
	}

	return matchingSymbols
}

// extractSymbolsForSearch extracts symbols from AST for fallback search.
// This is a simplified version that doesn't use the full indexer machinery.
func extractSymbolsForSearch(uri string, programAST *ast.Program, results *[]SymbolLocation) {
	if programAST == nil {
		return
	}

	// Traverse top-level statements
	for _, stmt := range programAST.Statements {
		if stmt == nil {
			continue
		}

		switch node := stmt.(type) {
		case *ast.FunctionDecl:
			if node.Name != nil {
				start := node.Pos()
				end := node.End()
				*results = append(*results, SymbolLocation{
					Name: node.Name.Value,
					Kind: protocol.SymbolKindFunction,
					Location: protocol.Location{
						URI: uri,
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      uint32(max(0, start.Line-1)),
								Character: uint32(max(0, start.Column-1)),
							},
							End: protocol.Position{
								Line:      uint32(max(0, end.Line-1)),
								Character: uint32(max(0, end.Column-1)),
							},
						},
					},
					ContainerName: "",
					Detail:        "",
				})
			}

		case *ast.VarDeclStatement:
			for _, name := range node.Names {
				if name != nil {
					start := node.Pos()
					end := node.End()
					*results = append(*results, SymbolLocation{
						Name: name.Value,
						Kind: protocol.SymbolKindVariable,
						Location: protocol.Location{
							URI: uri,
							Range: protocol.Range{
								Start: protocol.Position{
									Line:      uint32(max(0, start.Line-1)),
									Character: uint32(max(0, start.Column-1)),
								},
								End: protocol.Position{
									Line:      uint32(max(0, end.Line-1)),
									Character: uint32(max(0, end.Column-1)),
								},
							},
						},
						ContainerName: "",
						Detail:        "",
					})
				}
			}

		case *ast.ConstDecl:
			if node.Name != nil {
				start := node.Pos()
				end := node.End()
				*results = append(*results, SymbolLocation{
					Name: node.Name.Value,
					Kind: protocol.SymbolKindConstant,
					Location: protocol.Location{
						URI: uri,
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      uint32(max(0, start.Line-1)),
								Character: uint32(max(0, start.Column-1)),
							},
							End: protocol.Position{
								Line:      uint32(max(0, end.Line-1)),
								Character: uint32(max(0, end.Column-1)),
							},
						},
					},
					ContainerName: "",
					Detail:        "",
				})
			}

		case *ast.ClassDecl:
			if node.Name != nil {
				start := node.Pos()
				end := node.End()
				*results = append(*results, SymbolLocation{
					Name: node.Name.Value,
					Kind: protocol.SymbolKindClass,
					Location: protocol.Location{
						URI: uri,
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      uint32(max(0, start.Line-1)),
								Character: uint32(max(0, start.Column-1)),
							},
							End: protocol.Position{
								Line:      uint32(max(0, end.Line-1)),
								Character: uint32(max(0, end.Column-1)),
							},
						},
					},
					ContainerName: "",
					Detail:        "",
				})
			}

		case *ast.RecordDecl:
			if node.Name != nil {
				start := node.Pos()
				end := node.End()
				*results = append(*results, SymbolLocation{
					Name: node.Name.Value,
					Kind: protocol.SymbolKindStruct,
					Location: protocol.Location{
						URI: uri,
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      uint32(max(0, start.Line-1)),
								Character: uint32(max(0, start.Column-1)),
							},
							End: protocol.Position{
								Line:      uint32(max(0, end.Line-1)),
								Character: uint32(max(0, end.Column-1)),
							},
						},
					},
					ContainerName: "",
					Detail:        "",
				})
			}

		case *ast.EnumDecl:
			if node.Name != nil {
				start := node.Pos()
				end := node.End()
				*results = append(*results, SymbolLocation{
					Name: node.Name.Value,
					Kind: protocol.SymbolKindEnum,
					Location: protocol.Location{
						URI: uri,
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      uint32(max(0, start.Line-1)),
								Character: uint32(max(0, start.Column-1)),
							},
							End: protocol.Position{
								Line:      uint32(max(0, end.Line-1)),
								Character: uint32(max(0, end.Column-1)),
							},
						},
					},
					ContainerName: "",
					Detail:        "",
				})
			}
		}
	}
}

// Compile-time check to ensure io.Reader is implemented if needed.
var _ io.Reader
