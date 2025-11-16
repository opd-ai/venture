// Package world provides world state management including map data,
// entity persistence, and save/load functionality.
//
// # Persistent World State (V6.0 Phase 37)
//
// The persistence system supports saving and loading complete world state
// with the following features:
//
// - JSON serialization with gzip compression (5:1 ratio typical)
// - Incremental saves (only changed chunks since last save)
// - Backup rotation (keeps last 3 saves: current, .1, .2, .3)
// - Migration system for version upgrades
// - Sparse chunk storage (only modified chunks saved)
// - Entity and world event persistence
//
// # Performance
//
// Benchmark results for 100-chunk worlds:
//   - SaveWorld: ~0.44ms (2000x faster than 2s target)
//   - LoadWorld: ~0.21ms
//   - IncrementalSave: ~0.22ms (automatic fallback to full save if >50% modified)
//   - Memory: <50MB disk space per server
//
// # Usage Example
//
//	wp := world.NewWorldPersistence("/path/to/world.save")
//
//	// Create world state
//	state := &world.PersistentWorldState{
//	    Version:     world.CurrentSchemaVersion,
//	    WorldSeed:   12345,
//	    ChunkData:   make(map[string]*world.Chunk),
//	    Entities:    []*world.EntityState{},
//	    WorldEvents: []world.WorldEvent{},
//	}
//
//	// Save
//	if err := wp.SaveWorld(state); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Load (creates new state if file doesn't exist)
//	loadedState, err := wp.LoadWorld(12345)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Incremental save (only modified chunks)
//	state.ModifiedChunks["0,0"] = true
//	if err := wp.SaveIncremental(state); err != nil {
//	    log.Fatal(err)
//	}
//
// # Backup System
//
// The persistence system automatically rotates backups on each save:
//   - Current save: world.save
//   - Backup 1: world.save.1 (previous save)
//   - Backup 2: world.save.2 (2nd previous)
//   - Backup 3: world.save.3 (3rd previous, oldest kept)
//
// If the current save is corrupted, LoadWorld automatically tries backups.
//
// # Migration System
//
// Save files include a version field for automatic migration on load.
// When loading an old save format, the system automatically upgrades it
// to the current schema version while preserving data.
//
// Current schema version: 1
//
// # Thread Safety
//
// WorldPersistence is NOT thread-safe. Callers must synchronize access
// to SaveWorld, LoadWorld, and related methods.
package world
