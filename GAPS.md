# Implementation Gaps — 2026-04-24 (rev 2)

> **Rev 2 — forward-pass re-audit.** This file supersedes the prior `GAPS.md`
> (2026-04-24 implementation gaps, IDs G1–G14). All prior findings have been
> re-verified against the current tree.
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
**Status**: ⚠️ PARTIAL

- **Prior State**: `NewModBrowserSystem` and `NewModRepositoryFS` had no
  production callers; players could not browse or install mods.
- **Partial Resolution**: `ModBrowserSystem` is now registered
  (`system_init.go:2107`); install/uninstall callbacks are wired in
  `cmd/client/init_versions.go:653-744`; modding client-side is fully enabled
  (G5 resolved).
- **Remaining Gap**: The production wiring at
  `cmd/client/init_versions.go:720` sets the repository to
  `engine.NewInMemoryModRepository()` — documented in
  `pkg/engine/mod_browser_system.go:424` as "provides a simple in-memory mod
  repository **for testing**". The filesystem-backed implementation
  (`pkg/engine/mod_repository_fs.go`) is used only in
  `examples/mod_repository_fs_integration/main.go:62`.
  Players therefore browse an empty mod list rather than the `mods/` directory.
  See **G16** for the remaining gap.
- **Affected Files**: `cmd/client/init_versions.go:720` (wrong repository type).

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
**Status**: ⚠️ PARTIAL

- **Prior State**: `OpenPortraitDialog` returned a plain `fmt.Errorf`; UI
  still showed the Browse button on mobile.
- **Partial Resolution**: The error is now the typed sentinel
  `ErrPortraitDialogUnsupported` (`pkg/engine/character_creation_mobile.go:30`);
  the character-creation UI detects this sentinel and hides the Browse button
  on mobile builds.
- **Remaining Gap**: No alternative portrait selection path (preset gallery or
  native image picker) is wired in. Mobile players cannot set any custom
  portrait.
- **Implementation Path**:
  - Option A: Implement `pkg/mobile.OpenImagePicker(ctx)` with
    `//go:build android` (gomobile `Intent.ACTION_PICK`) and `//go:build ios`
    (`UIImagePickerController`).
  - Option B: Add a procedural preset-portrait gallery rendered without a file
    dialog.
- **Effort**: medium
- **Affected Files**: `pkg/engine/character_creation_mobile.go`,
  `pkg/mobile/` (Option A), character-creation UI code.

---

## G13 — `pkg/companion` Namespace Is Undocumented
**Status**: ✅ RESOLVED

- **Prior State**: No `pkg/companion/doc.go`; namespace purpose unclear.
- **Resolution**: `pkg/companion/doc.go` created with full namespace map
  describing the relationship between `pkg/companion/learning/`,
  `pkg/engine/companion_*.go`, and `pkg/procgen/companion/`.

---

## G14 — Carryover: Concurrency-Safety Gaps from Prior Root `GAPS.md`
**Status**: Mixed — Prior Gaps 1, 2, 6 ✅ RESOLVED; Prior Gaps 3, 4, 5 🔴 OPEN

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
**Status**: 🔴 OPEN

- **Location**: `pkg/network/server.go:213-248`
- **Issue**: `Start()` calls `s.clientsMu.Lock()` then immediately
  `defer s.clientsMu.Unlock()`. Mid-function, it manually calls
  `s.clientsMu.Unlock()` (to avoid holding the lock while starting goroutines)
  followed by `s.clientsMu.Lock()` again. A future early `return` path added
  between the manual unlock and the manual re-lock would cause the deferred
  unlock to fire on an already-unlocked mutex (double-unlock panic).
- **Remediation**: Remove the `defer` and document the explicit lock/unlock
  sequence with a warning comment, or restructure to hold the mutex only for
  the slice modification scope using a helper function.
- **Effort**: small

### Prior Gap 4 — Non-`defer` mutex unlock patterns
**Status**: 🔴 OPEN

- **Locations**:
  - `pkg/procgen/audit_registry.go:22,24,29,32` — package-level mutex
    with explicit `Lock()`/`Unlock()` (no defer)
  - `pkg/procgen/terrain/async_loader.go:54,57,67,70,81,84,142,144` —
    manual lock/unlock inside goroutine callbacks; a panic in the loop
    body would leave the mutex permanently locked
- **Remediation**: Convert each critical section to use
  `defer s.mu.Unlock()` immediately after `s.mu.Lock()`, or add
  `defer recover()` guards in the goroutine callbacks.
- **Effort**: small

### Prior Gap 5 — `startLegacyMetricsMonitor` goroutine leaks on shutdown
**Status**: 🔴 OPEN

- **Location**: `cmd/client/init_monitoring.go:52-63`
- **Issue**: Goroutine uses `for range ticker.C` with no `select` on a
  cancellation channel or `context.Context`. On clean game exit or signal,
  the goroutine runs until process termination. The companion
  `startStabilityMonitor` at line 98 correctly accepts and uses a context.
- **Remediation**: Add a `context.Context` parameter to
  `startLegacyMetricsMonitor`; change the loop to:
  ```go
  for {
      select {
      case <-ctx.Done():
          ticker.Stop()
          return
      case <-ticker.C:
          // existing metric collection
      }
  }
  ```
- **Effort**: small

### Prior Gap 6 — `CleanupTask` stop channel undocumented
**Status**: ✅ RESOLVED — `pkg/network/projectile_sync.go:441-463`
now carries doc comment: "Returns a stop channel — send a value to stop
the cleanup task."

---

## G15 — `HotReloadSystem` Defined and Fully Implemented But Never Registered
**Status**: 🔴 OPEN — **HIGH**

- **Intended Behavior**: `HotReloadSystem` watches the `mods/` directory for
  JSON changes and hot-reloads mod rules without restarting the game. The
  system has full ECS wiring (`Update(entities []*Entity, dt float64)`),
  reload/rollback/hash callbacks, and a `FileWatcher` injection point.
- **Current State**: `pkg/engine/hot_reload_system.go:45`
  (`NewHotReloadSystem`) is never called in any production binary.
  `grep -rn 'NewHotReloadSystem\|HotReloadSystem' --include='*.go'
  --exclude='*_test.go' .` returns hits only in:
  - `pkg/engine/hot_reload_system.go` (definition)
  - `examples/file_watcher_demo/main.go:142-144` (demo only)
  The companion `HotReloadComponent` (`hot_reload_component.go`) is
  similarly never attached in production.
- **Affected Files**:
  - `pkg/engine/hot_reload_system.go:45` (constructor)
  - `pkg/engine/hot_reload_component.go` (component, never attached)
  - `cmd/client/init_versions.go` (missing registration)
  - `pkg/engine/system_init.go` (alternatively: missing registration)
- **Blocked Goal**: README "Modding System" live hot-reload feature;
  mod changes require game restart without this system.
- **Implementation Path**:
  1. In `cmd/client/init_versions.go` after mod manager creation:
     ```go
     modsDir := // resolved mods/ path
     hotReload := engine.NewHotReloadSystem(game.World)
     hotReload.SetReloadCallback(func(modID string) error {
         return sys.modManager.ReloadMod(modID)
     })
     hotReload.SetRollbackCallback(func(modID string) error {
         return sys.modManager.RollbackMod(modID)
     })
     hotReload.SetFileWatcher(engine.NewFileWatcherFS(modsDir, logger))
     game.World.AddSystem(hotReload)
     ```
  2. Validate: `make feature-audit` should no longer list `HotReloadSystem`
     as dangling; add an integration test asserting the system is present in
     `world.GetSystems()` after init.
- **Dependencies**: G5 (resolved — modManager already in client).
- **Effort**: small

---

## G16 — `FileSystemModRepository` Never Used in Production
**Status**: 🔴 OPEN — **MEDIUM**

- **Intended Behavior**: The in-game mod browser should list mods from the
  local `mods/` directory, enabling players to discover and install mods from
  the filesystem.
- **Current State**: `pkg/engine/mod_repository_fs.go` implements
  `FileSystemModRepository` (described as "production mod repository backed by
  local directory"). At `cmd/client/init_versions.go:720`, the mod browser's
  repository is set to `engine.NewInMemoryModRepository()`, which is
  documented at `pkg/engine/mod_browser_system.go:424` as "simple in-memory
  mod repository **for testing**". A code comment at line 721 reads:
  "A network-backed repository can be injected here in the future."
  `FileSystemModRepository` is used only in
  `examples/mod_repository_fs_integration/main.go:62`.
  Players therefore browse an empty list instead of their `mods/` directory.
- **Affected Files**:
  - `cmd/client/init_versions.go:720` (wrong repository)
  - `pkg/engine/mod_repository_fs.go` (unused in production)
- **Blocked Goal**: "Browse and install mods in-game" — the ModBrowserSystem's
  core use case. Install/uninstall callbacks are correctly wired; only the
  listing source is wrong.
- **Implementation Path**: In `initializeModBrowserWiring`
  (`cmd/client/init_versions.go:720`), replace:
  ```go
  sys.modBrowserSys.SetRepository(engine.NewInMemoryModRepository())
  ```
  with:
  ```go
  sys.modBrowserSys.SetRepository(engine.NewFileSystemModRepository(modsDir))
  ```
  where `modsDir` is the already-resolved mods directory path used by
  `loader.LoadAll()`. Add a test asserting `FetchMods()` returns at least one
  entry when `mods/` contains a valid JSON mod file.
- **Dependencies**: G3 (partially resolved — ModBrowserSystem registered and
  callbacks wired).
- **Effort**: small

---

## Severity Summary (rev 2)

| ID | Title | Status | Severity | Effort |
|----|-------|--------|----------|--------|
| G1 | OpenXR Controller / Headset Input Stubbed | ✅ RESOLVED | — | — |
| G2 | Eleven Engine Systems Never Registered | ✅ RESOLVED | — | — |
| G3 | Mod Browser Unreachable from Any Binary | ⚠️ PARTIAL | — | — |
| G4 | Seasonal Event Subsystem Has No Spawner | ✅ RESOLVED | — | — |
| G5 | Modding System Wired Server-Only | ✅ RESOLVED | — | — |
| G6 | `vr_webxr_adapters.go` Documented but Missing | ✅ RESOLVED | — | — |
| G7 | Client Has No Observability/Health Endpoint | ✅ RESOLVED | — | — |
| G8 | Dead `Server` Type in `pkg/hostplay` | ✅ RESOLVED | — | — |
| G9 | `EnableShadows` Deprecation Not Enforced | ✅ RESOLVED | — | — |
| G10 | `ExtendedAchievementSystem` Shadows Wired System | ✅ RESOLVED | — | — |
| G11 | Menu "Exit Game" Returns Error | ✅ RESOLVED | — | — |
| G12 | Mobile Portrait Picker Returns Error (no alternative) | ⚠️ PARTIAL | MEDIUM | medium |
| G13 | `pkg/companion` Namespace Undocumented | ✅ RESOLVED | — | — |
| G14-1 | `FederatedMarket.Stop` lacks `sync.Once` (Prior Gap 1) | ✅ RESOLVED | — | — |
| G14-2 | `FederatedMarket.Start` lacks `sync.Once` (Prior Gap 2) | ✅ RESOLVED | — | — |
| G14-3 | `TCPServer.Start` defer-unlock fragility (Prior Gap 3) | 🔴 OPEN | MEDIUM | small |
| G14-4 | Non-defer mutex unlock patterns (Prior Gap 4) | 🔴 OPEN | MEDIUM | small |
| G14-5 | `startLegacyMetricsMonitor` goroutine leaks (Prior Gap 5) | 🔴 OPEN | LOW | small |
| G14-6 | `CleanupTask` stop channel undocumented (Prior Gap 6) | ✅ RESOLVED | — | — |
| G15 | `HotReloadSystem` never registered | 🔴 OPEN | HIGH | small |
| G16 | `FileSystemModRepository` unused in production | 🔴 OPEN | MEDIUM | small |
