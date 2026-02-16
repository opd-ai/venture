# Audit: pkg/network/federation
**Date**: 2026-02-16
**Status**: Needs Work

## Summary
The federation package implements server-to-server communication with 35 Go files (20,554 LOC). Overall code quality is high with strong security (ed25519 auth), comprehensive error handling, and thread-safe design. Main package test coverage cannot be measured due to Ebiten test initialization failure; subdirectories have excellent coverage (80-87%). Primary concerns: Ebiten test dependency blocking main package tests, incomplete gossip/guild transmission stubs, and deterministic RNG usage in retry logic differs from project procgen standards.

## Issues Found
- [ ] **high** Test coverage — Main package tests fail with Ebiten GLFW initialization panic; cannot measure coverage (subdirs: 80-87% ✅) (`federation_test.go` imports cause UI init)
- [ ] **high** Stub/incomplete code — Gossip propagation builds message but doesn't transmit; documented as incomplete (`discovery.go:423-463`)
- [ ] **high** Stub/incomplete code — Guild federation broadcast builds message but doesn't send; documented as missing transport integration (`guild/federation.go:184`)
- [ ] **med** Deterministic procgen — `retry.go:63` uses `time.Now().UnixNano()` for RNG seed when config.Seed=0; violates project standard of deterministic generation (`retry.go:63`)
- [ ] **med** Network interfaces — `discovery.go:292` uses concrete `*net.UDPAddr` via `ResolveUDPAddr`, then assigns to `net.Addr` interface; acceptable but could be cleaner (`discovery.go:292`)
- [ ] **low** Integration points — Portal system update loop checks collision radius but doesn't trigger transfers; incomplete activation logic (`protocol.go:336-371`)
- [ ] **low** Error handling — `auth.go` uses crypto/rand for security tokens (correct usage; not procgen context) (`auth.go:196,216`)

## Test Coverage
**Main package: BLOCKED** (Ebiten dependency prevents test execution)  
**Subdirectories: 84.0% average** (target: 65%) ✅

Subdirectory breakdown:
- `guild/`: 87.2% ✅ (27 tests)
- `webrtc/`: 83.9% ✅ (23 tests)
- `mobile/`: 80.8% ✅ (18 tests)
- **Main package: FAIL** — Tests panic during init: `glfw: The GLFW library is not initialized`

**Root Cause**: Test files in main package import types that transitively depend on Ebiten UI initialization. The test suite cannot run in headless environments.

**Test Infrastructure**:
- 27 test files total (`*_test.go`)
- Subdirectories use build tags correctly and pass
- Main package needs Ebiten test mocking or refactoring to remove UI dependencies from testable code paths

**Coverage Gaps** (cannot verify due to test failure):
- `DiscoverySystem` gossip propagation edge cases
- `FederationProtocol` transfer workflow rollback scenarios
- `TransferManager` state machine transitions
- `FederatedMarket` price calculation boundary conditions
- `PortalSystem` activation and collision detection

## Integration Status
**Excellent integration surface** - connects to 9+ packages:

1. **Engine Integration**: Uses `engine.World`, `engine.Entity`, component types (Position, Health, Inventory, Portal)
2. **Recovery Integration**: All goroutines use `recovery.RecoverPanicWithLogger` for panic handling
3. **Version Integration**: Delegates protocol version to `version.ProtocolVersion` (centralized)
4. **Logging Integration**: Consistent `logrus.WithFields` structured logging throughout
5. **Subdirectory Integration**: 3 subdirs (`guild/`, `mobile/`, `webrtc/`) well-integrated

**Registration Status**:
- Discovery system requires manual `Start()` - not auto-registered
- Portal system requires explicit `Update()` call in game loop - not in system registry
- Guild federation manager has no auto-sync - requires external integration
- TransferManager callbacks must be configured (onTransferStart/Commit/Fail)

**Missing Serialization**:
- `ServerInfo`, `FederationState`, `PlayerTransfer` lack `Serialize()`/`Deserialize()` methods
- Hand-rolled JSON encoding via `json.Marshal` instead of component pattern

## Recommendations
1. **Fix Ebiten test dependency (CRITICAL)** — Refactor main package tests to avoid Ebiten UI imports; use test doubles/interfaces or move UI-dependent code to separate testable package
2. **Complete gossip transmission** — Integrate `FederationProtocol.SendGossip()` method with `DiscoverySystem.PropagateGossip()` to enable multi-hop server discovery (`discovery.go:423-463`)
3. **Complete guild broadcast** — Add TCP/TLS transport layer to `GuildFederationManager.BroadcastGuildUpdate()` for cross-server guild sync (`guild/federation.go:184`)
4. **Implement portal activation** — Complete `PortalSystem.Update()` collision detection to trigger `ActivatePortal()` on player proximity (`protocol.go:336-371`)
5. **Fix retry RNG non-determinism** — Change `retry.go:63` to require explicit seed parameter or document why network retry jitter doesn't need determinism (unlike procgen)
6. **Add structured logging** — Main package uses only 5 `logrus.WithFields` calls; add contextual logging to discovery, protocol, and transfer operations for better observability

## Architecture Strengths
- **Security**: ed25519 signatures, nonce-based replay prevention, fingerprint verification
- **Resilience**: Circuit breaker, retry with exponential backoff, connection pooling
- **Observability**: Comprehensive metrics (CircuitBreakerMetrics, MarketStats, FederationHealth)
- **Concurrency**: RWMutex throughout, goroutine-safe design, recovery wrappers
- **Interface-based networking**: Uses `net.Conn`, `net.PacketConn`, `net.Addr` (minor exception in discovery UDP resolution)

## ECS Compliance
**N/A** - This is network infrastructure, not game logic. No components or systems in ECS sense.

## Error Handling
**Excellent** - All errors checked, wrapped with context via `fmt.Errorf("%w")`, structured logging with `logrus.Fields` on error paths. No swallowed errors detected.
