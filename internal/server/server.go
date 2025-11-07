// Package server provides the core LSP server state and management.
package server

import (
	"sync"
)

// Server holds the state of the LSP server.
type Server struct {
	// documents stores all open documents
	documents *DocumentStore

	// symbolIndex caches references for workspace documents (even when not open)
	symbolIndex *SymbolIndex

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
		documents:   NewDocumentStore(),
		symbolIndex: NewSymbolIndex(),
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
