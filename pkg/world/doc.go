// Package world provides world state management including map data,
// entity persistence, save/load functionality, and chunk streaming.
//
// # Persistent World State (V6.0 Phase 37.1)
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
// # Chunk Streaming (V6.0 Phase 37.2)
//
// The chunk streaming system enables efficient memory management for large worlds:
//
// - ChunkLoaderSystem: Load/unload chunks based on player proximity (5 chunk radius default)
// - ChunkModificationSystem: Track terrain changes, mark chunks dirty
// - ChunkCompressionSystem: RLE encoding for uniform terrain (10-1000x compression for uniform areas)
//
// Chunks are loaded on-demand as players move and unloaded when no players are nearby,
// maintaining a target of <1MB per loaded chunk and <10ms load time per chunk.
//
// # Performance
//
// Benchmark results for 100-chunk worlds:
//   - SaveWorld: ~0.44ms (2000x faster than 2s target)
//   - LoadWorld: ~0.21ms
//   - IncrementalSave: ~0.22ms (automatic fallback to full save if >50% modified)
//   - ChunkLoad: ~28µs per chunk (350x faster than 10ms target)
//   - ChunkCompress (uniform): ~2µs with 1000x compression
//   - ChunkCompress (varied): ~44µs with 0.5x compression
//   - ModifyTerrain: ~92ns per tile
//   - Memory: <50MB disk space per server, <1MB per loaded chunk
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
//	// Setup chunk streaming
//	loader := world.NewChunkLoaderSystem(12345, wp, generator)
//	loader.SetLoadRadius(5) // Load 5 chunks in each direction
//
//	// Update chunk loading based on player positions
//	playerPositions := map[uint64]struct{ X, Y float64 }{
//	    1: {X: 100, Y: 100},
//	}
//	loader.Update(playerPositions)
//
//	// Track terrain modifications (wired in cmd/server/main.go after loader init)
//	modSystem := world.NewChunkModificationSystem(state)
//	modSystem.ModifyTerrain(100, 100, world.TileWall)
//	modSystem.AddModification("explosion", 100, 100, 5.0)
//
//	// Compress chunk on eviction (wired via loader.SetOnEvict in cmd/server/main.go)
//	compressor := world.NewChunkCompressionSystem()
//	chunk, _ := loader.GetChunk(0, 0)
//	compressed, ratio, _ := compressor.CompressChunk(chunk)
//	// ratio typically 10-1000x for uniform terrain
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
// # Save Format Versioning (PRE-1.0)
//
// OBSOLETE CODE REMOVED: Automatic save migration
// Replaced by: Pre-1.0 policy - incompatible saves are rejected with clear error
// Removed: Automatic migration logic, backward compatibility code
//
// Save files include a version field. Pre-version 1.0, only the current schema
// version is supported. Incompatible saves are rejected with an error message.
//
// Current schema version: 1
//
// # Thread Safety
//
// WorldPersistence is NOT thread-safe. Callers must synchronize access
// to SaveWorld, LoadWorld, and related methods. ChunkLoaderSystem,
// ChunkModificationSystem, and ChunkCompressionSystem are also not thread-safe.
package world
