// Package terrain provides procedural terrain and dungeon generation algorithms.
//
// This package implements multiple generation strategies:
//   - BSP (Binary Space Partitioning) for structured dungeon layouts
//   - Cellular Automata for organic cave-like structures
//
// All generators are deterministic based on seed values and follow the
// Generator interface from the parent procgen package.
package terrain
