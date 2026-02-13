# Audit: cmd/mobile/config
**Date**: 2026-02-13
**Status**: Needs Work

## Summary
Mobile configuration package providing seed and genre retrieval from environment variables with fallback to time-based seed. Package is small (75 LOC production code), well-tested (73.9% coverage), and properly integrated into mobile initialization. One medium-severity issue regarding time-based seed usage requires documentation clarification, and missing doc.go reduces discoverability.

## Issues Found
- [ ] med deterministic_procgen — Uses `time.Now()` for fallback seed generation without exemption documentation (`seed.go:32`)
- [ ] low doc_coverage — Missing `doc.go` package documentation file (package root)
- [ ] low integration — No explicit validation that genre list matches `pkg/procgen/genre` registry (seed.go:44-73)
- [ ] low error_handling — Genre validation loop could benefit from early return optimization (`seed.go:47-56`)

## Test Coverage
73.9% (target: 65%) ✅

**Breakdown:**
- Unit tests: `seed_test.go` with 7 table-driven test cases per function
- Integration tests: `integration_test.go` with 6 scenarios including determinism validation
- Benchmarks: 4 benchmarks covering both valid and fallback paths
- Edge cases: Max int64, negative seeds, overflow, invalid strings all covered

## Integration Status
✅ **Properly integrated** into mobile initialization flow:
- Called in `cmd/mobile/mobile.go:34-37` during `init()`
- Seed used for terrain generation, RNG initialization, and system setup
- Genre passed to procgen systems via `GenerationParams.GenreID`
- Logger integration functional (optional parameter pattern)

**Missing integrations:** None identified

## Recommendations
1. **Add package documentation** (`doc.go`) explaining mobile-specific config pattern vs desktop CLI flags and clarifying that time-based seed is intentional fallback behavior (not a violation of determinism rules when user doesn't specify explicit seed)
2. **Document exemption status** in package godoc if mobile config is exempt from strict determinism (like network/auth packages) OR document that fallback is acceptable for user-facing defaults
3. **Genre validation improvement** (optional): Import genre list from `pkg/procgen/genre.Registry()` instead of hardcoded slice in `cmd/mobile/mobile.go:35`
4. **Optimization** (minor): Use early return in genre validation loop to avoid unnecessary iterations after match found

## Notes
- No stub code, TODO, FIXME, or placeholder comments found ✅
- No ECS components (configuration package) ✅
- No network types (not applicable) ✅
- Proper structured logging with `logrus.WithFields` ✅
- All exported functions have godoc comments ✅
- Error handling follows best practices (no swallowed errors) ✅
- Table-driven tests with comprehensive edge case coverage ✅
- Benchmarks present for performance-critical paths ✅

## Time-Based Seed Justification
The use of `time.Now().UnixNano()` on line 32 is a **fallback mechanism** for mobile users who don't set `VENTURE_SEED` environment variable. This differs from procedural generation code where all randomness must be seed-based. The pattern here is:
1. User specifies explicit seed → deterministic ✅
2. User doesn't specify seed → time-based for variety ✅

Similar to desktop client accepting `-seed` flag with random fallback. If this is intentional design, it should be documented in package godoc to clarify it's not a determinism violation.
