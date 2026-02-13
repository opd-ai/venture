# Audit: pkg/integration/narrative_world
**Date**: 2026-02-12
**Status**: Needs Work

## Summary
The narrative_world package implements companion-driven story events, personal quests, memory-based dialogue, and conflict detection for V9 narrative systems. The package is well-structured with comprehensive tests (>95% estimated coverage based on test files), proper ECS integration, and full client/server registration. However, it has 2 high-severity determinism violations and 4 medium-severity issues that must be addressed before production use.

## Issues Found
- [x] **high** Deterministic procgen — Non-deterministic timestamp in RecordMemory using `time.Now()` breaks replay/networking (`manager.go:431`)
- [x] **high** Deterministic procgen — Memory pruning uses `time.Now()` for recency scoring, non-deterministic (`manager.go:475`)
- [x] **med** Integration points — No Serialize/Deserialize methods for StoryEventManager state (quests, memories, conflicts persist only in memory)
- [x] **med** Integration points — CompanionMemory and PersonalQuest types lack serialization for save/load support
- [x] **med** Test coverage — Tests cannot run in CI/headless environment due to Ebiten GUI initialization (package imports engine which imports Ebiten)
- [x] **low** Documentation — MemoryEvent.Timestamp field usage unclear (should use game-time ticks instead of wall clock)
- [x] **low** Error handling — CheckConflict creates conflict on probabilistic basis but doesn't log when conflict is NOT created (hard to debug)

## Test Coverage
Unable to measure (Ebiten GUI dependency blocks headless test execution). Test files comprehensively cover all public APIs with table-driven tests + 4 benchmarks. **Estimated >95%** based on test file analysis (26 test functions + benchmarks cover all manager/system methods).

## Integration Status
**Fully integrated:**
- ✅ Client: Registered in `cmd/client/handlers.go` as `narrativeWorldSystem`, created with `narrative_world.NewSystem(game.World, game.GetWorldSeed())`
- ✅ Server: Imported in `cmd/server/v9_systems.go` as `narrativeworld` alias
- ✅ Engine: `pkg/engine/narrative_system.go` defines `CompanionStoryProvider` interface for integration
- ✅ Engine tests: `pkg/engine/narrative_world_integration_test.go` validates System integration
- ✅ System registration: NewSystem creates manager, Update() processes companion entities
- ⚠️ **Missing:** No persistence layer for quests/memories/conflicts (requires serialization)

## Recommendations
1. **CRITICAL:** Replace `time.Now()` with deterministic game ticks/timestamps (manager.go:431, 475)
   - Change `MemoryEvent.Timestamp` from `time.Time` to `int64` (game tick count)
   - Update RecordMemory to accept tick parameter instead of calling time.Now()
   - Modify pruneOldMemories to use tick-based recency calculation
   
2. **HIGH:** Implement serialization for persistence:
   - Add `Serialize()/Deserialize()` to StoryEventManager
   - Add `Serialize()/Deserialize()` to CompanionMemory, PersonalQuest, CompanionConflict, CrossCompanionStory
   - Integrate with pkg/saveload for quest/memory persistence
   
3. **MEDIUM:** Fix test execution in headless environments:
   - Create mock/stub World implementation that doesn't import Ebiten
   - Or add build tags to skip Ebiten init in test mode
   
4. **LOW:** Add structured logging for conflict detection failures:
   - Log with logrus.WithFields when CheckConflict returns false (no conflict created)
   - Include personality trait details for debugging conflict probability
