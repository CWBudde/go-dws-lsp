package document

import (
	"testing"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

const (
	testVarDeclaration       = "var x: Integer;"
	testMultiVarDeclarations = "var x: Integer;\nvar y: String;\nvar z: Float;"
)

func TestApplyContentChange_FullSync(t *testing.T) {
	originalText := "var x: Integer;\nvar y: String;"
	newText := "var z: Float;"

	change := protocol.TextDocumentContentChangeEvent{
		Range: nil, // Full sync
		Text:  newText,
	}

	result, err := ApplyContentChange(originalText, change)
	if err != nil {
		t.Fatalf("ApplyContentChange returned error: %v", err)
	}

	if result != newText {
		t.Errorf("Result = %q, want %q", result, newText)
	}
}

func TestApplyContentChange_SingleLineReplacement(t *testing.T) {
	originalText := testVarDeclaration

	// Replace "Integer" (positions 7-14) with "String"
	change := protocol.TextDocumentContentChangeEvent{
		Range: &protocol.Range{
			Start: protocol.Position{Line: 0, Character: 7},
			End:   protocol.Position{Line: 0, Character: 14},
		},
		Text: "String",
	}

	result, err := ApplyContentChange(originalText, change)
	if err != nil {
		t.Fatalf("ApplyContentChange returned error: %v", err)
	}

	expected := "var x: String;"
	if result != expected {
		t.Errorf("Result = %q, want %q", result, expected)
	}
}

func TestApplyContentChange_MultiLineReplacement(t *testing.T) {
	originalText := testMultiVarDeclarations

	// Delete the entire second line (including newline)
	change := protocol.TextDocumentContentChangeEvent{
		Range: &protocol.Range{
			Start: protocol.Position{Line: 1, Character: 0},
			End:   protocol.Position{Line: 2, Character: 0},
		},
		Text: "",
	}

	result, err := ApplyContentChange(originalText, change)
	if err != nil {
		t.Fatalf("ApplyContentChange returned error: %v", err)
	}

	expected := "var x: Integer;\nvar z: Float;"
	if result != expected {
		t.Errorf("Result = %q, want %q", result, expected)
	}
}

func TestApplyContentChange_Insertion(t *testing.T) {
	originalText := "var x: Integer;\nPrintLn(x);"

	// Insert a new line at the end of the first line (position 15 is at the end)
	change := protocol.TextDocumentContentChangeEvent{
		Range: &protocol.Range{
			Start: protocol.Position{Line: 0, Character: 15},
			End:   protocol.Position{Line: 0, Character: 15},
		},
		Text: "\nx := 42;",
	}

	result, err := ApplyContentChange(originalText, change)
	if err != nil {
		t.Fatalf("ApplyContentChange returned error: %v", err)
	}

	expected := "var x: Integer;\nx := 42;\nPrintLn(x);"
	if result != expected {
		t.Errorf("Result = %q, want %q", result, expected)
	}
}

func TestApplyContentChange_InsertionAtStartOfLine(t *testing.T) {
	originalText := "var x: Integer;\nPrintLn(x);"

	// Insert at the start of the second line
	change := protocol.TextDocumentContentChangeEvent{
		Range: &protocol.Range{
			Start: protocol.Position{Line: 1, Character: 0},
			End:   protocol.Position{Line: 1, Character: 0},
		},
		Text: "x := 42;\n",
	}

	result, err := ApplyContentChange(originalText, change)
	if err != nil {
		t.Fatalf("ApplyContentChange returned error: %v", err)
	}

	expected := "var x: Integer;\nx := 42;\nPrintLn(x);"
	if result != expected {
		t.Errorf("Result = %q, want %q", result, expected)
	}
}

func TestApplyContentChange_DeletionWithinLine(t *testing.T) {
	originalText := testVarDeclaration

	// Delete ": Integer" (positions 5-14)
	change := protocol.TextDocumentContentChangeEvent{
		Range: &protocol.Range{
			Start: protocol.Position{Line: 0, Character: 5},
			End:   protocol.Position{Line: 0, Character: 14},
		},
		Text: "",
	}

	result, err := ApplyContentChange(originalText, change)
	if err != nil {
		t.Fatalf("ApplyContentChange returned error: %v", err)
	}

	expected := "var x;"
	if result != expected {
		t.Errorf("Result = %q, want %q", result, expected)
	}
}

func TestApplyContentChange_UTF16Handling(t *testing.T) {
	// Test with characters outside the BMP (emoji that takes 2 UTF-16 code units)
	originalText := "Hello 😀 World"

	// Replace the emoji (which is at UTF-16 position 6-8) with a simple smiley
	// In UTF-16: "Hello " = 6 code units, "😀" = 2 code units (surrogate pair)
	change := protocol.TextDocumentContentChangeEvent{
		Range: &protocol.Range{
			Start: protocol.Position{Line: 0, Character: 6},
			End:   protocol.Position{Line: 0, Character: 8},
		},
		Text: "🙂",
	}

	result, err := ApplyContentChange(originalText, change)
	if err != nil {
		t.Fatalf("ApplyContentChange returned error: %v", err)
	}

	expected := "Hello 🙂 World"
	if result != expected {
		t.Errorf("Result = %q, want %q", result, expected)
	}
}

func TestApplyContentChange_InvalidRange_StartLineOutOfBounds(t *testing.T) {
	originalText := testVarDeclaration

	change := protocol.TextDocumentContentChangeEvent{
		Range: &protocol.Range{
			Start: protocol.Position{Line: 5, Character: 0},
			End:   protocol.Position{Line: 5, Character: 5},
		},
		Text: "test",
	}

	_, err := ApplyContentChange(originalText, change)
	if err == nil {
		t.Error("ApplyContentChange should return error for out-of-bounds start line")
	}
}

func TestApplyContentChange_InvalidRange_EndLineOutOfBounds(t *testing.T) {
	originalText := testVarDeclaration

	change := protocol.TextDocumentContentChangeEvent{
		Range: &protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
			End:   protocol.Position{Line: 5, Character: 0},
		},
		Text: "test",
	}

	_, err := ApplyContentChange(originalText, change)
	if err == nil {
		t.Error("ApplyContentChange should return error for out-of-bounds end line")
	}
}

func TestUTF16CharOffsetToByteOffset_ASCII(t *testing.T) {
	line := "Hello World"
	tests := []struct {
		utf16Offset int
		wantByte    int
	}{
		{0, 0},
		{5, 5},
		{11, 11},
	}

	for _, tt := range tests {
		got, err := utf16CharOffsetToByteOffset(line, tt.utf16Offset)
		if err != nil {
			t.Errorf("utf16CharOffsetToByteOffset(%q, %d) returned error: %v", line, tt.utf16Offset, err)
			continue
		}

		if got != tt.wantByte {
			t.Errorf("utf16CharOffsetToByteOffset(%q, %d) = %d, want %d", line, tt.utf16Offset, got, tt.wantByte)
		}
	}
}

func TestUTF16CharOffsetToByteOffset_Emoji(t *testing.T) {
	// "Hello 😀" - emoji is 4 bytes in UTF-8, 2 code units in UTF-16
	line := "Hello 😀"

	tests := []struct {
		utf16Offset int
		wantByte    int
		description string
	}{
		{0, 0, "start of string"},
		{6, 6, "before emoji (after space)"},
		{8, 10, "after emoji"},
	}

	for _, tt := range tests {
		got, err := utf16CharOffsetToByteOffset(line, tt.utf16Offset)
		if err != nil {
			t.Errorf("utf16CharOffsetToByteOffset(%q, %d) [%s] returned error: %v",
				line, tt.utf16Offset, tt.description, err)

			continue
		}

		if got != tt.wantByte {
			t.Errorf("utf16CharOffsetToByteOffset(%q, %d) [%s] = %d, want %d",
				line, tt.utf16Offset, tt.description, got, tt.wantByte)
		}
	}
}

func TestUTF16CharOffsetToByteOffset_MultiByteChars(t *testing.T) {
	// Test with various multi-byte UTF-8 characters
	line := "Héllo Wörld" // é and ö are 2 bytes each in UTF-8, 1 code unit in UTF-16

	tests := []struct {
		utf16Offset int
		wantByte    int
	}{
		{0, 0},  // H
		{1, 1},  // é (starts at byte 1, is 2 bytes)
		{2, 3},  // l (after é)
		{6, 7},  // W
		{7, 8},  // ö (starts at byte 8, is 2 bytes)
		{8, 10}, // r (after ö)
		{9, 11}, // l
	}

	for _, tt := range tests {
		got, err := utf16CharOffsetToByteOffset(line, tt.utf16Offset)
		if err != nil {
			t.Errorf("utf16CharOffsetToByteOffset(%q, %d) returned error: %v", line, tt.utf16Offset, err)
			continue
		}

		if got != tt.wantByte {
			t.Errorf("utf16CharOffsetToByteOffset(%q, %d) = %d, want %d", line, tt.utf16Offset, got, tt.wantByte)
		}
	}
}

func TestPositionToOffset(t *testing.T) {
	text := testMultiVarDeclarations

	tests := []struct {
		line       int
		character  int
		wantOffset int
	}{
		{0, 0, 0},   // Start of first line
		{0, 4, 4},   // "var "
		{1, 0, 16},  // Start of second line
		{1, 4, 20},  // "var " on second line
		{2, 0, 31},  // Start of third line
		{2, 10, 41}, // "var z: Flo"
	}

	for _, tt := range tests {
		got, err := PositionToOffset(text, tt.line, tt.character)
		if err != nil {
			t.Errorf("PositionToOffset(line=%d, char=%d) returned error: %v", tt.line, tt.character, err)
			continue
		}

		if got != tt.wantOffset {
			t.Errorf("PositionToOffset(line=%d, char=%d) = %d, want %d", tt.line, tt.character, got, tt.wantOffset)
		}
	}
}

func TestOffsetToPosition(t *testing.T) {
	text := testMultiVarDeclarations

	tests := []struct {
		offset   int
		wantLine int
		wantChar int
	}{
		{0, 0, 0},   // Start of first line
		{4, 0, 4},   // "var "
		{16, 1, 0},  // Start of second line
		{20, 1, 4},  // "var " on second line
		{31, 2, 0},  // Start of third line
		{41, 2, 10}, // "var z: Flo"
	}

	for _, tt := range tests {
		gotLine, gotChar, err := OffsetToPosition(text, tt.offset)
		if err != nil {
			t.Errorf("OffsetToPosition(offset=%d) returned error: %v", tt.offset, err)
			continue
		}

		if gotLine != tt.wantLine || gotChar != tt.wantChar {
			t.Errorf("OffsetToPosition(offset=%d) = (line=%d, char=%d), want (line=%d, char=%d)",
				tt.offset, gotLine, gotChar, tt.wantLine, tt.wantChar)
		}
	}
}

func TestByteOffsetToUTF16Offset_ASCII(t *testing.T) {
	line := "Hello World"

	tests := []struct {
		byteOffset int
		wantUTF16  int
	}{
		{0, 0},
		{5, 5},
		{11, 11},
	}

	for _, tt := range tests {
		got, err := byteOffsetToUTF16Offset(line, tt.byteOffset)
		if err != nil {
			t.Errorf("byteOffsetToUTF16Offset(%q, %d) returned error: %v", line, tt.byteOffset, err)
			continue
		}

		if got != tt.wantUTF16 {
			t.Errorf("byteOffsetToUTF16Offset(%q, %d) = %d, want %d", line, tt.byteOffset, got, tt.wantUTF16)
		}
	}
}

func TestByteOffsetToUTF16Offset_Emoji(t *testing.T) {
	// "Hello 😀" - emoji is 4 bytes in UTF-8, 2 code units in UTF-16
	line := "Hello 😀"

	tests := []struct {
		byteOffset  int
		wantUTF16   int
		description string
	}{
		{0, 0, "start"},
		{6, 6, "before emoji"},
		{10, 8, "after emoji"},
	}

	for _, tt := range tests {
		got, err := byteOffsetToUTF16Offset(line, tt.byteOffset)
		if err != nil {
			t.Errorf("byteOffsetToUTF16Offset(%q, %d) [%s] returned error: %v",
				line, tt.byteOffset, tt.description, err)

			continue
		}

		if got != tt.wantUTF16 {
			t.Errorf("byteOffsetToUTF16Offset(%q, %d) [%s] = %d, want %d",
				line, tt.byteOffset, tt.description, got, tt.wantUTF16)
		}
	}
}

func TestRoundTripConversion(t *testing.T) {
	// Test that converting UTF-16 -> byte -> UTF-16 yields the same result
	testCases := []string{
		"Hello World",
		"Héllo Wörld",
		"Hello 😀 World",
		"🎉🎊🎈",
		"var x: Integer;",
	}

	for _, line := range testCases {
		// Test various UTF-16 offsets
		for utf16Offset := 0; utf16Offset <= len(line); utf16Offset += 2 {
			// Convert UTF-16 -> byte
			byteOffset, err := utf16CharOffsetToByteOffset(line, utf16Offset)
			if err != nil {
				// If this offset is invalid, skip it
				continue
			}

			// Convert byte -> UTF-16
			gotUTF16, err := byteOffsetToUTF16Offset(line, byteOffset)
			if err != nil {
				t.Errorf("Round trip failed for %q at UTF-16 offset %d: byteOffsetToUTF16Offset returned error: %v",
					line, utf16Offset, err)

				continue
			}

			if gotUTF16 != utf16Offset {
				t.Errorf("Round trip failed for %q: UTF-16 %d -> byte %d -> UTF-16 %d",
					line, utf16Offset, byteOffset, gotUTF16)
			}
		}
	}
}
