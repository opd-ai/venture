# Audit: github.com/opd-ai/venture/pkg/world/raids
**Date**: 2026-02-12
**Status**: Needs Work

## Summary
The raids package provides complete procedural raid dungeon generation and instance management for endgame multiplayer content. The package is well-structured with excellent test coverage (91.8%) and comprehensive documentation. However, it has one critical ECS violation in the related engine component (`RaidLockoutComponent`), lacks structured logging with logrus, and has no serialize/deserialize support for persistent raid instance state. The package correctly uses deterministic generation patterns and proper error handling throughout.

## Issues Found
- [x] **high** ECS compliance — `RaidLockoutComponent` in `pkg/engine/raid_component.go` has logic methods `IsLockedOut()` and `SetLockout()` which violates ECS principle that components must be pure data only (`pkg/engine/raid_component.go:65-85`)
- [x] **med** Error handling — Package lacks structured logging with `logrus.WithFields`; errors are returned but not logged with context for debugging (`generator.go`, `instance.go`, `lockout.go`, `manager.go`)
- [x] **med** Integration points — No serialize/deserialize support for `RaidInstance`, `PlayerLockout`, or `RaidDungeon` types; raid state cannot be persisted across server restarts (all files)
- [x] **low** Integration points — Raid components (`RaidInstanceComponent`, `RaidBossComponent`, `RaidLockoutComponent`) not registered in save/load system for persistence (`pkg/engine/raid_component.go`)
- [x] **low** Doc coverage — `mechanic.go` lacks package-level comment explaining mechanic generation algorithm (line 1-4 has only basic comment)

## Test Coverage
91.8% (target: 65%) ✅

**Breakdown by file:**
- `generator.go` - Full coverage with table-driven tests
- `instance.go` - Full coverage with concurrent safety tests
- `lockout.go` - Full coverage with time-based expiry tests
- `manager.go` - Full coverage with integration tests
- `names.go` - Full coverage with determinism tests
- All tests use deterministic seeds and table-driven patterns

## Integration Status
**How this package connects:**
- ✅ **V2 Terrain**: Uses `terrain.BSPGenerator` for dungeon layouts via `generator.go:29,198`
- ✅ **V2 Entities**: Uses `entity.EntityGenerator` for boss generation via `generator.go:20,336`
- ✅ **V8 Guilds**: Integrated in group validation and guild coordination via `manager.go:58-67`
- ✅ **V9 Economy**: Loot table integration via `generator.go:391-434`
- ✅ **Engine**: `RaidSystem` consumes this package via `pkg/engine/raid_system.go`
- ✅ **Client**: Raid UI handler in `cmd/client/handlers.go:786`
- ✅ **Server**: Server initialization in `cmd/server/v4_systems.go:185-186`

**Missing registrations:**
- ❌ Raid components not registered in `pkg/saveload/` for persistence
- ❌ No serialization support for saving active raid instances across restarts

**Deterministic generation verified:**
- ✅ All randomness uses `rand.New(rand.NewSource(seed))` pattern
- ✅ No use of global `rand` functions
- ✅ `time.Now()` only used in appropriate contexts (instance expiry, lockout tracking - NOT in procedural generation)
- ✅ Seed derivation uses deterministic hash function (`hashString()` at `generator.go:453-459`)

## Recommendations
1. **[HIGH PRIORITY]** Refactor `RaidLockoutComponent` in `pkg/engine/raid_component.go` - Move `IsLockedOut()` and `SetLockout()` logic to `RaidSystem` or create a dedicated lockout system. Component should only contain data fields.
2. **[MEDIUM PRIORITY]** Add structured logging with `logrus.WithFields` throughout error paths in `generator.go`, `instance.go`, `lockout.go`, `manager.go`. Use standard field names: `tier`, `instance_id`, `group_id`, `player_id`, `seed`.
3. **[MEDIUM PRIORITY]** Implement `Serialize()`/`Deserialize()` methods for `RaidInstance`, `PlayerLockout`, and `RaidDungeon` to support persistent state across server restarts.
4. **[LOW PRIORITY]** Register raid components in save/load system to ensure raid state persists.
5. **[LOW PRIORITY]** Expand package-level documentation in `mechanic.go` to explain the mechanic generation algorithm and scaling formulas.
