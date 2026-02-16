# Engine Progression Systems Sub-Audit

**Scope**: Progression systems — `progression_system.go`, `progression_components.go`, `skill_progression_system.go`, `class_progression_component.go`, `class_progression_system.go`, `achievement.go`, `stat_growth.go`, `reputation_system.go`, `base_stats_component.go`
**Date**: 2026-02-16
**Status**: Complete

## Files Audited

| File | Lines | Description |
|------|-------|-------------|
| `progression_system.go` | 342 | XP awarding, leveling, stat scaling, skill point management |
| `progression_components.go` | 281 | ExperienceComponent, LevelScalingComponent, SkillTreeComponent |
| `skill_progression_system.go` | 547 | Skill bonus calculation and application to entity stats |
| `class_progression_component.go` | 653 | ClassProgressionComponent, specializations, serialization |
| `class_progression_system.go` | 438 | Class leveling, specialization, dual-classing, stat growth |
| `stat_growth.go` | 382 | Per-class stat growth rates and calculation formulas |
| `achievement.go` | 354 | Achievement tracking and unlocking system |
| `reputation_system.go` | 274 | Reputation tracking, deed recording, alignment system |
| `base_stats_component.go` | 148 | Base stat storage to prevent bonus compounding |

## Issues Found: 6 (0 high, 3 medium, 3 low)

### MED-1: Unsafe type assertion in `GetLevel` — FIXED

**File**: `progression_system.go:241`
**Description**: `GetLevel()` used a bare type assertion `expComp.(*ExperienceComponent).Level` without checking the `ok` return value. If the component stored under the "experience" key was not actually an `*ExperienceComponent` (e.g., due to component replacement bugs or serialization errors), this would panic at runtime.
**Fix**: Added safe type assertion with `ok` check, returning default level 1 on failure.
**Risk**: Medium — panics crash the game/server.

### MED-2: `time.Now()` used instead of game clock — FIXED

**File**: `reputation_system.go:63`
**Description**: `RecordAction()` used `time.Now()` to timestamp deeds, but other systems (achievement.go, collection_system.go, etc.) correctly use `s.world.Clock.Now()`. Using system time breaks determinism in replays and makes deed timestamps inconsistent with the game's simulation clock.
**Fix**: Replaced `time.Now()` with `s.world.Clock.Now()` and removed unused `time` import.
**Risk**: Medium — breaks timestamp consistency across systems.

### MED-3: `getEntityFaction` always returned empty string — FIXED

**File**: `reputation_system.go:263`
**Description**: `getEntityFaction()` fetched the "faction" component but never read `FactionID` from it, always returning empty string. This meant reputation changes from kills, helps, and thefts never applied faction-specific impacts, making the entire faction reputation system non-functional.
**Fix**: Added proper `*FactionComponent` type assertion and return of `FactionID`.
**Risk**: Medium — faction reputation was completely broken.

### LOW-1: Division by zero in `updateHealthComponent` — FIXED

**File**: `class_progression_system.go:113`
**Description**: `healthPercent := health.Current / health.Max` would divide by zero if `health.Max` was 0, producing NaN that would propagate to all health calculations. This could occur with freshly created entities or corrupted component data.
**Fix**: Added zero check for `health.Max`, defaulting to 100% health preservation when Max is 0.

### LOW-2: Division by zero in `applyHealthGrowth` — FIXED

**File**: `class_progression_system.go:353`
**Description**: Same division by zero pattern as LOW-1 in the secondary class growth path.
**Fix**: Added zero check for `health.Max`, defaulting to 100% health preservation.

### LOW-3: `RecalculateSkillBonuses` missing nil entity check — FIXED

**File**: `skill_progression_system.go:533`
**Description**: Public function `RecalculateSkillBonuses()` accessed `entity.ID` without nil check, which would panic if called with a nil entity. Other public functions in the progression system (e.g., `AwardXP`, `SpendSkillPoint`) properly validate nil entities.
**Fix**: Added nil check with early return and warning log.

## Test Coverage

Existing test files provide good coverage:
- `progression_test.go`: XP awarding, leveling, callbacks, XP curves, error cases
- `progression_components_test.go`: Component initialization, skill learning/unlearning
- `class_progression_system_test.go`: Class leveling, specializations, dual-classing
- `skill_progression_system_test.go`: Bonus calculation, skill level scaling, integration
- `achievement_test.go`: Achievement unlocking and tracking

Note: Engine tests cannot run in headless CI due to Ebiten GLFW dependency. Tests are verified via `//go:build ignore` tagged test files or via the game client.

## Code Quality Notes

- All progression components follow the pure data + `Type()` pattern correctly
- Stat growth formulas are well-documented with clear per-class differentiation
- Base stats tracking (GAP-008 repair) correctly prevents bonus compounding
- Achievement system properly uses game clock (`world.Clock.Now()`)
- Skill tree component handles dependencies correctly for respec
