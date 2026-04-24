# Concurrency Safety Gaps — 2026-04-24

> **Legacy Compatibility** — the previous gaps in this file (2026-04-23) carried IDs for
> implementation gaps (OpenXR stubs, WebXR missing file, etc.). Those entries are preserved
> verbatim in the **[Legacy: Implementation Gaps 2026-04-23]** section at the bottom so
> downstream references remain valid.

---

## Gap 1 — `FederatedMarket.Stop()` Unguarded Channel Close

- **Stated Goal**: `FederatedMarket` should be safely stoppable at any time, including by deferred calls or concurrent shutdown routines.
- **Current State**: `pkg/network/federation/market.go:112` calls `close(m.stopChan)` with no idempotency guard. A second call to `Stop()` panics with `close of closed channel`. No `sync.Once`, `stopped bool`, or nil-check protects the close.
- **Risk**: **HIGH**. Server shutdown paths routinely use `defer market.Stop()` alongside explicit `market.Stop()` calls. A double-call during error handling or test teardown produces an unrecoverable panic.
- **Closing the Gap**:
  1. Add `stopOnce sync.Once` field to `FederatedMarket`.
  2. Wrap the body of `Stop()` in `m.stopOnce.Do(func() { ... })`.
  3. Add a unit test: call `Stop()` twice, assert no panic.
  4. Grep for other structs in `pkg/network/federation/` with bare `close(stopChan)` and apply the same fix. Note: `PeriodicSyncManager.Stop()` in `pkg/network/desync.go` uses a `running` boolean flag — an alternative to `sync.Once` that is acceptable when the struct already tracks running state under a mutex, but `sync.Once` is preferred for new code since it requires no additional lock.

---

## Gap 2 — `FederatedMarket.Start()` Lacks Idempotency Guard

- **Stated Goal**: Calling `market.Start()` multiple times should be a no-op after the first call, consistent with `PeriodicSyncManager.Start()` and `RouteManager.Start()`.
- **Current State**: `pkg/network/federation/market.go:92-106` creates a new `time.Ticker` and launches a new goroutine unconditionally. On double-call: (1) `m.updateTicker` is overwritten, leaking the first ticker's goroutine; (2) two goroutines call `UpdatePrices()` concurrently every 60 s.
- **Risk**: **MEDIUM**. Goroutine and ticker resource leak; redundant concurrent `UpdatePrices()` calls (safe but wasteful).
- **Closing the Gap**:
  1. Add `startOnce sync.Once` field to `FederatedMarket`.
  2. Wrap the body of `Start()` in `m.startOnce.Do(func() { ... })`.
  3. Add a unit test: call `Start()` twice, assert `Stop()` terminates cleanly and only one ticker existed.

---

## Gap 3 — `TCPServer.Start()` Defer-Unlock Fragility

- **Stated Goal**: `TCPServer.Start()` should be panic-safe; any early exit or panic should release `clientsMu` exactly once.
- **Current State**: `pkg/network/server.go:213-248` uses `Lock()` + `defer Unlock()` then explicitly calls `Unlock()` and later `Lock()` around goroutine launches. The deferred `Unlock()` fires against the re-acquired lock on clean return — correct. However, if a panic fires in the window between the explicit `Unlock()` (line 230) and the explicit `Lock()` (line 237), the deferred `Unlock()` fires on an already-unlocked mutex, causing a secondary panic that masks the original error.
- **Risk**: **MEDIUM**. The three lines between unlock and re-lock (`wg.Add`, `go s.acceptLoop()`, `wg.Add`) cannot currently panic. Risk is future fragility: any addition of a fallible call in that window silently introduces a double-unlock bug.
- **Closing the Gap**:
  1. Replace the mixed `defer`/manual pattern with fully manual lock management (explicit `Unlock()` on every exit path), or
  2. Refactor to release the lock before goroutine spawning and not re-acquire it, eliminating the sandwich entirely.
  3. Add a test that verifies `Start()` is race-free when called from two goroutines concurrently.

---

## Gap 4 — Non-`defer` Mutex Unlock Patterns

- **Stated Goal**: All mutex unlock operations should be deferred immediately after the corresponding lock to be resilient against future code additions that introduce early returns or panics.
- **Current State**: Several files use explicit `Lock()` / `Unlock()` pairs without `defer`:
  - `pkg/procgen/audit_registry.go:22,29`
  - `pkg/companion/learning/manager.go:112,155`
  - `pkg/procgen/legendary/manager.go:104,143,395,406,431,439,478`
  - `pkg/world/territory.go:195-206` (`GetResourceBonus`)
  - `pkg/procgen/terrain/async_loader.go:54,67,81`
  All current call paths call `Unlock()` unconditionally; no bugs exist today.
- **Risk**: **LOW**. Non-idiomatic; fragile under future edits. No current correctness issue.
- **Closing the Gap**:
  1. Replace manual `Lock()` / `Unlock()` pairs with `Lock()` + `defer Unlock()` in each listed location.
  2. Where a check-then-act pattern needs to release the lock early (e.g., `audit_registry.go:GetAuditEntries` copies the slice under lock then returns), use a helper function scoped to just the locked region so `defer` still applies.

---

## Gap 5 — `startLegacyMetricsMonitor` Goroutine Has No Cancellation

- **Stated Goal**: All long-lived background goroutines should accept a cancellation mechanism (context or done channel) so they can be terminated cleanly on client shutdown or in tests.
- **Current State**: `cmd/client/init_monitoring.go:52-63` launches a goroutine that ticks forever with no `context.Context`, stop channel, or WaitGroup. The goroutine cannot be joined or cancelled.
- **Risk**: **LOW**. In production desktop use, the process exits when the goroutine is abandoned; no crash. In unit tests or tools that create multiple client instances, goroutines accumulate.
- **Closing the Gap**:
  1. Change `startLegacyMetricsMonitor` to accept a `context.Context`.
  2. Replace `for range ticker.C` with `select { case <-ctx.Done(): return; case <-ticker.C: ... }`.
  3. Thread the client's shutdown context through to this call site.

---

## Gap 6 — `ProjectileNetworkSync.CleanupTask()` Stop Channel Undocumented

- **Stated Goal**: The cleanup goroutine started by `CleanupTask()` should be clearly stoppable, with callers not required to maintain undocumented stop semantics.
- **Current State**: `pkg/network/projectile_sync.go:441-463` returns a `chan<- struct{}` that callers must send to in order to stop the goroutine. If callers ignore or lose the channel, the goroutine leaks silently.
- **Risk**: **LOW**. No current caller loses the channel, but the API is error-prone.
- **Closing the Gap**:
  1. Change `CleanupTask()` to accept a `context.Context` and return nothing (or return an error).
  2. Select on `ctx.Done()` instead of the stop channel.
  3. Update all call sites to pass the server's shutdown context.

---

---

# Legacy: Implementation Gaps 2026-04-23

> The following section is preserved verbatim from the prior GAPS.md to maintain legacy ID
> compatibility with downstream references.

## OpenXR Adapter Logic Is Stubbed (Desktop VR)
- **Intended Behavior**: In `-tags vr` builds, desktop VR adapters should bind to OpenXR runtime/session and provide real headset/controller state.
- **Current State**: `pkg/engine/vr_openxr_adapters.go` contains TODO-backed placeholders and default-return methods (`0/false/no-op`) at lines 94, 117, 126, 147, 160, 176, 183, 191, 199, 207, 215, 224.
- **Blocked Goal**: ROADMAP Priority 4 items requiring real head tracking/controller input and real-hardware validation (`ROADMAP.md:157-162`).
- **Implementation Path**:
  1. Implement OpenXR runtime initialization in `NewOpenXRHeadsetAdapter` (`xrCreateInstance`, `xrGetSystem`, `xrCreateSession`).
  2. Implement pose reads (`xrLocateViews`) in `GetHeadOrientation`/`GetHeadPosition`.
  3. Implement controller action setup and state polling (`xrSyncActions`, `xrGetActionState*`) in controller adapter methods.
  4. Implement haptics via `xrApplyHapticFeedback`.
  5. Add build-tagged integration tests for successful runtime session and non-zero state reads when hardware/runtime is available.
- **Dependencies**: OpenXR SDK availability in CI/dev environments and cgo linker configuration.
- **Effort**: large

## OpenXR Runtime Path Never Selected at Runtime
- **Intended Behavior**: `NewRuntimeHeadsetAdapter` / `NewRuntimeControllerAdapter` should choose OpenXR adapters when a valid VR runtime/hardware session is available.
- **Current State**: Factory code (`pkg/engine/vr_adapter_factory_openxr.go:8-12`, `18-22`) gates on `IsConnected*`; OpenXR adapters never set `connected` true due missing integration (`pkg/engine/vr_openxr_adapters.go:82,112,177`).
- **Blocked Goal**: Real VR runtime route cannot activate, so VR remains stub-only even in vr-tagged builds.
- **Implementation Path**:
  1. Set `connected` true only after successful runtime/session/action initialization.
  2. Add explicit adapter-state tests proving both branches: fallback-to-stub and connected-OpenXR.
  3. Ensure error logging distinguishes "SDK unavailable" vs "runtime unavailable" vs "hardware unavailable".
- **Dependencies**: Completion of OpenXR adapter integration gap above.
- **Effort**: medium

## LightingConfig.EnableShadows Is a No-Op API Field
- **Intended Behavior**: Public config fields should either affect behavior or be removed/deprecated with migration path.
- **Current State**: `pkg/rendering/lighting/types.go:116-122` explicitly states `EnableShadows` has no implementation; no operational usage in lighting runtime path.
- **Blocked Goal**: Misleading API and configuration contract for lighting behavior.
- **Implementation Path**:
  1. Decide ownership: rendering-lighting shadow toggle vs engine-level shadow system toggle.
  2. If retained, wire field into active render/update logic and add tests for enabled/disabled behavior.
  3. If removed, deprecate and migrate callers to canonical shadow control path with clear release note.
- **Dependencies**: Alignment between `pkg/rendering/lighting` and `pkg/engine` shadow systems.
- **Effort**: medium

## WebXR Adapter File Is Documented but Missing
- **Intended Behavior**: WASM VR path should have a js-build adapter implementation per package documentation.
- **Current State**: `pkg/vr/doc.go:79,84` references future `pkg/engine/vr_webxr_adapters.go`, but no such file exists.
- **Blocked Goal**: Documented WASM VR integration path is not yet implementable.
- **Implementation Path**:
  1. Add `pkg/engine/vr_webxr_adapters.go` with `//go:build js` constraints.
  2. Implement `VRHeadsetAdapter` and `VRControllerAdapter` via `syscall/js` WebXR APIs.
  3. Add compile-time interface assertions and js-target smoke tests.
- **Dependencies**: Browser WebXR support and js-target test strategy.
- **Effort**: medium
