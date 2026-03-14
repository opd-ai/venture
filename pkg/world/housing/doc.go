// Package housing provides player housing functionality for Venture.
//
// The housing system allows players to place and manage buildings in the game world,
// with support for procedural generation, persistence, cross-server synchronization,
// and blueprint sharing.
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
// # Blueprints
//
// The blueprint system allows players to share building and furniture layouts:
//   - Export/import blueprints as JSON files with gzip compression
//   - Filter blueprints by author, tags, genre, rating, size
//   - Sort by rating, downloads, created/modified date, name
//   - Rate and review blueprints
//   - Track download statistics
//
// Blueprints are portable files that contain:
//   - Building parameters (type, style, dimensions, floors, seed)
//   - Furniture placements (positions, rotations, materials)
//   - Metadata (author, description, tags, genre)
//
// # Persistence
//
// Housing structures and blueprints are saved using JSON serialization with gzip compression,
// targeting <50MB per player. Incremental saves are used to minimize I/O overhead.
//
// # Usage Example
//
// Housing Management:
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
//	    logrus.WithError(err).WithField("plot_id", plot.ID).Error("Failed to place plot")
//	}
//
//	// Save housing data
//	err = manager.Save("saves/housing_player123.json.gz")
//
// Blueprint Management:
//
//	library := housing.NewBlueprintLibrary()
//
//	// Create a blueprint
//	bp := &housing.Blueprint{
//	    ID:      "bp001",
//	    Name:    "Fantasy Manor",
//	    Author:  "Player1",
//	    GenreID: "fantasy",
//	    Tags:    []string{"medieval", "manor"},
//	    BuildingDef: &housing.BuildingDefinition{
//	        Type:   building.TypeManor,
//	        Style:  building.StyleMedieval,
//	        Width:  24,
//	        Height: 24,
//	        Floors: 3,
//	        Seed:   12345,
//	    },
//	}
//
// # Architecture
//
// The housing package is composed of four primary components:
//   - Manager: Core housing lifecycle (CreateHouse, PlacePlot, furniture CRUD)
//   - BlueprintManager: Save/load/share building blueprints (compressed JSON)
//   - SpatialManager: Overlap detection and spatial queries for plot placement
//   - GuildhallManager: Multi-plot guild headquarters with permission levels
//
// Manager and GuildhallManager delegate spatial queries to SpatialManager to
// prevent overlapping placements. BlueprintManager is independent and provides
// export/import functionality for sharing designs between players.
//
//	library.Add(bp)
//
//	// Export blueprint
//	bp.Export("blueprints/fantasy_manor.json.gz")
//
//	// Import blueprint
//	imported, err := housing.ImportBlueprint("blueprints/fantasy_manor.json.gz")
//
//	// Search blueprints
//	results := library.Filter(housing.FilterOptions{
//	    Author:    "Player1",
//	    MinRating: 4.0,
//	    Tags:      []string{"medieval"},
//	})
//
//	// Sort by rating
//	library.Sort(results, housing.SortByRating, true)
package housing
