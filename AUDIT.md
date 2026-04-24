# IMPLEMENTATION GAP AUDIT — 2026-04-24
## Repo: opd-ai/venture — Procedural Co-op RPG

> **Scope**: Read-only implementation gap audit per the 2026-04-24 audit directive. The
> previous root-level `AUDIT.md` (2026-04-24 sync audit + legacy 2026-04-23 implementation
> audit) and `GAPS.md` (concurrency safety gaps + legacy 2026-04-23 implementation gaps)
> have been replaced by this fresh report. Findings carried over from the prior implementation
> audit are explicitly re-verified against the current tree. Subpackage audits at
> `pkg/*/AUDIT.md` are unchanged and remain authoritative for their respective packages.

---

## Project Architecture Overview

Venture is a fully-procedural multiplayer action-RPG written in Go (1.24.5) on Ebiten v2.9.3.
Per `README.md` and `ROADMAP.md`, the project's core stated goals are:

- 100 % procedural content (graphics, audio, terrain, items, quests, NPCs) — no asset files
- Single-binary distribution (desktop, WASM, iOS, Android)
- Entity-Component-System (ECS) architecture; deterministic seed-based generation
- High-latency multiplayer (200–5 000 ms, Tor/onion-friendly)
- Cross-server federation, player housing, guild/raid/territory systems
- Procedural audio synthesis (music, SFX, ADPCM voice chat)
- Genre-based theming (fantasy / sci-fi / horror / cyberpunk / post-apocalyptic)
- Modding system (JSON-based, sandboxed)
- VR support — explicitly marked **experimental** (mock adapters only) in `README.md`
  and ROADMAP Priority 4.

`ROADMAP.md` records 25 / 26 stated goals as ✅ Achieved and 1 (VR) as ⚠️ Partial. This audit
is a forward-looking gap scan against that baseline; goals already marked Achieved are
spot-checked rather than re-validated.

### Package responsibility map (audited)

| Package | Responsibility | Status (this audit) |
|---|---|---|
| `pkg/engine` | Core ECS + 343 game systems | Mostly wired; 11 dangling systems (≈ 5 700 LOC, see Findings) |
| `pkg/network` | TCP/WebRTC transport, packets, federation | Wired in client + server |
| `pkg/audio` | Music / SFX / voice synthesis + Manager | Wired (`cmd/client/handlers.go:1083`) |
| `pkg/hostplay` | Embedded LAN-party server | Wired (`cmd/client/util.go:199-220`) |
| `pkg/modding` | JSON mod loader, sandbox, adapter | Wired in **server only** (`cmd/server/main.go:120`); not in client |
| `pkg/vr` | Hardware detection | Wired (`cmd/client/init_versions.go:539`) |
| `pkg/mobile` | iOS / Android input + platform abstraction | Wired via `cmd/mobile/` |
| `pkg/procgen` | 25+ generators (terrain, item, quest, …) | Wired across cmd/client + cmd/server |
| `pkg/companion/learning` | Companion learning manager | Wired (`pkg/engine/companion_learning_system.go:7`) |
| `pkg/observability` | Prometheus exporter + health endpoints | Wired in **server only** (`cmd/server/main.go:1270-1274`) |
| `pkg/network/federation` | Cross-server discovery, auth, sync, market, WebRTC | Wired in server |
| `pkg/network/chat` | Chat channels | Wired via `pkg/engine/chat_system.go` |

---

## Gap Summary

| Category | Count | Critical | High | Medium | Low |
|----------|------:|---------:|-----:|-------:|----:|
| Stubs / TODOs | 14 | 0 | 0 | 0 | 14 |
| Dead Code (dangling systems / providers) | 11 | 0 | 1 | 7 | 3 |
| Partially Wired (client vs server asymmetry, missing files) | 4 | 0 | 1 | 2 | 1 |
| Interface Gaps | 0 | 0 | 0 | 0 | 0 |
| Dependency Gaps | 0 | 0 | 0 | 0 | 0 |
| **Total** | **29** | **0** | **2** | **9** | **18** |

All 14 TODO/stub markers in the entire non-test codebase live in **one file**:
`pkg/engine/vr_openxr_adapters.go` (11 method-level TODOs + 3 file-level integration notes).
There are **zero** `TODO` / `FIXME` / `HACK` / `XXX` markers anywhere else in `pkg/`, `cmd/`,
or `examples/` (verified via `grep -rn 'TODO\|FIXME\|HACK\|XXX' --include='*.go' --exclude='*_test.go'`).

## Implementation Completeness by Package

Counts below are based on direct file reads and grep across the non-test tree.

| Package | Notable Symbols | Implemented | Stubs | Dangling | Notes |
|---------|-----------------|-------------|-------|----------|-------|
| `pkg/engine` | 343 systems, 728 `New*` constructors | 332 | 11 OpenXR methods | 11 systems / providers | See Findings G2–G4 |
| `pkg/network` | `TCPServer`, `Client`, federation, prediction | full | 0 | 0 | Interface-only nettypes enforced |
| `pkg/audio` | `Manager`, music, sfx, voice ADPCM | full | 0 | 0 | Reaches `cmd/client/handlers.go:1083` |
| `pkg/hostplay` | `Server`, `ServerManager`, `InputHandler` | full | 0 | 0 | Used by `cmd/client/util.go:210` |
| `pkg/modding` | `Loader`, `Manager`, `Sandbox`, `ProviderAdapter` | full | 0 | 0 (pkg) | Client never calls `SetModRules` (G5) |
| `pkg/vr` | `Detector` (hardware probe) | full | 0 | 0 | Build-tag-gated OpenXR is the stub layer |
| `pkg/mobile` | input/touch/keyboard adapters, dual-joystick | full | 0 | 0 | `OpenPortraitDialog` returns error on mobile by design |
| `pkg/procgen` | 25 generator subdirs | full | 0 | 0 | Substantive, not placeholder |
| `pkg/companion/learning` | `Manager` + system | full | 0 | 0 | Wired into `pkg/engine/companion_learning_system.go` |
| `pkg/observability` | `MetricsExporter`, `/health*`, `/ready*`, `/metrics` | full | 0 | 0 (pkg) | Client side has no exporter (G7, low) |
| `pkg/network/federation` | discovery, auth, sync, market, webrtc, portal, guild | full | 0 | 0 | Concurrency gaps tracked elsewhere |

---

## Findings

### CRITICAL
*(none — no stated, currently-prioritized goal is blocked by a stub)*

### HIGH

- [x] **G1 — OpenXR controller/headset adapters return zero values** —
  `pkg/engine/vr_openxr_adapters.go:91-230` — All 11 controller/headset methods
  (`GetTrigger`, `GetGrip`, `GetThumbstick`, `IsThumbstickPressed`, `GetButton`,
  `SetHaptic`, `GetHeadOrientation`, `GetHeadPosition`, plus the two
  `IsConnected`/`NewOpenXR*Adapter` constructors that never set `connected = true`)
  carry `TODO(vr-sdk)` markers and return `0` / `false` / `(0,0,0)` placeholders.
  The `connected` field is therefore always `false`, so the runtime selector in
  `pkg/engine/vr_adapter_factory_openxr.go:8-22` never activates the OpenXR path —
  even with `-tags vr`. Severity is **HIGH** rather than CRITICAL because VR is
  documented as **experimental** in `README.md` and `ROADMAP.md` Priority 4.
  **Blocked goal**: ROADMAP Priority 4 items §157-162 (real head tracking + controller
  input + real-hardware validation). Mock fallback (`vr_stub_adapters.go`) keeps non-VR
  builds and `--force-vr` test mode functional.
  **Remediation**: Implement OpenXR loader cgo in `NewOpenXRHeadsetAdapter`
  (`xrCreateInstance` → `xrGetSystem` → `xrCreateSession`), implement `xrLocateViews` for
  pose, set up XrActionSet/XrAction bindings for controller input, implement
  `xrApplyHapticFeedback`. Set `connected = true` only after successful runtime+session
  init. Validation: `go build -tags vr ./...` + new `-tags vr` integration test that
  reports non-zero pose when an OpenXR runtime is available.

- [x] **G2 — Eleven engine systems are defined but never registered with the World** —
  collectively ~5 672 LOC of working ECS code whose `Update()` is never called at runtime.
  Confirmed via `grep -rn 'New<SystemName>' pkg/ cmd/ --include='*.go'` returning hits only
  inside the system's own definition file plus its tests. Files:
  - `pkg/engine/event_calendar_system.go` (273 lines) — seasonal-event lifecycle
  - `pkg/engine/event_quest_system.go` (344 lines) — event-bound quest progression
  - `pkg/engine/event_decoration_system.go` (300 lines) — event décor placement
  - `pkg/engine/event_reward_system.go` (648 lines) — event reward distribution
  - `pkg/engine/destructible_object_system.go` (384 lines) — distinct from
    `pkg/engine/physics/destruction/`; targets *world objects* not terrain
  - `pkg/engine/carry_system.go` (412 lines) — pick-up/carry mechanic
  - `pkg/engine/commerce_system.go` (814 lines) — vendor / shop logic
  - `pkg/engine/extended_achievement_system.go` (815 lines) — meta-achievements layer
  - `pkg/engine/mod_compatibility_system.go` (677 lines) — mod conflict detection
  - `pkg/engine/mod_browser_system.go` (564 lines) — in-game mod browser
  - `pkg/engine/terrain_modification_system.go` (441 lines) — pickaxe / explosion terrain
    destruction (server-authoritative per file header)
  
  Each defines an `Update(entities []*Entity, deltaTime float64)` method that satisfies
  the `System` interface, but no `world.AddSystem(...)` call exists in `cmd/client/`,
  `cmd/server/`, or `pkg/engine/system_init.go`. Per the project's own
  ⚠️ "Zero Dangling Features" rule (custom instructions §"The Integration Chain"),
  step 2 (Instantiation → Registration) is missing for every one of these.
  **Blocked goal**: each system implements a feature implied by README/ROADMAP
  (events / commerce / destructible terrain / mod browser) but is invisible at
  runtime. Severity **MEDIUM** for G3 below; this aggregate finding is **HIGH**
  because of total LOC + breadth.
  **Remediation**: For each system, decide registration owner (`cmd/server/main.go`,
  `cmd/client/handlers.go`, or `pkg/engine/system_init.go`) and call
  `world.AddSystem(NewXxxSystem(...))` in the appropriate genre/feature bucket.
  Validation: `make feature-audit` + new tests asserting the system is present in
  `world.systems` after `InitializeGameSystems`.

### MEDIUM

- [x] **G3 — Mod browser & repository unreachable from any binary** —
  `pkg/engine/mod_browser_system.go:37` (`NewModBrowserSystem`) and
  `pkg/engine/mod_repository_fs.go` (`NewModRepositoryFS` — confirmed via grep) are
  referenced only by their own tests, `pkg/engine/mod_browser_integration_test.go`,
  and one demo `examples/mod_repository_fs_integration/main.go`. Players cannot install
  or browse mods through the running game. **Blocked**: README "Modding System" feature.
  **Remediation**: Wire `NewModBrowserSystem(world)` in `cmd/client/handlers.go` after
  `modManager` creation; call `SetRepository(NewModRepositoryFS(...))` and
  `SetInstallCallback`/`SetUninstallCallback` to bridge into `modding.Manager`. Add a
  smoke test that lists then installs a mod via the system.

- [x] **G4 — Seasonal-event subsystem is internally complete but un-instantiated** —
  `EventCalendarSystem` + `EventQuestSystem` + `EventDecorationSystem` +
  `EventRewardSystem` + the matching components (`EventQuestComponent`,
  `EventDecorationComponent`, `EventRewardComponent`) form a cohesive feature cluster
  (`pkg/engine/event_*_system.go`, ~1 565 LOC). No spawner ever attaches a
  `SeasonalEventComponent` to the world entity, so even if the systems were registered
  they would no-op. **Blocked**: not in stated ROADMAP — qualifies as scaffolded
  feature awaiting decision (keep / wire / remove).
  **Remediation**: Either (a) wire all four systems in `cmd/server/main.go` and seed
  a `SeasonalEventComponent` on the world entity from `pkg/procgen/`, or (b) move
  these files behind a `//go:build seasonal_events` tag and document the deferral.

- [x] **G5 — Modding system not wired into the client** —
  `cmd/server/main.go:120-121` calls `world.SetModRules(modding.NewProviderAdapter(modManager))`,
  but `cmd/client/` never imports `pkg/modding` (verified by `grep -rn 'modding\.' cmd/client/`
  returning zero hits) and never calls `SetModRules`. In single-player / host-and-play
  mode the local world therefore ignores all rule overrides defined in `mods/*.json`.
  **Blocked**: README "JSON-based, sandboxed" modding for desktop play without
  dedicated server.
  **Remediation**: In `cmd/client/handlers.go` after world creation, call
  `mgr := modding.NewManager(); _ = mgr.LoadAll(); world.SetModRules(modding.NewProviderAdapter(mgr))`
  inside the host-and-play / single-player branch (avoid double-loading when an
  embedded server is already managing mods).

- [x] **G6 — `pkg/engine/vr_webxr_adapters.go` documented but missing** —
  `pkg/vr/doc.go:79,84` and `pkg/engine/vr_openxr_adapters.go:9,31-35` both reference
  a `pkg/engine/vr_webxr_adapters.go` file with `//go:build js` constraints for the
  WASM VR path. The file does not exist (`ls pkg/engine/vr_*.go` confirms only
  `vr_openxr_adapters.go`, `vr_stub_adapters.go`, and the factory files). Browser VR
  is therefore unreachable.
  **Blocked**: ROADMAP Priority 4 §160 ("Consider WebXR for WASM builds as alternative")
  marked complete, but only the *documentation* is in place.
  **Remediation**: Create `pkg/engine/vr_webxr_adapters.go` with `//go:build js`,
  implement `VRHeadsetAdapter` + `VRControllerAdapter` via `syscall/js` calls into
  `navigator.xr` (`requestSession("immersive-vr")`, `XRReferenceSpace`, frame
  callbacks). Add a `//go:build js` smoke test.

- [x] **G7 — `cmd/client` does not expose Prometheus/health endpoints** —
  `pkg/observability/MetricsExporter` is used only in `cmd/server/main.go:1270`. The
  client has no `/metrics`, `/healthz`, `/ready`, or `/status` server. ROADMAP §64
  records "Health Check Endpoints" as ✅ Achieved but the implementation is
  server-only. For host-and-play / single-player, no observability surface exists
  on the client process even though the same `MetricsExporter` would work.
  **Blocked**: partial coverage of stated goal; not strictly broken.
  **Remediation**: In `cmd/client/handlers.go` (host-and-play path) start a
  `MetricsExporter` on a separate port (e.g. `localhost:9091`), guarded by an
  opt-in flag `--metrics-port` to avoid surprising desktop users. Register the
  client-side perf monitor and embedded server (if any) as sources.

- [x] **G8 — Dead `Server` constructor in `pkg/hostplay`** —
  `pkg/hostplay/host_and_play.go:45` (`hostplay.New(...)` returning `*Server`) and the
  `*Server` type are not referenced outside the package's own tests. The live
  host-and-play path uses `pkg/hostplay.NewServerManager` (`cmd/client/util.go:210`),
  which is a separate type living in `pkg/hostplay/server_manager.go`. The older
  `*Server` API duplicates `FindAvailablePort`/`Shutdown` semantics and risks
  confusing future callers.
  **Blocked**: none — pure tech debt.
  **Remediation**: Either delete `pkg/hostplay/host_and_play.go` and its test (the
  package retains `server_manager.go` + `input_handler.go` + `state_broadcaster.go`),
  or document `*Server` as the low-level API and `*ServerManager` as the high-level
  wrapper. Validation: `go build ./... && go test ./pkg/hostplay/...`.

- [x] **G9 — `EnableShadows` deprecation path not enforced** —
  `pkg/rendering/lighting/types.go:116-123` marks `EnableShadows` as `// Deprecated:` but
  no `staticcheck` configuration or build-time deprecation warning is emitted; new
  callers can still set the legacy field without notice. The legacy 2026-04-23 finding
  ("`EnableShadows` is a No-Op API field") was *partially* closed — the field now
  forces base-AO on (`pkg/rendering/lighting/system.go:341-349`) — but the
  deprecation lifecycle is not enforced.
  **Blocked**: API contract clarity.
  **Remediation**: Add a `staticcheck.conf` rule or runtime `logrus.Warn` when
  `EnableShadows == true && !AOConfig.Enabled`, plus a unit test asserting the warn
  fires once.

- [x] **G10 — `extended_achievement_system.go` shadows `pkg/engine` achievement system** —
  `pkg/engine/extended_achievement_system.go` (815 LOC) defines an unregistered
  parallel system to the wired achievement system. Whether it is meant to *replace*,
  *augment*, or *be deleted* is undocumented. Two parallel achievement systems can
  silently double-fire rewards if both are ever registered.
  **Blocked**: feature ambiguity; data-corruption risk if both wired.
  **Remediation**: Decide ownership (likely merge "extended" features into the
  primary system or gate behind a flag), then delete the loser. Add an integration
  test asserting only one achievement system instance exists in the world.

### LOW

- [x] **G11 — Menu "Exit Game" returns an error rather than exiting** —
  `pkg/engine/menu_system.go:613` returns `fmt.Errorf("exit not implemented (close window manually)")`.
  This is the lone TODO-equivalent in the engine outside `vr_openxr_adapters.go`. The
  surrounding code closes the menu and bubbles the error — the user must Alt-F4 to
  actually quit. `pkg/engine/AUDIT.md:40` already tracks this as acceptable platform
  variance, but it is still a user-facing rough edge.
  **Remediation**: Inject a `func() error` exit callback from the `Game` (where
  `os.Exit` or `ebiten.Termination` is appropriate per platform) and call it from the
  confirm action.

- [x] **G12 — `pkg/engine/character_creation_mobile.go` — portrait dialog returns error** —
  `OpenPortraitDialog()` on `js || android || ios` returns
  `fmt.Errorf("file dialogs are not supported on mobile/WASM platforms")`
  (`pkg/engine/character_creation_mobile.go:15-17`). On mobile, character creation
  cannot import a custom portrait. This is a deliberate "no zenity on mobile" choice
  (call sites must surface it as a UI message), but no in-game replacement (camera roll
  picker / preset gallery) is wired in.
  **Remediation**: Either route mobile builds to `pkg/mobile.OpenImagePicker` (does
  not yet exist — would require a small native bridge) or replace the error with a
  documented preset-only flow and update the character-creation UI to hide the import
  button on mobile.

- [x] **G13 — `pkg/companion` namespace contains only `learning` subpackage** —
  Per the audit prompt's expectation of a separate companion package distinct from
  `pkg/engine/companion_*.go`, `pkg/companion/` ships a single subdirectory
  `pkg/companion/learning/`. `pkg/engine/companion_*.go` plus `pkg/procgen/companion/`
  cover the rest of the companion feature surface. Not a bug — the structure is
  intentional — but `pkg/companion/doc.go` is missing, so the role of the namespace
  is undocumented.
  **Remediation**: Add `pkg/companion/doc.go` describing the namespace and its
  relationship to `pkg/engine/companion_*.go` + `pkg/procgen/companion/`.

- [x] **G14 — Stop / Start idempotency gaps in federation (carried from prior GAPS.md)** —
  `pkg/network/federation/market.go:92-112` (`FederatedMarket.Start/Stop`) still
  lacks `sync.Once` guards (concurrency Gaps 1 & 2 in the prior root `GAPS.md`).
  These remain LOW for the implementation-gap perspective (no missing feature) but
  are surfaced here so the prior IDs are not lost when this audit replaces the file.
  **Remediation**: See prior `GAPS.md` Gap 1 / Gap 2; both call for `sync.Once`
  fields plus regression tests.

---

## False Positives Considered and Rejected

| Candidate | Reason Rejected |
|-----------|-----------------|
| `pkg/engine/character_creation_mobile.go` (582 B) is "almost certainly a stub" | Verified intentional — file ships `OpenPortraitDialog` returning a guarded error under `//go:build js || android || ios`; the desktop counterpart at `character_creation_desktop.go` provides the real zenity dialog. Build-tag pair is complete. (Surface concern remains as **G12**.) |
| `pkg/engine/guild_component.go` (366 B) is an empty/no-logic struct | Verified intentional — pure-data ECS component (`GuildComponent` with `GuildID/Rank/JoinedAt`) per the project's "components are data only" rule. Logic lives in `pkg/network/federation/guild/manager.go` and `pkg/engine/guild_*_system.go`. |
| `pkg/engine/behavior_tree_entity.go` (589 B) is a stub | Verified intentional — sole purpose is the `entityFromContext` helper used by every leaf-node Tick to assert `aitypes.EntityContext` → `*Entity`. Per stored memory `architecture` (behavior_tree_actions.go:30-35). |
| `pkg/engine/game_clock.go` (1 806 B) is too small to drive time-dependent systems | Verified — file holds two complete clock implementations (`SimulationClock` deterministic + `RealTimeClock` wall-clock). Time-dependent systems read `time.Time` from injected `Clock`-shaped interfaces; this file is not the only time source. |
| `EnableShadows` is a no-op API field (legacy 2026-04-23 finding) | Re-verified: now wired at `pkg/rendering/lighting/system.go:341-349` to force base-AO on. Field still has a deprecation comment but is no longer a no-op. (Tracked separately as **G9** for the deprecation-enforcement gap.) |
| `pkg/observability` defines metrics that are never emitted | Verified: `MetricsExporter` is fully wired in `cmd/server/main.go:1270`; PerformanceMonitor / NetworkServer / World adapters (`pkg/observability/metrics.go:39-59`) are populated from the live server's monitor + network + world objects. |
| `pkg/audio/manager.go` (`Manager`) might be unused given `pkg/engine/audio_manager.go` exists | Verified: both are used. `audio.NewManager` is the new unified manager wired at `cmd/client/handlers.go:1083`; `engine.NewAudioManager` is the legacy compatibility wrapper kept for in-engine systems (`cmd/client/handlers.go:1100` documents this explicitly). Dual-track is intentional. |
| All TODOs flagged in `pkg/engine/vr_openxr_adapters.go` | Confirmed real (G1) — but only because the file is the sole TODO-bearing non-test file in the entire repo. The codebase is otherwise TODO/FIXME-free. |
| `pkg/engine/InitializeGameSystems` complexity (CC=19) | Re-verified: documented justification comment present per `ROADMAP.md:200-203`; sequential system registration. |
| `pkg/network/desync.go` `PeriodicSyncManager.Start/Stop` lacks `sync.Once` | Verified: uses an explicit `running bool` under the existing mutex (legacy GAPS.md Gap 1 explicitly accepts this as an alternative). Not a duplicate of G14. |

---

## Cross-Package Wiring Verification (Phase 3)

| Chain | Result |
|-------|--------|
| Audio: `pkg/engine/audio_manager.go` → `pkg/audio/Manager` → synthesis | ✅ Wired (`cmd/client/handlers.go:1080-1130`) |
| VR: `pkg/engine/vr_openxr_adapters.go` → `pkg/vr/Detector` → game loop | ⚠️ Detector wired (`init_versions.go:539`); OpenXR adapter never reports connected (G1) |
| Modding: `pkg/modding/loader.go` → `pkg/engine` ModRules | ⚠️ Server-only wired (G5) |
| Modding: `ModBrowserSystem` → repository → ECS | ❌ Unreachable from any binary (G3) |
| Observability: `pkg/observability` → `pkg/engine`/`pkg/network` | ⚠️ Server-only wired (G7) |
| Host-and-play: `pkg/hostplay/server_manager.go` → `cmd/client/util.go` | ✅ Wired (`util.go:210`) |
| Host-and-play: `pkg/hostplay/host_and_play.go` → anywhere | ❌ Dead code (G8) |
| Mobile input: `pkg/mobile/*` → `cmd/mobile/` | ✅ Wired |
| Procgen: `pkg/procgen/*` → `cmd/server/entity_spawning.go` + `cmd/client` | ✅ Wired (multiple subgenerators imported) |
| Companion learning: `pkg/companion/learning` → `pkg/engine/companion_learning_system.go` | ✅ Wired |
| Network chat: `pkg/network/chat` → `pkg/engine/chat_system.go` | ✅ Wired |
| Federation: `pkg/network/federation/*` → `cmd/server/main.go` | ✅ Wired |

---

## Methodology

1. **Phase 0** — Read `README.md`, `ROADMAP.md`, `REFACTORING_SUMMARY.md`, `go.mod`, prior
   root `AUDIT.md` + `GAPS.md`, and key subpackage `AUDIT.md` files.
2. **Phase 1** — Direct read of all enumerated high-priority files (VR adapters,
   `vr/detection.go`, `audio/manager.go`, `hostplay/host_and_play.go`,
   `modding/loader.go` + `adapter.go`, mobile character creation pair, `guild_component.go`,
   `behavior_tree_entity.go`, `game_clock.go`, `observability/metrics.go`, plus directory
   listings for `pkg/audio`, `pkg/mobile`, `pkg/procgen`, `pkg/companion`,
   `pkg/network/federation`, `pkg/network/chat`).
3. **Phase 2** — Codebase-wide scans:
   - `grep -rn 'TODO\|FIXME\|HACK\|XXX' --include='*.go' --exclude='*_test.go' .`
     → 15 hits, all but one in `vr_openxr_adapters.go` (the lone other:
     `menu_system.go:613` "exit not implemented" string, G11).
   - `New<Symbol>` constructor cross-reference: enumerated all 728 `New*` constructors in
     `pkg/engine/*.go` (non-test) and verified callers across `pkg/`, `cmd/`, `examples/`,
     producing 168 candidates with ≤ 1 caller. Manual triage isolated the 11
     architecturally-significant `*System` constructors listed in **G2**.
4. **Phase 3** — Wiring verification table above.
5. **Phase 4** — False-positive scrub against the rejection table above.

No source files were modified.
