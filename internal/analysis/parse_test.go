package analysis

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/dwscript"
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
			program, diagnostics, err := ParseDocument(tt.source, "test.dws")
			if err != nil {
				t.Fatalf("ParseDocument returned unexpected error: %v", err)
			}

			if program == nil {
				t.Error("Expected non-nil program for valid code")
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
			_, diagnostics, err := ParseDocument(tt.source, "test.dws")
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
	_, diagnostics, err := ParseDocument("", "test.dws")
	if err != nil {
		t.Fatalf("ParseDocument returned unexpected error: %v", err)
	}

	// Empty source is actually valid in DWScript - it's an empty program
	// So we don't expect diagnostics
	_ = diagnostics // Just verify we got a result
}

func TestConvertStructuredErrors(t *testing.T) {
	tests := []struct {
		name   string
		errors []*dwscript.Error
	}{
		{
			name:   "empty errors",
			errors: []*dwscript.Error{},
		},
		{
			name: "single error",
			errors: []*dwscript.Error{
				{
					Message:  "Syntax Error: unexpected token",
					Line:     5,
					Column:   10,
					Length:   5,
					Severity: dwscript.SeverityError,
					Code:     "",
				},
			},
		},
		{
			name: "multiple errors with different severities",
			errors: []*dwscript.Error{
				{
					Message:  "Undefined identifier",
					Line:     1,
					Column:   5,
					Length:   8,
					Severity: dwscript.SeverityError,
					Code:     "E_UNDEFINED_VAR",
				},
				{
					Message:  "Unused variable",
					Line:     3,
					Column:   10,
					Length:   4,
					Severity: dwscript.SeverityWarning,
					Code:     "W_UNUSED_VAR",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnostics := convertStructuredErrors(tt.errors)

			if len(diagnostics) != len(tt.errors) {
				t.Errorf("Expected %d diagnostics, got %d", len(tt.errors), len(diagnostics))
			}

			for i, diag := range diagnostics {
				if diag.Message == "" {
					t.Errorf("Diagnostic %d has empty message", i)
				}
				if diag.Severity == nil {
					t.Errorf("Diagnostic %d has nil severity", i)
				}
				if diag.Source == nil || *diag.Source != "go-dws" {
					t.Errorf("Diagnostic %d has incorrect source", i)
				}
			}
		})
	}
}

func TestConvertStructuredError(t *testing.T) {
	tests := []struct {
		name         string
		error        *dwscript.Error
		expectedLine uint32
		expectedCol  uint32
	}{
		{
			name: "1-based to 0-based conversion",
			error: &dwscript.Error{
				Message:  "Test error",
				Line:     5,
				Column:   10,
				Length:   5,
				Severity: dwscript.SeverityError,
				Code:     "",
			},
			expectedLine: 4, // 0-based
			expectedCol:  9, // 0-based
		},
		{
			name: "warning with unused tag",
			error: &dwscript.Error{
				Message:  "Unused variable 'x'",
				Line:     1,
				Column:   5,
				Length:   1,
				Severity: dwscript.SeverityWarning,
				Code:     "W_UNUSED_VAR",
			},
			expectedLine: 0,
			expectedCol:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnostic := convertStructuredError(tt.error)

			if diagnostic.Range.Start.Line != tt.expectedLine {
				t.Errorf("Expected line %d, got %d", tt.expectedLine, diagnostic.Range.Start.Line)
			}
			if diagnostic.Range.Start.Character != tt.expectedCol {
				t.Errorf("Expected column %d, got %d", tt.expectedCol, diagnostic.Range.Start.Character)
			}
			if diagnostic.Message != tt.error.Message {
				t.Errorf("Expected message '%s', got '%s'", tt.error.Message, diagnostic.Message)
			}

			// Verify tags for unused variables
			if tt.error.Code == "W_UNUSED_VAR" {
				if len(diagnostic.Tags) == 0 {
					t.Error("Expected Unnecessary tag for unused variable")
				} else if diagnostic.Tags[0] != protocol.DiagnosticTagUnnecessary {
					t.Errorf("Expected Unnecessary tag, got %v", diagnostic.Tags[0])
				}
			}
		})
	}
}

func TestMapSeverity(t *testing.T) {
	tests := []struct {
		name             string
		severity         dwscript.ErrorSeverity
		expectedSeverity protocol.DiagnosticSeverity
	}{
		{
			name:             "error severity",
			severity:         dwscript.SeverityError,
			expectedSeverity: protocol.DiagnosticSeverityError,
		},
		{
			name:             "warning severity",
			severity:         dwscript.SeverityWarning,
			expectedSeverity: protocol.DiagnosticSeverityWarning,
		},
		{
			name:             "info severity",
			severity:         dwscript.SeverityInfo,
			expectedSeverity: protocol.DiagnosticSeverityInformation,
		},
		{
			name:             "hint severity",
			severity:         dwscript.SeverityHint,
			expectedSeverity: protocol.DiagnosticSeverityHint,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapSeverity(tt.severity)
			if result != tt.expectedSeverity {
				t.Errorf("Expected severity %v, got %v", tt.expectedSeverity, result)
			}
		})
	}
}

func TestMapDiagnosticTags(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectedTag *protocol.DiagnosticTag
	}{
		{
			name:        "unused variable",
			code:        "W_UNUSED_VAR",
			expectedTag: func() *protocol.DiagnosticTag { t := protocol.DiagnosticTagUnnecessary; return &t }(),
		},
		{
			name:        "unused parameter",
			code:        "W_UNUSED_PARAM",
			expectedTag: func() *protocol.DiagnosticTag { t := protocol.DiagnosticTagUnnecessary; return &t }(),
		},
		{
			name:        "deprecated",
			code:        "W_DEPRECATED",
			expectedTag: func() *protocol.DiagnosticTag { t := protocol.DiagnosticTagDeprecated; return &t }(),
		},
		{
			name:        "no tag for other codes",
			code:        "E_UNDEFINED_VAR",
			expectedTag: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := mapDiagnosticTags(tt.code)
			if tt.expectedTag == nil {
				if len(tags) != 0 {
					t.Errorf("Expected no tags, got %v", tags)
				}
			} else {
				if len(tags) == 0 {
					t.Error("Expected a tag, got none")
				} else if tags[0] != *tt.expectedTag {
					t.Errorf("Expected tag %v, got %v", *tt.expectedTag, tags[0])
				}
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
			_, diagnostics, err := ParseDocument(tt.source, "test.dws")
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
