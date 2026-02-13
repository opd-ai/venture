# Audit: github.com/opd-ai/venture/pkg/benchmark/fps
**Date**: 2026-02-13
**Status**: Complete

## Summary
Dedicated FPS benchmark package providing 4 benchmarks and 2 validation tests for the 60 FPS @ 2000 entities performance target. Code quality is excellent with no stub implementations, proper component usage, and comprehensive documentation. One medium-severity integration gap: CI/CD workflows reference `pkg/engine/` benchmarks instead of this dedicated package, reducing the value of having separated FPS/memory test infrastructure.

## Issues Found
- [ ] medium integration — CI workflow `.github/workflows/benchmark.yml` uses `pkg/engine/BenchmarkWorldUpdateWith2000Entities` instead of `pkg/benchmark/fps/BenchmarkFPS2000Entities`, defeating purpose of dedicated FPS package separation mentioned in README (`benchmark.yml:44`)
- [ ] low doc-coverage — Package has excellent godoc and README but no `go test -cover` percentage (N/A for test-only package, but could report if production code existed) (N/A)
- [ ] low test-pattern — Validation tests use direct benchmark calls instead of table-driven pattern (acceptable for performance threshold tests, but less flexible for multiple scenarios) (`fps_test.go:165-247`)

## Test Coverage
N/A (test-only package with benchmarks and validation tests, no production code to cover)

## Integration Status
**Purpose**: Validates 60 FPS performance target referenced in `docs/PERFORMANCE.md`, `docs/BENCHMARK_WORKFLOW.md`, and across 50+ documentation files.

**Dependencies**:
- Imports: `github.com/opd-ai/venture/pkg/engine` (ECS world, components, systems)
- Uses: `PositionComponent`, `VelocityComponent`, `HealthComponent`, `StatsComponent`, `ColliderComponent`, `MovementSystem`

**Integrations**:
- ✅ **Documentation**: Referenced in package README, `docs/BENCHMARK_WORKFLOW.md` (lines 16-26), separation from `pkg/benchmark/memory/` explained
- ❌ **CI/CD**: NOT used in `.github/workflows/benchmark.yml` — workflow still calls legacy `pkg/engine/performance_targets_test.go` benchmarks
- ✅ **Engine Tests**: Properly uses engine components and ECS world API
- ✅ **Benchmark Infrastructure**: Separated from memory benchmarks for cleaner CI reporting (per `fps/README.md` line 69)

**Missing Registrations**: None required (benchmark/test package)

**Critical Gap**: CI integration — the dedicated FPS package exists but is not referenced in the benchmark workflow, reducing its utility and potentially creating drift between "official" benchmarks (in workflow) and "dedicated" benchmarks (in this package).

## Recommendations
1. **HIGH PRIORITY**: Update `.github/workflows/benchmark.yml` to use `pkg/benchmark/fps/BenchmarkFPS2000Entities` instead of `pkg/engine/BenchmarkWorldUpdateWith2000Entities` to align with documented separation of FPS/memory benchmarks
2. **MEDIUM**: Add `TestFPS60TargetWith5000Entities` validation test to complement existing stress benchmark `BenchmarkFPS5000Entities` for CI pass/fail thresholds at scale
3. **LOW**: Consider adding table-driven test pattern to validation tests for easier addition of future entity count scenarios (100, 500, 1000, 2000, 5000)
4. **LOW**: Document expected ns/op ranges in README based on historical benchmark runs for regression detection
5. **INFO**: All components are pure data (ECS compliance verified), no deterministic procgen applies (benchmark package), no network types present, error handling appropriate for test code
