# BUG AND GAP AUDIT — 2026-04-25 — Rev 7

> Supersedes Rev 6 (2026-04-24, GAPS.md). Prior gap IDs G1–G39 preserved.
> New findings: **G40–G46**.
> Status legend: ✅ RESOLVED | ⚠️ PARTIAL | 🔴 OPEN

This revision focuses on **Axis S1 / S2 — System Lifecycle Completeness** in the
desktop client. The investigation enumerated every `*engine.*System` field on
`systemsContainer` (`cmd/client/handlers.go:147`–`618`) and reconciled it against
every constructor invocation and every `game.World.AddSystem(...)` /
`addGenreSystem(...)` call in `cmd/client/`.

The result is a previously-undocumented cluster of **state-affecting systems
that exist only in the mobile path** (constructed by `engine.InitializeGameSystems`
in `pkg/engine/system_init.go`, called only from `cmd/mobile/mobile.go:164`).
On the desktop client, the corresponding `systemsContainer` fields stay nil
forever, so all of the day/night-cycle and terrain-bonus gameplay silently no-ops
for the platform that the README explicitly lists as the primary target.

## Build Status

- `go build ./...` — **environment-blocked**. The agent sandbox lacks X11/GL
  development headers, so `github.com/hajimehoshi/ebiten/v2/internal/glfw`
  fails with `fatal error: X11/Xlib.h: No such file or directory`. CI installs
  the equivalent packages on `ubuntu-latest` (`.github/workflows/test.yml:22`–`57`).
- `go vet ./...` — same environment block.
- `go test -race ./pkg/engine/... ./pkg/combat/... ./pkg/network/...` — same
  environment block.

All findings below were derived from static reading and `grep`-driven cross-
reference, citing `file.go:line` for every claim.

## Summary

| Axis | Category | Count | Critical | High | Medium | Low |
|------|----------|-------|----------|------|--------|-----|
| S1   | System Lifecycle             | 6 | 0 | 5 | 1 | 0 |
| S2   | Server/Client Symmetry       | 0 | 0 | 0 | 0 | 0 |
| S3   | Component Orphans            | 0 | 0 | 0 | 0 | 0 |
| S4   | Optional Dependencies        | 1 | 0 | 0 | 1 | 0 |
| S5   | Cross-System Wiring          | 0 | 0 | 0 | 0 | 0 |
| S6   | Entity Construction          | 0 | 0 | 0 | 0 | 0 |
| B1   | Stat Mutation                | 0 | 0 | 0 | 0 | 0 |
| B2   | Unconsumed Output            | 0 | 0 | 0 | 0 | 0 |
| B3   | Ordering                     | 0 | 0 | 0 | 0 | 0 |
| B4   | Error Path Symmetry          | 0 | 0 | 0 | 0 | 0 |
| B5   | Precision/Convention         | 0 | 0 | 0 | 0 | 0 |
| B6   | Concurrency                  | 0 | 0 | 0 | 0 | 0 |
| B7   | Callbacks                    | 0 | 0 | 0 | 0 | 0 |
| B8   | Platform Parity              | 0 | 0 | 0 | 0 | 0 |
| B9   | ECS Cache                    | 0 | 0 | 0 | 0 | 0 |

The Axis S4 entry (G46) is a downstream consequence of G42 — counted
separately because its surface (the fishing system's `SetLightingSystem` call
at `init_versions.go:207`) is observable independently and would need its
own remediation step (a nil-aware fallback or an init-order assertion).

## Findings

### HIGH

- [x] **[G40] `SpecializationCritDamageSystem` declared but never instantiated and never registered in any client/server entry point** — `cmd/client/handlers.go:393` — Axis **S1**

  **Evidence:**
  ```go
  // cmd/client/handlers.go:393
  specializationCritDamageSys  *engine.SpecializationCritDamageSystem  // Connects class specialization with crit damage bonuses
  ```
  Field declared. Constructor exists (`pkg/engine/specialization_crit_damage_system.go:17`,
  `NewSpecializationCritDamageSystem` at line of same file). Test suite is
  comprehensive (`pkg/engine/specialization_crit_damage_system_test.go`, ~480
  lines). However, `grep -rn "NewSpecializationCritDamageSystem|specializationCritDamageSys"
  cmd/ pkg/engine/system_init.go --include='*.go' | grep -v _test.go` returns
  **only the field declaration**. No constructor call anywhere in `cmd/`,
  no `world.AddSystem` registration, no inclusion in
  `pkg/engine/system_init.go::InitializeGameSystems` either. Sister systems
  `specializationManaBoostSys`, `specializationHealthRegenSys`,
  `specializationSpellDamageSys`, `specializationAttackSpeedSys`,
  `specializationDefenseSys`, `specializationLifestealSys` are all constructed
  in `cmd/client/init_versions.go:87`–`113` and registered at
  `cmd/client/handlers.go:2179`–`2194`. CritDamage was added to the struct but
  the matching `New*` and `AddSystem` lines were never written.

  **Incorrect/missing behavior:** Picking a specialization (Berserker,
  Champion, Battlemaster, etc.) confers zero crit-damage bonus despite the
  system being fully implemented and tested. Players see the
  documented per-specialization crit-damage multipliers in tooltips but the
  combat formula receives no actual bonus.

  **Expected behavior:** Like the other six specialization systems registered
  on the immediately-adjacent lines, CritDamage should be constructed in
  `initializeProgressionSystems` and registered in the same block at
  `handlers.go:2191–2194`.

  **Remediation:** In `cmd/client/init_versions.go` add the missing
  constructor call alongside the others:
  ```go
  sys.specializationCritDamageSys = engine.NewSpecializationCritDamageSystem(game.World, *seed+seedOffsetSpecCritDamage)
  sys.specializationCritDamageSys.SetGenre(*genreID)
  ```
  In `cmd/client/handlers.go` immediately after the `specializationDefenseSys`
  registration (line 2191), add:
  ```go
  game.World.AddSystem(sys.specializationCritDamageSys)
  ```

  **Validation:** Add a regression test in `cmd/client/handlers_test.go` that
  asserts `sys.specializationCritDamageSys != nil` and that the
  `engine.World.GetSystems()` slice contains a `*engine.SpecializationCritDamageSystem`.

- [x] **[G41] `SpecializationEvasionSystem` declared but never instantiated and never registered in any client/server entry point** — `cmd/client/handlers.go:394` — Axis **S1**

  **Evidence:**
  ```go
  // cmd/client/handlers.go:394
  specializationEvasionSys     *engine.SpecializationEvasionSystem     // Connects class specialization with evasion bonuses
  ```
  Same anti-pattern as G40. Constructor at
  `pkg/engine/specialization_evasion_system.go:54` (full implementation,
  documents 8 specialization-specific evasion bonuses ranging from +10% to
  +30%, with five genre modifiers). Test file exists with full coverage.
  Search across `cmd/` and `pkg/engine/system_init.go` for
  `NewSpecializationEvasionSystem` or `specializationEvasionSys` returns only
  the systemsContainer field declaration — no `New` call, no `AddSystem` call.

  **Incorrect/missing behavior:** Players choosing Shadowdancer (+30%
  documented), Windwalker (+25%), Assassin (+20%), Trickster (+20%),
  Marksman (+15%), Duelist (+15%), Beastmaster (+10%), or Exorcist (+10%)
  receive **no evasion bonus from their specialization** despite the system
  being fully implemented. The doc comment at lines 13–33 of
  `specialization_evasion_system.go` describes the intended behavior in detail.

  **Expected behavior:** Constructor in `init_versions.go:107`–`113` block,
  registration in `handlers.go:2188`–`2194` block, alongside the six sister
  specialization systems already wired.

  **Remediation:** Identical pattern to G40:
  ```go
  // cmd/client/init_versions.go (add near line 113)
  sys.specializationEvasionSys = engine.NewSpecializationEvasionSystem(game.World, *seed+seedOffsetSpecEvasion)
  sys.specializationEvasionSys.SetGenre(*genreID)

  // cmd/client/handlers.go (add near line 2194)
  game.World.AddSystem(sys.specializationEvasionSys)
  ```

  **Validation:** Same as G40, plus an integration assertion that an entity
  with `ClassProgressionComponent.SpecializationID == "shadowdancer"` shows
  `StatsComponent.Evasion` increase by ~30% after `Update` runs.

- [x] **[G42] Eleven Time-of-Day stat-affecting systems never constructed in the desktop client** — `cmd/client/handlers.go:353`–`363` — Axis **S1**

  **Evidence:** systemsContainer fields declared:
  ```go
  // cmd/client/handlers.go:353-363
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
  All eleven `New*` constructors live in `pkg/engine/timeofday_*.go`. They are
  invoked **only** in `pkg/engine/system_init.go:1904`–`1990` (in the body of
  `InitializeGameSystems`). Searching `cmd/client/` for
  `engine.NewTimeOfDayLightingSystem` and the other ten constructors returns
  zero non-test hits:
  ```bash
  $ grep -rnE "engine\.New(TimeOfDayLighting|TimeOfDayStealth|TimeOfDayXPBonus|TimeOfDayManaCost|TimeOfDayCriticalChance|TimeOfDayCompanionBonus|TimeOfDayManaRegen|TimeOfDayBlockChance|TimeOfDayEvasion|TimeOfDaySpellDamage|TimeOfDayAttackSpeed)" cmd/client/ --include='*.go' | grep -v _test.go
  # (no output)
  ```
  The desktop client constructs only one ToD system —
  `timeOfDayShadowDirectionSystem` at `cmd/client/handlers.go:1736`, which is
  presentation-only. `InitializeGameSystems` is referenced only by
  `cmd/mobile/mobile.go:164` (verified by `grep -rn "InitializeGameSystems\b"
  cmd/ --include='*.go' | grep -v _test.go`), so the desktop binary never
  registers the eleven systems. The connector at
  `cmd/client/terrain_collision_helpers.go:110`–`144` (`connectTimeOfDaySystems`)
  iterates `game.World`'s registered systems via type assertion, so when the
  systems were never `AddSystem`'d the loop finds nothing and the fields stay
  `nil`.

  **Incorrect/missing behavior:**
  - On desktop (and WebAssembly, which uses the same `cmd/client/` entry
    point — `cmd/client/main.go`), the day/night cycle has **no effect on
    gameplay**. Stealth detection thresholds, XP gains, spell mana cost,
    crit chance, companion stats, mana regen, block chance, evasion, spell
    damage, attack speed, and the ambient lighting tint itself never vary
    with time of day.
  - Mobile (iOS/Android) gets the full feature set via
    `pkg/engine/system_init.go`. This is a hard platform-parity asymmetry
    — the same save imported on phone vs. desktop produces materially
    different stat modifiers from the same time-of-day state.
  - Cascading silent failures: see G46.

  **Expected behavior:** Eleven `New*` calls plus matching
  `world.AddSystem` registrations in the desktop client's
  `initializeEnvironmentalSystems` (`cmd/client/handlers.go:1230`) or in the
  V4–V19 init sequence, mirroring the layout in
  `pkg/engine/system_init.go:1904`–`1990`.

  **Remediation:** Either (a) add the eleven `New*` + `AddSystem` lines to
  the desktop client init, mirroring `pkg/engine/system_init.go:1904`–`1990`
  exactly (preserving the sub-seed offsets `+7500`, `+7550`, `+7575`, `+7600`,
  `+7625`, `+7650`, `+7675`, `+7700`, `+7725`, `+7750`, `+7775`); or (b)
  refactor so the desktop client also calls `engine.InitializeGameSystems`
  (the current dual-init path is the underlying cause of both this finding
  and G43, G44, G45). Option (b) eliminates an entire class of future drift.

  **Validation:** Add an integration test in
  `cmd/client/handlers_test.go` that constructs a `systemsContainer` via the
  real init path (or its closest test-callable equivalent) and asserts each
  of the eleven fields is non-nil after lazy init completes. Mirror the
  existing system-existence assertions in `cmd/client/lazy_init_test.go`.

- [x] **[G43] Ten Terrain stat-affecting systems never constructed in the desktop client** — `cmd/client/handlers.go:299`–`311` — Axis **S1**

  **Evidence:** systemsContainer fields declared:
  ```go
  // cmd/client/handlers.go:299-311 (relevant subset)
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
  All ten constructors are invoked **only** in
  `pkg/engine/system_init.go:1229`–`1334`. Searching `cmd/` for any of these
  ten `NewTerrain*` constructors returns zero non-test hits. Confirmed by:
  ```bash
  $ grep -rnE "engine\.NewTerrain(MovementSpeed|CombatBonusSystem|StealthSystem|AmbushCrit|StatusEffectSystem|ManaRegen|SpellDamage|EquipmentDurability|RangedAccuracy|CompanionBonus)" cmd/client/ --include='*.go' | grep -v _test.go
  # (no output)
  ```
  The connector pattern is identical to G42 —
  `cmd/client/terrain_collision_helpers.go:25`–`97`
  (`connectTerrainMovementSystems`, `connectTerrainCombatSystems`,
  `connectTerrainStealthSystems`, `connectTerrainEffectSystems`) all
  type-assert against `game.World.GetSystems()`. Without registration in
  `cmd/client/`, the conditional bodies never execute and the fields remain
  nil.

  **Incorrect/missing behavior:** On desktop, terrain has **no gameplay
  effect** beyond pathing collision and rendering. Walking through tall
  grass or shallow water doesn't change move speed; standing on high ground
  doesn't grant the documented combat bonus; concealment terrain doesn't
  alter AI detection; lava/water tiles don't apply burning/wet status
  effects; no terrain-derived mana regen, spell damage modifier, ranged
  accuracy modifier, or companion bonus is ever applied. The
  `*ParticleSystem` particle effects that depend on these systems
  (e.g. `terrainCombatBonusParticleSystem.SetTerrainCombatBonusSystem` at
  `terrain_collision_helpers.go:41`) are all nil-guarded one level up
  (`if c.sys.terrainCombatBonusSystem != nil`), so even the particle
  feedback is silently suppressed.

  **Expected behavior:** Ten `New*` calls plus matching
  `world.AddSystem` registrations in the desktop client's terrain init
  block, with the systems' `SetTerrain(...)` setter invoked once the
  terrain object is materialized (the connector at
  `terrain_collision_helpers.go:25`–`87` already encodes the correct order
  — it just needs the systems to be present in `World.GetSystems()`).

  **Remediation:** Same two-option choice as G42. The simpler patch is to
  add the ten `New*` + `AddSystem` calls to the desktop client init
  before the terrain object is constructed, then let the existing
  connector wire `SetTerrain`. The strategically-correct fix is to make
  `cmd/client/` and `cmd/mobile/` share `engine.InitializeGameSystems`.

  **Validation:** Same shape as G42 — assert each of the ten fields is
  non-nil after init in `cmd/client/handlers_test.go`.

- [x] **[G44] `WeatherCritChanceSystem` constructed and configured but never registered** — `cmd/client/handlers.go:1390`–`1391` — Axis **S1**

  **Evidence:**
  ```go
  // cmd/client/handlers.go:1388-1391
  // WeatherCritChanceSystem - modifies critical hit chance based on weather conditions
  // Fog and dust increase crit (concealment), rain/snow decrease crit (precision penalty)
  sys.weatherCritChanceSystem = engine.NewWeatherCritChanceSystem(game.World, *seed+6775)
  sys.weatherCritChanceSystem.SetGenre(*genreID)
  ```
  Constructor and `SetGenre` are present, so the field is non-nil. But:
  ```bash
  $ grep -nE "weatherCritChanceSystem\)|AddSystem.*weatherCritChance" cmd/client/handlers.go cmd/client/init_versions.go
  # (no output beyond the constructor line above)
  ```
  No `game.World.AddSystem(sys.weatherCritChanceSystem)` call exists in
  `cmd/`. The system therefore has all three other lifecycle stages
  (New, Configure, Wire — implicit) but never **Register**, so its
  `Update()` method never runs.

  **Incorrect/missing behavior:** Crit chance never varies with weather.
  Fog/dust/sandstorm fail to deliver their documented "concealment crit
  bonus" and rain/snow fail to deliver their documented "precision penalty".
  Sister Weather*Systems on the immediately-following lines
  (`weatherXPBonusSystem` at line 1395, `weatherRangedAccuracySystem` at
  line 1385) follow the same construct-and-configure pattern but are
  registered later (search `addGenreSystem` and `world.AddSystem` for
  those names confirms registration). CritChance was the single one that
  fell off the registration list.

  **Expected behavior:** A `game.World.AddSystem(sys.weatherCritChanceSystem)`
  call alongside the other Weather* registrations.

  **Remediation:** Add a single line in `registerNonCriticalSystems` (or
  the symmetric helper that registers `weatherXPBonusSystem` and
  `weatherRangedAccuracySystem`):
  ```go
  game.World.AddSystem(sys.weatherCritChanceSystem)
  ```

  **Validation:** Assert in
  `cmd/client/handlers_test.go` that `game.World.GetSystems()` contains a
  `*engine.WeatherCritChanceSystem` after `lazyInit` completes.

- [x] **[G45] `WeatherBlockChanceSystem` field declared but never instantiated in the desktop client** — `cmd/client/handlers.go:334` — Axis **S1**

  **Evidence:**
  ```go
  // cmd/client/handlers.go:334
  weatherBlockChanceSystem  *engine.WeatherBlockChanceSystem  // Connects weather to block chance modifiers
  ```
  Constructor exists at `pkg/engine/weather_block_chance_system.go:52`. It
  is invoked **only** in `pkg/engine/system_init.go:1687` (the mobile path).
  Searching `cmd/client/` for `weatherBlockChanceSystem` returns just the
  field declaration:
  ```bash
  $ grep -rnE "weatherBlockChanceSystem|NewWeatherBlockChanceSystem" cmd/ --include='*.go' | grep -v _test.go
  cmd/client/handlers.go:334:	weatherBlockChanceSystem ...
  # (only one hit — the declaration)
  ```

  **Incorrect/missing behavior:** Block chance never varies with weather on
  desktop. Same severity as G44 — a single weather-driven combat stat is
  silently fixed. This row of the platform-parity matrix differs from G44
  only in that no construction was attempted at all (vs. G44 where
  construction ran but registration was forgotten).

  **Expected behavior:** Construct, configure, and register alongside the
  other Weather*ChanceSystem instances.

  **Remediation:**
  ```go
  // alongside cmd/client/handlers.go:1390
  sys.weatherBlockChanceSystem = engine.NewWeatherBlockChanceSystem(game.World, *seed+6780)
  sys.weatherBlockChanceSystem.SetGenre(*genreID)
  // and later, alongside the other Weather* AddSystem calls
  game.World.AddSystem(sys.weatherBlockChanceSystem)
  ```
  The seed offset `+6780` is chosen to match
  `pkg/engine/system_init.go:1687`, preserving the deterministic-hierarchy
  contract.

  **Validation:** Same shape as G44.

### MEDIUM

- [x] **[G46] `TimeOfDayFishingBonusSystem.SetLightingSystem` is called with a perpetually-nil reference on the desktop client** — `cmd/client/init_versions.go:207` — Axis **S4**

  **Evidence:**
  ```go
  // cmd/client/init_versions.go:205-209
  sys.timeOfDayFishingBonusSystem = engine.NewTimeOfDayFishingBonusSystem(game.World, *seed+seedOffsetFishing+100)
  sys.timeOfDayFishingBonusSystem.SetGenre(*genreID)
  sys.timeOfDayFishingBonusSystem.SetLightingSystem(sys.timeOfDayLightingSystem)  // ← always nil on desktop
  sys.timeOfDayFishingBonusSystem.SetFishingSystem(sys.fishingSystem)
  game.World.AddSystem(sys.timeOfDayFishingBonusSystem)
  ```
  `sys.timeOfDayLightingSystem` is never assigned in the desktop client (see
  G42), so the setter receives `nil`. The receiver guards it inside
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
  The `terrain_collision_helpers.finalizeTimeOfDayConnections` re-attempt
  at `cmd/client/terrain_collision_helpers.go:159`–`161` is also nil-guarded:
  ```go
  if c.sys.timeOfDayFishingBonusSystem != nil && c.sys.timeOfDayLightingSystem != nil {
      c.sys.timeOfDayFishingBonusSystem.SetLightingSystem(c.sys.timeOfDayLightingSystem)
  }
  ```
  so it never fires either (because `c.sys.timeOfDayLightingSystem` is
  also nil — see G42).

  **Incorrect/missing behavior:** Fishing catch rates do not vary by time
  of day on the desktop client. The system is registered, runs every frame,
  and silently early-returns on every tick.

  **Expected behavior:** Once G42 is fixed (lighting system actually
  registered), this dependency injection works as intended. Until then, a
  defensive log at the call site would surface the misconfiguration so
  future runs of this audit catch it earlier.

  **Remediation:** Resolve G42 first; this finding closes automatically.
  Optionally add a `clientLogger.Warn(...)` when
  `sys.timeOfDayLightingSystem == nil` at the `SetLightingSystem` call site
  so the silent dependency surfaces in the log.

  **Validation:** Once G42 is fixed, assert in
  `cmd/client/handlers_test.go` that
  `sys.timeOfDayFishingBonusSystem.lightingSystem != nil` after lazy init
  (requires either an exported accessor or a package-internal test).

## Prior Findings Status

| ID  | Title | Status | Evidence |
|-----|-------|--------|----------|
| G1–G31 | (Various, see git history) | ✅ RESOLVED | Reserved IDs preserved for code-comment back-references; all confirmed resolved in Rev 6 (`GAPS.md` Rev 6 historical notes). |
| G32 | `AdvancedClassSystem` per-frame stat re-application | ✅ RESOLVED | `pkg/engine/advanced_class_system.go:14,67–73,87–88,106–110,125–135` |
| G33 | `StatusEffectCriticalChanceSystem` accumulating crit chance | ✅ RESOLVED | `pkg/engine/status_effect_crit_chance_system.go:42,72–87,105–112` |
| G34 | Equipment-set defense bonus not folded into damage formula | ✅ RESOLVED | `pkg/engine/combat_system.go:562–565,580,599,618` |
| G35 | Shield-absorption / damage-floor ordering inversion | ✅ RESOLVED | `pkg/engine/combat_system.go:577–583` |
| G36 | `CancelCast` did not apply partial cooldown | ✅ RESOLVED | `pkg/engine/spell_casting.go:2734–2738` |
| G37 | `ClassAffinitySystem` mana regen not subtracted on level-up | ✅ RESOLVED | `pkg/engine/class_affinity_system.go:157–169` |
| G38 | Render interpolation skipped at world origin | ✅ RESOLVED | `pkg/engine/render_system.go:328–333` |
| G39 | `completeCast` applies cooldown after silent `executeCast` no-op | 🔴 OPEN | `pkg/engine/spell_casting.go:288–321` (no remediation merged since Rev 6) |

## Categories Audited Clean (this revision)

These categories were re-investigated under Rev 7 and produced no new
findings beyond those already documented in Rev 6.

- **B1 Stat Mutation** — `statBonusApplier` (`pkg/engine/statmod.go:25`–`123`)
  remains the canonical helper; spot checks of the
  `Specialization*` and `Reputation*` systems confirm Class-D delta-tracking.
- **B6 Concurrency** — `network.Server` lock ordering unchanged from Rev 6.
- **B9 ECS Cache** — `Entity.AddComponent` / `Entity.RemoveComponent`
  continue to keep typed cache pointers in sync.
- **S2 Server/Client Symmetry** — the new findings G40–G45 are technically
  also S2 issues (the missing systems are absent on the server too), but
  since the server doesn't pretend to register *any* class-specialization or
  time-of-day or terrain-stat systems, the asymmetry is intentional under
  the current "client predicts, server is permissive on these stats" design.
  Logged here as observation, not flagged as a separate finding.

## False Positives Considered and Rejected

| Candidate | Axis | Lines | Reason Rejected |
|-----------|------|-------|-----------------|
| `timeOfDayShadowDirectionSystem` "missing" | S1 | `cmd/client/handlers.go:1736`–`1737` | Actually constructed and registered via `addGenreSystem`. Visual-only system. Not in the G42 list. |
| `terrainReflectionTintSystem`, `terrainFishingBonusSystem` "missing" | S1 | `cmd/client/handlers.go:1837`, `cmd/client/init_versions.go:240` | Both are constructed in the desktop client. Excluded from G43. |
| `weatherFactionResistanceSystem`, `weatherEquipmentSheenSystem`, `weatherEntityWetnessSystem` "missing" | S1 | `cmd/client/handlers.go:1604,1808,1978` | All constructed in the desktop client; subsequent `AddSystem` / `addGenreSystem` calls confirmed by grep. |
| `aiStateBubbleSystem`, `armorHitSparkSystem`, etc. "missing" from initial unreg list | S1 | various | All registered via `addGenreSystem(...)` from `cmd/client/system_reg.go:27`. Initial regex missed this helper; second pass with `(AddSystem|addGenreSystem)` confirmed registration. |
| `combatEquipmentDurabilityParticleSystem` "missing" | S1 | (search) | Registered via `addGenreSystem` after a `SetParticleSystem` call. Verified non-nil registration. |
| `WeatherCritChanceSystem` "intentionally disabled" | S1 | `cmd/client/handlers.go:1388`–`1391` | The comment block at lines 1388–1389 explicitly documents the intended bonus/penalty effect. The construct + `SetGenre` calls would be dead code if disablement were intended. Treating G44 as a real omission, not an intentional disable. |
| Re-reporting G39 | B4 | `pkg/engine/spell_casting.go:299–300` | Already documented in `GAPS.md` Rev 6, unresolved. Listed in Prior Findings Status as still 🔴 OPEN, not as a new finding. |
