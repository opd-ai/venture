# Audit: github.com/opd-ai/venture/pkg/engine/prestige
**Date**: 2026-02-25
**Auditor**: GitHub Copilot (META_AUDIT v2)
**Status**: Needs Work
<!--
Status criteria:
- Complete: All automated checks passed and fewer than 5 non-critical issues identified.
- Incomplete: Audit was stopped early or one or more required checks (e.g., go test, go vet, race) were not run.
- Needs Work: 5 or more issues identified, or any critical/priority-0 failure (e.g., panics, data corruption, security issues).
-->

## Summary
The prestige package provides post-max-level progression with prestige levels, paragon points, and prestige abilities. The package has excellent test coverage (103.5% test-to-source ratio) and clean architecture with proper ECS separation via interface adapters. However, it suffers from one critical UI integration gap and several medium-severity issues with time.Now() usage that violate determinism guidelines for metadata timestamps.

## Automated Check Results
| Check | Result |
|---|---|
| `go vet` | ✅ Pass |
| `go test -cover` | Unmeasurable (requires X11; 103.5% test-to-source ratio) |
| `go test -race` | ⚠️ Requires X11 (cannot run in headless environment) |
| WASM vet | N/A (not WASM-specific) |
| TODO/FIXME count | 0 |
| Non-deterministic rand | 0 occurrences |
| Concrete net types | 0 occurrences |

## Issues Found

### High Severity
- [ ] **UI Integration** — PrestigeUI is fully implemented with keyboard/touch controls and callbacks but is NOT registered in `cmd/client/handlers.go` or `cmd/client/main.go`. The prestige system is active (via `prestigeSystemWrapper`), but players have no way to access the UI to allocate paragon points or view prestige progress. This is a complete feature gap. (`ui.go:1-510`, `cmd/client/handlers.go:945-1323`)

### Medium Severity
- [ ] **Non-Deterministic Timestamps** — Uses `time.Now()` for `LastUpdated` metadata in 10 locations. While documented as "for audit/debugging purposes only" in manager.go:15, this violates Coding Guideline #2 (Deterministic Generation). Metadata timestamps should use a injected clock interface (`GameClock` already exists in `pkg/engine/interfaces.go:521-528`) to maintain determinism. (`manager.go:49, 61, 91, 124, 147, 174, 269, 306`; `manager_test.go:552`)
- [x] **Input Abstraction Violation** — ✅ **ALREADY FIXED** — PrestigeUI now uses the standard `InputProvider` interface from `pkg/engine/interfaces.go:171-236`. UI correctly uses `IsMenuUpJustPressed()`, `IsMenuDownJustPressed()`, `IsMenuConfirmJustPressed()`, `IsMenuBackJustPressed()` methods. No custom `MenuInputProvider` interface exists in codebase. (`ui.go:98-99, 175-177, 244-277`)
- [x] **Missing Struct Logging** — Manager and System constructors do not use structured logging with `logrus.WithFields`. Manager has no logger at all; System logs creation but should include constructor parameters. Add `logrus.Fields{"playerID", "className", "accountID"}` where relevant. (`manager.go:27-33`, `system.go:29-48`) — **FIXED 2026-02-27**: Added logger field to Manager with NewManagerWithLogger constructor. Added structured logging to NewManagerWithLogger (component field) and CreatePlayer (playerID, className, accountID fields). Updated System constructors to log creation with Debug level. All constructors now use logrus.WithFields for structured logging.

### Low Severity
- [ ] **Doc Coverage** — Package has excellent `doc.go` (104 lines) with comprehensive usage examples and integration notes. All exported types and methods are documented. No issues.
- [ ] **Missing Benchmarks** — No benchmarks exist for hot-path operations (`AddPrestigeXP`, `AllocateParagonPoint`, `GetStatBonus`, `GetAccountXPBonus`). The doc.go:89-93 specifies performance targets (<1ms XP addition, <5ms point allocation, <0.1ms ability check, <10ms account bonus), but these are not validated. Add benchmarks in `manager_test.go`.
- [ ] **Component Cache** — `PrestigeComponent` is not listed in the engine's hot-path component cache (`pkg/engine/ecs.go:Entity` fields like `positionCache`, `velocityCache`, etc.). If prestige checks are frequent (e.g., every frame for visual tier), consider adding `prestigeCache *PrestigeComponent` to Entity. (`types.go:136-151`, integration note)
- [x] **Error Context** — Error returns in `Manager.AllocateParagonPoint` and `Manager.RespecParagonPoints` use `fmt.Errorf` without wrapping context. Use `fmt.Errorf("context: %w", err)` pattern for error chains where applicable. (`manager.go:134, 159`) - **FIXED 2026-02-27**: Added sentinel errors (ErrPlayerNotFound, ErrNoParagonPoints, ErrInvalidStat, ErrAccountNotFound, ErrUnknownParagonCategory) to manager.go and updated all error returns to use fmt.Errorf with %w wrapping for proper error chains

## Input Integration
| Input Source | Status | Notes |
|---|---|---|
| Keyboard | ⚠️ Partial | UI handles arrow keys, Enter, Space, ESC via `MenuInputProvider` interface, but should use canonical `InputProvider` interface from engine (Medium Severity issue) |
| Mouse | ❌ | No mouse support for clicking menu options; keyboard-only navigation |
| Gamepad | ❌ | No gamepad navigation support (could map D-pad to Up/Down) |
| Touch | ✅ | Excellent touch support with `mobile.TouchInputHandler` and touch buttons for all actions (allocate, respec, back, close) |
| VR | N/A | Not applicable for menu UI |
| Stub/Test | ✅ | `SetInputProvider()` method enables injection of test input stub; `MenuInputProvider` interface used in tests |

## Menu/UI Integration
| Menu | Reachable | Input-Complete | Backing System Wired | Notes |
|---|---|---|---|---|
| Prestige Menu | ❌ | ✅ | ✅ | **HIGH SEVERITY**: UI is fully implemented (`ui.go`) with keyboard (Up/Down/Enter/ESC) and touch controls, but is NOT registered in client entry point. No keybind or menu option exists to open it. System is active via `prestigeSystemWrapper` in `handlers.go:1323`. |

## Documentation Coverage
- Package `doc.go`: ✅ (104 lines with usage examples, integration notes, performance targets, testing section)
- Exported symbols documented: 100% (all types, methods, constants, and functions have godoc comments)
- Complex algorithms commented: ✅ (XP curve calculation, account bonus stacking formula, visual tier thresholds all documented inline)

## Integration Status
The prestige package integrates with the engine via interface adapters to avoid circular dependencies.

- System registration: ✅ — System is registered in `cmd/client/handlers.go:1323` via `prestigeSystemWrapper`. The wrapper adapts `[]*engine.Entity` to `[]prestige.Entity` using `prestigeEntityAdapter` (system_wrappers.go:432-473).
- Component registration: ✅ — `PrestigeComponent` implements `Component` interface with `Type() string` returning "prestige". Component is added via `InitializePlayer()` in `system.go:99-116`.
- Serialize/Deserialize: ✅ — `PrestigeComponent` implements `Serialize() ([]byte, error)` and `Deserialize(data []byte) error` using JSON marshaling (types.go:154-161). Manager has `Save()/Load()` with gzip compression (manager.go:358-418).
- Network sync: ❌ — No network sync documented. `PrestigeComponent.ActiveAbilities []string` may need replication for multiplayer if prestige abilities affect combat. Check if prestige data is server-authoritative or client-side only.
- Genre theming: N/A — Prestige abilities are class-based, not genre-based. `generateAbilitiesForClass()` uses class name for deterministic generation (manager.go:320-356).
- Mod compatibility: N/A — No mod integration documented. Prestige abilities could be extended via mod system (`pkg/modding/`) to override ability names/descriptions/stats.

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ✅ | Works on Linux/macOS/Windows. Uses `ebiten.Key`, `inpututil`, and `ebitenutil` for rendering. |
| WASM | ✅ | No platform-specific code. Should work in browser if UI is registered. Touch controls present via `mobile.TouchInputHandler`. |
| Mobile | ✅ | Touch support is first-class with touch buttons for all actions (`ui.go:113-183`). Uses `pkg/mobile` package. |

## Recommendations
1. **[HIGH]** Register PrestigeUI in `cmd/client/handlers.go` or `cmd/client/main.go` with a keybind (e.g., `P` for Prestige). Add to pause menu or character sheet as tab. Wire `SetBackCallback()` to hide UI and return to previous menu. Wire `SetRespecCallback()` to check player gold and deduct cost.
2. **[HIGH]** Add keybind constant to `pkg/engine/menu_keys.go` (e.g., `PrestigeMenu ebiten.Key = ebiten.KeyP`). Update `input_system.go` to handle prestige menu toggle similar to inventory/character/quest UIs.
3. **[MED]** Replace `time.Now()` calls with injected `GameClock` interface (already exists in `interfaces.go:521-528`). Add `clock GameClock` field to Manager and System. Pass clock in constructors. Use `clock.Now()` instead of `time.Now()`. This maintains determinism for metadata timestamps.
4. **[MED]** Refactor PrestigeUI to use canonical `InputProvider` interface from `pkg/engine/interfaces.go:171-236` instead of custom `MenuInputProvider`. Map `IsMenuUpJustPressed()`, `IsMenuDownJustPressed()`, `IsMenuConfirmJustPressed()`, `IsMenuBackJustPressed()` to navigation. This aligns with project input abstraction pattern.
5. **[MED]** Add structured logging to Manager constructor: `logger := logrus.WithField("system_name", "prestige")` and log creation with `logger.Debug("Creating prestige manager")`. System constructor already logs but should include any configuration parameters.
6. **[LOW]** Add benchmarks for hot-path operations: `BenchmarkAddPrestigeXP`, `BenchmarkAllocateParagonPoint`, `BenchmarkGetStatBonus`, `BenchmarkGetAccountXPBonus`. Validate against targets in doc.go:89-93 (<1ms, <5ms, <0.1ms, <10ms respectively).
7. **[LOW]** Consider adding `prestigeCache *PrestigeComponent` to `pkg/engine/ecs.go:Entity` if prestige visual tier checks are frequent (e.g., every frame for aura rendering). Current implementation queries component via `entity.GetComponent("prestige")` which has map lookup overhead.
8. **[LOW]** Add mouse click support to PrestigeUI menu options. Currently keyboard + touch but no mouse. Users expect to click menu items on desktop. Check other menus (inventory, quest, etc.) for mouse handling pattern.
