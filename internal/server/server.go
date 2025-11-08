// Package server provides the core LSP server state and management.
package server

import (
	"sync"

	"github.com/CWBudde/go-dws-lsp/internal/workspace"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// Server holds the state of the LSP server.
type Server struct {
	// documents stores all open documents
	documents *DocumentStore

	// symbolIndex caches references for workspace documents (even when not open)
	symbolIndex *SymbolIndex

	// workspaceIndex stores workspace-wide symbol definitions for global symbol search
	workspaceIndex *workspace.SymbolIndex

	// workspaceFolders stores the workspace folders from the client
	workspaceFolders []string

	// clientCapabilities stores the client's capabilities from the initialize request
	clientCapabilities *protocol.ClientCapabilities

	// config holds server configuration
	config *Config

	// mutex protects server state
	mu sync.RWMutex

	// shutting down flag
	shuttingDown bool
}

// Config holds server configuration options.
type Config struct {
	// MaxProblems limits the number of diagnostics reported
	MaxProblems int

	// Trace controls logging verbosity
	Trace string
}

// New creates a new LSP server instance.
func New() *Server {
	return &Server{
		documents:      NewDocumentStore(),
		symbolIndex:    NewSymbolIndex(),
		workspaceIndex: workspace.NewSymbolIndex(),
		config: &Config{
			MaxProblems: 100,
			Trace:       "off",
		},
	}
}

// IsShuttingDown returns true if the server is shutting down.
func (s *Server) IsShuttingDown() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.shuttingDown
}

// SetShuttingDown marks the server as shutting down.
func (s *Server) SetShuttingDown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shuttingDown = true
}

// Documents returns the document store.
func (s *Server) Documents() *DocumentStore {
	return s.documents
}

// Symbols returns the workspace symbol index.
func (s *Server) Symbols() *SymbolIndex {
	return s.symbolIndex
}

// WorkspaceIndex returns the workspace-wide symbol index.
func (s *Server) WorkspaceIndex() *workspace.SymbolIndex {
	return s.workspaceIndex
}

// Config returns the server configuration.
func (s *Server) Config() *Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// UpdateConfig updates the server configuration atomically.
// The update function is called with the current config under a write lock.
func (s *Server) UpdateConfig(update func(*Config)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	update(s.config)
}

// SetWorkspaceFolders sets the workspace folders.
func (s *Server) SetWorkspaceFolders(folders []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workspaceFolders = folders
}

// GetWorkspaceFolders returns the workspace folders.
func (s *Server) GetWorkspaceFolders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.workspaceFolders
}

// SetClientCapabilities sets the client's capabilities.
func (s *Server) SetClientCapabilities(capabilities *protocol.ClientCapabilities) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clientCapabilities = capabilities
}

// GetClientCapabilities returns the client's capabilities.
func (s *Server) GetClientCapabilities() *protocol.ClientCapabilities {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clientCapabilities
}

// SupportsSnippets returns true if the client supports snippet completions.
func (s *Server) SupportsSnippets() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.clientCapabilities == nil {
		return false
	}

	if s.clientCapabilities.TextDocument == nil {
		return false
	}

	if s.clientCapabilities.TextDocument.Completion == nil {
		return false
	}

	if s.clientCapabilities.TextDocument.Completion.CompletionItem == nil {
		return false
	}

	if s.clientCapabilities.TextDocument.Completion.CompletionItem.SnippetSupport == nil {
		return false
	}

	return *s.clientCapabilities.TextDocument.Completion.CompletionItem.SnippetSupport
}
