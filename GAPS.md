# Bug and Gap Details — 2026-04-24 (Rev 5)

> **Rev 5 — deep behavioral-correctness pass (2026-04-24).** This file supersedes
> `GAPS.md` Rev 3 (2026-04-25). All G1–G31 findings from prior revisions are
> confirmed resolved. Seven new gaps G32–G38 are documented below.
>
> **Legacy ID compatibility**: G1–G31 are preserved in `AUDIT.md` and in the
> historical `GAPS.md` entries; this file contains only the new Rev 5 findings.
>
> **Status legend**: ✅ RESOLVED | ⚠️ PARTIAL | 🔴 OPEN

---

## G32 — `AdvancedClassSystem` Adds Stat Bonuses Every Frame Without Guard

- **Category**: Per-Frame Mutation
- **Status**: 🔴 OPEN
- **Severity**: CRITICAL

**Evidence** (`pkg/engine/advanced_class_system.go`):
```go
// Update — called every frame (lines 27–48)
func (acs *AdvancedClassSystem) Update(entities []*Entity, deltaTime float64) {
    classEntities = acs.world.GetEntitiesWith("advanced_class")
    for _, entity := range classEntities {
        playerID := strconv.FormatUint(entity.ID, 10)
        stats, err := acs.manager.CalculateTotalStats(playerID)
        if err != nil { continue }
        acs.applyStatBonuses(entity, stats)   // ← called unconditionally each frame
    }
}

// applyHealthBonuses (line 68)
health.Max += float64(bonuses.Health)   // no subtraction of prior value

// applyManaBonuses (line 84)
mana.Max += bonuses.Mana               // no subtraction of prior value

// applyStatsBonuses (lines 100–104)
stats.Attack     += float64(bonuses.Strength)
stats.Defense    += float64(bonuses.Defense)
stats.MagicPower += float64(bonuses.Intelligence)
stats.CritChance += bonuses.CritChance
stats.CritDamage += bonuses.CritDamage
```

- **Incorrect Behavior**: For a Warrior with `bonuses.Strength = 10` (attack
  bonus per strength point), `stats.Attack` grows by 600 per second at 60 FPS.
  After 10 seconds the player has +6 000 attack on top of the base 10, making
  the game trivially winnable. `health.Max` and `mana.Max` grow similarly
  without bound. Values are never clamped.

- **Root Cause**: The system applies bonuses additively on every Update tick
  without storing the previously-applied value and subtracting it first. The
  identical pattern was fixed for `TalentSystem` (G23) by adding
  `removeTalentBonuses` + `AppliedDeltas` tracking, but `AdvancedClassSystem`
  received no equivalent fix.

- **Impact**: Every player who has an advanced class assigned (all players after
  calling `InitializePlayerClass` from `cmd/client/handlers.go:3473`) develops
  unbounded stats within seconds of gameplay, breaking all combat balance.

- **Remediation**: Store the last-applied bonuses on the system or in the
  `AdvancedClassComponent`. Before applying, subtract the previous bonuses; after
  applying, cache the new bonuses. Pattern from `talent_system.go:77`:
  ```go
  acs.removeBonuses(entity, acs.lastApplied[entity.ID])
  acs.applyStatBonuses(entity, stats)
  acs.lastApplied[entity.ID] = stats
  ```
  Alternatively, use the existing `statBonusApplier` helper
  (`pkg/engine/statmod.go`) which already implements this pattern correctly.

- **Dependencies**: None — standalone fix.

- **Effort**: small

---

## G33 — `StatusEffectCriticalChanceSystem` Permanently Corrupts `CritChance`

- **Category**: Per-Frame Mutation
- **Status**: 🔴 OPEN
- **Severity**: CRITICAL

**Evidence** (`pkg/engine/status_effect_crit_chance_system.go:60–92`):
```go
func (s *StatusEffectCriticalChanceSystem) Update(entities []*Entity, dt float64) {
    // Clear cache each frame (lines 61–64)
    for k := range s.critCache {
        delete(s.critCache, k)
    }

    for _, entity := range entities {
        stats := entity.GetStats()
        if stats == nil { continue }

        modifier := s.calculateCritModifier(entity)
        if modifier == 0.0 {
            continue              // expired effect — no subtraction of prior modifier
        }

        s.critCache[entity.ID] = modifier
        stats.CritChance += modifier   // unconditional ADD every frame (line 82)

        if stats.CritChance < 0.0 { stats.CritChance = 0.0 }
        else if stats.CritChance > 1.0 { stats.CritChance = 1.0 }
    }
}
```
Because `critCache` is cleared at lines 61–64, `calculateCritModifier` always
recomputes from scratch and never returns a cached value. The modifier is then
added to `stats.CritChance` on **every frame** without subtracting the
contribution from the previous frame.

- **Incorrect Behavior (positive modifier, e.g. "blessed" = +0.10)**:
  - Frame 1: CritChance = 0.05 + 0.10 = 0.15
  - Frame 10: CritChance = 0.95 + 0.10 = 1.05 → clamped to 1.0
  - Blessed expires: `modifier = 0.0`, entity skipped via `continue`
  - CritChance stays permanently at 1.0 (100% crit forever)

- **Incorrect Behavior (negative modifier, e.g. "cursed" = −0.10)**:
  - Frame 1: CritChance = 0.05 − 0.10 = −0.05 → clamped to 0.0
  - All frames: CritChance stays at 0.0 (clamped each frame)
  - Cursed expires: entity skipped via `continue`
  - CritChance stays permanently at 0.0 (0% crit forever)

- **Root Cause**: The `critCache` serves as a "what modifier did we store for
  combat lookup" cache, but it was repurposed without adding a "what was the last
  crit delta applied to stats" cache. Without knowing the previous delta, it is
  impossible to undo it. The `continue` on `modifier == 0.0` skips undo.

- **Impact**: Any player or enemy briefly affected by a crit-modifier status
  effect will have their CritChance permanently locked at 0 or 1 after the
  effect expires. All PvP and PvE combat balance is affected.

- **Remediation**: Add a `prevApplied map[uint64]float64` field tracking the
  last delta written to each entity's `stats.CritChance`. Each frame:
  ```go
  // Undo previous contribution
  stats.CritChance -= s.prevApplied[entity.ID]

  // Compute and apply new contribution
  modifier = s.calculateCritModifier(entity)
  stats.CritChance += modifier
  s.prevApplied[entity.ID] = modifier

  // Clamp
  ```
  On entity removal or effect expiry, ensure `delete(s.prevApplied, id)`.

  This is the same delta-tracking pattern correctly implemented in
  `TerrainAmbushCritSystem` (lines 127–148, which tracks `s.critBonuses` and
  subtracts `currentBonus` before adding `newBonus`).

- **Dependencies**: None.

- **Effort**: small

---

## G34 — `EquipmentSetBonusSystem` Output Never Consumed

- **Category**: System Output Never Used (Dangling Integration)
- **Status**: 🔴 OPEN
- **Severity**: HIGH

**Evidence** (`pkg/engine/equipment_set_bonus_system.go:207–299`,
`pkg/engine/equipment_set_bonus_component.go:83–110`):

The system correctly detects equipment changes (via `currentHash !=
setBonus.LastEquipmentHash`), recalculates which sets are active, and
populates `EquipmentSetBonusComponent.ActiveSets` and `CombinedBonus`. It
exposes these bonus values through the following methods:
- `GetTotalDamageBonus() int`
- `GetTotalDefenseBonus() int`
- `GetTotalHealthBonus() int`
- `GetTotalAttackSpeed() float64`
- `GetTotalCritBonus() float64`

A grep of all non-test Go files in `pkg/` and `cmd/` for any of these method
names returns **zero results** outside of the component file itself and test
files. Neither `CombatSystem.calculateDamage` nor `InventorySystem` nor any
other system reads from `EquipmentSetBonusComponent`.

```bash
grep -rn "GetTotalDamageBonus\|GetTotalDefenseBonus\|GetTotalHealthBonus\|GetTotalAttackSpeed\|GetTotalCritBonus" \
  pkg/ cmd/ --include="*.go" | grep -v "_test.go" | grep -v "equipment_set_bonus"
# Returns no output
```

- **Incorrect Behavior**: Wearing any combination of set items (e.g. the
  "Inferno" 2-piece or 4-piece set) grants no stat benefit whatsoever. The
  system faithfully tracks set membership but the computed bonuses are stored
  in a component that nothing ever reads.

- **Root Cause**: Integration step 5 of the six-link chain ("Output → Consumer")
  is missing. The system's output is computed correctly; no downstream consumer
  exists.

- **Impact**: The entire equipment set system is non-functional. Players who
  collect set items for their bonuses receive no reward.

- **Remediation**: Options (in order of invasiveness):
  1. **In `CombatSystem.calculateDamage`**: retrieve
     `EquipmentSetBonusComponent` from the attacker entity and add
     `GetTotalDamageBonus()` to `baseDamage`.
  2. **New applicator system**: Register an `EquipmentSetBonusApplicatorSystem`
     after `EquipmentSetBonusSystem` in the update order. When the component's
     `Dirty` flag is set, subtract old bonuses from stats and apply new ones.
  3. **Inline in `InventorySystem`**: When equipment changes fire,
     read set bonuses and propagate to `StatsComponent`.

  Option 2 is recommended for clean separation of concerns.

- **Dependencies**: None.

- **Effort**: medium

---

## G35 — Minimum Damage Floor Applied Before Shield Absorption

- **Category**: Combat Formula Bug
- **Status**: 🔴 OPEN
- **Severity**: MEDIUM

**Evidence** (`pkg/engine/combat_system.go:569–574`):
```go
finalDamage := damageAfterResist
if finalDamage < 1.0 {
    finalDamage = 1.0          // floor at lines 570–572
}
finalDamage = s.applyShieldAbsorption(target, finalDamage)  // line 574
```
The 1.0 floor at lines 570–572 fires **before** shield absorption at line 574.
`applyShieldAbsorption` subtracts up to `shield.AbsorbAmount` from
`finalDamage` and can return 0.0. However, because `finalDamage` was already
floored to 1.0, a shield that fully absorbs ≤ 1 damage still leaks 1 point per
hit.

- **Incorrect Behavior**: A target with a charged shield (`AbsorbAmount = 5`,
  incoming `damageAfterResist = 0.3`) should take 0 damage (shield absorbs all).
  Instead: `0.3 → floored to 1.0 → shield absorbs 1.0 → finalDamage = 0`.
  If `AbsorbAmount = 0.8` then: `0.3 → floored to 1.0 → shield absorbs 0.8 →
  finalDamage = 0.2 → target takes 0.2 damage`. In both cases more damage leaks
  through than the attacker's raw damage warrants.

  More practically: any attack that deals 0.x damage after resistance
  (resistances > ~0.9 and moderate defense) will always deal 1 damage regardless
  of shield because the floor clamps before shield can act.

- **Root Cause**: The 1.0 floor is placed two lines before the shield call. This
  ordering made sense before shields were added but was not re-evaluated when
  `applyShieldAbsorption` was introduced.

- **Impact**: Shield mechanics do not function as intended against highly-resisted
  attacks. Affects any target that both has resistances and equips a shield.

- **Remediation**: Move the floor after shield absorption, or apply it only to
  `damageAfterResist` values that are positive but below 1.0 (which guards
  against floating-point underflow, not intentional full blocks):
  ```go
  finalDamage = s.applyShieldAbsorption(target, damageAfterResist)
  if finalDamage > 0 && finalDamage < 1.0 {
      finalDamage = 1.0
  }
  ```

- **Dependencies**: None.

- **Effort**: small

---

## G36 — `CancelCast` Does Not Apply Cooldown to Interrupted Slot

- **Category**: Spell/Mana Bug
- **Status**: 🔴 OPEN
- **Severity**: MEDIUM

**Evidence** (`pkg/engine/spell_casting.go:2697–2741`):
```go
func (s *SpellCastingSystem) CancelCast(entity *Entity) {
    ...
    if slots.IsCasting() {
        // Only records the cancel, no cooldown applied
        slots.Casting    = -1
        slots.CastingBar = 0
        // slots.Cooldowns[slotIdx] is never written
    }
}
```
In `completeCast` (the success path), the cooldown is set:
```go
slots.Cooldowns[slots.Casting] = spell.Stats.Cooldown
```
`CancelCast` does not reproduce this write.

- **Incorrect Behavior**: A player initiates a 3-second cast, advances the bar
  to 2.9 seconds (97% progress), and then cancels. `slots.Cooldowns[slotIndex]`
  remains 0. The player can immediately start the same cast again. This cycle
  can be repeated indefinitely, creating 100% cast uptime and allowing
  continuous cast-animation pressure without any spell economy cost.

- **Root Cause**: Cooldown assignment is only in the success path
  (`completeCast`). The cancel path was not updated to apply a proportional
  penalty.

- **Impact**: Spell economy is broken. Spells with long cast times and high
  cooldowns are effectively free to "attempt" repeatedly, removing the
  risk/reward trade-off of slow-cast spells.

- **Remediation**:
  ```go
  if slots.IsCasting() {
      slotIdx := slots.Casting
      spell   := slots.GetSlot(slotIdx)
      if spell != nil && slots.CastingBar > 0 {
          // Partial cooldown proportional to how far the cast progressed
          slots.Cooldowns[slotIdx] = spell.Stats.Cooldown * slots.CastingBar
      }
      slots.Casting    = -1
      slots.CastingBar = 0
  }
  ```
  A simpler alternative: apply a fixed interrupt penalty (e.g. 25% of full
  cooldown) regardless of progress to avoid encouraging fast-cancel optimisation.

- **Dependencies**: None.

- **Effort**: small

---

## G37 — `ClassAffinitySystem` Mana Regen Removal Uses Stale `mana.Max`

- **Category**: Stat Removal Precision Bug
- **Status**: 🔴 OPEN
- **Severity**: MEDIUM

**Evidence** (`pkg/engine/class_affinity_system.go:157–163`):
```go
if oldLevel, exists := comp.BonusesApplied[affinityType]; exists {
    oldBonuses := GetAffinityBonuses(affinityType, oldLevel)
    // Removal uses CURRENT mana.Max (line 160)
    mana.Regen -= oldBonuses.ManaRegenBonus * float64(mana.Max) * effectiveness
}
// Application also uses CURRENT mana.Max (line 163)
mana.Regen += bonuses.ManaRegenBonus * float64(mana.Max) * effectiveness
```

Both the removal and the new application use the same current `mana.Max`. If
`mana.Max` at removal equals `mana.Max` at the original application, the
arithmetic is correct. But `mana.Max` is modified by many other systems
(talents, items, buffs) and will routinely differ.

**Worked example:**
- T1: `mana.Max = 100`, mana affinity level 1 applied:
  `mana.Regen += 0.05 × 100 × 1.0 = 5.0`
- T2: Talent increases `mana.Max` to 150.
- T3: Affinity level 1 → level 2 upgrade fires:
  - Removal: `mana.Regen -= 0.05 × 150 × 1.0 = 7.5` (only 5.0 was added at T1)
  - Net removal: −7.5 instead of −5.0 → permanent −2.5 mana regen drain
  - Application: `mana.Regen += 0.08 × 150 × 1.0 = 12.0`
  - Expected net (if removal was correct): +7.0; actual: +4.5

- **Incorrect Behavior**: Any affinity level-up that occurs after `mana.Max`
  increased causes a permanent mana regen shortfall (drain). If `mana.Max`
  decreased, the shortfall is inverted (regen bonus is overstated).

- **Root Cause**: The contribution of the bonus was computed relative to
  `mana.Max` at application time, but the removal computation uses the current
  `mana.Max`, which may differ. The absolute regen value added was never stored.

- **Impact**: Mana regeneration drifts from the intended value every time a
  player levels up an affinity after gaining or losing max mana from any source.
  In a typical session (multiple affinity upgrades, multiple talent purchases)
  the cumulative drift can render mana regen negligible or extremely overpowered.

- **Remediation**: Store the absolute regen value that was applied in
  `ClassAffinityComponent.AppliedManaRegen map[AffinityType]float64`. Replace
  the current removal formula with a lookup into this map:
  ```go
  mana.Regen -= comp.AppliedManaRegen[affinityType]  // exact previously-applied value
  newRegen := bonuses.ManaRegenBonus * float64(mana.Max) * effectiveness
  mana.Regen += newRegen
  comp.AppliedManaRegen[affinityType] = newRegen
  ```

- **Dependencies**: `ClassAffinityComponent` needs a new `AppliedManaRegen`
  map field. Existing save-files may not have this field; zero-value on load
  is safe (first level-up will apply correctly, subsequent level-ups will
  track correctly).

- **Effort**: small

---

## G38 — Render Interpolation Guard False-Positive at World Origin

- **Category**: Render Bug
- **Status**: 🔴 OPEN
- **Severity**: LOW

**Evidence** (`pkg/engine/render_system.go:331`):
```go
func (r *EbitenRenderSystem) interpolatePosition(pos *PositionComponent) (float64, float64) {
    if r.renderAlpha >= 1.0 || (pos.PrevX == 0 && pos.PrevY == 0 && (pos.X != 0 || pos.Y != 0)) {
        return r.cameraSystem.WorldToScreen(pos.X, pos.Y)
    }
    interpX := pos.PrevX + (pos.X-pos.PrevX)*r.renderAlpha
    interpY := pos.PrevY + (pos.Y-pos.PrevY)*r.renderAlpha
    return r.cameraSystem.WorldToScreenInterpolated(interpX, interpY, r.renderAlpha)
}
```
The guard `(pos.PrevX == 0 && pos.PrevY == 0 && (pos.X != 0 || pos.Y != 0))`
is designed to skip interpolation when `PrevX`/`PrevY` are uninitialized
(default zero) so that a newly spawned entity does not blend from the world
origin to its real position. However it also fires for entities that are
genuinely located at pixel position (0,0) — i.e., the top-left tile of the
world — and begin moving in their first rendered frame.

`MovementSystem.Update` stores `pos.PrevX = pos.X; pos.PrevY = pos.Y` before
applying velocity. For a fresh entity at (0,0) in the same frame as its first
velocity tick: `PrevX = 0, PrevY = 0, X = newX`. The guard triggers, skipping
interpolation and snapping the entity directly to `newX, newY`.

- **Incorrect Behavior**: An entity spawned at world origin (0,0) that moves in
  its first simulated frame renders without position interpolation for exactly
  one frame, causing a single-frame visual snap from (0,0) to the new position
  instead of a smooth blend.

- **Root Cause**: The "uninitialized" state is inferred from a zero-value check
  rather than an explicit flag. The world origin is a legitimate position, not
  the same thing as "never had a physics tick."

- **Impact**: Low — visually observable only when entities spawn at or near
  (0,0), which is the world origin. Starting room spawns (first room, first
  corridor intersection) at `(room.X + room.Width/2) * 32` are almost never at
  literal pixel (0,0). The issue surfaces primarily in server test harnesses
  that spawn entities without a terrain context (default `spawnX = 400.0`,
  `spawnY = 300.0` — not affected).

- **Remediation**: Add `Initialized bool` to `PositionComponent`. Set it to
  `true` in `MovementSystem.Update` on first tick (or in `AddComponent`). Guard
  becomes:
  ```go
  if r.renderAlpha >= 1.0 || !pos.Initialized {
      return r.cameraSystem.WorldToScreen(pos.X, pos.Y)
  }
  ```

- **Dependencies**: `PositionComponent` struct change; `MovementSystem.Update`
  must set `Initialized = true` before copying `PrevX`/`PrevY`.

- **Effort**: small

---

## Prior Gap Compatibility Table

| Prior ID | Title | Current Status |
|----------|-------|---------------|
| G1–G16   | Various (see Rev 2 GAPS.md) | ✅ All resolved |
| G17      | WebRTC federation simulated | ✅ Resolved (pion/webrtc WASM) |
| G18      | ClassProgressionSystem no-op | ⚠️ By design; documented |
| G19      | Companion scout velocity | ⚠️ Partial (4-dir cycle, not pathfinding) |
| G20      | Ambush node random offset | ⚠️ Partial (fallback, not cover-based) |
| G21      | Mobile input Type() collision | Verify separately |
| G22      | XP double-award on kill | ✅ Resolved |
| G23      | TalentSystem stat accumulation | ✅ Resolved |
| G24      | HUD no mana bar | ✅ Resolved |
| G25      | consumeItem heals by Defense | ✅ Resolved |
| G26      | AttributeEffects not applied | ✅ Resolved |
| G27      | HUD health bar overflow | ✅ Resolved |
| G28      | Death callback fires every frame | ✅ Resolved |
| G29      | Streak decay hardcoded time=0 | ✅ Resolved |
| G30      | Self-damage guard missing | ✅ Resolved |
| G31      | CarryOverSystem not in system_init | ⚠️ By design (client-only) |
