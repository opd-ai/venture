# Audit: github.com/opd-ai/venture/pkg/integration/political_warfare
**Date**: 2026-02-12
**Status**: Needs Work

## Summary
Integration package for guild-level political warfare mechanics (war declarations, peace treaties, embargoes, diplomatic victories). The package is well-structured with 96.3% test coverage and comprehensive functionality. However, it has several integration gaps: lacks persistence methods for manager state, doesn't use world seed for deterministic RNG, and has incomplete reputation penalty implementation that only records but doesn't apply penalties to faction system.

## Issues Found
- [ ] **med** Integration — Manager not initialized with world seed; uses DefaultSeed instead of world-specific seed for deterministic RNG (`system.go:21`, compare to `cmd/server/v9_systems.go:61` which passes seed to narrativeworld.NewSystem)
- [ ] **med** Integration — No Save/Load/Serialize methods for Manager persistence; wars, treaties, embargoes would be lost on server restart (`manager.go:14-26`)
- [ ] **med** Incomplete — Reputation penalty only records but doesn't apply to faction system; see comment "In a full implementation, this would interact with the faction system" (`manager.go:372-374`)
- [ ] **med** Error handling — Gold concession silently swallows error if attacker guild not found; gold deducted from defender but not added to attacker on error (`manager.go:453`)
- [ ] **low** Error handling — No structured logging for errors in Manager; System has logger but doesn't use it in Update() (`system.go:13,27-29`)
- [ ] **low** Doc coverage — Manager methods lack logrus.WithFields logging on error paths for debugging

## Test Coverage
96.3% (target: 65%) ✅

## Integration Status
**Registered Systems:**
- ✅ Server: `cmd/server/v9_systems.go:67` - Creates system, adds to world
- ✅ Client: `cmd/client/handlers.go:1423` - Creates system for client-side state

**Integration Points:**
- **Dependencies**: `pkg/engine` (World, Entity), `pkg/network/federation/guild` (Guild Manager)
- **Data Flow**: Manager state accessed via `System.GetManager()` for direct API calls from network handlers
- **Missing**: 
  - No world seed propagation from NewSystem to NewManagerWithSeed
  - No persistence integration with `pkg/saveload`
  - No faction system integration for reputation penalties (manager.go:373-374)
  - Territory transfer integration incomplete (only tracks pending transfers, no actual transfer logic)
  - Item tribute integration incomplete (only tracks pending tributes, no inventory transfer)
  - Apology broadcast integration incomplete (only tracks pending apologies, no messaging system integration)

**Thread Safety:** ✅ All public Manager methods use sync.RWMutex

**Deterministic RNG:** ⚠️ Manager uses seed-based RNG correctly (`manager.go:48`), but System doesn't pass world seed to Manager

**Time Usage:** ✅ Uses `time.Now()` for runtime event timestamps (war declarations, treaty signing, expiration checks) - appropriate for real-time gameplay events, not procedural generation

## Recommendations
1. **HIGH PRIORITY**: Modify `NewSystem` to accept seed parameter and pass to `NewManagerWithSeed(world, guildManager, seed)` for deterministic RNG tied to world generation
2. **HIGH PRIORITY**: Add Manager persistence methods (`SaveState() []byte`, `LoadState([]byte) error`) and integrate with saveload system
3. **MEDIUM**: Complete reputation penalty implementation by integrating with faction system (likely `pkg/engine/faction_system.go` or similar)
4. **MEDIUM**: Add error logging in `applyGoldConcession` when attacker guild not found; consider returning error or using transaction pattern
5. **LOW**: Add structured logging with logrus.WithFields in Manager error paths and System.Update() for observability
6. **LOW**: Add integration helpers for pending concessions: methods to clear processed concessions after territory/item/apology systems handle them
