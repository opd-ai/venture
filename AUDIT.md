# IMPLEMENTATION GAP AUDIT — 2026-04-20

## Project Architecture Overview

**Venture** is a fully procedural multiplayer action-RPG built with Go 1.24.5 and Ebiten v2.9.3. The stated philosophy is "zero external assets" — every graphic, sound, and gameplay element is generated at runtime from a single distributable binary. Stated goals from `README.md` and `ROADMAP.md`:

- 100% procedural content (terrain, items, NPCs, quests, dialog, music, SFX, sprites)
- ECS architecture with components-as-data and 100+ systems
- Genre-based theming (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)
- High-latency multiplayer (200–5000ms) with **client-side prediction**, lag compensation, snapshot synchronization, suitable for Tor/onion services
- Cross-server federation (discovery, auth, WebRTC, portals, marketplaces, multi-server guilds)
- Voice chat with party/guild/proximity/private channels using built-in ADPCM codec and spatial audio
- Player housing, guildhalls, territory control, raids, marketplaces, persistent world state
- Sandboxed JSON-based modding system with scriptable entities
- Mobile (iOS/Android) and WebAssembly targets in addition to Linux/macOS/Windows
- 60 FPS / <500MB client memory targets, ≥40% per-package test coverage

**Package responsibilities** (123 packages, 754 non-test `.go` files, ~399k LOC):

| Package | Responsibility |
|---|---|
| `cmd/client`, `cmd/server`, `cmd/mobile`, `cmd/balance-validator` | Application entry points |
| `pkg/engine` | ECS core, entity/world/component management, 100+ gameplay systems |
| `pkg/procgen/*` (25+ subdirs) | Deterministic, seed-based content generators |
| `pkg/rendering/*` | Runtime sprite/tile/lighting/particle/UI generation |
| `pkg/audio/*` | Procedural music, SFX, synthesis, voice codec |
| `pkg/network/*` | Multiplayer client/server, federation, prediction, lag compensation, voice transport |
| `pkg/world/*` | Persistent world state, housing, economy, territory, raids, chunked streaming |
| `pkg/integration/*` | Cross-system feature integrations |
| `pkg/saveload`, `pkg/modding`, `pkg/security`, `pkg/observability`, `pkg/validation` | Cross-cutting infrastructure |

**Baseline quality state** (verified 2026-04-20):

- `go build ./...` — **clean** (after installing X11/GL dev headers).
- `go vet ./...` — **clean** (no warnings).
- `go-stats-generator` invocation timed out / did not complete on this large codebase within the available budget; substituted manual `grep`/`go list` analysis with explicit verification of every claim.
- An existing `GAPS.md` and `ROADMAP.md` track many resolved items; multiple entries in the prior `GAPS.md` are now stale and were re-verified for this audit (see "False Positives Considered and Rejected").

---

## Gap Summary

| Category | Count | Critical | High | Medium | Low |
|----------|-------|----------|------|--------|-----|
| Stubs/TODOs | 5 | 0 | 1 | 0 | 4 |
| Dead Code | 11 | 0 | 1 | 8 | 2 |
| Partially Wired | 4 | 1 | 2 | 1 | 0 |
| Interface Gaps | 0 | 0 | 0 | 0 | 0 |
| Dependency Gaps | 0 | 0 | 0 | 0 | 0 |
| **Total** | **20** | **1** | **4** | **9** | **6** |

---

## Implementation Completeness by Package

Based on grep-verified instantiation analysis of the highest-signal packages. "Coverage" reflects the prior project assessment in `ROADMAP.md` (not re-measured for this audit; stats generator did not complete).

| Package | Notable Exported Constructors | Wired in Runtime | Dead/Unwired | Notes |
|---|---|---|---|---|
| `pkg/engine` | ~140 `NewXxxSystem` | ~135 | 4 (`NewAchievementNotificationSystem`, `NewAvailabilitySystem`, `NewCarrySystem`, `NewCommerceSystem`) | Bulk of unwired systems are scaffolded but never registered. |
| `pkg/world` | 12 chunk/world systems | 9 | 2 (`NewChunkCompressionSystem`, `NewChunkModificationSystem`) | Chunk loader is wired in `cmd/server/main.go:507`; compression/modification companions are not. |
| `pkg/network` | 18 (predictor, sync, bandwidth, federation managers) | ~14 | 4 (`NewClientPredictor`, `NewHighLatencyClientPredictor`, `NewAnimationSyncManager`, `NewBandwidthMonitor`) | Gap impacts a stated core multiplayer feature. |
| `pkg/network/federation` | 30+ | ~29 | 1 (`NewAuthManager`) | One auth manager type defined but unused; federation auth flow uses `auth.Authenticator` interface in handshake. |
| `pkg/world/housing` | ~20 | ~19 | 1 (`NewBlueprint`) | `NewBlueprintWithTime` is the only used constructor. |
| `pkg/mobile` | platform shims | all wired | 0 | `ShowKeyboard()` non-WASM is intentional no-op (documented). |
| `pkg/audio` | core synth+codec | all wired | 0 | `InitializeVoice` carries an integration TODO for server routing. |
| `pkg/procgen/quest` | 12 template constructors | all called | 0 | TODO is a refactor suggestion, not a missing template. |
| `pkg/procgen/story` | fragment generator | all called | 0 | TODO is a refactor suggestion (data tables vs. switches). |

---

## Findings

### CRITICAL

- [x] **Voice chat server-side packet routing missing** — `pkg/network/server.go` (no voice handler present); receive/decode path lives in `pkg/network/voice_transport.go:279` (`HandleReceivedPacket`) — README claims "Voice chat is integrated with party, guild, proximity, and private channels"; clients can encode and send voice packets but the server never broadcasts them to other channel members, so multiplayer voice never reaches any other player. Blocks the README's voice-chat goal end-to-end. — **Remediation:** In `pkg/network/server.go`, add a branch in the input dispatcher matching `InputType=="voice"` (or a new packet type) that looks up the sender's active voice channel via `engine.VoiceChannelSystem`, iterates channel members, and enqueues the payload onto each member's outbound queue using the existing client send path. Add an integration test `TestVoiceEndToEnd` in `pkg/network/voice_transport_test.go` that spins up two TCP test servers and asserts a packet sent by client A is delivered to client B. Validate with `xvfb-run -a go test ./pkg/network/...`.

### HIGH

- [x] **Client-side prediction never instantiated** — `pkg/network/prediction.go:55` (`NewClientPredictor`) and `:68` (`NewHighLatencyClientPredictor`) — README and `pkg/network/doc.go` advertise client-side prediction as a core mechanism for the 200–5000 ms latency target; the predictor type is fully implemented (state history, reconciliation, replay) but is never constructed in `cmd/client/`, `pkg/network/client.go`, or any system. Without a live predictor, the client cannot predict-then-reconcile any input under network latency, defeating one of the project's headline multiplayer goals. — **Remediation:** In `pkg/network/client.go` (`TCPClient`), add a `predictor *ClientPredictor` field and instantiate it in `NewTCPClient`. Choose `NewHighLatencyClientPredictor()` when the existing `--high-latency` flag is set, else `NewClientPredictor()`. Call `PredictInput` from the input send path and `ReconcileServerState` when the server snapshot arrives. Add a unit test `TestClientPredictorReconciliation` confirming the client diverges from the server then converges within `errorThreshold`. Validate with `go build ./... && go test ./pkg/network/...`.

- [x] **Four `pkg/engine` systems defined but never registered** — `pkg/engine/achievement_notification_system.go:55`, `pkg/engine/availability_system.go:20`, `pkg/engine/carry_system.go:32`, `pkg/engine/commerce_system.go:60` — Each `NewXxxSystem` constructor has zero non-test, non-definition references (verified by grep across `cmd/`, `pkg/engine/system_init.go`, `cmd/client/init_versions.go`, `cmd/server/v4_systems.go`). Each implements the `System` interface, so `Update()` is never called. README/architecture lists 100+ active systems; whichever player-visible features these encode (achievement toast UI, NPC time-of-day availability, carry/grapple, NPC commerce/shop initiation) are silently inert. — **Remediation:** For each system, audit its `Update()` and dependencies; either (a) register it in `pkg/engine/system_init.go` adjacent to the related domain (`AchievementSystem`, `InventorySystem`, etc.) and surface it through `SystemInitResult`, or (b) delete the file and its test if the feature is genuinely unwanted. Validate by adding a one-line `world.AddSystem(...)` and asserting via `make feature-audit` (`examples/featureaudit`) that the system count increases.

- [x] **`AnimationSyncManager` never instantiated** — `pkg/network/animation_sync.go:206` — Implements throttled animation-state synchronization for multiplayer, including jitter buffering (`BufferState`, `GetNextState`). Zero non-test references. Without it, the project cannot deliver smooth remote-player animations under high latency, a stated multiplayer feature. — **Remediation:** In `pkg/engine/system_init.go` (network/multiplayer section) instantiate `mgr := network.NewAnimationSyncManager()` and inject it into `AnimationSystem` (add a setter `SetSyncManager`). On the client, call `mgr.ShouldSync` before sending an animation-state packet and `mgr.BufferState` when one arrives. Add an integration test in `pkg/network/animation_sync_test.go` covering the round-trip. Validate with `go build ./... && go test ./pkg/network/...`.

### MEDIUM

- [x] **`ChunkCompressionSystem` and `ChunkModificationSystem` never instantiated** — `pkg/world/chunk_compression.go:14`, `pkg/world/chunk_modification.go:15` — Sibling chunk system `ChunkLoaderSystem` is wired in `cmd/server/main.go:507`, but the compression and modification-tracking companions are referenced only in `pkg/world/doc.go` example comments. Memory budget and persistent-world goals (READMEs "memory-efficient large worlds") rely on chunk eviction compression and modification dirty-tracking. — **Remediation:** In `cmd/server/main.go` immediately after `chunkLoader := worldpkg.NewChunkLoaderSystem(...)` (line 507), construct `compressor := worldpkg.NewChunkCompressionSystem()` and `mods := worldpkg.NewChunkModificationSystem(persistentState)`. Wire `compressor` into the loader's eviction callback and have `mods` register on chunk write. Add tests `TestChunkEvictionCompresses` and `TestChunkModificationTracked`. Validate with `go test ./pkg/world/...`.

- [x] **`AuthManager` defined but never used** — `pkg/network/federation/auth.go:30` — Federation README/doc references authentication as a federation primitive. Zero non-test references; the federation handshake currently uses other types (`auth.Authenticator` interface). Either premature abstraction or the planned upgrade was abandoned. — **Remediation:** Decide between (a) deleting `AuthManager` and its tests (lowest risk; removes ~few hundred LOC of dead surface) or (b) wiring it into `pkg/network/federation/handshake.go` as the canonical token store. If retained, add a `TestFederationHandshakeUsesAuthManager` integration test. Validate with `go vet ./pkg/network/federation/... && go test ./pkg/network/federation/...`.

- [x] **`BandwidthMonitor` defined but never used** — `pkg/network/bandwidth.go:38` — Provides sliding-window bandwidth measurement; `pkg/observability` exposes Prometheus metrics. Zero non-test references; bandwidth is therefore never observable, violating the spirit of the observability goal for multiplayer health. — **Remediation:** Instantiate `bw := network.NewBandwidthMonitor(time.Second)` in `pkg/network/server.go` (`TCPServer`) and `pkg/network/client.go` (`TCPClient`). Call `bw.RecordSent`/`RecordReceived` from the existing send/receive paths and expose its rolling rate via a new `pkg/observability` gauge. Add a unit test `TestBandwidthMonitorIntegration`. Validate with `go test ./pkg/network/...`.

- [x] **`NewBlueprint` constructor unreachable** — `pkg/world/housing/types.go:391` — `NewBlueprintWithTime` is the only used constructor. `NewBlueprint` (zero non-test references) calls `NewBlueprintWithTime(time.Now())` per its body but is dead. Causes confusion for contributors choosing a constructor. — **Remediation:** Delete `NewBlueprint` (and any test that exercises it that does not also cover the time-injecting variant) so that callers use the deterministic, testable `NewBlueprintWithTime`. Validate with `go build ./pkg/world/housing/... && go test ./pkg/world/housing/...`.

- [x] **XP never awarded on enemy death** — `cmd/client/util.go:1430` (`createDeathCallback` signature) — `progressionSystem` is not in the parameter list, and `pkg/engine/progression_system.go:97` (`AwardXP`) is fully implemented yet never called from the death callback. README ties character progression to combat; without XP-on-kill the player cannot level. — **Remediation:** Add `progressionSystem *engine.ProgressionSystem` to the `createDeathCallback` signature; after `enemy.AddComponent(deadComp)` (line ~1460), call `progressionSystem.AwardXP(*playerEntity, calculateXP(enemy))` (`calculateXP` reads enemy level from `HealthComponent.MaxHP` or a new `XPRewardComponent`). Thread `sys.progressionSystem` from `cmd/client/handlers.go:3561`. Add table-driven test `TestDeathCallbackAwardsXP`. Validate with `xvfb-run -a go test ./cmd/client/... && go vet ./...`.

- [x] **Refactoring TODO: quest template tables** — `pkg/procgen/quest/types.go:286` (`GetFantasyKillTemplates`) — `TODO(REM-144)` proposes converting genre→template switch cases into package-level data tables to avoid editing function bodies for new templates. Functional today, but accretive complexity. — **Remediation:** Introduce `var killTemplates = map[string][]QuestTemplate{...}` populated at package init for all five genres, and replace the per-genre `GetXxxKillTemplates` functions with a single lookup. Keep public-API surface stable by re-exporting wrappers. Validate with `go test ./pkg/procgen/quest/...` (table-driven tests already exist and should keep passing).

- [ ] **Refactoring TODO: story theme tables** — `pkg/procgen/story/generator.go:303` (`selectTheme`) — `TODO(REM-096)` proposes lifting the duplicated genre→theme switch out of `selectTheme` and `getThemesForGenre` into a single `var genreThemes = map[string][]string`. Functional today; reduces drift risk. — **Remediation:** Define `var genreThemes = map[string][]string{...}` in `pkg/procgen/story/generator.go` and rewrite both helpers to read from it. Add `TestSelectTheme_AllGenresCovered` to lock the contract. Validate with `go test ./pkg/procgen/story/...`.

- [ ] **`InitializeVoice` integration TODO** — `pkg/audio/manager.go:310` — `TODO(integration): Wire VoiceTransport to network/chat subsystem`. The function itself is implemented and is called by `cmd/client/handlers.go:1133`; the TODO restates the same gap as the CRITICAL voice-routing finding (server-side broadcast). Tracked here for completeness so its resolution is paired with the CRITICAL fix. — **Remediation:** Resolve concurrently with the CRITICAL voice-routing finding; remove the TODO line and add a comment pointing at the server's voice broadcast handler. Validate with `go vet ./pkg/audio/... && go test ./pkg/audio/...`.

- [ ] **Native mobile keyboard integration TODO** — `pkg/mobile/keyboard_default.go:19` (`ShowKeyboard`) — TODO states UIKeyboard (iOS) / InputMethodManager (Android) not yet integrated. `IsKeyboardSupported()` correctly returns `false` and the WASM bridge handles browser keyboards; the function fulfils its documented "no-op on non-WASM platforms" contract. Mobile support is listed but stated to use OS keyboards. Low severity because text-entry on mobile is currently delegated to OS controls; only programmatic show/hide is missing. — **Remediation:** When mobile text entry is prioritized: add `pkg/mobile/keyboard_android.go` and `pkg/mobile/keyboard_ios.go` using `gomobile`/`ebitenmobile` JNI bridges to call `InputMethodManager.showSoftInput` and `UIResponder.becomeFirstResponder`. Update `IsKeyboardSupported()` to return true on those builds. Validate with `make android-apk` / `make ios-simulator` and a manual focus-text-field test.

### LOW

- [ ] **`HighLatencyClientPredictor` never instantiated** — `pkg/network/prediction.go:68` — Variant of the dead `ClientPredictor` (covered above). Listed separately since it carries the high-latency tuning that is specifically called out in the README's 200–5000 ms goal. Same remediation: select this variant when `--high-latency` is set, in the same change that wires the base predictor.

- [ ] **`pkg/world` chunk doc-only references** — `pkg/world/doc.go:53,63,68` — Doc comments demonstrate `NewChunkLoaderSystem`, `NewChunkCompressionSystem`, `NewChunkModificationSystem` usage but are the only non-`func` references for the latter two. Once the MEDIUM finding above is closed, these examples will match runtime reality. — **Remediation:** Update `pkg/world/doc.go` after wiring chunk compression/modification to point at the actual `cmd/server/main.go` integration site. Validate with `go test ./pkg/world/...`.

- [ ] **Disabled test file present** — `pkg/engine/high_throughput_integration_test_disabled.go` — Sub-200-byte file (heuristic empty/scaffold) ending in `_disabled.go` (no build tag, just naming convention so it is not compiled). Indicates a deliberately disabled integration test, not a stub. — **Remediation:** Decide whether to (a) re-enable with a documented `//go:build integration` tag plus a Makefile target, or (b) delete the file. Validate with `grep -r '_disabled.go' pkg/ cmd/` returning no other matches.

- [ ] **Quest TODO is a refactor, not a missing feature** — `pkg/procgen/quest/types.go:286` — Already covered as MEDIUM-Low refactor; flagged here as the canonical "tracked TODO" for completeness so it appears in low-priority sweep lists.

- [ ] **Story TODO is a refactor, not a missing feature** — `pkg/procgen/story/generator.go:303` — Same as above.

- [ ] **Audio integration TODO is a forward-reference** — `pkg/audio/manager.go:310` — Same as the CRITICAL/MEDIUM voice findings; listed here so it can be closed in a single PR.

---

## False Positives Considered and Rejected

| Candidate Finding | Reason Rejected |
|---|---|
| `WorldPersistenceSystem` not instantiated (per pre-existing `GAPS.md` Gap 18) | **Resolved.** Verified at `pkg/engine/system_init.go:2179` `worldPersistenceSystem := NewWorldPersistenceSystem()`. The prior `GAPS.md` entry is stale. |
| `ScriptingSystem` not instantiated (per pre-existing `GAPS.md` Gap 17) | **Resolved.** Verified at `pkg/engine/system_init.go:1190` `scriptingSystem := NewScriptingSystem(game.World)`. Prior `GAPS.md` entry is stale. |
| `HousingUI` missing from `shouldUpdateWorld`/`anyUIOpen` (per Gap 19) | **Resolved.** Verified at `pkg/engine/game.go:1374` (`shouldUpdateWorld` clause) and `:1337` (`anyUIOpen` clause) — both include `HousingUI`. |
| `LoadingUI` GPU image leak (`drawRect`, per Gap 15) | **Resolved.** Verified `pkg/engine/loading_ui.go:88` already has `defer img.Dispose()`. |
| `ChunkLoaderSystem` not instantiated (per Gap 20) | **Resolved.** Verified at `cmd/server/main.go:507` `chunkLoader := worldpkg.NewChunkLoaderSystem(...)`. The compression and modification companions, however, remain unwired (recorded as MEDIUM). |
| Mobile `CraftingSystem` initialized with nil `itemGenerator` (per Gap 16) | **Resolved.** `pkg/engine/crafting_system.go:84` now exposes `SetItemGenerator`; mobile path can call it. |
| `enableMetrics` flag parsed but never used (per "wiring" subagent) | **False positive.** Flag is read at `cmd/server/main.go:138` (`if *enableMetrics { ... }`). |
| 324 systems "unregistered" (per "wiring" subagent's broad grep) | **False positive (over-counting).** That count assumed registration only happens in `system_init.go`. Verified that `cmd/client/init_versions.go`, `cmd/client/handlers.go` (266 `AddSystem` calls), `cmd/server/v4_systems.go`, and `cmd/server/main.go` (12) collectively register the vast majority. Only the four `pkg/engine` systems listed under HIGH (`AchievementNotificationSystem`, `AvailabilitySystem`, `CarrySystem`, `CommerceSystem`) survive grep verification across all four sites. |
| `pkg/engine/chat_system.go:23` empty `Update()` | **Not a stub.** Method body is documentation-explained as event-driven (`SendMessage`-triggered); design decision, not unimplemented work. The system is registered and used. |
| `pkg/engine/alignment_system.go:32` empty `Update()` | **Not a stub.** Same pattern: state is mutated by `RecordDeed` calls, not by per-frame ticks. System is wired (`cmd/client/handlers.go:2066`, `cmd/server/v4_systems.go:132`). |
| `pkg/engine/achievement.go:154` empty `Update()` (`AchievementSystem`) | **Not a stub.** Achievement evaluation is event-driven via `CheckOnEvent`; system is wired (`cmd/server/v4_systems.go:158`, `cmd/client/init_versions.go:139`). |
| `pkg/mobile/keyboard_default.go` no-op `ShowKeyboard`/`HideKeyboard` | **Intentional minimalism.** Documented as no-op on non-WASM; OS provides keyboard. Flagged the TODO separately at LOW. |
| VR system using `StubHeadsetAdapter` / `StubControllerAdapter` | **Documented as experimental** in README; the project explicitly states VR is mock-only pending SDK integration (ROADMAP Priority 4). Not an undisclosed gap. |
| `NewXxxWithLogger` variants unused while `NewXxx` is used (e.g., `NewAlignmentSystemWithLogger`) | **Not a critical gap.** A "with logger" variant is a public-API affordance for callers wanting custom logging; absence of internal callers is by design (the basic `NewXxx` covers the runtime path). Refactoring to a single constructor with options is a code-quality improvement, not an implementation gap. Not enumerated. |
| Missing `panic("not implemented")` patterns | **None found** in `pkg/`/`cmd/` non-test code. Confirms no naked stubs exist. |
| `go.mod` unused dependencies | **None found.** All direct dependencies (`ebiten`, `logrus`, `uuid`, `golang.org/x/image`, `golang.org/x/text`, `zenity`) are imported by production code. Indirect deps are transitive only. |

---

## Methodology Notes

- `go-stats-generator analyze . --skip-tests` did not complete within the available wall-clock budget on this codebase. Substituted manual evidence collection: `grep -rn`, `go list ./...`, `go build ./...`, `go vet ./...`, plus targeted file reads.
- Every finding above was verified by an explicit `grep` for the constructor or symbol across `cmd/` and `pkg/` excluding `_test.go` and the symbol's own definition file before being recorded.
- All severity ratings follow the project's stated goals (README + ROADMAP). A symbol that "should exist" hypothetically but is not promised by docs is **not** flagged.
- Outputs are limited to this `AUDIT.md` and `GAPS.md`. No source code changes were made.
