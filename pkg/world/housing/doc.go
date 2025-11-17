// Package housing provides player housing functionality for Venture.
//
// The housing system allows players to place and manage buildings in the game world,
// with support for procedural generation, persistence, and cross-server synchronization.
//
// # Key Concepts
//
// Building Size Tiers:
//   - Small: 8×8 tiles (256 square units)
//   - Medium: 16×16 tiles (1024 square units)
//   - Large: 24×24 tiles (2304 square units)
//   - Estate: 32×32 tiles (4096 square units)
//
// # Collision Detection
//
// The plot placement system validates that new buildings:
//   - Do not overlap with existing buildings
//   - Are placed on walkable terrain
//   - Respect minimum spacing requirements (1 tile margin)
//
// # Persistence
//
// Housing structures are saved using JSON serialization with gzip compression,
// targeting <50MB per player. Incremental saves are used to minimize I/O overhead.
//
// # Usage Example
//
//	manager := housing.NewManager(world)
//
//	// Place a building
//	plot := &housing.Plot{
//	    OwnerID: "player123",
//	    Position: Vector2{X: 100, Y: 100},
//	    Size: housing.SizeMedium,
//	}
//
//	err := manager.PlacePlot(plot)
//	if err != nil {
//	    log.Printf("Failed to place plot: %v", err)
//	}
//
//	// Save housing data
//	err = manager.Save("saves/housing_player123.json.gz")
package housing
