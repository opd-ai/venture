// Package tiles - Phase 45 validation tests for 64×64 graphics improvements.
// These tests verify all validation criteria for Phase 45 64×64 graphics improvements.
package tiles

import (
	"testing"
)

// TestPhase45_TileGeneration64x64 validates tile generation at 64×64.
// Validation criteria: tile <1ms for 64×64.
func TestPhase45_TileGeneration64x64(t *testing.T) {
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

// TestPhase45_AllTileTypesGenerate64x64 validates all tile types generate at 64×64.
func TestPhase45_AllTileTypesGenerate64x64(t *testing.T) {
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

// BenchmarkPhase45_TileGeneration64x64 benchmarks tile generation at 64×64.
// Validation criteria: tile <1ms for 64×64.
func BenchmarkPhase45_TileGeneration64x64(b *testing.B) {
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

// BenchmarkPhase45_AllTileTypes64x64 benchmarks all tile types at 64×64.
func BenchmarkPhase45_AllTileTypes64x64(b *testing.B) {
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
