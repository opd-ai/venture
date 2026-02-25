// Package housing_crafting integrates V8 housing system with V4 crafting and skill systems.
//
// Phase 55.1: Crafting Stations & Skill Training
//
// This package connects player-owned houses with crafting mechanics, allowing players to:
//   - Place crafting stations in their houses for crafting bonuses
//   - Train skills at dedicated facilities for increased XP gain
//   - Unlock advanced recipes through high-quality stations
//
// # Station Types
//
// Eight crafting station types are supported:
//   - Forge: Metalworking, weapon crafting
//   - Alchemy: Potion brewing, elixir creation
//   - Enchanting: Magic item enhancement
//   - Cooking: Food preparation, buff items
//   - Tailoring: Cloth armor, garment creation
//   - Woodworking: Bows, furniture, wooden items
//   - Inscription: Scrolls, books, magical writings
//   - Engineering: Mechanical devices, gadgets
//
// # Quality Tiers
//
// Crafting stations have four quality tiers affecting bonuses:
//   - Basic (1.0x): No bonus, standard crafting
//   - Standard (1.2x): 20% faster/better crafting
//   - Advanced (1.5x): 50% improvement
//   - Master (2.0x): Double effectiveness
//
// # Skill Training
//
// When crafting in an owned house with appropriate stations, players receive:
//   - +50% skill XP compared to crafting in the field
//   - Access to station-specific advanced recipes
//   - Faster crafting times based on station quality
//
// # Integration Points
//
// This package integrates with:
//   - pkg/world/housing/: House ownership, furniture placement
//   - pkg/procgen/recipe/: Recipe generation and unlock system
//   - pkg/procgen/skills/: Skill tree progression
//   - pkg/engine/: Crafting components and systems
//
// # Example Usage
//
//	// Create station manager
//	manager := housing_crafting.NewStationManager()
//
//	// Register a master-quality forge in player's house
//	station := &housing_crafting.CraftingStation{
//	    Type:    housing_crafting.StationTypeForge,
//	    Quality: housing_crafting.QualityMaster,
//	    OwnerID: playerID,
//	    HouseID: houseID,
//	}
//	manager.RegisterStation(station)
//
//	// Get crafting bonus when player crafts
//	bonus := manager.GetCraftingBonus(playerID, "sword_recipe")
//	// bonus = 2.0 for master quality
//
//	// Get unlocked recipes
//	recipes := manager.UnlockRecipes(housing_crafting.StationTypeForge, housing_crafting.QualityMaster)
//	// Returns all forge recipes available at master tier
//
// # Performance
//
// Target: <0.1ms per crafting operation bonus calculation
// Memory: ~2KB per station, <50KB total for typical player house
//
// # Testing
//
// Target test coverage: ≥40%
// All functions are deterministic and thread-safe.
package housing_crafting
