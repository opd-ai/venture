// Package terrain provides procedural terrain and dungeon generation algorithms.
//
// This package implements multiple generation strategies for diverse environments:
//   - BSP (Binary Space Partitioning) for structured dungeon layouts with rooms and corridors
//   - Cellular Automata for organic cave-like structures with smooth walls
//   - Maze for labyrinthine corridors using recursive backtracking
//   - Forest for natural outdoor areas with trees, clearings, and water features
//   - City for urban environments with buildings, streets, and plazas
//   - Composite for multi-biome levels combining 2-4 different terrain types
//   - Multi-Level for connected dungeon levels with stair systems
//
// All generators are deterministic based on seed values and follow the
// Generator interface from the parent procgen package.
//
// # Tile Types
//
// The package supports 13+ tile types including:
//   - Basic: Wall, Floor, Corridor, Door
//   - Water: Shallow (walkable but slow), Deep (impassable), Bridge
//   - Natural: Tree (impassable obstacle)
//   - Navigation: StairsUp, StairsDown, TrapDoor, SecretDoor
//   - Urban: Structure (buildings/ruins)
//
// # Usage Examples
//
// Basic dungeon generation:
//
//	gen := terrain.NewBSPGenerator()
//	params := procgen.GenerationParams{
//	    Difficulty: 0.5,
//	    Depth:      1,
//	    GenreID:    "fantasy",
//	    Custom: map[string]interface{}{
//	        "width":  100,
//	        "height": 80,
//	    },
//	}
//	result, err := gen.Generate(12345, params)
//	terrain := result.(*terrain.Terrain)
//
// Composite multi-biome generation:
//
//	gen := terrain.NewCompositeGenerator()
//	params := procgen.GenerationParams{
//	    Difficulty: 0.5,
//	    Depth:      5,
//	    GenreID:    "fantasy",
//	    Custom: map[string]interface{}{
//	        "width":      100,
//	        "height":     80,
//	        "biomeCount": 3,  // Combine 3 different terrain types
//	    },
//	}
//	result, err := gen.Generate(12345, params)
//	terrain := result.(*terrain.Terrain)
//	// Result contains 3 biome regions (e.g., dungeon + cave + forest)
//	// with smooth transition zones and guaranteed connectivity
//
// Multi-level dungeon:
//
//	gen := terrain.NewLevelGenerator()
//	params := procgen.GenerationParams{
//	    Difficulty: 0.5,
//	    Depth:      1,
//	    GenreID:    "fantasy",
//	    Custom: map[string]interface{}{
//	        "width":  100,
//	        "height": 80,
//	    },
//	}
//	levels, err := gen.GenerateMultiLevel(5, 12345, params)
//	// Returns 5 connected levels with stairs
//
// # Performance Targets
//
// All generators meet performance requirements:
//   - 100x100: <150ms (composite <300ms)
//   - 200x200: <600ms (composite <1.2s)
//   - 500x500: <3.0s (composite <5.0s)
//
// # Validation
//
// All generators implement validation ensuring:
//   - Minimum 25-30% walkable area
//   - 90%+ connectivity (reachable via flood-fill)
//   - Proper stair placement in accessible locations
//   - Deterministic output (same seed = same terrain)
package terrain
