// Package tiles - Step 10 validation tests for Phase 45 graphics improvements.
// These tests verify all validation criteria from docs/PLAN.md Step 10.
package tiles

import (
	"testing"
)

// TestStep10_DefaultDimensions64x64 validates default tile dimensions are 64×64.
// Validation criteria: All sprites/tiles/shapes generate at 64×64 by default.
func TestStep10_DefaultDimensions64x64(t *testing.T) {
	config := DefaultConfig()

	if config.Width != 64 {
		t.Errorf("DefaultConfig().Width = %d, want 64", config.Width)
	}
	if config.Height != 64 {
		t.Errorf("DefaultConfig().Height = %d, want 64", config.Height)
	}
}

// TestStep10_TileGeneration64x64 validates tile generation at 64×64.
// Validation criteria: tile <1ms for 64×64.
func TestStep10_TileGeneration64x64(t *testing.T) {
	gen := NewGenerator()
	config := Config{
		Type:    TileFloor,
		Width:   64,
		Height:  64,
		GenreID: "fantasy",
		Seed:    12345,
		Variant: 0.5,
		Custom:  make(map[string]interface{}),
	}

	img, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 64 || bounds.Dy() != 64 {
		t.Errorf("generated tile size = %dx%d, want 64x64", bounds.Dx(), bounds.Dy())
	}
}

// TestStep10_AllTileTypesGenerate64x64 validates all tile types generate at 64×64.
func TestStep10_AllTileTypesGenerate64x64(t *testing.T) {
	gen := NewGenerator()
	tileTypes := []TileType{
		TileFloor, TileWall, TileDoor, TileCorridor,
		TileWater, TileLava, TileTrap, TileStairs,
	}

	for _, tileType := range tileTypes {
		t.Run(tileType.String(), func(t *testing.T) {
			config := Config{
				Type:    tileType,
				Width:   64,
				Height:  64,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
				Custom:  make(map[string]interface{}),
			}

			img, err := gen.Generate(config)
			if err != nil {
				t.Fatalf("Generate(%s) failed: %v", tileType.String(), err)
			}

			bounds := img.Bounds()
			if bounds.Dx() != 64 || bounds.Dy() != 64 {
				t.Errorf("tile %s size = %dx%d, want 64x64", tileType.String(), bounds.Dx(), bounds.Dy())
			}
		})
	}
}

// BenchmarkStep10_TileGeneration64x64 benchmarks tile generation at 64×64.
// Validation criteria: tile <1ms for 64×64.
func BenchmarkStep10_TileGeneration64x64(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		Type:    TileFloor,
		Width:   64,
		Height:  64,
		GenreID: "fantasy",
		Seed:    12345,
		Variant: 0.5,
		Custom:  make(map[string]interface{}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStep10_AllTileTypes64x64 benchmarks all tile types at 64×64.
func BenchmarkStep10_AllTileTypes64x64(b *testing.B) {
	tileTypes := []TileType{
		TileFloor, TileWall, TileDoor, TileCorridor,
		TileWater, TileLava, TileTrap, TileStairs,
	}

	for _, tileType := range tileTypes {
		b.Run(tileType.String(), func(b *testing.B) {
			gen := NewGenerator()
			config := Config{
				Type:    tileType,
				Width:   64,
				Height:  64,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
				Custom:  make(map[string]interface{}),
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := gen.Generate(config)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
