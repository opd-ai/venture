# Bug and Gap Details — 2026-04-25 (Rev 7)

> **Rev 7 — desktop-client system-lifecycle pass (2026-04-25).** This file
> supersedes `GAPS.md` Rev 6 (2026-04-24). G39 from Rev 6 remains **🔴 OPEN**
> and is reproduced verbatim below for continuity. New findings: **G40–G46**,
> all Axis S1 (System Lifecycle Completeness) or Axis S4 (Optional
> Dependency) defects in the desktop client (`cmd/client/`). G32–G38 remain
> ✅ RESOLVED — see `AUDIT.md` Rev 7 "Prior Findings Status" for the per-ID
> resolution evidence.
>
> **Legacy ID compatibility**: G1–G38 IDs remain reserved for cross-references
> in code comments and historical issue trackers. All are resolved.
>
> **Status legend**: ✅ RESOLVED | ⚠️ PARTIAL | 🔴 OPEN

---

## G39 — `completeCast` Always Applies Cooldown Even When `executeCast` Silently No-Ops

- **Axis**: B4 — Error Path Symmetry
- **Gap Class**: Behavioral — Asymmetric error-path cleanup (cooldown applied
  even when the resource-consuming portion of the action silently fails)
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
    if mana == nil { return }                  // ← spell vanishes silently
    pos := s.getCasterPosition(caster, spell)
    if pos == nil { return }                   // ← spell vanishes silently
    s.dispatchSpellByType(caster, spell, pos)
    s.applySpellEffects(caster, spell, pos, slotIndex)
}
```

- **Root Cause**: `completeCast` writes the cooldown unconditionally on
  the line immediately after the `executeCast` call, with no
  success/failure signal flowing back from `executeCast` or
  `validateAndConsumeMana`. The function treats "the cast bar reached
  100%" as equivalent to "the spell executed", which is incorrect when
  preconditions checked at `StartCast` (mana availability, mana
  component presence) no longer hold at completion.

- **Player Impact**: A spell that produces no effect locks out its slot for
  the full cooldown. In PvP, repeated mana-drain trivially denies a caster's
  entire spellbook with no resource cost to the drainer.

- **Remediation**: Have `executeCast` return a `bool`. Apply cooldown only
  on success (see Rev 6 detail for full code suggestion). G36 already
  precedents partial-cooldown behaviour for cancelled casts.

- **Client needed**: yes
- **Server needed**: yes (combat authority)
- **Dependencies**: None
- **Effort**: small

---

## G40 — `SpecializationCritDamageSystem` Never Instantiated and Never Registered

- **Axis**: S1 — System Lifecycle Completeness
- **Gap Class**: Structural — Never-instantiated (no `New*` call) and
  Never-registered (no `World.AddSystem` call). Six of seven sister systems
  are wired; this one was forgotten.
- **Status**: 🔴 OPEN
- **Severity**: HIGH

**Evidence**:

```go
// cmd/client/handlers.go:393
specializationCritDamageSys  *engine.SpecializationCritDamageSystem  // Connects class specialization with crit damage bonuses
```

The struct field is declared. The constructor exists at
`pkg/engine/specialization_crit_damage_system.go:17` (`type
SpecializationCritDamageSystem`) plus `NewSpecializationCritDamageSystem`
in the same file. A 480-line test file
(`pkg/engine/specialization_crit_damage_system_test.go`) exercises every
branch.

But searching all `cmd/` and `pkg/engine/system_init.go` for either
`NewSpecializationCritDamageSystem` or the field name returns **only the
field declaration above** — no construction, no `SetGenre` call, no
`AddSystem` registration. Compare to the six sister systems that **are**
wired:

```go
// cmd/client/init_versions.go:87,107,112  (constructor + SetGenre)
sys.specializationManaBoostSys = engine.NewSpecializationManaBoostSystem(...)
sys.specializationDefenseSys   = engine.NewSpecializationDefenseSystem(...)
sys.specializationLifestealSys = engine.NewSpecializationLifestealSystem(...)
// (and four more between)

// cmd/client/handlers.go:2179-2194  (registration block)
game.World.AddSystem(sys.specializationManaBoostSys)
game.World.AddSystem(sys.specializationHealthRegenSys)
game.World.AddSystem(sys.specializationSpellDamageSys)
game.World.AddSystem(sys.specializationAttackSpeedSys)
game.World.AddSystem(sys.specializationDefenseSys)
game.World.AddSystem(sys.specializationLifestealSys)
// CritDamage missing here
```

- **Root Cause**: When the new specialization stat surface was rolled out,
  the systemsContainer field was added but the matching `New*`/`SetGenre`/
  `AddSystem` triple was forgotten. No test in `cmd/client/` asserts that
  the field is non-nil, so the omission slipped past CI.

- **Player Impact**: Specializations that document a crit-damage bonus
  (Berserker, Champion, Battlemaster, etc.) confer no actual crit-damage
  multiplier in combat. Tooltips and UI may show the intended bonus while
  the combat formula receives the unmodified value.

- **Remediation**:
  1. Add a `seedOffsetSpecCritDamage` constant near the other
     `seedOffsetSpec*` constants.
  2. In `cmd/client/init_versions.go` near line 113 (next to the existing
     `specializationLifestealSys` block), add:
     ```go
     sys.specializationCritDamageSys = engine.NewSpecializationCritDamageSystem(game.World, *seed+seedOffsetSpecCritDamage)
     sys.specializationCritDamageSys.SetGenre(*genreID)
     ```
  3. In `cmd/client/handlers.go` near line 2194, add:
     ```go
     game.World.AddSystem(sys.specializationCritDamageSys)
     ```
- **Client needed**: yes
- **Server needed**: optional — not currently registered on server (matches
  pattern of G41 and other specialization systems)
- **Dependencies**: None
- **Effort**: small (3 lines + 1 constant + 1 regression test)

---

## G41 — `SpecializationEvasionSystem` Never Instantiated and Never Registered

- **Axis**: S1 — System Lifecycle Completeness
- **Gap Class**: Structural — Never-instantiated, Never-registered. Same
  anti-pattern as G40, sister system.
- **Status**: 🔴 OPEN
- **Severity**: HIGH

**Evidence**:

```go
// cmd/client/handlers.go:394
specializationEvasionSys     *engine.SpecializationEvasionSystem     // Connects class specialization with evasion bonuses
```

Constructor at `pkg/engine/specialization_evasion_system.go:54`. Doc
comment (lines 13–33) describes eight specialization-specific evasion
bonuses (Shadowdancer +30%, Windwalker +25%, Assassin +20%, Trickster +20%,
Marksman +15%, Duelist +15%, Beastmaster +10%, Exorcist +10%) and five
genre modifiers (fantasy ±0%, scifi +10%, horror −10%, cyberpunk +15%,
postapoc −5%).

Search across `cmd/` and `pkg/engine/system_init.go` returns only the
field declaration above. No construction, no `AddSystem`.

- **Root Cause**: Identical to G40. The two stat-surface additions were
  apparently checked in as a pair — both struct fields landed but neither
  registration did.

- **Player Impact**: All eight evasion-focused specializations grant zero
  evasion bonus despite tooltip/UI documentation stating otherwise.
  Particularly impactful for high-mobility builds (Shadowdancer with
  documented +30% evasion is the headline build for stealth play).

- **Remediation**: Mirror G40's three-step fix with `Evasion` substituted
  for `CritDamage` and `seedOffsetSpecEvasion` for `seedOffsetSpecCritDamage`.

- **Client needed**: yes
- **Server needed**: optional — matches existing specialization-system
  asymmetry
- **Dependencies**: None
- **Effort**: small

---

## G42 — Eleven Time-of-Day Stat Systems Never Constructed in the Desktop Client

- **Axis**: S1 — System Lifecycle Completeness
- **Gap Class**: Structural — Never-instantiated in the primary entry
  point. The systems exist and run on mobile (via
  `engine.InitializeGameSystems`) but not on desktop or WASM.
- **Status**: 🔴 OPEN
- **Severity**: HIGH

**Evidence**:

```go
// cmd/client/handlers.go:353-363  (eleven systemsContainer fields declared)
timeOfDayLightingSystem        *engine.TimeOfDayLightingSystem
timeOfDayStealthSystem         *engine.TimeOfDayStealthSystem
timeOfDayXPBonusSystem         *engine.TimeOfDayXPBonusSystem
timeOfDayManaCostSystem        *engine.TimeOfDayManaCostSystem
timeOfDayCriticalChanceSystem  *engine.TimeOfDayCriticalChanceSystem
timeOfDayCompanionBonusSystem  *engine.TimeOfDayCompanionBonusSystem
timeOfDayManaRegenSystem       *engine.TimeOfDayManaRegenSystem
timeOfDayBlockChanceSystem     *engine.TimeOfDayBlockChanceSystem
timeOfDayEvasionSystem         *engine.TimeOfDayEvasionSystem
timeOfDaySpellDamageSystem     *engine.TimeOfDaySpellDamageSystem
timeOfDayAttackSpeedSystem     *engine.TimeOfDayAttackSpeedSystem
```

All eleven `New*` constructors exist in `pkg/engine/timeofday_*.go` and
are invoked **only** in `pkg/engine/system_init.go:1904`–`1990`:

```go
// pkg/engine/system_init.go (excerpt)
1904:    NewTimeOfDayLightingSystem(game.World, config.Seed+7500), config.GenreID)
1908:    timeOfDayStealthSystem := NewTimeOfDayStealthSystem(game.World, config.Seed+7550)
1909:    timeOfDayStealthSystem.SetLightingSystem(result.TimeOfDayLightingSystem)
1916:    timeOfDayXPBonusSystem := NewTimeOfDayXPBonusSystem(game.World, config.Seed+7575)
...
```

`InitializeGameSystems` is called only by `cmd/mobile/mobile.go:164`.
Confirmed by:

```bash
$ grep -rn 'InitializeGameSystems\b' cmd/ --include='*.go' | grep -v _test.go
cmd/client/init_versions.go:720:	// Reuse the ModBrowserSystem already registered by InitializeGameSystems (if  ← comment only
cmd/mobile/mobile.go:164:	systemsInitResult, err = engine.InitializeGameSystems(gameInstance, config)
```

The desktop client's `cmd/client/handlers.go` constructs only one
ToD system — `timeOfDayShadowDirectionSystem` at line 1736 (presentation
only — modifies sprite drop-shadow direction, not gameplay stats).
Searching all eleven constructor names in `cmd/client/`:

```bash
$ grep -rnE 'engine\.New(TimeOfDayLighting|TimeOfDayStealth|TimeOfDayXPBonus|TimeOfDayManaCost|TimeOfDayCriticalChance|TimeOfDayCompanionBonus|TimeOfDayManaRegen|TimeOfDayBlockChance|TimeOfDayEvasion|TimeOfDaySpellDamage|TimeOfDayAttackSpeed)' cmd/client/ --include='*.go' | grep -v _test.go
# (no output — none of the eleven are constructed)
```

The connector at
`cmd/client/terrain_collision_helpers.go:110`–`144`
(`connectTimeOfDaySystems`) iterates `game.World.GetSystems()` and
captures references via type assertion. Because none of the eleven
systems are ever `AddSystem`'d in the desktop init, every type assertion
fails and every field remains nil. The connector silently returns; the
fields stay nil for the lifetime of the process.

- **Root Cause**: The desktop client (`cmd/client/handlers.go`) and the
  mobile client (via `pkg/engine/system_init.go`) maintain two separate
  system-init paths. The eleven ToD stat systems were added in the
  shared `pkg/engine/system_init.go` but the mirror entries were never
  added to the bespoke `cmd/client/` initialisation.

- **Player Impact**:
  - **Desktop and WebAssembly users** see no time-of-day variation in
    stats, lighting, AI detection, XP gains, mana cost, crit chance,
    companion stats, mana regen, block chance, evasion, spell damage, or
    attack speed.
  - **Mobile users** get the full feature set.
  - Same save imported on phone vs. desktop produces materially different
    combat outcomes, which is observable as desync if a mobile player
    plays alongside a desktop player on the same federated server.
  - Cascading silent failure — see G46 for the dead `SetLightingSystem`
    call that depends on this fix.

- **Remediation**: Two options.
  1. **Tactical (small effort, preserves dual-path):** Add eleven `New*`
     + `world.AddSystem` lines to the desktop client init, mirroring
     `pkg/engine/system_init.go:1904`–`1990` exactly. Preserve the
     sub-seed offsets (`+7500`, `+7550`, `+7575`, `+7600`, `+7625`,
     `+7650`, `+7675`, `+7700`, `+7725`, `+7750`, `+7775`) so saves are
     deterministic across platforms. After registration, the existing
     `connectTimeOfDaySystems` connector picks them up automatically.
  2. **Strategic (medium effort, eliminates entire defect class):**
     Refactor so `cmd/client/` and `cmd/mobile/` both call
     `engine.InitializeGameSystems`. This single change closes G42, G43,
     G45, and prevents the same pattern recurring as new systems are
     added to `system_init.go`.

- **Client needed**: yes (this is a desktop-client gap)
- **Server needed**: no — server is intentionally permissive on
  time-of-day stat modifiers under the current "client predicts, server
  is authoritative on damage but not on per-frame stat envelopes" design
- **Dependencies**: G46 closes automatically once this is fixed.
- **Effort**: small (option 1) or medium (option 2)

---

## G43 — Ten Terrain Stat Systems Never Constructed in the Desktop Client

- **Axis**: S1 — System Lifecycle Completeness
- **Gap Class**: Structural — Never-instantiated. Same root cause as G42:
  shared `system_init.go` registers them, desktop client init does not.
- **Status**: 🔴 OPEN
- **Severity**: HIGH

**Evidence**:

```go
// cmd/client/handlers.go:299-311  (relevant subset)
terrainMovementSpeedSystem     *engine.TerrainMovementSpeedSystem
terrainCombatBonusSystem       *engine.TerrainCombatBonusSystem
terrainStealthSystem           *engine.TerrainStealthSystem
terrainAmbushCritSystem        *engine.TerrainAmbushCritSystem
terrainStatusEffectSystem      *engine.TerrainStatusEffectSystem
terrainManaRegenSystem         *engine.TerrainManaRegenSystem
terrainSpellDamageSystem       *engine.TerrainSpellDamageSystem
terrainEquipmentDurabilitySys  *engine.TerrainEquipmentDurabilitySystem
terrainRangedAccuracySys       *engine.TerrainRangedAccuracySystem
terrainCompanionBonusSystem    *engine.TerrainCompanionBonusSystem
```

All ten `New*` constructors are invoked **only** in
`pkg/engine/system_init.go:1229`–`1334`. Search of `cmd/`:

```bash
$ grep -rnE 'engine\.NewTerrain(MovementSpeed|CombatBonusSystem|StealthSystem|AmbushCrit|StatusEffectSystem|ManaRegen|SpellDamage|EquipmentDurability|RangedAccuracy|CompanionBonus)' cmd/client/ --include='*.go' | grep -v _test.go
# (no output — none of the ten are constructed)
```

The connector pattern is identical to G42 —
`cmd/client/terrain_collision_helpers.go:25`–`97` performs a series of
type-asserted captures over `game.World.GetSystems()`, then sets
`SetTerrain(c.terrain)` on each capture. With no entries ever
registered, the conditional bodies never execute and the fields stay
nil. Particle systems that depend on the stat systems (e.g.
`terrainCombatBonusParticleSys.SetTerrainCombatBonusSystem` at line 41)
are also nil-guarded one level up, so even particle feedback is
suppressed.

- **Root Cause**: Same dual-init divergence as G42.

- **Player Impact**: On desktop, terrain has no gameplay effect beyond
  pathfinding and rendering. Tall grass / shallow water doesn't change
  move speed; high ground doesn't grant the documented combat bonus;
  concealment terrain doesn't alter AI detection thresholds; lava/water
  tiles don't apply burning/wet status effects; no terrain-derived mana
  regen, spell damage modifier, ranged accuracy modifier, or companion
  bonus is ever applied. Mobile gets all ten.

- **Remediation**: Same two-option choice as G42. Tactically: add ten
  `New*` + `AddSystem` lines to the desktop client init **before** the
  terrain object is constructed (so the existing connector wires
  `SetTerrain` correctly when the terrain materialises). Strategically:
  consolidate on `engine.InitializeGameSystems`.

- **Client needed**: yes
- **Server needed**: no (same authority/permissive design as G42)
- **Dependencies**: None
- **Effort**: small (option 1) or medium (option 2 — closes G42 and G45 too)

---

## G44 — `WeatherCritChanceSystem` Constructed and Configured but Never Registered

- **Axis**: S1 — System Lifecycle Completeness
- **Gap Class**: Structural — Never-registered. The constructor and
  configuration calls are present; only the `world.AddSystem` line is
  missing.
- **Status**: 🔴 OPEN
- **Severity**: HIGH

**Evidence**:

```go
// cmd/client/handlers.go:1388-1391
// WeatherCritChanceSystem - modifies critical hit chance based on weather conditions
// Fog and dust increase crit (concealment), rain/snow decrease crit (precision penalty)
sys.weatherCritChanceSystem = engine.NewWeatherCritChanceSystem(game.World, *seed+6775)
sys.weatherCritChanceSystem.SetGenre(*genreID)
```

Confirmed missing AddSystem:

```bash
$ grep -nE 'AddSystem.*weatherCritChance|weatherCritChanceSystem\)' cmd/client/handlers.go cmd/client/init_versions.go
# (no output beyond the constructor line above)
```

Sister systems on the immediately surrounding lines —
`weatherRangedAccuracySystem` (line 1385),
`weatherXPBonusSystem` (line 1395) — follow the same construct +
`SetGenre` pattern and *are* registered later. CritChance is the single
one that fell off the registration list.

- **Root Cause**: The construction and configuration triples for the
  Weather*Systems were added in one commit; the corresponding
  `AddSystem` lines were added later in a separate commit and one entry
  was overlooked.

- **Player Impact**: Crit chance never varies with weather on desktop.
  Fog/dust/sandstorm fail to deliver the documented "concealment crit
  bonus"; rain/snow fail to deliver the documented "precision penalty".
  All other weather-driven combat modifiers work correctly; this one
  outlier creates inconsistent player expectations.

- **Remediation**: Add a single line in the same `register*` helper
  that calls `game.World.AddSystem(sys.weatherRangedAccuracySystem)` and
  `game.World.AddSystem(sys.weatherXPBonusSystem)`:

  ```go
  game.World.AddSystem(sys.weatherCritChanceSystem)
  ```

- **Client needed**: yes
- **Server needed**: no (consistent with sister Weather* systems)
- **Dependencies**: None
- **Effort**: small (1 line + 1 regression test)

---

## G45 — `WeatherBlockChanceSystem` Field Declared but Never Instantiated in the Desktop Client

- **Axis**: S1 — System Lifecycle Completeness
- **Gap Class**: Structural — Never-instantiated. Same dual-init
  divergence as G42 / G43, but only one system this time.
- **Status**: 🔴 OPEN
- **Severity**: HIGH

**Evidence**:

```go
// cmd/client/handlers.go:334
weatherBlockChanceSystem  *engine.WeatherBlockChanceSystem  // Connects weather to block chance modifiers
```

Constructor exists at `pkg/engine/weather_block_chance_system.go:52`.
Invoked only in `pkg/engine/system_init.go:1687`. All `cmd/` references
are exhaustively the field declaration above:

```bash
$ grep -rnE 'weatherBlockChanceSystem|NewWeatherBlockChanceSystem' cmd/ --include='*.go' | grep -v _test.go
cmd/client/handlers.go:334:	weatherBlockChanceSystem  *engine.WeatherBlockChanceSystem ...
```

- **Root Cause**: Same dual-init divergence as G42 / G43.

- **Player Impact**: Block chance never varies with weather on desktop.
  Same "platform-asymmetric stat modifier" failure mode as G44 — on
  mobile, weather affects block chance; on desktop, it does not.

- **Remediation**: Mirror G44's fix. Construct, configure, and register
  alongside the other Weather*ChanceSystem instances (use seed offset
  `+6780` to match `pkg/engine/system_init.go:1687`):

  ```go
  // alongside cmd/client/handlers.go:1390
  sys.weatherBlockChanceSystem = engine.NewWeatherBlockChanceSystem(game.World, *seed+6780)
  sys.weatherBlockChanceSystem.SetGenre(*genreID)
  // and later, alongside the other Weather* AddSystem calls
  game.World.AddSystem(sys.weatherBlockChanceSystem)
  ```

- **Client needed**: yes
- **Server needed**: no
- **Dependencies**: None — but if the strategic remediation for G42 is
  taken (consolidate on `engine.InitializeGameSystems`), this finding
  closes automatically.
- **Effort**: small

---

## G46 — `TimeOfDayFishingBonusSystem.SetLightingSystem(nil)` on the Desktop Client

- **Axis**: S4 — Optional Dependency Audit
- **Gap Class**: Structural — Never-wired (the setter is called, but with
  a perpetually-nil reference, because the dependency itself is never
  instantiated — see G42).
- **Status**: 🔴 OPEN
- **Severity**: MEDIUM

**Evidence**:

```go
// cmd/client/init_versions.go:205-209
sys.timeOfDayFishingBonusSystem = engine.NewTimeOfDayFishingBonusSystem(game.World, *seed+seedOffsetFishing+100)
sys.timeOfDayFishingBonusSystem.SetGenre(*genreID)
sys.timeOfDayFishingBonusSystem.SetLightingSystem(sys.timeOfDayLightingSystem)  // ← always nil on desktop
sys.timeOfDayFishingBonusSystem.SetFishingSystem(sys.fishingSystem)
game.World.AddSystem(sys.timeOfDayFishingBonusSystem)
```

`sys.timeOfDayLightingSystem` is never assigned in the desktop client
(see G42), so the setter receives `nil`. The receiver guards it inside
`Update`:

```go
// pkg/engine/timeofday_fishing_bonus_system.go:84-94
func (s *TimeOfDayFishingBonusSystem) Update(entities []*Entity, deltaTime float64) {
    s.timeSinceCheck += deltaTime
    if s.timeSinceCheck < s.updateInterval { return }
    s.timeSinceCheck = 0
    if s.world == nil || s.lightingSystem == nil {
        return  // ← silent no-op forever on desktop
    }
    currentTime := s.lightingSystem.GetCurrentTimeOfDay()
    ...
}
```

The reattempt at `cmd/client/terrain_collision_helpers.go:158`–`164`
(`finalizeTimeOfDayConnections`) is also nil-guarded:

```go
if c.sys.timeOfDayFishingBonusSystem != nil && c.sys.timeOfDayLightingSystem != nil {
    c.sys.timeOfDayFishingBonusSystem.SetLightingSystem(c.sys.timeOfDayLightingSystem)
}
```

so it never fires either (since `c.sys.timeOfDayLightingSystem` is also
nil — see G42).

- **Root Cause**: Downstream consequence of G42. The init code optimistically
  assumed `sys.timeOfDayLightingSystem` would be populated by the time the
  setter ran, but the lighting system itself was never constructed in the
  desktop client.

- **Player Impact**: Fishing catch rates do not vary by time of day on the
  desktop client. The system is registered, runs every frame, and silently
  early-returns on every tick. Mobile gets the documented dawn/dusk/night
  catch-rate modifiers; desktop does not.

- **Remediation**: Resolve G42 first. This finding closes automatically
  once the lighting system is registered. Optionally, harden the call
  site at `init_versions.go:207` with:

  ```go
  if sys.timeOfDayLightingSystem == nil {
      clientLogger.Warn("timeOfDayFishingBonusSystem: SetLightingSystem received nil — fishing time-of-day bonuses will not apply")
  }
  sys.timeOfDayFishingBonusSystem.SetLightingSystem(sys.timeOfDayLightingSystem)
  ```

  so the silent dependency surfaces in production logs and future audits
  catch it without grep archaeology.

- **Client needed**: yes
- **Server needed**: no
- **Dependencies**: G42 (must close first; this finding will then
  close automatically)
- **Effort**: small (zero lines if G42 fix is option 2; one optional
  defensive log otherwise)

---

## Resolved Findings (G32–G38) — Verification Citations

The following table records the exact lines in the *current* source
tree that demonstrate each prior finding has been remediated. Cite
these locations rather than re-opening the issues.

| ID  | Resolution evidence (current code) |
|-----|-----------------------------------|
| G32 | `pkg/engine/advanced_class_system.go:14` (`lastApplied` field), `:67–73` (`prev` subtracted in `applyStatBonuses`), `:87–88,106–110,125–135` (per-domain `prev` subtraction) |
| G33 | `pkg/engine/status_effect_crit_chance_system.go:42` (`prevCache` field), `:72–87` (cache swap + `stats.CritChance -= prev`), `:105–112` (expiry path clamp) |
| G34 | `pkg/engine/combat_system.go:562–565` (defense bonus folded in via `applyEquipmentSetDefenseBonus`), `:580` (final damage), `:599` (`getEquipmentSetDamageBonus`), `:618` (defense bonus reader) |
| G35 | `pkg/engine/combat_system.go:577–583` — `applyShieldAbsorption` precedes the `< 1.0` floor, and the floor only triggers when `finalDamage > 0` |
| G36 | `pkg/engine/spell_casting.go:2734–2738` — `slots.Cooldowns[slotIdx] = spell.Stats.Cooldown * castProgress` on `CancelCast` |
| G37 | `pkg/engine/class_affinity_system.go:157–169` — `comp.AppliedManaRegen[affinityType]` stores absolute regen; subtracted on level-up |
| G38 | `pkg/engine/render_system.go:328–333` — guard uses `!pos.Initialized` instead of the (PrevX==0 && PrevY==0) heuristic |

---

## Categories With No New Findings (this revision)

Re-investigated under Rev 7; produced no defects beyond those already
documented.

- **B1 Per-Frame Stat Mutation** — `statBonusApplier`
  (`pkg/engine/statmod.go:25`–`123`) remains the canonical helper. Spot
  checks of `Specialization*` and `Reputation*` systems confirm
  Class-D delta-tracking with symmetric removal.
- **B2 Unconsumed Output** — for systems that *are* registered, the
  output flow remains intact. (The G40–G45 systems produce zero output
  not because nothing reads it but because they don't run.)
- **B3 Ordering** — system-execution order in `cmd/client/handlers.go`
  registration blocks remains compatible with the delta-tracking fixes
  from G32–G34.
- **B6 Concurrency** — `network.Server` lock ordering unchanged.
- **B9 ECS Cache** — `Entity.AddComponent` / `Entity.RemoveComponent`
  continue to keep typed cache pointers in sync.
