# Implementation Gaps — 2026-04-25 (rev 3)

> **Rev 3 — forward-pass re-audit (2026-04-25).** This file supersedes the
> `GAPS.md` rev 2 (2026-04-24). All G1–G16 findings from that revision remain
> resolved. Four new gaps G17–G20 were identified and are documented below.
>
> **Rev 2 baseline (2026-04-24)**: All gaps G1–G16 confirmed resolved in code.
>
> **Legacy ID compatibility**: G1–G14 are preserved below with updated status.
> The legacy "Gap 1"–"Gap 6" identifiers (from the prior concurrency-safety
> `GAPS.md`) remain locatable under **G14** sub-items "Prior Gap 1"–"Prior
> Gap 6" — CI scripts and issue-tracker references to those IDs are still valid.
>
> **Status legend**: ✅ RESOLVED | ⚠️ PARTIAL | 🔴 OPEN

---

## G1 — OpenXR Controller / Headset Input Stubbed (Desktop VR)
**Status**: ✅ RESOLVED

- **Prior State**: All 11 controller/headset methods returned zero values; both
  constructors left `connected = false`; 14 `TODO(vr-sdk)` markers covered every
  required OpenXR call. The runtime selector never activated the OpenXR path
  even with `-tags vr`.
- **Resolution**: `pkg/engine/vr_openxr_adapters.go` rewritten — 615 lines of
  cgo OpenXR implementation (`//go:build vr && !js`). The file now contains
  `xrCreateInstance`, `xrGetSystem`, `xrCreateSession`, `xrLocateViews`,
  `xrSyncActions`, `xrGetActionStateFloat/Vector2f/Boolean`, and
  `xrApplyHapticFeedback` call sites. Zero TODO markers remain in the file
  (verified: `grep -rn 'TODO\|FIXME\|HACK\|XXX' --include='*.go'
  --exclude='*_test.go' .` returns 0 hits in the entire non-test codebase).
  The stub fallback (`vr_stub_adapters.go`) still provides graceful degradation
  for non-VR builds.
- **Affected Files**: `pkg/engine/vr_openxr_adapters.go` (fully implemented),
  `pkg/engine/vr_adapter_factory_openxr.go` (selector now fires under `-tags vr`).

---

## G2 — Eleven Engine Systems Defined But Never Registered
**Status**: ✅ RESOLVED

- **Prior State**: 11 systems (~5 672 LOC) had no `world.AddSystem(...)` call
  in any `cmd/` binary or `pkg/engine/system_init.go`.
- **Resolution**: All 11 systems are now registered in
  `pkg/engine/system_init.go`:
  - `CommerceSystem` — line 1094
  - `DestructibleObjectSystem` — line 1864
  - `CarrySystem` — line 1876
  - `EventCalendarSystem` — line 2077
  - `EventQuestSystem` — line 2081
  - `EventDecorationSystem` — line 2085
  - `EventRewardSystem` — line 2089
  - `ModCompatibilitySystem` — line 2103
  - `ModBrowserSystem` — line 2107
  - `ExtendedAchievementSystem` — lines 2112–2114
  - `TerrainModificationSystem` — registered server-side in
    `cmd/server/v4_systems.go:392` and client-side in
    `cmd/client/handlers.go:2284`

---

## G3 — Mod Browser & FS Repository Unreachable from Any Binary
**Status**: ✅ RESOLVED

- **Prior State**: `NewModBrowserSystem` and `NewModRepositoryFS` had no
  production callers; players could not browse or install mods.
- **Resolution**: All four layers are now wired:
  1. `ModBrowserSystem` registered (`system_init.go:2107`).
  2. Install/uninstall callbacks wired in `cmd/client/init_versions.go:653-744`.
  3. Modding client-side fully enabled (G5 resolved).
  4. Repository corrected from `NewInMemoryModRepository()` to
     `NewFileSystemModRepository(modCfg.ModsDirectory)` at
     `cmd/client/init_versions.go:722` (G16 resolved). Players now see
     mods from the `mods/` directory in the in-game mod browser.
- **Affected Files**: `cmd/client/init_versions.go`, `pkg/engine/system_init.go`.

---

## G4 — Seasonal Event Subsystem Has No Spawner
**Status**: ✅ RESOLVED

- **Prior State**: All four event systems were unregistered and no
  `SeasonalEventComponent` was ever attached to the world entity.
- **Resolution**: All four systems registered (G2). `SeasonalEventComponent`
  seeded on the world entity at `pkg/engine/system_init.go:2093` with
  magic constant `seed ^ 0x53454153` ("SEAS") for deterministic derivation.

---

## G5 — Modding System Wired Server-Only
**Status**: ✅ RESOLVED

- **Prior State**: `cmd/client/` had zero references to `pkg/modding`; local
  worlds ignored all `mods/*.json` rule overrides.
- **Resolution**: `cmd/client/init_versions.go:657-700`
  (`initializeModBrowserWiring`) loads `modding.NewManager()`, calls
  `mgr.LoadAll()`, and calls
  `game.World.SetModRules(modding.NewProviderAdapter(sys.modManager))` at
  line 700. Single-player and host-and-play builds now honour mod rules.

---

## G6 — `pkg/engine/vr_webxr_adapters.go` Documented but Missing
**Status**: ✅ RESOLVED

- **Prior State**: Documentation referenced this file but it did not exist;
  WASM VR was unreachable.
- **Resolution**: `pkg/engine/vr_webxr_adapters.go` created — 451 lines,
  `//go:build js`, full `syscall/js` WebXR implementation covering
  `navigator.xr.requestSession`, `XRReferenceSpace`, frame callbacks,
  `XRViewerPose`, and `XRInputSource.gamepad` controller mapping.

---

## G7 — Client Has No Observability/Health Endpoint
**Status**: ✅ RESOLVED

- **Prior State**: `MetricsExporter` was server-only; client had no
  `/metrics`, `/healthz`, `/readyz`, or `/status` endpoint.
- **Resolution**: `cmd/client/init_monitoring.go:154-158`
  (`initObservabilityExporter`) and the `--enable-metrics` flag
  (`cmd/client/util.go:70-72`) provide opt-in client-side Prometheus/health
  endpoints. The exporter is started in the host-and-play path when the flag
  is set.

---

## G8 — Dead `Server` Type in `pkg/hostplay`
**Status**: ✅ RESOLVED

- **Prior State**: `pkg/hostplay/host_and_play.go` defined a `*Server` type
  with no production callers alongside the used `*ServerManager`.
- **Resolution**: `pkg/hostplay/host_and_play.go` has been removed. The
  package now contains only `server_manager.go`, `input_handler.go`,
  `state_broadcaster.go`, and `doc.go`. `cmd/client/util.go:214` correctly
  uses `NewServerManager`.

---

## G9 — `EnableShadows` Deprecation Not Enforced
**Status**: ✅ RESOLVED

- **Prior State**: The `// Deprecated:` comment existed but no runtime warning
  or static-analysis enforcement fired.
- **Resolution**: `pkg/rendering/lighting/system.go:44-51` emits
  `logrus.Warn("EnableShadows is deprecated; set AOConfig.Enabled instead")`
  in the system constructor when the deprecated combination is detected.

---

## G10 — `ExtendedAchievementSystem` Shadows Wired Achievement System
**Status**: ✅ RESOLVED

- **Prior State**: Unregistered parallel system; potential double-fire if both
  wired; ownership undocumented.
- **Resolution**: Decision documented at `pkg/engine/system_init.go:2109-2114`.
  Both systems are registered deliberately: the primary achievement system
  handles kills/quests/crafting; `ExtendedAchievementSystem` handles
  expression/social/meta achievements. The comment explicitly states they are
  complementary and confirms no overlap in event handlers.

---

## G11 — Menu "Exit Game" Returns Error Instead of Exiting
**Status**: ✅ RESOLVED

- **Prior State**: `pkg/engine/menu_system.go:613` returned
  `fmt.Errorf("exit not implemented")`.
- **Resolution**: `pkg/engine/menu_system.go:158-162` exposes
  `SetExitCallback(func() error)`. `cmd/client/handlers.go:3190` injects a
  callback returning `ebiten.Termination`. Exit now works on all desktop and
  WASM platforms.

---

## G12 — Mobile Portrait Picker Returns Error With No Replacement
**Status**: ✅ RESOLVED (Option B implemented)

- **Prior State**: `OpenPortraitDialog` returned a plain `fmt.Errorf`; UI
  still showed the Browse button on mobile.
- **Resolution**: Two layers of resolution:
  1. The error is the typed sentinel `ErrPortraitDialogUnsupported`
     (`pkg/engine/character_creation_mobile.go:30`); the character-creation
     UI detects this sentinel and hides the Browse button on mobile builds.
  2. **Option B implemented**: `pkg/engine/character_creation.go` provides
     a procedural preset-portrait gallery rendered without a file dialog,
     with the implementation anchored by `portraitPresetButtons`,
     `portraitPresets`, `handlePortraitPreset`, and the gallery layout
     logic in that file.  Mobile players select a color-based procedural
     portrait from the gallery instead of importing a custom image.
- **Remaining item** (out of scope / won't-fix for now): A native
  image-picker bridge via `pkg/mobile.OpenImagePicker` (Option A) was not
  implemented; the preset gallery is the accepted alternative per the
  zero-asset, procedural-content philosophy.
- **Affected Files**: `pkg/engine/character_creation_mobile.go`,
  `pkg/engine/character_creation.go`

---

## G13 — `pkg/companion` Namespace Is Undocumented
**Status**: ✅ RESOLVED

- **Prior State**: No `pkg/companion/doc.go`; namespace purpose unclear.
- **Resolution**: `pkg/companion/doc.go` created with full namespace map
  describing the relationship between `pkg/companion/learning/`,
  `pkg/engine/companion_*.go`, and `pkg/procgen/companion/`.

---

## G14 — Carryover: Concurrency-Safety Gaps from Prior Root `GAPS.md`
**Status**: ✅ All six prior gaps resolved

> **Legacy ID note**: The sub-items below were originally labelled "Gap 1"–"Gap 6"
> in the concurrency-safety `GAPS.md` that preceded the 2026-04-24 implementation
> audit. CI scripts and issue-tracker references to those IDs remain valid here.

### Prior Gap 1 — `FederatedMarket.Stop()` lacks `sync.Once`
**Status**: ✅ RESOLVED — `stopOnce sync.Once` field added at
`pkg/network/federation/market.go:24`; `Stop()` at line 117 now uses
`s.stopOnce.Do(...)`.

### Prior Gap 2 — `FederatedMarket.Start()` lacks `sync.Once`
**Status**: ✅ RESOLVED — `startOnce sync.Once` field added at
`pkg/network/federation/market.go:23`; `Start()` at line 98 now uses
`s.startOnce.Do(...)`.

### Prior Gap 3 — `TCPServer.Start()` defer-unlock fragility
**Status**: ✅ RESOLVED

- **Location**: `pkg/network/server.go:213-254`
- **Resolution**: `Start()` now uses a closure pattern — Phase 1 (lock
  acquisition, listener creation, `running = true`) executes inside an
  anonymous closure with `defer s.clientsMu.Unlock()` so every early return
  releases the lock automatically.  Phase 2 (goroutine spawning) runs outside
  the closure with no lock held.  A comment at line 216 references this fix.
  The fragile manual-lock/manual-unlock/re-lock sequence no longer exists.
- **Affected Files**: `pkg/network/server.go`

### Prior Gap 4 — Non-`defer` mutex unlock patterns
**Status**: ✅ RESOLVED

- **Locations** (all now use `defer`):
  - `pkg/procgen/audit_registry.go` — both `RegisterAuditEntry` and
    `GetAuditEntries` use `defer auditMu.Unlock()` immediately after `Lock()`.
  - `pkg/procgen/terrain/async_loader.go` — every goroutine callback wraps
    its critical section in a closure with `defer l.mu.Unlock()`, isolating
    the lock scope and guaranteeing release even on panic.
- **Affected Files**: `pkg/procgen/audit_registry.go`,
  `pkg/procgen/terrain/async_loader.go`

### Prior Gap 5 — `startLegacyMetricsMonitor` goroutine leaks on shutdown
**Status**: ✅ RESOLVED

- **Location**: `cmd/client/init_monitoring.go:54-69`
- **Resolution**: `startLegacyMetricsMonitor` now accepts a
  `context.Context` parameter (line 56).  The goroutine body uses a
  `select` with a `<-ctx.Done()` case (line 62) that stops the ticker
  and returns on context cancellation, matching the clean-shutdown
  pattern already present in `startStabilityMonitor`.
- **Affected Files**: `cmd/client/init_monitoring.go`

### Prior Gap 6 — `CleanupTask` stop channel undocumented
**Status**: ✅ RESOLVED — `pkg/network/projectile_sync.go:441-463`
now carries doc comment: "Returns a stop channel — send a value to stop
the cleanup task."

---

## G15 — `HotReloadSystem` Defined and Fully Implemented But Never Registered
**Status**: ✅ RESOLVED

- **Prior State**: `NewHotReloadSystem` had no production callers; the system
  existed only in the definition file and a demo example.
- **Resolution**: `cmd/client/init_versions.go:748-816`
  (`initializeModBrowserWiring`) now:
  1. Creates a `FileSystemFileWatcher` pointed at `modCfg.ModsDirectory`.
  2. Calls `engine.NewHotReloadSystem(game.World)` and wires the file
     watcher, hash callback, reload callback (re-parse + Manager reload),
     and rollback callback.
  3. Calls `game.World.AddSystem(hotReload)` to register the system.
  4. Creates a world-level entity with `HotReloadComponent` and calls
     `hotReload.StartWatchingMod` for every currently enabled mod so the
     system begins monitoring immediately at startup.
  Mod changes in `mods/` are now detected and applied without a game restart.
- **Affected Files**: `cmd/client/init_versions.go:748-816`

---

## G16 — `FileSystemModRepository` Never Used in Production
**Status**: ✅ RESOLVED

- **Prior State**: `initializeModBrowserWiring` used
  `engine.NewInMemoryModRepository()` (a testing stub), leaving the mod
  browser empty for all players.
- **Resolution**: `cmd/client/init_versions.go:719-722`
  (`initializeModBrowserWiring`) now calls:
  ```go
  // G16 (AUDIT.md): Use the filesystem-backed mod repository …
  sys.modBrowserSys.SetRepository(engine.NewFileSystemModRepository(modCfg.ModsDirectory))
  ```
  `FileSystemModRepository.FetchMods()` scans the `mods/` directory and
  returns every valid `*.json` mod file as a `ModListing`, so players see
  their locally installed mods in the browser immediately.
- **Affected Files**: `cmd/client/init_versions.go:719-722`

---

## Severity Summary (rev 2)

| ID | Title | Status | Severity | Effort |
|----|-------|--------|----------|--------|
| G1 | OpenXR Controller / Headset Input Stubbed | ✅ RESOLVED | — | — |
| G2 | Eleven Engine Systems Never Registered | ✅ RESOLVED | — | — |
| G3 | Mod Browser Unreachable from Any Binary | ✅ RESOLVED | — | — |
| G4 | Seasonal Event Subsystem Has No Spawner | ✅ RESOLVED | — | — |
| G5 | Modding System Wired Server-Only | ✅ RESOLVED | — | — |
| G6 | `vr_webxr_adapters.go` Documented but Missing | ✅ RESOLVED | — | — |
| G7 | Client Has No Observability/Health Endpoint | ✅ RESOLVED | — | — |
| G8 | Dead `Server` Type in `pkg/hostplay` | ✅ RESOLVED | — | — |
| G9 | `EnableShadows` Deprecation Not Enforced | ✅ RESOLVED | — | — |
| G10 | `ExtendedAchievementSystem` Shadows Wired System | ✅ RESOLVED | — | — |
| G11 | Menu "Exit Game" Returns Error | ✅ RESOLVED | — | — |
| G12 | Mobile Portrait Picker Returns Error (no alternative) | ✅ RESOLVED | — | — |
| G13 | `pkg/companion` Namespace Undocumented | ✅ RESOLVED | — | — |
| G14-1 | `FederatedMarket.Stop` lacks `sync.Once` (Prior Gap 1) | ✅ RESOLVED | — | — |
| G14-2 | `FederatedMarket.Start` lacks `sync.Once` (Prior Gap 2) | ✅ RESOLVED | — | — |
| G14-3 | `TCPServer.Start` defer-unlock fragility (Prior Gap 3) | ✅ RESOLVED | — | — |
| G14-4 | Non-defer mutex unlock patterns (Prior Gap 4) | ✅ RESOLVED | — | — |
| G14-5 | `startLegacyMetricsMonitor` goroutine leaks (Prior Gap 5) | ✅ RESOLVED | — | — |
| G14-6 | `CleanupTask` stop channel undocumented (Prior Gap 6) | ✅ RESOLVED | — | — |
| G15 | `HotReloadSystem` never registered | ✅ RESOLVED | — | — |
| G16 | `FileSystemModRepository` unused in production | ✅ RESOLVED | — | — |
| G17 | WebRTC browser-to-browser federation is simulated | 🔴 OPEN | — | see below |
| G18 | `ClassProgressionSystem.Update()` is a no-op | 🔴 OPEN | — | see below |
| G19 | Companion scout behavior uses hardcoded velocity | 🔴 OPEN | — | see below |
| G20 | BehaviorTree ambush node uses random position offset | 🔴 OPEN | — | see below |

---

## G17 — WebRTC Browser-to-Browser Federation is Simulated, Not Real

**Status**: 🔴 OPEN  
**Severity**: HIGH

- **Finding**: The entire `pkg/network/federation/webrtc/` package is a
  simulation harness, not a real WebRTC implementation. The package header at
  `peer.go:4` states: _"This is a stub implementation for testing; real WebRTC
  integration requires `github.com/pion/webrtc/v3`."_ The `Connect()` method
  (`peer.go:77`) calls `simulateConnection()` (line 112), which sleeps for a
  random interval then sets state to `StateConnected` artificially. No ICE
  candidate gathering, no DTLS handshake, and no data channel creation occur.
  The signaling server (`signaling.go:76,299`) is also simulated. The WASM
  initializer `initWebRTCFederation()` (`cmd/client/webrtc_wasm.go:20`) is
  defined but never called from any build entrypoint. `NewWebRTCTransport()`
  (`transport_webrtc.go:31`) is never instantiated in production code — only in
  `_test.go` files. `github.com/pion/webrtc/v3` does not appear in `go.mod`.

- **Impact**: Browser-to-browser (WASM) federation — where two browser tabs
  connect to each other without a dedicated TCP relay server — is listed as a
  feature in `README.md` (line 60: "federation/WebRTC, portals") but is
  non-functional. Desktop server federation over TCP works correctly and is
  unaffected by this gap.

- **Affected Files**:
  - `pkg/network/federation/webrtc/peer.go:4,69,77,112`
  - `pkg/network/federation/webrtc/signaling.go:76,299`
  - `pkg/network/federation/webrtc/nat_traversal.go:188`
  - `pkg/network/federation/transport_webrtc.go:31` (never called in production)
  - `cmd/client/webrtc_wasm.go:20` (initializer never invoked)
  - `go.mod` (missing `github.com/pion/webrtc/v3`)

- **Remediation Path**:
  1. Add `github.com/pion/webrtc/v3` to `go.mod` (`go get
     github.com/pion/webrtc/v3`). The pion library is WASM-safe and requires no
     CGo.
  2. Replace `simulateConnection()` in `peer.go` with `pion.NewPeerConnection`,
     data-channel setup, and SDP offer/answer exchange via the existing signaling
     transport interface.
  3. Call `initWebRTCFederation(clientID)` from `cmd/client/main.go`'s WASM
     startup path and wire the returned `*Peer` into the federation protocol via
     `NewWebRTCTransport`.
  4. Update README to clarify the feature is experimental / in progress until
     implementation is complete.

---

## G18 — `ClassProgressionSystem.Update()` is a No-op

**Status**: 🔴 OPEN  
**Severity**: LOW

- **Finding**: The `Update` method body of `ClassProgressionSystem` at
  `pkg/engine/class_progression_system.go:18–20` contains only a comment:
  _"Currently a stub - progression happens through LevelUp() calls / This
  system could be extended to apply passive effects."_ The system is registered
  in the ECS world at `cmd/client/handlers.go:2170` and
  `cmd/server/v4_systems.go:99`, runs every frame, and consumes a scheduler
  slot while producing zero output.

- **Impact**: Time-based passive class effects (per-second stamina modifiers,
  class-specific buff tick-downs, passive aura reapplication) cannot be
  expressed declaratively via the system. Core progression (XP, level-up
  events, stat bonuses) is unaffected.

- **Affected Files**:
  - `pkg/engine/class_progression_system.go:18–20`
  - `cmd/client/handlers.go:2170` (registration)
  - `cmd/server/v4_systems.go:99` (registration)

- **Remediation**: If passive effects are intentionally deferred, add a
  `// Passive-effect processing is deferred to LevelUp() calls; see GAPS.md
  G18` comment and note in ROADMAP.md. If passive effects are desired, iterate
  entities with the `class_progression` component and apply class-specific regen
  ticks per frame using the component's current class/level state.

---

## G19 — Companion Scout Behavior Uses Hardcoded Diagonal Velocity

**Status**: 🔴 OPEN  
**Severity**: MEDIUM

- **Finding**: `CompanionSystem.executeScout()` at
  `pkg/engine/companion_system.go:494–502` contains the comment _"This is a
  stub - full implementation would use pathfinding"_ and unconditionally sets
  `velocityComp.VX = 80.0`, `velocityComp.VY = 80.0`. A companion placed in
  Scout mode will always move diagonally north-east at maximum speed, ignoring
  walls, owner position, visibility radius, and exploration targets.

- **Impact**: Any companion assigned `BehaviorScout` mode exhibits broken
  movement. The Companion system is wired and registered; the damage is
  isolated to the scout behavior path.

- **Affected Files**:
  - `pkg/engine/companion_system.go:494–502` (stub body)
  - `pkg/engine/companion_system.go:160` (caller: `executeBehavior` switch case)

- **Remediation**: Minimum fix — add angular variation so scouts cycle through
  cardinal directions using a per-companion step counter. Full fix — query
  `SpatialPartition` for walkable cells at increasing radii from the owner,
  drive the companion toward the least-recently-visited cell, and return it to
  the owner when the radius is exhausted.

---

## G20 — BehaviorTree Ambush Node Uses Random Position Offset

**Status**: 🔴 OPEN  
**Severity**: LOW

- **Finding**: The ambush action node in
  `pkg/engine/behavior_tree_advanced_nodes.go:391–396` contains the comment
  _"In a full implementation, this would use pathfinding data."_ The ambush
  position is computed as the entity's current position plus a random
  `(rng.Float64()-0.5)*100` X/Y offset. No cover detection, line-of-sight
  check, or walkability validation is performed; enemies may target positions
  inside solid terrain or in full view of the player.

- **Impact**: Enemy AI for stealth/ambush archetypes (assassins, hunters, traps)
  is degraded. The node evaluates to `NodeRunning` and enemies move to a
  semi-random nearby location rather than seeking genuine cover.

- **Affected Files**:
  - `pkg/engine/behavior_tree_advanced_nodes.go:385–397` (ambush node action)

- **Remediation**: Replace the random offset with a query to
  `SpatialPartition.Query()` filtering for passable tiles with low
  `VisibilityComponent.Visibility` score (populated by the lighting system).
  Fall back to the random offset if no low-visibility tiles are found nearby.

---

> **Rev 4 additions — 2026-04-25.** Gaps G21–G31 found by tracing live
> data-flow through the ECS update loop, HUD rendering path, combat
> callbacks, and the mobile input pipeline. G1–G20 statuses unchanged.

---

## G21 — Mobile Input Completely Non-Functional

**Status**: 🔴 OPEN  
**Severity**: CRITICAL

- **Finding**: `MobileInputAdapter.Type()` returns `"input"` — the same key
  as `EbitenInput.Type()` (`pkg/engine/input_system.go:159`). When
  `cmd/mobile/mobile.go:309` calls `playerEntity.AddComponent(mobileInput)`,
  `Entity.AddComponent` stores by type key (`e.Components[c.Type()] = c`,
  `ecs.go:71`), overwriting the existing `*EbitenInput`. In
  `InputSystem.processEntityInputs` (`input_system.go:1033–1041`):
  ```go
  input, ok := inputComp.(*EbitenInput)
  if !ok { continue }
  ```
  The assertion fails for `*MobileInputAdapter`; the entity is silently
  skipped and `applyInputToVelocity` is never called. The mobile player
  cannot move, attack, or interact. Additionally, `MobileInputAdapter.Update()`
  — which reads live touch positions from `DualJoystickLayout` — is never
  called during the game loop, so joystick state is always stale.

- **Impact**: Complete loss of player control on iOS and Android.

- **Affected Files**:
  - `pkg/engine/input_system.go:1033–1041` (type assertion skip)
  - `pkg/mobile/input_adapter.go:49–52` (type key collision with `EbitenInput`)
  - `cmd/mobile/mobile.go:308–309` (overwrites EbitenInput component)

- **Remediation**:
  1. Extend `processEntityInputs` to handle `InputProvider` interface when
     `*EbitenInput` assertion fails:
     ```go
     if provider, ok := inputComp.(InputProvider); ok {
         s.processInputProvider(entity, provider, deltaTime)
     }
     ```
  2. Call `mobileInput.Update()` each frame — either register a pre-process
     hook in `InputSystem.Update()`, or give `MobileInputAdapter` a unique
     type key `"input_mobile"` so both components coexist and the
     `*EbitenInput` assertion still succeeds for the desktop component.

---

## G22 — XP Double-Award on Every Kill

**Status**: 🔴 OPEN  
**Severity**: HIGH

- **Finding**: Two independent callbacks both call `AwardXP` for the same kill.
  `SetKillCallback` at `pkg/engine/system_init.go:918–931` calls
  `progressionSystem.AwardXP(attacker, xp)` at line 931 for every combat kill.
  Separately, `configureDeathCallback` at `cmd/client/handlers.go:3585`
  wires `createDeathCallback` which calls
  `(*progressionSystem).AwardXP(*playerEntity, xpAmount)` at
  `cmd/client/client_loot.go:512`. Both fire on the same entity death.
  The two XP formulae (`CalculateXPReward` vs `calculateEnemyXP`) yield
  different amounts; the player receives both every kill.

- **Impact**: XP gain is roughly doubled every kill. Players level up twice
  as fast as designed; combat balance and progression pacing are broken.

- **Affected Files**:
  - `pkg/engine/system_init.go:931` (kill callback — primary AwardXP call)
  - `cmd/client/client_loot.go:512` (death callback — duplicate AwardXP call)
  - `cmd/client/handlers.go:3585`

- **Remediation**: Remove `AwardXP` from the kill callback in `system_init.go`
  and retain the death-callback path which handles loot, animation, and
  `DeadComponent` attachment in one transaction. Or merge both XP paths
  into a single calculation shared by the kill callback.

---

## G23 — TalentSystem Stat Accumulation — Old Bonuses Never Removed

**Status**: 🔴 OPEN  
**Severity**: HIGH

- **Finding**: `TalentSystem.applyStatsBonuses` at
  `pkg/engine/talent_system.go:183–212` directly adds flat bonuses to
  `StatsComponent` fields (`stats.Attack += bonuses.FlatDamage`, etc.)
  without first subtracting previously applied values. When `talent.Dirty`
  is re-triggered (reset via `ResetAll()` or reallocation),
  `applyStatsBonuses` is called again — adding the new bonuses on top of
  the already-baked-in old ones. `AttributeAllocationSystem` correctly calls
  `removeAppliedBonuses` at `attribute_allocation_system.go:204` before
  reapplying; `TalentSystem` has no equivalent.

- **Impact**: Talent reset and reallocation yield permanent unbounded stat
  growth. Combat balance is broken for any player who uses the respec flow.

- **Affected Files**:
  - `pkg/engine/talent_system.go:183–212`
  - `pkg/engine/talent_component.go` (no `AppliedBonuses` field — needs adding)

- **Remediation**: Add `AppliedBonuses TalentBonus` to `TalentComponent`.
  Before each `applyStatsBonuses` call, subtract `c.AppliedBonuses` from
  stats, then apply new bonuses and update `c.AppliedBonuses`.

---

## G24 — Desktop HUD Has No Mana Bar

**Status**: 🔴 OPEN  
**Severity**: HIGH

- **Finding**: `HUDSystem.Draw()` at `pkg/engine/hud_system.go:71–99` calls
  `drawHealthBar()`, `drawStatsPanel()`, `drawExperienceBar()`,
  `drawNetworkStatus()`, and `drawTerritoryBonuses()` — no `drawManaBar()`.
  Mana is a primary resource consumed by all spells (100+ mana-related
  systems in `system_init.go`); `ManaComponent` is attached to every player
  entity. The mobile HUD (`pkg/mobile/ui.go`) includes and renders a
  `ManaBar ProgressBar` correctly.

- **Impact**: Desktop players have zero mana feedback. Spell failures occur
  silently; players cannot manage mana economy or gauge regen rate.

- **Affected Files**:
  - `pkg/engine/hud_system.go:71–99`

- **Remediation**: Implement `drawManaBar(screen *ebiten.Image, entity *Entity)`
  modeled on `drawHealthBar`. Read `ManaComponent.Current`/`.Max`, draw a
  blue fill bar below the health bar, and clamp fill fraction to `[0, 1]`.

---

## G25 — Server `consumeItem` Heals Player by `item.Stats.Defense`

**Status**: 🔴 OPEN  
**Severity**: HIGH

- **Finding**: `cmd/server/player_management.go:298`:
  ```go
  healAmount := float64(item.Stats.Defense)
  ```
  Healing potions carry no `Defense` value; that field stores armor
  contribution. All standard consumables heal for 0 HP server-side.

- **Impact**: In dedicated-server mode all potion heals are no-ops.
  In solo play, client-side prediction shows a heal but the server
  corrects to 0 HP, causing a visible snap.

- **Affected Files**:
  - `cmd/server/player_management.go:298`

- **Remediation**: Use `item.Stats.Healing` (add to `ItemStats` if absent)
  or `item.Value` as the heal amount.

---

## G26 — `AttributeEffects` Fields Defined But Never Applied

**Status**: 🔴 OPEN  
**Severity**: MEDIUM

- **Finding**: `DefaultAttributeEffects()` at
  `pkg/engine/attribute_allocation_component.go:93` returns non-zero values
  for `CarryCapPerStr` (5.0), `SpeedBonusPerAgi` (0.5), `ManaRegenPerInt`
  (0.1), `HealthRegenPerVit` (0.05), `StaminaPerEnd` (10.0). The function
  `applyAttributeBonuses` at `attribute_allocation_system.go:113` reads
  these into local variables but never writes them to any component.

- **Impact**: Five advertised per-attribute effects are silently zero.
  Players who invest in STR for carry capacity, AGI for speed, INT for mana
  regen, VIT for health regen, or END for stamina receive no benefit.

- **Affected Files**:
  - `pkg/engine/attribute_allocation_system.go:113–200`
  - `pkg/engine/attribute_allocation_component.go:93–108`

- **Remediation**: Write derived values to the appropriate components
  (`InventoryComponent.MaxCarryWeight`, `VelocityComponent.MaxSpeed`,
  `ManaComponent.Regen`, `HealthComponent.RegenRate`, `StaminaComponent.Max`)
  inside `applyAttributeBonuses`.

---

## G27 — HUD Health Bar Overflows on Overheal

**Status**: 🔴 OPEN  
**Severity**: MEDIUM

- **Finding**: `pkg/engine/hud_system.go:126`:
  ```go
  healthPct := float32(health.Current / health.Max)
  fillWidth  := int(float32(barWidth) * healthPct)
  ```
  No clamping. If `health.Current > health.Max` (overheal from spell/buff),
  `fillWidth > barWidth` and the fill rect is drawn past the background
  boundary, overwriting adjacent HUD elements. `pkg/mobile/ui.go:620`
  correctly clamps to `[0, 1]`.

- **Affected Files**:
  - `pkg/engine/hud_system.go:124–130`

- **Remediation**:
  ```go
  healthPct := float32(health.Current) / float32(health.Max)
  if healthPct > 1.0 { healthPct = 1.0 }
  if healthPct < 0.0 { healthPct = 0.0 }
  ```

---

## G28 — Entity Death Callback Fires Every Frame Until Entity Removed

**Status**: 🔴 OPEN  
**Severity**: MEDIUM

- **Finding**: `combat_system.go:218` (`handleEntityDeath`) calls
  `s.onDeathCallback(entity, attacker)` when `health.Current <= 0` but does
  NOT add `DeadComponent`. On subsequent frames the entity still exists,
  health is still ≤ 0, and the callback fires again every frame.
  The client-side `createDeathCallback` (`client_loot.go:490`) guards with
  `if enemy.HasComponent("dead") { return }` and then adds `NewDeadComponent`,
  but this only protects the one client-side callback — not server-side
  callbacks or any future callbacks added without their own guard.

- **Impact**: Any death callback without an explicit guard processes N times
  per death, where N = frames from death detection to entity removal.
  Duplicated loot, XP, achievements, and analytics events are possible.

- **Affected Files**:
  - `pkg/engine/combat_system.go:218–250`

- **Remediation**: Add `entity.AddComponent(NewDeadComponent())` inside
  `handleEntityDeath` before invoking `onDeathCallback`. Update
  `processEntity` entry guard to skip entities with `DeadComponent`.

---

## G29 — `ClassAffinitySystem.decayStreaks` Uses Hardcoded `currentTime = 0`

**Status**: 🔴 OPEN  
**Severity**: MEDIUM

- **Finding**: `pkg/engine/class_affinity_system.go:103`:
  ```go
  currentTime := 0.0 // Would be game time in real implementation
  ```
  `timeSinceActivity := currentTime - data.LastActivityTime` is always
  negative, so `timeSinceActivity > s.streakDecayTime` is never true.
  Streaks never decay regardless of elapsed time.

- **Impact**: Class affinity streaks are permanent once built. The
  time-based risk/reward design of the streak system is non-functional.

- **Affected Files**:
  - `pkg/engine/class_affinity_system.go:100–113`

- **Remediation**: Accumulate `s.elapsedTime += deltaTime` in `Update()`
  and use it as `currentTime`.

---

## G30 — No Self-Damage Guard in `validateAttackEntities`

**Status**: 🔴 OPEN  
**Severity**: LOW

- **Finding**: `pkg/engine/combat_system.go:278` does not check
  `attacker.ID == target.ID`. Entities can be set as their own target
  (via AOE, reflection spells, or target-selection bugs), triggering the
  full damage pipeline against themselves.

- **Impact**: Combinable with G22 (XP double-award) for self-kill XP exploit.
  Low severity in normal gameplay.

- **Affected Files**:
  - `pkg/engine/combat_system.go:278–310`

- **Remediation**:
  ```go
  if attacker.ID == target.ID { return false }
  ```
  Add at the top of `validateAttackEntities`.

---

## G31 — `CarryOverSystem` Not Registered in `system_init.go`

**Status**: 🔴 OPEN  
**Severity**: LOW

- **Finding**: `CarryOverSystem` is instantiated and registered only in
  `cmd/client/init_versions.go:259`. It is absent from
  `pkg/engine/system_init.go`. The Rev-3 integration chain table in
  AUDIT.md listed it as verified (erroneously: it was only verified in
  the client path, not the shared engine path).

- **Impact**: Server builds and headless integration tests cannot exercise
  the prestige carry-over path. No player-facing regression in desktop solo.

- **Affected Files**:
  - `pkg/engine/system_init.go` (absent)
  - `cmd/client/init_versions.go:259` (client-only registration)

- **Remediation**: Move registration to `system_init.go` behind a
  `config.EnablePrestige` guard, or add a comment documenting intentional
  client-only placement.
