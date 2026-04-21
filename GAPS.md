# Performance Gaps — 2026-04-21

> This file documents gaps between the project's stated performance targets and current
> measured/analyzed state. Each gap identifies a concrete bottleneck, its severity, and
> actionable remediation steps. See AUDIT.md for full findings inventory.

---

## Gap P1 — Per-Frame String Allocation in `AdvancedClassSystem`

- **Stated Goal**: Client must sustain 60 FPS with <500 MB memory; ECS `Update()` methods must complete within the 16.6 ms frame budget.
- **Current State**: `AdvancedClassSystem.Update` calls `fmt.Sprintf("%d", entity.ID)` for every entity that has an `"advanced_class"` component, every frame. It also calls `entity.GetComponent("advanced_class")` (uncached map lookup) for every entity in the world before filtering. At 50 entities and 60 FPS this produces ≥3,000 heap allocations per second from a single system.
- **Bottleneck**: `pkg/engine/advanced_class_system.go:27–32` — generic `GetComponent` + `fmt.Sprintf` inside `Update` loop.
- **Closing the Gap**:
  1. Replace `entity.GetComponent("advanced_class")` with `world.GetEntitiesWith("advanced_class")` pre-filter so only relevant entities are iterated.
  2. Replace `playerID := fmt.Sprintf("%d", entity.ID)` with `playerID := strconv.FormatUint(entity.ID, 10)` (avoids reflect machinery), or better: rekey `advanced.Manager` maps by `uint64` to eliminate the string conversion entirely.
- **Expected Improvement**: Eliminates ≥3,000 allocs/sec and ≥30 µs/frame of GC pressure from this system. Also reduces `Update` time by skipping non-advanced-class entities.

---

## Gap P2 — AI System Iterates All Entities with Uncached Component Lookup

- **Stated Goal**: Support 500+ simultaneous entities at 60 FPS; AI `Update` must remain within budget.
- **Current State**: `AISystem.Update` iterates the full entity list and calls `entity.GetComponent("ai")` (a `map[string]Component` lookup) for every entity. At 500 entities this is 30,000 map lookups/sec just for the outer filter. The `"ai"` component type is not in the 21-slot fast-path cache on `Entity`. Inside the decision path, 10 more uncached `GetComponent` calls occur per AI entity.
- **Bottleneck**: `pkg/engine/ai_system.go:82` — uncached `GetComponent("ai")` on all entities every frame; also `GetComponent("attack")`, `GetComponent("animation")`, `GetComponent("aim")` inside the hot path.
- **Closing the Gap**:
  1. Immediate: change the first line of `AISystem.Update` to `entities = s.world.GetEntitiesWith("ai")` so only AI entities are processed (the query cache returns the same slice on cache hits with 0 allocs).
  2. Medium-term: add `"ai"` to `entity.updateComponentCache` in `ecs.go` alongside the existing 21 cached types; add typed getter `entity.GetAI() *AIComponent`.
- **Expected Improvement**: For a world with 500 entities and 50 AI entities, the outer loop shrinks 10× (500 → 50 iterations). The 10× reduction in `GetComponent` calls translates to ~1 µs/frame recovered, and eliminates map-lookup cache misses from non-AI entities.

---

## Gap P3 — `MovementSystem.visitedCells` Grows Without Bound

- **Stated Goal**: Server memory target <1 GB for 8 players; long-running session reliability.
- **Current State**: `MovementSystem` maintains a `map[uint64]map[int64]bool` (`visitedCells`) that records every grid cell ever visited by every entity. Inner maps are created lazily per entity during `trackAreaVisit`. The map is never trimmed when entities die or leave. In a long session with hundreds of NPCs, `visitedCells` accumulates unbounded per-entity inner maps (one `map[int64]bool` per ever-seen entity, each growing as entities move).
- **Bottleneck**: `pkg/engine/movement.go:286-292` — lazy inner-map creation in `trackAreaVisit`, no cleanup on entity removal.
- **Closing the Gap**:
  1. Hook entity removal: call `s.ClearVisitedCells(entity.ID)` in the entity removal path (e.g. a `System.OnEntityRemoved` callback or from `RemoveEntity` post-processing).
  2. Consider replacing the map-of-maps with a flat `map[uint64]bitset` (fixed-size bitset for explored grid) to cap memory per entity.
  3. Pre-initialize the inner map with reasonable capacity (`make(map[int64]bool, 64)`) to reduce rehashing.
- **Expected Improvement**: Eliminates unbounded memory growth. For a 2-hour session with 200 NPC entities, the current implementation can accumulate 200 × (N-cells-visited × 8 bytes) in map overhead with no upper bound.

---

## Gap P4 — Query Cache Invalidation Stampede on Entity Spawn/Despawn

- **Stated Goal**: 60 FPS with NPCs that spawn and despawn (dungeon room clear, market traders, etc.).
- **Current State**: `World.invalidateQueryCache()` marks **every** cached query dirty on every entity addition or removal, regardless of which components the entity has. With 260 registered systems many of which call `GetEntitiesWith` with distinct component combinations, a single NPC death triggers O(q) dirty marks and subsequently O(q × n) entity rescans on the next frame as each system's query rebuilds.
- **Bottleneck**: `pkg/engine/ecs.go:858-862` — blanket `queryCacheDirty[key] = true` for all keys; `filterEntitiesByComponents` is O(n) per dirty query.
- **Closing the Gap**:
  1. Track which component types each cached query depends on: `queryDependencies map[string][]string` (key → component types).
  2. On `AddEntity(e)` / `RemoveEntity(id)`, compute the entity's component type set and only dirty queries whose dependency set intersects it.
  3. This reduces invalidation from O(q) to O(k × c) where k = number of distinct component types the entity has and c = average queries per component.
- **Expected Improvement**: In a scenario where a single skeleton NPC (components: `position`, `health`, `ai`, `collider`) dies, only the 4–8 queries that involve those components are dirtied instead of all 100+. Frame time after spawn/despawn events drops from O(q × n) to near-zero.

---

## Gap P5 — `GenerateGradient` Heap-Escapes One `color.RGBA` Per Pixel

- **Stated Goal**: Zero-asset rendering pipeline must be fast enough to generate gradients during level load without stutter.
- **Current State**: `GenerateGradient` in `pkg/rendering/palette/gradient.go` produces 65,538 heap allocations for a 256×256 image (4.57 ms / 524 KB). The root cause is `interpolateColors` returning `color.Color` (interface), which causes the concrete `color.RGBA` value to escape to the heap on every `img.Set(x, y, c)` call.
- **Bottleneck**: `pkg/rendering/palette/gradient.go:66,174,214` — `interpolateColors` returns `color.Color` interface; `image.RGBA.Set` accepts `color.Color`; no direct pixel-write path used.
- **Closing the Gap**:
  1. Change `interpolateColors` return type to `color.RGBA` (concrete, stack-allocatable).
  2. Replace `img.Set(x, y, c)` with direct `img.Pix` writes:
     ```go
     off := img.PixOffset(x, y)
     img.Pix[off+0] = c.R
     img.Pix[off+1] = c.G
     img.Pix[off+2] = c.B
     img.Pix[off+3] = c.A
     ```
  3. This reduces per-call allocations from 65,538 to ~2 (the image itself and its pixel buffer).
- **Expected Improvement**: 65,536 fewer heap allocations per `GenerateGradient` call; runtime drops from ~4.5 ms to an estimated ~0.5–1 ms for a 256×256 gradient. Zero GC pressure from gradient generation.

---

## Gap P6 — Terrain Cache Warm-Hit Allocates 55 Objects / 33 KB

- **Stated Goal**: Zone transitions should be near-instant with the terrain cache; restart performance target is "675× faster than uncached" (from `terrain/ASYNC_LOADING.md`).
- **Current State**: `BenchmarkTerrainCache_Get` shows **55 allocs / 33,540 B per cache hit**. If the cache returns a pointer to an already-generated `*Terrain`, a warm hit should allocate nothing beyond the pointer return. The 55-alloc, 33 KB cost suggests the read path is either (a) deserializing from gob in the memory cache path, or (b) copying large slices inside `cachedTerrain`.
- **Bottleneck**: `pkg/procgen/terrain/cache.go` — `Get` implementation; suspect LRU `updateAccessOrder` or internal copy is causing the alloc count.
- **Closing the Gap**:
  1. Run `go test -memprofile=mem.out -bench=BenchmarkTerrainCache_Get ./pkg/procgen/terrain/` and inspect with `go tool pprof`.
  2. Ensure the in-memory path returns the cached `*Terrain` pointer directly without gob round-trip.
  3. If `accessOrder []string` is being rebuilt with a new slice on every hit, replace with a doubly-linked list approach (already used in the sprite LRU cache in `pkg/rendering/sprites/cache.go`).
- **Expected Improvement**: Warm cache hits should drop to <5 allocs / <1 KB. This directly impacts zone transition latency, which affects all players including those on high-latency (Tor) connections where "instant" cache hits are especially valuable.

---

## Gap P7 — Unguarded `logrus.WithFields` Calls in Per-Frame Systems

- **Stated Goal**: Logging must not measurably impact frame time at production (non-debug) log levels.
- **Current State**: 226 non-test, non-AI, non-movement engine source locations call `logrus.WithFields` or `logrus.WithField` without a guard check. The AI system (`ai_system.go`) and movement system (`movement.go`) both cache the debug flag (`aiDebugEnabled`, `movementDebugEnabled`) and skip allocation entirely when debug is off. Other per-frame systems (e.g. `adaptive_soundtrack.go`, `achievement_notification_system.go`, `advanced_class_system.go`) call `logrus.WithFields` unconditionally, allocating a `logrus.Fields` map every frame per matching entity.
- **Bottleneck**: Scattered across `pkg/engine/*.go` — systems that lack cached debug flags.
- **Closing the Gap**:
  1. Identify all `System.Update` implementations that call `logrus.WithFields` without a guard.
  2. Add a package-level `var xyzDebugEnabled bool` and `SetXyzDebugEnabled(logger)` pattern (matching `ai_system.go`).
  3. Guard all `logrus` calls inside `Update` behind `if xyzDebugEnabled`.
- **Expected Improvement**: Each guarded `logrus.WithFields` call that is now skipped saves ~150–300 ns and 1–2 heap allocations (the `Fields` map). Across 10 affected systems at 60 FPS this recovers up to 180 µs/frame of GC pause reduction.

---

## Gap P8 — Missing Per-System Frame-Budget Warnings

- **Stated Goal**: Maintain 60 FPS; detect regressions from newly added systems before they reach players.
- **Current State**: `World.Update` records per-system timing via `PerformanceMetrics.RecordSystemTime` but there is no mechanism to warn when a system exceeds its frame budget. A slow system (e.g. one that accidentally calls a blocking I/O operation or triggers a cache stampede) silently degrades the frame rate.
- **Bottleneck**: `pkg/engine/ecs.go:700-708` — per-system timing is recorded but not acted upon.
- **Closing the Gap**:
  1. Add a `SystemBudget map[string]time.Duration` field to `World` (defaulting to 2 ms per system).
  2. After `RecordSystemTime`, compare elapsed to budget and emit a rate-limited `logrus.Warn` if exceeded: `"system %s exceeded frame budget: %v > %v"`.
  3. Optionally expose per-system timing via the `/metrics` Prometheus endpoint in `pkg/observability`.
- **Expected Improvement**: Regressions are caught in development/CI rather than at player reports. Enables data-driven optimization targeting (identify which systems are slowest before profiling).

---

## Previously Documented Gaps (Retained from Prior Audit)

> The following gaps from prior audits are non-performance gaps retained for continuity.
> They have been migrated to the issue tracker but are summarized here for reference.

### Gap R1 — Inconsistent `gzip.Writer` Close-Error Handling Across Persistence Layer

- **Stated Goal**: Continuous-uptime server with persistent world state and data integrity.
- **Current State**: Four files use `defer gzWriter.Close()` with discarded error: `pkg/world/economy/guild_bank.go:546`, `pkg/world/housing/blueprint.go:248`, `pkg/world/housing/guildhall_manager.go:235`, `pkg/integration/choice_consequences/choice_tracker.go:595`. A failed `Close()` silently corrupts the gzip stream.
- **Bottleneck**: Silent error discard on `gzip.Writer.Close()` in four persistence paths.
- **Closing the Gap**: Adopt the named-return deferred pattern from `pkg/world/housing/persistence.go` at all four sites.
- **Expected Improvement**: Eliminates silent data corruption risk for guild vault, blueprints, guildhall state, and player choice history.

### Gap Signal-1 — Signal Handler Integration Test Infrastructure Missing

- **Stated Goal**: Graceful shutdown on SIGTERM for server reliability.
- **Current State**: No integration test validates graceful shutdown under SIGTERM.
- **Closing the Gap**: Add a test that sends SIGTERM to a running server process and asserts clean exit within timeout.

### Gap Trade-1 — `TradeValidator` Lacks Per-Item Quantity Concept

- **Stated Goal**: Player trade system with quantity enforcement.
- **Current State**: `TradeValidator` exists but the trade model has no quantity field; quantity validation is a no-op.
- **Closing the Gap**: Add `Quantity int` to the trade item model; add quantity bounds checks to `TradeValidator`.

### Gap Log-1 — `pkg/memprofile/profile.go` Uses `fmt.Printf`

- **Stated Goal**: Structured logging throughout (`logrus`).
- **Current State**: `pkg/memprofile/profile.go` uses `fmt.Printf` for output.
- **Closing the Gap**: Replace with `logrus.WithFields` or document as an explicit exemption.
