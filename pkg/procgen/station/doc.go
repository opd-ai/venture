// Package station provides procedural generation of crafting stations for the Venture game.
// StationType enum maps to recipe categories: Alchemy→potions/consumables,
// Forge→enchanted equipment, Workbench→magic items, Kitchen→food/buffs,
// Anvil→weapons and armor.
//
// # Overview
//
// This package generates crafting stations (alchemy tables, forges, workbenches, kitchens, anvils)
// with genre-appropriate names and properties. Stations provide bonuses to crafting success
// and speed when used by players.
//
// # Station Types
//
// Five station types are supported:
//   - Alchemy Table: For potion brewing (+5% success, 25% faster)
//   - Forge: For enchanting equipment (+5% success, 25% faster)
//   - Workbench: For crafting magic items (+5% success, 25% faster)
//   - Kitchen: For cooking food (+5% success, 25% faster)
//   - Anvil: For smithing weapons and armor (+5% success, 25% faster)
//
// # Generation
//
// Station generation is deterministic and genre-aware:
//
//	gen := station.NewStationGenerator()
//	params := procgen.GenerationParams{
//	    Difficulty: 0.5,
//	    Depth:      1,
//	    GenreID:    "fantasy",
//	}
//
//	result, err := gen.Generate(12345, params)
//	if err != nil {
//	    // Note: Production code should use logrus.WithError(err).Fatal()
//	    return err
//	}
//
//	stations := result.([]*station.StationData)
//	// Note: Production code should use logrus.WithFields for structured logging
//	for _, s := range stations {
//	    // Example: logrus.WithFields(logrus.Fields{"type": s.StationType, "name": s.Name}).Info("Generated crafting station")
//	}
//
// # Genre Support
//
// All five genres are supported with appropriate naming themes:
//   - Fantasy: "Ancient Arcane Table", "Dwarven Flaming Forge"
//   - Sci-Fi: "Molecular Synthesis Station", "Plasma Fabrication Unit"
//   - Horror: "Cursed Necromantic Altar", "Blood Infernal Forge"
//   - Cyberpunk: "Synth Mixing Terminal", "Cyber Modification Rig"
//   - Post-Apocalyptic: "Makeshift Brewing Station", "Scrap Welding Forge"
//
// # Integration
//
// Stations are spawned in the game world using the engine's station_spawn.go helpers:
//
//	import "github.com/opd-ai/venture/pkg/engine"
//
//	count := engine.SpawnStationsInTerrain(
//	    world,
//	    stationGen,
//	    terrain,
//	    tileSize,
//	    seed,
//	    genreID,
//	)
//
// # Determinism
//
// All generation is deterministic based on seed:
//   - Same seed + parameters always produces same station names
//   - Station types are always generated in same order (alchemy, forge, workbench, kitchen, anvil)
//   - Name variations come from seed-based RNG, not system randomness
//
// This ensures multiplayer synchronization and reproducible worlds.
//
// # Performance
//
// Station generation is fast:
//   - <1ms to generate 5 stations
//   - No external I/O or heavy computation
//   - Suitable for real-time world generation
package station
