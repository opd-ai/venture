# Audit: github.com/opd-ai/venture/pkg/network/federation/mobile
**Date**: 2026-02-13
**Status**: Needs Work

## Summary
Mobile federation package provides battery-aware federation for phones/tablets acting as servers. The package is well-designed with 80.8% test coverage and uses appropriate concurrency patterns. However, it has several violations of project standards: non-deterministic use of `time.Now()` in network code, missing error logging, and a lack of integration with the broader federation system beyond the stub implementation in MobileFederationSystem.

## Issues Found
- [x] **high** **deterministic procgen** — Non-deterministic time usage in network code: `time.Now()` used for bandwidth token bucket timing, should use monotonic clock or injected time provider for testability (`adapter.go:212`)
- [x] **high** **deterministic procgen** — Non-deterministic task ID generation uses `time.Now().Unix()` making IDs unpredictable and untestable (`adapter.go:285`)
- [x] **high** **error handling** — Error from sync handler not logged with structured logging in `performSync()` — errors silently recorded to state without diagnostic info (`adapter.go:185-187`)
- [x] **med** **error handling** — Missing structured logging when bandwidth limit exceeded in `executeSyncWithBandwidthLimit` — should log wait time, tokens needed, etc. (`adapter.go:226-239`)
- [x] **med** **stub/incomplete** — Simulated byte counts in `RecordSyncSuccess()` calls instead of actual transferred bytes (`adapter.go:191,323`)
- [x] **med** **stub/incomplete** — Fixed estimated bytes (10KB) instead of actual measurement in bandwidth limiter (`adapter.go:208`)
- [x] **med** **integration points** — Not registered in any system initialization despite MobileFederationSystem existing — client/server need explicit initialization code
- [x] **low** **error handling** — `ScheduleBackgroundTask()` error handling could be improved with structured logging when background sync disabled (`adapter.go:265-267`)
- [x] **low** **deterministic procgen** — `RecordSyncSuccess()` uses `time.Now()` for LastSyncTime, acceptable for network timing but should be documented as exempt from seed-based determinism (`types.go:177`)

## Test Coverage
80.8% (target: 65%) ✅ — Exceeds target

**Coverage breakdown:**
- `adapter.go`: Well-covered with table-driven tests for all major paths
- `capabilities.go`: Platform detection covered across matrix
- `types.go`: State management thread-safety covered
- Integration tests demonstrate graceful degradation

## Integration Status
Package is properly integrated with engine layer via `MobileFederationSystem` (`pkg/engine/mobile_federation_system.go`). The system is imported by:
- `pkg/engine/mobile_federation_system.go` — ECS wrapper with Start/Stop/Update
- `cmd/client/handlers.go` — Client-side integration point

**Missing registrations:**
- Not auto-initialized in any system_init.go — requires manual instantiation
- MobileFederationSystem.syncHandler is a placeholder (line 47-56 in mobile_federation_system.go) marked as "For now, this is a placeholder that returns success"
- No server-side mobile federation support beyond client integration

**Serialization:** N/A — Mobile federation state is transient, no persistence needed for sync state

## Recommendations
1. **HIGH PRIORITY**: Replace `time.Now()` with monotonic clock (`time.Since()` with stored start time) or inject time provider interface for bandwidth token bucket to enable deterministic testing
2. **HIGH PRIORITY**: Add structured logging with `logrus.WithFields()` on all error paths in performSync and executeSyncWithBandwidthLimit
3. **HIGH PRIORITY**: Document exemption from seed-based determinism for network timing operations in package doc — network sync inherently uses wall-clock time
4. **MEDIUM**: Implement actual byte counting from sync operations instead of hardcoded estimates (requires federation protocol changes)
5. **MEDIUM**: Add system registration example in docs/MOBILE.md showing how to initialize MobileFederationSystem in client/server
6. **LOW**: Replace `time.Now().Unix()` task ID with UUID or sequential counter for better testability
