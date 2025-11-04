package engine

import (
	"strings"
	"testing"

	"golang.org/x/image/font/basicfont"
)

// TestWrapText_EmptyString tests wrapping empty string.
func TestWrapText_EmptyString(t *testing.T) {
	result := WrapText("", 100, basicfont.Face7x13)
	if len(result) != 0 {
		t.Errorf("WrapText(\"\") = %v, want empty slice", result)
	}
}

// TestWrapText_SingleWord tests wrapping single word that fits.
func TestWrapText_SingleWord(t *testing.T) {
	result := WrapText("Hello", 100, basicfont.Face7x13)
	if len(result) != 1 {
		t.Errorf("WrapText(\"Hello\") returned %d lines, want 1", len(result))
	}
	if len(result) > 0 && result[0] != "Hello" {
		t.Errorf("WrapText(\"Hello\") = %q, want \"Hello\"", result[0])
	}
}

// TestWrapText_MultipleWords tests wrapping multiple words that fit on one line.
func TestWrapText_MultipleWords(t *testing.T) {
	result := WrapText("Hello World", 200, basicfont.Face7x13)
	if len(result) != 1 {
		t.Errorf("WrapText(\"Hello World\") returned %d lines, want 1", len(result))
	}
	if len(result) > 0 && result[0] != "Hello World" {
		t.Errorf("WrapText(\"Hello World\") = %q, want \"Hello World\"", result[0])
	}
}

// TestWrapText_WordWrap tests word wrapping at word boundaries.
func TestWrapText_WordWrap(t *testing.T) {
	// "Hello World Test" should wrap to multiple lines at small width
	result := WrapText("Hello World Test", 50, basicfont.Face7x13)

	// Should have at least 2 lines for narrow width
	if len(result) < 2 {
		t.Errorf("WrapText with narrow width returned %d lines, want >= 2", len(result))
	}

	// Verify no line is empty
	for i, line := range result {
		if line == "" {
			t.Errorf("Line %d is empty, want non-empty", i)
		}
	}
}

// TestWrapText_LongWord tests hyphenation of very long words.
func TestWrapText_LongWord(t *testing.T) {
	// Create a very long word that must be hyphenated
	longWord := "Supercalifragilisticexpialidocious"
	result := WrapText(longWord, 70, basicfont.Face7x13)

	// Should be broken into multiple lines
	if len(result) < 2 {
		t.Errorf("WrapText with very long word returned %d lines, want >= 2", len(result))
	}

	// All lines except the last should end with hyphen
	for i := 0; i < len(result)-1; i++ {
		if !strings.HasSuffix(result[i], "-") {
			t.Errorf("Line %d = %q, want to end with hyphen", i, result[i])
		}
	}

	// Last line should NOT end with hyphen
	if len(result) > 0 {
		lastLine := result[len(result)-1]
		if strings.HasSuffix(lastLine, "-") {
			t.Errorf("Last line %q ends with hyphen, want no hyphen", lastLine)
		}
	}
}

// TestWrapText_InvalidWidth tests behavior with invalid width.
func TestWrapText_InvalidWidth(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		text     string
		wantLen  int
		wantText string
	}{
		{"zero width", 0, "Hello", 1, "Hello"},
		{"negative width", -10, "World", 1, "World"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapText(tt.text, tt.width, basicfont.Face7x13)
			if len(result) != tt.wantLen {
				t.Errorf("WrapText returned %d lines, want %d", len(result), tt.wantLen)
			}
			if len(result) > 0 && result[0] != tt.wantText {
				t.Errorf("WrapText = %q, want %q", result[0], tt.wantText)
			}
		})
	}
}

// TestWrapText_ProcedurallyGeneratedNames tests wrapping of realistic procedural names.
func TestWrapText_ProcedurallyGeneratedNames(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxWidth int
		wantMin  int // minimum number of lines expected
	}{
		{
			name:     "fantasy weapon",
			text:     "Legendary Enchanted Sword of the Ancient Dragon",
			maxWidth: 150,
			wantMin:  2,
		},
		{
			name:     "sci-fi weapon",
			text:     "Advanced Cybernetic Plasma Rifle Mark VII",
			maxWidth: 150,
			wantMin:  2,
		},
		{
			name:     "horror item",
			text:     "Cursed Amulet of the Forgotten Elder God",
			maxWidth: 150,
			wantMin:  2,
		},
		{
			name:     "quest description",
			text:     "Find the lost artifact in the depths of the ancient temple ruins",
			maxWidth: 200,
			wantMin:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapText(tt.text, tt.maxWidth, basicfont.Face7x13)

			if len(result) < tt.wantMin {
				t.Errorf("WrapText returned %d lines, want >= %d", len(result), tt.wantMin)
			}

			// Verify all text is preserved
			joined := strings.Join(result, " ")
			// Remove hyphens that were added for wrapping
			joined = strings.ReplaceAll(joined, "- ", "")

			// Count words (approximate check)
			originalWords := strings.Fields(tt.text)
			wrappedWords := strings.Fields(joined)

			if len(wrappedWords) < len(originalWords) {
				t.Errorf("Some words were lost: original %d words, wrapped %d words",
					len(originalWords), len(wrappedWords))
			}
		})
	}
}

// TestMeasureText tests text width measurement.
func TestMeasureText(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantMin int
		wantMax int
	}{
		{"empty", "", 0, 0},
		{"single char", "A", 5, 10},
		{"short word", "Hello", 30, 50},
		{"long word", "Supercalifragilistic", 100, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width := MeasureText(tt.text, basicfont.Face7x13)

			if width < tt.wantMin || width > tt.wantMax {
				t.Errorf("MeasureText(%q) = %d, want range [%d, %d]",
					tt.text, width, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// TestMeasureText_Consistency tests that measured width is consistent.
func TestMeasureText_Consistency(t *testing.T) {
	text := "Hello World"

	// Measure same text multiple times
	width1 := MeasureText(text, basicfont.Face7x13)
	width2 := MeasureText(text, basicfont.Face7x13)
	width3 := MeasureText(text, basicfont.Face7x13)

	if width1 != width2 || width2 != width3 {
		t.Errorf("MeasureText inconsistent: %d, %d, %d", width1, width2, width3)
	}
}

// TestWrapTextWithHeight tests wrapping with height calculation.
func TestWrapTextWithHeight(t *testing.T) {
	text := "This is a longer text that should wrap to multiple lines"
	maxWidth := 100
	lineSpacing := 2

	lines, height := WrapTextWithHeight(text, maxWidth, basicfont.Face7x13, lineSpacing)

	// Should have multiple lines
	if len(lines) < 2 {
		t.Errorf("WrapTextWithHeight returned %d lines, want >= 2", len(lines))
	}

	// Height should be positive
	if height <= 0 {
		t.Errorf("WrapTextWithHeight height = %d, want > 0", height)
	}

	// Height should increase with more lines
	// Expected: lineHeight * numLines + lineSpacing * (numLines - 1)
	metrics := basicfont.Face7x13.Metrics()
	lineHeight := int(metrics.Height >> 6) // Convert from fixed.Int26_6
	expectedMin := lineHeight * len(lines)

	if height < expectedMin {
		t.Errorf("WrapTextWithHeight height = %d, want >= %d", height, expectedMin)
	}
}

// TestWrapTextWithHeight_EmptyString tests height calculation for empty string.
func TestWrapTextWithHeight_EmptyString(t *testing.T) {
	lines, height := WrapTextWithHeight("", 100, basicfont.Face7x13, 2)

	if len(lines) != 0 {
		t.Errorf("WrapTextWithHeight(\"\") returned %d lines, want 0", len(lines))
	}

	if height != 0 {
		t.Errorf("WrapTextWithHeight(\"\") height = %d, want 0", height)
	}
}

// TestHyphenateWord tests word hyphenation directly.
func TestHyphenateWord(t *testing.T) {
	tests := []struct {
		name     string
		word     string
		maxWidth int
		wantMin  int // minimum segments expected
	}{
		{"short word", "Hello", 100, 1},
		{"long word narrow", "Supercalifragilistic", 50, 2},
		{"very long word", "Pneumonoultramicroscopicsilicovolcanoconiosis", 70, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hyphenateWord(tt.word, tt.maxWidth, basicfont.Face7x13)

			if len(result) < tt.wantMin {
				t.Errorf("hyphenateWord returned %d segments, want >= %d",
					len(result), tt.wantMin)
			}

			// All segments except last should end with hyphen
			for i := 0; i < len(result)-1; i++ {
				if !strings.HasSuffix(result[i], "-") {
					t.Errorf("Segment %d = %q, want to end with hyphen", i, result[i])
				}
			}
		})
	}
}

// BenchmarkWrapText benchmarks text wrapping performance.
func BenchmarkWrapText(b *testing.B) {
	text := "This is a long procedurally generated item name that needs wrapping"
	maxWidth := 150

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = WrapText(text, maxWidth, basicfont.Face7x13)
	}
}

// BenchmarkMeasureText benchmarks text measurement performance.
func BenchmarkMeasureText(b *testing.B) {
	text := "Legendary Enchanted Sword of the Ancient Dragon"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MeasureText(text, basicfont.Face7x13)
	}
}
