// Package entity provides procedural generation for game entities (monsters, NPCs).
//
// The entity generator creates diverse creatures and characters with varied stats,
// abilities, and behaviors. Generation is deterministic based on seed values.
//
// # Entity Generation
//
// Entities are generated with:
//   - Base stats (health, damage, defense, speed)
//   - Size and rarity classifications
//   - Descriptive name generation
//   - Level-based scaling
//
// # Usage
//
// Basic entity generation:
//
//	gen := entity.NewEntityGenerator()
//	params := procgen.GenerationParams{
//	    Difficulty: 0.5,
//	    Depth:      5,
//	    GenreID:    "fantasy",
//	}
//	result, err := gen.Generate(seed, params)
//	entities := result.([]*entity.Entity)
//
// # Entity Types
//
// The generator supports various entity classes:
//   - Monsters: Hostile creatures with combat stats
//   - NPCs: Non-hostile characters (merchants, questgivers)
//   - Bosses: Rare, powerful entities with enhanced stats
//   - Minions: Weak, common entities in groups
//
// # Deterministic Generation
//
// All entity generation is deterministic. The same seed and parameters
// will always produce the same entities. This is critical for multiplayer
// synchronization.
package entity
