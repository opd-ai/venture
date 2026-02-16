# AI & Behavior Systems Sub-Audit

**Date**: 2026-02-16
**Scope**: AI state machine, behavior trees, squad coordination, and AI bridge systems
**Status**: Complete

## Files Audited

### Core AI
- `ai_system.go` — Main AI state machine system (1270 lines)
- `ai_components.go` — AIComponent with state machine, patrol, and detection
- `ai_state_bubble_system.go` — Genre-aware AI state indicators

### Behavior Trees
- `behavior_tree_system.go` — BehaviorTreeSystem execution
- `behavior_tree_nodes.go` — Node types (Sequence, Selector, Parallel, Inverter, Repeat)
- `behavior_tree_actions.go` — Action/Condition leaf nodes
- `behavior_tree_component.go` — BehaviorTreeComponent data
- `behavior_tree_archetypes.go` — Pre-built enemy archetypes (Melee, Ranged, Tank, Support, Stealth)

### Squad System
- `squad_system.go` — Squad coordination and formations
- `squad_component.go` — SquadComponent with roles and formations

### AI Bridge Systems
- `faction_aware_ai_system.go` — Faction reputation → AI hostility
- `status_effect_ai_system.go` — Status effects → AI disabling
- `weather_aware_ai_system.go` — Weather → AI detection range

## Issues Found & Fixed

### MED-1: Duplicate code blocks in `faction_aware_ai_system.go` (FIXED)

`getPlayerFactionStandings()` and `getEntityFaction()` contained identical duplicate type assertion blocks with misleading comments ("Also handle non-pointer faction components" / "Handle value type"). Both blocks checked `comp.(*FactionComponent)` identically. Removed the duplicate blocks.

### MED-2: Nil pointer dereference in `ai_system.go` debug logging (FIXED)

`processDetect()` (line 405) and `logChaseState()` (line 481) accessed `aiComp.Target.ID` in debug log statements without nil-checking `aiComp.Target` first. If Target was cleared between state transitions, this would panic. Added `aiComp.Target != nil` guard to both log conditions.

### LOW-1: `AdvanceToNextWaypoint` single-waypoint edge case (FIXED)

With a single-waypoint patrol route and `PatrolReverse=true`, the index arithmetic would produce out-of-bounds indices (-1 or 1 for a len-1 slice), causing `GetCurrentWaypoint()` to return nil on subsequent calls. Fixed by returning early when `len(PatrolWaypoints) <= 1`.

## Remaining Notes (No Action Required)

### LOW: `NewBlackboard()` defaults to seed 0
All behavior trees created via `NewBehaviorTreeComponent` get a blackboard with seed 0, making random behavior identical across all entities unless the caller explicitly sets a seed via `NewBlackboardWithSeed`. This is documented in the code comments but callers should be aware.

### LOW: `companion_ai_system.go` has 0% test coverage
The companion AI system exists but has no test coverage. It was excluded from this audit scope (companion systems are a separate domain).

## Test Coverage Summary

| File | Coverage |
|------|----------|
| `ai_components.go` | 100% |
| `ai_system.go` | ~55% (many debug-only log helpers) |
| `ai_state_bubble_system.go` | ~85% |
| `behavior_tree_nodes.go` | ~65% (Parallel/Inverter/Repeat at 0%) |
| `behavior_tree_actions.go` | ~35% (many action builders untested via integration) |
| `behavior_tree_component.go` | 100% |
| `behavior_tree_system.go` | ~67% |
| `behavior_tree_archetypes.go` | ~5% (builder functions uncovered) |
| `squad_system.go` | ~85% |
| `squad_component.go` | ~85% |
| `faction_aware_ai_system.go` | ~75% |
| `status_effect_ai_system.go` | ~92% |
| `weather_aware_ai_system.go` | ~88% |

**Overall AI subsystem coverage**: ~70% (meets 65% minimum threshold)
