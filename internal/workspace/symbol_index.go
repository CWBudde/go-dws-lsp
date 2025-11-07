// Package workspace provides workspace-wide symbol indexing and management.
package workspace

import (
	"log"
	"strings"
	"sync"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

// SymbolLocation represents a location where a symbol is defined.
type SymbolLocation struct {
	Name          string                  // Symbol name
	Kind          protocol.SymbolKind     // Symbol kind (Function, Class, Variable, etc.)
	Location      protocol.Location       // Full location with URI and range
	ContainerName string                  // Name of containing scope (e.g., class name for methods)
	Detail        string                  // Additional detail (e.g., signature)
}

// FileInfo stores metadata about an indexed file.
type FileInfo struct {
	URI     string   // Document URI
	Version int32    // Document version
	Symbols []string // List of symbol names defined in this file
}

// SymbolIndex maintains a workspace-wide index of symbols.
// It provides thread-safe access to symbol information across all files.
type SymbolIndex struct {
	// symbols maps symbol names to their locations
	// Multiple locations support function overloading and same names in different files
	symbols map[string][]SymbolLocation

	// files maps document URIs to file metadata
	files map[string]*FileInfo

	// mutex protects concurrent access to the index
	mutex sync.RWMutex
}

// NewSymbolIndex creates a new empty symbol index.
func NewSymbolIndex() *SymbolIndex {
	return &SymbolIndex{
		symbols: make(map[string][]SymbolLocation),
		files:   make(map[string]*FileInfo),
	}
}

// AddSymbol adds a symbol to the index.
// If the symbol already exists from the same file, it will be updated.
func (si *SymbolIndex) AddSymbol(name string, kind protocol.SymbolKind, uri string, symbolRange protocol.Range, containerName string, detail string) {
	si.mutex.Lock()
	defer si.mutex.Unlock()

	location := SymbolLocation{
		Name: name,
		Kind: kind,
		Location: protocol.Location{
			URI:   uri,
			Range: symbolRange,
		},
		ContainerName: containerName,
		Detail:        detail,
	}

	// Add to symbols map
	si.symbols[name] = append(si.symbols[name], location)

	// Update file info
	fileInfo, exists := si.files[uri]
	if !exists {
		fileInfo = &FileInfo{
			URI:     uri,
			Symbols: make([]string, 0),
		}
		si.files[uri] = fileInfo
	}

	// Track that this file defines this symbol
	fileInfo.Symbols = append(fileInfo.Symbols, name)

	log.Printf("Indexed symbol '%s' (%v) in %s", name, kind, uri)
}

// FindSymbol searches for all locations where a symbol is defined.
// Returns an empty slice if the symbol is not found.
func (si *SymbolIndex) FindSymbol(name string) []SymbolLocation {
	si.mutex.RLock()
	defer si.mutex.RUnlock()

	locations, exists := si.symbols[name]
	if !exists {
		return nil
	}

	// Return a copy to avoid external modifications
	result := make([]SymbolLocation, len(locations))
	copy(result, locations)
	return result
}

// FindSymbolsByKind searches for symbols of a specific kind.
func (si *SymbolIndex) FindSymbolsByKind(kind protocol.SymbolKind) []SymbolLocation {
	si.mutex.RLock()
	defer si.mutex.RUnlock()

	var result []SymbolLocation
	for _, locations := range si.symbols {
		for _, loc := range locations {
			if loc.Kind == kind {
				result = append(result, loc)
			}
		}
	}
	return result
}

// FindSymbolsInFile returns all symbols defined in a specific file.
func (si *SymbolIndex) FindSymbolsInFile(uri string) []SymbolLocation {
	si.mutex.RLock()
	defer si.mutex.RUnlock()

	fileInfo, exists := si.files[uri]
	if !exists {
		return nil
	}

	var result []SymbolLocation
	for _, symbolName := range fileInfo.Symbols {
		locations := si.symbols[symbolName]
		for _, loc := range locations {
			if loc.Location.URI == uri {
				result = append(result, loc)
			}
		}
	}
	return result
}

// RemoveFile removes all symbols from a file.
// This should be called when a file is deleted or before re-indexing.
func (si *SymbolIndex) RemoveFile(uri string) {
	si.mutex.Lock()
	defer si.mutex.Unlock()

	fileInfo, exists := si.files[uri]
	if !exists {
		return
	}

	// Remove symbols that belong to this file
	for _, symbolName := range fileInfo.Symbols {
		locations := si.symbols[symbolName]

		// Filter out locations from this file
		var remaining []SymbolLocation
		for _, loc := range locations {
			if loc.Location.URI != uri {
				remaining = append(remaining, loc)
			}
		}

		// Update or remove the symbol entry
		if len(remaining) > 0 {
			si.symbols[symbolName] = remaining
		} else {
			delete(si.symbols, symbolName)
		}
	}

	// Remove file info
	delete(si.files, uri)

	log.Printf("Removed all symbols from file: %s", uri)
}

// UpdateFileVersion updates the version number for a file.
func (si *SymbolIndex) UpdateFileVersion(uri string, version int32) {
	si.mutex.Lock()
	defer si.mutex.Unlock()

	if fileInfo, exists := si.files[uri]; exists {
		fileInfo.Version = version
	}
}

// GetFileCount returns the number of files in the index.
func (si *SymbolIndex) GetFileCount() int {
	si.mutex.RLock()
	defer si.mutex.RUnlock()
	return len(si.files)
}

// GetSymbolCount returns the total number of unique symbol names in the index.
func (si *SymbolIndex) GetSymbolCount() int {
	si.mutex.RLock()
	defer si.mutex.RUnlock()
	return len(si.symbols)
}

// GetTotalLocationCount returns the total number of symbol locations across all symbols.
func (si *SymbolIndex) GetTotalLocationCount() int {
	si.mutex.RLock()
	defer si.mutex.RUnlock()

	count := 0
	for _, locations := range si.symbols {
		count += len(locations)
	}
	return count
}

// Clear removes all symbols and file information from the index.
func (si *SymbolIndex) Clear() {
	si.mutex.Lock()
	defer si.mutex.Unlock()

	si.symbols = make(map[string][]SymbolLocation)
	si.files = make(map[string]*FileInfo)

	log.Println("Symbol index cleared")
}

// Search searches for symbols matching the query string.
// Returns symbols whose names contain the query string (case-insensitive).
// If query is empty, returns all symbols (up to a reasonable limit).
func (si *SymbolIndex) Search(query string, maxResults int) []SymbolLocation {
	si.mutex.RLock()
	defer si.mutex.RUnlock()

	// Normalize query for case-insensitive search
	queryLower := strings.ToLower(query)

	var results []SymbolLocation

	// If query is empty, return all symbols (up to limit)
	if query == "" {
		for _, locations := range si.symbols {
			for _, loc := range locations {
				results = append(results, loc)
				if maxResults > 0 && len(results) >= maxResults {
					return results
				}
			}
		}
		return results
	}

	// Search for symbols containing the query
	for symbolName, locations := range si.symbols {
		nameLower := strings.ToLower(symbolName)
		if strings.Contains(nameLower, queryLower) {
			for _, loc := range locations {
				results = append(results, loc)
				if maxResults > 0 && len(results) >= maxResults {
					return results
				}
			}
		}
	}

	return results
}
