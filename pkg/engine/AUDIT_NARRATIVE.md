# Engine Narrative & Dialog Systems Sub-Audit

**Date**: 2026-02-16
**Auditor**: Copilot
**Scope**: Narrative, dialog, investigation, and quest tracking systems in `pkg/engine/`

## Files Audited

| File | Lines | Description |
|------|-------|-------------|
| `narrative_system.go` | 386 | Core narrative system with event tracking and story progression |
| `narrative_component.go` | 349 | Narrative component with events, triggers, relationships |
| `branching_narrative_system.go` | 224 | Branching story arc system with player choices |
| `branching_narrative_component.go` | 23 | Branching narrative component (pure data) |
| `dialog_system.go` | 204 | Simple choice-based dialog system |
| `dialog_ui.go` | 401 | Dialog UI rendering and input handling |
| `npcdialog_system.go` | 385 | NPC dialog with Markov chain generation |
| `npcdialog_component.go` | 171 | NPC dialog component with conversation state |
| `markov_dialog_provider.go` | 108 | Markov dialog provider implementation |
| `investigation_system.go` | ~280 | Investigation/fragment revelation system |
| `investigation_component.go` | 152 | Investigation component with cooldowns and skills |
| `quest_tracker.go` | 267 | Quest tracking component with story unlocks |

## Issues Found and Fixed

### Issue 1 — Non-deterministic conversation ID generation (Medium)
- **File**: `npcdialog_system.go:122-124`
- **Problem**: Used `time.Now().UnixNano()` for conversation ID generation, violating the project's deterministic generation requirement. Same NPC interactions could produce different conversation IDs across runs with the same seed.
- **Fix**: Changed to use entity ID and conversation length (`GetConversationLength()`) for deterministic, reproducible conversation IDs. Removed unused `time` import.

### Issue 2 — Excessive debug logging in investigation system (Medium)
- **File**: `investigation_system.go` (entire file)
- **Problem**: Every function had redundant entry/exit debug logs, including simple getters like `IsFragmentHidden()` and `IsInvestigating()`. This created enormous log noise at debug level, harmed performance through unnecessary log allocations, and was inconsistent with the gated logging pattern used in other audited systems (narrative, dialog, combat).
- **Fix**: Refactored to use a local `*logrus.Entry` logger field (consistent with NarrativeSystem, BranchingNarrativeSystem). Added log level gating (`GetLevel() >= log.InfoLevel`). Removed entry/exit debug logs from simple getters. Simplified internal method signatures by removing unused entity ID parameters. Reduced ~550 lines to ~280 lines with same functionality.

### Issue 3 — Low test coverage on NarrativeSystem.Update() (Low)
- **File**: `narrative_system.go:60-70`, `narrative_system_test.go`
- **Problem**: Core `Update()` loop and all its helper methods (`getNarrativeComponent`, `processTriggeredEvents`, `processActProgression`) had 0% coverage.
- **Fix**: Added `TestNarrativeSystem_Update`, `TestNarrativeSystem_Update_ActProgression`, `TestNarrativeSystem_Update_SkipsNonNarrative` tests. Coverage for `Update()` now 100%.

### Issue 4 — Low test coverage on NPCDialogComponent methods (Low)
- **File**: `npcdialog_component.go`, `npcdialog_component_test.go`
- **Problem**: Several methods had 0% coverage: `HasDiscussedTopic`, `GetTimeSinceLastInteraction`, `IsFirstInteraction`, `GetConversationLength`, `SetDeterministicMode`, `IsDeterministicMode`, `ClearHistory`.
- **Fix**: Added 5 new tests covering all uncovered methods. All now at 100% coverage.

### Issue 5 — StoryAct.String() untested (Low)
- **File**: `narrative_component.go:70`
- **Problem**: `StoryAct.String()` method had 0% test coverage.
- **Fix**: Added `TestStoryAct_AllStrings` test covering all acts including unknown default case. Now at 100% coverage.

## Remaining Low-Priority Items

1. **dialog_ui.go Draw/render methods**: `Draw()`, `drawWrappedText()`, `findSelectedOption()` at 0% coverage. These require Ebiten rendering context and are impractical to unit test.
2. **dialog_system.go Update()**: Empty no-op method reserved for future use (timed dialogs, auto-close). Not worth testing in current state.
3. **npcdialog_system.go**: Several multi-party conversation methods at 0% (`StartMultiPartyConversation`, `GetConversationMessages`, `AddConversationMessage`, `CleanupStaleConversations`). These delegate to ConversationManager which has its own tests.

## Coverage Summary

| File | Before | After |
|------|--------|-------|
| narrative_system.go | ~60% | ~85% |
| narrative_component.go | ~90% | ~95% |
| npcdialog_component.go | ~55% | ~92% |
| investigation_system.go | ~80% | ~85% |
| investigation_component.go | 100% | 100% |
| quest_tracker.go | ~95% | ~95% |
| branching_narrative_system.go | ~70% | ~70% |
| dialog_system.go | ~75% | ~75% |
| markov_dialog_provider.go | ~90% | ~90% |

## Audit Result

**5 issues fixed** (0 high, 2 medium, 3 low); 3 low remaining (UI rendering methods, empty no-op, delegating methods)
