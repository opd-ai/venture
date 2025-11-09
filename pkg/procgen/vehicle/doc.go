// Package vehicle provides procedural vehicle generation.
//
// This package generates vehicles with genre-specific templates,
// stat scaling based on depth/rarity, and visual variations.
//
// # Vehicle Types
//
// Five base vehicle types are supported:
//   - Mount: Living creatures (horses, dragons, alien beasts)
//   - Cart: Wheeled vehicles (wagons, trucks, hovercrafts)
//   - Boat: Water vehicles (ships, submarines, hydrofoils)
//   - Glider: Air vehicles (gliders, jetpacks, flying mounts)
//   - Mech: Mechanical suits (power armor, walkers, combat suits)
//
// # Genre Adaptation
//
// Vehicles adapt to genre themes:
//   - Fantasy: Horses, dragons, wagons, ships
//   - Sci-Fi: Hovercrafts, mechs, energy gliders, submarines
//   - Horror: Possessed carriages, bone wagons, flesh mechs
//   - Cyberpunk: Motorcycles, combat mechs, street racers
//   - Post-Apocalyptic: Scrap vehicles, wasteland bikes, makeshift armor
//
// # Usage
//
//	gen := vehicle.NewVehicleGenerator()
//	params := procgen.GenerationParams{
//	    GenreID:    "fantasy",
//	    Depth:      5,
//	    Difficulty: 0.5,
//	}
//	result, err := gen.Generate(12345, params)
//	vehicles := result.([]*vehicle.Vehicle)
//
// # Determinism
//
// All generation is deterministic - same seed with same parameters
// produces identical vehicles. This ensures multiplayer synchronization
// and reproducible content.
//
// Phase 21.1: Vehicle Foundation
package vehicle
