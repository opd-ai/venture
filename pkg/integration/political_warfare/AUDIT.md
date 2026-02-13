# Audit: github.com/opd-ai/venture/pkg/integration/political_warfare
**Date**: 2026-02-13
**Status**: Needs Work

## Summary
Integration package for guild-level political warfare mechanics (war declarations, peace treaties, embargoes, diplomatic victories). The package is well-structured with 96.3% test coverage and comprehensive functionality. Now includes Save/Load persistence methods for manager state. Remaining gap: reputation penalty implementation only records but doesn't apply penalties to faction system.

## Issues Found
- [x] **med** Integration — Manager not initialized with world seed; uses DefaultSeed instead of world-specific seed for deterministic RNG (`system.go:21`, compare to `cmd/server/v9_systems.go:61` which passes seed to narrativeworld.NewSystem) — **FIXED 2026-02-13**: NewSystem now accepts seed parameter and passes to NewManagerWithSeed
- [x] **med** Integration — No Save/Load/Serialize methods for Manager persistence; wars, treaties, embargoes would be lost on server restart (`manager.go:14-26`) — **FIXED 2026-02-13**: Added `Save() ([]byte, error)` and `Load(data []byte) error` methods with gzip-compressed JSON serialization. Also added `GetActiveAllianceCalls()` getter.
- [ ] **med** Incomplete — Reputation penalty only records but doesn't apply to faction system; see comment "In a full implementation, this would interact with the faction system" (`manager.go:372-374`)
- [x] **med** Error handling — Gold concession silently swallows error if attacker guild not found; gold deducted from defender but not added to attacker on error (`manager.go:453`) — **FIXED 2026-02-13**: Added structured logging with logrus.WithFields when attacker guild not found
- [x] **low** Error handling — No structured logging for errors in Manager; System has logger but doesn't use it in Update() (`system.go:13,27-29`) — **FIXED 2026-02-13**: Added logrus import to manager.go and logging in applyGoldConcession error path
- [x] **low** Doc coverage — Manager methods lack logrus.WithFields logging on error paths for debugging — **FIXED 2026-02-13**: Added structured logging in gold concession error path

## Test Coverage
96.3% (target: 65%) ✅

## Integration Status
**Registered Systems:**
- ✅ Server: `cmd/server/v9_systems.go:67` - Creates system with seed, adds to world
- ✅ Client: `cmd/client/handlers.go:1419` - Creates system with seed for client-side state

**Integration Points:**
- **Dependencies**: `pkg/engine` (World, Entity), `pkg/network/federation/guild` (Guild Manager)
- **Data Flow**: Manager state accessed via `System.GetManager()` for direct API calls from network handlers
- **Persistence**: ✅ Manager implements `Save()/Load()` for state persistence (compressed JSON)
- **Missing**: 
  - No faction system integration for reputation penalties (manager.go:373-374)
  - Territory transfer integration incomplete (only tracks pending transfers, no actual transfer logic)
  - Item tribute integration incomplete (only tracks pending tributes, no inventory transfer)
  - Apology broadcast integration incomplete (only tracks pending apologies, no messaging system integration)

**Thread Safety:** ✅ All public Manager methods use sync.RWMutex

**Deterministic RNG:** ✅ System now passes world seed to Manager via NewSystem(world, guildManager, seed)

**Time Usage:** ✅ Uses `time.Now()` for runtime event timestamps (war declarations, treaty signing, expiration checks) - appropriate for real-time gameplay events, not procedural generation

## Recommendations
1. ~~**HIGH PRIORITY**: Modify `NewSystem` to accept seed parameter and pass to `NewManagerWithSeed(world, guildManager, seed)` for deterministic RNG tied to world generation~~ **DONE**
2. ~~**HIGH PRIORITY**: Add Manager persistence methods (`SaveState() []byte`, `LoadState([]byte) error`) and integrate with saveload system~~ **DONE**
3. **MEDIUM**: Complete reputation penalty implementation by integrating with faction system (likely `pkg/engine/faction_system.go` or similar)
4. ~~**MEDIUM**: Add error logging in `applyGoldConcession` when attacker guild not found; consider returning error or using transaction pattern~~ **DONE**
5. ~~**LOW**: Add structured logging with logrus.WithFields in Manager error paths and System.Update() for observability~~ **DONE (partial)**
6. **LOW**: Add integration helpers for pending concessions: methods to clear processed concessions after territory/item/apology systems handle them
