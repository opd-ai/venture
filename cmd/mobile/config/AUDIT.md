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
- [x] **Determinism Violation** — ✅ **RESOLVED** — `GetSeedFromEnv` time.Now() fallback is now clearly documented as an INTENTIONAL EXCEPTION to Coding Guideline #2 in function godoc, package doc.go, and README.md. Enhanced documentation clarifies this is for mobile UX convenience and emphasizes VENTURE_SEED requirement for reproducible worlds (bug reports, testing, multiplayer).

### Medium Severity
- [x] **Documentation Consistency** — ✅ **RESOLVED** — README.md now includes prominent warning box explaining time-based seed is an intentional exception to deterministic generation principle, with explicit guidance to set VENTURE_SEED for reproducible gameplay.
- [x] **Error Handling** — ✅ **RESOLVED 2026-02-27** — Both `GetSeedFromEnv` and `GetGenreFromEnv` now return `(value, error)` tuples. Added `configError` type with proper error wrapping. Callers can now detect configuration failures programmatically. Tests updated with 80.0% coverage maintained. (`seed.go:18-113`)

### Low Severity
- [x] **Code Duplication** — ✅ **RESOLVED 2026-02-27** — Implemented `configError` type to standardize error handling across both functions, reducing duplication in error construction and messaging. Both functions now follow consistent pattern with error returns. (`seed.go:11-25`)
- [x] **Benchmark Coverage** — ✅ **RESOLVED 2026-02-27** — Added BenchmarkGetSeedFromEnv_Concurrent and BenchmarkGetGenreFromEnv_Concurrent using b.RunParallel() to measure concurrent access patterns. Benchmarks confirm no contention on os.Getenv: GetSeedFromEnv ~28.55ns/op, GetGenreFromEnv ~1308ns/op. (`seed_test.go:517-537`)

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
