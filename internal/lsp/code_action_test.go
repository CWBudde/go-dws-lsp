package lsp

import (
	"testing"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

// Test diagnostic pattern recognition functions
func TestIsUndeclaredIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		diagnostic protocol.Diagnostic
		want       bool
	}{
		{
			name: "undeclared identifier message",
			diagnostic: protocol.Diagnostic{
				Message: "undeclared identifier 'x'",
			},
			want: true,
		},
		{
			name: "unknown identifier message",
			diagnostic: protocol.Diagnostic{
				Message: "unknown identifier: foo",
			},
			want: true,
		},
		{
			name: "different error",
			diagnostic: protocol.Diagnostic{
				Message: "syntax error",
			},
			want: false,
		},
		{
			name: "error code E_UNDECLARED",
			diagnostic: protocol.Diagnostic{
				Message: "some message",
				Code:    &protocol.IntegerOrString{Value: "E_UNDECLARED"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUndeclaredIdentifier(tt.diagnostic)
			if got != tt.want {
				t.Errorf("isUndeclaredIdentifier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsMissingSemicolon(t *testing.T) {
	tests := []struct {
		name       string
		diagnostic protocol.Diagnostic
		want       bool
	}{
		{
			name: "missing semicolon message",
			diagnostic: protocol.Diagnostic{
				Message: "missing semicolon",
			},
			want: true,
		},
		{
			name: "expected semicolon message",
			diagnostic: protocol.Diagnostic{
				Message: "expected ';'",
			},
			want: true,
		},
		{
			name: "different error",
			diagnostic: protocol.Diagnostic{
				Message: "syntax error",
			},
			want: false,
		},
		{
			name: "error code E_MISSING_SEMICOLON",
			diagnostic: protocol.Diagnostic{
				Message: "some message",
				Code:    &protocol.IntegerOrString{Value: "E_MISSING_SEMICOLON"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isMissingSemicolon(tt.diagnostic)
			if got != tt.want {
				t.Errorf("isMissingSemicolon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUnusedVariable(t *testing.T) {
	tests := []struct {
		name       string
		diagnostic protocol.Diagnostic
		want       bool
	}{
		{
			name: "unused variable message",
			diagnostic: protocol.Diagnostic{
				Message: "unused variable 'temp'",
			},
			want: true,
		},
		{
			name: "variable not used message",
			diagnostic: protocol.Diagnostic{
				Message: "variable not used: foo",
			},
			want: true,
		},
		{
			name: "different error",
			diagnostic: protocol.Diagnostic{
				Message: "syntax error",
			},
			want: false,
		},
		{
			name: "error code W_UNUSED_VAR",
			diagnostic: protocol.Diagnostic{
				Message: "some message",
				Code:    &protocol.IntegerOrString{Value: "W_UNUSED_VAR"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUnusedVariable(tt.diagnostic)
			if got != tt.want {
				t.Errorf("isUnusedVariable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractIdentifierName(t *testing.T) {
	tests := []struct {
		name       string
		diagnostic protocol.Diagnostic
		want       string
	}{
		{
			name: "identifier in single quotes",
			diagnostic: protocol.Diagnostic{
				Message: "undeclared identifier 'myVar'",
			},
			want: "myVar",
		},
		{
			name: "identifier in double quotes",
			diagnostic: protocol.Diagnostic{
				Message: `unknown identifier "foo"`,
			},
			want: "foo",
		},
		{
			name: "identifier with colon",
			diagnostic: protocol.Diagnostic{
				Message: "identifier: bar not found",
			},
			want: "bar",
		},
		{
			name: "no identifier found",
			diagnostic: protocol.Diagnostic{
				Message: "syntax error",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractIdentifierName(tt.diagnostic)
			if got != tt.want {
				t.Errorf("extractIdentifierName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractVariableName(t *testing.T) {
	tests := []struct {
		name       string
		diagnostic protocol.Diagnostic
		want       string
	}{
		{
			name: "variable in single quotes",
			diagnostic: protocol.Diagnostic{
				Message: "unused variable 'temp'",
			},
			want: "temp",
		},
		{
			name: "variable in double quotes",
			diagnostic: protocol.Diagnostic{
				Message: `variable "myVar" not used`,
			},
			want: "myVar",
		},
		{
			name: "variable with space",
			diagnostic: protocol.Diagnostic{
				Message: "variable foo not used",
			},
			want: "foo",
		},
		{
			name: "no variable found",
			diagnostic: protocol.Diagnostic{
				Message: "syntax error",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractVariableName(tt.diagnostic)
			if got != tt.want {
				t.Errorf("extractVariableName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseUnitsFromUsesClause(t *testing.T) {
	tests := []struct {
		name      string
		usesText  string
		wantUnits []string
	}{
		{
			name:      "single unit",
			usesText:  "uses MyUnit;",
			wantUnits: []string{"MyUnit"},
		},
		{
			name:      "multiple units",
			usesText:  "uses UnitA, UnitB, UnitC;",
			wantUnits: []string{"UnitA", "UnitB", "UnitC"},
		},
		{
			name:      "units with extra spaces",
			usesText:  "uses  UnitA ,  UnitB  ;",
			wantUnits: []string{"UnitA", "UnitB"},
		},
		{
			name:      "multiline uses",
			usesText:  "uses\n  UnitA,\n  UnitB;",
			wantUnits: []string{"UnitA", "UnitB"},
		},
		{
			name:      "empty uses",
			usesText:  "uses;",
			wantUnits: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseUnitsFromUsesClause(tt.usesText)
			if len(got) != len(tt.wantUnits) {
				t.Errorf("parseUnitsFromUsesClause() returned %d units, want %d", len(got), len(tt.wantUnits))
				return
			}
			for i, unit := range got {
				if unit != tt.wantUnits[i] {
					t.Errorf("parseUnitsFromUsesClause()[%d] = %v, want %v", i, unit, tt.wantUnits[i])
				}
			}
		})
	}
}

func TestSortUnits(t *testing.T) {
	tests := []struct {
		name  string
		units []string
		want  []string
	}{
		{
			name:  "already sorted",
			units: []string{"Alpha", "Beta", "Gamma"},
			want:  []string{"Alpha", "Beta", "Gamma"},
		},
		{
			name:  "reverse order",
			units: []string{"Gamma", "Beta", "Alpha"},
			want:  []string{"Alpha", "Beta", "Gamma"},
		},
		{
			name:  "case insensitive sort",
			units: []string{"zebra", "Apple", "banana"},
			want:  []string{"Apple", "banana", "zebra"},
		},
		{
			name:  "single unit",
			units: []string{"Single"},
			want:  []string{"Single"},
		},
		{
			name:  "empty",
			units: []string{},
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy since sortUnits modifies in place
			units := make([]string, len(tt.units))
			copy(units, tt.units)

			sortUnits(units)

			if len(units) != len(tt.want) {
				t.Errorf("sortUnits() resulted in %d units, want %d", len(units), len(tt.want))
				return
			}
			for i, unit := range units {
				if unit != tt.want[i] {
					t.Errorf("sortUnits()[%d] = %v, want %v", i, unit, tt.want[i])
				}
			}
		})
	}
}

func TestFormatUsesClause(t *testing.T) {
	tests := []struct {
		name  string
		units []string
		want  string
	}{
		{
			name:  "single unit",
			units: []string{"MyUnit"},
			want:  "uses MyUnit;",
		},
		{
			name:  "multiple units",
			units: []string{"UnitA", "UnitB", "UnitC"},
			want:  "uses UnitA, UnitB, UnitC;",
		},
		{
			name:  "empty units",
			units: []string{},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatUsesClause(tt.units)
			if got != tt.want {
				t.Errorf("formatUsesClause() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUnitNameFromURI(t *testing.T) {
	tests := []struct {
		name string
		uri  string
		want string
	}{
		{
			name: "unix path",
			uri:  "file:///home/user/project/MyUnit.dws",
			want: "MyUnit",
		},
		{
			name: "windows path",
			uri:  "file:///C:/Users/user/project/MyUnit.dws",
			want: "MyUnit",
		},
		{
			name: "no file prefix",
			uri:  "/home/user/project/MyUnit.dws",
			want: "MyUnit",
		},
		{
			name: "nested directories",
			uri:  "file:///home/user/project/src/units/Helper.dws",
			want: "Helper",
		},
		{
			name: "no extension",
			uri:  "file:///home/user/project/MyUnit",
			want: "MyUnit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getUnitNameFromURI(tt.uri)
			if got != tt.want {
				t.Errorf("getUnitNameFromURI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsKeyword(t *testing.T) {
	tests := []struct {
		name string
		word string
		want bool
	}{
		{name: "begin keyword", word: "begin", want: true},
		{name: "end keyword", word: "end", want: true},
		{name: "var keyword", word: "var", want: true},
		{name: "function keyword", word: "function", want: true},
		{name: "not a keyword", word: "myVariable", want: false},
		{name: "uppercase keyword", word: "BEGIN", want: false}, // Should be lowercase
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isKeyword(tt.word)
			if got != tt.want {
				t.Errorf("isKeyword(%q) = %v, want %v", tt.word, got, tt.want)
			}
		})
	}
}

func TestIsBuiltinIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		want       bool
	}{
		{name: "Integer type", identifier: "Integer", want: true},
		{name: "String type", identifier: "String", want: true},
		{name: "WriteLn function", identifier: "WriteLn", want: true},
		{name: "Length function", identifier: "Length", want: true},
		{name: "custom identifier", identifier: "MyCustomType", want: false},
		{name: "lowercase integer", identifier: "integer", want: false}, // Case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBuiltinIdentifier(tt.identifier)
			if got != tt.want {
				t.Errorf("isBuiltinIdentifier(%q) = %v, want %v", tt.identifier, got, tt.want)
			}
		})
	}
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		str   string
		want  bool
	}{
		{
			name:  "string present",
			slice: []string{"apple", "banana", "cherry"},
			str:   "banana",
			want:  true,
		},
		{
			name:  "string not present",
			slice: []string{"apple", "banana", "cherry"},
			str:   "orange",
			want:  false,
		},
		{
			name:  "empty slice",
			slice: []string{},
			str:   "test",
			want:  false,
		},
		{
			name:  "single element match",
			slice: []string{"test"},
			str:   "test",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsString(tt.slice, tt.str)
			if got != tt.want {
				t.Errorf("containsString() = %v, want %v", got, tt.want)
			}
		})
	}
}
