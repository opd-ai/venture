// Package sprites provides procedural sprite generation.
// This file tests anti-aliasing functionality for sprites.
package sprites

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/opd-ai/venture/pkg/rendering/shapes"
)

// TestDefaultConfig_AntiAliasSet verifies default config includes anti-aliasing setting.
func TestDefaultConfig_AntiAliasSet(t *testing.T) {
	config := DefaultConfig()

	if config.AntiAlias != shapes.AntiAliasLow {
		t.Errorf("DefaultConfig AntiAlias = %v, want AntiAliasLow", config.AntiAlias)
	}
}

// TestSpriteGeneration_WithAntiAlias verifies sprites can be generated with anti-aliasing.
func TestSpriteGeneration_WithAntiAlias(t *testing.T) {
	tests := []struct {
		name      string
		antiAlias shapes.AntiAliasQuality
	}{
		{"off", shapes.AntiAliasOff},
		{"low", shapes.AntiAliasLow},
		{"medium", shapes.AntiAliasMedium},
		{"high", shapes.AntiAliasHigh},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator()
			config := Config{
				Type:       SpriteEntity,
				Width:      64,
				Height:     64,
				Seed:       12345,
				GenreID:    "fantasy",
				Complexity: 0.5,
				AntiAlias:  tt.antiAlias,
			}

			img, err := gen.Generate(config)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			if img == nil {
				t.Error("Expected non-nil image")
			}

			bounds := img.Bounds()
			if bounds.Dx() != 64 || bounds.Dy() != 64 {
				t.Errorf("Image size = %dx%d, want 64x64", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

// TestSpriteGeneration_AllTypesWithAntiAlias verifies all sprite types support anti-aliasing.
func TestSpriteGeneration_AllTypesWithAntiAlias(t *testing.T) {
	types := []SpriteType{
		SpriteEntity,
		SpriteItem,
		SpriteTile,
		SpriteParticle,
		SpriteUI,
	}

	for _, spriteType := range types {
		t.Run(spriteType.String(), func(t *testing.T) {
			gen := NewGenerator()
			config := Config{
				Type:       spriteType,
				Width:      64,
				Height:     64,
				Seed:       54321,
				GenreID:    "fantasy",
				Complexity: 0.5,
				AntiAlias:  shapes.AntiAliasMedium,
			}

			img, err := gen.Generate(config)
			if err != nil {
				t.Fatalf("Generate(%s) failed: %v", spriteType, err)
			}

			if img == nil {
				t.Errorf("Generate(%s) returned nil image", spriteType)
			}
		})
	}
}

// TestSpriteGeneration_AntiAliasDeterminism verifies anti-aliasing is deterministic.
func TestSpriteGeneration_AntiAliasDeterminism(t *testing.T) {
	gen := NewGenerator()
	config := Config{
		Type:       SpriteEntity,
		Width:      64,
		Height:     64,
		Seed:       99999,
		GenreID:    "sci-fi",
		Complexity: 0.7,
		AntiAlias:  shapes.AntiAliasHigh,
	}

	// Generate first image
	img1, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("First generation failed: %v", err)
	}

	// Generate second image with same config
	img2, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("Second generation failed: %v", err)
	}

	// Compare pixel data (deterministic check)
	bounds := img1.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c1 := img1.At(x, y)
			c2 := img2.At(x, y)

			r1, g1, b1, a1 := c1.RGBA()
			r2, g2, b2, a2 := c2.RGBA()

			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				t.Fatalf("Pixel at (%d,%d) differs: (%d,%d,%d,%d) vs (%d,%d,%d,%d)",
					x, y, r1, g1, b1, a1, r2, g2, b2, a2)
			}
		}
	}
}

// TestSpriteGeneration_ConfigPreservation verifies anti-alias setting is preserved.
func TestSpriteGeneration_ConfigPreservation(t *testing.T) {
	config := Config{
		Type:       SpriteEntity,
		Width:      32,
		Height:     32,
		Seed:       11111,
		GenreID:    "horror",
		Complexity: 0.3,
		AntiAlias:  shapes.AntiAliasMedium,
	}

	originalAA := config.AntiAlias

	gen := NewGenerator()
	_, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify config wasn't modified
	if config.AntiAlias != originalAA {
		t.Errorf("Config.AntiAlias modified during generation: was %v, now %v",
			originalAA, config.AntiAlias)
	}
}

// TestSpriteGeneration_WithPalette verifies anti-aliasing works with provided palette.
func TestSpriteGeneration_WithPalette(t *testing.T) {
	paletteGen := palette.NewGenerator()
	pal, err := paletteGen.Generate("cyberpunk", 77777)
	if err != nil {
		t.Fatalf("Palette generation failed: %v", err)
	}

	config := Config{
		Type:       SpriteEntity,
		Width:      64,
		Height:     64,
		Seed:       77777,
		Palette:    pal,
		Complexity: 0.6,
		AntiAlias:  shapes.AntiAliasLow,
	}

	gen := NewGenerator()
	img, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("Generate with palette failed: %v", err)
	}

	if img == nil {
		t.Error("Expected non-nil image with palette")
	}
}

// TestSpriteGeneration_ComplexEntityWithAntiAlias verifies complex entities support anti-aliasing.
func TestSpriteGeneration_ComplexEntityWithAntiAlias(t *testing.T) {
	gen := NewGenerator()
	config := Config{
		Type:       SpriteEntity,
		Width:      128,
		Height:     128,
		Seed:       55555,
		GenreID:    "fantasy",
		Complexity: 0.9, // High complexity
		Custom: map[string]interface{}{
			"entityType": "warrior",
			"facing":     "down",
			"hasWeapon":  true,
		},
		AntiAlias: shapes.AntiAliasHigh,
	}

	img, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("Generate complex entity failed: %v", err)
	}

	if img == nil {
		t.Error("Expected non-nil image for complex entity")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 128 || bounds.Dy() != 128 {
		t.Errorf("Image size = %dx%d, want 128x128", bounds.Dx(), bounds.Dy())
	}
}

// TestSpriteGeneration_ItemWithAntiAlias verifies item sprites support anti-aliasing.
func TestSpriteGeneration_ItemWithAntiAlias(t *testing.T) {
	gen := NewGenerator()
	config := Config{
		Type:       SpriteItem,
		Width:      48,
		Height:     48,
		Seed:       33333,
		GenreID:    "fantasy",
		Complexity: 0.5,
		Custom: map[string]interface{}{
			"itemType": "sword",
			"rarity":   "rare",
		},
		AntiAlias: shapes.AntiAliasMedium,
	}

	img, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("Generate item failed: %v", err)
	}

	if img == nil {
		t.Error("Expected non-nil item image")
	}
}

// TestSpriteGeneration_BackwardCompatibility verifies sprites work without explicit AntiAlias.
func TestSpriteGeneration_BackwardCompatibility(t *testing.T) {
	gen := NewGenerator()
	
	// Config without AntiAlias field set (relies on default from DefaultConfig)
	config := Config{
		Type:       SpriteEntity,
		Width:      64,
		Height:     64,
		Seed:       44444,
		GenreID:    "fantasy",
		Complexity: 0.5,
		// AntiAlias not set - should use default
	}

	img, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("Generate without explicit AntiAlias failed: %v", err)
	}

	if img == nil {
		t.Error("Expected non-nil image for backward compatibility")
	}
}

// BenchmarkSpriteGeneration_AntiAliasOff benchmarks sprite generation without anti-aliasing.
func BenchmarkSpriteGeneration_AntiAliasOff(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		Type:       SpriteEntity,
		Width:      64,
		Height:     64,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 0.5,
		AntiAlias:  shapes.AntiAliasOff,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(config)
	}
}

// BenchmarkSpriteGeneration_AntiAliasLow benchmarks sprite generation with low quality AA.
func BenchmarkSpriteGeneration_AntiAliasLow(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		Type:       SpriteEntity,
		Width:      64,
		Height:     64,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 0.5,
		AntiAlias:  shapes.AntiAliasLow,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(config)
	}
}

// BenchmarkSpriteGeneration_AntiAliasMedium benchmarks sprite generation with medium quality AA.
func BenchmarkSpriteGeneration_AntiAliasMedium(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		Type:       SpriteEntity,
		Width:      64,
		Height:     64,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 0.5,
		AntiAlias:  shapes.AntiAliasMedium,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(config)
	}
}

// BenchmarkSpriteGeneration_AntiAliasHigh benchmarks sprite generation with high quality AA.
func BenchmarkSpriteGeneration_AntiAliasHigh(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		Type:       SpriteEntity,
		Width:      64,
		Height:     64,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 0.5,
		AntiAlias:  shapes.AntiAliasHigh,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(config)
	}
}
