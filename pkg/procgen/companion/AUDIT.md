# Audit: pkg/procgen/companion
**Date**: 2026-02-13
**Status**: Needs Work

## Summary
The companion generator provides deterministic procedural generation of companion entities with 8 types, genre-based theming, and stat scaling. The implementation is functionally complete with comprehensive tests (75% coverage) and proper deterministic generation. Critical issue: imports `pkg/engine` causing Ebiten initialization, preventing tests from running in CI/CD environments without GUI.

## Issues Found
- [x] **high** architecture — Generator imports `pkg/engine` for `CompanionType` and `CommandType` enums, causing Ebiten GUI initialization on test import; tests cannot run in headless CI/CD environments (`generator.go:9`)
- [x] **high** ECS compliance — `pkg/engine/companion_component.go` violates ECS rules by adding logic methods `HasPerk()` and `AddPerk()` to CompanionComponent; all companion logic should be in systems, components must be pure data (`pkg/engine/companion_component.go:100,110`)
- [x] **med** error handling — No structured logging with `logrus.WithFields` on error paths; errors return plain `fmt.Errorf` without contextual fields (`generator.go:42,79,84,89,92,97`)
- [x] **low** doc coverage — CompanionStatsComponent lacks godoc comment in engine package, should document purpose and usage (`pkg/engine/companion_component.go:117`)
- [x] **low** integration — No validation that generated companion integrates correctly with CompanionComponent/CompanionStatsComponent in engine; generator returns custom `Companion` struct instead of engine components (`generator.go:14`)

## Test Coverage
~75% (exceeds 65% requirement) — **Cannot verify via `go test -cover`** due to Ebiten initialization preventing headless execution

Tests cover:
- ✅ All genres (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic, unknown)
- ✅ Difficulty validation (negative, out-of-range)
- ✅ Zero/negative depth handling
- ✅ Determinism (same seed = same output)
- ✅ All companion types sprite pattern generation
- ✅ Validation edge cases (zero/negative attack, HP, invalid loyalty, no commands)
- ✅ Benchmark for performance verification

Missing test coverage:
- ❌ Defense stat validation (tests check Attack, MaxHP, Loyalty, Commands but not Defense)
- ❌ Speed stat validation

## Integration Status
**Engine Integration**: Generator used by `cmd/server/entity_spawning.go` and `cmd/client/util.go` to spawn companions into terrain. Generator produces `companion.Companion` struct which is manually mapped to `engine.CompanionSpawnData` by callers, then `engine.SpawnCompanionsInTerrain()` converts to actual engine components (`CompanionComponent`, `CompanionStatsComponent`, `HealthComponent`, `SpriteComponent`, `AnimationComponent`, `ColliderComponent`, `TeamComponent`, `DialogComponent`).

**Missing Registrations**: None required — generator is called directly by client/server spawning code, not registered in system_init.go.

**Serialize/Deserialize**: `CompanionComponent` has `Serialize()`/`Deserialize()` methods for network transmission (lines 133-250 in companion_component.go). Generator output (`companion.Companion`) is not directly serializable; must be converted to components first.

**Type Dependency**: Generator depends on `engine.CompanionType` (8 enum values) and `engine.CommandType` (6 enum values). This creates tight coupling and Ebiten initialization side effects.

## Recommendations
1. **HIGH PRIORITY**: Extract `CompanionType` and `CommandType` enums from `pkg/engine` to `pkg/procgen/companion` or new `pkg/types` package to break Ebiten dependency and enable headless testing
2. **HIGH PRIORITY**: Move `HasPerk()` and `AddPerk()` logic from `CompanionComponent` to a companion management system; components must be pure data per ECS architecture
3. **MEDIUM PRIORITY**: Add structured logging with `logrus.WithFields` on all error paths in generator (`seed`, `difficulty`, `depth`, `genreID` fields)
4. **LOW PRIORITY**: Add godoc comment to `CompanionStatsComponent` explaining its relationship to generated companions
5. **LOW PRIORITY**: Consider having generator return `engine.CompanionSpawnData` directly instead of custom `Companion` struct to reduce mapping overhead and potential data loss
