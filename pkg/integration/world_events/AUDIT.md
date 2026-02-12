# Audit: github.com/opd-ai/venture/pkg/integration/world_events
**Date**: 2026-02-12
**Status**: Needs Work

## Summary
World events integration package connects world state management with dynamic event systems. The package achieves 91.1% test coverage (exceeds 65% target) with well-structured table-driven tests. It demonstrates proper deterministic generation patterns and good error handling. However, it violates deterministic procgen standards through extensive use of `time.Now()` for event timing, lacks structured logging with logrus.WithFields, and has no error logging. The package is functionally complete with no stub implementations.

## Issues Found
- [x] **high** deterministic procgen — `manager.go:32,44,87,111,173,426` use `time.Now()` for event timing; should use seed-based time simulation or accept explicit currentTime parameters (`manager.go:32,44,87,111,173,426`)
- [x] **high** deterministic procgen — `events.go:217` `ShouldSpawnEvent` uses `time.Since(lastEventTime)` which depends on real-world clock; non-deterministic for replays (`events.go:217`)
- [x] **med** error handling — No structured logging with `logrus.WithFields` on error paths in `manager.go:54,58,77`; errors returned but not logged (`manager.go:54,58,77`)
- [x] **med** error handling — `manager.go:81` swallows error from `GenerateEvent` in error logging path (checks `err == nil` but doesn't log when err != nil) (`manager.go:81`)
- [x] **low** doc coverage — Missing godoc comments for exported types: `EventType`, `TriggerType`, `Severity`, `ImpactType` constants lack individual explanations (`types.go:11-51`)
- [x] **low** integration points — No component serialization/deserialization for persistence; events are runtime-only and cannot be saved/loaded (`types.go:54-134`)

## Test Coverage
91.1% (target: 65%) — ✅ Exceeds target

**Test Quality**: Excellent table-driven tests for all major functions. Tests cover:
- Event manager creation with custom config
- Event generation for all trigger types
- Validation edge cases
- Event chains and propagation
- Cleanup and expiration logic
- Concurrent access patterns (mutex safety)

## Integration Status
**How this package connects to engine, client, server**:
- ✅ Integrated with ECS via `pkg/engine/world_events_system.go` (WorldEventsSystem)
- ✅ Client integration in `cmd/client/handlers.go` (worldEventManager field) and `cmd/client/system_wrappers.go` (adapter wrapper)
- ✅ Engine system properly uses `world_events.EventManager` as a subsystem
- ⚠️ No server-side integration detected in `cmd/server/` (events appear client-side only)
- ⚠️ No component registration in ECS — events exist outside entity system (acceptable for integration layer)
- ⚠️ No save/load support — events don't persist across sessions

**Dependencies**: Only depends on `pkg/procgen` (SeedGenerator) and standard library. Clean dependency graph.

## Recommendations
1. **HIGH PRIORITY**: Replace all `time.Now()` calls with deterministic time simulation. Add `currentTime time.Time` parameter to EventManager constructor and Update method. Use `currentTime` plus seed-based delays instead of real clock.
2. **HIGH PRIORITY**: Refactor `ShouldSpawnEvent` to accept `currentTime time.Time` parameter instead of using `time.Since()`. This enables deterministic replays.
3. **MEDIUM PRIORITY**: Add structured logging with `logrus.WithFields` on all error paths. Include context fields: `event_id`, `trigger_type`, `severity`, `server_id`.
4. **MEDIUM PRIORITY**: Fix error swallowing in `world_events_system.go:81` — log errors when event generation fails.
5. **LOW PRIORITY**: Add godoc comments for all exported constant values (EventType, TriggerType, Severity, ImpactType).
6. **LOW PRIORITY**: Consider adding Serialize/Deserialize methods to WorldEvent, EventChain types for persistence support (enables event state to survive save/load).
