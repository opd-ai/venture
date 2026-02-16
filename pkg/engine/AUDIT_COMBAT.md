# Engine Combat Systems Sub-Audit

**Scope**: Combat system files — `combat_system.go`, `combat_components.go`, `player_combat_system.go`
**Date**: 2026-02-16
**Status**: Complete

## Files Audited

| File | Lines | Description |
|------|-------|-------------|
| `combat_system.go` | 1457 | CombatSystem: damage calculation, status effects, projectiles, feedback |
| `combat_components.go` | 373 | HealthComponent, StatsComponent, AttackComponent, StatusEffectComponent, TeamComponent, ShieldComponent, DeadComponent |
| `player_combat_system.go` | 213 | PlayerCombatSystem: player input → combat action bridge |

## Issues Found: 5 (0 high, 3 medium, 2 low)

### MED-1: `getEntityStats` used unsafe type assertion — FIXED

**File**: `combat_system.go:592`
**Description**: `getEntityStats` used unchecked type assertion `statsComp.(*StatsComponent)` which would panic if a non-`*StatsComponent` was stored under the "stats" key. While unlikely in normal operation, this violates defensive coding practices and could cause crashes from corrupted entity state.
**Fix**: Changed to comma-ok pattern with nil return on type mismatch.
**Test**: `TestGetEntityStats_SafeTypeAssertion`

### MED-2: `applyDamageAndFeedback` recalculated damage consuming extra RNG state — FIXED

**File**: `combat_system.go:495`
**Description**: `applyDamageAndFeedback` called `s.calculateDamage()` a second time to get `baseDamage` for logging. This call consumed an additional random number from the RNG (for crit roll), breaking determinism and potentially logging a different baseDamage than what was actually used for the attack. This violated the deterministic generation requirement.
**Fix**: Changed `computeFinalDamage` to return `(finalDamage, baseDamage, isCrit)` and passed `baseDamage` through the call chain instead of recalculating.
**Test**: `TestComputeFinalDamage_ReturnsBaseDamage`, `TestCombatSystem_DeterministicRNG`

### MED-3: `additionalDamageCallbacks` never invoked — FIXED

**File**: `combat_system.go:46,1016`
**Description**: `AddDamageCallback` appended to `additionalDamageCallbacks` slice, but these callbacks were never invoked during attack processing. Only `onDamageCallback` (set via `SetDamageCallback`) was called. Any system registering via `AddDamageCallback` would silently have its callback ignored.
**Fix**: Added invocation loop for `additionalDamageCallbacks` in `applyDamageAndFeedback`, matching the existing pattern used for `additionalCriticalHitCallbacks`.
**Test**: `TestAdditionalDamageCallbacks`

### LOW-1: `triggerScreenShake` used unsafe type assertion — FIXED

**File**: `combat_system.go:853`
**Description**: `targetHealthComp.(*HealthComponent).Max` used unchecked type assertion that could panic on corrupted component data.
**Fix**: Changed to comma-ok pattern with safe fallback to default maxHP value.

### LOW-2: Misplaced godoc comment on `getEntityStats`/`applyAttackFeedback` — FIXED

**File**: `combat_system.go:547-548`
**Description**: The godoc comment for `getEntityStats` was erroneously placed above `applyAttackFeedback`, and the actual `getEntityStats` function had no godoc. This caused incorrect documentation for both functions.
**Fix**: Moved godoc to correct function, wrote proper comment for `applyAttackFeedback`.

## Architecture Assessment

The combat system is well-designed with clear separation of concerns:
- **CombatSystem**: Authoritative damage calculation using `combat.CombatResolver` for consistent formulas
- **PlayerCombatSystem**: Clean bridge between input and combat, properly decomposed into small methods
- **Components**: Pure data structures following ECS architecture (HealthComponent, StatsComponent, etc.)
- **Callback system**: Extensible event system for visual/audio feedback (death, damage, crit, evasion, block, shield, heal)
- **Projectile support**: Ranged weapons spawn projectile entities with physics integration
- **Deterministic RNG**: Seed-based `rand.New(rand.NewSource(seed))` for reproducible combat

## ECS Compliance

⚠️ **MINOR NOTE** — Combat components have convenience methods (e.g., `HealthComponent.Heal()`, `AttackComponent.CanAttack()`, `ShieldComponent.AbsorbDamage()`). These are simple state mutations, not behavioral logic, and are consistent with the project's established pattern (used throughout the codebase). The actual combat logic resides in `CombatSystem.Attack()` and its helper methods.

## Determinism Compliance

✅ **PASS** — Uses `rand.New(rand.NewSource(seed))` for all random decisions (crit rolls, evasion, block). After MED-2 fix, no redundant RNG consumption occurs.

## Test Coverage

New tests added:
- `TestGetEntityStats_SafeTypeAssertion` — 2 sub-tests for safe type assertion
- `TestAdditionalDamageCallbacks` — Verifies all registered callbacks fire with correct damage
- `TestComputeFinalDamage_ReturnsBaseDamage` — Verifies baseDamage return value
- `TestCombatSystem_DeterministicRNG` — 5-attack sequence determinism verification

Existing tests (28 test functions + 1 benchmark) all pass without modification.
