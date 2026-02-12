# Audit: github.com/opd-ai/venture/pkg/world/housing
**Date**: 2026-02-12
**Status**: Needs Work

## Summary
The housing package provides player housing, guild halls, blueprint sharing, and spatial management with comprehensive implementation (2,365 LOC + 3,220 test LOC). Overall code quality is high with thread-safe concurrency, good test coverage, and clean architecture. Critical issues include non-deterministic timestamp usage in Plot/GuildHall creation (violates deterministic procgen standard), missing error logging with structured fields, and lack of component serialization for persistence.

## Issues Found
- [x] **high** Deterministic procgen — Non-deterministic `time.Now()` used in Plot/GuildHall creation timestamps instead of seed-based generation (`types.go:124`, `types.go:317`, `guildhall.go:123`, `guildhall.go:152`, `guildhall.go:157`, `guildhall.go:193`)
- [x] **high** Error handling — No structured logging with `logrus.WithFields` on error paths; errors returned without logging context (`manager.go:42-55`, `manager.go:69`, `persistence.go:22`, `persistence.go:38`, `blueprint.go:241`, `guildhall_manager.go:36`, `guildhall_manager.go:49`)
- [x] **med** Integration points — HousingComponent missing Serialize/Deserialize methods for persistence support (`component.go:4-11`)
- [x] **med** Integration points — HousingComponent not registered in engine component registry (no references found in `pkg/engine/*.go`)
- [x] **low** Doc coverage — Integration test placeholder comments should be removed or implemented (`integration_test.go:333`, `integration_test.go:556`)
- [x] **low** Error handling — `ContributeMaterial` silently swallows `AdvancePhase` errors with comment justification, but doesn't log the error (`guildhall_manager.go:88-92`)

## Test Coverage
Unable to measure coverage (requires GUI environment for Ebiten runtime). Package has extensive test suite: 3,220 test LOC covering API, blueprints, guildhalls, integration, manager, persistence, spatial grid, types, and UI.

Estimated coverage: 70-85% based on:
- ✅ All public methods have corresponding tests
- ✅ Manager/GuildHallManager have comprehensive unit tests
- ✅ Blueprint library filtering/sorting tested
- ✅ Spatial grid collision detection tested
- ✅ Persistence save/load cycle tested
- ✅ UI rendering tested with stub screen

## Integration Status
**World Management**: Housing package integrates with world persistence layer through `Manager.Save/Load` and federation sync via `SyncHouseFromFederation`.

**Procedural Generation**: Blueprint system references `pkg/procgen/building` (BuildingDefinition) and UI integrates with `pkg/procgen/furniture` for catalog rendering.

**Client UI**: HousingUI implements menu system with building construction, furniture placement, and guild hall management. Uses Ebiten for rendering. Provides `HousingUIProvider` interface for InputSystem ESC key handling (`ui.go:85`).

**Missing Registrations**:
1. HousingComponent not found in `pkg/engine/components.go` component registry
2. No housing system found in `pkg/engine/` to process HousingComponent entities
3. Blueprint library not wired to client/server startup (library instances created but not connected to world state)

**Deterministic Generation**: Manager uses seed-based position generation correctly (`manager.go:211-229`) with documented justification for math/rand usage, but Plot/GuildHall creation uses non-deterministic timestamps.

## Recommendations
1. **Replace `time.Now()` with seed-based timestamps** — Modify `NewPlot`, `NewBlueprint`, `NewGuildHall` to accept seed parameter and generate deterministic CreatedAt/ModifiedAt timestamps using `time.Unix(seed, 0)` or similar. Update callers to pass appropriate seeds. (Addresses high-severity deterministic procgen violation)

2. **Add structured logging to all error paths** — Import logrus, add `logrus.WithFields` calls before returning errors. Use standard field names: `plotID`, `ownerID`, `guildID`, `filename`, `error`. Example: `logrus.WithFields(logrus.Fields{"plotID": plot.ID, "error": err.Error()}).Error("failed to validate plot placement")` (Addresses high-severity error handling gap)

3. **Implement HousingComponent persistence** — Add `Serialize() ([]byte, error)` and `Deserialize([]byte) error` methods to HousingComponent. Serialize PlotID as JSON. Add tests for round-trip serialization. (Addresses med-severity integration gap)

4. **Register HousingComponent in engine** — Add HousingComponent to component registry in `pkg/engine/components.go`. Create or extend housing system in `pkg/engine/` to process entities with HousingComponent (e.g., sync plot positions with entity positions). (Addresses med-severity integration gap)

5. **Log swallowed errors** — In `guildhall_manager.go:88-92`, add `logrus.WithFields(logrus.Fields{"guildID": guildID, "error": err.Error()}).Warn("auto-advance phase failed")` before the comment. (Addresses low-severity error handling)
