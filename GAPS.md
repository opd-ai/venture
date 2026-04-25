# Bug and Gap Details — 2026-04-24 (Rev 6)

> **Rev 6 — fresh deep behavioral-correctness pass (2026-04-24).** This
> file supersedes `GAPS.md` Rev 5 (2026-04-24, earlier). All previously
> documented findings G32–G38 are confirmed resolved against the
> current source tree (see `AUDIT.md` for the per-finding evidence
> table). One newly-discovered defect, **G39**, is documented below.
>
> **Legacy ID compatibility**: G1–G38 IDs remain reserved for
> cross-references in code comments and historical issue trackers; all
> are resolved.
>
> **Status legend**: ✅ RESOLVED | ⚠️ PARTIAL | 🔴 OPEN

---

## G39 — `completeCast` Always Applies Cooldown Even When `executeCast` Silently No-Ops

- **Category**: Spell Bug
- **Status**: 🔴 OPEN
- **Severity**: MEDIUM

**Evidence** (`pkg/engine/spell_casting.go`):

```go
// completeCast — lines 288–303
func (s *SpellCastingSystem) completeCast(entity *Entity, slots *SpellSlotComponent, spell *magic.Spell) {
    if s.logger != nil {
        s.logger.WithFields(logrus.Fields{
            "entity_id":  entity.ID,
            "spell_name": spell.Name,
            "spell_type": spell.Type.String(),
            "slot_index": slots.Casting,
            "mana_cost":  spell.Stats.ManaCost,
        }).Info("Spell cast completed")
    }
    s.executeCast(entity, spell, slots.Casting)        // line 299 — silent no-op possible
    slots.Cooldowns[slots.Casting] = spell.Stats.Cooldown  // line 300 — UNCONDITIONAL
    slots.Casting = -1
    slots.CastingBar = 0
}

// executeCast — lines 305–321
func (s *SpellCastingSystem) executeCast(caster *Entity, spell *magic.Spell, slotIndex int) {
    s.logSpellExecution(caster, spell, slotIndex)
    mana := s.validateAndConsumeMana(caster, spell)
    if mana == nil {
        return        // ← spell vanishes silently
    }
    pos := s.getCasterPosition(caster, spell)
    if pos == nil {
        return        // ← spell vanishes silently
    }
    s.dispatchSpellByType(caster, spell, pos)
    s.applySpellEffects(caster, spell, pos, slotIndex)
}

// validateAndConsumeMana — lines 336–350
func (s *SpellCastingSystem) validateAndConsumeMana(caster *Entity, spell *magic.Spell) *ManaComponent {
    mana := s.getManaComponent(caster, spell.Name)
    if mana == nil { return nil }                     // missing component path
    if !s.hasEnoughMana(mana, spell) { return nil }   // mid-cast drain path
    s.consumeMana(caster.ID, mana, spell.Stats.ManaCost)
    return mana
}
```

- **Incorrect Behavior**:

  Concrete trace (Mage, Fireball, cost=50, cooldown=8s, cast time=3s):

  1. `mana.Current = 60`. Player calls `StartCast(slot=0)`.
     `checkManaAvailability` (line 2584) succeeds (60 ≥ 50).
     `slots.Casting = 0`, `slots.CastingBar` begins advancing.
  2. At T=2.5s an enemy applies a mana-drain debuff that deducts 30
     mana. Now `mana.Current = 30`.
  3. At T=3.0s `slots.CastingBar >= 1.0`, `Update` calls
     `completeCast`.
  4. `completeCast` calls `executeCast` (line 299).
  5. `executeCast` calls `validateAndConsumeMana`. `hasEnoughMana`
     returns false (30 < 50), shows the "Not enough mana!"
     notification (line 397), and returns nil. `executeCast` returns
     immediately at line 311 — no projectile spawned, no spell effect
     applied, no mana consumed.
  6. Control returns to `completeCast` line 300, which writes
     `slots.Cooldowns[0] = 8.0` regardless of the silent failure.
  7. The Mage now has Fireball locked out for 8 seconds despite
     producing zero output.

  The same trace applies if the entity loses its mana component
  between `StartCast` and completion (e.g. polymorph effect,
  component swap). `getManaComponent` returns nil → `executeCast`
  returns early → cooldown is still applied.

- **Root Cause**: `completeCast` writes the cooldown unconditionally
  on the line immediately after the `executeCast` call, with no
  success/failure signal flowing back from `executeCast` or
  `validateAndConsumeMana`. The function treats "the cast bar reached
  100%" as equivalent to "the spell executed", which is incorrect when
  preconditions checked at `StartCast` (mana availability, mana
  component presence) no longer hold at completion.

- **Impact**:
  - **Player-visible**: A spell that produces no effect locks out its
    slot for the full cooldown. In PvP, repeated mana-drain trivially
    denies a caster's entire spellbook with no resource cost to the
    drainer — the drainer pays mana once per drain cast, the caster
    pays the full cooldown of their own spell with no benefit.
  - **Hidden state**: Internal state is consistent (`slots.Casting`
    correctly resets to `-1`, `slots.CastingBar` to 0); the issue is
    purely the spurious cooldown write.

- **Remediation**: Have `executeCast` (and/or
  `validateAndConsumeMana`) return a `bool` indicating whether the
  spell actually executed. Apply cooldown only on success.

  ```go
  func (s *SpellCastingSystem) completeCast(entity *Entity, slots *SpellSlotComponent, spell *magic.Spell) {
      // Snapshot slot before clearing casting state so executeCast
      // cannot observe inconsistent state via side effects.
      slotIdx := slots.Casting
      slots.Casting = -1
      slots.CastingBar = 0

      if s.executeCast(entity, spell, slotIdx) {
          slots.Cooldowns[slotIdx] = spell.Stats.Cooldown
      }
      // Optional: apply a small "wasted attempt" cooldown (e.g. 0.5s GCD)
      // when the cast fizzles, mirroring the partial cooldown applied
      // by CancelCast (G36 fix at line 2738).
  }

  func (s *SpellCastingSystem) executeCast(caster *Entity, spell *magic.Spell, slotIndex int) bool {
      s.logSpellExecution(caster, spell, slotIndex)
      mana := s.validateAndConsumeMana(caster, spell)
      if mana == nil { return false }
      pos := s.getCasterPosition(caster, spell)
      if pos == nil { return false }
      s.dispatchSpellByType(caster, spell, pos)
      s.applySpellEffects(caster, spell, pos, slotIndex)
      return true
  }
  ```

  Update the existing call site in `updateCastingProgress` (line 248)
  — no API change needed since the return value can be ignored there.

- **Dependencies**: None. G36 (`CancelCast` cooldown) is already in
  place and provides a precedent for partial-cooldown behavior on a
  failed cast attempt.

- **Effort**: small — three signature changes (`executeCast`,
  `completeCast`) and one decision about whether to apply a partial
  GCD on failure. Add a regression test that drains mana between
  `StartCast` and `completeCast` and asserts `slots.Cooldowns[idx] == 0`.

---

## Resolved Findings (G32–G38) — Verification Citations

The following table records the exact lines in the *current* source
tree that demonstrate each prior finding has been remediated. Cite
these locations rather than re-opening the issues.

| ID  | Resolution evidence (current code)                                                                                                     |
|-----|----------------------------------------------------------------------------------------------------------------------------------------|
| G32 | `pkg/engine/advanced_class_system.go:14` (`lastApplied` field), `:67–73` (`prev` subtracted in `applyStatBonuses`), `:87–88,106–110,125–135` (per-domain `prev` subtraction)            |
| G33 | `pkg/engine/status_effect_crit_chance_system.go:42` (`prevCache` field), `:72–87` (cache swap + `stats.CritChance -= prev`), `:105–112` (expiry path clamp) |
| G34 | `pkg/engine/combat_system.go:562–565` (defense bonus folded in via `applyEquipmentSetDefenseBonus`), `:580` (final damage), `:599` (`getEquipmentSetDamageBonus`), `:618` (defense bonus reader) |
| G35 | `pkg/engine/combat_system.go:577–583` — `applyShieldAbsorption` precedes the `< 1.0` floor, and the floor only triggers when `finalDamage > 0` |
| G36 | `pkg/engine/spell_casting.go:2734–2738` — `slots.Cooldowns[slotIdx] = spell.Stats.Cooldown * castProgress` on `CancelCast`         |
| G37 | `pkg/engine/class_affinity_system.go:157–169` — `comp.AppliedManaRegen[affinityType]` stores absolute regen; subtracted on level-up |
| G38 | `pkg/engine/render_system.go:328–333` — guard uses `!pos.Initialized` instead of the (PrevX==0 && PrevY==0) heuristic              |

---

## Categories With No New Findings

These categories were investigated and produced no new defects this
revision. See `AUDIT.md` "Categories Audited Clean" for the supporting
citations.

- Per-Frame Stat Mutation (3a) — all known systems use delta-tracking.
- Combat Formula Correctness (3b) — shield/floor ordering fixed
  (G35); `MinDamageMultiplier` is intentional balance protection.
- Server Concurrency (3d) — `running` consistently locked; lock order
  `clientsMu → c.mu` is consistent across all paths.
- ECS Component Cache Consistency (3e) — `AddComponent` keeps the
  typed cache pointer in sync with the map entry.
- Input and UI Correctness (3f) — dead-zone applied before
  normalization; modal input fully suppressed; ESC handler dismisses
  in priority order.
- UI Display Bugs (3g) — health/mana bars clamp; render interpolation
  uses `Initialized`; `consumeItem` uses `Stats.Healing`.
- Procedural Generation Determinism (3i) — pooled-RNG re-seeding is
  intentional for deterministic replay.
- System Registration Ordering (3j) — `AdvancedClassSystem` is
  registered (`cmd/client/handlers.go:2162`); registration order is
  no longer correctness-critical thanks to delta-tracking fixes.
