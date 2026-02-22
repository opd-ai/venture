# Audit: github.com/opd-ai/venture/pkg/world/housing
**Date**: 2026-02-22
**Auditor**: GitHub Copilot (META_AUDIT v2)
**Status**: Needs Work
<!--
Status criteria:
- Complete: All automated checks passed and fewer than 5 non-critical issues identified.
- Incomplete: Audit was stopped early or one or more required checks (e.g., go test, go vet, race) were not run.
- Needs Work: 5 or more issues identified, or any critical/priority-0 failure (e.g., panics, data corruption, security issues).
-->

## Summary
The housing package provides plot placement, guild halls, blueprints, persistence, and UI for player housing. Good overall quality with 74.8% test coverage exceeding the 65% target. Primary concern is direct Ebiten input calls in `ui.go` bypassing the `Input` interface, which violates the codebase's input abstraction guidelines.

## Automated Check Results
| Check | Result |
|---|---|
| `go vet` | ✅ Pass |
| `go test -cover` | 74.8% (target: 65%) ✅ |
| `go test -race` | ✅ Pass |
| WASM vet | ✅ Pass |
| TODO/FIXME count | 0 |
| Non-deterministic rand | 0 occurrences |
| Concrete net types | 0 occurrences |

## Issues Found

### High Severity
- [ ] **Input Abstraction Violation** — `HousingUI.Update()` and `handleSubmenuInput()` call `ebiten.IsKeyPressed()` directly instead of using the `Input` interface. This prevents testing without Ebiten runtime and violates Coding Guideline #2 (Input interface abstraction). (`ui.go:100`, `ui.go:111`, `ui.go:232`, `ui.go:236`, `ui.go:245`, `ui.go:249`)

### Medium Severity
- [ ] **Missing Input Interface Integration** — `HousingUI` has no method to accept an `InputProvider` and thus cannot be tested with `StubInput` or support controller/touch input sources. Add `SetInput(InputProvider)` method. (`ui.go:16-44`)
- [ ] **time.Now Usage in TimeProvider** — While `RealTimeProvider.Now()` correctly calls `time.Now()` (appropriate for production), tests and multiplayer sync must use `MockTimeProvider` to ensure determinism. Currently, `NewPlot()` and `NewBlueprint()` default to `RealTimeProvider`. (`types.go:26`)

### Low Severity
- [ ] **Missing Gamepad Navigation** — `HousingUI` only handles keyboard input (ESC, Tab, Up/Down). No gamepad button mapping for D-pad navigation or confirm/cancel buttons. (`ui.go:100-130`, `ui.go:227-256`)
- [ ] **Missing Touch Input Support** — `HousingUI` has no touch/tap handling for mobile builds. Should support tap-to-select and swipe gestures. (`ui.go`)
- [ ] **doc.go Example Uses log.Printf** — Documentation example shows `log.Printf` instead of `logrus.WithFields` for error logging. Should follow structured logging guidelines. (`doc.go:56`)

## Input Integration
| Input Source | Status | Notes |
|---|---|---|
| Keyboard | ❌ | Direct `ebiten.IsKeyPressed()` calls bypass `Input` interface |
| Mouse | N/A | UI is keyboard-navigated, no mouse click handling needed |
| Gamepad | ❌ | No gamepad/controller support implemented |
| Touch | ❌ | No touch gesture support for mobile |
| VR | N/A | Housing UI not VR-relevant |
| Stub/Test | ❌ | Cannot use `StubInput` due to direct Ebiten calls |

## Menu/UI Integration
| Menu | Reachable | Input-Complete | Backing System Wired | Notes |
|---|---|---|---|---|
| Housing Main | ✅ | ❌ | ✅ | Reachable via keybind; keyboard-only |
| Building Menu | ✅ | ❌ | ✅ | Tab-switchable from main; building types listed |
| Furniture Menu | ✅ | ❌ | ✅ | Tab-switchable; furniture categories functional |
| Guild Hall Menu | ✅ | ❌ | ✅ | Tab-switchable; displays construction progress |

## Test Coverage
**Coverage**: 74.8% (target: 65%) ✅
- Missing test areas: `HousingUI.Update()` input handling (requires Ebiten), `Draw()` rendering
- Missing benchmarks: `SpatialGrid.Query` with large datasets (partially covered), `BlueprintLibrary.Filter` performance
- Table-driven test compliance: ✅ (types_test.go, manager_test.go, blueprint_test.go use table-driven patterns)

## Documentation Coverage
- Package `doc.go`: ✅ Present with comprehensive examples
- Exported symbols documented: ~95% (199 doc comments for 103 methods)
- Complex algorithms commented: ✅ (`spatial.go` grid logic, `guildhall.go` material calculation)

## Integration Status
- System registration: ✅ — `HousingUI` implements `HousingUIProvider` interface for ESC key handling
- Component registration: ✅ — `HousingComponent` implements `Type() string` returning `"housing"`
- Serialize/Deserialize: ✅ — `HousingComponent`, `Manager`, `GuildHallManager`, `Blueprint` all support JSON/gzip serialization
- Network sync: ✅ — `SyncHouseFromFederation()` supports cross-server replication
- Genre theming: ✅ — `HousingUI.GenreID` and `Blueprint.GenreID` support genre filtering
- Mod compatibility: N/A — Housing data structures are JSON-based but no explicit mod override points

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ✅ | Full functionality, keyboard navigation |
| WASM | ✅ | `go vet GOOS=js GOARCH=wasm` passes; no WASM-specific issues |
| Mobile | ❌ | No touch input support in `HousingUI`; keyboard-only |

## Recommendations
1. **[HIGH]** Refactor `HousingUI` to accept `InputProvider` interface instead of calling `ebiten.IsKeyPressed()` directly. Add `SetInput(input InputProvider)` method and route all input checks through the interface.
2. **[HIGH]** Create `StubHousingUI` or add input injection to enable unit testing of `Update()` and `handleSubmenuInput()` without Ebiten runtime.
3. **[MED]** Add gamepad support with D-pad navigation and A/B button mappings for confirm/cancel.
4. **[MED]** Add touch input support with tap-to-select for mobile builds.
5. **[LOW]** Update `doc.go` example to use `logrus.WithFields` instead of `log.Printf`.
