# Memory Safety Gaps — 2026-04-21

> **Legacy gap ID note**: This memory-safety gap list (dated 2026-04-21) supersedes prior `GAPS.md` numbering. Existing references elsewhere in the repository to historical IDs such as `GAPS.md Gap 1` may refer to older versions. For this memory-focused document, use stable aliases `MS-1`, `MS-2`, and `MS-3`.
>
> **Scope note**: This file now tracks memory-specific gaps only. Broader non-memory backlog items remain in existing project trackers/docs (for example `ROADMAP.md` and package-scoped audit docs).

## Gap 1 (MS-1) — Priority P1 / Severity MEDIUM: Timer Allocation Churn in WebRTC Peer Loop
- **Stated Goal**: Maintain frame/network responsiveness under sustained multiplayer load with constrained memory/GC overhead (`docs/PERFORMANCE_TUNING.md:3,16-19`).
- **Current State**: `pkg/network/federation/webrtc/peer.go:149` uses `time.After(100 * time.Millisecond)` inside a long-running goroutine loop.
- **Risk**: Continuous timer allocations create unnecessary GC pressure as peer count grows.
- **Closing the Gap**: Replace loop-local `time.After` with one reusable `time.Ticker` (`NewTicker` + `Stop`), and validate via `go test -run '^$' -bench BenchmarkPeer_Connect -benchmem ./pkg/network/federation/webrtc`.

## Gap 2 (MS-2) — Priority P1 / Severity MEDIUM: Per-Message Timeout Timer Allocation on Success Path
- **Stated Goal**: High-throughput multiplayer messaging with stable memory behavior (network target and low per-packet overhead in `docs/PERFORMANCE.md:11,97-117`).
- **Current State**: signaling send/relay functions (`peer.go:182`; `signaling.go:154,179,204,228,371`) allocate timeout timers even when channel sends succeed immediately.
- **Risk**: Avoidable allocations on every message increase GC work and reduce throughput headroom in busy signaling scenarios.
- **Closing the Gap**: Refactor timeout handling to reuse timers or adopt non-allocating send/backpressure patterns; verify reduction with `BenchmarkPeer_Send`, `BenchmarkSignalingClient_SendOffer`, and `BenchmarkSignalingServer_RelayMessage` using `-benchmem`.

## Gap 3 (MS-3) — Priority P2 / Severity LOW: Cancellation-Path Timer Retention in Mobile Bandwidth Throttling
- **Stated Goal**: Efficient operation on constrained/mobile environments (`docs/PERFORMANCE_TUNING.md:3,51-59,158-165`).
- **Current State**: `pkg/network/federation/mobile/adapter.go:263` uses `time.After(waitTime)` within retry loop.
- **Risk**: Under frequent context cancellations, timers can outlive useful work until expiration, causing transient memory retention.
- **Closing the Gap**: Use `time.NewTimer(waitTime)` with explicit `Stop`/drain on cancellation; confirm lower allocation behavior in cancellation-heavy benchmark/test path.
