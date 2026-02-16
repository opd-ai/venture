# Audit: github.com/opd-ai/venture/pkg/network/federation/guild
**Date**: 2026-02-16
**Status**: Needs Work

## Summary
Cross-server guild management package with 87.3% test coverage. Implements core guild operations (create, member management, treasury, federation sync) with thread-safe design and deterministic procedural identity generation. Has 2 high-severity ECS violations (logic methods on Guild component) and several medium-priority issues with time.Now() usage violating deterministic generation principles.

## Issues Found
- [x] **high** ECS compliance — Guild component has logic methods `HasPermission()` and `GetMember()` - violates ECS purity, should be system-level functions (`types.go:75`, `types.go:92`)
- [x] **high** ECS compliance — Guild component implements `Type() string` correctly but also has behavior methods - all logic must move to Manager or helper functions (`types.go:67-99`)
- [x] **med** Deterministic procgen — Uses `time.Now()` for timestamps in non-procgen context (member joins, treasury txns, guild updates) - acceptable for metadata but inconsistent with strict determinism (`manager.go:102`, `manager.go:107-108`, `manager.go:165-166`, `manager.go:168`, `manager.go:198`, `manager.go:282`, `manager.go:328`, `federation.go:90`, `treasury.go:35`, `treasury.go:66`)
- [x] **med** Integration status — GuildTransport interface defined but no concrete implementation provided - transport layer stubbed (`federation.go:22-25`)
- [x] **med** Integration status — Manager.transport field optional (nil check at federation.go:94) - federation sync messages prepared but not transmitted without transport wiring
- [x] **low** Error handling — Manager uses uuid.New() for serverID generation which is non-deterministic - should accept serverID in constructor or use seeded generation (`manager.go:56`)
- [x] **low** Documentation — Missing package-level doc comment explaining integration with pkg/network/federation parent package (`doc.go:1`)
- [x] **low** Thread safety — Manager.guildCounter incremented without atomic operations - potential race condition in concurrent CreateGuild calls (`manager.go:78`)

## Test Coverage
87.3% (target: 65%) ✅

**Strengths**:
- Comprehensive table-driven tests for all operations (create, members, treasury, federation)
- Concurrent access tests (manager_test.go demonstrates thread-safety validation)
- Federation message handler tests with type conversions
- Decompression bomb protection tests
- Deterministic identity generation tests

**Gaps**:
- No tests for transport.BroadcastGuildUpdate integration
- Missing benchmark tests for high-frequency operations (member lookups, permission checks)
- No tests for guildCounter race conditions

## Integration Status
**Fully Integrated** ✅

The package is integrated with:
1. **Engine**: `pkg/engine/guild_system.go` uses guild.Manager, `pkg/engine/guild_ui.go` renders guild interfaces
2. **Client**: `cmd/client/handlers.go` instantiates guild.NewManager() for client-side guild management
3. **Integration**: `pkg/integration/political_warfare/` uses guild.Manager for political warfare features
4. **Server**: `cmd/server/v8_systems.go` returns guild.Manager for V9 integration

**Missing Integrations**:
- No concrete GuildTransport implementation wired up (federation broadcast stubbed)
- pkg/network/federation parent package does not yet invoke guild sync protocol
- No server-side persistence hooks (Save/Load not called from cmd/server)

**Registration**: Not applicable - package provides Manager as injectable dependency rather than auto-registered system

## Recommendations
1. **HIGH PRIORITY**: Refactor Guild component to remove `HasPermission()` and `GetMember()` methods - move to Manager helper functions or standalone package-level functions to achieve ECS purity
2. **HIGH PRIORITY**: Extract behavioral methods from Guild component - create `guildHelpers.go` with `HasPermission(guild *Guild, rank Rank, perm Permission) bool` and `GetMember(guild *Guild, playerID string) *Member`
3. **MEDIUM PRIORITY**: Implement concrete GuildTransport adapter that wraps pkg/network/federation protocol - wire up actual federation broadcast in Manager.SyncGuildState
4. **MEDIUM PRIORITY**: Make Manager.guildCounter atomic using sync/atomic.AddInt64 to prevent race conditions
5. **LOW PRIORITY**: Accept serverID as NewManager parameter instead of generating via uuid.New() for deterministic testing
6. **LOW PRIORITY**: Document integration with parent pkg/network/federation package in doc.go header
7. **LOW PRIORITY**: Add benchmark tests for permission checks and member lookups (hot path operations)
8. **LOW PRIORITY**: Consider TimeProvider abstraction (like pkg/companion/learning) to make timestamps deterministic in tests while preserving real-time behavior in production
