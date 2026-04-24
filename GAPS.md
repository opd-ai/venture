# Implementation Gaps — 2026-04-24 (rev 2)

> **Rev 2 — forward-pass re-audit.** This file supersedes the prior `GAPS.md`
> (2026-04-24 implementation gaps, IDs G1–G14). All prior findings have been
> re-verified against the current tree.
>
> **Rev 2 final pass (2026-04-24)**: All gaps G1–G16 confirmed resolved in
> code. GAPS.md updated to reflect current state.
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
- **Resolution**: All three layers are now wired:
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
     a procedural preset-portrait gallery rendered without a file dialog
     (line 250: `portraitPresetButtons`, line 401: `portraitPresets`,
     line 438: `handlePortraitPreset`, line 570: gallery layout).  Mobile
     players select a color-based procedural portrait from the gallery
     instead of importing a custom image.
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

**All gaps fully resolved.**
