# Engine World Systems Sub-Audit

**Date**: 2026-02-16
**Scope**: Weather, World Events, City Evolution, Economy, Terrain Modification, Commerce, World Persistence systems
**Files Audited**:
- `weather_system.go` — Weather particle management and transitions
- `world_events_system.go` — Dynamic world event generation and application
- `city_evolution_system.go` — City state changes based on triggers
- `economy_system.go` — Marketplace and guild bank integration
- `terrain_modification_system.go` — Destructible terrain and tile damage
- `commerce_system.go` — Buy/sell transactions between players and merchants
- `world_persistence_system.go` — World state save/load and time progression

## Issues Found and Fixed

### Medium Severity

#### W1: Dead code in `checkGuildWarfareTriggers` (world_events_system.go)
- **Problem**: Loop iterated guild entities with `_ = entity` but never incremented `guildWarCount`, making the `if guildWarCount > 0` condition permanently false. Guild warfare events could never trigger naturally.
- **Fix**: Replaced dead loop with `len(s.world.GetEntitiesWith("guild"))` count check. Triggers when >1 guild exists (multiple guilds suggest warfare potential).
- **Impact**: Guild warfare events now trigger correctly during world event checks.

#### W2: Non-deterministic timing in `EconomySystem` (economy_system.go)
- **Problem**: Used `time.Now()` and `time.Duration` for marketplace/interest interval tracking. Inconsistent with other world systems that use deltaTime accumulation. In multiplayer, wall-clock differences between server and client could cause timing drift.
- **Fix**: Converted to deltaTime-based accumulator pattern (`updateAccumulator`/`interestAccumulator` float64 fields) matching `CityEvolutionSystem` and `WorldEventsSystem` patterns.
- **Impact**: Economy system timing is now deterministic and consistent with other world systems.

### Low Severity

#### W3: Timer accumulator drift in `CityEvolutionSystem` (city_evolution_system.go)
- **Problem**: `s.timeAccumulator = 0.0` on interval trigger discards any accumulated excess time. Over many frames, this causes subtle timing drift (updates occur slightly less frequently than intended).
- **Fix**: Changed to `s.timeAccumulator -= s.updateInterval` to preserve remainder.
- **Impact**: City evolution timing is now precise over long play sessions.

#### W4: Timer accumulator drift in `WorldEventsSystem` (world_events_system.go)
- **Problem**: Same as W3 — `s.updateTimer = 0` discards excess time.
- **Fix**: Changed to `s.updateTimer -= s.updateInterval` to preserve remainder.
- **Impact**: World event check timing is now precise over long play sessions.

## Issues Documented (Not Fixed)

### Low Severity

#### W5: `findDestructibleEntityAt` iterates all entities (terrain_modification_system.go)
- **Description**: Iterates all entities via `s.world.GetEntities()` to find a destructible entity at specific coordinates. Could use spatial partitioning for O(1) lookup.
- **Status**: Acceptable for now — terrain destruction is infrequent and entity count per destructible search is bounded.

#### W6: `time.Now()` usage in event expiration (world_events_system.go)
- **Description**: `updateActiveEvents` and `GetActiveEvents` use `time.Now()` for event lifecycle tracking. This is appropriate for real-time event expiration but noted for awareness.
- **Status**: Acceptable — event expiration is a real-time concern, not a game-time concern.

## Test Coverage

| File | Coverage | Notes |
|------|----------|-------|
| `weather_system.go` | Good | Comprehensive tests in weather_system_test.go |
| `world_events_system.go` | ~72% | Good coverage of triggers and impacts |
| `city_evolution_system.go` | ~92% | Excellent coverage with trigger tests |
| `economy_system.go` | 100% | Full coverage including trade impact |
| `terrain_modification_system.go` | ~85% | Good coverage of weapon/area damage |
| `commerce_system.go` | ~65% | Adequate — buy/sell paths well tested |
| `world_persistence_system.go` | ~88% | Good save/load and time progression coverage |

## Summary

- **Total issues found**: 6 (0 high, 2 medium, 4 low)
- **Issues fixed**: 4 (0 high, 2 medium, 2 low)
- **Issues documented**: 2 (0 high, 0 medium, 2 low)
- **Tests updated**: economy_system_test.go, world_events_system_test.go
