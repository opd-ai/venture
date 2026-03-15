# Audit: github.com/opd-ai/venture/pkg/network/federation/guild
**Date**: 2026-02-25
**Auditor**: GitHub Copilot (META_AUDIT v2)
**Status**: Complete
<!--
Status criteria:
- Complete: All automated checks passed and fewer than 5 non-critical issues identified.
- Incomplete: Audit was stopped early or one or more required checks (e.g., go test, go vet, race) were not run.
- Needs Work: 5 or more issues identified, or any critical/priority-0 failure (e.g., panics, data corruption, security issues).
-->

## Summary
The `pkg/network/federation/guild` package provides cross-server guild management with federation synchronization. All automated checks passed (go vet, race detector, WASM compilation). Test coverage is 88.2%, exceeding the 40% target. The package follows ECS purity, uses deterministic procedural generation with seeds, and implements proper concurrency control with RWMutex. Only 4 minor issues found, all low severity.

## Automated Check Results
| Check | Result |
|---|---|
| `go vet` | ✅ Pass |
| `go test -cover` | 88.2% (target: 40%) |
| `go test -race` | ✅ Pass |
| WASM vet | ✅ Pass |
| TODO/FIXME count | 0 |
| Non-deterministic rand | 0 occurrences |
| Concrete net types | 0 occurrences |

## Issues Found

### High Severity
None found.

### Medium Severity
None found.

### Low Severity
- [x] **Documentation** — Package-level `doc.go` exists but individual files lack file-level comments explaining their purpose in the larger architecture (`constants.go:1`, `treasury.go:1`, `persistence.go:1`) — **ALREADY RESOLVED**: constants.go, treasury.go, and persistence.go all have file-level comments explaining their purpose
- [x] **Time Dependency** — `time_provider.go:23` uses `time.Now()` in production RealTimeProvider, which is non-deterministic. This is acceptable for timestamps but documented here for awareness. The package correctly provides MockTimeProvider for deterministic testing. — **ACCEPTABLE**: documented, MockTimeProvider provided for testing
- [x] **API Consistency** — `Manager.SetServerID()` method exists (`federation.go:39`) but is redundant with `WithServerID()` constructor option; prefer single initialization path via functional options to avoid runtime ID changes **COMPLETED 2026-02-27** - Removed SetServerID() method, updated all test usages to use WithServerID() functional option. Coverage: 93.5%

## Input Integration
| Input Source | Status | Notes |
|---|---|---|
| Keyboard | N/A | No input handling (server-side data layer) |
| Mouse | N/A | No input handling (server-side data layer) |
| Gamepad | N/A | No input handling (server-side data layer) |
| Touch | N/A | No input handling (server-side data layer) |
| VR | N/A | No input handling (server-side data layer) |
| Stub/Test | N/A | No input handling (server-side data layer) |

## Menu/UI Integration
| Menu | Reachable | Input-Complete | Backing System Wired | Notes |
|---|---|---|---|---|
| N/A | N/A | N/A | N/A | Package is server-side data layer with no UI components. Guild UI is in `pkg/engine/guild_ui.go` which uses this package as a backend. |

## Documentation Coverage
- Package `doc.go`: ✅ (comprehensive package overview with examples)
- Exported symbols documented: 44/44 (100%)
- Complex algorithms commented: ✅ (procedural generation, federation sync, permission validation all explained)

## Integration Status
This package bridges three major subsystems: network federation, guild management, and ECS world state. It serves as the authoritative source for cross-server guild data.

- System registration: ✅ — GuildSystem in `pkg/engine/guild_system.go` wraps the Manager; registered in `cmd/server/v8_systems.go` via `world.AddSystem(guildSystem)` (line ~50)
- Component registration: ✅ — GuildComponent defined in `pkg/engine/components.go`; Guild type implements `Type() string` method
- Serialize/Deserialize: ✅ — `Save()` and `Load()` methods in `persistence.go` use gzip-compressed JSON; decompression bomb protection with MaxGuildDataSize (50MB limit)
- Network sync: ✅ — Federation messages defined in `types.go` (GuildMessage, MemberJoinData, etc.); `SyncGuildState()` broadcasts via GuildTransport interface; message handlers registered in `NewManager()`
- Genre theming: ✅ — `GenerateIdentity()` in `identity.go` uses genre parameter to select name/emblem templates (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)
- Mod compatibility: N/A — Guild system operates at runtime data layer; no mod hooks needed (guild creation/management is data-driven via ECS)

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ✅ | Standard Go library usage, no platform-specific code |
| WASM | ✅ | `GOOS=js GOARCH=wasm go vet` passes; no syscall dependencies |
| Mobile | ✅ | No mobile-specific concerns; uses standard Go concurrency |

## Recommendations
1. **[LOW]** Add file-level doc comments to `constants.go`, `treasury.go`, and `persistence.go` explaining their role in the architecture
2. **[LOW]** Remove redundant `Manager.SetServerID()` method in favor of `WithServerID()` functional option to enforce single initialization path
3. **[LOW]** Add table-driven tests for `HandleGuildMessage()` covering malformed JSON, missing fields, and type mismatches
4. **[LOW]** Add benchmark for guild lookup performance with 1000+ guilds to validate O(1) map access scales

## Audit Details

### ECS Compliance
✅ **Pass** — Guild type is a pure data component with only `Type() string` method (`types.go:67-69`). All behavior is in Manager and standalone functions (`HasPermission()`, `GetMember()`). No logic methods on components.

### Deterministic Procedural Generation
✅ **Pass** — `GenerateIdentity()` uses `rand.New(rand.NewSource(seed))` for deterministic generation (`identity.go:26`). All randomness is seeded. Tests verify same seed produces same guild name and emblem colors.

### Network Interface Abstraction
✅ **Pass** — `GuildTransport` interface uses abstract `[]byte` for data (`federation.go:21-26`). No concrete `net.UDPConn` or `net.TCPConn` types. Fully mockable for testing.

### Error Handling
✅ **Pass** — All errors use structured logging via `logrus.WithFields()` with standard field names (`guild_id`, `player_id`, `server_id`). Error chains use `fmt.Errorf(..., %w, err)` for context preservation. No swallowed errors.

### Concurrency Safety
✅ **Pass** — Manager uses `sync.RWMutex` for all shared state access. Read operations use `RLock()`/`RUnlock()`, write operations use `Lock()`/`Unlock()`. Atomic operations for `guildCounter` (`manager.go:124`). Race detector passes.

### Structured Logging
✅ **Pass** — All logging uses `logrus.WithFields()` with standard field names. No `fmt.Println()` or `log.Println()` in production code. Logger initialized with `component: guild_manager` field.

### Resource Management
✅ **Pass** — Gzip reader/writer properly closed with defer (`persistence.go:56`, `persistence.go:39`). No goroutine leaks (no spawned goroutines). LimitReader protects against decompression bombs (`persistence.go:60`).

### API Design
✅ **Pass** — Constructor follows `NewManager(opts ...ManagerOption)` pattern. Functional options (`WithServerID`, `WithTimeProvider`) enable flexible configuration. System constructors log creation with structured fields.

### Integration Points

#### Server Registration
✅ Registered in `cmd/server/v8_systems.go` via `world.AddSystem(guildSystem)`. GuildSystem wraps Manager and handles ECS entity integration.

#### Client Registration
⚠️ Not registered in `cmd/client/` — Guild UI (`pkg/engine/guild_ui.go`) is present but registration in client systems not found. This is acceptable for server-authoritative guild management; client only needs UI to display server-provided guild data.

#### Federation Transport
✅ GuildTransport interface (`federation.go:21`) implemented by federation layer. SetTransport() allows runtime wiring. Example usage in `pkg/network/federation/transport_webrtc.go`.

#### Time Provider Abstraction
✅ TimeProvider interface enables deterministic timestamps. RealTimeProvider uses `time.Now()` in production; MockTimeProvider enables fixed-time testing. Constructor accepts `WithTimeProvider()` option.

### Cross-Server Synchronization Architecture

The package implements a message-passing federation protocol:

1. **Guild Creation** — Server A creates guild → broadcasts `MsgTypeGuildSync` with full state
2. **Member Join** — Server B adds member → broadcasts `MsgTypeMemberJoin` with delta
3. **State Merge** — All servers update local guild state from broadcast messages
4. **Conflict Resolution** — Last-write-wins via timestamp (Guild.UpdatedAt field)

Message handlers:
- `handleGuildSync()` — Full state replacement (line 135)
- `handleMemberJoin()` — Idempotent member addition (line 164)
- `handleMemberLeave()` — Idempotent member removal (line 217)
- `handleTerritoryChange()` — Reputation tracking (line 261)

### Performance Characteristics

From benchmark comments in `doc.go`:
- Guild creation: <0.1ms (1000x faster than 100ms target)
- Member add: 0.6µs (83,333x faster than 50ms target)
- Treasury operations: 0.2µs (50,000x faster than 10ms target)
- Save 100 guilds: 0.73ms (137x faster than 100ms target)

Map-based guild lookup is O(1). Thread-safe RWMutex allows concurrent reads.

### Test Quality

`manager_test.go` (1706 lines) provides comprehensive coverage:
- Table-driven tests for all manager operations
- Concurrent access tests (race detector enabled)
- Decompression bomb protection test
- Federation message handling tests
- Mock time provider tests for deterministic timestamps
- Permission validation tests
- Error path coverage

Example test structure:
```go
func TestManager_CreateGuild(t *testing.T) {
    tests := []struct {
        name    string
        genre   string
        seed    int64
        wantErr bool
    }{
        {"fantasy guild", "fantasy", 12345, false},
        {"sci-fi guild", "sci-fi", 67890, false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) { ... })
    }
}
```

### Security Considerations

1. **Decompression Bomb Protection** — `Load()` uses `io.LimitReader` to cap decompressed data at 50MB (`persistence.go:60`)
2. **Permission Enforcement** — All operations validate caller permissions via `HasPermission()` checks
3. **Input Validation** — Guild IDs, player IDs validated before operations; empty checks prevent nil panics
4. **No Arbitrary Code Execution** — All data structures are pure JSON; no executable code in guild state

### Dependencies

- `github.com/google/uuid` v1.6.0 — UUID generation for server IDs
- `github.com/sirupsen/logrus` v1.9.3 — Structured logging
- Standard library: `encoding/json`, `compress/gzip`, `sync`, `time`, `math/rand`

No Ebiten dependencies — pure server-side logic package.

### Future Enhancements

Per doc.go, the following are mentioned as "ready for integration":
- Guild bank (item storage) — types defined but no item storage logic implemented yet
- Territory control — reputation tracking exists but no territory capture mechanics
- Guild wars — PermissionDeclareWar defined but no war system implemented

These are out of scope for this package (require integration with `pkg/world/territory` and combat systems).
