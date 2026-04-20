# Implementation Gaps — 2026-04-20

This document catalogs verified implementation gaps between Venture's stated goals and its current runtime behavior. Each gap was confirmed by direct file inspection and `grep` verification across `cmd/` and `pkg/` (excluding `_test.go` files) on 2026-04-20.

For the executive summary and false-positive log, see `GAPS.md` (this file). The root `AUDIT.md` has been retired after all findings were resolved.

---

## Gap A: Voice Chat Server-Side Packet Routing Missing

- **Intended Behavior**: README — "Voice chat is integrated with party, guild, proximity, and private channels using a built-in codec with spatial audio support." Clients should be able to speak in a channel and have all other channel members hear them.
- **Current State**: Client-side encode/send and decode/playback are implemented (`pkg/audio/voice.go`, `pkg/network/voice_transport.go`). `TCPVoiceTransport.HandleReceivedPacket` (`pkg/network/voice_transport.go:279`) processes incoming packets locally. `pkg/network/server.go` contains **no** voice-packet handling — the server never broadcasts received voice payloads to other channel members. `pkg/audio/manager.go:310` carries an explicit `TODO(integration): Wire VoiceTransport to network/chat subsystem`.
- **Blocked Goal**: Multiplayer voice chat (README "Voice chat" feature). Two clients can each talk to themselves but never to each other.
- **Implementation Path**:
  1. In `pkg/network/server.go`, add a packet-type branch (or `InputType=="voice"` arm in the existing input dispatcher) that reads the voice payload and the sender's claimed channel ID.
  2. Look up channel members via `engine.VoiceChannelSystem` (already instantiated and has `GetChannelMembers`-style API).
  3. For each member ≠ sender, enqueue the payload onto that client's outbound queue using the existing send path.
  4. For proximity channels, intersect channel members with `engine.SpatialVoiceSystem` distance results before forwarding.
  5. Add `TestVoiceEndToEnd` and `TestSpatialVoiceRouting` integration tests under `pkg/network/`.
- **Dependencies**: None (`VoiceChannelSystem` and `SpatialVoiceSystem` are already wired per pre-existing GAPS.md Gap 1 status).
- **Effort**: small (1–2 days)

---

## Gap B: Client-Side Prediction Never Instantiated

- **Intended Behavior**: README — "high-latency multiplayer (200–5000ms) … with client-side prediction, lag compensation, and snapshot synchronization". `pkg/network/doc.go` advertises "client-side prediction, server reconciliation, entity interpolation".
- **Current State**: `ClientPredictor` (`pkg/network/prediction.go:30`) is fully implemented with state history, sequence numbers, `PredictInput`, `ReconcileServerState`, `correctAndReplay`, and a high-latency-tuned variant `NewHighLatencyClientPredictor` (`:68`). Both constructors have **zero non-test references** in `cmd/` or `pkg/`. `pkg/network/client.go` (`TCPClient`) does not hold a predictor field. The header comment of `client.go` even claims "this file implements TCPClient which handles client-side networking, prediction, …" but no prediction calls exist.
- **Blocked Goal**: Responsive gameplay under the project's stated 200–5000 ms latency target. Players see only server-authoritative positions with full RTT delay; the project's headline multiplayer design promise is unfulfilled.
- **Implementation Path**:
  1. Add `predictor *ClientPredictor` field to `TCPClient` in `pkg/network/client.go`.
  2. In `NewTCPClient`, instantiate `network.NewClientPredictor()` by default; instantiate `network.NewHighLatencyClientPredictor()` if a new constructor parameter `highLatency bool` is true. Plumb this from the existing `--high-latency` flag in `cmd/client/main.go`.
  3. In the input-send path, call `predictor.PredictInput(dx, dy, dt)` and apply the returned `PredictedState` to the local player entity *before* the server confirms.
  4. In the snapshot-receive path, call `predictor.ReconcileServerState(seq, pos, vel)` and reconcile (correct + replay) the local player.
  5. Add `TestClientPredictorReconciliation` (forces a divergence, asserts convergence within `errorThreshold`).
- **Dependencies**: None.
- **Effort**: medium (3–5 days; requires careful integration with the existing entity update loop)

---

## Gap C: Four `pkg/engine` Systems Defined But Never Registered

- **Intended Behavior**: 100+ ECS systems registered with the world, each `Update()` invoked per frame.
- **Current State**: The following constructors have **zero** non-test, non-definition references across `cmd/` and `pkg/engine/system_init.go` and the per-version init files:
  - `NewAchievementNotificationSystem` — `pkg/engine/achievement_notification_system.go:55`
  - `NewAvailabilitySystem` — `pkg/engine/availability_system.go:20`
  - `NewCarrySystem` — `pkg/engine/carry_system.go:32`
  - `NewCommerceSystem` — `pkg/engine/commerce_system.go:60`

  Each implements `Update(entities, dt)` so the absence of registration means the corresponding gameplay never runs. Achievement notification UI never appears, NPC time-of-day availability never gates interactions, players cannot carry physical objects, and procedural NPC commerce/shop initiation never occurs.
- **Blocked Goal**: Procedural NPC behavior, achievement feedback, and physical-interaction gameplay claimed by the engine architecture.
- **Implementation Path** (per system):
  1. In `pkg/engine/system_init.go`, instantiate alongside the related domain (e.g., `availabilitySystem := NewAvailabilitySystem(world, clock)` in the world-time block; `commerceSystem := NewCommerceSystem(world, inventorySystem)` after `inventorySystem` is constructed).
  2. Call `world.AddSystem(...)` and add the typed pointer to `SystemInitResult`.
  3. For systems with UI surface (`AchievementNotificationSystem`), wire the renderer call from `pkg/engine/game.go`'s `Draw()` path.
  4. If a system is genuinely unwanted, delete the file and its tests; do **not** leave a constructor that has no caller.
- **Dependencies**: For `CommerceSystem`, ensure `InventorySystem` is constructed before it (already true in the standard init order).
- **Effort**: small per system (1 day each); medium aggregate (≈4 days) including UI/renderer wiring.

---

## Gap D: `AnimationSyncManager` Never Instantiated

- **Intended Behavior**: Multiplayer animation states sync between clients with throttling and jitter buffering, producing smooth remote-player animations under high latency.
- **Current State**: `AnimationSyncManager` (`pkg/network/animation_sync.go:191`) is fully implemented (throttle policy, `BufferState`, `GetNextState`, per-entity buffers). `NewAnimationSyncManager` (`:206`) has **zero non-test references**. `AnimationSystem` has no sync-manager field.
- **Blocked Goal**: Smooth multiplayer animation under high latency (subset of the high-latency multiplayer goal). Without sync, remote players exhibit either over-frequent updates (bandwidth) or stuttering (no jitter buffer).
- **Implementation Path**:
  1. Add a `syncManager *network.AnimationSyncManager` field and a `SetSyncManager` setter to `engine.AnimationSystem`.
  2. In `pkg/engine/system_init.go` (or `cmd/client/init_versions.go`), create the manager and inject: `mgr := network.NewAnimationSyncManager(); animationSystem.SetSyncManager(mgr)`.
  3. On the network sender, call `mgr.ShouldSync(entityID, state)` to decide whether to emit a packet, and `mgr.RecordSync` after sending.
  4. On receive, call `mgr.BufferState(packet)` and pull with `mgr.GetNextState(entityID)` from the playback frame.
  5. Add `TestAnimationSyncRoundTrip` covering buffering and ordered playback.
- **Dependencies**: None.
- **Effort**: small (1–2 days)

---

## Gap E: `ChunkCompressionSystem` and `ChunkModificationSystem` Never Instantiated

- **Intended Behavior**: Persistent, memory-efficient large worlds — chunks stream in/out, compressed on eviction, with player modifications dirty-tracked per chunk.
- **Current State**: `ChunkLoaderSystem` is wired (`cmd/server/main.go:507`). Its companions are not:
  - `NewChunkCompressionSystem` — `pkg/world/chunk_compression.go:14` — zero non-test references; only mentioned in `pkg/world/doc.go:68` example.
  - `NewChunkModificationSystem` — `pkg/world/chunk_modification.go:15` — zero non-test references; only in `pkg/world/doc.go:63` example.
- **Blocked Goal**: <500 MB client memory target on large worlds (no eviction compression → memory grows unboundedly with explored area). Persistent world (no modification tracking → world snapshot persists every chunk every save).
- **Implementation Path**:
  1. In `cmd/server/main.go` immediately after the existing `NewChunkLoaderSystem` line: `compressor := worldpkg.NewChunkCompressionSystem(); modSystem := worldpkg.NewChunkModificationSystem(persistentState)`.
  2. Add an eviction-callback hook to `ChunkLoaderSystem` (`SetOnEvict func(chunkID, *Chunk)`) and use `compressor.Compress` inside it before persistence.
  3. Replace direct chunk writes in chunk-modifying paths with `modSystem.MarkModified(chunkID)`; only persist chunks where `modSystem.IsDirty(chunkID)`.
  4. Add `TestChunkEvictionCompresses` and `TestChunkModificationDirtyFlag`.
- **Dependencies**: A `ChunkLoaderSystem` eviction-callback API may need to be added (small surface change).
- **Effort**: medium (3–5 days)

---

## Gap F: XP Never Awarded on Enemy Death

- **Intended Behavior**: Players gain XP from kills and level up; progression is integral to RPG gameplay.
- **Current State**: `engine.ProgressionSystem.AwardXP` (`pkg/engine/progression_system.go:97`) is fully implemented and is called by `FactionXPBonusSystem`. The death pipeline (`cmd/client/util.go:1430` `createDeathCallback`, called from `cmd/client/handlers.go:3561`) does **not** include `progressionSystem` in its parameter list and never calls `AwardXP`. Verified by reading the entire callback (lines 1430–1485).
- **Blocked Goal**: Character progression from combat (RPG core).
- **Implementation Path**:
  1. Add `progressionSystem *engine.ProgressionSystem` to the `createDeathCallback` parameter list.
  2. After `enemy.AddComponent(deadComp)` (line ~1460), compute XP via a helper `calculateXP(enemy)` (use enemy `MaxHP` × scaling, or read a new `XPRewardComponent` if introduced) and call `_ = progressionSystem.AwardXP(*playerEntity, xp)` when both `progressionSystem != nil` and `playerEntity != nil`.
  3. Thread `sys.progressionSystem` into the call site at `cmd/client/handlers.go:3561`.
  4. Add `TestDeathCallbackAwardsXP` table-driven test in `cmd/client/util_test.go`.
- **Dependencies**: None (`AwardXP` is already thread-safe).
- **Effort**: small (< 1 day)

---

## Gap G: `AuthManager` Defined But Never Used

- **Intended Behavior**: Federation authentication primitive available for cross-server credential storage.
- **Current State**: `pkg/network/federation/auth.go:30` `NewAuthManager` has zero non-test references. Federation handshake (see `pkg/network/federation/handshake.go`) currently uses the `auth.Authenticator` interface and other concrete types.
- **Blocked Goal**: None directly — the federation auth flow does function via different types — but the orphan abstraction adds maintenance burden and confuses contributors choosing where to extend auth.
- **Implementation Path**: Pick one:
  - **Delete**: Remove `pkg/network/federation/auth.go` and any tests that exclusively cover `AuthManager`. Update `pkg/network/federation/doc.go` to reflect the actual auth path.
  - **Adopt**: Refactor `handshake.go` to acquire and persist tokens through `AuthManager`, providing a single coherent credential store. Add `TestFederationHandshakeUsesAuthManager`.
- **Dependencies**: None.
- **Effort**: small (1 day either direction)

---

## Gap H: `BandwidthMonitor` Defined But Never Used

- **Intended Behavior**: Sliding-window bandwidth measurement available to observability for multiplayer health monitoring.
- **Current State**: `pkg/network/bandwidth.go:38` `NewBandwidthMonitor` has zero non-test references. `pkg/observability/metrics.go` exposes Prometheus metrics but does not include a bandwidth gauge.
- **Blocked Goal**: Observability completeness for multiplayer health (relates to ROADMAP performance/observability priorities). Operators cannot see bandwidth in `/metrics`.
- **Implementation Path**:
  1. In `pkg/network/server.go` (`TCPServer`) and `pkg/network/client.go` (`TCPClient`) add `bw *BandwidthMonitor` fields constructed with `NewBandwidthMonitor(time.Second)`.
  2. From the existing send/receive loops, call `bw.RecordSent(n)` / `bw.RecordReceived(n)`.
  3. In `pkg/observability/metrics.go`, register a new gauge `venture_network_bandwidth_bytes_per_sec{direction=...}` and update it on a tick from `bw.Rate()`.
  4. Add `TestBandwidthMonitorRollingWindow` and `TestObservabilityExposesBandwidth`.
- **Dependencies**: None.
- **Effort**: small (1–2 days)

---

## Gap I: `NewBlueprint` Constructor Unreachable

- **Intended Behavior**: One canonical constructor for housing blueprints.
- **Current State**: `pkg/world/housing/types.go:391` `NewBlueprint()` calls `NewBlueprintWithTime(time.Now())`. `NewBlueprintWithTime` is the only constructor used in production. `NewBlueprint` has zero non-test references and creates a determinism hazard for callers who don't realize the alternative exists.
- **Blocked Goal**: Determinism of housing tests/seeds (low risk because no caller uses the non-deterministic variant today, but the surface invites future regressions).
- **Implementation Path**: Delete `NewBlueprint` and any tests that solely exercise it (keep tests that pass `time.Now()` explicitly to `NewBlueprintWithTime`). Update `pkg/world/housing/doc.go` if it references the deleted constructor.
- **Dependencies**: None.
- **Effort**: trivial (<30 minutes)

---

## Gap J: Quest Template Refactor (Tracked TODO)

- **Intended Behavior**: New genre quest templates can be added by editing data tables, not function bodies.
- **Current State**: `pkg/procgen/quest/types.go:286` carries `TODO(REM-144)` proposing the refactor. Functions (`GetFantasyKillTemplates`, `GetSciFiKillTemplates`, etc.) work today but each is 200+ lines of literals.
- **Blocked Goal**: Modding/extensibility (low impact today; mods currently target other extension points).
- **Implementation Path**:
  1. Define `var killTemplates = map[string][]QuestTemplate{ "fantasy": {...}, "scifi": {...}, "horror": {...}, "cyberpunk": {...}, "postapoc": {...} }` at package scope.
  2. Replace each `GetXxxKillTemplates() []QuestTemplate` with `return killTemplates["xxx"]` (or remove and switch callers to a single `GetKillTemplates(genre string)`).
  3. Repeat for fetch/escort/etc. template families.
  4. Run existing `pkg/procgen/quest/quest_test.go` to confirm output equivalence.
- **Dependencies**: None.
- **Effort**: small (1 day)

---

## Gap K: Story Theme Tables Refactor (Tracked TODO)

- **Intended Behavior**: Genre→theme mapping centralized so it cannot drift between `selectTheme` and `getThemesForGenre`.
- **Current State**: `pkg/procgen/story/generator.go:303` carries `TODO(REM-096)`. Both helpers contain duplicated `switch genreID` blocks.
- **Blocked Goal**: Procgen maintainability (low impact today).
- **Implementation Path**:
  1. Define `var genreThemes = map[string][]string{ "fantasy": {...}, "scifi": {...}, "horror": {...}, "cyberpunk": {...}, "postapoc": {...} }` at package scope.
  2. Rewrite `selectTheme` and `getThemesForGenre` to read from the table.
  3. Add `TestSelectTheme_AllGenresCovered` to assert each genre has ≥1 theme.
- **Dependencies**: None.
- **Effort**: trivial (<1 day)

---

## Gap L: Native Mobile Keyboard Integration (Tracked TODO)

- **Intended Behavior**: Programmatic show/hide of the on-screen keyboard on iOS and Android, matching the WASM bridge's capability.
- **Current State**: `pkg/mobile/keyboard_default.go:19` documents the gap with a TODO. `IsKeyboardSupported()` returns `false` on non-WASM, and OS-level keyboard input still works (system soft-keyboard appears on text-field focus). The gap is purely the lack of programmatic control.
- **Blocked Goal**: Polished mobile UX (low impact; OS keyboard suffices for typing).
- **Implementation Path**:
  1. Create `pkg/mobile/keyboard_android.go` with `//go:build android` invoking `InputMethodManager.showSoftInput` via the `gomobile`/`ebitenmobile` JNI bridge.
  2. Create `pkg/mobile/keyboard_ios.go` with `//go:build ios` invoking `becomeFirstResponder` on a hidden `UITextField`.
  3. Update `IsKeyboardSupported()` to return `true` on those platforms.
  4. Manual validation via `make android-apk` / `make ios-simulator`.
- **Dependencies**: None.
- **Effort**: medium (2–4 days; cgo/JNI plumbing)

---

## Gap M: Disabled Test File Should Be Re-Enabled or Removed

- **Intended Behavior**: Either the integration test runs as part of `make test` (possibly behind a build tag) or the file does not exist.
- **Current State**: `pkg/engine/high_throughput_integration_test_disabled.go` is present but has the `_disabled.go` suffix that makes Go's test runner skip it (no `_test.go` suffix means it is also not a test).
- **Blocked Goal**: Confidence in high-throughput integration scenarios (the test was presumably written for a reason).
- **Implementation Path**: Pick one:
  - Re-enable: rename to `high_throughput_integration_test.go`, add a `//go:build integration` build tag, and add a `make test-integration` target (or extend the existing one) to run it.
  - Remove: `git rm pkg/engine/high_throughput_integration_test_disabled.go`.
- **Dependencies**: None.
- **Effort**: trivial (<30 minutes for either path)

---

## Summary

| Gap | Severity | Effort | Stated Goal Affected |
|---|---|---|---|
| A — Voice server-side routing | 🔴 CRITICAL | small | Voice chat |
| B — Client-side prediction unwired | 🟠 HIGH | medium | High-latency multiplayer |
| C — 4 engine systems unregistered | 🟠 HIGH | small per system | Achievements, NPC behavior, carry, commerce |
| D — `AnimationSyncManager` unwired | 🟠 HIGH | small | Multiplayer animation smoothness |
| E — Chunk compression/modification unwired | 🟡 MEDIUM | medium | Memory budget, persistent world |
| F — XP not awarded on kill | 🟡 MEDIUM | small | Character progression |
| G — `AuthManager` orphan | 🟡 MEDIUM | small | Code clarity (federation auth works via other path) |
| H — `BandwidthMonitor` orphan | 🟡 MEDIUM | small | Observability completeness |
| I — `NewBlueprint` orphan | 🟡 MEDIUM | trivial | Determinism hazard |
| J — Quest template refactor | 🟢 LOW | small | Modding extensibility |
| K — Story theme refactor | 🟢 LOW | trivial | Maintainability |
| L — Native mobile keyboard | 🟢 LOW | medium | Mobile UX polish |
| M — Disabled integration test | 🟢 LOW | trivial | Test coverage hygiene |

**Totals**: 1 CRITICAL · 3 HIGH · 5 MEDIUM · 4 LOW = 13 verified open gaps.

*Validated 2026-04-20 against `git build ./... && go vet ./...` (both clean) and `grep`-verified for every constructor/symbol cited.*
