# Audit: github.com/opd-ai/venture/pkg/hostplay
**Date**: 2026-02-13
**Status**: Needs Work

## Summary
The hostplay package provides in-process server lifecycle management for host-and-play mode, enabling single-command local multiplayer. The package is well-architected with comprehensive test coverage (2431 test LOC vs 1211 source LOC), good documentation, and proper integration with the client. However, it contains 3 medium-severity issues (non-deterministic timestamps, network type assertion, stub implementation) and 3 low-severity issues (silent error handling) that should be addressed before production use.

## Issues Found
- [ ] **severity:med** deterministic-procgen — Uses `time.Now()` for timestamps in state broadcasting; should use deterministic time source or accept timestamp from caller (`state_broadcaster.go:86,97`)
- [ ] **severity:med** network-interfaces — Type assertion to concrete `*net.IPNet` in `GetLANAddress()`; should use interface methods only (`server_manager.go:637`)
- [ ] **severity:med** stub-code — `CreateDeltaSnapshot()` is incomplete stub returning full snapshot with TODO comment (`state_broadcaster.go:197-200`)
- [ ] **severity:low** error-handling — JSON marshal errors silently ignored in `addPositionComponent`, `addVelocityComponent`, `addHealthComponent`, `addRotationComponent` (`server_manager.go:481,500,519,537`)
- [ ] **severity:low** error-handling — JSON unmarshal error logged but data processing continues without validation (`input_handler.go:44-50`)
- [ ] **severity:low** test-coverage — Cannot measure coverage due to Ebiten GUI dependency; tests require display environment

## Test Coverage
Unable to measure (requires X11/display; tests fail with "DISPLAY environment variable missing")

**Estimated coverage**: High (2431 test LOC vs 1211 source LOC = 2:1 ratio)
- Unit tests: `host_and_play_test.go` (264 lines), `input_handler_test.go` (452 lines), `server_manager_test.go` (427 lines), `state_broadcaster_test.go` (440 lines)
- Lifecycle tests: `server_lifecycle_test.go` (576 lines)
- Integration tests: `integration_test.go` (272 lines, requires `go test -tags integration`)

**go vet**: ✅ PASS (no issues)

## Integration Status
**Fully integrated** with client (`cmd/client/util.go`):
- Used to start local server in host-and-play mode
- Configuration wired to CLI flags: `-server-port`, `-server-players`, `-host-lan`, `-server-tick`
- Proper lifecycle management with `defer manager.Stop()` pattern

**No system registration needed**: Package is standalone networking infrastructure, not an ECS system.

**No persistence needed**: Package manages ephemeral server state only; world state persisted via `pkg/world` and `pkg/saveload`.

## Recommendations
1. **[HIGH]** Replace `time.Now()` in `StateBroadcaster` with deterministic timestamp source (accept `timestamp` parameter in `CreateSnapshot()` or use monotonic counter)
2. **[MED]** Refactor `GetLANAddress()` to avoid `*net.IPNet` type assertion; use `net.Addr` interface methods or document exemption for local IP discovery
3. **[MED]** Implement `CreateDeltaSnapshot()` with delta compression or remove method if not planned for 1.0 release
4. **[LOW]** Add structured error logging for JSON marshal failures in `add*Component()` methods using `logrus.WithFields`
5. **[LOW]** Return early from `ProcessInputRaw()` when JSON unmarshal fails instead of calling `ProcessInput()` with empty data
