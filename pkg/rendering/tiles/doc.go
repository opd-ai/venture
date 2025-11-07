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
//   - Phase 16.2: Smooth terrain transitions with auto-tiling (Marching Squares)
//   - Edge blending between different tile types
//   - Corner rounding for organic feel
//   - Edge smoothing for visual polish
//
// Phase 16.2 Transition System:
//
// The transition system implements Marching Squares algorithm for seamless
// tile connections. It analyzes 8-directional neighbors and generates
// 47 unique tile variants including:
//   - Single edge connections (4 types)
//   - Adjacent corners (4 types)
//   - Opposite corridors (2 types)
//   - T-junctions (4 types)
//   - Inner corners (4 types)
//   - Full connections and isolated tiles
//
// The system provides gradient blending at tile boundaries, corner rounding
// for walls, and edge smoothing for organic appearance. All transitions are
// deterministic and maintain performance targets (<3% frame time increase).
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
//
// Example with Transitions:
//
//	neighbors := tiles.TileNeighbors{N: true, E: true, S: true}
//	transitionType := tiles.DetermineTransition(neighbors)
//
//	transConfig := tiles.TransitionConfig{
//	    BaseConfig:   config,
//	    Transition:   transitionType,
//	    Neighbors:    neighbors,
//	    BlendRadius:  0.3,
//	    CornerRadius: 0.25,
//	    Smoothness:   0.5,
//	}
//	transitionTile, err := gen.GenerateWithTransition(transConfig)
//	if err != nil {
//	    log.Fatal(err)
//	}
package tiles
