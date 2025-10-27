// Package sprites provides tests for directional sprite generation (Phase 4).
package sprites

import (
	"image/color"
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

// TestGenerateDirectionalSprites tests 4-directional sprite sheet generation.
func TestGenerateDirectionalSprites(t *testing.T) {
	gen := NewGenerator()

	config := Config{
		Type:       SpriteEntity,
		Width:      28,
		Height:     28,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 0.7,
		Custom: map[string]interface{}{
			"entityType": "humanoid",
			"useAerial":  true,
		},
	}

	sprites, err := gen.GenerateDirectionalSprites(config)
	if err != nil {
		t.Fatalf("GenerateDirectionalSprites failed: %v", err)
	}

	// Verify we got 4 sprites
	if len(sprites) != 4 {
		t.Errorf("Expected 4 sprites, got %d", len(sprites))
	}

	// Verify all directions are present
	directions := []int{0, 1, 2, 3} // Up, Down, Left, Right
	for _, dir := range directions {
		sprite, exists := sprites[dir]
		if !exists {
			t.Errorf("Missing sprite for direction %d", dir)
			continue
		}
		if sprite == nil {
			t.Errorf("Sprite for direction %d is nil", dir)
			continue
		}

		// Verify sprite dimensions
		bounds := sprite.Bounds()
		if bounds.Dx() != config.Width {
			t.Errorf("Direction %d: width = %d, want %d", dir, bounds.Dx(), config.Width)
		}
		if bounds.Dy() != config.Height {
			t.Errorf("Direction %d: height = %d, want %d", dir, bounds.Dy(), config.Height)
		}
	}
}

// TestGenerateDirectionalSprites_Determinism tests deterministic generation.
func TestGenerateDirectionalSprites_Determinism(t *testing.T) {
	gen := NewGenerator()

	config := Config{
		Type:       SpriteEntity,
		Width:      28,
		Height:     28,
		Seed:       54321,
		GenreID:    "scifi",
		Complexity: 0.5,
		Custom: map[string]interface{}{
			"entityType": "humanoid",
			"useAerial":  true,
		},
	}

	// Generate twice with same seed
	sprites1, err1 := gen.GenerateDirectionalSprites(config)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	sprites2, err2 := gen.GenerateDirectionalSprites(config)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}

	// Verify both generated same number of sprites
	if len(sprites1) != len(sprites2) {
		t.Errorf("Sprite count mismatch: %d vs %d", len(sprites1), len(sprites2))
	}

	// Verify dimensions match for each direction
	for dir := 0; dir < 4; dir++ {
		bounds1 := sprites1[dir].Bounds()
		bounds2 := sprites2[dir].Bounds()

		if bounds1 != bounds2 {
			t.Errorf("Direction %d: bounds mismatch: %v vs %v", dir, bounds1, bounds2)
		}
	}
}

// TestGenerateDirectionalSprites_WithoutAerialFlag tests fallback to side-view.
func TestGenerateDirectionalSprites_WithoutAerialFlag(t *testing.T) {
	gen := NewGenerator()

	config := Config{
		Type:       SpriteEntity,
		Width:      28,
		Height:     28,
		Seed:       99999,
		GenreID:    "fantasy",
		Complexity: 0.6,
		Custom: map[string]interface{}{
			"entityType": "humanoid",
			// No useAerial flag - should use side-view templates
		},
	}

	sprites, err := gen.GenerateDirectionalSprites(config)
	if err != nil {
		t.Fatalf("GenerateDirectionalSprites failed: %v", err)
	}

	// Should still generate 4 sprites (using side-view templates)
	if len(sprites) != 4 {
		t.Errorf("Expected 4 sprites, got %d", len(sprites))
	}
}

// TestGenerateDirectionalSprites_DifferentGenres tests multiple genres.
func TestGenerateDirectionalSprites_DifferentGenres(t *testing.T) {
	gen := NewGenerator()
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			config := Config{
				Type:       SpriteEntity,
				Width:      28,
				Height:     28,
				Seed:       12345,
				GenreID:    genre,
				Complexity: 0.7,
				Custom: map[string]interface{}{
					"entityType": "humanoid",
					"useAerial":  true,
					"genre":      genre,
				},
			}

			sprites, err := gen.GenerateDirectionalSprites(config)
			if err != nil {
				t.Fatalf("GenerateDirectionalSprites failed for %s: %v", genre, err)
			}

			if len(sprites) != 4 {
				t.Errorf("%s: expected 4 sprites, got %d", genre, len(sprites))
			}
		})
	}
}

// TestGenerateDirectionalSprites_NoPalette tests palette generation.
func TestGenerateDirectionalSprites_NoPalette(t *testing.T) {
	gen := NewGenerator()

	config := Config{
		Type:       SpriteEntity,
		Width:      28,
		Height:     28,
		Seed:       11111,
		GenreID:    "fantasy",
		Complexity: 0.5,
		Palette:    nil, // No palette provided - should generate one
		Custom: map[string]interface{}{
			"entityType": "humanoid",
			"useAerial":  true,
		},
	}

	sprites, err := gen.GenerateDirectionalSprites(config)
	if err != nil {
		t.Fatalf("GenerateDirectionalSprites failed: %v", err)
	}

	if len(sprites) != 4 {
		t.Errorf("Expected 4 sprites, got %d", len(sprites))
	}
}

// TestGenerateDirectionalSprites_WithPalette tests using provided palette.
func TestGenerateDirectionalSprites_WithPalette(t *testing.T) {
	gen := NewGenerator()

	// Generate a palette
	pal, err := gen.GetPaletteGenerator().Generate("fantasy", 12345)
	if err != nil {
		t.Fatalf("Palette generation failed: %v", err)
	}

	config := Config{
		Type:       SpriteEntity,
		Width:      28,
		Height:     28,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 0.5,
		Palette:    pal, // Provide palette
		Custom: map[string]interface{}{
			"entityType": "humanoid",
			"useAerial":  true,
		},
	}

	sprites, err := gen.GenerateDirectionalSprites(config)
	if err != nil {
		t.Fatalf("GenerateDirectionalSprites failed: %v", err)
	}

	if len(sprites) != 4 {
		t.Errorf("Expected 4 sprites, got %d", len(sprites))
	}
}

// TestGenerateDirectionalSprites_InvalidConfig tests error handling.
func TestGenerateDirectionalSprites_InvalidConfig(t *testing.T) {
	gen := NewGenerator()

	// Config with missing entityType
	config := Config{
		Type:       SpriteEntity,
		Width:      28,
		Height:     28,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 0.3, // Low complexity - will use random generation
		Custom: map[string]interface{}{
			"useAerial": true,
			// No entityType - should still work but use random generation
		},
	}

	sprites, err := gen.GenerateDirectionalSprites(config)
	if err != nil {
		t.Fatalf("GenerateDirectionalSprites failed: %v", err)
	}

	// Should still generate 4 sprites (using fallback generation)
	if len(sprites) != 4 {
		t.Errorf("Expected 4 sprites, got %d", len(sprites))
	}
}

// TestGenerateEntityWithTemplate_UseAerial tests useAerial flag in template selection.
func TestGenerateEntityWithTemplate_UseAerial(t *testing.T) {
	gen := NewGenerator()

	// Create test palette
	pal := &palette.Palette{
		Primary:    testColor{100, 150, 200, 255},
		Secondary:  testColor{200, 100, 50, 255},
		Accent1:    testColor{50, 200, 100, 255},
		Background: testColor{30, 30, 40, 255},
		Colors: []color.Color{
			testColor{100, 150, 200, 255},
			testColor{200, 100, 50, 255},
			testColor{50, 200, 100, 255},
		},
	}

	tests := []struct {
		name      string
		useAerial bool
		direction string
	}{
		{"aerial up", true, "up"},
		{"aerial down", true, "down"},
		{"aerial left", true, "left"},
		{"aerial right", true, "right"},
		{"side-view down", false, "down"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Type:       SpriteEntity,
				Width:      28,
				Height:     28,
				Seed:       12345,
				GenreID:    "fantasy",
				Complexity: 0.7,
				Palette:    pal,
				Custom: map[string]interface{}{
					"entityType": "humanoid",
					"useAerial":  tt.useAerial,
					"facing":     tt.direction,
				},
			}

			sprite, err := gen.Generate(config)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			if sprite == nil {
				t.Errorf("Generated sprite is nil")
			}

			bounds := sprite.Bounds()
			if bounds.Dx() != config.Width || bounds.Dy() != config.Height {
				t.Errorf("Sprite dimensions %dx%d, want %dx%d",
					bounds.Dx(), bounds.Dy(), config.Width, config.Height)
			}
		})
	}
}

// BenchmarkGenerateDirectionalSprites benchmarks 4-sprite generation.
func BenchmarkGenerateDirectionalSprites(b *testing.B) {
	gen := NewGenerator()

	config := Config{
		Type:       SpriteEntity,
		Width:      28,
		Height:     28,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 0.7,
		Custom: map[string]interface{}{
			"entityType": "humanoid",
			"useAerial":  true,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Seed = int64(i) // Vary seed
		_, err := gen.GenerateDirectionalSprites(config)
		if err != nil {
			b.Fatalf("GenerateDirectionalSprites failed: %v", err)
		}
	}
}

// Helper type for test palette
type testColor [4]uint8

func (c testColor) RGBA() (r, g, b, a uint32) {
	r = uint32(c[0])
	r |= r << 8
	g = uint32(c[1])
	g |= g << 8
	b = uint32(c[2])
	b |= b << 8
	a = uint32(c[3])
	a |= a << 8
	return
}
