package main

import (
	"flag"
	"fmt"
	"os"
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

	// TODO: Initialize server
	// TODO: Setup logging
	// TODO: Start STDIO or TCP transport
	
	fmt.Fprintf(os.Stderr, "Server initialization not yet implemented\n")
	os.Exit(1)
}
