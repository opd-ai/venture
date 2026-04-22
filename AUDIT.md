# MEMORY AUDIT — 2026-04-21

## Legacy Cross-Reference Compatibility (for existing repo comments)
This file keeps historical `AUDIT.md` identifiers stable for existing references in code comments and docs.

> **Legacy ID note**: Existing references to historical IDs (for example `Gap 1`, `Gap #3`, `Task 7`, `G5`) refer to prior audit/backlog semantics. New memory findings in this document use `MA-*` identifiers.

| Legacy ID | Historical topic (from existing references) |
|-----------|---------------------------------------------|
| `Gap 1` | Voice transport wiring / multiplayer desync follow-up |
| `Gap #3` | First-touch/input-delay issue references |
| `Task 2` | Guild fleet integration wiring |
| `Task 3` | Enhanced chat / validation integration |
| `Task 4` | Post office / courier integration wiring |
| `Task 6` | Housing + crafting integration wiring |
| `Task 7` | VR hardware detection and conditional initialization |
| `Task 8` | Server-side chat history persistence |
| `Task 9` | Companion housing integration wiring |
| `Task 10` | Narrative world integration wiring |
| `Task 12` | Political warfare integration wiring |
| `G4` | Archaeological/timeline spawning integration |
| `G5` | Story journal UI wiring and input binding |
| `G6` | Genre component reward/theming integration |

## Project Memory Profile
- **Type**: long-running client + dedicated server game runtime (leak-sensitive), plus short-lived tools/tests.
- **Stated goals**: 60 FPS target and memory targets of **<500MB client / <1GB server** (`docs/PERFORMANCE_TUNING.md:3,16-19`; `docs/PERFORMANCE.md:13-15`).
- **Scale indicators**: 123 packages as of 2026-04-21 snapshot (`go list ./...`), high-throughput subsystems in `pkg/engine`, `pkg/network`, `pkg/rendering`, `pkg/procgen`, `pkg/world`.
- **Memory-sensitive deps from `go.mod`**: Ebiten (image/runtime memory behavior), Logrus (field map allocations), stdlib `sync.Pool` heavily used.

## Baseline & Tooling Evidence
- `go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages,structs` reported:
  - LOC: 215,840
  - Functions: 4,069
  - Methods: 12,878
  - Structs: 2,259
  - Packages: 106 (stats scope)
- `go list ./...`: 123 packages.
- `go test -race -count=1 ./...`: partially blocked by missing X11 headers in this environment (`X11/Xlib.h`), but many non-Ebiten packages executed.
- `go vet ./...`: blocked by same X11 header issue.

Reproduction commands:

```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages,structs
go-stats-generator analyze . --skip-tests
go list ./...
go test -race -count=1 ./...
go vet ./...
go test -run '^$' -bench 'BenchmarkTerrainCache_Get|BenchmarkTerrainCache_Put|BenchmarkGenerateGradient_Linear|BenchmarkGenerateGradient_Radial|BenchmarkGenerateGradient_Angular' -benchmem ./pkg/procgen/terrain ./pkg/rendering/palette
go test -run '^$' -bench 'BenchmarkPeer_Send|BenchmarkSignalingClient_SendOffer|BenchmarkSignalingServer_RelayMessage' -benchmem ./pkg/network/federation/webrtc
go build -gcflags='-m' ./pkg/network/federation/mobile ./pkg/network/federation/webrtc ./pkg/network ./pkg/engine
```

## Memory Inventory
| Package | unsafe | sync.Pool | Large Allocs | Closures | cgo | Global State |
|---------|--------|-----------|-------------|----------|-----|-------------|
| `pkg/network/federation/webrtc` | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ bounded maps/tickers |
| `pkg/network/federation/mobile` | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ bounded state |
| `pkg/network` | ❌ (non-test) | ✅ (`buffer_pool.go`) | ⚠️ message buffers | ✅ | ❌ | ✅ |
| `pkg/engine` | ❌ (non-test) | ✅ (multiple pools) | ⚠️ world/entity slices | ✅ | ❌ | ✅ |
| `pkg/rendering/pool` | ❌ | ✅ | ✅ image buffers | ✅ | ❌ | ✅ |
| `pkg/rendering/particles` | ❌ | ✅ | ✅ particle buffers | ✅ | ❌ | ✅ |
| `pkg/procgen/terrain` | ❌ | ✅ (`hasherPool`) | ✅ terrain clone/copy paths | ✅ | ❌ | ✅ cache w/ capacity |
| `pkg/mobile` | ✅ (`keyboard_android.go`) | ❌ | ❌ | ✅ | ✅ (JNI bridge, no C malloc usage found) | ✅ |

Pattern scan highlights:
- `reflect.SliceHeader`/`reflect.StringHeader`: none found.
- `runtime.SetFinalizer`: none found.
- `C.CString`/`C.CBytes`/`C.malloc`: none found.
- `unsafe.Pointer` in non-test code: `pkg/mobile/keyboard_android.go` JNI bridging only.

## Phase 1 Research (brief)
- No repository issues found for memory/OOM/leak queries (`github issue search` for `repo:opd-ai/venture memory leak allocation GC OOM`).
- Relevant external guidance confirmed and applied in review criteria:
  - `time.After` in loops/selects can create avoidable timer allocations/retention.
  - `sync.Pool` requires explicit object reset before `Put` (project generally follows this).
  - Structured logging field maps can allocate in hot paths.

## Findings
### CRITICAL
- [ ] None.

### HIGH
- [ ] None.

### MEDIUM
- [ ] **MA-1** `time.After` in long-lived loop allocates a new timer every iteration — `pkg/network/federation/webrtc/peer.go:149` — **Evidence:** `processMessages()` loop uses `case <-time.After(100 * time.Millisecond)` instead of a reusable ticker; this loop runs for peer lifetime. — **Impact:** cumulative timer allocations and GC churn per connected peer. — **Remediation:** replace loop-local `time.After` with `ticker := time.NewTicker(100 * time.Millisecond)` + `defer ticker.Stop()` and use `case <-ticker.C`. **Validation:** `go test -run '^$' -bench BenchmarkPeer_Connect -benchmem ./pkg/network/federation/webrtc` and compare allocs/op before/after.

- [ ] **MA-2** Fast-path signaling send APIs allocate timeout timers even when send succeeds immediately — `pkg/network/federation/webrtc/peer.go:182`, `pkg/network/federation/webrtc/signaling.go:154`, `pkg/network/federation/webrtc/signaling.go:179`, `pkg/network/federation/webrtc/signaling.go:204`, `pkg/network/federation/webrtc/signaling.go:228`, `pkg/network/federation/webrtc/signaling.go:371` — **Evidence:** each method uses `select { case sendChan <- ...; case <-time.After(5 * time.Second) }`; `time.After` is evaluated on entry, so timer allocation occurs even on success. Benchmarks show per-call allocations: `BenchmarkPeer_Send` = **496 B/op, 6 allocs/op** and `BenchmarkSignalingClient_SendOffer` = **344 B/op, 4 allocs/op**. — **Impact:** avoidable per-message heap pressure on network hot paths. — **Remediation:** use reusable timers (`time.NewTimer` + `Stop`/drain) or non-allocating backpressure strategy (e.g., bounded non-blocking send + caller-level retry policy). **Validation:** rerun `go test -run '^$' -bench 'BenchmarkPeer_Send|BenchmarkSignalingClient_SendOffer|BenchmarkSignalingServer_RelayMessage' -benchmem ./pkg/network/federation/webrtc` and confirm reduced B/op + allocs/op.

### LOW
- [ ] **MA-3** Bandwidth-throttle retry path uses `time.After` inside a retry loop — `pkg/network/federation/mobile/adapter.go:263` — **Evidence:** `executeWithBandwidthLimit` loops with `case <-time.After(waitTime)` / `case <-ctx.Done()`. On cancellation, timers can remain until firing. — **Impact:** minor transient timer retention; mostly relevant under frequent cancellations. — **Remediation:** replace with `time.NewTimer(waitTime)` and explicit `Stop`/drain on `ctx.Done()`. **Validation:** add/extend micro-benchmark for cancellation-heavy path and compare allocs/op with `-benchmem`.

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| Terrain cache warm-hit allocations in `pkg/procgen/terrain/cache.go:159-177` | **Rejected as new issue**: code explicitly documents defensive clone-on-get behavior to prevent cache corruption (`cache.go:159-166`), and tests enforce immutability expectations (`cache_test.go:467-473`). This is an acknowledged memory/perf tradeoff. |
| Gradient generation allocation concern in `pkg/rendering/palette/gradient.go` | Rejected: benchmark shows only 2 allocs/op (`BenchmarkGenerateGradient_*`) corresponding to expected image/buffer allocation; no excess interface-boxing pattern remains (uses `SetRGBA`, concrete `color.RGBA`). |
| MovementSystem visited-cell unbounded growth | Rejected: cleanup hook is implemented and wired (`pkg/engine/movement.go:310-324`; call sites in `cmd/client/handlers.go:1480`, `cmd/server/main.go:384`, `pkg/engine/system_init.go:362`). |
| HTTP response body leaks | Rejected: no production-path `http.Get`/`Do` body leak pattern found; hits were test paths with close handling. |
