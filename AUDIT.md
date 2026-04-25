# BUG AND GAP AUDIT — 2026-04-24

> **Audit revision**: Rev 6 — fresh deep behavioral-correctness pass
> (2026-04-24). Supersedes Rev 5 (2026-04-24, same date, earlier in day).
> All Rev 5 findings (G32–G38) were re-verified against the current source
> tree and are confirmed resolved. One newly-discovered finding (G39) is
> reported here.
>
> **Legacy ID compatibility**: G1–G38 IDs remain reserved; code comments
> and docs that reference G1–G38 (e.g. `combat_system.go:577`,
> `class_affinity_system.go:157`, `render_system.go:328`) continue to map
> to the historical findings, all now resolved.
>
> **Status legend**: ✅ RESOLVED | ⚠️ PARTIAL | 🔴 OPEN

---

## Audit Scope

This audit targets **behavioral correctness bugs** — logic errors that
produce wrong results at runtime, silent state corruption that
accumulates across frames, concurrency hazards, and precision errors.
Structural gaps (stubs, TODOs, unwired components) are only reported if
newly discovered. The full Phase-3 checklist from the audit playbook was
exercised; categories with no new defects are explicitly listed under
"Categories Audited Clean".

---

## Build Status

`go build ./...` cannot complete in the CI runner because the runner
lacks X11 headers (`X11/Xlib.h`) required by Ebiten's GLFW C bindings.
This is a runner-environment limitation, not a code defect; the project
CI uses `xvfb-run` and produces clean builds on `ubuntu-latest`
(`.github/workflows/test.yml:22-57`). `go test ./pkg/combat/...` runs
successfully without a display.

---

## Summary

| Category                          | Count | Critical | High | Medium | Low |
|-----------------------------------|-------|----------|------|--------|-----|
| Per-Frame Stat Mutation           | 0     | 0        | 0    | 0      | 0   |
| System Output Never Used          | 0     | 0        | 0    | 0      | 0   |
| Combat Formula Bugs               | 0     | 0        | 0    | 0      | 0   |
| Spell/Mana Bugs                   | 1     | 0        | 0    | 1      | 0   |
| Server Concurrency                | 0     | 0        | 0    | 0      | 0   |
| ECS Cache Bugs                    | 0     | 0        | 0    | 0      | 0   |
| Input Bugs                        | 0     | 0        | 0    | 0      | 0   |
| UI Display Bugs                   | 0     | 0        | 0    | 0      | 0   |
| System Registration Bugs          | 0     | 0        | 0    | 0      | 0   |
| New Structural Gaps               | 0     | 0        | 0    | 0      | 0   |
| **Total (G39)**                   | **1** | **0**    | **0**| **1**  | **0**|

---

## Findings

### MEDIUM

- [x] **[G39] `completeCast` Always Applies Cooldown Even When `executeCast` Silently No-Ops**
  — `pkg/engine/spell_casting.go:288–303`,
    `pkg/engine/spell_casting.go:305–321`,
    `pkg/engine/spell_casting.go:336–350` —

  **Evidence:**
  ```go
  // completeCast (lines 288–303)
  func (s *SpellCastingSystem) completeCast(entity *Entity, slots *SpellSlotComponent, spell *magic.Spell) {
      ...
      s.executeCast(entity, spell, slots.Casting)        // line 299 — may silently return
      slots.Cooldowns[slots.Casting] = spell.Stats.Cooldown  // line 300 — ALWAYS sets cooldown
      slots.Casting = -1
      slots.CastingBar = 0
  }

  // executeCast (lines 305–321)
  func (s *SpellCastingSystem) executeCast(caster *Entity, spell *magic.Spell, slotIndex int) {
      s.logSpellExecution(caster, spell, slotIndex)
      mana := s.validateAndConsumeMana(caster, spell)
      if mana == nil {
          return            // ← spell vanishes, but caller still applies cooldown
      }
      ...
  }

  // validateAndConsumeMana (lines 336–350)
  func (s *SpellCastingSystem) validateAndConsumeMana(caster *Entity, spell *magic.Spell) *ManaComponent {
      mana := s.getManaComponent(caster, spell.Name)
      if mana == nil { return nil }                     // missing component → nil
      if !s.hasEnoughMana(mana, spell) { return nil }   // drained mid-cast → nil
      s.consumeMana(caster.ID, mana, spell.Stats.ManaCost)
      return mana
  }
  ```

  `executeCast` returns early (no spell projectile, no effect, no damage
  application) when:
  1. The mana component is missing entirely (e.g. an entity that lost
     its mana component between `StartCast` and the cast-bar reaching
     1.0), or
  2. Mana was drained mid-cast by another effect (debuff, sap, channel
     drain) so `mana.Current < spell.Stats.ManaCost` at completion.

  In both cases `completeCast` unconditionally writes
  `slots.Cooldowns[slots.Casting] = spell.Stats.Cooldown` at line 300
  immediately after the silent `executeCast` call. A "Not enough mana!"
  notification fires from `hasEnoughMana` (line 397), so the player is
  informed, but the spell slot still goes on its full cooldown despite
  no effect being produced.

  **Incorrect behavior:** A Mage starts a 3-second Fireball cast
  (cooldown 8s, cost 50 mana) at 60 mana. Mid-cast, an enemy applies a
  mana-drain debuff that reduces them to 30 mana. At cast completion:
  - No fireball is produced.
  - No mana is consumed.
  - Slot 0 cooldown is set to 8 seconds.

  The player loses all access to that spell for 8 seconds despite
  paying no mana cost and receiving no benefit. Repeated mana-drain in
  PvP makes spell rotations unrecoverable.

  **Expected behavior:** Cooldown should only be applied when the spell
  actually executed. If `executeCast` silently fails, the slot should
  return to ready (or, at most, a small "GCD" cooldown analogous to the
  partial-cooldown applied by `CancelCast` G36 fix at line 2738).

  **Remediation:** Have `executeCast` (or `validateAndConsumeMana`)
  signal success to the caller. Apply cooldown only on success:
  ```go
  func (s *SpellCastingSystem) completeCast(entity *Entity, slots *SpellSlotComponent, spell *magic.Spell) {
      slotIdx := slots.Casting
      slots.Casting = -1
      slots.CastingBar = 0
      if s.executeCast(entity, spell, slotIdx) {
          slots.Cooldowns[slotIdx] = spell.Stats.Cooldown
      }
  }

  // executeCast returns true on full execution.
  func (s *SpellCastingSystem) executeCast(caster *Entity, spell *magic.Spell, slotIndex int) bool {
      mana := s.validateAndConsumeMana(caster, spell)
      if mana == nil { return false }
      ...
      return true
  }
  ```

  **Validation:** Start a cast with 50/50 mana. Drain caster to 0 mana
  before `CastingBar >= 1.0`. After `Update` completes, assert
  `slots.Cooldowns[castSlot] == 0` (no cooldown applied) and
  `slots.IsCasting() == false`.

---

## Prior Findings Status (G32–G38) — All Confirmed Resolved

Each Rev-5 finding was re-verified against the current source. Evidence
lines below cite the *current* code that demonstrates the fix.

| ID  | Title                                                                              | Status | Evidence in current code                                                                                               |
|-----|------------------------------------------------------------------------------------|--------|------------------------------------------------------------------------------------------------------------------------|
| G32 | `AdvancedClassSystem` adds stat bonuses every frame without guard                  | ✅      | `pkg/engine/advanced_class_system.go:14,68,72,87–88,106–110,125–135` — `lastApplied` map; prev subtracted before re-add |
| G33 | `StatusEffectCriticalChanceSystem` permanently corrupts `CritChance`                | ✅      | `pkg/engine/status_effect_crit_chance_system.go:42,72–87,105–112` — `prevCache` swapped each frame; prev subtracted     |
| G34 | `EquipmentSetBonusSystem` computes bonuses but nothing consumes them               | ✅      | `pkg/engine/combat_system.go:562–565,580,599,618` — set damage/defense bonuses now applied in `computeFinalDamage`     |
| G35 | Minimum damage floor applied before shield absorption                              | ✅      | `pkg/engine/combat_system.go:577–583` — shield absorption now precedes the floor; floor only fires when `>0`           |
| G36 | `CancelCast` does not apply cooldown to interrupted slot                           | ✅      | `pkg/engine/spell_casting.go:2734–2738` — partial cooldown `spell.Stats.Cooldown * castProgress` applied                |
| G37 | `ClassAffinitySystem` mana regen removal uses current `mana.Max` (drift on level-up)| ✅      | `pkg/engine/class_affinity_system.go:157–169` — `AppliedManaRegen` map stores absolute regen; that exact value removed |
| G38 | Render interpolation guard snaps first-frame movement at world origin              | ✅      | `pkg/engine/render_system.go:328–333` — guard now uses `pos.Initialized` instead of (PrevX==0 && PrevY==0) heuristic   |

---

## Categories Audited Clean (no defects found this revision)

The following categories from the Phase-3 checklist were investigated
and yielded no new findings beyond what is already documented as
resolved. Citations below show the specific code paths verified.

### Per-Frame Stat Mutation (3a)

- `AdvancedClassSystem` — `pkg/engine/advanced_class_system.go:67–73` —
  guarded by `lastApplied` map and prev-subtraction (G32 fix).
- `TalentSystem` — `pkg/engine/talent_system.go:77` — `removeTalentBonuses`
  called before `applyBonuses` (G23 fix).
- `AttributeAllocationSystem` — uses `removeAppliedBonuses` →
  `applyAttributeBonuses` pattern (G26).
- `ClassAffinitySystem` — uses `BonusesApplied` map for delta-tracked
  remove/apply, with absolute mana-regen storage (G37 fix).
- `EquipmentSetBonusSystem` — Dirty hash detects equipment changes; set
  bonus values are now consumed in `CombatSystem.computeFinalDamage`
  (G34 fix).
- `StatusEffectCriticalChanceSystem` — uses swapped `critCache` /
  `prevCache` (G33 fix).
- `TerrainAmbushCritSystem` — verified delta-tracked via lines 133–134
  (cited in Rev-5 false-positives table; pattern still valid).

### Combat Formula Correctness (3b)

- Minimum damage floor is now applied **after** shield absorption and
  only when `finalDamage > 0`, preserving the full-block semantic
  (`pkg/engine/combat_system.go:577–583`, G35 fix).
- `DefaultCombatResolver.CalculateDamage` continues to enforce
  `minDamage = base * MinDamageMultiplier` (default 0.1). The package
  documentation describes this as deliberate balance protection — there
  is no documented "immunity at resistance == 1.0" semantic. Carried
  over from Rev-5 false-positives table; not reopened.
- Critical-hit and evasion callbacks fire with the correct values; no
  new defects observed.

### Spell and Mana Correctness (3c)

- Mana check **at cast start** is correctly performed by
  `checkManaAvailability` at `pkg/engine/spell_casting.go:2584` (the
  Rev-4 G15 fix is preserved).
- `CancelCast` applies a partial cooldown proportional to cast progress
  (`spell_casting.go:2734–2738`, G36 fix).
- See **G39** above for the newly-discovered cooldown-on-silent-fail
  bug at the cast **completion** path.

### Server Concurrency (3d)

- `running` is consistently accessed under `clientsMu` —
  `pkg/network/server.go:218–231,280–287,355–357`. No reads outside the
  lock observed.
- No `RLock` → `Lock` upgrade pattern detected in `clientsMu` paths.
- `disconnectAllClients` (`server.go:299–308`) holds `clientsMu.Lock()`
  then calls `client.disconnect()` which acquires `c.mu.Lock()` — lock
  order is `clientsMu → c.mu`. No reverse-order paths found:
  `client.disconnect()` and `client.sendStateUpdate()` never
  re-acquire `clientsMu`.
- Channel-send safety is preserved by non-blocking `select-default`
  patterns at the input/error send sites; the existing test
  `pkg/network/...` race coverage exercises shutdown paths.

### ECS Component Cache Consistency (3e)

- `Entity.AddComponent` calls `updateComponentCache`, which assigns the
  same pointer that the `Components` map holds. Replacing a component
  via `AddComponent` updates the cached field in lock-step.
- No duplicate `Type()` strings observed across the engine component
  set during this pass.

### Input and UI Correctness (3f)

- Mobile dual-joystick dead-zone is applied **before** normalization at
  `pkg/mobile/dual_joystick.go:281–305`. Dead zone correctly produces
  `Magnitude = 0` when `distance < DeadZone`.
- Modal input suppression is comprehensive in
  `pkg/engine/game.go:1326–1342` (12 modal panels checked); the world
  is paused while any modal panel is visible.
- ESC handler dismisses panels in priority order with explicit
  `return`s after each successful dismissal (verified manually in
  game.go input dispatch).

### UI Display Bugs (3g)

- HUD draws both health bar (`hud_system.go:105–151`) and mana bar
  (`hud_system.go:155+`, G24 fix).
- Health bar fill is clamped to `[0, 1]` at lines 132–137 (G27 fix);
  mana bar uses the same pattern.
- Render interpolation uses `pos.Initialized` (G38 fix).
- `consumeItem` correctly heals by `item.Stats.Healing`, not
  `item.Stats.Defense` (`cmd/server/player_management.go:300`, G25
  fix).

### Procedural Generation Determinism (3i)

- `rngSourcePool` and `DefaultAmbienceConfig` were considered. The
  ambience-particle case is intentional (deterministic seeds for
  identical scenes are desirable for replay/testing); not a bug.

### System Registration Ordering (3j)

- `AdvancedClassSystem` is registered in `cmd/client/handlers.go:2162`
  via `game.World.AddSystem(sys.advancedClassSystem)` after class and
  talent systems. Stat-bonus delta tracking (G32 fix) makes the
  registration order non-critical for correctness.
- `EquipmentSetBonusSystem` is consumed by `CombatSystem` lazily via
  the `equipment_set_bonus` component lookup, not by registration
  order.
- No newly-introduced ordering bugs detected.

---

## False Positives Considered and Rejected

| Candidate                                                                                       | Evidence Examined                                                                          | Reason Rejected                                                                                                                                                |
|-------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `AdvancedClassSystem` per-frame mutation (G32)                                                  | `pkg/engine/advanced_class_system.go:67–73,87–88,106–110,125–135`                          | `lastApplied` map and prev-subtraction now in place; bonuses are net (new − prev) per call.                                                                    |
| `StatusEffectCriticalChanceSystem` permanent corruption (G33)                                   | `pkg/engine/status_effect_crit_chance_system.go:72–87`                                     | `critCache`/`prevCache` swap each frame; previous modifier subtracted before adding new one; expiry path correctly subtracts last contribution.                |
| `EquipmentSetBonusSystem` output never used (G34)                                               | `pkg/engine/combat_system.go:562–599,618`                                                  | Set bonuses are now read by `CombatSystem` for both attacker damage and target defense.                                                                        |
| Minimum damage floor before shield (G35)                                                        | `pkg/engine/combat_system.go:577–583`                                                      | Floor is now after shield absorption and only when `finalDamage > 0`; full-block returns 0.                                                                    |
| `CancelCast` no cooldown (G36)                                                                  | `pkg/engine/spell_casting.go:2734–2738`                                                    | Partial cooldown `spell.Stats.Cooldown * castProgress` is applied on interrupt.                                                                                |
| `ClassAffinitySystem` mana regen drift (G37)                                                    | `pkg/engine/class_affinity_system.go:157–169`                                              | Absolute regen value is now stored in `AppliedManaRegen[affinityType]` and that exact value is subtracted on level-up.                                         |
| Render interpolation snap at origin (G38)                                                       | `pkg/engine/render_system.go:328–333`                                                      | Guard now keys on `pos.Initialized` rather than the (PrevX==0 && PrevY==0) heuristic; entities at origin interpolate correctly.                                |
| `consumeItem` heals by Defense (G25)                                                            | `cmd/server/player_management.go:300`                                                      | Uses `item.Stats.Healing`.                                                                                                                                     |
| `DefaultCombatResolver` 10% floor breaks immunity                                               | `pkg/combat/resolver.go` (resolver formula and `MinDamageMultiplier` field)                | No documented immunity-at-resistance==1.0 contract; the `MinDamageMultiplier` is the documented floor, intentional.                                            |
| `running` data race on `TCPServer`                                                              | `pkg/network/server.go:221,231,283,287,355–357`                                            | All five accesses are inside `clientsMu` Lock/Unlock; no unprotected read or write.                                                                            |
| `clientsMu` deadlock via `client.mu`                                                            | `pkg/network/server.go:299–308 (Lock→c.mu.Lock); 398–408 (RLock→c.mu.RLock)`               | Lock order is consistent (clientsMu → c.mu); no reverse-order acquisitions; RLock paths only RLock c.mu.                                                       |
| Mobile dead-zone applied after normalization                                                    | `pkg/mobile/dual_joystick.go:278–286`                                                      | Dead zone is checked against raw `distance` before any normalization; correct order.                                                                           |
| Modal input not suppressed                                                                      | `pkg/engine/game.go:1326–1342`                                                             | All 12 modal UIs checked in `anyUIOpen`; world simulation paused via `shouldUpdateWorld`.                                                                      |
| HUD missing mana bar                                                                            | `pkg/engine/hud_system.go:88,155`                                                          | `drawManaBar()` is called from `Draw()` and renders below the health bar (G24 fix).                                                                            |

---

## Legacy Gap ID Compatibility

- **G1–G16**: Documented in pre-Rev-2 history; all resolved.
- **G17–G31**: Documented in Rev 4 (2026-04-25); all confirmed resolved
  in Rev 5 verification table.
- **G32–G38**: Introduced and resolved in Rev 5 (2026-04-24);
  re-verified in this revision (Rev 6, 2026-04-24).
- **G39**: New finding in this revision — see the **MEDIUM** section above.
