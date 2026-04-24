# SYNC AUDIT — 2026-04-24

> **Legacy Compatibility** — the previous audit in this file (2026-04-23) carried IDs tied to
> implementation-gap findings. Those findings are preserved verbatim in the section
> **[Legacy: Implementation Gap Audit 2026-04-23]** at the bottom of this file so that any
> downstream references (CI scripts, issue trackers, GAPS.md cross-links) remain valid.

---

## Project Concurrency Model

Venture uses a hybrid concurrency model across its 126 packages:

- **Main game loop** (`pkg/engine`, `cmd/client`): Ebiten-driven single-thread ECS update/draw; concurrency is used only during startup (`initializeUIComponents` fan-out, `initializeCriticalSystems` fan-out) and for background I/O goroutines.
- **Multiplayer server** (`cmd/server`, `pkg/network`): goroutine-per-connection model protected by `sync.RWMutex` on shared maps; channels carry cross-goroutine messages.
- **Procedural generation** (`pkg/procgen`, `pkg/rendering/sprites`, `pkg/rendering/cache`): worker-pool fan-out for CPU-bound batch generation; `sync.WaitGroup` + buffered channels; most generators are stateless or carry their own `sync.Mutex`.
- **Federation / WebRTC** (`pkg/network/federation`): long-lived background goroutines controlled by context cancellation or `done`/`stopChan` channels; connection pool has its own `sync.WaitGroup`.
- **Object pools** (`pkg/rendering/particles`, `pkg/engine/projectile_pool`, etc.): `sync.Pool` for GC-pressure reduction; package-level atomics for pool statistics.

Go toolchain: **go 1.24.5** (per `go.mod`). All loop variables are per-iteration by language spec; no Go ≤1.21 loop-capture bugs apply.

---

## Concurrency Inventory

Counts are over non-test `.go` files.

| Metric | Count |
|--------|------:|
| `go func` goroutine launches | 73 |
| Channel declarations (`chan `) | 125 |
| `sync.*` usages | 304 |
| `atomic.*` usages | 44 |
| `context.Context` / `ctx.Done` / `ctx.Err` | 166 |

**Top files by `sync.*` density (non-test):**

| File | `sync.*` hits |
|------|-------------:|
| `pkg/rendering/particles/pool.go` | 11 |
| `pkg/engine/animation_system.go` | 9 |
| `pkg/engine/projectile_pool.go` | 8 |
| `pkg/network/desync.go` | 6 |
| `pkg/rendering/sprites/cache.go` | 6 |
| `pkg/engine/status_effect_pool.go` | 6 |
| `cmd/server/v9_validation.go` | 6 |
| `pkg/rendering/sprites/pool.go` | 5 |
| `pkg/network/voice_transport.go` | 5 |
| `pkg/network/server.go` | 5 |

**Packages with most goroutine launches (non-test):**

| File | `go func` count |
|------|---------------:|
| `pkg/engine/game.go` | 15 |
| `cmd/client/handlers.go` | 11 |
| `pkg/network/federation/webrtc/signaling.go` | 2 |
| `pkg/network/federation/webrtc/relay.go` | 2 |
| `pkg/rendering/sprites/cache.go` | 2 |
| `pkg/rendering/sprites/pool.go` | 2 |
| `cmd/server/main.go` | 3 |

---

## Race Detector Results

`go test -race` was executed against every package buildable in the headless CI environment (Ebiten-dependent packages fail to build due to missing `X11/Xlib.h`; this is a known environmental constraint, not a code defect).

**Packages tested with `-race` — all PASSED:**

| Package | Result |
|---------|--------|
| `pkg/procgen/terrain` | ✅ pass |
| `pkg/procgen/legendary` | ✅ pass |
| `pkg/procgen/building` | ✅ pass |
| `pkg/procgen/item` | ✅ pass |
| `pkg/procgen/dialog` | ✅ pass |
| `pkg/procgen/entity` | ✅ pass |
| `pkg/procgen/environment` | ✅ pass |
| `pkg/procgen/furniture` | ✅ pass |
| `pkg/procgen/genre` | ✅ pass |
| `pkg/procgen/magic` | ✅ pass |
| `pkg/procgen/narrative` | ✅ pass |
| `pkg/procgen/puzzle` | ✅ pass |
| `pkg/procgen/quest` | ✅ pass |
| `pkg/procgen/skills` | ✅ pass |
| `pkg/procgen/station` | ✅ pass |
| `pkg/procgen/story` | ✅ pass |
| `pkg/world` | ✅ pass |
| `pkg/world/economy` | ✅ pass |
| `pkg/world/territory` | ✅ pass |
| `pkg/world/raids` | ✅ pass |
| `pkg/audio` | ✅ pass |
| `pkg/audio/music` | ✅ pass |
| `pkg/audio/sfx` | ✅ pass |
| `pkg/audio/synthesis` | ✅ pass |
| `pkg/network/federation/guild` | ✅ pass |
| `pkg/network/federation/mobile` | ✅ pass |
| `pkg/network/federation/webrtc` | ✅ pass |
| `pkg/network/resilience` | ✅ pass |

**Packages that failed to BUILD (Ebiten/X11 dependency, not a race):**
`pkg/network`, `pkg/network/chat`, `pkg/network/federation`, `pkg/network/trade`,
`pkg/procgen/audit`, `pkg/procgen/book`, `pkg/procgen/class`, `pkg/procgen/companion`,
`pkg/procgen/faction`, `pkg/procgen/minigame`, `pkg/procgen/recipe`, `pkg/procgen/vehicle`,
`pkg/world/housing`, `cmd/server`.

No data races were reported by the detector on any testable package.

---

## go vet Results

`go vet ./pkg/...` and `go vet ./cmd/...` both fail with the same X11/GLFW build error that prevents Ebiten from compiling in the headless environment. This is an environmental constraint; no vet output is available. Manually reviewed files show no vet-detectable patterns (shadow, printf format mismatch, unreachable code, etc.) in the concurrency-critical paths examined during this audit.

---

## Findings

### HIGH

- [ ] **`FederatedMarket.Stop()` — unguarded `close(stopChan)` causes panic on double-call**
  - **File:** `pkg/network/federation/market.go:112`
  - **Execution path:** Any caller that calls `market.Stop()` more than once — including the common `defer market.Stop()` + explicit `market.Stop()` pattern, or two concurrent goroutines both calling `Stop()`.
  - **Root cause:** `Stop()` calls `close(m.stopChan)` with no idempotency guard (no `sync.Once`, no `stopped bool` flag, no nil/closed check). Go panics on `close` of an already-closed channel.
  - **Evidence:** `pkg/network/federation/market.go:108-113`:
    ```go
    func (m *FederatedMarket) Stop() {
        if m.updateTicker != nil {
            m.updateTicker.Stop()
        }
        close(m.stopChan)   // ← panics on second call
    }
    ```
  - **Remediation:** Add a `sync.Once` or a `sync.Mutex`-protected `stopped bool` flag:
    ```go
    func (m *FederatedMarket) Stop() {
        m.stopOnce.Do(func() {
            if m.updateTicker != nil {
                m.updateTicker.Stop()
            }
            close(m.stopChan)
        })
    }
    ```
    Add a `stopOnce sync.Once` field to `FederatedMarket`. Update `NewFederatedMarket` accordingly. Add a unit test calling `Stop()` twice.

---

### MEDIUM

- [ ] **`TCPServer.Start()` — deferred unlock fires on manually-unlocked mutex if panic occurs during goroutine launch window**
  - **File:** `pkg/network/server.go:213-248`
  - **Execution path:** `TCPServer.Start()` is called during server startup (`cmd/server/main.go`).
  - **Root cause:** The function uses `Lock()` → `defer Unlock()` → explicit `Unlock()` → goroutine spawns → explicit `Lock()`. The `defer` was registered against the first `Lock()`. If a panic fires between the explicit `Unlock()` (line 230) and the explicit `Lock()` (line 237) — during `wg.Add(1)` or the goroutine launch — the deferred `Unlock()` fires against an already-unlocked mutex, causing a secondary panic that masks the original error.
  - **Evidence:** `pkg/network/server.go:229-237`:
    ```go
    // VULNERABLE WINDOW — between explicit Unlock() and explicit Lock():
    s.clientsMu.Unlock()   // manual unlock; defer still registered
    s.wg.Add(1)            // currently cannot panic, but any future fallible
    go s.acceptLoop()      // call added here would trigger double-unlock on panic
    s.wg.Add(1)
    go s.cleanupLoop()
    s.clientsMu.Lock()     // re-lock so defer can unlock it on return
    ```
  - **Risk:** Low probability (the intervening code cannot currently panic), but the pattern is fragile: any future addition of a fallible operation between lines 230–237 silently introduces a double-unlock bug.
  - **Remediation:** Replace the sandwich with `defer`-free manual locking for the entire function, or restructure to avoid the unlock/re-lock:
    ```go
    func (s *TCPServer) Start() error {
        s.clientsMu.Lock()
        if s.running {
            s.clientsMu.Unlock()
            return fmt.Errorf("server already running")
        }
        listener, err := net.Listen("tcp", s.config.Address)
        if err != nil {
            s.clientsMu.Unlock()
            return fmt.Errorf("failed to listen on %s: %w", s.config.Address, err)
        }
        s.listener = listener
        s.running = true
        s.clientsMu.Unlock()
        s.wg.Add(1)
        go s.acceptLoop()
        s.wg.Add(1)
        go s.cleanupLoop()
        // ... log and return
        return nil
    }
    ```

- [ ] **`FederatedMarket.Start()` — no idempotency guard causes goroutine and ticker leak on double-call**
  - **File:** `pkg/network/federation/market.go:92-106`
  - **Execution path:** Any code that calls `market.Start()` more than once (no documentation or guard prevents this).
  - **Root cause:** `Start()` creates a new `time.Ticker` and assigns it to `m.updateTicker`, overwriting the previous value without calling `Stop()` on it, and launches a new goroutine that also listens on `m.stopChan`. On double-call: (1) the first ticker is leaked (its goroutine fires every 60 s forever), (2) two goroutines call `UpdatePrices()` concurrently (safe because `UpdatePrices` acquires `m.mu`, but wasteful).
  - **Evidence:** `pkg/network/federation/market.go:92-106`:
    ```go
    func (m *FederatedMarket) Start() {
        m.updateTicker = time.NewTicker(60 * time.Second) // overwrites without stopping
        go func() { ... }()                               // second goroutine on re-call
    }
    ```
  - **Remediation:** Add a `sync.Once` or a `sync.Mutex`-protected `started bool`:
    ```go
    func (m *FederatedMarket) Start() {
        m.startOnce.Do(func() {
            m.updateTicker = time.NewTicker(60 * time.Second)
            go func() { ... }()
        })
    }
    ```
    Add a `startOnce sync.Once` field; add a unit test calling `Start()` twice and verifying only one ticker goroutine runs.

---

### LOW

- [ ] **Multiple non-`defer` `Lock()`/`Unlock()` pairs in short code paths — fragile on future edits**
  - **Files and representative lines:**
    - `pkg/procgen/audit_registry.go:22,29` — package-level `auditMu` manually locked/unlocked in `RegisterAuditEntry` and `GetAuditEntries`
    - `pkg/companion/learning/manager.go:112,155` — `m.mu.Lock()` … `m.mu.Unlock()` in `AddCompanion` / `RemoveCompanion`
    - `pkg/procgen/legendary/manager.go:104,143,395,406,431,439,478` — multiple manual lock/unlock pairs
    - `pkg/world/territory.go:195-206` — `tm.mu.RLock()` loop `tm.mu.RUnlock()` in `GetResourceBonus`
    - `pkg/procgen/terrain/async_loader.go:54,67,81` — manual lock/unlock in error branches of background goroutine
  - **Root cause:** All locks protect only a single assignment or map operation with no intermediate returns; the pattern is currently safe. However, `defer` is idiomatic Go for mutex unlock and prevents lock-leak if an early return or panic is added later.
  - **Remediation:** Replace non-`defer` `Unlock()` calls with `defer mu.Unlock()` immediately after `mu.Lock()`. Where early exit before the critical section is needed, extract the guarded block into a helper that uses `defer`.

- [ ] **`startLegacyMetricsMonitor` goroutine has no cancellation path — permanent goroutine leak**
  - **File:** `cmd/client/init_monitoring.go:52-63`
  - **Execution path:** Called once during client startup; the goroutine ticks forever with no `context.Context`, stop channel, or `done` signal.
  - **Evidence:**
    ```go
    go func() {
        ticker := time.NewTicker(perfMonitorInterval * time.Second)
        defer ticker.Stop()
        for range ticker.C {          // no select on ctx.Done or done channel
            metrics := perfMonitor.GetMetrics()
            ...
        }
    }()
    ```
  - **Risk:** On client shutdown the goroutine is abandoned (not joined). For a single-process desktop client this is benign; for test harnesses that create multiple clients it accumulates leaked goroutines.
  - **Remediation:** Accept a `context.Context` and select on `ctx.Done()`:
    ```go
    go func() {
        ticker := time.NewTicker(perfMonitorInterval * time.Second)
        defer ticker.Stop()
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                ...
            }
        }
    }()
    ```

- [ ] **`ProjectileNetworkSync.CleanupTask()` stop channel is undocumented and callers may leak the goroutine**
  - **File:** `pkg/network/projectile_sync.go:441-463`
  - **Root cause:** `CleanupTask()` returns a `chan<- struct{}` that callers must send to in order to stop the background cleanup goroutine. The function comment says "send a value to stop" but there is no `defer`-safe idiom at call sites, and omitting the send leaks the goroutine.
  - **Remediation:** Change the API to accept a `context.Context` (consistent with other systems), or at minimum add a `//nolint:contextcheck` + prominent GoDoc warning at each call site, and add a test that verifies the goroutine exits when the stop channel is signalled.

---

## False Positives Considered and Rejected

| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| Loop variable capture in goroutines (`pkg/network/federation/webrtc/relay.go:417`, `pkg/rendering/sprites/cache.go:460-462`, etc.) | Go 1.24.5 (`go.mod: go 1.24.5`) introduces per-iteration loop variable semantics; closure capture is safe. |
| `TCPVoiceTransport.BufferVoicePacket` — TOCTOU between outer `jitterBuffersMu` release and inner `jb.mu` acquisition (`pkg/network/voice_transport.go:333-334`) | `jb` pointer is obtained while holding the outer lock. `Close()` replaces the map (`t.jitterBuffers = make(...)`) under the same lock, but the `jb` variable in `BufferVoicePacket` holds a reference to the old struct which the GC keeps live. The packet safely goes to the orphaned buffer (no use-after-free, no crash). Not a safety issue. |
| `PeriodicSyncManager.Start()` goroutine accesses `p.syncFunc` and `p.detector` without lock (`pkg/network/desync.go:381-393`) | Both fields are written exactly once at construction (`NewPeriodicSyncManager`) and never modified afterward. The goroutine is launched after construction completes, establishing a happens-before relationship via the goroutine start. Read-only access from that goroutine is safe without a lock. |
| `pkg/engine/ecs.go` dual-mutex ordering (`w.mu` RWMutex + `w.entityMu` Mutex) | Code always releases `w.mu` before acquiring `w.entityMu` (sequential, not nested); no lock-order inversion is possible. |
| `pkg/engine/game.go` `initializeUIComponents` WaitGroup usage | `wg.Add(N)` is called before the corresponding `go func()` launches; all goroutines call `defer wg.Done()`; `wg.Wait()` is called after all launches. Textbook-correct pattern. |
| `pkg/rendering/sprites/cache.go` `BatchGenerate` — goroutine closing `resultsChan` | The goroutine calls `wg.Wait()` then `close(resultsChan)`. The `wg` is passed by pointer and all worker goroutines call `defer wg.Done()`. `collectCachedResults` drains `resultsChan` until close. No goroutine leak; no double-close. |
| `go vet ./...` failures | Environmental; caused by missing `X11/Xlib.h` for Ebiten's GLFW bindings, not a code defect. |
| `pkg/world/economy`, `pkg/world/territory`, `pkg/procgen/legendary` mutex patterns | Race detector passes on all three packages; manual review confirms all lock/unlock pairs are well-bracketed. |

---

---

# Legacy: Implementation Gap Audit 2026-04-23

> The following section is preserved verbatim from the prior audit to maintain legacy ID compatibility.
> Findings marked `[x]` were resolved before the sync audit was conducted.

## Project Architecture Overview (2026-04-23)
Venture is a zero-asset, deterministic procedural multiplayer action-RPG (`README.md`) built around ECS (`docs/ARCHITECTURE.md` ADR-001), runtime-generated content (ADR-002), authoritative high-latency networking (ADR-003), and package-domain ownership (`engine`, `procgen`, `rendering`, `audio`, `network`, `world`).

Phase 0 evidence reviewed:
- Purpose and goals: `/home/runner/work/venture/venture/README.md`
- Architecture decisions: `/home/runner/work/venture/venture/docs/ARCHITECTURE.md`
- Planned/unimplemented areas: `/home/runner/work/venture/venture/ROADMAP.md`
- Module/deps: `/home/runner/work/venture/venture/go.mod`
- Package graph: `go list ./...` (126 packages enumerated)

Phase 1 online research (brief):
- Open issues: none (`repo:opd-ai/venture is:issue is:open`)
- Open PRs are dependency/CI bumps (Dependabot)
- Recent merged PRs show prior audit backlogs already addressed (e.g. #485, #487, #489), leaving a small set of remaining implementation gaps mostly around real VR hardware support.

Phase 2 baseline:
- `go-stats-generator analyze . --skip-tests --format json ...` completed (`/tmp/gap-audit-metrics.json`)
- `go-stats-generator analyze . --skip-tests` completed (`/tmp/gap-audit-summary.txt`)
- `go build ./...` fails in this environment due missing X11 headers (`X11/Xlib.h`) from Ebiten native prerequisites
- `go vet ./...` fails for same environment prerequisite reason

## Gap Summary (2026-04-23)
| Category | Count | Critical | High | Medium | Low |
|----------|-------|----------|------|--------|-----|
| Stubs/TODOs | 1 | 0 | 1 | 0 | 0 |
| Dead Code | 1 | 0 | 0 | 1 | 0 |
| Partially Wired | 2 | 0 | 0 | 1 | 1 |
| Interface Gaps | 1 | 0 | 0 | 0 | 1 |
| Dependency Gaps | 0 | 0 | 0 | 0 | 0 |

## Findings (2026-04-23)
### HIGH
- [ ] OpenXR runtime adapters are still placeholder stubs — `/home/runner/work/venture/venture/pkg/engine/vr_openxr_adapters.go:94`, `:117`, `:126`, `:147`, `:160`, `:176`, `:183`, `:191`, `:199`, `:207`, `:215`, `:224` — methods return static defaults/no-op and TODO markers remain; real SDK calls are absent — blocks ROADMAP Priority 4 validation that VR works with real hardware (`/home/runner/work/venture/venture/ROADMAP.md:157-162`) — **Remediation:** implement OpenXR session/action plumbing in `NewOpenXRHeadsetAdapter`/`NewOpenXRControllerAdapter` and replace placeholder getters with real `xrLocateViews`/`xrGetActionState*` reads, then add hardware-backed integration tests behind `//go:build vr` and verify with `go build -tags vr ./...` and targeted VR tests.

### MEDIUM
- [ ] VR runtime OpenXR path is effectively unreachable in current behavior — `/home/runner/work/venture/venture/pkg/engine/vr_adapter_factory_openxr.go:8-12`, `:18-22`; `/home/runner/work/venture/venture/pkg/engine/vr_openxr_adapters.go:82`, `:112`, `:177` — factory only selects OpenXR adapter when `IsConnected()` is true, but adapter never sets `connected` true due missing SDK integration, so flow always falls back to stubs — player-visible VR hardware path cannot activate even in `-tags vr` builds — **Remediation:** set `connected` based on successful OpenXR runtime/session/action initialization and add explicit adapter-state tests covering both fallback and connected-path selection.
- [x] `LightingConfig.EnableShadows` is now consumed by rendering logic — it forces ambient-occlusion processing in `ApplyFullPostProcessing` as a legacy compatibility toggle, with test coverage (`pkg/rendering/lighting/system.go`, `pkg/rendering/lighting/ambient_occlusion_test.go`) and updated field contract in `types.go`.

### LOW
- [ ] WebXR adapter contract is documented but implementation file is still absent — `/home/runner/work/venture/venture/pkg/vr/doc.go:79`, `:84` references future `pkg/engine/vr_webxr_adapters.go`; file does not exist in current tree — WASM VR implementation remains a documented placeholder — **Remediation:** add `pkg/engine/vr_webxr_adapters.go` under appropriate build tags (`js`) implementing `VRHeadsetAdapter`/`VRControllerAdapter` via `syscall/js`, with smoke tests or compile-time interface checks for js builds.

## False Positives Considered and Rejected (2026-04-23)
| Candidate Finding | Reason Rejected |
|-------------------|-----------------|
| `go build ./...` and `go vet ./...` failures indicate core implementation breakage | Rejected: failures are environmental (`X11/Xlib.h` missing) and match documented native prerequisites in README build instructions. |
| Large number of `return nil` matches are stubs | Rejected: broad text search mostly captured normal error-handling/control-flow returns; not evidence of incomplete implementation by itself. |
| Interfaces with `implementation_count=0` in go-stats output are unimplemented gaps | Rejected: many are intentional boundary abstractions/public contracts and may be implemented by external callers, build-tagged files, or runtime injection; no direct stub evidence. |
| Deprecated wrappers are implementation gaps | Rejected: deprecations include replacement guidance and generally preserve backward compatibility, not unfinished functionality. |
| Roadmap historical gaps (trade quantity, race CI, staticcheck) remain open | Rejected: recent merged PRs (#485, #489) indicate those items were implemented; direct code checks confirm quantity-aware trade path exists (`/home/runner/work/venture/venture/pkg/network/trade/system.go:140-244`). |
