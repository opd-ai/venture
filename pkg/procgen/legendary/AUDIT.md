# Audit: github.com/opd-ai/venture/pkg/procgen/legendary
**Date**: 2026-02-13
**Status**: Needs Work

## Summary
The legendary quest generation package implements multi-phase legendary quests with cross-server requirements, raid integration, and unique rewards. The package has excellent test coverage (88.1%), comprehensive documentation, and well-structured code. However, it violates deterministic generation standards by using `time.Now()` for progress tracking, lacks structured logging throughout, and has missing error handling in several critical paths.

## Issues Found
- [ ] **high** deterministic-procgen — Manager uses `time.Now()` for timestamps in progress tracking, violating deterministic generation requirement. All quest state timestamps (StartedAt, LastUpdated, CompletedAt) use system time instead of deterministic time sources. (`manager.go:125`, `manager.go:527`, `types.go:511-512`, `types.go:519`, `types.go:525-528`, `types.go:549-550`, `types.go:561`, `types.go:581-582`, `types.go:593`, `types.go:613-614`, `types.go:620`)
- [ ] **med** error-handling — No structured logging with `logrus.WithFields` on error paths in manager functions. Silent failures in `ValidateServerVisit`, `ValidateRaidCompletion`, `ValidateCraftingCompletion` don't log context (playerID, questID, serverID, etc.). (`manager.go:134-169`, `manager.go:172-203`, `manager.go:206-249`)
- [ ] **med** error-handling — `grantRewards` silently skips already-claimed rewards without logging or notifying caller. Players may not receive expected rewards with no indication why. (`manager.go:432-437`, `manager.go:442-445`)
- [ ] **low** doc-coverage — Exported function `NewLegendaryQuestGenerator` lacks godoc comment explaining purpose and usage. (`generator.go:17`)
- [ ] **low** doc-coverage — Exported function `NewQuestManager` lacks comprehensive godoc explaining the raidMgr parameter requirement. (`manager.go:52`)
- [ ] **low** integration — Package has no verification that generated quests are registered/persisted beyond manager's internal state. Consider integration with quest tracking systems documented in manager. (`manager.go:81-102`)

## Test Coverage
88.1% (target: 65%) ✅

## Integration Status
Package integrates with:
- **pkg/world/raids**: Manager requires raids.Manager for raid validation integration (Phase 59.1)
- **pkg/procgen**: Implements procgen.Generator interface for consistent generation patterns
- Quest persistence via Save/Load methods serializing to JSON
- Progress tracking via ProgressTracker maintaining quest/player state
- Cross-server validation via ServerValidator tracking federated servers

**Missing integrations:**
- No registration with global quest registry or quest tracking systems
- No event emission for quest state changes (started, phase complete, quest complete)
- No integration with reward granting systems beyond internal catalog
- Progress tracker timestamps use `time.Now()` which breaks determinism for replay/testing

## Recommendations
1. **[HIGH PRIORITY]** Replace all `time.Now()` calls with deterministic time sources. Add a `Clock` interface parameter to QuestManager/ProgressTracker allowing injectable time for deterministic generation and testing. Use seed-derived timestamps or monotonic counters for quest progression.
2. **[MEDIUM PRIORITY]** Add structured logging with `logrus.WithFields` to all validation methods. Log errors with context: `logger.WithFields(logrus.Fields{"playerID": playerID, "questID": questID, "serverID": serverID}).Error("server visit validation failed")`.
3. **[MEDIUM PRIORITY]** Improve reward granting to return detailed results including skipped rewards. Change `grantRewards` signature to return `(*QuestRewards, []string, error)` where second return value lists already-claimed rewards for user notification.
4. Add godoc comments to exported constructors (`NewLegendaryQuestGenerator`, `NewQuestManager`) explaining parameters and dependencies.
5. Consider adding quest lifecycle event hooks for better integration with engine quest systems (OnQuestStarted, OnPhaseCompleted, OnQuestCompleted callbacks).
