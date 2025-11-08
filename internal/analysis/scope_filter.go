package analysis

import (
	"log"
	"os"

	"github.com/CWBudde/go-dws-lsp/internal/server"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/dwscript"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// FilterByScope removes references whose scope does not match the target scope.
// It inspects each reference location, determines its scope, and keeps it only if
// the scope matches. For documents not currently open, it attempts to parse them
// from disk on-demand.
func FilterByScope(references []protocol.Location, docStore *server.DocumentStore, targetName string, targetScope *ScopeInfo) []protocol.Location {
	if len(references) == 0 || targetScope == nil || targetScope.Type == ScopeUnknown {
		return references
	}

	cache := make(map[string]*dwscript.Program)
	var filtered []protocol.Location

	for _, loc := range references {
		program := programForURI(loc.URI, docStore, cache)
		if program == nil || program.AST() == nil {
			// Can't determine scope, keep the reference to avoid false negatives.
			filtered = append(filtered, loc)
			continue
		}

		refScope := DetermineScope(program.AST(), targetName, Position{
			Line:   int(loc.Range.Start.Line) + 1,
			Column: int(loc.Range.Start.Character) + 1,
		})

		if scopeMatches(targetScope, refScope) {
			filtered = append(filtered, loc)
		}
	}

	return filtered
}

func programForURI(uri string, docStore *server.DocumentStore, cache map[string]*dwscript.Program) *dwscript.Program {
	if prog, ok := cache[uri]; ok {
		return prog
	}

	if docStore != nil {
		if doc, ok := docStore.Get(uri); ok && doc != nil && doc.Program != nil {
			cache[uri] = doc.Program
			return doc.Program
		}
	}

	prog := parseProgramFromDisk(uri)
	cache[uri] = prog

	return prog
}

func parseProgramFromDisk(uri string) *dwscript.Program {
	path, err := uriToPath(uri)
	if err != nil {
		log.Printf("FilterByScope: unable to resolve path for %s: %v", uri, err)
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("FilterByScope: unable to read %s: %v", path, err)
		return nil
	}

	program, _, err := ParseDocument(string(data), uri)
	if err != nil {
		log.Printf("FilterByScope: unable to parse %s: %v", uri, err)
		return nil
	}

	return program
}

func scopeMatches(target, candidate *ScopeInfo) bool {
	if candidate == nil {
		return false
	}

	if target.Type != candidate.Type {
		return false
	}

	switch target.Type {
	case ScopeLocal, ScopeParameter:
		return sameFunction(target.Function, candidate.Function)
	case ScopeClassMember:
		return sameClass(target.Class, candidate.Class)
	default:
		return true
	}
}

func sameFunction(a, b *ast.FunctionDecl) bool {
	if a == nil || b == nil {
		return false
	}

	if a.Name == nil || b.Name == nil {
		return false
	}

	return a.Name.Value == b.Name.Value
}

func sameClass(a, b *ast.ClassDecl) bool {
	if a == nil || b == nil {
		return false
	}

	if a.Name == nil || b.Name == nil {
		return false
	}

	return a.Name.Value == b.Name.Value
}
