package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	glspserver "github.com/tliron/glsp/server"

	"github.com/CWBudde/go-dws-lsp/internal/lsp"
	"github.com/CWBudde/go-dws-lsp/internal/server"
)

const (
	version = "0.1.0"
)

var (
	tcpMode  bool
	tcpPort  int
	logLevel string
	logFile  string
)

func init() {
	// Command-line flags
	flag.BoolVar(&tcpMode, "tcp", false, "Run server in TCP mode (for debugging)")
	flag.IntVar(&tcpPort, "port", 8765, "TCP port to listen on (used with -tcp)")
	flag.StringVar(&logLevel, "log-level", "error", "Log level: debug, info, warn, error")
	flag.StringVar(&logFile, "log-file", "", "Log file path (default: stderr)")
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "go-dws-lsp version %s\n\n", version)
	fmt.Fprintf(os.Stderr, "Usage: go-dws-lsp [options]\n\n")
	fmt.Fprintf(os.Stderr, "Language Server Protocol implementation for DWScript\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	// Print version if requested
	if flag.NArg() > 0 && flag.Arg(0) == "version" {
		fmt.Printf("go-dws-lsp version %s\n", version)
		os.Exit(0)
	}

	fmt.Fprintf(os.Stderr, "go-dws-lsp version %s starting...\n", version)
	fmt.Fprintf(os.Stderr, "Transport: ")
	if tcpMode {
		fmt.Fprintf(os.Stderr, "TCP (port %d)\n", tcpPort)
	} else {
		fmt.Fprintf(os.Stderr, "STDIO\n")
	}
	fmt.Fprintf(os.Stderr, "Log level: %s\n", logLevel)

	// Initialize server state
	srv := server.New()

	// Set up logging
	setupLogging()

	// Create GLSP handler
	handler := protocol.Handler{
		Initialize:  lsp.Initialize,
		Initialized: lsp.Initialized,
		Shutdown:    lsp.Shutdown,
		SetTrace:    func(context *glsp.Context, params *protocol.SetTraceParams) error { return nil },
	}

	// Create GLSP server
	glspServer := glspserver.NewServer(&handler, "go-dws-lsp", false)

	// Store our server instance for handler access
	lsp.SetServer(srv)

	// Start server with appropriate transport
	if tcpMode {
		fmt.Fprintf(os.Stderr, "Starting TCP server on port %d...\n", tcpPort)
		if err := glspServer.RunTCP(fmt.Sprintf("127.0.0.1:%d", tcpPort)); err != nil {
			log.Fatalf("TCP server error: %v", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Starting STDIO server...\n")
		if err := glspServer.RunStdio(); err != nil {
			log.Fatalf("STDIO server error: %v", err)
		}
	}
}

// setupLogging configures the logging system based on command-line flags.
func setupLogging() {
	// Set log output
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
			os.Exit(1)
		}
		log.SetOutput(f)
	} else {
		log.SetOutput(os.Stderr)
	}

	// Set log flags
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Log level filtering is basic with standard log package
	// For now, we'll just log everything and rely on server-side filtering if needed
}
