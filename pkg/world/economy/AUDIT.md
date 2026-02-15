# Audit: github.com/opd-ai/venture/pkg/world/economy
**Date**: 2026-02-15
**Status**: Complete

## Summary
The economy package implements cross-server marketplace and guild banking systems. Overall health is **strong** with 88.5% test coverage (exceeding 65% target), passing `go vet`, comprehensive godoc, and clean architecture. One **critical** non-deterministic issue found in time.Now() usage. Integration with engine complete via pkg/engine/economy_system.go, but marketplace/guild bank operations need active guild vault tracking.

## Issues Found
- [x] **high** deterministic-procgen — Extensive use of `time.Now()` for timestamps breaks determinism requirement (multiple locations)
  - `marketplace.go:66` - listing creation timestamp
  - `marketplace.go:69` - listing expiration timestamp (7-day default)
  - `marketplace.go:307` - transaction timestamp
  - `guild_bank.go:117,177,178,213,238,239,273,292,332,341,342,385,404,502` - 15+ instances for audit entries, withdrawals, interest
- [x] **med** error-handling — Incomplete implementation in `PurchaseItem()` silently ignores gold transfer (`marketplace.go:328`)
- [x] **low** integration — EconomySystem integration exists in `pkg/engine/economy_system.go` but `calculateInterest()` and `updateMarketplace()` are stubs (`pkg/engine/economy_system.go:58-69`)
- [x] **low** integration — No automatic interest calculation: `calculateInterest()` logs but doesn't call `guildBank.CalculateInterest()` for any guilds (`pkg/engine/economy_system.go:65-69`)
- [x] **low** logging — Zero structured logging in economy package files (no `logrus.WithFields` usage for errors/operations)

## Test Coverage
88.5% (target: 65%) ✅

**File-by-file breakdown**:
- `marketplace.go`: Well-tested (search, purchase, expiration)
- `guild_bank.go`: Comprehensive (deposits, withdrawals, interest, limits, audit log)
- `pricing_engine.go`: Complete (trend tracking, trade impact)
- `system.go`: Basic coverage (system wrapper)
- `types.go`: Full coverage (utility methods)

## Integration Status
**Engine Integration**: ✅ Complete via `pkg/engine/economy_system.go`
- `EconomySystem` wraps `FederatedMarketplace` and `GuildBankManager`
- Registered with World via `Update([]*Entity, float64)` pattern
- Direct accessors: `GetMarketplace()`, `GetGuildBank()`

**Client/Server Integration**: ✅ Active imports
- `cmd/client/handlers.go` — imports economy package
- `cmd/client/system_wrappers.go` — imports economy package

**Missing Integrations**:
- No automatic interest calculation loop in `EconomySystem.calculateInterest()` (stub)
- No guild vault iteration for cross-server sync
- No marketplace price trend updates in `EconomySystem.updateMarketplace()` (stub)
- No persistence hooks in system wrapper (Save/Load not called)

## Recommendations
1. **HIGH PRIORITY**: Refactor time.Now() usage to support deterministic replay/testing
   - Add `TimeProvider` interface with `Now()` method (prod: real time, test: fixed/seed-based)
   - Inject `TimeProvider` into `FederatedMarketplace` and `GuildBankManager` constructors
   - Example: `type TimeProvider interface { Now() time.Time }`
   - Alternative: Accept `currentTime time.Time` parameters for timestamp-critical methods

2. **MEDIUM PRIORITY**: Complete `marketplace.go:323-328` implementation
   - Integrate with player inventory/gold systems for actual item/gold transfer
   - Add `OnPurchase` callback or return `*Transaction` for engine to process
   - Document interim behavior in godoc if deferred to integration layer

3. **MEDIUM PRIORITY**: Add structured logging with `logrus.WithFields`
   - Log marketplace operations: `listingID`, `sellerID`, `buyerID`, `price`, `quantity`
   - Log guild bank operations: `guildID`, `memberID`, `action`, `amount`, `balanceBefore`, `balanceAfter`
   - Log errors with contextual fields for debugging federation/sync issues

4. **LOW PRIORITY**: Implement active guild vault iteration in `pkg/engine/economy_system.go:65-69`
   - Maintain registry of active guild vaults (map or entity components)
   - Call `guildBank.CalculateInterest(guildID)` for each vault during daily cycle
   - Add error handling/logging for interest calculation failures

5. **LOW PRIORITY**: Implement marketplace price updates in `pkg/engine/economy_system.go:58-62`
   - Call `marketplace.CleanupExpiredListings()` during update cycle
   - Add price trend recalculation or decay logic
   - Consider marketplace statistics logging (listings, transactions, capacity)
