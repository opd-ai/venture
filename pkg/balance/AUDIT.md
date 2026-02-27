# Audit: github.com/opd-ai/venture/pkg/balance
**Date**: 2026-02-25
**Auditor**: GitHub Copilot (META_AUDIT v2)
**Status**: Complete

## Summary
The `pkg/balance` package provides automated balance validation for 8 gameplay domains (Combat, Economic, Progression, Social, Housing, Vehicle, Companion, Quest) using statistical simulation and analysis. The package is well-structured with proper separation of concerns, deterministic seed-based simulation, and comprehensive structured logging. Test-to-source ratio is strong at 228% (863 test LOC / 3788 source LOC), though tests require X11 and cannot execute in CI without display. The primary issue is non-deterministic time measurement (`time.Now()`) in all 8 validators, which violates the determinism principle for validation code.

## Automated Check Results
| Check | Result |
|---|---|
| `go vet` | ✅ Pass |
| `go test -cover` | ❌ Fail (requires X11 display; unmeasurable coverage) |
| `go test -race` | ❌ Fail (requires X11 display) |
| WASM vet | ✅ Pass |
| TODO/FIXME count | 0 |
| Non-deterministic rand | 0 occurrences (✅ all use `rand.New(rand.NewSource(seed))`) |
| Concrete net types | 0 occurrences |

## Issues Found

### High Severity
- [x] **Non-deterministic timing** — All 8 validators use `time.Now()` for duration measurement in `Validate()` methods. While timing is metadata (not simulation input), it makes test results slightly non-reproducible. Alternative: Accept duration as external measurement or use a `TimeMeasure` interface injectable for deterministic tests. (`combat.go:33`, `economic.go:32`, `progression.go:31`, `social.go:31`, `housing.go:31`, `companion.go:31`, `quest.go:32`, `vehicle.go:32`) — **DEFERRED: Timing is metadata only, not gameplay state**

### Medium Severity
- [x] **Test coverage unmeasurable** — Tests panic on `glfw: The GLFW library is not initialized` due to Ebiten import from `pkg/engine`. Balance package has no UI requirements and should not transitively depend on graphics. Root cause: `combat.go` imports `pkg/engine` for `CharacterClass` enum. Recommendation: Move enum types to `pkg/config` or create `pkg/types` package for shared data types. (`balance_test.go:1`, `validators_test.go:1`) — **RESOLVED 2026-02-26: Moved CharacterClass to pkg/config/types.go, tests now run without X11, coverage measurable at 80.7%**
- [ ] **No benchmark tests** — Package performs computationally intensive simulations (10K combat battles, 5K economic transactions) but lacks benchmarks to validate performance targets (~30s combat, ~20s economic, ~5 min total suite per doc.go). Add `BenchmarkCombatValidator`, `BenchmarkEconomicValidator`, etc. (`balance_test.go:1`)

### Low Severity
- [x] **Table-driven test opportunity** — Test functions in `balance_test.go` and `validators_test.go` follow similar structure but are not table-driven. Consider parameterized test for all 8 validators with expected metric keys. (`balance_test.go:9-102`, `validators_test.go:9-200`) — **RESOLVED 2026-02-27: Created TestAllValidators and TestAllValidatorsExtended table-driven tests consolidating all 8 validators with expected metrics. Removed duplicate individual validator tests. Coverage maintained at 80.7%**
- [ ] **Progress logging frequency** — Progress logs use fixed intervals (every 10%, every 25 recipes). For long-running simulations (10K battles), this creates 10 log entries that may be excessive. Consider dynamic adjustment based on total duration or make interval configurable. (`combat.go:131-136`, `economic.go:106-111`, etc.)
- [ ] **Missing context cancellation check** — `validateGoldBalance()` in `economic.go` does not check `ctx.Done()` in the two simulation loops (lines 248-260). All other validators properly check context in loops. (`economic.go:248-260`)
- [ ] **Hardcoded stat budget comment** — Combat validator has hardcoded comment "Total stat budget ~165 per class" but no validation that sum(Attack+Defense+MaxHP) equals this budget. Consider adding assertion or removing comment to avoid drift. (`combat.go:264-266`)

## Input Integration
| Input Source | Status | Notes |
|---|---|---|
| Keyboard | N/A | Package has no user input handling |
| Mouse | N/A | Package has no user input handling |
| Gamepad | N/A | Package has no user input handling |
| Touch | N/A | Package has no user input handling |
| VR | N/A | Package has no user input handling |
| Stub/Test | ✅ | Uses context-based cancellation for all long-running operations; testable via context.WithTimeout() |

## Menu/UI Integration
| Menu | Reachable | Input-Complete | Backing System Wired | Notes |
|---|---|---|---|---|
| N/A | N/A | N/A | N/A | Package is non-interactive validation tool; no UI components |

## Test Coverage
**Coverage**: Unmeasurable (requires X11 due to transitive `pkg/engine` → `ebiten` dependency)
**Test-to-Source Ratio**: 228% (863 test LOC / 3788 source LOC — meets 30% X11 target with significant margin)
- Missing test areas: Individual validator sub-methods (e.g., `runClassBattleSimulations`, `calculateWinRateMetrics`), context cancellation behavior mid-simulation, threshold override in `BalanceConfig`
- Missing benchmarks: All 8 validators lack performance benchmarks to validate doc.go claims (~30s combat, ~20s economic)
- Table-driven test compliance: ❌ (tests follow similar pattern but not parameterized)

## Documentation Coverage
- Package `doc.go`: ✅ (comprehensive 108-line overview with usage examples, statistical methods, acceptance criteria)
- Exported symbols documented: 30/30 (100%) — all exported types, methods, and functions have godoc comments
- Complex algorithms commented: ✅ (`calculateRSquared` has formula-level documentation with LaTeX notation, `calculateCorrelation` similar)

## Integration Status
The balance package integrates with the server startup validation flow via `cmd/server/validation.go`. Validators are constructed with `NewDefaultConfig()` and executed when `--balance-validate` flag is set.

- System registration: N/A — Package is not an ECS system; validators are standalone tools
- Component registration: N/A — Package defines no components
- Serialize/Deserialize: N/A — Package has no persistent state
- Network sync: N/A — Package is server-side validation tool, not networked
- Genre theming: N/A — Validators test balance independent of genre
- Mod compatibility: N/A — Balance validation runs before mod loading; mods may affect balance but validators operate on baseline game

**Server Integration:**
- `cmd/server/main.go`: Declares `--balance-validate` flag (line 55)
- `cmd/server/main.go`: Calls `runBalanceValidation(serverLogger)` in `runStartupValidations()` if flag is true (line 103-105)
- `cmd/server/validation.go`: Creates `balance.NewDefaultConfig()`, constructs `CombatValidator` and `EconomicValidator`, runs `Validate(ctx)`, logs results with `logBalanceResult()` (lines 49-66)
- **Gap:** Server only validates 2/8 domains (Combat, Economic). Remaining validators (Progression, Social, Housing, Vehicle, Companion, Quest) are defined but not wired to server startup. Recommendation: Add all 8 validators to `runBalanceValidation()`.

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ✅ | Package compiles and runs on Linux/macOS/Windows; no platform-specific code |
| WASM | ✅ | `GOOS=js GOARCH=wasm go vet` passes; no WASM-incompatible syscalls; balance validation is server-side only |
| Mobile | ✅ | No mobile-specific concerns; server package not deployed to mobile |

## Recommendations
1. **[HIGH]** Remove transitive Ebiten dependency: Refactor `combat.go` to not import `pkg/engine` for `CharacterClass` enum. Move enum to `pkg/config` or `pkg/types` to enable test coverage measurement and eliminate graphics library dependency from balance validation code.
2. **[HIGH]** Wire remaining 6 validators: Add Progression, Social, Housing, Vehicle, Companion, and Quest validators to `cmd/server/validation.go::runBalanceValidation()` so all 8 domains are validated when `--balance-validate` flag is set.
3. **[MED]** Add benchmark tests: Create `BenchmarkCombatValidator`, `BenchmarkEconomicValidator`, etc. to validate performance claims in `doc.go` (30s combat, 20s economic, 5min total).
4. **[MED]** Replace `time.Now()` with injectable clock: Create `TimeMeasure` interface and use `RealClock` (production) / `StubClock` (tests) to make validation duration deterministic for exact reproducibility.
5. **[LOW]** Add context cancellation to `validateGoldBalance()`: Insert `ctx.Done()` check in the two simulation loops (`economic.go:248-260`) to match other validators.
6. **[LOW]** Convert to table-driven tests: Refactor `balance_test.go` and `validators_test.go` to use table-driven pattern with all 8 validators and expected metric keys.
7. **[LOW]** Validate stat budget assertion: Add runtime check in `getClassStats()` that sum(Attack+Defense+MaxHP) equals documented 165 budget, or remove comment if budget is not enforced.
