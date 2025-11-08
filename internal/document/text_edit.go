// Package document provides utilities for text document manipulation.
package document

import (
	"fmt"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

// ApplyContentChange applies a TextDocumentContentChangeEvent to the given text
// and returns the updated text. This handles LSP's UTF-16 based positions.
func ApplyContentChange(text string, change protocol.TextDocumentContentChangeEvent) (string, error) {
	if change.Range == nil {
		// This shouldn't happen in incremental mode, but handle gracefully
		return change.Text, nil
	}

	// Convert the document to lines for easier manipulation
	lines := strings.Split(text, "\n")

	// Get start and end positions
	startLine := int(change.Range.Start.Line)
	startChar := int(change.Range.Start.Character)
	endLine := int(change.Range.End.Line)
	endChar := int(change.Range.End.Character)

	// Validate line numbers
	if startLine < 0 || startLine >= len(lines) {
		return "", fmt.Errorf("start line %d out of range (0-%d)", startLine, len(lines)-1)
	}

	if endLine < 0 || endLine >= len(lines) {
		return "", fmt.Errorf("end line %d out of range (0-%d)", endLine, len(lines)-1)
	}

	if startLine > endLine {
		return "", fmt.Errorf("start line %d after end line %d", startLine, endLine)
	}

	// Convert UTF-16 character positions to UTF-8 byte offsets within their lines
	startByteOffset, err := utf16CharOffsetToByteOffset(lines[startLine], startChar)
	if err != nil {
		return "", fmt.Errorf("invalid start position: %w", err)
	}

	endByteOffset, err := utf16CharOffsetToByteOffset(lines[endLine], endChar)
	if err != nil {
		return "", fmt.Errorf("invalid end position: %w", err)
	}

	// Build the new text
	var result strings.Builder

	if startLine == endLine {
		// Change is within a single line
		before := lines[startLine][:startByteOffset]
		after := lines[startLine][endByteOffset:]
		newLine := before + change.Text + after

		// Reconstruct document
		for i := range startLine {
			result.WriteString(lines[i])
			result.WriteString("\n")
		}

		result.WriteString(newLine)

		for i := startLine + 1; i < len(lines); i++ {
			result.WriteString("\n")
			result.WriteString(lines[i])
		}
	} else {
		// Change spans multiple lines
		before := lines[startLine][:startByteOffset]
		after := lines[endLine][endByteOffset:]

		// Reconstruct document
		for i := range startLine {
			result.WriteString(lines[i])
			result.WriteString("\n")
		}

		result.WriteString(before)
		result.WriteString(change.Text)
		result.WriteString(after)

		for i := endLine + 1; i < len(lines); i++ {
			result.WriteString("\n")
			result.WriteString(lines[i])
		}
	}

	return result.String(), nil
}

// utf16CharOffsetToByteOffset converts a UTF-16 character offset (as used by LSP)
// to a UTF-8 byte offset within the given line.
// LSP uses UTF-16 code units for character positions.
func utf16CharOffsetToByteOffset(line string, utf16Offset int) (int, error) {
	if utf16Offset == 0 {
		return 0, nil
	}

	// Convert the line to UTF-16 to count code units correctly
	utf16Units := utf16.Encode([]rune(line))

	// Validate offset
	if utf16Offset > len(utf16Units) {
		// Allow offset at end of line for insertions
		if utf16Offset == len(utf16Units) {
			return len(line), nil
		}

		return 0, fmt.Errorf("UTF-16 offset %d exceeds line length %d", utf16Offset, len(utf16Units))
	}

	// Count UTF-8 bytes up to the UTF-16 offset
	byteOffset := 0
	utf16Count := 0

	for _, r := range line {
		if utf16Count >= utf16Offset {
			break
		}

		// Count how many UTF-16 code units this rune takes
		// Runes in BMP (U+0000 to U+FFFF) take 1 code unit
		// Runes outside BMP take 2 code units (surrogate pair)
		if r <= 0xFFFF {
			utf16Count++
		} else {
			utf16Count += 2
		}

		byteOffset += utf8.RuneLen(r)
	}

	return byteOffset, nil
}

// PositionToOffset converts a line/character position to a byte offset in the text.
// This is a helper function for general position-based operations.
func PositionToOffset(text string, line, character int) (int, error) {
	lines := strings.Split(text, "\n")

	if line < 0 || line >= len(lines) {
		return 0, fmt.Errorf("line %d out of range (0-%d)", line, len(lines)-1)
	}

	// Calculate offset: sum of all previous lines + newlines + character offset in current line
	offset := 0
	for i := range line {
		offset += len(lines[i]) + 1 // +1 for newline
	}

	byteOffset, err := utf16CharOffsetToByteOffset(lines[line], character)
	if err != nil {
		return 0, err
	}

	return offset + byteOffset, nil
}

// OffsetToPosition converts a byte offset to a line/character position.
// Returns positions in UTF-16 code units (as expected by LSP).
func OffsetToPosition(text string, offset int) (line, character int, err error) {
	if offset < 0 || offset > len(text) {
		return 0, 0, fmt.Errorf("offset %d out of range (0-%d)", offset, len(text))
	}

	currentOffset := 0
	currentLine := 0

	lines := strings.Split(text, "\n")
	for i, lineText := range lines {
		lineLen := len(lineText)
		if currentOffset+lineLen >= offset {
			// Offset is in this line
			byteOffsetInLine := offset - currentOffset

			utf16Offset, err := byteOffsetToUTF16Offset(lineText, byteOffsetInLine)
			if err != nil {
				return 0, 0, err
			}

			return i, utf16Offset, nil
		}

		currentOffset += lineLen + 1 // +1 for newline
		currentLine++
	}

	// Offset is at the very end
	if offset == len(text) {
		return len(lines) - 1, len(utf16.Encode([]rune(lines[len(lines)-1]))), nil
	}

	return 0, 0, fmt.Errorf("offset %d not found in text", offset)
}

// byteOffsetToUTF16Offset converts a UTF-8 byte offset within a line
// to a UTF-16 code unit offset.
func byteOffsetToUTF16Offset(line string, byteOffset int) (int, error) {
	if byteOffset < 0 || byteOffset > len(line) {
		return 0, fmt.Errorf("byte offset %d out of range (0-%d)", byteOffset, len(line))
	}

	if byteOffset == 0 {
		return 0, nil
	}

	utf16Offset := 0
	currentByteOffset := 0

	for _, r := range line {
		if currentByteOffset >= byteOffset {
			break
		}

		// Count UTF-16 code units for this rune
		if r <= 0xFFFF {
			utf16Offset++
		} else {
			utf16Offset += 2
		}

		currentByteOffset += utf8.RuneLen(r)
	}

	return utf16Offset, nil
}
