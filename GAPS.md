# Implementation Gaps — 2026-04-24

> Replaces the prior root `GAPS.md` (concurrency-safety gaps + legacy 2026-04-23
> implementation gaps). Concurrency Gaps 1–6 from the prior file are still open
> and remain authoritative in their original form; they are summarised under **G14**
> below so that downstream references (CI scripts, issue tracker IDs `Gap 1` …
> `Gap 6`) remain locatable. Findings G1, G6 carry forward (re-verified) the
> 2026-04-23 entries "OpenXR Adapter Logic Is Stubbed" and "WebXR Adapter File Is
> Documented but Missing"; the 2026-04-23 "OpenXR Runtime Path Never Selected" entry
> is folded into G1 (single root cause); the 2026-04-23 "EnableShadows Is a No-Op"
> entry is partially closed (now wired) and replaced by **G9** (deprecation
> enforcement).

---

## G1 — OpenXR Controller / Headset Input Stubbed (Desktop VR)
- **Intended Behavior**: Under `-tags vr` builds, `OpenXRHeadsetAdapter` and
  `OpenXRControllerAdapter` should bind to a live OpenXR runtime/session and report
  real pose, axis, button, and haptic state. The runtime adapter selector
  (`pkg/engine/vr_adapter_factory_openxr.go:8-22`) should activate the OpenXR path
  when hardware is present.
- **Current State**: All 11 controller/headset methods return zero values; both
  constructors leave `connected = false`; 11 `TODO(vr-sdk)` markers cover every
  required `xrCreateInstance` / `xrGetSystem` / `xrCreateSession` /
  `xrLocateViews` / `xrSyncActions` / `xrGetActionState*` / `xrApplyHapticFeedback`
  call. The runtime selector therefore never escapes the stub branch even with
  `-tags vr`. **The entire non-test codebase contains 14 TODOs and all 14 are in
  this single file** (verified with `grep -rn 'TODO\|FIXME\|HACK\|XXX'
  --include='*.go' --exclude='*_test.go' .`).
- **Affected Files**: `pkg/engine/vr_openxr_adapters.go:91-230` (constructors +
  every method); `pkg/engine/vr_adapter_factory_openxr.go:8-22` (selector that
  cannot succeed).
- **Blocked Goal**: ROADMAP Priority 4 §157-162 — real head tracking adapter, real
  controller input adapter, removal of "experimental" label.
- **Implementation Path**:
  1. Add cgo block including `<openxr/openxr.h>` with platform-specific
     `LDFLAGS: -lopenxr_loader`.
  2. In `NewOpenXRHeadsetAdapter`: call `xrCreateInstance` →
     `xrGetSystem(XR_FORM_FACTOR_HEAD_MOUNTED_DISPLAY)` →
     `xrCreateSession`; on success store the handles and set `connected = true`.
  3. In `GetHeadOrientation`/`GetHeadPosition`: call `xrLocateViews` against the
     `XR_REFERENCE_SPACE_TYPE_LOCAL` reference space; convert `XrPosef.orientation`
     quaternion to Euler.
  4. In `NewOpenXRControllerAdapter`: build an `XrActionSet` with `XrAction` entries
     for trigger / grip / thumbstick / thumbstick-click / face-buttons;
     `xrSuggestInteractionProfileBindings` for `khr/simple_controller`,
     `valve/index_controller`, `oculus/touch_controller`,
     `microsoft/motion_controller`.
  5. In `IsConnected/GetTrigger/GetGrip/GetThumbstick/IsThumbstickPressed/GetButton`:
     call `xrSyncActions` once per frame (cache in adapter), then
     `xrGetActionStateFloat/Vector2f/Boolean`.
  6. In `SetHaptic`: call `xrApplyHapticFeedback` with `XrHapticVibration{
     amplitude: intensity, duration: ns, frequency: 0 }`.
- **Dependencies**: Khronos OpenXR loader installed in CI/dev; cgo enabled in the
  `vr` build flavour; OpenXR runtime (Monado on Linux CI / SteamVR on dev) for
  `-tags vr` integration tests.
- **Effort**: large

## G2 — Eleven Engine Systems Defined But Never Registered
- **Intended Behavior**: Per the project's own "Zero Dangling Features" rule, every
  `*System` whose `Update(entities, dt)` exists must be passed to
  `world.AddSystem(...)` exactly once during initialization, otherwise its logic
  never runs.
- **Current State**: The following 11 systems (~5 672 LOC total) are constructed
  *only* in their own test files. `grep -rn 'New<Sys>' pkg/ cmd/ --include='*.go'`
  excluding `*_test.go` returns hits **only inside the system's own definition file**:
  - `pkg/engine/event_calendar_system.go:34` — `EventCalendarSystem.Update`
  - `pkg/engine/event_quest_system.go` — `EventQuestSystem.Update`
  - `pkg/engine/event_decoration_system.go` — `EventDecorationSystem.Update`
  - `pkg/engine/event_reward_system.go` — `EventRewardSystem.Update`
  - `pkg/engine/destructible_object_system.go:99` — `DestructibleObjectSystem.Update`
    *(distinct from `pkg/engine/physics/destruction/`)*
  - `pkg/engine/carry_system.go:59` — `CarrySystem.Update`
  - `pkg/engine/commerce_system.go:811` — `CommerceSystem.Update`
  - `pkg/engine/extended_achievement_system.go:502` — `ExtendedAchievementSystem.Update`
  - `pkg/engine/mod_compatibility_system.go` — `ModCompatibilitySystem.Update`
  - `pkg/engine/mod_browser_system.go:37` — `ModBrowserSystem`
  - `pkg/engine/terrain_modification_system.go:67` — `TerrainModificationSystem.Update`
- **Affected Files**: see list above, plus `pkg/engine/system_init.go`,
  `cmd/client/handlers.go`, `cmd/client/init_versions.go`, `cmd/server/main.go`,
  `cmd/server/v8_systems.go` — none of which currently register them.
- **Blocked Goal**: Internal "Zero Dangling Features" rule + implied features
  (events / commerce / destructible terrain / mod browsing / mod conflicts).
- **Implementation Path**: For each system decide owner:
  - Server-authoritative (commerce, terrain modification, destructible objects,
    event reward, mod compatibility) → `cmd/server/main.go` after world init.
  - Client/UI (mod browser, extended achievements UI, event decoration) →
    `cmd/client/handlers.go` or `pkg/engine/system_init.go`.
  - Shared (event calendar, event quest, carry) → `pkg/engine/system_init.go`
    inside a feature-flag gate so single-binary builds opt in.
  Add `world.AddSystem(...)` calls and the corresponding component spawners (e.g.
  attach `SeasonalEventComponent` to the world entity from `pkg/procgen/`).
- **Dependencies**: G4 (seasonal-event spawning), G3 (mod browser repo wiring),
  G10 (extended-achievement vs achievement deduplication).
- **Effort**: large (LOC) but mechanically simple per-system.

## G3 — Mod Browser & FS Repository Unreachable from Any Binary
- **Intended Behavior**: Players should be able to browse and install mods in-game
  via `ModBrowserSystem` backed by `ModRepositoryFS`, integrated with
  `modding.Manager` for sandboxed loading.
- **Current State**: `pkg/engine/mod_browser_system.go:37` `NewModBrowserSystem`
  and `pkg/engine/mod_repository_fs.go` `NewModRepositoryFS` are referenced only
  by their own tests, `pkg/engine/mod_browser_integration_test.go`, and the demo
  `examples/mod_repository_fs_integration/main.go`. No `cmd/client/*.go` or
  `cmd/server/*.go` instantiates either type.
- **Affected Files**: `pkg/engine/mod_browser_system.go`,
  `pkg/engine/mod_repository_fs.go`, `cmd/client/handlers.go` (missing wiring).
- **Blocked Goal**: README "Modding System (JSON-based, sandboxed)" — installation
  flow and in-game browser are unreachable.
- **Implementation Path**:
  1. In `cmd/client/handlers.go` after `modManager` creation, add:
     `repo := engine.NewModRepositoryFS(modsDir, logger)`,
     `browser := engine.NewModBrowserSystem(world)`,
     `browser.SetRepository(repo)`,
     `browser.SetInstallCallback(...)`, `browser.SetUninstallCallback(...)`,
     `world.AddSystem(browser)`.
  2. Wire install/uninstall callbacks into `modding.Manager.LoadFromFile` and a
     mirror of `os.Remove` for uninstall.
  3. Add a smoke test that lists, installs, and uninstalls a fixture mod.
- **Dependencies**: G2 (registration), G5 (client-side mod manager existence).
- **Effort**: medium

## G4 — Seasonal Event Subsystem Has No Spawner
- **Intended Behavior**: A seasonal-event feature consisting of
  `EventCalendarSystem` + `EventQuestSystem` + `EventDecorationSystem` +
  `EventRewardSystem` should manage upcoming/active/ending phases driven by a
  `SeasonalEventComponent` attached to the world entity.
- **Current State**: All four systems (~1 565 LOC) are well-formed but
  unregistered (G2). Additionally, no procgen path attaches a
  `SeasonalEventComponent` to the world entity, so even after registration the
  systems would no-op. `grep -rn SeasonalEventComponent` returns hits only inside
  the engine system files themselves.
- **Affected Files**: `pkg/engine/event_*_system.go`, `pkg/engine/event_*_component.go`,
  `pkg/procgen/` (missing seeder).
- **Blocked Goal**: feature implied by 4 systems' presence; not on current
  ROADMAP; qualifies as scaffolded feature awaiting decision.
- **Implementation Path**:
  Either (a) implement a `pkg/procgen/event/` generator that produces seasonal
  events from `(seed, calendar_date)`, attach `SeasonalEventComponent` to the
  world entity in `cmd/server/main.go`, register the four systems; or
  (b) gate all four files behind `//go:build seasonal_events` and document the
  deferral in `pkg/engine/AUDIT.md`.
- **Dependencies**: G2.
- **Effort**: large (path a) / small (path b)

## G5 — Modding System Wired Server-Only
- **Intended Behavior**: All play modes — dedicated server, embedded host-and-play,
  pure single-player — should respect rule overrides defined in `mods/*.json`.
- **Current State**: `cmd/server/main.go:120-121` calls
  `world.SetModRules(modding.NewProviderAdapter(modManager))`, but
  `cmd/client/` contains zero references to `pkg/modding`. In single-player or
  host-and-play, `world.ModRules` is `nil` and `world.GetModRuleFloat64` returns
  the default at every call site.
- **Affected Files**: `cmd/client/handlers.go`, `cmd/client/util.go` (host-and-play
  init), `pkg/engine/ecs.go:520-1053` (consumers).
- **Blocked Goal**: README "Modding System" — single-player and host-and-play
  honour neither rule overrides nor mod events.
- **Implementation Path**: In `cmd/client/handlers.go` after world creation, add:
  ```go
  if !connectingToRemoteServer {
      mgr := modding.NewManager()
      _, _ = mgr.LoadAll() // tolerate empty mods dir
      world.SetModRules(modding.NewProviderAdapter(mgr))
  }
  ```
  When host-and-play spawns an embedded server (`cmd/client/util.go:210`), have
  the embedded server own the manager and let the client read overrides through
  its handle to the same world (already the case). Add a unit test asserting
  `world.GetModRules() != nil` after single-player init.
- **Dependencies**: none.
- **Effort**: small

## G6 — `pkg/engine/vr_webxr_adapters.go` Documented but Missing
- **Intended Behavior**: WASM (`//go:build js`) builds should provide WebXR-backed
  `VRHeadsetAdapter` + `VRControllerAdapter` implementations using `navigator.xr`.
- **Current State**: `pkg/vr/doc.go:79,84` and
  `pkg/engine/vr_openxr_adapters.go:9,31-35` reference a future
  `pkg/engine/vr_webxr_adapters.go` with `//go:build js` constraints. The file
  does not exist (`ls pkg/engine/vr_*.go` confirms only `vr_openxr_adapters.go`,
  `vr_stub_adapters.go`, the factory, controller, UI files). WASM VR is therefore
  unreachable.
- **Affected Files**: `pkg/engine/vr_webxr_adapters.go` (does not exist),
  `pkg/vr/doc.go:79`, `pkg/engine/vr_openxr_adapters.go:9`.
- **Blocked Goal**: ROADMAP Priority 4 §160 marked complete in documentation but
  only the *research* note is in place — no implementation.
- **Implementation Path**:
  1. Create `pkg/engine/vr_webxr_adapters.go` with `//go:build js`.
  2. Implement `WebXRHeadsetAdapter` using `syscall/js` to call
     `navigator.xr.requestSession("immersive-vr")`, hold an `XRSession`, register
     a frame callback, read `XRViewerPose`.
  3. Implement `WebXRControllerAdapter` reading `XRInputSource.gamepad` axes and
     buttons each frame; map standard XR Controller mapping to trigger / grip /
     thumbstick / buttons.
  4. Update `pkg/engine/vr_adapter_factory.go` (default factory) to select
     WebXR variants under `//go:build js`.
  5. Add a `//go:build js` smoke test in `pkg/engine/`.
- **Dependencies**: none structural.
- **Effort**: medium

## G7 — Client Has No Observability/Health Endpoint
- **Intended Behavior**: `pkg/observability.MetricsExporter` provides `/metrics`,
  `/health`, `/healthz`, `/ready`, `/readyz`, `/status` endpoints. ROADMAP §64
  records this as ✅ Achieved.
- **Current State**: The exporter is instantiated only in
  `cmd/server/main.go:1270` (`initializeMetricsExporter`). The desktop / WASM /
  mobile client process exposes none of these endpoints, so host-and-play and
  single-player runs cannot be probed even though the same `MetricsExporter`
  would work.
- **Affected Files**: `cmd/client/handlers.go` (missing); `cmd/client/init_versions.go`
  (missing); `pkg/observability/metrics.go:170-175` (endpoint registration).
- **Blocked Goal**: Stated goal is partially met (server-side); not strictly broken.
- **Implementation Path**: Add an opt-in `--metrics-port int` flag (default `0` =
  disabled). When > 0, start a `MetricsExporter` from `cmd/client/handlers.go`
  bound to `localhost:<port>`. Register a client `PerformanceMonitor` and (in
  host-and-play mode) the embedded server.
- **Dependencies**: none.
- **Effort**: small

## G8 — Dead `Server` Type in `pkg/hostplay`
- **Intended Behavior**: `pkg/hostplay` exposes a single canonical entry point
  for host-and-play.
- **Current State**: `pkg/hostplay/host_and_play.go:36-104` defines `Server` +
  `New` + `FindAvailablePort` + `Shutdown`. The live integration in
  `cmd/client/util.go:210` uses `pkg/hostplay/server_manager.go`'s
  `NewServerManager`. The `*Server` API has no production caller (verified by
  grep) — only its own test file. Two parallel APIs invite future regressions.
- **Affected Files**: `pkg/hostplay/host_and_play.go`,
  `pkg/hostplay/host_and_play_test.go`, `pkg/hostplay/server_manager.go`.
- **Blocked Goal**: none — pure tech debt / API clarity.
- **Implementation Path**: Either delete `host_and_play.go` (and its test) since
  `server_manager.go` covers all production needs, or document `*Server` as the
  low-level primitive and `*ServerManager` as the high-level wrapper, plus add
  `pkg/hostplay/doc.go` clarifying the layering.
- **Dependencies**: none.
- **Effort**: small

## G9 — `EnableShadows` Deprecation Not Enforced
- **Intended Behavior**: Deprecated config fields should generate compile-time or
  load-time warnings so callers can migrate before removal.
- **Current State**: `pkg/rendering/lighting/types.go:116-123` carries a
  `// Deprecated:` comment, and the field is now wired (no longer a no-op — see
  `pkg/rendering/lighting/system.go:341-349` where it forces base AO on). However
  no `staticcheck.conf` rule and no runtime warning fires when callers set
  `EnableShadows = true && AOConfig.Enabled == false`.
- **Affected Files**: `pkg/rendering/lighting/types.go:116`,
  `pkg/rendering/lighting/system.go:341`, no `staticcheck.conf`.
- **Blocked Goal**: API contract clarity; latent risk that callers continue to
  rely on the legacy toggle past its removal.
- **Implementation Path**: Add a one-shot `logrus.Warn` in the lighting system
  constructor when the deprecated combination is detected. Optionally add a
  `staticcheck.conf` SA1019 enforcement rule.
- **Dependencies**: none.
- **Effort**: small

## G10 — `ExtendedAchievementSystem` Shadows the Wired Achievement System
- **Intended Behavior**: A single, well-defined achievement system tracks unlocks
  and rewards.
- **Current State**: `pkg/engine/extended_achievement_system.go` (815 LOC) defines
  a parallel system never registered (G2). Whether it should *replace*, *augment*,
  or *be deleted* relative to the wired achievement system is undocumented. If
  ever both are registered, achievement events would double-fire.
- **Affected Files**: `pkg/engine/extended_achievement_system.go`,
  `pkg/engine/achievement_*` (the wired system), `pkg/engine/system_init.go`.
- **Blocked Goal**: Feature ambiguity; data-corruption risk if both wired.
- **Implementation Path**: Decide ownership (most likely merge "extended"
  features into the primary system or gate behind a `//go:build extended_achievements`
  tag), delete the loser, add an integration test asserting only one
  achievement system instance exists in `world.systems` after init.
- **Dependencies**: G2.
- **Effort**: medium

## G11 — Menu "Exit Game" Returns Error Instead of Exiting
- **Intended Behavior**: Selecting "Exit Game" → "Yes" in the in-game menu should
  cleanly terminate the process.
- **Current State**: `pkg/engine/menu_system.go:613` returns
  `fmt.Errorf("exit not implemented (close window manually)")`. Confirm action
  closes the menu but the error is swallowed by the menu dispatcher; the user must
  Alt-F4 (desktop) or task-kill (mobile) to actually quit. Documented as
  acceptable in `pkg/engine/AUDIT.md:40` but still a user-facing rough edge.
- **Affected Files**: `pkg/engine/menu_system.go:606-619`,
  `pkg/engine/game.go` (would own the exit callback).
- **Blocked Goal**: UX completeness only.
- **Implementation Path**: Add a `func() error` exit callback field on
  `MenuSystem`, set it from `cmd/client/handlers.go` to a function that calls
  `ebiten.Termination` (desktop / WASM) or platform-appropriate shutdown
  (mobile). Replace the `fmt.Errorf` with the callback invocation.
- **Dependencies**: none.
- **Effort**: small

## G12 — Mobile Portrait Picker Returns Error With No Replacement
- **Intended Behavior**: Mobile players should be able to choose a portrait
  during character creation (preset gallery, system image picker, or camera roll).
- **Current State**: `pkg/engine/character_creation_mobile.go:15-17`
  `OpenPortraitDialog` returns
  `fmt.Errorf("file dialogs are not supported on mobile/WASM platforms")` and no
  alternative flow is wired into the character-creation UI. The desktop
  counterpart at `pkg/engine/character_creation_desktop.go:17-41` uses zenity
  successfully. Build-tag pair is structurally complete; UX is not.
- **Affected Files**: `pkg/engine/character_creation_mobile.go`, `pkg/mobile/`
  (would own a native bridge), character-creation UI in `pkg/engine/`.
- **Blocked Goal**: README mobile-support coverage of character-creation parity.
- **Implementation Path**: Either (a) add `pkg/mobile.OpenImagePicker(ctx)` with
  `//go:build android` calling `Intent.ACTION_PICK` via gomobile and
  `//go:build ios` calling `UIImagePickerController`; or (b) replace the error
  with a documented preset-gallery flow and hide the import button on mobile in
  the character-creation UI. Update the call sites of `OpenPortraitDialog` to
  surface the platform-appropriate flow.
- **Dependencies**: none.
- **Effort**: medium

## G13 — `pkg/companion` Namespace Is Undocumented
- **Intended Behavior**: A top-level `pkg/companion/` namespace should document
  its purpose and relationship to `pkg/engine/companion_*.go` and
  `pkg/procgen/companion/`.
- **Current State**: `pkg/companion/` contains a single subpackage `learning/`
  (companion learning manager). No `pkg/companion/doc.go` exists. New contributors
  cannot tell from the namespace what belongs there vs in `pkg/engine`.
- **Affected Files**: `pkg/companion/` (missing `doc.go`).
- **Blocked Goal**: Documentation completeness.
- **Implementation Path**: Add `pkg/companion/doc.go` with a `// Package
  companion` block describing the namespace as the home of cross-cutting
  companion subsystems decoupled from ECS, listing `learning` and any planned
  subpackages, and explicitly noting that ECS-specific companion code lives in
  `pkg/engine/companion_*.go`.
- **Dependencies**: none.
- **Effort**: small

## G14 — Carryover: Concurrency-Safety Gaps from Prior Root `GAPS.md`
- **Intended Behavior**: Federation `Start`/`Stop` lifecycle should be idempotent;
  long-lived background goroutines should accept cancellation; mutex unlocks
  should be deferred for panic safety.
- **Current State**: All six gaps from the prior 2026-04-24 root `GAPS.md`
  (concurrency-safety audit) remain open at the time of this audit:
  - **Prior Gap 1** — `FederatedMarket.Stop()` lacks `sync.Once`
    (`pkg/network/federation/market.go:112`)
  - **Prior Gap 2** — `FederatedMarket.Start()` lacks `sync.Once`
    (`pkg/network/federation/market.go:92-106`)
  - **Prior Gap 3** — `TCPServer.Start()` defer-unlock fragility
    (`pkg/network/server.go:213-248`)
  - **Prior Gap 4** — Non-`defer` mutex unlock patterns
    (`pkg/procgen/audit_registry.go:22,29`,
    `pkg/companion/learning/manager.go:112,155`,
    `pkg/procgen/legendary/manager.go:104,143,395,406,431,439,478`,
    `pkg/world/territory.go:195-206`,
    `pkg/procgen/terrain/async_loader.go:54,67,81`)
  - **Prior Gap 5** — `startLegacyMetricsMonitor` goroutine has no cancellation
    (`cmd/client/init_monitoring.go:52-63`)
  - **Prior Gap 6** — `ProjectileNetworkSync.CleanupTask()` stop channel
    undocumented (`pkg/network/projectile_sync.go:441-463`)
- **Affected Files**: see list above.
- **Blocked Goal**: Concurrency robustness; not blocking any stated feature.
- **Implementation Path**: Per the prior root `GAPS.md` Gaps 1–6 — add
  `sync.Once` guards (Gaps 1–2), refactor mixed defer/manual lock pattern
  (Gap 3), convert manual `Lock`/`Unlock` to `defer` (Gap 4), thread
  `context.Context` into `startLegacyMetricsMonitor` and
  `ProjectileNetworkSync.CleanupTask` (Gaps 5–6).
- **Dependencies**: none cross-gap.
- **Effort**: small per gap, medium aggregate.

---

## Severity Summary

| ID | Title | Severity | Effort |
|----|-------|----------|--------|
| G1 | OpenXR Controller / Headset Input Stubbed | HIGH | large |
| G2 | Eleven Engine Systems Defined But Never Registered | HIGH | large |
| G3 | Mod Browser & FS Repository Unreachable | MEDIUM | medium |
| G4 | Seasonal Event Subsystem Has No Spawner | MEDIUM | large / small |
| G5 | Modding System Wired Server-Only | MEDIUM | small |
| G6 | `vr_webxr_adapters.go` Documented but Missing | MEDIUM | medium |
| G7 | Client Has No Observability/Health Endpoint | MEDIUM | small |
| G8 | Dead `Server` Type in `pkg/hostplay` | MEDIUM | small |
| G9 | `EnableShadows` Deprecation Not Enforced | MEDIUM | small |
| G10 | `ExtendedAchievementSystem` Shadows Wired System | MEDIUM | medium |
| G11 | Menu "Exit Game" Returns Error | LOW | small |
| G12 | Mobile Portrait Picker Returns Error | LOW | medium |
| G13 | `pkg/companion` Namespace Undocumented | LOW | small |
| G14 | Concurrency-Safety Gaps Carryover (6 sub-items) | LOW–HIGH | small each |
