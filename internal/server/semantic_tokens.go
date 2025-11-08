// Package server provides semantic tokens support for DWScript.
package server

import (
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// SemanticToken represents a raw semantic token with position and classification.
type SemanticToken struct {
	Line      uint32 // 0-based line number
	StartChar uint32 // 0-based start character
	Length    uint32 // Token length
	TokenType uint32 // Index into legend.TokenTypes
	Modifiers uint32 // Bit flags for modifiers
}

// SemanticTokensLegend defines the token types and modifiers used by the server.
// The legend must remain consistent across all requests to ensure proper highlighting.
type SemanticTokensLegend struct {
	// TokenTypes is an ordered array of token type strings.
	// The index in this array is used to encode token types in the semantic tokens response.
	TokenTypes []string

	// TokenModifiers is an ordered array of token modifier strings.
	// Modifiers are encoded as bit flags where each index represents a bit position.
	TokenModifiers []string
}

// NewSemanticTokensLegend creates a new SemanticTokensLegend with standard DWScript token types and modifiers.
func NewSemanticTokensLegend() *SemanticTokensLegend {
	return &SemanticTokensLegend{
		TokenTypes: []string{
			// Index 0: namespace - for unit names in uses clauses
			"namespace",
			// Index 1: type - for type names (classes, records, aliases)
			"type",
			// Index 2: class - for class definitions and references
			"class",
			// Index 3: enum - for enumeration types
			"enum",
			// Index 4: interface - for interface types
			"interface",
			// Index 5: struct - for record types
			"struct",
			// Index 6: typeParameter - for generic type parameters (future)
			"typeParameter",
			// Index 7: parameter - for function/method parameters
			"parameter",
			// Index 8: variable - for local and global variables
			"variable",
			// Index 9: property - for class properties
			"property",
			// Index 10: enumMember - for enum members
			"enumMember",
			// Index 11: function - for function declarations and calls
			"function",
			// Index 12: method - for class method declarations and calls
			"method",
			// Index 13: keyword - for language keywords (var, begin, end, etc.)
			"keyword",
			// Index 14: string - for string literals
			"string",
			// Index 15: number - for numeric literals
			"number",
			// Index 16: comment - for comments
			"comment",
		},
		TokenModifiers: []string{
			// Bit 0: declaration - marks symbol definitions
			"declaration",
			// Bit 1: readonly - for constants and readonly properties
			"readonly",
			// Bit 2: static - for static/class methods and fields
			"static",
			// Bit 3: deprecated - for deprecated symbols
			"deprecated",
			// Bit 4: abstract - for abstract classes/methods
			"abstract",
			// Bit 5: modification - for assignments and modifications
			"modification",
			// Bit 6: documentation - for doc comments
			"documentation",
		},
	}
}

// ToProtocolLegend converts the legend to the LSP protocol format.
func (l *SemanticTokensLegend) ToProtocolLegend() protocol.SemanticTokensLegend {
	return protocol.SemanticTokensLegend{
		TokenTypes:     l.TokenTypes,
		TokenModifiers: l.TokenModifiers,
	}
}

// GetTokenTypeIndex returns the index of a token type in the legend.
// Returns -1 if the token type is not found.
func (l *SemanticTokensLegend) GetTokenTypeIndex(tokenType string) int {
	for i, t := range l.TokenTypes {
		if t == tokenType {
			return i
		}
	}
	return -1
}

// GetModifierMask returns the bit mask for the given modifiers.
// Multiple modifiers can be combined using bitwise OR.
func (l *SemanticTokensLegend) GetModifierMask(modifiers ...string) uint32 {
	var mask uint32
	for _, modifier := range modifiers {
		for i, m := range l.TokenModifiers {
			if m == modifier {
				mask |= 1 << uint32(i)
				break
			}
		}
	}
	return mask
}

// Token type constants for easier reference
const (
	TokenTypeNamespace     = "namespace"
	TokenTypeType          = "type"
	TokenTypeClass         = "class"
	TokenTypeEnum          = "enum"
	TokenTypeInterface     = "interface"
	TokenTypeStruct        = "struct"
	TokenTypeTypeParameter = "typeParameter"
	TokenTypeParameter     = "parameter"
	TokenTypeVariable      = "variable"
	TokenTypeProperty      = "property"
	TokenTypeEnumMember    = "enumMember"
	TokenTypeFunction      = "function"
	TokenTypeMethod        = "method"
	TokenTypeKeyword       = "keyword"
	TokenTypeString        = "string"
	TokenTypeNumber        = "number"
	TokenTypeComment       = "comment"
)

// Token modifier constants for easier reference
const (
	TokenModifierDeclaration   = "declaration"
	TokenModifierReadonly      = "readonly"
	TokenModifierStatic        = "static"
	TokenModifierDeprecated    = "deprecated"
	TokenModifierAbstract      = "abstract"
	TokenModifierModification  = "modification"
	TokenModifierDocumentation = "documentation"
)

// SemanticToken represents a raw semantic token with position and classification.
type SemanticToken struct {
	Line      uint32 // 0-based line number
	StartChar uint32 // 0-based start character
	Length    uint32 // Token length
	TokenType uint32 // Index into legend.TokenTypes
	Modifiers uint32 // Bit flags for modifiers
}
