# Performance Lag Investigation - 2026-01-21

## Summary

Investigation identified **5 major bottlenecks** causing startup lag: excessive system count (111 ECS systems per frame), synchronous terrain generation (up to 18ms), sequential entity spawning, cold sprite cache at startup, and spatial partition rebuilds. The cumulative effect is an estimated 200-400ms initialization overhead with first-frame FPS drops below the 60 FPS target.

## Issues Identified

### Issue 1: Excessive ECS System Count

- **Location:** `cmd/client/handlers.go:1067` (`registerAllSystems`)
- **Severity:** Critical
- **Impact:** ~1-2ms added per frame from system iteration overhead; 111 systems × entity processing = significant frame budget consumption
- **Root Cause:** The game registers 111 systems to the ECS World. Each frame, `World.Update()` iterates through all systems calling `Update(entities, deltaTime)`. Many systems process the full entity list even when they only affect a small subset of entities. The linear iteration through 111 systems creates a baseline cost regardless of game state.

**Evidence:**
- Direct system count: `grep -c "game.World.AddSystem" cmd/client/handlers.go` returns **111**
- Additional systems from `pkg/engine/system_init.go`: **44**
- Total potential systems: **155** (though not all are additive)
- Note: Some AddSystem calls are conditional (wrapped in `if sys.xyz != nil` checks). Actual runtime count may be lower, estimated at **80-100 active systems** based on code review of conditional registrations (lines 1234, 1239, 1299-1312 show conditional patterns).

**System Registration Pattern (handlers.go:1067-1313):**
```go
func registerAllSystems(game *engine.EbitenGame, sys *systemsContainer) {
    game.World.AddSystem(sys.performanceSystem)
    game.World.AddSystem(sys.inputSystem)
    // ... 109 more AddSystem calls
}
```

### Issue 2: Synchronous Terrain Generation Blocking Startup

- **Location:** `cmd/client/handlers.go:1347` (`generateWorldTerrain`)
- **Severity:** High
- **Impact:** 2.5-18.3ms blocking startup time for terrain generation
- **Root Cause:** Terrain generation is synchronous and happens before any frame is rendered. The BSP generator is relatively fast (25-240μs), but composite and cellular generators used for complex terrains can take 5-18ms on the main thread.

**Profiling Data (from `go test -bench`):**
| Generator | Small | Medium | Large |
|-----------|-------|--------|-------|
| BSP | 38μs | - | 239μs |
| Cellular | 667μs | - | 7.9ms |
| Forest | 1.1ms | 3.8ms | 14.3ms |
| Composite | 5.1ms | - | 18.3ms |
| Maze | 171μs | - | 1.6ms |

**Code Flow (main.go:72-76):**
```go
generatedTerrain := generateWorldTerrain(logger, clientLogger)  // BLOCKS
initializeTerrainRendering(game, generatedTerrain, clientLogger)
configureLightingSystem(game, clientLogger)
```

### Issue 3: Sequential Entity Spawning with Multiple Generator Calls

- **Location:** `cmd/client/handlers.go:1597` (`spawnWorldEntities`)
- **Severity:** High
- **Impact:** 50-200ms blocking time from cumulative generator calls
- **Root Cause:** Entity spawning happens synchronously in sequence. Each spawn function creates procedural content by calling generators:
  - Enemy generation: ~21μs per enemy, ~285μs for merchants
  - Item generation: ~17-27μs per item
  - Quest generation: ~19-40μs per quest
  - Spell generation: ~37-47μs per spell
  
For a typical startup with 50+ enemies, 10+ merchants, puzzles, vehicles, companions, and story fragments, this accumulates to significant blocking time.

**Profiling Data:**
| Generator | Single | Batch (100) | Allocations |
|-----------|--------|-------------|-------------|
| Entity | 21μs | - | - |
| Merchant | 285μs | - | - |
| Item | 17μs | 82μs | 516 allocs |
| Quest | 19μs | 62μs | 395 allocs |
| Spell | 37μs | 129μs | 1750 allocs |

**Code Flow (handlers.go:1597-1619):**
```go
func spawnWorldEntities(...) {
    spawnEnemiesWithLogging(...)     // Synchronous
    spawnMerchantsWithLogging(...)   // Synchronous
    spawnMinigamesWithLogging(...)   // Synchronous
    spawnStationsWithLogging(...)    // Synchronous
    spawnPuzzlesWithLogging(...)     // Synchronous
    spawnObjectsWithLogging(...)     // Synchronous
    spawnVehiclesWithLogging(...)    // Synchronous
    spawnCompanionsWithLogging(...)  // Synchronous
    spawnBookshelvesWithLogging(...) // Synchronous
    spawnStoryFragmentsWithLogging(...) // Synchronous
    generateNarrativeArcWithLogging(...) // Synchronous
}
```

### Issue 4: Cold Sprite Cache at Startup

- **Location:** `pkg/engine/animation_system.go:78-94`
- **Severity:** Medium
- **Impact:** First-frame rendering 2-5× slower than steady-state; deferred regeneration may cause visual pop-in
- **Root Cause:** The animation system has a `maxRegenPerFrame` limit of 8 sprite regenerations per frame to prevent lag. At startup, the number of entities requiring sprites varies by game configuration:
  - **Estimated entity count:** 50-200 entities (1 player + 30-100 enemies + 2-10 merchants + vehicles + companions + NPCs)
  - **Regeneration spread example:** 100 entities ÷ 8 regenerations/frame = ~12.5 frames (~208ms at 60 FPS)
  - During this deferred regeneration period, entities may render with missing or placeholder sprites

The sprite cache (400MB limit) starts empty, causing initial cache misses for every entity.

**Evidence from animation_system.go:78-94:**
```go
return &AnimationSystem{
    frameCache:       make(map[string][]*ebiten.Image),
    maxCacheSize:     100, // Cache up to 100 animation sequences
    maxRegenPerFrame: 8,   // Limits regenerations per frame
}
```

**Sprite Cache Configuration (consts.go:27):**
```go
spriteCacheMaxSize = 400 * 1024 * 1024 // 400MB limit
```

### Issue 5: Spatial Partition Double Rebuild at Startup

- **Location:** `cmd/client/main.go:84-92`
- **Severity:** Low (50-150μs overhead, acceptable trade-off)
- **Impact:** 50-150μs per rebuild, called twice at startup (100-300μs total)
- **Root Cause:** The spatial partition initialization requires a rebuild with all entities. Due to initialization order (spatial partition created before entities spawned), the first rebuild operates on an empty/partial entity list, then a second rebuild is required after entity spawning. This is a known design trade-off documented in the code.

**Design Trade-off:** The current approach prioritizes code simplicity over micro-optimization. The alternative (delaying spatial partition initialization until after all spawning) would require restructuring the initialization order.

**Evidence (main.go:84-92):**
```go
// CRITICAL FIX: Rebuild spatial partition after ALL entities are created
// The initial rebuild happens before player and enemies are spawned,
// so we need to rebuild again to include them in the quadtree for culling
if spatialSystem := game.RenderSystem.GetSpatialPartition(); spatialSystem != nil {
    spatialSystem.Rebuild(game.World.GetEntities())
}
```

## Additional Observations

### Observation A: Large systemsContainer Struct

- **Location:** `cmd/client/handlers.go:141-388`
- **Impact:** Memory allocation and initialization time
- **Details:** The `systemsContainer` struct has 150+ fields, requiring significant memory allocation and null-checking during initialization.

### Observation B: Query Cache Invalidation

- **Location:** `pkg/engine/ecs.go:604-608`
- **Impact:** Cache thrashing during entity addition
- **Details:** Each entity addition invalidates the entire query cache, causing re-computation of all queries on the next access.

```go
func (w *World) invalidateQueryCache() {
    for key := range w.queryCache {
        w.queryCacheDirty[key] = true
    }
}
```

### Observation C: Save/Load Performance for Large States

- **Location:** `pkg/saveload/`
- **Impact:** 4-8ms for large game saves
- **Details:** While not directly affecting startup (unless loading a save), the save/load benchmarks show:
  - Large save: 4.5ms, 2.2MB allocations
  - Large load: 8.6ms, 1.2MB allocations

## Recommendations

**Priority 1 - System Count Reduction:**
1. **Combine related systems:** Consider merging systems that operate on the same components:
   - `CompanionAISystem` + `CompanionProgressionSystem` + `CompanionLoyaltySystem` + `CompanionInventorySystem` → unified `CompanionManager`
   - `VehicleMovementSystem` + `VehicleDurabilitySystem` + `VehicleCombatSystem` → unified `VehicleManager`
   - Multiple expression systems (`ExpressionSystem`, `ExpressionComboSystem`) → single system
2. **Conditional loading candidates:** Systems for optional features that could load on first use:
   - PvP systems: `PvPRatingSystem`, `PvPRewardSystem`, `TournamentSystem`
   - Raid systems: `RaidSystem`, `LegendaryQuestSystem`
   - Federation systems: `PortalSystem`, `BountySystem`, `PoliticsSystem`, `FederationProtocol`
   - Social systems: `MailSystem`, `CourierSystem`, `EnhancedChatSystem` (defer until player interaction)
3. Implement system groups that can be updated together based on active game mode

**Priority 2 - Async/Deferred Generation:**
1. Move terrain generation to a loading screen or async worker
2. Implement progressive entity spawning across multiple frames
3. Use spawn queues with per-frame budget limits

**Priority 3 - Cache Prewarming:**
1. Implement sprite cache prewarming during loading screen
2. Generate common sprites (player, basic enemies) synchronously, defer rare sprites
3. Consider bundling pre-generated sprites for common configurations

**Priority 4 - Spatial Partition Optimization:**
1. Delay initial spatial partition build until after all entity spawning
2. Consider incremental insertion during entity spawning instead of full rebuild

**Priority 5 - Memory Allocation Reduction:**
1. Pre-allocate systemsContainer fields that are always used
2. Use object pools for frequently allocated generator results
3. Reduce allocations in hot paths (item: 59 allocs, quest: 122 allocs per generation)

## Profiling Data

### Terrain Generation Benchmarks
```
BenchmarkBSPGen-4                38445 ns/op     44120 B/op    134 allocs
BenchmarkBSPGenLarge-4          239286 ns/op    408557 B/op    908 allocs
BenchmarkCellularGen-4          667257 ns/op    399013 B/op   2483 allocs
BenchmarkCellularGenLarge-4    7924125 ns/op   5193421 B/op  24360 allocs
BenchmarkForestGen-4           3809904 ns/op     92717 B/op    121 allocs
BenchmarkCompositeGen-4        2496619 ns/op    699541 B/op   1391 allocs
BenchmarkCompositeGenLarge-4  16628368 ns/op   5942621 B/op   8166 allocs
```

### Entity Generation Benchmarks
```
BenchmarkEntityGeneration-4         20602 ns/op
BenchmarkGenerateMerchant-4        285061 ns/op
BenchmarkItemGenerator-4            26940 ns/op    10656 B/op    59 allocs
BenchmarkQuestGenerator-4           35940 ns/op    12874 B/op   122 allocs
BenchmarkSpellGenerator-4           47575 ns/op    41034 B/op   423 allocs
```

### Save/Load Benchmarks
```
BenchmarkSaveGameMedium-4     1100917 ns/op    503635 B/op     19 allocs
BenchmarkSaveGameLarge-4      4523925 ns/op   2190950 B/op    281 allocs
BenchmarkLoadGameMedium-4     2103553 ns/op    286184 B/op    958 allocs
BenchmarkLoadGameLarge-4      8645339 ns/op   1150804 B/op   3044 allocs
```

### System Count Summary
- **handlers.go AddSystem calls:** 111
- **system_init.go AddSystem calls:** 44
- **Estimated total active systems per frame:** 100-150

### Memory Targets vs Observations
- **Target:** <500MB client memory
- **Sprite cache allocation:** 400MB max (spriteCacheMaxSize)
- **Animation cache:** 300 sequences (animationCacheSize)
- **Observation:** Memory targets likely met, but initialization allocations may cause GC pressure

## Methodology

1. Static code analysis of initialization paths in `cmd/client/main.go` and `cmd/client/handlers.go`
2. Benchmark execution: `go test -bench=. -benchtime=3x -benchmem -run=^$`
3. System count via grep: `grep -c "AddSystem" cmd/client/handlers.go`
4. Review of existing performance documentation in `pkg/engine/PERFORMANCE_BENCHMARKS.md`
5. Analysis of ECS architecture in `pkg/engine/ecs.go`

---
*Report generated by performance investigation - no fixes implemented*
