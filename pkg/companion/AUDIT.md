# Audit: pkg/companion/learning

**Date**: 2026-02-16
**Coverage**: 92.5%
**Status**: Complete

## Summary

The `learning` package implements companion AI skill progression (24 skills across 8 categories), personality evolution (10 opposing traits), and behavioral memory (LRU-based, up to 1000 events). It uses deterministic time providers for reproducible state.

## Issues Found

### Fixed: 2 (1 med, 1 low)

1. **MED** (fixed): Race condition in `CompanionLearningSystem.Update()` — accessed `s.manager.companions` map without holding a lock while `AddCompanion`/`RemoveCompanion` could mutate it concurrently. Fixed by taking a snapshot under `RLock` before iterating.

2. **LOW** (fixed): Nil pointer dereference risk in `ProcessCombatAction`, `ProcessSocialInteraction`, `ProcessExploration`, `AdaptBehaviorToCombatStyle`, and `GeneratePersonalityDescription` — functions accessed `comp` fields without nil checks. Added nil guards with early return.

### Remaining: 2 (0 high, 0 med, 2 low)

1. **LOW**: Hardcoded magic numbers for skill decay (0.1), trait clamping (0.0-1.0), LRU limits (1000), and XP values are not configurable.

2. **LOW**: `Deserialize()` rebuilds prerequisites as empty and defaults cost to 1 — deserialized skill trees have incomplete structure.

## Test Coverage

- Concurrency test added for Update + Add/Remove races
- Nil guard tests added for all Process* functions
- Nil guard tests for GeneratePersonalityDescription and AdaptBehaviorToCombatStyle
