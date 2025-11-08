// Package environment provides procedural generation of environmental objects.
// This package generates furniture, decorations, obstacles, and hazards for
// dungeon and world environments. All objects are generated procedurally with
// collision detection, interaction properties, and genre-specific styling.
//
// Object types include:
//   - Furniture (tables, chairs, beds, shelves, chests)
//   - Decorations (plants, statues, paintings, banners)
//   - Obstacles (barrels, crates, rubble, pillars)
//   - Hazards (spikes, fire pits, acid pools, bear traps)
//
// Phase 20.1 enhancements:
//   - 10 new decoration types (sconces, wall cracks, bloodstains, grass, mushrooms,
//     skulls, chains, spider webs, moss, graffiti)
//   - Visual variation system with rotation, scale, color shift, and flipping
//   - Smart room placement with natural decoration density (5-10 items per room)
//   - Biome-specific decoration pools for genre consistency
//
// Usage example:
//
//	// Generate a single decoration
//	gen := environment.NewGenerator()
//	config := environment.Config{
//	    SubType: environment.SubTypeSkull,
//	    Width:   64,
//	    Height:  64,
//	    GenreID: "horror",
//	    Seed:    12345,
//	}
//	obj, err := gen.Generate(config)
//
//	// Place decorations in a room
//	placer := environment.NewPlacer()
//	placementConfig := environment.PlacementConfig{
//	    RoomWidth:  10,
//	    RoomHeight: 10,
//	    Density:    0.5,
//	    GenreID:    "fantasy",
//	    Seed:       99999,
//	}
//	placements, err := placer.PlaceDecorations(placementConfig)
//
// All generation is deterministic based on seed values, ensuring reproducible
// content across different game sessions and clients.
package environment
