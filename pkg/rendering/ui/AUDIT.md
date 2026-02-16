# Audit: pkg/rendering/ui

**Date**: 2026-02-16
**Auditor**: Automated
**Coverage**: 79.9%

## Summary

The `pkg/rendering/ui` package provides procedural UI element generation including buttons, panels, health bars, decorative frames, chat UI, notifications, settings management, keybindings, quick travel, tooltips, transitions, tutorials, and accessibility features.

**Files**: 16 source files, ~4,625 lines of code
**Tests**: 12 test files with comprehensive coverage

## Issues Found and Fixed

### Issue 1 — Settings Save() data race (Medium)
- **File**: `settings.go:620`
- **Problem**: `Save()` acquired `RLock()` but modified `sm.modified = false` on line 640, causing a data race since `RLock` allows concurrent readers.
- **Fix**: Changed `sm.mu.RLock()` to `sm.mu.Lock()` in `Save()` since it mutates state.
- **Severity**: Medium — potential data race under concurrent access.

### Issue 2 — Default keybind conflict: VehicleMount/VehicleDismount (Medium)
- **File**: `keybinds.go:251-252`
- **Problem**: Both `ActionVehicleMount` and `ActionVehicleDismount` were bound to `KeyV`, causing `keyMap` to map `KeyV` to whichever action was registered last, silently losing the first mapping.
- **Fix**: Changed `ActionVehicleDismount` to use `NumPad5` to eliminate the conflict.
- **Severity**: Medium — silent default keybind conflict, one action unreachable by key.

### Issue 3 — Keybind Save/Load loses Description field (Low)
- **File**: `keybinds.go:404-464`
- **Problem**: `Save()` serialized only `primary` and `secondary` keys. `Load()` created `Keybind` structs without `Description`, losing it on a save/load round-trip.
- **Fix**: Added `description` to both `Save()` serialization and `Load()` deserialization.
- **Severity**: Low — descriptions lost after save/load cycle, UI display issue.

### Issue 4 — Division by zero in calculateGradientAlpha (Low)
- **File**: `hierarchy.go:272`
- **Problem**: `calculateGradientAlpha()` computed `fadeWidth = width / 4` then divided by `fadeWidth` without checking for zero. When `width < 4`, `fadeWidth` would be 0, causing a divide-by-zero panic.
- **Fix**: Added early return when `fadeWidth == 0`, returning `1.0` (full opacity).
- **Severity**: Low — only triggered with very small separator widths (< 4px).

## Remaining Observations (Not Fixed)

- **Hardcoded colors**: Chat, notifications, trade, and image preview UIs have hardcoded color values. These are visual constants and don't affect correctness.
- **Approximate text wrapping**: `notifications.go` and `story_journal.go` use character-count-based wrapping (~7px/char) rather than font metrics. Acceptable for procedural UI.
- **time.Now() usage**: Chat, notifications, image preview, and tutorial use `time.Now()` for UI lifecycle management (cursor blink, expiry, timestamps). This is expected for interactive UI components and does not affect determinism of generation.
- **Tests require display**: Test suite requires X11/Xvfb due to ebiten dependency in source files. Run with `xvfb-run -a go test ./pkg/rendering/ui/`.

## Test Coverage

| Area | Coverage |
|------|----------|
| Overall | 79.9% |
| Settings | High |
| Keybinds | High |
| Quick Travel / Tooltips | High |
| Hierarchy / Separators | High |
| Generator | High |
| Transitions | High |
| Tutorial / Accessibility | High |
