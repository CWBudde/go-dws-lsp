// Package analysis provides semantic token analysis for DWScript.
package analysis

import (
	"log"
	"sort"
	"unicode/utf16"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"

	"github.com/CWBudde/go-dws-lsp/internal/server"
)

// CollectSemanticTokens traverses the AST and collects semantic tokens.
func CollectSemanticTokens(astRoot *ast.Program, legend *server.SemanticTokensLegend) ([]server.SemanticToken, error) {
	if astRoot == nil || legend == nil {
		return nil, nil
	}

	collector := &tokenCollector{
		legend: legend,
		tokens: make([]server.SemanticToken, 0),
	}

	// Traverse the AST
	ast.Inspect(astRoot, collector.visit)

	// Sort tokens by position (line, then character)
	sort.Slice(collector.tokens, func(i, j int) bool {
		if collector.tokens[i].Line != collector.tokens[j].Line {
			return collector.tokens[i].Line < collector.tokens[j].Line
		}
		return collector.tokens[i].StartChar < collector.tokens[j].StartChar
	})

	return collector.tokens, nil
}

// tokenCollector holds state during AST traversal.
type tokenCollector struct {
	legend *server.SemanticTokensLegend
	tokens []server.SemanticToken
}

// visit is called for each AST node during traversal.
func (tc *tokenCollector) visit(node ast.Node) bool {
	if node == nil {
		return true
	}

	// Get node position
	pos := node.Pos()
	if !pos.IsValid() {
		return true // Skip nodes without valid positions
	}

	// Classify node and add tokens
	switch n := node.(type) {
	// Literals
	case *ast.StringLiteral:
		tc.addToken(pos, utf16Length(n.Value)+2, server.TokenTypeString, 0) // +2 for quotes
	case *ast.CharLiteral:
		tc.addToken(pos, utf16Length(n.Token.Literal), server.TokenTypeString, 0)
	case *ast.IntegerLiteral:
		tc.addToken(pos, utf16Length(n.Token.Literal), server.TokenTypeNumber, 0)
	case *ast.FloatLiteral:
		tc.addToken(pos, utf16Length(n.Token.Literal), server.TokenTypeNumber, 0)
	case *ast.BooleanLiteral:
		tc.addToken(pos, utf16Length(n.Token.Literal), server.TokenTypeKeyword, 0)
	case *ast.NilLiteral:
		tc.addToken(pos, 3, server.TokenTypeKeyword, 0) // "nil" - always 3 chars

	// Variable declarations with declaration modifier
	case *ast.VarDeclStatement:
		for _, name := range n.Names {
			if name != nil {
				namePos := name.Pos()
				tc.addToken(namePos, utf16Length(name.Value), server.TokenTypeVariable,
					tc.legend.GetModifierMask(server.TokenModifierDeclaration))
			}
		}

	// Constant declarations with declaration and readonly modifiers
	case *ast.ConstDecl:
		if n.Name != nil {
			namePos := n.Name.Pos()
			tc.addToken(namePos, utf16Length(n.Name.Value), server.TokenTypeVariable,
				tc.legend.GetModifierMask(server.TokenModifierDeclaration, server.TokenModifierReadonly))
		}

	// Function declarations
	case *ast.FunctionDecl:
		if n.Name != nil {
			namePos := n.Name.Pos()
			modifiers := tc.legend.GetModifierMask(server.TokenModifierDeclaration)

			// Check if it's a method (has ClassName)
			if n.ClassName != nil {
				// It's a method
				if n.IsAbstract {
					modifiers |= tc.legend.GetModifierMask(server.TokenModifierAbstract)
				}
				tc.addToken(namePos, utf16Length(n.Name.Value), server.TokenTypeMethod, modifiers)
			} else {
				// It's a function
				tc.addToken(namePos, utf16Length(n.Name.Value), server.TokenTypeFunction, modifiers)
			}
		}
		// Mark parameters with declaration modifier
		if n.Parameters != nil {
			for _, param := range n.Parameters {
				if param.Name != nil {
					paramPos := param.Name.Pos()
					tc.addToken(paramPos, utf16Length(param.Name.Value), server.TokenTypeParameter,
						tc.legend.GetModifierMask(server.TokenModifierDeclaration))
				}
			}
		}

	// Class declarations
	case *ast.ClassDecl:
		if n.Name != nil {
			namePos := n.Name.Pos()
			tc.addToken(namePos, utf16Length(n.Name.Value), server.TokenTypeClass,
				tc.legend.GetModifierMask(server.TokenModifierDeclaration))
		}

	// Interface declarations
	case *ast.InterfaceDecl:
		if n.Name != nil {
			namePos := n.Name.Pos()
			tc.addToken(namePos, utf16Length(n.Name.Value), server.TokenTypeInterface,
				tc.legend.GetModifierMask(server.TokenModifierDeclaration))
		}

	// Field declarations (class fields)
	case *ast.FieldDecl:
		if n.Name != nil {
			fieldPos := n.Name.Pos()
			modifiers := tc.legend.GetModifierMask(server.TokenModifierDeclaration)
			if n.IsClassVar {
				modifiers |= tc.legend.GetModifierMask(server.TokenModifierStatic)
			}
			tc.addToken(fieldPos, utf16Length(n.Name.Value), server.TokenTypeProperty, modifiers)
		}

	// Property declarations
	case *ast.PropertyDecl:
		if n.Name != nil {
			propPos := n.Name.Pos()
			modifiers := tc.legend.GetModifierMask(server.TokenModifierDeclaration)
			// Add readonly modifier if property has no setter (WriteSpec is nil)
			if n.WriteSpec == nil {
				modifiers |= tc.legend.GetModifierMask(server.TokenModifierReadonly)
			}
			tc.addToken(propPos, utf16Length(n.Name.Value), server.TokenTypeProperty, modifiers)
		}

	// Type declarations
	case *ast.TypeDeclaration:
		if n.Name != nil {
			typePos := n.Name.Pos()
			tc.addToken(typePos, utf16Length(n.Name.Value), server.TokenTypeType,
				tc.legend.GetModifierMask(server.TokenModifierDeclaration))
		}

	// Enum declarations
	case *ast.EnumDecl:
		if n.Name != nil {
			enumPos := n.Name.Pos()
			tc.addToken(enumPos, utf16Length(n.Name.Value), server.TokenTypeEnum,
				tc.legend.GetModifierMask(server.TokenModifierDeclaration))
		}
		// Mark enum members - Note: EnumValue.Name is a string
		for _, member := range n.Values {
			if member.Name != "" && len(member.Name) > 0 {
				// We can't get position for enum members easily as Name is just a string
				// Skip for now - would need token position from parser
			}
		}

	// Member access (e.g., obj.field)
	case *ast.MemberAccessExpression:
		if n.Member != nil {
			memberPos := n.Member.Pos()
			tc.addToken(memberPos, utf16Length(n.Member.Value), server.TokenTypeProperty, 0)
		}

	// Function calls (e.g., Foo(), not method calls)
	case *ast.CallExpression:
		// If the function is a simple identifier (not a member access), tag it as function
		if ident, ok := n.Function.(*ast.Identifier); ok && ident != nil {
			funcPos := ident.Pos()
			tc.addToken(funcPos, utf16Length(ident.Value), server.TokenTypeFunction, 0)
		}
		// If it's a member access, it will be handled by MethodCallExpression or MemberAccessExpression

	// Method calls (e.g., obj.Method())
	case *ast.MethodCallExpression:
		if n.Method != nil {
			methodPos := n.Method.Pos()
			tc.addToken(methodPos, utf16Length(n.Method.Value), server.TokenTypeMethod, 0)
		}

	// Type annotations - Note: Name is a string
	case *ast.TypeAnnotation:
		if n.Name != "" && len(n.Name) > 0 {
			// TypeAnnotation has position from Token
			tc.addToken(n.Token.Pos, utf16Length(n.Name), server.TokenTypeType, 0)
		}
	}

	return true // Continue traversal
}

// addToken adds a semantic token to the collection.
func (tc *tokenCollector) addToken(pos token.Position, length int, tokenType string, modifiers uint32) {
	if !pos.IsValid() || length <= 0 {
		return
	}

	// Convert 1-based position to 0-based
	line := uint32(pos.Line - 1)
	if line < 0 {
		line = 0
	}
	startChar := uint32(pos.Column - 1)
	if startChar < 0 {
		startChar = 0
	}

	// Get token type index
	typeIndex := tc.legend.GetTokenTypeIndex(tokenType)
	if typeIndex < 0 {
		log.Printf("Warning: unknown token type: %s\n", tokenType)
		return
	}

	tc.tokens = append(tc.tokens, server.SemanticToken{
		Line:      line,
		StartChar: startChar,
		Length:    uint32(length),
		TokenType: uint32(typeIndex),
		Modifiers: modifiers,
	})
}

// EncodeSemanticTokens encodes tokens in LSP delta format.
// The LSP protocol uses a delta encoding where each token is represented as:
// [deltaLine, deltaStartChar, length, tokenType, tokenModifiers]
func EncodeSemanticTokens(tokens []server.SemanticToken) []uint32 {
	if len(tokens) == 0 {
		return []uint32{}
	}

	encoded := make([]uint32, 0, len(tokens)*5)
	var prevLine, prevChar uint32

	for _, token := range tokens {
		deltaLine := token.Line - prevLine
		deltaChar := token.StartChar
		if deltaLine == 0 {
			deltaChar = token.StartChar - prevChar
		}

		encoded = append(encoded,
			deltaLine,
			deltaChar,
			token.Length,
			token.TokenType,
			token.Modifiers,
		)

		prevLine = token.Line
		prevChar = token.StartChar
	}

	return encoded
}

// utf16Length calculates the length of a string in UTF-16 code units.
// LSP uses UTF-16 for character positions and lengths.
func utf16Length(s string) int {
	return len(utf16.Encode([]rune(s)))
}
