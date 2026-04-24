# IMPLEMENTATION GAP AUDIT — 2026-04-25 (rev 4)

> **Audit revision**: Rev 4 — deep-pass bug audit (2026-04-25). Supersedes Rev 3
> of this file. Rev 3 findings (G17–G20) remain open and are preserved verbatim
> below. Eleven new gaps G21–G31 were discovered by tracing live data-flow paths
> through the ECS update loop, HUD rendering, combat callbacks, and the mobile
> input pipeline.
>
> Rev 3 (2026-04-25): G17–G20 — WebRTC stub, ClassProgression no-op, companion
> scout velocity, ambush random offset.
>
> **Status legend**: ✅ RESOLVED | ⚠️ PARTIAL | 🔴 OPEN

---

## Project Architecture Overview

Venture is a fully procedural multiplayer action-RPG built on Go 1.24.5 and
Ebiten v2.9.3. **All graphics, audio, and gameplay content are generated at
runtime from a single binary; no external asset files exist.** Key architectural
pillars:

- **ECS**: `pkg/engine/` — ~400 files, 100+ systems, 183 `world.AddSystem` calls
  in `pkg/engine/system_init.go`.
- **Procedural generation**: `pkg/procgen/` — 25+ generators (terrain, entities,
  items, quests, spells, dialog, narrative).
- **Runtime rendering**: `pkg/rendering/` — sprites, tiles, lighting, particles,
  post-processing.
- **Procedural audio**: `pkg/audio/` — synthesis engine, adaptive music, SFX.
- **Multiplayer**: `pkg/network/` — TCP-based with prediction, lag compensation,
  federation, voice chat.
- **Platforms**: Linux, macOS, Windows, WebAssembly, iOS, Android.

---

## Build Status

`go build ./...` was attempted but failed due to missing X11 / OpenGL headers
(`X11/Xlib.h`) in the CI runner — a runner environment issue, not a code defect.
The failure is confined to Ebiten's GLFW C bindings and does not reflect any
Go-level error. `go vet` could not be run for the same reason. All static
analysis in this audit is based on source-level inspection.

---

## Gap Summary

| Category            | Count  | Critical | High | Medium | Low |
|---------------------|--------|----------|------|--------|-----|
| Stubs / Simulated   |   2    |    0     |  1   |   1    |  0  |
| Dead Code           |   1    |    0     |  0   |   0    |  1  |
| Partially Wired     |   2    |    0     |  0   |   0    |  2  |
| Gameplay Logic Bugs |   7    |    0     |  4   |   3    |  0  |
| Mobile Input        |   1    |    1     |  0   |   0    |  0  |
| **Total (G17–G31)** | **13** |  **1**   |**5** | **4**  |**3**|

> G1–G16 are all resolved (see Legacy section).  G17–G20 carry over from Rev 3.

---

## Implementation Completeness by Package

| Package | Assessment | Notes |
|---------|-----------|-------|
| `pkg/engine` | ⚠️ Bugs | 183 systems registered; logic bugs in combat callbacks, talent stat tracking, HUD (G22–G28) |
| `pkg/procgen` | ✅ Complete | All generators dispatch correctly by genre |
| `pkg/rendering` | ✅ Complete | Headless stubs are intentional (build-tag guarded) |
| `pkg/audio` | ✅ Complete | 91–98% coverage per prior benchmarks |
| `pkg/network` | ⚠️ Partial | TCP federation complete; WebRTC transport is simulated (G17) |
| `pkg/network/federation/webrtc` | 🔴 Stub | `Peer.Connect()` calls `simulateConnection()` (G17) |
| `pkg/world` | ✅ Complete | Housing, economy, raids, territory all wired |
| `pkg/mobile` | 🔴 Bug | `MobileInputAdapter.Type()=="input"` replaces EbitenInput; `processEntityInputs` skips it (G21) |
| `pkg/modding` | ✅ Complete | FileSystemModRepository + HotReloadSystem both wired (G3, G15, G16 resolved) |
| `pkg/vr` | ✅ Complete (experimental) | OpenXR + WebXR adapters present; deliberately labeled experimental |
| `pkg/memprofile` | ✅ Complete | `fmt.Printf` exemption formally documented at `profile.go:195` |
| `pkg/engine/class_progression_system.go` | ⚠️ Partial | `Update()` is a no-op; progression via explicit API only (G18) |
| `pkg/engine/companion_system.go` | ⚠️ Stub | `executeScout()` uses hardcoded velocity instead of pathfinding (G19) |
| `pkg/engine/behavior_tree_advanced_nodes.go` | ⚠️ Stub | Ambush node uses random offset, not cover-based pathfinding (G20) |
| `cmd/client` | ⚠️ Bug | XP double-award: kill callback + death callback both call `AwardXP` (G22) |
| `cmd/server` | ⚠️ Bug | `consumeItem` heals from `item.Stats.Defense` instead of heal stat (G25) |
| `cmd/mobile` | 🔴 Bug | Mobile input wiring broken — see `pkg/mobile` entry (G21) |

---

## Findings

### HIGH

- [ ] **[G17] WebRTC Browser-to-Browser Federation is Simulated, Not Real**
  — `pkg/network/federation/webrtc/peer.go:4,77` —
  The package header states "This is a stub implementation for testing; real
  WebRTC integration requires `github.com/pion/webrtc/v3`." The `Connect()`
  method calls `simulateConnection()` (line 77), which queues fake
  `StateConnected` transitions without any real ICE/DTLS negotiation. The
  signaling server is similarly simulated (`signaling.go:76,299`). The
  WASM-specific initializer `initWebRTCFederation()` at
  `cmd/client/webrtc_wasm.go:20` is defined but never called from any entry
  point. `NewWebRTCTransport()` (`transport_webrtc.go:31`) is never instantiated
  outside of `_test.go` files.

  **Impact**: Browser-to-browser federation (WASM client ↔ WASM server without a
  dedicated TCP host) is advertised — README line 60 lists "federation/WebRTC,
  portals" — but non-functional. Desktop TCP-based federation is fully
  operational; this gap only blocks the WebRTC code path for WASM peers.

  **Remediation**:
  1. Add `github.com/pion/webrtc/v3` (WASM-safe, no CGo) to `go.mod`.
  2. In `pkg/network/federation/webrtc/peer.go`, replace `simulateConnection()`
     body with `pion.NewPeerConnection`, data-channel creation, and SDP
     offer/answer exchange via the existing signaling transport.
  3. In `cmd/client/webrtc_wasm.go`, call `initWebRTCFederation(clientID)` and
     wire the returned `*Peer` into `sys.federationProtocol` via
     `NewWebRTCTransport()`.

  **Validation**: WASM build; two browser tabs can exchange a "ping" via the
  WebRTC data channel without a dedicated TCP server.

---

### MEDIUM

- [ ] **[G19] `CompanionSystem.executeScout()` Uses Hardcoded Velocity**
  — `pkg/engine/companion_system.go:496–502` —
  The comment reads "This is a stub - full implementation would use pathfinding".
  The body unconditionally sets `VX = 80.0`, `VY = 80.0`. A companion in Scout
  mode moves diagonally north-east indefinitely; it never returns to the owner,
  never avoids walls, and never changes direction.

  **Impact**: Companions assigned the Scout role (set via
  `CompanionComponent.BehaviorMode = BehaviorScout`) exhibit broken movement.
  Any player who relies on scout companions for exploration or detection will see
  nonsensical behavior.

  **Remediation**: Either (a) implement expanding-circle motion using the entity's
  `SpatialPartition` to select walkable cells at increasing radii, or (b) as a
  minimal fix, add angular variation using a per-companion counter so scouts move
  in four cardinal directions sequentially rather than one fixed diagonal.

  **Validation**: Spawn a companion with `BehaviorScout` mode; confirm it moves
  away from owner, turns at obstacles, and returns after a configurable radius.

---

### LOW

- [ ] **[G18] `ClassProgressionSystem.Update()` is a No-op**
  — `pkg/engine/class_progression_system.go:18–20` —
  The `Update` method body is a single comment: "Currently a stub - progression
  happens through LevelUp() calls / This system could be extended to apply
  passive effects." The system is registered (`cmd/client/handlers.go:2170`,
  `cmd/server/v4_systems.go:99`) and called every frame, consuming a scheduler
  slot without producing any output.

  **Impact**: Any time-based passive class effects (mana regen rate modifiers,
  class-specific buff cooldowns, etc.) cannot be applied automatically. Game
  designers must wire every class effect through explicit `LevelUp()` callsites.
  The core leveling path is unaffected.

  **Remediation**: Implement at minimum an empty-but-documented contract: if
  passive effects are genuinely out of scope, add a `_ = deltaTime` comment and
  a `// COMPLEXITY JUSTIFICATION: passive effects are event-driven` note to
  satisfy the quality gate. If passive effects are desired, iterate entities with
  `class_progression` component and apply class-specific regen ticks.

  **Validation**: `go test ./pkg/engine/ -run TestClassProgressionSystem`; all
  cases pass; no regression in leveling tests.

---

- [ ] **[G20] BehaviorTree Ambush Node Uses Random Positional Offset**
  — `pkg/engine/behavior_tree_advanced_nodes.go:391–396` —
  The comment states "In a full implementation, this would use pathfinding data."
  The ambush position is computed as the enemy's current position plus a ±50-unit
  random X/Y offset (`rng.Float64() - 0.5) * 100`). This does not account for
  walls, cover, or line-of-sight; enemies may "ambush" from inside solid terrain.

  **Impact**: Enemy AI quality for stealth/ambush archetypes is reduced. No
  feature is hard-blocked; the node evaluates to `NodeRunning` and the enemy
  eventually moves somewhere nearby.

  **Remediation**: Query `SpatialPartition` for passable tiles near the current
  position; prefer cells with low `VisibilityComponent.Visibility` (if populated
  by the LightingSystem) as cover candidates. Falls back to the current random
  offset if no cover tiles are found.

  **Validation**: `go test ./pkg/engine/ -run TestBehaviorTreeAmbush`; enemy
  ambush positions resolve to walkable cells.

---

### CRITICAL (Rev 4)

- [ ] **[G21] Mobile Player Entity Receives No Input — `*MobileInputAdapter` Silently Dropped**
  — `pkg/engine/input_system.go:1035–1041`, `pkg/mobile/input_adapter.go:49–52`,
  `cmd/mobile/mobile.go:308–309` —
  `AddComponent` stores components keyed by `Type()`. Both `EbitenInput` and
  `MobileInputAdapter` return `"input"` from `Type()`. When
  `playerEntity.AddComponent(mobileInput)` runs at `mobile.go:308`, it
  overwrites any existing `*EbitenInput` in the entity's component map.
  `InputSystem.processEntityInputs` then calls
  `entity.GetComponent("input")` — which succeeds — but the subsequent
  type assertion `inputComp.(*EbitenInput)` fails (the stored value is
  `*MobileInputAdapter`), and the entity is silently skipped (`continue`).
  As a result, `applyInputToVelocity` is never called for the mobile player:
  the player entity cannot move, attack, or interact via touch on iOS or
  Android. In addition, `MobileInputAdapter.Update()` — which reads live
  touch positions from `DualJoystickLayout` — is never called during the
  game loop, so even if the type-assertion bug were fixed the joystick state
  would remain stale.

  **Impact**: iOS and Android players are completely unable to control their
  character. All movement and combat input is discarded every frame.

  **Remediation**:
  1. Extend `processEntityInputs` to handle both concrete types. Add a
     second branch that reads from `InputProvider` interface when
     `*EbitenInput` assertion fails:
     ```go
     if provider, ok := inputComp.(InputProvider); ok {
         s.processInputProvider(entity, provider, deltaTime)
     }
     ```
  2. Call `mobileInput.Update()` once per frame in `InputSystem.Update()`
     by keeping a slice of registered `MobileInputAdapter` instances or by
     calling `Update()` on all components that implement an `Updater`
     interface before `processEntityInputs`.
  3. Alternatively, make `MobileInputAdapter` a thin wrapper that *delegates*
     to an `EbitenInput` it owns, writing touch results into the inner
     `EbitenInput` during its own `Update()`, and returning `"input_mobile"`
     from `Type()` to avoid the collision.

  **Affected Files**:
  - `pkg/engine/input_system.go:1033–1041` (type assertion skip)
  - `pkg/mobile/input_adapter.go:49–52` (`Type()` collision)
  - `cmd/mobile/mobile.go:308–309` (overwrites EbitenInput)

  **Validation**: On Android/iOS simulator, player entity responds to
  virtual joystick movement and attack button within one game tick.

---

### HIGH (Rev 4)

- [ ] **[G22] XP Double-Award on Every Kill**
  — `pkg/engine/system_init.go:929–933`, `cmd/client/client_loot.go:511–513` —
  Two independent callbacks both call `AwardXP` for the same kill event.
  The `SetKillCallback` registered at `system_init.go:929` calls
  `progressionSystem.CalculateXPReward(target)` then
  `progressionSystem.AwardXP(attacker, xp)` for every combat kill.
  Separately, `configureDeathCallback` (wired at `cmd/client/handlers.go:3585`)
  calls `createDeathCallback` which calls `calculateEnemyXP(enemy)` then
  `(*progressionSystem).AwardXP(*playerEntity, xpAmount)` at `client_loot.go:512`.
  Both fire on the same entity death. The two XP formulae are also different
  functions (`CalculateXPReward` vs `calculateEnemyXP`), so the award amounts
  do not even cancel out — the player receives both.

  **Impact**: Every enemy killed awards XP twice. Players level up roughly
  twice as fast as designed. Exploitable by farming low-level enemies.

  **Remediation**: Remove the `AwardXP` call from one site. The recommended
  approach is to remove it from `system_init.go:929–933` (the kill callback)
  and retain the richer `createDeathCallback` path which also handles loot
  drops, death animation, and `DeadComponent` attachment in one transaction.
  Alternatively, remove it from `client_loot.go` and extend the system_init
  kill callback to cover loot logic.

  **Affected Files**:
  - `pkg/engine/system_init.go:918–936` (kill callback XP award)
  - `cmd/client/client_loot.go:511–513` (death callback XP award)
  - `cmd/client/handlers.go:3585` (wires death callback)

  **Validation**: Kill one enemy; verify `ProgressionComponent.XP` increases
  by exactly `calculateEnemyXP()` once, not twice.

---

- [ ] **[G23] TalentSystem Stat Accumulation — Old Bonuses Never Removed**
  — `pkg/engine/talent_system.go:183–212` —
  `applyStatsBonuses(stats, bonuses)` at line 183 directly adds flat bonuses
  to live `StatsComponent` fields:
  ```go
  stats.Attack  += bonuses.FlatDamage
  stats.Defense += bonuses.FlatDefense
  ```
  There is no corresponding subtraction of previously-applied bonuses before
  the new values are added. The system fires whenever `talent.Dirty == true`,
  which is set on initial allocation, on reallocation, and on `ResetAll()`.
  After a talent reset, `applyStatsBonuses` is called with a zero-valued
  `TalentBonus` (adding 0), leaving the old bonuses permanently baked into
  the stats. After a reallocation to different talents, both the old and new
  bonuses stack.

  Compare with `AttributeAllocationSystem`, which correctly calls
  `removeAppliedBonuses(stats, prevBonuses)` at `attribute_allocation_system.go:204`
  before applying new values.

  **Impact**: Players who reset or reallocate talents retain the old stat
  bonuses indefinitely. Repeated reallocation allows unbounded stat growth
  without any talent cost. Combat balance is completely broken for any player
  who uses the talent respec flow.

  **Remediation**: Store the last applied `TalentBonus` in
  `TalentComponent.AppliedBonuses` (analogous to
  `AttributeAllocationComponent.AppliedBonuses`) and call a new
  `removeTalentBonuses(stats, c.AppliedBonuses)` before every
  `applyStatsBonuses` call.

  **Affected Files**:
  - `pkg/engine/talent_system.go:183–212` (missing subtraction step)
  - `pkg/engine/talent_component.go` (no `AppliedBonuses` field — needs adding)

  **Validation**: Allocate talents, check `stats.Attack`; reset all talents,
  check `stats.Attack` returns to pre-talent baseline.

---

- [ ] **[G24] Desktop HUD Has No Mana Bar**
  — `pkg/engine/hud_system.go:71–99` —
  `HUDSystem.Draw()` calls `drawHealthBar()`, `drawStatsPanel()`,
  `drawExperienceBar()`, `drawNetworkStatus()`, and `drawTerritoryBonuses()`.
  There is no `drawManaBar()` call. Mana is a primary gameplay resource
  consumed by spells (100+ mana-related systems registered in
  `system_init.go`), and `ManaComponent` is attached to every player entity.
  The mobile HUD in `pkg/mobile/ui.go` includes a `ManaBar ProgressBar` field
  and renders it correctly; the desktop HUD does not.

  **Impact**: Desktop players have no visual feedback on mana availability.
  Spells fail silently (or with a log message) when mana is insufficient;
  players cannot gauge remaining mana or regen rate.

  **Remediation**: Add a `drawManaBar(screen, entity)` method to `HUDSystem`
  modeled on `drawHealthBar`. Read `ManaComponent.Current`/`Max` from the
  player entity, draw a blue bar below the health bar, and clamp the fill
  fraction to `[0, 1]`.

  **Affected Files**:
  - `pkg/engine/hud_system.go:71–99` (`Draw()` missing mana bar call)
  - `pkg/engine/spell_casting.go:23` (`ManaComponent` definition, for reference)

  **Validation**: Launch desktop client; blue mana bar visible below health
  bar; bar decreases when a spell is cast and refills during regen.

---

- [ ] **[G25] Server `consumeItem` Heals Player by `item.Stats.Defense`**
  — `cmd/server/player_management.go:298` —
  The server-side item consumption handler computes:
  ```go
  healAmount := float64(item.Stats.Defense)
  ```
  `item.Stats.Defense` is the item's armor contribution and is typically 0
  for consumable potions (which have no armor value). Healing potions that
  set `item.Stats.Healing` or derive heal amount from `item.Value` are
  silently healed for 0 HP.

  **Impact**: All server-authoritative health restores from consumables
  produce no effect. Client-side prediction may show a heal visually, but
  the server authoritative state will not reflect it, causing a correction
  snap. In pure server mode (dedicated server), potions do nothing.

  **Remediation**: Replace `item.Stats.Defense` with the appropriate field.
  If a `Healing` stat exists: `healAmount := float64(item.Stats.Healing)`.
  If no dedicated healing stat: use `item.Value` as the heal amount and add
  a `Healing int` field to `ItemStats` for future use.

  **Affected Files**:
  - `cmd/server/player_management.go:298` (wrong stat field)

  **Validation**: Consume a healing potion via server command; player's
  `HealthComponent.Current` increases by the intended amount.

---

### MEDIUM (Rev 4)

- [ ] **[G26] `AttributeEffects` Fields Defined and Defaulted But Never Applied**
  — `pkg/engine/attribute_allocation_system.go:113–200`,
  `pkg/engine/attribute_allocation_component.go:93–108` —
  `AttributeEffects` (returned by `DefaultAttributeEffects()`) defines five
  fields with non-zero defaults: `CarryCapPerStr` (5.0), `SpeedBonusPerAgi`
  (0.5), `ManaRegenPerInt` (0.1), `HealthRegenPerVit` (0.05),
  `StaminaPerEnd` (10.0). The `applyAttributeBonuses` function at line 113
  applies only: STR→Attack, AGI→Evasion, INT→MagicPower+ManaMax,
  VIT→HealthMax, END→Defense+BlockChance, LCK→CritChance. The five
  `AttributeEffects` fields are read into local variables but never written
  to any component.

  **Impact**: STR does not increase carry capacity; AGI does not increase
  movement speed; INT does not increase mana regen rate; VIT does not
  increase health regen rate; END does not increase stamina. Players who
  specialize in these attributes gain no benefit from the advertised effects.

  **Remediation**: After each attribute loop in `applyAttributeBonuses`,
  write derived values to the appropriate components:
  - `CarryCapPerStr` → `InventoryComponent.MaxCarryWeight`
  - `SpeedBonusPerAgi` → `VelocityComponent.MaxSpeed`
  - `ManaRegenPerInt` → `ManaComponent.Regen`
  - `HealthRegenPerVit` → `HealthComponent.RegenRate` (if field exists)
  - `StaminaPerEnd` → `StaminaComponent.Max` (if applicable)

  **Affected Files**:
  - `pkg/engine/attribute_allocation_system.go:113–200`
  - `pkg/engine/attribute_allocation_component.go:93–108`

  **Validation**: Allocate 10 STR points; confirm `InventoryComponent.MaxCarryWeight`
  increases by `10 * effects.CarryCapPerStr`.

---

- [ ] **[G27] HUD Health Bar Overflows Background on Overheal**
  — `pkg/engine/hud_system.go:126` —
  ```go
  healthPct := float32(health.Current / health.Max)
  fillWidth  := int(float32(barWidth) * healthPct)
  ```
  If `health.Current > health.Max` (possible via healing spells, buffs, or
  rounding), `healthPct > 1.0` and `fillWidth > barWidth`. The fill
  rectangle is drawn outside its background, overwriting adjacent HUD
  elements. Compare with `pkg/mobile/ui.go:620` which correctly clamps:
  `value = clamp(value, 0, 1)`.

  **Impact**: Any overheal effect causes visual corruption in the HUD health
  bar area on desktop. Severity increases if adjacent elements are drawn at
  fixed offsets.

  **Remediation**:
  ```go
  healthPct := float32(health.Current) / float32(health.Max)
  if healthPct > 1.0 { healthPct = 1.0 }
  if healthPct < 0.0 { healthPct = 0.0 }
  ```
  Also fix the integer-division bug: `health.Current / health.Max` performs
  integer division (both fields are `float64` according to
  `components.go`—verify the actual type and use `float32(health.Current) /
  float32(health.Max)` regardless).

  **Affected Files**:
  - `pkg/engine/hud_system.go:124–130`

  **Validation**: Apply an overheal buff that sets `health.Current = health.Max * 1.5`;
  confirm fill bar stops at the background boundary.

---

- [ ] **[G28] Entity Death Callback Fires Every Frame Until Entity Is Removed**
  — `pkg/engine/combat_system.go:218–250` —
  `handleEntityDeath` detects `health.Current <= 0` and calls
  `s.onDeathCallback(entity, attacker)`, but does NOT add `DeadComponent`
  or any other guard to the entity. On the next frame the entity still exists,
  `health.Current` is still ≤ 0, and the callback fires again. This repeats
  until the entity is removed from the world (which may take several frames
  or never happen server-side for persistent NPCs).
  The client-side `createDeathCallback` (`client_loot.go:490`) guards with
  `if enemy.HasComponent("dead") { return }` and then adds `NewDeadComponent`,
  so the *client* path is protected. But the `onDeathCallback` itself fires
  multiple times before the guard can act, and any future or server-side death
  callback that lacks its own guard will process multiple times.

  **Impact**: Any death callback without an internal guard (e.g., server-side
  stat tracking, achievement callbacks, drop-chance callbacks) runs N times
  per entity death, where N is the number of frames from death detection to
  entity removal. Duplicated loot drops, XP, achievement triggers, and
  analytics events are possible.

  **Remediation**: In `handleEntityDeath`, immediately add `DeadComponent`
  to the entity before calling `onDeathCallback`:
  ```go
  entity.AddComponent(NewDeadComponent())
  if s.onDeathCallback != nil {
      s.onDeathCallback(entity, attacker)
  }
  ```
  Then change `processEntity`'s entry guard to skip entities that already have
  `DeadComponent`.

  **Affected Files**:
  - `pkg/engine/combat_system.go:218–250`

  **Validation**: Kill one entity; assert `onDeathCallback` is called exactly once.

---

- [ ] **[G29] `ClassAffinitySystem.decayStreaks` Uses Hardcoded `currentTime = 0`**
  — `pkg/engine/class_affinity_system.go:103` —
  ```go
  currentTime := 0.0 // Would be game time in real implementation
  ```
  `timeSinceActivity := currentTime - data.LastActivityTime` is always
  negative (since `LastActivityTime` is set to positive game-time values),
  so `timeSinceActivity > s.streakDecayTime` is never true. Streaks never
  decay regardless of how much time has passed since the last affinity action.

  **Impact**: Combo/streak bonuses for class affinities are permanent once
  acquired. A player can build a maxed streak early and retain it indefinitely
  with no further action. This invalidates the risk/reward design of the
  affinity streak system.

  **Remediation**: Pass the current world time as a parameter or store it on
  the system:
  ```go
  currentTime := s.worldTime  // set via SetWorldTime(dt) each frame
  ```
  Or use a monotonic counter incremented in `Update()`:
  ```go
  s.elapsedTime += deltaTime
  currentTime := s.elapsedTime
  ```

  **Affected Files**:
  - `pkg/engine/class_affinity_system.go:100–113`

  **Validation**: Build max streak; wait (simulated) past `streakDecayTime`;
  confirm `CurrentStreak` resets to 0.

---

### LOW (Rev 4)

- [ ] **[G30] Entities Can Attack Themselves — No Self-Damage Guard**
  — `pkg/engine/combat_system.go:278` —
  `validateAttackEntities(attacker, target *Entity)` checks that both
  entities are non-nil and live, but does not check
  `attacker.ID == target.ID`. An entity can therefore be set as its own
  target, triggering the full damage pipeline against itself. This can occur
  via AOE effects, reflection spells, or a bug in target selection.

  **Impact**: Low severity for normal gameplay. Exploitable via AoE or
  reflection mechanics to trigger kill callbacks on the attacker, awarding
  self-kill XP and potentially looping on the XP-double-award bug (G22).

  **Remediation**: Add one line at the top of `validateAttackEntities`:
  ```go
  if attacker.ID == target.ID { return false }
  ```

  **Affected Files**:
  - `pkg/engine/combat_system.go:278–310`

  **Validation**: `go test ./pkg/engine/ -run TestCombatSelfAttack`; verify
  attack with attacker == target returns false.

---

- [ ] **[G31] `CarryOverSystem` Not Registered in `system_init.go`**
  — `pkg/engine/system_init.go` (absent), `cmd/client/init_versions.go:259` —
  `CarryOverSystem` is instantiated and registered in
  `cmd/client/init_versions.go:259` (client-only), but is absent from
  `pkg/engine/system_init.go`. Server builds and the engine's canonical
  system list do not include it. The integration chain spot-check table in
  Rev 3 listed this as verified, but the Rev 3 entry was erroneous: wiring
  in `init_versions.go` is correct for the desktop client but the system is
  not available in shared/server contexts.

  **Impact**: Low — `CarryOverSystem` is a New Game+ prestige feature that
  only makes sense for single-player/client sessions. However its absence
  from `system_init.go` means server test harnesses and headless integration
  tests cannot exercise the prestige carry-over path.

  **Remediation**: Either (a) move the `CarryOverSystem` registration into
  `system_init.go` behind a `config.EnablePrestige` flag, or (b) add a
  comment in `system_init.go` explaining the intentional client-only wiring.

  **Affected Files**:
  - `pkg/engine/system_init.go` (missing entry)
  - `cmd/client/init_versions.go:259` (only caller)
|---|---|
| `NewMySystem` (dangling) | In a doc comment at `pkg/engine/statmod.go:17` — not a real constructor |
| `NewStaminaRegenSystem` (dangling) | In a doc comment at `pkg/engine/doc.go:213` — documentation example only |
| `pkg/memprofile/profile.go` using `fmt.Printf` | Formal exemption documented at line 195–197 with guideline reference |
| `tradeRouteManager` not added via `world.AddSystem` in client | `RouteManager` is a goroutine-based manager (`Start()`/`Stop()`), not an ECS System; goroutine-based pattern is intentional |
| `ClassProgressionSystem` no callers for `Update` side-effects | Documented design: all progression goes through explicit `LevelUp()` calls |
| VR adapters use stub headset/controller on non-`vr`-tagged builds | Intentional via build tags; `vr_stub_adapters.go` is production-correct fallback |
| GPU bloom / GPU post-process headless stubs | Build-tag guarded (`headless`); correct test-infrastructure pattern |
| WebRTC `signaling.go:76` "simulate" comment | Part of the same simulated-peer stack as G17 — same root gap |
| `initWebRTCFederation` not called in main.go (non-WASM) | WASM-only file (`//go:build js && wasm`); correctly excluded from desktop builds |

---

## Legacy Gap ID Compatibility

Prior audit IDs from the 2026-04-24 `GAPS.md` (G1–G16) are mapped below. Code
comments, CI scripts, or issue-tracker entries referencing these IDs remain valid
cross-references.

| Prior ID | Title | Status in this audit |
|----------|-------|---------------------|
| G1 | OpenXR Controller / Headset Input Stubbed | ✅ RESOLVED (per rev-2 GAPS.md) |
| G2 | Eleven Engine Systems Never Registered | ✅ RESOLVED (per rev-2 GAPS.md) |
| G3 | Mod Browser Unreachable from Any Binary | ✅ RESOLVED (per rev-2 GAPS.md) |
| G4 | Seasonal Event Subsystem Has No Spawner | ✅ RESOLVED (per rev-2 GAPS.md) |
| G5 | Modding System Wired Server-Only | ✅ RESOLVED (per rev-2 GAPS.md) |
| G6 | `vr_webxr_adapters.go` Documented but Missing | ✅ RESOLVED (per rev-2 GAPS.md) |
| G7 | Client Has No Observability/Health Endpoint | ✅ RESOLVED (per rev-2 GAPS.md) |
| G8 | Dead `Server` Type in `pkg/hostplay` | ✅ RESOLVED (per rev-2 GAPS.md) |
| G9 | `EnableShadows` Deprecation Not Enforced | ✅ RESOLVED (per rev-2 GAPS.md) |
| G10 | `ExtendedAchievementSystem` Shadows Wired System | ✅ RESOLVED (per rev-2 GAPS.md) |
| G11 | Menu "Exit Game" Returns Error | ✅ RESOLVED (per rev-2 GAPS.md) |
| G12 | Mobile Portrait Picker Returns Error | ✅ RESOLVED (per rev-2 GAPS.md) |
| G13 | `pkg/companion` Namespace Undocumented | ✅ RESOLVED (per rev-2 GAPS.md) |
| G14-1 | `FederatedMarket.Stop` lacks `sync.Once` | ✅ RESOLVED (per rev-2 GAPS.md) |
| G14-2 | `FederatedMarket.Start` lacks `sync.Once` | ✅ RESOLVED (per rev-2 GAPS.md) |
| G14-3 | `TCPServer.Start` defer-unlock fragility | ✅ RESOLVED (per rev-2 GAPS.md) |
| G14-4 | Non-defer mutex unlock patterns | ✅ RESOLVED (per rev-2 GAPS.md) |
| G14-5 | `startLegacyMetricsMonitor` goroutine leaks | ✅ RESOLVED (per rev-2 GAPS.md) |
| G14-6 | `CleanupTask` stop channel undocumented | ✅ RESOLVED (per rev-2 GAPS.md) |
| G15 | `HotReloadSystem` never registered | ✅ RESOLVED (per rev-2 GAPS.md) |
| G16 | `FileSystemModRepository` unused in production | ✅ RESOLVED (per rev-2 GAPS.md) |
| **Prior Gap 4** | `pkg/memprofile` uses `fmt.Printf` (custom instructions) | ✅ RESOLVED — Exemption formally documented at `pkg/memprofile/profile.go:195–197` |

---

## Appendix: Verified Integration Chain Spot-Checks

| System | Constructor | Registered | Update Fires | Output Consumed |
|--------|------------|-----------|-------------|----------------|
| AchievementSystem | `init_versions.go:140` | `handlers.go:2198` | ✅ | EventBus handlers |
| ExtendedAchievementSystem | `system_init.go:2112` | `system_init.go:2114` | ✅ | Achievement components |
| CompanionAISystem | `init_versions.go:61` | `handlers.go` | ✅ | Companion entity velocity |
| ClassProgressionSystem | `init_versions.go:79` | `handlers.go:2170` | ✅ (no-op body) | via LevelUp() |
| CarryOverSystem | `init_versions.go:259` | `handlers.go:2214` | ✅ | CarryoverComponent |
| ChallengeSystem | `system_init.go:2041` | `system_init.go` | ✅ | DailyChallengeComponent |
| GuildSystem | `init_versions.go:607` | registered | ✅ | GuildComponent |
| GuildCombatBonusSystem | `system_init.go:458` | `system_init.go` | ✅ | GuildCombatBonusComponent |
| HotReloadSystem | `init_versions.go:752` | `init_versions.go:~810` | ✅ | HotReloadComponent |
| CommerceSystem | `system_init.go:1094` | `system_init.go:1095` | ✅ | Commerce events |
| TradeRouteManager (client) | `init_versions.go:635` | `Start()` goroutine | ✅ | Route updates via goroutine |
| TradeRouteManager (server) | `v4_systems.go:~280` | `AddSystem` wrapper | ✅ | ECS update |
| MobileInputAdapter | `mobile.go:308` | entity attachment | 🔴 BROKEN — `processEntityInputs` skips via `*EbitenInput` assertion (G21) |
| EbitenHUDSystem | `game.go:188` | `game.go:363` | ✅ | `game.go:1616` Draw |
| WebRTCTransport | — **never instantiated** — | 🔴 N/A | 🔴 | 🔴 |
