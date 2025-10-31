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
// All generation is deterministic based on seed values, ensuring reproducible
// content across different game sessions and clients.
package environment
