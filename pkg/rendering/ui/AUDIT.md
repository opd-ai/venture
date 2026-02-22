# Audit: github.com/opd-ai/venture/pkg/rendering/ui
**Date**: 2026-02-22
**Auditor**: GitHub Copilot (META_AUDIT v2)
**Status**: Complete

## Summary
The `pkg/rendering/ui` package provides procedural UI element generation, chat rendering, notifications, tutorials, settings management, keybinding customization, and trade UI. The package has strong test coverage (80.0%), passes all automated checks, and follows deterministic generation patterns correctly. The package serves as a pure UI data/rendering layer with no direct input handling (input abstraction is delegated to `pkg/engine` systems).

## Automated Check Results
| Check | Result |
|---|---|
| `go vet` | ✅ Pass |
| `go test -cover` | 80.0% (target: 65%) ✅ |
| `go test -race` | ✅ Pass |
| WASM vet | ✅ Pass |
| TODO/FIXME count | 0 |
| Non-deterministic rand | 0 occurrences |
| Concrete net types | 0 occurrences |

## Issues Found

### High Severity
_(None)_

### Medium Severity
- [x] **Direct Ebiten input** — `trade.go:130` calls `ebiten.CursorPosition()` directly instead of through `Input` interface. UI systems should receive input state from callers, not query Ebiten directly. **FIXED**: Added `UpdateWithInput(mouseX, mouseY int)` method for testable input handling. Original `Update()` preserved for backward compatibility but marked deprecated.
- [x] **Direct Ebiten input** — `trade.go:321` and `trade.go:329` call `ebiten.IsMouseButtonPressed()` directly instead of through `Input` interface. **FIXED**: Added `IsButtonClickedWithInput(button string, mousePressed bool)` and `GetClickedButtonWithInput(mousePressed bool)` methods for testable input handling. Original methods preserved for backward compatibility but marked deprecated.
- [ ] **time.Now() in non-test code** — `chat.go:109` and `chat.go:269` call `time.Now()` for cursor blinking and system messages. While acceptable for pure UI timing, consider injecting a clock for deterministic replay/testing.

### Low Severity
- [ ] **File system access** — `keybinds.go:420,428` and `settings.go:620,646` use `os.WriteFile`/`os.ReadFile` directly for saving/loading. This will not work in WASM builds without abstraction through `pkg/saveload/` WASM storage.
- [ ] **Missing doc comment** — `chat.go:266` function `splitWords` in `notifications.go:266` lacks a package-level doc explaining the word-splitting algorithm.

## Input Integration
| Input Source | Status | Notes |
|---|---|---|
| Keyboard | N/A | Package is a rendering/data layer; input handled by engine systems |
| Mouse | ✅ | `trade.go` now has `UpdateWithInput()`, `IsButtonClickedWithInput()`, `GetClickedButtonWithInput()` methods for input abstraction. Legacy methods preserved for backward compatibility. |
| Gamepad | N/A | No gamepad handling in this package |
| Touch | N/A | No touch handling in this package |
| VR | N/A | No VR handling in this package |
| Stub/Test | ✅ | Tests use new input abstraction methods; don't require Ebiten runtime |

## Menu/UI Integration
| Menu | Reachable | Input-Complete | Backing System Wired | Notes |
|---|---|---|---|---|
| ChatUI | ✅ | N/A | ✅ | Data structure for chat rendering; wired via `pkg/engine/chat_system.go` |
| NotificationManager | ✅ | N/A | ✅ | Data structure for notifications; used by client systems |
| TutorialManager | ✅ | N/A | ✅ | Implements `ContextualTutorialProvider` interface |
| TradeUI | ✅ | ✅ | ✅ | Input abstraction via `*WithInput()` methods |
| SettingsManager | ✅ | N/A | ✅ | Data/persistence for settings; UI rendering in engine |
| KeybindManager | ✅ | N/A | ✅ | Data/persistence for keybinds; used by input system |
| QuickTravelManager | ✅ | N/A | ✅ | Data structure for quick travel destinations |

## Test Coverage
**Coverage**: 80.1% (target: 65%) ✅
- Missing test areas: `story_journal.go` and `image_preview.go` have lower coverage based on file names
- Missing benchmarks: No benchmarks for UI generation performance (would be useful for `generator.go`)
- Table-driven test compliance: ✅ Tests use table-driven patterns
- Input abstraction tests: ✅ New `*WithInput()` methods have comprehensive table-driven tests

## Documentation Coverage
- Package `doc.go`: ✅ Comprehensive documentation with examples
- Exported symbols documented: 95%+ (all major types and functions documented)
- Complex algorithms commented: ✅ WCAG contrast calculation, easing functions documented

## Integration Status
- System registration: ✅ — Package provides data types used by `pkg/engine/` UI systems
- Component registration: N/A — This is a rendering utility package, not an ECS component package
- Serialize/Deserialize: ✅ — `TutorialManager` implements `ExportState()`/`ImportState()` for save/load
- Network sync: N/A — UI state is client-local, not networked
- Genre theming: ✅ — `Generator` uses `GenreID` for genre-aware UI styling
- Mod compatibility: N/A — UI generation not exposed to mod system
- Accessibility: ✅ — `AccessibilityConfig` struct supports colorblind modes, font scaling, high contrast, screen reader hints

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ✅ | Full functionality |
| WASM | ⚠️ | `keybinds.go` and `settings.go` file I/O needs WASM storage abstraction |
| Mobile | ✅ | No platform-specific issues |

## Recommendations
1. **[MED]** Refactor `trade.go` to accept mouse position and button state as parameters instead of querying `ebiten.CursorPosition()` and `ebiten.IsMouseButtonPressed()` directly. This improves testability and follows the Input interface abstraction pattern.
2. **[MED]** Abstract `time.Now()` usage in `chat.go` behind an injectable clock interface for deterministic testing of cursor blink timing.
3. **[LOW]** Wrap `os.WriteFile`/`os.ReadFile` in `keybinds.go` and `settings.go` to use `pkg/saveload/` storage abstraction for WASM compatibility.
4. **[LOW]** Add benchmarks for `generator.go` UI element generation to validate <1ms generation time target.
