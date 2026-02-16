# Engine Social Systems Sub-Audit

**Date**: 2026-02-16
**Scope**: Chat, Mail, Trade, Guild, and Faction systems in `pkg/engine/`
**Files Audited**:
- `chat_system.go`, `chat_trade_components.go`
- `enhanced_chat_system.go`
- `mail_system.go`
- `trade_system.go`
- `guild_system.go`, `guild_component.go`
- `faction_system.go`, `faction_component.go`

## Issues Found and Fixed

### HIGH Severity

1. **FactionComponent stored by value causes panic** (`faction_system.go:122`)
   - `playerEntity.AddComponent(*factionComp)` stored `FactionComponent` as a value type
   - All other code paths assert to `*FactionComponent` (pointer), causing panic on second reputation change
   - **Fix**: Changed to `playerEntity.AddComponent(factionComp)` to store as pointer
   - **Test**: Added `TestFactionSystem_ApplyReputationChange_PointerConsistency`

### MEDIUM Severity

2. **Unchecked type assertions in faction_system.go** (lines 94, 175, 207, 232)
   - Four type assertions used `comp.(*FactionComponent)` without comma-ok pattern
   - Would panic if component type was unexpected
   - **Fix**: Changed all four to use `comp.(*FactionComponent); ok` pattern with early return

3. **Unchecked type assertions in mail_system.go** (lines 92, 128, 168)
   - Three type assertions used `mailComp.(*MailComponent)` without comma-ok pattern
   - **Fix**: Changed all three to use comma-ok pattern with appropriate error returns

### LOW Severity

4. **fmt.Printf for error logging** (`enhanced_chat_system.go:189`)
   - Used `fmt.Printf` instead of structured logrus logging, inconsistent with codebase
   - **Fix**: Replaced with `log.WithFields(log.Fields{...}).WithError(err).Warn(...)`

5. **Non-deterministic message ID generation** (`chat_system.go:320`)
   - `generateMessageID()` uses `time.Now().UnixNano()` which could produce collisions
   - **Status**: Not fixed — low risk for chat message IDs, and `EnhancedChatSystem` already uses crypto/rand

## Summary

- **Total issues**: 5 (1 high, 2 medium, 2 low)
- **Fixed**: 4 (1 high, 2 medium, 1 low)
- **Remaining**: 1 low (non-deterministic message ID)
- **Tests**: All existing tests pass; 1 new test added for pointer consistency
