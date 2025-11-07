package analysis

import (
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
)

// uriToPath converts a file:// URI into an OS-specific absolute path.
func uriToPath(u string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	if parsed.Scheme != "file" && parsed.Scheme != "" {
		return "", fmt.Errorf("unsupported URI scheme: %s", parsed.Scheme)
	}

	path := parsed.Path
	if path == "" {
		path = parsed.Opaque
	}

	decoded, err := url.PathUnescape(path)
	if err == nil {
		path = decoded
	}

	if runtime.GOOS == "windows" {
		if strings.HasPrefix(path, "/") && len(path) >= 3 && path[2] == ':' {
			path = path[1:]
		}
	}

	if path == "" {
		return "", fmt.Errorf("empty path extracted from URI: %s", u)
	}

	return filepath.FromSlash(path), nil
}
