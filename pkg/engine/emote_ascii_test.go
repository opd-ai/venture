package engine

import (
	"strings"
	"testing"
)

func TestGenerateASCIIEmote(t *testing.T) {
	tests := []struct {
		name           string
		expressionType ExpressionType
		seed           int64
	}{
		{"Wave", ExpressionWave, 12345},
		{"Cheer", ExpressionCheer, 12345},
		{"Dance", ExpressionDance, 12345},
		{"Laugh", ExpressionLaugh, 12345},
		{"Cry", ExpressionCry, 12345},
		{"Sit", ExpressionSit, 12345},
		{"Point", ExpressionPoint, 12345},
		{"Salute", ExpressionSalute, 12345},
		{"Shrug", ExpressionShrug, 12345},
		{"ThumbsUp", ExpressionThumbsUp, 12345},
		{"Facepalm", ExpressionFacepalm, 12345},
		{"Sleep", ExpressionSleep, 12345},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emote := GenerateASCIIEmote(tt.expressionType, tt.seed)
			if emote == nil {
				t.Errorf("GenerateASCIIEmote() returned nil")
				return
			}
			if emote.Art == "" {
				t.Errorf("GenerateASCIIEmote() returned empty art")
			}
			if emote.Width <= 0 {
				t.Errorf("GenerateASCIIEmote() width = %d, want > 0", emote.Width)
			}
			if emote.Height <= 0 {
				t.Errorf("GenerateASCIIEmote() height = %d, want > 0", emote.Height)
			}
		})
	}
}

func TestGenerateASCIIEmote_Deterministic(t *testing.T) {
	seed := int64(54321)
	expressionType := ExpressionDance

	// Generate twice with same seed
	emote1 := GenerateASCIIEmote(expressionType, seed)
	emote2 := GenerateASCIIEmote(expressionType, seed)

	if emote1.Art != emote2.Art {
		t.Errorf("GenerateASCIIEmote() not deterministic: got different art")
	}
	if emote1.Width != emote2.Width {
		t.Errorf("GenerateASCIIEmote() not deterministic: got different width")
	}
	if emote1.Height != emote2.Height {
		t.Errorf("GenerateASCIIEmote() not deterministic: got different height")
	}
}

func TestGenerateASCIIEmote_DifferentSeeds(t *testing.T) {
	expressionType := ExpressionWave

	// Generate with different seeds (may produce same art due to limited variants)
	emote1 := GenerateASCIIEmote(expressionType, 111)
	emote2 := GenerateASCIIEmote(expressionType, 222)

	// Both should be valid
	if emote1 == nil || emote2 == nil {
		t.Errorf("GenerateASCIIEmote() returned nil")
	}
}

func TestEmoteASCII_FormatForChat_SingleLine(t *testing.T) {
	emote := &EmoteASCII{
		Art:    "XD",
		Width:  2,
		Height: 1,
	}

	formatted := emote.FormatForChat("> ")
	if formatted != "XD" {
		t.Errorf("FormatForChat() = %q, want %q", formatted, "XD")
	}
}

func TestEmoteASCII_FormatForChat_MultiLine(t *testing.T) {
	emote := &EmoteASCII{
		Art:    "  o/\n |/\n/ \\",
		Width:  3,
		Height: 3,
	}

	formatted := emote.FormatForChat("> ")
	expected := ">   o/\n>  |/\n> / \\"
	if formatted != expected {
		t.Errorf("FormatForChat() = %q, want %q", formatted, expected)
	}
}

func TestEmoteASCII_String(t *testing.T) {
	emote := &EmoteASCII{
		Art:    "test",
		Width:  4,
		Height: 1,
	}

	str := emote.String()
	if !strings.Contains(str, "Width: 4") {
		t.Errorf("String() = %q, should contain 'Width: 4'", str)
	}
	if !strings.Contains(str, "Height: 1") {
		t.Errorf("String() = %q, should contain 'Height: 1'", str)
	}
	if !strings.Contains(str, "test") {
		t.Errorf("String() = %q, should contain 'test'", str)
	}
}

func TestGenerateASCIIEmote_UnknownType(t *testing.T) {
	// Test with an invalid expression type
	emote := GenerateASCIIEmote(ExpressionType(999), 12345)
	if emote == nil {
		t.Errorf("GenerateASCIIEmote() returned nil for unknown type")
		return
	}
	// Should return default shrug emote
	if emote.Art == "" {
		t.Errorf("GenerateASCIIEmote() returned empty art for unknown type")
	}
}

// Benchmark ASCII emote generation
func BenchmarkGenerateASCIIEmote(b *testing.B) {
	seed := int64(12345)
	for i := 0; i < b.N; i++ {
		GenerateASCIIEmote(ExpressionWave, seed)
	}
}

func BenchmarkGenerateASCIIEmote_AllTypes(b *testing.B) {
	expressionTypes := []ExpressionType{
		ExpressionWave, ExpressionCheer, ExpressionDance, ExpressionLaugh,
		ExpressionCry, ExpressionSit, ExpressionPoint, ExpressionSalute,
		ExpressionShrug, ExpressionThumbsUp, ExpressionFacepalm, ExpressionSleep,
	}
	seed := int64(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, expType := range expressionTypes {
			GenerateASCIIEmote(expType, seed)
		}
	}
}
