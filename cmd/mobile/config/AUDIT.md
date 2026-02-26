# Audit: github.com/opd-ai/venture/cmd/mobile/config
**Date**: 2026-02-26
**Auditor**: GitHub Copilot (META_AUDIT v2)
**Status**: Needs Work
<!--
Status criteria:
- Complete: All automated checks passed and fewer than 5 non-critical issues identified.
- Incomplete: Audit was stopped early or one or more required checks (e.g., go test, go vet, race) were not run.
- Needs Work: 5 or more issues identified, or any critical/priority-0 failure (e.g., panics, data corruption, security issues).
-->

## Summary
The `cmd/mobile/config` package provides environment-based configuration for mobile game initialization (seed and genre selection). The package has 73.9% test coverage exceeding the 40% target, passes all automated checks, and has excellent documentation. However, it contains one **critical architectural issue**: intentional non-deterministic seed generation via `time.Now()` that contradicts the project's core deterministic generation principles (Coding Guideline #2), even though the README correctly claims "time-based seed with a warning" as fallback behavior.

## Automated Check Results
| Check | Result |
|---|---|
| `go vet` | ✅ Pass |
| `go test -cover` | 73.9% (target: 40%) |
| `go test -race` | ✅ Pass |
| WASM vet | N/A (mobile-specific package) |
| TODO/FIXME count | 0 |
| Non-deterministic rand | 0 occurrences (rng parameter correctly requires *rand.Rand) |
| Concrete net types | 0 occurrences |

## Issues Found

### High Severity
- [ ] **Determinism Violation** — `GetSeedFromEnv` uses `time.Now().UnixNano()` as fallback seed when VENTURE_SEED is unset, violating Coding Guideline #2 (Deterministic Generation). While the README states this is "intentional for mobile UX", it contradicts the project's architectural principle that "all randomness must use seed-based `rand.New(rand.NewSource(seed))` with no time-based seeding". This creates non-reproducible worlds when VENTURE_SEED is unset. (`seed.go:36`)
  - **Impact**: Bug reports from mobile users without VENTURE_SEED set cannot be reproduced; automated testing without explicit seed produces different results each run; violates design principle that same seed = same output.
  - **Recommendation**: Either (1) require VENTURE_SEED as mandatory with no fallback, or (2) document this as an **intentional exception** to Coding Guideline #2 for mobile UX, and ensure all test/debug documentation emphasizes VENTURE_SEED requirement.

### Medium Severity
- [ ] **Documentation Consistency** — README.md claims "Random time-based seed" as default but does not warn that this violates the core deterministic generation principle. Documentation should explicitly state this is an exception to normal procedural generation rules for mobile convenience. (`README.md:20`)
- [ ] **Error Handling** — `GetSeedFromEnv` and `GetGenreFromEnv` log warnings for invalid input but do not return errors, making it impossible for callers to detect configuration failures programmatically. Consider returning `(int64, error)` and `(string, error)` to enable error handling. (`seed.go:18, seed.go:48`)

### Low Severity
- [ ] **Code Duplication** — Both functions follow similar validation patterns (env var → parse/validate → fallback → log). Consider extracting a generic `getConfigFromEnv[T any](key string, parser func(string) (T, error), validator func(T) bool, fallback func() T, logger *logrus.Logger)` helper to reduce duplication. (`seed.go:18-78`)
- [ ] **Test Coverage Gap** — Tests verify determinism for VENTURE_GENRE (line 179-193) but do not test concurrent access to verify thread-safety claim in doc.go. Add `t.Run("concurrent", func(t *testing.T) { t.Parallel(); ... })` tests. (`seed_test.go`)
- [ ] **Benchmark Coverage** — Benchmarks exist but do not measure concurrent access patterns. Add `BenchmarkGetSeedFromEnv_Concurrent` to verify no contention on os.Getenv. (`seed_test.go:214-254`)

## Input Integration
| Input Source | Status | Notes |
|---|---|---|
| Keyboard | N/A | Config package has no input handling |
| Mouse | N/A | Config package has no input handling |
| Gamepad | N/A | Config package has no input handling |
| Touch | N/A | Config package has no input handling |
| VR | N/A | Config package has no input handling |
| Stub/Test | ✅ | Tests use os.Setenv/os.Unsetenv without mocking; sufficient for environment variable parsing |

## Menu/UI Integration
| Menu | Reachable | Input-Complete | Backing System Wired | Notes |
|---|---|---|---|---|
| N/A | N/A | N/A | N/A | Config package provides initialization utilities only; no UI |

## Test Coverage
**Coverage**: 73.9% (target: 40%)
- Missing test areas: None critical; all core paths tested
- Missing benchmarks: Concurrent access benchmarks
- Table-driven test compliance: ✅ Both test functions use table-driven patterns

## Documentation Coverage
- Package `doc.go`: ✅ Comprehensive package documentation with usage examples
- Exported symbols documented: 2/2 (100%) - Both `GetSeedFromEnv` and `GetGenreFromEnv` have godoc comments
- Complex algorithms commented: ✅ Inline comments explain validation and fallback logic
- Additional documentation: ✅ Detailed README.md with usage examples, environment variable formats, and testing guidance

## Integration Status
This package provides environment-based configuration for cmd/mobile initialization. It is a leaf utility package with no dependencies on engine, procgen, or rendering.

- System registration: N/A — Utility package, not an ECS system
- Component registration: N/A — No components defined
- Serialize/Deserialize: N/A — No persistent data
- Network sync: N/A — Local configuration only
- Genre theming: ✅ — `GetGenreFromEnv` provides genre string for procgen.GenerationParams
- Mod compatibility: N/A — Configuration occurs before mod loading

**Integration Points:**
1. **cmd/mobile/mobile.go** imports and calls both functions during initialization (line 7)
2. Seed value flows to `rand.New(rand.NewSource(worldSeed))` for procedural generation
3. Genre value flows to procgen generators via `GenerationParams.GenreID`

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ❌ | Package is mobile-specific (cmd/mobile/config); desktop uses CLI flags |
| WASM | ❌ | Package is mobile-specific; WASM uses different initialization |
| Mobile | ✅ | Primary target platform; Android/iOS builds use environment variables for seed/genre configuration |

**Build Tags:** No build tags present; package compiles on all platforms but is only imported by cmd/mobile

## Recommendations
1. **[HIGH]** Resolve determinism violation: Either (a) remove time-based fallback and require VENTURE_SEED, (b) document this as intentional exception to Coding Guideline #2 with clear warnings in README.md and doc.go, or (c) add a `--require-seed` flag that forces explicit seed for testing/CI builds (`seed.go:36`)
2. **[HIGH]** Update documentation to explicitly state that time-based seed fallback is an exception to deterministic generation principles and should not be used for testing/debugging (`README.md:20`, `doc.go:15-17`)
3. **[MED]** Consider returning errors from both functions to enable programmatic error detection instead of silent fallback with warning logs (`seed.go:18, seed.go:48`)
4. **[LOW]** Add concurrent access tests and benchmarks to validate thread-safety claims (`seed_test.go`)
5. **[LOW]** Extract common validation pattern into generic helper function to reduce code duplication (`seed.go`)
