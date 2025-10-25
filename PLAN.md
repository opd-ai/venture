# Implementation Plan - Critical Gaps & Bugs

**Project:** Venture - Procedural Action RPG  
**Date:** October 25, 2025  
**Status:** Phase 8 Polish & Optimization - Moving to Production Release  
**Priority:** Game-breaking bugs first, then core mechanics, then polish

---

## 1. Death & Revival System (CRITICAL - Priority 1)

### Issue Identification
**Current State:** Death detection exists (`IsDead()` method), but no immobilization or action prevention. No loot drops. No revival mechanics.

**Impact:** Players/monsters can attack while dead. No consequence for death. Multiplayer revival mentioned in docs but not implemented.

### Root Cause
- Death callback exists but only triggers cleanup, not state changes
- No "dead" state component to disable actions
- No item spawning system for dropped loot
- No proximity detection for revival mechanics

### Implementation Steps

**1.1 Death State Component** (`pkg/engine/death_component.go`)
```go
type DeadComponent struct {
    TimeOfDeath float64
    DroppedItems []uint64 // Entity IDs of dropped items
}
```

**1.2 Movement Prevention** (Modify `pkg/engine/movement.go`)
- Check for DeadComponent before applying movement
- Return early if entity has dead state

**1.3 Combat Prevention** (Modify `pkg/engine/combat_system.go`)
- Skip dead entities in Update loop
- Prevent dead entities from being attack targets
- Block action inputs for dead player entities

**1.4 Loot Drop System** (Modify death callback in `cmd/client/main.go`)
- Access inventory component from dead entity
- Create item entities at death position with physics
- Add dropped items to world
- Clear inventory after spawning items

**1.5 Multiplayer Revival** (New system in `pkg/engine/revival_system.go`)
```go
type RevivalSystem struct {
    revivalRange float64 // 32.0 units
    revivalAmount float64 // 0.2 (20% health)
}
```
- Check proximity between living teammates and dead players
- Require input action (E key) to revive
- Restore 20% health and remove DeadComponent
- Network sync revival events

### Testing
- Unit test: Dead entity cannot move/attack
- Integration test: Loot spawns at death location
- Multiplayer test: Revival mechanics with 2 clients
- Performance test: Death with 100+ entities doesn't drop FPS

### Success Criteria
- ✅ Dead entities frozen in place
- ✅ Dead entities drop inventory items
- ✅ Living players can revive dead teammates
- ✅ Network synchronized across clients
- ✅ No performance impact (<1ms per death)

---

## 2. Input System Fixes (HIGH - Priority 2)

### Issue Identification
**Current State:** Input system exists but has documented bugs (GAP-001, GAP-002) with frame-persistent vs immediate consumption flags. Tutorial may miss key presses.

**Gaps:**
- UI labels may not match actual keybindings
- Menu navigation focus issues
- Input conflicts between game states
- No input buffering for combat responsiveness

### Root Cause
- Action flags consumed before all systems read them
- No centralized key binding registry
- State machine doesn't block inputs properly
- Tutorial checks flags after consumption

### Implementation Steps

**2.1 Key Binding Registry** (`pkg/engine/keybindings.go`)
- Central mapping of actions to keys
- UI automatically reads from registry
- Support remapping (future enhancement)

**2.2 Input State Separation** (Modify `pkg/engine/input_system.go`)
- Keep frame-persistent flags for UI/tutorial (already partially done)
- Separate consumption flags for gameplay systems
- Add input buffer (last 3 frames) for combo detection

**2.3 State-Based Input Filtering** (Modify `pkg/engine/input_system.go`)
```go
type GameState int
const (
    StateExploring = iota
    StateCombat
    StateMenu
    StateDialogue
)
```
- Block game inputs when menu open
- Allow only menu navigation during menu state
- Prevent movement during dialogue

**2.4 UI Label Validation** (Audit all UI systems)
- Map UI: Check keybinding display matches InputSystem
- Inventory UI: Verify key labels
- Skills UI: Confirm hotkey display
- Help System: Validate all key references

### Testing
- Manual test: Press all keys in all game states
- Automated test: Input consumption order
- Tutorial test: Complete tutorial without missed inputs
- UI test: All labels match actual bindings

### Success Criteria
- ✅ Tutorial completes without input issues
- ✅ UI labels 100% accurate
- ✅ No input conflicts between states
- ✅ Menu navigation smooth and responsive

---

## 3. Status Effects System (MEDIUM - Priority 3)

### Issue Identification
**Current State:** Status effects implemented (burning, poison, buffs) but potential edge cases with stacking, expiration, and conflicts.

**Potential Issues:**
- Multiple same-type effects may stack incorrectly
- Stat modifiers may not cleanly revert
- Network desync on effect application
- Visual feedback missing for some effects

### Root Cause
- Effect refresh logic overwrites instead of stacking
- Removal multiplies instead of subtracting modifiers
- No validation of stat values after effect removal

### Implementation Steps

**3.1 Stacking Validation** (Modify `pkg/engine/status_effect_system.go`)
- Track effect stacks per entity
- Decide stacking rules: refresh duration, stack magnitude, or ignore
- Add max stack limits per effect type

**3.2 Stat Modifier Tracking** (New component in `pkg/engine/combat_components.go`)
```go
type StatModifierTracker struct {
    BaseAttack float64
    BaseDefense float64
    ActiveModifiers map[string]float64 // effectID -> modifier
}
```
- Store base stats before modifications
- Track each effect's contribution
- Recalculate stats when effects change

**3.3 Network Synchronization** (Modify `pkg/network/protocol.go`)
- Add status effect application messages
- Sync effect timers periodically
- Validate effect state on client reconciliation

**3.4 Visual Feedback** (Connect to particle system)
- Burning: Fire particles
- Poison: Green drip particles
- Buffs: Glow aura
- Debuffs: Dark swirls

### Testing
- Edge case: Apply strength buff twice
- Edge case: Effect expires during combat
- Network test: Effect sync with 200ms latency
- Visual test: All effects show particles

### Success Criteria
- ✅ Effects stack/refresh correctly
- ✅ Stats revert to exact base values
- ✅ Network synchronized
- ✅ Visual feedback for all effects

---

## 4. Fog of War System (MEDIUM - Priority 4)

### Issue Identification
**Current State:** Fog of war mentioned in save/load (GAP-005), MapUI has GetFogOfWar/SetFogOfWar methods, but visibility calculations may be buggy.

**Suspected Issues:**
- FOW not updating during gameplay
- Visibility range incorrect
- Rendering artifacts
- Not integrated with lighting system

### Root Cause
- MapUI fog system not connected to movement
- No tile exploration tracking in movement system
- Rendering may not respect fog state

### Implementation Steps

**4.1 Visibility Calculation** (Modify `pkg/engine/movement.go`)
- Track player position changes
- Mark tiles as explored in radius (e.g., 8 tiles)
- Call MapUI.ExploreTile() for each visible tile

**4.2 Rendering Integration** (Modify rendering system)
- Unexplored tiles: Fully black
- Explored but not visible: 50% darkened
- Currently visible: Full brightness
- Add subtle fog texture overlay

**4.3 Performance Optimization**
- Only recalculate FOW when player moves
- Cache visible tile set
- Use spatial partitioning for large maps

### Testing
- Visual test: FOW updates as player moves
- Save/load test: FOW state persists
- Performance test: FOW calculation <0.5ms per frame

### Success Criteria
- ✅ Fog reveals as player explores
- ✅ Explored areas remain visible but dimmed
- ✅ No rendering artifacts
- ✅ Saves/loads correctly

---

## 5. Combat Calculations Validation (LOW - Priority 5)

### Issue Identification
**Current State:** Combat system fully implemented with damage, defense, criticals. Need validation of edge cases.

**Potential Issues:**
- Negative damage from high defense
- Critical hit calculation edge cases
- Resistance overflow (>100%)
- Shield absorption edge cases

### Implementation Steps

**5.1 Damage Formula Audit** (`pkg/engine/combat_system.go`)
- Ensure minimum damage (1.0) even with high defense
- Cap resistance at 90% (not 100%)
- Validate critical damage multipliers
- Test edge cases with extreme stats

**5.2 Shield System Validation** 
- Verify shield absorbs before health
- Test shield expiration during damage
- Check shield + resistance interaction

**5.3 Add Combat Logging**
- Debug flag for damage calculations
- Log: raw damage, after defense, after resistance, final
- Helps identify balance issues

### Testing
- Unit test: Extreme stat values
- Unit test: Shield edge cases
- Balance test: Damage feels appropriate at all levels

### Success Criteria
- ✅ No negative or zero damage bugs
- ✅ All edge cases handled gracefully
- ✅ Combat feels balanced

---

## 6. System Integration Audit (LOW - Priority 6)

### Issue Identification
**Current State:** All systems instantiated in cmd/client/main.go. Need verification of connection points.

**Check List:**
- ✅ Input → Movement → Collision (connected)
- ✅ Combat → StatusEffects → UI (connected)
- ✅ Combat → Camera shake (connected via GAP-012)
- ✅ Combat → Particles (connected via GAP-016)
- ✅ Death → Audio (connected via GAP-010)
- ✅ Death → Objectives (connected via GAP-004)
- ? Death → Loot (needs implementation)
- ? Revival → Network (needs implementation)

### Implementation Steps

**6.1 Event Flow Validation**
- Trace code paths for major game events
- Ensure all callbacks properly wired
- Add missing connections

**6.2 System Initialization Order**
- Review AddSystem() call sequence
- Ensure dependencies initialized first
- Document system dependencies

### Success Criteria
- ✅ All system connections documented
- ✅ Event flow works end-to-end
- ✅ No orphaned systems

---

## 7. Multiplayer Synchronization (LOW - Priority 7)

### Issue Identification
**Current State:** Network prediction and lag compensation implemented. Need testing of death/revival sync.

**Requirements:**
- Death state synchronized
- Revival actions validated by server
- Dropped items spawned on all clients
- No desync during combat

### Implementation Steps

**7.1 Death Event Protocol**
- Client predicts death locally
- Server validates and broadcasts
- All clients spawn loot items deterministically

**7.2 Revival Action Protocol**
- Client sends revival request with target ID
- Server validates proximity and living state
- Server broadcasts revival to all clients
- Clients update health and remove dead state

**7.3 Testing**
- 2 clients: simultaneous death
- 2 clients: revival during enemy attack
- High latency: death with 1000ms delay

### Success Criteria
- ✅ Death synchronized <100ms
- ✅ Revival works reliably
- ✅ No item duplication bugs

---

## Implementation Order

### Week 1: Critical Systems
1. **Day 1-2:** Death state and action prevention (Priority 1.1-1.3)
2. **Day 3:** Loot drop system (Priority 1.4)
3. **Day 4-5:** Revival mechanics (Priority 1.5)

### Week 2: Input & Polish
4. **Day 6-7:** Input system fixes (Priority 2)
5. **Day 8:** Status effects validation (Priority 3)
6. **Day 9:** Fog of war fixes (Priority 4)
7. **Day 10:** Combat validation and testing (Priority 5-7)

---

## Testing Strategy

### Automated Tests
- Unit tests for each new component
- Integration tests for death/revival flow
- Network tests for multiplayer sync
- Performance benchmarks for critical paths

### Manual Testing
- Playthrough: Complete dungeon run with deaths
- Multiplayer: 2-player co-op session
- Edge cases: Stress test with extreme scenarios
- UI audit: Check all labels and bindings

### Performance Targets
- 60 FPS minimum with all systems active
- <500ms world generation time
- <100KB/s network bandwidth per player
- <500MB client memory usage

---

## Risk Assessment

**High Risk:**
- Revival system network desync → Mitigation: Server authority validation
- Loot duplication bugs → Mitigation: Deterministic spawning with seed
- Performance impact of death events → Mitigation: Object pooling

**Medium Risk:**
- Input system regression → Mitigation: Comprehensive test suite
- Status effect edge cases → Mitigation: Stat modifier tracking

**Low Risk:**
- FOW rendering artifacts → Mitigation: Incremental testing
- Combat balance issues → Mitigation: Configurable damage formulas

---

## Definition of Done

**Each feature complete when:**
1. ✅ Implementation code written and tested
2. ✅ Unit tests with >80% coverage
3. ✅ Integration tests passing
4. ✅ Manual testing completed
5. ✅ Documentation updated
6. ✅ Performance benchmarks met
7. ✅ Code review passed
8. ✅ No regressions introduced

**Project complete when:**
- All Priority 1-2 items implemented and tested
- All Priority 3-4 items implemented
- Priority 5-7 validated or documented as acceptable
- Full multiplayer session tested without issues
- Performance targets met across all systems
- Production release candidate ready

---

## Notes

- Follow ECS architecture patterns throughout
- Maintain deterministic behavior for multiplayer sync
- Keep network bandwidth under targets
- All changes must pass existing test suite
- Document any breaking changes or new APIs
- Update user manual with new features (revival system)

**Last Updated:** October 25, 2025
