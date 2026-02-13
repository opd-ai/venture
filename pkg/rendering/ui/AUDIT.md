# Audit: github.com/opd-ai/venture/pkg/rendering/ui
**Date**: 2026-02-13
**Status**: Needs Work

## Summary
The `pkg/rendering/ui` package provides procedural UI element generation for menus, buttons, panels, and interface components. The package is feature-complete with 16 source files covering visual hierarchy, transitions, decorations, and specialized UI (chat, settings, keybinds, notifications, etc.). While implementation quality is high with documented 92.6% test coverage, several issues exist around non-deterministic time usage and placeholder implementations.

## Issues Found
- [ ] **high** deterministic-procgen — Uses `time.Now()` for UI state (cursor blink, notifications, chat timestamps) in non-test files. UI state timing should use deterministic frame counters or explicit timestamps passed from caller. (`chat.go:90,109,268,319`, `tutorial.go:261`, `notifications.go:107,145`, `image_preview.go:129`)
- [ ] **med** stub-code — Placeholder text rendering implementation in StoryJournalUI.drawText() marked as "simple placeholder" that draws colored rectangles instead of actual text (`story_journal.go:442-445`)
- [ ] **low** test-infrastructure — Tests fail in headless/CI environment due to Ebiten GLFW initialization requirement. Package documentation claims 92.6% coverage but tests cannot run without GUI environment (`go test` output shows panic)
- [ ] **low** integration — Package has zero direct importers outside tests. UI generators are not currently integrated into client or engine systems (0 imports found via grep)
- [ ] **low** doc-coverage — Chat.go placeholder text usage documented in comment but not in package godoc (`chat.go:237-240`)

## Test Coverage
92.6% (documented in `doc.go`, cannot verify via `go test -cover` due to GUI dependency)

## Integration Status
The package provides UI generation utilities but is not yet integrated into the game client. No imports found in `cmd/client/` or `pkg/engine/`. The package appears to be a standalone library waiting for integration. UI systems in client likely use direct Ebiten drawing rather than these procedural generators.

Key integration points that should exist:
- Client menu rendering (should use `Generator.Generate()` for buttons/panels)
- HUD elements (should use health bars, icons, labels)
- Settings UI (should use `SettingsManager` from `settings.go`)
- Chat UI (should use `ChatUI` from `chat.go`)
- Keybind customization (should use `KeybindManager` from `keybinds.go`)

## Recommendations
1. **HIGH PRIORITY**: Refactor time-based UI state to use deterministic frame counters. Replace all `time.Now()` calls in non-test production code with frame-based timing or externally provided timestamps. This is critical for maintaining deterministic behavior and replay capability.
2. **HIGH PRIORITY**: Integrate UI generators into client. Add imports in `cmd/client/` to actually use the procedural UI generation system instead of letting this code remain unused.
3. **MEDIUM PRIORITY**: Complete StoryJournalUI.drawText() implementation with proper font rendering using existing font libraries (already imported `golang.org/x/image/font`)
4. **LOW PRIORITY**: Add test build tag to skip GUI tests in headless environments (`// +build !headless` or use Ebiten's test mode)
5. **LOW PRIORITY**: Document placeholder text behavior in package godoc for completeness
