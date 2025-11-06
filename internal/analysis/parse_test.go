package analysis

import (
	"strings"
	"testing"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestParseDocument_ValidCode(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "simple variable declaration",
			source: `var x: Integer;
begin
	x := 42;
end.`,
		},
		{
			name: "function definition",
			source: `function Add(a, b: Integer): Integer;
begin
	Result := a + b;
end;

begin
	PrintLn(Add(5, 3));
end.`,
		},
		{
			name: "empty program",
			source: `begin
end.`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnostics, err := ParseDocument(tt.source, "test.dws")
			if err != nil {
				t.Fatalf("ParseDocument returned unexpected error: %v", err)
			}

			if len(diagnostics) > 0 {
				t.Errorf("Expected no diagnostics for valid code, got %d:", len(diagnostics))
				for i, diag := range diagnostics {
					t.Errorf("  [%d] Line %d: %s", i+1, diag.Range.Start.Line, diag.Message)
				}
			}
		})
	}
}

func TestParseDocument_SyntaxErrors(t *testing.T) {
	tests := []struct {
		name              string
		source            string
		expectedErrorsMin int
	}{
		{
			name: "missing semicolon",
			source: `var x: Integer
begin
	x := 42;
end.`,
			expectedErrorsMin: 1,
		},
		{
			name: "unclosed string",
			source: `begin
	PrintLn('Hello World);
end.`,
			expectedErrorsMin: 1,
		},
		{
			name: "undefined variable",
			source: `begin
	x := 42;
end.`,
			expectedErrorsMin: 1,
		},
		{
			name: "missing end keyword",
			source: `begin
	PrintLn('test');
.`,
			expectedErrorsMin: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnostics, err := ParseDocument(tt.source, "test.dws")
			if err != nil {
				t.Fatalf("ParseDocument returned unexpected error: %v", err)
			}

			if len(diagnostics) < tt.expectedErrorsMin {
				t.Errorf("Expected at least %d diagnostic(s), got %d",
					tt.expectedErrorsMin, len(diagnostics))
			}

			// Verify all diagnostics have proper structure
			for i, diag := range diagnostics {
				if diag.Message == "" {
					t.Errorf("Diagnostic %d has empty message", i)
				}
				if diag.Severity == nil {
					t.Errorf("Diagnostic %d has nil severity", i)
				}
				if diag.Source == nil || *diag.Source == "" {
					t.Errorf("Diagnostic %d has empty source", i)
				}
			}
		})
	}
}

func TestParseDocument_EmptySource(t *testing.T) {
	diagnostics, err := ParseDocument("", "test.dws")
	if err != nil {
		t.Fatalf("ParseDocument returned unexpected error: %v", err)
	}

	// Empty source is actually valid in DWScript - it's an empty program
	// So we don't expect diagnostics
	_ = diagnostics // Just verify we got a result
}

func TestConvertCompileErrors(t *testing.T) {
	tests := []struct {
		name   string
		errors []string
		source string
	}{
		{
			name:   "empty errors",
			errors: []string{},
			source: "test",
		},
		{
			name: "simple error message",
			errors: []string{
				"Syntax Error: unexpected token [line 5]",
			},
			source: "test source",
		},
		{
			name: "multiple errors",
			errors: []string{
				"Error at line 1, col 5: missing semicolon",
				"Error at line 3, col 10: undefined identifier",
			},
			source: "test source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnostics := convertCompileErrors(tt.errors, tt.source)

			if len(diagnostics) != len(tt.errors) {
				t.Errorf("Expected %d diagnostics, got %d", len(tt.errors), len(diagnostics))
			}

			for i, diag := range diagnostics {
				if diag.Message == "" {
					t.Errorf("Diagnostic %d has empty message", i)
				}
				if diag.Severity == nil || *diag.Severity != protocol.DiagnosticSeverityError {
					t.Errorf("Diagnostic %d should have Error severity", i)
				}
			}
		})
	}
}

func TestParseErrorMessage(t *testing.T) {
	tests := []struct {
		name         string
		errorMsg     string
		expectedLine uint32
		expectedCol  uint32
	}{
		{
			name:         "line only format",
			errorMsg:     "Syntax Error: unexpected token [line 5]",
			expectedLine: 4, // 0-based
			expectedCol:  0,
		},
		{
			name:         "line and column format",
			errorMsg:     "Error at line 10, col 25: missing semicolon",
			expectedLine: 9,  // 0-based
			expectedCol:  24, // 0-based
		},
		{
			name:         "parentheses format",
			errorMsg:     "Type mismatch (15,8)",
			expectedLine: 14, // 0-based
			expectedCol:  7,  // 0-based
		},
		{
			name:         "no position info",
			errorMsg:     "General compilation error",
			expectedLine: 0,
			expectedCol:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnostic := parseErrorMessage(tt.errorMsg, "test source")

			if diagnostic.Range.Start.Line != tt.expectedLine {
				t.Errorf("Expected line %d, got %d", tt.expectedLine, diagnostic.Range.Start.Line)
			}
			if diagnostic.Range.Start.Character != tt.expectedCol {
				t.Errorf("Expected column %d, got %d", tt.expectedCol, diagnostic.Range.Start.Character)
			}
			if diagnostic.Message == "" {
				t.Error("Diagnostic message should not be empty")
			}
		})
	}
}

func TestCleanErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "remove line suffix",
			message:  "Syntax Error: unexpected token [line 5]",
			expected: "Syntax Error: unexpected token",
		},
		{
			name:     "remove line col prefix",
			message:  "line 10, col 25: missing semicolon",
			expected: "missing semicolon",
		},
		{
			name:     "remove parentheses position",
			message:  "Type mismatch (15,8)",
			expected: "Type mismatch",
		},
		{
			name:     "no position info",
			message:  "General compilation error",
			expected: "General compilation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaned := cleanErrorMessage(tt.message)
			// Trim spaces for comparison
			cleaned = strings.TrimSpace(cleaned)
			expected := strings.TrimSpace(tt.expected)

			if cleaned != expected {
				t.Errorf("Expected '%s', got '%s'", expected, cleaned)
			}
		})
	}
}

func TestCreateDiagnostic(t *testing.T) {
	tests := []struct {
		name         string
		line         int
		col          int
		length       int
		message      string
		expectedLine uint32
		expectedCol  uint32
	}{
		{
			name:         "convert 1-based to 0-based",
			line:         5,
			col:          10,
			length:       5,
			message:      "Test error",
			expectedLine: 4,
			expectedCol:  9,
		},
		{
			name:         "handle zero line",
			line:         0,
			col:          0,
			length:       1,
			message:      "Test error",
			expectedLine: 0,
			expectedCol:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnostic := createDiagnostic(tt.line, tt.col, tt.length, tt.message)

			if diagnostic.Range.Start.Line != tt.expectedLine {
				t.Errorf("Expected line %d, got %d", tt.expectedLine, diagnostic.Range.Start.Line)
			}
			if diagnostic.Range.Start.Character != tt.expectedCol {
				t.Errorf("Expected column %d, got %d", tt.expectedCol, diagnostic.Range.Start.Character)
			}
			if diagnostic.Severity == nil {
				t.Error("Severity should not be nil")
			}
		})
	}
}

func TestParseDocument_SemanticErrors(t *testing.T) {
	tests := []struct {
		name              string
		source            string
		expectedErrorsMin int
	}{
		{
			name: "type mismatch",
			source: `var x: Integer;
begin
	x := 'string value';
end.`,
			expectedErrorsMin: 1,
		},
		{
			name: "wrong argument count",
			source: `function Add(a, b: Integer): Integer;
begin
	Result := a + b;
end;

begin
	PrintLn(Add(5));
end.`,
			expectedErrorsMin: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnostics, err := ParseDocument(tt.source, "test.dws")
			if err != nil {
				t.Fatalf("ParseDocument returned unexpected error: %v", err)
			}

			if len(diagnostics) < tt.expectedErrorsMin {
				t.Errorf("Expected at least %d diagnostic(s), got %d",
					tt.expectedErrorsMin, len(diagnostics))
				for i, diag := range diagnostics {
					t.Logf("  [%d] Line %d: %s", i+1, diag.Range.Start.Line, diag.Message)
				}
			}
		})
	}
}
