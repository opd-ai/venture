// Package tiles provides procedural tile image generation for terrain rendering.
//
// The tiles package generates visual representations of terrain tiles using
// procedural techniques. It supports multiple tile types (floor, wall, door,
// corridor) and can generate genre-specific visual styles.
//
// Features:
//   - Deterministic tile generation using seeds
//   - Genre-aware styling using color palettes
//   - Pattern variations for visual diversity
//   - Integration with terrain generation
//   - Configurable tile sizes
//
// Example Usage:
//
//	gen := tiles.NewGenerator()
//	config := tiles.Config{
//	    Type:    tiles.TileFloor,
//	    Width:   32,
//	    Height:  32,
//	    GenreID: "fantasy",
//	    Seed:    12345,
//	}
//	tileImg, err := gen.Generate(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
package tiles
