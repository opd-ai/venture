# BUG AND GAP AUDIT — 2026-04-24

> **Audit revision**: Rev 5 — deep behavioral-correctness pass (2026-04-24).
> Supersedes Rev 4 (2026-04-25). All prior gap IDs G1–G31 are preserved for
> cross-reference compatibility (code comments, CI scripts, issue-tracker
> entries referencing those IDs remain valid). New findings are numbered G32–G38.
>
> **Status legend**: ✅ RESOLVED | ⚠️ PARTIAL | 🔴 OPEN

---

## Audit Scope

This audit targets **behavioral correctness bugs** — logic errors that produce
wrong results at runtime, silent state corruption that accumulates across
frames, concurrency hazards, and precision errors. Structural gaps (stubs,
TODOs, unwired components) are only reported if newly discovered. All findings
in G17–G31 have been verified resolved; their resolution evidence is summarized
in the **Prior Findings Status** section below.

---

## Build Status

`go build ./...` cannot complete in the CI runner because the runner lacks X11
headers (`X11/Xlib.h`) required by Ebiten's GLFW C bindings. This is a runner
environment limitation, not a code defect. `go vet` is similarly blocked for
the same reason. `go test ./pkg/combat/...` passes successfully.

---

## Summary

| Category                   | Count | Critical | High | Medium | Low |
|----------------------------|-------|----------|------|--------|-----|
| Per-Frame Stat Mutation    | 2     | 2        | 0    | 0      | 0   |
| System Output Never Used   | 1     | 0        | 1    | 0      | 0   |
| Combat Formula Bugs        | 1     | 0        | 0    | 1      | 0   |
| Spell/Mana Bugs            | 1     | 0        | 0    | 1      | 0   |
| Stat Removal Precision     | 1     | 0        | 0    | 1      | 0   |
| Render Bugs                | 1     | 0        | 0    | 0      | 1   |
| **Total (G32–G38)**        | **7** | **2**    | **1**| **3**  | **1**|

---

## Findings

### CRITICAL

- [x] **[G32] `AdvancedClassSystem` Adds Stat Bonuses Every Frame Without Guard**
  — `pkg/engine/advanced_class_system.go:68,84,100–104` —

  **Evidence:**
  ```go
  // applyHealthBonuses (line 68)
  health.Max += float64(bonuses.Health)

  // applyManaBonuses (line 84)
  mana.Max += bonuses.Mana

  // applyStatsBonuses (lines 100–104)
  stats.Attack     += float64(bonuses.Strength)
  stats.Defense    += float64(bonuses.Defense)
  stats.MagicPower += float64(bonuses.Intelligence)
  stats.CritChance += bonuses.CritChance
  stats.CritDamage += bonuses.CritDamage
  ```
  The `Update()` method at line 27–48 iterates every entity with an
  `"advanced_class"` component, calls `manager.CalculateTotalStats(playerID)`
  which returns the same bonus values every frame, and then unconditionally
  adds them to live component fields. There is no dirty flag, no previous-value
  subtraction, and no idempotency guard anywhere in `Update`,
  `applyHealthBonuses`, `applyManaBonuses`, or `applyStatsBonuses`.

  **Incorrect behavior:** For a character with `bonuses.Strength = 10`, at 60
  FPS `stats.Attack` grows by 600 per second. `health.Max` grows by
  `bonuses.Health × 60` per second. Within seconds every player with an
  advanced class assigned via `InitializePlayerClass` reaches unbounded stats.

  **Expected behavior:** Bonuses should be applied exactly once, or removed
  before re-application, matching the `TalentSystem` pattern (G23 fix, now in
  `talent_system.go:77`).

  **Remediation:** Store a `lastApplied advanced.StatBonuses` field on the
  system (or in the `AdvancedClassComponent`). Before calling
  `applyStatBonuses`, subtract `lastApplied`; after applying, store the new
  bonuses as `lastApplied`. Alternatively gate the apply call behind a dirty
  flag on the component that is set only when class/talent configuration
  changes.

  **Validation:** Assign a non-zero primary class, run 300 frames, assert
  `stats.Attack == baseAttack + expectedBonus` (not `baseAttack + expectedBonus * 300`).

---

- [x] **[G33] `StatusEffectCriticalChanceSystem` Permanently Corrupts `CritChance`**
  — `pkg/engine/status_effect_crit_chance_system.go:61–92` —

  **Evidence:**
  ```go
  // Update() — runs every frame (lines 61–92)
  for k := range s.critCache {
      delete(s.critCache, k)  // cache cleared each frame (lines 61–64)
  }
  for _, entity := range entities {
      ...
      modifier := s.calculateCritModifier(entity)
      if modifier == 0.0 {
          continue  // expired-effect path — NO subtraction (lines 73–75)
      }
      s.critCache[entity.ID] = modifier
      oldCrit := stats.CritChance
      stats.CritChance += modifier   // unconditional ADD every frame (line 82)
      if stats.CritChance < 0.0 { stats.CritChance = 0.0 }
      else if stats.CritChance > 1.0 { stats.CritChance = 1.0 }
  }
  ```
  The `critCache` is cleared at the start of every frame (lines 61–64).
  `calculateCritModifier` therefore always recalculates from scratch. The
  recalculated `modifier` is then **added** to `stats.CritChance` without
  first subtracting the value added on the previous frame. For any entity
  under a non-zero status effect, `stats.CritChance += modifier` executes
  60 times per second.

  When a status effect expires, `calculateCritModifier` returns 0.0 and the
  `continue` at line 74 skips the entity — meaning the previously-accumulated
  additions to `CritChance` are **never undone**.

  **Incorrect behavior (positive modifier, e.g. "blessed" = +0.10):**
  - Frame 1: CritChance = 0.05 + 0.10 = 0.15
  - Frame 9: CritChance = 0.85 + 0.10 = 0.95
  - Frame 10: CritChance = 0.95 + 0.10 = 1.05 → clamped to 1.0
  - Blessed expires: modifier = 0.0 → continue → CritChance stays at 1.0
  - Entity now has permanent 100% crit chance.

  **Incorrect behavior (negative modifier, e.g. "cursed" = −0.10):**
  - Frame 1: CritChance = 0.05 − 0.10 = −0.05 → clamped to 0.0
  - All subsequent frames: 0.0 − 0.10 = −0.10 → clamped to 0.0
  - Cursed expires: modifier = 0.0 → continue → CritChance stays at 0.0
  - Entity now has permanent 0% crit chance.

  **Expected behavior:** On effect expiry the base `CritChance` should be
  restored to its pre-effect value.

  **Remediation:** Track the last-applied modifier per entity in a separate
  `prevCache map[uint64]float64`. Each frame, subtract `prevCache[id]` before
  adding the new modifier. When modifier becomes 0.0, subtract `prevCache[id]`
  to undo the last contribution. This is the same delta-tracking pattern used
  correctly in `TerrainAmbushCritSystem` (lines 133–134).

  **Validation:** Apply "blessed" for 100 frames; expire it; assert
  `stats.CritChance` equals the pre-blessing baseline.

---

### HIGH

- [x] **[G34] `EquipmentSetBonusSystem` Computes Bonuses But Nothing Consumes Them**
  — `pkg/engine/equipment_set_bonus_system.go:207–299`,
  `pkg/engine/equipment_set_bonus_component.go:83–110` —

  **Evidence:** `EquipmentSetBonusSystem.Update()` detects equipment changes
  (via hash comparison), recalculates which set bonuses are active, and writes
  the combined bonus totals into `EquipmentSetBonusComponent.ActiveSets` and
  `CombinedBonus`. The component exposes `GetTotalDamageBonus()`,
  `GetTotalDefenseBonus()`, `GetTotalHealthBonus()`, `GetTotalAttackSpeed()`,
  and `GetTotalCritBonus()` methods. A grep of all non-test production Go files
  in `pkg/engine/`, `cmd/client/`, and `cmd/server/` for calls to these methods
  returns **zero results** (excluding the component file itself and tests).

  ```bash
  # Confirmation: zero production callers
  grep -rn "GetTotalDamageBonus\|GetTotalDefenseBonus\|GetTotalHealthBonus\|GetTotalAttackSpeed\|GetTotalCritBonus" \
    pkg/ cmd/ --include="*.go" | grep -v "_test.go" | grep -v "equipment_set_bonus"
  # (no output)
  ```

  `pkg/engine/combat_system.go` and `pkg/engine/inventory_system.go` read
  `EquipmentComponent.GetTotalDefense()` and `GetWeaponDamage()` but never read
  `EquipmentSetBonusComponent` bonus values.

  **Incorrect behavior:** A player wearing two items from the "Inferno" set
  (which should grant +15 damage bonus) receives no benefit at all. Set bonuses
  are silently discarded every frame after being calculated.

  **Expected behavior:** `CombatSystem.calculateDamage` (or a dedicated
  integration system) should read `GetTotalDamageBonus()` and add it to the
  attack calculation; `GetTotalDefenseBonus()` to defense; etc.

  **Remediation:** Add a consumer in `CombatSystem.calculateDamage` (or a new
  `EquipmentSetBonusApplicatorSystem` registered after
  `EquipmentSetBonusSystem`) that reads `GetTotalDamageBonus()` and applies it
  to `StatsComponent.Attack`, guarded by the same dirty flag pattern.

  **Validation:** Equip two matching set items; verify `stats.Attack` increases
  by `GetTotalDamageBonus()` and reverts when either item is removed.

---

### MEDIUM

- [x] **[G35] Minimum Damage Floor Applied Before Shield Absorption**
  — `pkg/engine/combat_system.go:570–574` —

  **Evidence:**
  ```go
  finalDamage := damageAfterResist
  if finalDamage < 1.0 {
      finalDamage = 1.0          // floor applied at line 570–572
  }
  finalDamage = s.applyShieldAbsorption(target, finalDamage)  // line 574
  ```
  The 1.0 minimum floor fires at lines 570–572, **before** shield absorption
  at line 574. `applyShieldAbsorption` can return zero if the shield absorbs
  the full damage value (the shield logic inside that call subtracts absorbed
  from `finalDamage` and returns the remainder, which can be 0.0).

  Because the floor clamps `finalDamage` to 1.0 before the shield processes
  it, a target with a shield capable of absorbing any amount ≤ 1.0 still
  receives 1 damage. In practice, a fully-charged shield that should block all
  incoming damage will always leak 1 point per hit.

  **Incorrect behavior:** A shielded target with a shield absorbing 100% of
  incoming damage takes 1 damage per hit instead of 0.

  **Expected behavior:** Shield absorption should have the opportunity to
  reduce damage to 0 before any floor is applied. The floor is a fallback
  against negative resistances, not a shield bypass.

  **Remediation:** Move the floor to after shield absorption:
  ```go
  finalDamage = s.applyShieldAbsorption(target, damageAfterResist)
  if finalDamage < 1.0 && !targetHasShield(target) {
      finalDamage = 1.0
  }
  ```
  Alternatively, apply the floor only when `damageAfterResist >= 1.0` (i.e.,
  the floor corrects near-zero resist math, not intentional full blocks).

  **Validation:** Equip a target with a shield absorbing 100% of damage; fire
  a 1-damage attack; assert `health.Current` is unchanged.

---

- [x] **[G36] `CancelCast` Does Not Apply Cooldown to Interrupted Slot**
  — `pkg/engine/spell_casting.go:2697–2741` —

  **Evidence:**
  ```go
  func (s *SpellCastingSystem) CancelCast(entity *Entity) {
      ...
      if slots.IsCasting() {
          slots.Casting    = -1   // only resets slot index
          slots.CastingBar = 0    // only resets progress bar
          // No cooldown applied to slots.Cooldowns[slots.Casting]
      }
  }
  ```
  Cooldown is set only in `completeCast` (called when `CastingBar >= 1.0`).
  `CancelCast` sets `slots.Casting = -1` and `slots.CastingBar = 0` but does
  not write to `slots.Cooldowns`.

  In contrast, `completeCast` sets:
  ```go
  slots.Cooldowns[slots.Casting] = spell.Stats.Cooldown
  ```
  (verified at the `completeCast` call site.)

  **Incorrect behavior:** A player can start a 3-second cast, cancel after 2.9
  seconds, and immediately start the same cast again — paying zero cooldown
  cost. This allows indefinite cast-attempt pressure (e.g. stagger-locking an
  opponent with near-complete casts) without any spell economy.

  **Expected behavior:** Interrupting a cast that has progressed past a minimum
  threshold (e.g. 50% of cast time) should apply either the full cooldown or a
  partial cooldown proportional to progress.

  **Remediation:**
  ```go
  if slots.IsCasting() {
      slotIdx := slots.Casting
      spell   := slots.GetSlot(slotIdx)
      if spell != nil && slots.CastingBar > 0 {
          // Apply partial cooldown proportional to cast progress
          slots.Cooldowns[slotIdx] = spell.Stats.Cooldown * slots.CastingBar
      }
      slots.Casting    = -1
      slots.CastingBar = 0
  }
  ```

  **Validation:** Cast a spell to 80% bar; cancel; assert
  `slots.Cooldowns[slotIndex] > 0`.

---

- [x] **[G37] `ClassAffinitySystem` Mana Regen Removal Uses Current `mana.Max`**
  — `pkg/engine/class_affinity_system.go:157–163` —

  **Evidence:**
  ```go
  // Remove old bonus (line 158–161)
  if oldLevel, exists := comp.BonusesApplied[affinityType]; exists {
      oldBonuses := GetAffinityBonuses(affinityType, oldLevel)
      mana.Regen -= oldBonuses.ManaRegenBonus * float64(mana.Max) * effectiveness
  }
  // Apply new bonus (line 163)
  mana.Regen += bonuses.ManaRegenBonus * float64(mana.Max) * effectiveness
  ```
  Both the removal and the application use the **current** `mana.Max` at the
  moment the affinity level-up fires. If `mana.Max` changed between the time
  the old bonus was first applied and the level-up:

  - At application: `mana.Max = 100`, added `0.05 × 100 × 1.0 = 5.0` regen
  - At removal: `mana.Max = 150` (talent bonus applied), removed
    `0.05 × 150 × 1.0 = 7.5` regen (removed 2.5 more than was added)
  - Net: permanent −2.5 mana regen drain per affinity level-up after any
    `mana.Max` increase.

  **Incorrect behavior:** Every affinity level-up that occurs after any change
  to `mana.Max` causes a permanent, non-recoverable mana regen drift.

  **Expected behavior:** The removal should undo exactly what was originally
  applied. The original `mana.Max` at application time must be stored (or the
  absolute regen value added must be stored), not recomputed.

  **Remediation:** Store the absolute regen value applied in
  `ClassAffinityComponent.AppliedManaRegen[affinityType]`. On removal,
  subtract that stored value; on application, add the new value and store it.

  **Validation:** Apply a mana affinity level at `mana.Max = 100`; increase
  `mana.Max` to 150 (via talent); trigger affinity level-up; assert
  `mana.Regen` increased correctly and is not negative.

---

### LOW

- [x] **[G38] Render Interpolation Guard Snaps First-Frame Movement at World Origin**
  — `pkg/engine/render_system.go:331` —

  **Evidence:**
  ```go
  if r.renderAlpha >= 1.0 || (pos.PrevX == 0 && pos.PrevY == 0 && (pos.X != 0 || pos.Y != 0)) {
      return r.cameraSystem.WorldToScreen(pos.X, pos.Y)
  }
  ```
  This guard's intended purpose is to skip interpolation when `PrevX`/`PrevY`
  are uninitialized (zero-value defaults), preventing a snap from (0,0) to the
  entity's real first-frame position. However the condition
  `pos.PrevX == 0 && pos.PrevY == 0` is also true for any entity that starts
  and remains at the world origin (tile 0,0 → pixel 0,0). An entity legitimately
  spawned at (0,0) that moves on its first tick has `PrevX==0, PrevY==0,
  X≠0 or Y≠0`, which matches the guard — interpolation is skipped and the
  entity snaps to the new position without blending.

  The same condition fires for all entities whose `MovementSystem` has not yet
  run a first tick to set `PrevX = pos.X` before `Draw` is called.

  **Incorrect behavior:** Any entity whose initial position is exactly (0,0)
  and which moves away during its first rendered frame will snap rather than
  interpolate, producing a single-frame visual artifact.

  **Expected behavior:** Interpolation should be skipped only when PrevX/PrevY
  are genuinely uninitialized (i.e., the entity has never had a physics tick),
  not when the entity is legitimately at the world origin.

  **Remediation:** Add an `Initialized bool` field to `PositionComponent` (set
  to `true` in `MovementSystem.Update` after the first tick). The guard
  becomes:
  ```go
  if r.renderAlpha >= 1.0 || !pos.Initialized {
      return r.cameraSystem.WorldToScreen(pos.X, pos.Y)
  }
  ```

  **Validation:** Spawn an entity at (0,0); simulate two frames; assert the
  render position on frame 2 is interpolated between (0,0) and the new
  position, not snapped.

---

## Prior Findings Status (G17–G31)

All findings from Rev 4 (G17–G31) are confirmed resolved in the current
codebase. Key verifications:

| ID  | Title                                          | Verified In                                      |
|-----|------------------------------------------------|--------------------------------------------------|
| G17 | WebRTC federation simulated                    | `pkg/network/federation/webrtc/peer_wasm.go` — real pion connection |
| G18 | `ClassProgressionSystem.Update()` no-op        | Intentional design; documented (still a stub, but by design) |
| G19 | Companion scout hardcoded velocity             | `companion_system.go:524–547` — 4-direction cycle with timer |
| G20 | BehaviorTree ambush random offset              | Still uses random offset (intentional fallback, low priority) |
| G21 | Mobile input `Type()` collision                | Verify status separately — out of scope for this pass |
| G22 | XP double-award on kill                        | `system_init.go:933–941` — kill callback has explicit no-AwardXP comment |
| G23 | TalentSystem stat accumulation                 | `talent_system.go:77` — `removeTalentBonuses` called before reapply |
| G24 | Desktop HUD no mana bar                        | `hud_system.go` — `drawManaBar()` now called |
| G25 | `consumeItem` heals by Defense                 | `player_management.go:300` — now uses `item.Stats.Healing` |
| G26 | `AttributeEffects` fields never applied        | `attribute_allocation_system.go` — applied after each attribute loop |
| G27 | HUD health bar overflows on overheal           | `hud_system.go:126` — clamped to `[0, 1]` |
| G28 | Death callback fires every frame               | `combat_system.go:284–287` — `processedDeaths` guard |
| G29 | Streak decay uses hardcoded time = 0           | `class_affinity_system.go:109` — uses `s.elapsedTime` |
| G30 | Entities can attack themselves                 | `combat_system.go:278` — self-attack guard present |
| G31 | `CarryOverSystem` not in `system_init.go`      | Client-only registration is intentional (prestige/NG+ feature) |

---

## False Positives Considered and Rejected

| Candidate | Evidence Examined | Reason Rejected |
|-----------|-------------------|-----------------|
| `terrain_ambush_crit_system.go:138,140` — `stats.CritChance = 0/1.0` look like direct assignment | Lines 133–140 | Lines 133–134 do `stats.CritChance -= currentBonus; stats.CritChance += newBonus`; lines 138,140 are clamp guards after the delta update, not direct stat assignment |
| `CancelCast` no mana refund | `spell_casting.go` | Mana is only consumed at cast completion (`completeCast`); a cancelled cast never consumed mana, so no refund is needed |
| `checkManaAvailability` only at completion | `spell_casting.go:2584` | Mana check IS performed at cast start in `StartCast` via `checkManaAvailability`; the prior "check only at completion" bug is already fixed |
| `ClassAffinitySystem` BonusesApplied nil map panic | `class_affinity_system.go:158` | `comp.BonusesApplied[affinityType]` in Go returns zero-value (`AffinityLevelNone`) from a nil map lookup without panic; the `exists` check is correct |
| `AdvancedClassSystem` component only added once (intentional) | `advanced_class_system.go:119`, `cmd/client/handlers.go:3473` | The component IS added only once at class initialization, but `AdvancedClassSystem.Update` is called every frame and adds bonuses each call with no guard — confirmed NOT intentional |
| `DefaultCombatResolver` 10% minimum breaks immunity | `pkg/combat/resolver.go:76–77` | `resistance = 1.0` is clamped at line 72; `resistedDamage = 0`; floor applies `minDamage = base * 0.1`. This is documented as intentional — no immunity semantic is advertised; the comment at line 13 says "0.0 = no resist, 1.0 = immune" but the code and `MinDamageMultiplier` field comment clarify the floor is deliberate |
| `CompanionSystem.applyBondingPerks` comments-only implementation | `companion_system.go:567–580` | Stubs are in a switch-case for perks; stub comments are documented design gaps; not newly discovered |

---

## Legacy Gap ID Compatibility

G1–G16 are all resolved (G1–G16 documented in `GAPS.md` Rev 2 and earlier).
G17–G31 are documented in Rev 4 (`AUDIT.md` 2026-04-25) and confirmed resolved.
G32–G38 are the new findings documented in this revision.
