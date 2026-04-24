# IMPLEMENTATION GAP AUDIT — 2026-04-24 (rev 2)
## Repo: opd-ai/venture — Procedural Co-op RPG

> **Scope**: Forward-pass re-audit executed 2026-04-24 (rev 2). The prior root `AUDIT.md`
> (2026-04-24) and `GAPS.md` (concurrency safety gaps + 2026-04-24 implementation gaps,
> IDs G1–G14) are superseded by this document. All prior findings (G1–G14 plus legacy
> "Gap 1"–"Gap 6") have been re-verified against the current tree; resolved gaps are marked
> ✅ closed and carry-forwards retain their original IDs. New findings (G15–G16) are appended.
> Subpackage audits at `pkg/*/AUDIT.md` are unchanged and remain authoritative for their
> respective packages. This file is referenced by `cmd/client/handlers.go:115,594,766,853,1136`
> and `cmd/server/main.go:130–521`; the path and major section headings are preserved.

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

### Package responsibility map (audited — rev 2)

| Package | Responsibility | Status (rev 2) |
|---|---|---|
| `pkg/engine` | Core ECS + 358 system constructors | 357 wired; 1 dangling (`HotReloadSystem`, G15) |
| `pkg/network` | TCP/WebRTC transport, packets, federation | Wired in client + server |
| `pkg/audio` | Music / SFX / voice synthesis + Manager | Wired (`cmd/client/handlers.go:1083`) |
| `pkg/hostplay` | Embedded LAN-party server | Wired (`cmd/client/util.go:199-220`); legacy `*Server` type removed |
| `pkg/modding` | JSON mod loader, sandbox, adapter | ✅ Wired in client **and** server (`cmd/client/init_versions.go:657-700`) |
| `pkg/vr` | Hardware detection | Wired; OpenXR cgo impl merged (`pkg/engine/vr_openxr_adapters.go`); WebXR impl merged (`vr_webxr_adapters.go`) |
| `pkg/mobile` | iOS / Android input + platform abstraction | Wired via `cmd/mobile/`; portrait dialog surfaced as typed error |
| `pkg/procgen` | 25+ generators (terrain, item, quest, …) | Wired across cmd/client + cmd/server |
| `pkg/companion/learning` | Companion learning manager | Wired (`pkg/engine/companion_learning_system.go:7`) |
| `pkg/observability` | Prometheus exporter + health endpoints | ✅ Wired in client (opt-in `--enable-metrics`, `cmd/client/init_monitoring.go:157`) and server |
| `pkg/network/federation` | Cross-server discovery, auth, sync, market, WebRTC | Wired in server; `FederatedMarket` sync.Once guards applied |
| `pkg/network/chat` | Chat channels | Wired via `pkg/engine/chat_system.go` |

---

## Gap Summary (rev 2)

| Category | Count | Critical | High | Medium | Low |
|----------|------:|---------:|-----:|-------:|----:|
| Stubs / TODOs | 0 | 0 | 0 | 0 | 0 |
| Dead Code (dangling systems / providers) | 2 | 0 | 1 | 1 | 0 |
| Partially Wired | 3 | 0 | 0 | 1 | 2 |
| Concurrency Safety | 3 | 0 | 0 | 1 | 2 |
| Interface Gaps | 0 | 0 | 0 | 0 | 0 |
| Dependency Gaps | 0 | 0 | 0 | 0 | 0 |
| **Total open** | **8** | **0** | **1** | **3** | **4** |

**TODOs re-verified**: `grep -rn 'TODO\|FIXME\|HACK\|XXX' --include='*.go' --exclude='*_test.go' .`
returns **zero hits** in the non-test codebase. The 14 `TODO(vr-sdk)` markers that existed in
`pkg/engine/vr_openxr_adapters.go` in the prior audit have been replaced by a full cgo OpenXR
implementation (615 lines, 22 OpenXR API call sites). All other previously-open G2–G13 findings
are ✅ closed (see Findings section).

## Implementation Completeness by Package (rev 2)

| Package | Notable Symbols | Implemented | Stubs | Dangling | Notes |
|---------|-----------------|-------------|-------|----------|-------|
| `pkg/engine` | 358 system constructors | 357 | 0 | 1 (`HotReloadSystem`, G15) | G2–G4 fully resolved |
| `pkg/network` | `TCPServer`, `Client`, federation, prediction | full | 0 | 0 | Interface-only nettypes enforced; concurrency gaps G14 sub-items 3,4,5 open |
| `pkg/audio` | `Manager`, music, sfx, voice ADPCM | full | 0 | 0 | Voice transport wired (`cmd/client/handlers.go:773`) |
| `pkg/hostplay` | `ServerManager`, `InputHandler` | full | 0 | 0 | Legacy `*Server` type removed; `cmd/client/util.go:214` uses `ServerManager` |
| `pkg/modding` | `Loader`, `Manager`, `Sandbox`, `ProviderAdapter` | full | 0 | 0 | ✅ Client calls `SetModRules` (`cmd/client/init_versions.go:700`) |
| `pkg/vr` | `Detector`, OpenXR/WebXR adapters | full (hardware-gated) | 0 | 0 | cgo (`//go:build vr`) + WebXR (`//go:build js`) both implemented |
| `pkg/mobile` | input/touch/keyboard adapters, dual-joystick | full | 0 | 0 | Portrait dialog: typed `ErrPortraitDialogUnsupported`; UI hides Browse on mobile |
| `pkg/procgen` | 25 generator subdirs | full | 0 | 0 | Substantive, not placeholder |
| `pkg/companion/learning` | `Manager` + system | full | 0 | 0 | Wired into `pkg/engine/companion_learning_system.go:7` |
| `pkg/companion` | top-level namespace | full | 0 | 0 | `doc.go` added; namespace map documented |
| `pkg/observability` | `MetricsExporter`, `/health*`, `/ready*`, `/metrics` | full | 0 | 0 | ✅ Client opt-in via `--enable-metrics`; G7 closed |
| `pkg/network/federation` | discovery, auth, sync, market, webrtc, portal, guild | full | 0 | 0 | `FederatedMarket` sync.Once guards added (G14 Gaps 1–2 closed) |

---

## Findings

### CRITICAL
*(none — no stated, currently-prioritized goal is blocked by a stub)*

### HIGH

- [ ] **G15 — `HotReloadSystem` defined and fully implemented but never registered** —
  `pkg/engine/hot_reload_system.go:45` (`NewHotReloadSystem`) provides a complete ECS
  system with `Update(entities []*Entity, deltaTime float64)` (line 92), callbacks for
  mod reload/rollback/hash, a `FileWatcher` injection point, and a companion
  `HotReloadComponent` (`hot_reload_component.go`). No `cmd/client/*.go`,
  `cmd/server/*.go`, or `pkg/engine/system_init.go` instantiates or registers it
  (verified: `grep -rn 'NewHotReloadSystem\|HotReloadSystem' --include='*.go'
  --exclude='*_test.go' .` returns only the definition file and `examples/file_watcher_demo/main.go`).
  Without registration the live-mod-reload feature — already surfaced in the demo —
  is invisible to every runtime binary.
  **Blocked goal**: README "Modding System"; live hot-reload is its most visible in-session
  feature and is cited in `examples/file_watcher_demo/main.go:142-144`.
  **Remediation**: In `cmd/client/init_versions.go` (after mod manager creation):
  1. Instantiate: `hotReload := engine.NewHotReloadSystem(game.World)`
  2. Wire callbacks: `SetReloadCallback`, `SetRollbackCallback`, `SetHashCallback`
     delegating to `sys.modManager`.
  3. Wire file watcher: `hotReload.SetFileWatcher(engine.NewFileWatcherFS(modsDir, logger))`
  4. Register: `game.World.AddSystem(hotReload)`
  Validation: `make feature-audit` must no longer report `HotReloadSystem` as dangling; add
  an integration test asserting the system is present in `world.GetSystems()` after init.
  **Effort**: small

### MEDIUM

- [ ] **G16 — `FileSystemModRepository` never used in production; ModBrowserSystem backed by
  in-memory stub** —
  `pkg/engine/mod_repository_fs.go` provides `FileSystemModRepository`, described as "production
  mod repository backed by local directory" (file header), implementing the `ModRepository`
  interface. In production (`cmd/client/init_versions.go:720`), `initializeModBrowserWiring`
  sets the repository to `engine.NewInMemoryModRepository()`, which is documented in
  `pkg/engine/mod_browser_system.go:424` as "provides a simple in-memory mod repository
  **for testing**". `FileSystemModRepository` is used only in
  `examples/mod_repository_fs_integration/main.go:62`. Players launching the game therefore
  browse an empty (in-memory) mod list rather than the `mods/` directory on disk.
  **Blocked goal**: "Browse and install mods in-game" feature implied by the ModBrowserSystem's
  presence; the install/uninstall callbacks are correctly wired but the listing source is wrong.
  **Remediation**: In `initializeModBrowserWiring` (`cmd/client/init_versions.go:720`), replace:
  ```go
  sys.modBrowserSys.SetRepository(engine.NewInMemoryModRepository())
  ```
  with:
  ```go
  sys.modBrowserSys.SetRepository(engine.NewFileSystemModRepository(modsDir))
  ```
  where `modsDir` is the same path already resolved for `loader.LoadAll()`. Add a test
  asserting `FetchMods()` returns at least one entry when `mods/` contains a valid JSON mod.
  **Effort**: small

- [ ] **G12 (carry-forward, partially resolved) — Mobile portrait dialog has no alternative
  flow** — `pkg/engine/character_creation_mobile.go:30-31` now returns a typed
  `ErrPortraitDialogUnsupported` sentinel (improvement over the plain `fmt.Errorf` in the
  prior audit) and the UI hides the Browse button on mobile builds. However, no alternative
  path (preset gallery or native image picker) is wired; players cannot set any custom portrait
  on iOS/Android.
  **Blocked goal**: README mobile-support coverage of character-creation parity.
  **Remediation**: Either implement `pkg/mobile.OpenImagePicker` with `//go:build android`
  calling `Intent.ACTION_PICK` and `//go:build ios` calling `UIImagePickerController`, or add
  a procedural preset-portrait gallery to the character-creation UI that renders without a
  file dialog.
  **Effort**: medium

### LOW

- [ ] **G14 (carry-forward, sub-items 3–5 still open) — Concurrency safety gaps** —
  From the prior `GAPS.md` G14 group (legacy "Gap 1"–"Gap 6"), the following remain:

  - [ ] **G14-3 (Prior Gap 3)** — `pkg/network/server.go:213-248`
    `TCPServer.Start()` uses `defer s.clientsMu.Unlock()` at entry and then manually
    calls `s.clientsMu.Unlock()` + `s.clientsMu.Lock()` in the middle to avoid holding
    the lock while starting goroutines. The mixed defer/manual pattern is fragile; a
    future `return` path added before the manual unlock will silently double-unlock.
    **Remediation**: Restructure to use a temporary closure or explicitly document
    the unlock sequence with a comment warning contributors not to add early returns.

  - [ ] **G14-4 (Prior Gap 4)** — Non-deferred mutex patterns remain in:
    `pkg/procgen/audit_registry.go:22,24,29,32` (explicit `Lock` + `Unlock` with no
    defer); `pkg/procgen/terrain/async_loader.go:54,57,67,70,81,84,142,144`
    (same pattern in a goroutine callback — panic in the loop body would leave the
    mutex locked).
    **Remediation**: Convert to `defer mu.Unlock()` at the top of each critical
    section, or at minimum add `recover` guards in the goroutine callbacks.

  - [ ] **G14-5 (Prior Gap 5)** — `cmd/client/init_monitoring.go:52-63`
    `startLegacyMetricsMonitor` spawns a goroutine with `for range ticker.C`
    and no `select` on a cancellation channel or context. The goroutine therefore
    runs for the process lifetime; there is no shutdown hook.
    **Remediation**: Accept a `context.Context` parameter; add a `select` case
    `case <-ctx.Done(): return` so the ticker goroutine shuts down cleanly on
    game exit or signal handler.

---

## Resolved Findings (closed in rev 2)

The following findings from the prior `AUDIT.md` (2026-04-24) are **✅ closed** as of
the rev-2 audit:

| ID | Title | Evidence |
|----|-------|----------|
| G1 | OpenXR controller/headset adapters return zero values | `pkg/engine/vr_openxr_adapters.go` now 615 lines of cgo OpenXR implementation (`//go:build vr && !js`); 0 TODO markers; `xrCreateInstance/GetSystem/CreateSession/LocateViews/SyncActions/ApplyHaptic` all implemented |
| G2 | 11 engine systems never registered | All 11 registered in `pkg/engine/system_init.go:1864,1876,2077-2114`; seasonal-event cluster seeded with `SeasonalEventComponent` at line 2093 |
| G3 | Mod browser unreachable | `ModBrowserSystem` registered in `system_init.go:2107`; callbacks wired in `cmd/client/init_versions.go:653-744`; partial — see new G16 for remaining repository gap |
| G4 | Seasonal-event subsystem no spawner | `SeasonalEventComponent` seeded on world entity at `system_init.go:2093`; all 4 event systems registered |
| G5 | Modding system wired server-only | `cmd/client/init_versions.go:657-700` calls `world.SetModRules(modding.NewProviderAdapter(sys.modManager))` for single-player/host-and-play |
| G6 | `vr_webxr_adapters.go` documented but missing | `pkg/engine/vr_webxr_adapters.go` exists (451 lines, `//go:build js`); full WebXR implementation via `syscall/js` |
| G7 | Client has no observability endpoint | `cmd/client/init_monitoring.go:157` + `cmd/client/util.go:70-72` implement opt-in `--enable-metrics` flag with `initObservabilityExporter` |
| G8 | Dead `Server` type in `pkg/hostplay` | `pkg/hostplay/host_and_play.go` removed; package now contains only `server_manager.go`, `input_handler.go`, `state_broadcaster.go`, `doc.go` |
| G9 | `EnableShadows` deprecation not enforced | `pkg/rendering/lighting/system.go:44-51` emits `logrus.Warn` on deprecated combination |
| G10 | `ExtendedAchievementSystem` shadows wired system | Decision documented at `system_init.go:2109-2114`: both registered; G10 comment states they complement each other (kills/quests/crafting vs expression/social) |
| G11 | Menu "Exit Game" returns error | `pkg/engine/menu_system.go:158-162` exposes `SetExitCallback`; `cmd/client/handlers.go:3190` injects `ebiten.Termination` |
| G13 | `pkg/companion` namespace undocumented | `pkg/companion/doc.go` added with namespace map |
| G14-1 | `FederatedMarket.Stop` lacks `sync.Once` | `pkg/network/federation/market.go:24,117` — `stopOnce sync.Once` field added |
| G14-2 | `FederatedMarket.Start` lacks `sync.Once` | `pkg/network/federation/market.go:23,98` — `startOnce sync.Once` field added |
| G14-6 | `ProjectileNetworkSync.CleanupTask` stop channel undocumented | `pkg/network/projectile_sync.go:441-463` now has doc comment |

---

## False Positives Considered and Rejected

| Candidate | Reason Rejected |
|-----------|-----------------|
| `pkg/engine/character_creation_mobile.go` described as "stub" | Intentional — ships `ErrPortraitDialogUnsupported` typed error under `//go:build js \|\| android \|\| ios`; desktop counterpart provides real zenity dialog. Residual UX gap tracked as G12. |
| `pkg/engine/guild_component.go` is an empty/no-logic struct | Verified intentional — pure-data ECS component; logic lives in `pkg/network/federation/guild/manager.go` and `pkg/engine/guild_*_system.go`. |
| `pkg/engine/behavior_tree_entity.go` is a stub | Verified intentional — single helper `entityFromContext` used by every BehaviorNode.Tick via `aitypes.EntityContext`. |
| `InMemoryModRepository` used as mod browser backing store | The code comment explicitly labels it "for testing", but it is currently used in production as a placeholder. Tracked as G16, not dismissed. |
| `startLegacyMetricsMonitor` goroutine | Real concurrency gap (no cancellation path). Retained as G14-5. |
| `pkg/engine/vr_openxr_adapters.go` — previously had 14 TODOs | All 14 replaced by cgo implementation in rev 2. G1 closed. |
| `pkg/engine/HotReloadSystem` — only in demo | Demo is `examples/`, not production code. Tracked as G15 (HIGH). |
| `pkg/hostplay` `*Server` type | Removed in rev 2. G8 closed. |

---

## Cross-Package Wiring Verification (rev 2)

| Chain | Result |
|-------|--------|
| Audio: `pkg/engine/audio_manager.go` → `pkg/audio/Manager` → synthesis | ✅ Wired (`cmd/client/handlers.go:1080-1130`) |
| Voice: `pkg/audio/voice.go` → `pkg/network.TCPVoiceTransport` → network | ✅ Wired (`cmd/client/handlers.go:773,1189`) |
| VR OpenXR: `vr_openxr_adapters.go` → `pkg/vr/Detector` → game loop | ✅ Wired under `-tags vr`; graceful stub fallback without tag |
| VR WebXR: `vr_webxr_adapters.go` → adapter factory → game loop | ✅ Wired under `//go:build js` |
| Modding: `pkg/modding/loader.go` → `world.SetModRules` | ✅ Both client (`init_versions.go:700`) and server (`main.go:120`) |
| Mod browser: `ModBrowserSystem` → `ModRepository` → ECS | ⚠️ Registered + callbacks wired; backed by `InMemoryModRepository` (G16) |
| Hot reload: `HotReloadSystem` → `FileWatcher` → ECS | ❌ System defined but never registered (G15) |
| Observability: `pkg/observability` → `pkg/engine`/`pkg/network` | ✅ Server always; client opt-in via `--enable-metrics` |
| Host-and-play: `pkg/hostplay/server_manager.go` → `cmd/client/util.go` | ✅ Wired (`util.go:214`) |
| Mobile input: `pkg/mobile/*` → `cmd/mobile/` | ✅ Wired |
| Procgen: `pkg/procgen/*` → `cmd/server/entity_spawning.go` + `cmd/client` | ✅ Wired |
| Companion learning: `pkg/companion/learning` → `pkg/engine/companion_learning_system.go` | ✅ Wired |
| Seasonal events: event_*_system.go → `world.AddSystem` + `SeasonalEventComponent` | ✅ Wired (`system_init.go:2075-2093`) |
| Commerce: `CommerceSystem` → `world.AddSystem` | ✅ Wired (`system_init.go:1094`) |
| Extended achievements: `ExtendedAchievementSystem` → `world.AddSystem` | ✅ Wired (`system_init.go:2112-2114`) |

---

## Methodology

1. **Phase 0** — Read `README.md`, `ROADMAP.md`, `REFACTORING_SUMMARY.md`, `go.mod`, prior
   root `AUDIT.md` + `GAPS.md` (IDs G1–G14, Gaps 1–6).
2. **Phase 1** — Direct read of all previously-flagged files; re-verified each prior finding.
3. **Phase 2** — Codebase-wide scans:
   - `grep -rn 'TODO\|FIXME\|HACK\|XXX' --include='*.go' --exclude='*_test.go' .`
     → **0 hits** in non-test code (all prior TODO markers replaced).
   - `New<Symbol>` constructor cross-reference across 358 system constructors in
     `pkg/engine/` — isolated `HotReloadSystem` as the sole unregistered system.
   - `AddSystem` call count in `system_init.go` → 183 calls (up from prior audit).
4. **Phase 3** — Gap status re-verification: each G1–G14 item re-checked against current
   source, 13 of 14 closed (G14 sub-items 3, 4, 5 still open).
5. **Phase 4** — New gap discovery: G15 (HotReloadSystem), G16 (FileSystemModRepository).
6. **Phase 5** — False-positive scrub.

No source files were modified.
