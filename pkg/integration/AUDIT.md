# Audit: github.com/opd-ai/venture/pkg/integration
**Date**: 2026-02-16
**Status**: Needs Work

## Summary
The integration package coordinates cross-system features across 10 sub-packages (73 Go files, ~15K LOC). Five sub-packages have excellent test coverage (87-98%) and deterministic design. Four sub-packages have test infrastructure that requires Ebiten runtime (transitive dependency via engine.World), causing CI/headless test failures. Critical issues include non-deterministic time usage in guild_housing, trade_routes, political_warfare, and world_events packages that violates the deterministic generation requirement.

## Issues Found
- [ ] <severity:high> deterministic procgen — Non-deterministic `time.Now()` used for ID generation in guild housing, breaking multiplayer synchronization requirements (`guild_housing/guild_housing_manager.go:78,246,355,443,568`)
- [ ] <severity:high> deterministic procgen — Non-deterministic `time.Now()` used for route timing in trade routes, creating desync potential (`trade_routes/manager.go:216,217,261,262,273,524`)
- [ ] <severity:high> deterministic procgen — Non-deterministic `time.Now()` used for war timing and embargo tracking in political warfare (`political_warfare/manager.go:105,148,153,158,212,252,284,338,371,420,455,557`)
- [ ] <severity:high> deterministic procgen — Non-deterministic `time.Now()` used for event timing in world events (`world_events/manager.go:32,44,87,111,173,426`)
- [ ] <severity:high> deterministic procgen — Non-deterministic `time.Now()` used for fleet management timestamps (`guild_vehicle/fleet_manager.go:48,49,80,81,99,100,104,125,143,164,202,219`)
- [ ] <severity:high> test coverage — 4 packages fail in headless/CI environments due to Ebiten init requirement via engine.World dependency: guild_housing, narrative_world, political_warfare, trade_routes (causes panic: "GLFW library is not initialized")
- [ ] <severity:med> integration points — No clear system registration pattern documented; systems appear manually wired in client/server initialization
- [ ] <severity:low> doc coverage — Root `doc.go` describes integration *tests* but package contains production integration systems (choice_consequences, companion_housing, housing_crafting, etc.) — misleading scope statement

## Test Coverage
**Testable packages**: 87.1-98.5% (choice_consequences: 87.1%, companion_housing: 93.5%, guild_vehicle: 94.6%, housing_crafting: 98.5%, world_events: 91.2%)
**Failing packages**: guild_housing, narrative_world, political_warfare, trade_routes (Ebiten init panic)
**Overall**: 5/9 sub-packages pass; strong table-driven test patterns where functional

## Integration Status
**Cross-system dependencies**:
- `companion_housing` ✅ Integrates engine (CompanionComponent), world/housing (blueprints), deterministic via injectable TimeProvider
- `housing_crafting` ✅ Integrates world/housing (furniture), procgen/recipe (generation), deterministic seed-based
- `guild_housing` ❌ Integrates network/federation/guild, world/housing; uses time.Now() for IDs/timestamps (non-deterministic)
- `guild_vehicle` ⚠️ Integrates network/federation/guild, procgen/vehicle; uses time.Now() for fleet timestamps
- `choice_consequences` ✅ Integrates engine (CompanionComponent), narrative; injectable TimeProvider for determinism
- `narrative_world` ⚠️ Integrates engine (CompanionComponent), companion/learning, procgen/quest; injectable TimeProvider but test failures due to Ebiten
- `political_warfare` ❌ Integrates engine.World, network/federation/guild; uses time.Now() throughout (wars, treaties, embargoes)
- `trade_routes` ❌ Integrates procgen/vehicle; uses time.Now() for route timestamps and mission timing
- `world_events` ❌ Uses time.Now() for event scheduling and timing

**System registration**: No centralized registry found. Systems like NarrativeWorldSystem, PoliticalWarfareSystem appear manually constructed in client/server initialization. Guild housing and trade routes use Manager pattern without ECS System wrapper (inconsistent).

## Recommendations
1. **HIGH PRIORITY**: Replace all `time.Now()` calls with injectable TimeProvider pattern (already exists in choice_consequences and narrative_world) for guild_housing, trade_routes, political_warfare, world_events, guild_vehicle
2. **HIGH PRIORITY**: Fix Ebiten test dependency — use stub engine.World or mock World interface to break transitive Ebiten dependency in tests for guild_housing, narrative_world, political_warfare, trade_routes
3. **MEDIUM PRIORITY**: Document system registration pattern — create INTEGRATION.md showing how to register new integration systems in client/server initialization
4. **LOW PRIORITY**: Update root doc.go to clarify package contains both integration *systems* and integration *tests*, not just tests
5. **LOW PRIORITY**: Standardize on System wrapper pattern — guild_housing and trade_routes use Manager-only (no ECS System), while narrative_world and political_warfare have System wrappers for consistency
