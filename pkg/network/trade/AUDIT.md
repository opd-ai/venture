# Audit: pkg/network/trade
**Date**: 2026-02-15
**Status**: Complete

## Summary
Trade system implements two-phase commit protocol for multiplayer item trading with validation and rate limiting. High code quality with 75.4% test coverage, but has critical integration issues: duplicate implementation in `pkg/engine/trade_system.go` (608 LOC) creates unclear responsibility boundaries; non-deterministic time usage violates procgen standards; missing trust/rarity validation tests; no structured logging; TradeComponent lacks serialize/deserialize for persistence.

## Issues Found
- [x] **high** architecture — Duplicate TradeSystem implementation in `pkg/engine/trade_system.go` (608 LOC) vs `pkg/network/trade/system.go` (783 LOC) creates unclear ownership; server uses `engine.TradeSystem`, client uses both (`handlers.go:19,25`)
- [x] **high** deterministic-procgen — Non-deterministic time usage via `time.Now()` in `system.go:92,234,740` violates deterministic generation requirement (should accept injectable time source)
- [x] **med** test-coverage — Missing trust score validation tests: no tests for low-trust (<0.3) item limits, high-trust (>0.8) legendary items, or mid-trust rarity restrictions (`system_test.go`: all tests use `RarityCommon`)
- [x] **med** error-handling — No structured logging with `logrus.WithFields` for error paths; uses only `fmt.Errorf` without contextual logging (`system.go`: no logrus imports)
- [x] **med** integration — TradeComponent lacks `Serialize()`/`Deserialize()` methods for persistence support (`pkg/engine/chat_trade_components.go:294-304`)
- [x] **low** test-coverage — Missing timeout validation tests: no tests verify 30-second proposal timeout or proximity-based auto-cancel (`system_test.go`: Update() tests don't verify timeout logic)
- [x] **low** test-coverage — Missing inventory space validation tests: no tests for weight/slot limits edge cases (`system_test.go`: no weight capacity tests)
- [x] **low** test-coverage — Missing rollback scenario tests: no tests verify atomic rollback on partial transfer failure (`system_test.go`: no rollback tests)
- [x] **low** doc-coverage — Missing godoc for helper methods `resolveItems`, `validateTrust`, `validateTradability`, `validateInventorySpace` (`system.go:629-722`)

## Test Coverage
75.4% (target: 65%) — **PASS** but missing critical edge cases (trust limits, timeouts, rollback, inventory constraints)

## Integration Status

**Current State**:
- Server: Uses `engine.NewTradeSystem` (`cmd/server/v4_systems.go`)
- Client: Uses BOTH `engine.TradeSystem` AND `network/trade.TradeSystem` (`cmd/client/handlers.go:19,25`)
- Network layer: Has `TradeProposalPacket` serialization (`pkg/network/packets.go:89,167,201`)
- Components: `TradeComponent` and `TradeProposal` defined in `pkg/engine/chat_trade_components.go:269,294`

**Missing Integrations**:
- No registration in network packet handlers for `PacketTypeTradeProposal` (type 53)
- TradeComponent missing serialize/deserialize for save/load persistence
- No system registration in `pkg/engine/system_init.go` (likely handled manually in cmd/server)
- Unclear delineation: `pkg/network/trade` appears to be legacy/unused in favor of `pkg/engine/trade_system.go`

**Architecture Conflict**:
The dual implementation suggests refactoring need: either consolidate into single canonical implementation or clarify that `network/trade` is network-layer protocol handler while `engine/trade_system.go` is authoritative game logic.

## Recommendations
1. **HIGH PRIORITY**: Clarify/consolidate duplicate TradeSystem implementations — choose canonical location and deprecate/remove other, OR document clear separation of concerns (network protocol vs game logic)
2. **HIGH PRIORITY**: Make time usage deterministic — accept injectable `TimeSource` interface instead of `time.Now()` to enable deterministic testing/replay
3. **MEDIUM PRIORITY**: Add trust validation tests — cover low-trust item limits (5 items max, common/uncommon only), high-trust legendary trading, mid-trust epic restrictions
4. **MEDIUM PRIORITY**: Add structured logging with `logrus.WithFields` — log trade events with fields: `proposerID`, `recipientID`, `status`, `failure_reason`, `trust_score`
5. **MEDIUM PRIORITY**: Implement TradeComponent serialization — add `Serialize()`/`Deserialize()` for save/load support
6. **LOW PRIORITY**: Add edge case tests — timeout validation, proximity auto-cancel, inventory weight/slot limits, rollback on partial failure
