// Package server provides the core LSP server state and management.
package server

import (
	"sync"
)

// Server holds the state of the LSP server.
type Server struct {
	// documents stores all open documents
	documents *DocumentStore
	
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
		documents: NewDocumentStore(),
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

// Config returns the server configuration.
func (s *Server) Config() *Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}
