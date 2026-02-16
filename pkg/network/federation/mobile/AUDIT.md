# Audit: github.com/opd-ai/venture/pkg/network/federation/mobile
**Date**: 2026-02-16
**Status**: Needs Work

## Summary
Mobile federation package provides battery-aware federation support for mobile devices (iOS/Android). Overall architecture is solid with good thread-safety and comprehensive testing (80.8% coverage), but has 5 high-severity issues: non-deterministic time.Now() usage throughout, incomplete federation sync implementation in engine integration, and lack of TimeProvider abstraction. Critical for cross-platform multiplayer experience.

## Issues Found
- [x] **severity:high** Stub/incomplete code — Placeholder sync handler in engine integration never syncs actual player/guild data (`pkg/engine/mobile_federation_system.go:47-57`)
- [x] **severity:high** Deterministic procgen — Uses `time.Now()` for token bucket bandwidth limiting instead of injected time provider (`adapter.go:212`)
- [x] **severity:high** Deterministic procgen — Uses `time.Now().Unix()` for BackgroundTask ID generation instead of deterministic seed (`adapter.go:285`)
- [x] **severity:high** Deterministic procgen — Uses `time.Now()` for BackgroundTask scheduling instead of injected time provider (`adapter.go:286`, `adapter.go:307`)
- [x] **severity:high** Deterministic procgen — RecordSyncSuccess uses `time.Now()` for LastSyncTime tracking instead of injected time provider (`types.go:177`)
- [x] **severity:med** Error handling — Bandwidth limit wait loop uses recursive call that could stack overflow on long waits; should use iterative loop (`adapter.go:236`)
- [x] **severity:low** Integration points — Not registered in cmd/server (only cmd/client); mobile devices cannot act as federated servers as documented (`cmd/client/handlers.go:1032`)

## Test Coverage
80.8% (target: 65%) ✓

**Coverage exceeds target**, but determinism violations prevent reliable testing of time-dependent features (battery modes, sync intervals, background tasks).

## Integration Status
**Client Integration**: ✓ Fully integrated via `cmd/client/handlers.go` with MobileFederationSystem in engine layer.

**Server Integration**: ✗ **Not integrated** with `cmd/server` despite documentation claiming mobile devices can act as federated servers.

**Engine System**: Present (`pkg/engine/mobile_federation_system.go`) but sync handler is a **placeholder** (lines 47-57) that never performs actual federation sync (player data, guild info, world state).

**Persistence**: No serialize/deserialize support for State (not an ECS component, so acceptable for runtime-only state).

**Dependencies**: 
- Imports `logrus` for structured logging ✓
- No concrete network types (net.UDPAddr, etc.) ✓
- Platform detection uses `runtime.GOOS` for capability detection ✓

## Recommendations
1. **HIGH PRIORITY**: Replace all `time.Now()` calls with TimeProvider abstraction pattern (see `pkg/companion/learning` for example: `TimeProvider` interface with `RealTimeProvider` and `FakeTimeProvider`)
2. **HIGH PRIORITY**: Implement actual federation sync logic in `mobile_federation_system.go` sync handler (currently placeholder that returns success after 10ms)
3. **MEDIUM PRIORITY**: Refactor `executeSyncWithBandwidthLimit` to use iterative loop instead of recursion for wait/retry logic (prevents potential stack overflow)
4. **LOW PRIORITY**: Add mobile federation support to `cmd/server` if mobile-as-server is a supported use case (or update documentation to clarify client-only scope)
5. **ENHANCEMENT**: Add serialize/deserialize for BackgroundTask for persistence across app restarts (iOS/Android lifecycle management)

## Architecture Strengths
- ✓ Thread-safe design with comprehensive `sync.RWMutex` usage
- ✓ Battery-aware optimization with three-tier mode system (Normal/Low/Critical)
- ✓ Graceful degradation with WebRTC capability detection and WebSocket/HTTP fallback
- ✓ Token bucket bandwidth limiting algorithm
- ✓ Platform-specific capability detection (iOS/Android/WASM vs desktop)
- ✓ Comprehensive test coverage with table-driven tests
- ✓ Excellent godoc documentation with usage examples
- ✓ No ECS violations (infrastructure package, not game logic)
- ✓ Structured logging with logrus.Fields
- ✓ Context-based cancellation and timeout handling

## ECS Compliance
N/A - This is a network infrastructure package, not part of the ECS game logic layer. No components defined. The engine wrapper (`MobileFederationSystem`) correctly implements the System interface with an empty `Update()` method since all operations are async.
