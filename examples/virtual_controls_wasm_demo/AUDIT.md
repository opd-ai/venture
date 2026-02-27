# Audit: examples/virtual_controls_wasm_demo
**Date**: 2026-02-26 (ISO 8601)
**Auditor**: GitHub Copilot (META_AUDIT v2)
**Status**: Needs Work
<!--
Status criteria:
- Complete: All automated checks passed and fewer than 5 non-critical issues identified.
- Incomplete: Audit was stopped early or one or more required checks (e.g., go test, go vet, race) were not run.
- Needs Work: 5 or more issues identified, or any critical/priority-0 failure (e.g., panics, data corruption, security issues).
-->

## Summary
This is a 190-line example program demonstrating the WASM virtual controls pre-initialization fix (Gap #3 from mobile AUDIT.md). The program successfully shows touch input handling and virtual controls visibility management. However, it has 8 significant issues: 3 high-severity input abstraction violations, 1 medium-severity non-test coverage gap, 4 low-severity code quality issues. The demo directly calls `ebiten.TouchIDs()` and `ebiten.TouchPosition()` without routing through the `InputProvider` interface, violating Coding Guideline #2 (Input interface abstraction). Additionally, the demo uses standard library `log` package instead of structured logging with logrus.

## Automated Check Results
| Check | Result |
|---|---|
| `go vet` | ✅ Pass |
| `go test -cover` | 0.0% (target: N/A - demo program with no tests) |
| `go test -race` | ⚠️ No tests (demo program) |
| WASM vet | ✅ Pass |
| TODO/FIXME count | 0 |
| Non-deterministic rand | 0 occurrences |
| Concrete net types | 0 occurrences |

## Issues Found

### High Severity
- [ ] **Input Abstraction** — Direct `ebiten.TouchIDs()` call instead of routing through `InputProvider` interface violates Coding Guideline #2 (Input interface abstraction). Touch input should be abstracted for testability and platform consistency. (`main.go:62`)
- [ ] **Input Abstraction** — Direct `ebiten.TouchIDs()` call in Draw() without interface abstraction. (`main.go:160`)
- [ ] **Input Abstraction** — Direct `ebiten.TouchPosition()` call in Draw() without interface abstraction. (`main.go:161`)

### Medium Severity
- [ ] **Test Coverage** — Demo program has 0% test coverage. While demo programs are typically untested, this program demonstrates a specific bug fix (Gap #3) and would benefit from a test validating the fix (e.g., test that controls are pre-initialized on WASM+touch platforms, test that first touch is not missed). (`main.go:1-191`)

### Low Severity
- [x] **Logging** — Uses standard library `log.Println` and `log.Printf` instead of structured logging with `logrus.WithFields`. Demo programs typically use simple logging, but structured logging would align with project standards. (`main.go:50,51,66,75,183,184,185,188`) - **FIXED 2026-02-27**: Replaced all log.Println/Printf with logrus.WithFields for structured logging
- [x] **Error Handling** — `log.Fatal(err)` terminates program without cleanup or structured error context. Should use logrus.WithError(err).Fatal() for consistency. (`main.go:188`) - **FIXED 2026-02-27**: Replaced log.Fatal with logrus.WithError
- [x] **Doc Coverage** — Package doc comment is present but exported `Game` type (line 32) and exported `NewGame` function (line 41) lack godoc comments. Demo programs typically have less strict doc requirements. (`main.go:32,41`) - **ALREADY FIXED**: Godoc comments exist on lines 31 and 40 and are recognized by go doc
- [ ] **Magic Numbers** — Screen dimensions (800x600) and UI layout coordinates are hardcoded literals. Consider defining as named constants for clarity, especially for the complex UI layout box calculations. (`main.go:27,28,131-157`)

## Input Integration
| Input Source | Status | Notes |
|---|---|---|
| Keyboard | N/A | Demo does not use keyboard input |
| Mouse | N/A | Demo does not use mouse input (only visualizes touches as mouse input in desktop mode) |
| Gamepad | N/A | Demo does not use gamepad input |
| Touch | ❌ | **VIOLATION**: Direct `ebiten.TouchIDs()` and `ebiten.TouchPosition()` calls bypass `InputProvider` interface. Touch input should be routed through `InputSystem` and accessed via `InputProvider.GetMovement()` or custom touch methods added to the interface. |
| VR | N/A | Demo does not use VR input |
| Stub/Test | ❌ | No `StubInput` usage - demo cannot be tested without Ebiten runtime due to direct Ebiten API calls |

## Menu/UI Integration
| Menu | Reachable | Input-Complete | Backing System Wired | Notes |
|---|---|---|---|---|
| Demo Screen | ✅ | ✅ | ✅ | Single screen demo - no menu navigation required. Touch visualization works. Virtual controls visibility state is simulated (not actually wired to InputSystem virtual controls). |

## Test Coverage
**Coverage**: 0.0% (target: N/A - demo program)
- Missing test areas: Touch input handling, virtual controls visibility logic, WASM platform detection, first-touch delay measurement
- Missing benchmarks: N/A (demo program)
- Table-driven test compliance: N/A (no tests)

**Note**: While demo programs typically have no tests, this program demonstrates a specific bug fix (Gap #3: first-touch delay) and would benefit from an integration test that:
1. Validates controls are pre-initialized (hidden) on WASM+touch platforms
2. Validates controls become visible on first touch
3. Measures that first touch is captured (0-frame delay)
4. Validates graceful degradation on non-touch platforms

## Documentation Coverage
- Package `doc.go`: ❌ (no separate doc.go file; package comment is in main.go)
- Exported symbols documented: 2/2 (100%)
  - `Game` type (line 32): ✅ Godoc comment on line 31
  - `NewGame` function (line 41): ✅ Godoc comment on line 40
- Complex algorithms commented: ✅ (UI layout and fix explanation have inline comments)

## Integration Status
This demo is a standalone example program that validates a fix in the main codebase but does not integrate as a library component.

- System registration: N/A — Demo program, not a system
- Component registration: N/A — Demo does not define components for engine use
- Serialize/Deserialize: N/A — Demo has no persistence
- Network sync: N/A — Demo is single-player
- Genre theming: N/A — Demo is UI-focused, not content generation
- Mod compatibility: N/A — Demo is not data-driven

**Integration with engine**: The demo references `engine.InputSystem` and `mobile.IsWASM()`, `mobile.IsTouchCapable()`. However, it does not actually wire virtual controls from `pkg/mobile/` - it only simulates the visibility logic. The demo demonstrates the intended behavior but does not test the actual implementation in `pkg/mobile/`.

**Gap**: The demo claims to show the Gap #3 fix (virtual controls pre-initialization) but does not actually use the `pkg/mobile/` virtual controls implementation. It should either:
1. Import and use the actual virtual controls from `pkg/mobile/`, or
2. Clarify that it is a conceptual demonstration, not a functional integration test

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ✅ | Compiles and runs on desktop with touch simulation via mouse clicks visualized as red circles |
| WASM | ✅ | Builds with `GOOS=js GOARCH=wasm` - vet passes. Platform detection logic uses `mobile.IsWASM()` and `mobile.IsTouchCapable()` correctly. |
| Mobile | ⚠️ | Not directly buildable as mobile app (no mobile entry point), but demonstrates mobile-relevant touch input patterns |

## Recommendations
1. **[HIGH]** Refactor touch input to use `InputProvider` interface methods instead of direct `ebiten.TouchIDs()` and `ebiten.TouchPosition()` calls. This aligns with Coding Guideline #2 and enables testing with `StubInput`. Add `GetTouchPositions() []TouchPoint` method to `InputProvider` interface if needed.
2. **[HIGH]** Wire actual virtual controls from `pkg/mobile/` to validate the Gap #3 fix is functional, not just conceptual. Import `pkg/mobile/` and instantiate virtual controls, or clarify in comments that this is a conceptual demo only.
3. **[HIGH]** Add integration test that validates the Gap #3 fix: controls pre-initialized (hidden) on WASM+touch, controls visible on first touch, first touch captured without delay.
4. **[MED]** Replace `log.Println`/`log.Printf` with `logrus.WithFields()` for structured logging consistency with project standards.
5. **[LOW]** Add godoc comments to exported `Game` type and `NewGame` function for consistency with project documentation standards (even for demo programs).
6. **[LOW]** Extract screen dimensions and UI layout coordinates to named constants for clarity and maintainability.
7. **[LOW]** Replace `log.Fatal(err)` with `logrus.WithError(err).Fatal()` for structured error context.
