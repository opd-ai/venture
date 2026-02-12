# Performance Benchmarks Workflow

This document describes the automated performance validation workflow that ensures Venture meets its documented performance targets.

## Overview

The `benchmark.yml` workflow validates two critical performance targets documented in `docs/ARCHITECTURE.md`:

1. **60 FPS minimum framerate** with 2000 entities
2. **<500MB client memory** usage

## Workflow Jobs

### 1. FPS Target Validation (`fps-validation`)

**Purpose:** Validates that the ECS framework can update 2000 entities at 60+ FPS.

**Benchmark:** `BenchmarkWorldUpdateWith2000Entities`
- Creates 2000 entities with Position, Velocity, Health, and Stats components
- Uses MovementSystem for realistic update overhead
- Measures nanoseconds per frame update

**Target:** ≤16,666,666 ns/op (60 FPS)
**Current Performance:** ~130,000 ns/op (~7,800 FPS) ✅

**Failure Condition:** If ns/op exceeds 16,666,666 (below 60 FPS)

### 2. Memory Target Validation (`memory-validation`)

**Purpose:** Validates that 2000 entities stay within the 500MB memory budget.

**Test:** `TestMemoryBounds2000Entities`
- Creates 2000 entities with multiple components
- Measures heap allocation after stabilization
- Compares against 500MB target

**Target:** <500MB heap allocation
**Current Performance:** ~2.5MB (0.5% of budget) ✅

**Failure Condition:** If heap allocation ≥500MB

### 3. Comprehensive Benchmarks (`comprehensive-benchmarks`)

**Purpose:** Runs all engine benchmarks for regression tracking.

**Action:** Executes `go test -bench=. -benchmem ./pkg/engine/`
- Captures all benchmark results for historical comparison
- Uploads results as artifacts (30-day retention)
- Helps identify performance regressions across all systems

## Trigger Conditions

The workflow runs on:
- Pull requests to `main` branch
- Pushes to `main` branch
- Manual trigger via `workflow_dispatch`

## Artifacts

All jobs upload their results as GitHub Actions artifacts:
- `fps-benchmark-results` - FPS validation output
- `memory-test-results` - Memory test output
- `comprehensive-benchmark-results` - All benchmark results

Artifacts are retained for 30 days for historical analysis.

## Interpreting Results

### FPS Calculation

```bash
FPS = 1,000,000,000 ns/s ÷ ns/op
```

Example:
- 16,666,666 ns/op = 60 FPS (minimum target)
- 130,000 ns/op = 7,692 FPS (current performance)

### Memory Measurement

The test uses `runtime.MemStats.Alloc` after `runtime.GC()` to measure heap allocation:

```go
runtime.GC()
var mem runtime.MemStats
runtime.ReadMemStats(&mem)
totalMemoryMB := float64(mem.Alloc) / (1024 * 1024)
```

## Performance Targets Documentation

The targets validated by this workflow are documented in:
- `docs/ARCHITECTURE.md` - Performance requirements (60 FPS, <500MB)
- Project README.md context section - Performance guidelines

## Benchmark Implementation

### Test File: `pkg/engine/performance_targets_test.go`

Contains three key functions:

1. **`BenchmarkWorldUpdateWith2000Entities`**
   - Validates core ECS update performance
   - Uses MovementSystem only (lightweight)
   - Measures frame update time with 2000 entities

2. **`TestMemoryBounds2000Entities`**
   - Validates memory usage under target
   - Creates realistic entity mix
   - Measures heap after stabilization

3. **`BenchmarkWorldUpdateWith2000EntitiesRealistic`**
   - More comprehensive benchmark with varied components
   - Includes multiple component types per entity
   - Useful for regression detection

### Design Decisions

**Why MovementSystem only?**
- Collision detection performance depends heavily on spatial partitioning configuration
- MovementSystem represents core ECS overhead without spatial complexity
- Collision is tested separately with specific spatial grid setups

**Why 2000 entities?**
- Documented target in ARCHITECTURE.md
- Represents realistic gameplay scenario
- Historical performance benchmarks used this count

**Why separate memory test?**
- Memory measurement requires different tooling (`runtime.MemStats`)
- Benchmarks focus on time, tests validate absolute bounds
- Allows clear pass/fail criteria for CI

## Related Files

- `.github/workflows/benchmark.yml` - Workflow definition
- `pkg/engine/performance_targets_test.go` - Test implementation
- `docs/ARCHITECTURE.md` - Performance target documentation
- `scripts/benchmark-baseline.json` - Additional benchmark baselines
- `scripts/benchmark-regression.sh` - Regression detection script

## Future Enhancements

Potential improvements to the benchmark workflow:

1. **Trend Analysis:** Compare against previous runs to detect gradual regressions
2. **Platform Matrix:** Run on multiple OS/architectures
3. **Collision Benchmarks:** Add optimized collision system benchmarks
4. **Rendering Benchmarks:** Add sprite generation/rendering performance tests
5. **Network Benchmarks:** Add packet serialization performance tests

## Maintenance

When updating performance targets:

1. Update targets in `docs/ARCHITECTURE.md`
2. Update `MAX_NS_PER_FRAME` in workflow (line 62)
3. Update memory target check in workflow (line 147)
4. Update this documentation

When adding new benchmarks:

1. Add to `pkg/engine/*_test.go` with `Benchmark` prefix
2. Consider adding to `scripts/benchmark-baseline.json`
3. Update workflow if specific validation is needed
4. Update this documentation
