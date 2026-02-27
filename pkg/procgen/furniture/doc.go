/*
Package furniture provides procedural generation of furniture items for player housing and guild halls.

This package implements Phase 51.3 of ROADMAP_V8.md: Furniture Generation & Placement.

# Overview

The furniture package generates 30+ furniture types across 8 categories with material
variation, rarity tiers, and placement validation. All generation is deterministic using
seed-based algorithms, ensuring consistent results across clients in multiplayer.

# Features

- 30+ furniture types across 8 categories
- 5 material types (Wood, Metal, Stone, Crystal, Fabric)
- 5 rarity tiers affecting visual detail and functionality
- Placement validation with collision detection
- 4-way and 8-way rotation support
- Genre-specific naming and appearance

# Furniture Categories

 1. Seating: Chair, Bench, Stool, Throne, Couch
 2. Storage: Chest, Wardrobe, Shelf, Barrel, Cabinet, Crate
 3. Crafting: Anvil, Workbench, Forge, Alchemy Table, Enchanting Table
 4. Decoration: Statue, Painting, Vase, Tapestry, Plant
 5. Lighting: Torch, Chandelier, Lantern, Crystal Light
 6. Bedding: Bed, Hammock, Bedroll
 7. Tables: Table, Desk, Counter, Altar
 8. Utility: Fireplace, Mirror, Fountain, Brazier

# Generation

Generate furniture using the Generator with seed-based parameters:

	gen := furniture.NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"SubType": "Chair", // Optional, random if not specified
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		logrus.WithError(err).Fatal("failed to generate furniture")
	}

	furniture := result.(*furniture.Furniture)
	logrus.WithField("name", furniture.Name).Info("generated furniture")
	// Output: "Fine Wood Chair" or "Legendary Metal Throne of Power"

# Placement Validation

Validate furniture placement in rooms using PlacementValidator:

	validator := furniture.NewPlacementValidator(16.0, 12.0) // 16×12 room

	// Validate placement
	err := validator.ValidatePlacement(furnitureItem, 2.0, 3.0, furniture.DirNorth)
	if err != nil {
		logrus.WithError(err).Warn("invalid furniture placement")
	}

	// Place furniture
	err = validator.PlaceFurniture(furnitureItem, 2.0, 3.0, furniture.DirNorth)
	if err != nil {
		logrus.WithError(err).Warn("cannot place furniture")
	}

	// Auto-find valid placement
	x, y, dir, ok := validator.FindValidPlacement(furnitureItem, furniture.DirNorth)
	if ok {
		validator.PlaceFurniture(furnitureItem, x, y, dir)
	}

# Material Properties

Each material type has distinct visual properties and genre preferences:

  - Wood: Browns and tans, preferred in fantasy and post-apocalyptic
  - Metal: Grays and silvers, preferred in sci-fi and cyberpunk
  - Stone: Grays and browns, preferred in fantasy and horror
  - Crystal: Bright saturated colors, preferred for high-rarity items
  - Fabric: Wide color variety, used for seating and bedding

# Rarity System

Rarity affects visual detail, naming, and functionality:

  - Common (1.0x): Base appearance, simple names
  - Uncommon (1.2x): Quality prefix, slight enhancement
  - Rare (1.5x): Material prefix, improved appearance
  - Epic (2.0x): Legendary prefix + quality suffix, high detail
  - Legendary (3.0x): Mythical prefix + epic suffix, maximum detail

Higher rarity increases:
- Visual detail level (DetailMultiplier)
- Storage capacity (for storage furniture)
- Dimensional scaling (10% per tier)
- Exotic material selection probability

# Rotation

Furniture supports 4-way and 8-way rotation:

	4-way: North, East, South, West (90° increments)
	8-way: Adds NE, SE, SW, NW (45° increments)

Rotation affects collision dimensions:
- N/S: Uses width × depth
- E/W: Uses depth × width (swapped)
- Diagonal: Uses diagonal distance

# Performance

Generation targets:
- Generation time: <10ms per furniture item
- Visual variety: >95% unique items (achieved through material/rarity/color variation)
- Placement validation: <5ms per item

Determinism:
- Same seed + same parameters = identical furniture
- Suitable for multiplayer synchronization
- Reproducible for testing

# Testing

The package includes comprehensive tests covering:
- Template validation (30+ templates)
- Generation determinism
- Material and rarity distribution
- Naming conventions
- Placement collision detection
- Rotation logic
- Occupancy calculation

Run tests:

	go test ./pkg/procgen/furniture/

# Integration

This package integrates with:
- pkg/world/housing/: Building interior decoration
- pkg/network/federation/guild/: Guild hall furnishing
- pkg/procgen/: Standard generator interface
- pkg/rendering/: Visual representation (future work)

# Example: Furnishing a Room

	gen := furniture.NewGenerator()
	validator := furniture.NewPlacementValidator(16.0, 12.0)

	// Generate various furniture types
	subtypes := []string{"Bed", "Chest", "Table", "Chair", "Torch"}
	seed := int64(54321)

	for i, subtype := range subtypes {
		params := procgen.GenerationParams{
			Difficulty: 0.5,
			Depth:      5,
			GenreID:    "fantasy",
			Custom:     map[string]interface{}{"SubType": subtype},
		}

		result, _ := gen.Generate(seed+int64(i), params)
		item := result.(*furniture.Furniture)

		// Find and place
		x, y, dir, ok := validator.FindValidPlacement(item, furniture.DirNorth)
		if ok {
			validator.PlaceFurniture(item, x, y, dir)
			logrus.WithFields(logrus.Fields{
				"name": item.Name,
				"x":    x,
				"y":    y,
			}).Info("placed furniture")
		}
	}

	logrus.WithField("occupancy_percent", validator.GetOccupancyPercent()).Info("room furnishing complete")

# Future Enhancements

Planned features (Phase 51.4 and beyond):
- Visual rendering integration with pkg/rendering/
- Furniture durability and damage states
- Interactive furniture (crafting stations, storage UI)
- Furniture trading and blueprints
- Custom furniture creation via player designs
*/
package furniture
