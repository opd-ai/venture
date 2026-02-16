# cmd/server/ Audit — 2026-02-16

## Summary

Audited all source files in `cmd/server/` (11 source files, ~2800 LOC). Found 4 issues:
2 high, 1 medium, 1 low. Fixed 3 issues (2 high, 1 medium). 1 low remaining.

**Coverage**: 65.6% of statements

## Issues Found

### Fixed

1. **[HIGH] Nil pointer dereference in `createPlayerEntity`** (`player_management.go:28`)
   - `terrain.Rooms` accessed without nil check on `terrain` parameter
   - `startPlayerManagementHandlers` passes nil terrain from `executeGameLoop`
   - **Fix**: Added `terrain != nil` guard before accessing `terrain.Rooms`

2. **[HIGH] Build error in `v9_validation_test.go`** (lines 269, 329, 467)
   - `guildMgr.CreateGuildHouse()` returns `(*GuildHouse, error)` but tests only captured 1 value
   - API signature changed but tests not updated
   - **Fix**: Updated 3 call sites to capture both return values with proper error handling

3. **[MED] Manual byte encoding in `putFloat64`** (`main.go:862-872`)
   - Manual little-endian byte manipulation instead of `encoding/binary`
   - Error-prone and harder to maintain
   - **Fix**: Replaced with `binary.LittleEndian.PutUint64(b, math.Float64bits(v))`

### Remaining

4. **[LOW] Global `v9ValidationService` synchronization** (`v9_validation.go:306`)
   - Global mutable variable set once during init, read concurrently
   - Added `sync.Once` wrapper and `SetV9ValidationService()` setter for safe initialization
   - Low risk since assignment happens once at startup before concurrent access begins

## Pre-existing Test Failures (Not Related to Audit)

- `TestWorldEventsSystemUnification` — expects `engine.WorldEventsSystem` in server systems
- `TestWorldEventsSystemServerClientParity` — server missing `engine.WorldEventsSystem`

These are pre-existing failures related to world events system unification (tracked separately).

## Files Changed

| File | Changes |
|------|---------|
| `player_management.go` | Added nil terrain guard (line 28) |
| `main.go` | Used `encoding/binary` for `putFloat64`, used `SetV9ValidationService()` |
| `v9_validation.go` | Added `sync.Once` for global service, added `SetV9ValidationService()` |
| `v9_validation_test.go` | Fixed 3 `CreateGuildHouse` call sites to handle 2 return values |
| `main_test.go` | Added 6 new tests: nil terrain, putFloat64 roundtrip, serialization, env vars |

## Test Coverage

Tests cover:
- Logger initialization (default, verbose, env override)
- World snapshot building (empty, position, velocity, multiple entities, all components)
- Network resilience initialization (all scenarios)
- Mod system initialization (valid/invalid directories)
- Stability monitoring
- System wrappers (15 wrapper types)
- Player management (entity creation, input commands, item use)
- Entity spawning (vehicles, companions, bookshelves, determinism)
- V9 validation service (crafting, loyalty, training, permissions, guild upgrades)
- Serialization (position, velocity, float64 roundtrips)
- Snapshot-to-state-update conversion
- Environment variable helpers
